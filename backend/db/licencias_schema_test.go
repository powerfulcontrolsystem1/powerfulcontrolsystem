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

func TestUpdateLicenciaRepairsMissingFechaActualizacionColumn(t *testing.T) {
	dbConn := openLicenciasSchemaSQLite(t, "licencias_missing_fecha_actualizacion.db")
	if _, err := dbConn.Exec(`CREATE TABLE licencias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tipo_id INTEGER,
		nombre TEXT,
		descripcion TEXT,
		valor REAL DEFAULT 0,
		duracion_dias INTEGER DEFAULT 0,
		modulos_habilitados TEXT,
		super_rol_habilitado INTEGER DEFAULT 0,
		fecha_creacion TEXT,
		activo INTEGER DEFAULT 1
	)`); err != nil {
		t.Fatalf("create licencias legacy schema without fecha_actualizacion: %v", err)
	}

	res, err := dbConn.Exec(`INSERT INTO licencias (tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, activo)
		VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), 1)`, 1, "Plan Basico", "Licencia legacy", 99000.0, 30, "ventas", 0)
	if err != nil {
		t.Fatalf("insert legacy licencia: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	if err := UpdateLicencia(dbConn, id, 1, "Plan Basico Ajustado", "Licencia legacy actualizada", 123456.78, 60, "ventas,finanzas", 1); err != nil {
		t.Fatalf("update licencia legacy without fecha_actualizacion: %v", err)
	}
	if err := SetLicenciaActivo(dbConn, id, 0); err != nil {
		t.Fatalf("set licencia activo on legacy schema without fecha_actualizacion: %v", err)
	}

	updated, err := GetLicenciaByID(dbConn, id)
	if err != nil {
		t.Fatalf("get licencia after legacy update: %v", err)
	}
	if updated == nil {
		t.Fatal("expected licencia after legacy update")
	}
	if math.Abs(updated.Valor-123456.78) > 0.0001 {
		t.Fatalf("expected valor actualizado 123456.78, got %v", updated.Valor)
	}
	if updated.DuracionDias != 60 {
		t.Fatalf("expected duracion_dias 60, got %d", updated.DuracionDias)
	}
	if updated.Activo != 0 {
		t.Fatalf("expected activo 0, got %d", updated.Activo)
	}

}