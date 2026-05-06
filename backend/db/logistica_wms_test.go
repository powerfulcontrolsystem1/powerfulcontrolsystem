package db

import "testing"

func TestNormalizeWMSCodigo(t *testing.T) {
	cases := map[string]string{
		" bod a/p01 r01 ": "BOD-A-P01-R01",
		"pack_01":         "PACK-01",
		" A..B ":          "A-B",
	}
	for input, want := range cases {
		if got := normalizeWMSCodigo(input); got != want {
			t.Fatalf("normalizeWMSCodigo(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCalcularProgresoWMS(t *testing.T) {
	pick, pack := calcularProgresoWMS(20, 5, 2)
	if pick != 25 || pack != 10 {
		t.Fatalf("progreso inesperado pick %.2f pack %.2f", pick, pack)
	}
	pick, pack = calcularProgresoWMS(0, 5, 2)
	if pick != 0 || pack != 0 {
		t.Fatalf("progreso con total cero debe ser 0, got %.2f %.2f", pick, pack)
	}
}

func TestInferWMSItemEstado(t *testing.T) {
	if got := inferWMSItemEstado(10, 0, 0); got != "pendiente" {
		t.Fatalf("estado esperado pendiente, got %s", got)
	}
	if got := inferWMSItemEstado(10, 4, 0); got != "en_picking" {
		t.Fatalf("estado esperado en_picking, got %s", got)
	}
	if got := inferWMSItemEstado(10, 10, 0); got != "pickeado" {
		t.Fatalf("estado esperado pickeado, got %s", got)
	}
	if got := inferWMSItemEstado(10, 10, 10); got != "completado" {
		t.Fatalf("estado esperado completado, got %s", got)
	}
}
