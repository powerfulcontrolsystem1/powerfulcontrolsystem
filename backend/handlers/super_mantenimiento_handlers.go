package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	dbpkg "github.com/you/pos-backend/db"
)

// MantenimientoPayload define los datos de configuración de mantenimiento
type MantenimientoPayload struct {
	Activo bool `json:"mantenimiento_activo"`
}

// SuperMantenimientoConfigHandler maneja GET y PUT de modo mantenimiento.
func SuperMantenimientoConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			val, _, err := dbpkg.GetConfigValue(dbSuper, "mantenimiento_activo")
			if err != nil {
				log.Printf("[mantenimiento] error reading value: %v", err)
			}
			isActivo := val == "true"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"mantenimiento_activo": isActivo,
			})
			return

		case http.MethodPut:
			var payload MantenimientoPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON inválido", http.StatusBadRequest)
				return
			}
			
			valToSet := "false"
			if payload.Activo {
				valToSet = "true"
			}
			
			if err := dbpkg.SetConfigValue(dbSuper, "mantenimiento_activo", valToSet, false); err != nil {
				log.Printf("[mantenimiento] failed to update: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
			return

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
