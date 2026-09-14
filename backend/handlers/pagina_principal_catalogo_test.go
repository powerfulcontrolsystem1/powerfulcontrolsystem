package handlers

import "testing"

func TestPaginaPrincipalNormalizeConfigExcluyeTarjetasRetiradas(t *testing.T) {
	cfg := paginaPrincipalConfig{
		Cantidad: 10,
		Tarjetas: []paginaPrincipalCard{
			{Titulo: "Oferta vigente", Enlace: "/administrar_empresa.html?module=punto_venta"},
			{Titulo: "Taller de motos", Enlace: "/administrar_empresa.html?module=taller_mecanico"},
			{Titulo: "Oferta heredada", Enlace: "/administrar_empresa.html?module=taxi_system"},
			{Titulo: "Drogueria y farmacia", Enlace: "/login.html"},
			{Titulo: "Parqueadero", Enlace: "/administrar_empresa.html?module=parqueadero"},
			{Titulo: "Domicilios y entregas", Enlace: "/login.html"},
			{Titulo: "Alquileres de activos", Enlace: "/login.html"},
			{Titulo: "Veterinaria y pet shop", Enlace: "/login.html"},
			{Titulo: "Lavanderia y tintoreria", Enlace: "/login.html"},
			{Titulo: "Parque recreativo", Enlace: "/login.html"},
		},
	}

	got := paginaPrincipalNormalizeConfig(cfg)
	if got.Cantidad != 2 || len(got.Tarjetas) != 2 {
		t.Fatalf("se esperaban 2 tarjetas activas, cantidad=%d tarjetas=%d", got.Cantidad, len(got.Tarjetas))
	}
	if got.Tarjetas[0].Titulo != "Oferta vigente" || got.Tarjetas[1].Titulo != "Taller de motos" {
		t.Fatalf("las tarjetas vigentes fueron reemplazadas: %#v", got.Tarjetas)
	}
}

func TestInformacionModulosNormalizeConfigFijaCatalogoVigente(t *testing.T) {
	cfg := informacionModulosConfig{
		Titulo: "53 modulos activos y 13 plantillas empresariales",
		Modulos: []informacionModuloItem{
			{Titulo: "Gimnasio", Caracteristicas: []string{"Socios"}},
			{Titulo: "Plantillas listas", Caracteristicas: []string{"Hotel", "Motel"}},
		},
	}

	got := informacionModulosNormalizeConfig(cfg)
	if got.Titulo != "Modulos empresariales, 7 preconfiguraciones basicas y Taller de motos" {
		t.Fatalf("titulo vigente inesperado: %q", got.Titulo)
	}
	foundTemplates := false
	foundFeatured := false
	for _, item := range got.Modulos {
		if paginaPrincipalCardIsRetired(paginaPrincipalCard{Titulo: item.Titulo}) {
			t.Fatalf("el resumen conserva un modulo retirado: %q", item.Titulo)
		}
		if informacionModulosIsTemplatesTitle(item.Titulo) {
			if item.Titulo != "7 preconfiguraciones basicas" {
				t.Fatalf("titulo de plantillas inesperado: %q", item.Titulo)
			}
			if len(item.Caracteristicas) != 7 {
				t.Fatalf("se esperaban 7 preconfiguraciones, se obtuvieron %d", len(item.Caracteristicas))
			}
			foundTemplates = true
		}
		if item.Titulo == "Sistema destacado" {
			foundFeatured = len(item.Caracteristicas) > 0 && item.Caracteristicas[0] == "Taller de motos"
		}
	}
	if !foundTemplates {
		t.Fatal("no se publico el resumen de preconfiguraciones vigente")
	}
	if !foundFeatured {
		t.Fatal("no se publico Taller de motos como sistema destacado")
	}
}
