package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func ensureAdminAuthTestSchema(t *testing.T, dbSuper *sql.DB) {
	t.Helper()

	_, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT DEFAULT 'administrador',
		photo TEXT,
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
	);`)
	if err != nil {
		t.Fatalf("create administradores schema: %v", err)
	}

	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS sesiones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		admin_email TEXT,
		token TEXT,
		ip TEXT,
		user_agent TEXT,
		fecha_inicio TEXT,
		fecha_fin TEXT,
		fecha_creacion TEXT,
		activo INTEGER DEFAULT 1
	);`)
	if err != nil {
		t.Fatalf("create sesiones schema: %v", err)
	}
}

func TestAdminRegisterHandlerCreatesPendingAdminAndCapturesConfirmationMail(t *testing.T) {
	t.Setenv("PCS_MAIL_TEST_MODE", "1")

	dbSuper := openTestSQLite(t, "admin_register_handler.db")
	ensureAdminAuthTestSchema(t, dbSuper)

	body := `{"email":"nuevo_admin@empresa.com","name":"Nuevo Administrador","telefono":"3001234567","pais":"Colombia","ciudad":"Bogota","password":"ClaveSegura99"}`
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/super/api/administradores/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	AdminRegisterHandler(dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	admin, err := dbpkg.GetAdminByEmailFull(dbSuper, "nuevo_admin@empresa.com")
	if err != nil {
		t.Fatalf("get admin by email: %v", err)
	}
	if admin == nil {
		t.Fatal("expected created admin")
	}
	if admin.Name != "Nuevo Administrador" {
		t.Fatalf("expected name Nuevo Administrador, got %q", admin.Name)
	}
	if admin.Telefono != "3001234567" {
		t.Fatalf("expected telefono 3001234567, got %q", admin.Telefono)
	}
	if admin.Pais != "Colombia" {
		t.Fatalf("expected pais Colombia, got %q", admin.Pais)
	}
	if admin.Ciudad != "Bogota" {
		t.Fatalf("expected ciudad Bogota, got %q", admin.Ciudad)
	}
	if admin.PasswordSet != 1 || strings.TrimSpace(admin.PasswordHash) == "" {
		t.Fatalf("expected password_set=1 with hash, got %+v", admin)
	}
	if admin.EmailConfirmado != 0 || strings.TrimSpace(admin.EmailConfirmToken) == "" {
		t.Fatalf("expected pending confirmation token, got %+v", admin)
	}

	notifications, err := dbpkg.ListSuperCorreoNotificacionesPrueba(dbSuper, dbpkg.SuperCorreoNotificacionPruebaFilter{Destinatario: "nuevo_admin@empresa.com", Limit: 10})
	if err != nil {
		t.Fatalf("list captured notifications: %v", err)
	}
	if len(notifications) == 0 {
		t.Fatal("expected at least one captured notification")
	}
	if notifications[0].Tipo != "confirmacion_correo_admin" {
		t.Fatalf("expected confirmacion_correo_admin, got %q", notifications[0].Tipo)
	}
	if !strings.Contains(notifications[0].Cuerpo, "/auth/confirmar_admin?token=") {
		t.Fatalf("expected admin confirm URL in captured body, got %q", notifications[0].Cuerpo)
	}
}

func TestAdminRegisterHandlerRejectsConfirmedExistingAdmin(t *testing.T) {
	dbSuper := openTestSQLite(t, "admin_register_conflict.db")
	ensureAdminAuthTestSchema(t, dbSuper)

	if err := dbpkg.UpsertAdministrador(dbSuper, "existente@empresa.com", "Administrador Existente", "administrador", ""); err != nil {
		t.Fatalf("upsert admin: %v", err)
	}
	if _, err := dbSuper.Exec(`UPDATE administradores SET email_confirmado = 1 WHERE lower(email) = lower(?)`, "existente@empresa.com"); err != nil {
		t.Fatalf("confirm existing admin: %v", err)
	}

	body := `{"email":"existente@empresa.com","name":"Administrador Existente","telefono":"3001234567","pais":"Colombia","ciudad":"Bogota","password":"ClaveSegura99"}`
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/super/api/administradores/register", strings.NewReader(body))
	rr := httptest.NewRecorder()

	AdminRegisterHandler(dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}
}

func TestAdminLoginHandlerCreatesSessionForConfirmedAdmin(t *testing.T) {
	dbSuper := openTestSQLite(t, "admin_login_handler.db")
	ensureAdminAuthTestSchema(t, dbSuper)

	if err := dbpkg.UpsertAdministrador(dbSuper, "login_admin@empresa.com", "Login Admin", "super_administrador", ""); err != nil {
		t.Fatalf("upsert admin: %v", err)
	}
	hash, salt, err := generateEmpresaUsuarioPasswordHash("ClaveSegura99")
	if err != nil {
		t.Fatalf("generate hash: %v", err)
	}
	if err := dbpkg.SetAdministradorPassword(dbSuper, "login_admin@empresa.com", hash, salt); err != nil {
		t.Fatalf("set admin password: %v", err)
	}
	if _, err := dbSuper.Exec(`UPDATE administradores SET email_confirmado = 1 WHERE lower(email) = lower(?)`, "login_admin@empresa.com"); err != nil {
		t.Fatalf("confirm admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/super/api/administradores/login", strings.NewReader(`{"email":"login_admin@empresa.com","password":"ClaveSegura99"}`))
	rr := httptest.NewRecorder()

	AdminLoginHandler(dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode login response: %v body=%s", err, rr.Body.String())
	}
	if got := body["redirect_url"]; got != "/super_administrador.html" {
		t.Fatalf("expected super redirect, got %v", got)
	}

	var sessionsCount int
	if err := dbSuper.QueryRow(`SELECT COUNT(1) FROM sesiones WHERE lower(admin_email) = lower(?)`, "login_admin@empresa.com").Scan(&sessionsCount); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionsCount != 1 {
		t.Fatalf("expected 1 session, got %d", sessionsCount)
	}
	if !strings.Contains(rr.Header().Get("Set-Cookie"), "session_token=") {
		t.Fatalf("expected session cookie, got headers=%v", rr.Header())
	}
}

func TestAdminRequestAndResetPasswordHandlersUseCapturedMailAndCreateSession(t *testing.T) {
	t.Setenv("PCS_MAIL_TEST_MODE", "1")

	dbSuper := openTestSQLite(t, "admin_reset_handler.db")
	ensureAdminAuthTestSchema(t, dbSuper)

	if err := dbpkg.UpsertAdministrador(dbSuper, "reset_admin@empresa.com", "Reset Admin", "administrador", ""); err != nil {
		t.Fatalf("upsert admin: %v", err)
	}
	hash, salt, err := generateEmpresaUsuarioPasswordHash("ClaveAnterior88")
	if err != nil {
		t.Fatalf("generate initial hash: %v", err)
	}
	if err := dbpkg.SetAdministradorPassword(dbSuper, "reset_admin@empresa.com", hash, salt); err != nil {
		t.Fatalf("set admin password: %v", err)
	}
	if _, err := dbSuper.Exec(`UPDATE administradores SET email_confirmado = 1 WHERE lower(email) = lower(?)`, "reset_admin@empresa.com"); err != nil {
		t.Fatalf("confirm admin: %v", err)
	}

	reqRecovery := httptest.NewRequest(http.MethodPost, "http://localhost:8080/super/api/administradores/solicitar_recuperacion", strings.NewReader(`{"email":"reset_admin@empresa.com"}`))
	rrRecovery := httptest.NewRecorder()
	AdminRequestPasswordRecoveryHandler(dbSuper).ServeHTTP(rrRecovery, reqRecovery)

	if rrRecovery.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrRecovery.Code, rrRecovery.Body.String())
	}

	admin, err := dbpkg.GetAdminByEmailFull(dbSuper, "reset_admin@empresa.com")
	if err != nil {
		t.Fatalf("reload admin after recovery request: %v", err)
	}
	if strings.TrimSpace(admin.PasswordResetToken) == "" {
		t.Fatalf("expected recovery token after request, got %+v", admin)
	}

	notifications, err := dbpkg.ListSuperCorreoNotificacionesPrueba(dbSuper, dbpkg.SuperCorreoNotificacionPruebaFilter{Destinatario: "reset_admin@empresa.com", Limit: 10})
	if err != nil {
		t.Fatalf("list captured notifications: %v", err)
	}
	if len(notifications) == 0 {
		t.Fatal("expected captured recovery notification")
	}
	if notifications[0].Tipo != "recuperacion_password_admin" {
		t.Fatalf("expected recuperacion_password_admin, got %q", notifications[0].Tipo)
	}
	if !strings.Contains(notifications[0].Cuerpo, "view=reset") {
		t.Fatalf("expected reset URL with view=reset, got %q", notifications[0].Cuerpo)
	}

	reqReset := httptest.NewRequest(http.MethodPost, "http://localhost:8080/super/api/administradores/restablecer_password", strings.NewReader(`{"email":"reset_admin@empresa.com","token":"`+admin.PasswordResetToken+`","password":"NuevaClave99"}`))
	rrReset := httptest.NewRecorder()
	AdminResetPasswordHandler(dbSuper).ServeHTTP(rrReset, reqReset)

	if rrReset.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrReset.Code, rrReset.Body.String())
	}

	adminAfterReset, err := dbpkg.GetAdminByEmailFull(dbSuper, "reset_admin@empresa.com")
	if err != nil {
		t.Fatalf("reload admin after reset: %v", err)
	}
	if strings.TrimSpace(adminAfterReset.PasswordResetToken) != "" {
		t.Fatalf("expected cleared reset token, got %+v", adminAfterReset)
	}
	if hashEmpresaUsuarioPassword("NuevaClave99", adminAfterReset.PasswordSalt) != strings.TrimSpace(adminAfterReset.PasswordHash) {
		t.Fatalf("expected updated password hash, got %+v", adminAfterReset)
	}

	var sessionsCount int
	if err := dbSuper.QueryRow(`SELECT COUNT(1) FROM sesiones WHERE lower(admin_email) = lower(?)`, "reset_admin@empresa.com").Scan(&sessionsCount); err != nil {
		t.Fatalf("count sessions after reset: %v", err)
	}
	if sessionsCount != 1 {
		t.Fatalf("expected 1 session after reset, got %d", sessionsCount)
	}
	if !strings.Contains(rrReset.Header().Get("Set-Cookie"), "session_token=") {
		t.Fatalf("expected session cookie after reset, got headers=%v", rrReset.Header())
	}
}

