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
			Payload string `json:"payload"`
			Token   string `json:"token"`
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

		// La validación reCAPTCHA se ha eliminado por decisión del producto.
		// Se omite cualquier token y se considera la verificación como satisfactoria.
		// Nota: el campo `token` del payload se ignora para mantener compatibilidad.

		// Persistir aceptación y crear sesión
		if err := dbpkg.SetAdministradorAceptaContrato(dbSuper, data.Email, true); err != nil {
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
			Secure:   (r.TLS != nil),
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)
		// Limpiar cookie legacy para no reusar señal global de aceptación entre cuentas distintas.
		http.SetCookie(w, &http.Cookie{
			Name:     "accepted_contract",
			Value:    "",
			Path:     "/",
			HttpOnly: false,
			MaxAge:   -1,
			Secure:   (r.TLS != nil),
			SameSite: http.SameSiteLaxMode,
		})

		redirectTo := "/seleccionar_empresa.html"
		next := strings.TrimSpace(data.Next)
		if next == "/super_administrador.html" || next == "/seleccionar_empresa.html" {
			redirectTo = next
		} else if dbSuper != nil {
			if admin, err := dbpkg.GetAdminByEmail(dbSuper, data.Email); err == nil && admin != nil {
				if strings.TrimSpace(admin.Role) == "super_administrador" {
					redirectTo = "/super_administrador.html"
				}
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "redirect": redirectTo})
	}
}
