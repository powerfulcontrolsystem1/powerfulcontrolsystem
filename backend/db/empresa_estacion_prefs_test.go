package db

import (
	"encoding/json"
	"database/sql"
	"reflect"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestEmpresaEstacionPrefs_UpsertGetList(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaEstacionPrefsSchema(dbConn); err != nil {
		t.Fatalf("ensure prefs schema: %v", err)
	}

	payload := EmpresaEstacionPref{
		EmpresaID:      1,
		EstacionID:     0,
		Clave:          "estaciones_config",
		Valor:          `{"cantidad":2,"estaciones":[{"id":1,"nombre":"A","venta_simple_habilitada":true},{"id":2,"nombre":"B","venta_simple_habilitada":false}]}`,
		UsuarioCreador: "test",
		Estado:         "activo",
	}

	id, err := UpsertEmpresaEstacionPref(dbConn, payload)
	if err != nil {
		t.Fatalf("upsert error: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	got, err := GetEmpresaEstacionPref(dbConn, 1, 0, "estaciones_config")
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if got == nil {
		t.Fatal("expected pref, got nil")
	}

	var a interface{}
	var b interface{}
	if err := json.Unmarshal([]byte(got.Valor), &a); err != nil {
		t.Fatalf("unmarshal got valor: %v", err)
	}
	if err := json.Unmarshal([]byte(payload.Valor), &b); err != nil {
		t.Fatalf("unmarshal payload valor: %v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("valor mismatch: expected %v, got %v", b, a)
	}

	rows, err := ListEmpresaEstacionPrefs(dbConn, 1, 0, false)
	if err != nil {
		t.Fatalf("list error: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected list length > 0")
	}

	// Update valor
	payload.Valor = `{"cantidad":2,"estaciones":[{"id":1,"nombre":"A","venta_simple_habilitada":false}]}`
	id2, err := UpsertEmpresaEstacionPref(dbConn, payload)
	if err != nil {
		t.Fatalf("upsert update error: %v", err)
	}
	if id2 != id {
		t.Fatalf("expected same id after update: before=%d after=%d", id, id2)
	}

	got2, err := GetEmpresaEstacionPref(dbConn, 1, 0, "estaciones_config")
	if err != nil {
		t.Fatalf("get after update error: %v", err)
	}
	if got2 == nil {
		t.Fatal("expected pref after update, got nil")
	}
	var c interface{}
	if err := json.Unmarshal([]byte(got2.Valor), &c); err != nil {
		t.Fatalf("unmarshal got2 valor: %v", err)
	}
	var d interface{}
	if err := json.Unmarshal([]byte(payload.Valor), &d); err != nil {
		t.Fatalf("unmarshal payload valor after update: %v", err)
	}
	if !reflect.DeepEqual(c, d) {
		t.Fatalf("valor mismatch after update: expected %v, got %v", d, c)
	}
}

func TestEmpresaEstacionPrefs_UpsertSinEstadoSigueActivoEnListado(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaEstacionPrefsSchema(dbConn); err != nil {
		t.Fatalf("ensure prefs schema: %v", err)
	}

	payload := EmpresaEstacionPref{
		EmpresaID:      88,
		EstacionID:     0,
		Clave:          "estaciones_config",
		Valor:          `{"cantidad":10,"estaciones":[{"id":1,"nombre":"E1"}]}`,
		UsuarioCreador: "test",
		Estado:         "",
	}

	if _, err := UpsertEmpresaEstacionPref(dbConn, payload); err != nil {
		t.Fatalf("upsert sin estado: %v", err)
	}

	rows, err := ListEmpresaEstacionPrefs(dbConn, 88, 0, false)
	if err != nil {
		t.Fatalf("list prefs activas: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected estaciones_config visible as activo when estado is empty")
	}
	if rows[0].Estado != "activo" {
		t.Fatalf("expected estado activo normalizado, got %q", rows[0].Estado)
	}
}

func TestSyncEmpresaEstacionCarritosCreatesAndUpdatesLinkedDefaults(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	ensureClientesTableForStationSyncTest(t, dbConn)

	result, err := SyncEmpresaEstacionCarritos(dbConn, 55, `{"cantidad":2,"estaciones":[{"id":1,"nombre":"Mesa 1"},{"id":2,"nombre":"Mesa 2"}]}`, "test")
	if err != nil {
		t.Fatalf("sync initial station carritos: %v", err)
	}
	if result.Created != 2 {
		t.Fatalf("expected 2 created carritos, got %+v", result)
	}

	rows, err := GetCarritosCompraByEmpresa(dbConn, 55, true, "")
	if err != nil {
		t.Fatalf("list synced carritos: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 synced carritos, got %d", len(rows))
	}

	byCode := map[string]CarritoCompra{}
	for _, item := range rows {
		byCode[item.Codigo] = item
	}
	if got := byCode["EST-55-1"]; got.ReferenciaExterna != "ESTACION_1" || got.Nombre != "Mesa 1" {
		t.Fatalf("unexpected carrito for station 1: %+v", got)
	}
	if got := byCode["EST-55-2"]; got.ReferenciaExterna != "ESTACION_2" || got.Nombre != "Mesa 2" {
		t.Fatalf("unexpected carrito for station 2: %+v", got)
	}
	for _, item := range rows {
		if strings.ToLower(strings.TrimSpace(item.Estado)) != "inactivo" {
			t.Fatalf("expected carrito estado inactivo, got %+v", item)
		}
		if strings.ToLower(strings.TrimSpace(item.EstadoCarrito)) != "cerrado" {
			t.Fatalf("expected carrito estado_carrito cerrado, got %+v", item)
		}
	}

	result, err = SyncEmpresaEstacionCarritos(dbConn, 55, `{"cantidad":2,"estaciones":[{"id":1,"nombre":"Mesa 1"},{"id":2,"nombre":"Mesa VIP 2"}]}`, "test")
	if err != nil {
		t.Fatalf("sync updated station carritos: %v", err)
	}
	if result.Created != 0 {
		t.Fatalf("expected 0 created carritos on second sync, got %+v", result)
	}

	rows, err = GetCarritosCompraByEmpresa(dbConn, 55, true, "")
	if err != nil {
		t.Fatalf("list synced carritos after update: %v", err)
	}
	byCode = map[string]CarritoCompra{}
	for _, item := range rows {
		byCode[item.Codigo] = item
	}
	if got := byCode["EST-55-2"]; got.Nombre != "Mesa VIP 2" {
		t.Fatalf("expected updated station name in linked carrito, got %+v", got)
	}
	if len(rows) != 2 {
		t.Fatalf("expected no duplicate carritos after second sync, got %d", len(rows))
	}
}

func ensureClientesTableForStationSyncTest(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	if _, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS clientes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre_razon_social TEXT
	)`); err != nil {
		t.Fatalf("create clientes table for sync test: %v", err)
	}
}
