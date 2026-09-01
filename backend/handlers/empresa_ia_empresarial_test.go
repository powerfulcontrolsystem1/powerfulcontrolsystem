package handlers

import "testing"

func TestNormalizeEmpresaIAAccionPreservesCatalogIDs(t *testing.T) {
	tests := map[string]string{
		"diagnostico_erp":       "diagnostico_erp",
		"borrador_factura":      "borrador_factura",
		"cobranza_pagos":        "cobranza_pagos",
		"inventario_productos":  "inventario_productos",
		"conciliacion_bancaria": "conciliacion_bancaria",
		"compras_gastos":        "compras_gastos",
		"cumplimiento_dian":     "cumplimiento_dian",
	}
	for input, want := range tests {
		if got := normalizeEmpresaIAAccion(input); got != want {
			t.Errorf("normalizeEmpresaIAAccion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestNormalizeEmpresaIAAccionFailsClosedToDiagnosis(t *testing.T) {
	if got := normalizeEmpresaIAAccion("accion_no_permitida"); got != "diagnostico_erp" {
		t.Fatalf("accion desconocida = %q, want diagnostico_erp", got)
	}
}
