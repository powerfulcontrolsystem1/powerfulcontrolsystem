package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestSuperRolesPostgresSessionPrincipalAndCurrentAdmin(t *testing.T) {
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	adminDB, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal("open isolated PostgreSQL failed")
	}
	suffix := time.Now().UnixNano()
	schema := fmt.Sprintf("super_roles_session_%d", suffix)
	if _, err := adminDB.Exec("CREATE SCHEMA " + schema); err != nil {
		adminDB.Close()
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	conn, err := sql.Open(dbpkg.PostgresCompatDriverName(), u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn.Close()
		if _, err := adminDB.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Error(err)
		}
		adminDB.Close()
	})
	exec := func(query string, args ...interface{}) {
		t.Helper()
		if _, err := conn.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`CREATE TABLE administradores (
		id BIGINT PRIMARY KEY, email TEXT UNIQUE, name TEXT DEFAULT '', role TEXT,
		photo TEXT DEFAULT '', usuario_creador TEXT DEFAULT '', fecha_creacion TEXT DEFAULT '',
		fecha_actualizacion TEXT DEFAULT '', estado TEXT, acepta_contrato INT DEFAULT 0
	)`)
	exec(`CREATE TABLE sesiones (
		id BIGSERIAL PRIMARY KEY, admin_email TEXT, token TEXT, token_hash TEXT UNIQUE,
		ip TEXT, user_agent TEXT, fecha_inicio TEXT, fecha_fin TEXT, fecha_creacion TEXT,
		activo INT, principal_type TEXT, principal_id BIGINT, empresa_id BIGINT, principal_role TEXT
	)`)
	email := fmt.Sprintf("same_identity_%d@example.invalid", suffix)
	exec(`INSERT INTO administradores(id,email,role,estado) VALUES(1,?,'super_administrador','activo')`, email)
	adminToken, operationalToken := "synthetic-admin-token", "synthetic-operational-token"
	if err := dbpkg.CreateSession(conn, email, "127.0.0.1", "QA", adminToken); err != nil {
		t.Fatal(err)
	}
	if err := dbpkg.CreateEmpresaUsuarioSession(conn, email, "127.0.0.1", "QA", operationalToken, 77, 71001, "cajero"); err != nil {
		t.Fatal(err)
	}
	check := func(token string, wantStatus int, wantAllowed bool) {
		t.Helper()
		r := httptest.NewRequest(http.MethodGet, "/super/api/roles_de_usuario", nil)
		if token != "" {
			r.AddCookie(&http.Cookie{Name: "session_token", Value: token})
		}
		// These browser-controlled headers must never upgrade the session.
		r.Header.Set("X-Admin-Role", "super_administrador")
		r.Header.Set("X-Admin-Email", email)
		w := httptest.NewRecorder()
		gotEmail, allowed := paginaPrincipalRequireSuperAdmin(w, r, conn)
		if allowed != wantAllowed || w.Code != wantStatus || (allowed && gotEmail != email) || (!allowed && gotEmail != "") {
			t.Fatalf("Super boundary allowed=%v status=%d; want allowed=%v status=%d", allowed, w.Code, wantAllowed, wantStatus)
		}
	}
	check("", http.StatusUnauthorized, false)
	check(operationalToken, http.StatusForbidden, false)
	check(adminToken, http.StatusOK, true)
	// Prime the old presentation cache, then mutate without its invalidators.
	// Authorization must still observe the next database state immediately.
	if _, err := dbpkg.GetAdminByEmail(conn, email); err != nil {
		t.Fatal(err)
	}
	exec(`UPDATE administradores SET role='admin_empresa' WHERE id=1`)
	check(adminToken, http.StatusForbidden, false)
	exec(`UPDATE administradores SET role='super_administrador', estado='inactivo' WHERE id=1`)
	check(adminToken, http.StatusUnauthorized, false)
	exec(`UPDATE administradores SET estado='activo' WHERE id=1`)
	check(adminToken, http.StatusOK, true)
	exec(`UPDATE sesiones SET principal_id=999 WHERE principal_type='admin'`)
	check(adminToken, http.StatusUnauthorized, false)
	exec(`UPDATE sesiones SET principal_id=0, empresa_id=71001 WHERE principal_type='admin'`)
	check(adminToken, http.StatusForbidden, false)
	exec(`UPDATE sesiones SET empresa_id=0 WHERE principal_type='admin'`)
	if err := dbpkg.RevokeSessionByToken(conn, adminToken); err != nil {
		t.Fatal(err)
	}
	check(adminToken, http.StatusUnauthorized, false)
	check(operationalToken, http.StatusForbidden, false)
}
