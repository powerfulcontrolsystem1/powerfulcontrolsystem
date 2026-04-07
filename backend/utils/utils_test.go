package utils

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func withTempWorkingDir(t *testing.T) string {
	t.Helper()
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir to temp dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})
	return tmpDir
}

func TestGenerateSecureToken(t *testing.T) {
	token, err := GenerateSecureToken(16)
	if err != nil {
		t.Fatalf("GenerateSecureToken returned error: %v", err)
	}
	if len(token) != 32 {
		t.Fatalf("expected token len 32, got %d", len(token))
	}
	if _, err := hex.DecodeString(token); err != nil {
		t.Fatalf("token should be hex: %v", err)
	}
}

func TestEncryptDecryptStringRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(key))

	plain := "secreto-operativo"
	encrypted, err := EncryptString(plain)
	if err != nil {
		t.Fatalf("EncryptString returned error: %v", err)
	}
	if encrypted == plain {
		t.Fatal("expected encrypted value to differ from plain text")
	}

	decrypted, err := DecryptString(encrypted)
	if err != nil {
		t.Fatalf("DecryptString returned error: %v", err)
	}
	if decrypted != plain {
		t.Fatalf("expected %q, got %q", plain, decrypted)
	}
}

func TestEncryptStringWithoutKey(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", "")
	if _, err := EncryptString("demo"); err == nil {
		t.Fatal("expected error when CONFIG_ENC_KEY is missing")
	}
}

func TestDecryptStringInvalidPayload(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(key))
	if _, err := DecryptString("%%%invalid-base64%%%"); err == nil {
		t.Fatal("expected error for invalid encrypted payload")
	}
}

func TestEncryptionAvailable(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", "")
	if EncryptionAvailable() {
		t.Fatal("expected EncryptionAvailable=false without key")
	}

	key := []byte("0123456789abcdef0123456789abcdef")
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(key))
	if !EncryptionAvailable() {
		t.Fatal("expected EncryptionAvailable=true with valid key")
	}
}

func TestParsePositiveInt64(t *testing.T) {
	cases := []struct {
		raw  string
		want int64
	}{
		{"", 0},
		{"abc", 0},
		{"-4", 0},
		{"0", 0},
		{"15", 15},
	}

	for _, tc := range cases {
		if got := parsePositiveInt64(tc.raw); got != tc.want {
			t.Fatalf("parsePositiveInt64(%q) expected %d, got %d", tc.raw, tc.want, got)
		}
	}
}

func TestExtractEmpresaIDFromBody(t *testing.T) {
	if got := extractEmpresaIDFromBody([]byte(`{"empresa_id":12}`)); got != 12 {
		t.Fatalf("expected empresa_id=12, got %d", got)
	}
	if got := extractEmpresaIDFromBody([]byte(`{"empresa":{"id":"44"}}`)); got != 44 {
		t.Fatalf("expected empresa.id=44, got %d", got)
	}
	if got := extractEmpresaIDFromBody([]byte(`{"empresaId":7}`)); got != 7 {
		t.Fatalf("expected empresaId=7, got %d", got)
	}
	if got := extractEmpresaIDFromBody([]byte(`{"x":1}`)); got != 0 {
		t.Fatalf("expected no empresa id, got %d", got)
	}
}

func TestInferEmpresaIDFromRequestAndBodyPreserved(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/demo", strings.NewReader(`{"empresa_id":21,"foo":"bar"}`))
	req.Header.Set("Content-Type", "application/json")

	if got := inferEmpresaIDFromRequest(req); got != 21 {
		t.Fatalf("expected empresa id 21, got %d", got)
	}

	raw, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("read request body after inference: %v", err)
	}
	if !strings.Contains(string(raw), `"empresa_id":21`) {
		t.Fatalf("expected request body to remain readable, got %q", string(raw))
	}
}

func TestInferEmpresaIDFromRequestPrecedence(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/demo?empresa_id=5", strings.NewReader(`{"empresa_id":21}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Empresa-ID", "8")
	ctx := context.WithValue(req.Context(), ctxKeyEmpresaID, int64(13))
	req = req.WithContext(ctx)

	if got := inferEmpresaIDFromRequest(req); got != 5 {
		t.Fatalf("expected query empresa_id precedence with value 5, got %d", got)
	}
}

