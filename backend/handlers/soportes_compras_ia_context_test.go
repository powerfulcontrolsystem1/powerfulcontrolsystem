package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSoportesComprasIAGetPropagaContextoYNoEjecutaDDL(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("soportes_compras_ia.go"))
	if err != nil {
		t.Fatalf("leer handler: %v", err)
	}
	body := string(source)
	for _, expected := range []string{
		"BuildEmpresaSoportesComprasIADashboardContext(r.Context(), dbEmp",
		"ListEmpresaSoportesComprasIARegistroContext(r.Context(), dbEmp",
		"ListEmpresaSoportesComprasIARetencionContext(r.Context(), dbEmp",
		"ListEmpresaSoportesComprasIAEventosContext(r.Context(), dbEmp",
		"UpdateEmpresaSoporteComprasIARegistroEstadoContext(r.Context(), dbEmp",
		"UpdateEmpresaSoporteComprasIAEstadoContext(r.Context(), dbEmp",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("falta propagacion cancelable: %s", expected)
		}
	}

	dbSource, err := os.ReadFile(filepath.Join("..", "db", "soportes_compras_ia.go"))
	if err != nil {
		t.Fatalf("leer repositorio: %v", err)
	}
	dbBody := string(dbSource)
	for _, expected := range []string{
		"func EmpresaSoportesComprasIASchemaReadyContext(ctx context.Context",
		"func ListEmpresaSoportesComprasIARegistroContext(ctx context.Context",
		"EmpresaSoportesComprasIASchemaReadyContext(ctx, dbConn)",
		"querySQLCompatContext(ctx, dbConn",
		"func UpdateEmpresaSoporteComprasIARegistroEstadoContext(ctx context.Context",
		"func UpdateEmpresaSoporteComprasIAEstadoContext(ctx context.Context",
		"dbConn.BeginTx(ctx, nil)",
	} {
		if !strings.Contains(dbBody, expected) {
			t.Fatalf("falta repositorio cancelable/readiness: %s", expected)
		}
	}
}
