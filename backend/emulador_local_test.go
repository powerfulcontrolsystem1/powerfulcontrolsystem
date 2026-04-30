package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalEmulatorHandlerServesIndexAndROMCatalog(t *testing.T) {
	juegosDir := filepath.Join("..", "juegos")
	handler := http.StripPrefix("/emulador", localEmulatorHandler(juegosDir))
	mustRegisterLocalEmulatorMIMETypes()

	indexReq := httptest.NewRequest(http.MethodGet, "/emulador/", nil)
	indexRec := httptest.NewRecorder()
	handler.ServeHTTP(indexRec, indexReq)
	if indexRec.Code != http.StatusOK {
		t.Fatalf("GET /emulador/ status = %d, want %d", indexRec.Code, http.StatusOK)
	}
	if body := indexRec.Body.String(); body == "" || !strings.Contains(body, "Juegos - Emulador Web") {
		t.Fatalf("GET /emulador/ did not return emulator index")
	}

	romsReq := httptest.NewRequest(http.MethodGet, "/emulador/api/roms", nil)
	romsRec := httptest.NewRecorder()
	handler.ServeHTTP(romsRec, romsReq)
	if romsRec.Code != http.StatusOK {
		t.Fatalf("GET /emulador/api/roms status = %d, want %d", romsRec.Code, http.StatusOK)
	}
	if contentType := romsRec.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("GET /emulador/api/roms content-type = %q, want json", contentType)
	}

	loaderReq := httptest.NewRequest(http.MethodGet, "/emulador/emulator/data/loader.js", nil)
	loaderRec := httptest.NewRecorder()
	handler.ServeHTTP(loaderRec, loaderReq)
	if loaderRec.Code != http.StatusOK {
		t.Fatalf("GET /emulador/emulator/data/loader.js status = %d, want %d", loaderRec.Code, http.StatusOK)
	}
	if contentType := loaderRec.Header().Get("Content-Type"); !strings.Contains(contentType, "javascript") {
		t.Fatalf("GET /emulador/emulator/data/loader.js content-type = %q, want javascript", contentType)
	}
	if body := loaderRec.Body.String(); !strings.Contains(body, "EJS_pathtodata") {
		t.Fatalf("GET /emulador/emulator/data/loader.js did not return EmulatorJS loader")
	}
}
