package handlers

import "testing"

func TestParseOpenAIResponsesFunctionCallsRejectsMalformedArguments(t *testing.T) {
	valid := []byte(`{"output":[{"type":"function_call","call_id":"call_1","name":"catalog.create_product","arguments":"{}"}]}`)
	calls, err := parseOpenAIResponsesFunctionCalls(valid)
	if err != nil || len(calls) != 1 {
		t.Fatalf("valid call rejected: %v", err)
	}
	invalid := []byte(`{"output":[{"type":"function_call","call_id":"call_1","name":"catalog.create_product","arguments":"not-json"}]}`)
	if _, err := parseOpenAIResponsesFunctionCalls(invalid); err == nil {
		t.Fatal("malformed arguments accepted")
	}
}
