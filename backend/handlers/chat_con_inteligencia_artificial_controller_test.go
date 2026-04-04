package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
	_ "modernc.org/sqlite"
)

func openChatIAHandlerTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "chat_ia_handler_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	return dbConn
}

func ensureEmpresasTableForChatIATest(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	_, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT,
		nit TEXT,
		usuario_creador TEXT
	)`)
	if err != nil {
		t.Fatalf("create empresas table: %v", err)
	}
}

func TestModelosHandlerRequiresGoogleAccount(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)

	ctrl := NewEmpresaAIChatController(dbEmp, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/chat_con_inteligencia_artificial/modelos?empresa_id=1", nil)
	rr := httptest.NewRecorder()

	ctrl.ModelosHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 when google account missing, got %d", rr.Code)
	}
}

func TestModelosHandlerReturnsPreferredModelForGoogleAccount(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)

	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 7, "Empresa Test", "900123", "admin@example.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	if err := dbpkg.UpsertEmpresaAIModeloPreferido(dbEmp, 7, "admin@example.com", "google:gemini-2.0-flash", "admin@example.com"); err != nil {
		t.Fatalf("upsert modelo preferido: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/chat_con_inteligencia_artificial/modelos?empresa_id=7", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.ModelosHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode json response: %v", err)
	}
	if payload["google_account"] != "admin@example.com" {
		t.Fatalf("expected google_account admin@example.com, got %#v", payload["google_account"])
	}
	if payload["modelo_preferido"] != "google:gemini-2.0-flash" {
		t.Fatalf("expected modelo_preferido google:gemini-2.0-flash, got %#v", payload["modelo_preferido"])
	}
}

func TestConsultarHandlerRejectsEmpresaFueraDeAlcance(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)

	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 11, "Empresa Scope", "900999", "owner@scope.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, nil)
	body := `{"empresa_id":11,"model_id":"google:gemini-2.0-flash","pregunta":"Hola"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/chat_con_inteligencia_artificial/consultar", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.ConsultarHandler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for out-of-scope empresa, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "fuera del alcance") {
		t.Fatalf("expected out-of-scope message, got body=%s", rr.Body.String())
	}
}

func TestModelosHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)

	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 21, "Empresa Scope Modelos", "900321", "owner@scope.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/chat_con_inteligencia_artificial/modelos?empresa_id=21", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.ModelosHandler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for out-of-scope empresa, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "fuera del alcance") {
		t.Fatalf("expected out-of-scope message, got body=%s", rr.Body.String())
	}
}