func TestTruncateLogMessage(t *testing.T) {
	if got := truncateLogMessage("  hola  ", 100); got != "hola" {
		t.Fatalf("expected trimmed text, got %q", got)
	}
	if got := truncateLogMessage("abcdef", 3); got != "abc" {
		t.Fatalf("expected truncated text 'abc', got %q", got)
	}
}

func TestRequestIDAndEmpresaIDFromContext(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxKeyRequestID, "req-demo")
	ctx = context.WithValue(ctx, ctxKeyEmpresaID, int64(23))

	if got := RequestIDFromContext(ctx); got != "req-demo" {
		t.Fatalf("expected request id req-demo, got %q", got)
	}
	if got := empresaIDFromContext(ctx); got != 23 {
		t.Fatalf("expected empresa id 23, got %d", got)
	}
}

func TestRequestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:8080"
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")

	if got := requestClientIP(req); got != "10.0.0.1" {
		t.Fatalf("expected forwarded ip 10.0.0.1, got %q", got)
	}

	req.Header.Del("X-Forwarded-For")
	req.Header.Set("X-Real-IP", "192.168.1.9")
	if got := requestClientIP(req); got != "192.168.1.9" {
		t.Fatalf("expected real ip 192.168.1.9, got %q", got)
	}
}

func TestLoggingMiddlewareSetsContextAndWritesLogs(t *testing.T) {
	tmpDir := withTempWorkingDir(t)

	var seenReqID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenReqID = RequestIDFromContext(r.Context())
		if seenReqID == "" {
			t.Fatalf("expected request id in context")
		}
		w.Header().Set("X-Empresa-ID", "77")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})

	h := LoggingMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/api/demo?empresa_id=21", nil)
	req.Header.Set("User-Agent", "utils-test")
	req.RemoteAddr = "127.0.0.1:9090"
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rr.Code)
	}
	if got := rr.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("expected X-Request-ID header")
	}
	if got := rr.Header().Get("X-Empresa-ID"); got != "77" {
		t.Fatalf("expected X-Empresa-ID=77, got %q", got)
	}
	if body := rr.Body.String(); body != "ok" {
		t.Fatalf("expected body ok, got %q", body)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "logs", "empresa_21.log")); err != nil {
		t.Fatalf("expected start log file for empresa 21: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "logs", "empresa_77.log")); err != nil {
		t.Fatalf("expected end log file for empresa 77: %v", err)
	}
}

func TestJSONErrorMiddlewareBypassesNonAPI(t *testing.T) {
	withTempWorkingDir(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("plain-error"))
	})

	h := JSONErrorMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/home", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}
	if rr.Body.String() != "plain-error" {
		t.Fatalf("expected passthrough body, got %q", rr.Body.String())
	}
}

func TestJSONErrorMiddlewareWrapsNonJSONError(t *testing.T) {
	tmpDir := withTempWorkingDir(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("recurso no encontrado"))
	})

	h := JSONErrorMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/api/demo?empresa_id=9", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
	if ct := strings.ToLower(rr.Header().Get("Content-Type")); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected JSON content type, got %q", ct)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode json body: %v body=%s", err, rr.Body.String())
	}
	if got := int(payload["status"].(float64)); got != http.StatusNotFound {
		t.Fatalf("expected status field 404, got %d", got)
	}
	if got := payload["error"].(string); !strings.Contains(got, "recurso no encontrado") {
		t.Fatalf("unexpected error field: %q", got)
	}
	if got := int(payload["empresa_id"].(float64)); got != 9 {
		t.Fatalf("expected empresa_id=9, got %d", got)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "logs", "empresa_9.log")); err != nil {
		t.Fatalf("expected api error log file for empresa 9: %v", err)
	}
}

func TestJSONErrorMiddlewarePreservesJSONErrorBody(t *testing.T) {
	withTempWorkingDir(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"dato invalido"}`))
	})

	h := JSONErrorMiddleware(next)
	req := httptest.NewRequest(http.MethodGet, "/api/demo?empresa_id=12", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status 422, got %d", rr.Code)
	}
	if body := strings.TrimSpace(rr.Body.String()); body != `{"error":"dato invalido"}` {
		t.Fatalf("expected original json error body, got %q", body)
	}
}
