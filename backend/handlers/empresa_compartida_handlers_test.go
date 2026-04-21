package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func ensureEmpresaCompartidaTestSchemas(t *testing.T, dbEmp, dbSuper *sql.DB) {
	t.Helper()
	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	_, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT,
		photo TEXT,
		usuario_creador TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		estado TEXT DEFAULT 'activo',
		acepta_contrato INTEGER DEFAULT 0,
		telefono TEXT,
		pais TEXT,
		ciudad TEXT,
		email_confirmado INTEGER DEFAULT 0,
		email_confirm_token TEXT,
		email_confirm_expira TEXT,
		email_confirmado_en TEXT,
		password_hash TEXT,
		password_salt TEXT,
		password_set INTEGER DEFAULT 0,
		password_reset_token TEXT,
		password_reset_expira TEXT
	)`)
	if err != nil {
		t.Fatalf("create administradores schema: %v", err)
	}
	if err := dbpkg.EnsureAdminEmpresaCompartidaSchema(dbSuper); err != nil {
		t.Fatalf("ensure admin empresa compartida schema: %v", err)
	}
}

func seedEmpresaCompartidaAdmin(t *testing.T, dbSuper *sql.DB, email, name, role, creator string) {
	t.Helper()
	nowValue := time.Now().Format("2006-01-02 15:04:05")
	creator = strings.TrimSpace(creator)
	if _, err := dbSuper.Exec(`
		INSERT INTO administradores (
			email, name, role, usuario_creador, fecha_creacion, fecha_actualizacion, estado, email_confirmado, password_set
		) VALUES (?, ?, ?, ?, ?, ?, 'activo', 1, 1)
	`, email, name, role, creator, nowValue, nowValue); err != nil {
		t.Fatalf("insert admin %s: %v", email, err)
	}
}

func seedEmpresaCompartidaEmpresa(t *testing.T, dbEmp *sql.DB, id int64, nombre, creador string) {
	t.Helper()
	nowValue := time.Now().Format("2006-01-02 15:04:05")
	_, err := dbpkg.ExecCompat(dbEmp, `
		INSERT INTO empresas (id, empresa_id, nombre, usuario_creador, estado, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, 'activo', ?, ?)
	`, id, id, nombre, creador, nowValue, nowValue)
	if err != nil {
		t.Fatalf("insert empresa %d: %v", id, err)
	}
}

func TestEmpresasHandlerListsSharedEmpresaForInvitedAdmin(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compartidas_list.db")
	dbSuper := openTestSQLite(t, "super_compartidas_list.db")
	ensureEmpresaCompartidaTestSchemas(t, dbEmp, dbSuper)

	seedEmpresaCompartidaAdmin(t, dbSuper, "owner@test.com", "Owner", "administrador", "")
	seedEmpresaCompartidaAdmin(t, dbSuper, "shared@test.com", "Shared", "administrador", "")
	seedEmpresaCompartidaEmpresa(t, dbEmp, 41, "Empresa Compartida", "owner@test.com")

	if _, err := dbpkg.UpsertAdminEmpresaCompartidaAcceso(dbSuper, dbpkg.AdminEmpresaCompartidaAcceso{
		EmpresaID:          41,
		AdminEmail:         "shared@test.com",
		CompartidoPorEmail: "owner@test.com",
		FechaAceptada:      time.Now().Format("2006-01-02 15:04:05"),
		UsuarioCreador:     "owner@test.com",
		Estado:             "activo",
	}); err != nil {
		t.Fatalf("upsert shared access: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/super/api/empresas", nil)
	req.Header.Set("X-Admin-Email", "shared@test.com")
	rr := httptest.NewRecorder()

	EmpresasHandler(dbEmp, dbSuper).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var empresas []dbpkg.Empresa
	if err := json.Unmarshal(rr.Body.Bytes(), &empresas); err != nil {
		t.Fatalf("decode empresas response: %v body=%s", err, rr.Body.String())
	}
	if len(empresas) != 1 {
		t.Fatalf("expected 1 empresa, got %d body=%s", len(empresas), rr.Body.String())
	}
	if empresas[0].EmpresaID != 41 {
		t.Fatalf("expected empresa_id 41, got %d", empresas[0].EmpresaID)
	}
	if empresas[0].AccessSource != "shared" {
		t.Fatalf("expected access_source shared, got %q", empresas[0].AccessSource)
	}
	if !strings.EqualFold(empresas[0].CompartidaPor, "owner@test.com") {
		t.Fatalf("expected compartida_por owner@test.com, got %q", empresas[0].CompartidaPor)
	}
}

func TestEmpresasHandlerBlocksSharedAdminFromUpdatingEmpresa(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compartidas_put.db")
	dbSuper := openTestSQLite(t, "super_compartidas_put.db")
	ensureEmpresaCompartidaTestSchemas(t, dbEmp, dbSuper)

	seedEmpresaCompartidaAdmin(t, dbSuper, "owner@test.com", "Owner", "administrador", "")
	seedEmpresaCompartidaAdmin(t, dbSuper, "shared@test.com", "Shared", "administrador", "")
	seedEmpresaCompartidaEmpresa(t, dbEmp, 52, "Empresa Bloqueada", "owner@test.com")

	if _, err := dbpkg.UpsertAdminEmpresaCompartidaAcceso(dbSuper, dbpkg.AdminEmpresaCompartidaAcceso{
		EmpresaID:          52,
		AdminEmail:         "shared@test.com",
		CompartidoPorEmail: "owner@test.com",
		FechaAceptada:      time.Now().Format("2006-01-02 15:04:05"),
		UsuarioCreador:     "owner@test.com",
		Estado:             "activo",
	}); err != nil {
		t.Fatalf("upsert shared access: %v", err)
	}

	body := strings.NewReader(`{"nombre":"Empresa Editada","tipo_id":0,"tipo_nombre":"","nit":"","observaciones":""}`)
	req := httptest.NewRequest(http.MethodPut, "/super/api/empresas?id=52", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Email", "shared@test.com")
	rr := httptest.NewRecorder()

	EmpresasHandler(dbEmp, dbSuper).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, rr.Code, rr.Body.String())
	}
}

func TestEmpresaCompartidaAcceptHandlerCreatesAccessAndMarksInvitationAccepted(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compartidas_accept.db")
	dbSuper := openTestSQLite(t, "super_compartidas_accept.db")
	ensureEmpresaCompartidaTestSchemas(t, dbEmp, dbSuper)

	seedEmpresaCompartidaAdmin(t, dbSuper, "owner@test.com", "Owner", "administrador", "")
	seedEmpresaCompartidaAdmin(t, dbSuper, "shared@test.com", "Shared", "administrador", "")
	seedEmpresaCompartidaEmpresa(t, dbEmp, 63, "Empresa Invitada", "owner@test.com")

	const rawToken = "token-invitacion-demo"
	invID, err := dbpkg.CreateAdminEmpresaCompartidaInvitacion(dbSuper, dbpkg.AdminEmpresaCompartidaInvitacion{
		EmpresaID:        63,
		AdminEmail:       "shared@test.com",
		InvitadoPorEmail: "owner@test.com",
		TokenHash:        hashAdminEmpresaCompartidaToken(rawToken),
		Mensaje:          "Te comparto esta empresa",
		ExpiraEn:         time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		UsuarioCreador:   "owner@test.com",
		Estado:           "pendiente",
	})
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/super/api/empresas/compartidos/aceptar", strings.NewReader(`{"token":"`+rawToken+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Email", "shared@test.com")
	rr := httptest.NewRecorder()

	EmpresaCompartidaAcceptHandler(dbEmp, dbSuper).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	access, err := dbpkg.GetActiveAdminEmpresaCompartidaAcceso(dbSuper, 63, "shared@test.com")
	if err != nil {
		t.Fatalf("get active shared access: %v", err)
	}
	if access == nil {
		t.Fatal("expected active shared access after accepting invitation")
	}
	if access.InvitacionID != invID {
		t.Fatalf("expected invitacion_id %d, got %d", invID, access.InvitacionID)
	}

	inv, err := dbpkg.GetAdminEmpresaCompartidaInvitacionByID(dbSuper, invID)
	if err != nil {
		t.Fatalf("get invitation by id: %v", err)
	}
	if inv == nil || !strings.EqualFold(inv.Estado, "aceptada") {
		t.Fatalf("expected invitation accepted, got %+v", inv)
	}
}

