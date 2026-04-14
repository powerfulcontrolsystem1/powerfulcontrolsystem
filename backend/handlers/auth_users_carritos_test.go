package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
	_ "modernc.org/sqlite"
)

func openTestSQLite(t *testing.T, name string) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), name)
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = dbConn.Close()
		_ = os.Remove(dbPath)
	})
	return dbConn
}

func ensureEmpresaUsersSchema(t *testing.T, dbEmp *sql.DB) {
	t.Helper()

	_, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT,
		empresa_id INTEGER,
		documento_identidad TEXT,
		password_hash TEXT,
		password_salt TEXT,
		password_set INTEGER DEFAULT 0,
		password_actualizada_en TEXT,
		login_failed_attempts INTEGER DEFAULT 0,
		login_failed_last_at TEXT,
		login_locked_until TEXT,
		password_reset_token TEXT,
		password_reset_expira TEXT,
		password_reset_requested_en TEXT,
		rol_usuario_id INTEGER,
		email_confirmado INTEGER DEFAULT 0,
		email_confirmado_en TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`)
	if err != nil {
		t.Fatalf("create users schema: %v", err)
	}
}

func ensureSuperSchema(t *testing.T, dbSuper *sql.DB) {
	t.Helper()

	_, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT,
		photo TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		estado TEXT DEFAULT 'activo'
	);`)
	if err != nil {
		t.Fatalf("create administradores schema: %v", err)
	}

	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS sesiones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		admin_email TEXT,
		token TEXT,
		ip TEXT,
		user_agent TEXT,
		fecha_inicio TEXT,
		fecha_fin TEXT,
		fecha_creacion TEXT,
		activo INTEGER DEFAULT 1
	);`)
	if err != nil {
		t.Fatalf("create sesiones schema: %v", err)
	}
}

func ensureClientesSchema(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	_, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS clientes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre_razon_social TEXT
	);`)
	if err != nil {
		t.Fatalf("create clientes schema: %v", err)
	}
}

func ensureCarritosVentasSchema(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCodigosDescuentoSchema(dbEmp); err != nil {
		t.Fatalf("ensure codigos descuento schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaPropinasSchema(dbEmp); err != nil {
		t.Fatalf("ensure propinas schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaComisionesServicioSchema(dbEmp); err != nil {
		t.Fatalf("ensure comisiones servicio schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaConfiguracionOperativaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion operativa schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}
}

func TestHandleGoogleLoginRedirectIncludesLoginHint(t *testing.T) {
	h := HandleGoogleLogin("client-123", "http://localhost:8080/auth/google/callback")
	req := httptest.NewRequest(http.MethodGet, "/auth/google/login?login_hint=usuario@example.com", nil)
	req.Host = "localhost:8080"
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rr.Code)
	}
	loc := rr.Header().Get("Location")
	if loc == "" {
		t.Fatal("expected redirect location")
	}
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse redirect url: %v", err)
	}
	q := parsed.Query()
	if q.Get("client_id") != "client-123" {
		t.Fatalf("unexpected client_id: %q", q.Get("client_id"))
	}
	if q.Get("redirect_uri") != "http://localhost:8080/auth/google/callback" {
		t.Fatalf("unexpected redirect_uri: %q", q.Get("redirect_uri"))
	}
	if q.Get("login_hint") != "usuario@example.com" {
		t.Fatalf("unexpected login_hint: %q", q.Get("login_hint"))
	}
	if q.Get("prompt") != "select_account" {
		t.Fatalf("unexpected prompt: %q", q.Get("prompt"))
	}
}

func TestHandleGoogleLoginRedirectRewritesConfiguredLocalhostForPublicHost(t *testing.T) {
	h := HandleGoogleLogin("client-123", "http://localhost:8080/auth/google/callback")
	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "pos.example.com")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rr.Code)
	}
	loc := rr.Header().Get("Location")
	if loc == "" {
		t.Fatal("expected redirect location")
	}
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse redirect url: %v", err)
	}
	if got := parsed.Query().Get("redirect_uri"); got != "https://pos.example.com/auth/google/callback" {
		t.Fatalf("unexpected redirect_uri: %q", got)
	}
}

func TestHandleGoogleLoginRedirectUsesForwardedHostWhenRedirectNotConfigured(t *testing.T) {
	h := HandleGoogleLogin("client-123", "")
	req := httptest.NewRequest(http.MethodGet, "/auth/google/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "pos.example.com")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rr.Code)
	}
	loc := rr.Header().Get("Location")
	if loc == "" {
		t.Fatal("expected redirect location")
	}
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse redirect url: %v", err)
	}
	if got := parsed.Query().Get("redirect_uri"); got != "https://pos.example.com/auth/google/callback" {
		t.Fatalf("unexpected redirect_uri: %q", got)
	}

	hasRedirectCookie := false
	for _, ck := range rr.Result().Cookies() {
		if ck.Name == "oauth_redirect_url" {
			hasRedirectCookie = true
			break
		}
	}
	if !hasRedirectCookie {
		t.Fatal("expected oauth_redirect_url cookie")
	}
}

func TestEmpresaUsuarioLoginHandlerSuccess(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_login.db")
	dbSuper := openTestSQLite(t, "super_login.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	salt := "salt-login"
	hash := hashEmpresaUsuarioPassword("PasswordSegura1", salt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, 1, 'activo')`,
		"user@login.com", "Usuario Login", "vendedor", int64(10), "DOC-10", hash, salt, int64(2),
	)
	if err != nil {
		t.Fatalf("seed user login: %v", err)
	}

	h := EmpresaUsuarioLoginHandler(dbEmp, dbSuper)
	body := `{"email":"user@login.com","password":"PasswordSegura1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if ok, _ := resp["ok"].(bool); !ok {
		t.Fatalf("expected ok=true, got response=%v", resp)
	}

	hasSessionCookie := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session_token" && strings.TrimSpace(c.Value) != "" {
			hasSessionCookie = true
			break
		}
	}
	if !hasSessionCookie {
		t.Fatal("expected session_token cookie")
	}

	var sesionesCount int
	if err := dbSuper.QueryRow("SELECT COUNT(1) FROM sesiones").Scan(&sesionesCount); err != nil {
		t.Fatalf("count sesiones: %v", err)
	}
	if sesionesCount != 1 {
		t.Fatalf("expected 1 session, got %d", sesionesCount)
	}
}

