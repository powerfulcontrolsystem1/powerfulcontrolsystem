package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestControlElectricoScenesUseExistingActivationQueue(t *testing.T) {
	raw, err := os.ReadFile("control_electrico.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, marker := range []string{"executeEmpresaControlElectricoScene", "controlElectricoDispatchManual", `fmt.Sprintf("escena:%d", sceneID)`} {
		if !strings.Contains(source, marker) {
			t.Fatalf("ejecucion de escenas sin marcador %q", marker)
		}
	}
}

func TestControlElectricoScenesVisibleInAutomationUI(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "control_electrico.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(raw)
	for _, marker := range []string{"Escenas", "sceneDeviceList", "saveSceneBtn", "scene-run", "ejecutar_escena"} {
		if !strings.Contains(html, marker) {
			t.Fatalf("UI de escenas sin marcador %q", marker)
		}
	}
}
