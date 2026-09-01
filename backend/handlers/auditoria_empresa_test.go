package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestResolveEmpresaAuditoriaPermissionAction(t *testing.T) {
	tests := []struct {
		name   string
		method string
		action string
		want   string
	}{
		{name: "consulta", method: http.MethodGet, want: permActionRead},
		{name: "exportacion forense", method: http.MethodGet, action: "export_forense", want: permActionRead},
		{name: "conectividad", method: http.MethodPost, action: "conexion", want: permActionRead},
		{name: "retencion destructiva", method: http.MethodPut, action: "retener", want: permActionDelete},
		{name: "purga destructiva", method: http.MethodPost, action: "purgar", want: permActionDelete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/empresa/auditoria/eventos?action="+tt.action, nil)
			if got := resolveEmpresaAuditoriaPermissionAction(req); got != tt.want {
				t.Fatalf("accion = %q, esperado %q", got, tt.want)
			}
		})
	}
}

func TestAuditorRoleCannotPurgeEmpresaAuditoria(t *testing.T) {
	if roleAllowsModuleAction("auditor", permModuleAuditoria, permActionDelete) {
		t.Fatal("el rol auditor no debe poder purgar la auditoria empresarial")
	}
	if !roleAllowsModuleAction("admin_empresa", permModuleAuditoria, permActionDelete) {
		t.Fatal("admin_empresa debe conservar la autorizacion explicita para la purga")
	}
}

func TestEmpresaAuditoriaEventosHandlerRejectsInvalidForensicFormatBeforeDatabaseAccess(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/auditoria/eventos?empresa_id=7&action=export_forense&format=xml", nil)
	rec := httptest.NewRecorder()

	EmpresaAuditoriaEventosHandler(nil).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, esperado %d: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "format invalido") {
		t.Fatalf("respuesta inesperada: %s", rec.Body.String())
	}
}

func TestBuildAuditoriaForenseExportPayloadBuildsDeterministicCustodyChain(t *testing.T) {
	rows := []dbpkg.EmpresaAuditoriaEvento{
		{ID: 11, FechaEvento: "2026-08-31 09:00:00", Modulo: "ventas", Accion: "cerrar", Resultado: "ok", CodigoHTTP: http.StatusOK, MetadataJSON: `{"origen":"pos"}`},
		{ID: 12, FechaEvento: "2026-08-31 09:01:00", Modulo: "finanzas", Accion: "aprobar", Resultado: "ok", CodigoHTTP: http.StatusOK, MetadataJSON: `{}`},
	}

	payload, err := buildAuditoriaForenseExportPayload(7, dbpkg.EmpresaAuditoriaEventoFilter{Limit: 200, Offset: 0}, 5, rows)
	if err != nil {
		t.Fatalf("buildAuditoriaForenseExportPayload: %v", err)
	}
	if payload.Manifest.TotalCoincidencias != 5 || payload.Manifest.TotalRegistros != len(rows) {
		t.Fatalf("manifiesto inconsistente: %+v", payload.Manifest)
	}
	if payload.Registros[0].HashCadenaAnterior != "GENESIS" {
		t.Fatalf("inicio de cadena = %q", payload.Registros[0].HashCadenaAnterior)
	}
	if payload.Registros[1].HashCadenaAnterior != payload.Registros[0].HashCadena {
		t.Fatal("la cadena de custodia no enlaza los registros consecutivos")
	}
	if payload.Manifest.HashCadenaFinal != payload.Registros[1].HashCadena || payload.Manifest.HashGlobal == "" {
		t.Fatalf("manifiesto sin hashes verificables: %+v", payload.Manifest)
	}
}
