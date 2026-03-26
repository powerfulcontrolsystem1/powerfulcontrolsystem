package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"

	"github.com/you/pos-backend/auth"
	dbpkg "github.com/you/pos-backend/db"
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
func HandleGoogleCallback(db *sql.DB, clientID, clientSecret, redirectURL string) http.HandlerFunc {
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

		if err := dbpkg.UpsertUser(db, userinfo.Email, userinfo.Name); err != nil {
			log.Println("db upsert error:", err)
		}

		// Asegurar que el usuario tenga una empresa por defecto asociada
		if err := dbpkg.EnsureUserEmpresa(db, userinfo.Email, "Empresa de "+userinfo.Name); err != nil {
			log.Println("db ensure empresa error:", err)
		}

		http.Redirect(w, r, "/seleccionar_empresa.html", http.StatusFound)
	}
}
