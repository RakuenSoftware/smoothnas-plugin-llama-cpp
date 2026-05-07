// wrapper is a tiny reverse-proxy + bearer-auth gate that fronts
// upstream llama.cpp's llama-server inside the SmoothNAS plugin
// container. Upstream llama-server has no built-in auth; the wrapper
// adds the Authorization: Bearer check the SmoothNAS phase-7 nginx
// route relies on, then proxies valid requests through to the
// child server on 127.0.0.1.
//
// Configuration:
//
//   SMOOTHNAS_BEARER_EXPECTED   The bearer token the wrapper accepts.
//                                Must match the token tierd injects via
//                                the nginx route. tierd's "rotate token"
//                                flow updates this env var and restarts
//                                the container.
//
//   LLAMA_BIN                   Path to the upstream llama-server binary.
//                                Default: /llama-server
//
//   LLAMA_PORT                  Port the upstream llama-server listens on
//                                (loopback only). Default: 8081.
//
//   LISTEN_ADDR                 Address the wrapper itself binds. Default:
//                                :8080. Matches the manifest's exposed port.
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
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	listenAddr := envOr("LISTEN_ADDR", ":8080")
	llamaBin := envOr("LLAMA_BIN", "/llama-server")
	llamaPort := envOr("LLAMA_PORT", "8081")
	expected := os.Getenv("SMOOTHNAS_BEARER_EXPECTED")
	if expected == "" {
		log.Fatal("SMOOTHNAS_BEARER_EXPECTED is empty; refusing to start without auth")
	}

	// Pass everything else through to llama-server. We don't actually
	// parse the args; we forward them verbatim.
	llamaArgs := append([]string{"--host", "127.0.0.1", "--port", llamaPort}, os.Args[1:]...)

	// Spawn upstream llama-server. SIGTERM/SIGINT to the wrapper
	// propagates: we catch it, signal the child, wait briefly, exit.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

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
