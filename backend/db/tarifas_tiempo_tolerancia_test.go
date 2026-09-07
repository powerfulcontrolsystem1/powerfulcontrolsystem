package db

import "testing"

func TestCalcularDetalleTarifaPorMinutosRespetaMargenTolerancia(t *testing.T) {
	tarifa := EmpresaTarifaPorMinutos{
		ID:           1,
		EmpresaID:    7,
		EstacionID:   101,
		MinutosBase:  120,
		ValorBase:    100000,
		MinutosExtra: 60,
		ValorExtra:   50000,
		Moneda:       "COP",
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
