package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strings"

    dbpkg "github.com/you/pos-backend/db"
)

// VendedorConfigHandler gestiona la configuración avanzada del módulo vendedor (vendedor de licencia)
func VendedorConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodGet:
            defaultPct, _, _, pctUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "vendedor.default_comision_pct")
            meses, _, _, mesesUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "vendedor.default_meses_renovacion")
            fechaFin, _, _, fechaUpdated, _ := dbpkg.GetConfigEntry(dbSuper, "vendedor.comision_fecha_fin")

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{
                "default_comision_pct": defaultPct,
                "default_comision_pct_updated": pctUpdated,
                "default_meses_renovacion": meses,
                "default_meses_renovacion_updated": mesesUpdated,
                "comision_fecha_fin": fechaFin,
                "comision_fecha_fin_updated": fechaUpdated,
            })
            return

        case http.MethodPost, http.MethodPut:
            var payload struct {
                DefaultComisionPct string `json:"default_comision_pct"`
                MesesRenovacion    string `json:"meses_renovacion"`
                ComisionFechaFin    string `json:"comision_fecha_fin"`
            }
            if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
                http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
                return
            }

            // Guardar solo valores no vacíos (permitir dejar en blanco para no modificar)
            if strings.TrimSpace(payload.DefaultComisionPct) != "" {
                if err := dbpkg.SetConfigValue(dbSuper, "vendedor.default_comision_pct", strings.TrimSpace(payload.DefaultComisionPct), false); err != nil {
                    http.Error(w, "failed to save default_comision_pct: "+err.Error(), http.StatusInternalServerError)
                    return
                }
            }
            if strings.TrimSpace(payload.MesesRenovacion) != "" {
                if err := dbpkg.SetConfigValue(dbSuper, "vendedor.default_meses_renovacion", strings.TrimSpace(payload.MesesRenovacion), false); err != nil {
                    http.Error(w, "failed to save meses_renovacion: "+err.Error(), http.StatusInternalServerError)
                    return
                }
            }
            if strings.TrimSpace(payload.ComisionFechaFin) != "" {
                if err := dbpkg.SetConfigValue(dbSuper, "vendedor.comision_fecha_fin", strings.TrimSpace(payload.ComisionFechaFin), false); err != nil {
                    http.Error(w, "failed to save comision_fecha_fin: "+err.Error(), http.StatusInternalServerError)
                    return
                }
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(map[string]interface{}{"saved": true})
            return

        default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
    }
}
