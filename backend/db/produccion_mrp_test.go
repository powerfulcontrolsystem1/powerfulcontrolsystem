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

func TestValidateProduccionOrdenTransitionRespetaCalidadObligatoria(t *testing.T) {
	cfg := EmpresaProduccionMRPConfig{CerrarConCalidad: true}
	if err := validateProduccionOrdenTransition("en_proceso", "cerrada", cfg); err == nil {
		t.Fatalf("expected quality gate to block direct close")
	}
	if err := validateProduccionOrdenTransition("en_proceso", "calidad", cfg); err != nil {
		t.Fatalf("expected transition to quality to be valid: %v", err)
	}
	if err := validateProduccionOrdenTransition("calidad", "cerrada", cfg); err != nil {
		t.Fatalf("expected close after quality to be valid: %v", err)
	}
}

func TestValidateProduccionOrdenTransitionBloqueaEstadosFinales(t *testing.T) {
	cfg := EmpresaProduccionMRPConfig{}
	if err := validateProduccionOrdenTransition("cerrada", "en_proceso", cfg); err == nil {
		t.Fatalf("expected closed orders to reject further workflow changes")
	}
	if err := validateProduccionOrdenTransition("programada", "cerrada", cfg); err == nil {
		t.Fatalf("expected invalid transition to be rejected")
	}
}

func TestProduccionDateKey(t *testing.T) {
	if got := produccionDateKey("2026-05-08 10:30:00"); got != "2026-05-08" {
		t.Fatalf("date key = %q", got)
	}
	if got := produccionDateKey("  "); got != "" {
		t.Fatalf("empty date key = %q", got)
	}
}

func TestProduccionMRPDemoDefinitionsSonProfesionales(t *testing.T) {
	defs := produccionMRPDemoDefinitions("qa")
	if len(defs) < 3 {
		t.Fatalf("se esperaban varios ejemplos de produccion, got %d", len(defs))
	}
	seen := map[string]bool{}
	for _, def := range defs {
		code := normalizeProduccionReceta(def.Receta).Codigo
		if code == "" || seen[code] {
			t.Fatalf("codigo de receta demo invalido o repetido: %q", code)
		}
		seen[code] = true
		if len(def.Receta.Componentes) < 3 {
			t.Fatalf("la receta demo %s debe tener BOM completo", code)
		}
		if def.Orden.CantidadPlanificada <= 0 || def.Orden.ProductoTerminadoNombre == "" {
			t.Fatalf("orden demo incompleta para %s: %#v", code, def.Orden)
		}
	}
}
