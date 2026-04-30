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
}
