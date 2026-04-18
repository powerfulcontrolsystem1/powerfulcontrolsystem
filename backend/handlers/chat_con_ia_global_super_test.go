package handlers

import (
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

func openSuperAIHandlerTestDB(t *testing.T, name string) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), name)
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

func ensureEmpresasCoreForSuperAI(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	_, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT,
		nit TEXT,
		estado TEXT,
		usuario_creador TEXT
	)`)
	if err != nil {
		t.Fatalf("create empresas table: %v", err)
	}
	_, err = dbConn.Exec(`CREATE TABLE IF NOT EXISTS clientes (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER)`)
	if err != nil {
		t.Fatalf("create clientes table: %v", err)
	}
	_, err = dbConn.Exec(`CREATE TABLE IF NOT EXISTS productos (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER)`)
	if err != nil {
		t.Fatalf("create productos table: %v", err)
	}
	_, err = dbConn.Exec(`CREATE TABLE IF NOT EXISTS carritos_compras (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, estado_carrito TEXT, total REAL)`)
	if err != nil {
		t.Fatalf("create carritos_compras table: %v", err)
	}
	_, err = dbConn.Exec(`CREATE TABLE IF NOT EXISTS empresa_finanzas_movimientos (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, tipo TEXT, valor REAL)`)
	if err != nil {
		t.Fatalf("create empresa_finanzas_movimientos table: %v", err)
	}
}

func ensureConfigTableForSuperAITest(t *testing.T, dbConn *sql.DB) {
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

func createSuperSession(t *testing.T, dbSuper *sql.DB, email, role, token string) {
	t.Helper()
	ensureSuperSchema(t, dbSuper)
	if err := dbpkg.EnsureAdministradoresAuthSchema(dbSuper); err != nil {
		t.Fatalf("ensure administradores auth schema: %v", err)
	}
	if err := dbpkg.UpsertAdministrador(dbSuper, email, "Super Test", role, ""); err != nil {
		t.Fatalf("upsert admin: %v", err)
	}
	if err := dbpkg.CreateSession(dbSuper, email, "127.0.0.1", "unit-test", token); err != nil {
		t.Fatalf("create session: %v", err)
	}
}

func TestSuperAIModelosHandlerRequiresSuperSession(t *testing.T) {
	dbEmp := openSuperAIHandlerTestDB(t, "empresas.db")
	dbSuper := openSuperAIHandlerTestDB(t, "super.db")
	if err := dbpkg.EnsureSuperAIChatSchema(dbSuper); err != nil {
		t.Fatalf("ensure super ai schema: %v", err)
	}
	ensureEmpresasCoreForSuperAI(t, dbEmp)

	ctrl := NewSuperAIChatController(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/super/api/chat_con_ia_global/modelos", nil)
	rr := httptest.NewRecorder()

	ctrl.ModelosHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSuperAIModelosHandlerReturnsCatalog(t *testing.T) {
	dbEmp := openSuperAIHandlerTestDB(t, "empresas.db")
	dbSuper := openSuperAIHandlerTestDB(t, "super.db")
	if err := dbpkg.EnsureSuperAIChatSchema(dbSuper); err != nil {
		t.Fatalf("ensure super ai schema: %v", err)
	}
	ensureEmpresasCoreForSuperAI(t, dbEmp)
	createSuperSession(t, dbSuper, "super@pcs.com", "super_administrador", "token-super-ai")
	if err := dbpkg.UpsertSuperAIModeloPreferido(dbSuper, "super@pcs.com", "ollama:ambis", "super@pcs.com"); err != nil {
		t.Fatalf("upsert super model preferred: %v", err)
	}

	ctrl := NewSuperAIChatController(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/super/api/chat_con_ia_global/modelos", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-super-ai"})
	rr := httptest.NewRecorder()

	ctrl.ModelosHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rr.Code, rr.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode json response: %v", err)
	}
	if payload["admin_email"] != "super@pcs.com" {
		t.Fatalf("expected admin_email super@pcs.com, got %#v", payload["admin_email"])
	}
	if payload["modelo_preferido"] != "ollama:ambis" {
		t.Fatalf("expected modelo_preferido ollama:ambis, got %#v", payload["modelo_preferido"])
	}
	modelos, ok := payload["modelos"].([]interface{})
	if !ok || len(modelos) < 2 {
		t.Fatalf("expected modelos catalog, got %#v", payload["modelos"])
	}
}

func TestSuperAIModeloPreferidoHandlerRejectsNonSuper(t *testing.T) {
	dbEmp := openSuperAIHandlerTestDB(t, "empresas.db")
	dbSuper := openSuperAIHandlerTestDB(t, "super.db")
	if err := dbpkg.EnsureSuperAIChatSchema(dbSuper); err != nil {
		t.Fatalf("ensure super ai schema: %v", err)
	}
	ensureEmpresasCoreForSuperAI(t, dbEmp)
	createSuperSession(t, dbSuper, "admin@pcs.com", "administrador", "token-admin-ai")

	ctrl := NewSuperAIChatController(dbEmp, dbSuper)
	body := `{"model_id":"ollama:ambis"}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/chat_con_ia_global/modelo_preferido", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-admin-ai"})
	rr := httptest.NewRecorder()

	ctrl.ModeloPreferidoHandler(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSuperAIModelosHandlerRejectsWhenAIDisabled(t *testing.T) {
	dbEmp := openSuperAIHandlerTestDB(t, "empresas.db")
	dbSuper := openSuperAIHandlerTestDB(t, "super.db")
	if err := dbpkg.EnsureSuperAIChatSchema(dbSuper); err != nil {
		t.Fatalf("ensure super ai schema: %v", err)
	}
	ensureEmpresasCoreForSuperAI(t, dbEmp)
	ensureConfigTableForSuperAITest(t, dbSuper)
	createSuperSession(t, dbSuper, "super@pcs.com", "super_administrador", "token-super-disabled")
	if err := dbpkg.SetConfigValue(dbSuper, superAIEnabledConfigKey, "0", false); err != nil {
		t.Fatalf("disable ai: %v", err)
	}

	ctrl := NewSuperAIChatController(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/super/api/chat_con_ia_global/modelos", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-super-disabled"})
	rr := httptest.NewRecorder()

	ctrl.ModelosHandler(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d body=%s", rr.Code, rr.Body.String())
	}
}
