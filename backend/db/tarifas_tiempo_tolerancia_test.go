package db

import (
	"testing"
	"time"
)

func TestCalcularDetalleTarifaPorMinutosRespetaMargenTolerancia(t *testing.T) {
	tarifa := EmpresaTarifaPorMinutos{
		ID:                1,
		EmpresaID:         7,
		EstacionID:        101,
		MinutosBase:       120,
		ValorBase:         100000,
		MinutosExtra:      60,
		ValorExtra:        50000,
		CobrarPorFraccion: true,
		Moneda:            "COP",
	}
	cfg := defaultEmpresaTarifaPorMinutosConfiguracion(7)
	cfg.MargenToleranciaEntradaMinutos = 10

	dentro := CalcularDetalleTarifaPorMinutos(tarifa, 129, cfg)
	if dentro.BloquesExtra != 0 {
		t.Fatalf("2h + 9min con 10min de tolerancia no debe cobrar extra, bloques=%d", dentro.BloquesExtra)
	}
	if dentro.MontoTotal != 100000 {
		t.Fatalf("monto dentro de tolerancia = %.2f, want 100000", dentro.MontoTotal)
	}
	if dentro.MinutosFacturables != 120 {
		t.Fatalf("minutos facturables dentro de tolerancia = %.2f, want 120", dentro.MinutosFacturables)
	}

	fuera := CalcularDetalleTarifaPorMinutos(tarifa, 131, cfg)
	if fuera.BloquesExtra != 1 {
		t.Fatalf("2h + 11min con 10min de tolerancia debe cobrar 1 bloque extra, bloques=%d", fuera.BloquesExtra)
	}
	if fuera.MontoTotal != 150000 {
		t.Fatalf("monto fuera de tolerancia = %.2f, want 150000", fuera.MontoTotal)
	}
}

func TestCalcularDetalleTarifaPorMinutosRepiteToleranciaEnCadaBloqueExtra(t *testing.T) {
	tarifa := EmpresaTarifaPorMinutos{
		ID: 1, EmpresaID: 7, EstacionID: 101,
		MinutosBase: 120, ValorBase: 20000,
		MinutosExtra: 60, ValorExtra: 10000,
		CobrarPorFraccion: true, Moneda: "COP",
	}
	cfg := defaultEmpresaTarifaPorMinutosConfiguracion(7)
	cfg.MargenToleranciaEntradaMinutos = 10

	for _, tc := range []struct {
		minutos int
		bloques int
		total   float64
	}{
		{130, 0, 20000},
		{131, 1, 30000},
		{200, 1, 30000},
		{201, 2, 40000},
		{270, 2, 40000},
		{271, 3, 50000},
	} {
		got := CalcularDetalleTarifaPorMinutos(tarifa, float64(tc.minutos), cfg)
		if got.BloquesExtra != tc.bloques || got.MontoTotal != tc.total {
			t.Fatalf("%d min: bloques=%d total=%.2f; want bloques=%d total=%.2f", tc.minutos, got.BloquesExtra, got.MontoTotal, tc.bloques, tc.total)
		}
	}

	tarifa.CobrarPorFraccion = false
	for _, tc := range []struct {
		minutos int
		bloques int
	}{
		{189, 0},
		{190, 1},
		{259, 1},
		{260, 2},
	} {
		got := CalcularDetalleTarifaPorMinutos(tarifa, float64(tc.minutos), cfg)
		if got.BloquesExtra != tc.bloques {
			t.Fatalf("sin fraccion, %d min: bloques=%d; want %d", tc.minutos, got.BloquesExtra, tc.bloques)
		}
	}
}

func TestResolveCarritoTarifaPorMinutosCurrentEndIncluyeToleranciaRecurrente(t *testing.T) {
	activadoEn := time.Date(2026, time.September, 13, 12, 0, 0, 0, time.UTC)
	tarifa := EmpresaTarifaPorMinutos{MinutosBase: 120, MinutosExtra: 60}
	detalle := EmpresaTarifaPorMinutosCalculo{MinutosTolerancia: 10, BloquesExtra: 2}
	got := resolveCarritoTarifaPorMinutosCurrentEnd(activadoEn, tarifa, detalle)
	want := activadoEn.Add(260 * time.Minute)
	if !got.Equal(want) {
		t.Fatalf("fin de dos bloques con tolerancia = %s, want %s", got, want)
	}
}

