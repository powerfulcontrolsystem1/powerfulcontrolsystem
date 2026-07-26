package db

import "testing"

func TestEmpresaAIMemoryRejectsCredentialMarkers(t *testing.T) {
	if !empresaAIMemoryContainsSecret("preferencia", "api_key", `"valor"`) {
		t.Fatal("api_key debe rechazarse")
	}
	if empresaAIMemoryContainsSecret("preferencia", "idioma", `"es-CO"`) {
		t.Fatal("preferencia normal no debe rechazarse")
	}
}
