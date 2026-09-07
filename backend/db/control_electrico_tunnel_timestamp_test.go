package db

import (
	"os"
	"strings"
	"testing"
)

func TestControlElectricoTunnelQueueComparesRFC3339WithTimezone(t *testing.T) {
	raw, err := os.ReadFile("control_electrico_tunnel.go")
	if err != nil {
		t.Fatalf("read control_electrico_tunnel.go: %v", err)
	}
	body := string(raw)
	for _, column := range []string{"disponible_desde", "expira_en", "entregado_en"} {
		want := "CAST(NULLIF(" + column + ",'') AS TIMESTAMPTZ)"
		if !strings.Contains(body, want) {
			t.Fatalf("la cola debe comparar %s conservando Z/offset: falta %q", column, want)
		}
	}
	if strings.Contains(body, "AS TIMESTAMP)") {
		t.Fatal("la cola no debe eliminar la zona horaria de fechas RFC3339")
	}
}

func TestControlElectricoTunnelRestoreRetriesDiscardedDatabaseConnection(t *testing.T) {
	raw, err := os.ReadFile("control_electrico_tunnel.go")
	if err != nil {
		t.Fatalf("read control_electrico_tunnel.go: %v", err)
	}
	body := string(raw)
	for _, want := range []string{
		`"database/sql/driver"`,
		`for attempt := 0; attempt < 3; attempt++`,
		`errors.Is(err, driver.ErrBadConn)`,
		`queueEmpresaControlElectricoTunnelRestoreOnBootOnce`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("la recuperacion de arranque debe reintentar conexiones descartadas: falta %q", want)
		}
	}
}

func TestControlElectricoTunnelRestoreClosesReadCursorBeforeQueueWrites(t *testing.T) {
	raw, err := os.ReadFile("control_electrico_tunnel.go")
	if err != nil {
		t.Fatalf("read control_electrico_tunnel.go: %v", err)
	}
	body := string(raw)
	closeAt := strings.Index(body, `if err := rows.Close(); err != nil {`)
	reserveAfterClose := -1
	if closeAt >= 0 {
		reserveAfterClose = strings.Index(body[closeAt:], `reserveEmpresaControlElectricoActivationSlotTx`)
	}
	if closeAt < 0 || reserveAfterClose <= 0 {
		t.Fatal("la recuperacion debe cerrar el cursor de relés antes de reservar turnos de la cola")
	}
	if !strings.Contains(body, `activeRelays = append(activeRelays, relay)`) {
		t.Fatal("la recuperacion debe materializar los relés antes de escribir comandos")
	}
}
