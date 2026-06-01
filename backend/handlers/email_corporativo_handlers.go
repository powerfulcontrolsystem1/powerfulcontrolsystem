package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
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
	corporateEmailAPIBaseURLKey    = "email_corporativo.mailu_api_base_url"
	corporateEmailAPIAdminKey      = "email_corporativo.mailu_admin"
	corporateEmailAPIPasswordKey   = "email_corporativo.mailu_api_token"
	corporateEmailQuotaMBKey       = "email_corporativo.quota_mb"
	corporateEmailDirectCommandKey = "email_corporativo.direct_provision_command"
	corporateEmailAutologinKey     = "email_corporativo.autologin_secret"
	corporateEmailEmpresaPrefsKey  = "email_corporativo_config"
	corporateEmailMaxAccountsKey   = "email_corporativo.max_accounts_per_empresa"
	corporateEmailDefaultMax       = 5
)

type CorporateEmailConfig struct {
	Enabled        bool   `json:"enabled"`
	AutoCreate     bool   `json:"auto_create"`
	Domain         string `json:"domain"`
	WebmailURL     string `json:"webmail_url"`
	ProvisionMode  string `json:"provision_mode"`
	APIBaseURL     string `json:"mailu_api_base_url"`
	APIAdmin       string `json:"mailu_admin"`
	APIPasswordSet bool   `json:"mailu_api_token_set"`
	QuotaMB        int    `json:"quota_mb"`
	MaxAccounts    int    `json:"max_accounts_per_empresa"`
	DirectCommand  string `json:"direct_provision_command,omitempty"`
}

