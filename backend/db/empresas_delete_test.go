package db

import "testing"

func TestEmpresaDeleteBuildEmpresaIDPredicateUsesTextArgForTextColumns(t *testing.T) {
	predicate, arg := empresaDeleteBuildEmpresaIDPredicate(empresaDeleteCandidateTable{
		Name:              "empresa_configuracion_avanzada",
		EmpresaIDDataType: "text",
		EmpresaIDUDTName:  "text",
	}, 29)

	if predicate != "TRIM(empresa_id) = ?" {
		t.Fatalf("expected text predicate, got %q", predicate)
	}
	if value, ok := arg.(string); !ok || value != "29" {
		t.Fatalf("expected string empresa id arg 29, got %#v", arg)
	}
}

func TestEmpresaDeleteBuildEmpresaIDPredicateUsesNumericArgForNumericColumns(t *testing.T) {
	predicate, arg := empresaDeleteBuildEmpresaIDPredicate(empresaDeleteCandidateTable{
		Name:              "productos",
		EmpresaIDDataType: "bigint",
		EmpresaIDUDTName:  "int8",
	}, 29)

	if predicate != "empresa_id = ?" {
		t.Fatalf("expected numeric predicate, got %q", predicate)
	}
	if value, ok := arg.(int64); !ok || value != 29 {
		t.Fatalf("expected int64 empresa id arg 29, got %#v", arg)
	}
}

func TestEmpresaDeleteEmpresaIDIsTextColumnCoversPostgresAliases(t *testing.T) {
	textColumns := []empresaDeleteCandidateTable{
		{EmpresaIDDataType: "character varying"},
		{EmpresaIDDataType: "character"},
		{EmpresaIDUDTName: "varchar"},
		{EmpresaIDUDTName: "bpchar"},
	}

	for _, column := range textColumns {
		if !empresaDeleteEmpresaIDIsTextColumn(column) {
			t.Fatalf("expected column metadata %#v to be treated as text", column)
		}
	}
}
