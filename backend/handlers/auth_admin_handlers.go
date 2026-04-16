package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/you/pos-backend/auth"
	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const googleOAuthRedirectCookieName = "oauth_redirect_url"
const browserSessionStateCookieName = "browser_session_active"
const minAdminPasswordLength = 8
const googlePasswordSetupPagePath = "/registrar_contrasena_usuario_de_google.html"

// SessionCookieSecure resuelve si una cookie de sesión debe emitirse como Secure
// considerando terminación TLS local o por proxy inverso.
func SessionCookieSecure(r *http.Request) bool {
	return resolveOAuthScheme(r) == "https"
}

func writeAdminAuthJSON(w http.ResponseWriter, status int, payload map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeAdminAuthError(w http.ResponseWriter, status int, message string) {
	writeAdminAuthJSON(w, status, map[string]interface{}{"ok": false, "message": message})
}

func countAdminPhoneDigits(raw string) int {
	total := 0
	for _, ch := range strings.TrimSpace(raw) {
		if ch >= '0' && ch <= '9' {
			total++
		}
	}
	return total
}

func resolveAdminPostLoginRedirect(admin *dbpkg.Admin) string {
	if admin == nil {
		return "/seleccionar_empresa.html"
	}
	if admin.PasswordSet != 1 || strings.TrimSpace(admin.PasswordHash) == "" {
		return googlePasswordSetupPagePath
	}
	if strings.EqualFold(strings.TrimSpace(admin.Role), "super_administrador") {
		return "/super_administrador.html"
	}
	return "/seleccionar_empresa.html"
}

// AdminRegisterHandler registra un administrador (envía email de confirmación).
func AdminRegisterHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeAdminAuthError(w, http.StatusMethodNotAllowed, "Método no permitido.")
			return
		}
		var payload struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Telefono string `json:"telefono"`
			Pais     string `json:"pais"`
			Ciudad   string `json:"ciudad"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "El formulario de registro es inválido.")
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		payload.Name = strings.TrimSpace(payload.Name)
		payload.Telefono = strings.TrimSpace(payload.Telefono)
		payload.Pais = strings.TrimSpace(payload.Pais)
		payload.Ciudad = strings.TrimSpace(payload.Ciudad)
		payload.Password = strings.TrimSpace(payload.Password)
		if payload.Email == "" || payload.Name == "" || payload.Telefono == "" || payload.Pais == "" || payload.Ciudad == "" || payload.Password == "" {
			writeAdminAuthError(w, http.StatusBadRequest, "Debes completar correo, nombre completo, celular, pais, ciudad y contraseña.")
			return
		}
		if _, err := mail.ParseAddress(payload.Email); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "El correo electrónico no es válido.")
			return
		}
		if countAdminPhoneDigits(payload.Telefono) < 7 {
			writeAdminAuthError(w, http.StatusBadRequest, "El teléfono debe contener al menos 7 dígitos.")
			return
		}
		if len([]rune(payload.Ciudad)) < 2 {
			writeAdminAuthError(w, http.StatusBadRequest, "Debes indicar una ciudad valida.")
			return
		}
		if len(payload.Password) < minAdminPasswordLength {
			writeAdminAuthError(w, http.StatusBadRequest, "La contraseña debe tener mínimo 8 caracteres.")
			return
		}

		existing, err := dbpkg.GetAdminByEmailFull(dbSuper, payload.Email)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			log.Println("AdminRegisterHandler get existing admin error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo validar el estado de la cuenta.")
			return
		}
		if err == nil && existing != nil && existing.EmailConfirmado == 1 {
			writeAdminAuthError(w, http.StatusConflict, "Ya existe una cuenta administrativa confirmada con ese correo. Inicia sesión o recupera tu contraseña.")
			return
		}

		// crear o actualizar administrador básico
		if err := dbpkg.UpsertAdministrador(dbSuper, payload.Email, payload.Name, "administrador", ""); err != nil {
			log.Println("AdminRegisterHandler upsert error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo crear la cuenta administrativa.")
			return
		}

		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, payload.Email)
		if err != nil || admin == nil {
			log.Println("AdminRegisterHandler reload admin error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo completar el registro de la cuenta.")
			return
		}
		if err := dbpkg.UpdateAdministradorProfile(dbSuper, admin.ID, payload.Name, payload.Telefono, payload.Email, payload.Pais, payload.Ciudad); err != nil {
			log.Println("AdminRegisterHandler update profile error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo guardar la información de contacto.")
			return
		}

		hash, salt, err := generateEmpresaUsuarioPasswordHash(payload.Password)
		if err != nil {
			log.Println("AdminRegisterHandler hash error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo proteger la contraseña del administrador.")
			return
		}
		if err := dbpkg.SetAdministradorPassword(dbSuper, payload.Email, hash, salt); err != nil {
			log.Println("AdminRegisterHandler set password error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo guardar la contraseña del administrador.")
			return
		}

		// generar token confirmación
		token, expira, nerr := newEmailConfirmationTokenAndExpiration()
		if nerr != nil {
			log.Println("AdminRegisterHandler token gen error:", nerr)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo generar el enlace de confirmación.")
			return
		} else {
			if err := dbpkg.SetAdministradorConfirmToken(dbSuper, payload.Email, token, expira); err != nil {
				log.Println("AdminRegisterHandler set confirm token error:", err)
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo activar la confirmación por correo.")
				return
			}
		}

		// enviar correo de confirmación
		if _, err := sendAdminConfirmationEmail(r, dbSuper, payload.Email, payload.Name, token); err != nil {
			log.Println("AdminRegisterHandler send email error:", err)
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"email_sent": false,
				"message":    "La cuenta fue creada, pero no se pudo enviar el correo de confirmación. Revisa la configuración SMTP.",
				"error":      err.Error(),
			})
			return
		}

		writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"email_sent": true,
			"message":    "Registro exitoso. Revisa tu correo para confirmar la cuenta antes de iniciar sesión.",
		})
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
			writeAdminAuthError(w, http.StatusMethodNotAllowed, "Método no permitido.")
			return
		}
		var payload struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "El formulario de acceso es inválido.")
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		payload.Password = strings.TrimSpace(payload.Password)
		if payload.Email == "" || payload.Password == "" {
			writeAdminAuthError(w, http.StatusBadRequest, "Debes ingresar correo y contraseña.")
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, payload.Email)
		if err != nil {
			log.Println("AdminLoginHandler get admin error:", err)
			writeAdminAuthError(w, http.StatusUnauthorized, "Credenciales inválidas.")
			return
		}
		if admin.EmailConfirmado != 1 {
			writeAdminAuthError(w, http.StatusForbidden, "Debes confirmar tu correo antes de iniciar sesión.")
			return
		}
		if admin.PasswordSet != 1 || strings.TrimSpace(admin.PasswordHash) == "" {
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": false, "password_setup_required": true, "message": "Tu cuenta todavía no tiene una contraseña activa."})
			return
		}
		// verificar contraseña
		expected := hashEmpresaUsuarioPassword(payload.Password, admin.PasswordSalt)
		if expected != strings.TrimSpace(admin.PasswordHash) {
			writeAdminAuthError(w, http.StatusUnauthorized, "Credenciales inválidas.")
			return
		}
		// crear sesión
		token, terr := utils.GenerateSecureToken(32)
		if terr != nil {
			log.Println("AdminLoginHandler token gen error:", terr)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo crear la sesión administrativa.")
			return
		}
		if err := dbpkg.CreateSession(dbSuper, admin.Email, r.RemoteAddr, r.UserAgent(), token); err != nil {
			log.Println("AdminLoginHandler create session error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo guardar la sesión administrativa.")
			return
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

		redirectURL := "/seleccionar_empresa.html"
		if strings.ToLower(strings.TrimSpace(admin.Role)) == "super_administrador" {
			redirectURL = "/super_administrador.html"
		}
		writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "redirect_url": redirectURL})
	}
}

// AdminRequestPasswordRecoveryHandler solicita envío de token de recuperación.
func AdminRequestPasswordRecoveryHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeAdminAuthError(w, http.StatusMethodNotAllowed, "Método no permitido.")
			return
		}
		var payload struct {
			Email string `json:"email"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "La solicitud de recuperación es inválida.")
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		if payload.Email == "" {
			writeAdminAuthError(w, http.StatusBadRequest, "Debes indicar el correo de la cuenta.")
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, payload.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "email_sent": false, "message": "Si la cuenta existe y ya fue confirmada, enviaremos instrucciones para restablecer la contraseña."})
				return
			}
			log.Println("AdminRequestPasswordRecoveryHandler get admin error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo validar la cuenta administrativa.")
			return
		}
		if admin == nil || admin.EmailConfirmado != 1 {
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "email_sent": false, "message": "Si la cuenta existe y ya fue confirmada, enviaremos instrucciones para restablecer la contraseña."})
			return
		}
		// generar token
		token, expira, nerr := newPasswordRecoveryTokenAndExpiration()
		if nerr != nil {
			log.Println("AdminRequestPasswordRecoveryHandler token gen error:", nerr)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo generar el token de recuperación.")
			return
		}
		if err := dbpkg.SetAdministradorPasswordResetToken(dbSuper, payload.Email, token, expira); err != nil {
			log.Println("AdminRequestPasswordRecoveryHandler set token error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo iniciar la recuperación de contraseña.")
			return
		}
		if _, err := sendAdminPasswordRecoveryEmail(r, dbSuper, payload.Email, "", token); err != nil {
			log.Println("AdminRequestPasswordRecoveryHandler send mail error:", err)
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "email_sent": false, "message": "No se pudo enviar el correo de recuperación. Revisa la configuración SMTP.", "error": err.Error()})
			return
		}
		writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "email_sent": true, "message": "Si la cuenta existe y ya fue confirmada, enviaremos instrucciones para restablecer la contraseña."})
	}
}

