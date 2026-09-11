package db

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestAdminEmpresaAuthorizationPostgresCurrentMembership(t *testing.T) {
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	admin, err := sql.Open(PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("admin_access_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer admin.Close()
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("clean fixture: %v", err)
		}
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	conn, err := sql.Open(PostgresCompatDriverName(), u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	exec := func(statement string) {
		t.Helper()
		if _, err := conn.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	for _, statement := range []string{
		`CREATE TABLE administradores (id BIGINT PRIMARY KEY, email TEXT, name TEXT, role TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE empresas (id BIGINT PRIMARY KEY, empresa_id BIGINT, usuario_creador TEXT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE users (empresa_id BIGINT, email TEXT, role TEXT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE tipos_de_empresas (id BIGINT PRIMARY KEY, nombre TEXT)`,
		`CREATE TABLE tipo_empresa_preconfiguraciones (id BIGINT PRIMARY KEY, tipo_empresa_id BIGINT, enabled INT, nombre TEXT, descripcion TEXT, config_json TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT)`,
		`CREATE TABLE admin_principal_delegaciones (admin_email TEXT, principal_email TEXT, estado TEXT, fecha_revocada TEXT)`,
		`CREATE TABLE admin_empresa_compartida (id BIGINT PRIMARY KEY, empresa_id BIGINT, admin_email TEXT, compartido_por_email TEXT, invitacion_id BIGINT, nivel_acceso TEXT, modulos_permitidos TEXT, puede_compartir BOOL, fecha_aceptada TEXT, fecha_revocada TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT)`,
		`INSERT INTO administradores (id,email,role,usuario_creador) VALUES (1,'owner@example.invalid','administrador',''),(2,'delegate@example.invalid','administrador',''),(3,'super@example.invalid','super_administrador',''),(4,'child@example.invalid','administrador','owner@example.invalid'),(5,'other@example.invalid','administrador','')`,
		`INSERT INTO empresas (id,empresa_id,usuario_creador) VALUES (101,101,'owner@example.invalid'),(202,202,'other@example.invalid'),(303,303,'operational@example.invalid')`,
		`INSERT INTO users (empresa_id,email,role) VALUES (101,'owner@example.invalid','super_administrador'),(303,'operational@example.invalid','super_administrador')`,
		`INSERT INTO tipos_de_empresas (id,nombre) VALUES (1,'QA')`,
		`INSERT INTO tipo_empresa_preconfiguraciones (id,tipo_empresa_id,enabled,nombre,config_json,estado) VALUES (1,1,1,'QA','{}','activo')`,
	} {
		exec(statement)
	}
	assertAccess := func(email string, company int64, want bool) {
		t.Helper()
		allowed, err := CanAdminAccessEmpresaIA(conn, conn, email, company)
		if err != nil {
			t.Fatal(err)
		}
		if allowed != want {
			t.Fatalf("access company=%d account=%s = %v want=%v", company, email, allowed, want)
		}
	}
	assertAccess("owner@example.invalid", 101, true)
	assertAccess("owner@example.invalid", 202, false)
	assertAccess("operational@example.invalid", 303, false)
	assertAccess("super@example.invalid", 202, true)
	exec(`UPDATE administradores SET role='administrador' WHERE id=3`)
	assertAccess("super@example.invalid", 202, false)
	identity, err := GetAdminAuthorizationIdentity(conn, "super@example.invalid")
	if err != nil || identity.Role != "administrador" {
		t.Fatalf("stale administrative role: %v", err)
	}
	exec(`UPDATE administradores SET role='super_administrador' WHERE id=3`)
	exec(`UPDATE empresas SET estado='inactivo' WHERE empresa_id=202`)
	assertAccess("super@example.invalid", 202, false)
	exec(`UPDATE empresas SET estado='activo' WHERE empresa_id=202`)
	exec(`UPDATE administradores SET estado='inactivo' WHERE id=1`)
	assertAccess("owner@example.invalid", 101, false)
	assertAccess("child@example.invalid", 101, false)
	exec(`UPDATE administradores SET estado='activo' WHERE id=1`)
	assertAccess("child@example.invalid", 101, true)
	exec(`UPDATE administradores SET usuario_creador='' WHERE id=4`)
	assertAccess("child@example.invalid", 101, false)
	exec(`UPDATE empresas SET usuario_creador='other@example.invalid' WHERE empresa_id=101`)
	assertAccess("owner@example.invalid", 101, false)
	exec(`UPDATE empresas SET usuario_creador='owner@example.invalid' WHERE empresa_id=101`)

	assertAccess("delegate@example.invalid", 101, false)
	exec(`INSERT INTO admin_principal_delegaciones (admin_email,principal_email,estado) VALUES ('delegate@example.invalid','owner@example.invalid','activo')`)
	assertAccess("delegate@example.invalid", 101, true)
	assertAccess("delegate@example.invalid", 202, false)
	exec(`UPDATE admin_principal_delegaciones SET fecha_revocada='revoked' WHERE admin_email='delegate@example.invalid'`)
	assertAccess("delegate@example.invalid", 101, false)
	exec(`INSERT INTO admin_empresa_compartida (id,empresa_id,admin_email,compartido_por_email,nivel_acceso,modulos_permitidos) VALUES (1,101,'delegate@example.invalid','owner@example.invalid','acceso_total','ventas,inventario')`)
	assertAccess("delegate@example.invalid", 101, true)
	first, err := GetActiveAdminEmpresaCompartidaAcceso(conn, 101, "delegate@example.invalid")
	if err != nil || first == nil || first.NivelAcceso != "acceso_total" {
		t.Fatalf("initial shared access: %v", err)
	}
	exec(`UPDATE admin_empresa_compartida SET nivel_acceso='solo_lectura',modulos_permitidos='ventas' WHERE id=1`)
	current, err := GetActiveAdminEmpresaCompartidaAcceso(conn, 101, "delegate@example.invalid")
	if err != nil || current == nil || current.NivelAcceso != "solo_lectura" || current.ModulosPermitidos != "ventas" {
		t.Fatalf("shared ceiling not current: %v", err)
	}

	// The access path must work with PostgreSQL read-only transactions; no Ensure/DDL.
	readQuery := u.Query()
	readQuery.Set("default_transaction_read_only", "on")
	u.RawQuery = readQuery.Encode()
	readOnly, err := sql.Open(PostgresCompatDriverName(), u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close()
	if allowed, err := CanAdminAccessEmpresaIA(readOnly, readOnly, "delegate@example.invalid", 101); err != nil || !allowed {
		t.Fatalf("read-only authorization: allowed=%v error=%v", allowed, err)
	}
	if template, err := GetTipoEmpresaPreconfiguracionByTipoID(readOnly, 1); err != nil || template == nil || !template.Enabled {
		t.Fatalf("read-only template lookup: %v", err)
	}

	exec(`UPDATE admin_empresa_compartida SET fecha_revocada='revoked' WHERE id=1`)
	assertAccess("delegate@example.invalid", 101, false)
	exec(`INSERT INTO admin_empresa_compartida (id,empresa_id,admin_email,compartido_por_email) VALUES (2,101,'other@example.invalid','delegate@example.invalid')`)
	assertAccess("delegate@example.invalid", 101, false)
	exec(`UPDATE admin_empresa_compartida SET fecha_revocada='' WHERE id=1`)
	exec(`UPDATE administradores SET estado='inactivo' WHERE id=2`)
	assertAccess("delegate@example.invalid", 101, false)
	if item, err := GetActiveAdminEmpresaCompartidaAcceso(conn, 101, "delegate@example.invalid"); err != nil || item != nil {
		t.Fatalf("inactive shared recipient: %v", err)
	}
	exec(`UPDATE administradores SET estado='activo' WHERE id=2`)
	exec(`DROP TABLE admin_principal_delegaciones`)
	if allowed, err := CanAdminAccessEmpresaIA(conn, conn, "delegate@example.invalid", 101); err == nil || allowed {
		t.Fatal("missing authorization schema failed open")
	}
}
