package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
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

func TestModelErrorHandlerReturnsVisibleHTML(t *testing.T) {
	h := modelErrorHandler(errors.New(`downloaded model is not a GGUF file: magic "<!do"`))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d want 503", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("content-type = %q want text/html", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Model configuration error") {
		t.Fatalf("body missing title: %q", body)
	}
	if !strings.Contains(body, "MODEL_URL") {
		t.Fatalf("body missing MODEL_URL guidance: %q", body)
	}
	if strings.Contains(body, `magic "<!do"`) {
		t.Fatalf("body did not escape HTML-sensitive error text: %q", body)
	}
}

func TestModelErrorHandlerReturnsJSONForAPI(t *testing.T) {
	h := modelErrorHandler(errors.New("download model: HTTP 404"))
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d want 503", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("content-type = %q want application/json", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"type":"model_configuration_error"`) {
		t.Fatalf("body missing error type: %q", body)
	}
	if !strings.Contains(body, "download model: HTTP 404") {
		t.Fatalf("body missing error detail: %q", body)
	}
}

func TestModelProgressHandlerReturnsVisibleHTML(t *testing.T) {
	h := modelProgressHandler(startupSnapshot{
		phase:   "Downloading model",
		message: "Downloading model 1.0 / 2.0 GiB (50%).",
		written: 1024,
		total:   2048,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d want 200", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("content-type = %q want text/html", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Downloading model") {
		t.Fatalf("body missing phase: %q", body)
	}
	if !strings.Contains(body, `value="1024" max="2048"`) {
		t.Fatalf("body missing progress element: %q", body)
	}
	if !strings.Contains(body, `http-equiv="refresh"`) {
		t.Fatalf("body missing refresh: %q", body)
	}
}

func TestModelProgressHandlerReturnsJSONForAPI(t *testing.T) {
	h := modelProgressHandler(startupSnapshot{
		phase:   "Downloading model",
		message: "Downloading model.",
		written: 7,
		total:   11,
	})
	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	req.Header.Set("Accept", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d want 503", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("content-type = %q want application/json", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"status":"preparing"`) {
		t.Fatalf("body missing status: %q", body)
	}
	if !strings.Contains(body, `"writtenBytes":7`) || !strings.Contains(body, `"totalBytes":11`) {
		t.Fatalf("body missing byte counts: %q", body)
	}
}

func TestStartupStateServesReadyHandler(t *testing.T) {
	state := newStartupState()
	state.setReady(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "ready")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	state.ServeHTTP(rr, req)

	if rr.Body.String() != "ready" {
		t.Fatalf("body = %q want ready", rr.Body.String())
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

func TestBuildLlamaArgs_SpeculativeModeDisabled(t *testing.T) {
	env := map[string]string{
		"SPECULATIVE_MODE":      "none",
		"SPEC_DRAFT_MODEL_PATH": "/models/mtp.gguf",
	}
	got := buildLlamaArgs("8081", []string{"--model", "/models/main.gguf"}, mapEnv(env))
	want := []string{"--host", "127.0.0.1", "--port", "8081", "--model", "/models/main.gguf"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v want %#v", got, want)
	}
}

func TestBuildLlamaArgs_AppendsTemperature(t *testing.T) {
	env := map[string]string{
		"LLAMA_ARG_TEMP": "0.35",
	}
	got := buildLlamaArgs("8081", []string{"--model", "/models/main.gguf"}, mapEnv(env))
	want := []string{
		"--host", "127.0.0.1", "--port", "8081",
		"--model", "/models/main.gguf",
		"--temp", "0.35",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v want %#v", got, want)
	}
}

func TestBuildLlamaArgs_DoesNotDuplicateTemperature(t *testing.T) {
	env := map[string]string{
		"LLAMA_ARG_TEMP": "0.35",
	}
	got := buildLlamaArgs("8081", []string{"--model", "/models/main.gguf", "--temp", "0.8"}, mapEnv(env))
	want := []string{
		"--host", "127.0.0.1", "--port", "8081",
		"--model", "/models/main.gguf",
		"--temp", "0.8",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v want %#v", got, want)
	}
}

func TestAppendModelArgIfMissing(t *testing.T) {
	got := appendModelArgIfMissing([]string{"--ctx-size", "4096"}, "/models/model.gguf")
	want := []string{"--ctx-size", "4096", "--model", "/models/model.gguf"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v want %#v", got, want)
	}

	got = appendModelArgIfMissing([]string{"--model", "/custom/model.gguf"}, "/models/model.gguf")
	want = []string{"--model", "/custom/model.gguf"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args with existing model = %#v want %#v", got, want)
	}
}

func TestEnsureModelDownloadsAndCaches(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = io.WriteString(w, "GGUF-data")
	}))
	defer srv.Close()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.gguf")
	env := map[string]string{
		"MODEL_URL":  srv.URL + "/model.gguf",
		"MODEL_PATH": modelPath,
	}

	got, err := ensureModel(context.Background(), mapEnv(env), srv.Client())
	if err != nil {
		t.Fatalf("ensureModel: %v", err)
	}
	if got != modelPath {
		t.Fatalf("model path = %q want %q", got, modelPath)
	}
	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("read model: %v", err)
	}
	if string(data) != "GGUF-data" {
		t.Fatalf("model data = %q", data)
	}
	if hits != 1 {
		t.Fatalf("hits after first download = %d want 1", hits)
	}

	got, err = ensureModel(context.Background(), mapEnv(env), srv.Client())
	if err != nil {
		t.Fatalf("ensureModel cached: %v", err)
	}
	if got != modelPath {
		t.Fatalf("cached path = %q want %q", got, modelPath)
	}
	if hits != 1 {
		t.Fatalf("cached ensureModel should not re-download, hits=%d", hits)
	}
}

