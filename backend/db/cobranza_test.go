package db

import (
	"testing"
	"time"
)

func TestCobranzaDiasMora(t *testing.T) {
	today := time.Date(2026, 5, 6, 12, 0, 0, 0, time.Local)
	cases := []struct {
		name     string
		vence    string
		fallback int
		want     int
	}{
		{name: "vencida", vence: "2026-05-01", want: 5},
		{name: "por vencer", vence: "2026-05-10", want: 0},
		{name: "sin fecha usa fallback", fallback: 12, want: 12},
		{name: "fallback negativo se limpia", fallback: -3, want: 0},
		{name: "fecha invalida usa fallback", vence: "sin-fecha", fallback: 7, want: 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cobranzaDiasMora(tc.vence, tc.fallback, today); got != tc.want {
				t.Fatalf("cobranzaDiasMora() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestCobranzaNormalizaciones(t *testing.T) {
	if got := normalizeCobranzaCanal("correo"); got != "email" {
		t.Fatalf("canal correo = %q", got)
	}
	if got := normalizeCobranzaResultado("promesa"); got != "promesa_pago" {
		t.Fatalf("resultado promesa = %q", got)
	}
	if got := normalizeCobranzaEstadoCampana("desconocido"); got != "borrador" {
		t.Fatalf("estado campana default = %q", got)
	}
	if got := normalizeCobranzaEstadoPromesa("cumplida"); got != "cumplida" {
		t.Fatalf("estado promesa = %q", got)
	}
	if got := normalizeCobranzaEstadoCartera("mora"); got != "vencida" {
		t.Fatalf("estado cartera mora = %q", got)
	}
}
