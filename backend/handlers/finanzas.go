package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func buildTableroResumenExportPayload(resumen *dbpkg.EmpresaReportesTableroResumen) map[string]interface{} {
	if resumen == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"empresa_id":        resumen.EmpresaID,
		"desde":             resumen.Desde,
		"hasta":             resumen.Hasta,
		"generado_en":       resumen.GeneradoEn,
		"operativo":         resumen.Operativo,
		"financiero":        resumen.Financiero,
		"contable":          resumen.Contable,
		"estado_resultados": resumen.EstadoResultados,
		"balance_general":   resumen.BalanceGeneral,
	}
}

func formatTableroMetricFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func buildTableroResumenCSVRows(resumen *dbpkg.EmpresaReportesTableroResumen) [][]string {
	rows := [][]string{}
	if resumen == nil {
		return rows
	}

	empresaID := strconv.FormatInt(resumen.EmpresaID, 10)
	addRow := func(bloque, metrica, valor string) {
		rows = append(rows, []string{empresaID, resumen.Desde, resumen.Hasta, resumen.GeneradoEn, bloque, metrica, valor})
	}

	addRow("operativo", "ventas_cerradas", strconv.FormatInt(resumen.Operativo.VentasCerradas, 10))
	addRow("operativo", "ventas_hoy", strconv.FormatInt(resumen.Operativo.VentasHoy, 10))
	addRow("operativo", "ingresos_ventas", formatTableroMetricFloat(resumen.Operativo.IngresosVentas))
	addRow("operativo", "ticket_promedio", formatTableroMetricFloat(resumen.Operativo.TicketPromedio))
	addRow("operativo", "clientes_activos", strconv.FormatInt(resumen.Operativo.ClientesActivos, 10))
	addRow("operativo", "productos_activos", strconv.FormatInt(resumen.Operativo.ProductosActivos, 10))
	addRow("operativo", "productos_bajo_minimo", strconv.FormatInt(resumen.Operativo.ProductosBajoMinimo, 10))
	addRow("operativo", "compras_movimientos", strconv.FormatInt(resumen.Operativo.ComprasMovimientos, 10))
	addRow("operativo", "compras_costo", formatTableroMetricFloat(resumen.Operativo.ComprasCosto))

	addRow("financiero", "movimientos_ingresos", strconv.FormatInt(resumen.Financiero.MovimientosIngresos, 10))
	addRow("financiero", "movimientos_egresos", strconv.FormatInt(resumen.Financiero.MovimientosEgresos, 10))
	addRow("financiero", "ingresos", formatTableroMetricFloat(resumen.Financiero.Ingresos))
	addRow("financiero", "egresos", formatTableroMetricFloat(resumen.Financiero.Egresos))
	addRow("financiero", "balance", formatTableroMetricFloat(resumen.Financiero.Balance))
	addRow("financiero", "periodos_abiertos", strconv.FormatInt(resumen.Financiero.PeriodosAbiertos, 10))
	addRow("financiero", "periodos_cerrados", strconv.FormatInt(resumen.Financiero.PeriodosCerrados, 10))

	addRow("contable", "eventos_pendientes", strconv.FormatInt(resumen.Contable.EventosPendientes, 10))
	addRow("contable", "eventos_procesados", strconv.FormatInt(resumen.Contable.EventosProcesados, 10))
	addRow("contable", "eventos_total", strconv.FormatInt(resumen.Contable.EventosTotal, 10))
	addRow("contable", "eventos_monto_total", formatTableroMetricFloat(resumen.Contable.EventosMontoTotal))
	addRow("contable", "asientos_generados", strconv.FormatInt(resumen.Contable.AsientosGenerados, 10))
	addRow("contable", "asientos_monto_total", formatTableroMetricFloat(resumen.Contable.AsientosMontoTotal))
	addRow("contable", "documentos_facturacion_activos", strconv.FormatInt(resumen.Contable.DocumentosFacturacionActivos, 10))
	addRow("contable", "documentos_compras_activos", strconv.FormatInt(resumen.Contable.DocumentosComprasActivos, 10))

	addRow("estado_resultados", "ingresos", formatTableroMetricFloat(resumen.EstadoResultados.Ingresos))
	addRow("estado_resultados", "gastos", formatTableroMetricFloat(resumen.EstadoResultados.Gastos))
	addRow("estado_resultados", "utilidad_operacional", formatTableroMetricFloat(resumen.EstadoResultados.UtilidadOperacional))

	addRow("balance_general", "activos", formatTableroMetricFloat(resumen.BalanceGeneral.Activos))
	addRow("balance_general", "pasivos", formatTableroMetricFloat(resumen.BalanceGeneral.Pasivos))
	addRow("balance_general", "patrimonio", formatTableroMetricFloat(resumen.BalanceGeneral.Patrimonio))
	addRow("balance_general", "resultado_ejercicio", formatTableroMetricFloat(resumen.BalanceGeneral.ResultadoEjercicio))
	addRow("balance_general", "cuadre", formatTableroMetricFloat(resumen.BalanceGeneral.Cuadre))

	return rows
}

