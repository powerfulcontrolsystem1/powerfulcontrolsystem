package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestBuildLicenciaSoftwarePDFGeneraDocumentoValidoConMarca(t *testing.T) {
	pdf := buildLicenciaSoftwarePDF(licenciaSoftwarePDFInput{
		Title:       "Licencia de software Powerful Control System - Motel Calipso",
		Body:        "LICENCIA DE USO\n\nEmpresa: Motel Calipso\nPlan: Plan empresarial",
		CompanyName: "Motel Calipso",
		LicenseName: "Plan empresarial",
		LicenseCode: "PCS-7-99",
		IssuedAt:    time.Date(2026, 5, 31, 9, 0, 0, 0, time.UTC),
	})
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		t.Fatalf("pdf no inicia con cabecera PDF: %q", string(pdf[:8]))
	}
	if !bytes.Contains(pdf, []byte("Powerful Control System")) {
		t.Fatalf("pdf no incluye marca del sistema")
	}
	if !bytes.Contains(pdf, []byte("Motel Calipso")) {
		t.Fatalf("pdf no incluye empresa")
	}
	if !bytes.Contains(pdf, []byte("%%EOF")) {
		t.Fatalf("pdf no incluye cierre EOF")
	}
}

func TestBuildLicenciaActivationEmailMessageAdjuntaPDF(t *testing.T) {
	pdf := buildLicenciaSoftwarePDF(licenciaSoftwarePDFInput{CompanyName: "Motel Calipso"})
	msg := buildLicenciaActivationEmailMessage(
		"Powerful Control System",
		"no-reply@example.com",
		"cliente@example.com",
		"Tu licencia ya quedo activa",
		"Adjunto encontraras tu licencia.",
		"licencia.pdf",
		pdf,
	)
	text := string(msg)
	for _, want := range []string{
		"Content-Type: multipart/mixed",
		"Content-Type: application/pdf",
		"Content-Transfer-Encoding: base64",
		"Content-Disposition: attachment; filename=\"licencia.pdf\"",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("mensaje no contiene %q", want)
		}
	}
}

func TestBuildLicenciaActivationEmailMessageAdjuntaSoloFacturaElectronica(t *testing.T) {
	facturaPDF := []byte("%PDF-1.4\nfactura electronica\n%%EOF")
	msg := buildLicenciaActivationEmailMessageWithAttachments(
		"Powerful Control System",
		"no-reply@example.com",
		"cliente@example.com",
		"Bienvenido a Powerful Control System",
		"Adjunto encontraras la factura electronica. La licencia se descarga desde Administrar empresa.",
		[]licenciaEmailAttachment{
			{Filename: "factura-electronica.pdf", ContentType: "application/pdf", Data: facturaPDF},
		},
	)
	text := string(msg)
	for _, want := range []string{
		"Content-Type: multipart/mixed",
		"Content-Disposition: attachment; filename=\"factura-electronica.pdf\"",
		"La licencia se descarga desde Administrar empresa.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("mensaje no contiene %q", want)
		}
	}
	if strings.Contains(text, "filename=\"licencia.pdf\"") {
		t.Fatalf("el correo de bienvenida no debe adjuntar la licencia del software")
	}
	if got := strings.Count(text, "Content-Type: application/pdf"); got != 1 {
		t.Fatalf("adjuntos PDF = %d, want 1", got)
	}
}

func TestBuildLicenciaFacturaElectronicaPDFGeneraDocumentoValido(t *testing.T) {
	doc := dbpkg.EmpresaDocumentoFacturacion{
		DocumentoCodigo:  "LIC-WOMPI-REF 123",
		NumeroLegal:      "FE-PCS-000123",
		CodigoValidacion: "CUFE-123",
		EstadoDocumento:  "emitida",
		PaisCodigo:       "CO",
		AmbienteFE:       "habilitacion",
		MontoTotal:       60000,
		Moneda:           "COP",
		FechaDocumento:   "2026-06-04 12:00:00",
	}
	pdf, filename := buildLicenciaFacturaElectronicaPDF(doc, "Cliente Prueba", "Plan 60000", "wompi", "REF 123")
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		t.Fatalf("pdf no inicia con cabecera PDF")
	}
	for _, want := range []string{"Factura electronica por compra de licencia", "FE-PCS-000123", "Cliente Prueba", "Plan 60000", "60000 COP"} {
		if !bytes.Contains(pdf, []byte(want)) {
			t.Fatalf("pdf no contiene %q", want)
		}
	}
	if filename != "factura-electronica-lic-wompi-ref-123.pdf" {
		t.Fatalf("filename = %s", filename)
	}
}

func TestBuildLicenciaSoftwarePDFForEmpresaUsaFormatoDefault(t *testing.T) {
	empresa := &dbpkg.Empresa{ID: 7, EmpresaID: 7, Nombre: "Motel Calipso", Nit: "900123456"}
	lic := &dbpkg.Licencia{
		ID:          99,
		EmpresaID:   7,
		Nombre:      "Plan empresarial",
		Valor:       60000,
		FechaInicio: "2026-05-01",
		FechaFin:    "2026-06-01",
	}
	pdf, filename, err := buildLicenciaSoftwarePDFForEmpresa(nil, empresa, lic, "Sistema", "REF-123", "60000", time.Date(2026, 5, 31, 10, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("no esperaba error generando pdf: %v", err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		t.Fatalf("pdf no inicia con cabecera PDF")
	}
	if !strings.Contains(filename, "motel-calipso") {
		t.Fatalf("filename no incluye empresa normalizada: %s", filename)
	}
	for _, want := range []string{"Motel Calipso", "Plan empresarial", "PCS-7-99"} {
		if !bytes.Contains(pdf, []byte(want)) {
			t.Fatalf("pdf no contiene %q", want)
		}
	}
}

func TestEmpresaLicenciaSistemaPDFHandlerRechazaMetodoNoGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/licencia_sistema/pdf?empresa_id=7", nil)
	rr := httptest.NewRecorder()
	EmpresaLicenciaSistemaPDFHandler(nil, nil).ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}
