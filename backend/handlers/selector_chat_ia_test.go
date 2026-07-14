package handlers

import (
	"strings"
	"testing"
)

func TestSelectorAIContextRedactsTechnicalCompanyIdentifiers(t *testing.T) {
	context := selectorAIContext([]string{"CONTEXTO EMPRESA\n- empresa_id: 77\n- nombre: Empresa autorizada\n- nit: 900123456\n- ventas_cerradas: 4\n"})
	for _, forbidden := range []string{"empresa_id: 77", "900123456"} {
		if strings.Contains(context, forbidden) {
			t.Fatalf("selector context leaked %q: %s", forbidden, context)
		}
	}
	if !strings.Contains(context, "Empresa autorizada") || !strings.Contains(context, "ventas_cerradas: 4") {
		t.Fatalf("selector context omitted allowed summary: %s", context)
	}
}
