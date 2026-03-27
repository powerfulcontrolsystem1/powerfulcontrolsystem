package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/url"

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
		authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + url.Values{
			"client_id":              {clientID},
			"redirect_uri":           {redirectURL},
			"response_type":          {"code"},
			"scope":                  {"openid email profile"},
			"prompt":                 {"select_account consent"},
			"include_granted_scopes": {"true"},
			"access_type":            {"offline"},
			"state":                  {state},
		}.Encode()

		// Loguear la URL de autorización (sin exponer secretos) para depuración
		log.Printf("Auth URL: %s", authURL)
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// HandleGoogleCallback devuelve un http.HandlerFunc que procesa el callback y persiste el usuario
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

		// Registrar/actualizar en la base de datos de superadministrador (tabla administradores)
		// Mantener el role existente si ya está presente (no sobrescribir super_administrador)
		roleToSet := "administrador"
		if existingAdmin, err := dbpkg.GetAdminByEmail(dbSuper, userinfo.Email); err == nil && existingAdmin != nil {
			if existingAdmin.Role != "" {
				roleToSet = existingAdmin.Role
			}
		}
		if err := dbpkg.UpsertAdministrador(dbSuper, userinfo.Email, userinfo.Name, roleToSet); err != nil {
			log.Println("db upsert administradores error:", err)
		}

		// Registrar/actualizar también en la base de empresas (tabla users) por compatibilidad
		if err := dbpkg.UpsertUser(dbEmpresas, userinfo.Email, userinfo.Name); err != nil {
			log.Println("db upsert users error:", err)
		}

		// Asegurar que el usuario tenga una empresa por defecto asociada en la DB de empresas
		if err := dbpkg.EnsureUserEmpresa(dbEmpresas, userinfo.Email, "Empresa de "+userinfo.Name); err != nil {
			log.Println("db ensure empresa error:", err)
		}

		// Registrar sesión en la base superadministrador (usar token seguro)
		token, err := utils.GenerateSecureToken(32)
		if err != nil {
			log.Println("failed to generate session token:", err)
			token = userinfo.Sub // fallback
		}
		ip := r.RemoteAddr
		ua := r.UserAgent()
		if err := dbpkg.CreateSession(dbSuper, userinfo.Email, ip, ua, token); err != nil {
			log.Println("create session error:", err)
		}
		// Establecer cookie de sesión segura (httpOnly)
		cookie := &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400,
			Secure:   (r.TLS != nil), // habilitar Secure solo si la conexión es TLS
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)

		// Obtener rol del administrador y redirigir según rol
		admin, err := dbpkg.GetAdminByEmail(dbSuper, userinfo.Email)
		if err != nil {
			// Si no se encuentra o hay error, redirigir a seleccionar_empresa por defecto
			log.Println("warning: no se pudo obtener admin para redireccion, usando seleccionar_empresa:", err)
			http.Redirect(w, r, "/seleccionar_empresa.html", http.StatusFound)
			return
		}
		log.Printf("post-login: admin=%s role=%q token=%s", admin.Email, admin.Role, token)
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
