package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLicenciaModulosCSVControlsModuleAccess(t *testing.T) {
	allowed, ordered := parseLicenciaModulosCSV(" ventas, SEGURIDAD, carnets, modulo_desconocido, ventas ")
	if len(ordered) != 3 {
		t.Fatalf("expected only known unique modules, got %v", ordered)
	}
	if !isModuloPermitidoByLicencia(permModuleVentas, allowed) {
		t.Fatal("expected ventas to be enabled by licencia")
	}
	if !isModuloPermitidoByLicencia(permModuleSeguridad, allowed) {
		t.Fatal("expected seguridad to be enabled by licencia")
	}
	if isModuloPermitidoByLicencia(permModuleFinanzas, allowed) {
		t.Fatal("expected finanzas to be disabled when omitted from licencia")
	}
	if !isModuloPermitidoByLicencia(permModuleCarnets, allowed) {
		t.Fatal("expected carnets to be enabled by licencia")
	}
}

func TestApplyLicenciaRestriccionesDisablesActionsForInactiveModules(t *testing.T) {
	rows := []permissionModuleMatrixRow{
		{
			Modulo: permModuleVentas,
			Acciones: map[string]bool{
				permActionRead:   true,
				permActionCreate: true,
			},
		},
		{
			Modulo: permModuleFinanzas,
			Acciones: map[string]bool{
				permActionRead:    true,
				permActionCreate:  true,
				permActionUpdate:  true,
				permActionDelete:  true,
				permActionApprove: true,
			},
		},
	}
	allowed, _ := parseLicenciaModulosCSV(permModuleVentas)

	filtered := applyLicenciaRestriccionesToModuleRows(rows, allowed)

	if !filtered[0].Acciones[permActionRead] || !filtered[0].Acciones[permActionCreate] {
		t.Fatal("expected enabled licencia module actions to remain active")
	}
	for _, action := range permissionActionsCatalogOrdered {
		if filtered[1].Acciones[action] {
			t.Fatalf("expected finanzas action %s to be disabled by licencia", action)
		}
	}
}

func TestValidateEmpresaIDConsistencyRejectsBodyQueryMismatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/empresa/productos?empresa_id=7", bytes.NewBufferString(`{"empresa_id":8,"nombre":"Producto"}`))
	r.Header.Set("Content-Type", "application/json")

	if err := validateEmpresaIDConsistency(r, 7); err == nil {
		t.Fatal("expected mismatch between query empresa_id and JSON empresa_id to be rejected")
	}

	body := new(bytes.Buffer)
	if _, err := body.ReadFrom(r.Body); err != nil {
		t.Fatalf("reading restored body: %v", err)
	}
	if got := body.String(); !strings.Contains(got, `"empresa_id":8`) {
		t.Fatalf("expected JSON body to be restored for downstream handlers, got %q", got)
	}
}

func TestValidateEmpresaIDConsistencyAllowsMatchingSources(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/api/empresa/productos?empresa_id=7", bytes.NewBufferString(`{"empresaId":7,"nombre":"Producto"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Empresa-ID", "7")

	if err := validateEmpresaIDConsistency(r, 7); err != nil {
		t.Fatalf("expected matching empresa_id sources to pass, got %v", err)
	}
}
