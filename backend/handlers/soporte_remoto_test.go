package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func decodeSoporteRemotoBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	out := map[string]interface{}{}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode body json: %v. body=%s", err, rr.Body.String())
	}
	return out
}

func TestEmpresaSoporteRemotoHandlerFlow(t *testing.T) {
	dbEmp := openPermsTestDB(t, "soporte_remoto_handler.db")
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	h := EmpresaSoporteRemotoHandler(dbEmp)

	reqCfg := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=config&empresa_id=901", strings.NewReader(`{"habilitado":true,"proveedor_preferido":"guacamole","modo_operacion":"agente_web","requiere_aprobacion_operador":false,"auto_cerrar_minutos":35}`))
	reqCfg.Header.Set("Content-Type", "application/json")
	reqCfg.Header.Set("X-Admin-Email", "admin901@test.local")
	rrCfg := httptest.NewRecorder()
	h.ServeHTTP(rrCfg, reqCfg)
	if rrCfg.Code != http.StatusOK {
		t.Fatalf("expected 200 config, got %d body=%s", rrCfg.Code, rrCfg.Body.String())
	}
	cfgResp := decodeSoporteRemotoBody(t, rrCfg)
	cfgRaw, ok := cfgResp["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected config object in response: %#v", cfgResp)
	}
	if provider := strings.TrimSpace(strings.ToLower(stringValue(cfgRaw["proveedor_preferido"]))); provider != "guacamole" {
		t.Fatalf("expected proveedor_preferido guacamole, got %q", provider)
	}

	reqDevice := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=crear_dispositivo&empresa_id=901", strings.NewReader(`{"nombre_equipo":"Caja 901","alias_operativo":"Punto principal","stream_url":"https://remote.example/901","acceso_pin":"9010"}`))
	reqDevice.Header.Set("Content-Type", "application/json")
	reqDevice.Header.Set("X-Admin-Email", "admin901@test.local")
	rrDevice := httptest.NewRecorder()
	h.ServeHTTP(rrDevice, reqDevice)
	if rrDevice.Code != http.StatusCreated {
		t.Fatalf("expected 201 create device, got %d body=%s", rrDevice.Code, rrDevice.Body.String())
	}
	deviceResp := decodeSoporteRemotoBody(t, rrDevice)
	deviceRaw, ok := deviceResp["dispositivo"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected dispositivo object in response: %#v", deviceResp)
	}
	deviceID := int64(numberValue(deviceRaw["id"]))
	if deviceID <= 0 {
		t.Fatalf("expected device id > 0, got %d", deviceID)
	}

	reqSession := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=solicitar_sesion&empresa_id=901", strings.NewReader(`{"dispositivo_id":`+itoa64(deviceID)+`,"motivo":"soporte rapido","duracion_min":20}`))
	reqSession.Header.Set("Content-Type", "application/json")
	reqSession.Header.Set("X-Admin-Email", "admin901@test.local")
	rrSession := httptest.NewRecorder()
	h.ServeHTTP(rrSession, reqSession)
	if rrSession.Code != http.StatusCreated {
		t.Fatalf("expected 201 create session, got %d body=%s", rrSession.Code, rrSession.Body.String())
	}
	sessionResp := decodeSoporteRemotoBody(t, rrSession)
	sessionRaw, ok := sessionResp["session"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected session object in response: %#v", sessionResp)
	}
	codigoSesion := stringValue(sessionRaw["codigo_sesion"])
	token := stringValue(sessionRaw["token_visualizacion"])
	if strings.TrimSpace(codigoSesion) == "" || strings.TrimSpace(token) == "" {
		t.Fatalf("expected codigo_sesion and token_visualizacion, got session=%#v", sessionRaw)
	}

	reqResolve := httptest.NewRequest(http.MethodGet, "/api/empresa/soporte_remoto?action=resolver_visualizacion&empresa_id=901&codigo_sesion="+codigoSesion+"&token="+token, nil)
	rrResolve := httptest.NewRecorder()
	h.ServeHTTP(rrResolve, reqResolve)
	if rrResolve.Code != http.StatusOK {
		t.Fatalf("expected 200 resolve, got %d body=%s", rrResolve.Code, rrResolve.Body.String())
	}
	resolveResp := decodeSoporteRemotoBody(t, rrResolve)
	if !boolValue(resolveResp["acceso_permitido"]) {
		t.Fatalf("expected acceso_permitido=true, got response=%#v", resolveResp)
	}
	if strings.TrimSpace(stringValue(resolveResp["embed_url"])) == "" {
		t.Fatalf("expected non-empty embed_url, got response=%#v", resolveResp)
	}

	reqExport := httptest.NewRequest(http.MethodGet, "/api/empresa/soporte_remoto?action=export_sesiones&empresa_id=901&format=csv", nil)
	rrExport := httptest.NewRecorder()
	h.ServeHTTP(rrExport, reqExport)
	if rrExport.Code != http.StatusOK {
		t.Fatalf("expected 200 export, got %d body=%s", rrExport.Code, rrExport.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrExport.Header().Get("Content-Type")), "csv") {
		t.Fatalf("expected csv content type, got %q", rrExport.Header().Get("Content-Type"))
	}
}

