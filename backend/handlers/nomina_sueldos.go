package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaNominaSueldosHandler gestiona configuracion, empleados y liquidaciones de nomina integradas con asistencia.
func EmpresaNominaSueldosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureEmpresaNominaSchema(dbEmp); err != nil {
			log.Printf("[nomina] ensure schema error: %v", err)
			http.Error(w, "No se pudo preparar el modulo de nomina", http.StatusInternalServerError)
			return
		}

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

			case "pagos", "pagos_nomina":
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
				rows, err := dbpkg.ListEmpresaNominaPagos(dbEmp, empresaID, dbpkg.EmpresaNominaPagoFilter{
					PeriodoDesde:     strings.TrimSpace(r.URL.Query().Get("periodo_desde")),
					PeriodoHasta:     strings.TrimSpace(r.URL.Query().Get("periodo_hasta")),
					EmpleadoNominaID: empleadoNominaID,
					IncludeInactive:  queryBool(r, "include_inactive"),
					Limit:            limit,
				})
				if err != nil {
					http.Error(w, "No se pudo listar pagos de nomina", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "provisiones", "resumen_empresarial", "aportes":
				empleadoNominaID, err := parseInt64QueryOptional(r, "empleado_nomina_id")
				if err != nil {
					http.Error(w, "empleado_nomina_id invalido", http.StatusBadRequest)
					return
				}
				resumen, err := dbpkg.GetEmpresaNominaProvisionesResumen(
					dbEmp,
					empresaID,
					strings.TrimSpace(r.URL.Query().Get("periodo_desde")),
					strings.TrimSpace(r.URL.Query().Get("periodo_hasta")),
					empleadoNominaID,
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return

			case "dashboard", "resumen", "resumen_operativo":
				periodoDesde := strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				periodoHasta := strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				if periodoDesde == "" || periodoHasta == "" {
					now := time.Now()
					periodoDesde = now.Format("2006-01") + "-01"
					periodoHasta = now.Format("2006-01-02")
				}
				empleadoNominaID, err := parseInt64QueryOptional(r, "empleado_nomina_id")
				if err != nil {
					http.Error(w, "empleado_nomina_id invalido", http.StatusBadRequest)
					return
				}
				dashboard, err := buildEmpresaNominaDashboard(dbEmp, empresaID, periodoDesde, periodoHasta, empleadoNominaID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, dashboard)
				return

			case "desprendible", "desprendible_nomina":
				empleadoNominaID, err := parseInt64Query(r, "empleado_nomina_id")
				if err != nil {
					http.Error(w, "empleado_nomina_id es obligatorio", http.StatusBadRequest)
					return
				}
				periodoDesde := strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				periodoHasta := strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				if periodoDesde == "" {
					periodoDesde = strings.TrimSpace(r.URL.Query().Get("desde"))
				}
				if periodoHasta == "" {
					periodoHasta = strings.TrimSpace(r.URL.Query().Get("hasta"))
				}
				doc, err := dbpkg.GetEmpresaNominaDesprendible(dbEmp, empresaID, empleadoNominaID, periodoDesde, periodoHasta)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "No se encontro liquidacion para desprendible", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, doc)
				return

			case "conciliacion_asistencia", "conciliar_asistencia":
				periodoDesde := strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				periodoHasta := strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				if periodoDesde == "" {
					periodoDesde = strings.TrimSpace(r.URL.Query().Get("desde"))
				}
				if periodoHasta == "" {
					periodoHasta = strings.TrimSpace(r.URL.Query().Get("hasta"))
				}
				empleadoNominaID, err := parseInt64QueryOptional(r, "empleado_nomina_id")
				if err != nil {
					http.Error(w, "empleado_nomina_id invalido", http.StatusBadRequest)
					return
				}
				result, err := dbpkg.ConciliarEmpresaNominaAsistencia(dbEmp, dbpkg.EmpresaNominaConciliacionRequest{
					EmpresaID:        empresaID,
					PeriodoDesde:     periodoDesde,
					PeriodoHasta:     periodoHasta,
					EmpleadoNominaID: empleadoNominaID,
					AutoRecalcular:   false,
					UsuarioCreador:   strings.TrimSpace(adminEmailFromRequest(r)),
				})
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, result)
				return

			case "conceptos_colombia", "conceptos":
				rows, err := dbpkg.ListEmpresaNominaConceptosColombia(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("tipo")), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar conceptos Colombia", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "novedades_colombia", "novedades":
				rows, err := dbpkg.ListEmpresaNominaNovedadesColombia(
					dbEmp,
					empresaID,
					strings.TrimSpace(r.URL.Query().Get("periodo_desde")),
					strings.TrimSpace(r.URL.Query().Get("periodo_hasta")),
					strings.TrimSpace(r.URL.Query().Get("estado_aprobacion")),
					500,
				)
				if err != nil {
					http.Error(w, "No se pudieron listar novedades Colombia", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "pila_colombia", "pila":
				rows, err := dbpkg.ListEmpresaNominaPILAResumenColombia(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")), 1000)
				if err != nil {
					http.Error(w, "No se pudo listar resumen PILA", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "dashboard_colombia", "colombia_avanzada":
				ds, err := dbpkg.BuildEmpresaNominaColombiaAvanzadaDashboard(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")))
				if err != nil {
					http.Error(w, "No se pudo consultar nomina Colombia avanzada", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, ds)
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

			case "conciliar_asistencia", "conciliacion_asistencia":
				var req dbpkg.EmpresaNominaConciliacionRequest
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
				if strings.TrimSpace(req.PeriodoDesde) == "" {
					req.PeriodoDesde = strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				}
				if strings.TrimSpace(req.PeriodoHasta) == "" {
					req.PeriodoHasta = strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				}
				if strings.TrimSpace(req.PeriodoDesde) == "" {
					req.PeriodoDesde = strings.TrimSpace(r.URL.Query().Get("desde"))
				}
				if strings.TrimSpace(req.PeriodoHasta) == "" {
					req.PeriodoHasta = strings.TrimSpace(r.URL.Query().Get("hasta"))
				}
				if req.EmpleadoNominaID <= 0 {
					if id, err := parseInt64QueryOptional(r, "empleado_nomina_id"); err == nil && id > 0 {
						req.EmpleadoNominaID = id
					}
				}
				req.AutoRecalcular = req.AutoRecalcular || queryBool(r, "auto_recalcular") || queryBool(r, "recalcular")
				req.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				result, err := dbpkg.ConciliarEmpresaNominaAsistencia(dbEmp, req)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, result)
				return

			case "generar_pagos", "pagar_nomina":
				var payload struct {
					EmpresaID        int64  `json:"empresa_id"`
					PeriodoDesde     string `json:"periodo_desde"`
					PeriodoHasta     string `json:"periodo_hasta"`
					EmpleadoNominaID int64  `json:"empleado_nomina_id"`
					MetodoPago       string `json:"metodo_pago"`
					CuentaBancaria   string `json:"cuenta_bancaria"`
				}
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
				if strings.TrimSpace(payload.PeriodoDesde) == "" {
					payload.PeriodoDesde = strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				}
				if strings.TrimSpace(payload.PeriodoHasta) == "" {
					payload.PeriodoHasta = strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				}
				if payload.EmpleadoNominaID <= 0 {
					if id, err := parseInt64QueryOptional(r, "empleado_nomina_id"); err == nil && id > 0 {
						payload.EmpleadoNominaID = id
					}
				}
				result, err := dbpkg.GenerateEmpresaNominaPagos(
					dbEmp,
					payload.EmpresaID,
					strings.TrimSpace(payload.PeriodoDesde),
					strings.TrimSpace(payload.PeriodoHasta),
					payload.EmpleadoNominaID,
					strings.TrimSpace(payload.MetodoPago),
					strings.TrimSpace(payload.CuentaBancaria),
					strings.TrimSpace(adminEmailFromRequest(r)),
				)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, result)
				return

			case "concepto_colombia", "conceptos_colombia":
				var payload dbpkg.EmpresaNominaConceptoColombia
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaNominaConceptoColombia(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return

			case "novedad_colombia", "novedades_colombia":
				var payload dbpkg.EmpresaNominaNovedadColombia
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CreateEmpresaNominaNovedadColombia(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return

			case "generar_pila_colombia", "generar_pila":
				var payload struct {
					EmpresaID    int64  `json:"empresa_id"`
					PeriodoDesde string `json:"periodo_desde"`
					PeriodoHasta string `json:"periodo_hasta"`
				}
				if r.Body != nil {
					_ = json.NewDecoder(r.Body).Decode(&payload)
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if strings.TrimSpace(payload.PeriodoDesde) == "" {
					payload.PeriodoDesde = strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				}
				if strings.TrimSpace(payload.PeriodoHasta) == "" {
					payload.PeriodoHasta = strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				}
				if payload.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.GenerarEmpresaNominaPILAResumenColombia(dbEmp, payload.EmpresaID, payload.PeriodoDesde, payload.PeriodoHasta, strings.TrimSpace(adminEmailFromRequest(r)))
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "seed_colombia", "seed_colombia_avanzada":
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil || empresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SeedEmpresaNominaColombiaAvanzadaDemo(dbEmp, empresaID, strings.TrimSpace(adminEmailFromRequest(r))); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return

			default:
				http.Error(w, "action invalida. Use: empleado, festivo, calcular o conciliar_asistencia", http.StatusBadRequest)
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

type empresaNominaDashboard struct {
	EmpresaID            int64                                  `json:"empresa_id"`
	PeriodoDesde         string                                 `json:"periodo_desde"`
	PeriodoHasta         string                                 `json:"periodo_hasta"`
	EmpleadosActivos     int                                    `json:"empleados_activos"`
	EmpleadosInactivos   int                                    `json:"empleados_inactivos"`
	FestivosActivos      int                                    `json:"festivos_activos"`
	Liquidaciones        int                                    `json:"liquidaciones"`
	PagosGenerados       int                                    `json:"pagos_generados"`
	TotalDevengado       float64                                `json:"total_devengado"`
	TotalDeducciones     float64                                `json:"total_deducciones"`
	TotalNeto            float64                                `json:"total_neto"`
	TotalPagado          float64                                `json:"total_pagado"`
	CostoEmpresaEstimado float64                                `json:"costo_empresa_estimado"`
	Provisiones          *dbpkg.EmpresaNominaProvisionesResumen `json:"provisiones,omitempty"`
	Alertas              []string                               `json:"alertas,omitempty"`
}

func buildEmpresaNominaDashboard(dbEmp *sql.DB, empresaID int64, periodoDesde, periodoHasta string, empleadoNominaID int64) (*empresaNominaDashboard, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id es obligatorio")
	}
	out := &empresaNominaDashboard{
		EmpresaID:    empresaID,
		PeriodoDesde: strings.TrimSpace(periodoDesde),
		PeriodoHasta: strings.TrimSpace(periodoHasta),
		Alertas:      make([]string, 0),
	}

	empleados, err := dbpkg.ListEmpresaNominaEmpleados(dbEmp, empresaID, true, "", 5000)
	if err != nil {
		return nil, err
	}
	for _, row := range empleados {
		if strings.EqualFold(strings.TrimSpace(row.Estado), "activo") {
			out.EmpleadosActivos++
		} else {
			out.EmpleadosInactivos++
		}
	}

	festivos, err := dbpkg.ListEmpresaNominaFestivos(dbEmp, empresaID, false, out.PeriodoDesde, out.PeriodoHasta, 5000)
	if err == nil {
		out.FestivosActivos = len(festivos)
	}

	liquidaciones, err := dbpkg.ListEmpresaNominaLiquidaciones(dbEmp, empresaID, dbpkg.EmpresaNominaLiquidacionFilter{
		PeriodoDesde:     out.PeriodoDesde,
		PeriodoHasta:     out.PeriodoHasta,
		EmpleadoNominaID: empleadoNominaID,
		IncludeInactive:  false,
		Limit:            5000,
	})
	if err != nil {
		return nil, err
	}
	out.Liquidaciones = len(liquidaciones)
	for _, item := range liquidaciones {
		out.TotalDevengado += item.DevengadoTotal
		out.TotalDeducciones += item.DeduccionTotal
		out.TotalNeto += item.NetoPagar
	}

	pagos, err := dbpkg.ListEmpresaNominaPagos(dbEmp, empresaID, dbpkg.EmpresaNominaPagoFilter{
		PeriodoDesde:     out.PeriodoDesde,
		PeriodoHasta:     out.PeriodoHasta,
		EmpleadoNominaID: empleadoNominaID,
		IncludeInactive:  false,
		Limit:            5000,
	})
	if err == nil {
		out.PagosGenerados = len(pagos)
		for _, pago := range pagos {
			out.TotalPagado += pago.NetoPagado
		}
	}

	provisiones, err := dbpkg.GetEmpresaNominaProvisionesResumen(dbEmp, empresaID, out.PeriodoDesde, out.PeriodoHasta, empleadoNominaID)
	if err == nil && provisiones != nil {
		out.Provisiones = provisiones
		out.CostoEmpresaEstimado = provisiones.CostoEmpresaEstimado
	}

	if out.EmpleadosActivos == 0 {
		out.Alertas = append(out.Alertas, "No hay empleados activos vinculados a nómina.")
	}
	if out.Liquidaciones == 0 {
		out.Alertas = append(out.Alertas, "Todavía no hay liquidaciones generadas para el rango seleccionado.")
	}
	if out.Liquidaciones > 0 && out.PagosGenerados == 0 {
		out.Alertas = append(out.Alertas, "Hay liquidaciones calculadas sin pagos de nómina registrados.")
	}
	if out.PagosGenerados > 0 && out.PagosGenerados < out.Liquidaciones {
		out.Alertas = append(out.Alertas, "Existen liquidaciones pendientes por pagar o ya pagadas parcialmente.")
	}

	return out, nil
}
