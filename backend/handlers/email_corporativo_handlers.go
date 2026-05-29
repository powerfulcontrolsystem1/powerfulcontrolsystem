package handlers

import (
	"crypto/tls"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	corporateEmailEnabledKey       = "email_corporativo.enabled"
	corporateEmailAutoCreateKey    = "email_corporativo.auto_create"
	corporateEmailDomainKey        = "email_corporativo.domain"
	corporateEmailWebmailURLKey    = "email_corporativo.webmail_url"
	corporateEmailProvisionModeKey = "email_corporativo.provision_mode"
	corporateEmailAPIBaseURLKey    = "email_corporativo.iredadmin_api_base_url"
	corporateEmailAPIAdminKey      = "email_corporativo.iredadmin_admin"
	corporateEmailAPIPasswordKey   = "email_corporativo.iredadmin_password"
	corporateEmailQuotaMBKey       = "email_corporativo.quota_mb"
)

var errIredAdminAPINotAvailable = errors.New("iRedAdmin-Pro API no esta disponible en la URL configurada")

type CorporateEmailConfig struct {
	Enabled        bool   `json:"enabled"`
	AutoCreate     bool   `json:"auto_create"`
	Domain         string `json:"domain"`
	WebmailURL     string `json:"webmail_url"`
	ProvisionMode  string `json:"provision_mode"`
	APIBaseURL     string `json:"iredadmin_api_base_url"`
	APIAdmin       string `json:"iredadmin_admin"`
	APIPasswordSet bool   `json:"iredadmin_password_set"`
	QuotaMB        int    `json:"quota_mb"`
}

type corporateEmailProvisionResult struct {
	OK     bool   `json:"ok"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type corporateWebmailCheck struct {
	Checked bool   `json:"checked"`
	OK      bool   `json:"ok"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type corporateEmailDiagnostics struct {
	Enabled              bool                  `json:"enabled"`
	AutoCreate           bool                  `json:"auto_create"`
	ProvisionMode        string                `json:"provision_mode"`
	IredAdminAPIEnabled  bool                  `json:"iredadmin_api_enabled"`
	IredAdminAPIURLSet   bool                  `json:"iredadmin_api_url_set"`
	IredAdminAdminSet    bool                  `json:"iredadmin_admin_set"`
	IredAdminPasswordSet bool                  `json:"iredadmin_password_set"`
	EncryptionAvailable  bool                  `json:"encryption_available"`
	Webmail              corporateWebmailCheck `json:"webmail"`
	Accounts             map[string]int        `json:"accounts"`
	RecommendedAction    string                `json:"recommended_action"`
}

func getCorporateEmailConfig(dbSuper *sql.DB) CorporateEmailConfig {
	cfg := CorporateEmailConfig{
		Enabled:       false,
		AutoCreate:    true,
		Domain:        "powerfulcontrolsystem.com",
		WebmailURL:    "https://mail.powerfulcontrolsystem.com/mail/",
		ProvisionMode: "manual",
		QuotaMB:       1024,
	}
	if dbSuper == nil {
		return cfg
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailEnabledKey); err == nil && strings.TrimSpace(value) != "" {
		cfg.Enabled = parseConfigBool(value, false)
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailAutoCreateKey); err == nil && strings.TrimSpace(value) != "" {
		cfg.AutoCreate = parseConfigBool(value, true)
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailDomainKey); err == nil && strings.TrimSpace(value) != "" {
		cfg.Domain = normalizeCorporateEmailDomain(value)
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailWebmailURLKey); err == nil && strings.TrimSpace(value) != "" {
		cfg.WebmailURL = strings.TrimSpace(value)
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailProvisionModeKey); err == nil && strings.TrimSpace(value) != "" {
		cfg.ProvisionMode = normalizeCorporateEmailProvisionMode(value)
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailAPIBaseURLKey); err == nil {
		cfg.APIBaseURL = strings.TrimRight(strings.TrimSpace(value), "/")
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailAPIAdminKey); err == nil {
		cfg.APIAdmin = strings.TrimSpace(value)
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailQuotaMBKey); err == nil && strings.TrimSpace(value) != "" {
		if parsed, parseErr := strconv.Atoi(strings.TrimSpace(value)); parseErr == nil && parsed >= 0 {
			cfg.QuotaMB = parsed
		}
	}
	if raw, _, err := dbpkg.GetConfigValue(dbSuper, corporateEmailAPIPasswordKey); err == nil && strings.TrimSpace(raw) != "" {
		cfg.APIPasswordSet = true
	}
	return cfg
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func corporateEmailEnvBool(keys []string, fallback bool) bool {
	value := firstNonEmptyEnv(keys...)
	if value == "" {
		return fallback
	}
	return parseConfigBool(value, fallback)
}

func corporateEmailAPIAdminPassword(dbSuper *sql.DB) (string, error) {
	value, err := getDecryptedConfigValue(dbSuper, corporateEmailAPIPasswordKey)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value), nil
	}
	return firstNonEmptyEnv("IREDADMIN_PASSWORD", "IREDMAIL_ADMIN_PASSWORD", "EMAIL_CORPORATIVO_IREDADMIN_PASSWORD"), nil
}

