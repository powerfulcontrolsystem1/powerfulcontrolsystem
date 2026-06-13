package db

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildEmpresaAsientoVentaConImpuestosCuadra(t *testing.T) {
	payload, _ := json.Marshal(map[string]interface{}{
		"subtotal":         1000,
		"iva":              190,
		"retencion_fuente": 25,
		"metodo_pago":      "credito_cliente",
	})
	evento := EmpresaEventoContable{
		EmpresaID:     7,
		Modulo:        "facturacion",
		Evento:        "factura_emitida",
		MontoTotal:    1165,
		PayloadJSON:   string(payload),
		DocumentoTipo: "factura",
	}

	lineas := buildEmpresaAsientoContableLineas(evento, defaultEmpresaFinanzasConfiguracion(7))
	lineas, debito, credito, diferencia, err := normalizeAndValidateEmpresaAsientoLineas(evento, lineas)
	if err != nil {
		t.Fatalf("asiento de factura con impuestos debe cuadrar: %v", err)
	}
	if len(lineas) != 4 {
		t.Fatalf("se esperaban 4 lineas, got %d: %#v", len(lineas), lineas)
	}
	if debito != 1190 || credito != 1190 || diferencia != 0 {
		t.Fatalf("totales inesperados debito=%.2f credito=%.2f diferencia=%.2f", debito, credito, diferencia)
	}
}

func TestNormalizeAndValidateEmpresaAsientoLineasRechazaSinPartidaDoble(t *testing.T) {
	evento := EmpresaEventoContable{EmpresaID: 7, Modulo: "compras", Evento: "orden_compra_creada"}
	_, _, _, _, err := normalizeAndValidateEmpresaAsientoLineas(evento, nil)
	if err == nil || !strings.Contains(err.Error(), "partida doble") {
		t.Fatalf("debe rechazar eventos sin asiento real, got %v", err)
	}
}

func TestEmpresaEventoContableRequiereAsientoDistingueHitosPrecontables(t *testing.T) {
	precontable := EmpresaEventoContable{EmpresaID: 7, Modulo: "compras", Evento: "orden_compra_pendiente_aprobacion"}
	if empresaEventoContableRequiereAsiento(precontable) {
		t.Fatalf("un hito de aprobacion de compra debe quedar auditable sin asiento")
	}

	monetario := EmpresaEventoContable{EmpresaID: 7, Modulo: "compras", Evento: "compra_contabilizada"}
	if !empresaEventoContableRequiereAsiento(monetario) {
		t.Fatalf("una compra contabilizada debe exigir asiento")
	}
}
