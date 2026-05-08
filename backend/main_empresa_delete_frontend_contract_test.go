package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmpresaDeleteFrontendHiddenProgressContract(t *testing.T) {
	cssRaw, err := os.ReadFile(filepath.Join("..", "web", "estilos.css"))
	if err != nil {
		t.Fatalf("read estilos.css: %v", err)
	}
	css := string(cssRaw)
	requiredCSS := []string{
		".empresa-delete-progress[hidden]",
		"display:none !important",
		".selector-delete-progress[hidden]",
	}
	for _, expected := range requiredCSS {
		if !strings.Contains(css, expected) {
			t.Fatalf("estilos.css debe mantener el contrato de ocultamiento del progreso de eliminacion; falta %q", expected)
		}
	}
}

func TestEditarEmpresaDeleteButtonRequiresSafeConfirmation(t *testing.T) {
	jsRaw, err := os.ReadFile(filepath.Join("..", "web", "js", "editar_empresa.js"))
	if err != nil {
		t.Fatalf("read editar_empresa.js: %v", err)
	}
	js := string(jsRaw)
	requiredJS := []string{
		"function getDeleteValidationState()",
		"ready: nameOk && phraseOk && riskOk && !isSharedEmpresa()",
		"btn.disabled = state.deleting || !validation.ready",
		"if (!busy) {\n      updateDeleteChecklist();\n    }",
	}
	for _, expected := range requiredJS {
		if !strings.Contains(js, expected) {
			t.Fatalf("editar_empresa.js debe bloquear eliminacion hasta confirmar nombre, frase y riesgo; falta %q", expected)
		}
	}

	htmlRaw, err := os.ReadFile(filepath.Join("..", "web", "editar_empresa.html"))
	if err != nil {
		t.Fatalf("read editar_empresa.html: %v", err)
	}
	html := string(htmlRaw)
	if !strings.Contains(html, `id="empresaDeleteBtn"`) || !strings.Contains(html, `id="empresaDeleteBtn" type="button" class="btn danger" disabled`) {
		t.Fatalf("editar_empresa.html debe cargar el boton de eliminacion deshabilitado hasta completar las validaciones")
	}
}
