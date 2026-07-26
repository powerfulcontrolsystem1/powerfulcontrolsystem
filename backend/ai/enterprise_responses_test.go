package ai

import "testing"

func TestResponsesToolDefinitionsAreStrictAndTenantFree(t *testing.T) {
	tools := ResponsesToolDefinitions(ExecutionContext{Permissions: []string{"inventario:R", "inventario:C"}})
	if len(tools) != 2 {
		t.Fatalf("got %d tools", len(tools))
	}
	for _, tool := range tools {
		if tool["strict"] != true {
			t.Fatal("tool must be strict")
		}
		params := tool["parameters"].(map[string]interface{})
		if params["additionalProperties"] != false {
			t.Fatal("tool must reject extra properties")
		}
		if _, found := params["empresa_id"]; found {
			t.Fatal("tool must not accept tenant")
		}
		properties := params["properties"].(map[string]interface{})
		required := params["required"].([]string)
		if len(required) != len(properties) {
			t.Fatal("strict tool must require every property")
		}
	}
}