func TestEmpresaUsuarioLoginHandlerRejectsWrongEmpresaScope(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_login_wrong_scope.db")
	dbSuper := openTestSQLite(t, "super_login_wrong_scope.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	salt := "salt-login-scope"
	hash := hashEmpresaUsuarioPassword("PasswordSegura1", salt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, 1, 'activo')`,
		"scope@login.com", "Usuario Scope", "vendedor", int64(10), "DOC-10", hash, salt, int64(2),
	)
	if err != nil {
		t.Fatalf("seed user login wrong scope: %v", err)
	}

	h := EmpresaUsuarioLoginHandler(dbEmp, dbSuper)
	body := `{"empresa_id":99,"email":"scope@login.com","password":"PasswordSegura1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "credenciales") {
		t.Fatalf("expected credentials message, got body=%s", rr.Body.String())
	}

	for _, c := range rr.Result().Cookies() {
		if c.Name == "session_token" && strings.TrimSpace(c.Value) != "" {
			t.Fatal("session_token must not be issued when empresa scope is invalid")
		}
	}

	var sesionesCount int
	if err := dbSuper.QueryRow("SELECT COUNT(1) FROM sesiones").Scan(&sesionesCount); err != nil {
		t.Fatalf("count sesiones: %v", err)
	}
	if sesionesCount != 0 {
		t.Fatalf("expected 0 sessions, got %d", sesionesCount)
	}
}

func TestEmpresaUsuarioLoginHandlerRejectsWrongEmpresaScopeFromQuery(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_login_wrong_scope_query.db")
	dbSuper := openTestSQLite(t, "super_login_wrong_scope_query.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	salt := "salt-login-query-scope"
	hash := hashEmpresaUsuarioPassword("PasswordSegura1", salt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, 1, 'activo')`,
		"scope-query@login.com", "Usuario Scope Query", "vendedor", int64(10), "DOC-10", hash, salt, int64(2),
	)
	if err != nil {
		t.Fatalf("seed user login wrong scope query: %v", err)
	}

	h := EmpresaUsuarioLoginHandler(dbEmp, dbSuper)
	body := `{"email":"scope-query@login.com","password":"PasswordSegura1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login?empresa_id=99", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "credenciales") {
		t.Fatalf("expected credentials message, got body=%s", rr.Body.String())
	}

	for _, c := range rr.Result().Cookies() {
		if c.Name == "session_token" && strings.TrimSpace(c.Value) != "" {
			t.Fatal("session_token must not be issued when empresa scope from query is invalid")
		}
	}

	var sesionesCount int
	if err := dbSuper.QueryRow("SELECT COUNT(1) FROM sesiones").Scan(&sesionesCount); err != nil {
		t.Fatalf("count sesiones: %v", err)
	}
	if sesionesCount != 0 {
		t.Fatalf("expected 0 sessions, got %d", sesionesCount)
	}
}

func TestEmpresaUsuarioSetPasswordHandlerSuccess(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_set_password.db")
	dbSuper := openTestSQLite(t, "super_set_password.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, '', '', 0, ?, 1, 'activo')`,
		"nuevo@empresa.com", "Nuevo Usuario", "auxiliar", int64(12), "DOC-22", int64(3),
	)
	if err != nil {
		t.Fatalf("seed user set password: %v", err)
	}

	h := EmpresaUsuarioSetPasswordHandler(dbEmp, dbSuper)
	body := `{"email":"nuevo@empresa.com","documento_identidad":"DOC-22","password":"ClaveNueva88","password_confirm":"ClaveNueva88"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/establecer_password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var passwordSet int
	var hash string
	var salt string
	err = dbEmp.QueryRow("SELECT COALESCE(password_set,0), COALESCE(password_hash,''), COALESCE(password_salt,'') FROM users WHERE email = ?", "nuevo@empresa.com").Scan(&passwordSet, &hash, &salt)
	if err != nil {
		t.Fatalf("query password fields: %v", err)
	}
	if passwordSet != 1 {
		t.Fatalf("expected password_set=1, got %d", passwordSet)
	}
	if strings.TrimSpace(hash) == "" || strings.TrimSpace(salt) == "" {
		t.Fatalf("expected non-empty password hash and salt, got hash=%q salt=%q", hash, salt)
	}

	hasSessionCookie := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "session_token" && strings.TrimSpace(c.Value) != "" {
			hasSessionCookie = true
			break
		}
	}
	if !hasSessionCookie {
		t.Fatal("expected session_token cookie")
	}
}

func TestEmpresaUsuarioSetPasswordHandlerRejectsWrongEmpresaScope(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_set_password_wrong_scope.db")
	dbSuper := openTestSQLite(t, "super_set_password_wrong_scope.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, '', '', 0, ?, 1, 'activo')`,
		"scopepass@empresa.com", "Usuario Scope Password", "auxiliar", int64(12), "DOC-22", int64(3),
	)
	if err != nil {
		t.Fatalf("seed user set password wrong scope: %v", err)
	}

	h := EmpresaUsuarioSetPasswordHandler(dbEmp, dbSuper)
	body := `{"empresa_id":99,"email":"scopepass@empresa.com","documento_identidad":"DOC-22","password":"ClaveNueva88","password_confirm":"ClaveNueva88"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/establecer_password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "usuario no encontrado") {
		t.Fatalf("expected not found message, got body=%s", rr.Body.String())
	}

	var passwordSet int
	if err := dbEmp.QueryRow("SELECT COALESCE(password_set,0) FROM users WHERE email = ?", "scopepass@empresa.com").Scan(&passwordSet); err != nil {
		t.Fatalf("query password_set: %v", err)
	}
	if passwordSet != 0 {
		t.Fatalf("expected password_set=0 for wrong scope, got %d", passwordSet)
	}

	var sesionesCount int
	if err := dbSuper.QueryRow("SELECT COUNT(1) FROM sesiones").Scan(&sesionesCount); err != nil {
		t.Fatalf("count sesiones: %v", err)
	}
	if sesionesCount != 0 {
		t.Fatalf("expected 0 sessions, got %d", sesionesCount)
	}
}

func TestEmpresaUsuarioSetPasswordHandlerRejectsWrongEmpresaScopeFromQuery(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_set_password_wrong_scope_query.db")
	dbSuper := openTestSQLite(t, "super_set_password_wrong_scope_query.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, '', '', 0, ?, 1, 'activo')`,
		"scopepass-query@empresa.com", "Usuario Scope Password Query", "auxiliar", int64(12), "DOC-22", int64(3),
	)
	if err != nil {
		t.Fatalf("seed user set password wrong scope query: %v", err)
	}

	h := EmpresaUsuarioSetPasswordHandler(dbEmp, dbSuper)
	body := `{"email":"scopepass-query@empresa.com","documento_identidad":"DOC-22","password":"ClaveNueva88","password_confirm":"ClaveNueva88"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/establecer_password?empresa_id=99", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "usuario no encontrado") {
		t.Fatalf("expected not found message, got body=%s", rr.Body.String())
	}

	var passwordSet int
	if err := dbEmp.QueryRow("SELECT COALESCE(password_set,0) FROM users WHERE email = ?", "scopepass-query@empresa.com").Scan(&passwordSet); err != nil {
		t.Fatalf("query password_set: %v", err)
	}
	if passwordSet != 0 {
		t.Fatalf("expected password_set=0 for wrong scope from query, got %d", passwordSet)
	}

	var sesionesCount int
	if err := dbSuper.QueryRow("SELECT COUNT(1) FROM sesiones").Scan(&sesionesCount); err != nil {
		t.Fatalf("count sesiones: %v", err)
	}
	if sesionesCount != 0 {
		t.Fatalf("expected 0 sessions, got %d", sesionesCount)
	}
}

func TestEmpresaCarritosCompraAndItemsFlow(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createCarritoBody := `{"empresa_id":1,"nombre":"Caja Principal","canal_venta":"mostrador","moneda":"COP"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(createCarritoBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create carrito: %v", err)
	}
	carritoIDFloat, ok := createResp["id"].(float64)
	if !ok || carritoIDFloat <= 0 {
		t.Fatalf("invalid carrito id in response: %v", createResp)
	}
	carritoID := int64(carritoIDFloat)

	itemsHandler := EmpresaCarritoItemsHandler(dbEmp)
	createItemBody := `{"empresa_id":1,"carrito_id":` + strconv.FormatInt(carritoID, 10) + `,"descripcion":"Producto A","cantidad":2,"precio_unitario":1500}`
	createItemReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra/items", strings.NewReader(createItemBody))
	createItemReq.Header.Set("Content-Type", "application/json")
	createItemRR := httptest.NewRecorder()
	itemsHandler.ServeHTTP(createItemRR, createItemReq)
	if createItemRR.Code != http.StatusCreated {
		t.Fatalf("expected item create status %d, got %d body=%s", http.StatusCreated, createItemRR.Code, createItemRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=1", nil)
	listRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}

	var rows []dbpkg.CarritoCompra
	if err := json.Unmarshal(listRR.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list carritos: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 carrito, got %d", len(rows))
	}
	if rows[0].ItemCount != 1 {
		t.Fatalf("expected item_count=1, got %d", rows[0].ItemCount)
	}
	if rows[0].Total <= 0 {
		t.Fatalf("expected total > 0 after item creation, got %v", rows[0].Total)
	}
	if rows[0].EstadoVenta != "venta_abierta" {
		t.Fatalf("expected estado_venta=venta_abierta, got %q", rows[0].EstadoVenta)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"total_pagado":3000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected pay status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	var payResp map[string]interface{}
	if err := json.Unmarshal(payRR.Body.Bytes(), &payResp); err != nil {
		t.Fatalf("decode pay response: %v", err)
	}
	if got, _ := payResp["estado_venta"].(string); got != "venta_pagada" {
		t.Fatalf("expected pay response estado_venta=venta_pagada, got %q", got)
	}

	listPaidReq := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=1&include_inactive=1", nil)
	listPaidRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(listPaidRR, listPaidReq)
	if listPaidRR.Code != http.StatusOK {
		t.Fatalf("expected list after pay status %d, got %d body=%s", http.StatusOK, listPaidRR.Code, listPaidRR.Body.String())
	}

	var paidRows []dbpkg.CarritoCompra
	if err := json.Unmarshal(listPaidRR.Body.Bytes(), &paidRows); err != nil {
		t.Fatalf("decode list after pay: %v", err)
	}
	if len(paidRows) == 0 {
		t.Fatalf("expected at least one carrito after pay, got %d", len(paidRows))
	}
	if paidRows[0].EstadoVenta != "venta_pagada" {
		t.Fatalf("expected estado_venta=venta_pagada after pay, got %q", paidRows[0].EstadoVenta)
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 1, dbpkg.EmpresaEventoContableFilter{Modulo: "ventas", Limit: 10})
	if err != nil {
		t.Fatalf("list eventos contables: %v", err)
	}
	if len(eventos) == 0 {
		t.Fatalf("expected at least one evento contable de ventas")
	}
	if eventos[0].Evento != "venta_pagada" {
		t.Fatalf("expected latest evento venta_pagada, got %q", eventos[0].Evento)
	}
}

func TestEmpresaCarritosCompraMetricasEstacionIncluyeCorrecciones(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_metricas_estacion.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"codigo":"EST-1-7","nombre":"Estacion 7","canal_venta":"mostrador","moneda":"COP","referencia_externa":"ESTACION_7"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:           1,
		CarritoID:           carritoID,
		TipoItem:            "otro",
		CodigoItem:          "METRICA-ITEM-" + strconv.FormatInt(carritoID, 10),
		Descripcion:         "Consumo base estacion",
		UnidadMedida:        "unidad",
		Cantidad:            1,
		PrecioUnitario:      10000,
		DescuentoPorcentaje: 0,
		ImpuestoPorcentaje:  0,
		ImpuestoCodigo:      "IVA",
		UsuarioCreador:      "test@empresa.com",
		Estado:              "activo",
	}); err != nil {
		t.Fatalf("seed carrito item: %v", err)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"efectivo","total_pagado":10000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected pay status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	correctionReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=anular_cierre_parcial", strings.NewReader(`{"monto_anulado":2000,"motivo":"correccion de caja"}`))
	correctionReq.Header.Set("Content-Type", "application/json")
	correctionRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(correctionRR, correctionReq)
	if correctionRR.Code != http.StatusOK {
		t.Fatalf("expected correction status %d, got %d body=%s", http.StatusOK, correctionRR.Code, correctionRR.Body.String())
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=1&action=metricas_estacion&estacion_id=7&days=30&limit=5", nil)
	metricsRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(metricsRR, metricsReq)
	if metricsRR.Code != http.StatusOK {
		t.Fatalf("expected metrics status %d, got %d body=%s", http.StatusOK, metricsRR.Code, metricsRR.Body.String())
	}

	var metricsResp struct {
		Rows    []dbpkg.CarritoStationMetricSummary `json:"rows"`
		Resumen map[string]interface{}              `json:"resumen"`
	}
	if err := json.Unmarshal(metricsRR.Body.Bytes(), &metricsResp); err != nil {
		t.Fatalf("decode metrics response: %v", err)
	}
	if len(metricsResp.Rows) == 0 {
		t.Fatalf("expected at least one station summary row, got %d", len(metricsResp.Rows))
	}
	row := metricsResp.Rows[0]
	if row.EstacionID != 7 {
		t.Fatalf("expected estacion_id=7, got %d", row.EstacionID)
	}
	if row.VentasPagadas < 1 {
		t.Fatalf("expected ventas_pagadas >= 1, got %d", row.VentasPagadas)
	}
	if row.Correcciones < 1 {
		t.Fatalf("expected correcciones >= 1, got %d", row.Correcciones)
	}
	if row.MontoAnulado < 2000 {
		t.Fatalf("expected monto_anulado >= 2000, got %.2f", row.MontoAnulado)
	}

	ventasResumen, ok := metricsResp.Resumen["ventas_pagadas"].(float64)
	if !ok || ventasResumen < 1 {
		t.Fatalf("expected resumen ventas_pagadas >= 1, got %v", metricsResp.Resumen["ventas_pagadas"])
	}
}

func TestEmpresaCarritosCompraEstadoVentaSuspendida(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_suspendida.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createCarritoBody := `{"empresa_id":1,"nombre":"Caja Secundaria","canal_venta":"mostrador","moneda":"COP"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(createCarritoBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create carrito: %v", err)
	}
	carritoIDFloat, ok := createResp["id"].(float64)
	if !ok || carritoIDFloat <= 0 {
		t.Fatalf("invalid carrito id in response: %v", createResp)
	}
	carritoID := int64(carritoIDFloat)

	disableReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=desactivar", nil)
	disableRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(disableRR, disableReq)

	if disableRR.Code != http.StatusOK {
		t.Fatalf("expected disable status %d, got %d body=%s", http.StatusOK, disableRR.Code, disableRR.Body.String())
	}

	var disableResp map[string]interface{}
	if err := json.Unmarshal(disableRR.Body.Bytes(), &disableResp); err != nil {
		t.Fatalf("decode disable response: %v", err)
	}
	if got, _ := disableResp["estado_venta"].(string); got != "venta_suspendida" {
		t.Fatalf("expected disable response estado_venta=venta_suspendida, got %q", got)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=1&include_inactive=1", nil)
	listRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}

	var rows []dbpkg.CarritoCompra
	if err := json.Unmarshal(listRR.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list carritos: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected at least one carrito, got %d", len(rows))
	}
	if rows[0].EstadoVenta != "venta_suspendida" {
		t.Fatalf("expected estado_venta=venta_suspendida, got %q", rows[0].EstadoVenta)
	}
}

func TestEmpresaCarritosCompraRejectsDoublePago(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_double_pay.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Doble Pago","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	payURL := "/api/empresa/carritos_compra?empresa_id=1&id=" + strconv.FormatInt(carritoID, 10) + "&action=pagar_estacion"
	payReq := httptest.NewRequest(http.MethodPut, payURL, strings.NewReader(`{"total_pagado":0}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected first pay status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	payAgainReq := httptest.NewRequest(http.MethodPut, payURL, strings.NewReader(`{"total_pagado":0}`))
	payAgainReq.Header.Set("Content-Type", "application/json")
	payAgainRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payAgainRR, payAgainReq)
	if payAgainRR.Code != http.StatusConflict {
		t.Fatalf("expected second pay status %d, got %d body=%s", http.StatusConflict, payAgainRR.Code, payAgainRR.Body.String())
	}
}

func TestEmpresaCarritosCompraRejectsReabrirVentaPagada(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_reopen_paid.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Reabrir","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"total_pagado":0}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected pay status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	reopenReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=reabrir", nil)
	reopenRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(reopenRR, reopenReq)
	if reopenRR.Code != http.StatusConflict {
		t.Fatalf("expected reopen paid status %d, got %d body=%s", http.StatusConflict, reopenRR.Code, reopenRR.Body.String())
	}
}

func TestEmpresaCarritosCompraRejectsActivarEstacionPagadaSinReset(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_activate_paid.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Activar Pagada","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"total_pagado":0}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected pay status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	activateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=activar_estacion", nil)
	activateRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(activateRR, activateReq)
	if activateRR.Code != http.StatusConflict {
		t.Fatalf("expected activar_estacion pagada sin reset status %d, got %d body=%s", http.StatusConflict, activateRR.Code, activateRR.Body.String())
	}

	activateResetReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=activar_estacion&reset_items=1", nil)
	activateResetRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(activateResetRR, activateResetReq)
	if activateResetRR.Code != http.StatusOK {
		t.Fatalf("expected activar_estacion con reset status %d, got %d body=%s", http.StatusOK, activateResetRR.Code, activateResetRR.Body.String())
	}
}