func corporateEmailInternalURL(rawURL string, envKeys ...string) string {
	rawURL = strings.TrimSpace(rawURL)
	internalURL := firstNonEmptyEnv(envKeys...)
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return rawURL
	}
	host := strings.ToLower(parsed.Hostname())
	if strings.TrimSpace(internalURL) != "" && (host == "iredmail" || host == "mail.powerfulcontrolsystem.com" || strings.HasPrefix(host, "mail.")) {
		internalURL = strings.TrimRight(strings.TrimSpace(internalURL), "/")
		if internalParsed, parseErr := url.Parse(internalURL); parseErr == nil && strings.EqualFold(internalParsed.Hostname(), "iredmail") {
			internalParsed.Scheme = "http"
			return strings.TrimRight(internalParsed.String(), "/")
		}
		return internalURL
	}
	if host == "iredmail" && parsed.Scheme == "https" {
		parsed.Scheme = "http"
		return strings.TrimRight(parsed.String(), "/")
	}
	return rawURL
}

func corporateEmailEffectiveAPIBaseURL(rawURL string) string {
	return corporateEmailInternalURL(rawURL, "EMAIL_CORPORATIVO_INTERNAL_IREDADMIN_API_BASE_URL", "IREDADMIN_INTERNAL_API_BASE_URL")
}

func corporateEmailEffectiveWebmailURL(rawURL string) string {
	internalURL := corporateEmailInternalURL(rawURL, "EMAIL_CORPORATIVO_INTERNAL_WEBMAIL_URL", "IREDMAIL_INTERNAL_WEBMAIL_URL")
	if internalURL == rawURL {
		return rawURL
	}
	if strings.HasSuffix(rawURL, "/") && !strings.HasSuffix(internalURL, "/") {
		return internalURL + "/"
	}
	return internalURL
}

func corporateEmailHTTPClient(timeout time.Duration, jar http.CookieJar, endpoint string) *http.Client {
	client := &http.Client{Timeout: timeout, Jar: jar}
	parsed, err := url.Parse(endpoint)
	if err == nil && strings.EqualFold(parsed.Hostname(), "iredmail") {
		// El certificado interno del contenedor iRedMail es autofirmado/legacy.
		// Esta excepcion solo aplica a la red Docker interna; el acceso publico conserva TLS valido.
		client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	}
	return client
}

