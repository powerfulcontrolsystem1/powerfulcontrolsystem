package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaUsuarioChangePasswordFlow(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_change_password.db")
	dbSuper := openTestSQLite(t, "super_change_password.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)

	oldSalt := "salt-change-old"
	oldHash := hashEmpresaUsuarioPassword("ClaveActual101", oldSalt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set, password_actualizada_en,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, 1, datetime('now','-45 day','localtime'), ?, 1, 'activo')`,
		"change@empresa.com", "Usuario Change", "vendedor", int64(44), "DOC-CHANGE", oldHash, oldSalt, int64(2),
	)
	if err != nil {
		t.Fatalf("seed user change password: %v", err)
	}

	h := EmpresaUsuarioChangePasswordHandler(dbEmp, dbSuper)
	body := `{"empresa_id":44,"email":"change@empresa.com","current_password":"ClaveActual101","new_password":"ClaveNueva202","new_password_confirm":"ClaveNueva202"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/cambiar_password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var updatedHash string
	var passwordSet int
	if err := dbEmp.QueryRow(`SELECT COALESCE(password_hash,''), COALESCE(password_set,0) FROM users WHERE email = ?`, "change@empresa.com").Scan(&updatedHash, &passwordSet); err != nil {
		t.Fatalf("query user after change password: %v", err)
	}
	if strings.TrimSpace(updatedHash) == "" || updatedHash == oldHash {
		t.Fatalf("expected password hash updated after change password old=%q new=%q", oldHash, updatedHash)
	}
	if passwordSet != 1 {
		t.Fatalf("expected password_set=1, got %d", passwordSet)
	}

	loginH := EmpresaUsuarioLoginHandler(dbEmp, dbSuper)
	loginOldReq := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(`{"empresa_id":44,"email":"change@empresa.com","password":"ClaveActual101"}`))
	loginOldReq.Header.Set("Content-Type", "application/json")
	loginOldRR := httptest.NewRecorder()
	loginH.ServeHTTP(loginOldRR, loginOldReq)
	if loginOldRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password login status %d, got %d body=%s", http.StatusUnauthorized, loginOldRR.Code, loginOldRR.Body.String())
	}

	loginNewReq := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(`{"empresa_id":44,"email":"change@empresa.com","password":"ClaveNueva202"}`))
	loginNewReq.Header.Set("Content-Type", "application/json")
	loginNewRR := httptest.NewRecorder()
	loginH.ServeHTTP(loginNewRR, loginNewReq)
	if loginNewRR.Code != http.StatusOK {
		t.Fatalf("expected new password login status %d, got %d body=%s", http.StatusOK, loginNewRR.Code, loginNewRR.Body.String())
	}
}

