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
