package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	whatsAppConfigEnabled       = "whatsapp.notifications.enabled"
	whatsAppConfigProvider      = "whatsapp.notifications.provider"
	whatsAppConfigAPIVersion    = "whatsapp.notifications.api_version"
	whatsAppConfigPhoneNumberID = "whatsapp.notifications.phone_number_id"
	whatsAppConfigAccessToken   = "whatsapp.notifications.access_token"
	whatsAppConfigTestMode      = "whatsapp.notifications.test_mode"
	whatsAppUsagePrefix         = "whatsapp.usage."
)

type whatsAppEventConfig struct {
	Key             string `json:"key"`
	Nombre          string `json:"nombre"`
	Descripcion     string `json:"descripcion"`
	EmailEnabled    bool   `json:"email_enabled"`
	WhatsAppEnabled bool   `json:"whatsapp_enabled"`
}

type whatsAppNotificationsConfig struct {
	Enabled               bool                  `json:"enabled"`
	Provider              string                `json:"provider"`
	APIVersion            string                `json:"api_version"`
	PhoneNumberID         string                `json:"phone_number_id"`
	AccessTokenConfigured bool                  `json:"access_token_configured"`
	TestMode              bool                  `json:"test_mode"`
	Events                []whatsAppEventConfig `json:"events"`
	UpdatedAt             string                `json:"updated_at,omitempty"`
}

func defaultWhatsAppNotificationEvents() []whatsAppEventConfig {
	return []whatsAppEventConfig{
		{Key: "admin_registro", Nombre: "Registro de administrador", Descripcion: "Confirmaciones y avisos al crear administradores.", EmailEnabled: true, WhatsAppEnabled: false},
		{Key: "usuario_empresa_confirmacion", Nombre: "Invitacion de usuario", Descripcion: "Invitacion y confirmacion de usuarios creados desde empresa.", EmailEnabled: true, WhatsAppEnabled: false},
		{Key: "licencia_activada_pago", Nombre: "Licencia comprada o activada", Descripcion: "Aviso cuando se compra, activa o renueva una licencia.", EmailEnabled: true, WhatsAppEnabled: false},
		{Key: licenciaVencimientoMailType, Nombre: "Vencimiento de licencia", Descripcion: "Recordatorio de licencia proxima a vencer.", EmailEnabled: true, WhatsAppEnabled: false},
		{Key: licenciaRetencionEmpresaMailType, Nombre: "Retencion de empresas vencidas", Descripcion: "Preavisos antes de limpieza de empresas vencidas.", EmailEnabled: true, WhatsAppEnabled: false},
		{Key: "alerta_sistema", Nombre: "Alertas del sistema", Descripcion: "Alertas operativas, seguridad y mantenimiento.", EmailEnabled: true, WhatsAppEnabled: false},
		{Key: "agente_mantenimiento_dian", Nombre: "Agente DIAN", Descripcion: "Noticias o cambios DIAN detectados por agentes.", EmailEnabled: true, WhatsAppEnabled: false},
		{Key: cobranzaClienteNotificationType, Nombre: "Cobranza a clientes", Descripcion: "Recordatorios de cartera configurados por cada empresa.", EmailEnabled: true, WhatsAppEnabled: false},
	}
}

func whatsAppEventConfigKey(eventKey, field string) string {
	return "whatsapp.event." + strings.ToLower(strings.TrimSpace(eventKey)) + "." + field
}