func TestEmpresaUsuarioChangePasswordPolicyRejectsWeakPassword(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_change_policy.db")
	dbSuper := openTestSQLite(t, "super_change_policy.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "usuarios.password_require_symbol", "1", false); err != nil {
		t.Fatalf("set policy require symbol: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "usuarios.password_min_length", "10", false); err != nil {
		t.Fatalf("set policy min length: %v", err)
	}

	salt := "salt-policy"
	hash := hashEmpresaUsuarioPassword("ClaveActual101", salt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, 1, 'activo')`,
		"policy@empresa.com", "Usuario Policy", "vendedor", int64(54), "DOC-POL", hash, salt, int64(2),
	)
	if err != nil {
		t.Fatalf("seed user policy change: %v", err)
	}

	h := EmpresaUsuarioChangePasswordHandler(dbEmp, dbSuper)
	body := `{"empresa_id":54,"email":"policy@empresa.com","current_password":"ClaveActual101","new_password":"ClaveNueva101","new_password_confirm":"ClaveNueva101"}`
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/cambiar_password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "símbolo") && !strings.Contains(strings.ToLower(rr.Body.String()), "simbolo") {
		t.Fatalf("expected weak password policy message, got body=%s", rr.Body.String())
	}
}

func TestEmpresaUsuarioLoginRequiresRotationWhenPolicyEnabled(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_login_rotation.db")
	dbSuper := openTestSQLite(t, "super_login_rotation.db")
	ensureEmpresaUsersSchema(t, dbEmp)
	ensureSuperSchema(t, dbSuper)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "usuarios.password_rotation_days", "30", false); err != nil {
		t.Fatalf("set policy rotation days: %v", err)
	}

	oldDate := time.Now().AddDate(0, 0, -45).Format("2006-01-02 15:04:05")
	salt := "salt-rotation"
	hash := hashEmpresaUsuarioPassword("PasswordSegura1", salt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad,
		password_hash, password_salt, password_set, password_actualizada_en,
		rol_usuario_id, email_confirmado, estado
	) VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?, ?, 1, 'activo')`,
		"rotation@empresa.com", "Usuario Rotation", "vendedor", int64(64), "DOC-ROT", hash, salt, oldDate, int64(2),
	)
	if err != nil {
		t.Fatalf("seed user rotation: %v", err)
	}

	h := EmpresaUsuarioLoginHandler(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(`{"empresa_id":64,"email":"rotation@empresa.com","password":"PasswordSegura1"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if required, _ := body["password_rotation_required"].(bool); !required {
		t.Fatalf("expected password_rotation_required=true got body=%v", body)
	}

	var sesionesCount int
	if err := dbSuper.QueryRow("SELECT COUNT(1) FROM sesiones").Scan(&sesionesCount); err != nil {
		t.Fatalf("count sesiones: %v", err)
	}
	if sesionesCount != 0 {
		t.Fatalf("expected 0 sessions when rotation is required, got %d", sesionesCount)
	}
}

func TestEmpresaUsuarioNotificationsCaptureInMailTestMode(t *testing.T) {
	t.Setenv("PCS_MAIL_TEST_MODE", "1")

	dbSuper := openTestSQLite(t, "super_mail_test_mode.db")
	ensureSuperSchema(t, dbSuper)

	reqConfirm := httptest.NewRequest(http.MethodPost, "http://localhost:8080/super/api/usuarios", nil)
	confirmURL, err := sendEmpresaUsuarioConfirmationEmail(reqConfirm, dbSuper, 77, "captura@empresa.com", "Usuario Captura", "token-confirm-001")
	if err != nil {
		t.Fatalf("send confirmation in test mode: %v", err)
	}
	if !strings.Contains(confirmURL, "auth/confirmar_correo") {
		t.Fatalf("expected confirm URL, got %q", confirmURL)
	}

	reqRecovery := httptest.NewRequest(http.MethodPost, "http://localhost:8080/api/empresa/usuarios/solicitar_recuperacion_password", nil)
	recoveryURL, err := sendEmpresaUsuarioPasswordRecoveryEmail(reqRecovery, dbSuper, 77, "captura@empresa.com", "Usuario Captura", "token-reset-001")
	if err != nil {
		t.Fatalf("send recovery in test mode: %v", err)
	}
	if !strings.Contains(recoveryURL, "login_usuario.html") {
		t.Fatalf("expected recovery URL, got %q", recoveryURL)
	}

	rows, err := dbpkg.ListSuperCorreoNotificacionesPrueba(dbSuper, dbpkg.SuperCorreoNotificacionPruebaFilter{EmpresaID: 77, Limit: 10})
	if err != nil {
		t.Fatalf("list captured mail notifications: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected at least 2 captured notifications, got %d", len(rows))
	}

	tipos := map[string]bool{}
	tokens := map[string]bool{}
	for _, row := range rows {
		tipos[row.Tipo] = true
		tokens[row.TokenRef] = true
	}
	if !tipos[dbpkg.SuperCorreoNotificacionTipoConfirmacion] {
		t.Fatalf("expected captured confirmation notification, rows=%v", rows)
	}
	if !tipos[dbpkg.SuperCorreoNotificacionTipoRecuperacion] {
		t.Fatalf("expected captured recovery notification, rows=%v", rows)
	}
	if !tokens["token-confirm-001"] || !tokens["token-reset-001"] {
		t.Fatalf("expected captured tokens in notifications, tokens=%v", tokens)
	}
}
