package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaNominaHandlerDoesNotCreateSchema(t *testing.T) {
	body, err := os.ReadFile("nomina_sueldos.go")
	if err != nil {
		t.Fatalf("read payroll handler: %v", err)
	}
	source := string(body)
	start := strings.Index(source, "func EmpresaNominaSueldosHandler")
	if start < 0 {
		t.Fatal("payroll handler not found")
	}
	section := source[start:]
	if strings.Contains(section, "EnsureEmpresaNominaSchema(") {
		t.Fatal("payroll requests must not create schema")
	}
	if !strings.Contains(section, "EmpresaNominaSchemaReady(") {
		t.Fatal("payroll handler must verify the migrated schema")
	}
}
