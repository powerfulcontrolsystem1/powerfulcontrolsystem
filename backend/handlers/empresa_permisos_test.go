package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
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

func ensurePermsRoleConfigSchema(t *testing.T, dbSuper *sql.DB) {
	t.Helper()
	_, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS tipos_de_empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT NOT NULL,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`)
	if err != nil {
		t.Fatalf("create tipos_de_empresas table: %v", err)
	}
	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS roles_de_usuario (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tipo_empresa_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		descripcion TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`)
	if err != nil {
		t.Fatalf("create roles_de_usuario table: %v", err)
	}
	if err := dbpkg.EnsureRolesPermisosSchema(dbSuper); err != nil {
		t.Fatalf("ensure roles permisos schema: %v", err)
	}
}

func ensurePermsLicenciasSchema(t *testing.T, dbSuper *sql.DB) {
	t.Helper()
	_, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS licencias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre TEXT,
		modulos_habilitados TEXT,
		super_rol_habilitado INTEGER DEFAULT 0,
		fecha_inicio TEXT,
		fecha_fin TEXT,
		activo INTEGER DEFAULT 1
	)`)
	if err != nil {
		t.Fatalf("create licencias table: %v", err)
	}
}

func seedPermsLicencia(t *testing.T, dbSuper *sql.DB, empresaID int64, nombre, modulos string, superRol bool) {
	t.Helper()
	superRolInt := 0
	if superRol {
		superRolInt = 1
	}
	_, err := dbSuper.Exec(`INSERT INTO licencias (
		empresa_id,
		nombre,
		modulos_habilitados,
		super_rol_habilitado,
		fecha_inicio,
		fecha_fin,
		activo
	) VALUES (?, ?, ?, ?, datetime('now','-1 day','localtime'), datetime('now','+30 day','localtime'), 1)`, empresaID, nombre, modulos, superRolInt)
	if err != nil {
		t.Fatalf("insert licencia: %v", err)
	}
}

func seedPermsRolDeUsuario(t *testing.T, dbSuper *sql.DB, nombreRol string) int64 {
	t.Helper()
	res, err := dbSuper.Exec(`INSERT INTO tipos_de_empresas (nombre, estado) VALUES ('Tipo Test', 'activo')`)
	if err != nil {
		t.Fatalf("insert tipos_de_empresas: %v", err)
	}
	tipoID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("tipo last insert id: %v", err)
	}
	res, err = dbSuper.Exec(`INSERT INTO roles_de_usuario (tipo_empresa_id, nombre, descripcion, estado) VALUES (?, ?, ?, 'activo')`, tipoID, nombreRol, "rol de prueba")
	if err != nil {
		t.Fatalf("insert roles_de_usuario: %v", err)
	}
	rolID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("rol last insert id: %v", err)
	}
	return rolID
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

func TestWithEmpresaVentasPermissionsInjectsEmpresaIDContextForParsers(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 71, "admin@scope.com")
	seedPermsAdmin(t, dbSuper, "admin@scope.com", "administrador")

	h := WithEmpresaVentasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		empresaIDByRequired, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "parseEmpresaIDQuery fallo: "+err.Error(), http.StatusBadRequest)
			return
		}
		empresaIDByOptional, err := parseInt64QueryOptional(r, "empresa_id")
		if err != nil {
			http.Error(w, "parseInt64QueryOptional fallo: "+err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]int64{
			"empresa_id_required": empresaIDByRequired,
			"empresa_id_optional": empresaIDByOptional,
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/ventas/context_scope", strings.NewReader(`{"empresa_id":71,"descripcion":"prueba"}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@scope.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for context injected empresa_id, got %d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("X-Empresa-ID"); got != "71" {
		t.Fatalf("expected response header X-Empresa-ID=71, got %q", got)
	}

	var payload map[string]int64
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if payload["empresa_id_required"] != 71 {
		t.Fatalf("expected empresa_id_required=71, got %d", payload["empresa_id_required"])
	}
	if payload["empresa_id_optional"] != 71 {
		t.Fatalf("expected empresa_id_optional=71, got %d", payload["empresa_id_optional"])
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

func TestWithEmpresaFinanzasPermissionsDeniesCajeroAprobarCierreCaja(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 41, "cajero@cierres.com")
	seedPermsAdmin(t, dbSuper, "cajero@cierres.com", "cajero")

	nextCalled := false
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/cierres_caja?action=aprobar&empresa_id=41&id=101", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@cierres.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cajero approving cierres_caja, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaFinanzasPermissionsDeniesSupervisorAprobarCierreCaja(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 42, "supervisor@cierres.com")
	seedPermsAdmin(t, dbSuper, "supervisor@cierres.com", "supervisor_sucursal")

	nextCalled := false
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/cierres_caja?action=aprobar&empresa_id=42&id=102", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "supervisor@cierres.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for supervisor approving cierres_caja, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaFinanzasPermissionsAllowsAdminAprobarCierreCaja(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 43, "admin@cierres.com")
	seedPermsAdmin(t, dbSuper, "admin@cierres.com", "administrador")

	nextCalled := false
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/cierres_caja?action=aprobar&empresa_id=43&id=103", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@cierres.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin approving cierres_caja, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called when permission is granted")
	}
}

