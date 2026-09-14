package handlers

import (
	"strings"
	"testing"
)

func TestDefaultContextoIALogicaNegocioIncluyeCatalogoVigente(t *testing.T) {
	ctx := defaultContextoIALogicaNegocioText()

	required := []string{
		"siete preconfiguraciones basicas",
		"Hotel, Motel, Restaurante, Bar, Pymes, Salon de belleza y Lavadero de autos",
		"Taller de motos",
		"no es una octava preconfiguracion",
		"Configuracion del menu",
		"ocho licencias globales",
		"COP 200000",
		"dos compras adicionales",
	}
	for _, want := range required {
		if !strings.Contains(ctx, want) {
			t.Fatalf("default IA context is missing %q", want)
		}
	}
}
