package db

import (
	"database/sql"
	"os"
	"testing"
)

func TestEmpresaUsuarioSessionRevocationRequiresScope(t *testing.T) {
	for _, tc := range []struct {
		empresa, usuario int64
		email            string
	}{
		{0, 5, "shared@example.test"}, {12, 0, "shared@example.test"}, {12, 5, ""},
	} {
		if err := RevokeEmpresaUsuarioSessions(nil, tc.empresa, tc.usuario, tc.email); err == nil {
			t.Fatal("invalid principal must not revoke any session")
		}
	}
}

func TestEmpresaUsuarioSessionRevocationPostgres(t *testing.T) {
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal("open isolated PostgreSQL")
	}
	defer conn.Close()
	conn.SetMaxOpenConns(1)
	if _, err = conn.Exec(`CREATE TEMP TABLE sesiones (
		id BIGINT, admin_email TEXT, principal_type TEXT, principal_id BIGINT,
		empresa_id BIGINT, activo INTEGER, fecha_fin TIMESTAMPTZ)`); err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(`INSERT INTO sesiones VALUES
		(1,'shared@example.test','empresa_usuario',5,12,1,NULL),
		(2,'shared@example.test','empresa_usuario',5,13,1,NULL),
		(3,'shared@example.test','admin',0,0,1,NULL),
		(4,'shared@example.test','empresa_usuario',6,12,1,NULL),
		(5,'another@example.test','empresa_usuario',5,12,1,NULL),
		(6,'shared@example.test','empresa_usuario',5,12,0,NULL)`); err != nil {
		t.Fatal(err)
	}
	if err = RevokeEmpresaUsuarioSessions(conn, 12, 5, " Shared@Example.Test "); err != nil {
		t.Fatal(err)
	}
	rows, err := conn.Query(`SELECT id,activo FROM sesiones ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id, active int
		if err = rows.Scan(&id, &active); err != nil {
			t.Fatal(err)
		}
		want := 1
		if id == 1 || id == 6 {
			want = 0
		}
		if active != want {
			t.Fatalf("session %d active=%d want=%d", id, active, want)
		}
		count++
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 6 {
		t.Fatalf("expected all scoped and unscoped controls, got %d", count)
	}
}
