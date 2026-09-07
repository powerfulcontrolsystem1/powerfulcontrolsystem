package db

import "testing"

func TestPeriodoRangeDeclaracionMensual(t *testing.T) {
	desde, hasta := periodoRangeDeclaracion("2026-02")
	if desde != "2026-02-01" || hasta != "2026-02-28" {
		t.Fatalf("rango inesperado: %s a %s", desde, hasta)
	}
}

func TestCalcularSaldosDeclaracionIVA(t *testing.T) {
	got := calcularSaldosDeclaracion(EmpresaDeclaracionTributaria{
		TipoDeclaracion:    "iva",
		IVAGenerado:        900000,
		IVADescontable:     320000,
		SaldoFavorAnterior: 100000,
		Sanciones:          15000,
		Intereses:          5000,
	})
	if got.SaldoPagar != 500000 || got.SaldoFavor != 0 {
		t.Fatalf("saldo IVA inesperado: pagar %.2f favor %.2f", got.SaldoPagar, got.SaldoFavor)
	}
}

func TestCalcularSaldosDeclaracionSaldoFavor(t *testing.T) {
	got := calcularSaldosDeclaracion(EmpresaDeclaracionTributaria{
		TipoDeclaracion:    "iva",
		IVAGenerado:        100000,
		IVADescontable:     250000,
		SaldoFavorAnterior: 50000,
	})
	if got.SaldoPagar != 0 || got.SaldoFavor != 200000 {
		t.Fatalf("saldo favor inesperado: pagar %.2f favor %.2f", got.SaldoPagar, got.SaldoFavor)
	}
}

func TestNormalizeDeclaracionTipo(t *testing.T) {
	cases := map[string]string{
		" ReteIVA ": "reteiva",
		"ICA":       "ica",
		"raro":      "iva",
	}
	for input, want := range cases {
		if got := normalizeDeclaracionTipo(input); got != want {
			t.Fatalf("normalizeDeclaracionTipo(%q) = %q, want %q", input, got, want)
		}
	}
}
