package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedactEmailForLog(t *testing.T) {
	tests := map[string]string{
		"ana@example.com":   "a***@example.com",
		"a@example.com":     "a***@example.com",
		" invalid-address ": "[redacted]",
		"":                  "[redacted]",
	}
	for input, want := range tests {
		if got := redactEmailForLog(input); got != want {
			t.Fatalf("redactEmailForLog(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSensitiveLogValuesAreRedacted(t *testing.T) {
	if got := redactPersonalDocumentForLog("1.234.567.890"); got != "1***90" {
		t.Fatalf("document redaction = %q", got)
	}
	if got := redactPhoneForLog("+57 300 123 4567"); got != "5***67" {
		t.Fatalf("phone redaction = %q", got)
	}
	if got := redactTokenForLog("very-secret-token"); got != "[redacted-token]" {
		t.Fatalf("token redaction = %q", got)
	}
	if got := redactAuthorizationForLog("Bearer very-secret-token"); got != "[authorization-present]" {
		t.Fatalf("authorization redaction = %q", got)
	}
	if got := normalizeLogValue("line1\r\nline2"); got != "line1\\r\\nline2" {
		t.Fatalf("log normalization = %q", got)
	}
}

func TestSensitiveRequestHeadersAreRedacted(t *testing.T) {
	headers := http.Header{
		"Authorization": {"Bearer secret"},
		"Cookie":        {"session=secret"},
		"X-Request-ID":  {"req-123"},
	}
	redacted := redactRequestHeadersForLog(headers)
	for key, value := range redacted {
		if strings.Contains(value, "secret") {
			t.Fatalf("header %s leaked secret: %q", key, value)
		}
	}
	if got := redacted["x-request-id"]; got != "req-123" {
		t.Fatalf("request id = %q", got)
	}
}

func TestLegacyEpaycoRuntimeToolDoesNotReadOrPrintSensitiveData(t *testing.T) {
	_, err := os.Stat(filepath.Join("..", "tools", "query_epayco_runtime.go"))
	if !os.IsNotExist(err) {
		t.Fatalf("legacy sensitive payment runtime tool must remain removed, stat err=%v", err)
	}
}

func TestLegacyAdminLoginInspectionToolRemainsRemoved(t *testing.T) {
	_, err := os.Stat(filepath.Join("..", "tools", "inspect_admin_login", "main.go"))
	if !os.IsNotExist(err) {
		t.Fatalf("legacy admin login inspection tool must remain removed, stat err=%v", err)
	}
}

func TestLegacyPlaintextRustDeskDeviceImplementationRemainsRemoved(t *testing.T) {
	for _, path := range []string{
		filepath.Join("rustdesk.go"),
		filepath.Join("..", "db", "rustdesk.go"),
	} {
		_, err := os.Stat(path)
		if !os.IsNotExist(err) {
			t.Fatalf("legacy plaintext RustDesk implementation must remain removed: %s, stat err=%v", path, err)
		}
	}
}

func TestStationPreferencesDoNotLogRawUserEmail(t *testing.T) {
	path := filepath.Join("estaciones_columnas_pref.go")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if strings.Contains(string(contents), "user=%s error: %v\", empresaID, p.UsuarioEmail") {
		t.Fatal("station preferences must redact user email before logging")
	}
}
