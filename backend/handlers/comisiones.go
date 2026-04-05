package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaComisionesServicioHandler gestiona configuracion, movimientos y reporte de comisiones por servicio.
func EmpresaComisionesServicioHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			switch action {
			case "", "config", "configuracion":
				cfg, err := dbpkg.GetEmpresaComisionesServicioConfiguracion(dbEmp, empresaID)
				if err != nil {
					log.Printf("[comisiones] get config empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo consultar la configuracion de comisiones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return

			case "reporte", "resumen", "dashboard":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				report, err := dbpkg.GetEmpresaComisionesServicioReporte(dbEmp, empresaID, dbpkg.EmpresaComisionServicioMovimientoFilter{
					Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
					UsuarioLavador:  strings.TrimSpace(r.URL.Query().Get("usuario_lavador")),
					ServicioFiltro:  strings.TrimSpace(r.URL.Query().Get("servicio_filtro")),
					IncludeInactive: queryBool(r, "include_inactive"),
					Limit:           limit,
				})
				if err != nil {
					log.Printf("[comisiones] get report empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo construir el reporte de comisiones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, report)
				return

			case "movimientos", "detalle":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaComisionServicioMovimientos(dbEmp, empresaID, dbpkg.EmpresaComisionServicioMovimientoFilter{
					Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
					UsuarioLavador:  strings.TrimSpace(r.URL.Query().Get("usuario_lavador")),
					ServicioFiltro:  strings.TrimSpace(r.URL.Query().Get("servicio_filtro")),
					IncludeInactive: queryBool(r, "include_inactive"),
					Limit:           limit,
				})
				if err != nil {
					log.Printf("[comisiones] list movimientos empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo listar movimientos de comisiones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			default:
				http.Error(w, "action invalida. Use: config, reporte o movimientos", http.StatusBadRequest)
				return
			}

		case http.MethodPost, http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			var payload dbpkg.EmpresaComisionesServicioConfiguracion
			if r.Body != nil {
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
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
			if action == "activar" {
				payload.HabilitarComisiones = true
			}
			if action == "desactivar" {
				payload.HabilitarComisiones = false
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertEmpresaComisionesServicioConfiguracion(dbEmp, payload)
			if err != nil {
				log.Printf("[comisiones] upsert config empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			cfg, err := dbpkg.GetEmpresaComisionesServicioConfiguracion(dbEmp, payload.EmpresaID)
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