func TestEnsureModelReportsDownloadProgress(t *testing.T) {
	body := "GGUF-data"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.gguf")
	env := map[string]string{
		"MODEL_URL":  srv.URL + "/model.gguf",
		"MODEL_PATH": modelPath,
	}
	var events []startupSnapshot
	_, err := ensureModelWithProgress(context.Background(), mapEnv(env), srv.Client(), func(written, total int64) {
		events = append(events, startupSnapshot{written: written, total: total})
	})
	if err != nil {
		t.Fatalf("ensureModel: %v", err)
	}
	if len(events) < 2 {
		t.Fatalf("events = %#v want start and progress events", events)
	}
	last := events[len(events)-1]
	if last.written != int64(len(body)) || last.total != int64(len(body)) {
		t.Fatalf("last event = %#v want written/total %d", last, len(body))
	}
}

func TestEnsureModelReplacesWhenURLChanges(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "GGUF-"+strings.TrimPrefix(r.URL.Path, "/"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.gguf")
	env := map[string]string{
		"MODEL_URL":  srv.URL + "/first.gguf",
		"MODEL_PATH": modelPath,
	}
	if _, err := ensureModel(context.Background(), mapEnv(env), srv.Client()); err != nil {
		t.Fatalf("first ensureModel: %v", err)
	}
	env["MODEL_URL"] = srv.URL + "/second.gguf"
	if _, err := ensureModel(context.Background(), mapEnv(env), srv.Client()); err != nil {
		t.Fatalf("second ensureModel: %v", err)
	}
	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("read replaced model: %v", err)
	}
	if string(data) != "GGUF-second.gguf" {
		t.Fatalf("model data = %q want GGUF-second.gguf", data)
	}
}

func TestEnsureModelNormalizesHuggingFaceBlobURL(t *testing.T) {
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.gguf")
	env := map[string]string{
		"MODEL_URL":  "https://huggingface.co/bartowski/model/blob/main/subdir/model.gguf?download=true",
		"MODEL_PATH": modelPath,
	}
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://huggingface.co/bartowski/model/resolve/main/subdir/model.gguf?download=true" {
			t.Fatalf("download URL = %s", req.URL.String())
		}
		body := "GGUF-data"
		return &http.Response{
			StatusCode:    http.StatusOK,
			Body:          io.NopCloser(strings.NewReader(body)),
			ContentLength: int64(len(body)),
			Header:        make(http.Header),
			Request:       req,
		}, nil
	})}

	if _, err := ensureModel(context.Background(), mapEnv(env), client); err != nil {
		t.Fatalf("ensureModel: %v", err)
	}
}