type corporateEmailProvisionResult struct {
	OK     bool   `json:"ok"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type corporateEmailEmpresaPrefs struct {
	AutoOpen bool `json:"auto_open"`
}

var errCorporateEmailAutologinRejected = errors.New("credenciales del buzon no aceptadas por Mailu")

type corporateWebmailCheck struct {
	Checked bool   `json:"checked"`
	OK      bool   `json:"ok"`
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type corporateEmailDiagnostics struct {
	Enabled             bool                  `json:"enabled"`
	AutoCreate          bool                  `json:"auto_create"`
	ProvisionMode       string                `json:"provision_mode"`
	MailuDirectEnabled  bool                  `json:"mailu_direct_enabled"`
	MailuAPIURLSet      bool                  `json:"mailu_api_url_set"`
	MailuAdminSet       bool                  `json:"mailu_admin_set"`
	MailuAPITokenSet    bool                  `json:"mailu_api_token_set"`
	EncryptionAvailable bool                  `json:"encryption_available"`
	Webmail             corporateWebmailCheck `json:"webmail"`
	Accounts            map[string]int        `json:"accounts"`
	RecommendedAction   string                `json:"recommended_action"`
}

func getCorporateEmailConfig(dbSuper *sql.DB) CorporateEmailConfig {
	cfg := CorporateEmailConfig{
		Enabled:       false,
		AutoCreate:    true,
		Domain:        "powerfulcontrolsystem.com",
		WebmailURL:    "https://mail.powerfulcontrolsystem.com/webmail/",
		ProvisionMode: "manual",
		QuotaMB:       1024,
		MaxAccounts:   corporateEmailDefaultMax,
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
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailMaxAccountsKey); err == nil && strings.TrimSpace(value) != "" {
		if parsed, parseErr := strconv.Atoi(strings.TrimSpace(value)); parseErr == nil {
			cfg.MaxAccounts = normalizeCorporateEmailMaxAccounts(parsed)
		}
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailDirectCommandKey); err == nil {
		cfg.DirectCommand = strings.TrimSpace(value)
	}
	if raw, _, err := dbpkg.GetConfigValue(dbSuper, corporateEmailAPIPasswordKey); err == nil && strings.TrimSpace(raw) != "" {
		cfg.APIPasswordSet = true
	}
	return cfg
}

func normalizeCorporateEmailMaxAccounts(value int) int {
	if value <= 0 {
		return corporateEmailDefaultMax
	}
	if value > 500 {
		return 500
	}
	return value
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
	return firstNonEmptyEnv("MAILU_API_TOKEN", "EMAIL_CORPORATIVO_MAILU_API_TOKEN"), nil
}

func corporateEmailInternalURL(rawURL string, envKeys ...string) string {
	rawURL = strings.TrimSpace(rawURL)
	internalURL := firstNonEmptyEnv(envKeys...)
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return rawURL
	}
	host := strings.ToLower(parsed.Hostname())
	if strings.TrimSpace(internalURL) != "" && (host == "mailu-front" || host == "mail.powerfulcontrolsystem.com" || strings.HasPrefix(host, "mail.")) {
		internalURL = strings.TrimRight(strings.TrimSpace(internalURL), "/")
		if internalParsed, parseErr := url.Parse(internalURL); parseErr == nil && strings.EqualFold(internalParsed.Hostname(), "mailu-front") {
			internalParsed.Scheme = "http"
			return strings.TrimRight(internalParsed.String(), "/")
		}
		return internalURL
	}
	if host == "mailu-front" && parsed.Scheme == "https" {
		parsed.Scheme = "http"
		return strings.TrimRight(parsed.String(), "/")
	}
	return rawURL
}

func corporateEmailEffectiveAPIBaseURL(rawURL string) string {
	return corporateEmailInternalURL(rawURL, "EMAIL_CORPORATIVO_INTERNAL_MAILU_API_BASE_URL", "MAILU_INTERNAL_API_BASE_URL")
}

func corporateEmailEffectiveWebmailURL(rawURL string) string {
	internalURL := corporateEmailInternalURL(rawURL, "EMAIL_CORPORATIVO_INTERNAL_WEBMAIL_URL", "MAILU_INTERNAL_WEBMAIL_URL")
	if internalURL == rawURL {
		return rawURL
	}
	if strings.HasSuffix(rawURL, "/") && !strings.HasSuffix(internalURL, "/") {
		return internalURL + "/"
	}
	return internalURL
}

func corporateEmailIsLegacyWebmailURL(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}
	path := strings.ToLower(strings.Trim(parsed.Path, "/"))
	return path == "mail"
}

func corporateEmailAccountWebmailURL(accountURL, configURL string) string {
	accountURL = strings.TrimSpace(accountURL)
	if accountURL != "" && !corporateEmailIsLegacyWebmailURL(accountURL) {
		return accountURL
	}
	return strings.TrimSpace(configURL)
}

// EnsureCorporateEmailConfigFromEnv registra en base la configuracion Mailu
// definida en variables de entorno de la VPS. No imprime secretos y guarda el
// token API cifrado si CONFIG_ENC_KEY esta disponible.
func EnsureCorporateEmailConfigFromEnv(dbSuper *sql.DB) error {
	if dbSuper == nil {
		return nil
	}
	envKeys := []string{
		"EMAIL_CORPORATIVO_ENABLED", "MAILU_ENABLED", "EMAIL_CORPORATIVO_AUTO_CREATE", "MAILU_AUTO_CREATE",
		"EMAIL_CORPORATIVO_DOMAIN", "MAILU_DOMAIN", "EMAIL_CORPORATIVO_WEBMAIL_URL", "MAILU_WEBMAIL_URL",
		"EMAIL_CORPORATIVO_PROVISION_MODE", "MAILU_PROVISION_MODE", "MAILU_API_BASE_URL",
		"EMAIL_CORPORATIVO_MAILU_API_BASE_URL", "MAILU_ADMIN", "MAILU_POSTMASTER",
		"EMAIL_CORPORATIVO_MAILU_ADMIN", "MAILU_API_TOKEN", "EMAIL_CORPORATIVO_MAILU_API_TOKEN",
		"MAILU_QUOTA_MB", "EMAIL_CORPORATIVO_QUOTA_MB",
		"EMAIL_CORPORATIVO_MAX_ACCOUNTS_PER_EMPRESA", "MAILU_MAX_ACCOUNTS_PER_EMPRESA",
		"EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND", "MAILU_DIRECT_PROVISION_COMMAND",
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
	cfg.Enabled = corporateEmailEnvBool([]string{"EMAIL_CORPORATIVO_ENABLED", "MAILU_ENABLED"}, cfg.Enabled)
	cfg.AutoCreate = corporateEmailEnvBool([]string{"EMAIL_CORPORATIVO_AUTO_CREATE", "MAILU_AUTO_CREATE"}, cfg.AutoCreate)
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_DOMAIN", "MAILU_DOMAIN"); value != "" {
		cfg.Domain = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_WEBMAIL_URL", "MAILU_WEBMAIL_URL"); value != "" {
		cfg.WebmailURL = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_PROVISION_MODE", "MAILU_PROVISION_MODE"); value != "" {
		cfg.ProvisionMode = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_MAILU_API_BASE_URL", "MAILU_API_BASE_URL"); value != "" {
		cfg.APIBaseURL = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_MAILU_ADMIN", "MAILU_ADMIN", "MAILU_POSTMASTER"); value != "" {
		cfg.APIAdmin = value
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_QUOTA_MB", "MAILU_QUOTA_MB"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			cfg.QuotaMB = parsed
		}
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_MAX_ACCOUNTS_PER_EMPRESA", "MAILU_MAX_ACCOUNTS_PER_EMPRESA"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			cfg.MaxAccounts = normalizeCorporateEmailMaxAccounts(parsed)
		}
	}
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND", "MAILU_DIRECT_PROVISION_COMMAND"); value != "" {
		cfg.DirectCommand = value
	}
	plainPassword := firstNonEmptyEnv("EMAIL_CORPORATIVO_MAILU_API_TOKEN", "MAILU_API_TOKEN")
	if cfg.ProvisionMode == "manual" && strings.TrimSpace(cfg.DirectCommand) != "" {
		cfg.ProvisionMode = "mailu_direct"
	}
	if plainPassword != "" && !utils.EncryptionAvailable() {
		return fmt.Errorf("CONFIG_ENC_KEY no esta disponible para registrar MAILU_API_TOKEN cifrado")
	}
	if err := saveCorporateEmailConfig(dbSuper, cfg, plainPassword); err != nil {
		return err
	}
	return nil
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

func normalizeCorporateEmailTheme(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "dark", "oscuro", "nocturno", "negro", "dark-corporate", "dark-neon", "dark-absolute", "negro-absoluto", "dark-obsidian", "super-oscuro":
		return "dark"
	default:
		return "light"
	}
}

func corporateEmailThemeName(value string) string {
	if normalizeCorporateEmailTheme(value) == "dark" {
		return "PCSDark"
	}
	return "PCSLight"
}

func corporateEmailSnappyMailTheme(value string) string {
	return corporateEmailThemeName(value) + "@custom"
}

func getCorporateEmailEmpresaPrefs(dbEmp *sql.DB, empresaID int64) corporateEmailEmpresaPrefs {
	prefs := corporateEmailEmpresaPrefs{AutoOpen: true}
	if dbEmp == nil || empresaID <= 0 {
		return prefs
	}
	item, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, corporateEmailEmpresaPrefsKey)
	if err != nil || item == nil || strings.TrimSpace(item.Valor) == "" {
		return prefs
	}
	var stored corporateEmailEmpresaPrefs
	if err := json.Unmarshal([]byte(item.Valor), &stored); err != nil {
		return prefs
	}
	prefs.AutoOpen = stored.AutoOpen
	return prefs
}

func saveCorporateEmailEmpresaPrefs(dbEmp *sql.DB, empresaID int64, prefs corporateEmailEmpresaPrefs, usuario string) error {
	if dbEmp == nil {
		return fmt.Errorf("base empresarial no disponible")
	}
	if empresaID <= 0 {
		return fmt.Errorf("empresa_id invalido")
	}
	payload, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	_, err = dbpkg.UpsertEmpresaEstacionPref(dbEmp, dbpkg.EmpresaEstacionPref{
		EmpresaID:      empresaID,
		EstacionID:     0,
		Clave:          corporateEmailEmpresaPrefsKey,
		Valor:          string(payload),
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
		Observaciones:  "Preferencias de bandeja de email corporativo por empresa",
	})
	return err
}

func validateCorporateEmailNewPassword(password, confirm string) (string, error) {
	password = strings.TrimSpace(password)
	confirm = strings.TrimSpace(confirm)
	if password == "" {
		return "", nil
	}
	if password != confirm {
		return "", fmt.Errorf("la confirmacion de la contrasena no coincide")
	}
	if len(password) < 10 || len(password) > 128 {
		return "", fmt.Errorf("la contrasena debe tener entre 10 y 128 caracteres")
	}
	for _, r := range password {
		if r < 32 || r == 127 {
			return "", fmt.Errorf("la contrasena contiene caracteres no permitidos")
		}
	}
	return password, nil
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
	case "mailu_direct", "direct", "docker_direct", "cli", "direct_sql":
		return "mailu_direct"
	case "mailu_api", "api":
		return "manual"
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

type corporateEmailAutologinToken struct {
	EmpresaID int64  `json:"empresa_id"`
	Email     string `json:"email"`
	Exp       int64  `json:"exp"`
	Nonce     string `json:"nonce"`
}

func corporateEmailAutologinSecret(dbSuper *sql.DB) (string, error) {
	if value := firstNonEmptyEnv("EMAIL_CORPORATIVO_AUTOLOGIN_SECRET", "MAILU_AUTOLOGIN_SECRET"); strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value), nil
	}
	if value, err := getDecryptedConfigValue(dbSuper, corporateEmailAutologinKey); err == nil && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value), nil
	} else if err != nil {
		return "", err
	}
	if !utils.EncryptionAvailable() {
		return "", fmt.Errorf("CONFIG_ENC_KEY no esta disponible para proteger autologin")
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(buf)
	encrypted, err := utils.EncryptString(secret)
	if err != nil {
		return "", err
	}
	if err := dbpkg.SetConfigValue(dbSuper, corporateEmailAutologinKey, encrypted, true); err != nil {
		return "", err
	}
	return secret, nil
}

func signCorporateEmailAutologinPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func createCorporateEmailAutologinToken(dbSuper *sql.DB, account dbpkg.EmpresaEmailCorporativo) (string, error) {
	if account.EmpresaID <= 0 || strings.TrimSpace(account.Email) == "" {
		return "", fmt.Errorf("cuenta corporativa invalida")
	}
	secret, err := corporateEmailAutologinSecret(dbSuper)
	if err != nil {
		return "", err
	}
	nonceBytes := make([]byte, 12)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	payload := corporateEmailAutologinToken{
		EmpresaID: account.EmpresaID,
		Email:     strings.ToLower(strings.TrimSpace(account.Email)),
		Exp:       time.Now().Add(2 * time.Minute).Unix(),
		Nonce:     base64.RawURLEncoding.EncodeToString(nonceBytes),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signature := signCorporateEmailAutologinPayload([]byte(payloadPart), secret)
	return payloadPart + "." + signature, nil
}

func validateCorporateEmailAutologinToken(dbSuper *sql.DB, token string) (corporateEmailAutologinToken, error) {
	var payload corporateEmailAutologinToken
	token = strings.TrimSpace(token)
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return payload, fmt.Errorf("token invalido")
	}
	secret, err := corporateEmailAutologinSecret(dbSuper)
	if err != nil {
		return payload, err
	}
	expected := signCorporateEmailAutologinPayload([]byte(parts[0]), secret)
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return payload, fmt.Errorf("firma invalida")
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return payload, fmt.Errorf("payload invalido")
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return payload, fmt.Errorf("payload invalido")
	}
	if payload.EmpresaID <= 0 || strings.TrimSpace(payload.Email) == "" || time.Now().Unix() > payload.Exp {
		return payload, fmt.Errorf("token vencido o incompleto")
	}
	return payload, nil
}

func corporateEmailAutologinPublicURL(webmailURL, token, theme string) string {
	parsed, err := url.Parse(strings.TrimSpace(webmailURL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || strings.TrimSpace(token) == "" {
		return ""
	}
	parsed.Path = "/pcs-mail-autologin"
	query := parsed.Query()
	query.Set("token", token)
	query.Set("theme", normalizeCorporateEmailTheme(theme))
	query.Set("mail_theme", corporateEmailSnappyMailTheme(theme))
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String()
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
	cfg.DirectCommand = strings.TrimSpace(cfg.DirectCommand)
	if cfg.QuotaMB < 0 {
		cfg.QuotaMB = 0
	}
	cfg.MaxAccounts = normalizeCorporateEmailMaxAccounts(cfg.MaxAccounts)
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
		{corporateEmailMaxAccountsKey, strconv.Itoa(cfg.MaxAccounts), false},
		{corporateEmailDirectCommandKey, cfg.DirectCommand, false},
	}
	for _, pair := range pairs {
		if err := dbpkg.SetConfigValue(dbSuper, pair.key, pair.value, pair.enc); err != nil {
			return err
		}
	}
	if strings.TrimSpace(plainPassword) != "" {
		if !utils.EncryptionAvailable() {
			return fmt.Errorf("CONFIG_ENC_KEY no esta disponible para cifrar el token Mailu")
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
	if count, err := dbpkg.CountEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID); err != nil {
		return nil, err
	} else if count >= normalizeCorporateEmailMaxAccounts(cfg.MaxAccounts) {
		return nil, fmt.Errorf("la empresa ya alcanzo el limite de %d cuentas de correo corporativo", normalizeCorporateEmailMaxAccounts(cfg.MaxAccounts))
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
	} else if cfg.Enabled && cfg.ProvisionMode == "mailu_direct" {
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
		ProvisionProvider: "mailu",
		UsuarioCreador:    usuario,
		Observaciones:     "Generado automaticamente al crear la empresa",
	}, encryptedPassword)
	if err != nil {
		return nil, err
	}
	if cfg.Enabled && cfg.ProvisionMode == "mailu_direct" && encryptedPassword != "" {
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
	return dbpkg.EnsureEmpresaEmailRowsForExistingEmpresas(dbSuper, dbEmp, cfg.Domain, cfg.WebmailURL, usuario, normalizeCorporateEmailMaxAccounts(cfg.MaxAccounts))
}

func provisionEmpresaEmailAccount(dbSuper *sql.DB, cfg CorporateEmailConfig, account dbpkg.EmpresaEmailCorporativo, password string) corporateEmailProvisionResult {
	return provisionEmpresaEmailAccountWithTheme(dbSuper, cfg, account, password, "")
}

func provisionEmpresaEmailAccountWithTheme(dbSuper *sql.DB, cfg CorporateEmailConfig, account dbpkg.EmpresaEmailCorporativo, password, theme string) corporateEmailProvisionResult {
	if !cfg.Enabled {
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_modulo_desactivado", "El modulo global esta desactivado", false)
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_modulo_desactivado", Error: "modulo desactivado"}
	}
	if cfg.ProvisionMode != "mailu_direct" {
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_provision_manual", "Modo manual: crear o validar la cuenta en Mailu", false)
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_provision_manual", Error: "modo manual"}
	}
	return provisionEmpresaEmailAccountDirect(dbSuper, cfg, account, password, theme)
}

func provisionEmpresaEmailAccountDirect(dbSuper *sql.DB, cfg CorporateEmailConfig, account dbpkg.EmpresaEmailCorporativo, password, theme string) corporateEmailProvisionResult {
	commandPath := strings.TrimSpace(cfg.DirectCommand)
	if commandPath == "" {
		commandPath = strings.TrimSpace(firstNonEmptyEnv("EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND", "MAILU_DIRECT_PROVISION_COMMAND"))
	}
	if commandPath == "" {
		msg := "Falta EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND para crear el buzon directo en Mailu"
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_comando", msg, false)
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_comando", Error: msg}
	}
	if strings.TrimSpace(password) == "" {
		msg := "La clave inicial de la cuenta no esta disponible para crear el buzon"
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "pendiente_clave", msg, false)
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_clave", Error: msg}
	}
	if err := validateCorporateEmailAccountForProvision(account); err != nil {
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "error_validacion", err.Error(), false)
		return corporateEmailProvisionResult{OK: false, Status: "error_validacion", Error: err.Error()}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, commandPath)
	cmd.Env = append(os.Environ(),
		"PCS_MAILU_EMAIL="+strings.ToLower(strings.TrimSpace(account.Email)),
		"PCS_MAILU_PASSWORD="+password,
		"PCS_MAILU_NAME="+strings.TrimSpace(account.EmpresaNombre),
		"PCS_MAILU_DOMAIN="+normalizeCorporateEmailDomain(account.Domain),
		"PCS_MAILU_QUOTA_MB="+strconv.Itoa(cfg.QuotaMB),
		"PCS_MAILU_THEME_MODE="+normalizeCorporateEmailTheme(theme),
		"PCS_MAILU_THEME="+corporateEmailThemeName(theme),
	)
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		msg := "Tiempo agotado creando el buzon en Mailu"
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "error_timeout", msg, false)
		return corporateEmailProvisionResult{OK: false, Status: "error_timeout", Error: msg}
	}
	if err != nil {
		msg := sanitizeProvisionCommandOutput(string(output))
		if msg == "" {
			msg = "El comando directo de Mailu fallo"
		}
		_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "error_provision", msg, false)
		return corporateEmailProvisionResult{OK: false, Status: "error_provision", Error: msg}
	}
	_ = dbpkg.MarkEmpresaEmailProvisionResult(dbSuper, account.EmpresaID, "provisionado", "", true)
	return corporateEmailProvisionResult{OK: true, Status: "provisionado"}
}

func validateCorporateEmailAccountForProvision(account dbpkg.EmpresaEmailCorporativo) error {
	email := strings.ToLower(strings.TrimSpace(account.Email))
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9._%+-]{0,126}@[a-z0-9.-]+\.[a-z]{2,}$`).MatchString(email) {
		return fmt.Errorf("email corporativo invalido")
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("email corporativo invalido")
	}
	domain := normalizeCorporateEmailDomain(account.Domain)
	if domain != "" && domain != parts[1] {
		return fmt.Errorf("el dominio del email no coincide con la configuracion de la empresa")
	}
	return nil
}

