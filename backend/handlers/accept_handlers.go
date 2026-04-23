package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// AcceptCompleteHandler procesa la aceptación del contrato desde la página /accept.html
// Recibe { payload: string, token: string } donde payload es un valor cifrado con utils.EncryptString
func AcceptCompleteHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Payload        string `json:"payload"`
			Token          string `json:"token"`
			RecaptchaToken string `json:"recaptcha_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(in.Payload) == "" {
			http.Error(w, "payload required", http.StatusBadRequest)
			return
		}

		dec, err := utils.DecryptString(in.Payload)
		if err != nil {
			log.Printf("accept: decrypt payload failed: %v", err)
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}
		var data struct {
			Email string `json:"email"`
			Exp   int64  `json:"exp"`
			Next  string `json:"next"`
		}
		if err := json.Unmarshal([]byte(dec), &data); err != nil {
			http.Error(w, "invalid token data", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(data.Email) == "" {
			http.Error(w, "invalid token data", http.StatusBadRequest)
			return
		}
		if time.Now().Unix() > data.Exp {
			http.Error(w, "token expired", http.StatusBadRequest)
			return
		}

		if err := dbpkg.EnsureSuperContractSchema(dbSuper); err != nil {
			log.Printf("accept: ensure contract schema failed: %v", err)
			http.Error(w, "contract unavailable", http.StatusInternalServerError)
			return
		}
		currentContract, err := dbpkg.GetCurrentSuperContract(dbSuper)
		if err != nil || currentContract == nil {
			log.Printf("accept: load current contract failed: %v", err)
			http.Error(w, "contract unavailable", http.StatusInternalServerError)
			return
		}

		captchaToken := strings.TrimSpace(in.RecaptchaToken)
		if captchaToken == "" {
			captchaToken = strings.TrimSpace(in.Token)
		}
		if err := validateRecaptchaToken(dbSuper, r, captchaToken); err != nil {
			writeRecaptchaValidationError(w, err)
			return
		}

		// Persistir aceptación y crear sesión
		if err := dbpkg.SetAdministradorContratoAceptado(dbSuper, data.Email, currentContract.Version); err != nil {
			log.Printf("warning: failed to persist acepta_contrato: %v", err)
		}
		token, err := utils.GenerateSecureToken(32)
		if err != nil {
			log.Printf("failed to generate session token: %v", err)
			token = data.Email
		}
		ip := r.RemoteAddr
		ua := r.UserAgent()
		if err := dbpkg.CreateSession(dbSuper, data.Email, ip, ua, token); err != nil {
			log.Printf("create session error: %v", err)
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
		// Limpiar cookie legacy para no reusar señal global de aceptación entre cuentas distintas.
		http.SetCookie(w, &http.Cookie{
			Name:     "accepted_contract",
			Value:    "",
			Path:     "/",
			HttpOnly: false,
			MaxAge:   -1,
			Secure:   SessionCookieSecure(r),
			SameSite: http.SameSiteLaxMode,
		})

		redirectTo := "/seleccionar_empresa.html"
		if dbSuper != nil {
			if admin, err := dbpkg.GetAdminByEmailFull(dbSuper, data.Email); err == nil && admin != nil {
				redirectTo = resolveAdminPostLoginRedirect(admin)
			} else {
				next := strings.TrimSpace(data.Next)
				if next == "/super_administrador.html" || next == "/seleccionar_empresa.html" {
					redirectTo = next
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "redirect": redirectTo, "contract_version": currentContract.Version})
	}
}
