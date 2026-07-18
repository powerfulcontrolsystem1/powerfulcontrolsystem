package db

import (
	"os"
	"strings"
	"testing"
)

func TestAuditoriaPostgresNormalizesTimestampExpressions(t *testing.T) {
	for _, file := range []string{"auditoria_empresa.go", "auditoria_super.go"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		src := string(raw)
		for _, want := range []string{
			"COALESCE(CAST(fecha_evento AS TEXT), CAST(fecha_creacion AS TEXT), '')",
			"COALESCE(CAST(fecha_evento AS TEXT), '')",
		} {
			if !strings.Contains(src, want) {
				t.Fatalf("%s must normalize timestamp expression %q", file, want)
			}
		}
	}

	raw, err := os.ReadFile("auditoria_empresa.go")
	if err != nil {
		t.Fatalf("read auditoria_empresa.go: %v", err)
	}
	if !strings.Contains(string(raw), "COALESCE(NULLIF(?, ''), CAST(CURRENT_TIMESTAMP AS TEXT))") {
		t.Fatal("empresa audit insert must avoid mixing text and PostgreSQL timestamps")
	}
}