// EnsureCorporateEmailConfigFromEnv registra en base la configuracion iRedMail
// definida en variables de entorno de la VPS. No imprime secretos y guarda la
// clave iRedAdmin cifrada si CONFIG_ENC_KEY esta disponible.
func EnsureCorporateEmailConfigFromEnv(dbSuper *sql.DB) error {
	if dbSuper == nil {
		return nil
	}
	envKeys := []string{
		"EMAIL_CORPORATIVO_ENABLED", "IREDMAIL_ENABLED", "EMAIL_CORPORATIVO_AUTO_CREATE", "IREDMAIL_AUTO_CREATE",
		"EMAIL_CORPORATIVO_DOMAIN", "IREDMAIL_DOMAIN", "EMAIL_CORPORATIVO_WEBMAIL_URL", "IREDMAIL_WEBMAIL_URL",
		"EMAIL_CORPORATIVO_PROVISION_MODE", "IREDMAIL_PROVISION_MODE", "IREDADMIN_API_BASE_URL",
		"EMAIL_CORPORATIVO_IREDADMIN_API_BASE_URL", "IREDADMIN_ADMIN", "IREDMAIL_ADMIN_EMAIL",
		"EMAIL_CORPORATIVO_IREDADMIN_ADMIN", "IREDADMIN_PASSWORD", "IREDMAIL_ADMIN_PASSWORD",
		"EMAIL_CORPORATIVO_IREDADMIN_PASSWORD", "IREDMAIL_QUOTA_MB", "EMAIL_CORPORATIVO_QUOTA_MB",
	}
	hasEnv := false
	for _, key := range envKeys {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			hasEnv = true
			break
		}
	}
	if !hasEnv {
		return nil
	}
	cfg := getCorporateEmailConfig(dbSuper)
	cfg.Enabled = corporateEmailEnvBool([]string{"EMAIL_CORPORATIVO_ENABLED", "IREDMAIL_ENABLED"}, cfg.Enabled)
	cfg.AutoCreate = corporateEmailEnvBool([]string{"EMAIL_CORPORATIVO_AUTO_CREATE", "IREDMAIL_AUTO_CREATE"}, cfg.AutoCreate)
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_DOMAIN", "IREDMAIL_DOMAIN"); value != "" {
		cfg.Domain = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_WEBMAIL_URL", "IREDMAIL_WEBMAIL_URL"); value != "" {
		cfg.WebmailURL = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_PROVISION_MODE", "IREDMAIL_PROVISION_MODE"); value != "" {
		cfg.ProvisionMode = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_IREDADMIN_API_BASE_URL", "IREDADMIN_API_BASE_URL"); value != "" {
		cfg.APIBaseURL = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_IREDADMIN_ADMIN", "IREDADMIN_ADMIN", "IREDMAIL_ADMIN_EMAIL"); value != "" {
		cfg.APIAdmin = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_QUOTA_MB", "IREDMAIL_QUOTA_MB"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			cfg.QuotaMB = parsed
		}
	}
	plainPassword := firstNonEmptyEnv("EMAIL_CORPORATIVO_IREDADMIN_PASSWORD", "IREDADMIN_PASSWORD", "IREDMAIL_ADMIN_PASSWORD")
	if cfg.ProvisionMode == "manual" && cfg.APIBaseURL != "" && cfg.APIAdmin != "" && plainPassword != "" {
		cfg.ProvisionMode = "iredadmin_api"
	}
	if plainPassword != "" && !utils.EncryptionAvailable() {
		return fmt.Errorf("CONFIG_ENC_KEY no esta disponible para registrar IREDADMIN_PASSWORD cifrado")
	}
	return saveCorporateEmailConfig(dbSuper, cfg, plainPassword)
}

func parseConfigBool(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "on", "yes", "si", "activo", "activa", "enabled":
		return true
	case "0", "false", "off", "no", "inactivo", "inactiva", "disabled":
		return false
	default:
		return fallback
	}
}

func normalizeCorporateEmailDomain(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.TrimPrefix(value, "http://")
	value = strings.TrimPrefix(value, "https://")
	value = strings.Trim(value, "/ ")
	if value == "" {
		return "powerfulcontrolsystem.com"
	}
	return value
}

func normalizeCorporateEmailProvisionMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "iredadmin_api", "iredmail_api", "api":
		return "iredadmin_api"
	default:
		return "manual"
	}
}

func generateCorporateEmailPassword() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.TrimRight(base64.URLEncoding.EncodeToString(buf), "=") + "Aa1!", nil
}

