package db

import "testing"

func TestNormalizeProduccionRecetaDefaults(t *testing.T) {
	got := normalizeProduccionReceta(EmpresaProduccionReceta{
		Codigo:          " bom-001 ",
		Nombre:          "  Kit terminado  ",
		Unidad:          "",
		CantidadBase:    -4,
		MermaPorcentaje: 150,
		Estado:          "raro",
	})
	if got.Codigo != "BOM-001" || got.Nombre != "Kit terminado" {
		t.Fatalf("campos base no normalizados: %#v", got)
	}
	if got.ProductoTerminadoNombre != "Kit terminado" || got.Unidad != "und" || got.CantidadBase != 1 {
		t.Fatalf("defaults de receta inesperados: %#v", got)
	}
	if got.MermaPorcentaje != 100 || got.Estado != "activo" {
		t.Fatalf("catalogos de receta inesperados: merma=%v estado=%q", got.MermaPorcentaje, got.Estado)
	}
}

func TestNormalizeProduccionOrdenPermiteFlujoProfesional(t *testing.T) {
	got := normalizeProduccionOrden(EmpresaProduccionOrden{
		Codigo:              " op-77 ",
		CantidadPlanificada: 0,
		CantidadProducida:   -1,
		Estado:              "En proceso",
		Prioridad:           "URGENTE",
	})
	if got.Codigo != "OP-77" {
		t.Fatalf("codigo de orden inesperado: %q", got.Codigo)
	}
	if got.CantidadPlanificada != 1 || got.CantidadProducida != 0 {
		t.Fatalf("cantidades normalizadas incorrectas: %#v", got)
	}
	if got.Estado != "en_proceso" || got.Prioridad != "urgente" {
		t.Fatalf("estado/prioridad inesperados: estado=%q prioridad=%q", got.Estado, got.Prioridad)
	}
}

func TestNormalizeProduccionCalidad(t *testing.T) {
	got := normalizeProduccionCalidad(EmpresaProduccionCalidad{
		Resultado:         "APROBADO",
		CantidadAprobada:  -10,
		CantidadRechazada: 2,
		Responsable:       " Calidad ",
	})
	if got.Resultado != "aprobado" || got.CantidadAprobada != 0 || got.CantidadRechazada != 2 || got.Responsable != "Calidad" {
		t.Fatalf("calidad no normalizada: %#v", got)
	}
}
