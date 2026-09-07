package db

import "testing"

func TestNormalizeAlquilerTipoActivoUniversal(t *testing.T) {
	cases := map[string]string{
		"":                      "equipo",
		"motos":                 "moto",
		"Motocicleta":           "moto",
		"Herramienta Electrica": "herramienta_electrica",
		"muebles":               "mobiliario",
		"sonido eventos":        "sonido_eventos",
		"audio":                 "sonido_eventos",
		"tecnologia":            "tecnologia",
		"objeto":                "objeto",
		"desconocido":           "objeto",
	}
	for input, want := range cases {
		if got := normalizeAlquilerTipoActivo(input); got != want {
			t.Fatalf("normalizeAlquilerTipoActivo(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeAlquilerTipoRegistro(t *testing.T) {
	cases := map[string]string{
		"":              "alquiler",
		"renta":         "alquiler",
		"reserva":       "reserva",
		"presupuesto":   "cotizacion",
		"devolucion":    "devolucion",
		"mantenimiento": "mantenimiento",
		"otro":          "alquiler",
	}
	for input, want := range cases {
		if got := normalizeAlquilerTipoRegistro(input); got != want {
			t.Fatalf("normalizeAlquilerTipoRegistro(%q) = %q, want %q", input, got, want)
		}
	}
}