func TestWithEmpresaFinanzasPermissionsDeniesCajeroProcesarAsientos(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 44, "cajero@asientos.com")
	seedPermsAdmin(t, dbSuper, "cajero@asientos.com", "cajero")

	nextCalled := false
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/asientos_contables?action=procesar_asientos&empresa_id=44", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@asientos.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cajero processing asientos, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when permission is denied")
	}
}

func TestWithEmpresaFinanzasPermissionsAllowsContabilidadProcesarAsientos(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 45, "conta@asientos.com")
	seedPermsAdmin(t, dbSuper, "conta@asientos.com", "contabilidad")

	nextCalled := false
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/asientos_contables?action=procesar_asientos&empresa_id=45", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "conta@asientos.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for contabilidad processing asientos, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called when permission is granted")
	}
}

func TestEmpresaPermisosContextoHandlerRetornaPermisosPorRol(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 46, "conta@contexto.com")
	seedPermsAdmin(t, dbSuper, "conta@contexto.com", "contabilidad")

	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaPermisosContextoHandler(dbSuper))
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/permisos_contexto?empresa_id=46", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "conta@contexto.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for permisos_contexto, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp empresaPermisosContextResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode permisos_contexto response: %v body=%s", err, rr.Body.String())
	}

	if resp.EmpresaID != 46 {
		t.Fatalf("expected empresa_id=46, got %d", resp.EmpresaID)
	}
	if resp.Rol != "contabilidad" {
		t.Fatalf("expected rol=contabilidad, got %q", resp.Rol)
	}
	if resp.IncluyeMatriz {
		t.Fatalf("expected incluye_matriz=false by default")
	}
	if got, want := len(resp.Modulos), len(permissionModulesCatalogOrdered); got != want {
		t.Fatalf("expected %d modulos, got %d", want, got)
	}

	finanzas, ok := findPermissionModuleRow(resp.Modulos, permModuleFinanzas)
	if !ok {
		t.Fatalf("finanzas module must exist in response")
	}
	if !finanzas.Read || !finanzas.Create || !finanzas.Update || !finanzas.Approve {
		t.Fatalf("contabilidad must have read/create/update/approve over finanzas: %+v", finanzas)
	}

	seguridad, ok := findPermissionModuleRow(resp.Modulos, permModuleSeguridad)
	if !ok {
		t.Fatalf("seguridad module must exist in response")
	}
	if !seguridad.Read {
		t.Fatalf("contabilidad must keep read on seguridad")
	}
	if seguridad.Create || seguridad.Update || seguridad.Delete || seguridad.Approve {
		t.Fatalf("contabilidad must not escalate seguridad write/approve actions: %+v", seguridad)
	}
}

