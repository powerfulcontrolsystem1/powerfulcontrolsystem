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
