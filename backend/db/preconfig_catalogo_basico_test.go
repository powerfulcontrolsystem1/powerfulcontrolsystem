package db

import "testing"

func TestPreconfiguracionesBasicasAprobadasExactas(t *testing.T) {
	aprobadas := []string{
		"Hotel",
		"Motel",
		"Restaurante",
		"Bar",
		"Pymes",
		"Salon de belleza",
		"Lavadero de autos",
	}
	for _, nombre := range aprobadas {
		if !IsTipoEmpresaPreconfiguracionBasica(nombre) {
			t.Fatalf("%q debe ser una preconfiguracion basica", nombre)
		}
	}

	retiradas := []string{
		"Parqueadero",
		"Domicilios",
		"Alquileres",
		"Construccion / AIU",
		"Eventos y boleteria",
		"Salon, barberia y spa",
		"Veterinaria y pet shop",
		"Lavanderia y tintoreria",
		"Taller mecanico",
		"Transporte de carga / TMS",
		"Servicios tecnicos",
		"Funeraria y servicios exequiales",
		"Parque recreativo",
	}
	for _, nombre := range retiradas {
		if IsTipoEmpresaPreconfiguracionBasica(nombre) {
			t.Fatalf("%q no debe pertenecer al catalogo basico", nombre)
		}
	}
}

func TestCatalogoAdicionalSoloConservaTallerDeMotos(t *testing.T) {
	items := NuevasPlantillasTipoEmpresaCatalog()
	if len(items) != 1 || items[0].Modulo != "taller_mecanico" || items[0].Nombre != "Taller de motos" {
		t.Fatalf("catalogo adicional inesperado: %+v", items)
	}
}
