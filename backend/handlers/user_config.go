package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	dbpkg "github.com/you/pos-backend/db"
)

func UserConfiguracionHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validar sesión
		c, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		session, err := dbpkg.GetSessionByToken(dbSuper, c.Value)
		if err != nil || session == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			apariencia, err := dbpkg.GetUsuarioApariencia(dbSuper, session.AdminEmail)
			if err != nil {
				log.Println("Error obteniendo apariencia:", err)
				apariencia = "dark" // default seguro
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":         true,
				"apariencia": apariencia,
			})
			return
		}

		if r.Method == http.MethodPost {
			var payload struct {
				Apariencia string `json:"apariencia"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			if payload.Apariencia != "" {
				err := dbpkg.SetUsuarioApariencia(dbSuper, session.AdminEmail, payload.Apariencia)
				if err != nil {
					log.Println("Error almacenando apariencia en DB:", err)
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