func TestEmpresaCarritosCompraRejectsMetodoPagoInvalido(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_metodo_invalido.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Metodo Invalido","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:           1,
		CarritoID:           carritoID,
		TipoItem:            "otro",
		CodigoItem:          "PROPINA-ITEM-" + strconv.FormatInt(carritoID, 10),
		Descripcion:         "Consumo base para propina",
		UnidadMedida:        "unidad",
		Cantidad:            1,
		PrecioUnitario:      10000,
		DescuentoPorcentaje: 0,
		ImpuestoPorcentaje:  0,
		ImpuestoCodigo:      "IVA",
		UsuarioCreador:      "test@empresa.com",
		Estado:              "activo",
	}); err != nil {
		t.Fatalf("seed carrito item para propina: %v", err)
	}

	var subtotalSeed float64
	var totalSeed float64
	if err := dbEmp.QueryRow(`SELECT COALESCE(subtotal, 0), COALESCE(total, 0) FROM carritos_compras WHERE empresa_id = 1 AND id = ?`, carritoID).Scan(&subtotalSeed, &totalSeed); err != nil {
		t.Fatalf("read seeded carrito totals: %v", err)
	}
	if math.Abs(subtotalSeed-10000) > 0.001 || math.Abs(totalSeed-10000) > 0.001 {
		t.Fatalf("expected seeded totals subtotal=10000 total=10000, got subtotal=%.4f total=%.4f", subtotalSeed, totalSeed)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"cripto","total_pagado":10000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)

	if payRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid payment method status %d, got %d body=%s", http.StatusBadRequest, payRR.Code, payRR.Body.String())
	}
}

