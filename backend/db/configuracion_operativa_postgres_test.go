package db

import (
	"os"
	"strings"
	"testing"
)

func TestConfiguracionOperativaWritesUsePostgresReturningID(t *testing.T) {
	raw, err := os.ReadFile("configuracion_operativa.go")
	if err != nil {
		t.Fatalf("read configuracion_operativa.go: %v", err)
	}
	src := string(raw)

	for _, fn := range []string{
		"func UpsertEmpresaConfiguracionOperativa(",
		"func UpsertEmpresaConfiguracionOperativaRol(",
		"func UpsertEmpresaConfiguracionOperativaPolitica(",
		"func CreateEmpresaConfiguracionOperativaHistorialSnapshot(",
	} {
		body := extractConfiguracionOperativaFunctionForTest(t, src, fn)
		if strings.Contains(body, "LastInsertId(") {
			t.Fatalf("%s no debe depender de LastInsertId en PostgreSQL: %s", fn, body)
		}
		if !strings.Contains(body, "QueryRowCompat") {
			t.Fatalf("%s debe usar QueryRowCompat para rebind PostgreSQL: %s", fn, body)
		}
		if !strings.Contains(body, "RETURNING id") {
			t.Fatalf("%s debe retornar id con RETURNING id: %s", fn, body)
		}
		if !strings.Contains(body, "sqlNowExpr()") {
			t.Fatalf("%s debe usar sqlNowExpr para fechas runtime: %s", fn, body)
		}
	}
}

func extractConfiguracionOperativaFunctionForTest(t *testing.T, src, startMarker string) string {
	t.Helper()

	start := strings.Index(src, startMarker)
	if start < 0 {
		t.Fatalf("no se encontro %s", startMarker)
	}
	rest := src[start+len(startMarker):]
	next := strings.Index(rest, "\nfunc ")
	if next < 0 {
		return src[start:]
	}
	return src[start : start+len(startMarker)+next]
}