func getWhatsAppNotificationsConfig(dbSuper *sql.DB) whatsAppNotificationsConfig {
	enabledRaw, _, updatedAt, _, _ := dbpkg.GetConfigEntry(dbSuper, whatsAppConfigEnabled)
	provider, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, whatsAppConfigProvider)
	apiVersion, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, whatsAppConfigAPIVersion)
	phoneNumberID, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, whatsAppConfigPhoneNumberID)
	tokenRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, whatsAppConfigAccessToken)
	testModeRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, whatsAppConfigTestMode)
	cfg := whatsAppNotificationsConfig{
		Enabled:               parseEmpresaUsuarioBool(enabledRaw, false),
		Provider:              firstNonEmptyWhatsApp(provider, "meta_cloud"),
		APIVersion:            firstNonEmptyWhatsApp(apiVersion, "v20.0"),
		PhoneNumberID:         strings.TrimSpace(phoneNumberID),
		AccessTokenConfigured: strings.TrimSpace(tokenRaw) != "",
		TestMode:              parseEmpresaUsuarioBool(testModeRaw, false),
		UpdatedAt:             strings.TrimSpace(updatedAt),
		Events:                defaultWhatsAppNotificationEvents(),
	}
	for i := range cfg.Events {
		emailRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, whatsAppEventConfigKey(cfg.Events[i].Key, "email_enabled"))
		waRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, whatsAppEventConfigKey(cfg.Events[i].Key, "whatsapp_enabled"))
		cfg.Events[i].EmailEnabled = parseEmpresaUsuarioBool(emailRaw, cfg.Events[i].EmailEnabled)
		cfg.Events[i].WhatsAppEnabled = parseEmpresaUsuarioBool(waRaw, cfg.Events[i].WhatsAppEnabled)
	}
	return cfg
}

func saveWhatsAppNotificationsConfig(dbSuper *sql.DB, cfg whatsAppNotificationsConfig, accessToken string) error {
	if dbSuper == nil {
		return fmt.Errorf("db super no disponible")
	}
	if err := dbpkg.SetConfigValue(dbSuper, whatsAppConfigEnabled, strconv.FormatBool(cfg.Enabled), false); err != nil {
		return err
	}
	if err := dbpkg.SetConfigValue(dbSuper, whatsAppConfigProvider, firstNonEmptyWhatsApp(cfg.Provider, "meta_cloud"), false); err != nil {
		return err
	}
	if err := dbpkg.SetConfigValue(dbSuper, whatsAppConfigAPIVersion, firstNonEmptyWhatsApp(cfg.APIVersion, "v20.0"), false); err != nil {
		return err
	}
	if err := dbpkg.SetConfigValue(dbSuper, whatsAppConfigPhoneNumberID, strings.TrimSpace(cfg.PhoneNumberID), false); err != nil {
		return err
	}
	if strings.TrimSpace(accessToken) != "" {
		tokenToStore := strings.TrimSpace(accessToken)
		encrypted := false
		if utils.EncryptionAvailable() {
			if enc, err := utils.EncryptString(tokenToStore); err == nil {
				tokenToStore = enc
				encrypted = true
			} else {
				return fmt.Errorf("no se pudo cifrar el token de WhatsApp: %w", err)
			}
		}
		if err := dbpkg.SetConfigValue(dbSuper, whatsAppConfigAccessToken, tokenToStore, encrypted); err != nil {
			return err
		}
	}
	if err := dbpkg.SetConfigValue(dbSuper, whatsAppConfigTestMode, strconv.FormatBool(cfg.TestMode), false); err != nil {
		return err
	}
	for _, event := range cfg.Events {
		key := strings.ToLower(strings.TrimSpace(event.Key))
		if key == "" {
			continue
		}
		if err := dbpkg.SetConfigValue(dbSuper, whatsAppEventConfigKey(key, "email_enabled"), strconv.FormatBool(event.EmailEnabled), false); err != nil {
			return err
		}
		if err := dbpkg.SetConfigValue(dbSuper, whatsAppEventConfigKey(key, "whatsapp_enabled"), strconv.FormatBool(event.WhatsAppEnabled), false); err != nil {
			return err
		}
	}
	return nil
}

func SuperWhatsAppNotificationsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, getWhatsAppNotificationsConfig(dbSuper))
		case http.MethodPut, http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "test" {
				var payload struct {
					To      string `json:"to"`
					Message string `json:"message"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				status, err := sendPCSWhatsAppNotification(dbSuper, "alerta_sistema", payload.To, firstNonEmptyWhatsApp(payload.Message, "Prueba de WhatsApp desde Powerful Control System."), "", "super_whatsapp_test")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "status": status})
				return
			}
			var payload struct {
				Enabled       bool                  `json:"enabled"`
				Provider      string                `json:"provider"`
				APIVersion    string                `json:"api_version"`
				PhoneNumberID string                `json:"phone_number_id"`
				AccessToken   string                `json:"access_token"`
				TestMode      bool                  `json:"test_mode"`
				Events        []whatsAppEventConfig `json:"events"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			cfg := whatsAppNotificationsConfig{
				Enabled:       payload.Enabled,
				Provider:      payload.Provider,
				APIVersion:    payload.APIVersion,
				PhoneNumberID: payload.PhoneNumberID,
				TestMode:      payload.TestMode,
				Events:        mergeWhatsAppEventConfig(payload.Events),
			}
			if err := saveWhatsAppNotificationsConfig(dbSuper, cfg, payload.AccessToken); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, getWhatsAppNotificationsConfig(dbSuper))
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func mergeWhatsAppEventConfig(input []whatsAppEventConfig) []whatsAppEventConfig {
	byKey := map[string]whatsAppEventConfig{}
	for _, event := range input {
		key := strings.ToLower(strings.TrimSpace(event.Key))
		if key != "" {
			event.Key = key
			byKey[key] = event
		}
	}
	out := defaultWhatsAppNotificationEvents()
	for i := range out {
		if incoming, ok := byKey[strings.ToLower(strings.TrimSpace(out[i].Key))]; ok {
			out[i].EmailEnabled = incoming.EmailEnabled
			out[i].WhatsAppEnabled = incoming.WhatsAppEnabled
		}
	}
	return out
}

func isPCSWhatsAppEventEnabled(dbSuper *sql.DB, eventKey string) bool {
	if dbSuper == nil {
		return false
	}
	cfg := getWhatsAppNotificationsConfig(dbSuper)
	if !cfg.Enabled {
		return false
	}
	key := strings.ToLower(strings.TrimSpace(eventKey))
	for _, event := range cfg.Events {
		if strings.EqualFold(event.Key, key) {
			return event.WhatsAppEnabled
		}
	}
	return false
}

func isPCSEmailEventEnabled(dbSuper *sql.DB, eventKey string) bool {
	if dbSuper == nil || strings.TrimSpace(eventKey) == "" {
		return true
	}
	cfg := getWhatsAppNotificationsConfig(dbSuper)
	for _, event := range cfg.Events {
		if strings.EqualFold(event.Key, eventKey) {
			return event.EmailEnabled
		}
	}
	return true
}

func sendPCSWhatsAppForEmailRecipient(dbSuper *sql.DB, eventKey, toEmail, subject, textBody, metadataJSON, actorEmail string) {
	if !isPCSWhatsAppEventEnabled(dbSuper, eventKey) {
		return
	}
	phone := resolveAdminWhatsAppPhoneByEmail(dbSuper, toEmail)
	if phone == "" {
		log.Printf("[whatsapp] evento=%s omitido: destinatario sin telefono asociado", strings.TrimSpace(eventKey))
		return
	}
	message := strings.TrimSpace(subject)
	if body := strings.TrimSpace(textBody); body != "" {
		if message != "" {
			message += "\n\n"
		}
		message += body
	}
	if _, err := sendPCSWhatsAppNotification(dbSuper, eventKey, phone, message, metadataJSON, actorEmail); err != nil {
		log.Printf("[whatsapp] evento=%s destinatario=%s error=%v", strings.TrimSpace(eventKey), safeWhatsAppPhoneForLog(phone), err)
	}
}

func resolveAdminWhatsAppPhoneByEmail(dbSuper *sql.DB, email string) string {
	email = strings.ToLower(strings.TrimSpace(email))
	if _, err := mail.ParseAddress(email); err != nil {
		return ""
	}
	admin, err := dbpkg.GetAdminByEmailFull(dbSuper, email)
	if err != nil || admin == nil {
		return ""
	}
	return normalizeWhatsAppPhone(admin.Telefono)
}

func sendPCSWhatsAppNotification(dbSuper *sql.DB, eventKey, toPhone, message, metadataJSON, actorEmail string) (status string, err error) {
	eventKey = strings.ToLower(strings.TrimSpace(eventKey))
	if eventKey == "" {
		eventKey = "alerta_sistema"
	}
	status = "unknown"
	defer func() {
		recordPCSWhatsAppUsage(dbSuper, eventKey, status)
	}()
	if !isPCSWhatsAppEventEnabled(dbSuper, eventKey) {
		status = "disabled"
		return "disabled", nil
	}
	cfg := getWhatsAppNotificationsConfig(dbSuper)
	to := normalizeWhatsAppPhone(toPhone)
	if to == "" {
		status = "invalid_phone"
		return "invalid_phone", fmt.Errorf("telefono WhatsApp invalido")
	}
	body := truncateWhatsAppMessage(message)
	if body == "" {
		status = "empty_message"
		return "empty_message", fmt.Errorf("mensaje WhatsApp vacio")
	}
	if cfg.TestMode {
		status = "captured_test_mode"
		return "captured_test_mode", nil
	}
	token, err := getDecryptedConfigValue(dbSuper, whatsAppConfigAccessToken)
	if err != nil {
		status = "config_error"
		return "config_error", fmt.Errorf("no se pudo leer token WhatsApp: %w", err)
	}
	if strings.TrimSpace(cfg.PhoneNumberID) == "" || strings.TrimSpace(token) == "" {
		status = "not_configured"
		return "not_configured", fmt.Errorf("WhatsApp API no esta configurada")
	}
	if strings.TrimSpace(cfg.Provider) != "" && !strings.EqualFold(strings.TrimSpace(cfg.Provider), "meta_cloud") {
		status = "provider_unsupported"
		return "provider_unsupported", fmt.Errorf("proveedor WhatsApp no soportado: %s", cfg.Provider)
	}
	status, err = sendMetaCloudWhatsAppText(cfg, token, to, body)
	return status, err
}

func recordPCSWhatsAppUsage(dbSuper *sql.DB, eventKey, status string) {
	if dbSuper == nil {
		return
	}
	date := time.Now().Format("2006-01-02")
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "unknown"
	}
	keys := []string{
		whatsAppUsageCounterKey(date, "total"),
		whatsAppUsageCounterKey(date, normalizeWhatsAppUsageStatus(status)),
	}
	if strings.TrimSpace(eventKey) != "" {
		keys = append(keys, whatsAppUsageCounterKey(date, "event."+strings.ToLower(strings.TrimSpace(eventKey))))
	}
	for _, key := range keys {
		raw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, key)
		current, _ := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
		_ = dbpkg.SetConfigValue(dbSuper, key, strconv.FormatInt(current+1, 10), false)
	}
}

