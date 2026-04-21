package handlers

import (
	"encoding/base64"
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

func TestPublicSoporteRemotoResolverAccesoExponeDescargasRustDesk(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))

	dbEmp := openPermsTestDB(t, "soporte_remoto_public_access.db")
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	if _, err := dbpkg.UpsertEmpresaSoporteRemotoConfig(dbEmp, dbpkg.EmpresaSoporteRemotoConfig{
		EmpresaID:               778,
		Habilitado:              true,
		ProveedorPreferido:      "rustdesk_oss",
		ModoOperacion:           "cliente_local",
		PortalPublicoHabilitado: true,
		RustDeskServerHost:      "rustdesk.powerfulcontrolsystem.com:21116",
		RustDeskServerKey:       "PUB-KEY-778",
		ClienteWindowsURL:       "https://downloads.example/rustdesk-client-win.exe",
		ClienteLinuxURL:         "https://downloads.example/rustdesk-client-linux.deb",
		ClienteMacURL:           "https://downloads.example/rustdesk-client-mac.dmg",
		ServidorWindowsURL:      "https://downloads.example/rustdesk-server-win.zip",
		ServidorLinuxURL:        "https://downloads.example/rustdesk-server-linux.tar.gz",
		ServidorMacURL:          "https://downloads.example/rustdesk-server-mac.pkg",
		CarpetaTransferencia:    "/transferencias/empresa-778",
		InstruccionesPublicas:   "Descarga el cliente, agrega el host y comparte el ID visible con soporte.",
		UsuarioCreador:          "admin778@test.local",
		Estado:                  "activo",
	}); err != nil {
		t.Fatalf("UpsertEmpresaSoporteRemotoConfig: %v", err)
	}

	deviceID, err := dbpkg.CreateEmpresaSoporteRemotoDispositivo(dbEmp, dbpkg.EmpresaSoporteRemotoDispositivo{
		EmpresaID:               778,
		CodigoDispositivo:       "DEV-778-A",
		NombreEquipo:            "Recepcion 778",
		RustDeskDeviceID:        "778-DEVICE",
		RustDeskPasswordEnc:     "clave-778",
		CarpetaTransferencia:    "/transferencias/recepcion-778",
		AccesoPublicoHabilitado: true,
		UsuarioCreador:          "admin778@test.local",
		Estado:                  "activo",
	}, "7780")
	if err != nil {
		t.Fatalf("CreateEmpresaSoporteRemotoDispositivo: %v", err)
	}

	session, err := dbpkg.CreateEmpresaSoporteRemotoSession(dbEmp, 778, deviceID, "admin778@test.local", "Operador 778", "op778@test.local", "acceso publico", 30, false)
	if err != nil {
		t.Fatalf("CreateEmpresaSoporteRemotoSession: %v", err)
	}

	publicHandler := PublicEmpresaSoporteRemotoAgentHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/public/soporte_remoto?action=resolver_acceso_publico&empresa_id=778&codigo_sesion="+session.CodigoSesion+"&token="+session.TokenVisualizacionRaw, nil)
	rr := httptest.NewRecorder()
	publicHandler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 public resolve, got %d body=%s", rr.Code, rr.Body.String())
	}
	body := decodeSoporteRemotoBody(t, rr)
	if !boolValue(body["acceso_permitido"]) {
		t.Fatalf("expected acceso_permitido=true, got %#v", body)
	}
	access, ok := body["access"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected access object, got %#v", body)
	}
	if strings.TrimSpace(stringValue(access["cliente_windows_url"])) == "" || strings.TrimSpace(stringValue(access["servidor_linux_url"])) == "" {
		t.Fatalf("expected download urls in access bundle, got %#v", access)
	}
	if strings.TrimSpace(stringValue(access["cliente_mac_url"])) != "https://downloads.example/rustdesk-client-mac.dmg" {
		t.Fatalf("expected cliente_mac_url in access bundle, got %#v", access)
	}
	if strings.TrimSpace(stringValue(access["servidor_mac_url"])) != "https://downloads.example/rustdesk-server-mac.pkg" {
		t.Fatalf("expected servidor_mac_url in access bundle, got %#v", access)
	}
	if strings.TrimSpace(stringValue(access["rustdesk_device_id"])) != "778-DEVICE" {
		t.Fatalf("expected rustdesk_device_id, got %#v", access)
	}
	if strings.TrimSpace(stringValue(access["portal_publico_url"])) == "" {
		t.Fatalf("expected portal_publico_url, got %#v", access)
	}
}

func TestSuperServidoresProbeHandlerReturnsRustDeskStatus(t *testing.T) {
	h := SuperServidoresProbeHandler()
	req := httptest.NewRequest(http.MethodGet, "/super/api/servidores/probar?id=rustdesk", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 probe response, got %d body=%s", rr.Code, rr.Body.String())
	}
	body := decodeSoporteRemotoBody(t, rr)
	service, ok := body["servicio"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected servicio object, got %#v", body)
	}
	if strings.TrimSpace(stringValue(service["id"])) != "rustdesk" {
		t.Fatalf("expected rustdesk service, got %#v", service)
	}
	prueba, ok := service["prueba"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected prueba payload, got %#v", service)
	}
	if strings.TrimSpace(stringValue(prueba["resumen"])) == "" {
		t.Fatalf("expected probe summary, got %#v", prueba)
	}
}

