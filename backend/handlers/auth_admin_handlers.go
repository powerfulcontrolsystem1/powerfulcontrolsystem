package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"net/mail"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"fmt"
	"net/smtp"

	"github.com/you/pos-backend/auth"
	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const googleOAuthRedirectCookieName = "oauth_redirect_url"
const browserSessionStateCookieName = "browser_session_active"

// SessionCookieSecure resuelve si una cookie de sesión debe emitirse como Secure
// considerando terminación TLS local o por proxy inverso.
func SessionCookieSecure(r *http.Request) bool {
	return resolveOAuthScheme(r) == "https"
}

// AdminRegisterHandler registra un administrador (envía email de confirmación).
func AdminRegisterHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		payload.Name = strings.TrimSpace(payload.Name)
		payload.Password = strings.TrimSpace(payload.Password)
		if payload.Email == "" {
			http.Error(w, "email required", http.StatusBadRequest)
			return
		}
		if _, err := mail.ParseAddress(payload.Email); err != nil {
			http.Error(w, "invalid email", http.StatusBadRequest)
			return
		}

		// crear o actualizar administrador básico
		if err := dbpkg.UpsertAdministrador(dbSuper, payload.Email, payload.Name, "administrador", ""); err != nil {
			log.Println("AdminRegisterHandler upsert error:", err)
			http.Error(w, "failed to create administrador", http.StatusInternalServerError)
			return
		}

		// si enviaron contraseña, guardarla (hash+salt)
		if payload.Password != "" {
			hash, salt, err := generateEmpresaUsuarioPasswordHash(payload.Password)
			if err != nil {
				log.Println("AdminRegisterHandler hash error:", err)
			} else {
				if err := dbpkg.SetAdministradorPassword(dbSuper, payload.Email, hash, salt); err != nil {
					log.Println("AdminRegisterHandler set password error:", err)
				}
			}
		}

		// generar token confirmación
		token, expira, nerr := newEmailConfirmationTokenAndExpiration()
		if nerr != nil {
			log.Println("AdminRegisterHandler token gen error:", nerr)
		} else {
			if err := dbpkg.SetAdministradorConfirmToken(dbSuper, payload.Email, token, expira); err != nil {
				log.Println("AdminRegisterHandler set confirm token error:", err)
			}
		}

		// enviar correo de confirmación
		if _, err := sendAdminConfirmationEmail(r, dbSuper, payload.Email, payload.Name, token); err != nil {
			log.Println("AdminRegisterHandler send email error:", err)
			// no fallamos la creación por error de email; devolver info
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "email_sent": false, "error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "email_sent": true})
	}
}

// ConfirmarAdminHandler confirma el correo vía token y muestra una página simple.
func ConfirmarAdminHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		token := strings.TrimSpace(q.Get("token"))
		if token == "" {
			http.Error(w, "token required", http.StatusBadRequest)
			return
		}
		if _, err := dbpkg.ConfirmAdministradorByToken(dbSuper, token); err != nil {
			log.Println("ConfirmarAdminHandler error:", err)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<html><body><h3>Token inválido o expirado</h3><p>Si ya confirmaste, intenta iniciar sesión: <a href="/login.html">Iniciar</a></p></body></html>`))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body><h3>Correo confirmado</h3><p>Tu cuenta ha sido confirmada. Ahora puedes <a href="/login.html">iniciar sesión</a>.</p></body></html>`))
	}
}

