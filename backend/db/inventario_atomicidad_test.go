package db

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSalidasInventarioUsanDescuentoCondicional(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate source")
	}
	body, err := os.ReadFile(filepath.Join(filepath.Dir(file), "productos.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(body)
	for _, operation := range []string{"TransferirProductoEntreBodegas", "RegistrarMovimientoInventario", "RegistrarCambioProducto"} {
		start := strings.Index(source, "func "+operation)
		if start < 0 {
			t.Fatalf("%s missing", operation)
		}
		fragment := source[start:]
		if end := strings.Index(fragment, "\nfunc "); end >= 0 {
			fragment = fragment[:end]
		}
		if !strings.Contains(fragment, "decrementarExistenciaTx") {
			t.Fatalf("%s must use the canonical conditional stock decrement", operation)
		}
	}
	start := strings.Index(source, "func decrementarExistenciaTx")
	if start < 0 {
		t.Fatal("decrementarExistenciaTx missing")
	}
	helper := source[start:]
	if end := strings.Index(helper, "\nfunc "); end >= 0 {
		helper = helper[:end]
	}
	if !strings.Contains(helper, "cantidad >= ?") || !strings.Contains(helper, "RowsAffected") {
		t.Fatal("canonical stock decrement must be conditional and verify the affected row")
	}
}

func TestInventarioResumenCuentaCatalogosSinCargarListasCompletas(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func GetInventarioResumenByEmpresa")
	if start < 0 {
		t.Fatal("GetInventarioResumenByEmpresa missing")
	}
	fragment := source[start:]
	if end := strings.Index(fragment, "\nfunc "); end >= 0 {
		fragment = fragment[:end]
	}
	helperStart := strings.Index(source, "func getInventarioCatalogTotalsByEmpresa")
	if helperStart < 0 {
		t.Fatal("getInventarioCatalogTotalsByEmpresa missing")
	}
	helper := source[helperStart:]
	if end := strings.Index(helper, "\nfunc "); end >= 0 {
		helper = helper[:end]
	}
	if !strings.Contains(fragment, "getInventarioCatalogTotalsByEmpresa(dbConn, empresaID, &resumen)") {
		t.Fatal("inventory summary must use the bounded catalog aggregate helper")
	}
	for _, table := range []string{"productos p", "bodegas b", "servicios s", "categorias_productos c"} {
		if !strings.Contains(helper, "FROM "+table+" WHERE") {
			t.Errorf("inventory summary missing tenant-scoped count for %s", table)
		}
	}
	for _, field := range []string{"ProductosTotal", "ProductosPorVencer", "BodegasTotal", "ServiciosTotal", "CategoriasTotal"} {
		if !strings.Contains(helper, "&resumen."+field) {
			t.Errorf("inventory summary does not scan %s", field)
		}
	}
	if strings.Count(fragment+helper, "empresa_id = ?") < 8 {
		t.Fatalf("inventory summary must keep every catalog/stock aggregate tenant-scoped")
	}
}

func TestEntradasInventarioUsanUpsertAtomico(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func upsertExistenciaTx")
	if start < 0 {
		t.Fatal("upsertExistenciaTx missing")
	}
	fragment := source[start:]
	if end := strings.Index(fragment, "\nfunc "); end >= 0 {
		fragment = fragment[:end]
	}
	for _, required := range []string{
		"INSERT INTO inventario_existencias",
		"ON CONFLICT (empresa_id, producto_id, bodega_id) DO UPDATE",
		"inventario_existencias.cantidad + EXCLUDED.cantidad",
	} {
		if !strings.Contains(fragment, required) {
			t.Errorf("atomic stock upsert missing %q", required)
		}
	}
	if !strings.Contains(source, "CREATE UNIQUE INDEX IF NOT EXISTS ux_existencias_empresa_prod_bodega") {
		t.Fatal("inventory stock upsert requires a tenant/product/warehouse unique index")
	}
}

func TestListadoProductosAcotaAgregadoDeStockPorEmpresa(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func GetProductosByEmpresa")
	if start < 0 {
		t.Fatal("GetProductosByEmpresa missing")
	}
	fragment := source[start:]
	if end := strings.Index(fragment, "\nfunc "); end >= 0 {
		fragment = fragment[:end]
	}
	for _, required := range []string{
		"FROM inventario_existencias\n\t\tWHERE empresa_id = ?",
		"args := []interface{}{empresaID, empresaID}",
		"LIMIT ? OFFSET ?",
	} {
		if !strings.Contains(fragment, required) {
			t.Errorf("scalable tenant product listing missing %q", required)
		}
	}
	migrationRaw, err := os.ReadFile("productos_inventario_migration.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(migrationRaw), "CREATE INDEX IF NOT EXISTS ix_productos_empresa_estado_id") {
		t.Fatal("product listing requires tenant/state/id index")
	}
}

func TestConsumoCostoPEPSBloqueaCapasYConservaTenant(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func consumirCostoLotesPEPSTx")
	if start < 0 {
		t.Fatal("consumirCostoLotesPEPSTx missing")
	}
	fragment := source[start:]
	if end := strings.Index(fragment, "\nfunc "); end >= 0 {
		fragment = fragment[:end]
	}
	for _, required := range []string{
		"ORDER BY COALESCE(fecha_lote, fecha_creacion, '') ASC, id ASC\n\t\tFOR UPDATE",
		"WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ? AND id = ?",
	} {
		if !strings.Contains(fragment, required) {
			t.Errorf("atomic PEPS cost consumption missing %q", required)
		}
	}
	if strings.Count(fragment, "WHERE empresa_id = ? AND producto_id = ? AND bodega_id = ? AND id = ?") < 2 {
		t.Fatal("every PEPS cost-layer update must retain tenant/product/warehouse ownership")
	}
}
