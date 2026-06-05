package handlers

import "testing"

func TestInformacionModulosNormalizeConfig(t *testing.T) {
	cfg := informacionModulosNormalizeConfig(informacionModulosConfig{
		Titulo: "  ",
		Modulos: []informacionModuloItem{
			{
				Titulo:          "  Ventas nuevas  ",
				IconoURL:        "/img/money.svg",
				Caracteristicas: []string{"  Caja  ", "", "Caja", "POS"},
			},
			{
				Titulo:          "",
				IconoURL:        "https://externo.test/icon.png",
				Caracteristicas: []string{},
			},
		},
	})

	if cfg.Titulo != "Modulos y caracteristicas principales" {
		t.Fatalf("titulo = %q, want default", cfg.Titulo)
	}
	if got := len(cfg.Modulos); got != 5 {
		t.Fatalf("modulos len = %d, want 5 con destacados nuevos", got)
	}
	if cfg.Modulos[0].Titulo != "Ventas nuevas" {
		t.Fatalf("module title = %q", cfg.Modulos[0].Titulo)
	}
	if cfg.Modulos[0].IconoURL != "/img/money.svg" {
		t.Fatalf("icon = %q", cfg.Modulos[0].IconoURL)
	}
	if got := cfg.Modulos[0].Caracteristicas; len(got) != 2 || got[0] != "Caja" || got[1] != "POS" {
		t.Fatalf("features = %#v, want unique trimmed values", got)
	}
	if cfg.Modulos[1].IconoURL != "/img/punto_venta.png" {
		t.Fatalf("invalid icon fallback = %q", cfg.Modulos[1].IconoURL)
	}
	if len(cfg.Modulos[1].Caracteristicas) == 0 {
		t.Fatalf("second module must receive fallback features")
	}
	seen := map[string]bool{}
	for _, mod := range cfg.Modulos {
		seen[mod.Titulo] = true
	}
	for _, expected := range []string{"GRAFOLOGIX", "Camaras y DVR", "Energia solar"} {
		if !seen[expected] {
			t.Fatalf("configuracion antigua debe completar modulo destacado %q; got %#v", expected, cfg.Modulos)
		}
	}
}
