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
	for _, required := range []string{
		`if(guarded > 0) return guarded;`,
		`if(contextual > 0) return contextual;`,
		`return Number(p.get("empresa_id") || p.get("id") || 0);`,
	} {
		if !strings.Contains(page, required) {
			t.Fatalf("Centro IA must fall back to URL tenant when a visual guard resolves zero: %s", required)
		}
	}
}
