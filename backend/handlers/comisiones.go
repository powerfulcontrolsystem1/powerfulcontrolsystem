package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaComisionesServicioHandler gestiona configuracion, escalas, ajustes, movimientos y reporte de comisiones por servicio.
func EmpresaComisionesServicioHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				cfg, err := dbpkg.GetEmpresaComisionesServicioConfiguracion(dbEmp, empresaID)
				if err != nil {
					log.Printf("[comisiones] get config empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo consultar la configuracion de comisiones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return

			case "escalas":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaComisionServicioEscalas(
					dbEmp,
					empresaID,
					queryBool(r, "include_inactive"),
					strings.TrimSpace(r.URL.Query().Get("rol_operacion")),
					limit,
				)
				if err != nil {
					log.Printf("[comisiones] list escalas empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo listar escalas de comision", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "reporte", "resumen", "dashboard":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				liquidacionNominaID, err := parseInt64QueryOptional(r, "liquidacion_nomina_id")
				if err != nil {
					http.Error(w, "liquidacion_nomina_id invalido", http.StatusBadRequest)
					return
				}
				report, err := dbpkg.GetEmpresaComisionesServicioReporte(dbEmp, empresaID, dbpkg.EmpresaComisionServicioMovimientoFilter{
					Desde:               strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:               strings.TrimSpace(r.URL.Query().Get("hasta")),
					UsuarioLavador:      strings.TrimSpace(r.URL.Query().Get("usuario_lavador")),
					RolOperacion:        strings.TrimSpace(r.URL.Query().Get("rol_operacion")),
					ServicioFiltro:      strings.TrimSpace(r.URL.Query().Get("servicio_filtro")),
					OrigenMovimiento:    strings.TrimSpace(r.URL.Query().Get("origen")),
					AjusteEstado:        strings.TrimSpace(r.URL.Query().Get("ajuste_estado")),
					LiquidacionNominaID: liquidacionNominaID,
					SoloAjustes:         queryBool(r, "solo_ajustes"),
					SoloPendientes:      queryBool(r, "solo_pendientes"),
					NoLiquidado:         queryBool(r, "no_liquidado"),
					IncludeInactive:     queryBool(r, "include_inactive"),
					Limit:               limit,
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
				liquidacionNominaID, err := parseInt64QueryOptional(r, "liquidacion_nomina_id")
				if err != nil {
					http.Error(w, "liquidacion_nomina_id invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaComisionServicioMovimientos(dbEmp, empresaID, dbpkg.EmpresaComisionServicioMovimientoFilter{
					Desde:               strings.TrimSpace(r.URL.Query().Get("desde")),
					Hasta:               strings.TrimSpace(r.URL.Query().Get("hasta")),
					UsuarioLavador:      strings.TrimSpace(r.URL.Query().Get("usuario_lavador")),
					RolOperacion:        strings.TrimSpace(r.URL.Query().Get("rol_operacion")),
					ServicioFiltro:      strings.TrimSpace(r.URL.Query().Get("servicio_filtro")),
					OrigenMovimiento:    strings.TrimSpace(r.URL.Query().Get("origen")),
					AjusteEstado:        strings.TrimSpace(r.URL.Query().Get("ajuste_estado")),
					LiquidacionNominaID: liquidacionNominaID,
					SoloAjustes:         queryBool(r, "solo_ajustes"),
					SoloPendientes:      queryBool(r, "solo_pendientes"),
					NoLiquidado:         queryBool(r, "no_liquidado"),
					IncludeInactive:     queryBool(r, "include_inactive"),
					Limit:               limit,
				})
				if err != nil {
					log.Printf("[comisiones] list movimientos empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo listar movimientos de comisiones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "resumen_liquidacion":
				periodoDesde := strings.TrimSpace(r.URL.Query().Get("periodo_desde"))
				periodoHasta := strings.TrimSpace(r.URL.Query().Get("periodo_hasta"))
				if periodoDesde == "" {
					periodoDesde = strings.TrimSpace(r.URL.Query().Get("desde"))
				}
				if periodoHasta == "" {
					periodoHasta = strings.TrimSpace(r.URL.Query().Get("hasta"))
				}

				aliases := dbpkg.BuildEmpresaComisionServicioAliases(
					strings.TrimSpace(r.URL.Query().Get("usuario_lavador")),
					strings.TrimSpace(r.URL.Query().Get("empleado_codigo")),
					strings.TrimSpace(r.URL.Query().Get("empleado_documento")),
					strings.TrimSpace(r.URL.Query().Get("empleado_nombre")),
					strings.TrimSpace(r.URL.Query().Get("identificador")),
				)
				resumen, err := dbpkg.GetEmpresaComisionServicioLiquidacionResumen(dbEmp, empresaID, aliases, periodoDesde, periodoHasta)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return

			default:
				http.Error(w, "action invalida. Use: config, escalas, reporte, movimientos o resumen_liquidacion", http.StatusBadRequest)
				return
			}

		case http.MethodPost:
			switch action {
			case "escala":
				var payload dbpkg.EmpresaComisionServicioEscala
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
				id, err := dbpkg.CreateEmpresaComisionServicioEscala(dbEmp, payload)
				if err != nil {
					log.Printf("[comisiones] create escala empresa_id=%d error: %v", payload.EmpresaID, err)
					http.Error(w, err.Error(), comisionesWriteStatus(err))
					return
				}
				escala, err := dbpkg.GetEmpresaComisionServicioEscalaByID(dbEmp, payload.EmpresaID, id)
				if err != nil {
					writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id, "escala": escala})
				return

			case "ajuste_manual":
				var payload struct {
					EmpresaID         int64   `json:"empresa_id"`
					CarritoID         int64   `json:"carrito_id"`
					CarritoItemID     int64   `json:"carrito_item_id"`
					ServicioID        int64   `json:"servicio_id"`
					ServicioCodigo    string  `json:"servicio_codigo"`
					ServicioNombre    string  `json:"servicio_nombre"`
					ServicioCategoria string  `json:"servicio_categoria"`
					UsuarioLavador    string  `json:"usuario_lavador"`
					UsuarioLavadorID  int64   `json:"usuario_lavador_id"`
					RolOperacion      string  `json:"rol_operacion"`
					VentaReferencia   string  `json:"venta_referencia"`
					Moneda            string  `json:"moneda"`
					BaseServicio      float64 `json:"base_servicio"`
					MontoAjuste       float64 `json:"monto_ajuste"`
					MontoComision     float64 `json:"monto_comision"`
					Motivo            string  `json:"motivo"`
					ReferenciaAjuste  string  `json:"referencia_ajuste"`
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

				montoAjuste := payload.MontoAjuste
				if math.Abs(montoAjuste) < 0.0001 {
					montoAjuste = payload.MontoComision
				}
				if math.Abs(montoAjuste) < 0.0001 {
					http.Error(w, "monto_ajuste es obligatorio y debe ser diferente de cero", http.StatusBadRequest)
					return
				}

				motivo := strings.TrimSpace(payload.Motivo)
				if motivo == "" {
					http.Error(w, "motivo es obligatorio para ajuste manual", http.StatusBadRequest)
					return
				}

				usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))
				mov := dbpkg.EmpresaComisionServicioMovimiento{
					EmpresaID:          payload.EmpresaID,
					CarritoID:          payload.CarritoID,
					CarritoItemID:      payload.CarritoItemID,
					ServicioID:         payload.ServicioID,
					ServicioCodigo:     strings.TrimSpace(payload.ServicioCodigo),
					ServicioNombre:     strings.TrimSpace(payload.ServicioNombre),
					ServicioCategoria:  strings.TrimSpace(payload.ServicioCategoria),
					UsuarioOrigen:      usuarioOperacion,
					UsuarioLavador:     strings.TrimSpace(payload.UsuarioLavador),
					UsuarioLavadorID:   payload.UsuarioLavadorID,
					RolOperacion:       strings.TrimSpace(payload.RolOperacion),
					VentaReferencia:    strings.TrimSpace(payload.VentaReferencia),
					Moneda:             strings.TrimSpace(payload.Moneda),
					BaseServicio:       payload.BaseServicio,
					MontoComisionBruto: montoAjuste,
					MontoComision:      montoAjuste,
					ReferenciaAjuste:   strings.TrimSpace(payload.ReferenciaAjuste),
					UsuarioCreador:     usuarioOperacion,
					Estado:             "pendiente",
					Observaciones:      motivo,
				}

				id, err := dbpkg.CreateEmpresaComisionServicioAjusteManual(dbEmp, mov)
				if err != nil {
					log.Printf("[comisiones] ajuste manual empresa_id=%d error: %v", payload.EmpresaID, err)
					http.Error(w, err.Error(), comisionesWriteStatus(err))
					return
				}

				registrarAuditoriaComisionAjusteNoBloqueante(dbEmp, r, payload.EmpresaID, id, "ajuste_manual_solicitado", montoAjuste, motivo)

				item, _ := dbpkg.GetEmpresaComisionServicioMovimientoByID(dbEmp, payload.EmpresaID, id)
				writeJSON(w, http.StatusCreated, map[string]interface{}{
					"ok":            true,
					"id":            id,
					"empresa_id":    payload.EmpresaID,
					"monto_ajuste":  montoAjuste,
					"estado_ajuste": dbpkg.EmpresaComisionServicioAjustePendiente,
					"movimiento":    item,
				})
				return

			default:
				http.Error(w, "action invalida. Use: escala o ajuste_manual", http.StatusBadRequest)
				return
			}

		case http.MethodPut:
			switch action {
			case "escala":
				var payload dbpkg.EmpresaComisionServicioEscala
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.ID <= 0 {
					if escalaID, err := parseInt64QueryOptional(r, "escala_id"); err == nil && escalaID > 0 {
						payload.ID = escalaID
					}
				}
				if payload.EmpresaID <= 0 || payload.ID <= 0 {
					http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
					return
				}
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpdateEmpresaComisionServicioEscala(dbEmp, payload)
				if err != nil {
					log.Printf("[comisiones] update escala empresa_id=%d escala_id=%d error: %v", payload.EmpresaID, payload.ID, err)
					http.Error(w, err.Error(), comisionesWriteStatus(err))
					return
				}
				escala, err := dbpkg.GetEmpresaComisionServicioEscalaByID(dbEmp, payload.EmpresaID, id)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "escala": escala})
				return

			case "activar_escala", "desactivar_escala":
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				escalaID, err := parseInt64QueryOptional(r, "escala_id")
				if err != nil || escalaID <= 0 {
					http.Error(w, "escala_id es obligatorio", http.StatusBadRequest)
					return
				}
				activo := action == "activar_escala"
				usuario := strings.TrimSpace(adminEmailFromRequest(r))
				_, err = dbpkg.SetEmpresaComisionServicioEscalaEstado(dbEmp, empresaID, escalaID, activo, usuario, "")
				if err != nil {
					http.Error(w, err.Error(), comisionesWriteStatus(err))
					return
				}
				escala, _ := dbpkg.GetEmpresaComisionServicioEscalaByID(dbEmp, empresaID, escalaID)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": escalaID, "escala": escala})
				return

			case "aprobar_ajuste", "rechazar_ajuste":
				var payload struct {
					EmpresaID     int64  `json:"empresa_id"`
					MovimientoID  int64  `json:"movimiento_id"`
					Observaciones string `json:"observaciones"`
				}
				if r.Body != nil {
					_ = json.NewDecoder(r.Body).Decode(&payload)
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.MovimientoID <= 0 {
					if movimientoID, err := parseInt64QueryOptional(r, "movimiento_id"); err == nil && movimientoID > 0 {
						payload.MovimientoID = movimientoID
					}
				}
				if payload.EmpresaID <= 0 || payload.MovimientoID <= 0 {
					http.Error(w, "empresa_id y movimiento_id son obligatorios", http.StatusBadRequest)
					return
				}

				aprobar := action == "aprobar_ajuste"
				usuario := strings.TrimSpace(adminEmailFromRequest(r))
				item, err := dbpkg.ResolverEmpresaComisionServicioAjusteManual(dbEmp, payload.EmpresaID, payload.MovimientoID, aprobar, usuario, payload.Observaciones)
				if err != nil {
					log.Printf("[comisiones] resolver ajuste empresa_id=%d movimiento_id=%d error: %v", payload.EmpresaID, payload.MovimientoID, err)
					http.Error(w, err.Error(), comisionesWriteStatus(err))
					return
				}

				auditAction := "ajuste_manual_rechazado"
				if aprobar {
					auditAction = "ajuste_manual_aprobado"
				}
				registrarAuditoriaComisionAjusteNoBloqueante(dbEmp, r, payload.EmpresaID, payload.MovimientoID, auditAction, item.MontoComision, item.Observaciones)

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":         true,
					"id":         payload.MovimientoID,
					"movimiento": item,
				})
				return

			default:
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

		case http.MethodDelete:
			if action != "escala" && action != "desactivar_escala" {
				http.Error(w, "action invalida. Use: escala", http.StatusBadRequest)
				return
			}
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			escalaID, err := parseInt64QueryOptional(r, "escala_id")
			if err != nil || escalaID <= 0 {
				http.Error(w, "escala_id es obligatorio", http.StatusBadRequest)
				return
			}
			usuario := strings.TrimSpace(adminEmailFromRequest(r))
			_, err = dbpkg.SetEmpresaComisionServicioEscalaEstado(dbEmp, empresaID, escalaID, false, usuario, "desactivada por solicitud")
			if err != nil {
				http.Error(w, err.Error(), comisionesWriteStatus(err))
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": escalaID, "estado": "inactivo"})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func comisionesWriteStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound
	}
	lower := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(lower, "obligatorio") ||
		strings.Contains(lower, "invalido") ||
		strings.Contains(lower, "debe") ||
		strings.Contains(lower, "procesado") {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

func registrarAuditoriaComisionAjusteNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID, recursoID int64, accion string, monto float64, observaciones string) {
	if dbEmp == nil || empresaID <= 0 || recursoID <= 0 {
		return
	}
	metadata, _ := json.Marshal(map[string]interface{}{
		"monto_ajuste": monto,
		"accion":       accion,
	})
	usuario := strings.TrimSpace(adminEmailFromRequest(r))
	if usuario == "" {
		usuario = "sistema"
	}
	requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
	if requestID == "" {
		requestID = strings.TrimSpace(r.Header.Get("X-Request-Id"))
	}
	if _, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "comisiones",
		Accion:         accion,
		Recurso:        "empresa_comisiones_servicio_movimientos",
		RecursoID:      recursoID,
		MetodoHTTP:     r.Method,
		Endpoint:       r.URL.Path,
		Resultado:      "ok",
		CodigoHTTP:     http.StatusOK,
		RequestID:      requestID,
		IPOrigen:       strings.TrimSpace(r.RemoteAddr),
		UserAgent:      strings.TrimSpace(r.UserAgent()),
		MetadataJSON:   string(metadata),
		UsuarioCreador: usuario,
		Observaciones:  strings.TrimSpace(observaciones),
	}); err != nil {
		log.Printf("[comisiones] auditoria ajuste empresa_id=%d recurso_id=%d error: %v", empresaID, recursoID, err)
	}
}
