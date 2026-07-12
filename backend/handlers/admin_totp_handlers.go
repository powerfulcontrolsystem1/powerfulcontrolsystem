package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const adminTOTPIssuer = "Powerful Control System"
const superAdmin2FAEnabledConfigKey = "security.admin_2fa.enabled"

func generateAdminTOTPSecret() (string, error) {
	raw := make([]byte, 20)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return strings.TrimRight(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw), "="), nil
}

func normalizeTOTPCode(code string) string {
	var b strings.Builder
	for _, ch := range strings.TrimSpace(code) {
		if ch >= '0' && ch <= '9' {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func totpCodeAt(secret string, counter int64) (string, error) {
	if counter < 0 {
		return "", fmt.Errorf("contador TOTP invalido")
	}
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", err
	}
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(counter))
	mac := hmac.New(sha1.New, key) // #nosec G505 -- RFC 6238 interoperable TOTP uses HMAC-SHA-1; the secret is independently encrypted at rest.
	_, _ = mac.Write(msg[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	bin := (int(sum[offset])&0x7f)<<24 | (int(sum[offset+1])&0xff)<<16 | (int(sum[offset+2])&0xff)<<8 | (int(sum[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", bin%1000000), nil
}

func verifyAdminTOTPCode(secret, code string, now time.Time) bool {
	_, ok := matchingAdminTOTPCounter(secret, code, now)
	return ok
}

func matchingAdminTOTPCounter(secret, code string, now time.Time) (int64, bool) {
	code = normalizeTOTPCode(code)
	secret = strings.TrimSpace(secret)
	if secret == "" || len(code) != 6 {
		return 0, false
	}
	counter := now.Unix() / 30
	for _, drift := range []int64{-1, 0, 1} {
		expected, err := totpCodeAt(secret, counter+drift)
		if err != nil {
			return 0, false
		}
		if hmac.Equal([]byte(expected), []byte(code)) {
			return counter + drift, true
		}
	}
	return 0, false
}

func adminTOTPProvisioningURI(email, secret string) string {
	label := url.QueryEscape(adminTOTPIssuer + ":" + strings.ToLower(strings.TrimSpace(email)))
	q := url.Values{}
	q.Set("secret", strings.TrimSpace(secret))
	q.Set("issuer", adminTOTPIssuer)
	q.Set("digits", "6")
	q.Set("period", "30")
	return "otpauth://totp/" + label + "?" + q.Encode()
}

func isAdminTOTPLoginEnabled(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return false
	}
	value, err := getDecryptedConfigValue(dbSuper, superAdmin2FAEnabledConfigKey)
	if err != nil {
		return false
	}
	return parseTruthyConfigValue(value, false)
}

func adminTOTPLoginRequiredForAdmin(admin *dbpkg.Admin, globalEnabled bool) bool {
	return globalEnabled && admin != nil && admin.TOTPEnabled == 1 && strings.TrimSpace(admin.TOTPSecret) != ""
}

func adminTOTPSecretForVerification(admin *dbpkg.Admin) (string, error) {
	if admin == nil || strings.TrimSpace(admin.TOTPSecret) == "" {
		return "", fmt.Errorf("TOTP not configured")
	}
	secret, err := dbpkg.DecryptAdministradorTOTPSecret(admin.TOTPSecret)
	if err == nil {
		return secret, nil
	}
	// Legacy plaintext is only tolerated while the startup migration has not
	// yet run. New writes always use the purpose-bound encrypted envelope.
	if !strings.HasPrefix(strings.TrimSpace(admin.TOTPSecret), "v1:") {
		return admin.TOTPSecret, nil
	}
	return "", err
}

func generateAdminTOTPRecoveryCodes(count int) ([]string, string, error) {
	if count <= 0 {
		return nil, "", fmt.Errorf("recovery code count required")
	}
	batchRaw := make([]byte, 16)
	if _, err := rand.Read(batchRaw); err != nil {
		return nil, "", err
	}
	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			return nil, "", err
		}
		codes = append(codes, base64.RawURLEncoding.EncodeToString(raw))
	}
	return codes, base64.RawURLEncoding.EncodeToString(batchRaw), nil
}

func verifyAndConsumeAdminTOTP(dbSuper *sql.DB, admin *dbpkg.Admin, code string, now time.Time, allowRecovery bool) bool {
	secret, err := adminTOTPSecretForVerification(admin)
	if err == nil {
		if counter, ok := matchingAdminTOTPCounter(secret, code, now); ok {
			consumed, consumeErr := dbpkg.ConsumeAdministradorTOTPCounter(dbSuper, admin.Email, counter)
			return consumeErr == nil && consumed
		}
	}
	if !allowRecovery {
		return false
	}
	consumed, consumeErr := dbpkg.ConsumeAdministradorTOTPRecoveryCode(dbSuper, admin.Email, strings.TrimSpace(code))
	return consumeErr == nil && consumed
}

func adminPasswordReauthenticated(admin *dbpkg.Admin, password string) bool {
	if admin == nil || strings.TrimSpace(password) == "" || strings.TrimSpace(admin.PasswordHash) == "" || strings.TrimSpace(admin.PasswordSalt) == "" {
		return false
	}
	expected := hashEmpresaUsuarioPassword(password, admin.PasswordSalt)
	stored := strings.TrimSpace(admin.PasswordHash)
	return len(expected) == len(stored) && hmac.Equal([]byte(expected), []byte(stored))
}

func issueReplacementAdminSession(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB, adminEmail string) error {
	if err := dbpkg.RevokeSessionsByAdminEmail(dbSuper, adminEmail); err != nil {
		return err
	}
	utils.InvalidateAuthCacheForAdmin(adminEmail)
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		return err
	}
	if err := dbpkg.CreateSession(dbSuper, adminEmail, r.RemoteAddr, r.UserAgent(), token); err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: token, Path: "/", HttpOnly: true, MaxAge: 86400, Secure: SessionCookieSecure(r), SameSite: http.SameSiteLaxMode})
	SetBrowserSessionStateCookie(w, r, true)
	return nil
}

func AdminTwoFactorGlobalConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if r.Method == http.MethodGet {
			raw, _, _, updatedAt, _ := dbpkg.GetConfigEntry(dbSuper, superAdmin2FAEnabledConfigKey)
			updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, superAdmin2FAEnabledConfigKey+".updated_by")
			enabled := parseTruthyConfigValue(raw, false)
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
				"ok":          true,
				"enabled":     enabled,
				"updated_at":  updatedAt,
				"updated_by":  strings.TrimSpace(updatedBy),
				"config_key":  superAdmin2FAEnabledConfigKey,
				"description": "Gobierna si el login de administradores muestra y exige codigo 2FA cuando la cuenta tiene TOTP activo.",
			})
			return
		}

		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			writeAdminAuthError(w, http.StatusMethodNotAllowed, "Metodo no permitido.")
			return
		}
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
		if adminEmail == "" || adminEmail == "sistema" {
			writeAdminAuthError(w, http.StatusUnauthorized, "Sesion no autenticada.")
			return
		}
		var payload struct {
			Enabled *bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "Solicitud de configuracion 2FA invalida.")
			return
		}
		if payload.Enabled == nil {
			writeAdminAuthError(w, http.StatusBadRequest, "Debes enviar enabled.")
			return
		}
		value := "0"
		if *payload.Enabled {
			value = "1"
		}
		if err := dbpkg.SetConfigValue(dbSuper, superAdmin2FAEnabledConfigKey, value, false); err != nil {
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo guardar la configuracion global 2FA.")
			return
		}
		_ = dbpkg.SetConfigValue(dbSuper, superAdmin2FAEnabledConfigKey+".updated_by", adminEmail, false)
		writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
			"ok":      true,
			"enabled": *payload.Enabled,
		})
	}
}

func AdminTwoFactorHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, adminEmail)
		if err != nil {
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo cargar el estado 2FA.")
			return
		}

		if r.Method == http.MethodGet {
			globalEnabled := isAdminTOTPLoginEnabled(dbSuper)
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                 true,
				"enabled":            admin.TOTPEnabled == 1,
				"configured":         strings.TrimSpace(admin.TOTPSecret) != "",
				"confirmed_at":       admin.TOTPConfirmadoEn,
				"issuer":             adminTOTPIssuer,
				"global_enabled":     globalEnabled,
				"login_requires_otp": adminTOTPLoginRequiredForAdmin(admin, globalEnabled),
			})
			return
		}
		if r.Method != http.MethodPost {
			writeAdminAuthError(w, http.StatusMethodNotAllowed, "Metodo no permitido.")
			return
		}

		var payload struct {
			Action          string `json:"action"`
			Code            string `json:"code"`
			CurrentPassword string `json:"current_password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "Solicitud 2FA invalida.")
			return
		}
		switch strings.ToLower(strings.TrimSpace(payload.Action)) {
		case "setup":
			if !adminPasswordReauthenticated(admin, payload.CurrentPassword) {
				writeAdminAuthError(w, http.StatusUnauthorized, "Debes confirmar tu contraseña actual para configurar 2FA.")
				return
			}
			secret, err := generateAdminTOTPSecret()
			if err != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo generar el secreto 2FA.")
				return
			}
			if err := dbpkg.SetAdministradorTOTPSecret(dbSuper, adminEmail, secret); err != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo guardar el secreto 2FA.")
				return
			}
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
				"ok":               true,
				"secret":           secret,
				"provisioning_uri": adminTOTPProvisioningURI(adminEmail, secret),
				"digits":           6,
				"period":           30,
			})
		case "confirm":
			if !adminPasswordReauthenticated(admin, payload.CurrentPassword) {
				writeAdminAuthError(w, http.StatusUnauthorized, "Debes confirmar tu contraseña actual para activar 2FA.")
				return
			}
			if !verifyAndConsumeAdminTOTP(dbSuper, admin, payload.Code, time.Now(), false) {
				writeAdminAuthError(w, http.StatusUnauthorized, "Codigo 2FA invalido.")
				return
			}
			if err := dbpkg.EnableAdministradorTOTP(dbSuper, adminEmail); err != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo activar 2FA.")
				return
			}
			codes, batchID, err := generateAdminTOTPRecoveryCodes(10)
			if err != nil || dbpkg.ReplaceAdministradorTOTPRecoveryCodes(dbSuper, adminEmail, batchID, codes) != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudieron crear los codigos de recuperacion.")
				return
			}
			if err := issueReplacementAdminSession(w, r, dbSuper, adminEmail); err != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo rotar la sesión después de activar 2FA.")
				return
			}
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "enabled": true, "recovery_codes": codes})
		case "disable":
			if !adminPasswordReauthenticated(admin, payload.CurrentPassword) {
				writeAdminAuthError(w, http.StatusUnauthorized, "Debes confirmar tu contraseña actual para desactivar 2FA.")
				return
			}
			if admin.TOTPEnabled == 1 && !verifyAndConsumeAdminTOTP(dbSuper, admin, payload.Code, time.Now(), true) {
				writeAdminAuthError(w, http.StatusUnauthorized, "Codigo 2FA invalido.")
				return
			}
			if err := dbpkg.DisableAdministradorTOTP(dbSuper, adminEmail); err != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo desactivar 2FA.")
				return
			}
			if err := issueReplacementAdminSession(w, r, dbSuper, adminEmail); err != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo rotar la sesión después de desactivar 2FA.")
				return
			}
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "enabled": false})
		case "regenerate_recovery_codes":
			if admin.TOTPEnabled != 1 || !adminPasswordReauthenticated(admin, payload.CurrentPassword) || !verifyAndConsumeAdminTOTP(dbSuper, admin, payload.Code, time.Now(), true) {
				writeAdminAuthError(w, http.StatusUnauthorized, "No se pudo confirmar la regeneración de códigos.")
				return
			}
			codes, batchID, err := generateAdminTOTPRecoveryCodes(10)
			if err != nil || dbpkg.ReplaceAdministradorTOTPRecoveryCodes(dbSuper, adminEmail, batchID, codes) != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudieron regenerar los códigos de recuperación.")
				return
			}
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "recovery_codes": codes})
		case "verify":
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": verifyAndConsumeAdminTOTP(dbSuper, admin, payload.Code, time.Now(), true)})
		default:
			writeAdminAuthError(w, http.StatusBadRequest, "Accion 2FA no soportada: "+strconv.Quote(payload.Action))
		}
	}
}
