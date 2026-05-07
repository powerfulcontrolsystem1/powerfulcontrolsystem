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

			case "control_contable", "validacion_contable", "auditoria_contable":
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
				control, err := buildEmpresaNominaControlContable(dbEmp, empresaID, periodoDesde, periodoHasta, empleadoNominaID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, control)
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
					EmpresaID             int64  `json:"empresa_id"`
					PeriodoDesde          string `json:"periodo_desde"`
					PeriodoHasta          string `json:"periodo_hasta"`
					EmpleadoNominaID      int64  `json:"empleado_nomina_id"`
					MetodoPago            string `json:"metodo_pago"`
					CuentaBancaria        string `json:"cuenta_bancaria"`
					ConfirmarAdvertencias bool   `json:"confirmar_advertencias"`
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
				control, err := buildEmpresaNominaControlContable(dbEmp, payload.EmpresaID, payload.PeriodoDesde, payload.PeriodoHasta, payload.EmpleadoNominaID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if !control.PuedeGenerarPagos {
					writeNominaControlConflict(w, "El control contable no permite generar pagos de nomina para este periodo.", control)
					return
				}
				if control.RequiereConfirmacion && !payload.ConfirmarAdvertencias {
					writeNominaControlConflict(w, "El control contable tiene advertencias que deben revisarse antes de generar pagos.", control)
					return
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
	ControlContable      *empresaNominaControlContable          `json:"control_contable,omitempty"`
	Alertas              []string                               `json:"alertas,omitempty"`
}

type empresaNominaControlContable struct {
	EmpresaID            int64    `json:"empresa_id"`
	PeriodoDesde         string   `json:"periodo_desde"`
	PeriodoHasta         string   `json:"periodo_hasta"`
	EmpleadoNominaID     int64    `json:"empleado_nomina_id,omitempty"`
	Estado               string   `json:"estado"`
	PuedeGenerarPagos    bool     `json:"puede_generar_pagos"`
	RequiereConfirmacion bool     `json:"requiere_confirmacion"`
	Liquidaciones        int      `json:"liquidaciones"`
	PagosGenerados       int      `json:"pagos_generados"`
	PendientesPago       int      `json:"pendientes_pago"`
	NovedadesPendientes  int      `json:"novedades_pendientes"`
	RegistrosPILA        int      `json:"registros_pila"`
	ConceptosSinCuenta   []string `json:"conceptos_sin_cuenta,omitempty"`
	TotalDevengado       float64  `json:"total_devengado"`
	TotalDeducciones     float64  `json:"total_deducciones"`
	TotalNeto            float64  `json:"total_neto"`
	TotalPagado          float64  `json:"total_pagado"`
	SaldoPendiente       float64  `json:"saldo_pendiente"`
	TotalIBC             float64  `json:"total_ibc"`
	CostoEmpresaEstimado float64  `json:"costo_empresa_estimado"`
	TotalAportesPILA     float64  `json:"total_aportes_pila"`
	Bloqueos             []string `json:"bloqueos,omitempty"`
	Alertas              []string `json:"alertas,omitempty"`
}

func writeNominaControlConflict(w http.ResponseWriter, message string, control *empresaNominaControlContable) {
	writeJSON(w, http.StatusConflict, map[string]interface{}{
		"error":                 message,
		"requiere_confirmacion": control != nil && control.RequiereConfirmacion,
		"control_contable":      control,
	})
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

	if control, err := buildEmpresaNominaControlContable(dbEmp, empresaID, out.PeriodoDesde, out.PeriodoHasta, empleadoNominaID); err == nil {
		out.ControlContable = control
		for _, item := range control.Bloqueos {
			out.Alertas = append(out.Alertas, item)
		}
		for _, item := range control.Alertas {
			out.Alertas = append(out.Alertas, item)
		}
	}

	return out, nil
}

func buildEmpresaNominaControlContable(dbEmp *sql.DB, empresaID int64, periodoDesde, periodoHasta string, empleadoNominaID int64) (*empresaNominaControlContable, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id es obligatorio")
	}
	desde, hasta, err := normalizeNominaControlPeriodo(periodoDesde, periodoHasta)
	if err != nil {
		return nil, err
	}
	control := &empresaNominaControlContable{
		EmpresaID:          empresaID,
		PeriodoDesde:       desde,
		PeriodoHasta:       hasta,
		EmpleadoNominaID:   empleadoNominaID,
		Estado:             "listo",
		ConceptosSinCuenta: make([]string, 0),
		Bloqueos:           make([]string, 0),
		Alertas:            make([]string, 0),
		PuedeGenerarPagos:  true,
	}

	empleados, err := dbpkg.ListEmpresaNominaEmpleados(dbEmp, empresaID, false, "", 5000)
	if err != nil {
		return nil, err
	}
	if len(empleados) == 0 {
		control.Bloqueos = append(control.Bloqueos, "No hay empleados activos vinculados a nomina.")
	}

	liquidaciones, err := dbpkg.ListEmpresaNominaLiquidaciones(dbEmp, empresaID, dbpkg.EmpresaNominaLiquidacionFilter{
		PeriodoDesde:     desde,
		PeriodoHasta:     hasta,
		EmpleadoNominaID: empleadoNominaID,
		IncludeInactive:  false,
		Limit:            5000,
	})
	if err != nil {
		return nil, err
	}
	control.Liquidaciones = len(liquidaciones)
	for _, item := range liquidaciones {
		control.TotalDevengado += item.DevengadoTotal
		control.TotalDeducciones += item.DeduccionTotal
		control.TotalNeto += item.NetoPagar
		control.TotalIBC += item.IngresoBaseCotizacion
	}
	if control.Liquidaciones == 0 {
		control.Bloqueos = append(control.Bloqueos, "No hay liquidaciones activas para el periodo seleccionado.")
	}

	pagos, err := dbpkg.ListEmpresaNominaPagos(dbEmp, empresaID, dbpkg.EmpresaNominaPagoFilter{
		PeriodoDesde:     desde,
		PeriodoHasta:     hasta,
		EmpleadoNominaID: empleadoNominaID,
		IncludeInactive:  false,
		Limit:            5000,
	})
	if err != nil {
		return nil, err
	}
	pagosPorLiquidacion := make(map[int64]bool, len(pagos))
	control.PagosGenerados = len(pagos)
	for _, pago := range pagos {
		pagosPorLiquidacion[pago.LiquidacionID] = true
		control.TotalPagado += pago.NetoPagado
	}
	for _, liq := range liquidaciones {
		if !pagosPorLiquidacion[liq.ID] {
			control.PendientesPago++
		}
	}
	control.SaldoPendiente = roundNominaControl(control.TotalNeto - control.TotalPagado)
	if control.Liquidaciones > 0 && control.PendientesPago == 0 {
		control.Alertas = append(control.Alertas, "Todas las liquidaciones del periodo ya tienen pago activo registrado.")
	}
	if control.TotalPagado-control.TotalNeto > 0.01 {
		control.Bloqueos = append(control.Bloqueos, "El total pagado supera el neto liquidado del periodo.")
	}
	if control.PagosGenerados > 0 && absNominaControl(control.SaldoPendiente) > 0.01 && control.PendientesPago == 0 {
		control.Alertas = append(control.Alertas, "El saldo pagado no coincide exactamente con el neto liquidado.")
	}

	if provisiones, err := dbpkg.GetEmpresaNominaProvisionesResumen(dbEmp, empresaID, desde, hasta, empleadoNominaID); err == nil && provisiones != nil {
		control.CostoEmpresaEstimado = provisiones.CostoEmpresaEstimado
		if control.TotalIBC <= 0 {
			control.TotalIBC = provisiones.TotalIBC
		}
	}

	conceptos, err := dbpkg.ListEmpresaNominaConceptosColombia(dbEmp, empresaID, "", 500)
	if err == nil {
		for _, concepto := range conceptos {
			if !strings.EqualFold(strings.TrimSpace(concepto.Estado), "activo") {
				continue
			}
			if !concepto.AfectaPILA && !concepto.AfectaNominaElectronica {
				continue
			}
			if strings.TrimSpace(concepto.CuentaContable) == "" {
				control.ConceptosSinCuenta = append(control.ConceptosSinCuenta, strings.TrimSpace(concepto.Codigo+" - "+concepto.Nombre))
			}
		}
		if len(control.ConceptosSinCuenta) > 0 {
			control.Alertas = append(control.Alertas, "Hay conceptos de nomina Colombia sin cuenta contable configurada.")
		}
	}

	novedades, err := dbpkg.ListEmpresaNominaNovedadesColombia(dbEmp, empresaID, desde, hasta, "pendiente", 500)
	if err == nil {
		control.NovedadesPendientes = len(novedades)
		if control.NovedadesPendientes > 0 {
			control.Bloqueos = append(control.Bloqueos, "Existen novedades Colombia pendientes de aprobacion en el periodo.")
		}
	}

	periodoPILA := nominaControlPeriodoPILA(desde)
	pila, err := dbpkg.ListEmpresaNominaPILAResumenColombia(dbEmp, empresaID, periodoPILA, 2000)
	if err == nil {
		control.RegistrosPILA = len(pila)
		for _, row := range pila {
			control.TotalAportesPILA += row.TotalAportes
		}
		if control.Liquidaciones > 0 && control.RegistrosPILA == 0 {
			control.Alertas = append(control.Alertas, "No se ha generado resumen PILA para el mes del periodo.")
		}
	}

	control.PuedeGenerarPagos = len(control.Bloqueos) == 0 && control.Liquidaciones > 0 && control.PendientesPago > 0
	control.RequiereConfirmacion = len(control.Alertas) > 0
	switch {
	case len(control.Bloqueos) > 0:
		control.Estado = "bloqueado"
	case len(control.Alertas) > 0:
		control.Estado = "advertencia"
	default:
		control.Estado = "listo"
	}
	control.TotalDevengado = roundNominaControl(control.TotalDevengado)
	control.TotalDeducciones = roundNominaControl(control.TotalDeducciones)
	control.TotalNeto = roundNominaControl(control.TotalNeto)
	control.TotalPagado = roundNominaControl(control.TotalPagado)
	control.CostoEmpresaEstimado = roundNominaControl(control.CostoEmpresaEstimado)
	control.TotalIBC = roundNominaControl(control.TotalIBC)
	control.TotalAportesPILA = roundNominaControl(control.TotalAportesPILA)
	return control, nil
}

