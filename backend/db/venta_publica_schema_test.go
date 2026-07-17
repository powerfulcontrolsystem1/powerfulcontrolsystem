package db

import (
	"os"
	"strings"
	"testing"
)

func TestVentaPublicaSchemaUsesPostgresTimestampTypes(t *testing.T) {
	content, err := os.ReadFile("venta_publica.go")
	if err != nil {
		t.Fatalf("read venta_publica.go: %v", err)
	}
	source := string(content)
	if strings.Contains(source, "DATETIME") {
		t.Fatal("venta publica schema must not use SQLite DATETIME in PostgreSQL DDL")
	}
	if strings.Count(source, "TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP") < 6 {
		t.Fatal("expected explicit PostgreSQL timestamp columns in venta publica schema")
	}
}