func TestEmpresaPermisosContextoHandlerAplicaOverridesPorRol(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	ensurePermsRoleConfigSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 48, "conta@override.com")
	seedPermsAdmin(t, dbSuper, "conta@override.com", "contabilidad")
	rolID := seedPermsRolDeUsuario(t, dbSuper, "contabilidad")

	if err := dbpkg.ReplaceRolPermisosDeUsuario(dbSuper, rolID,
		[]dbpkg.RolPermisoModulo{{RolID: rolID, Modulo: permModuleFinanzas, Accion: permActionApprove, Permitido: false}},
		[]dbpkg.RolPermisoPagina{{RolID: rolID, PaginaClave: "linkFinanzas", Permitido: false}},
		"tester",
	); err != nil {
		t.Fatalf("seed role overrides: %v", err)
	}

	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaPermisosContextoHandler(dbSuper))
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/permisos_contexto?empresa_id=48", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "conta@override.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for permisos_contexto with overrides, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp empresaPermisosContextResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode permisos_contexto response: %v body=%s", err, rr.Body.String())
	}

	finanzas, ok := findPermissionModuleRow(resp.Modulos, permModuleFinanzas)
	if !ok {
		t.Fatalf("finanzas module must exist in response")
	}
	if finanzas.Approve {
		t.Fatalf("override must remove approve on finanzas for contabilidad")
	}
	if visible, ok := resp.Paginas["linkFinanzas"]; !ok || visible {
		t.Fatalf("override must hide linkFinanzas in paginas map, got exists=%v visible=%v", ok, visible)
	}
}

func TestWithEmpresaFinanzasPermissionsRespetaOverrideDenegado(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	ensurePermsRoleConfigSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 49, "conta@override.com")
	seedPermsAdmin(t, dbSuper, "conta@override.com", "contabilidad")
	rolID := seedPermsRolDeUsuario(t, dbSuper, "contabilidad")

	if err := dbpkg.ReplaceRolPermisosDeUsuario(dbSuper, rolID,
		[]dbpkg.RolPermisoModulo{{RolID: rolID, Modulo: permModuleFinanzas, Accion: permActionApprove, Permitido: false}},
		nil,
		"tester",
	); err != nil {
		t.Fatalf("seed role module override: %v", err)
	}

	nextCalled := false
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/periodos?empresa_id=49&action=cerrar", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "conta@override.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when override denies approve, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when override denies permission")
	}
}

func TestEmpresaPermisosContextoHandlerIncluyeMatrizRoles(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 47, "admin@contexto.com")
	seedPermsAdmin(t, dbSuper, "admin@contexto.com", "administrador")

	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaPermisosContextoHandler(dbSuper))
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/permisos_contexto?empresa_id=47&include_matrix=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@contexto.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for permisos_contexto include_matrix, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp empresaPermisosContextResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode permisos_contexto include_matrix response: %v body=%s", err, rr.Body.String())
	}

	if !resp.IncluyeMatriz {
		t.Fatalf("expected incluye_matriz=true when include_matrix=1")
	}
	if got, want := len(resp.MatrizRoles), len(permissionRolesCatalogOrdered); got != want {
		t.Fatalf("expected %d role rows in matriz_roles, got %d", want, got)
	}

	superAdmin, ok := findPermissionRoleRow(resp.MatrizRoles, "super_administrador")
	if !ok {
		t.Fatalf("matriz_roles must include super_administrador")
	}
	seguridad, ok := findPermissionModuleRow(superAdmin.Modulos, permModuleSeguridad)
	if !ok {
		t.Fatalf("seguridad module must exist in super_administrador matrix")
	}
	if !seguridad.Read || !seguridad.Create || !seguridad.Update || !seguridad.Delete || !seguridad.Approve {
		t.Fatalf("super_administrador must have full seguridad permissions: %+v", seguridad)
	}
	if superAdmin.Resumen.AccionesHabilitadas == 0 {
		t.Fatalf("summary for super_administrador must expose enabled actions")
	}
}

