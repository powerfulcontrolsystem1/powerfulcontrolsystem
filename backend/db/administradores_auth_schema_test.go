package db

import (
	"database/sql"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func openAdminSchemaSQLite(t *testing.T, name string) *sql.DB {
	t.Helper()
	dbConn, err := sql.Open("sqlite", t.TempDir()+"/"+name)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = dbConn.Close() })
	return dbConn
}

func TestEnsureAdministradoresAuthSchemaAddsMissingColumnsInSQLite(t *testing.T) {
	dbConn := openAdminSchemaSQLite(t, "admin_schema_columns.db")
	if _, err := dbConn.Exec(`CREATE TABLE administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		estado TEXT
	)`); err != nil {
		t.Fatalf("create administradores: %v", err)
	}

	if err := EnsureAdministradoresAuthSchema(dbConn); err != nil {
		t.Fatalf("ensure admin auth schema: %v", err)
	}

	columns := map[string]bool{}
	rows, err := dbConn.Query("PRAGMA table_info(administradores)")
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
	for _, required := range []string{"acepta_contrato", "telefono", "pais", "ciudad", "email_confirmado", "password_hash", "password_salt", "password_set", "password_reset_token", "password_reset_expira"} {
		if !columns[required] {
			t.Fatalf("expected column %s to exist after schema ensure", required)
		}
	}
}

func TestSetAdministradorPasswordRepairsMissingSecurityColumns(t *testing.T) {
	dbConn := openAdminSchemaSQLite(t, "admin_password_schema_fix.db")
	if _, err := dbConn.Exec(`CREATE TABLE administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT,
		photo TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		estado TEXT
	)`); err != nil {
		t.Fatalf("create administradores: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO administradores (email, name, role, photo, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), 'activo')`, "google_admin@empresa.com", "Google Admin", "administrador", ""); err != nil {
		t.Fatalf("seed admin: %v", err)
	}

	if err := SetAdministradorPassword(dbConn, "google_admin@empresa.com", "hash-demo", "salt-demo"); err != nil {
		t.Fatalf("set admin password: %v", err)
	}

	admin, err := GetAdminByEmailFull(dbConn, "google_admin@empresa.com")
	if err != nil {
		t.Fatalf("reload admin: %v", err)
	}
	if admin == nil {
		t.Fatal("expected admin after password setup")
	}
	if admin.PasswordSet != 1 {
		t.Fatalf("expected password_set=1, got %d", admin.PasswordSet)
	}
	if strings.TrimSpace(admin.PasswordHash) != "hash-demo" {
		t.Fatalf("unexpected password hash: %q", admin.PasswordHash)
	}
	if strings.TrimSpace(admin.PasswordSalt) != "salt-demo" {
		t.Fatalf("unexpected password salt: %q", admin.PasswordSalt)
	}
}