func TestEmpresaSoporteRemotoHandlerBlocksSessionWhenPlanLimitReached(t *testing.T) {
	dbEmp := openPermsTestDB(t, "soporte_remoto_limit_handler.db")
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	h := EmpresaSoporteRemotoHandler(dbEmp)

	reqCfg := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=config&empresa_id=915", strings.NewReader(`{"habilitado":true,"requiere_aprobacion_operador":false,"auto_cerrar_minutos":30,"max_conexiones_mes":1,"max_minutos_mes":1,"max_dispositivos":1}`))
	reqCfg.Header.Set("Content-Type", "application/json")
	reqCfg.Header.Set("X-Admin-Email", "admin915@test.local")
	rrCfg := httptest.NewRecorder()
	h.ServeHTTP(rrCfg, reqCfg)
	if rrCfg.Code != http.StatusOK {
		t.Fatalf("expected 200 config with limits, got %d body=%s", rrCfg.Code, rrCfg.Body.String())
	}

	reqDevice := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=crear_dispositivo&empresa_id=915", strings.NewReader(`{"nombre_equipo":"Caja 915","stream_url":"https://remote.example/915","acceso_pin":"9150"}`))
	reqDevice.Header.Set("Content-Type", "application/json")
	reqDevice.Header.Set("X-Admin-Email", "admin915@test.local")
	rrDevice := httptest.NewRecorder()
	h.ServeHTTP(rrDevice, reqDevice)
	if rrDevice.Code != http.StatusCreated {
		t.Fatalf("expected 201 create device, got %d body=%s", rrDevice.Code, rrDevice.Body.String())
	}
	deviceResp := decodeSoporteRemotoBody(t, rrDevice)
	deviceRaw := deviceResp["dispositivo"].(map[string]interface{})
	deviceID := int64(numberValue(deviceRaw["id"]))

	reqSession1 := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=solicitar_sesion&empresa_id=915", strings.NewReader(`{"dispositivo_id":`+itoa64(deviceID)+`,"motivo":"Primera sesion","duracion_min":1}`))
	reqSession1.Header.Set("Content-Type", "application/json")
	reqSession1.Header.Set("X-Admin-Email", "admin915@test.local")
	rrSession1 := httptest.NewRecorder()
	h.ServeHTTP(rrSession1, reqSession1)
	if rrSession1.Code != http.StatusCreated {
		t.Fatalf("expected 201 first session, got %d body=%s", rrSession1.Code, rrSession1.Body.String())
	}
	sessionResp := decodeSoporteRemotoBody(t, rrSession1)
	sessionRaw := sessionResp["session"].(map[string]interface{})
	codigoSesion := stringValue(sessionRaw["codigo_sesion"])

	reqFinish := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=finalizar_sesion&empresa_id=915", strings.NewReader(`{"codigo_sesion":"`+codigoSesion+`"}`))
	reqFinish.Header.Set("Content-Type", "application/json")
	rrFinish := httptest.NewRecorder()
	h.ServeHTTP(rrFinish, reqFinish)
	if rrFinish.Code != http.StatusOK {
		t.Fatalf("expected 200 finish session, got %d body=%s", rrFinish.Code, rrFinish.Body.String())
	}

	reqSession2 := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=solicitar_sesion&empresa_id=915", strings.NewReader(`{"dispositivo_id":`+itoa64(deviceID)+`,"motivo":"Segunda sesion bloqueada","duracion_min":1}`))
	reqSession2.Header.Set("Content-Type", "application/json")
	reqSession2.Header.Set("X-Admin-Email", "admin915@test.local")
	rrSession2 := httptest.NewRecorder()
	h.ServeHTTP(rrSession2, reqSession2)
	if rrSession2.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412 blocked session, got %d body=%s", rrSession2.Code, rrSession2.Body.String())
	}
	body := decodeSoporteRemotoBody(t, rrSession2)
	usoRaw, ok := body["uso"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected uso object in blocked response: %#v", body)
	}
	if numberValue(usoRaw["intentos_bloqueados_mes"]) < 1 {
		t.Fatalf("expected blocked attempts >= 1, got %#v", usoRaw)
	}
}

