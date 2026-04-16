package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

func ensureEmpresasCoreSchemaForSuper(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	_, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre TEXT,
		nit TEXT,
		tipo_id INTEGER,
		tipo_nombre TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`)
	if err != nil {
		t.Fatalf("create empresas schema: %v", err)
	}
}

func ensureEmpresasImpactSchemaForSuper(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	_, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		email TEXT,
		estado TEXT DEFAULT 'activo'
	);`)
	if err != nil {
		t.Fatalf("create users impact schema: %v", err)
	}

	_, err = dbEmp.Exec(`CREATE TABLE IF NOT EXISTS carritos_compras (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		estado_carrito TEXT,
		estado TEXT DEFAULT 'activo'
	);`)
	if err != nil {
		t.Fatalf("create carritos impact schema: %v", err)
	}

	_, err = dbEmp.Exec(`CREATE TABLE IF NOT EXISTS reservas_hotel (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		estado_reserva TEXT,
		estado TEXT DEFAULT 'activo'
	);`)
	if err != nil {
		t.Fatalf("create reservas impact schema: %v", err)
	}
}

func ensureSuperConfigSchemaForSuper(t *testing.T, dbSuper *sql.DB) {
	t.Helper()
	_, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS configuraciones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_key TEXT UNIQUE,
		value TEXT,
		encrypted INTEGER DEFAULT 0,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT
	);`)
	if err != nil {
		t.Fatalf("create configuraciones schema: %v", err)
	}

	_, err = dbSuper.Exec(`CREATE TABLE IF NOT EXISTS licencias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		activo INTEGER DEFAULT 1,
		fecha_fin TEXT
	);`)
	if err != nil {
		t.Fatalf("create licencias schema: %v", err)
	}
}

func seedEmpresaEstadoForSuper(t *testing.T, dbEmp *sql.DB, id int64, nombre, estado string) {
	t.Helper()
	_, err := dbEmp.Exec(`
		INSERT INTO empresas (id, empresa_id, nombre, estado, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))
	`, id, id, nombre, estado)
	if err != nil {
		t.Fatalf("insert empresa seed: %v", err)
	}
}

func TestEmpresasHandlerDesactivarConImpactoYForce(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_super_estado.db")
	dbSuper := openTestSQLite(t, "super_super_estado.db")

	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureEmpresasImpactSchemaForSuper(t, dbEmp)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	seedEmpresaEstadoForSuper(t, dbEmp, 1, "Empresa Uno", "activo")
	if _, err := dbEmp.Exec(`INSERT INTO users (empresa_id, email, estado) VALUES (1, 'user@empresa.com', 'activo')`); err != nil {
		t.Fatalf("insert users impact: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO carritos_compras (empresa_id, estado_carrito, estado) VALUES (1, 'abierto', 'activo')`); err != nil {
		t.Fatalf("insert carritos impact: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO reservas_hotel (empresa_id, estado_reserva, estado) VALUES (1, 'confirmada', 'activo')`); err != nil {
		t.Fatalf("insert reservas impact: %v", err)
	}
	if _, err := dbSuper.Exec(`INSERT INTO licencias (empresa_id, activo) VALUES (1, 1)`); err != nil {
		t.Fatalf("insert licencias impact: %v", err)
	}

	h := EmpresasHandler(dbEmp, dbSuper)

	req := httptest.NewRequest(http.MethodPut, "/super/api/empresas?id=1&action=desactivar", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d without force, got %d body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}

	var conflictBody map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &conflictBody); err != nil {
		t.Fatalf("decode conflict body: %v body=%s", err, rr.Body.String())
	}
	if ok, _ := conflictBody["requiere_confirmacion"].(bool); !ok {
		t.Fatalf("expected requiere_confirmacion=true, got %v", conflictBody["requiere_confirmacion"])
	}

	var estado string
	if err := dbEmp.QueryRow(`SELECT estado FROM empresas WHERE id = 1`).Scan(&estado); err != nil {
		t.Fatalf("query empresa estado after conflict: %v", err)
	}
	if strings.TrimSpace(strings.ToLower(estado)) != "activo" {
		t.Fatalf("estado should remain activo after conflict, got %q", estado)
	}

	forceReq := httptest.NewRequest(http.MethodPut, "/super/api/empresas?id=1&action=desactivar&force=1", nil)
	forceRR := httptest.NewRecorder()
	h.ServeHTTP(forceRR, forceReq)

	if forceRR.Code != http.StatusOK {
		t.Fatalf("expected status %d with force, got %d body=%s", http.StatusOK, forceRR.Code, forceRR.Body.String())
	}

	if err := dbEmp.QueryRow(`SELECT estado FROM empresas WHERE id = 1`).Scan(&estado); err != nil {
		t.Fatalf("query empresa estado after force: %v", err)
	}
	if strings.TrimSpace(strings.ToLower(estado)) != "inactivo" {
		t.Fatalf("estado should be inactivo after force deactivation, got %q", estado)
	}
}

