package db

import "testing"

func TestNormalizeRolCatalogKeyAliases(t *testing.T) {
	cases := map[string]string{
		"Administrador":         "admin_empresa",
		"Administrador empresa": "admin_empresa",
		"Servicio de limpieza":  "servicio_limpieza",
		"servicio_limpieza":     "servicio_limpieza",
		"Jefe de bodega":        "jefe_bodega",
		"Técnico solar":         "tecnico_solar",
		"Dueño":                 "empresario",
	}
	for input, want := range cases {
		if got := normalizeRolCatalogKey(input); got != want {
			t.Fatalf("normalizeRolCatalogKey(%q) = %q, want %q", input, got, want)
		}
	}
}
