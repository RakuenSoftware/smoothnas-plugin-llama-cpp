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
//	LLAMA_ARG_FLASH_ATTN        Optional flash-attention mode. Appended as
//	                             --flash-attn unless already supplied.
//
//	LLAMA_ARG_CACHE_TYPE_K      Optional K cache quantization. Appended as
//	                             --cache-type-k unless already supplied.
//
//	LLAMA_ARG_CACHE_TYPE_V      Optional V cache quantization. Appended as
//	                             --cache-type-v unless already supplied.
//
//	LLAMA_ARG_N_CPU_MOE         Optional MoE expert layers kept on CPU.
//	                             Appended as --n-cpu-moe unless supplied.
//
//	LLAMA_ARG_FIT               Optional llama.cpp fit mode. Appended as
//	                             --fit unless already supplied.
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
	"sync"
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

	state := newStartupState()
	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           authHandler(expected, state),
		ReadHeaderTimeout: 10 * time.Second,
	}
	serverDone := make(chan error, 1)
	go func() { serverDone <- srv.ListenAndServe() }()
	log.Printf("wrapper listening on %s; bearer auth required", listenAddr)

	modelPath, err := ensureModelWithProgress(ctx, os.Getenv, http.DefaultClient, state.setDownloadProgress)
	if err != nil {
		log.Printf("prepare model: %v", err)
		state.setError(err)
		waitForServerExit(ctx, srv, serverDone)
		return
	}
	state.setPhase("Starting llama.cpp", "Model downloaded. Starting llama.cpp server.", 0, 0)
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
	state.setReady(proxy)

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

func waitForServerExit(ctx context.Context, srv *http.Server, serverDone <-chan error) {
	select {
	case <-ctx.Done():
		log.Printf("signal received; shutting down")
	case err := <-serverDone:
		log.Printf("listener exited: %v", err)
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = srv.Shutdown(shutdownCtx)
}

type startupState struct {
	mu      sync.RWMutex
	phase   string
	message string
	written int64
	total   int64
	err     error
	ready   http.Handler
}

type startupSnapshot struct {
	phase   string
	message string
	written int64
	total   int64
	err     error
	ready   http.Handler
}

func newStartupState() *startupState {
	return &startupState{
		phase:   "Preparing model",
		message: "Preparing configured model.",
	}
}

func (s *startupState) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snap := s.snapshot()
	if snap.ready != nil {
		snap.ready.ServeHTTP(w, r)
		return
	}
	if snap.err != nil {
		modelErrorHandler(snap.err).ServeHTTP(w, r)
		return
	}
	modelProgressHandler(snap).ServeHTTP(w, r)
}

func (s *startupState) setPhase(phase, message string, written, total int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.phase = phase
	s.message = message
	s.written = written
	s.total = total
}

func (s *startupState) setDownloadProgress(written, total int64) {
	s.setPhase("Downloading model", modelDownloadMessage(written, total), written, total)
}

func (s *startupState) setError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.err = err
	s.ready = nil
}

func (s *startupState) setReady(handler http.Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = handler
	s.err = nil
}

func (s *startupState) snapshot() startupSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return startupSnapshot{
		phase:   s.phase,
		message: s.message,
		written: s.written,
		total:   s.total,
		err:     s.err,
		ready:   s.ready,
	}
}

func modelProgressHandler(snap startupSnapshot) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wantsJSON(r) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintf(w, `{"status":"preparing","phase":%q,"message":%q,"writtenBytes":%d,"totalBytes":%d}`+"\n",
				snap.phase, snap.message, snap.written, snap.total)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="refresh" content="5">
  <title>llama.cpp starting</title>
  <style>
    body { margin: 0; font-family: system-ui, sans-serif; color: #202124; background: #f7f7f5; }
    main { max-width: 720px; margin: 10vh auto; padding: 32px; background: #fff; border: 1px solid #ddd; border-radius: 8px; }
    h1 { margin: 0 0 12px; font-size: 24px; }
    p { line-height: 1.5; }
    progress { width: 100%%; height: 18px; }
  </style>
</head>
<body>
<main>
  <h1>%s</h1>
  <p>%s</p>
  %s
</main>
</body>
</html>
`, htmlEscape(snap.phase), htmlEscape(snap.message), progressElement(snap))
	})
}

func wantsJSON(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept"), "application/json") || strings.HasPrefix(r.URL.Path, "/v1/")
}

func modelErrorHandler(prepareErr error) http.Handler {
	message := modelErrorMessage(prepareErr)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if wantsJSON(r) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = fmt.Fprintf(w, `{"error":{"message":%q,"type":"model_configuration_error"}}`+"\n", message)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>llama.cpp model configuration error</title>
  <style>
    body { margin: 0; font-family: system-ui, sans-serif; color: #202124; background: #f7f7f5; }
    main { max-width: 720px; margin: 10vh auto; padding: 32px; background: #fff; border: 1px solid #ddd; border-radius: 8px; }
    h1 { margin: 0 0 12px; font-size: 24px; }
    p { line-height: 1.5; }
    code { background: #f0f0ed; padding: 2px 5px; border-radius: 4px; }
  </style>
</head>
<body>
<main>
  <h1>Model configuration error</h1>
  <p>%s</p>
  <p>Update <code>MODEL_URL</code> to a direct HTTP(S) GGUF download URL, then restart the plugin.</p>
</main>
</body>
</html>
`, htmlEscape(message))
	})
}

func modelErrorMessage(prepareErr error) string {
	if prepareErr == nil {
		return "The configured model could not be prepared."
	}
	return "The configured MODEL_URL could not be prepared: " + prepareErr.Error()
}

