package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestSuperErroresSistemaHandlerFiltersResults(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_errores_panel.db")
	ensureSuperSchema(t, dbSuper)
	if err := dbpkg.EnsureSuperErroresSistemaSchema(dbSuper); err != nil {
		t.Fatalf("ensure error schema: %v", err)
	}
	if err := dbpkg.UpsertAdministrador(dbSuper, "super@pcs.com", "Super", "super_administrador", ""); err != nil {
		t.Fatalf("seed admin: %v", err)
	}
	if err := dbpkg.CreateSession(dbSuper, "super@pcs.com", "127.0.0.1", "test-agent", "token-super-errores"); err != nil {
		t.Fatalf("seed session: %v", err)
	}
	if _, err := dbpkg.CreateSuperErrorSistema(dbSuper, dbpkg.SuperErrorSistema{
		Nivel:          "ERROR",
		TipoError:      "db_timeout",
		Mensaje:        "Timeout consultando inventario",
		MensajePublico: "No fue posible consultar inventario en este momento",
		EmpresaID:      14,
		UsuarioEmail:   "operador@empresa.com",
		Endpoint:       "/api/empresa/productos",
		Modulo:         "empresa/productos",
		MetodoHTTP:     http.MethodGet,
		CodigoHTTP:     http.StatusInternalServerError,
		RequestID:      "req-err-1",
		Origen:         "http",
		FechaError:     "2026-04-15 10:00:00",
	}); err != nil {
		t.Fatalf("seed error row 1: %v", err)
	}
	if _, err := dbpkg.CreateSuperErrorSistema(dbSuper, dbpkg.SuperErrorSistema{
		Nivel:          "WARNING",
		TipoError:      "validation_error",
		Mensaje:        "Parametro limit invalido",
		MensajePublico: "Debes corregir el filtro enviado",
		EmpresaID:      15,
		UsuarioEmail:   "cajero@empresa.com",
		Endpoint:       "/api/empresa/reportes",
		Modulo:         "empresa/reportes",
		MetodoHTTP:     http.MethodGet,
		CodigoHTTP:     http.StatusBadRequest,
		RequestID:      "req-err-2",
		Origen:         "http",
		FechaError:     "2026-04-15 12:00:00",
	}); err != nil {
		t.Fatalf("seed error row 2: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/super/api/errores?empresa_id=14&nivel=ERROR&search=timeout", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-super-errores"})
	rr := httptest.NewRecorder()

	SuperErroresSistemaHandler(dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status inesperado: got=%d want=%d body=%s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if ok, _ := payload["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got %#v", payload["ok"])
	}
	if total := int(payload["total"].(float64)); total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	items, _ := payload["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item, _ := items[0].(map[string]interface{})
	if got := item["tipo_error"].(string); got != "db_timeout" {
		t.Fatalf("expected tipo_error db_timeout, got %q", got)
	}
	summary, _ := payload["summary"].(map[string]interface{})
	if got := int(summary["error"].(float64)); got != 1 {
		t.Fatalf("expected summary.error=1, got %d", got)
	}
	if got := int(summary["warning"].(float64)); got != 0 {
		t.Fatalf("expected summary.warning=0 after filters, got %d", got)
	}
}

func TestSuperErroresSistemaHandlerMethodNotAllowed(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_errores_method.db")
	req := httptest.NewRequest(http.MethodPost, "/super/api/errores", nil)
	rr := httptest.NewRecorder()

	SuperErroresSistemaHandler(dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status inesperado: got=%d want=%d", rr.Code, http.StatusMethodNotAllowed)
	}
}