func TestEmpresaSoporteRemotoHandlerBlocksApprovalWhenRustDeskDailyLimitReached(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))

	dbEmp := openPermsTestDB(t, "soporte_remoto_daily_approval_handler.db")
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaSoporteRemotoSchema: %v", err)
	}

	h := EmpresaSoporteRemotoHandler(dbEmp)

	reqCfg := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=config&empresa_id=916", strings.NewReader(`{"habilitado":true,"proveedor_preferido":"rustdesk_oss","modo_operacion":"cliente_local","requiere_aprobacion_operador":true,"auto_cerrar_minutos":30,"max_conexiones_mes":10,"max_minutos_mes":100,"max_minutos_dia_rustdesk":1,"max_dispositivos":2}`))
	reqCfg.Header.Set("Content-Type", "application/json")
	reqCfg.Header.Set("X-Admin-Email", "admin916@test.local")
	rrCfg := httptest.NewRecorder()
	h.ServeHTTP(rrCfg, reqCfg)
	if rrCfg.Code != http.StatusOK {
		t.Fatalf("expected 200 config with daily rustdesk limit, got %d body=%s", rrCfg.Code, rrCfg.Body.String())
	}

	reqDevice := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=crear_dispositivo&empresa_id=916", strings.NewReader(`{"nombre_equipo":"RustDesk 916","rustdesk_device_id":"916-RUSTDESK","rustdesk_password":"clave-916","acceso_pin":"9160"}`))
	reqDevice.Header.Set("Content-Type", "application/json")
	reqDevice.Header.Set("X-Admin-Email", "admin916@test.local")
	rrDevice := httptest.NewRecorder()
	h.ServeHTTP(rrDevice, reqDevice)
	if rrDevice.Code != http.StatusCreated {
		t.Fatalf("expected 201 create rustdesk device, got %d body=%s", rrDevice.Code, rrDevice.Body.String())
	}
	deviceResp := decodeSoporteRemotoBody(t, rrDevice)
	deviceRaw := deviceResp["dispositivo"].(map[string]interface{})
	deviceID := int64(numberValue(deviceRaw["id"]))

	reqSession1 := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=solicitar_sesion&empresa_id=916", strings.NewReader(`{"dispositivo_id":`+itoa64(deviceID)+`,"motivo":"Primera sesion diaria","duracion_min":1}`))
	reqSession1.Header.Set("Content-Type", "application/json")
	reqSession1.Header.Set("X-Admin-Email", "admin916@test.local")
	rrSession1 := httptest.NewRecorder()
	h.ServeHTTP(rrSession1, reqSession1)
	if rrSession1.Code != http.StatusCreated {
		t.Fatalf("expected 201 first pending session, got %d body=%s", rrSession1.Code, rrSession1.Body.String())
	}
	firstSession := decodeSoporteRemotoBody(t, rrSession1)["session"].(map[string]interface{})
	firstCode := stringValue(firstSession["codigo_sesion"])

	reqApprove1 := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=aprobar_sesion&empresa_id=916", strings.NewReader(`{"codigo_sesion":"`+firstCode+`"}`))
	reqApprove1.Header.Set("Content-Type", "application/json")
	rrApprove1 := httptest.NewRecorder()
	h.ServeHTTP(rrApprove1, reqApprove1)
	if rrApprove1.Code != http.StatusOK {
		t.Fatalf("expected 200 first approve, got %d body=%s", rrApprove1.Code, rrApprove1.Body.String())
	}

	reqFinish1 := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=finalizar_sesion&empresa_id=916", strings.NewReader(`{"codigo_sesion":"`+firstCode+`"}`))
	reqFinish1.Header.Set("Content-Type", "application/json")
	rrFinish1 := httptest.NewRecorder()
	h.ServeHTTP(rrFinish1, reqFinish1)
	if rrFinish1.Code != http.StatusOK {
		t.Fatalf("expected 200 finish first rustdesk session, got %d body=%s", rrFinish1.Code, rrFinish1.Body.String())
	}

	reqSession2 := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=solicitar_sesion&empresa_id=916", strings.NewReader(`{"dispositivo_id":`+itoa64(deviceID)+`,"motivo":"Segunda sesion diaria","duracion_min":1}`))
	reqSession2.Header.Set("Content-Type", "application/json")
	reqSession2.Header.Set("X-Admin-Email", "admin916@test.local")
	rrSession2 := httptest.NewRecorder()
	h.ServeHTTP(rrSession2, reqSession2)
	if rrSession2.Code != http.StatusCreated {
		t.Fatalf("expected 201 second pending session, got %d body=%s", rrSession2.Code, rrSession2.Body.String())
	}
	secondSession := decodeSoporteRemotoBody(t, rrSession2)["session"].(map[string]interface{})
	secondCode := stringValue(secondSession["codigo_sesion"])

	reqApprove2 := httptest.NewRequest(http.MethodPost, "/api/empresa/soporte_remoto?action=aprobar_sesion&empresa_id=916", strings.NewReader(`{"codigo_sesion":"`+secondCode+`"}`))
	reqApprove2.Header.Set("Content-Type", "application/json")
	rrApprove2 := httptest.NewRecorder()
	h.ServeHTTP(rrApprove2, reqApprove2)
	if rrApprove2.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412 second approve blocked by rustdesk daily limit, got %d body=%s", rrApprove2.Code, rrApprove2.Body.String())
	}
	body := decodeSoporteRemotoBody(t, rrApprove2)
	usoRaw, ok := body["uso"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected uso object in blocked approval response: %#v", body)
	}
	if numberValue(usoRaw["minutos_consumidos_dia_rustdesk"]) < 1 {
		t.Fatalf("expected minutos_consumidos_dia_rustdesk >= 1, got %#v", usoRaw)
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
