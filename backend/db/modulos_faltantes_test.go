package db

import "testing"

func TestEmpresaGenericAllowedTablesIncluyeDocumentosTransaccionales(t *testing.T) {
	for _, table := range []string{"empresa_compras_documentos", "empresa_facturacion_documentos"} {
		if !isAllowedGenericTable(table) {
			t.Fatalf("tabla transaccional debe estar permitida para reportes genericos: %s", table)
		}
	}
	if isAllowedGenericTable("empresa_compras_documentos;DROP TABLE users") {
		t.Fatalf("tabla con SQL no debe estar permitida")
	}
}
