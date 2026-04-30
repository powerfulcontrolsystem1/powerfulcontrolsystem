package handlers

import (
	"archive/zip"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestDynamicDocumentFileGeneration(t *testing.T) {
	record := dynamicDocumentRecord{
		ID:           "0123456789abcdef0123456789abcdef",
		EmpresaID:    77,
		Title:        "Reporte de prueba",
		Content:      "| Producto | Total |\n| --- | --- |\n| Cafe | 12000 |",
		InputFormat:  "markdown",
		TemplateName: "reporte",
		HTML:         "<html><body><h1>Reporte de prueba</h1><table><tr><td>Cafe</td></tr></table></body></html>",
		PlainText:    "Reporte de prueba\nCafe 12000",
		Variables: map[string]interface{}{
			"cliente": "Cliente Demo",
			"total":   12000,
		},
		CreatedAt: "2026-04-29T12:00:00Z",
		CreatedBy: "admin@test.local",
	}
	dir, err := ensureDynamicDocumentDir()
	if err != nil {
		t.Fatalf("ensureDynamicDocumentDir: %v", err)
	}
	for _, ext := range []string{"pdf", "docx", "xlsx", "txt", "json", "html", "record.json"} {
		_ = os.Remove(filepath.Join(dir, record.ID+"."+ext))
	}
	if err := saveDynamicDocumentRecord(record); err != nil {
		t.Fatalf("saveDynamicDocumentRecord: %v", err)
	}
	t.Cleanup(func() {
		for _, ext := range []string{"pdf", "docx", "xlsx", "txt", "json", "html", "record.json"} {
			_ = os.Remove(filepath.Join(dir, record.ID+"."+ext))
		}
	})

	for _, format := range []string{"pdf", "docx", "xlsx", "txt", "json", "html"} {
		path, contentType, err := ensureDynamicDocumentFile(context.Background(), record, format)
		if err != nil {
			t.Fatalf("ensureDynamicDocumentFile(%s): %v", format, err)
		}
		if contentType == "" {
			t.Fatalf("content type vacio para %s", format)
		}
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", format, err)
		}
		if info.Size() == 0 {
			t.Fatalf("archivo %s vacio", format)
		}
	}

	assertPDFHeader(t, filepath.Join(dir, record.ID+".pdf"))
	assertDOCXDocumentXML(t, filepath.Join(dir, record.ID+".docx"))
	assertXLSXContent(t, filepath.Join(dir, record.ID+".xlsx"))
	assertJSONRecord(t, filepath.Join(dir, record.ID+".json"))
}

func TestDynamicDocumentProfessionalFilenameAndRedaction(t *testing.T) {
	filename := buildDynamicDocumentProfessionalBaseFilename("Mi Empresa SAS", "contrato", mustParseDocTestDate(t, "2026-04-30"))
	if filename != "mi_empresa_sas_contrato_2026-04-30" {
		t.Fatalf("filename = %q", filename)
	}

	redacted := redactDynamicDocumentSecrets("Cliente: Demo\napi_key: sk-secret\npassword=abc123\nTotal: 10")
	if strings.Contains(redacted, "sk-secret") || strings.Contains(redacted, "abc123") {
		t.Fatalf("redaction leaked secret: %q", redacted)
	}
	if !strings.Contains(redacted, "[REDACTADO]") {
		t.Fatalf("expected redaction marker, got %q", redacted)
	}
}

func TestDynamicDocumentRecordFromChatContentSanitizesHTMLAndKeepsMetadata(t *testing.T) {
	record, err := buildDynamicDocumentRecordFromContent(DynamicDocumentRequest{
		EmpresaID:    42,
		Title:        "Contrato corto",
		Content:      `<h2 onclick="alert(1)">Contrato</h2><a href="javascript:alert(1)">firmar</a><script>alert(1)</script><p>Total {{.total}}</p>`,
		InputFormat:  "html",
		TemplateName: "contrato",
		ModelID:      dynamicDocumentModelID,
		Variables:    map[string]interface{}{"total": "12000"},
		Metadata:     map[string]interface{}{"origin": "chat_ia", "document_type": "contrato"},
	}, "admin@test.local", "Empresa Demo")
	if err != nil {
		t.Fatalf("buildDynamicDocumentRecordFromContent: %v", err)
	}
	if strings.Contains(strings.ToLower(record.HTML), "<script") {
		t.Fatalf("html contiene script sin sanitizar: %s", record.HTML)
	}
	if strings.Contains(strings.ToLower(record.HTML), "onclick") {
		t.Fatalf("html contiene event handler sin sanitizar: %s", record.HTML)
	}
	if strings.Contains(strings.ToLower(record.HTML), "javascript:") {
		t.Fatalf("html contiene URL javascript sin sanitizar: %s", record.HTML)
	}
	if !strings.Contains(record.PlainText, "12000") {
		t.Fatalf("variables no aplicadas al texto plano: %q", record.PlainText)
	}
	if record.Metadata["empresa_nombre"] != "Empresa Demo" {
		t.Fatalf("empresa_nombre metadata = %v", record.Metadata["empresa_nombre"])
	}
	if dynamicDocumentDownloadFilename(record, "docx") != "contrato_corto.docx" {
		t.Fatalf("filename fallback inesperado: %s", dynamicDocumentDownloadFilename(record, "docx"))
	}
}

func TestDynamicDocumentFormatsAndTableExtraction(t *testing.T) {
	formats := normalizeDynamicDocumentFormats([]string{"xlsx", "pdf", "pdf", "exe", "json"})
	got := strings.Join(formats, ",")
	if got != "xlsx,pdf,json" {
		t.Fatalf("formatos normalizados = %q", got)
	}
	rows := extractFirstMarkdownTable("| Item | Total |\n| --- | ---: |\n| Corte | 35000 |")
	if len(rows) != 2 || rows[0][0] != "Item" || rows[1][1] != "35000" {
		t.Fatalf("tabla markdown no preservada: %#v", rows)
	}
}

func mustParseDocTestDate(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		t.Fatalf("parse date: %v", err)
	}
	return parsed
}

func assertPDFHeader(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pdf: %v", err)
	}
	if !strings.HasPrefix(string(raw), "%PDF") {
		t.Fatalf("pdf no tiene cabecera PDF")
	}
}

func assertDOCXDocumentXML(t *testing.T, path string) {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open docx zip: %v", err)
	}
	defer reader.Close()
	for _, file := range reader.File {
		if file.Name == "word/document.xml" {
			return
		}
	}
	t.Fatalf("docx no contiene word/document.xml")
}

func assertXLSXContent(t *testing.T, path string) {
	t.Helper()
	file, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("open xlsx: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()
	value, err := file.GetCellValue("Documento", "A1")
	if err != nil {
		t.Fatalf("read xlsx cell: %v", err)
	}
	if value != "Producto" {
		t.Fatalf("xlsx A1 = %q, want Producto", value)
	}
}

func assertJSONRecord(t *testing.T, path string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read json: %v", err)
	}
	var record dynamicDocumentRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		t.Fatalf("json invalido: %v", err)
	}
	if record.Title != "Reporte de prueba" {
		t.Fatalf("titulo json = %q", record.Title)
	}
}