func TestAccountSetGooglePasswordHandlerCreatesInitialPassword(t *testing.T) {
	dbSuper := openTestSQLite(t, "admin_google_password_setup.db")
	ensureAdminAuthTestSchema(t, dbSuper)

	if err := dbpkg.UpsertAdministrador(dbSuper, "google_admin@empresa.com", "Google Admin", "administrador", ""); err != nil {
		t.Fatalf("upsert admin: %v", err)
	}
	if _, err := dbSuper.Exec(`UPDATE administradores SET email_confirmado = 1, password_set = 0, password_hash = '', password_salt = '' WHERE lower(email) = lower(?)`, "google_admin@empresa.com"); err != nil {
		t.Fatalf("prepare admin: %v", err)
	}
	if err := dbpkg.CreateSession(dbSuper, "google_admin@empresa.com", "127.0.0.1:1234", "test-agent", "token-google-setup"); err != nil {
		t.Fatalf("create session: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/account/set_google_password", strings.NewReader(`{"password":"NuevaClave99","password_confirm":"NuevaClave99"}`))
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "token-google-setup"})
	rr := httptest.NewRecorder()

	AccountSetGooglePasswordHandler(nil, dbSuper).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	admin, err := dbpkg.GetAdminByEmailFull(dbSuper, "google_admin@empresa.com")
	if err != nil {
		t.Fatalf("reload admin: %v", err)
	}
	if admin.PasswordSet != 1 || strings.TrimSpace(admin.PasswordHash) == "" || strings.TrimSpace(admin.PasswordSalt) == "" {
		t.Fatalf("expected password configured, got %+v", admin)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got := body["redirect_url"]; got != "/seleccionar_empresa.html" {
		t.Fatalf("expected seleccionar_empresa redirect, got %v", got)
	}
}