func TestEmpresasHandlerImpactoDesactivacion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_super_impacto.db")
	dbSuper := openTestSQLite(t, "super_super_impacto.db")

	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureEmpresasImpactSchemaForSuper(t, dbEmp)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	seedEmpresaEstadoForSuper(t, dbEmp, 5, "Empresa Impacto", "activo")
	if _, err := dbEmp.Exec(`INSERT INTO users (empresa_id, email, estado) VALUES (5, 'u1@empresa.com', 'activo')`); err != nil {
		t.Fatalf("insert users impacto: %v", err)
	}

	h := EmpresasHandler(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/super/api/empresas?id=5&action=impacto_desactivacion", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode impacto response: %v body=%s", err, rr.Body.String())
	}
	impacto, ok := body["impacto"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected impacto object, got %T", body["impacto"])
	}
	if got, _ := impacto["empresa_id"].(float64); int64(got) != 5 {
		t.Fatalf("expected impacto.empresa_id=5, got %v", impacto["empresa_id"])
	}
	if got, _ := impacto["usuarios_activos"].(float64); int64(got) != 1 {
		t.Fatalf("expected impacto.usuarios_activos=1, got %v", impacto["usuarios_activos"])
	}
}

func TestSuperConfigBackupHandlerExportYRestore(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_config_backup.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "wompi.public_key", "pub_test_demo", false); err != nil {
		t.Fatalf("seed wompi.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.mode", "sandbox", false); err != nil {
		t.Fatalf("seed wompi.mode: %v", err)
	}

	h := SuperConfigBackupHandler(dbSuper)
	exportReq := httptest.NewRequest(http.MethodGet, "/super/api/config/backup", nil)
	exportRR := httptest.NewRecorder()
	h.ServeHTTP(exportRR, exportReq)

	if exportRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on export, got %d body=%s", http.StatusOK, exportRR.Code, exportRR.Body.String())
	}

	var backup superConfigBackupPayload
	if err := json.Unmarshal(exportRR.Body.Bytes(), &backup); err != nil {
		t.Fatalf("decode backup export: %v body=%s", err, exportRR.Body.String())
	}
	if backup.Version != superConfigBackupVersion {
		t.Fatalf("expected backup version %q, got %q", superConfigBackupVersion, backup.Version)
	}

	updatedMode := false
	for i := range backup.Items {
		if strings.TrimSpace(backup.Items[i].Key) == "wompi.mode" {
			backup.Items[i].Value = "production"
			backup.Items[i].Configured = true
			updatedMode = true
			break
		}
	}
	if !updatedMode {
		t.Fatalf("expected wompi.mode key inside backup payload")
	}

	rawRestore, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("encode restore payload: %v", err)
	}

	restoreReq := httptest.NewRequest(http.MethodPut, "/super/api/config/backup", strings.NewReader(string(rawRestore)))
	restoreReq.Header.Set("Content-Type", "application/json")
	restoreRR := httptest.NewRecorder()
	h.ServeHTTP(restoreRR, restoreReq)

	if restoreRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on restore, got %d body=%s", http.StatusOK, restoreRR.Code, restoreRR.Body.String())
	}

	modeValue, _, err := dbpkg.GetConfigValue(dbSuper, "wompi.mode")
	if err != nil {
		t.Fatalf("read restored wompi.mode: %v", err)
	}
	if strings.TrimSpace(modeValue) != "production" {
		t.Fatalf("expected wompi.mode restored to production, got %q", modeValue)
	}
}

