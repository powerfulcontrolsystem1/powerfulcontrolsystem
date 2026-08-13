package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

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
		_, _, _, err := controller.callOpenAIResponsesWithSystemPromptContext(ctx, model, "extraer", nil, "sistema", nil, nil, nil)
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
