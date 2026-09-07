package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestReportesProgramacionUsesPostgresReturningIDs(t *testing.T) {
	raw, err := os.ReadFile("reportes_programacion.go")
	if err != nil {
		t.Fatalf("read reportes_programacion.go: %v", err)
	}
	src := string(raw)
	if strings.Contains(src, "LastInsertId(") {
		t.Fatal("reportes programados no deben depender de LastInsertId en PostgreSQL")
	}
	if got := strings.Count(src, "RETURNING id"); got < 3 {
		t.Fatalf("reportes programados deben recuperar tres inserciones con RETURNING id, got %d", got)
	}
}

func TestFacturacionReconciliacionHasNoSQLiteFallback(t *testing.T) {
	raw, err := os.ReadFile("facturacion_electronica.go")
	if err != nil {
		t.Fatalf("read facturacion_electronica.go: %v", err)
	}
	src := string(raw)
	if strings.Contains(strings.ToLower(src), "no such table: clientes") {
		t.Fatal("facturación no debe mantener rutas runtime específicas de SQLite")
	}
}

func TestAdminCompanyAndUserFlowsPropagateRequestContext(t *testing.T) {
	checks := map[string][]string{
		"system_empresas_handlers.go": {
			"buildEmpresaImpactoDesactivacion(r.Context(), dbEmp, dbSuper, id)",
			"QueryRowContext(ctx",
		},
		"usuarios_empresa.go": {
			"resolveRolNombreValidoParaEmpresa(r.Context()",
			"QueryRowContext(ctx",
		},
	}
	for filename, required := range checks {
		raw, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("read %s: %v", filename, err)
		}
		src := string(raw)
		for _, contract := range required {
			if !strings.Contains(src, contract) {
				t.Fatalf("%s debe conservar %q", filename, contract)
			}
		}
	}
}
