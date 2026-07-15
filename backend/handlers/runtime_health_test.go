package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRuntimeHealthHandlerIsPublicProbeSafe(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	RuntimeHealthHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("health response = %d %q", rec.Code, rec.Body.String())
	}
}

func TestRuntimeReadyWithCheckDoesNotExposeReleaseFailure(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	RuntimeReadyWithCheck(func(context.Context) error {
		return errors.New("internal schema migration detail")
	}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "schema") {
		t.Fatalf("ready response exposed release state: %d %q", rec.Code, rec.Body.String())
	}
}

func TestRuntimeReadyHandlerRejectsMissingDependency(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	RuntimeReadyHandler(nil).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if rec.Code != http.StatusServiceUnavailable || strings.Contains(strings.ToLower(rec.Body.String()), "database") {
		t.Fatalf("ready response exposed dependency or status = %d %q", rec.Code, rec.Body.String())
	}
}