func TestEmpresaPermisosContextoHandlerMatrizRolesCumplePoliticaPorModuloAccion(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 74, "admin@matriz.com")
	seedPermsAdmin(t, dbSuper, "admin@matriz.com", "administrador")

	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaPermisosContextoHandler(dbSuper))
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/permisos_contexto?empresa_id=74&include_matrix=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@matriz.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for include_matrix policy test, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp empresaPermisosContextResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode include_matrix policy response: %v body=%s", err, rr.Body.String())
	}

	if got, want := len(resp.MatrizRoles), len(permissionRolesCatalogOrdered); got != want {
		t.Fatalf("expected %d roles in matrix, got %d", want, got)
	}

	for _, role := range permissionRolesCatalogOrdered {
		roleRow, ok := findPermissionRoleRow(resp.MatrizRoles, role)
		if !ok {
			t.Fatalf("role %q must be present in matriz_roles", role)
		}
		for _, modulo := range permissionModulesCatalogOrdered {
			moduleRow, ok := findPermissionModuleRow(roleRow.Modulos, modulo)
			if !ok {
				t.Fatalf("module %q must be present for role %q", modulo, role)
			}
			for _, accion := range permissionActionsCatalogOrdered {
				expected := expectedPolicyPermission(role, modulo, accion)
				actual := moduleRow.Acciones[accion]
				if actual != expected {
					t.Fatalf("unexpected permission role=%s modulo=%s accion=%s expected=%v got=%v", role, modulo, accion, expected, actual)
				}
			}
		}
	}
}

func TestWithEmpresaSeguridadPermissionsRequiereAprobacionParaCambioPermisos(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 80, "admin@seguridad.com")
	seedPermsAdmin(t, dbSuper, "admin@seguridad.com", "administrador")

	nextCalled := false
	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios", strings.NewReader(`{"empresa_id":80,"email":"nuevo@empresa.com","nombre":"Nuevo","rol_usuario_id":1}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@seguridad.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when approval evidence is missing, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called without approval evidence")
	}
}

