package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextualHelpStaticHandlerInjectsHTML(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!doctype html><html><body><h1>Modulo</h1></body></html>"))
	})
	req := httptest.NewRequest(http.MethodGet, "/administrar_empresa/modulo.html", nil)
	rr := httptest.NewRecorder()

	contextualHelpStaticHandler(next).ServeHTTP(rr, req)

	body := rr.Body.String()
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(body, `/js/contextual_help.js`) {
		t.Fatalf("expected contextual help script in body: %s", body)
	}
	if strings.Count(body, `/js/contextual_help.js`) != 1 {
		t.Fatalf("expected one contextual help script, got body: %s", body)
	}
}

func TestContextualHelpStaticHandlerLeavesAssetsUntouched(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write([]byte("body{color:red}"))
	})
	req := httptest.NewRequest(http.MethodGet, "/estilos.css", nil)
	rr := httptest.NewRecorder()

	contextualHelpStaticHandler(next).ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, `contextual_help.js`) {
		t.Fatalf("asset response was modified: %s", body)
	}
	if body != "body{color:red}" {
		t.Fatalf("asset body = %q", body)
	}
}
