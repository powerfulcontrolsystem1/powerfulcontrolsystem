package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinanzasDoesNotExposeRawAPIErrorBodies(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "finanzas.html"))
	if err != nil {
		t.Fatalf("read finanzas UI: %v", err)
	}
	page := string(raw)
	if strings.Contains(page, "const txt = await res.text();") {
		t.Fatal("finanzas must not render raw API response bodies")
	}
	for _, expected := range []string{
		"res.status === 401",
		"Tu sesión venció. Inicia sesión nuevamente para continuar.",
		"res.status === 403",
		"No tienes permiso para realizar esta acción.",
	} {
		if !strings.Contains(page, expected) {
			t.Fatalf("missing safe finance error contract %q", expected)
		}
	}
}