func TestWithEmpresaSeguridadPermissionsAceptaAprobacionTrazableYRegistraMetadata(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 81, "admin@seguridad.com")
	seedPermsAdmin(t, dbSuper, "admin@seguridad.com", "administrador")

	nextCalled := false
	capturedApprovedBy := ""
	capturedApprovalCode := ""
	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		capturedApprovedBy = r.Header.Get(permissionApprovalHeaderBy)
		capturedApprovalCode = r.Header.Get(permissionApprovalHeaderCode)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/empresa/usuarios?aprobado_por=director.seguridad%40empresa.com&codigo_aprobacion=APR-SEG-001&motivo_aprobacion=ajuste_rol",
		strings.NewReader(`{"empresa_id":81,"email":"nuevo@empresa.com","nombre":"Nuevo","rol_usuario_id":1}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin@seguridad.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when approval evidence is present, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called with valid approval evidence")
	}
	if capturedApprovedBy != "director.seguridad@empresa.com" {
		t.Fatalf("expected approved_by header in request context, got %q", capturedApprovedBy)
	}
	if capturedApprovalCode != "APR-SEG-001" {
		t.Fatalf("expected approval_code header in request context, got %q", capturedApprovalCode)
	}

	var metadataJSON string
	err := dbEmp.QueryRow(`SELECT metadata_json FROM empresa_auditoria_eventos WHERE empresa_id = ? ORDER BY id DESC LIMIT 1`, 81).Scan(&metadataJSON)
	if err != nil {
		t.Fatalf("expected audit metadata for permission change, query error: %v", err)
	}

	metadata := map[string]interface{}{}
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		t.Fatalf("decode audit metadata: %v metadata=%s", err, metadataJSON)
	}

	if required, ok := metadata["permission_approval_required"].(bool); !ok || !required {
		t.Fatalf("expected permission_approval_required=true in metadata, got %v", metadata["permission_approval_required"])
	}
	if got := strings.TrimSpace(toStringForTest(metadata["permission_approved_by"])); got != "director.seguridad@empresa.com" {
		t.Fatalf("expected permission_approved_by in metadata, got %q", got)
	}
	if got := strings.TrimSpace(toStringForTest(metadata["permission_approval_code"])); got != "APR-SEG-001" {
		t.Fatalf("expected permission_approval_code in metadata, got %q", got)
	}
}

func expectedPolicyPermission(role, modulo, accion string) bool {
	role = normalizePermissionRole(role)
	if role == "super_administrador" {
		return true
	}

	allReadRoles := map[string]bool{
		"admin_empresa":       true,
		"supervisor_sucursal": true,
		"cajero":              true,
		"inventario":          true,
		"compras":             true,
		"contabilidad":        true,
		"auditor":             true,
	}
	if accion == permActionRead {
		return allReadRoles[role]
	}

	allowedByModule := map[string]map[string]map[string]bool{
		permModuleVentas: {
			permActionCreate:  {"admin_empresa": true, "supervisor_sucursal": true, "cajero": true},
			permActionUpdate:  {"admin_empresa": true, "supervisor_sucursal": true, "cajero": true},
			permActionDelete:  {"admin_empresa": true, "supervisor_sucursal": true, "cajero": true},
			permActionApprove: {"admin_empresa": true, "supervisor_sucursal": true, "cajero": true},
		},
		permModuleInventario: {
			permActionCreate:  {"admin_empresa": true, "supervisor_sucursal": true, "inventario": true},
			permActionUpdate:  {"admin_empresa": true, "supervisor_sucursal": true, "inventario": true},
			permActionDelete:  {"admin_empresa": true, "supervisor_sucursal": true, "inventario": true},
			permActionApprove: {"admin_empresa": true, "supervisor_sucursal": true, "inventario": true},
		},
		permModuleFinanzas: {
			permActionCreate:  {"admin_empresa": true, "contabilidad": true},
			permActionUpdate:  {"admin_empresa": true, "contabilidad": true},
			permActionDelete:  {"contabilidad": true},
			permActionApprove: {"admin_empresa": true, "contabilidad": true},
		},
		permModuleClientes: {
			permActionCreate:  {"admin_empresa": true, "supervisor_sucursal": true, "cajero": true},
			permActionUpdate:  {"admin_empresa": true, "supervisor_sucursal": true, "cajero": true},
			permActionDelete:  {},
			permActionApprove: {"admin_empresa": true, "supervisor_sucursal": true, "cajero": true},
		},
		permModuleCompras: {
			permActionCreate:  {"admin_empresa": true, "supervisor_sucursal": true, "compras": true},
			permActionUpdate:  {"admin_empresa": true, "supervisor_sucursal": true, "compras": true},
			permActionDelete:  {},
			permActionApprove: {"admin_empresa": true, "supervisor_sucursal": true, "compras": true},
		},
		permModuleFacturacion: {
			permActionCreate:  {"admin_empresa": true, "cajero": true},
			permActionUpdate:  {"admin_empresa": true, "cajero": true},
			permActionDelete:  {},
			permActionApprove: {"admin_empresa": true, "cajero": true},
		},
		permModuleSeguridad: {
			permActionCreate:  {"admin_empresa": true},
			permActionUpdate:  {"admin_empresa": true},
			permActionDelete:  {"admin_empresa": true},
			permActionApprove: {"admin_empresa": true},
		},
	}

	moduleMap, ok := allowedByModule[modulo]
	if !ok {
		return false
	}
	actionMap, ok := moduleMap[accion]
	if !ok {
		return false
	}
	return actionMap[role]
}

func TestEmpresaPermisosContextoHandlerRestringeModulosPorLicencia(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	ensurePermsLicenciasSchema(t, dbSuper)

	seedPermsEmpresa(t, dbEmp, 81, "admin.licencia@test.com")
	seedPermsAdmin(t, dbSuper, "admin.licencia@test.com", "administrador")
	seedPermsLicencia(t, dbSuper, 81, "Plan Ventas+Clientes", "ventas,clientes", false)

	h := WithEmpresaSeguridadPermissions(dbEmp, dbSuper, EmpresaPermisosContextoHandler(dbSuper))
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/permisos_contexto?empresa_id=81", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin.licencia@test.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for permisos_contexto with licencia restriction, got %d body=%s", rr.Code, rr.Body.String())
	}

	var resp empresaPermisosContextResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode permisos_contexto response: %v body=%s", err, rr.Body.String())
	}

	if resp.RolEfectivo != "admin_empresa" {
		t.Fatalf("expected rol_efectivo admin_empresa, got %q", resp.RolEfectivo)
	}
	if resp.Licencia == nil {
		t.Fatalf("expected licencia context in response")
	}
	if !resp.Licencia.RestringeModulos {
		t.Fatalf("expected licencia to restrict modules")
	}

	ventas, ok := findPermissionModuleRow(resp.Modulos, permModuleVentas)
	if !ok {
		t.Fatalf("ventas module not found")
	}
	if !ventas.Read {
		t.Fatalf("expected ventas.read true when module enabled by licencia")
	}

	seguridad, ok := findPermissionModuleRow(resp.Modulos, permModuleSeguridad)
	if !ok {
		t.Fatalf("seguridad module not found")
	}
	if seguridad.Read || seguridad.Create || seguridad.Update || seguridad.Delete || seguridad.Approve {
		t.Fatalf("expected seguridad module fully blocked by licencia restriction")
	}

	if resp.Paginas["linkCreditos"] {
		t.Fatalf("expected linkCreditos to be hidden when finanzas is not enabled by licencia")
	}
}

func TestWithEmpresaFinanzasPermissionsSupervisorConSuperRolLicencia(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	ensurePermsLicenciasSchema(t, dbSuper)

	seedPermsEmpresa(t, dbEmp, 82, "supervisor.licencia@test.com")
	seedPermsAdmin(t, dbSuper, "supervisor.licencia@test.com", "supervisor")
	seedPermsLicencia(t, dbSuper, 82, "Plan Finanzas SuperRol", "finanzas", true)

	nextCalled := false
	effectiveRole := ""
	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		effectiveRole = r.Header.Get("X-Admin-Role-Efectivo")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/movimientos", strings.NewReader(`{"empresa_id":82,"tipo_movimiento":"ingreso","concepto":"test","total":120}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "supervisor.licencia@test.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for supervisor with super_rol license in finanzas, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !nextCalled {
		t.Fatalf("next handler must be called for supervisor with super_rol enabled")
	}
	if effectiveRole != "admin_empresa" {
		t.Fatalf("expected effective role admin_empresa, got %q", effectiveRole)
	}
}