func TestSuperConfigBackupHandlerRestoreEncryptsSensitivePlaintext(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(i + 11)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_config_backup_encrypt_sensitive.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	payload := superConfigBackupPayload{
		Version:   superConfigBackupVersion,
		Scope:     "super_config_critica",
		CreatedAt: "2026-04-08T00:00:00Z",
		Items: []superConfigBackupItem{
			{
				Key:        "ai.model.deepseek.deepseek_chat.api_key",
				Value:      "deepseek_test_secret",
				Encrypted:  false,
				Configured: true,
			},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal restore payload: %v", err)
	}

	h := SuperConfigBackupHandler(dbSuper)
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/backup", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on restore, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "ai.model.deepseek.deepseek_chat.api_key")
	if err != nil {
		t.Fatalf("read restored sensitive key: %v", err)
	}
	if !encrypted {
		t.Fatal("expected restored sensitive key to be marked as encrypted")
	}
	if strings.TrimSpace(stored) == "deepseek_test_secret" {
		t.Fatal("expected stored sensitive key to be encrypted payload, got plaintext")
	}

	decrypted, decErr := utils.DecryptString(stored)
	if decErr != nil {
		t.Fatalf("decrypt restored sensitive key: %v", decErr)
	}
	if decrypted != "deepseek_test_secret" {
		t.Fatalf("expected decrypted sensitive value %q, got %q", "deepseek_test_secret", decrypted)
	}
}

func TestSuperConfigBackupHandlerRestoreSensitivePlaintextRequiresEncryptionKey(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", "")

	dbSuper := openTestSQLite(t, "super_config_backup_encrypt_required.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	payload := superConfigBackupPayload{
		Version:   superConfigBackupVersion,
		Scope:     "super_config_critica",
		CreatedAt: "2026-04-08T00:00:00Z",
		Items: []superConfigBackupItem{
			{
				Key:        "gmail.smtp_app_password",
				Value:      "app-password-demo",
				Encrypted:  false,
				Configured: true,
			},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal restore payload: %v", err)
	}

	h := SuperConfigBackupHandler(dbSuper)
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/backup", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d when CONFIG_ENC_KEY is missing, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestAIModelsConfigHandlerSaveDeepSeekEncrypted(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(77 + i)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_ai_config_handler.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := AIModelsConfigHandler(dbSuper)
	body := `{"credentials":[{"model_id":"deepseek:deepseek-chat","api_key":"sk_deepseek_prueba"}]}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/ai", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Email", "super@empresa.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on ai save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "ai.model.deepseek.deepseek_chat.api_key")
	if err != nil {
		t.Fatalf("read ai.model.deepseek.deepseek_chat.api_key: %v", err)
	}
	if !encrypted {
		t.Fatal("expected DeepSeek model key to be encrypted")
	}

	decrypted, decErr := utils.DecryptString(stored)
	if decErr != nil {
		t.Fatalf("decrypt DeepSeek model key: %v", decErr)
	}
	if decrypted != "sk_deepseek_prueba" {
		t.Fatalf("expected decrypted DeepSeek key %q, got %q", "sk_deepseek_prueba", decrypted)
	}

	providerStored, providerEncrypted, err := dbpkg.GetConfigValue(dbSuper, "ai.provider.deepseek.api_key")
	if err != nil {
		t.Fatalf("read ai.provider.deepseek.api_key: %v", err)
	}
	if !providerEncrypted {
		t.Fatal("expected provider DeepSeek key to be encrypted")
	}
	providerDecrypted, providerDecErr := utils.DecryptString(providerStored)
	if providerDecErr != nil {
		t.Fatalf("decrypt provider DeepSeek key: %v", providerDecErr)
	}
	if providerDecrypted != "sk_deepseek_prueba" {
		t.Fatalf("expected decrypted provider DeepSeek key %q, got %q", "sk_deepseek_prueba", providerDecrypted)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/ai", nil)
	getReq.Header.Set("X-Admin-Email", "super@empresa.com")
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on ai get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode ai get response: %v body=%s", err, getRR.Body.String())
	}
	modelos, ok := getBody["modelos"].([]interface{})
	if !ok || len(modelos) == 0 {
		t.Fatalf("expected modelos in ai get response, got %T", getBody["modelos"])
	}
	first, ok := modelos[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected first modelo object, got %T", modelos[0])
	}
	if masked, _ := first["masked"].(string); masked != "********" {
		t.Fatalf("expected masked value ********, got %q", masked)
	}
}

func TestGmailConfigHandlerSaveRestartAlertTo(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_gmail_restart_alert.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := GmailConfigHandler(dbSuper)
	body := `{"restart_alert_to":"ops@empresa.com"}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/gmail", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "gmail.restart_alert_to")
	if err != nil {
		t.Fatalf("read gmail.restart_alert_to: %v", err)
	}
	if strings.TrimSpace(stored) != "ops@empresa.com" {
		t.Fatalf("expected gmail.restart_alert_to %q, got %q", "ops@empresa.com", stored)
	}
	if encrypted {
		t.Fatal("expected gmail.restart_alert_to to be non-encrypted")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/gmail", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode gmail get response: %v body=%s", err, getRR.Body.String())
	}
	if got := strings.TrimSpace(fmt.Sprint(getBody["restart_alert_to"])); got != "ops@empresa.com" {
		t.Fatalf("expected restart_alert_to in response %q, got %q", "ops@empresa.com", got)
	}
	if setFlag, _ := getBody["restart_alert_to_set"].(bool); !setFlag {
		t.Fatalf("expected restart_alert_to_set=true, got %v", getBody["restart_alert_to_set"])
	}
}

func TestGmailConfigHandlerSaveRestartAlertToggle(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_gmail_restart_alert_toggle.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := GmailConfigHandler(dbSuper)
	body := `{"restart_alert_enabled":false}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/gmail", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail toggle save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "gmail.restart_alert_enabled")
	if err != nil {
		t.Fatalf("read gmail.restart_alert_enabled: %v", err)
	}
	if strings.TrimSpace(stored) != "0" {
		t.Fatalf("expected gmail.restart_alert_enabled %q, got %q", "0", stored)
	}
	if encrypted {
		t.Fatal("expected gmail.restart_alert_enabled to be non-encrypted")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/gmail", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode gmail get response: %v body=%s", err, getRR.Body.String())
	}
	if enabled, _ := getBody["restart_alert_enabled"].(bool); enabled {
		t.Fatalf("expected restart_alert_enabled=false, got %v", getBody["restart_alert_enabled"])
	}
	if setFlag, _ := getBody["restart_alert_enabled_set"].(bool); !setFlag {
		t.Fatalf("expected restart_alert_enabled_set=true, got %v", getBody["restart_alert_enabled_set"])
	}
}

func TestPublicLicenciasPaymentMethodsHandlerOrdersAndAvailability(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(31 + i)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_payment_methods_status.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "epayco.public_key", "pub_test_demo", false); err != nil {
		t.Fatalf("seed epayco.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.customer_id", "1579238", false); err != nil {
		t.Fatalf("seed epayco.customer_id: %v", err)
	}
	encEpaycoKey, err := utils.EncryptString("epayco_secret_demo")
	if err != nil {
		t.Fatalf("encrypt epayco.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.private_key", encEpaycoKey, true); err != nil {
		t.Fatalf("seed epayco.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.enabled", "1", false); err != nil {
		t.Fatalf("seed epayco.enabled: %v", err)
	}

	if err := dbpkg.SetConfigValue(dbSuper, "wompi.public_key", "pub_test_demo", false); err != nil {
		t.Fatalf("seed wompi.public_key: %v", err)
	}
	encWompiPrivate, err := utils.EncryptString("prv_test_demo")
	if err != nil {
		t.Fatalf("encrypt wompi.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.private_key", encWompiPrivate, true); err != nil {
		t.Fatalf("seed wompi.private_key: %v", err)
	}
	encWompiIntegrity, err := utils.EncryptString("test_integrity_demo")
	if err != nil {
		t.Fatalf("encrypt wompi.integrity_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.integrity_key", encWompiIntegrity, true); err != nil {
		t.Fatalf("seed wompi.integrity_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.enabled", "0", false); err != nil {
		t.Fatalf("seed wompi.enabled: %v", err)
	}

	h := PublicLicenciasPaymentMethodsHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/api/public/licencias/payment_methods", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body struct {
		Providers []struct {
			ID         string `json:"id"`
			Enabled    bool   `json:"enabled"`
			Configured bool   `json:"configured"`
			Available  bool   `json:"available"`
		} `json:"providers"`
		DefaultMethod string `json:"default_method"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode payment methods response: %v body=%s", err, rr.Body.String())
	}
	if len(body.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(body.Providers))
	}
	if body.Providers[0].ID != "epayco" || body.Providers[1].ID != "wompi" {
		t.Fatalf("expected order epayco,wompi got %+v", body.Providers)
	}
	if !body.Providers[0].Enabled || !body.Providers[0].Configured || !body.Providers[0].Available {
		t.Fatalf("expected epayco available, got %+v", body.Providers[0])
	}
	if !body.Providers[1].Configured || body.Providers[1].Available || body.Providers[1].Enabled {
		t.Fatalf("expected wompi configured but disabled, got %+v", body.Providers[1])
	}
	if body.DefaultMethod != "epayco" {
		t.Fatalf("expected default_method epayco, got %q", body.DefaultMethod)
	}
}

