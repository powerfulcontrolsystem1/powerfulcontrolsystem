package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestSuperSoporteRemotoHandlerListsCompaniesAndCreatesSession(t *testing.T) {
	dbEmp := openTestSQLite(t, "super_soporte_remoto_empresas.db")
	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	empresaID, err := dbpkg.CreateEmpresa(dbEmp, 1, "retail", "Empresa Remota", "90015", "test soporte remoto", "qa")
	if err != nil {
		t.Fatalf("CreateEmpresa: %v", err)
	}
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
	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	empresaID, err := dbpkg.CreateEmpresa(dbEmp, 1, "retail", "Empresa Config Soporte", "90117", "config soporte", "qa")
	if err != nil {
		t.Fatalf("CreateEmpresa: %v", err)
	}

	h := SuperSoporteRemotoHandler(dbEmp)

	reqSave := httptest.NewRequest(http.MethodPost, "/super/api/soporte_remoto?action=config&empresa_id="+itoa64(empresaID), strings.NewReader(`{"habilitado":true,"portal_publico_habilitado":true,"proveedor_preferido":"rustdesk_oss","modo_operacion":"cliente_local","max_minutos_dia_rustdesk":240,"rustdesk_server_host":"rustdesk.powerfulcontrolsystem.com:21116","rustdesk_server_key":"PUBKEY-90117","cliente_windows_url":"https://downloads.example/client-win.exe","cliente_linux_url":"https://downloads.example/client-linux.deb","servidor_windows_url":"https://downloads.example/server-win.zip","servidor_linux_url":"https://downloads.example/server-linux.tar.gz","carpeta_transferencia":"/srv/rustdesk/transfer","instrucciones_publicas":"Descarga el cliente y comparte el ID de RustDesk con soporte."}`))
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
	if numberValue(cfgGet["max_minutos_dia_rustdesk"]) != 240 {
		t.Fatalf("expected max_minutos_dia_rustdesk persisted, got %#v", cfgGet)
	}
}