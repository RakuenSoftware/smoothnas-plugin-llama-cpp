package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuthHandler_RejectsMissingBearer(t *testing.T) {
	h := authHandler("secret", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("downstream handler should not run")
	}))
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d want 401", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "missing bearer") {
		t.Errorf("body = %q", rr.Body.String())
	}
}

func TestAuthHandler_RejectsWrongBearer(t *testing.T) {
	h := authHandler("secret", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("downstream handler should not run")
	}))
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d want 401", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "invalid bearer") {
		t.Errorf("body = %q", rr.Body.String())
	}
}

func TestAuthHandler_AcceptsRightBearer(t *testing.T) {
	called := false
	h := authHandler("secret", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "downstream-ok")
	}))
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set("Authorization", "Bearer secret")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d want 200", rr.Code)
	}
	if !called {
		t.Error("downstream handler should have run")
	}
	if rr.Body.String() != "downstream-ok" {
		t.Errorf("body = %q", rr.Body.String())
	}
}

func TestAuthHandler_NotPrefixedBearerRejected(t *testing.T) {
	// "Basic ..." should land in the missing-bearer branch since the
	// prefix check fails before the value compare. Same for empty.
	h := authHandler("secret", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("downstream handler should not run")
	}))
	for _, header := range []string{"", "Basic c2VjcmV0", "bearer secret"} {
		req := httptest.NewRequest(http.MethodGet, "/foo", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Authorization=%q: status = %d want 401", header, rr.Code)
		}
	}
}

func TestWaitForUpstream_ReturnsWhenListenerOpens(t *testing.T) {
	// Bring up a TCP listener after a short delay; ensure waitForUpstream
	// returns once the connect succeeds.
	addr := "127.0.0.1:" + freePort(t)
	go func() {
		time.Sleep(100 * time.Millisecond)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			t.Errorf("listen: %v", err)
			return
		}
		t.Cleanup(func() { _ = ln.Close() })
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := waitForUpstream(ctx, addr, 1500*time.Millisecond); err != nil {
		t.Errorf("wait: %v", err)
	}
}

func TestWaitForUpstream_TimesOutOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err := waitForUpstream(ctx, "127.0.0.1:1", 5*time.Second)
	if err == nil {
		t.Error("expected an error on context cancel")
	}
}

func TestEnvOr(t *testing.T) {
	if got := envOr("DEFINITELY_NOT_SET_ABCDEF", "fallback"); got != "fallback" {
		t.Errorf("got %q, want fallback", got)
	}
	t.Setenv("WRAPPER_TEST_KEY", "explicit")
	if got := envOr("WRAPPER_TEST_KEY", "fallback"); got != "explicit" {
		t.Errorf("got %q, want explicit", got)
	}
}

// freePort grabs an OS-assigned ephemeral port and returns it as a
// string. The listener is closed immediately so the test code can
// re-bind to the same port.
func freePort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().(*net.TCPAddr)
	_ = ln.Close()
	// Tiny race window between Close and re-bind, but the test
	// timeline is short enough that it never fires in practice.
	return numToStr(addr.Port)
}

func numToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
