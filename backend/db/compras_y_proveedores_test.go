package db

import (
	"database/sql"
	"testing"
	_ "modernc.org/sqlite"
)

// TestEnsureEmpresasComprasSchema verifica que el esquema del módulo interdependiente
// de compras y proveedores de ERP se instancie y actualice correctamente sin romper.
func TestEnsureEmpresasComprasSchema(t *testing.T) {
	dbConn, err := sql.Open("sqlite", t.TempDir()+"/test_empresas_compras.db")
	if err != nil {
		t.Fatalf("no se pudo abrir bd: %v", err)
	}
	defer dbConn.Close()

	if err := EnsureEmpresasComprasSchema(dbConn); err != nil {
		t.Fatalf("Esperaba éxito asegurando esquema de compras, obtuve: %v", err)
	}

	// Ejecución una segunda vez asegura que es idempotente y los `ensureColumnIfMissing` no rompen
	if err := EnsureEmpresasComprasSchema(dbConn); err != nil {
		t.Fatalf("Esperaba idempotencia en schema de compras, obtuve: %v", err)
	}

	// Inserciones rápidas verificando la estructura.
	_, err = dbConn.Exec(`INSERT INTO empresa_proveedores (empresa_id, nombre_comercial, razon_social) VALUES (1, 'Distribuidor Test', 'DIST SAS')`)
	if err != nil {
		t.Fatalf("No se pudo insertar proveedor de prueba: %v", err)
	}

	var rowCount int
	if err := dbConn.QueryRow("SELECT COUNT(*) FROM empresa_proveedores WHERE empresa_id = 1").Scan(&rowCount); err != nil {
		t.Fatalf("Error verificando conteo de proveedores: %v", err)
	}
	if rowCount != 1 {
		t.Errorf("Se esperaba 1 registro proveedor, se encontraron %d", rowCount)
	}

	// Verify columns exist gracefully after idempotence check
	_, err = dbConn.Exec(`INSERT INTO empresa_ordenes_compra (empresa_id, proveedor_id, numero_orden, bodega_destino_id) VALUES (1, 1, 'OC-XYZ', 10)`)
	if err != nil {
		t.Fatalf("No se pudo insertar orden de compra de prueba: %v", err)
	}
}
