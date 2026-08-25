package handlers

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestParseDIANRUTAIResponseMapsReviewedFiscalFields(t *testing.T) {
	nit := "900123456"
	dv, ok := calculateColombianNITDV(nit)
	if !ok {
		t.Fatal("test NIT must be valid for DV calculation")
	}
	raw := fmt.Sprintf(`{
		"numero_formulario":"100000000001",
		"concepto":"Actualizacion",
		"fecha_generacion":"2026-08-25",
		"nit":%q,
		"dv":%q,
		"tipo_contribuyente":"persona_natural",
		"razon_social":"Persona Natural Prueba",
		"pais_codigo":"CO",
		"departamento":"Magdalena",
		"departamento_codigo_dane":"47",
		"municipio":"Santa Marta",
		"municipio_codigo_dane":"001",
		"direccion_fiscal":"Direccion fiscal de prueba",
		"email_facturacion":"fiscal@example.test",
		"telefono_facturacion":"3000000000",
		"actividad_economica_principal":"6202",
		"responsabilidades_rut":[49]
	}`, nit, strconv.Itoa(dv))

	fields, warnings, err := parseDIANRUTAIResponse(raw)
	if err != nil {
		t.Fatalf("parseDIANRUTAIResponse() error = %v", err)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "no fue almacenado") {
		t.Fatalf("warnings = %#v, want only temporary-processing notice", warnings)
	}
	wants := map[string]interface{}{
		"nit":                        nit,
		"dv":                         strconv.Itoa(dv),
		"tipo_documento_emisor":      "NIT",
		"tipo_persona_fiscal":        "persona_natural",
		"departamento_codigo_dane":   "47",
		"municipio_codigo_dane":      "47001",
		"iva_responsabilidad":        "no_responsable_iva",
		"responsabilidad_tributaria": "R-99-PN",
		"responsabilidades_rut_json": `["49"]`,
	}
	for key, want := range wants {
		if got := fields[key]; got != want {
			t.Fatalf("%s = %#v, want %#v; fields=%#v", key, got, want, fields)
		}
	}
}

func TestParseDIANRUTAIResponseMapsDIANTaxLevels(t *testing.T) {
	fields, _, err := parseDIANRUTAIResponse(`{
		"nit":"900123456",
		"tipo_contribuyente":"persona_juridica",
		"razon_social":"Empresa Prueba SAS",
		"departamento_codigo_dane":"11",
		"municipio_codigo_dane":"11001",
		"direccion_fiscal":"Direccion fiscal",
		"responsabilidades_rut":[13,15,47,48]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := fields["responsabilidad_tributaria"]; got != "O-13;O-15;O-47" {
		t.Fatalf("responsabilidad_tributaria = %#v", got)
	}
	if got := fields["regimen_tributario_colombia"]; got != "simple" {
		t.Fatalf("regimen_tributario_colombia = %#v", got)
	}
	if got := fields["iva_responsabilidad"]; got != "responsable_iva" {
		t.Fatalf("iva_responsabilidad = %#v", got)
	}
}

func TestReadDIANRUTPDFUploadRejectsFakePDF(t *testing.T) {
	req := dianRUTMultipartRequest(t, "rut.pdf", []byte("not a PDF"))
	if _, status, err := readDIANRUTPDFUpload(req); err == nil || status != 400 {
		t.Fatalf("status=%d err=%v, want invalid PDF 400", status, err)
	}
}

func TestReadDIANRUTPDFUploadKeepsEmpresaAndPDFBytes(t *testing.T) {
	pdf := []byte("%PDF-1.7\n%%EOF")
	req := dianRUTMultipartRequest(t, "rut-empresa.pdf", pdf)
	upload, status, err := readDIANRUTPDFUpload(req)
	if err != nil || status != 200 {
		t.Fatalf("status=%d err=%v", status, err)
	}
	if upload.EmpresaID != 12 || upload.FileName != "rut-empresa.pdf" || !bytes.Equal(upload.Bytes, pdf) {
		t.Fatalf("unexpected upload = %#v", upload)
	}
}

func TestReadDIANRUTPDFUploadUsesAuthenticatedTenantOverMultipart(t *testing.T) {
	req := dianRUTMultipartRequest(t, "rut-empresa.pdf", []byte("%PDF-1.7\n%%EOF"))
	req = requestWithTenantContext(req, TenantContext{EmpresaID: 53, Module: "facturacion", Action: "CREATE"})
	upload, status, err := readDIANRUTPDFUpload(req)
	if err != nil || status != 200 {
		t.Fatalf("status=%d err=%v", status, err)
	}
	if upload.EmpresaID != 53 {
		t.Fatalf("empresa_id=%d, want authenticated tenant 53 instead of multipart tenant", upload.EmpresaID)
	}
}

func TestFacturacionElectronicaHTMLSeparatesRUTFromFormulario1876(t *testing.T) {
	path := filepath.Join("..", "..", "web", "administrar_empresa", "facturacion_electronica.html")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	html := string(raw)
	for _, required := range []string{
		`id="dianRutPdfInput"`,
		`id="btnDianRutPdf"`,
		`action=importar_rut_pdf_ia`,
		`id="dianNumeracionPdfInput"`,
		`Formulario 1876`,
		`el PDF no quedo almacenado`,
		`function setReviewedValue(id, value)`,
		`if (!txt(value)) return;`,
		`setReviewedValue("adv_email_facturacion", fields.email_facturacion)`,
	} {
		if !strings.Contains(html, required) {
			t.Fatalf("facturacion electronica HTML missing %q", required)
		}
	}
}

func dianRUTMultipartRequest(t *testing.T, fileName string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("empresa_id", "12"); err != nil {
		t.Fatal(err)
	}
	part, err := writer.CreateFormFile("archivo", fileName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("POST", "/api/empresa/facturacion_electronica/dian?action=importar_rut_pdf_ia", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
