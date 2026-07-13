package db

import (
	"os"
	"strings"
	"testing"
)

func TestDisablePowerfulSystemEmpresaInternalLicenseAvoidsTextTimestampCASE(t *testing.T) {
	raw, err := os.ReadFile("licencias_empresa_sistema.go")
	if err != nil {
		t.Fatalf("read licencias_empresa_sistema.go: %v", err)
	}
	src := string(raw)
	if strings.Contains(src, "COALESCE(fecha_fin, '')") {
		t.Fatal("la licencia interna no debe mezclar fecha_fin timestamp con texto vacio en PostgreSQL")
	}
	if !strings.Contains(src, "fecha_fin = CASE WHEN fecha_fin IS NULL THEN") {
		t.Fatal("la licencia interna debe conservar fecha_fin con un CASE tipado por NULL")
	}
}
