package db

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaAsistenciaRuntimeUsesPostgresCompatibleSQL(t *testing.T) {
	raw, err := os.ReadFile("asistencia_empleados.go")
	if err != nil {
		t.Fatalf("read asistencia_empleados.go: %v", err)
	}
	src := string(raw)
	body := extractAsistenciaRuntimeForTest(t, src)

	for _, forbidden := range []string{"dbConn.Exec(", "dbConn.Query(", "dbConn.QueryRow(", "pcs_ts('now'"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("asistencia runtime no debe usar %s en PostgreSQL: %s", forbidden, body)
		}
	}
	for _, required := range []string{"queryRowSQLCompat", "querySQLCompat", "execSQLCompat", "insertSQLCompat", "sqlNowExpr()"} {
		if !strings.Contains(body, required) {
			t.Fatalf("asistencia runtime debe usar %s para compatibilidad PostgreSQL: %s", required, body)
		}
	}
}

func extractAsistenciaRuntimeForTest(t *testing.T, src string) string {
	t.Helper()

	start := strings.Index(src, "func GetEmpresaAsistenciaConfiguracion(")
	if start < 0 {
		t.Fatal("no se encontro GetEmpresaAsistenciaConfiguracion")
	}
	end := strings.Index(src[start:], "func normalizeAsistenciaDate(")
	if end < 0 {
		t.Fatal("no se encontro limite normalizeAsistenciaDate")
	}
	return src[start : start+end]
}
