package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextualHelpStaticHandlerDoesNotInjectHTML(t *testing.T) {
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
	if strings.Contains(body, `/js/contextual_help.js`) {
		t.Fatalf("contextual help script should not be injected: %s", body)
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

func TestButtonIconsStaticHandlerInjectsHTML(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!doctype html><html><body><button class=\"btn\">Guardar</button></body></html>"))
	})
	req := httptest.NewRequest(http.MethodGet, "/administrar_empresa/modulo.html", nil)
	rr := httptest.NewRecorder()

	buttonIconsStaticHandler(next).ServeHTTP(rr, req)

	body := rr.Body.String()
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(body, `/js/button_icons.js`) {
		t.Fatalf("button icons script should be injected: %s", body)
	}
	if !strings.Contains(body, `/js/csrf_fetch.js`) {
		t.Fatalf("csrf fetch script should be injected: %s", body)
	}
	if strings.Count(body, `/js/button_icons.js`) != 1 {
		t.Fatalf("button icons script should appear once: %s", body)
	}
}

func TestButtonIconsStaticHandlerLeavesAssetsUntouched(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write([]byte("body{color:red}"))
	})
	req := httptest.NewRequest(http.MethodGet, "/estilos.css", nil)
	rr := httptest.NewRecorder()

	buttonIconsStaticHandler(next).ServeHTTP(rr, req)

	body := rr.Body.String()
	if strings.Contains(body, `button_icons.js`) {
		t.Fatalf("asset response was modified: %s", body)
	}
	if strings.Contains(body, `csrf_fetch.js`) {
		t.Fatalf("asset response was modified: %s", body)
	}
	if body != "body{color:red}" {
		t.Fatalf("asset body = %q", body)
	}
}

func TestButtonIconsStaticHandlerDoesNotDuplicateScript(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!doctype html><html><body><button>Guardar</button><script src=\"/js/button_icons.js\" defer></script></body></html>"))
	})
	req := httptest.NewRequest(http.MethodGet, "/administrar_empresa/modulo.html", nil)
	rr := httptest.NewRecorder()

	buttonIconsStaticHandler(next).ServeHTTP(rr, req)

	if got := strings.Count(rr.Body.String(), `/js/button_icons.js`); got != 1 {
		t.Fatalf("script count = %d, want 1; body=%s", got, rr.Body.String())
	}
	if got := strings.Count(rr.Body.String(), `/js/csrf_fetch.js`); got != 1 {
		t.Fatalf("csrf script count = %d, want 1; body=%s", got, rr.Body.String())
	}
}

func TestInjectButtonIconsScriptUsesDocumentHeadBeforePrintableTemplate(t *testing.T) {
	body := []byte(`<!doctype html><html><head><title>Modulo</title></head><body><script>var printable = "</body>";</script></body></html>`)
	got := string(injectButtonIconsScript(body))
	csrfAt := strings.Index(got, `/js/csrf_fetch.js`)
	headEnd := strings.Index(got, `</head>`)
	printableAt := strings.Index(got, `var printable`)
	if csrfAt < 0 || headEnd < 0 || printableAt < 0 {
		t.Fatalf("missing expected content after injection: %s", got)
	}
	if csrfAt > headEnd || csrfAt > printableAt {
		t.Fatalf("csrf helper was not injected into document head: %s", got)
	}
	if strings.Contains(got, `var printable = "<script src=`) {
		t.Fatalf("printable template was modified: %s", got)
	}
}
