package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCentroIAEmpresarialRendersAIResponseSafely(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "centro_ia_empresarial.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(raw)
	for _, marker := range []string{
		"function renderIAResponse(value)",
		"var lines = String(value == null ? \"\" : value).replace(/\\r\\n?/g,\"\\n\").split(\"\\n\");",
		"function inline(raw){ return esc(raw)",
		"<h3>",
		"<li>",
		"renderIAResponse(ia.respuesta)",
	} {
		if !strings.Contains(page, marker) {
			t.Fatalf("Centro IA must keep safe formatted response marker %q", marker)
		}
	}
	if strings.Contains(page, `$("resultado").textContent = ia.respuesta`) {
		t.Fatal("Centro IA must not regress to literal markdown output")
	}
}
