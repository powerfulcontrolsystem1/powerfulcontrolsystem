package db

import "testing"

func TestEmpresaCamarasSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaCamarasSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}
