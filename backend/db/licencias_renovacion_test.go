package db

import (
	"database/sql"
	"testing"
	"time"
)

func TestBuildLicenciaStackWindowSumaPagoAnticipado(t *testing.T) {
	now := time.Date(2026, 6, 1, 10, 0, 0, 0, time.Local)
	venceActual := time.Date(2026, 6, 11, 10, 0, 0, 0, time.Local)

	inicio, fin := buildLicenciaStackWindow(now, venceActual, now.Format("2006-01-02 15:04:05"), now.AddDate(0, 0, 30).Format("2006-01-02 15:04:05"), 30)
	if inicio != "2026-06-11 10:00:00" {
		t.Fatalf("inicio renovacion anticipada = %s, want 2026-06-11 10:00:00", inicio)
	}
	if fin != "2026-07-11 10:00:00" {
		t.Fatalf("fin renovacion anticipada = %s, want 2026-07-11 10:00:00", fin)
	}

	segundaRenovacionAnchor := time.Date(2026, 7, 11, 10, 0, 0, 0, time.Local)
	inicio, fin = buildLicenciaStackWindow(now, segundaRenovacionAnchor, "", "", 30)
	if inicio != "2026-07-11 10:00:00" {
		t.Fatalf("inicio segunda renovacion = %s, want 2026-07-11 10:00:00", inicio)
	}
	if fin != "2026-08-10 10:00:00" {
		t.Fatalf("fin segunda renovacion = %s, want 2026-08-10 10:00:00", fin)
	}
}

func TestResolveLicenciaDurationDaysFallback(t *testing.T) {
	got := resolveLicenciaDurationDays(sql.NullInt64{}, "2026-06-01 10:00:00", "2026-06-16 10:00:00")
	if got != 15 {
		t.Fatalf("duracion derivada = %d, want 15", got)
	}
	got = resolveLicenciaDurationDays(sql.NullInt64{Int64: 60, Valid: true}, "2026-06-01 10:00:00", "2026-06-16 10:00:00")
	if got != 60 {
		t.Fatalf("duracion explicita = %d, want 60", got)
	}
}
