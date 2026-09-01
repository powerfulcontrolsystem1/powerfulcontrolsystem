package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEmpresaAISafetyIdentifierIsStableAndDoesNotExposeAccount(t *testing.T) {
	first := empresaAISafetyIdentifier("User@Example.com")
	if first == "" || first != empresaAISafetyIdentifier(" user@example.com ") {
		t.Fatalf("safety identifier must be stable: %q", first)
	}
	if len(first) != len("pcs-")+32 || first == "pcs-user@example.com" {
		t.Fatalf("safety identifier must be pseudonymous: %q", first)
	}
}

func TestPublicAIProviderErrorDoesNotExposeProviderBody(t *testing.T) {
	private := "provider detail that must not reach the browser"
	got := publicAIProviderError(&aiProviderHTTPError{Provider: "openai", Status: http.StatusBadGateway, Body: private})
	if strings.Contains(got, private) {
		t.Fatalf("provider body leaked in public error: %q", got)
	}
}

func TestOpenAIResponsesSendsPseudonymousSafetyIdentifier(t *testing.T) {
	var body map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"output":[{"content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer server.Close()
	controller := &EmpresaAIChatController{client: server.Client()}
	model := empresaAIModelDef{Provider: "openai", UpstreamModel: "gpt-test", Endpoint: server.URL + "/v1/responses", apiKeyOverride: "test-only"}
	safetyIdentifier := empresaAISafetyIdentifier("user@example.com")
	if _, _, _, err := controller.callOpenAIResponsesWithSystemPrompt(model, "hola", nil, "sistema", nil, nil, nil, safetyIdentifier); err != nil {
		t.Fatal(err)
	}
	if got, _ := body["safety_identifier"].(string); got != safetyIdentifier {
		t.Fatalf("safety_identifier = %q, want %q", got, safetyIdentifier)
	}
	if _, ok := body["store"]; !ok || body["store"] != false {
		t.Fatalf("Responses request must keep provider persistence disabled: %#v", body)
	}
}

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

func TestOpenAIResponsesContextCancellationStopsProviderRequest(t *testing.T) {
	started := make(chan struct{})
	releaseServer := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
		case <-releaseServer:
		}
	}))
	defer func() {
		close(releaseServer)
		server.Close()
	}()
	controller := &EmpresaAIChatController{client: &http.Client{Timeout: 5 * time.Second}}
	model := empresaAIModelDef{
		Provider:       "openai",
		UpstreamModel:  "gpt-test",
		Endpoint:       server.URL + "/v1/responses",
		apiKeyOverride: "test-only",
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, _, _, err := controller.callOpenAIResponsesWithSystemPromptContext(ctx, model, "extraer", nil, "sistema", nil, nil, nil, empresaAISafetyIdentifier("test-user"))
		done <- err
	}()
	select {
	case <-started:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not start")
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation not propagated: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not stop after cancellation")
	}
}

func TestOpenAIResponsesContextDeadlineStopsProviderRequest(t *testing.T) {
	started := make(chan struct{})
	releaseServer := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
		case <-releaseServer:
		}
	}))
	defer func() {
		close(releaseServer)
		server.Close()
	}()
	controller := &EmpresaAIChatController{client: &http.Client{Timeout: 5 * time.Second}}
	model := empresaAIModelDef{
		Provider: "openai", UpstreamModel: "gpt-test",
		Endpoint: server.URL + "/v1/responses", apiKeyOverride: "test-only",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, _, _, err := controller.callOpenAIResponsesWithSystemPromptContext(ctx, model, "extraer", nil, "sistema", nil, nil, nil)
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not start")
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("deadline not propagated: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not stop after deadline")
	}
}

func TestOpenAIChatCompletionsContextCancellationStopsProviderRequest(t *testing.T) {
	started := make(chan struct{})
	releaseServer := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
		case <-releaseServer:
		}
	}))
	defer func() {
		close(releaseServer)
		server.Close()
	}()
	controller := &EmpresaAIChatController{client: &http.Client{Timeout: 5 * time.Second}}
	model := empresaAIModelDef{
		Provider: "openai", UpstreamModel: "gpt-test",
		Endpoint: server.URL + "/v1/chat/completions", apiKeyOverride: "test-only",
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, _, _, err := controller.generateResponseWithSystemPromptContext(ctx, model, "consulta", nil, "sistema")
		done <- err
	}()
	select {
	case <-started:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not start")
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation not propagated: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not stop after cancellation")
	}
}

func TestGeminiContextCancellationStopsProviderRequest(t *testing.T) {
	started := make(chan struct{})
	releaseServer := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-r.Context().Done():
		case <-releaseServer:
		}
	}))
	defer func() {
		close(releaseServer)
		server.Close()
	}()
	controller := &EmpresaAIChatController{client: &http.Client{Timeout: 5 * time.Second}}
	model := empresaAIModelDef{
		Provider: "google", UpstreamModel: "gemini-test",
		Endpoint: server.URL + "/generateContent", apiKeyOverride: "test-only",
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, _, _, err := controller.generateResponseWithSystemPromptContext(ctx, model, "consulta", nil, "sistema")
		done <- err
	}()
	select {
	case <-started:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not start")
	}
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation not propagated: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("provider request did not stop after cancellation")
	}
}
