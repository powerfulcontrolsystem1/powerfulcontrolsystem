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