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

func ensureConfigTableForChatIATest(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	_, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS configuraciones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_key TEXT UNIQUE,
		value TEXT,
		encrypted INTEGER DEFAULT 0,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT
	)`)
	if err != nil {
		t.Fatalf("create configuraciones table: %v", err)
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
	modelos, ok := payload["modelos"].([]interface{})
	if !ok || len(modelos) != 1 {
		t.Fatalf("expected 1 modelo, got %#v", payload["modelos"])
	}
	item, _ := modelos[0].(map[string]interface{})
	if item["id"] != "google:gemini-2.0-flash" {
		t.Fatalf("expected google:gemini-2.0-flash in modelos response, got %#v", item["id"])
	}
}

func TestModeloPreferidoHandlerAcceptsGemini(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)

	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 8, "Empresa Gemini", "900555", "admin@example.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, nil)
	body := `{"empresa_id":8,"model_id":"google:gemini-2.0-flash"}`
	req := httptest.NewRequest(http.MethodPut, "/api/empresa/chat_con_inteligencia_artificial/modelo_preferido", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.ModeloPreferidoHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	modelID, err := dbpkg.GetEmpresaAIModeloPreferido(dbEmp, 8, "admin@example.com")
	if err != nil {
		t.Fatalf("get modelo preferido: %v", err)
	}
	if modelID != "google:gemini-2.0-flash" {
		t.Fatalf("expected google:gemini-2.0-flash, got %q", modelID)
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

func TestModeloPreferidoHandlerGetRejectsEmpresaFueraDeAlcanceByGoogleAccount(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)

	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 31, "Empresa Scope Preferido GET", "900111", "owner@scope.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/chat_con_inteligencia_artificial/modelo_preferido?empresa_id=31", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.ModeloPreferidoHandler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for out-of-scope empresa, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "fuera del alcance") {
		t.Fatalf("expected out-of-scope message, got body=%s", rr.Body.String())
	}
}

func TestModeloPreferidoHandlerPutRejectsEmpresaFueraDeAlcanceByGoogleAccount(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)

	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 32, "Empresa Scope Preferido PUT", "900112", "owner@scope.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, nil)
	body := `{"empresa_id":32,"model_id":"google:gemini-2.0-flash"}`
	req := httptest.NewRequest(http.MethodPut, "/api/empresa/chat_con_inteligencia_artificial/modelo_preferido", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.ModeloPreferidoHandler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for out-of-scope empresa, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "fuera del alcance") {
		t.Fatalf("expected out-of-scope message, got body=%s", rr.Body.String())
	}
}

func TestHistorialHandlerRejectsEmpresaFueraDeAlcanceByGoogleAccount(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)

	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 33, "Empresa Scope Historial", "900113", "owner@scope.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/chat_con_inteligencia_artificial/historial?empresa_id=33&limit=10", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.HistorialHandler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 for out-of-scope empresa, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "fuera del alcance") {
		t.Fatalf("expected out-of-scope message, got body=%s", rr.Body.String())
	}
}

func TestModelosHandlerRejectsWhenAIDisabled(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	dbSuper := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)
	ensureConfigTableForChatIATest(t, dbSuper)
	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 41, "Empresa IA Off", "900555", "admin@example.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, superAIEnabledConfigKey, "0", false); err != nil {
		t.Fatalf("disable ai: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/chat_con_inteligencia_artificial/modelos?empresa_id=41", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.ModelosHandler(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestModelosHandlerFiltersDisabledProvider(t *testing.T) {
	dbEmp := openChatIAHandlerTestDB(t)
	dbSuper := openChatIAHandlerTestDB(t)
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmp); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	ensureEmpresasTableForChatIATest(t, dbEmp)
	ensureConfigTableForChatIATest(t, dbSuper)
	ensureSuperSchema(t, dbSuper)
	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, nit, usuario_creador) VALUES (?, ?, ?, ?)`, 42, "Empresa IA Providers", "900556", "admin@example.com")
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "ai.provider.google.enabled", "0", false); err != nil {
		t.Fatalf("disable google provider: %v", err)
	}

	ctrl := NewEmpresaAIChatController(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/chat_con_inteligencia_artificial/modelos?empresa_id=42", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@example.com"))
	rr := httptest.NewRecorder()

	ctrl.ModelosHandler(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d body=%s", rr.Code, rr.Body.String())
	}
}
