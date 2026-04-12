package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaEstacionesColumnasPrefHandler gestiona preferencias de columnas por empresa/usuario/rol
func EmpresaEstacionesColumnasPrefHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			pref, err := dbpkg.GetEstacionColumnPreferences(dbEmp, empresaID, adminEmail, 0)
			if err != nil {
				if err == sql.ErrNoRows {
					// default columns
					defaultCols := map[string]bool{
						"nombre_estacion":    true,
						"activacion_carrito": true,
						"cliente_nombre":     true,
						"hora_salida":        true,
						"total_carrito":      true,
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "columnas": defaultCols})
					return
				}
				log.Printf("[estaciones_prefs] get empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudieron obtener preferencias", http.StatusInternalServerError)
				return
			}
			if pref.Columnas == nil {
				if pref.ColumnasJSON != "" {
					var mp map[string]bool
					_ = json.Unmarshal([]byte(pref.ColumnasJSON), &mp)
					pref.Columnas = mp
				} else {
					pref.Columnas = map[string]bool{}
				}
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "columnas": pref.Columnas})
			return

		case http.MethodPut, http.MethodPost:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			var payload struct {
				UsuarioEmail string          `json:"usuario_email,omitempty"`
				RolID        int64           `json:"rol_id,omitempty"`
				Columnas     map[string]bool `json:"columnas"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			usuario := strings.TrimSpace(payload.UsuarioEmail)
			if usuario == "" {
				usuario = strings.TrimSpace(adminEmailFromRequest(r))
			}
			if payload.Columnas == nil {
				http.Error(w, "columnas es obligatorio", http.StatusBadRequest)
				return
			}
			p := dbpkg.EstacionColumnPreferences{
				EmpresaID:      empresaID,
				UsuarioEmail:   usuario,
				RolID:          payload.RolID,
				Columnas:       payload.Columnas,
				UsuarioCreador: strings.TrimSpace(adminEmailFromRequest(r)),
				Estado:         "activo",
			}
			by, _ := json.Marshal(p.Columnas)
			p.ColumnasJSON = string(by)
			id, err := dbpkg.UpsertEstacionColumnPreferences(dbEmp, &p)
			if err != nil {
				log.Printf("[estaciones_prefs] upsert empresa_id=%d user=%s error: %v", empresaID, p.UsuarioEmail, err)
				http.Error(w, "No se pudieron guardar preferencias", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
