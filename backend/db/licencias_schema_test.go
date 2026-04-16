package db

import (
	"database/sql"
	"math"
	"testing"

	_ "modernc.org/sqlite"
)

func openLicenciasSchemaSQLite(t *testing.T, name string) *sql.DB {
	t.Helper()
	dbConn, err := sql.Open("sqlite", t.TempDir()+"/"+name)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = dbConn.Close() })
	return dbConn
}

func TestEnsureLicenciasSchemaAddsValorInSQLite(t *testing.T) {
	dbConn := openLicenciasSchemaSQLite(t, "licencias_schema_columns.db")
	if _, err := dbConn.Exec(`CREATE TABLE licencias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tipo_id INTEGER,
		nombre TEXT,
		descripcion TEXT,
		duracion_dias INTEGER,
		fecha_creacion TEXT,
		activo INTEGER DEFAULT 1
	)`); err != nil {
		t.Fatalf("create licencias: %v", err)
	}

	if err := EnsureLicenciasSchema(dbConn); err != nil {
		t.Fatalf("ensure licencias schema: %v", err)
	}

	columns := map[string]bool{}
	rows, err := dbConn.Query("PRAGMA table_info(licencias)")
	if err != nil {
		t.Fatalf("pragma table_info: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name string
		var ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan pragma row: %v", err)
		}
		columns[name] = true
	}

	for _, required := range []string{"valor", "modulos_habilitados", "super_rol_habilitado", "fecha_actualizacion", "estado"} {
		if !columns[required] {
			t.Fatalf("expected column %s to exist after schema ensure", required)
		}
	}
}

func TestCreateAndUpdateLicenciaRepairMissingValorColumn(t *testing.T) {
	dbConn := openLicenciasSchemaSQLite(t, "licencias_schema_repair.db")
	if _, err := dbConn.Exec(`CREATE TABLE licencias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tipo_id INTEGER,
		nombre TEXT,
		descripcion TEXT,
		duracion_dias INTEGER,
		fecha_creacion TEXT,
		activo INTEGER DEFAULT 1
	)`); err != nil {
		t.Fatalf("create licencias legacy schema: %v", err)
	}

	id, err := CreateLicencia(dbConn, 2, "Plan Pro", "Licencia corregida", 149900.75, 30, "ventas,clientes", 1)
	if err != nil {
		t.Fatalf("create licencia with schema repair: %v", err)
	}

	lic, err := GetLicenciaByID(dbConn, id)
	if err != nil {
		t.Fatalf("get licencia after create: %v", err)
	}
	if lic == nil {
		t.Fatal("expected licencia after create")
	}
	if math.Abs(lic.Valor-149900.75) > 0.0001 {
		t.Fatalf("expected valor 149900.75, got %v", lic.Valor)
	}

	if err := UpdateLicencia(dbConn, id, 2, "Plan Pro Plus", "Licencia actualizada", 189500.5, 45, "ventas,clientes,finanzas", 1); err != nil {
		t.Fatalf("update licencia with schema repair: %v", err)
	}

	updated, err := GetLicenciaByID(dbConn, id)
	if err != nil {
		t.Fatalf("get licencia after update: %v", err)
	}
	if updated == nil {
		t.Fatal("expected licencia after update")
	}
	if math.Abs(updated.Valor-189500.5) > 0.0001 {
		t.Fatalf("expected valor actualizado 189500.5, got %v", updated.Valor)
	}
	if updated.DuracionDias != 45 {
		t.Fatalf("expected duracion_dias 45, got %d", updated.DuracionDias)
	}
}