func saveCorporateEmailConfig(dbSuper *sql.DB, cfg CorporateEmailConfig, plainPassword string) error {
	if dbSuper == nil {
		return fmt.Errorf("base super no disponible")
	}
	cfg.Domain = normalizeCorporateEmailDomain(cfg.Domain)
	cfg.WebmailURL = strings.TrimSpace(cfg.WebmailURL)
	cfg.APIBaseURL = strings.TrimRight(strings.TrimSpace(cfg.APIBaseURL), "/")
	cfg.APIAdmin = strings.TrimSpace(cfg.APIAdmin)
	cfg.ProvisionMode = normalizeCorporateEmailProvisionMode(cfg.ProvisionMode)
	if cfg.QuotaMB < 0 {
		cfg.QuotaMB = 0
	}
	pairs := []struct {
		key   string
		value string
		enc   bool
	}{
		{corporateEmailEnabledKey, strconv.FormatBool(cfg.Enabled), false},
		{corporateEmailAutoCreateKey, strconv.FormatBool(cfg.AutoCreate), false},
		{corporateEmailDomainKey, cfg.Domain, false},
		{corporateEmailWebmailURLKey, cfg.WebmailURL, false},
		{corporateEmailProvisionModeKey, cfg.ProvisionMode, false},
		{corporateEmailAPIBaseURLKey, cfg.APIBaseURL, false},
		{corporateEmailAPIAdminKey, cfg.APIAdmin, false},
		{corporateEmailQuotaMBKey, strconv.Itoa(cfg.QuotaMB), false},
	}
	for _, pair := range pairs {
		if err := dbpkg.SetConfigValue(dbSuper, pair.key, pair.value, pair.enc); err != nil {
			return err
		}
	}
	if strings.TrimSpace(plainPassword) != "" {
		if !utils.EncryptionAvailable() {
			return fmt.Errorf("CONFIG_ENC_KEY no esta disponible para cifrar la clave iRedAdmin")
		}
		enc, err := utils.EncryptString(strings.TrimSpace(plainPassword))
		if err != nil {
			return err
		}
		if err := dbpkg.SetConfigValue(dbSuper, corporateEmailAPIPasswordKey, enc, true); err != nil {
			return err
		}
	}
	return nil
}

func EnsureEmpresaCorporateEmailAfterCreate(dbSuper *sql.DB, empresaID int64, empresaNombre, usuario string) (*dbpkg.EmpresaEmailCorporativo, error) {
	if dbSuper == nil || empresaID <= 0 {
		return nil, nil
	}
	cfg := getCorporateEmailConfig(dbSuper)
	if !cfg.AutoCreate {
		return nil, nil
	}
	if existing, err := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID); err == nil {
		return existing, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	email, local, err := dbpkg.ResolveUniqueCorporateEmail(dbSuper, empresaID, empresaNombre, cfg.Domain)
	if err != nil {
		return nil, err
	}
	initialPassword, err := generateCorporateEmailPassword()
	if err != nil {
		return nil, err
	}
	encryptedPassword := ""
	status := "pendiente_modulo_desactivado"
	if cfg.Enabled {
		status = "pendiente_provision"
	}
	if utils.EncryptionAvailable() {
		encryptedPassword, err = utils.EncryptString(initialPassword)
		if err != nil {
			return nil, err
		}
	} else if cfg.Enabled && cfg.ProvisionMode == "iredadmin_api" {
		status = "pendiente_cifrado"
	}
	item, err := dbpkg.UpsertEmpresaEmailCorporativo(dbSuper, dbpkg.EmpresaEmailCorporativo{
		EmpresaID:         empresaID,
		EmpresaNombre:     empresaNombre,
		Email:             email,
		LocalPart:         local,
		Domain:            cfg.Domain,
		WebmailURL:        cfg.WebmailURL,
		EstadoProvision:   status,
		ProvisionProvider: "iredmail",
		UsuarioCreador:    usuario,
		Observaciones:     "Generado automaticamente al crear la empresa",
	}, encryptedPassword)
	if err != nil {
		return nil, err
	}
	if cfg.Enabled && cfg.ProvisionMode == "iredadmin_api" && encryptedPassword != "" {
		result := provisionEmpresaEmailAccount(dbSuper, cfg, *item, initialPassword)
		if !result.OK {
			log.Printf("email corporativo empresa_id=%d provision warning: %s", empresaID, result.Error)
		}
		item, _ = dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID)
	}
	return item, nil
}

func EnsureCorporateEmailRowsForExistingCompanies(dbSuper, dbEmp *sql.DB, usuario string) (int, error) {
	if dbSuper == nil || dbEmp == nil {
		return 0, nil
	}
	cfg := getCorporateEmailConfig(dbSuper)
	if !cfg.AutoCreate {
		return 0, nil
	}
	return dbpkg.EnsureEmpresaEmailRowsForExistingEmpresas(dbSuper, dbEmp, cfg.Domain, cfg.WebmailURL, usuario)
}