func TestEmpresaCarritosCompraPermitePagoTransferenciaBancaria(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_transferencia_ok.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Transferencia","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:           1,
		CarritoID:           carritoID,
		TipoItem:            "otro",
		CodigoItem:          "PROPINA-ITEM-" + strconv.FormatInt(carritoID, 10),
		Descripcion:         "Consumo base para propina",
		UnidadMedida:        "unidad",
		Cantidad:            1,
		PrecioUnitario:      10000,
		DescuentoPorcentaje: 0,
		ImpuestoPorcentaje:  0,
		ImpuestoCodigo:      "IVA",
		UsuarioCreador:      "test@empresa.com",
		Estado:              "activo",
	}); err != nil {
		t.Fatalf("seed carrito item para propina: %v", err)
	}

	var subtotalSeed float64
	var totalSeed float64
	if err := dbEmp.QueryRow(`SELECT COALESCE(subtotal, 0), COALESCE(total, 0) FROM carritos_compras WHERE empresa_id = 1 AND id = ?`, carritoID).Scan(&subtotalSeed, &totalSeed); err != nil {
		t.Fatalf("read seeded carrito totals: %v", err)
	}
	if math.Abs(subtotalSeed-10000) > 0.001 || math.Abs(totalSeed-10000) > 0.001 {
		t.Fatalf("expected seeded totals subtotal=10000 total=10000, got subtotal=%.4f total=%.4f", subtotalSeed, totalSeed)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"transferencia","referencia_pago":"TRX-4455","total_pagado":10000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)

	if payRR.Code != http.StatusOK {
		t.Fatalf("expected transferencia payment status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	var metodoPago string
	var referenciaPago string
	err := dbEmp.QueryRow(`SELECT COALESCE(metodo_pago,''), COALESCE(referencia_pago,'') FROM carritos_compras WHERE empresa_id = 1 AND id = ?`, carritoID).Scan(&metodoPago, &referenciaPago)
	if err != nil {
		t.Fatalf("query paid carrito: %v", err)
	}
	if metodoPago != "transferencia_bancaria" {
		t.Fatalf("expected metodo_pago transferencia_bancaria, got %q", metodoPago)
	}
	if referenciaPago != "TRX-4455" {
		t.Fatalf("expected referencia_pago TRX-4455, got %q", referenciaPago)
	}
}

func TestEmpresaCarritosCompraExigeReferenciaTransferenciaBancaria(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_transferencia_ref.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Transferencia Ref","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	if _, err := dbEmp.Exec(`UPDATE carritos_compras SET subtotal = 10000, total = 10000, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = 1 AND id = ?`, carritoID); err != nil {
		t.Fatalf("seed carrito total: %v", err)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"transferencia_bancaria","referencia_pago":"12","total_pagado":10000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)

	if payRR.Code != http.StatusBadRequest {
		t.Fatalf("expected transferencia without reference status %d, got %d body=%s", http.StatusBadRequest, payRR.Code, payRR.Body.String())
	}
	if !strings.Contains(strings.ToLower(payRR.Body.String()), "referencia_pago") {
		t.Fatalf("expected referencia_pago message, got body=%s", payRR.Body.String())
	}
}

func TestEmpresaCarritosCompraAplicaPropinaSegunConfiguracion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_propina_ok.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	if _, err := dbpkg.UpsertEmpresaConfiguracionOperativa(dbEmp, dbpkg.EmpresaConfiguracionOperativa{
		EmpresaID:                       1,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		UsuarioCreador:                  "test@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert empresa configuracion operativa: %v", err)
	}

	if _, err := dbpkg.UpsertEmpresaPropinasConfiguracion(dbEmp, dbpkg.EmpresaPropinasConfiguracion{
		EmpresaID:              1,
		HabilitarPropina:       true,
		PorcentajePropina:      10,
		ModoDistribucion:       dbpkg.EmpresaPropinaModoPorUsuario,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "test@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("upsert propinas config: %v", err)
	}

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Propina","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:           1,
		CarritoID:           carritoID,
		TipoItem:            "otro",
		CodigoItem:          "PROPINA-ITEM-" + strconv.FormatInt(carritoID, 10),
		Descripcion:         "Consumo base para propina",
		UnidadMedida:        "unidad",
		Cantidad:            1,
		PrecioUnitario:      10000,
		DescuentoPorcentaje: 0,
		ImpuestoPorcentaje:  0,
		ImpuestoCodigo:      "IVA",
		UsuarioCreador:      "test@empresa.com",
		Estado:              "activo",
	}); err != nil {
		t.Fatalf("seed carrito item para propina: %v", err)
	}

	var subtotalSeed float64
	var totalSeed float64
	if err := dbEmp.QueryRow(`SELECT COALESCE(subtotal, 0), COALESCE(total, 0) FROM carritos_compras WHERE empresa_id = 1 AND id = ?`, carritoID).Scan(&subtotalSeed, &totalSeed); err != nil {
		t.Fatalf("read seeded carrito totals: %v", err)
	}
	if math.Abs(subtotalSeed-10000) > 0.001 || math.Abs(totalSeed-10000) > 0.001 {
		t.Fatalf("expected seeded totals subtotal=10000 total=10000, got subtotal=%.4f total=%.4f", subtotalSeed, totalSeed)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"efectivo","aplicar_propina":true,"total_pagado":11000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)

	if payRR.Code != http.StatusOK {
		t.Fatalf("expected payment status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	var payResp map[string]interface{}
	if err := json.Unmarshal(payRR.Body.Bytes(), &payResp); err != nil {
		t.Fatalf("decode pay response: %v", err)
	}
	propinaRaw, ok := payResp["propina"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected propina block in response, got %v", payResp)
	}
	if applied, _ := propinaRaw["aplicada"].(bool); !applied {
		t.Fatalf("expected propina aplicada=true, got %v", propinaRaw["aplicada"])
	}
	if monto, _ := propinaRaw["monto"].(float64); math.Abs(monto-1000) > 0.001 {
		t.Fatalf("expected propina monto 1000, got %.4f body=%s", monto, payRR.Body.String())
	}

	movs, err := dbpkg.ListEmpresaPropinaMovimientos(dbEmp, 1, dbpkg.EmpresaPropinaMovimientoFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list propina movements: %v", err)
	}
	if len(movs) == 0 {
		t.Fatal("expected at least one propina movement")
	}
	if movs[0].CarritoID != carritoID {
		t.Fatalf("expected propina carrito_id %d, got %d", carritoID, movs[0].CarritoID)
	}
	if math.Abs(movs[0].MontoPropina-1000) > 0.001 {
		t.Fatalf("expected stored propina 1000, got %.4f", movs[0].MontoPropina)
	}
}

func TestEmpresaCarritosCompraRegistraComisionServicioPorLavador(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_comision_servicio.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	if _, err := dbpkg.UpsertEmpresaComisionesServicioConfiguracion(dbEmp, dbpkg.EmpresaComisionesServicioConfiguracion{
		EmpresaID:              1,
		HabilitarComisiones:    true,
		PorcentajeComision:     15,
		FiltroServicio:         "lavado",
		AplicarAutomaticamente: true,
		UsuarioCreador:         "test@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("upsert comisiones config: %v", err)
	}

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Comision","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	servicioID, err := dbpkg.CreateServicio(dbEmp, dbpkg.Servicio{
		EmpresaID:          1,
		Codigo:             "LAV-001",
		Nombre:             "Lavado premium",
		Categoria:          "lavado",
		Precio:             10000,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     "test@empresa.com",
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create servicio: %v", err)
	}

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:           1,
		CarritoID:           carritoID,
		TipoItem:            "servicio",
		ReferenciaID:        servicioID,
		CodigoItem:          "LAV-001",
		Descripcion:         "Lavado premium de auto",
		UnidadMedida:        "servicio",
		Cantidad:            1,
		PrecioUnitario:      10000,
		DescuentoPorcentaje: 0,
		ImpuestoPorcentaje:  0,
		ImpuestoCodigo:      "IVA",
		UsuarioCreador:      "test@empresa.com",
		Estado:              "activo",
	}); err != nil {
		t.Fatalf("create carrito item servicio: %v", err)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"efectivo","total_pagado":10000,"usuario_lavador":"lavador1@empresa.com"}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)

	if payRR.Code != http.StatusOK {
		t.Fatalf("expected payment status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	var payResp map[string]interface{}
	if err := json.Unmarshal(payRR.Body.Bytes(), &payResp); err != nil {
		t.Fatalf("decode pay response: %v", err)
	}
	comisionRaw, ok := payResp["comision"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected comision block in response, got %v", payResp)
	}
	if applied, _ := comisionRaw["aplicada"].(bool); !applied {
		t.Fatalf("expected comision aplicada=true, got %v", comisionRaw["aplicada"])
	}
	if monto, _ := comisionRaw["monto_comision"].(float64); math.Abs(monto-1500) > 0.001 {
		t.Fatalf("expected comision monto 1500, got %.4f", monto)
	}

	movs, err := dbpkg.ListEmpresaComisionServicioMovimientos(dbEmp, 1, dbpkg.EmpresaComisionServicioMovimientoFilter{
		UsuarioLavador: "lavador1",
		Limit:          10,
	})
	if err != nil {
		t.Fatalf("list comisiones movimientos: %v", err)
	}
	if len(movs) == 0 {
		t.Fatal("expected at least one comision movement")
	}
	if movs[0].CarritoID != carritoID {
		t.Fatalf("expected comision carrito_id %d, got %d", carritoID, movs[0].CarritoID)
	}
	if movs[0].UsuarioLavador != "lavador1@empresa.com" {
		t.Fatalf("expected usuario_lavador lavador1@empresa.com, got %q", movs[0].UsuarioLavador)
	}
	if math.Abs(movs[0].MontoComision-1500) > 0.001 {
		t.Fatalf("expected stored comision 1500, got %.4f", movs[0].MontoComision)
	}
}

func TestEmpresaCarritosCompraCodigoDescuentoConsumeUso(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_codigo_uso.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)

	codigoID, err := dbpkg.CreateCodigoDescuento(dbEmp, dbpkg.CodigoDescuento{
		EmpresaID:      1,
		Codigo:         "PROMOUNO",
		TipoDescuento:  "valor_fijo",
		Valor:          15000,
		Moneda:         "COP",
		UsosMaximos:    1,
		UsuarioCreador: "test",
	})
	if err != nil {
		t.Fatalf("create discount code: %v", err)
	}
	if codigoID <= 0 {
		t.Fatalf("expected codigo id > 0, got %d", codigoID)
	}

	createAndSeedCarrito := func(nombre string) int64 {
		createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"`+nombre+`","canal_venta":"mostrador","moneda":"COP"}`))
		createReq.Header.Set("Content-Type", "application/json")
		createRR := httptest.NewRecorder()
		carritosHandler.ServeHTTP(createRR, createReq)
		if createRR.Code != http.StatusCreated {
			t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
		}

		var createResp map[string]interface{}
		if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
			t.Fatalf("decode create response: %v", err)
		}
		carritoID := int64(createResp["id"].(float64))
		if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
			EmpresaID:           1,
			CarritoID:           carritoID,
			TipoItem:            "otro",
			CodigoItem:          "PROMO-ITEM-" + strconv.FormatInt(carritoID, 10),
			Descripcion:         "Consumo base para descuento",
			UnidadMedida:        "unidad",
			Cantidad:            1,
			PrecioUnitario:      12000,
			DescuentoPorcentaje: 0,
			ImpuestoPorcentaje:  0,
			ImpuestoCodigo:      "IVA",
			UsuarioCreador:      "test@empresa.com",
			Estado:              "activo",
		}); err != nil {
			t.Fatalf("seed carrito item para descuento: %v", err)
		}
		return carritoID
	}

	firstCarritoID := createAndSeedCarrito("Caja Promo 1")
	firstPayReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(firstCarritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"codigo_descuento","descuento_tipo":"code","descuento_codigo":"PROMOUNO","codigo_descuento":"PROMOUNO","total_pagado":0}`))
	firstPayReq.Header.Set("Content-Type", "application/json")
	firstPayRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(firstPayRR, firstPayReq)
	if firstPayRR.Code != http.StatusOK {
		t.Fatalf("expected first discount payment status %d, got %d body=%s", http.StatusOK, firstPayRR.Code, firstPayRR.Body.String())
	}

	secondCarritoID := createAndSeedCarrito("Caja Promo 2")
	secondPayReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(secondCarritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"codigo_descuento","descuento_tipo":"code","descuento_codigo":"PROMOUNO","codigo_descuento":"PROMOUNO","total_pagado":0}`))
	secondPayReq.Header.Set("Content-Type", "application/json")
	secondPayRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(secondPayRR, secondPayReq)
	if secondPayRR.Code != http.StatusBadRequest {
		t.Fatalf("expected second discount payment status %d, got %d body=%s", http.StatusBadRequest, secondPayRR.Code, secondPayRR.Body.String())
	}
}

func TestEmpresaCarritosCompraBloqueaMetodoPagoSegunRol(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_config_operativa_rol.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	if _, err := dbpkg.UpsertEmpresaConfiguracionOperativa(dbEmp, dbpkg.EmpresaConfiguracionOperativa{
		EmpresaID:                       1,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert empresa configuracion operativa: %v", err)
	}
	if _, err := dbpkg.UpsertEmpresaConfiguracionOperativaRol(dbEmp, dbpkg.EmpresaConfiguracionOperativaRol{
		EmpresaID:                       1,
		Rol:                             "cajero",
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: false,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert role configuracion operativa: %v", err)
	}

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Config Rol","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	if _, err := dbEmp.Exec(`UPDATE carritos_compras SET subtotal = 10000, total = 10000, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = 1 AND id = ?`, carritoID); err != nil {
		t.Fatalf("seed carrito totals: %v", err)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"transferencia_bancaria","referencia_pago":"TRX-ROL-001","total_pagado":10000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payReq.Header.Set("X-Admin-Role", "cajero")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden status %d, got %d body=%s", http.StatusForbidden, payRR.Code, payRR.Body.String())
	}
}

func TestEmpresaCarritosCompraRespetaBloqueoPropinaYComisionPorRol(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_config_operativa_propina_comision.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	if _, err := dbpkg.UpsertEmpresaConfiguracionOperativa(dbEmp, dbpkg.EmpresaConfiguracionOperativa{
		EmpresaID:                       1,
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               true,
		HabilitarComisiones:             true,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert empresa configuracion operativa: %v", err)
	}
	if _, err := dbpkg.UpsertEmpresaConfiguracionOperativaRol(dbEmp, dbpkg.EmpresaConfiguracionOperativaRol{
		EmpresaID:                       1,
		Rol:                             "cajero",
		MetodoPagoEfectivo:              true,
		MetodoPagoTarjetaCredito:        true,
		MetodoPagoTarjetaDebito:         true,
		MetodoPagoTransferenciaBancaria: true,
		MetodoPagoMixto:                 true,
		MetodoPagoCodigoDescuento:       true,
		HabilitarPropinas:               false,
		HabilitarComisiones:             false,
		UsuarioCreador:                  "qa@empresa.com",
		Estado:                          "activo",
	}); err != nil {
		t.Fatalf("upsert role configuracion operativa: %v", err)
	}

	if _, err := dbpkg.UpsertEmpresaPropinasConfiguracion(dbEmp, dbpkg.EmpresaPropinasConfiguracion{
		EmpresaID:              1,
		HabilitarPropina:       true,
		PorcentajePropina:      10,
		ModoDistribucion:       dbpkg.EmpresaPropinaModoPorUsuario,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("upsert propinas config: %v", err)
	}

	if _, err := dbpkg.UpsertEmpresaComisionesServicioConfiguracion(dbEmp, dbpkg.EmpresaComisionesServicioConfiguracion{
		EmpresaID:              1,
		HabilitarComisiones:    true,
		PorcentajeComision:     15,
		FiltroServicio:         "lavado",
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("upsert comisiones config: %v", err)
	}

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Config Propina Comision","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	servicioID, err := dbpkg.CreateServicio(dbEmp, dbpkg.Servicio{
		EmpresaID:          1,
		Codigo:             "LAV-ROL-01",
		Nombre:             "Lavado rol",
		Categoria:          "lavado",
		Precio:             10000,
		ImpuestoPorcentaje: 0,
		UsuarioCreador:     "qa@empresa.com",
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create servicio: %v", err)
	}

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:           1,
		CarritoID:           carritoID,
		TipoItem:            "servicio",
		ReferenciaID:        servicioID,
		CodigoItem:          "LAV-ROL-01",
		Descripcion:         "Lavado rol por prueba",
		UnidadMedida:        "servicio",
		Cantidad:            1,
		PrecioUnitario:      10000,
		DescuentoPorcentaje: 0,
		ImpuestoPorcentaje:  0,
		ImpuestoCodigo:      "IVA",
		UsuarioCreador:      "qa@empresa.com",
		Estado:              "activo",
	}); err != nil {
		t.Fatalf("create carrito item: %v", err)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"efectivo","total_pagado":10000,"aplicar_propina":true,"usuario_lavador":"lavador-rol@empresa.com"}`))
	payReq.Header.Set("Content-Type", "application/json")
	payReq.Header.Set("X-Admin-Role", "cajero")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected payment status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	var payResp map[string]interface{}
	if err := json.Unmarshal(payRR.Body.Bytes(), &payResp); err != nil {
		t.Fatalf("decode pay response: %v", err)
	}

	propinaRaw, ok := payResp["propina"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected propina block in response, got %v", payResp)
	}
	if aplicada, _ := propinaRaw["aplicada"].(bool); aplicada {
		t.Fatalf("expected propina aplicada=false, got %v", propinaRaw["aplicada"])
	}
	if warning, _ := propinaRaw["warning"].(string); !strings.Contains(strings.ToLower(warning), "deshabilitadas") {
		t.Fatalf("expected propina warning about disabled policy, got %q", warning)
	}

	comisionRaw, ok := payResp["comision"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected comision block in response, got %v", payResp)
	}
	if aplicada, _ := comisionRaw["aplicada"].(bool); aplicada {
		t.Fatalf("expected comision aplicada=false, got %v", comisionRaw["aplicada"])
	}
	if warning, _ := comisionRaw["warning"].(string); !strings.Contains(strings.ToLower(warning), "deshabilitadas") {
		t.Fatalf("expected comision warning about disabled policy, got %q", warning)
	}

	propinaMovs, err := dbpkg.ListEmpresaPropinaMovimientos(dbEmp, 1, dbpkg.EmpresaPropinaMovimientoFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list propina movimientos: %v", err)
	}
	if len(propinaMovs) != 0 {
		t.Fatalf("expected no propina movements, got %d", len(propinaMovs))
	}

	comisionMovs, err := dbpkg.ListEmpresaComisionServicioMovimientos(dbEmp, 1, dbpkg.EmpresaComisionServicioMovimientoFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list comision movimientos: %v", err)
	}
	if len(comisionMovs) != 0 {
		t.Fatalf("expected no comision movements, got %d", len(comisionMovs))
	}
}

