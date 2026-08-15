package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestComprasDocumentosHTTPPropagaContextoAlRepositorio(t *testing.T) {
	raw, err := os.ReadFile("compras.go")
	if err != nil {
		t.Fatalf("read compras handler: %v", err)
	}
	source := string(raw)
	for _, forbidden := range []string{
		"ListEmpresaDocumentosCompraByEmpresa(dbEmp",
		"SetEmpresaDocumentoCompraEstadoByCodigo(dbEmp",
		"UpdateEmpresaDocumentoCompraComprobante(dbEmp",
		"GetEmpresaDocumentoCompraByCodigo(dbEmp",
		"UpsertEmpresaDocumentoCompra(dbEmp",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("compras handler bypasses request context with %q", forbidden)
		}
	}
	for _, expected := range []string{
		"ListEmpresaDocumentosCompraByEmpresaContext(r.Context(), dbEmp",
		"SetEmpresaDocumentoCompraEstadoByCodigoContext(r.Context(), dbEmp",
		"UpdateEmpresaDocumentoCompraComprobanteContext(r.Context(), dbEmp",
		"GetEmpresaDocumentoCompraByCodigoContext(r.Context(), dbEmp",
		"UpsertEmpresaDocumentoCompraContext(r.Context(), dbEmp",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("missing cancellable purchase operation %q", expected)
		}
	}
}