func provisionEmpresaEmailAccount(dbSuper *sql.DB, cfg CorporateEmailConfig, account dbpkg.EmpresaEmailCorporativo, password string) corporateEmailProvisionResult {
	if !cfg.Enabled {
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_modulo_desactivado", "El modulo global esta desactivado", false)
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_modulo_desactivado", Error: "modulo desactivado"}
	}
	if cfg.ProvisionMode != "iredadmin_api" {
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_provision_manual", "Modo manual: crear o validar la cuenta en iRedMail", false)
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_provision_manual", Error: "modo manual"}
	}
	adminPassword, err := corporateEmailAPIAdminPassword(dbSuper)
	if err != nil {
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "error", "No se pudo descifrar la clave iRedAdmin", false)
		return corporateEmailProvisionResult{OK: false, Status: "error", Error: err.Error()}
	}
	if cfg.APIBaseURL == "" || cfg.APIAdmin == "" || strings.TrimSpace(adminPassword) == "" {
		msg := "Faltan URL, usuario o clave de iRedAdmin-Pro"
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_api", msg, false)
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_api", Error: msg}
	}
	if strings.TrimSpace(password) == "" {
		msg := "La clave inicial de la cuenta no esta disponible para crear el buzon"
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_clave", msg, false)
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_clave", Error: msg}
	}
	apiBaseURL := corporateEmailEffectiveAPIBaseURL(cfg.APIBaseURL)
	jar, _ := cookiejar.New(nil)
	client := corporateEmailHTTPClient(20*time.Second, jar, apiBaseURL)
	loginValues := url.Values{}
	loginValues.Set("username", cfg.APIAdmin)
	loginValues.Set("password", adminPassword)
	if err := iredAdminAPIPostForm(client, apiBaseURL+"/api/login", loginValues); err != nil {
		if errors.Is(err, errIredAdminAPINotAvailable) {
			_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_provision_manual", errIredAdminAPINotAvailable.Error(), false)
			return corporateEmailProvisionResult{OK: false, Status: "pendiente_provision_manual", Error: errIredAdminAPINotAvailable.Error()}
		}
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "error_login", err.Error(), false)
		return corporateEmailProvisionResult{OK: false, Status: "error_login", Error: err.Error()}
	}
	userValues := url.Values{}
	userValues.Set("name", account.EmpresaNombre)
	userValues.Set("password", password)
	userValues.Set("language", "es_ES")
	userValues.Set("accountStatus", "active")
	userValues.Set("quota", strconv.Itoa(cfg.QuotaMB))
	if err := iredAdminAPIPostForm(client, apiBaseURL+"/api/user/"+url.PathEscape(account.Email), userValues); err != nil {
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "error_provision", err.Error(), false)
		return corporateEmailProvisionResult{OK: false, Status: "error_provision", Error: err.Error()}
	}
	_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "provisionado", "", true)
	return corporateEmailProvisionResult{OK: true, Status: "provisionado"}
}

func iredAdminAPIPostForm(client *http.Client, endpoint string, values url.Values) error {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	var payload struct {
		Success bool   `json:"_success"`
		Msg     string `json:"_msg"`
		Error   string `json:"error"`
	}
	_ = json.Unmarshal(body, &payload)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		msg := strings.TrimSpace(payload.Msg)
		if msg == "" {
			msg = strings.TrimSpace(payload.Error)
		}
		if res.StatusCode == http.StatusUnauthorized && strings.EqualFold(msg, "unauthorized") && (res.Header.Get("X-Request-Id") != "" || strings.Contains(string(body), `"request_id"`)) {
			return fmt.Errorf("iRedAdmin no esta publicado: el subdominio mail esta respondiendo el backend POS")
		}
		if res.StatusCode == http.StatusNotFound && strings.Contains(endpoint, "/api/login") {
			return errIredAdminAPINotAvailable
		}
		if msg == "" {
			msg = strings.TrimSpace(string(body))
		}
		return fmt.Errorf("iRedAdmin HTTP %d: %s", res.StatusCode, msg)
	}
	if strings.Contains(strings.TrimSpace(string(body)), "_success") && !payload.Success {
		msg := strings.TrimSpace(payload.Msg)
		if msg == "" {
			msg = "iRedAdmin rechazo la operacion"
		}
		return errors.New(msg)
	}
	return nil
}