// AdminLoginHandler maneja login por email/contraseña para administradores.
func AdminLoginHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct{
			Email string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		payload.Password = strings.TrimSpace(payload.Password)
		if payload.Email == "" || payload.Password == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, payload.Email)
		if err != nil {
			log.Println("AdminLoginHandler get admin error:", err)
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if admin.EmailConfirmado != 1 {
			http.Error(w, "email not confirmed", http.StatusForbidden)
			return
		}
		if admin.PasswordSet != 1 || strings.TrimSpace(admin.PasswordHash) == "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "password_setup_required": true})
			return
		}
		// verificar contraseña
		expected := hashEmpresaUsuarioPassword(payload.Password, admin.PasswordSalt)
		if expected != strings.TrimSpace(admin.PasswordHash) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		// crear sesión
		token, terr := utils.GenerateSecureToken(32)
		if terr != nil {
			log.Println("AdminLoginHandler token gen error:", terr)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.CreateSession(dbSuper, admin.Email, r.RemoteAddr, r.UserAgent(), token); err != nil {
			log.Println("AdminLoginHandler create session error:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		cookie := &http.Cookie{
			Name: "session_token",
			Value: token,
			Path: "/",
			HttpOnly: true,
			MaxAge: 86400,
			Secure: SessionCookieSecure(r),
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)
		SetBrowserSessionStateCookie(w, r, true)

		redirectURL := "/seleccionar_empresa.html"
		if strings.ToLower(strings.TrimSpace(admin.Role)) == "super_administrador" {
			redirectURL = "/super_administrador.html"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "redirect_url": redirectURL})
	}
}

// AdminRequestPasswordRecoveryHandler solicita envío de token de recuperación.
func AdminRequestPasswordRecoveryHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct{ Email string `json:"email"` }
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		if payload.Email == "" {
			http.Error(w, "email required", http.StatusBadRequest)
			return
		}
		// generar token
		token, expira, nerr := newPasswordRecoveryTokenAndExpiration()
		if nerr != nil {
			log.Println("AdminRequestPasswordRecoveryHandler token gen error:", nerr)
		} else {
			if err := dbpkg.SetAdministradorPasswordResetToken(dbSuper, payload.Email, token, expira); err != nil {
				log.Println("AdminRequestPasswordRecoveryHandler set token error:", err)
			}
		}
		if _, err := sendAdminPasswordRecoveryEmail(r, dbSuper, payload.Email, "", token); err != nil {
			log.Println("AdminRequestPasswordRecoveryHandler send mail error:", err)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "email_sent": false, "error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "email_sent": true})
	}
}

