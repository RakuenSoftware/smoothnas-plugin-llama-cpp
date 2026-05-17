// wrapper is a tiny reverse-proxy + bearer-auth gate that fronts
// upstream llama.cpp's llama-server inside the SmoothNAS plugin
// container. Upstream llama-server has no built-in auth; the wrapper
// adds the Authorization: Bearer check the SmoothNAS phase-7 nginx
// route relies on, then proxies valid requests through to the
// child server on 127.0.0.1.
//
// Configuration:
//
//	SMOOTHNAS_BEARER_EXPECTED   The bearer token the wrapper accepts.
//	                             Must match the token tierd injects via
//	                             the nginx route. tierd's "rotate token"
//	                             flow updates this env var and restarts
//	                             the container.
//
//	LLAMA_BIN                   Path to the upstream llama-server binary.
//	                             Default: /app/llama-server
//
//	LLAMA_PORT                  Port the upstream llama-server listens on
//	                             (loopback only). Default: 8081.
//
//	LLAMA_ARG_TEMP              Optional default sampling temperature.
//	                             Appended as --temp unless the manifest
//	                             already supplies --temp explicitly.
//
//	MODEL_URL                   Required model download URL. The wrapper
//	                             downloads this into MODEL_PATH before
//	                             starting llama-server.
//
//	MODEL_PATH                  Private in-container model destination.
//	                             Default: /models/model.gguf.
//
//	LISTEN_ADDR                 Address the wrapper itself binds. Default:
//	                             :8080. Matches the manifest's exposed port.
//
//	SPECULATIVE_MODE            llama.cpp speculative decoding mode. "none"
//	                             disables wrapper-managed speculative flags.
//	                             "draft-mtp" enables MTP decoding.
//
//	SPEC_DRAFT_MODEL_PATH       Optional draft/MTP GGUF path inside the
//	                             container. Empty means no separate draft
//	                             model flag is passed.
//
// Everything past the recognised wrapper flags is passed verbatim to
// llama-server, so operators can supply --model / --n-gpu-layers / etc.
// from the manifest's container.command.
package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	listenAddr := envOr("LISTEN_ADDR", ":8080")
	llamaBin := envOr("LLAMA_BIN", "/app/llama-server")
	llamaPort := envOr("LLAMA_PORT", "8081")
	expected := os.Getenv("SMOOTHNAS_BEARER_EXPECTED")
	if expected == "" {
		log.Fatal("SMOOTHNAS_BEARER_EXPECTED is empty; refusing to start without auth")
	}

	// Spawn upstream llama-server. SIGTERM/SIGINT to the wrapper
	// propagates: we catch it, signal the child, wait briefly, exit.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	modelPath, err := ensureModel(ctx, os.Getenv, http.DefaultClient)
	if err != nil {
		log.Fatalf("prepare model: %v", err)
	}
	llamaArgs := buildLlamaArgs(llamaPort, os.Args[1:], os.Getenv)
	llamaArgs = appendModelArgIfMissing(llamaArgs, modelPath)

	cmd := exec.CommandContext(ctx, llamaBin, llamaArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatalf("start %s: %v", llamaBin, err)
	}
	log.Printf("started %s pid=%d on 127.0.0.1:%s", llamaBin, cmd.Process.Pid, llamaPort)

	// Wait until upstream is ready before flipping the listener live.
	// Otherwise the very first SmoothNAS UI request races the child
	// startup and 502s. We poll with a short tick — typical first
	// model load takes minutes, but the HTTP listener comes up in
	// milliseconds.
	if err := waitForUpstream(ctx, "127.0.0.1:"+llamaPort, 30*time.Second); err != nil {
		log.Printf("upstream readiness wait: %v (proceeding anyway)", err)
	}

	upstream, err := url.Parse("http://127.0.0.1:" + llamaPort)
	if err != nil {
		log.Fatalf("parse upstream url: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(upstream)
	// llama.cpp returns SSE for streaming completions; the default
	// transport buffers, breaking token-by-token UX. Disable buffering
	// for the proxied responses.
	proxy.FlushInterval = -1
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, context.Canceled) {
			return
		}
		log.Printf("proxy error %s %s: %v", r.Method, r.URL.Path, err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           authHandler(expected, proxy),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Run the server in a goroutine; main goroutine waits on either
	// the upstream child exiting (we should die too) or the context
	// being cancelled (SIGTERM/SIGINT — we shut down cleanly).
	serverDone := make(chan error, 1)
	go func() { serverDone <- srv.ListenAndServe() }()
	log.Printf("wrapper listening on %s; bearer auth required", listenAddr)

	childDone := make(chan error, 1)
	go func() { childDone <- cmd.Wait() }()

	select {
	case err := <-childDone:
		log.Printf("upstream exited: %v", err)
	case <-ctx.Done():
		log.Printf("signal received; shutting down")
	case err := <-serverDone:
		log.Printf("listener exited: %v", err)
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(shutdownCtx)
	if cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		_, _ = cmd.Process.Wait()
	}
}

// authHandler wraps next so every request must carry
// Authorization: Bearer <expected>. Comparison is constant-time so
// the wrapper doesn't leak token length or contents through timing.
func authHandler(expected string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		got := strings.TrimPrefix(auth, "Bearer ")
		if subtle.ConstantTimeCompare([]byte(got), []byte(expected)) != 1 {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// waitForUpstream returns nil once a TCP connection to addr succeeds
// or the context cancels. Used at boot to delay the wrapper's
// listener until the child server can actually answer requests.
func waitForUpstream(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	d := &net.Dialer{Timeout: 500 * time.Millisecond}
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for %s", addr)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			conn, err := d.DialContext(ctx, "tcp", addr)
			if err == nil {
				_ = conn.Close()
				return nil
			}
		}
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

type envGetter func(string) string

const defaultModelPath = "/models/model.gguf"

func ensureModel(ctx context.Context, getenv envGetter, client *http.Client) (string, error) {
	rawURL := strings.TrimSpace(getenv("MODEL_URL"))
	if rawURL == "" {
		return "", errors.New("MODEL_URL is empty; configure a GGUF download URL")
	}
	downloadURL, err := normalizeModelURL(rawURL)
	if err != nil {
		return "", err
	}

	modelPath := strings.TrimSpace(getenv("MODEL_PATH"))
	if modelPath == "" {
		modelPath = defaultModelPath
	}
	if !filepath.IsAbs(modelPath) {
		return "", fmt.Errorf("MODEL_PATH must be absolute")
	}

	markerPath := modelPath + ".url"
	if modelReady(modelPath, markerPath, downloadURL) {
		log.Printf("using cached model %s", modelPath)
		return modelPath, nil
	}

	if client == nil {
		client = http.DefaultClient
	}
	if err := os.MkdirAll(filepath.Dir(modelPath), 0o750); err != nil {
		return "", fmt.Errorf("create model directory: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(modelPath), ".model-*.download")
	if err != nil {
		return "", fmt.Errorf("create temporary model file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) //nolint:errcheck // success path renames it away

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("build model request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		_ = tmp.Close()
		return "", fmt.Errorf("download model: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		_ = tmp.Close()
		return "", fmt.Errorf("download model: HTTP %d", resp.StatusCode)
	}

	written, err := copyModel(tmp, resp.Body, resp.ContentLength)
	closeErr := tmp.Close()
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", fmt.Errorf("close model file: %w", closeErr)
	}
	if written == 0 {
		return "", errors.New("downloaded model is empty")
	}
	if err := validateGGUFFile(tmpPath); err != nil {
		return "", err
	}
	if err := os.Chmod(tmpPath, 0o640); err != nil {
		return "", fmt.Errorf("chmod model file: %w", err)
	}
	if err := os.Rename(tmpPath, modelPath); err != nil {
		return "", fmt.Errorf("install model file: %w", err)
	}
	if err := os.WriteFile(markerPath, []byte(downloadURL+"\n"), 0o640); err != nil {
		return "", fmt.Errorf("write model url marker: %w", err)
	}
	log.Printf("installed model %s from %s", modelPath, downloadURL)
	return modelPath, nil
}

func normalizeModelURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u == nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return "", fmt.Errorf("MODEL_URL must be an http or https URL")
	}
	if strings.EqualFold(u.Hostname(), "huggingface.co") && strings.Contains(u.Path, "/blob/") {
		u.Path = strings.Replace(u.Path, "/blob/", "/resolve/", 1)
		u.RawPath = ""
	}
	return u.String(), nil
}

func modelReady(modelPath, markerPath, rawURL string) bool {
	info, err := os.Stat(modelPath)
	if err != nil || info.Size() <= 0 {
		return false
	}
	if err := validateGGUFFile(modelPath); err != nil {
		return false
	}
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == rawURL
}

func validateGGUFFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open downloaded model: %w", err)
	}
	defer f.Close()

	magic := make([]byte, 4)
	n, err := io.ReadFull(f, magic)
	if err != nil {
		return fmt.Errorf("downloaded model is not a GGUF file: read %d magic bytes: %w", n, err)
	}
	if string(magic) != "GGUF" {
		return fmt.Errorf("downloaded model is not a GGUF file: magic %q", string(magic))
	}
	return nil
}

func copyModel(dst io.Writer, src io.Reader, total int64) (int64, error) {
	buf := make([]byte, 1024*1024)
	var written int64
	lastProgress := time.Now()
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, err := dst.Write(buf[:n]); err != nil {
				return written, fmt.Errorf("write model file: %w", err)
			}
			written += int64(n)
			if time.Since(lastProgress) >= 10*time.Second {
				if total > 0 {
					log.Printf("downloading model %.1f / %.1f GiB", gib(written), gib(total))
				} else {
					log.Printf("downloading model %.1f GiB", gib(written))
				}
				lastProgress = time.Now()
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return written, fmt.Errorf("read model download: %w", readErr)
		}
	}
	return written, nil
}

func gib(v int64) float64 {
	return float64(v) / (1024 * 1024 * 1024)
}

func buildLlamaArgs(llamaPort string, passthrough []string, getenv envGetter) []string {
	args := append([]string{"--host", "127.0.0.1", "--port", llamaPort}, passthrough...)
	args = appendEnvArgIfMissing(args, getenv, "LLAMA_ARG_TEMP", "--temp")
	args = append(args, speculativeArgs(getenv)...)
	return args
}

func appendModelArgIfMissing(args []string, modelPath string) []string {
	if hasFlag(args, "--model") {
		return args
	}
	return append(args, "--model", modelPath)
}

func appendEnvArgIfMissing(args []string, getenv envGetter, envKey, flag string) []string {
	if hasFlag(args, flag) {
		return args
	}
	if v := strings.TrimSpace(getenv(envKey)); v != "" {
		return append(args, flag, v)
	}
	return args
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag || strings.HasPrefix(arg, flag+"=") {
			return true
		}
	}
	return false
}

func speculativeArgs(getenv envGetter) []string {
	mode := strings.TrimSpace(getenv("SPECULATIVE_MODE"))
	if mode == "" || mode == "none" {
		return nil
	}

	args := []string{"--spec-type", mode}
	appendIfSet := func(envKey, flag string) {
		if v := strings.TrimSpace(getenv(envKey)); v != "" {
			args = append(args, flag, v)
		}
	}

	appendIfSet("SPEC_DRAFT_MODEL_PATH", "--model-draft")
	appendIfSet("SPEC_DRAFT_MAX_TOKENS", "--spec-draft-n-max")
	appendIfSet("SPEC_DRAFT_MIN_TOKENS", "--spec-draft-n-min")
	appendIfSet("SPEC_DRAFT_P_MIN", "--spec-draft-p-min")
	appendIfSet("SPEC_DRAFT_P_SPLIT", "--spec-draft-p-split")
	appendIfSet("SPEC_DRAFT_CACHE_TYPE_K", "--cache-type-k-draft")
	appendIfSet("SPEC_DRAFT_CACHE_TYPE_V", "--cache-type-v-draft")
	appendIfSet("SPEC_DRAFT_GPU_LAYERS", "--n-gpu-layers-draft")
	appendIfSet("SPEC_DRAFT_CPU_MOE_LAYERS", "--n-cpu-moe-draft")

	return args
}