func checkCorporateWebmail(rawURL string) corporateWebmailCheck {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" || rawURL == "#" {
		return corporateWebmailCheck{Checked: false, OK: false, Message: "Webmail sin URL configurada"}
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return corporateWebmailCheck{Checked: true, OK: false, Message: "URL de webmail invalida"}
	}
	requestURL := corporateEmailEffectiveWebmailURL(rawURL)
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodHead, requestURL, nil)
	if err != nil {
		return corporateWebmailCheck{Checked: true, OK: false, Message: "No se pudo preparar la verificacion del webmail"}
	}
	req.Header.Set("User-Agent", "PowerfulControlSystem-WebmailCheck/1.0")
	res, err := client.Do(req)
	if err != nil {
		return corporateWebmailCheck{Checked: true, OK: false, Message: "No se pudo conectar con el webmail"}
	}
	defer res.Body.Close()
	status := res.StatusCode
	if status == http.StatusMethodNotAllowed {
		reqGet, reqErr := http.NewRequest(http.MethodGet, requestURL, nil)
		if reqErr == nil {
			reqGet.Header.Set("User-Agent", "PowerfulControlSystem-WebmailCheck/1.0")
			if resGet, getErr := client.Do(reqGet); getErr == nil {
				defer resGet.Body.Close()
				status = resGet.StatusCode
			}
		}
	}
	switch {
	case status >= 200 && status < 400:
		return corporateWebmailCheck{Checked: true, OK: true, Status: status, Message: "Webmail disponible"}
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return corporateWebmailCheck{Checked: true, OK: false, Status: status, Message: "El webmail responde autenticacion requerida"}
	default:
		return corporateWebmailCheck{Checked: true, OK: false, Status: status, Message: fmt.Sprintf("El webmail respondio HTTP %d", status)}
	}
}

func corporateEmailResponse(cfg CorporateEmailConfig, account *dbpkg.EmpresaEmailCorporativo, message string, checkWebmail bool) map[string]interface{} {
	webmailURL := cfg.WebmailURL
	if account != nil && strings.TrimSpace(account.WebmailURL) != "" {
		webmailURL = account.WebmailURL
	}
	resp := map[string]interface{}{
		"ok":          true,
		"enabled":     cfg.Enabled,
		"auto_create": cfg.AutoCreate,
		"account":     account,
		"webmail":     webmailURL,
		"domain":      cfg.Domain,
	}
	if strings.TrimSpace(message) != "" {
		resp["message"] = strings.TrimSpace(message)
	}
	if checkWebmail {
		resp["webmail_check"] = checkCorporateWebmail(webmailURL)
	}
	return resp
}

func corporateEmailAccountsSummary(accounts []dbpkg.EmpresaEmailCorporativo) map[string]int {
	out := map[string]int{
		"total":        len(accounts),
		"provisionado": 0,
		"pendiente":    0,
		"error":        0,
		"sin_clave":    0,
	}
	for _, item := range accounts {
		status := strings.ToLower(strings.TrimSpace(item.EstadoProvision))
		switch {
		case status == "provisionado":
			out["provisionado"]++
		case strings.HasPrefix(status, "error"):
			out["error"]++
		default:
			out["pendiente"]++
		}
		if !item.InitialPasswordSet {
			out["sin_clave"]++
		}
	}
	return out
}

func corporateEmailDiagnosticsFor(cfg CorporateEmailConfig, accounts []dbpkg.EmpresaEmailCorporativo) corporateEmailDiagnostics {
	summary := corporateEmailAccountsSummary(accounts)
	recommended := "Configuracion lista para asignar correos por empresa."
	if !cfg.Enabled {
		recommended = "Activa el modulo para que las empresas puedan abrir su buzon corporativo."
	} else if cfg.ProvisionMode != "iredadmin_api" {
		recommended = "El modo manual asigna correos, pero el buzon real se debe crear en iRedMail fuera del sistema."
	} else if cfg.APIBaseURL == "" || cfg.APIAdmin == "" || !cfg.APIPasswordSet {
		recommended = "Completa URL API, administrador y clave de iRedAdmin-Pro."
	} else if summary["error"] > 0 {
		recommended = "Prueba iRedAdmin y luego reintenta provisionar las cuentas con error."
	} else if summary["pendiente"] > 0 {
		recommended = "Provisiona las cuentas pendientes para crear los buzones reales."
	}
	return corporateEmailDiagnostics{
		Enabled:              cfg.Enabled,
		AutoCreate:           cfg.AutoCreate,
		ProvisionMode:        cfg.ProvisionMode,
		IredAdminAPIEnabled:  cfg.Enabled && cfg.ProvisionMode == "iredadmin_api",
		IredAdminAPIURLSet:   strings.TrimSpace(cfg.APIBaseURL) != "",
		IredAdminAdminSet:    strings.TrimSpace(cfg.APIAdmin) != "",
		IredAdminPasswordSet: cfg.APIPasswordSet,
		EncryptionAvailable:  utils.EncryptionAvailable(),
		Webmail:              checkCorporateWebmail(cfg.WebmailURL),
		Accounts:             summary,
		RecommendedAction:    recommended,
	}
}