// AdminResetPasswordHandler restablece contraseña usando token.
func AdminResetPasswordHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct{
			Email string `json:"email"`
			Token string `json:"token"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		payload.Token = strings.TrimSpace(payload.Token)
		payload.Password = strings.TrimSpace(payload.Password)
		if payload.Email == "" || payload.Token == "" || payload.Password == "" {
			http.Error(w, "email, token and password required", http.StatusBadRequest)
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, payload.Email)
		if err != nil {
			log.Println("AdminResetPasswordHandler get admin error:", err)
			http.Error(w, "invalid token or email", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(admin.PasswordResetToken) == "" || strings.TrimSpace(admin.PasswordResetToken) != payload.Token {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}
		// verificar expiración
		if admin.PasswordResetExpira != "" {
			if t, ok := parseEmpresaUsuarioDateTime(admin.PasswordResetExpira); ok {
				if time.Now().After(t) {
					http.Error(w, "token expirado", http.StatusBadRequest)
					return
				}
			}
		}
		// generar hash y guardar
		hash, salt, err := generateEmpresaUsuarioPasswordHash(payload.Password)
		if err != nil {
			log.Println("AdminResetPasswordHandler hash error:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.SetAdministradorPassword(dbSuper, payload.Email, hash, salt); err != nil {
			log.Println("AdminResetPasswordHandler set password error:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// limpiar token
		if err := dbpkg.ClearAdministradorPasswordResetToken(dbSuper, admin.ID); err != nil {
			log.Println("AdminResetPasswordHandler clear token error:", err)
		}
		// crear sesión y responder
		tokenSession, terr := utils.GenerateSecureToken(32)
		if terr != nil {
			log.Println("AdminResetPasswordHandler token gen error:", terr)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if err := dbpkg.CreateSession(dbSuper, admin.Email, r.RemoteAddr, r.UserAgent(), tokenSession); err != nil {
			log.Println("AdminResetPasswordHandler create session error:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		cookie := &http.Cookie{
			Name: "session_token",
			Value: tokenSession,
			Path: "/",
			HttpOnly: true,
			MaxAge: 86400,
			Secure: SessionCookieSecure(r),
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)
		SetBrowserSessionStateCookie(w, r, true)

		redirectURL := "/seleccionar_empresa.html"
		if strings.ToLower(strings.TrimSpace(admin.Role)) == "super_administrador" {
			redirectURL = "/super_administrador.html"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true, "redirect_url": redirectURL})
	}
}

// sendAdminConfirmationEmail envía el correo de confirmación para administradores.
func sendAdminConfirmationEmail(r *http.Request, dbSuper *sql.DB, toEmail, toName, token string) (string, error) {
	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	confirmURL := strings.TrimRight(baseURL, "/") + "/auth/confirmar_admin?token=" + url.QueryEscape(token)
	loginURL := strings.TrimRight(baseURL, "/") + "/login.html"

	safeName := strings.TrimSpace(toName)
	if safeName == "" {
		safeName = "administrador"
	}

	subject := "Confirma tu correo - Powerful Control System"
	body := "Hola " + safeName + ",\r\n\r\n" +
		"Para activar tu cuenta, haz clic en el siguiente enlace:\r\n" +
		confirmURL + "\r\n\r\n" +
		"Después de confirmar, inicia sesión aquí:\r\n" +
		loginURL + "\r\n\r\n" +
		"Si no solicitaste esta cuenta, ignora este mensaje.\r\n"

	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"confirm_url":%q,"login_url":%q,"mail_mode":"test"}`, confirmURL, loginURL)
		if err := captureEmpresaUsuarioMailNotification(dbSuper, "confirmacion_correo_admin", 0, toEmail, subject, body, token, metadataJSON, adminEmailFromRequest(r)); err != nil {
			return confirmURL, err
		}
		return confirmURL, nil
	}

	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil { return "", err }
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" { return "", fmt.Errorf("gmail.smtp_email no configurado") }
	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil { return "", err }
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" { return "", fmt.Errorf("gmail.smtp_app_password no configurado") }

	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")
	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" { smtpHost = "smtp.gmail.com" }
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" { smtpPort = "587" }
	fromName = strings.TrimSpace(fromName)
	if fromName == "" { fromName = "Powerful Control System" }

	mailHostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, err := net.SplitHostPort(smtpHost); err == nil {
			mailHostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}
	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body
	if err := smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, []byte(msg)); err != nil {
		return confirmURL, err
	}
	return confirmURL, nil
}

