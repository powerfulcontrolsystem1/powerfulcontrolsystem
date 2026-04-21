package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func insertEmpresaSoporteRemotoTest(t *testing.T, dbEmp *sql.DB, nombre, nit string) int64 {
	t.Helper()
	res, err := dbEmp.Exec(`INSERT INTO empresas (
		empresa_id,
		nombre,
		nit,
		tipo_id,
		tipo_nombre,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, 'activo', ?)` , 0, nombre, nit, 1, "retail", "qa", "soporte remoto test")
	if err != nil {
		t.Fatalf("insert empresa soporte remoto test: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id empresa soporte remoto test: %v", err)
	}
	if _, err := dbEmp.Exec(`UPDATE empresas SET empresa_id = ? WHERE id = ?`, id, id); err != nil {
		t.Fatalf("update empresa_id soporte remoto test: %v", err)
	}
	return id
}

func ensureEmpresasSchemaSoporteRemotoTest(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	_, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre TEXT,
		nit TEXT,
		tipo_id INTEGER,
		tipo_nombre TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`)
	if err != nil {
		t.Fatalf("create empresas schema soporte remoto test: %v", err)
	}
}

func TestSuperSoporteRemotoHandlerListsCompaniesAndCreatesSession(t *testing.T) {
	dbEmp := openTestSQLite(t, "super_soporte_remoto_empresas.db")
	ensureEmpresasSchemaSoporteRemotoTest(t, dbEmp)
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	empresaID := insertEmpresaSoporteRemotoTest(t, dbEmp, "Empresa Remota", "90015")
	if _, err := dbpkg.UpsertEmpresaSoporteRemotoConfig(dbEmp, dbpkg.EmpresaSoporteRemotoConfig{
		EmpresaID:                  empresaID,
		Habilitado:                 true,
		ProveedorPreferido:         "guacamole",
		ModoOperacion:              "agente_web",
		RequiereAprobacionOperador: false,
		AutoCerrarMinutos:          25,
		MaxConexionesMes:           5,
		MaxMinutosMes:              120,
		MaxDispositivos:            3,
		UsuarioCreador:             "qa",
		Estado:                     "activo",
	}); err != nil {
		t.Fatalf("UpsertEmpresaSoporteRemotoConfig: %v", err)
	}
	deviceID, err := dbpkg.CreateEmpresaSoporteRemotoDispositivo(dbEmp, dbpkg.EmpresaSoporteRemotoDispositivo{
		EmpresaID:      empresaID,
		NombreEquipo:   "Caja soporte",
		StreamURL:      "https://remote.example/super-support",
		UsuarioCreador: "qa",
		Estado:         "activo",
	}, "9015")
	if err != nil {
		t.Fatalf("CreateEmpresaSoporteRemotoDispositivo: %v", err)
	}

	h := SuperSoporteRemotoHandler(dbEmp)

	reqCompanies := httptest.NewRequest(http.MethodGet, "/super/api/soporte_remoto?action=empresas&q=remota", nil)
	rrCompanies := httptest.NewRecorder()
	h.ServeHTTP(rrCompanies, reqCompanies)
	if rrCompanies.Code != http.StatusOK {
		t.Fatalf("expected 200 companies, got %d body=%s", rrCompanies.Code, rrCompanies.Body.String())
	}
	bodyCompanies := decodeSoporteRemotoBody(t, rrCompanies)
	if numberValue(bodyCompanies["total"]) < 1 {
		t.Fatalf("expected at least one company in response, got %#v", bodyCompanies)
	}

	reqSession := httptest.NewRequest(http.MethodPost, "/super/api/soporte_remoto?action=solicitar_sesion", strings.NewReader(`{"empresa_id":`+itoa64(empresaID)+`,"dispositivo_id":`+itoa64(deviceID)+`,"motivo":"Soporte central","duracion_min":15}`))
	reqSession.Header.Set("Content-Type", "application/json")
	reqSession.Header.Set("X-Admin-Email", "mesa.central@test.local")
	rrSession := httptest.NewRecorder()
	h.ServeHTTP(rrSession, reqSession)
	if rrSession.Code != http.StatusCreated {
		t.Fatalf("expected 201 session from super, got %d body=%s", rrSession.Code, rrSession.Body.String())
	}
	bodySession := decodeSoporteRemotoBody(t, rrSession)
	if strings.TrimSpace(stringValue(bodySession["viewer_url"])) == "" {
		t.Fatalf("expected viewer_url in super session response, got %#v", bodySession)
	}

	reqSessions := httptest.NewRequest(http.MethodGet, "/super/api/soporte_remoto?action=sesiones&empresa_id="+itoa64(empresaID), nil)
	rrSessions := httptest.NewRecorder()
	h.ServeHTTP(rrSessions, reqSessions)
	if rrSessions.Code != http.StatusOK {
		t.Fatalf("expected 200 sessions from super, got %d body=%s", rrSessions.Code, rrSessions.Body.String())
	}
	bodySessions := decodeSoporteRemotoBody(t, rrSessions)
	if numberValue(bodySessions["total"]) < 1 {
		t.Fatalf("expected session rows in super response, got %#v", bodySessions)
	}
}

func TestSuperSoporteRemotoHandlerConfigGetAndUpdate(t *testing.T) {
	dbEmp := openTestSQLite(t, "super_soporte_remoto_config.db")
	ensureEmpresasSchemaSoporteRemotoTest(t, dbEmp)
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	empresaID := insertEmpresaSoporteRemotoTest(t, dbEmp, "Empresa Config Soporte", "90117")

	h := SuperSoporteRemotoHandler(dbEmp)

	reqDefault := httptest.NewRequest(http.MethodGet, "/super/api/soporte_remoto?action=config&empresa_id="+itoa64(empresaID), nil)
	rrDefault := httptest.NewRecorder()
	h.ServeHTTP(rrDefault, reqDefault)
	if rrDefault.Code != http.StatusOK {
		t.Fatalf("expected 200 get default config, got %d body=%s", rrDefault.Code, rrDefault.Body.String())
	}
	bodyDefault := decodeSoporteRemotoBody(t, rrDefault)
	cfgDefault, ok := bodyDefault["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected default config object, got %#v", bodyDefault)
	}
	if !boolValue(cfgDefault["habilitado"]) {
		t.Fatalf("expected default habilitado=true, got %#v", cfgDefault)
	}
	if !boolValue(cfgDefault["portal_publico_habilitado"]) {
		t.Fatalf("expected default portal_publico_habilitado=true, got %#v", cfgDefault)
	}
	if strings.TrimSpace(stringValue(cfgDefault["cliente_windows_url"])) != "/descargas/rustdesk-cliente-windows-x64.exe" {
		t.Fatalf("expected default cliente_windows_url local download, got %#v", cfgDefault)
	}
	if strings.TrimSpace(stringValue(cfgDefault["servidor_linux_url"])) != "/descargas/rustdesk-servidor-linux-amd64.zip" {
		t.Fatalf("expected default servidor_linux_url local download, got %#v", cfgDefault)
	}

	reqSave := httptest.NewRequest(http.MethodPost, "/super/api/soporte_remoto?action=config&empresa_id="+itoa64(empresaID), strings.NewReader(`{"habilitado":true,"portal_publico_habilitado":true,"proveedor_preferido":"rustdesk_oss","modo_operacion":"cliente_local","max_minutos_dia_rustdesk":240,"rustdesk_server_host":"rustdesk.powerfulcontrolsystem.com:21116","rustdesk_server_key":"PUBKEY-90117","cliente_windows_url":"https://downloads.example/client-win.exe","cliente_linux_url":"https://downloads.example/client-linux.deb","cliente_mac_url":"https://downloads.example/client-mac.dmg","servidor_windows_url":"https://downloads.example/server-win.zip","servidor_linux_url":"https://downloads.example/server-linux.tar.gz","servidor_mac_url":"https://downloads.example/server-mac.pkg","carpeta_transferencia":"/srv/rustdesk/transfer","instrucciones_publicas":"Descarga el cliente y comparte el ID de RustDesk con soporte."}`))
	reqSave.Header.Set("Content-Type", "application/json")
	reqSave.Header.Set("X-Admin-Email", "super@test.local")
	rrSave := httptest.NewRecorder()
	h.ServeHTTP(rrSave, reqSave)
	if rrSave.Code != http.StatusOK {
		t.Fatalf("expected 200 save config, got %d body=%s", rrSave.Code, rrSave.Body.String())
	}
	bodySave := decodeSoporteRemotoBody(t, rrSave)
	cfgSave, ok := bodySave["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected config object, got %#v", bodySave)
	}
	if strings.TrimSpace(stringValue(cfgSave["servidor_windows_url"])) == "" {
		t.Fatalf("expected servidor_windows_url saved, got %#v", cfgSave)
	}
	if strings.TrimSpace(stringValue(cfgSave["cliente_mac_url"])) == "" || strings.TrimSpace(stringValue(cfgSave["servidor_mac_url"])) == "" {
		t.Fatalf("expected mac urls saved, got %#v", cfgSave)
	}
	if numberValue(cfgSave["max_minutos_dia_rustdesk"]) != 240 {
		t.Fatalf("expected max_minutos_dia_rustdesk saved, got %#v", cfgSave)
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/super/api/soporte_remoto?action=config&empresa_id="+itoa64(empresaID), nil)
	rrGet := httptest.NewRecorder()
	h.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("expected 200 get config, got %d body=%s", rrGet.Code, rrGet.Body.String())
	}
	bodyGet := decodeSoporteRemotoBody(t, rrGet)
	cfgGet, ok := bodyGet["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected config object in get, got %#v", bodyGet)
	}
	if strings.TrimSpace(stringValue(cfgGet["rustdesk_server_host"])) != "rustdesk.powerfulcontrolsystem.com:21116" {
		t.Fatalf("expected rustdesk_server_host persisted, got %#v", cfgGet)
	}
	if strings.TrimSpace(stringValue(cfgGet["cliente_mac_url"])) != "https://downloads.example/client-mac.dmg" {
		t.Fatalf("expected cliente_mac_url persisted, got %#v", cfgGet)
	}
	if numberValue(cfgGet["max_minutos_dia_rustdesk"]) != 240 {
		t.Fatalf("expected max_minutos_dia_rustdesk persisted, got %#v", cfgGet)
	}
}