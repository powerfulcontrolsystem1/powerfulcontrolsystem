package db

import "testing"

func TestEmpresaCreateDedupKeyNormalizaNombreYNIT(t *testing.T) {
	a := empresaCreateDedupKey(3, " Restaurante ", "  Motel   Calipso ", "900.123-456 7", "ADMIN@EXAMPLE.COM ")
	b := empresaCreateDedupKey(3, "restaurante", "motel calipso", "9001234567", "admin@example.com")
	if a != b {
		t.Fatalf("dedup key should normalize equivalent payloads:\n%s\n%s", a, b)
	}
}

func TestEmpresaCreateDedupKeySeparaAdministrador(t *testing.T) {
	a := empresaCreateDedupKey(3, "Restaurante", "Motel Calipso", "9001234567", "admin1@example.com")
	b := empresaCreateDedupKey(3, "Restaurante", "Motel Calipso", "9001234567", "admin2@example.com")
	if a == b {
		t.Fatalf("dedup key must be scoped by administrator")
	}
}