func normalizeNominaControlPeriodo(periodoDesde, periodoHasta string) (string, string, error) {
	desde := strings.TrimSpace(periodoDesde)
	hasta := strings.TrimSpace(periodoHasta)
	if desde == "" || hasta == "" {
		return "", "", errors.New("periodo_desde y periodo_hasta son obligatorios")
	}
	start, err := time.Parse("2006-01-02", desde)
	if err != nil {
		return "", "", errors.New("periodo_desde invalido (use YYYY-MM-DD)")
	}
	end, err := time.Parse("2006-01-02", hasta)
	if err != nil {
		return "", "", errors.New("periodo_hasta invalido (use YYYY-MM-DD)")
	}
	if end.Before(start) {
		return "", "", errors.New("periodo_hasta no puede ser menor a periodo_desde")
	}
	return start.Format("2006-01-02"), end.Format("2006-01-02"), nil
}

func nominaControlPeriodoPILA(periodoDesde string) string {
	if len(periodoDesde) >= 7 {
		return periodoDesde[:7]
	}
	return strings.TrimSpace(periodoDesde)
}

func roundNominaControl(v float64) float64 {
	if v < 0 {
		return -roundNominaControl(-v)
	}
	return float64(int64(v*100+0.5)) / 100
}

func absNominaControl(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
