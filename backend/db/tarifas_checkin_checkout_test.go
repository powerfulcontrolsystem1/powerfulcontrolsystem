package db

import (
	"testing"
	"time"
)

func TestTarifaAmanecidaCruzaMedianocheSinCobroDoble(t *testing.T) {
	loc := time.FixedZone("America/Bogota", -5*60*60)
	tarifa := EmpresaTarifaPorDia{
		ID: 81, EmpresaID: 12, EstacionID: 4,
		NombreTarifa: "Amanecida", ValorDia: 100000,
		HoraCheckIn: "14:00", HoraCheckOut: "13:00", Moneda: "COP",
	}
	entrada := time.Date(2026, 9, 11, 14, 0, 0, 0, loc)
	salida := time.Date(2026, 9, 12, 13, 0, 0, 0, loc)

	got := CalcularDetalleTarifaPorDia(tarifa, entrada, salida)
	if got.DiasCobrados != 1 || got.MontoTotal != 100000 {
		t.Fatalf("amanecida 14:00 a 13:00 = dias %d total %.2f; want 1 y 100000", got.DiasCobrados, got.MontoTotal)
	}
	if got.HoraCheckIn != "14:00" || got.HoraCheckOut != "13:00" {
		t.Fatalf("horario devuelto = %s/%s; want 14:00/13:00", got.HoraCheckIn, got.HoraCheckOut)
	}
}

func TestTarifaAmanecidaCheckoutSiguienteDia(t *testing.T) {
	loc := time.FixedZone("America/Bogota", -5*60*60)
	entrada := time.Date(2026, 9, 11, 15, 45, 0, 0, loc)
	salida := resolveTarifaPorDiaNextCheckoutBoundary(entrada, "14:00", "13:00")
	want := time.Date(2026, 9, 12, 13, 0, 0, 0, loc)
	if !salida.Equal(want) {
		t.Fatalf("checkout = %s; want %s", salida, want)
	}
}

func TestNormalizeCarritoTarifaTiempoPermiteSinTarifa(t *testing.T) {
	if got := normalizeCarritoTarifaTiempoTipo("sin tarifa"); got != "sin_tarifa" {
		t.Fatalf("normalize sin tarifa = %q; want sin_tarifa", got)
	}
}