// AdminResetPasswordHandler restablece contraseña usando token.
func AdminResetPasswordHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeAdminAuthError(w, http.StatusMethodNotAllowed, "Método no permitido.")
			return
		}
		var payload struct {
			Email    string `json:"email"`
			Token    string `json:"token"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "La solicitud para restablecer la contraseña es inválida.")
			return
		}
		payload.Email = strings.TrimSpace(payload.Email)
		payload.Token = strings.TrimSpace(payload.Token)
		payload.Password = strings.TrimSpace(payload.Password)
		if payload.Email == "" || payload.Token == "" || payload.Password == "" {
			writeAdminAuthError(w, http.StatusBadRequest, "Debes indicar correo, token y nueva contraseña.")
			return
		}
		if len(payload.Password) < minAdminPasswordLength {
			writeAdminAuthError(w, http.StatusBadRequest, "La nueva contraseña debe tener mínimo 8 caracteres.")
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, payload.Email)
		if err != nil {
			log.Println("AdminResetPasswordHandler get admin error:", err)
			writeAdminAuthError(w, http.StatusBadRequest, "El correo o el token de recuperación no son válidos.")
			return
		}
		if strings.TrimSpace(admin.PasswordResetToken) == "" || strings.TrimSpace(admin.PasswordResetToken) != payload.Token {
			writeAdminAuthError(w, http.StatusBadRequest, "El token de recuperación no es válido.")
			return
		}
		// verificar expiración
		if admin.PasswordResetExpira != "" {
			if t, ok := parseEmpresaUsuarioDateTime(admin.PasswordResetExpira); ok {
				if time.Now().After(t) {
					writeAdminAuthError(w, http.StatusBadRequest, "El token de recuperación ya expiró.")
					return
				}
			}
		}
		// generar hash y guardar
		hash, salt, err := generateEmpresaUsuarioPasswordHash(payload.Password)
		if err != nil {
			log.Println("AdminResetPasswordHandler hash error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo proteger la nueva contraseña.")
			return
		}
		if err := dbpkg.SetAdministradorPassword(dbSuper, payload.Email, hash, salt); err != nil {
			log.Println("AdminResetPasswordHandler set password error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo guardar la nueva contraseña.")
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
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo crear la nueva sesión administrativa.")
			return
		}
		if err := dbpkg.CreateSession(dbSuper, admin.Email, r.RemoteAddr, r.UserAgent(), tokenSession); err != nil {
			log.Println("AdminResetPasswordHandler create session error:", err)
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo guardar la nueva sesión administrativa.")
			return
		}
		cookie := &http.Cookie{
			Name:     "session_token",
			Value:    tokenSession,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400,
			Secure:   SessionCookieSecure(r),
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)
		SetBrowserSessionStateCookie(w, r, true)

		redirectURL := "/seleccionar_empresa.html"
		if strings.ToLower(strings.TrimSpace(admin.Role)) == "super_administrador" {
			redirectURL = "/super_administrador.html"
		}
		writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "redirect_url": redirectURL, "message": "Contraseña restablecida correctamente."})
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
	if err != nil {
		return "", err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return "", fmt.Errorf("gmail.smtp_email no configurado")
	}
	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return "", err
	}
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" {
		return "", fmt.Errorf("gmail.smtp_app_password no configurado")
	}

	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")
	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" {
		smtpPort = "587"
	}
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Powerful Control System"
	}

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
	resetHintURL := strings.TrimRight(baseURL, "/") + "/login.html?view=reset&email=" + url.QueryEscape(toEmail) + "&token_recuperacion=" + url.QueryEscape(token)
	safeName := strings.TrimSpace(toName)
	if safeName == "" {
		safeName = "administrador"
	}
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
	if err != nil {
		return "", err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return "", fmt.Errorf("gmail.smtp_email no configurado")
	}
	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return "", err
	}
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" {
		return "", fmt.Errorf("gmail.smtp_app_password no configurado")
	}
	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")
	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" {
		smtpPort = "587"
	}
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Powerful Control System"
	}
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

			admin, err := dbpkg.GetAdminByEmailFull(dbSuper, userinfo.Email)
			if err != nil || admin == nil {
				log.Println("warning: no admin found, redirecting to seleccionar_empresa:", err)
				http.Redirect(w, r, "/seleccionar_empresa.html", http.StatusFound)
				return
			}
			http.Redirect(w, r, resolveAdminPostLoginRedirect(admin), http.StatusFound)
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
