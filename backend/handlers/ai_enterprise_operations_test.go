package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	aipkg "github.com/you/pos-backend/ai"
)

func TestEnterpriseOperationsRespectRoleFlagsAndExplanation(t *testing.T) {
	for _, flag := range []string{"AI_ENTERPRISE_ORCHESTRATOR_ENABLED", "AI_AGENT_MODE_ENABLED", "AI_WRITE_TOOLS_ENABLED", "AI_CATALOG_TOOLS_ENABLED", "AI_SALES_TOOLS_ENABLED"} {
		t.Setenv(flag, "true")
	}
	cases := []struct {
		name, question string
		permissions    []string
		wantWrite      bool
	}{
		{"administrator action", "crea un producto", []string{"inventario:C", "ventas:R", "ventas:C"}, true},
		{"read only", "crea un producto", []string{"inventario:R", "ventas:R"}, false},
		{"explanation", "Explícame cómo crear un producto paso a paso", []string{"inventario:C", "ventas:R", "ventas:C"}, false},
		{"no permissions", "agrega un producto", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tools := enterpriseAIChatTools(aipkg.ExecutionContext{Permissions: tc.permissions}, tc.question)
			writes := false
			for _, tool := range tools {
				def := aipkg.Registry()[tool["name"].(string)]
				writes = writes || def.Confirmation == "required"
			}
			if writes != tc.wantWrite {
				t.Fatalf("writes=%v", writes)
			}
		})
	}
	t.Setenv("AI_SALES_TOOLS_ENABLED", "")
	for _, tool := range enterpriseAIChatTools(aipkg.ExecutionContext{Permissions: []string{"ventas:R", "ventas:C"}}, "agrega dos cervezas") {
		if tool["name"] == aipkg.ToolSalesAddStationProduct {
			t.Fatal("disabled sales write exposed")
		}
	}
}

func TestEnterpriseOperationsRejectAuthorityAndInvalidQuantity(t *testing.T) {
	for _, raw := range []string{`{"estacion_id":1,"producto_id":2,"cantidad":2,"empresa_id":3}`, `{"estacion_id":1,"producto_id":2,"cantidad":2,"confirmed":true}`, `{} {}`, `{"cantidad":2.2}`} {
		var args enterpriseStationProductArgs
		if decodeEnterpriseTool(raw, &args) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	for _, n := range []int64{-1, 0, 100} {
		if validateEnterpriseStationProductArgs(enterpriseStationProductArgs{EstacionID: 1, ProductoID: 1, Cantidad: n}) == nil {
			t.Fatalf("accepted quantity %d", n)
		}
	}
	if key, _ := enterpriseAIReportDataset("SELECT * FROM empresas"); key != "" {
		t.Fatal("arbitrary dataset accepted")
	}
	_, err := dispatchEnterpriseAIOperation(nil, httptest.NewRequest("POST", "/", nil), aipkg.ExecutionContext{}, openAIResponsesFunctionCall{Name: aipkg.ToolSalesAddStationProduct, Arguments: `{}`})
	if err == nil {
		t.Fatal("unauthorized call reached database")
	}
}

func TestEnterpriseResponsesMultipleToolsAccumulateUsage(t *testing.T) {
	requests, dispatches := 0, 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if requests < 3 {
			fmt.Fprintf(w, `{"output":[{"type":"function_call","name":"catalog.search_products","call_id":"c%d","arguments":"{}"}],"usage":{"input_tokens":10,"output_tokens":2}}`, requests)
		} else {
			if len(body["input"].([]interface{})) != 6 {
				t.Errorf("previous tool results lost: %d", len(body["input"].([]interface{})))
			}
			fmt.Fprint(w, `{"output":[{"content":[{"type":"output_text","text":"Listo"}]}],"usage":{"input_tokens":10,"output_tokens":2}}`)
		}
	}))
	defer srv.Close()
	c := EmpresaAIChatController{client: srv.Client()}
	m := empresaAIModelDef{Provider: "openai", UpstreamModel: "test", Endpoint: srv.URL + "/v1/responses", apiKeyOverride: "test"}
	_, input, output, err := c.callOpenAIResponsesWithSystemPrompt(m, "consulta", nil, "sistema", nil, []map[string]interface{}{{"name": aipkg.ToolCatalogSearchProducts}}, func(call openAIResponsesFunctionCall) (string, error) { dispatches++; return `{}`, nil }, empresaAISafetyIdentifier("test"))
	if err != nil || input != 30 || output != 6 || dispatches != 2 {
		t.Fatalf("err=%v usage=%d/%d dispatches=%d", err, input, output, dispatches)
	}
}

func TestEnterpriseResponsesRejectsUnadvertisedTool(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"output":[{"type":"function_call","name":"catalog.create_product","call_id":"c1","arguments":"{}"}]}`)
	}))
	defer srv.Close()
	c := EmpresaAIChatController{client: srv.Client()}
	m := empresaAIModelDef{Provider: "openai", UpstreamModel: "test", Endpoint: srv.URL + "/v1/responses", apiKeyOverride: "test"}
	called := false
	_, _, _, err := c.callOpenAIResponsesWithSystemPrompt(m, "consulta", nil, "sistema", nil, []map[string]interface{}{{"name": aipkg.ToolCatalogSearchProducts}}, func(call openAIResponsesFunctionCall) (string, error) { called = true; return `{}`, nil }, empresaAISafetyIdentifier("test"))
	if err == nil || called {
		t.Fatal("unadvertised write was dispatched")
	}
}
