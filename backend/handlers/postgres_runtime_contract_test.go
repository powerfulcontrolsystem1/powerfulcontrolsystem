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
