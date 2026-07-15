package db

import (
	"strings"
	"testing"
)

func TestDefaultNuevoVerticalLicenciaPlans(t *testing.T) {
	catalog := NuevasPlantillasTipoEmpresaCatalog()
	if len(catalog) != 19 {
		t.Fatalf("expected 19 nuevas plantillas, got %d", len(catalog))
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
			if len(plans) != 8 {
				t.Fatalf("expected 8 plans, got %d", len(plans))
			}
			if plans[0].DuracionDias != 15 || plans[0].MaxDocumentosMensuales != 250 || plans[0].Valor != 0 {
				t.Fatalf("trial plan mismatch: %+v", plans[0])
			}
			if !strings.Contains(plans[0].ModulosHabilitados, item.Modulo) {
				t.Fatalf("modules %q missing %s", plans[0].ModulosHabilitados, item.Modulo)
			}
			if plans[4].Nombre != "Plan mensual COP 200000" || plans[4].MaxDocumentosMensuales != 4000 || plans[4].Valor != 200000 {
				t.Fatalf("plan COP 200000 mismatch: %+v", plans[4])
			}
			if plans[7].Nombre != "Plan anual COP 2200000" || plans[7].MaxDocumentosMensuales != 36000 || plans[7].DuracionDias != 365 {
				t.Fatalf("plan anual COP 2200000 mismatch: %+v", plans[7])
			}
		})
	}
}

func TestNuevasPlantillasProduccionMasivaLicenciasRecomendadas(t *testing.T) {
	selected := NuevasPlantillasProduccionMasivaSeleccionados()
	if len(selected) != 19 {
		t.Fatalf("seleccion produccion len=%d want 19", len(selected))
	}
	for _, modulo := range selected {
		t.Run(modulo, func(t *testing.T) {
			item, ok := getNuevoVerticalTipoEmpresaByModulo(modulo)
			if !ok {
				t.Fatalf("plantilla %s no existe en catalogo", modulo)
			}
			plans := DefaultNuevoVerticalLicenciaPlans(item)
			if len(plans) != 8 {
				t.Fatalf("planes %s len=%d want 8", modulo, len(plans))
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
	for _, item := range NuevasPlantillasTipoEmpresaCatalog() {
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
