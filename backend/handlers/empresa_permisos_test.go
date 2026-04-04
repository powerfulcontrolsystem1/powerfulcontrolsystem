package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func openPermsTestDB(t *testing.T, name string) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), name)
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	return dbConn
}

func ensurePermsAdminSchema(t *testing.T, dbSuper *sql.DB) {
	t.Helper()
	_, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		name TEXT,
		role TEXT,
		photo TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		estado TEXT DEFAULT 'activo'
	)`)
	if err != nil {
		t.Fatalf("create administradores table: %v", err)
	}
}

func ensurePermsEmpresasSchema(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	_, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT,
		usuario_creador TEXT
	)`)
	if err != nil {
		t.Fatalf("create empresas table: %v", err)
	}
}

func seedPermsAdmin(t *testing.T, dbSuper *sql.DB, email, role string) {
	t.Helper()
	_, err := dbSuper.Exec(`INSERT INTO administradores (email, name, role, estado) VALUES (?, ?, ?, 'activo')`, email, "Admin", role)
	if err != nil {
		t.Fatalf("insert admin: %v", err)
	}
}

func seedPermsEmpresa(t *testing.T, dbEmp *sql.DB, id int64, creador string) {
	t.Helper()
	_, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, usuario_creador) VALUES (?, ?, ?)`, id, "Empresa test", creador)
	if err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
}

func TestWithEmpresaFinanzasPermissionsDeniesInventarioWrite(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 7, "inventario@test.com")
	seedPermsAdmin(t, dbSuper, "inventario@test.com", "inventario")

	nextCalled := false
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/movimientos", strings.NewReader(`{"empresa_id":7,"tipo_movimiento":"ingreso","concepto":"test","total":1000}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "inventario@test.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for inventario role writing finanzas, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaFinanzasPermissionsAllowsContabilidadApprove(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 8, "conta@test.com")
	seedPermsAdmin(t, dbSuper, "conta@test.com", "contabilidad")

	nextCalled := false
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/periodos?empresa_id=8&action=cerrar&periodo=2026-04", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "conta@test.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for contabilidad approve, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called when permission is granted")
	}
}

func TestWithEmpresaInventarioPermissionsDeniesCajeroWrite(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 9, "cajero@test.com")
	seedPermsAdmin(t, dbSuper, "cajero@test.com", "cajero")

	nextCalled := false
	h := WithEmpresaInventarioPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/inventario/ajustar", strings.NewReader(`{"empresa_id":9,"producto_id":1,"bodega_id":1,"tipo":"entrada","cantidad":1}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@test.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cajero writing inventario, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaVentasPermissionsDeniesOutOfScopeEmpresa(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 10, "otro@test.com")
	seedPermsAdmin(t, dbSuper, "admin@test.com", "administrador")

	nextCalled := false
	h := WithEmpresaVentasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=10", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@test.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for empresa fuera de alcance, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when scope validation fails")
	}
}