func TestEnsureModelRejectsNonGGUFDownload(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, "<!doctype html>")
	}))
	defer srv.Close()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.gguf")
	env := map[string]string{
		"MODEL_URL":  srv.URL + "/model.gguf",
		"MODEL_PATH": modelPath,
	}

	_, err := ensureModel(context.Background(), mapEnv(env), srv.Client())
	if err == nil || !strings.Contains(err.Error(), "not a GGUF") {
		t.Fatalf("err = %v want GGUF validation error", err)
	}
	if _, statErr := os.Stat(modelPath); !os.IsNotExist(statErr) {
		t.Fatalf("model file stat err = %v want not exist", statErr)
	}
}

func TestEnsureModelReplacesInvalidCachedFile(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits++
		_, _ = io.WriteString(w, "GGUF-fixed")
	}))
	defer srv.Close()

	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.gguf")
	env := map[string]string{
		"MODEL_URL":  srv.URL + "/model.gguf",
		"MODEL_PATH": modelPath,
	}
	if err := os.WriteFile(modelPath, []byte("<!doctype html>"), 0o640); err != nil {
		t.Fatalf("write cached model: %v", err)
	}
	if err := os.WriteFile(modelPath+".url", []byte(env["MODEL_URL"]+"\n"), 0o640); err != nil {
		t.Fatalf("write marker: %v", err)
	}

	if _, err := ensureModel(context.Background(), mapEnv(env), srv.Client()); err != nil {
		t.Fatalf("ensureModel: %v", err)
	}
	if hits != 1 {
		t.Fatalf("hits = %d want 1", hits)
	}
	data, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatalf("read model: %v", err)
	}
	if string(data) != "GGUF-fixed" {
		t.Fatalf("model data = %q want GGUF-fixed", data)
	}
}

func TestEnsureModelRequiresURL(t *testing.T) {
	_, err := ensureModel(context.Background(), mapEnv(map[string]string{}), http.DefaultClient)
	if err == nil || !strings.Contains(err.Error(), "MODEL_URL") {
		t.Fatalf("err = %v want MODEL_URL error", err)
	}
}

func TestBuildLlamaArgs_DraftMTP(t *testing.T) {
	env := map[string]string{
		"SPECULATIVE_MODE":          "draft-mtp",
		"SPEC_DRAFT_MODEL_PATH":     "/models/mtp.gguf",
		"SPEC_DRAFT_MAX_TOKENS":     "8",
		"SPEC_DRAFT_MIN_TOKENS":     "1",
		"SPEC_DRAFT_P_MIN":          "0.60",
		"SPEC_DRAFT_P_SPLIT":        "0.10",
		"SPEC_DRAFT_CACHE_TYPE_K":   "q8_0",
		"SPEC_DRAFT_CACHE_TYPE_V":   "q8_0",
		"SPEC_DRAFT_GPU_LAYERS":     "999",
		"SPEC_DRAFT_CPU_MOE_LAYERS": "0",
	}
	got := buildLlamaArgs("8081", []string{"--model", "/models/main.gguf"}, mapEnv(env))
	want := []string{
		"--host", "127.0.0.1", "--port", "8081",
		"--model", "/models/main.gguf",
		"--spec-type", "draft-mtp",
		"--model-draft", "/models/mtp.gguf",
		"--spec-draft-n-max", "8",
		"--spec-draft-n-min", "1",
		"--spec-draft-p-min", "0.60",
		"--spec-draft-p-split", "0.10",
		"--cache-type-k-draft", "q8_0",
		"--cache-type-v-draft", "q8_0",
		"--n-gpu-layers-draft", "999",
		"--n-cpu-moe-draft", "0",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v want %#v", got, want)
	}
}

func TestBuildLlamaArgs_DraftMTPOmitsEmptyDraftModel(t *testing.T) {
	env := map[string]string{
		"SPECULATIVE_MODE":      "draft-mtp",
		"SPEC_DRAFT_MODEL_PATH": "   ",
		"SPEC_DRAFT_MAX_TOKENS": "4",
	}
	got := buildLlamaArgs("8081", nil, mapEnv(env))
	want := []string{
		"--host", "127.0.0.1", "--port", "8081",
		"--spec-type", "draft-mtp",
		"--spec-draft-n-max", "4",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v want %#v", got, want)
	}
}

func mapEnv(values map[string]string) envGetter {
	return func(key string) string {
		return values[key]
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
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
