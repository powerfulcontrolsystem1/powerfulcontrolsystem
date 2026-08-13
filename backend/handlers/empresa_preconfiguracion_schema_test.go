package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaPreconfiguracionDoesNotCreateSchemaDuringHTTPRequest(t *testing.T) {
	body, err := os.ReadFile("empresa_preconfiguracion.go")
	if err != nil {
		t.Fatalf("read empresa preconfiguration handler: %v", err)
	}
	source := string(body)
	for _, forbidden := range []string{
		"EnsureEmpresaProductosSchema(",
		"EnsureEmpresaUsuariosAuthSchema(",
		"EnsureEmpresaConfiguracionOperativaSchema(",
		"EnsureEmpresaComisionesServicioSchema(",
		"EnsureEmpresaTarifasPorMinutosSchema(",
		"EnsureEmpresaTarifasPorDiaSchema(",
		"EnsureEmpresaTarifasMotelSchema(",
		"EnsureEmpresaControlElectricoSchema(",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("empresa preconfiguration must not execute runtime DDL through %s", forbidden)
		}
	}
	for _, required := range []string{
		"EmpresaProductosSchemaReady(",
		"EmpresaUsuariosAuthSchemaReady(",
		"EmpresaConfiguracionOperativaSchemaReady(",
		"EmpresaComisionesServicioSchemaReady(",
		"EmpresaTarifasPorMinutosSchemaReady(",
		"EmpresaTarifasPorDiaSchemaReady(",
		"EmpresaTarifasMotelSchemaReady(",
		"EmpresaControlElectricoSchemaReady(",
	} {
		if !strings.Contains(source, required) {
			t.Fatalf("empresa preconfiguration must verify migrated schema through %s", required)
		}
	}
}
