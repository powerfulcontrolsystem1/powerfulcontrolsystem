package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
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
			// Leer cuerpo y decodificar tanto a mapa (para detectar keys presentes)
			// como a la estructura tipada. Luego fusionar sobre la configuración existente
			// para soportar parches parciales sin sobrescribir valores no enviados.
			body, err := io.ReadAll(r.Body)
			if err != nil || len(body) == 0 {
				http.Error(w, "JSON invalido o cuerpo vacio", http.StatusBadRequest)
				return
			}

			var raw map[string]interface{}
			if err := json.Unmarshal(body, &raw); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}

			var payload dbpkg.EmpresaConfiguracionAvanzada
			_ = json.Unmarshal(body, &payload) // ignore error; usamos raw map para presencia de keys

			// Si empresa_id no viene en el payload, intentar obtenerlo de la query
			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
					raw["empresa_id"] = empresaID
				} else if v, ok := raw["empresa_id"]; ok {
					// Si viene como numero en raw (float64), convertir
					switch vv := v.(type) {
					case float64:
						payload.EmpresaID = int64(vv)
					case string:
						// intentar parsear string numerico
						if parsed, perr := strconv.ParseInt(strings.TrimSpace(vv), 10, 64); perr == nil {
							payload.EmpresaID = parsed
						}
					}
				}
			}

			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}

			// Obtener configuración existente y fusionar con los campos enviados
			existingCfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, payload.EmpresaID)
			if err != nil {
				log.Printf("[empresa_config_avanzada] get existing empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, "No se pudo cargar la configuración existente", http.StatusInternalServerError)
				return
			}

			// serializar existing a map, sobreescribir con raw y volver a struct
			existBytes, _ := json.Marshal(existingCfg)
			var existMap map[string]interface{}
			_ = json.Unmarshal(existBytes, &existMap)
			for k, v := range raw {
				existMap[k] = v
			}
			mergedBytes, merr := json.Marshal(existMap)
			if merr != nil {
				log.Printf("[empresa_config_avanzada] marshal merged error: %v", merr)
				http.Error(w, "Error interno al procesar payload", http.StatusInternalServerError)
				return
			}

			var merged dbpkg.EmpresaConfiguracionAvanzada
			if err := json.Unmarshal(mergedBytes, &merged); err != nil {
				log.Printf("[empresa_config_avanzada] unmarshal merged error: %v", err)
				http.Error(w, "Error interno al procesar payload", http.StatusInternalServerError)
				return
			}

			merged.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, merged)
			if err != nil {
				log.Printf("[empresa_config_avanzada] upsert empresa_id=%d error: %v", merged.EmpresaID, err)
				http.Error(w, "No se pudo guardar la configuración avanzada", http.StatusInternalServerError)
				return
			}

			cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, merged.EmpresaID)
			if err != nil {
				log.Printf("[empresa_config_avanzada] get after upsert empresa_id=%d error: %v", merged.EmpresaID, err)
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
	err := dbpkg.QueryRowCompat(dbEmp, `SELECT COALESCE(nombre, ''), COALESCE(nit, '') FROM empresas WHERE id = ? LIMIT 1`, cfg.EmpresaID).Scan(&nombre, &nit)
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
