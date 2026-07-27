package handlers

import (
	"testing"
	"time"
)

func TestFacturacionDIANFechaEmisionFirmadaUsesColombiaSigningDate(t *testing.T) {
	instant := time.Date(2026, time.July, 27, 1, 2, 3, 0, time.UTC)
	got := facturacionDIANFechaEmisionFirmada(instant)
	if got != "2026-07-26T20:02:03-05:00" {
		t.Fatalf("fecha fiscal firmada=%q, se esperaba la fecha Colombia de firma", got)
	}
}
