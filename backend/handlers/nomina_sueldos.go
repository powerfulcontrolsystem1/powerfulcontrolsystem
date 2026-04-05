package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaNominaSueldosHandler gestiona configuracion, empleados y liquidaciones de nomina integradas con asistencia.
func EmpresaNominaSueldosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			switch action {
			case "", "config", "configuracion":
				cfg, err := dbpkg.GetEmpresaNominaConfiguracion(dbEmp, empresaID)
				if err != nil {
					log.Printf("[nomina] get config empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo consultar la configuracion de nomina", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return

			case "empleados", "empleado":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaNominaEmpleados(dbEmp, empresaID, queryBool(r, "include_inactive"), strings.TrimSpace(r.URL.Query().Get("q")), limit)
				if err != nil {
					log.Printf("[nomina] list empleados empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo listar empleados de nomina", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "festivos", "dias_festivos", "dia_festivo":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				rows, err := dbpkg.ListEmpresaNominaFestivos(dbEmp, empresaID, queryBool(r, "include_inactive"), desde, hasta, limit)
				if err != nil {
					log.Printf("[nomina] list festivos empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo listar dias festivos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "liquidaciones", "nominas":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				empleadoNominaID, err := parseInt64QueryOptional(r, "empleado_nomina_id")
				if err != nil {
					http.Error(w, "empleado_nomina_id invalido", http.StatusBadRequest)
					return
				}
				periodoDesde := strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				if periodoDesde == "" {
					periodoDesde = strings.TrimSpace(r.URL.Query().Get("desde"))
				}
				periodoHasta := strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				if periodoHasta == "" {
					periodoHasta = strings.TrimSpace(r.URL.Query().Get("hasta"))
				}

				rows, err := dbpkg.ListEmpresaNominaLiquidaciones(dbEmp, empresaID, dbpkg.EmpresaNominaLiquidacionFilter{
					PeriodoDesde:     periodoDesde,
					PeriodoHasta:     periodoHasta,
					EmpleadoNominaID: empleadoNominaID,
					IncludeInactive:  queryBool(r, "include_inactive"),
					Limit:            limit,
				})
				if err != nil {
					log.Printf("[nomina] list liquidaciones empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo listar liquidaciones de nomina", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			default:
				http.Error(w, "action invalida. Use: config, empleados, festivos o liquidaciones", http.StatusBadRequest)
				return
			}

		case http.MethodPost:
			switch action {
			case "empleado", "empleados":
				var payload dbpkg.EmpresaNominaEmpleado
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
				id, err := dbpkg.CreateEmpresaNominaEmpleado(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return

			case "festivo", "dia_festivo":
				var payload dbpkg.EmpresaNominaFestivo
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
				id, err := dbpkg.CreateEmpresaNominaFestivo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return

			case "calcular", "liquidar", "generar_nomina":
				var req dbpkg.EmpresaNominaCalculoRequest
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						http.Error(w, "JSON invalido", http.StatusBadRequest)
						return
					}
				}
				if req.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						req.EmpresaID = empresaID
					}
				}
				if req.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				if req.PeriodoDesde == "" {
					req.PeriodoDesde = strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				}
				if req.PeriodoHasta == "" {
					req.PeriodoHasta = strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				}
				if req.PeriodoDesde == "" {
					req.PeriodoDesde = strings.TrimSpace(r.URL.Query().Get("desde"))
				}
				if req.PeriodoHasta == "" {
					req.PeriodoHasta = strings.TrimSpace(r.URL.Query().Get("hasta"))
				}
				if req.EmpleadoNominaID <= 0 {
					if id, err := parseInt64QueryOptional(r, "empleado_nomina_id"); err == nil && id > 0 {
						req.EmpleadoNominaID = id
					}
				}
				req.Overwrite = req.Overwrite || queryBool(r, "overwrite") || queryBool(r, "recalcular")
				req.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				result, err := dbpkg.GenerateEmpresaNominaLiquidaciones(dbEmp, req)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, result)
				return

			default:
				http.Error(w, "action invalida. Use: empleado, festivo o calcular", http.StatusBadRequest)
				return
			}

		case http.MethodPut:
			switch action {
			case "", "config", "configuracion":
				var payload dbpkg.EmpresaNominaConfiguracion
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
				id, err := dbpkg.UpsertEmpresaNominaConfiguracion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				cfg, err := dbpkg.GetEmpresaNominaConfiguracion(dbEmp, payload.EmpresaID)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "configuracion": cfg})
				return

			case "empleado":
				var payload dbpkg.EmpresaNominaEmpleado
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 || payload.ID <= 0 {
					http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
					return
				}
				if err := dbpkg.UpdateEmpresaNominaEmpleado(dbEmp, payload); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "empleado de nomina no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return

			case "activar_empleado", "desactivar_empleado":
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id es obligatorio", http.StatusBadRequest)
					return
				}
				next := "activo"
				if action == "desactivar_empleado" {
					next = "inactivo"
				}
				if err := dbpkg.SetEmpresaNominaEmpleadoEstado(dbEmp, empresaID, id, next); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "empleado de nomina no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo actualizar estado del empleado de nomina", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": next})
				return

			default:
				http.Error(w, "action invalida. Use: config, empleado, activar_empleado o desactivar_empleado", http.StatusBadRequest)
				return
			}

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}

			switch action {
			case "empleado":
				if err := dbpkg.DeleteEmpresaNominaEmpleado(dbEmp, empresaID, id); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "empleado de nomina no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo eliminar empleado de nomina", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return

			case "festivo", "dia_festivo":
				if err := dbpkg.DeleteEmpresaNominaFestivo(dbEmp, empresaID, id); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "dia festivo no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo eliminar dia festivo", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return

			default:
				http.Error(w, "action invalida. Use: empleado o festivo", http.StatusBadRequest)
				return
			}
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
