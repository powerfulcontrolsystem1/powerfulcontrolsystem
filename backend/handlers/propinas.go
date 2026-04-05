package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaPropinasHandler gestiona configuracion, reportes y movimientos de propinas por empresa.
func EmpresaPropinasHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				cfg, err := dbpkg.GetEmpresaPropinasConfiguracion(dbEmp, empresaID)
				if err != nil {
					log.Printf("[propinas] get config empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo consultar la configuracion de propinas", http.StatusInternalServerError)
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
				report, err := dbpkg.GetEmpresaPropinasReporte(dbEmp, empresaID, dbpkg.EmpresaPropinaMovimientoFilter{
					Desde:            strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:            strings.TrimSpace(r.URL.Query().Get("hasta")),
					ModoDistribucion: strings.TrimSpace(r.URL.Query().Get("modo")),
					Usuario:          strings.TrimSpace(r.URL.Query().Get("usuario")),
					IncludeInactive:  queryBool(r, "include_inactive"),
					Limit:            limit,
				})
				if err != nil {
					log.Printf("[propinas] get report empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo construir el reporte de propinas", http.StatusInternalServerError)
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
				rows, err := dbpkg.ListEmpresaPropinaMovimientos(dbEmp, empresaID, dbpkg.EmpresaPropinaMovimientoFilter{
					Desde:            strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:            strings.TrimSpace(r.URL.Query().Get("hasta")),
					ModoDistribucion: strings.TrimSpace(r.URL.Query().Get("modo")),
					Usuario:          strings.TrimSpace(r.URL.Query().Get("usuario")),
					IncludeInactive:  queryBool(r, "include_inactive"),
					Limit:            limit,
				})
				if err != nil {
					log.Printf("[propinas] list movimientos empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo listar movimientos de propinas", http.StatusInternalServerError)
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
			var payload dbpkg.EmpresaPropinasConfiguracion
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
				payload.HabilitarPropina = true
			}
			if action == "desactivar" {
				payload.HabilitarPropina = false
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertEmpresaPropinasConfiguracion(dbEmp, payload)
			if err != nil {
				log.Printf("[propinas] upsert config empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			cfg, err := dbpkg.GetEmpresaPropinasConfiguracion(dbEmp, payload.EmpresaID)
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