func sanitizeProvisionCommandOutput(value string) string {
	value = strings.ReplaceAll(value, "\r", "\n")
	lines := strings.Split(value, "\n")
	clean := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") {
			line = "detalle sensible oculto"
		}
		clean = append(clean, line)
		if len(strings.Join(clean, " | ")) > 260 {
			break
		}
	}
	msg := strings.Join(clean, " | ")
	if len(msg) > 320 {
		msg = msg[:317] + "..."
	}
	return msg
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

func corporateEmailResponse(dbSuper *sql.DB, cfg CorporateEmailConfig, account *dbpkg.EmpresaEmailCorporativo, message string, checkWebmail bool, theme string, prefs corporateEmailEmpresaPrefs) map[string]interface{} {
	webmailURL := cfg.WebmailURL
	theme = normalizeCorporateEmailTheme(theme)
	if account != nil {
		webmailURL = corporateEmailAccountWebmailURL(account.WebmailURL, cfg.WebmailURL)
		if strings.TrimSpace(account.WebmailURL) != webmailURL {
			account.WebmailURL = webmailURL
		}
	}
	resp := map[string]interface{}{
		"ok":          true,
		"enabled":     cfg.Enabled,
		"auto_create": cfg.AutoCreate,
		"account":     account,
		"webmail":     webmailURL,
		"domain":      cfg.Domain,
		"theme":       theme,
		"mail_theme":  corporateEmailSnappyMailTheme(theme),
		"preferences": prefs,
	}
	if strings.TrimSpace(message) != "" {
		resp["message"] = strings.TrimSpace(message)
	}
	if checkWebmail {
		resp["webmail_check"] = checkCorporateWebmail(webmailURL)
	}
	accountCanAttemptAutologin := account != nil && (strings.EqualFold(strings.TrimSpace(account.EstadoProvision), "provisionado") || (cfg.ProvisionMode == "mailu_direct" && account.InitialPasswordSet))
	if cfg.Enabled && accountCanAttemptAutologin {
		if token, err := createCorporateEmailAutologinToken(dbSuper, *account); err == nil {
			if autologinURL := corporateEmailAutologinPublicURL(webmailURL, token, theme); autologinURL != "" {
				resp["autologin_url"] = autologinURL
				resp["autologin_expires_seconds"] = 120
			}
		} else {
			resp["autologin_error"] = "Autologin no disponible: " + err.Error()
		}
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
	} else if cfg.ProvisionMode == "mailu_direct" && strings.TrimSpace(cfg.DirectCommand) == "" {
		recommended = "Configura el comando directo de Mailu para crear buzones reales desde la VPS."
	} else if cfg.ProvisionMode == "mailu_direct" {
		recommended = "Provision directa activa: el sistema creara buzones reales en Mailu desde la VPS."
	} else if cfg.ProvisionMode != "mailu_direct" {
		recommended = "El modo manual asigna correos, pero el buzon real se debe crear en Mailu fuera del sistema."
	} else if summary["error"] > 0 {
		recommended = "Prueba Mailu y luego reintenta provisionar las cuentas con error."
	} else if summary["pendiente"] > 0 {
		recommended = "Provisiona las cuentas pendientes para crear los buzones reales."
	}
	return corporateEmailDiagnostics{
		Enabled:             cfg.Enabled,
		AutoCreate:          cfg.AutoCreate,
		ProvisionMode:       cfg.ProvisionMode,
		MailuDirectEnabled:  cfg.Enabled && cfg.ProvisionMode == "mailu_direct",
		MailuAPIURLSet:      strings.TrimSpace(cfg.APIBaseURL) != "",
		MailuAdminSet:       strings.TrimSpace(cfg.APIAdmin) != "",
		MailuAPITokenSet:    cfg.APIPasswordSet,
		EncryptionAvailable: utils.EncryptionAvailable(),
		Webmail:             checkCorporateWebmail(cfg.WebmailURL),
		Accounts:            summary,
		RecommendedAction:   recommended,
	}
}