func TestPublicSoporteRemotoAgentHeartbeatAndStateUpdate(t *testing.T) {
	dbEmp := openPermsTestDB(t, "soporte_remoto_public_agent.db")
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	deviceID, err := dbpkg.CreateEmpresaSoporteRemotoDispositivo(dbEmp, dbpkg.EmpresaSoporteRemotoDispositivo{
		EmpresaID:         777,
		CodigoDispositivo: "DEV-777-A",
		NombreEquipo:      "Recepcion 777",
		StreamURL:         "https://remote.example/777",
		UsuarioCreador:    "admin777@test.local",
		Estado:            "activo",
	}, "7770")
	if err != nil {
		t.Fatalf("create device: %v", err)
	}

	session, err := dbpkg.CreateEmpresaSoporteRemotoSession(dbEmp, 777, deviceID, "admin777@test.local", "Operador 777", "op777@test.local", "aprobacion por agente", 30, true)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	publicHandler := PublicEmpresaSoporteRemotoAgentHandler(dbEmp)

	reqHeartbeat := httptest.NewRequest(http.MethodPost, "/api/public/soporte_remoto?action=heartbeat", strings.NewReader(`{"empresa_id":777,"codigo_dispositivo":"DEV-777-A","acceso_pin":"7770","sistema_operativo":"Windows 11","agente_version":"1.2.0"}`))
	reqHeartbeat.Header.Set("Content-Type", "application/json")
	rrHeartbeat := httptest.NewRecorder()
	publicHandler.ServeHTTP(rrHeartbeat, reqHeartbeat)
	if rrHeartbeat.Code != http.StatusOK {
		t.Fatalf("expected 200 heartbeat, got %d body=%s", rrHeartbeat.Code, rrHeartbeat.Body.String())
	}

	reqApprove := httptest.NewRequest(http.MethodPost, "/api/public/soporte_remoto?action=aprobar_sesion&codigo_sesion="+session.CodigoSesion, strings.NewReader(`{"empresa_id":777,"codigo_dispositivo":"DEV-777-A","acceso_pin":"7770"}`))
	reqApprove.Header.Set("Content-Type", "application/json")
	rrApprove := httptest.NewRecorder()
	publicHandler.ServeHTTP(rrApprove, reqApprove)
	if rrApprove.Code != http.StatusOK {
		t.Fatalf("expected 200 approve, got %d body=%s", rrApprove.Code, rrApprove.Body.String())
	}

	updated, err := dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, 777, session.CodigoSesion)
	if err != nil {
		t.Fatalf("GetEmpresaSoporteRemotoSessionByCodigo after approve: %v", err)
	}
	if updated.EstadoSesion != "aprobada" {
		t.Fatalf("expected aprobada after agent update, got %q", updated.EstadoSesion)
	}

	reqFinish := httptest.NewRequest(http.MethodPost, "/api/public/soporte_remoto?action=finalizar_sesion&codigo_sesion="+session.CodigoSesion, strings.NewReader(`{"empresa_id":777,"codigo_dispositivo":"DEV-777-A","acceso_pin":"7770"}`))
	reqFinish.Header.Set("Content-Type", "application/json")
	rrFinish := httptest.NewRecorder()
	publicHandler.ServeHTTP(rrFinish, reqFinish)
	if rrFinish.Code != http.StatusOK {
		t.Fatalf("expected 200 finish, got %d body=%s", rrFinish.Code, rrFinish.Body.String())
	}

	finished, err := dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp, 777, session.CodigoSesion)
	if err != nil {
		t.Fatalf("GetEmpresaSoporteRemotoSessionByCodigo after finish: %v", err)
	}
	if finished.EstadoSesion != "finalizada" {
		t.Fatalf("expected finalizada after agent finish, got %q", finished.EstadoSesion)
	}
}

func stringValue(v interface{}) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func numberValue(v interface{}) float64 {
	n, _ := v.(float64)
	return n
}

func boolValue(v interface{}) bool {
	b, _ := v.(bool)
	return b
}