func TestEmpresaCarritosCompraRecuperarInterrumpidoConAuditoria(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_recover_interrumpido.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)

	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Recover","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	disableReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=desactivar", nil)
	disableRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(disableRR, disableReq)
	if disableRR.Code != http.StatusOK {
		t.Fatalf("expected desactivar status %d, got %d body=%s", http.StatusOK, disableRR.Code, disableRR.Body.String())
	}

	recoverReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=recuperar_interrumpido", nil)
	recoverReq.Header.Set("X-Admin-Email", "cajero@empresa.com")
	recoverRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(recoverRR, recoverReq)
	if recoverRR.Code != http.StatusOK {
		t.Fatalf("expected recuperar_interrumpido status %d, got %d body=%s", http.StatusOK, recoverRR.Code, recoverRR.Body.String())
	}

	var recoverResp map[string]interface{}
	if err := json.Unmarshal(recoverRR.Body.Bytes(), &recoverResp); err != nil {
		t.Fatalf("decode recover response: %v", err)
	}
	if got, _ := recoverResp["estado_venta"].(string); got != "venta_abierta" {
		t.Fatalf("expected estado_venta=venta_abierta, got %q", got)
	}

	auditoriaRows, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 1, dbpkg.EmpresaAuditoriaEventoFilter{
		Modulo: "ventas",
		Accion: "recuperar_interrumpido",
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(auditoriaRows) == 0 {
		t.Fatal("expected auditoria event for recuperar_interrumpido")
	}
	if auditoriaRows[0].RecursoID != carritoID {
		t.Fatalf("expected auditoria recurso_id=%d, got %d", carritoID, auditoriaRows[0].RecursoID)
	}
}