func TestPublicLicenciasPaymentMethodsHandlerAllowsEpaycoWithPublicKeyOnly(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_payment_methods_epayco_public_only.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "epayco.public_key", "pub_test_only_public", false); err != nil {
		t.Fatalf("seed epayco.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.enabled", "1", false); err != nil {
		t.Fatalf("seed epayco.enabled: %v", err)
	}

	h := PublicLicenciasPaymentMethodsHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/api/public/licencias/payment_methods", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body struct {
		Providers []struct {
			ID         string `json:"id"`
			Configured bool   `json:"configured"`
			Available  bool   `json:"available"`
		} `json:"providers"`
		DefaultMethod string `json:"default_method"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode payment methods response: %v body=%s", err, rr.Body.String())
	}
	if len(body.Providers) == 0 {
		t.Fatalf("expected providers in response, got %s", rr.Body.String())
	}
	if body.Providers[0].ID != "epayco" || !body.Providers[0].Configured || !body.Providers[0].Available {
		t.Fatalf("expected epayco available with public key only, got %+v", body.Providers[0])
	}
	if body.DefaultMethod != "epayco" {
		t.Fatalf("expected default_method epayco, got %q", body.DefaultMethod)
	}
}

func TestWompiConfigHandlerPersistsEnabledFlag(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_wompi_enabled_toggle.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := WompiConfigHandler(dbSuper)
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/wompi", strings.NewReader(`{"enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on wompi enabled save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "wompi.enabled")
	if err != nil {
		t.Fatalf("read wompi.enabled: %v", err)
	}
	if encrypted {
		t.Fatal("expected wompi.enabled to be non-encrypted")
	}
	if strings.TrimSpace(stored) != "1" {
		t.Fatalf("expected wompi.enabled stored as 1, got %q", stored)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/wompi", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on wompi get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode wompi get response: %v body=%s", err, getRR.Body.String())
	}
	if enabled, _ := body["enabled"].(bool); !enabled {
		t.Fatalf("expected enabled=true in wompi get response, got %v", body["enabled"])
	}
}

