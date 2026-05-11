package db

import (
	"strings"
	"testing"
)

func TestDefaultNuevoVerticalLicenciaPlans(t *testing.T) {
	catalog := NuevosVerticalesTipoEmpresaCatalog()
	if len(catalog) != 20 {
		t.Fatalf("expected 20 nuevos verticales, got %d", len(catalog))
	}
	for _, item := range catalog {
		t.Run(item.Modulo, func(t *testing.T) {
			plantilla := GetEmpresaModuloColombiaPlantilla(item.Modulo)
			if item.Nombre != plantilla.Titulo {
				t.Fatalf("catalog nombre=%q want plantilla titulo=%q", item.Nombre, plantilla.Titulo)
			}
			if strings.TrimSpace(item.Observaciones) == "" || !strings.Contains(strings.ToLower(item.Observaciones), "gestion profesional") {
				t.Fatalf("observaciones derivadas invalidas: %q", item.Observaciones)
			}
			plans := DefaultNuevoVerticalLicenciaPlans(item)
			if len(plans) != 4 {
				t.Fatalf("expected 4 plans, got %d", len(plans))
			}
			if plans[0].DuracionDias != 15 || plans[0].MaxDocumentosMensuales != 250 || plans[0].Valor != 0 {
				t.Fatalf("trial plan mismatch: %+v", plans[0])
			}
			if !strings.Contains(plans[0].ModulosHabilitados, item.Modulo) {
				t.Fatalf("modules %q missing %s", plans[0].ModulosHabilitados, item.Modulo)
			}
			if !strings.Contains(plans[3].Nombre, "4000 documentos") || plans[3].MaxDocumentosMensuales != 4000 {
				t.Fatalf("4000-doc plan mismatch: %+v", plans[3])
			}
		})
	}
}

func TestNuevosVerticalesProduccionMasivaLicenciasRecomendadas(t *testing.T) {
	selected := NuevosVerticalesProduccionMasivaSeleccionados()
	if len(selected) != 10 {
		t.Fatalf("seleccion v1 len=%d want 10", len(selected))
	}
	for _, modulo := range selected {
		t.Run(modulo, func(t *testing.T) {
			item, ok := getNuevoVerticalTipoEmpresaByModulo(modulo)
			if !ok {
				t.Fatalf("vertical %s no existe en catalogo", modulo)
			}
			plans := DefaultNuevoVerticalLicenciaPlans(item)
			if len(plans) != 4 {
				t.Fatalf("planes %s len=%d want 4", modulo, len(plans))
			}
			for _, plan := range plans {
				modules := strings.Split(plan.ModulosHabilitados, ",")
				seen := map[string]bool{}
				for _, module := range modules {
					seen[strings.TrimSpace(module)] = true
				}
				for _, required := range []string{modulo, "ventas", "clientes", "facturacion", "seguridad"} {
					if !seen[required] {
						t.Fatalf("plan %q no incluye modulo requerido %q en %q", plan.Nombre, required, plan.ModulosHabilitados)
					}
				}
			}
		})
	}
}

func TestDefaultNuevoVerticalTipoEmpresaPreconfigTemplate(t *testing.T) {
	for _, item := range NuevosVerticalesTipoEmpresaCatalog() {
		t.Run(item.Modulo, func(t *testing.T) {
			template, ok := defaultNuevoVerticalTipoEmpresaPreconfigTemplate(item.Nombre)
			if !ok {
				t.Fatalf("expected preconfig for %s", item.Nombre)
			}
			if !template.Asistente.Enabled || len(template.Productos) < 3 || len(template.Usuarios) == 0 || len(template.TareasGuia) < 3 {
				t.Fatalf("preconfig incompleta: %+v", template)
			}
			if got := template.Operacion.TipoNegocio; got != item.Modulo {
				t.Fatalf("tipo negocio=%q want %q", got, item.Modulo)
			}
		})
	}
}