func buildTableroResumenCSVContent(resumen *dbpkg.EmpresaReportesTableroResumen) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	if err := writer.Write([]string{"empresa_id", "desde", "hasta", "generado_en", "bloque", "metrica", "valor"}); err != nil {
		return "", err
	}
	for _, row := range buildTableroResumenCSVRows(resumen) {
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}
	return builder.String(), nil
}

// EmpresaFinanzasMovimientosHandler gestiona CRUD de ingresos/egresos por empresa.
func EmpresaFinanzasMovimientosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "tablero_export" || action == "tablero_exportar" || action == "export_tablero" {
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				resumen, err := dbpkg.GetEmpresaReportesTableroResumen(dbEmp, empresaID, desde, hasta)
				if err != nil {
					http.Error(w, "No se pudo construir el tablero de reportes", http.StatusInternalServerError)
					return
				}

				format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
				if format == "" {
					format = "json"
				}
				fileNameBase := "tablero_empresa_" + strconv.FormatInt(empresaID, 10) + "_" + time.Now().Format("20060102_150405")

				switch format {
				case "json":
					w.Header().Set("Content-Disposition", "attachment; filename=\""+fileNameBase+".json\"")
					writeJSON(w, http.StatusOK, buildTableroResumenExportPayload(resumen))
					return
				case "csv":
					content, err := buildTableroResumenCSVContent(resumen)
					if err != nil {
						http.Error(w, "No se pudo generar la exportacion CSV del tablero", http.StatusInternalServerError)
						return
					}
					w.Header().Set("Content-Type", "text/csv; charset=utf-8")
					w.Header().Set("Content-Disposition", "attachment; filename=\""+fileNameBase+".csv\"")
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte(content))
					return
				default:
					http.Error(w, "format invalido (use csv o json)", http.StatusBadRequest)
					return
				}
			}
			if action == "tablero" || action == "dashboard" || action == "resumen_kpi" {
				desde := strings.TrimSpace(r.URL.Query().Get("desde"))
				hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
				resumen, err := dbpkg.GetEmpresaReportesTableroResumen(dbEmp, empresaID, desde, hasta)
				if err != nil {
					http.Error(w, "No se pudo construir el tablero de reportes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			tipo := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("tipo")))
			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			rows, err := dbpkg.ListEmpresaFinanzasMovimientos(dbEmp, empresaID, dbpkg.EmpresaFinanzasMovimientoFilter{
				Tipo:            tipo,
				Desde:           desde,
				Hasta:           hasta,
				Periodo:         periodo,
				Q:               q,
				IncludeInactive: includeInactive,
				Limit:           limit,
			})
			if err != nil {
				http.Error(w, "No se pudieron listar los movimientos financieros", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaFinanzasMovimiento
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
			id, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, payload)
			if err != nil {
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			evento := "movimiento_ingreso_registrado"
			if strings.EqualFold(strings.TrimSpace(payload.TipoMovimiento), "egreso") {
				evento = "movimiento_egreso_registrado"
			}
			montoEvento := payload.Total
			if montoEvento <= 0 {
				montoEvento = payload.Monto
			}
			registrarEventoContableNoBloqueante(dbEmp, r, "finanzas", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "finanzas",
				Evento:          evento,
				Entidad:         "finanzas_movimiento",
				EntidadID:       id,
				DocumentoTipo:   strings.TrimSpace(payload.TipoComprobante),
				DocumentoCodigo: strings.TrimSpace(payload.Codigo),
				PeriodoContable: strings.TrimSpace(payload.PeriodoContable),
				MontoTotal:      montoEvento,
				Moneda:          strings.TrimSpace(payload.Moneda),
				Origen:          "api_finanzas_movimientos",
				Observaciones:   "movimiento financiero registrado desde API",
			}, map[string]interface{}{
				"tipo_movimiento":  strings.ToLower(strings.TrimSpace(payload.TipoMovimiento)),
				"concepto":         strings.TrimSpace(payload.Concepto),
				"categoria":        strings.TrimSpace(payload.Categoria),
				"periodo_contable": strings.TrimSpace(payload.PeriodoContable),
				"empresa_id":       payload.EmpresaID,
			})
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" || action == "anular" {
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
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if action == "anular" {
					estado = "anulado"
				}
				if err := dbpkg.SetEmpresaFinanzasMovimientoEstado(dbEmp, empresaID, id, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "movimiento no encontrado", http.StatusNotFound)
						return
					}
					if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
						http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
						return
					}
					http.Error(w, "No se pudo actualizar el estado", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.EmpresaFinanzasMovimiento
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "id y empresa_id son obligatorios", http.StatusBadRequest)
				return
			}
			if payload.UsuarioCreador == "" {
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			}
			if err := dbpkg.UpdateEmpresaFinanzasMovimiento(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "movimiento no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

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
			if err := dbpkg.DeleteEmpresaFinanzasMovimiento(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "movimiento no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrPeriodoFinancieroCerrado) {
					http.Error(w, "el periodo contable del movimiento esta cerrado", http.StatusConflict)
					return
				}
				http.Error(w, "No se pudo eliminar el movimiento", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasConfiguracionHandler gestiona configuracion por empresa del modulo financiero.
func EmpresaFinanzasConfiguracionHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetEmpresaFinanzasConfiguracion(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "No se pudo consultar la configuracion financiera", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, cfg)
			return

		case http.MethodPost, http.MethodPut:
			var payload dbpkg.EmpresaFinanzasConfiguracion
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
			id, err := dbpkg.UpsertEmpresaFinanzasConfiguracion(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar la configuracion financiera", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasPeriodosHandler gestiona periodos contables por empresa.
func EmpresaFinanzasPeriodosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			rows, err := dbpkg.ListEmpresaFinanzasPeriodos(dbEmp, empresaID, includeInactive)
			if err != nil {
				http.Error(w, "No se pudieron listar los periodos", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaFinanzasPeriodo
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
			id, err := dbpkg.UpsertEmpresaFinanzasPeriodo(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "cerrar" || action == "reabrir" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))
				if periodo == "" {
					http.Error(w, "periodo es obligatorio", http.StatusBadRequest)
					return
				}
				estado := "cerrado"
				if action == "reabrir" {
					estado = "abierto"
				}
				if err := dbpkg.SetEmpresaFinanzasPeriodoEstado(dbEmp, empresaID, periodo, estado, strings.TrimSpace(adminEmailFromRequest(r)), "actualizacion desde API"); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				evento := "periodo_contable_cerrado"
				if estado == "abierto" {
					evento = "periodo_contable_reabierto"
				}
				registrarEventoContableNoBloqueante(dbEmp, r, "finanzas", dbpkg.EmpresaEventoContable{
					EmpresaID:       empresaID,
					Modulo:          "finanzas",
					Evento:          evento,
					Entidad:         "finanzas_periodo",
					DocumentoTipo:   "periodo_contable",
					DocumentoCodigo: periodo,
					PeriodoContable: periodo,
					Origen:          "api_finanzas_periodos",
					Observaciones:   "actualizacion de estado de periodo contable",
				}, map[string]interface{}{
					"periodo":    periodo,
					"estado":     estado,
					"empresa_id": empresaID,
				})
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "periodo": periodo, "estado": estado})
				return
			}

			var payload dbpkg.EmpresaFinanzasPeriodo
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
			id, err := dbpkg.UpsertEmpresaFinanzasPeriodo(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasCierresCajaHandler gestiona apertura/arqueo/cierre de caja por empresa/sucursal.
func EmpresaFinanzasCierresCajaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			sucursalID, err := parseInt64QueryOptional(r, "sucursal_id")
			if err != nil {
				http.Error(w, "sucursal_id invalido", http.StatusBadRequest)
				return
			}
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			rows, err := dbpkg.ListEmpresaCierresCaja(dbEmp, empresaID, dbpkg.EmpresaCierreCajaFilter{
				SucursalID:      sucursalID,
				CajaCodigo:      strings.TrimSpace(r.URL.Query().Get("caja_codigo")),
				EstadoCierre:    strings.TrimSpace(r.URL.Query().Get("estado_cierre")),
				Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
				Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
				IncludeInactive: queryBool(r, "include_inactive"),
				Limit:           limit,
			})
			if err != nil {
				http.Error(w, "No se pudieron listar los cierres de caja", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaCierreCaja
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
			id, err := dbpkg.CreateEmpresaCierreCaja(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "cerrar" || action == "reabrir" || action == "aprobar" || action == "anular" || action == "activar" || action == "desactivar" {
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

				if action == "activar" || action == "desactivar" {
					estado := "activo"
					if action == "desactivar" {
						estado = "inactivo"
					}
					if err := dbpkg.SetEmpresaCierreCajaRegistroEstado(dbEmp, empresaID, id, estado); err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							http.Error(w, "cierre de caja no encontrado", http.StatusNotFound)
							return
						}
						http.Error(w, "No se pudo actualizar el estado del registro", http.StatusInternalServerError)
						return
					}
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
					return
				}

				estadoCierre := "cerrado"
				switch action {
				case "reabrir":
					estadoCierre = "abierto"
				case "aprobar":
					estadoCierre = "aprobado"
				case "anular":
					estadoCierre = "anulado"
				}

				var cajaFisica *float64
				if raw := strings.TrimSpace(r.URL.Query().Get("caja_fisica")); raw != "" {
					v, err := strconv.ParseFloat(raw, 64)
					if err != nil {
						http.Error(w, "caja_fisica invalida", http.StatusBadRequest)
						return
					}
					if v < 0 {
						v = 0
					}
					cajaFisica = &v
				}

				usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))

				if err := dbpkg.SetEmpresaCierreCajaEstado(
					dbEmp,
					empresaID,
					id,
					estadoCierre,
					cajaFisica,
					usuarioOperacion,
					strings.TrimSpace(r.URL.Query().Get("observaciones")),
				); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "cierre de caja no encontrado", http.StatusNotFound)
						return
					}
					if errors.Is(err, dbpkg.ErrCierreCajaTransicionInvalida) || errors.Is(err, dbpkg.ErrCierreCajaAprobadoBloqueado) {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}

				resp := map[string]interface{}{"ok": true, "estado_cierre": estadoCierre}
				if estadoCierre == "cerrado" || estadoCierre == "aprobado" {
					conciliacion, err := dbpkg.ConciliarEmpresaPropinasConCierreCaja(dbEmp, empresaID, id, usuarioOperacion)
					if err != nil {
						http.Error(w, "No se pudo conciliar propinas para el cierre de caja", http.StatusInternalServerError)
						return
					}
					resp["conciliacion_propinas"] = conciliacion
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}

			var payload dbpkg.EmpresaCierreCaja
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "id y empresa_id son obligatorios", http.StatusBadRequest)
				return
			}
			if payload.UsuarioCreador == "" {
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			}
			if err := dbpkg.UpdateEmpresaCierreCaja(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "cierre de caja no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrCierreCajaAprobadoBloqueado) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

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
			if err := dbpkg.DeleteEmpresaCierreCaja(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "cierre de caja no encontrado", http.StatusNotFound)
					return
				}
				if errors.Is(err, dbpkg.ErrCierreCajaAprobadoBloqueado) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, "No se pudo eliminar el cierre de caja", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaFinanzasAsientosContablesHandler procesa eventos contables pendientes y consulta asientos canónicos.
func EmpresaFinanzasAsientosContablesHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			if action == "conciliacion_periodo" || action == "conciliacion" || action == "conciliar" {
				resumen, err := dbpkg.GetEmpresaConciliacionContablePorPeriodo(dbEmp, empresaID, dbpkg.EmpresaConciliacionContableFilter{
					Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
					PeriodoContable: strings.TrimSpace(r.URL.Query().Get("periodo")),
					IncludeInactive: queryBool(r, "include_inactive"),
					Limit:           limit,
				})
				if err != nil {
					http.Error(w, "No se pudo construir la conciliacion contable", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return
			}
			rows, err := dbpkg.ListEmpresaAsientosContables(dbEmp, empresaID, dbpkg.EmpresaAsientoContableFilter{
				Modulo:          strings.TrimSpace(r.URL.Query().Get("modulo")),
				Evento:          strings.TrimSpace(r.URL.Query().Get("evento")),
				PeriodoContable: strings.TrimSpace(r.URL.Query().Get("periodo")),
				Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
				Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
				IncludeInactive: queryBool(r, "include_inactive"),
				Limit:           limit,
			})
			if err != nil {
				http.Error(w, "No se pudieron listar los asientos contables", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPut, http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "" {
				action = "procesar_asientos"
			}
			if action != "procesar_asientos" && action != "procesar" {
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}

			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}
			maxRetries, err := parseIntQueryOptional(r, "max_reintentos")
			if err != nil {
				http.Error(w, "max_reintentos invalido", http.StatusBadRequest)
				return
			}

			resultado, err := dbpkg.ProcessEmpresaEventosContablesPendientesConPolitica(dbEmp, empresaID, strings.TrimSpace(adminEmailFromRequest(r)), limit, maxRetries)
			if err != nil {
				http.Error(w, "No se pudieron procesar los eventos contables pendientes", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, resultado)
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