func testCorporateEmailIredAdminLogin(dbSuper *sql.DB, cfg CorporateEmailConfig) corporateEmailProvisionResult {
	if !cfg.Enabled {
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_modulo_desactivado", Error: "modulo desactivado"}
	}
	if cfg.ProvisionMode != "iredadmin_api" {
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_provision_manual", Error: "modo manual"}
	}
	adminPassword, err := corporateEmailAPIAdminPassword(dbSuper)
	if err != nil {
		return corporateEmailProvisionResult{OK: false, Status: "error", Error: "No se pudo descifrar la clave iRedAdmin"}
	}
	if cfg.APIBaseURL == "" || cfg.APIAdmin == "" || strings.TrimSpace(adminPassword) == "" {
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_api", Error: "Faltan URL, usuario o clave de iRedAdmin-Pro"}
	}
	apiBaseURL := corporateEmailEffectiveAPIBaseURL(cfg.APIBaseURL)
	jar, _ := cookiejar.New(nil)
	client := corporateEmailHTTPClient(20*time.Second, jar, apiBaseURL)
	loginValues := url.Values{}
	loginValues.Set("username", cfg.APIAdmin)
	loginValues.Set("password", adminPassword)
	if err := iredAdminAPIPostForm(client, apiBaseURL+"/api/login", loginValues); err != nil {
		if errors.Is(err, errIredAdminAPINotAvailable) {
			return corporateEmailProvisionResult{OK: false, Status: "pendiente_provision_manual", Error: errIredAdminAPINotAvailable.Error()}
		}
		return corporateEmailProvisionResult{OK: false, Status: "error_login", Error: err.Error()}
	}
	return corporateEmailProvisionResult{OK: true, Status: "login_ok"}
}

