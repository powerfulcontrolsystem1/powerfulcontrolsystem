package httpguard

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllowGetOrHead(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		response := httptest.NewRecorder()
		if !AllowGetOrHead(response, httptest.NewRequest(method, "/health", nil)) {
			t.Fatalf("%s debe estar permitido", method)
		}
	}
	response := httptest.NewRecorder()
	if AllowGetOrHead(response, httptest.NewRequest(http.MethodPost, "/health", nil)) {
		t.Fatal("POST no debe estar permitido")
	}
	if response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != "GET, HEAD" {
		t.Fatalf("respuesta inesperada: status=%d allow=%q", response.Code, response.Header().Get("Allow"))
	}
}