func whatsAppUsageCounterKey(date, metric string) string {
	return whatsAppUsagePrefix + strings.TrimSpace(date) + "." + strings.ToLower(strings.TrimSpace(metric))
}

func normalizeWhatsAppUsageStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "sent":
		return "sent"
	case "captured_test_mode":
		return "captured"
	case "disabled":
		return "disabled"
	case "invalid_phone", "empty_message", "config_error", "not_configured", "provider_unsupported", "provider_error", "send_error", "request_error":
		return "errors"
	default:
		return "other"
	}
}

func sendMetaCloudWhatsAppText(cfg whatsAppNotificationsConfig, token, to, message string) (string, error) {
	version := strings.Trim(strings.TrimSpace(cfg.APIVersion), "/")
	if version == "" {
		version = "v20.0"
	}
	endpoint := "https://graph.facebook.com/" + version + "/" + strings.TrimSpace(cfg.PhoneNumberID) + "/messages"
	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]interface{}{
			"preview_url": true,
			"body":        message,
		},
	}
	raw, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return "request_error", err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "send_error", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return "provider_error", fmt.Errorf("WhatsApp API HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	return "sent", nil
}

func normalizeWhatsAppPhone(raw string) string {
	digits := make([]rune, 0, len(raw))
	for _, ch := range strings.TrimSpace(raw) {
		if ch >= '0' && ch <= '9' {
			digits = append(digits, ch)
		}
	}
	value := strings.TrimLeft(string(digits), "0")
	if len(value) == 10 && strings.HasPrefix(value, "3") {
		value = "57" + value
	}
	if len(value) < 8 || len(value) > 15 {
		return ""
	}
	return value
}

func safeWhatsAppPhoneForLog(phone string) string {
	phone = normalizeWhatsAppPhone(phone)
	if len(phone) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

func truncateWhatsAppMessage(message string) string {
	message = strings.TrimSpace(message)
	if len([]rune(message)) <= 3900 {
		return message
	}
	runes := []rune(message)
	return string(runes[:3900])
}

func firstNonEmptyWhatsApp(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