// sendAdminPasswordRecoveryEmail envía token de recuperación para administrador.
func sendAdminPasswordRecoveryEmail(r *http.Request, dbSuper *sql.DB, toEmail, toName, token string) (string, error) {
	baseURL := resolveBaseURLForConfirmation(r, dbSuper)
	resetHintURL := strings.TrimRight(baseURL, "/") + "/login.html?email=" + url.QueryEscape(toEmail) + "&token_recuperacion=" + url.QueryEscape(token)
	safeName := strings.TrimSpace(toName)
	if safeName == "" { safeName = "administrador" }
	subject := "Recuperacion de contraseña - Powerful Control System"
	body := "Hola " + safeName + ",\r\n\r\n" +
		"Recibimos una solicitud para restablecer tu contraseña. Token de recuperación:\r\n" + token + "\r\n\r\n" +
		"Abre el login y usa el token para completar el restablecimiento:\r\n" + resetHintURL + "\r\n\r\n" +
		"Si no solicitaste este cambio, ignora este mensaje.\r\n"
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"reset_hint_url":%q,"mail_mode":"test"}`, resetHintURL)
		if err := captureEmpresaUsuarioMailNotification(dbSuper, "recuperacion_password_admin", 0, toEmail, subject, body, token, metadataJSON, adminEmailFromRequest(r)); err != nil {
			return resetHintURL, err
		}
		return resetHintURL, nil
	}
	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil { return "", err }
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" { return "", fmt.Errorf("gmail.smtp_email no configurado") }
	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil { return "", err }
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" { return "", fmt.Errorf("gmail.smtp_app_password no configurado") }
	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")
	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" { smtpHost = "smtp.gmail.com" }
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" { smtpPort = "587" }
	fromName = strings.TrimSpace(fromName)
	if fromName == "" { fromName = "Powerful Control System" }
	mailHostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, err := net.SplitHostPort(smtpHost); err == nil { mailHostForAuth = h }
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") { addr = smtpHost + ":" + smtpPort }
	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body
	if err := smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, []byte(msg)); err != nil {
		return resetHintURL, err
	}
	return resetHintURL, nil
}

// SetBrowserSessionStateCookie emite una señal visible para el cliente que indica
// si existe una sesión autenticada activa sin exponer el token real HttpOnly.
func SetBrowserSessionStateCookie(w http.ResponseWriter, r *http.Request, active bool) {
	if w == nil {
		return
	}

	value := ""
	maxAge := -1
	if active {
		value = "1"
		maxAge = 86400
	}

	http.SetCookie(w, &http.Cookie{
		Name:     browserSessionStateCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: false,
		MaxAge:   maxAge,
		Secure:   SessionCookieSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func firstForwardedValue(raw string) string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func resolveOAuthScheme(r *http.Request) string {
	if r == nil {
		return "http"
	}

	for _, header := range []string{"X-Forwarded-Proto", "X-Forwarded-Scheme"} {
		value := strings.ToLower(firstForwardedValue(r.Header.Get(header)))
		if value == "https" {
			return "https"
		}
		if value == "http" {
			return "http"
		}
	}

	if r.TLS != nil {
		return "https"
	}

	return "http"
}

func resolveOAuthHost(r *http.Request) string {
	if r == nil {
		return ""
	}

	if host := firstForwardedValue(r.Header.Get("X-Forwarded-Host")); host != "" {
		return host
	}

	return strings.TrimSpace(r.Host)
}

func splitHostPortSafe(rawHost string) string {
	trimmed := strings.TrimSpace(rawHost)
	if trimmed == "" {
		return ""
	}
	hostOnly, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		return strings.TrimSpace(hostOnly)
	}
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		return strings.Trim(strings.TrimSpace(trimmed), "[]")
	}
	return trimmed
}

func isLoopbackHost(rawHost string) bool {
	host := strings.ToLower(splitHostPortSafe(rawHost))
	if host == "" {
		return false
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func adaptConfiguredLoopbackRedirect(r *http.Request, configured string) string {
	trimmed := strings.TrimSpace(configured)
	if trimmed == "" {
		return ""
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}

	requestHost := resolveOAuthHost(r)
	if requestHost == "" {
		return trimmed
	}

	desiredScheme := resolveOAuthScheme(r)
	if !isLoopbackHost(requestHost) {
		desiredScheme = "https"
	}

	// Si host y esquema ya coinciden con el entorno actual, conservar configuración.
	configHost := splitHostPortSafe(parsed.Host)
	reqHost := splitHostPortSafe(requestHost)
	if strings.EqualFold(configHost, reqHost) && strings.EqualFold(parsed.Scheme, desiredScheme) && parsed.Path == "/auth/google/callback" {
		return trimmed
	}

	// Adaptar la URL al host/esquema real de la petición para que funcione
	// tanto en local (localhost) como en VPS (dominio real).
	parsed.Scheme = desiredScheme
	parsed.Host = requestHost
	parsed.Path = "/auth/google/callback"
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func resolveOAuthRedirectURL(r *http.Request, configuredRedirectURL string) string {
	configured := adaptConfiguredLoopbackRedirect(r, configuredRedirectURL)
	if configured != "" {
		return configured
	}

	host := resolveOAuthHost(r)
	if host == "" {
		host = "localhost:8080"
	}
	scheme := resolveOAuthScheme(r)
	if !isLoopbackHost(host) {
		scheme = "https"
	}

	return scheme + "://" + host + "/auth/google/callback"
}

func isValidOAuthRedirectURL(raw string) bool {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return false
	}
	if parsed.Host == "" {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return parsed.Path == "/auth/google/callback"
}

func normalizeGoogleLoginHint(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	parsed, err := mail.ParseAddress(trimmed)
	if err != nil {
		return ""
	}
	if !strings.EqualFold(strings.TrimSpace(parsed.Address), trimmed) {
		return ""
	}
	return trimmed
}

// HandleGoogleLogin devuelve un http.HandlerFunc configurado con clientID y redirectURL
func HandleGoogleLogin(clientID, redirectURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := "state-token"
		if clientID == "" {
			http.Error(w, "Acceso bloqueado: configuración incompleta (GOOGLE_CLIENT_ID no definido)", http.StatusInternalServerError)
			return
		}
		log.Printf("handleGoogleLogin: oauth redirect requested (client configured=%t)", clientID != "")
		effectiveRedirectURL := resolveOAuthRedirectURL(r, redirectURL)
		q := r.URL.Query()
		loginHint := normalizeGoogleLoginHint(q.Get("login_hint"))
		vals := url.Values{
			"client_id":              {clientID},
			"redirect_uri":           {effectiveRedirectURL},
			"response_type":          {"code"},
			"scope":                  {"openid email profile"},
			"include_granted_scopes": {"true"},
			"access_type":            {"offline"},
			"state":                  {state},
			// Forzar selección explícita de cuenta sin pedir consentimiento extra en cada login.
			"prompt": {"select_account"},
		}
		if loginHint != "" {
			vals.Set("login_hint", loginHint)
		}
		http.SetCookie(w, &http.Cookie{
			Name:     googleOAuthRedirectCookieName,
			Value:    url.QueryEscape(effectiveRedirectURL),
			Path:     "/auth/google",
			HttpOnly: true,
			MaxAge:   600,
			Secure:   SessionCookieSecure(r),
			SameSite: http.SameSiteLaxMode,
		})
		authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + vals.Encode()
		log.Printf("handleGoogleLogin: redirecting to OAuth provider")
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// HandleGoogleCallback procesa el callback OAuth y crea sesión/administrador
func HandleGoogleCallback(dbEmpresas *sql.DB, dbSuper *sql.DB, clientID, clientSecret, redirectURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errStr := q.Get("error"); errStr != "" {
			http.Error(w, "error from provider: "+errStr, http.StatusBadRequest)
			return
		}
		code := q.Get("code")
		if code == "" {
			http.Error(w, "code not found", http.StatusBadRequest)
			return
		}

		effectiveRedirectURL := resolveOAuthRedirectURL(r, redirectURL)
		if ck, err := r.Cookie(googleOAuthRedirectCookieName); err == nil {
			decodedValue, decodeErr := url.QueryUnescape(strings.TrimSpace(ck.Value))
			if decodeErr == nil && isValidOAuthRedirectURL(decodedValue) {
				effectiveRedirectURL = decodedValue
			}
			http.SetCookie(w, &http.Cookie{
				Name:     googleOAuthRedirectCookieName,
				Value:    "",
				Path:     "/auth/google",
				HttpOnly: true,
				MaxAge:   -1,
				Secure:   SessionCookieSecure(r),
				SameSite: http.SameSiteLaxMode,
			})
		}

		tokenResp, err := auth.ExchangeCodeForToken(code, clientID, clientSecret, effectiveRedirectURL)
		if err != nil {
			log.Println("token exchange error:", err)
			http.Error(w, "token exchange failed", http.StatusInternalServerError)
			return
		}

		userinfo, err := auth.FetchUserInfo(tokenResp.AccessToken)
		if err != nil {
			log.Println("fetch userinfo error:", err)
			http.Error(w, "failed to fetch userinfo", http.StatusInternalServerError)
			return
		}

		// Determinar rol existente (si aplica) para preservarlo
		existingAdmin, _ := dbpkg.GetAdminByEmail(dbSuper, userinfo.Email)
		roleToSet := "administrador"
		if existingAdmin != nil && existingAdmin.Role != "" {
			roleToSet = existingAdmin.Role
		}
		if err := dbpkg.UpsertAdministrador(dbSuper, userinfo.Email, userinfo.Name, roleToSet, userinfo.Picture); err != nil {
			log.Println("db upsert administradores error:", err)
		}

		if err := dbpkg.UpsertUser(dbEmpresas, userinfo.Email, userinfo.Name); err != nil {
			log.Println("db upsert users error:", err)
		}

		if err := dbpkg.EnsureUserEmpresa(dbEmpresas, userinfo.Email, "Empresa de "+userinfo.Name); err != nil {
			log.Println("db ensure empresa error:", err)
		}

		if err := dbpkg.EnsureSuperContractSchema(dbSuper); err != nil {
			log.Println("contract schema error:", err)
			http.Error(w, "failed to prepare contract metadata", http.StatusInternalServerError)
			return
		}
		currentContract, err := dbpkg.GetCurrentSuperContract(dbSuper)
		if err != nil || currentContract == nil {
			log.Println("load current contract error:", err)
			http.Error(w, "failed to load current contract", http.StatusInternalServerError)
			return
		}

		// La aceptación se decide únicamente por registro persistido por administrador.
		accepted := false
		if adminNow, err := dbpkg.GetAdminByEmail(dbSuper, userinfo.Email); err == nil && adminNow != nil {
			acceptance, acceptanceErr := dbpkg.GetAdministradorContratoAceptacion(dbSuper, userinfo.Email)
			if acceptanceErr == nil && adminNow.AceptaContrato == 1 && acceptance.Acepta && acceptance.Version >= currentContract.Version {
				accepted = true
			}
		}

		if accepted {
			token, err := utils.GenerateSecureToken(32)
			if err != nil {
				log.Println("failed to generate session token:", err)
				token = userinfo.Sub
			}
			ip := r.RemoteAddr
			ua := r.UserAgent()
			if err := dbpkg.CreateSession(dbSuper, userinfo.Email, ip, ua, token); err != nil {
				log.Println("create session error:", err)
			}
			cookie := &http.Cookie{
				Name:     "session_token",
				Value:    token,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   86400,
				Secure:   SessionCookieSecure(r),
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, cookie)
			SetBrowserSessionStateCookie(w, r, true)

			admin, err := dbpkg.GetAdminByEmail(dbSuper, userinfo.Email)
			if err != nil || admin == nil {
				log.Println("warning: no admin found, redirecting to seleccionar_empresa:", err)
				http.Redirect(w, r, "/seleccionar_empresa.html", http.StatusFound)
				return
			}
			if admin.Role == "super_administrador" {
				http.Redirect(w, r, "/super_administrador.html", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/seleccionar_empresa.html", http.StatusFound)
			return
		}

		// Si no aceptó, redirigir a página de aceptación server-side con payload cifrado.
		if userinfo.Email != "" {
			next := "/seleccionar_empresa.html"
			if roleToSet == "super_administrador" {
				next = "/super_administrador.html"
			}
			payload := map[string]interface{}{
				"email": userinfo.Email,
				"exp":   time.Now().Add(10 * time.Minute).Unix(),
				"next":  next,
			}
			pb, _ := json.Marshal(payload)
			enc, err := utils.EncryptString(string(pb))
			if err != nil {
				log.Printf("failed to encrypt accept payload: %v", err)
				http.Error(w, "failed to prepare contract acceptance", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/accept.html?payload="+url.QueryEscape(enc), http.StatusFound)
		} else {
			http.Redirect(w, r, "/login.html", http.StatusFound)
		}
		return
	}
}

// ListAdministradoresHandler devuelve JSON con la lista de administradores (super DB)
func ListAdministradoresHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		admins, err := dbpkg.GetAdministradores(dbSuper)
		if err != nil {
			http.Error(w, "failed to query administradores", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(admins)
	}
}

// ListSesionesHandler devuelve JSON con la lista de sesiones (super DB)
func ListSesionesHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sesiones, err := dbpkg.GetSesiones(dbSuper)
		if err != nil {
			http.Error(w, "failed to query sesiones", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sesiones)
	}
}

// AdministradoresHandler maneja CRUD de administradores y activar/desactivar
func AdministradoresHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			admins, err := dbpkg.GetAdministradores(dbSuper)
			if err != nil {
				http.Error(w, "failed to query administradores", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(admins)
			return
		case http.MethodPost:
			var payload struct{ Email, Name, Role, Photo string }
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.Email == "" {
				http.Error(w, "email required", http.StatusBadRequest)
				return
			}
			if payload.Role == "" {
				payload.Role = "administrador"
			}
			if err := dbpkg.UpsertAdministrador(dbSuper, payload.Email, payload.Name, payload.Role, payload.Photo); err != nil {
				http.Error(w, "failed to upsert administrador: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodPut:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			if q.Get("action") == "activar" {
				estado := q.Get("estado")
				if estado == "" {
					activoStr := q.Get("activo")
					if activoStr == "1" {
						estado = "activo"
					} else {
						estado = "inactivo"
					}
				}
				if err := dbpkg.SetAdministradorEstado(dbSuper, id, estado); err != nil {
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var payloadUpdate struct{ Name, Role string }
			if err := json.NewDecoder(r.Body).Decode(&payloadUpdate); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateAdministrador(dbSuper, id, payloadUpdate.Name, payloadUpdate.Role); err != nil {
				http.Error(w, "failed to update administrador: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteAdministrador(dbSuper, id); err != nil {
				http.Error(w, "failed to delete administrador: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// TiposEmpresasHandler maneja GET/POST/PUT/DELETE para tipos_de_empresas
func TiposEmpresasHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tipos, err := dbpkg.GetTiposEmpresas(dbSuper)
			if err != nil {
				http.Error(w, "failed to query tipos_de_empresas", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tipos)
			return
		case http.MethodPost:
			var payload struct{ Nombre, Observaciones string }
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.Nombre == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateTipoEmpresa(dbSuper, payload.Nombre, payload.Observaciones)
			if err != nil {
				http.Error(w, "failed to create tipo_empresa: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			// permitir activar/desactivar vía query param
			if q.Get("action") == "activar" {
				estado := q.Get("estado")
				if estado == "" {
					// soportar parámetro activo=1/0
					activoStr := q.Get("activo")
					if activoStr == "" {
						http.Error(w, "estado or activo required", http.StatusBadRequest)
						return
					}
					if activoStr == "1" {
						estado = "activo"
					} else {
						estado = "inactivo"
					}
				}
				if err := dbpkg.SetTipoEmpresaActivo(dbSuper, id, estado); err != nil {
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var payloadUpdate struct{ Nombre, Observaciones string }
			if err := json.NewDecoder(r.Body).Decode(&payloadUpdate); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateTipoEmpresa(dbSuper, id, payloadUpdate.Nombre, payloadUpdate.Observaciones); err != nil {
				http.Error(w, "failed to update: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			q := r.URL.Query()
			idStr := q.Get("id")
			if idStr == "" {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteTipoEmpresa(dbSuper, id); err != nil {
				http.Error(w, "failed to delete: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
