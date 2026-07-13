package handlers

import "testing"

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

func TestSupportedAIAttachmentsAreClosedList(t *testing.T) {
	valid := []string{"compra.pdf", "lista.xlsx", "nota.docx", "datos.csv", "texto.txt", "foto.png"}
	for _, filename := range valid {
		if !isSupportedAIAttachment(&aiAttachment{Filename: filename, Bytes: []byte("x")}) {
			t.Fatalf("adjunto valido rechazado: %s", filename)
		}
	}
	invalid := []string{"programa.exe", "pagina.html", "vector.svg", "archivo.zip"}
	for _, filename := range invalid {
		if isSupportedAIAttachment(&aiAttachment{Filename: filename, Bytes: []byte("x")}) {
			t.Fatalf("adjunto peligroso aceptado: %s", filename)
		}
	}
}
