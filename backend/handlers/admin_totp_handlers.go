package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base32"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const adminTOTPIssuer = "Powerful Control System"

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
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil {
		return "", err
	}
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(counter))
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(msg[:])
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	bin := (int(sum[offset])&0x7f)<<24 | (int(sum[offset+1])&0xff)<<16 | (int(sum[offset+2])&0xff)<<8 | (int(sum[offset+3]) & 0xff)
	return fmt.Sprintf("%06d", bin%1000000), nil
}

func verifyAdminTOTPCode(secret, code string, now time.Time) bool {
	code = normalizeTOTPCode(code)
	secret = strings.TrimSpace(secret)
	if secret == "" || len(code) != 6 {
		return false
	}
	counter := now.Unix() / 30
	for _, drift := range []int64{-1, 0, 1} {
		expected, err := totpCodeAt(secret, counter+drift)
		if err != nil {
			return false
		}
		if hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}
	return false
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
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                 true,
				"enabled":            admin.TOTPEnabled == 1,
				"configured":         strings.TrimSpace(admin.TOTPSecret) != "",
				"confirmed_at":       admin.TOTPConfirmadoEn,
				"issuer":             adminTOTPIssuer,
				"login_requires_otp": admin.TOTPEnabled == 1,
			})
			return
		}
		if r.Method != http.MethodPost {
			writeAdminAuthError(w, http.StatusMethodNotAllowed, "Metodo no permitido.")
			return
		}

		var payload struct {
			Action string `json:"action"`
			Code   string `json:"code"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "Solicitud 2FA invalida.")
			return
		}
		switch strings.ToLower(strings.TrimSpace(payload.Action)) {
		case "setup":
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
			if !verifyAdminTOTPCode(admin.TOTPSecret, payload.Code, time.Now()) {
				writeAdminAuthError(w, http.StatusUnauthorized, "Codigo 2FA invalido.")
				return
			}
			if err := dbpkg.EnableAdministradorTOTP(dbSuper, adminEmail); err != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo activar 2FA.")
				return
			}
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "enabled": true})
		case "disable":
			if admin.TOTPEnabled == 1 && !verifyAdminTOTPCode(admin.TOTPSecret, payload.Code, time.Now()) {
				writeAdminAuthError(w, http.StatusUnauthorized, "Codigo 2FA invalido.")
				return
			}
			if err := dbpkg.DisableAdministradorTOTP(dbSuper, adminEmail); err != nil {
				writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo desactivar 2FA.")
				return
			}
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "enabled": false})
		case "verify":
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{"ok": verifyAdminTOTPCode(admin.TOTPSecret, payload.Code, time.Now())})
		default:
			writeAdminAuthError(w, http.StatusBadRequest, "Accion 2FA no soportada: "+strconv.Quote(payload.Action))
		}
	}
}
