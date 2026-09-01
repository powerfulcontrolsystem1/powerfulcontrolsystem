package handlers

import "testing"

func TestPaginaPrincipalNormalizeConfigExcluyeTarjetasRetiradas(t *testing.T) {
	cfg := paginaPrincipalConfig{
		Cantidad: 3,
		Tarjetas: []paginaPrincipalCard{
			{Titulo: "Oferta vigente", Enlace: "/administrar_empresa.html?module=punto_venta"},
			{Titulo: "Oferta heredada", Enlace: "/administrar_empresa.html?module=taxi_system"},
			{Titulo: "Drogueria y farmacia", Enlace: "/login.html"},
		},
	}

	got := paginaPrincipalNormalizeConfig(cfg)
	if got.Cantidad != 1 || len(got.Tarjetas) != 1 {
		t.Fatalf("se esperaban 1 tarjeta activa, cantidad=%d tarjetas=%d", got.Cantidad, len(got.Tarjetas))
	}
	if got.Tarjetas[0].Titulo != "Oferta vigente" {
		t.Fatalf("la tarjeta activa fue reemplazada: %#v", got.Tarjetas[0])
	}
}

func TestInformacionModulosNormalizeConfigFijaCatalogoVigente(t *testing.T) {
	cfg := informacionModulosConfig{
		Titulo: "Resumen anterior",
		Modulos: []informacionModuloItem{
			{Titulo: "Gimnasio", Caracteristicas: []string{"Socios"}},
			{Titulo: "Plantillas listas", Caracteristicas: []string{"Hotel", "Motel"}},
		},
	}

	got := informacionModulosNormalizeConfig(cfg)
	for _, item := range got.Modulos {
		if paginaPrincipalCardIsRetired(paginaPrincipalCard{Titulo: item.Titulo}) {
			t.Fatalf("el resumen conserva un modulo retirado: %q", item.Titulo)
		}
		if informacionModulosIsTemplatesTitle(item.Titulo) {
			if item.Titulo != "13 plantillas listas" {
				t.Fatalf("titulo de plantillas inesperado: %q", item.Titulo)
			}
			if len(item.Caracteristicas) != 13 {
				t.Fatalf("se esperaban 13 plantillas, se obtuvieron %d", len(item.Caracteristicas))
			}
			return
		}
	}
	t.Fatal("no se publico el resumen de plantillas vigente")
}
