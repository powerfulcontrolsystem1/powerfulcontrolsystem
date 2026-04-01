package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaConfiguracionAvanzadaHandler expone la configuración avanzada por empresa
// para preparación de facturación electrónica en Colombia.
func EmpresaConfiguracionAvanzadaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, empresaID)
			if err != nil {
				log.Printf("[empresa_config_avanzada] get empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo cargar la configuración avanzada", http.StatusInternalServerError)
				return
			}
			hydrateEmpresaDefaults(dbEmp, cfg)
			writeJSON(w, http.StatusOK, cfg)
			return

		case http.MethodPost, http.MethodPut:
			var payload dbpkg.EmpresaConfiguracionAvanzada
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}

			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, payload)
			if err != nil {
				log.Printf("[empresa_config_avanzada] upsert empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, "No se pudo guardar la configuración avanzada", http.StatusBadRequest)
				return
			}

			cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, payload.EmpresaID)
			if err != nil {
				log.Printf("[empresa_config_avanzada] get after upsert empresa_id=%d error: %v", payload.EmpresaID, err)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			}
			hydrateEmpresaDefaults(dbEmp, cfg)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"configuracion": cfg,
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func hydrateEmpresaDefaults(dbEmp *sql.DB, cfg *dbpkg.EmpresaConfiguracionAvanzada) {
	if cfg == nil || cfg.EmpresaID <= 0 {
		return
	}
	if strings.TrimSpace(cfg.RazonSocial) != "" && strings.TrimSpace(cfg.NIT) != "" {
		return
	}

	var nombre string
	var nit string
	err := dbEmp.QueryRow(`SELECT COALESCE(nombre, ''), COALESCE(nit, '') FROM empresas WHERE id = ? LIMIT 1`, cfg.EmpresaID).Scan(&nombre, &nit)
	if err != nil {
		return
	}
	if strings.TrimSpace(cfg.RazonSocial) == "" {
		cfg.RazonSocial = strings.TrimSpace(nombre)
	}
	if strings.TrimSpace(cfg.NIT) == "" {
		cfg.NIT = strings.TrimSpace(nit)
	}
}