func TestWompiTermsHandlerRejectsWhenDisabled(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_wompi_terms_disabled.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "wompi.enabled", "0", false); err != nil {
		t.Fatalf("seed wompi.enabled: %v", err)
	}

	h := WompiTermsHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/wompi/terms", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusPreconditionFailed, rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode wompi disabled response: %v body=%s", err, rr.Body.String())
	}
	if got := strings.TrimSpace(fmt.Sprint(body["provider"])); got != "wompi" {
		t.Fatalf("expected provider wompi, got %q", got)
	}
}

func TestSuperEndpointsPermisosPorRol(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_super_roles.db")
	dbSuper := openTestSQLite(t, "super_super_roles.db")
	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureSuperSchema(t, dbSuper)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	mux := http.NewServeMux()
	mux.HandleFunc("/super/api/empresas", EmpresasHandler(dbEmp, dbSuper))
	mux.HandleFunc("/super/api/config/wompi", WompiConfigHandler(dbSuper))
	mux.HandleFunc("/super/api/config/gmail", GmailConfigHandler(dbSuper))
	mux.HandleFunc("/super/api/config/ai", AIModelsConfigHandler(dbSuper))
	mux.HandleFunc("/super/api/config/backup", SuperConfigBackupHandler(dbSuper))

	protected := utils.AuthMiddleware(dbSuper, mux)

	if err := dbpkg.UpsertAdministrador(dbSuper, "super@empresa.com", "Super", "super_administrador", ""); err != nil {
		t.Fatalf("upsert super admin: %v", err)
	}
	if err := dbpkg.CreateSession(dbSuper, "super@empresa.com", "127.0.0.1", "test-agent", "token-super"); err != nil {
		t.Fatalf("create super session: %v", err)
	}

	rolesBloqueados := []string{"administrador", "contabilidad", "cajero", "supervisor_sucursal", "auditor"}
	for i, role := range rolesBloqueados {
		email := fmt.Sprintf("role-%d@empresa.com", i+1)
		token := fmt.Sprintf("token-%d", i+1)
		if err := dbpkg.UpsertAdministrador(dbSuper, email, "RoleUser", role, ""); err != nil {
			t.Fatalf("upsert role admin %s: %v", role, err)
		}
		if err := dbpkg.CreateSession(dbSuper, email, "127.0.0.1", "test-agent", token); err != nil {
			t.Fatalf("create role session %s: %v", role, err)
		}
	}

	endpoints := []string{
		"/super/api/empresas",
		"/super/api/config/wompi",
		"/super/api/config/gmail",
		"/super/api/config/ai",
		"/super/api/config/backup",
	}

	for _, endpoint := range endpoints {
		reqNoAuth := httptest.NewRequest(http.MethodGet, endpoint, nil)
		rrNoAuth := httptest.NewRecorder()
		protected.ServeHTTP(rrNoAuth, reqNoAuth)
		if rrNoAuth.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 without token for %s, got %d body=%s", endpoint, rrNoAuth.Code, rrNoAuth.Body.String())
		}

		reqSuper := httptest.NewRequest(http.MethodGet, endpoint, nil)
		reqSuper.AddCookie(&http.Cookie{Name: "session_token", Value: "token-super"})
		rrSuper := httptest.NewRecorder()
		protected.ServeHTTP(rrSuper, reqSuper)
		if rrSuper.Code != http.StatusOK {
			t.Fatalf("expected 200 for super_admin in %s, got %d body=%s", endpoint, rrSuper.Code, rrSuper.Body.String())
		}

		for i := range rolesBloqueados {
			token := fmt.Sprintf("token-%d", i+1)
			reqRole := httptest.NewRequest(http.MethodGet, endpoint, nil)
			reqRole.AddCookie(&http.Cookie{Name: "session_token", Value: token})
			rrRole := httptest.NewRecorder()
			protected.ServeHTTP(rrRole, reqRole)
			if rrRole.Code != http.StatusForbidden {
				t.Fatalf("expected 403 for role=%s in %s, got %d body=%s", rolesBloqueados[i], endpoint, rrRole.Code, rrRole.Body.String())
			}
		}
	}
}
