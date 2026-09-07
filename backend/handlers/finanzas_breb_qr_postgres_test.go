package handlers

import (
	"strings"
	"testing"
)

func TestBrebQRSortExpressionsKeepPostgresTextTypesCompatible(t *testing.T) {
	t.Parallel()
	for name, expression := range map[string]string{
		"carritos": brebQRCarritosOrderExpr,
		"abonos":   brebQRAbonosOrderExpr,
		"bancos":   brebQRBancosOrderExpr,
	} {
		if strings.Contains(expression, "CURRENT_TIMESTAMP") {
			t.Fatalf("%s mixes text timestamps with CURRENT_TIMESTAMP: %s", name, expression)
		}
		if !strings.Contains(expression, "pcs_ts(COALESCE(") || !strings.Contains(expression, "''") {
			t.Fatalf("%s must normalize empty text through pcs_ts: %s", name, expression)
		}
	}
}

func TestValidateAndNormalizeBrebQRCuentasUsesColombianKeyFormats(t *testing.T) {
	rows, err := validateAndNormalizeBrebQRCuentas([]map[string]interface{}{{
		"activa": true, "nombre": "Caja principal", "proveedor": "breb", "tipo_llave": "celular", "llave": "3001234567", "qr_tipo": "estatico", "payload_oficial": "000201010211",
	}})
	if err != nil || len(rows) != 1 {
		t.Fatalf("expected valid Colombian Bre-B account, rows=%#v err=%v", rows, err)
	}
	if _, err := validateAndNormalizeBrebQRCuentas([]map[string]interface{}{{"activa": true, "proveedor": "breb", "tipo_llave": "celular", "llave": "2001234567", "qr_tipo": "estatico"}}); err == nil {
		t.Fatal("expected invalid Colombian mobile key to be rejected")
	}
	if _, err := validateAndNormalizeBrebQRCuentas([]map[string]interface{}{{"activa": true, "proveedor": "breb", "tipo_llave": "celular", "llave": "3001234567", "qr_tipo": "dinamico", "payload_oficial": "000201{valor}"}}); err == nil {
		t.Fatal("expected dynamic/template QR to be rejected without a provider connector")
	}
}
