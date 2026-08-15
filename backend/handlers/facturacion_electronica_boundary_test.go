package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFacturacionElectronicaMantieneFronteraDeEmpresaYAnulacionFiscal(t *testing.T) {
	handlerRaw, err := os.ReadFile("facturacion_electronica.go")
	if err != nil {
		t.Fatalf("read facturacion handler: %v", err)
	}
	source := string(handlerRaw)
	for _, expected := range []string{
		"empresaID, err := parseEmpresaIDQuery(r)",
		"EmpresaID:       empresaID,",
		"anular_factura_nota_credito",
		"la factura solo cambia a anulada cuando la nota credito queda aceptada por DIAN",
		"GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp, payload.EmpresaID",
		"ListEmpresaDocumentosFacturacionByEmpresaContext(r.Context(), dbEmp",
		"GetEmpresaDocumentoFacturacionByCodigoContext(r.Context(), dbEmp",
		"UpsertEmpresaDocumentoFacturacionContext(r.Context(), dbEmp",
		"UpdateEmpresaDocumentoFacturacionClienteContext(r.Context(), dbEmp",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("missing fiscal tenant boundary %q", expected)
		}
	}

	mainRaw, err := os.ReadFile(filepath.Join("..", "main.go"))
	if err != nil {
		t.Fatalf("read route registration: %v", err)
	}
	if !strings.Contains(string(mainRaw), "WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaHandler(dbEmpresas, dbSuper))") {
		t.Fatal("electronic invoicing route must keep the enterprise permission wrapper")
	}
}