func corporateEmailInitialPasswordForProvision(dbSuper *sql.DB, account dbpkg.EmpresaEmailCorporativo) (string, error) {
	encryptedPassword, err := dbpkg.GetEmpresaEmailCorporativoInitialPasswordEncrypted(dbSuper, account.EmpresaID)
	if err == nil && strings.TrimSpace(encryptedPassword) != "" {
		return utils.DecryptString(encryptedPassword)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if !utils.EncryptionAvailable() {
		return "", fmt.Errorf("CONFIG_ENC_KEY no esta disponible para cifrar la clave inicial")
	}
	initialPassword, err := generateCorporateEmailPassword()
	if err != nil {
		return "", err
	}
	encryptedPassword, err = utils.EncryptString(initialPassword)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(account.EstadoProvision) == "" {
		account.EstadoProvision = "pendiente_provision"
	}
	if strings.TrimSpace(account.ProvisionProvider) == "" {
		account.ProvisionProvider = "iredmail"
	}
	if strings.TrimSpace(account.Observaciones) == "" {
		account.Observaciones = "Clave inicial generada al reintentar provision"
	}
	if _, err := dbpkg.UpsertEmpresaEmailCorporativo(dbSuper, account, encryptedPassword); err != nil {
		return "", err
	}
	return initialPassword, nil
}

func SuperEmailCorporativoHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			cfg := getCorporateEmailConfig(dbSuper)
			accounts, err := dbpkg.ListEmpresaEmailCorporativo(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo listar emails corporativos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                    true,
				"config":                cfg,
				"accounts":              accounts,
				"diagnostics":           corporateEmailDiagnosticsFor(cfg, accounts),
				"encryption_available":  utils.EncryptionAvailable(),
				"iredadmin_api_enabled": cfg.Enabled && cfg.ProvisionMode == "iredadmin_api",
			})
			return
		case http.MethodPost, http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "sync" {
				cfg := getCorporateEmailConfig(dbSuper)
				count, err := dbpkg.EnsureEmpresaEmailRowsForExistingEmpresas(dbSuper, dbEmp, cfg.Domain, cfg.WebmailURL, adminEmailFromRequest(r))
				if err != nil {
					http.Error(w, "No se pudo sincronizar empresas: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "created": count})
				return
			}
			if action == "test_iredadmin" {
				result := testCorporateEmailIredAdminLogin(dbSuper, getCorporateEmailConfig(dbSuper))
				writeJSON(w, http.StatusOK, result)
				return
			}
			if action == "provision" {
				empresaID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("empresa_id")), 10, 64)
				if empresaID <= 0 {
					http.Error(w, "empresa_id requerido", http.StatusBadRequest)
					return
				}
				account, err := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID)
				if err != nil {
					http.Error(w, "No se encontro email corporativo: "+err.Error(), http.StatusNotFound)
					return
				}
				initialPassword, passErr := corporateEmailInitialPasswordForProvision(dbSuper, *account)
				if passErr != nil {
					_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, empresaID, "pendiente_clave", "No se pudo recuperar la clave inicial cifrada", false)
					http.Error(w, "No se pudo recuperar la clave inicial cifrada: "+passErr.Error(), http.StatusBadRequest)
					return
				}
				result := provisionEmpresaEmailAccount(dbSuper, getCorporateEmailConfig(dbSuper), *account, initialPassword)
				writeJSON(w, http.StatusOK, result)
				return
			}
			var payload struct {
				Enabled       *bool  `json:"enabled"`
				AutoCreate    *bool  `json:"auto_create"`
				Domain        string `json:"domain"`
				WebmailURL    string `json:"webmail_url"`
				ProvisionMode string `json:"provision_mode"`
				APIBaseURL    string `json:"iredadmin_api_base_url"`
				APIAdmin      string `json:"iredadmin_admin"`
				APIPassword   string `json:"iredadmin_password"`
				QuotaMB       int    `json:"quota_mb"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}
			cfg := getCorporateEmailConfig(dbSuper)
			if payload.Enabled != nil {
				cfg.Enabled = *payload.Enabled
			}
			if payload.AutoCreate != nil {
				cfg.AutoCreate = *payload.AutoCreate
			}
			if strings.TrimSpace(payload.Domain) != "" {
				cfg.Domain = payload.Domain
			}
			if strings.TrimSpace(payload.WebmailURL) != "" {
				cfg.WebmailURL = payload.WebmailURL
			}
			if strings.TrimSpace(payload.ProvisionMode) != "" {
				cfg.ProvisionMode = payload.ProvisionMode
			}
			cfg.APIBaseURL = payload.APIBaseURL
			cfg.APIAdmin = payload.APIAdmin
			if payload.QuotaMB >= 0 {
				cfg.QuotaMB = payload.QuotaMB
			}
			if err := saveCorporateEmailConfig(dbSuper, cfg, payload.APIPassword); err != nil {
				http.Error(w, "No se pudo guardar configuracion: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": getCorporateEmailConfig(dbSuper)})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func EmpresaEmailCorporativoHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("empresa_id")), 10, 64)
		if empresaID <= 0 {
			http.Error(w, "empresa_id requerido", http.StatusBadRequest)
			return
		}
		cfg := getCorporateEmailConfig(dbSuper)
		checkWebmail := parseConfigBool(r.URL.Query().Get("check_webmail"), false)
		account, err := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if cfg.AutoCreate && dbEmp != nil {
					if empresa, empresaErr := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID); empresaErr == nil && empresa != nil {
						if created, createErr := EnsureEmpresaCorporateEmailAfterCreate(dbSuper, empresa.EmpresaID, empresa.Nombre, adminEmailFromRequest(r)); createErr == nil {
							account = created
						} else {
							writeJSON(w, http.StatusOK, corporateEmailResponse(cfg, nil, "No se pudo generar el email corporativo", checkWebmail))
							return
						}
					}
				}
				if account != nil {
					if account.WebmailURL == "" {
						account.WebmailURL = cfg.WebmailURL
					}
					writeJSON(w, http.StatusOK, corporateEmailResponse(cfg, account, "Email corporativo generado", checkWebmail))
					return
				}
				writeJSON(w, http.StatusOK, corporateEmailResponse(cfg, nil, "Sin email corporativo generado", checkWebmail))
				return
			}
			http.Error(w, "No se pudo consultar email corporativo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if account.WebmailURL == "" {
			account.WebmailURL = cfg.WebmailURL
		}
		writeJSON(w, http.StatusOK, corporateEmailResponse(cfg, account, "", checkWebmail))
	}
}
