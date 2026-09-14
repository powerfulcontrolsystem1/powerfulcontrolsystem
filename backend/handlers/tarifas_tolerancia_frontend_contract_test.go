package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTarifasPorMinutosExplainsRecurringExtraBlockTolerance(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "tarifas_por_minutos.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read tarifas frontend: %v", err)
	}
	source := string(raw)
	for _, required := range []string{
		"Tolerancia por cada bloque extra (minutos)",
		"después de 130, 200, 270 min",
		"margen_tolerancia_entrada_minutos:",
		"Tolerancia por bloque:",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("tariff tolerance UI must keep %q", required)
		}
	}
}
