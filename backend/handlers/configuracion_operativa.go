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
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if action == "historial" {
				historialID, err := parseInt64QueryOptional(r, "historial_id")
				if err != nil {
					http.Error(w, "historial_id invalido", http.StatusBadRequest)
					return
				}

				if historialID > 0 {
					row, err := dbpkg.GetEmpresaConfiguracionOperativaHistorialSnapshotByID(dbEmp, empresaID, historialID)
					if err != nil {
						log.Printf("[empresa_config_operativa] get historial empresa_id=%d historial_id=%d error: %v", empresaID, historialID, err)
						http.Error(w, "No se pudo cargar el historial", http.StatusInternalServerError)
						return
					}
					if row == nil {
						http.Error(w, "Historial no encontrado", http.StatusNotFound)
						return
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"historial": row,
					})
					return
				}

				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaConfiguracionOperativaHistorialSnapshots(dbEmp, empresaID, limit)
				if err != nil {
					log.Printf("[empresa_config_operativa] list historial empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo cargar el historial", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"historial": rows,
					"total":     len(rows),
				})
				return
			}

			cfg, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, empresaID)
			if err != nil {
				log.Printf("[empresa_config_operativa] get empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo cargar la configuracion operativa", http.StatusInternalServerError)
				return
			}

			if action == "simular" {
				ctx := parseConfiguracionOperativaContextoFromQuery(r)
				if ctx.Rol == "" {
					http.Error(w, "rol es obligatorio para simular", http.StatusBadRequest)
					return
				}
				resolved := dbpkg.ResolveEmpresaConfiguracionOperativaConContexto(cfg, ctx)
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"configuracion":         cfg,
					"contexto":              ctx,
					"permisos_simulados":    resolved,
					"permisos_rol_resuelto": resolved,
				})
				return
			}

			ctx := parseConfiguracionOperativaContextoFromQuery(r)
			if ctx.Rol != "" || ctx.CanalVenta != "" || ctx.SucursalID > 0 || ctx.Turno != "" {
				resolved := dbpkg.ResolveEmpresaConfiguracionOperativaConContexto(cfg, ctx)
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"configuracion":              cfg,
					"rol":                        ctx.Rol,
					"contexto":                   ctx,
					"permisos_rol_resuelto":      resolved,
					"permisos_contexto_resuelto": resolved,
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

				historialID := registrarSnapshotConfiguracionOperativa(
					dbEmp,
					payload.EmpresaID,
					"publicar",
					payload.UsuarioCreador,
					"Actualizacion de configuracion por rol",
					nil,
				)

				cfg, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, payload.EmpresaID)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "historial_id": historialID})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            true,
					"id":            id,
					"historial_id":  historialID,
					"configuracion": cfg,
				})
				return
			}

			if action == "politica" {
				var payload dbpkg.EmpresaConfiguracionOperativaPolitica
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

				id, err := dbpkg.UpsertEmpresaConfiguracionOperativaPolitica(dbEmp, payload)
				if err != nil {
					log.Printf("[empresa_config_operativa] upsert politica empresa_id=%d canal=%q sucursal=%d turno=%q error: %v", payload.EmpresaID, payload.CanalVenta, payload.SucursalID, payload.Turno, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				historialID := registrarSnapshotConfiguracionOperativa(
					dbEmp,
					payload.EmpresaID,
					"publicar",
					payload.UsuarioCreador,
					"Actualizacion de politica contextual",
					nil,
				)

				cfg, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, payload.EmpresaID)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "historial_id": historialID})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            true,
					"id":            id,
					"historial_id":  historialID,
					"configuracion": cfg,
				})
				return
			}

			if action == "simular" {
				var payload struct {
					EmpresaID     int64                                       `json:"empresa_id"`
					Contexto      dbpkg.EmpresaConfiguracionOperativaContexto `json:"contexto"`
					Rol           string                                      `json:"rol"`
					CanalVenta    string                                      `json:"canal_venta"`
					SucursalID    int64                                       `json:"sucursal_id"`
					Turno         string                                      `json:"turno"`
					Guardar       bool                                        `json:"guardar"`
					Observaciones string                                      `json:"observaciones"`
				}
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

				ctx := payload.Contexto
				if strings.TrimSpace(ctx.Rol) == "" {
					ctx.Rol = payload.Rol
				}
				if strings.TrimSpace(ctx.CanalVenta) == "" {
					ctx.CanalVenta = payload.CanalVenta
				}
				if ctx.SucursalID <= 0 {
					ctx.SucursalID = payload.SucursalID
				}
				if strings.TrimSpace(ctx.Turno) == "" {
					ctx.Turno = payload.Turno
				}
				if strings.TrimSpace(ctx.Rol) == "" {
					http.Error(w, "rol es obligatorio para simular", http.StatusBadRequest)
					return
				}

				permisos, err := dbpkg.GetEmpresaConfiguracionOperativaPermisosContexto(dbEmp, payload.EmpresaID, ctx)
				if err != nil {
					log.Printf("[empresa_config_operativa] simular empresa_id=%d error: %v", payload.EmpresaID, err)
					http.Error(w, "No se pudo simular la configuracion", http.StatusInternalServerError)
					return
				}

				response := map[string]interface{}{
					"ok":                 true,
					"empresa_id":         payload.EmpresaID,
					"contexto":           ctx,
					"permisos_simulados": permisos,
				}

				var historialID int64
				if payload.Guardar {
					historialID = registrarSnapshotConfiguracionOperativa(
						dbEmp,
						payload.EmpresaID,
						"simular",
						strings.TrimSpace(adminEmailFromRequest(r)),
						strings.TrimSpace(payload.Observaciones),
						response,
					)
					response["historial_id"] = historialID
				}

				writeJSON(w, http.StatusOK, response)
				return
			}

			if action == "rollback" {
				var payload struct {
					EmpresaID     int64  `json:"empresa_id"`
					HistorialID   int64  `json:"historial_id"`
					Observaciones string `json:"observaciones"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.HistorialID <= 0 {
					if historialID, err := parseInt64QueryOptional(r, "historial_id"); err == nil && historialID > 0 {
						payload.HistorialID = historialID
					}
				}
				if payload.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				if payload.HistorialID <= 0 {
					http.Error(w, "historial_id es obligatorio", http.StatusBadRequest)
					return
				}

				rollbackID, err := dbpkg.ApplyEmpresaConfiguracionOperativaRollback(
					dbEmp,
					payload.EmpresaID,
					payload.HistorialID,
					strings.TrimSpace(adminEmailFromRequest(r)),
					strings.TrimSpace(payload.Observaciones),
				)
				if err != nil {
					log.Printf("[empresa_config_operativa] rollback empresa_id=%d historial_id=%d error: %v", payload.EmpresaID, payload.HistorialID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				cfg, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, payload.EmpresaID)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"ok":          true,
						"rollback_id": rollbackID,
					})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            true,
					"rollback_id":   rollbackID,
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

			for _, policyCfg := range payload.Politicas {
				policyCfg.EmpresaID = payload.EmpresaID
				if strings.TrimSpace(policyCfg.UsuarioCreador) == "" {
					policyCfg.UsuarioCreador = payload.UsuarioCreador
				}
				if _, errPolicy := dbpkg.UpsertEmpresaConfiguracionOperativaPolitica(dbEmp, policyCfg); errPolicy != nil {
					log.Printf("[empresa_config_operativa] upsert policy from payload empresa_id=%d canal=%q sucursal=%d turno=%q error: %v", policyCfg.EmpresaID, policyCfg.CanalVenta, policyCfg.SucursalID, policyCfg.Turno, errPolicy)
				}
			}

			historialID := registrarSnapshotConfiguracionOperativa(
				dbEmp,
				payload.EmpresaID,
				"publicar",
				payload.UsuarioCreador,
				"Actualizacion general de configuracion operativa",
				nil,
			)

			cfg, err := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, payload.EmpresaID)
			if err != nil {
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "historial_id": historialID})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"historial_id":  historialID,
				"configuracion": cfg,
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func parseConfiguracionOperativaContextoFromQuery(r *http.Request) dbpkg.EmpresaConfiguracionOperativaContexto {
	ctx := dbpkg.EmpresaConfiguracionOperativaContexto{
		Rol:        strings.TrimSpace(r.URL.Query().Get("rol")),
		CanalVenta: strings.TrimSpace(r.URL.Query().Get("canal_venta")),
		Turno:      strings.TrimSpace(r.URL.Query().Get("turno")),
	}
	if sucursalID, err := parseInt64QueryOptional(r, "sucursal_id"); err == nil && sucursalID > 0 {
		ctx.SucursalID = sucursalID
	}
	return ctx
}

func registrarSnapshotConfiguracionOperativa(dbEmp *sql.DB, empresaID int64, evento, usuario, observaciones string, simulacion interface{}) int64 {
	if empresaID <= 0 {
		return 0
	}
	var simulacionJSON string
	if simulacion != nil {
		raw, err := json.Marshal(simulacion)
		if err != nil {
			log.Printf("[empresa_config_operativa] marshal simulacion empresa_id=%d error: %v", empresaID, err)
		} else {
			simulacionJSON = string(raw)
		}
	}

	historialID, err := dbpkg.CreateEmpresaConfiguracionOperativaHistorialSnapshot(dbEmp, dbpkg.EmpresaConfiguracionOperativaHistorialSnapshot{
		EmpresaID:      empresaID,
		Evento:         strings.TrimSpace(evento),
		UsuarioCreador: strings.TrimSpace(usuario),
		Estado:         "activo",
		Observaciones:  strings.TrimSpace(observaciones),
		SimulacionJSON: simulacionJSON,
	})
	if err != nil {
		log.Printf("[empresa_config_operativa] snapshot empresa_id=%d evento=%q error: %v", empresaID, evento, err)
		return 0
	}
	return historialID
}
