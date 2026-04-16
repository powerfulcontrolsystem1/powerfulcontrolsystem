package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
	_ "modernc.org/sqlite"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

// TestE2E_AcceptContractCreatesSession valida el flujo completo:
// 1) callback OAuth de admin nuevo redirige a /accept.html con payload cifrado.
// 2) /accept/complete persiste acepta_contrato=1 y crea sesión.
// 3) siguiente callback del mismo admin entra directo sin pedir contrato.
func TestE2E_AcceptContractCreatesSession(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")

	// Mock HTTP client para endpoints de Google (token + userinfo)
	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		url := req.URL.String()
		if strings.Contains(url, "oauth2.googleapis.com/token") {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"token-abc"}`))}, nil
		}
		if strings.Contains(url, "www.googleapis.com/oauth2/v3/userinfo") {
			body := `{"sub":"u1","name":"Test User","email":"test@example.com","email_verified":true,"picture":"https://example.com/p.png"}`
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
		}
		return &http.Response{StatusCode: 404, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("not found"))}, nil
	})}
	defer func() { http.DefaultClient = original }()

	// DB en memoria para super y empresas
	dbSuper, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open super db: %v", err)
	}
	defer dbSuper.Close()
	dbEmpresas, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open empresas db: %v", err)
	}
	defer dbEmpresas.Close()

	// Crear tablas mínimas necesarias
	if _, err := dbSuper.Exec(`CREATE TABLE administradores (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE,
        name TEXT,
        role TEXT DEFAULT 'administrador',
        photo TEXT,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
        fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
        usuario_creador TEXT,
        estado TEXT DEFAULT 'activo',
        observaciones TEXT,
        acepta_contrato INTEGER DEFAULT 0
    );`); err != nil {
		t.Fatalf("create administradores: %v", err)
	}
	if _, err := dbSuper.Exec(`CREATE TABLE sesiones (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        admin_email TEXT,
        token TEXT,
        ip TEXT,
        user_agent TEXT,
        fecha_inicio TEXT DEFAULT (datetime('now','localtime')),
        fecha_fin TEXT,
        activo INTEGER DEFAULT 1,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime'))
    );`); err != nil {
		t.Fatalf("create sesiones: %v", err)
	}

	if _, err := dbEmpresas.Exec(`CREATE TABLE users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE,
        name TEXT,
        role TEXT DEFAULT 'administrador',
        empresa_id INTEGER,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
        fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
        usuario_creador TEXT,
        estado TEXT DEFAULT 'activo',
        observaciones TEXT
    );`); err != nil {
		t.Fatalf("create users: %v", err)
	}
	if _, err := dbEmpresas.Exec(`CREATE TABLE empresas (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        empresa_id INTEGER,
        nombre TEXT NOT NULL,
        nit TEXT,
        tipo_id INTEGER,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
        fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
        usuario_creador TEXT,
        estado TEXT DEFAULT 'activo',
        observaciones TEXT
    );`); err != nil {
		t.Fatalf("create empresas: %v", err)
	}

	callbackHandler := HandleGoogleCallback(dbEmpresas, dbSuper, "client-id", "client-secret", "http://localhost/callback")
	acceptHandler := AcceptCompleteHandler(dbSuper)

	// 1) Primer callback sin aceptación previa -> debe redirigir a accept.html con payload.
	firstReq := httptest.NewRequest("GET", "/?code=code123", nil)
	firstReq.RemoteAddr = "127.0.0.1:12345"
	firstRec := httptest.NewRecorder()
	callbackHandler.ServeHTTP(firstRec, firstReq)
	firstResp := firstRec.Result()
	defer firstResp.Body.Close()

	if firstResp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(firstResp.Body)
		t.Fatalf("expected redirect to accept, got status=%d body=%s", firstResp.StatusCode, string(body))
	}

	firstLoc := firstResp.Header.Get("Location")
	if !strings.HasPrefix(firstLoc, "/accept.html?") {
		t.Fatalf("expected redirect to /accept.html, got %s", firstLoc)
	}
	locURL, err := url.Parse(firstLoc)
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	payload := strings.TrimSpace(locURL.Query().Get("payload"))
	if payload == "" {
		t.Fatalf("expected encrypted payload in redirect")
	}

	// 2) Completar aceptación (bypass de desarrollo solo para test) -> crea sesión y marca contrato.
	t.Setenv("RECAPTCHA_DEV_BYPASS", "1")
	acceptReq := httptest.NewRequest("POST", "/accept/complete", strings.NewReader(`{"payload":"`+payload+`","token":""}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	acceptReq.RemoteAddr = "127.0.0.1:12345"
	acceptRec := httptest.NewRecorder()
	acceptHandler.ServeHTTP(acceptRec, acceptReq)
	acceptResp := acceptRec.Result()
	defer acceptResp.Body.Close()

	if acceptResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(acceptResp.Body)
		t.Fatalf("expected 200 from accept/complete, got status=%d body=%s", acceptResp.StatusCode, string(body))
	}

	var acceptJSON map[string]interface{}
	if err := json.NewDecoder(acceptResp.Body).Decode(&acceptJSON); err != nil {
		t.Fatalf("decode accept response: %v", err)
	}
	if got, _ := acceptJSON["redirect"].(string); got != "/seleccionar_empresa.html" {
		t.Fatalf("unexpected accept redirect: %q", got)
	}

	var token string
	hasBrowserSessionCookie := false
	for _, c := range acceptResp.Cookies() {
		if c.Name == "session_token" && strings.TrimSpace(c.Value) != "" {
			token = c.Value
		}
		if c.Name == browserSessionStateCookieName && strings.TrimSpace(c.Value) == "1" {
			hasBrowserSessionCookie = true
		}
	}
	if token == "" {
		t.Fatalf("session_token cookie not set after acceptance")
	}
	if !hasBrowserSessionCookie {
		t.Fatalf("browser session cookie not set after acceptance")
	}

	sess, err := dbpkg.GetSessionByToken(dbSuper, token)
	if err != nil {
		t.Fatalf("GetSessionByToken error: %v", err)
	}
	if sess == nil || sess.AdminEmail != "test@example.com" {
		t.Fatalf("unexpected session: %+v", sess)
	}

	admin, err := dbpkg.GetAdminByEmail(dbSuper, "test@example.com")
	if err != nil {
		t.Fatalf("GetAdminByEmail error: %v", err)
	}
	if admin == nil {
		t.Fatalf("expected admin record")
	}
	if admin.AceptaContrato != 1 {
		t.Fatalf("expected acepta_contrato=1, got %d", admin.AceptaContrato)
	}
	acceptance, err := dbpkg.GetAdministradorContratoAceptacion(dbSuper, "test@example.com")
	if err != nil {
		t.Fatalf("GetAdministradorContratoAceptacion error: %v", err)
	}
	if !acceptance.Acepta || acceptance.Version != 1 {
		t.Fatalf("expected accepted version 1, got acepta=%v version=%d", acceptance.Acepta, acceptance.Version)
	}

	// 3) Segundo callback con el mismo admin -> debe entrar directo sin pedir contrato.
	secondReq := httptest.NewRequest("GET", "/?code=code456", nil)
	secondReq.RemoteAddr = "127.0.0.1:12345"
	secondRec := httptest.NewRecorder()
	callbackHandler.ServeHTTP(secondRec, secondReq)
	secondResp := secondRec.Result()
	defer secondResp.Body.Close()

	if secondResp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(secondResp.Body)
		t.Fatalf("expected redirect on second callback, got status=%d body=%s", secondResp.StatusCode, string(body))
	}
	if got := secondResp.Header.Get("Location"); got != "/seleccionar_empresa.html" {
		t.Fatalf("expected direct redirect to seleccionar_empresa, got %s", got)
	}

	hasSecondSessionCookie := false
	hasSecondBrowserSessionCookie := false
	for _, c := range secondResp.Cookies() {
		if c.Name == "session_token" && strings.TrimSpace(c.Value) != "" {
			hasSecondSessionCookie = true
		}
		if c.Name == browserSessionStateCookieName && strings.TrimSpace(c.Value) == "1" {
			hasSecondBrowserSessionCookie = true
		}
	}
	if !hasSecondSessionCookie {
		t.Fatalf("expected session_token cookie on second callback")
	}
	if !hasSecondBrowserSessionCookie {
		t.Fatalf("expected browser session cookie on second callback")
	}
}

