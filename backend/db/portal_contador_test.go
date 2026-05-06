package db

import (
	"testing"
	"time"
)

func TestPortalContadorDiasParaVencer(t *testing.T) {
	today := time.Date(2026, 5, 6, 12, 0, 0, 0, time.Local)
	cases := []struct {
		name  string
		fecha string
		want  int
	}{
		{name: "vence pronto", fecha: "2026-05-10", want: 4},
		{name: "vencida", fecha: "2026-05-01", want: -5},
		{name: "sin fecha", fecha: "", want: 9999},
		{name: "invalida", fecha: "sin-fecha", want: 9999},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := portalContadorDiasParaVencer(tc.fecha, today); got != tc.want {
				t.Fatalf("portalContadorDiasParaVencer() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPortalContadorNormalizaciones(t *testing.T) {
	if got := normalizePortalContadorRegimen("simple"); got != "simple" {
		t.Fatalf("regimen simple = %q", got)
	}
	if got := normalizePortalContadorTipoObligacion("nomina_electronica"); got != "nomina_electronica" {
		t.Fatalf("tipo obligacion = %q", got)
	}
	if got := normalizePortalContadorEstadoSolicitud("desconocido"); got != "abierta" {
		t.Fatalf("estado solicitud default = %q", got)
	}
	if got := normalizePortalContadorPrioridad("critica"); got != "critica" {
		t.Fatalf("prioridad = %q", got)
	}
	if got := normalizePortalContadorCanal("whatsapp"); got != "whatsapp" {
		t.Fatalf("canal = %q", got)
	}
}
