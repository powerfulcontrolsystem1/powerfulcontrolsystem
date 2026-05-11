package db

import "testing"

func TestNuevosVerticalesTienenPlantillaYDemo(t *testing.T) {
	catalog := NuevosVerticalesTipoEmpresaCatalog()
	if len(catalog) != len(empresaModuloColombiaPlantillasVerticales) {
		t.Fatalf("catalogo tipos=%d plantillas=%d", len(catalog), len(empresaModuloColombiaPlantillasVerticales))
	}
	for _, item := range catalog {
		modulo := item.Modulo
		t.Run(modulo, func(t *testing.T) {
			if got := NormalizeEmpresaModuloColombia(modulo); got != modulo {
				t.Fatalf("NormalizeEmpresaModuloColombia(%q)=%q", modulo, got)
			}
			plantilla := GetEmpresaModuloColombiaPlantilla(modulo)
			if plantilla.Titulo == "" || len(plantilla.Tipos) < 2 || len(plantilla.Categorias) < 2 {
				t.Fatalf("plantilla incompleta: %+v", plantilla)
			}
			if len(plantilla.SeccionesFlujo) < 4 {
				t.Fatalf("secciones de flujo incompletas: %+v", plantilla.SeccionesFlujo)
			}
			rows := demoEmpresaModuloColombiaRows(1, modulo, "qa")
			if len(rows) < 2 {
				t.Fatalf("demo incompleto para %s: %d filas", modulo, len(rows))
			}
		})
	}
}