func testCorporateEmailMailuProvision(cfg CorporateEmailConfig) corporateEmailProvisionResult {
	if !cfg.Enabled {
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_modulo_desactivado", Error: "modulo desactivado"}
	}
	if cfg.ProvisionMode != "mailu_direct" {
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_provision_manual", Error: "modo manual"}
	}
	if strings.TrimSpace(cfg.DirectCommand) == "" && firstNonEmptyEnv("EMAIL_CORPORATIVO_DIRECT_PROVISION_COMMAND", "MAILU_DIRECT_PROVISION_COMMAND") == "" {
		return corporateEmailProvisionResult{OK: false, Status: "pendiente_comando", Error: "Falta comando directo de Mailu"}
	}
	return corporateEmailProvisionResult{OK: true, Status: "direct_ok"}
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
		account.ProvisionProvider = "mailu"
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
				"ok":                   true,
				"config":               cfg,
				"accounts":             accounts,
				"diagnostics":          corporateEmailDiagnosticsFor(cfg, accounts),
				"encryption_available": utils.EncryptionAvailable(),
				"mailu_direct_enabled": cfg.Enabled && cfg.ProvisionMode == "mailu_direct",
			})
			return
		case http.MethodPost, http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "sync" {
				cfg := getCorporateEmailConfig(dbSuper)
				count, err := dbpkg.EnsureEmpresaEmailRowsForExistingEmpresas(dbSuper, dbEmp, cfg.Domain, cfg.WebmailURL, adminEmailFromRequest(r), normalizeCorporateEmailMaxAccounts(cfg.MaxAccounts))
				if err != nil {
					http.Error(w, "No se pudo sincronizar empresas: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "created": count})
				return
			}
			if action == "test_mailu" {
				result := testCorporateEmailMailuProvision(getCorporateEmailConfig(dbSuper))
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
				APIBaseURL    string `json:"mailu_api_base_url"`
				APIAdmin      string `json:"mailu_admin"`
				APIPassword   string `json:"mailu_api_token"`
				QuotaMB       int    `json:"quota_mb"`
				MaxAccounts   int    `json:"max_accounts_per_empresa"`
				DirectCommand string `json:"direct_provision_command"`
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
			cfg.DirectCommand = payload.DirectCommand
			if payload.QuotaMB >= 0 {
				cfg.QuotaMB = payload.QuotaMB
			}
			if payload.MaxAccounts > 0 {
				cfg.MaxAccounts = normalizeCorporateEmailMaxAccounts(payload.MaxAccounts)
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
		empresaID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("empresa_id")), 10, 64)
		if empresaID <= 0 {
			http.Error(w, "empresa_id requerido", http.StatusBadRequest)
			return
		}
		cfg := getCorporateEmailConfig(dbSuper)
		checkWebmail := parseConfigBool(r.URL.Query().Get("check_webmail"), false)
		theme := normalizeCorporateEmailTheme(r.URL.Query().Get("theme"))
		prefs := getCorporateEmailEmpresaPrefs(dbEmp, empresaID)
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			var payload struct {
				AutoOpen        *bool  `json:"auto_open"`
				NoAutoOpen      *bool  `json:"no_auto_open"`
				Password        string `json:"password"`
				NewPassword     string `json:"new_password"`
				ConfirmPassword string `json:"confirm_password"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			if payload.AutoOpen != nil {
				prefs.AutoOpen = *payload.AutoOpen
			}
			if payload.NoAutoOpen != nil {
				prefs.AutoOpen = !*payload.NoAutoOpen
			}
			passwordValue := strings.TrimSpace(payload.NewPassword)
			if passwordValue == "" {
				passwordValue = strings.TrimSpace(payload.Password)
			}
			validatedPassword := ""
			if passwordValue != "" || strings.TrimSpace(payload.ConfirmPassword) != "" {
				newPassword, passErr := validateCorporateEmailNewPassword(passwordValue, payload.ConfirmPassword)
				if passErr != nil {
					http.Error(w, passErr.Error(), http.StatusBadRequest)
					return
				}
				validatedPassword = newPassword
			}
			if err := saveCorporateEmailEmpresaPrefs(dbEmp, empresaID, prefs, adminEmailFromRequest(r)); err != nil {
				http.Error(w, "No se pudo guardar la configuracion del correo corporativo", http.StatusInternalServerError)
				return
			}
			passwordChanged := false
			provisionStatus := ""
			if validatedPassword != "" {
				account, accountErr := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID)
				if accountErr != nil || account == nil {
					http.Error(w, "No se encontro el buzon corporativo de esta empresa", http.StatusNotFound)
					return
				}
				if !utils.EncryptionAvailable() {
					http.Error(w, "CONFIG_ENC_KEY no esta disponible para guardar la clave cifrada", http.StatusInternalServerError)
					return
				}
				encryptedPassword, encErr := utils.EncryptString(validatedPassword)
				if encErr != nil {
					http.Error(w, "No se pudo cifrar la nueva clave", http.StatusInternalServerError)
					return
				}
				if err := dbpkg.UpdateEmpresaEmailCorporativoInitialPassword(dbSuper, empresaID, encryptedPassword, adminEmailFromRequest(r)); err != nil {
					http.Error(w, "No se pudo guardar la nueva clave del buzon", http.StatusInternalServerError)
					return
				}
				passwordChanged = true
				provisionStatus = "clave_guardada"
				if cfg.Enabled && cfg.ProvisionMode == "mailu_direct" {
					result := provisionEmpresaEmailAccountWithTheme(dbSuper, cfg, *account, validatedPassword, theme)
					provisionStatus = result.Status
					if !result.OK {
						account.WebmailURL = corporateEmailAccountWebmailURL(account.WebmailURL, cfg.WebmailURL)
						writeJSON(w, http.StatusOK, map[string]interface{}{
							"ok":                true,
							"preferences":       prefs,
							"account":           account,
							"enabled":           cfg.Enabled,
							"webmail":           account.WebmailURL,
							"password_changed":  true,
							"provision_status":  result.Status,
							"provision_warning": "La clave quedo guardada cifrada, pero Mailu no pudo actualizar el buzon real en este momento.",
							"provision_error":   result.Error,
						})
						return
					}
				}
			}
			var accountResp interface{} = nil
			webmailResp := cfg.WebmailURL
			if account, accountErr := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID); accountErr == nil && account != nil {
				account.WebmailURL = corporateEmailAccountWebmailURL(account.WebmailURL, cfg.WebmailURL)
				webmailResp = account.WebmailURL
				accountResp = account
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":               true,
				"preferences":      prefs,
				"account":          accountResp,
				"enabled":          cfg.Enabled,
				"webmail":          webmailResp,
				"password_changed": passwordChanged,
				"provision_status": provisionStatus,
			})
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		account, err := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if cfg.AutoCreate && dbEmp != nil {
					if empresa, empresaErr := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID); empresaErr == nil && empresa != nil {
						if created, createErr := EnsureEmpresaCorporateEmailAfterCreate(dbSuper, empresa.EmpresaID, empresa.Nombre, adminEmailFromRequest(r)); createErr == nil {
							account = created
						} else {
							writeJSON(w, http.StatusOK, corporateEmailResponse(dbSuper, cfg, nil, "No se pudo generar el email corporativo", checkWebmail, theme, prefs))
							return
						}
					}
				}
				if account != nil {
					if account.WebmailURL == "" {
						account.WebmailURL = cfg.WebmailURL
					}
					writeJSON(w, http.StatusOK, corporateEmailResponse(dbSuper, cfg, account, "Email corporativo generado", checkWebmail, theme, prefs))
					return
				}
				writeJSON(w, http.StatusOK, corporateEmailResponse(dbSuper, cfg, nil, "Sin email corporativo generado", checkWebmail, theme, prefs))
				return
			}
			http.Error(w, "No se pudo consultar email corporativo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if account.WebmailURL == "" {
			account.WebmailURL = cfg.WebmailURL
		} else {
			account.WebmailURL = corporateEmailAccountWebmailURL(account.WebmailURL, cfg.WebmailURL)
		}
		writeJSON(w, http.StatusOK, corporateEmailResponse(dbSuper, cfg, account, "", checkWebmail, theme, prefs))
	}
}

func EmpresaEmailCorporativoAutologinHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		token, err := validateCorporateEmailAutologinToken(dbSuper, r.URL.Query().Get("token"))
		if err != nil {
			writeCorporateEmailAutologinError(w, http.StatusUnauthorized, "El acceso automatico al correo expiro. Vuelve al panel de administrar empresa y actualiza la bandeja.")
			return
		}
		cfg := getCorporateEmailConfig(dbSuper)
		if !cfg.Enabled {
			writeCorporateEmailAutologinError(w, http.StatusForbidden, "El modulo de email corporativo esta desactivado.")
			return
		}
		theme := normalizeCorporateEmailTheme(r.URL.Query().Get("theme"))
		account, err := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, token.EmpresaID)
		if err != nil || account == nil || !strings.EqualFold(strings.TrimSpace(account.Email), strings.TrimSpace(token.Email)) {
			writeCorporateEmailAutologinError(w, http.StatusNotFound, "No se encontro el buzon corporativo de esta empresa.")
			return
		}
		needsMailuProvision := cfg.ProvisionMode == "mailu_direct" && (corporateEmailWebmailEngine() == "snappymail" || !strings.EqualFold(strings.TrimSpace(account.EstadoProvision), "provisionado") || !strings.EqualFold(strings.TrimSpace(account.ProvisionProvider), "mailu"))
		if needsMailuProvision {
			password, passErr := corporateEmailInitialPasswordForProvision(dbSuper, *account)
			if passErr != nil {
				writeCorporateEmailAutologinError(w, http.StatusConflict, "El buzon todavia no esta listo y no se pudo recuperar su clave cifrada para provisionarlo.")
				return
			}
			if result := provisionEmpresaEmailAccountWithTheme(dbSuper, cfg, *account, password, theme); !result.OK {
				writeCorporateEmailAutologinError(w, http.StatusConflict, "El buzon corporativo todavia no pudo provisionarse en Mailu: "+result.Error)
				return
			}
			if refreshed, refreshErr := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, account.EmpresaID); refreshErr == nil && refreshed != nil {
				account = refreshed
			}
		}
		if !strings.EqualFold(strings.TrimSpace(account.EstadoProvision), "provisionado") {
			writeCorporateEmailAutologinError(w, http.StatusConflict, "El buzon corporativo todavia no esta provisionado en Mailu.")
			return
		}
		if strings.TrimSpace(account.WebmailURL) != "" {
			cfg.WebmailURL = corporateEmailAccountWebmailURL(account.WebmailURL, cfg.WebmailURL)
		}
		if corporateEmailWebmailEngine() == "snappymail" {
			password, passErr := corporateEmailInitialPasswordForProvision(dbSuper, *account)
			if passErr != nil {
				writeCorporateEmailAutologinError(w, http.StatusConflict, "No se pudo preparar el acceso automatico al buzon corporativo.")
				return
			}
			redirectURL, setCookies, redirectErr := snappyMailAutologinRedirectURL(cfg, account.Email, password, theme)
			if redirectErr != nil {
				writeCorporateEmailAutologinError(w, http.StatusBadGateway, "No se pudo iniciar sesion automaticamente en la bandeja de correo. "+redirectErr.Error())
				return
			}
			for _, cookieHeader := range setCookies {
				if strings.TrimSpace(cookieHeader) != "" {
					w.Header().Add("Set-Cookie", cookieHeader)
				}
			}
			http.Redirect(w, r, redirectURL, http.StatusFound)
			return
		}
		err = mailuProxyAutologinAndSetCookies(w, cfg, account.Email)
		if err != nil {
			writeCorporateEmailAutologinError(w, http.StatusBadGateway, "No se pudo iniciar sesion automaticamente en la bandeja de correo. "+err.Error())
			return
		}
		http.Redirect(w, r, "/webmail/?_task=mail&_mbox=INBOX", http.StatusFound)
	}
}

func writeCorporateEmailAutologinError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Correo corporativo</title><style>body{font-family:Arial,sans-serif;margin:0;background:#f5f7fb;color:#172033;display:grid;min-height:100vh;place-items:center}.box{max-width:560px;background:#fff;border:1px solid #dbe3ef;border-radius:10px;padding:24px;box-shadow:0 10px 30px rgba(15,23,42,.08)}h1{font-size:22px;margin:0 0 10px}p{line-height:1.45;margin:0}</style></head><body><main class="box"><h1>Bandeja de correo corporativo</h1><p>%s</p></main></body></html>`, htmlEscapeMinimal(message))
}

func htmlEscapeMinimal(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return replacer.Replace(value)
}

func corporateEmailWebmailEngine() string {
	value := strings.ToLower(strings.TrimSpace(firstNonEmptyEnv("MAILU_WEBMAIL", "EMAIL_CORPORATIVO_WEBMAIL_ENGINE")))
	switch value {
	case "roundcube":
		return "roundcube"
	default:
		return "snappymail"
	}
}

func snappyMailAutologinRedirectURL(cfg CorporateEmailConfig, email, password, theme string) (string, []string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return "", nil, fmt.Errorf("buzon sin credenciales internas disponibles")
	}
	webmailURL := strings.TrimSpace(firstNonEmptyEnv("EMAIL_CORPORATIVO_INTERNAL_SNAPPYMAIL_URL", "MAILU_INTERNAL_SNAPPYMAIL_URL"))
	if webmailURL == "" {
		webmailURL = "http://mailu-webmail/"
	}
	if !strings.HasSuffix(webmailURL, "/") {
		webmailURL += "/"
	}
	internalURL, err := url.Parse(webmailURL)
	if err != nil || internalURL.Scheme == "" || internalURL.Host == "" {
		return "", nil, fmt.Errorf("URL interna de SnappyMail invalida")
	}
	ssoURL := strings.TrimRight(internalURL.String(), "/") + "/sso.php"
	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodGet, ssoURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("no se pudo preparar autenticacion")
	}
	req.Header.Set("User-Agent", "PowerfulControlSystem-MailAutologin/1.0")
	req.Header.Set("X-Remote-User", email)
	req.Header.Set("X-Remote-User-Token", password)
	res, err := client.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("SnappyMail rechazo la conexion")
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 32*1024))
	_ = res.Body.Close()
	location := strings.TrimSpace(res.Header.Get("Location"))
	if res.StatusCode < 300 || res.StatusCode >= 400 || location == "" {
		return "", nil, errCorporateEmailAutologinRejected
	}
	return corporateEmailAppendThemeToURI(corporateEmailPublicWebmailRedirect(cfg.WebmailURL, location), theme), res.Header.Values("Set-Cookie"), nil
}