func TestEmpresaCarritosCompraPagoMixtoValidaSuma(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_pago_mixto_validacion.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)

	createAndSeed := func(nombre string) int64 {
		createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"`+nombre+`","canal_venta":"mostrador","moneda":"COP"}`))
		createReq.Header.Set("Content-Type", "application/json")
		createRR := httptest.NewRecorder()
		carritosHandler.ServeHTTP(createRR, createReq)
		if createRR.Code != http.StatusCreated {
			t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
		}

		var createResp map[string]interface{}
		if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
			t.Fatalf("decode create response: %v", err)
		}
		carritoID := int64(createResp["id"].(float64))

		if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
			EmpresaID:           1,
			CarritoID:           carritoID,
			TipoItem:            "otro",
			CodigoItem:          "MIX-ITEM-" + strconv.FormatInt(carritoID, 10),
			Descripcion:         "Consumo base para pago mixto",
			UnidadMedida:        "unidad",
			Cantidad:            1,
			PrecioUnitario:      10000,
			DescuentoPorcentaje: 0,
			ImpuestoPorcentaje:  0,
			ImpuestoCodigo:      "IVA",
			UsuarioCreador:      "qa@empresa.com",
			Estado:              "activo",
		}); err != nil {
			t.Fatalf("seed carrito item pago mixto: %v", err)
		}

		return carritoID
	}

	carritoOK := createAndSeed("Caja Mixto OK")
	payMixedReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoOK, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"mixto","pagos_mixtos":[{"metodo":"efectivo","monto":4000},{"metodo":"tarjeta_debito","monto":6000,"referencia":"TD-0001"}]}`))
	payMixedReq.Header.Set("Content-Type", "application/json")
	payMixedRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payMixedRR, payMixedReq)
	if payMixedRR.Code != http.StatusOK {
		t.Fatalf("expected mixed payment status %d, got %d body=%s", http.StatusOK, payMixedRR.Code, payMixedRR.Body.String())
	}

	carritoPaid, err := dbpkg.GetCarritoCompraByID(dbEmp, 1, carritoOK)
	if err != nil {
		t.Fatalf("get paid carrito: %v", err)
	}
	if strings.TrimSpace(carritoPaid.MetodoPago) != "mixto" {
		t.Fatalf("expected metodo_pago=mixto, got %q", carritoPaid.MetodoPago)
	}
	if !strings.Contains(strings.ToLower(strings.TrimSpace(carritoPaid.ReferenciaPago)), "mixto[") {
		t.Fatalf("expected referencia_pago to include mixed detail, got %q", carritoPaid.ReferenciaPago)
	}

	carritoInvalid := createAndSeed("Caja Mixto Error")
	payInvalidReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoInvalid, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"mixto","pagos_mixtos":[{"metodo":"efectivo","monto":3000},{"metodo":"tarjeta_debito","monto":6000,"referencia":"TD-0002"}]}`))
	payInvalidReq.Header.Set("Content-Type", "application/json")
	payInvalidRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payInvalidRR, payInvalidReq)
	if payInvalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected mixed invalid status %d, got %d body=%s", http.StatusBadRequest, payInvalidRR.Code, payInvalidRR.Body.String())
	}
	if !strings.Contains(strings.ToLower(payInvalidRR.Body.String()), "pagos mixtos") {
		t.Fatalf("expected mixed sum validation message, got body=%s", payInvalidRR.Body.String())
	}
}

func TestEmpresaCarritosCompraAnularCierreParcialConAuditoria(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carritos_anular_cierre_parcial.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)

	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":1,"nombre":"Caja Cierre Parcial","canal_venta":"mostrador","moneda":"COP"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	carritoID := int64(createResp["id"].(float64))

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:           1,
		CarritoID:           carritoID,
		TipoItem:            "otro",
		CodigoItem:          "CLOSE-ITEM-" + strconv.FormatInt(carritoID, 10),
		Descripcion:         "Consumo base para cierre parcial",
		UnidadMedida:        "unidad",
		Cantidad:            1,
		PrecioUnitario:      10000,
		DescuentoPorcentaje: 0,
		ImpuestoPorcentaje:  0,
		ImpuestoCodigo:      "IVA",
		UsuarioCreador:      "qa@empresa.com",
		Estado:              "activo",
	}); err != nil {
		t.Fatalf("seed carrito item: %v", err)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"efectivo","total_pagado":10000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected pay status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}

	partialReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=anular_cierre_parcial", strings.NewReader(`{"monto_anulado":2500,"motivo":"ajuste de cierre"}`))
	partialReq.Header.Set("Content-Type", "application/json")
	partialReq.Header.Set("X-Admin-Email", "auditor@empresa.com")
	partialRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(partialRR, partialReq)
	if partialRR.Code != http.StatusOK {
		t.Fatalf("expected partial close cancel status %d, got %d body=%s", http.StatusOK, partialRR.Code, partialRR.Body.String())
	}

	var partialResp map[string]interface{}
	if err := json.Unmarshal(partialRR.Body.Bytes(), &partialResp); err != nil {
		t.Fatalf("decode partial response: %v", err)
	}
	if got, _ := partialResp["total_pagado_nuevo"].(float64); math.Abs(got-7500) > 0.001 {
		t.Fatalf("expected total_pagado_nuevo=7500, got %.4f", got)
	}
	if got, _ := partialResp["devolucion_total"].(float64); math.Abs(got-2500) > 0.001 {
		t.Fatalf("expected devolucion_total=2500, got %.4f", got)
	}

	carritoActualizado, err := dbpkg.GetCarritoCompraByID(dbEmp, 1, carritoID)
	if err != nil {
		t.Fatalf("get carrito actualizado: %v", err)
	}
	if math.Abs(carritoActualizado.TotalPagado-7500) > 0.001 {
		t.Fatalf("expected stored total_pagado=7500, got %.4f", carritoActualizado.TotalPagado)
	}
	if math.Abs(carritoActualizado.DevolucionTotal-2500) > 0.001 {
		t.Fatalf("expected stored devolucion_total=2500, got %.4f", carritoActualizado.DevolucionTotal)
	}

	invalidReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=anular_cierre_parcial", strings.NewReader(`{"monto_anulado":7500}`))
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid partial close status %d, got %d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}

	auditoriaRows, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 1, dbpkg.EmpresaAuditoriaEventoFilter{
		Modulo: "ventas",
		Accion: "anular_cierre_parcial",
		Limit:  20,
	})
	if err != nil {
		t.Fatalf("list auditoria rows: %v", err)
	}
	if len(auditoriaRows) == 0 {
		t.Fatal("expected auditoria event for anular_cierre_parcial")
	}
}

func TestEmpresaUsuarioLoginHandlerBloqueaTrasIntentosFallidos(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_login_lockout.db")
	dbSuper := openTestSQLite(t, "super_login_lockout.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	salt := "salt-lockout"
	hash := hashEmpresaUsuarioPassword("PasswordSegura1", salt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, 1, 'activo')`,
		"lockout@empresa.com", "Usuario Lockout", "vendedor", int64(25), "DOC-LOCK", hash, salt, int64(2),
	)
	if err != nil {
		t.Fatalf("seed user lockout: %v", err)
	}

	h := EmpresaUsuarioLoginHandler(dbEmp, dbSuper)
	for i := 1; i <= 4; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(`{"empresa_id":25,"email":"lockout@empresa.com","password":"ClaveIncorrecta"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d expected status %d, got %d body=%s", i, http.StatusUnauthorized, rr.Code, rr.Body.String())
		}
	}

	lockedReq := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(`{"empresa_id":25,"email":"lockout@empresa.com","password":"ClaveIncorrecta"}`))
	lockedReq.Header.Set("Content-Type", "application/json")
	lockedRR := httptest.NewRecorder()
	h.ServeHTTP(lockedRR, lockedReq)
	if lockedRR.Code != http.StatusTooManyRequests {
		t.Fatalf("expected lockout status %d, got %d body=%s", http.StatusTooManyRequests, lockedRR.Code, lockedRR.Body.String())
	}

	var lockUntil string
	if err := dbEmp.QueryRow("SELECT COALESCE(login_locked_until,'') FROM users WHERE email = ?", "lockout@empresa.com").Scan(&lockUntil); err != nil {
		t.Fatalf("query login_locked_until: %v", err)
	}
	if strings.TrimSpace(lockUntil) == "" {
		t.Fatal("expected login_locked_until to be set")
	}

	blockedCorrectReq := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(`{"empresa_id":25,"email":"lockout@empresa.com","password":"PasswordSegura1"}`))
	blockedCorrectReq.Header.Set("Content-Type", "application/json")
	blockedCorrectRR := httptest.NewRecorder()
	h.ServeHTTP(blockedCorrectRR, blockedCorrectReq)
	if blockedCorrectRR.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d while user is locked, got %d body=%s", http.StatusTooManyRequests, blockedCorrectRR.Code, blockedCorrectRR.Body.String())
	}
}

func TestEmpresaUsuarioPasswordRecoveryFlow(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_password_recovery.db")
	dbSuper := openTestSQLite(t, "super_password_recovery.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	oldSalt := "salt-old"
	oldHash := hashEmpresaUsuarioPassword("ClaveAnterior99", oldSalt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, 1, 'activo')`,
		"recovery@empresa.com", "Usuario Recovery", "vendedor", int64(31), "DOC-REC", oldHash, oldSalt, int64(2),
	)
	if err != nil {
		t.Fatalf("seed user recovery: %v", err)
	}

	recoverH := EmpresaUsuarioRequestPasswordRecoveryHandler(dbEmp, dbSuper)
	recoverReq := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/solicitar_recuperacion_password", strings.NewReader(`{"empresa_id":31,"email":"recovery@empresa.com"}`))
	recoverReq.Header.Set("Content-Type", "application/json")
	recoverRR := httptest.NewRecorder()
	recoverH.ServeHTTP(recoverRR, recoverReq)
	if recoverRR.Code != http.StatusOK {
		t.Fatalf("expected recovery request status %d, got %d body=%s", http.StatusOK, recoverRR.Code, recoverRR.Body.String())
	}

	var resetToken string
	var resetExpira string
	if err := dbEmp.QueryRow("SELECT COALESCE(password_reset_token,''), COALESCE(password_reset_expira,'') FROM users WHERE email = ?", "recovery@empresa.com").Scan(&resetToken, &resetExpira); err != nil {
		t.Fatalf("query password reset fields: %v", err)
	}
	if strings.TrimSpace(resetToken) == "" || strings.TrimSpace(resetExpira) == "" {
		t.Fatalf("expected password reset token and expiration, got token=%q expira=%q", resetToken, resetExpira)
	}

	resetH := EmpresaUsuarioResetPasswordHandler(dbEmp, dbSuper)
	resetBody := `{"empresa_id":31,"email":"recovery@empresa.com","token":"` + resetToken + `","password":"ClaveNueva101","password_confirm":"ClaveNueva101"}`
	resetReq := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/restablecer_password", strings.NewReader(resetBody))
	resetReq.Header.Set("Content-Type", "application/json")
	resetRR := httptest.NewRecorder()
	resetH.ServeHTTP(resetRR, resetReq)
	if resetRR.Code != http.StatusOK {
		t.Fatalf("expected reset status %d, got %d body=%s", http.StatusOK, resetRR.Code, resetRR.Body.String())
	}

	var newHash string
	var tokenAfterReset string
	if err := dbEmp.QueryRow("SELECT COALESCE(password_hash,''), COALESCE(password_reset_token,'') FROM users WHERE email = ?", "recovery@empresa.com").Scan(&newHash, &tokenAfterReset); err != nil {
		t.Fatalf("query user after reset: %v", err)
	}
	if strings.TrimSpace(newHash) == "" || newHash == oldHash {
		t.Fatalf("expected password hash to change after reset, old=%q new=%q", oldHash, newHash)
	}
	if strings.TrimSpace(tokenAfterReset) != "" {
		t.Fatalf("expected password reset token to be cleared after reset, got %q", tokenAfterReset)
	}

	loginH := EmpresaUsuarioLoginHandler(dbEmp, dbSuper)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(`{"empresa_id":31,"email":"recovery@empresa.com","password":"ClaveNueva101"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRR := httptest.NewRecorder()
	loginH.ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusOK {
		t.Fatalf("expected login with new password status %d, got %d body=%s", http.StatusOK, loginRR.Code, loginRR.Body.String())
	}
}

func TestAuthMiddlewareRejectsReusedRevokedSessionToken(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_auth_revoke.db")
	ensureSuperSchema(t, dbSuper)

	if err := dbpkg.UpsertAdministrador(dbSuper, "security@empresa.com", "Security Admin", "administrador", ""); err != nil {
		t.Fatalf("upsert administrador: %v", err)
	}

	token := "token-revoke-001"
	if err := dbpkg.CreateSession(dbSuper, "security@empresa.com", "127.0.0.1", "go-test", token); err != nil {
		t.Fatalf("create session: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/privado", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h := utils.AuthMiddleware(dbSuper, mux)

	firstReq := httptest.NewRequest(http.MethodGet, "/api/privado", nil)
	firstReq.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	firstRR := httptest.NewRecorder()
	h.ServeHTTP(firstRR, firstReq)
	if firstRR.Code != http.StatusNoContent {
		t.Fatalf("expected first access status %d, got %d body=%s", http.StatusNoContent, firstRR.Code, firstRR.Body.String())
	}

	if err := dbpkg.RevokeSessionByToken(dbSuper, token); err != nil {
		t.Fatalf("revoke session token: %v", err)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/api/privado", nil)
	secondReq.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	secondRR := httptest.NewRecorder()
	h.ServeHTTP(secondRR, secondReq)
	if secondRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked token status %d, got %d body=%s", http.StatusUnauthorized, secondRR.Code, secondRR.Body.String())
	}
}

func TestAuthMiddlewareRejectsExpiredSessionToken(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_auth_expired.db")
	ensureSuperSchema(t, dbSuper)

	if err := dbpkg.UpsertAdministrador(dbSuper, "expired@empresa.com", "Expired Admin", "administrador", ""); err != nil {
		t.Fatalf("upsert administrador: %v", err)
	}

	token := "token-expired-001"
	if err := dbpkg.CreateSession(dbSuper, "expired@empresa.com", "127.0.0.1", "go-test", token); err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, err := dbSuper.Exec("UPDATE sesiones SET fecha_fin = datetime('now','-1 minute','localtime') WHERE token = ?", token); err != nil {
		t.Fatalf("expire session token: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/privado", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	h := utils.AuthMiddleware(dbSuper, mux)

	req := httptest.NewRequest(http.MethodGet, "/api/privado", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected expired token status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}