func TestEmpresaCompartidaInviteRejectsDuplicatePendingInvitation(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compartidas_duplicate.db")
	dbSuper := openTestSQLite(t, "super_compartidas_duplicate.db")
	ensureEmpresaCompartidaTestSchemas(t, dbEmp, dbSuper)

	seedEmpresaCompartidaAdmin(t, dbSuper, "owner@test.com", "Owner", "administrador", "")
	seedEmpresaCompartidaAdmin(t, dbSuper, "shared@test.com", "Shared", "administrador", "")
	seedEmpresaCompartidaEmpresa(t, dbEmp, 74, "Empresa Duplicada", "owner@test.com")

	if _, err := dbpkg.CreateAdminEmpresaCompartidaInvitacion(dbSuper, dbpkg.AdminEmpresaCompartidaInvitacion{
		EmpresaID:        74,
		AdminEmail:       "shared@test.com",
		InvitadoPorEmail: "owner@test.com",
		TokenHash:        hashAdminEmpresaCompartidaToken("token-pendiente"),
		Mensaje:          "Pendiente",
		ExpiraEn:         time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		UsuarioCreador:   "owner@test.com",
		Estado:           "pendiente",
	}); err != nil {
		t.Fatalf("seed pending invitation: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/super/api/empresas/compartidos", strings.NewReader(`{"empresa_id":74,"email":"shared@test.com","mensaje":"Otro intento"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Email", "owner@test.com")
	rr := httptest.NewRecorder()

	EmpresaCompartidaHandler(dbEmp, dbSuper).ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}
}

func TestCanAdminAccessEmpresaIAAllowsSharedAccess(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compartidas_acl.db")
	dbSuper := openTestSQLite(t, "super_compartidas_acl.db")
	ensureEmpresaCompartidaTestSchemas(t, dbEmp, dbSuper)

	seedEmpresaCompartidaAdmin(t, dbSuper, "owner@test.com", "Owner", "administrador", "")
	seedEmpresaCompartidaAdmin(t, dbSuper, "shared@test.com", "Shared", "administrador", "")
	seedEmpresaCompartidaEmpresa(t, dbEmp, 85, "Empresa ACL", "owner@test.com")

	if _, err := dbpkg.UpsertAdminEmpresaCompartidaAcceso(dbSuper, dbpkg.AdminEmpresaCompartidaAcceso{
		EmpresaID:          85,
		AdminEmail:         "shared@test.com",
		CompartidoPorEmail: "owner@test.com",
		FechaAceptada:      time.Now().Format("2006-01-02 15:04:05"),
		UsuarioCreador:     "owner@test.com",
		Estado:             "activo",
	}); err != nil {
		t.Fatalf("upsert shared access: %v", err)
	}

	allowed, err := dbpkg.CanAdminAccessEmpresaIA(dbEmp, dbSuper, "shared@test.com", 85)
	if err != nil {
		t.Fatalf("CanAdminAccessEmpresaIA returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected shared admin to have IA access to the shared empresa")
	}
}
