package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaConfiguracionOperativaHandler gestiona configuracion de cobro por empresa y por rol.
func EmpresaConfiguracionOperativaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			cfg, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, empresaID)
			if err != nil {
				log.Printf("[empresa_config_operativa] get empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo cargar la configuracion operativa", http.StatusInternalServerError)
				return
			}

			rol := strings.TrimSpace(r.URL.Query().Get("rol"))
			if rol != "" {
				resolved := dbpkg.ResolveEmpresaConfiguracionOperativaParaRol(cfg, rol)
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"configuracion":         cfg,
					"rol":                   rol,
					"permisos_rol_resuelto": resolved,
				})
				return
			}

			writeJSON(w, http.StatusOK, cfg)
			return

		case http.MethodPost, http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "rol" {
				var payload dbpkg.EmpresaConfiguracionOperativaRol
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
				id, err := dbpkg.UpsertEmpresaConfiguracionOperativaRol(dbEmp, payload)
				if err != nil {
					log.Printf("[empresa_config_operativa] upsert rol empresa_id=%d rol=%q error: %v", payload.EmpresaID, payload.Rol, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				cfg, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, payload.EmpresaID)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            true,
					"id":            id,
					"configuracion": cfg,
				})
				return
			}

			var payload dbpkg.EmpresaConfiguracionOperativa
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
			id, err := dbpkg.UpsertEmpresaConfiguracionOperativa(dbEmp, payload)
			if err != nil {
				log.Printf("[empresa_config_operativa] upsert empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			for _, roleCfg := range payload.Roles {
				if strings.TrimSpace(roleCfg.Rol) == "" {
					continue
				}
				roleCfg.EmpresaID = payload.EmpresaID
				if strings.TrimSpace(roleCfg.UsuarioCreador) == "" {
					roleCfg.UsuarioCreador = payload.UsuarioCreador
				}
				if _, errRole := dbpkg.UpsertEmpresaConfiguracionOperativaRol(dbEmp, roleCfg); errRole != nil {
					log.Printf("[empresa_config_operativa] upsert role from payload empresa_id=%d rol=%q error: %v", roleCfg.EmpresaID, roleCfg.Rol, errRole)
				}
			}

			cfg, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, payload.EmpresaID)
			if err != nil {
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			}
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