func modelDownloadMessage(written, total int64) string {
	if total > 0 {
		percent := (float64(written) / float64(total)) * 100
		if percent > 100 {
			percent = 100
		}
		return fmt.Sprintf("Downloading model %.1f / %.1f GiB (%.0f%%).", gib(written), gib(total), percent)
	}
	if written > 0 {
		return fmt.Sprintf("Downloading model %.1f GiB.", gib(written))
	}
	return "Downloading model."
}

func progressElement(snap startupSnapshot) string {
	if snap.total > 0 {
		value := snap.written
		if value > snap.total {
			value = snap.total
		}
		return fmt.Sprintf(`<progress value="%d" max="%d"></progress>`, value, snap.total)
	}
	return `<progress></progress>`
}

func htmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
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
type progressFunc func(written, total int64)

const defaultModelPath = "/models/model.gguf"

func ensureModel(ctx context.Context, getenv envGetter, client *http.Client) (string, error) {
	return ensureModelWithProgress(ctx, getenv, client, nil)
}

func ensureModelWithProgress(ctx context.Context, getenv envGetter, client *http.Client, progress progressFunc) (string, error) {
	rawURL := strings.TrimSpace(getenv("MODEL_URL"))
	if rawURL == "" {
		return "", errors.New("MODEL_URL is empty; configure a GGUF download URL")
	}
	downloadURL, err := normalizeModelURL(rawURL)
	if err != nil {
		return "", err
	}
	cacheKey := modelCacheKey(downloadURL)

	modelPath := strings.TrimSpace(getenv("MODEL_PATH"))
	if modelPath == "" {
		modelPath = defaultModelPath
	}
	if !filepath.IsAbs(modelPath) {
		return "", fmt.Errorf("MODEL_PATH must be absolute")
	}

	markerPath := modelPath + ".url"
	if modelReady(modelPath, markerPath, cacheKey, downloadURL) {
		log.Printf("using cached model %s", modelPath)
		return modelPath, nil
	}

	if client == nil {
		client = http.DefaultClient
	}
	if modelMarkerEmpty(markerPath) && adoptCachedModel(ctx, client, modelPath, markerPath, downloadURL, cacheKey) {
		return modelPath, nil
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

	written, err := copyModel(tmp, resp.Body, resp.ContentLength, progress)
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
	if err := os.WriteFile(markerPath, []byte(cacheKey+"\n"), 0o640); err != nil {
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

func modelCacheKey(downloadURL string) string {
	u, err := url.Parse(downloadURL)
	if err != nil || u == nil {
		return downloadURL
	}
	q := u.Query()
	for k := range q {
		if isVolatileModelURLParam(k) {
			delete(q, k)
		}
	}
	u.RawQuery = q.Encode()
	u.Fragment = ""
	return u.String()
}

func isVolatileModelURLParam(key string) bool {
	k := strings.ToLower(key)
	return k == "download" ||
		k == "expires" ||
		k == "signature" ||
		k == "policy" ||
		k == "key-pair-id" ||
		k == "awsaccesskeyid" ||
		k == "response-content-disposition" ||
		k == "response-content-type" ||
		k == "response-cache-control" ||
		k == "response-expires" ||
		strings.HasPrefix(k, "x-amz-") ||
		strings.HasPrefix(k, "x-goog-")
}

func modelReady(modelPath, markerPath string, cacheKeys ...string) bool {
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
	marker := strings.TrimSpace(string(data))
	for _, key := range cacheKeys {
		if marker == key || modelCacheKey(marker) == key {
			return true
		}
	}
	return false
}

func modelMarkerEmpty(markerPath string) bool {
	data, err := os.ReadFile(markerPath)
	return os.IsNotExist(err) || (err == nil && strings.TrimSpace(string(data)) == "")
}

func adoptCachedModel(ctx context.Context, client *http.Client, modelPath, markerPath, downloadURL, cacheKey string) bool {
	info, err := os.Stat(modelPath)
	if err != nil || info.Size() <= 0 {
		return false
	}
	if err := validateGGUFFile(modelPath); err != nil {
		return false
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, downloadURL, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("cached model probe failed: %v", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 || resp.ContentLength <= 0 {
		return false
	}
	if resp.ContentLength != info.Size() {
		return false
	}

	if err := os.WriteFile(markerPath, []byte(cacheKey+"\n"), 0o640); err != nil {
		log.Printf("using cached model %s, but could not write URL marker: %v", modelPath, err)
		return true
	}
	log.Printf("using cached model %s after matching remote size", modelPath)
	return true
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

func copyModel(dst io.Writer, src io.Reader, total int64, progress progressFunc) (int64, error) {
	buf := make([]byte, 1024*1024)
	var written int64
	lastProgress := time.Now()
	if progress != nil {
		progress(0, total)
	}
	for {
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, err := dst.Write(buf[:n]); err != nil {
				return written, fmt.Errorf("write model file: %w", err)
			}
			written += int64(n)
			if progress != nil {
				progress(written, total)
			}
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
	args = appendEnvArgIfMissing(args, getenv, "LLAMA_ARG_FLASH_ATTN", "--flash-attn")
	args = appendEnvArgIfMissing(args, getenv, "LLAMA_ARG_CACHE_TYPE_K", "--cache-type-k")
	args = appendEnvArgIfMissing(args, getenv, "LLAMA_ARG_CACHE_TYPE_V", "--cache-type-v")
	args = appendEnvArgIfMissing(args, getenv, "LLAMA_ARG_N_CPU_MOE", "--n-cpu-moe")
	args = appendEnvArgIfMissing(args, getenv, "LLAMA_ARG_FIT", "--fit")
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
