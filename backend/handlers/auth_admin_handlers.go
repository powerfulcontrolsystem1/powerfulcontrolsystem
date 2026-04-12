package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
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
		log.Printf("handleGoogleLogin: oauth redirect requested (client configured=%t)", clientID != "")
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
			// Forzar selección explícita de cuenta para evitar reutilizar sesión previa equivocada.
			"prompt": {"select_account consent"},
		}
		if loginHint != "" {
			vals.Set("login_hint", loginHint)
		}
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

		// La aceptación se decide únicamente por registro persistido por administrador.
		accepted := false
		if adminNow, err := dbpkg.GetAdminByEmail(dbSuper, userinfo.Email); err == nil && adminNow != nil {
			if adminNow.AceptaContrato == 1 {
				accepted = true
			}
		}

		if accepted {
			// Persistir marca de aceptación y crear sesión
			if err := dbpkg.SetAdministradorAceptaContrato(dbSuper, userinfo.Email, true); err != nil {
				log.Println("warning: failed to persist acepta_contrato:", err)
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
