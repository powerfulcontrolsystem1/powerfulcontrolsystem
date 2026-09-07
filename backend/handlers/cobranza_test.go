package handlers

import (
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestRenderCobranzaMessage(t *testing.T) {
	got := renderCobranzaMessage("{{empresa}} recuerda a {{cliente}} el saldo {{saldo}} de {{documento}}, vence {{vencimiento}}.", dbpkg.EmpresaCobranzaCuenta{ClienteNombre: "Cliente PCS", DocumentoCodigo: "FV-10", Saldo: 25000, Moneda: "COP", FechaVencimiento: "2026-07-09"}, "Powerful Control System")
	for _, expected := range []string{"Powerful Control System", "Cliente PCS", "25000 COP", "FV-10", "2026-07-09"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("mensaje no contiene %q: %s", expected, got)
		}
	}
}

func TestCobranzaWorkerDueHonorsConfiguredHourAndDailyRun(t *testing.T) {
	now := time.Date(2026, 7, 9, 9, 15, 0, 0, time.Local)
	cfg := dbpkg.EmpresaCobranzaConfiguracion{HoraLocal: "09:00"}
	if !cobranzaWorkerDue(cfg, now) {
		t.Fatal("se esperaba ejecucion durante la hora configurada")
	}
	cfg.UltimaEjecucion = "2026-07-09 09:02:00"
	if cobranzaWorkerDue(cfg, now) {
		t.Fatal("no debe ejecutar dos veces el mismo dia")
	}
	cfg.UltimaEjecucion = "2026-07-08 09:02:00"
	if cobranzaWorkerDue(cfg, now.Add(time.Hour)) {
		t.Fatal("no debe ejecutar fuera de la hora configurada")
	}
}
