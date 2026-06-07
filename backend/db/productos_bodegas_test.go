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

func TestEnsureEmpresaBodega1DefaultEsIdempotenteYSinStockDemo(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)
	start := strings.Index(src, "func EnsureEmpresaBodega1(")
	if start < 0 {
		t.Fatal("no se encontro EnsureEmpresaBodega1")
	}
	end := strings.Index(src[start:], "// GetBodegasByEmpresa")
	if end < 0 {
		t.Fatal("no se encontro limite de EnsureEmpresaBodega1")
	}
	body := src[start : start+end]
	for _, required := range []string{`"Bodega 1"`, "getEmpresaBodegaIDByNombre", "SetBodegaEstado", "CreateBodega"} {
		if !strings.Contains(body, required) {
			t.Fatalf("EnsureEmpresaBodega1 debe conservar %s para idempotencia: %s", required, body)
		}
	}
	for _, forbidden := range []string{"inventario_existencias", "CreateProducto", "InsertProducto", "upsertExistencia"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("EnsureEmpresaBodega1 no debe crear stock/productos demo; encontro %s en: %s", forbidden, body)
		}
	}
	if !strings.Contains(src, "func ApplyDefaultBodega1ToExistingEmpresas(") {
		t.Fatal("debe existir backfill independiente para empresas existentes")
	}
	if !strings.Contains(src, "20260607_bodega_1_default") {
		t.Fatal("el backfill debe tener version nueva para ejecutarse aunque migraciones anteriores ya esten aplicadas")
	}
}

func TestTransferirProductoEntreBodegasUsaSQLCompatPostgres(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatalf("read productos.go: %v", err)
	}
	src := string(raw)
	start := strings.Index(src, "func TransferirProductoEntreBodegas(")
	if start < 0 {
		t.Fatal("no se encontro TransferirProductoEntreBodegas")
	}
	end := strings.Index(src[start:], "// GetMovimientosByEmpresa")
	if end < 0 {
		t.Fatal("no se encontro limite de TransferirProductoEntreBodegas")
	}
	body := src[start : start+end]
	for _, forbidden := range []string{"tx.QueryRow(", "tx.Exec("} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("TransferirProductoEntreBodegas debe usar wrappers SQLCompat en runtime PostgreSQL; encontro %s en: %s", forbidden, body)
		}
	}
	for _, required := range []string{"queryRowTxSQLCompat", "execTxSQLCompat", "insertMovimientoTx"} {
		if !strings.Contains(body, required) {
			t.Fatalf("TransferirProductoEntreBodegas debe conservar %s en la transaccion: %s", required, body)
		}
	}
}
