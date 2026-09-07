package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEnergiaSolarReadingFresh(t *testing.T) {
	now := time.Date(2026, 8, 13, 22, 0, 0, 0, time.FixedZone("COT", -5*60*60))
	if !energiaSolarReadingFresh("2026-08-13 21:59:15-05", 30, now) {
		t.Fatal("lectura reciente debe marcarse conectada")
	}
	if energiaSolarReadingFresh("2026-08-13 21:50:00-05", 30, now) {
		t.Fatal("lectura vencida debe marcarse desconectada")
	}
}

func TestEnergiaSolarVictronRealtimeUIMarkers(t *testing.T) {
	root := filepath.Join("..", "..", "web")
	files := []string{filepath.Join(root, "administrar_empresa", "energia_solar.html"), filepath.Join(root, "js", "energia_solar.js")}
	combined := ""
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		combined += string(raw)
	}
	for _, marker := range []string{"Estado en tiempo real", "solarRealtimeGrid", "Conectado", "Desconectado", "No disponible", "estado_cargador", "vpv_v"} {
		if !strings.Contains(combined, marker) {
			t.Fatalf("interfaz solar sin marcador %q", marker)
		}
	}
}
