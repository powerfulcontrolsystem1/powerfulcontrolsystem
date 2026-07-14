package handlers

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestEmpresaAIModelCatalogIncludesAdvancedLuna(t *testing.T) {
	found := false
	for _, model := range empresaAIModelCatalog() {
		if model.ID == "openai:gpt-5.6-luna" {
			found = true
			if model.Endpoint != "https://api.openai.com/v1/responses" {
				t.Fatalf("endpoint avanzado inesperado: %q", model.Endpoint)
			}
		}
	}
	if !found {
		t.Fatal("el catalogo debe incluir GPT-5.6 Luna")
	}
}

func TestEmpresaAIModelCatalogIncludesTerraAndSolWithReasoning(t *testing.T) {
	for _, wanted := range []string{"openai:gpt-5.6-terra", "openai:gpt-5.6-sol"} {
		found := false
		for _, model := range empresaAIModelCatalog() {
			if model.ID != wanted {
				continue
			}
			found = true
			if len(model.ReasoningEfforts) != 6 || configuredAIReasoningEffort(nil, model) == "" {
				t.Fatalf("modelo %s sin esfuerzos configurables completos", wanted)
			}
		}
		if !found {
			t.Fatalf("no se encontro el modelo %s", wanted)
		}
	}
}

func TestSupportedAIAttachmentsAreClosedList(t *testing.T) {
	valid := []struct {
		filename string
		content  []byte
	}{
		{"compra.pdf", []byte("%PDF-1.7\\n")},
		{"lista.xlsx", testOfficeAttachment(t, "xl/workbook.xml")},
		{"nota.docx", testOfficeAttachment(t, "word/document.xml")},
		{"datos.csv", []byte("nombre,valor\\nCafe,1000\\n")},
		{"texto.txt", []byte("nota valida")},
		{"foto.png", []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}},
	}
	for _, candidate := range valid {
		if !isSupportedAIAttachment(&aiAttachment{Filename: candidate.filename, Bytes: candidate.content}) {
			t.Fatalf("adjunto valido rechazado: %s", candidate.filename)
		}
	}
	invalid := []struct {
		filename string
		content  []byte
	}{
		{"programa.exe", []byte("MZ")}, {"pagina.html", []byte("<html>")}, {"vector.svg", []byte("<svg>")}, {"archivo.zip", []byte("PK")},
		{"falsa.png", []byte("<script>alert(1)</script>")}, {"falso.pdf", []byte("not a pdf")}, {"falso.xlsx", []byte("PK-not-office")},
	}
	for _, candidate := range invalid {
		if isSupportedAIAttachment(&aiAttachment{Filename: candidate.filename, Bytes: candidate.content}) {
			t.Fatalf("adjunto peligroso aceptado: %s", candidate.filename)
		}
	}
}

func TestAIProviderErrorDoesNotExposeProviderBody(t *testing.T) {
	err := &aiProviderHTTPError{Provider: "openai", Status: 400, Body: `{"request_id":"private","message":"sensitive"}`}
	if got := err.Error(); bytes.Contains([]byte(got), []byte("sensitive")) || bytes.Contains([]byte(got), []byte("private")) {
		t.Fatalf("Error() no debe exponer el cuerpo del proveedor: %q", got)
	}
}

func testOfficeAttachment(t *testing.T, documentName string) []byte {
	t.Helper()
	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	for _, name := range []string{"[Content_Types].xml", documentName} {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write([]byte("<xml/>")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
