package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAIChatEntryPointsExposeOnlyServerOwnedAgentSwitch(t *testing.T) {
	pages := []string{
		"administrar_empresa.html",
		"seleccionar_empresa.html",
		"super_administrador.html",
		"venta_publica.html",
	}
	for _, pageName := range pages {
		raw, err := os.ReadFile(filepath.Join("..", "..", "web", pageName))
		if err != nil {
			t.Fatalf("read %s: %v", pageName, err)
		}
		page := string(raw)
		if strings.Contains(page, `aria-label="Modo del asistente IA"`) || strings.Contains(page, `Ayudante por pasos`) {
			t.Fatalf("%s still exposes a user-selectable assistant mode", pageName)
		}
		if !strings.Contains(page, `<input id="aiChatMode" type="hidden" value="operativo">`) {
			t.Fatalf("%s must keep the internal operational mode hidden", pageName)
		}
	}

	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "js", "ai_chat_drawer.js"))
	if err != nil {
		t.Fatalf("read drawer: %v", err)
	}
	drawer := string(raw)
	for _, marker := range []string{
		`modeEl.tagName || '').toLowerCase() === 'select'`,
		`hiddenMode.type = 'hidden'`,
		`aria-label="Activar modo agente"`,
	} {
		if !strings.Contains(drawer, marker) {
			t.Fatalf("chat drawer missing single-switch marker %q", marker)
		}
	}
}
