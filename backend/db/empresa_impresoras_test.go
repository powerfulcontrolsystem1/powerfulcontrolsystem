package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openEmpresaImpresorasTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "empresa_impresoras_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = dbConn.Close() })
	return dbConn
}

func ensureTestProductosTable(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	if _, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS productos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		estado TEXT DEFAULT 'activo'
	)`); err != nil {
		t.Fatalf("create productos table: %v", err)
	}
}

func TestEmpresaImpresorasResolvePriorities(t *testing.T) {
	dbConn := openEmpresaImpresorasTestDB(t)
	ensureTestProductosTable(t, dbConn)
	if err := EnsureEmpresaImpresorasSchema(dbConn); err != nil {
		t.Fatalf("ensure empresa_impresoras schema: %v", err)
	}

	impAID, err := UpsertEmpresaImpresora(dbConn, EmpresaImpresora{
		EmpresaID:        23,
		Codigo:           "COCINA_A",
		Nombre:           "Impresora Cocina A",
		TipoConexion:     "red",
		FormatoImpresion: "pos",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("upsert printer A: %v", err)
	}
	impBID, err := UpsertEmpresaImpresora(dbConn, EmpresaImpresora{
		EmpresaID:        23,
		Codigo:           "CAJA_B",
		Nombre:           "Impresora Caja B",
		TipoConexion:     "usb",
		FormatoImpresion: "carta",
		EsPredeterminada: true,
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("upsert printer B: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO productos (empresa_id, nombre, estado) VALUES (?,?,?)`, 23, "Hamburguesa", "activo"); err != nil {
		t.Fatalf("insert producto: %v", err)
	}

	if _, err := UpsertEmpresaImpresoraFuncionalidad(dbConn, EmpresaImpresoraFuncionalidad{
		EmpresaID:     23,
		Funcionalidad: "orden_servicio",
		ImpresoraID:   impAID,
		Estado:        "activo",
	}); err != nil {
		t.Fatalf("upsert funcionalidad: %v", err)
	}

	if _, err := UpsertEmpresaImpresoraProducto(dbConn, EmpresaImpresoraProducto{
		EmpresaID:   23,
		ProductoID:  1,
		ImpresoraID: impBID,
		Estado:      "activo",
	}); err != nil {
		t.Fatalf("upsert producto mapping: %v", err)
	}

	resProducto, err := ResolveEmpresaImpresora(dbConn, 23, "orden_servicio", 1)
	if err != nil {
		t.Fatalf("resolve producto: %v", err)
	}
	if resProducto == nil {
		t.Fatalf("expected resolver result for producto")
	}
	if resProducto.Fuente != "producto" {
		t.Fatalf("expected source producto, got %s", resProducto.Fuente)
	}
	if resProducto.Impresora.ID != impBID {
		t.Fatalf("expected printer B for producto mapping, got %d", resProducto.Impresora.ID)
	}

	resFunc, err := ResolveEmpresaImpresora(dbConn, 23, "orden_servicio", 999)
	if err != nil {
		t.Fatalf("resolve funcionalidad: %v", err)
	}
	if resFunc == nil {
		t.Fatalf("expected resolver result for funcionalidad")
	}
	if resFunc.Fuente != "funcionalidad" {
		t.Fatalf("expected source funcionalidad, got %s", resFunc.Fuente)
	}
	if resFunc.Impresora.ID != impAID {
		t.Fatalf("expected printer A for funcionalidad mapping, got %d", resFunc.Impresora.ID)
	}

	resDefault, err := ResolveEmpresaImpresora(dbConn, 23, "factura_caja", 999)
	if err != nil {
		t.Fatalf("resolve default: %v", err)
	}
	if resDefault == nil {
		t.Fatalf("expected resolver default result")
	}
	if resDefault.Fuente != "predeterminada" {
		t.Fatalf("expected source predeterminada, got %s", resDefault.Fuente)
	}
	if resDefault.Impresora.ID != impBID {
		t.Fatalf("expected default printer B, got %d", resDefault.Impresora.ID)
	}
}

func TestEmpresaImpresorasDefaultConsistency(t *testing.T) {
	dbConn := openEmpresaImpresorasTestDB(t)
	ensureTestProductosTable(t, dbConn)
	if err := EnsureEmpresaImpresorasSchema(dbConn); err != nil {
		t.Fatalf("ensure empresa_impresoras schema: %v", err)
	}

	impAID, err := UpsertEmpresaImpresora(dbConn, EmpresaImpresora{
		EmpresaID:        51,
		Codigo:           "IMP_A",
		Nombre:           "Impresora A",
		FormatoImpresion: "pos",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("upsert printer A: %v", err)
	}

	list1, err := ListEmpresaImpresorasByEmpresa(dbConn, 51, true)
	if err != nil {
		t.Fatalf("list printers step1: %v", err)
	}
	if len(list1) != 1 || !list1[0].EsPredeterminada {
		t.Fatalf("first printer must become default automatically")
	}

	impBID, err := UpsertEmpresaImpresora(dbConn, EmpresaImpresora{
		EmpresaID:        51,
		Codigo:           "IMP_B",
		Nombre:           "Impresora B",
		FormatoImpresion: "carta",
		EsPredeterminada: true,
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("upsert printer B: %v", err)
	}
	if impAID == impBID {
		t.Fatalf("expected different ids")
	}

	list2, err := ListEmpresaImpresorasByEmpresa(dbConn, 51, true)
	if err != nil {
		t.Fatalf("list printers step2: %v", err)
	}
	var defaultID int64
	for _, item := range list2 {
		if item.EsPredeterminada {
			defaultID = item.ID
		}
	}
	if defaultID != impBID {
		t.Fatalf("expected printer B as default, got %d", defaultID)
	}

	if err := SetEmpresaImpresoraEstado(dbConn, 51, impBID, "inactivo", "tester"); err != nil {
		t.Fatalf("deactivate printer B: %v", err)
	}
	list3, err := ListEmpresaImpresorasByEmpresa(dbConn, 51, true)
	if err != nil {
		t.Fatalf("list printers step3: %v", err)
	}
	defaultID = 0
	for _, item := range list3 {
		if item.EsPredeterminada {
			defaultID = item.ID
		}
	}
	if defaultID != impAID {
		t.Fatalf("expected fallback default printer A, got %d", defaultID)
	}
}
