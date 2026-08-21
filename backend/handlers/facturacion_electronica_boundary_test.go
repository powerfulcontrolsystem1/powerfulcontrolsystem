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
		"processFacturacionRetryQueueContext(r.Context(), dbEmp, dbSuper",
		"pg_try_advisory_lock($1::bigint)",
		"facturacionDocumentoAdvisoryLockKey(doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo)",
		"el documento ya tiene una integracion fiscal en proceso",
		"RunFacturacionElectronicaRetriesScheduled",
		"action == \"artefactos\"",
		"action == \"descargar_artefacto\"",
		"loadFacturacionFiscalArtifactItem(item)",
		"integridad SHA-256 del artefacto fiscal no coincide",
		"requireFiscalArtifacts := strings.EqualFold(strings.TrimSpace(doc.PaisCodigo), \"CO\")",
		"XML fiscal firmado no disponible o con integridad invalida",
		"representacion PDF fiscal no disponible o con integridad invalida",
		"validateDIANDocumentPreflight(dianCfg, doc.EmpresaID, docPayload, xmlFirmado, \"envio_real\")",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("missing fiscal tenant boundary %q", expected)
		}
	}
	connectionStart := strings.Index(source, `if action == "estado_conexion_dian" || action == "estado_conexion"`)
	if connectionStart < 0 {
		t.Fatal("DIAN connection status block was not found")
	}
	connectionEnd := strings.Index(source[connectionStart:], `if action == "catalogo_dian_colombia"`)
	if connectionEnd < 0 {
		t.Fatal("DIAN connection status block boundary was not found")
	}
	if strings.Contains(source[connectionStart:connectionStart+connectionEnd], "processFacturacionRetryQueue") {
		t.Fatal("GET connection status must not process or transmit fiscal retries")
	}

	mainRaw, err := os.ReadFile(filepath.Join("..", "main.go"))
	if err != nil {
		t.Fatalf("read route registration: %v", err)
	}
	if !strings.Contains(string(mainRaw), "WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaHandler(dbEmpresas, dbSuper))") {
		t.Fatal("electronic invoicing route must keep the enterprise permission wrapper")
	}
}
