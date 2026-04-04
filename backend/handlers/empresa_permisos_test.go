package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"mime/multipart"
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

func TestWithEmpresaClientesPermissionsAllowsCajeroWrite(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 20, "cajero@cliente.com")
	seedPermsAdmin(t, dbSuper, "cajero@cliente.com", "cajero")

	nextCalled := false
	h := WithEmpresaClientesPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/clientes", strings.NewReader(`{"empresa_id":20,"nombre":"cliente test"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@cliente.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for cajero writing clientes, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called when permission is granted")
	}
}

func TestWithEmpresaComprasPermissionsDeniesCajeroWrite(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 21, "cajero@compras.com")
	seedPermsAdmin(t, dbSuper, "cajero@compras.com", "cajero")

	nextCalled := false
	h := WithEmpresaComprasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/proveedores", strings.NewReader(`{"empresa_id":21,"nombre":"Proveedor Test"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@compras.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cajero writing compras/proveedores, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaFacturacionPermissionsDeniesSupervisorWrite(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 22, "supervisor@factura.com")
	seedPermsAdmin(t, dbSuper, "supervisor@factura.com", "supervisor_sucursal")

	nextCalled := false
	h := WithEmpresaFacturacionPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica", strings.NewReader(`{"empresa_id":22,"pais_codigo":"CO"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "supervisor@factura.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for supervisor writing facturacion, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaSeguridadPermissionsDeniesSupervisorWrite(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 30, "supervisor@seguridad.com")
	seedPermsAdmin(t, dbSuper, "supervisor@seguridad.com", "supervisor_sucursal")

	nextCalled := false
	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios", strings.NewReader(`{"empresa_id":30,"email":"nuevo@empresa.com"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "supervisor@seguridad.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for supervisor writing seguridad, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaSeguridadPermissionsAllowsSupervisorRead(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 31, "supervisor@seguridad.com")
	seedPermsAdmin(t, dbSuper, "supervisor@seguridad.com", "supervisor_sucursal")

	nextCalled := false
	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/empresa/configuracion_avanzada?empresa_id=31", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "supervisor@seguridad.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for supervisor reading seguridad, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called when permission is granted")
	}
}

func TestWithEmpresaSeguridadPermissionsAllowsAdminApprove(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 32, "admin@seguridad.com")
	seedPermsAdmin(t, dbSuper, "admin@seguridad.com", "administrador")

	nextCalled := false
	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/usuarios?empresa_id=32&id=9&action=reenviar_confirmacion", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@seguridad.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin approve in seguridad, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called when permission is granted")
	}
}

func TestWithEmpresaInventarioPermissionsDeniesCajeroWriteGPS(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 33, "cajero@gps.com")
	seedPermsAdmin(t, dbSuper, "cajero@gps.com", "cajero")

	nextCalled := false
	h := WithEmpresaInventarioPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/ubicacion_gps/dispositivos", strings.NewReader(`{"empresa_id":33,"nombre":"GPS 1"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@gps.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cajero writing ubicacion_gps, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaVentasPermissionsAllowsCajeroChatAdjuntoMultipart(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 34, "cajero@chat.com")
	seedPermsAdmin(t, dbSuper, "cajero@chat.com", "cajero")

	nextCalled := false
	h := WithEmpresaVentasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("empresa_id", "34"); err != nil {
		t.Fatalf("write multipart empresa_id: %v", err)
	}
	part, err := writer.CreateFormFile("archivo", "nota.txt")
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte("adjunto de prueba")); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/chat_tareas/mensajes/adjunto", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@chat.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for cajero posting chat adjunto, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called when permission is granted")
	}
	if got := rr.Header().Get("X-Empresa-ID"); got != "34" {
		t.Fatalf("expected response header X-Empresa-ID=34, got %q", got)
	}
}

func TestWithEmpresaVentasPermissionsRejectsChatAdjuntoWithoutAuth(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 35, "admin@chat.com")
	seedPermsAdmin(t, dbSuper, "admin@chat.com", "administrador")

	nextCalled := false
	h := WithEmpresaVentasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("empresa_id", "35"); err != nil {
		t.Fatalf("write multipart empresa_id: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/chat_tareas/mensajes/adjunto", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without auth context in chat adjunto, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when unauthenticated")
	}
}
