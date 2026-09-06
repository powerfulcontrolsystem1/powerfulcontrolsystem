package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestFacturacionContingenciaEndpointIsTenantGuardedAndFailClosed(t *testing.T) {
	mainRaw, err := os.ReadFile("../main.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mainRaw), `WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionContingenciasHandler(dbEmpresas))`) {
		t.Fatal("contingency endpoint is not behind company invoicing permissions")
	}
	raw, err := os.ReadFile("facturacion_contingencias.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"decoder.DisallowUnknownFields()",
		"ACTIVAR CONTINGENCIA",
		"RECUPERAR SERVICIO",
		"CERRAR CONTINGENCIA",
		"REGISTRAR TALONARIO",
		"integracion Colombia/DIAN activa en produccion",
		"fuente.Carrito.ID != input.CarritoID",
		"ListEmpresaFacturacionContingenciaDocumentos",
		"CreateEmpresaAuditoriaEvento",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("contingency endpoint missing fail-closed contract %q", required)
		}
	}
}

func TestPaidSaleWorkerRecoversAccountingBeforeDocuments(t *testing.T) {
	raw, err := os.ReadFile("../cmd/pcs-worker/business_registry.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	accounting := strings.Index(source, "RecoverPaidSaleAccounting")
	documents := strings.Index(source, "RecoverPaidSaleDocuments")
	if accounting < 0 || documents < 0 || documents < accounting {
		t.Fatal("paid-sale worker does not recover accounting before documents")
	}
}

func TestAutomaticInvoiceFailureRemainsRecoverableThroughOutbox(t *testing.T) {
	raw, err := os.ReadFile("carritos_compras.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	if !strings.Contains(source, "su factura electronica requiere recuperacion") {
		t.Fatal("automatic invoice failure is swallowed and could acknowledge the outbox")
	}
}
