package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/you/pos-backend/auth"
	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// HandleGoogleLogin devuelve un http.HandlerFunc configurado con clientID y redirectURL
func HandleGoogleLogin(clientID, redirectURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := "state-token"
		if clientID == "" {
			http.Error(w, "Acceso bloqueado: configuración incompleta (GOOGLE_CLIENT_ID no definido)", http.StatusInternalServerError)
			return
		}
		log.Printf("handleGoogleLogin: client_id set=%t, redirect_url=%q", clientID != "", redirectURL)
		q := r.URL.Query()
		loginHint := q.Get("login_hint")
		vals := url.Values{
			"client_id":              {clientID},
			"redirect_uri":           {redirectURL},
			"response_type":          {"code"},
			"scope":                  {"openid email profile"},
			"include_granted_scopes": {"true"},
			"access_type":            {"offline"},
			"state":                  {state},
		}
		if loginHint != "" {
			vals.Set("login_hint", loginHint)
			// Do not force account chooser when we have a login_hint; allow Google to select the hinted account if possible.
		} else {
			// Default: show account selector and consent to allow choosing among accounts.
			vals.Set("prompt", "select_account consent")
		}
		authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + vals.Encode()
		log.Printf("Auth URL: %s", authURL)
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

		tokenResp, err := auth.ExchangeCodeForToken(code, clientID, clientSecret, redirectURL)
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

		roleToSet := "administrador"
		if existingAdmin, err := dbpkg.GetAdminByEmail(dbSuper, userinfo.Email); err == nil && existingAdmin != nil {
			if existingAdmin.Role != "" {
				roleToSet = existingAdmin.Role
			}
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
			Secure:   (r.TLS != nil),
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)

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

// TiposLicenciasHandler placeholder (removed from UI)
func TiposLicenciasHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "tipos_de_licencia API removed", http.StatusNotFound)
	}
}

