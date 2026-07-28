package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmpresaIAEmpresarialPOSTKeepsTenantInPermissionURL(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "centro_ia_empresarial.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	page := string(raw)
	if !strings.Contains(page, `fetch(apiUrl(), { method: "POST"`) {
		t.Fatal("Centro IA POST must include apiUrl so permission middleware receives empresa_id")
	}
	if strings.Contains(page, `fetch("/api/empresa/ia_empresarial", { method: "POST"`) {
		t.Fatal("Centro IA POST must not rely on empresa_id only inside JSON")
	}
}
