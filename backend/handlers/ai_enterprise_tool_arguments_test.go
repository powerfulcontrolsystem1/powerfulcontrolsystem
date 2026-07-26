package handlers

import (
	"testing"

	aipkg "github.com/you/pos-backend/ai"
)

func TestDecodeEmpresaAIProductToolArgumentsIsStrict(t *testing.T) {
	valid := `{"nombre":"Cafe","sku":null,"descripcion":null,"categoria_id":null,"bodega_id":null,"unidad_medida":null,"costo":null,"precio":12000,"impuesto_porcentaje":null,"stock_inicial":null,"stock_minimo":null}`
	plan, err := decodeEmpresaAIProductToolArguments(valid)
	if err != nil || plan.Nombre != "Cafe" {
		t.Fatalf("valid tool plan rejected: %v", err)
	}
	for _, raw := range []string{`{"nombre":"Cafe","precio":1}`, `{"nombre":"Cafe","sku":null,"descripcion":null,"categoria_id":null,"bodega_id":null,"unidad_medida":null,"costo":null,"precio":1,"impuesto_porcentaje":null,"stock_inicial":null,"stock_minimo":null,"empresa_id":9}`} {
		if _, err := decodeEmpresaAIProductToolArguments(raw); err == nil {
			t.Fatalf("unsafe tool arguments accepted: %s", raw)
		}
	}
}

func TestDispatchEnterpriseAIResponsesRejectsUnknownToolBeforeDatabase(t *testing.T) {
	_, err := dispatchEnterpriseAIResponsesFunctionCall(nil, nil, aipkg.ExecutionContext{}, openAIResponsesFunctionCall{Name: "unknown", Arguments: `{}`})
	if err == nil {
		t.Fatal("unknown tool accepted")
	}
}
