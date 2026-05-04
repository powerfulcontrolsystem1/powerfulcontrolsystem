package db

import (
	"testing"
	"time"
)

func TestProductoVencimientoStatus(t *testing.T) {
	now := time.Date(2026, 5, 4, 10, 30, 0, 0, time.Local)
	tests := []struct {
		name       string
		fecha      string
		diasAlerta int
		wantEstado string
		wantDias   int
	}{
		{name: "sin fecha", fecha: "", diasAlerta: 10, wantEstado: "sin_fecha", wantDias: 0},
		{name: "vencido", fecha: "2026-05-03", diasAlerta: 10, wantEstado: "vencido", wantDias: -1},
		{name: "vence hoy", fecha: "2026-05-04", diasAlerta: 10, wantEstado: "vence_hoy", wantDias: 0},
		{name: "proximo a vencer", fecha: "2026-05-12", diasAlerta: 10, wantEstado: "proximo_vencer", wantDias: 8},
		{name: "vigente", fecha: "2026-06-20", diasAlerta: 10, wantEstado: "vigente", wantDias: 47},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEstado, gotDias := productoVencimientoStatus(tt.fecha, tt.diasAlerta, now)
			if gotEstado != tt.wantEstado || gotDias != tt.wantDias {
				t.Fatalf("productoVencimientoStatus() = (%q, %d), want (%q, %d)", gotEstado, gotDias, tt.wantEstado, tt.wantDias)
			}
		})
	}
}

func TestNormalizeProductoVencimiento(t *testing.T) {
	p := Producto{
		FechaVencimiento:      "2026-05-04 13:15:00",
		DiasAlertaVencimiento: -2,
		LoteCodigo:            " L-1 ",
	}
	if err := normalizeProductoVencimiento(&p); err != nil {
		t.Fatalf("normalizeProductoVencimiento() error = %v", err)
	}
	if !p.ManejaVencimiento {
		t.Fatalf("ManejaVencimiento = false, want true when fecha_vencimiento is present")
	}
	if p.FechaVencimiento != "2026-05-04" {
		t.Fatalf("FechaVencimiento = %q, want 2026-05-04", p.FechaVencimiento)
	}
	if p.DiasAlertaVencimiento != 30 {
		t.Fatalf("DiasAlertaVencimiento = %d, want 30", p.DiasAlertaVencimiento)
	}
	if p.LoteCodigo != "L-1" {
		t.Fatalf("LoteCodigo = %q, want L-1", p.LoteCodigo)
	}
}

func TestNormalizeProductoVencimientoRejectsInvalidDate(t *testing.T) {
	p := Producto{ManejaVencimiento: true, FechaVencimiento: "04/05/2026"}
	if err := normalizeProductoVencimiento(&p); err == nil {
		t.Fatalf("normalizeProductoVencimiento() error = nil, want invalid date error")
	}
}