func TestWithEmpresaVentasPermissionsBloqueaModuloNoHabilitadoPorLicencia(t *testing.T) {
	dbEmp := openPermsTestDB(t, "empresas.db")
	dbSuper := openPermsTestDB(t, "super.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	ensurePermsLicenciasSchema(t, dbSuper)

	seedPermsEmpresa(t, dbEmp, 83, "admin.finanzas@test.com")
	seedPermsAdmin(t, dbSuper, "admin.finanzas@test.com", "administrador")
	seedPermsLicencia(t, dbSuper, 83, "Plan Solo Finanzas", "finanzas", false)

	nextCalled := false
	h := WithEmpresaVentasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=83", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "admin.finanzas@test.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when modulo ventas is not enabled by licencia, got %d body=%s", rr.Code, rr.Body.String())
	}
	if nextCalled {
		t.Fatalf("next handler must not be called when modulo is blocked by licencia")
	}
}

func toStringForTest(v interface{}) string {
	switch typed := v.(type) {
	case string:
		return typed
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(typed, 10)
	case int:
		return strconv.Itoa(typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func findPermissionModuleRow(rows []permissionModuleMatrixRow, modulo string) (permissionModuleMatrixRow, bool) {
	for _, row := range rows {
		if row.Modulo == modulo {
			return row, true
		}
	}
	return permissionModuleMatrixRow{}, false
}

func findPermissionRoleRow(rows []empresaPermisosRolMatriz, role string) (empresaPermisosRolMatriz, bool) {
	for _, row := range rows {
		if row.Rol == role {
			return row, true
		}
	}
	return empresaPermisosRolMatriz{}, false
}