func TestCalcularDetalleTarifaPorMinutosDistingueFraccionYBloqueCompleto(t *testing.T) {
	base := EmpresaTarifaPorMinutos{
		ID: 1, EmpresaID: 7, EstacionID: 102,
		MinutosBase: 60, ValorBase: 30000,
		MinutosExtra: 15, ValorExtra: 7500, Moneda: "COP",
	}
	cfg := defaultEmpresaTarifaPorMinutosConfiguracion(7)

	fraccion := base
	fraccion.CobrarPorFraccion = true
	got := CalcularDetalleTarifaPorMinutos(fraccion, 61, cfg)
	if got.BloquesExtra != 1 || got.MontoTotal != 37500 {
		t.Fatalf("61 minutos con cobro por fraccion = bloques %d total %.2f; want 1 y 37500", got.BloquesExtra, got.MontoTotal)
	}

	bloqueCompleto := base
	bloqueCompleto.CobrarPorFraccion = false
	got = CalcularDetalleTarifaPorMinutos(bloqueCompleto, 74, cfg)
	if got.BloquesExtra != 0 || got.MontoTotal != 30000 {
		t.Fatalf("74 minutos sin fraccion = bloques %d total %.2f; want 0 y 30000", got.BloquesExtra, got.MontoTotal)
	}
	got = CalcularDetalleTarifaPorMinutos(bloqueCompleto, 75, cfg)
	if got.BloquesExtra != 1 || got.MontoTotal != 37500 {
		t.Fatalf("75 minutos sin fraccion = bloques %d total %.2f; want 1 y 37500", got.BloquesExtra, got.MontoTotal)
	}
}

func TestCalcularDetalleTarifaMotelRespetaToleranciaDelPlan(t *testing.T) {
	tarifa := EmpresaTarifaMotel{
		ID:                1,
		EmpresaID:         7,
		EstacionID:        201,
		NombrePlan:        "Express 2 horas",
		TipoPlan:          "express",
		MinutosIncluidos:  120,
		ValorBase:         100000,
		MinutosExtra:      60,
		ValorExtra:        50000,
		CobrarPorFraccion: true,
		ToleranciaMinutos: 10,
		Moneda:            "COP",
		AplicarAutomatico: true,
	}

	dentro := CalcularDetalleTarifaMotel(tarifa, 129)
	if dentro.BloquesExtra != 0 {
		t.Fatalf("motel 2h + 9min con 10min de tolerancia no debe cobrar extra, bloques=%d", dentro.BloquesExtra)
	}
	if dentro.MontoTotal != 100000 {
		t.Fatalf("monto motel dentro de tolerancia = %.2f, want 100000", dentro.MontoTotal)
	}

	fuera := CalcularDetalleTarifaMotel(tarifa, 131)
	if fuera.BloquesExtra != 1 {
		t.Fatalf("motel 2h + 11min con 10min de tolerancia debe cobrar 1 bloque extra, bloques=%d", fuera.BloquesExtra)
	}
	if fuera.MontoTotal != 150000 {
		t.Fatalf("monto motel fuera de tolerancia = %.2f, want 150000", fuera.MontoTotal)
	}
}

func TestCalcularDetalleTarifaMotelRepiteToleranciaEnCadaHoraExtra(t *testing.T) {
	tarifa := EmpresaTarifaMotel{
		ID: 1, EmpresaID: 7, EstacionID: 201,
		MinutosIncluidos: 120, ValorBase: 20000,
		MinutosExtra: 60, ValorExtra: 10000,
		CobrarPorFraccion: true, ToleranciaMinutos: 10, Moneda: "COP",
	}
	for _, tc := range []struct {
		minutos int
		bloques int
	}{
		{130, 0}, {131, 1}, {200, 1}, {201, 2},
	} {
		got := CalcularDetalleTarifaMotel(tarifa, float64(tc.minutos))
		if got.BloquesExtra != tc.bloques {
			t.Fatalf("motel %d min: bloques=%d; want %d", tc.minutos, got.BloquesExtra, tc.bloques)
		}
	}
}

func TestCalcularDetalleTarifaPorMinutosNoDividePorBloqueExtraInvalido(t *testing.T) {
	tarifa := EmpresaTarifaPorMinutos{
		ID:           1,
		EmpresaID:    7,
		EstacionID:   101,
		MinutosBase:  120,
		ValorBase:    100000,
		MinutosExtra: 0,
		ValorExtra:   50000,
		Moneda:       "COP",
	}
	cfg := defaultEmpresaTarifaPorMinutosConfiguracion(7)

	got := CalcularDetalleTarifaPorMinutos(tarifa, 240, cfg)
	if got.BloquesExtra != 0 || got.MontoTotal != 100000 {
		t.Fatalf("minutos_extra invalido debe quedar solo base, bloques=%d total=%.2f", got.BloquesExtra, got.MontoTotal)
	}
}
