package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestBuildLicenciaPaymentReceiptPDFNoExponePayloadYGeneraPDF(t *testing.T) {
	pdf, filename := buildLicenciaPaymentReceiptPDF(&dbpkg.Empresa{EmpresaID: 17, Nombre: "Empresa Prueba"}, dbpkg.EmpresaLicenciaPagoResumen{
		ID:             44,
		Proveedor:      "wompi",
		LicenciaNombre: "Plan mensual",
		Referencia:     "REF-44",
		TransaccionID:  "TX-44",
		Estado:         "APPROVED",
		FechaCreacion:  "2026-07-13 10:00:00",
	})
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) || !bytes.Contains(pdf, []byte("Empresa Prueba")) {
		t.Fatal("el comprobante debe ser un PDF identificable de la empresa")
	}
	if strings.Contains(strings.ToLower(string(pdf)), "raw_payload") {
		t.Fatal("el comprobante no debe exponer payload de pasarela")
	}
	if filename != "comprobante-licencia-ref-44.pdf" {
		t.Fatalf("filename = %q", filename)
	}
}

func TestEmpresaLicenciasComprobantesHandlerRechazaMutaciones(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/licencias/comprobantes", nil)
	rr := httptest.NewRecorder()
	EmpresaLicenciasComprobantesHandler(nil, nil).ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestEmpresaLicenciasComprobantesHandlerExigeEmpresaValidada(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/licencias/comprobantes?empresa_id=999", nil)
	rr := httptest.NewRecorder()
	EmpresaLicenciasComprobantesHandler(nil, nil).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}