// LicenciasHandler maneja CRUD de licencias
func LicenciasHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			licencias, err := dbpkg.GetLicencias(dbSuper)
			if err != nil {
				log.Println("GET /super/api/licencias error:", err)
				http.Error(w, "failed to query licencias: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(licencias)
			return
		case http.MethodPost:
			var payload struct {
				TipoID       int64   `json:"tipo_id"`
				Nombre       string  `json:"nombre"`
				Descripcion  string  `json:"descripcion"`
				Valor        float64 `json:"valor"`
				DuracionDias int     `json:"duracion_dias"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			log.Printf("POST /super/api/licencias payload: TipoID=%d Nombre=%q", payload.TipoID, payload.Nombre)
			if payload.Nombre == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateLicencia(dbSuper, payload.TipoID, payload.Nombre, payload.Descripcion, payload.Valor, payload.DuracionDias)
			if err != nil {
				log.Println("POST /super/api/licencias error:", err)
				http.Error(w, "failed to create licencia: "+err.Error(), http.StatusInternalServerError)
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
			// soporte para acción de activar/desactivar vía query param
			if q.Get("action") == "activar" {
				activoStr := q.Get("activo")
				if activoStr == "" {
					http.Error(w, "activo required (0 or 1)", http.StatusBadRequest)
					return
				}
				act, err := strconv.Atoi(activoStr)
				if err != nil || (act != 0 && act != 1) {
					http.Error(w, "invalid activo value", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetLicenciaActivo(dbSuper, id, act); err != nil {
					log.Println("ACTIVAR /super/api/licencias error:", err)
					http.Error(w, "failed to set activo: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			// actualización normal (payload JSON)
			var payloadUpdate struct {
				TipoID       int64   `json:"tipo_id"`
				Nombre       string  `json:"nombre"`
				Descripcion  string  `json:"descripcion"`
				Valor        float64 `json:"valor"`
				DuracionDias int     `json:"duracion_dias"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payloadUpdate); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateLicencia(dbSuper, id, payloadUpdate.TipoID, payloadUpdate.Nombre, payloadUpdate.Descripcion, payloadUpdate.Valor, payloadUpdate.DuracionDias); err != nil {
				log.Println("PUT /super/api/licencias error:", err)
				http.Error(w, "failed to update licencia: "+err.Error(), http.StatusInternalServerError)
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
			if err := dbpkg.DeleteLicencia(dbSuper, id); err != nil {
				log.Println("DELETE /super/api/licencias error:", err)
				http.Error(w, "failed to delete licencia: "+err.Error(), http.StatusInternalServerError)
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

// MercadoPagoCreatePreferenceHandler crea una preferencia en Mercado Pago y devuelve la respuesta API
func MercadoPagoCreatePreferenceHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			LicenciaID int64  `json:"licencia_id"`
			EmpresaID  int64  `json:"empresa_id,omitempty"`
			PayerEmail string `json:"payer_email,omitempty"`
			PayerName  string `json:"payer_name,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		lic, err := dbpkg.GetLicenciaByID(dbSuper, payload.LicenciaID)
		if err != nil || lic == nil {
			http.Error(w, "licencia not found", http.StatusBadRequest)
			return
		}
		// Preferir token almacenado en DB (configuraciones), fallback a variable de entorno
		token := ""
		if v, enc, err := dbpkg.GetConfigValue(dbSuper, "mercadopago.access_token"); err == nil && v != "" {
			if enc {
				if dec, derr := utils.DecryptString(v); derr == nil {
					token = dec
				} else {
					log.Println("warning: failed to decrypt stored mercadopago.access_token:", derr)
				}
			} else {
				token = v
			}
		}
		if token == "" {
			token = os.Getenv("MERCADOPAGO_ACCESS_TOKEN")
		}
		if token == "" {
			http.Error(w, "MERCADOPAGO_ACCESS_TOKEN not configured", http.StatusInternalServerError)
			return
		}

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseURL := scheme + "://" + r.Host
		successURL := baseURL + "/pagar_licencia.html?status=success"
		failureURL := baseURL + "/pagar_licencia.html?status=failure"
		pendingURL := baseURL + "/pagar_licencia.html?status=pending"
		notificationURL := baseURL + "/mercadopago/webhook"
		// Si el administrador configuró una URL de webhook externa (ej. ngrok o dominio público), usarla
		if wurl, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "mercadopago.webhook_url"); wurl != "" {
			notificationURL = wurl
		}
		// Si el administrador configuró una URL de webhook externa (ej. ngrok o dominio público), usarla
		if wurl, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "mercadopago.webhook_url"); wurl != "" {
			notificationURL = wurl
		}

		reqBody := map[string]interface{}{
			"items": []map[string]interface{}{map[string]interface{}{
				"title":       lic.Nombre,
				"quantity":    1,
				"currency_id": "ARS",
				"unit_price":  lic.Valor,
			}},
			"back_urls":          map[string]string{"success": successURL, "failure": failureURL, "pending": pendingURL},
			"notification_url":   notificationURL,
			"external_reference": fmt.Sprintf("licencia_%d_empresa_%d", payload.LicenciaID, payload.EmpresaID),
		}
		// include payer prefill if provided
		if payload.PayerEmail != "" || payload.PayerName != "" {
			payer := map[string]string{}
			if payload.PayerEmail != "" {
				payer["email"] = payload.PayerEmail
			}
			if payload.PayerName != "" {
				payer["first_name"] = payload.PayerName
			}
			reqBody["payer"] = payer
		}
		bodyBytes, _ := json.Marshal(reqBody)
		log.Printf("MercadoPago request body (test_pref): %s", string(bodyBytes))
		apiURL := "https://api.mercadopago.com/checkout/preferences"
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			http.Error(w, "failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "request error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			log.Println("MercadoPago API error:", resp.Status, string(respBody))
			http.Error(w, "mercadopago API error: "+string(respBody), http.StatusInternalServerError)
			return
		}
		var mpResp map[string]interface{}
		if err := json.Unmarshal(respBody, &mpResp); err != nil {
			http.Error(w, "invalid response from mercadopago: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// record preference in DB if possible
		prefID := ""
		if idv, ok := mpResp["id"]; ok {
			prefID = fmt.Sprint(idv)
		}
		raw := string(respBody)
		if _, err := dbpkg.CreateMPPaymentRecord(dbSuper, payload.LicenciaID, payload.EmpresaID, prefID, "", "preference_created", raw); err != nil {
			log.Println("warning: failed to record MP preference in DB:", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mpResp)
	}
}

// MercadoPagoConfigHandler maneja la obtención y guardado de credenciales de Mercado Pago
func MercadoPagoConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			access, enc, _, accessAct, _ := dbpkg.GetConfigEntry(dbSuper, "mercadopago.access_token")
			publicKey, _, _, publicAct, _ := dbpkg.GetConfigEntry(dbSuper, "mercadopago.public_key")
			webhookURL, _, _, webhookURLUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "mercadopago.webhook_url")
			webhookSecret, secretEnc, _, webhookSecretUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "mercadopago.webhook_secret")
			accessSet := false
			accessMasked := ""
			if access != "" {
				accessSet = true
				if enc {
					// mostrar solo sufijo si está cifrado
					if len(access) > 8 {
						accessMasked = "****" + access[len(access)-4:]
					} else {
						accessMasked = "****"
					}
				} else {
					if len(access) > 8 {
						accessMasked = access[:4] + "****" + access[len(access)-4:]
					} else {
						accessMasked = "****"
					}
				}
			}
			pubMasked := ""
			pubSet := false
			if publicKey != "" {
				pubSet = true
				// mostrar fragmento de la clave pública (no sensible)
				if len(publicKey) > 24 {
					pubMasked = publicKey[:8] + "..." + publicKey[len(publicKey)-8:]
				} else {
					pubMasked = publicKey
				}
			}
			// webhook masked info
			webhookSet := false
			webhookSecretSet := false
			webhookSecretMasked := ""
			if webhookURL != "" {
				webhookSet = true
			}
			if webhookSecret != "" {
				webhookSecretSet = true
				if secretEnc {
					if len(webhookSecret) > 8 {
						webhookSecretMasked = "****" + webhookSecret[len(webhookSecret)-4:]
					} else {
						webhookSecretMasked = "****"
					}
				} else {
					if len(webhookSecret) > 8 {
						webhookSecretMasked = webhookSecret[:4] + "****" + webhookSecret[len(webhookSecret)-4:]
					} else {
						webhookSecretMasked = "****"
					}
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token_set":       accessSet,
				"access_token_masked":    accessMasked,
				"access_token_updated":   accessAct,
				"public_key_set":         pubSet,
				"public_key_masked":      pubMasked,
				"public_key_updated":     publicAct,
				"webhook_url_set":        webhookSet,
				"webhook_url":            webhookURL,
				"webhook_url_updated":    webhookURLUpdated,
				"webhook_secret_set":     webhookSecretSet,
				"webhook_secret_masked":  webhookSecretMasked,
				"webhook_secret_updated": webhookSecretUpdated,
				"encryption_available":   utils.EncryptionAvailable(),
			})
			return
		case http.MethodPost, http.MethodPut:
			var payload struct {
				AccessToken    string `json:"access_token"`
				PublicKey      string `json:"public_key"`
				Encrypt        bool   `json:"encrypt"`
				SkipValidation bool   `json:"skip_validation"`
				WebhookURL     string `json:"webhook_url"`
				WebhookSecret  string `json:"webhook_secret"`
				WebhookEncrypt bool   `json:"webhook_encrypt"`
			}
			// payload for config save
			// decode request body into payload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			// If an access token is provided, validate it against Mercado Pago before saving
			var validatedInfo map[string]interface{}
			if payload.AccessToken != "" {
				if payload.SkipValidation {
					log.Println("MercadoPagoConfigHandler: skipping validation for access token save (admin requested)")
					// mark as skipped so frontend can know
					validatedInfo = map[string]interface{}{"skipped": true}
				} else {
					// validate token by calling /v1/users/me
					client := &http.Client{Timeout: 10 * time.Second}
					req, _ := http.NewRequest("GET", "https://api.mercadopago.com/v1/users/me", nil)
					req.Header.Set("Authorization", "Bearer "+payload.AccessToken)
					resp, err := client.Do(req)
					if err != nil {
						log.Printf("MercadoPago validation request error: %v", err)
						http.Error(w, "validation request failed: "+err.Error(), http.StatusBadGateway)
						return
					}
					defer resp.Body.Close()
					rb, _ := io.ReadAll(resp.Body)
					if resp.StatusCode >= 400 {
						log.Printf("MercadoPago validation failed: status=%s body=%s", resp.Status, string(rb))
						http.Error(w, "validation failed: "+string(rb), http.StatusBadRequest)
						return
					}
					// parse minimal info
					var info map[string]interface{}
					if err := json.Unmarshal(rb, &info); err == nil {
						validatedInfo = info
					}
				}

				// save token after successful validation or when skipping validation
				if payload.Encrypt {
					if !utils.EncryptionAvailable() {
						http.Error(w, "encryption failed: CONFIG_ENC_KEY not set", http.StatusBadRequest)
						return
					}
					encVal, err := utils.EncryptString(payload.AccessToken)
					if err != nil {
						http.Error(w, "encryption failed: "+err.Error(), http.StatusInternalServerError)
						return
					}
					if err := dbpkg.SetConfigValue(dbSuper, "mercadopago.access_token", encVal, true); err != nil {
						http.Error(w, "failed to save access token: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else {
					if err := dbpkg.SetConfigValue(dbSuper, "mercadopago.access_token", payload.AccessToken, false); err != nil {
						http.Error(w, "failed to save access token: "+err.Error(), http.StatusInternalServerError)
						return
					}
					// no payer prefill handled here (belongs to create_preference)
				}
			}

			if payload.PublicKey != "" {
				if err := dbpkg.SetConfigValue(dbSuper, "mercadopago.public_key", payload.PublicKey, false); err != nil {
					http.Error(w, "failed to save public key: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// Webhook URL
			if payload.WebhookURL != "" {
				// validar formato básico de URL
				if u, err := url.ParseRequestURI(payload.WebhookURL); err != nil || u.Scheme == "" || u.Host == "" {
					http.Error(w, "invalid webhook_url: must be a valid absolute URL", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, "mercadopago.webhook_url", payload.WebhookURL, false); err != nil {
					http.Error(w, "failed to save webhook URL: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// Webhook secret
			if payload.WebhookSecret != "" {
				if payload.WebhookEncrypt {
					if !utils.EncryptionAvailable() {
						http.Error(w, "encryption failed: CONFIG_ENC_KEY not set", http.StatusBadRequest)
						return
					}
					encVal, err := utils.EncryptString(payload.WebhookSecret)
					if err != nil {
						http.Error(w, "encryption failed: "+err.Error(), http.StatusInternalServerError)
						return
					}
					if err := dbpkg.SetConfigValue(dbSuper, "mercadopago.webhook_secret", encVal, true); err != nil {
						http.Error(w, "failed to save webhook secret: "+err.Error(), http.StatusInternalServerError)
						return
					}
				} else {
					if err := dbpkg.SetConfigValue(dbSuper, "mercadopago.webhook_secret", payload.WebhookSecret, false); err != nil {
						http.Error(w, "failed to save webhook secret: "+err.Error(), http.StatusInternalServerError)
						return
					}
				}
			}

			// Return validation result (if any) but do not expose tokens
			w.Header().Set("Content-Type", "application/json")
			if validatedInfo != nil {
				// include only safe fields
				safe := map[string]interface{}{}
				if id, ok := validatedInfo["id"]; ok {
					safe["id"] = id
				}
				if nick, ok := validatedInfo["nickname"]; ok {
					safe["nickname"] = nick
				}
				if email, ok := validatedInfo["email"]; ok {
					safe["email"] = email
				}
				json.NewEncoder(w).Encode(map[string]interface{}{"validated": true, "account": safe})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"validated": false})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// MercadoPagoTestPreferenceHandler crea una preferencia de prueba para verificar que el checkout funciona
func MercadoPagoTestPreferenceHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			Amount float64 `json:"amount"`
			Title  string  `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			// allow empty body
			payload.Amount = 1
			payload.Title = "Prueba de pago"
		}
		if payload.Amount <= 0 {
			payload.Amount = 1
		}
		if payload.Title == "" {
			payload.Title = "Prueba de pago"
		}

		// obtener token de DB o env
		token := ""
		if v, enc, err := dbpkg.GetConfigValue(dbSuper, "mercadopago.access_token"); err == nil && v != "" {
			if enc {
				if dec, derr := utils.DecryptString(v); derr == nil {
					token = dec
				} else {
					log.Println("warning: failed to decrypt stored mercadopago.access_token:", derr)
				}
			} else {
				token = v
			}
		}
		if token == "" {
			token = os.Getenv("MERCADOPAGO_ACCESS_TOKEN")
		}
		if token == "" {
			http.Error(w, "MERCADOPAGO_ACCESS_TOKEN not configured", http.StatusInternalServerError)
			return
		}

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseURL := scheme + "://" + r.Host
		successURL := baseURL + "/pagar_licencia.html?status=success"
		failureURL := baseURL + "/pagar_licencia.html?status=failure"
		pendingURL := baseURL + "/pagar_licencia.html?status=pending"
		notificationURL := baseURL + "/mercadopago/webhook"

		reqBody := map[string]interface{}{
			"items":              []map[string]interface{}{{"title": payload.Title, "quantity": 1, "currency_id": "ARS", "unit_price": payload.Amount}},
			"back_urls":          map[string]string{"success": successURL, "failure": failureURL, "pending": pendingURL},
			"notification_url":   notificationURL,
			"external_reference": fmt.Sprintf("test_config_%d", time.Now().Unix()),
		}
		bodyBytes, _ := json.Marshal(reqBody)
		log.Printf("MercadoPago request body (test_pref): %s", string(bodyBytes))
		apiURL := "https://api.mercadopago.com/checkout/preferences"
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			http.Error(w, "failed to create request: "+err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "request error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 400 {
			log.Println("MercadoPago API error test_pref:", resp.Status, string(respBody))
			http.Error(w, "mercadopago API error: "+string(respBody), http.StatusInternalServerError)
			return
		}
		var mpResp map[string]interface{}
		if err := json.Unmarshal(respBody, &mpResp); err != nil {
			http.Error(w, "invalid response from mercadopago: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// record preference
		prefID := ""
		if idv, ok := mpResp["id"]; ok {
			prefID = fmt.Sprint(idv)
		}
		if _, err := dbpkg.CreateMPPaymentRecord(dbSuper, 0, 0, prefID, "", "test_preference", string(respBody)); err != nil {
			log.Println("warning: failed to record test pref in DB:", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mpResp)
	}
}

// MercadoPagoWebhookHandler recibe notificaciones de Mercado Pago
func MercadoPagoWebhookHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusBadRequest)
			return
		}
		var obj map[string]interface{}
		_ = json.Unmarshal(body, &obj)

		// Verificación opcional de firma del webhook si el administrador configuró un secret
		if sVal, enc, err := dbpkg.GetConfigValue(dbSuper, "mercadopago.webhook_secret"); err == nil && sVal != "" {
			secret := sVal
			if enc {
				if dec, derr := utils.DecryptString(sVal); derr == nil {
					secret = dec
				} else {
					log.Println("warning: failed to decrypt webhook_secret:", derr)
				}
			}
			if secret != "" {
				// buscar encabezado de firma en varios nombres posibles
				sigHeaderKeys := []string{"X-Hub-Signature", "X-Mercadopago-Signature", "X-Meli-Signature", "X-Hub-Signature-256", "X-Hub-Signature-sha256"}
				var sigHeader string
				for _, hk := range sigHeaderKeys {
					if v := r.Header.Get(hk); v != "" {
						sigHeader = v
						break
					}
				}
				if sigHeader != "" {
					sig := sigHeader
					if strings.HasPrefix(strings.ToLower(sig), "sha256=") {
						sig = strings.SplitN(sig, "=", 2)[1]
					}
					mac := hmac.New(sha256.New, []byte(secret))
					mac.Write(body)
					expectedHex := hex.EncodeToString(mac.Sum(nil))
					// comparar hex (insensible a mayúsculas)
					if subtle.ConstantTimeCompare([]byte(strings.ToLower(sig)), []byte(strings.ToLower(expectedHex))) != 1 {
						// intentar también comparar con base64
						expectedB64 := base64.StdEncoding.EncodeToString(mac.Sum(nil))
						if subtle.ConstantTimeCompare([]byte(sig), []byte(expectedB64)) != 1 {
							log.Printf("mercadopago webhook signature mismatch header=%s expectedHex=%s expectedB64=%s", sigHeader, expectedHex, expectedB64)
							http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
							return
						}
						// else ok
					}
				} else {
					log.Println("webhook secret configured but no signature header present; rejecting webhook")
					http.Error(w, "missing signature header", http.StatusUnauthorized)
					return
				}
			}
		}

		// Determine payment id and preference id from payload
		var paymentID, prefID, status string
		if d, ok := obj["data"]; ok {
			if m, ok2 := d.(map[string]interface{}); ok2 {
				if idv, ok3 := m["id"]; ok3 {
					paymentID = fmt.Sprint(idv)
				} else if resource, ok4 := m["resource"]; ok4 {
					if rmap, ok5 := resource.(map[string]interface{}); ok5 {
						if idv2, ok6 := rmap["id"]; ok6 {
							paymentID = fmt.Sprint(idv2)
						}
						if p, ok7 := rmap["preference_id"]; ok7 {
							prefID = fmt.Sprint(p)
						}
						if ext, ok8 := rmap["external_reference"]; ok8 {
							if s, ok9 := ext.(string); ok9 && prefID == "" {
								// external_reference may contain licencia/empresa
								prefID = ""
								_ = s
							}
						}
					}
				}
			}
		}
		// top-level id sometimes references the preference
		if idv, ok := obj["id"]; ok && prefID == "" {
			prefID = fmt.Sprint(idv)
		}

		// If we found a payment id, fetch payment details from Mercado Pago to obtain status/preference
		// Preferir token almacenado en DB (configuraciones), fallback a variable de entorno
		token := ""
		if v, enc, err := dbpkg.GetConfigValue(dbSuper, "mercadopago.access_token"); err == nil && v != "" {
			if enc {
				if dec, derr := utils.DecryptString(v); derr == nil {
					token = dec
				} else {
					log.Println("warning: failed to decrypt stored mercadopago.access_token:", derr)
				}
			} else {
				token = v
			}
		}
		if token == "" {
			token = os.Getenv("MERCADOPAGO_ACCESS_TOKEN")
		}
		var payResp map[string]interface{}
		if paymentID != "" && token != "" {
			client := &http.Client{Timeout: 15 * time.Second}
			req, _ := http.NewRequest("GET", "https://api.mercadopago.com/v1/payments/"+paymentID, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := client.Do(req)
			if err == nil {
				defer resp.Body.Close()
				rb, _ := io.ReadAll(resp.Body)
				if resp.StatusCode < 400 {
					json.Unmarshal(rb, &payResp)
					if s, ok := payResp["status"]; ok {
						status = fmt.Sprint(s)
					}
					if p, ok := payResp["preference_id"]; ok {
						prefID = fmt.Sprint(p)
					}
					if ext, ok := payResp["external_reference"]; ok && prefID == "" {
						if s, ok2 := ext.(string); ok2 {
							// try to extract licencia/empresa from external_reference later
							_ = s
						}
					}
				} else {
					log.Println("MercadoPago returned error fetching payment:", resp.Status)
				}
			} else {
				log.Println("error fetching mercadopago payment:", err)
			}
		}

		// Try to resolve licencia_id and empresa_id: first by matching preference_id in our pagos table
		var licenciaID sql.NullInt64
		var empresaID sql.NullInt64
		if prefID != "" {
			row := dbSuper.QueryRow("SELECT licencia_id, empresa_id FROM pagos_mercadopago WHERE preference_id = ? LIMIT 1", prefID)
			if err := row.Scan(&licenciaID, &empresaID); err != nil {
				log.Println("no pagos_mercadopago record for preference:", prefID, "err:", err)
			}
		}

		// If not found, try to parse external_reference from payment response (format: licencia_<id>_empresa_<id>)
		if !licenciaID.Valid && payResp != nil {
			if ext, ok := payResp["external_reference"]; ok {
				if s, ok2 := ext.(string); ok2 {
					re := regexp.MustCompile(`licencia_(\d+)_empresa_(\d+)`)
					if m := re.FindStringSubmatch(s); len(m) == 3 {
						if lid, err := strconv.ParseInt(m[1], 10, 64); err == nil {
							licenciaID = sql.NullInt64{Int64: lid, Valid: true}
						}
						if eid, err := strconv.ParseInt(m[2], 10, 64); err == nil {
							empresaID = sql.NullInt64{Int64: eid, Valid: true}
						}
					}
				}
			}
		}

		// Update existing payment record (if we have prefID) or insert fallback
		if prefID != "" {
			if err := dbpkg.UpdateMPPaymentRecordByPreference(dbSuper, prefID, paymentID, status, string(body)); err != nil {
				log.Println("warning: failed to update MP payment record:", err)
			}
		} else {
			if _, err := dbSuper.Exec("INSERT INTO pagos_mercadopago (preference_id, payment_id, status, raw_payload, fecha_creacion) VALUES (?, ?, ?, ?, datetime('now','localtime'))", prefID, paymentID, status, string(body)); err != nil {
				log.Println("failed to persist mercadopago webhook:", err)
			}
		}

		// If payment approved and we have licencia & empresa, activate license
		if strings.ToLower(status) == "approved" && licenciaID.Valid && empresaID.Valid {
			lic, err := dbpkg.GetLicenciaByID(dbSuper, licenciaID.Int64)
			if err == nil && lic != nil {
				now := time.Now()
				fechaInicio := now.Format("2006-01-02 15:04:05")
				fechaFin := now.AddDate(0, 0, lic.DuracionDias).Format("2006-01-02 15:04:05")
				if err := dbpkg.ActivateLicenciaForEmpresa(dbSuper, licenciaID.Int64, empresaID.Int64, fechaInicio, fechaFin); err != nil {
					log.Println("failed to activate licencia:", err)
				} else {
					log.Println("Licencia activated:", licenciaID.Int64, "empresa:", empresaID.Int64)
				}
			} else {
				log.Println("failed to fetch licencia for activation:", err)
			}
		} else {
			log.Println("Payment status:", status, "— no activation (licenciaValid:", licenciaID.Valid, " empresaValid:", empresaID.Valid, ")")
		}

		w.WriteHeader(http.StatusOK)
	}
}

// MeHandler devuelve información del administrador autenticado usando la cookie session_token
func MeHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil || c == nil || c.Value == "" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		s, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
		if err != nil || s == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		admin, err := dbpkg.GetAdminByEmail(dbSuper, s.AdminEmail)
		if err != nil || admin == nil {
			http.Error(w, "no admin found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(admin)
	}
}

// SecurityPortsHandler intenta conexiones TCP a una lista de puertos y devuelve su estado.
func SecurityPortsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		ip := q.Get("ip")
		if ip == "" {
			ip = "127.0.0.1"
		}
		portsParam := q.Get("ports")
		var ports []int
		if portsParam == "" {
			ports = []int{22, 23, 80, 443, 3306, 5432, 8080, 8443}
		} else {
			for _, s := range strings.Split(portsParam, ",") {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				if n, err := strconv.Atoi(s); err == nil {
					ports = append(ports, n)
				}
			}
		}
		timeout := 500 * time.Millisecond
		if tms := q.Get("timeout_ms"); tms != "" {
			if ms, err := strconv.Atoi(tms); err == nil && ms > 0 {
				timeout = time.Duration(ms) * time.Millisecond
			}
		}

		type Entry struct {
			Puerto   int    `json:"puerto"`
			Estado   string `json:"estado"`
			IP       string `json:"ip"`
			Firewall string `json:"firewall"`
		}
		var resp []Entry
		for _, p := range ports {
			addr := fmt.Sprintf("%s:%d", ip, p)
			conn, err := net.DialTimeout("tcp", addr, timeout)
			estado := "cerrado"
			if err == nil {
				estado = "abierto"
				conn.Close()
			}
			resp = append(resp, Entry{Puerto: p, Estado: estado, IP: ip, Firewall: "Desconocido"})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// EmpresasHandler maneja CRUD de empresas en la base empresas.db
func EmpresasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresas, err := dbpkg.GetEmpresas(dbEmp)
			if err != nil {
				log.Println("GET /super/api/empresas error:", err)
				http.Error(w, "failed to query empresas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(empresas)
			return
		case http.MethodPost:
			var payload struct {
				TipoID         int64  `json:"tipo_id"`
				TipoNombre     string `json:"tipo_nombre"`
				Nombre         string `json:"nombre"`
				Nit            string `json:"nit"`
				Observaciones  string `json:"observaciones"`
				UsuarioCreador string `json:"usuario_creador"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.Nombre == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateEmpresa(dbEmp, payload.TipoID, payload.TipoNombre, payload.Nombre, payload.Nit, payload.Observaciones, payload.UsuarioCreador)
			if err != nil {
				log.Println("POST /super/api/empresas error:", err)
				http.Error(w, "failed to create empresa: "+err.Error(), http.StatusInternalServerError)
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
			if q.Get("action") == "activar" {
				activoStr := q.Get("activo")
				if activoStr == "" {
					http.Error(w, "activo required (0 or 1)", http.StatusBadRequest)
					return
				}
				act, err := strconv.Atoi(activoStr)
				if err != nil || (act != 0 && act != 1) {
					http.Error(w, "invalid activo value", http.StatusBadRequest)
					return
				}
				estado := "inactivo"
				if act == 1 {
					estado = "activo"
				}
				if err := dbpkg.SetEmpresaEstado(dbEmp, id, estado); err != nil {
					log.Println("ACTIVAR /super/api/empresas error:", err)
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var payloadUpdate struct {
				TipoID        int64  `json:"tipo_id"`
				TipoNombre    string `json:"tipo_nombre"`
				Nombre        string `json:"nombre"`
				Nit           string `json:"nit"`
				Observaciones string `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payloadUpdate); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateEmpresa(dbEmp, id, payloadUpdate.TipoID, payloadUpdate.TipoNombre, payloadUpdate.Nombre, payloadUpdate.Nit, payloadUpdate.Observaciones); err != nil {
				log.Println("PUT /super/api/empresas error:", err)
				http.Error(w, "failed to update empresa: "+err.Error(), http.StatusInternalServerError)
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
			if err := dbpkg.DeleteEmpresa(dbEmp, id); err != nil {
				log.Println("DELETE /super/api/empresas error:", err)
				http.Error(w, "failed to delete empresa: "+err.Error(), http.StatusInternalServerError)
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