func TestE2E_AcceptContractRequiresNewVersion(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=")

	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		url := req.URL.String()
		if strings.Contains(url, "oauth2.googleapis.com/token") {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"access_token":"token-abc"}`))}, nil
		}
		if strings.Contains(url, "www.googleapis.com/oauth2/v3/userinfo") {
			body := `{"sub":"u2","name":"Versioned User","email":"versioned@example.com","email_verified":true,"picture":"https://example.com/p.png"}`
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
		}
		return &http.Response{StatusCode: 404, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("not found"))}, nil
	})}
	defer func() { http.DefaultClient = original }()

	dbSuper, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open super db: %v", err)
	}
	defer dbSuper.Close()
	dbEmpresas, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open empresas db: %v", err)
	}
	defer dbEmpresas.Close()

	if _, err := dbSuper.Exec(`CREATE TABLE administradores (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE,
        name TEXT,
        role TEXT DEFAULT 'administrador',
        photo TEXT,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
        fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
        usuario_creador TEXT,
        estado TEXT DEFAULT 'activo',
        observaciones TEXT,
        acepta_contrato INTEGER DEFAULT 0
    );`); err != nil {
		t.Fatalf("create administradores: %v", err)
	}
	if _, err := dbSuper.Exec(`CREATE TABLE sesiones (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        admin_email TEXT,
        token TEXT,
        ip TEXT,
        user_agent TEXT,
        fecha_inicio TEXT DEFAULT (datetime('now','localtime')),
        fecha_fin TEXT,
        activo INTEGER DEFAULT 1,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime'))
    );`); err != nil {
		t.Fatalf("create sesiones: %v", err)
	}

	if _, err := dbEmpresas.Exec(`CREATE TABLE users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE,
        name TEXT,
        role TEXT DEFAULT 'administrador',
        empresa_id INTEGER,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
        fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
        usuario_creador TEXT,
        estado TEXT DEFAULT 'activo',
        observaciones TEXT
    );`); err != nil {
		t.Fatalf("create users: %v", err)
	}
	if _, err := dbEmpresas.Exec(`CREATE TABLE empresas (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        empresa_id INTEGER,
        nombre TEXT NOT NULL,
        nit TEXT,
        tipo_id INTEGER,
        fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
        fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
        usuario_creador TEXT,
        estado TEXT DEFAULT 'activo',
        observaciones TEXT
    );`); err != nil {
		t.Fatalf("create empresas: %v", err)
	}

	callbackHandler := HandleGoogleCallback(dbEmpresas, dbSuper, "client-id", "client-secret", "http://localhost/callback")
	acceptHandler := AcceptCompleteHandler(dbSuper)

	firstReq := httptest.NewRequest("GET", "/?code=code123", nil)
	firstReq.RemoteAddr = "127.0.0.1:12345"
	firstRec := httptest.NewRecorder()
	callbackHandler.ServeHTTP(firstRec, firstReq)
	firstResp := firstRec.Result()
	defer firstResp.Body.Close()
	if firstResp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(firstResp.Body)
		t.Fatalf("expected redirect to accept, got status=%d body=%s", firstResp.StatusCode, string(body))
	}
	locURL, err := url.Parse(firstResp.Header.Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	payload := strings.TrimSpace(locURL.Query().Get("payload"))
	if payload == "" {
		t.Fatal("expected accept payload")
	}

	acceptReq := httptest.NewRequest("POST", "/accept/complete", strings.NewReader(`{"payload":"`+payload+`"}`))
	acceptReq.Header.Set("Content-Type", "application/json")
	acceptReq.RemoteAddr = "127.0.0.1:12345"
	acceptRec := httptest.NewRecorder()
	acceptHandler.ServeHTTP(acceptRec, acceptReq)
	if acceptRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from accept/complete, got %d body=%s", acceptRec.Code, acceptRec.Body.String())
	}

	_, noChanges, err := dbpkg.SaveSuperContractVersion(dbSuper, dbpkg.SuperContractVersion{
		Titulo:         "Contrato PCS v2",
		Resumen:        "Resumen actualizado para exigir nueva aceptacion.",
		Contenido:      "1. Objeto\nContrato actualizado para exigir nueva version.",
		NotaAceptacion: "Acepto la nueva version vigente del contrato.",
		ResumenCambio:  "Actualizacion material del contrato.",
		UsuarioCreador: "legal@pcs.com",
	}, )
	if err != nil {
		t.Fatalf("save contract version 2: %v", err)
	}
	if noChanges {
		t.Fatal("expected a new contract version to be created")
	}

	secondReq := httptest.NewRequest("GET", "/?code=code456", nil)
	secondReq.RemoteAddr = "127.0.0.1:12345"
	secondRec := httptest.NewRecorder()
	callbackHandler.ServeHTTP(secondRec, secondReq)
	secondResp := secondRec.Result()
	defer secondResp.Body.Close()

	if secondResp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(secondResp.Body)
		t.Fatalf("expected redirect to accept after new version, got status=%d body=%s", secondResp.StatusCode, string(body))
	}
	if got := secondResp.Header.Get("Location"); !strings.HasPrefix(got, "/accept.html?") {
		t.Fatalf("expected redirect back to /accept.html after new contract version, got %s", got)
	}
}
