package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
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
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}
}

func TestHandleGoogleLoginRedirectIncludesLoginHint(t *testing.T) {
	h := HandleGoogleLogin("client-123", "http://localhost:8080/auth/google/callback")
	req := httptest.NewRequest(http.MethodGet, "/auth/google/login?login_hint=usuario@example.com", nil)
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
	if q.Get("prompt") != "" {
		t.Fatalf("prompt must be empty when login_hint is set, got %q", q.Get("prompt"))
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

	if _, err := dbEmp.Exec(`UPDATE carritos_compras SET subtotal = 10000, total = 10000, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = 1 AND id = ?`, carritoID); err != nil {
		t.Fatalf("seed carrito total: %v", err)
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=1&id="+strconv.FormatInt(carritoID, 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"cripto","total_pagado":10000}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)

	if payRR.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid payment method status %d, got %d body=%s", http.StatusBadRequest, payRR.Code, payRR.Body.String())
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
		if _, err := dbEmp.Exec(`UPDATE carritos_compras SET subtotal = 12000, total = 12000, fecha_actualizacion = datetime('now','localtime') WHERE empresa_id = 1 AND id = ?`, carritoID); err != nil {
			t.Fatalf("seed carrito total: %v", err)
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
