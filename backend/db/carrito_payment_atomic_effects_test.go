package db

import (
	"os"
	"strings"
	"testing"
)

func TestCarritoPaymentPersistsCashAndOutboxInSameTransaction(t *testing.T) {
	raw, err := os.ReadFile("carritos_compras.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func PayCarritoStationSession")
	if start < 0 {
		t.Fatal("no se encontro PayCarritoStationSession")
	}
	end := strings.Index(source[start:], "type carritoSaleItemSnapshot")
	if end < 0 {
		t.Fatal("no se encontro el final de PayCarritoStationSession")
	}
	body := source[start : start+end]
	for _, required := range []string{
		"UPDATE empresa_cierres_caja",
		"ingresos_efectivo = COALESCE(ingresos_efectivo, 0) + ?",
		"recordCarritoStationMetricTx(tx",
		"InsertOutboxEvent(tx",
		"document_intent",
		"if err := tx.Commit(); err != nil",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("el pago de carrito no conserva efecto atomico %q", required)
		}
	}
}
