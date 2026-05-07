package handlers

import "testing"

func TestNormalizeNominaControlPeriodo(t *testing.T) {
	desde, hasta, err := normalizeNominaControlPeriodo("2026-05-01", "2026-05-15")
	if err != nil {
		t.Fatalf("expected valid period, got %v", err)
	}
	if desde != "2026-05-01" || hasta != "2026-05-15" {
		t.Fatalf("unexpected normalized period %q/%q", desde, hasta)
	}
}

func TestNormalizeNominaControlPeriodoRejectsInvalidRange(t *testing.T) {
	if _, _, err := normalizeNominaControlPeriodo("2026-05-20", "2026-05-01"); err == nil {
		t.Fatal("expected invalid range error")
	}
}

func TestNominaControlPeriodoPILA(t *testing.T) {
	if got := nominaControlPeriodoPILA("2026-05-15"); got != "2026-05" {
		t.Fatalf("expected 2026-05, got %q", got)
	}
}

func TestRoundNominaControl(t *testing.T) {
	if got := roundNominaControl(123.456); got != 123.46 {
		t.Fatalf("expected 123.46, got %v", got)
	}
	if got := roundNominaControl(-123.456); got != -123.46 {
		t.Fatalf("expected -123.46, got %v", got)
	}
}
