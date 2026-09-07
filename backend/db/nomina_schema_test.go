package db

import "testing"

func TestEmpresaNominaSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaNominaSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}
