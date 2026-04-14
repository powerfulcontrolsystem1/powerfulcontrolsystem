package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostgresPerformanceHandlerMethodNotAllowed(t *testing.T) {
	dbEmpresas := openTestSQLite(t, "empresas_method_not_allowed.db")
	dbSuper := openTestSQLite(t, "super_method_not_allowed.db")

	req := httptest.NewRequest(http.MethodPost, "/super/api/postgres/performance", nil)
	rr := httptest.NewRecorder()

	PostgresPerformanceHandler(dbEmpresas, dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status inesperado: got=%d want=%d", rr.Code, http.StatusMethodNotAllowed)
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "method not allowed") {
		t.Fatalf("mensaje inesperado: %s", rr.Body.String())
	}
}

func TestPostgresPerformanceHandlerDialectGuard(t *testing.T) {
	t.Setenv("DB_DIALECT", "sqlite")
	dbEmpresas := openTestSQLite(t, "empresas_dialect_guard.db")
	dbSuper := openTestSQLite(t, "super_dialect_guard.db")

	req := httptest.NewRequest(http.MethodGet, "/super/api/postgres/performance", nil)
	rr := httptest.NewRecorder()

	PostgresPerformanceHandler(dbEmpresas, dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status inesperado: got=%d want=%d", rr.Code, http.StatusConflict)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("respuesta JSON invalida: %v", err)
	}

	okVal, okTyped := payload["ok"].(bool)
	if !okTyped || okVal {
		t.Fatalf("payload esperado ok=false, recibido: %#v", payload["ok"])
	}

	errMsg, _ := payload["error"].(string)
	if !strings.Contains(strings.ToLower(errMsg), "postgres") {
		t.Fatalf("mensaje de error inesperado: %s", errMsg)
	}
}