func corporateEmailAppendThemeToURI(rawURI, theme string) string {
	rawURI = strings.TrimSpace(rawURI)
	if rawURI == "" {
		rawURI = "/webmail/"
	}
	theme = normalizeCorporateEmailTheme(theme)
	parsed, err := url.Parse(rawURI)
	if err != nil {
		return rawURI
	}
	if corporateEmailIsSnappyMailSSORedirect(parsed) {
		extra := url.Values{}
		extra.Set("theme", theme)
		extra.Set("mail_theme", corporateEmailSnappyMailTheme(theme))
		extra.Set("pcs_theme", corporateEmailSnappyMailTheme(theme))
		separator := "&"
		if parsed.RawQuery == "" {
			separator = "?"
		}
		if strings.HasSuffix(rawURI, "?") || strings.HasSuffix(rawURI, "&") {
			separator = ""
		}
		return rawURI + separator + extra.Encode()
	}
	query := parsed.Query()
	query.Set("theme", theme)
	query.Set("mail_theme", corporateEmailSnappyMailTheme(theme))
	query.Set("pcs_theme", corporateEmailSnappyMailTheme(theme))
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func corporateEmailIsSnappyMailSSORedirect(parsed *url.URL) bool {
	if parsed == nil {
		return false
	}
	path := strings.ToLower(strings.TrimSpace(parsed.Path))
	rawQuery := strings.TrimSpace(parsed.RawQuery)
	return strings.HasSuffix(path, "/index.php") && (rawQuery == "sso" || strings.HasPrefix(rawQuery, "sso&") || strings.HasPrefix(rawQuery, "sso="))
}

func corporateEmailPublicWebmailRedirect(publicWebmailURL, location string) string {
	location = strings.TrimSpace(location)
	if location == "" {
		return "/webmail/"
	}
	parsedLocation, err := url.Parse(location)
	if err == nil && parsedLocation.IsAbs() {
		return parsedLocation.RequestURI()
	}
	if strings.HasPrefix(location, "/") {
		return location
	}
	publicPath := "/webmail/"
	if parsed, parseErr := url.Parse(strings.TrimSpace(publicWebmailURL)); parseErr == nil {
		pathValue := strings.TrimSpace(parsed.Path)
		if pathValue != "" && pathValue != "/" {
			publicPath = strings.TrimRight(pathValue, "/") + "/"
		}
	}
	return publicPath + location
}

func mailuProxyAutologinAndSetCookies(w http.ResponseWriter, cfg CorporateEmailConfig, email string) error {
	webmailURL := corporateEmailEffectiveWebmailURL(cfg.WebmailURL)
	webmailURL = strings.TrimSpace(webmailURL)
	if webmailURL == "" {
		return fmt.Errorf("webmail sin URL configurada")
	}
	if !strings.HasSuffix(webmailURL, "/") {
		webmailURL += "/"
	}
	baseURL, err := url.Parse(webmailURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return fmt.Errorf("URL de webmail invalida")
	}
	rootURL := *baseURL
	rootURL.Path = ""
	rootURL.RawQuery = ""
	rootURL.Fragment = ""
	ssoURL := strings.TrimRight(rootURL.String(), "/") + "/sso/login?url=/webmail/"
	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest(http.MethodGet, ssoURL, nil)
	if err != nil {
		return fmt.Errorf("no se pudo preparar autenticacion")
	}
	req.Header.Set("User-Agent", "PowerfulControlSystem-MailAutologin/1.0")
	req.Header.Set("X-Auth-Email", strings.ToLower(strings.TrimSpace(email)))
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Mailu rechazo la conexion")
	}
	resBody, _ := io.ReadAll(io.LimitReader(res.Body, 256*1024))
	_ = res.Body.Close()
	cookies := res.Cookies()
	if !mailuSsoLoginLooksSuccessful(cookies, string(resBody), res.StatusCode) {
		return errCorporateEmailAutologinRejected
	}
	for _, cookie := range cookies {
		if strings.TrimSpace(cookie.Name) == "" || strings.TrimSpace(cookie.Value) == "" {
			continue
		}
		http.SetCookie(w, &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	return nil
}

func mailuSsoLoginLooksSuccessful(cookies []*http.Cookie, body string, status int) bool {
	hasCookie := false
	for _, cookie := range cookies {
		if strings.TrimSpace(cookie.Name) != "" && strings.TrimSpace(cookie.Value) != "" {
			hasCookie = true
			break
		}
	}
	lowerBody := strings.ToLower(body)
	if strings.Contains(lowerBody, `name="pw"`) || strings.Contains(lowerBody, "submitadmin") || strings.Contains(lowerBody, "/sso/login") {
		return false
	}
	return status >= 200 && status < 400 && hasCookie
}
