package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaCamarasHandlerDoesNotCreateSchema(t *testing.T) {
	body, err := os.ReadFile("camaras.go")
	if err != nil {
		t.Fatalf("read cameras handler: %v", err)
	}
	source := string(body)
	start := strings.Index(source, "func EmpresaCamarasHandler")
	if start < 0 {
		t.Fatal("camera handler not found")
	}
	section := source[start:]
	if strings.Contains(section, "EnsureEmpresaCamarasSchema(") {
		t.Fatal("camera requests must not create schema")
	}
	if !strings.Contains(section, "EmpresaCamarasSchemaReady(") {
		t.Fatal("camera handler must verify the migrated schema")
	}
}
