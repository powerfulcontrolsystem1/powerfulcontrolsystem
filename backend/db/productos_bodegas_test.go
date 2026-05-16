package db

import (
	"os"
	"strings"
	"testing"
)

func TestCreateBodegaUsaTimestampPostgres(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)
	start := strings.Index(src, "func CreateBodega(")
	if start < 0 {
		t.Fatal("no se encontro CreateBodega")
	}
	end := strings.Index(src[start:], "// GetBodegasByEmpresa")
	if end < 0 {
		t.Fatal("no se encontro limite de CreateBodega")
	}
	body := src[start : start+end]
	if strings.Contains(body, "datetime(") {
		t.Fatalf("CreateBodega no debe usar datetime() en runtime PostgreSQL: %s", body)
	}
	if !strings.Contains(body, "sqlNowExpr()") {
		t.Fatalf("CreateBodega debe usar sqlNowExpr() para fecha_creacion/fecha_actualizacion: %s", body)
	}
}
