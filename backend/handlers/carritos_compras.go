package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type carritoPagoMixtoEntrada struct {
	Metodo     string  `json:"metodo"`
	Monto      float64 `json:"monto"`
	Referencia string  `json:"referencia"`
}

type carritoPagoMixtoNormalizado struct {
	Metodo     string
	Monto      float64
	Referencia string
}

type carritoCreditoVentaResultado struct {
	Aplica           bool    `json:"aplica"`
	CreditoID        int64   `json:"credito_id,omitempty"`
	Codigo           string  `json:"codigo,omitempty"`
	ClienteID        int64   `json:"cliente_id,omitempty"`
	ClienteNombre    string  `json:"cliente_nombre,omitempty"`
	MontoCredito     float64 `json:"monto_credito,omitempty"`
	CupoLimite       float64 `json:"cupo_limite,omitempty"`
	SaldoPrevio      float64 `json:"saldo_previo,omitempty"`
	SaldoDisponible  float64 `json:"saldo_disponible,omitempty"`
	SaldoDespues     float64 `json:"saldo_despues,omitempty"`
	FechaVencimiento string  `json:"fecha_vencimiento,omitempty"`
	Warning          string  `json:"warning,omitempty"`
}

type carritoBusinessPrerequisite struct {
	OK           bool                     `json:"ok"`
	Code         string                   `json:"code"`
	Title        string                   `json:"title"`
	Message      string                   `json:"message"`
	RobotMessage string                   `json:"robot_message"`
	Scope        string                   `json:"scope"`
	Steps        []string                 `json:"steps"`
	Missing      []string                 `json:"missing,omitempty"`
	Actions      []map[string]interface{} `json:"actions,omitempty"`
}

func effectiveAdminRoleFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if v := r.Context().Value("adminRoleEfectivo"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	if h := strings.TrimSpace(r.Header.Get("X-Admin-Role-Efectivo")); h != "" {
		return h
	}
	return strings.TrimSpace(adminRoleFromRequest(r))
}

func isPorteroCarritoRequest(r *http.Request) bool {
	return normalizePermissionRole(effectiveAdminRoleFromRequest(r)) == "portero"
}

func isServicioLimpiezaCarritoRequest(r *http.Request) bool {
	return normalizePermissionRole(effectiveAdminRoleFromRequest(r)) == "servicio_limpieza"
}

func isStationBoardOnlyCarritoRequest(r *http.Request) bool {
	role := normalizePermissionRole(effectiveAdminRoleFromRequest(r))
	return role == "portero" || role == "servicio_limpieza"
}

func isPorteroRestrictedCarritoRequest(r *http.Request, action string) bool {
	if !isPorteroCarritoRequest(r) {
		return false
	}
	if r.Method == http.MethodGet {
		return strings.TrimSpace(action) != ""
	}
	return !(r.Method == http.MethodPut && strings.EqualFold(strings.TrimSpace(action), "activar_estacion"))
}

func isServicioLimpiezaRestrictedCarritoRequest(r *http.Request, action string) bool {
	if !isServicioLimpiezaCarritoRequest(r) {
		return false
	}
	return !(r.Method == http.MethodGet && strings.TrimSpace(action) == "")
}

// EmpresaCarritosCompraHandler gestiona CRUD de carritos por empresa.
func EmpresaCarritosCompraHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if isPorteroRestrictedCarritoRequest(r, action) {
				http.Error(w, "forbidden: el rol portero solo puede ver y activar estaciones", http.StatusForbidden)
				return
			}
			if isServicioLimpiezaRestrictedCarritoRequest(r, action) {
				http.Error(w, "forbidden: el rol Servicio de limpieza solo puede ver estaciones y reportar aseo", http.StatusForbidden)
				return
			}
			if action == "totales_pago" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				usuarioActual := strings.TrimSpace(adminEmailFromRequest(r))
				estacionID, err := parseOptionalInt64CarritoQuery(r, "estacion_id")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if estacionID > 0 {
					if err := ensureCarritoStationAccessForStation(dbEmp, empresaID, usuarioActual, estacionID); err != nil {
						writeCarritoStationAccessError(w, err)
						return
					}
				}
				days, err := parseOptionalIntCarritoQuery(r, "days")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if days <= 0 {
					days = 7
				}

				query := `SELECT COALESCE(LOWER(TRIM(metodo_pago)),'efectivo') AS metodo_pago, COALESCE(SUM(COALESCE(monto_pagado,0)),0) AS total_pagado
					FROM empresa_ventas_estacion_metricas
					WHERE empresa_id = ?
						AND COALESCE(estado,'activo') = 'activo'
						AND evento_operacion = 'venta_pagada'
						AND pcs_ts(COALESCE(fecha_evento, fecha_creacion, CURRENT_TIMESTAMP)) >= pcs_ts('now','localtime', ?)`
				args := []interface{}{empresaID, fmt.Sprintf("-%d day", days)}
				if usuarioActual != "" {
					query += " AND LOWER(COALESCE(usuario_creador,'')) = LOWER(?)"
					args = append(args, usuarioActual)
				}
				if estacionID > 0 {
					query += " AND estacion_id = ?"
					args = append(args, estacionID)
				}
				query += " GROUP BY metodo_pago"

				rows, err := dbEmp.Query(query, args...)
				if err != nil {
					log.Printf("[carritos] totales_pago empresa_id=%d estacion_id=%d error: %v", empresaID, estacionID, err)
					http.Error(w, "No se pudieron consultar totales de pago", http.StatusInternalServerError)
					return
				}
				defer rows.Close()

				totals := map[string]float64{
					"efectivo":               0.0,
					"tarjeta_debito":         0.0,
					"tarjeta_credito":        0.0,
					"transferencia_bancaria": 0.0,
					"transferencia_bre_b":    0.0,
					"transferencia_nequi":    0.0,
					"transferencia_otro":     0.0,
					"credito_cliente":        0.0,
				}
				for rows.Next() {
					var metodo string
					var total float64
					if err := rows.Scan(&metodo, &total); err != nil {
						continue
					}
					metodo = strings.TrimSpace(strings.ToLower(metodo))
					totals[metodo] = roundMoneyCarritoHandler(total)
				}
				if err := rows.Err(); err != nil {
					log.Printf("[carritos] totales_pago rows error empresa_id=%d estacion_id=%d error: %v", empresaID, estacionID, err)
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"totales": totals,
					"filtros": map[string]interface{}{"empresa_id": empresaID, "estacion_id": estacionID, "days": days},
				})
				return
			}
			if action == "abonos" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				carritoID, err := parseInt64Query(r, "carrito_id")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				includeInactive := strings.TrimSpace(r.URL.Query().Get("include_inactive")) == "1"
				rows, err := dbpkg.ListCarritoCompraAbonos(dbEmp, empresaID, carritoID, includeInactive)
				if err != nil {
					log.Printf("[carritos] listar abonos empresa_id=%d carrito_id=%d error: %v", empresaID, carritoID, err)
					http.Error(w, "No se pudieron consultar los abonos del carrito", http.StatusInternalServerError)
					return
				}
				total := 0.0
				for _, row := range rows {
					if strings.TrimSpace(strings.ToLower(row.Estado)) == "activo" || strings.TrimSpace(row.Estado) == "" {
						total = roundMoneyCarritoHandler(total + row.Monto)
					}
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":     true,
					"abonos": rows,
					"total":  total,
				})
				return
			}
			if action == "metricas_estacion" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				estacionID, err := parseOptionalInt64CarritoQuery(r, "estacion_id")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				days, err := parseOptionalIntCarritoQuery(r, "days")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if days <= 0 {
					days = 7
				}
				limit, err := parseOptionalIntCarritoQuery(r, "limit")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if limit <= 0 {
					limit = 10
				}

				rows, err := dbpkg.ListCarritoStationMetricSummary(dbEmp, empresaID, estacionID, days, limit)
				if err != nil {
					log.Printf("[carritos] metricas_estacion empresa_id=%d estacion_id=%d error: %v", empresaID, estacionID, err)
					http.Error(w, "No se pudieron consultar metricas de estacion", http.StatusInternalServerError)
					return
				}

				resumen := map[string]interface{}{
					"estaciones":               len(rows),
					"ventas_pagadas":           int64(0),
					"correcciones":             int64(0),
					"monto_vendido":            0.0,
					"monto_pagado":             0.0,
					"monto_anulado":            0.0,
					"devolucion_total":         0.0,
					"tiempo_promedio_segundos": 0.0,
				}
				totalTiempoPonderado := 0.0
				totalVentasPonderadas := int64(0)
				for _, row := range rows {
					resumen["ventas_pagadas"] = resumen["ventas_pagadas"].(int64) + row.VentasPagadas
					resumen["correcciones"] = resumen["correcciones"].(int64) + row.Correcciones
					resumen["monto_vendido"] = roundMoneyCarritoHandler(resumen["monto_vendido"].(float64) + row.MontoVendido)
					resumen["monto_pagado"] = roundMoneyCarritoHandler(resumen["monto_pagado"].(float64) + row.MontoPagado)
					resumen["monto_anulado"] = roundMoneyCarritoHandler(resumen["monto_anulado"].(float64) + row.MontoAnulado)
					resumen["devolucion_total"] = roundMoneyCarritoHandler(resumen["devolucion_total"].(float64) + row.DevolucionTotal)
					if row.VentasPagadas > 0 && row.TiempoPromedioSegundos > 0 {
						totalTiempoPonderado += row.TiempoPromedioSegundos * float64(row.VentasPagadas)
						totalVentasPonderadas += row.VentasPagadas
					}
				}
				if totalVentasPonderadas > 0 {
					resumen["tiempo_promedio_segundos"] = roundMoneyCarritoHandler(totalTiempoPonderado / float64(totalVentasPonderadas))
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"rows":    rows,
					"resumen": resumen,
					"filtros": map[string]interface{}{
						"empresa_id":  empresaID,
						"estacion_id": estacionID,
						"days":        days,
						"limit":       limit,
					},
				})
				return
			}
			if action == "cajas_abiertas" || action == "cajas_abiertas_estacion" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				limit, err := parseOptionalIntCarritoQuery(r, "limit")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if limit <= 0 || limit > 100 {
					limit = 100
				}
				sucursalID, err := parseOptionalInt64CarritoQuery(r, "sucursal_id")
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaCierresCaja(dbEmp, empresaID, dbpkg.EmpresaCierreCajaFilter{
					SucursalID:     sucursalID,
					CajaCodigo:     strings.TrimSpace(r.URL.Query().Get("caja_codigo")),
					UsuarioCreador: strings.TrimSpace(adminEmailFromRequest(r)),
					EstadoCierre:   "abierto",
					Limit:          limit,
				})
				if err != nil {
					log.Printf("[carritos] cajas_abiertas empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron listar las cajas abiertas para cobrar", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
				log.Printf("[carritos] ensure schema empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo preparar la estructura de carritos", http.StatusInternalServerError)
				return
			}
			if err := dbpkg.RefreshCarritosActivosConTarifasTiempo(dbEmp, empresaID, time.Now()); err != nil {
				log.Printf("[carritos] refresh tarifas_tiempo empresa_id=%d error: %v", empresaID, err)
			}
			includeInactive := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "1") ||
				strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "true")
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			estacionID, err := parseOptionalInt64CarritoQuery(r, "estacion_id")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseOptionalInt64CarritoQuery(r, "id")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if isStationBoardOnlyCarritoRequest(r) && id > 0 {
				http.Error(w, "forbidden: este rol solo puede consultar el tablero de estaciones", http.StatusForbidden)
				return
			}
			if id > 0 {
				carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "Carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get empresa_id=%d id=%d error: %v", empresaID, id, err)
					http.Error(w, "No se pudo consultar el carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				writeJSON(w, http.StatusOK, carrito)
				return
			}

			rows, err := dbpkg.GetCarritosCompraByEmpresa(dbEmp, empresaID, includeInactive, q)
			if err != nil {
				log.Printf("[carritos] list empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudieron listar los carritos", http.StatusInternalServerError)
				return
			}
			if estacionID > 0 {
				filtered := make([]dbpkg.CarritoCompra, 0, len(rows))
				for _, row := range rows {
					currentStationID, _, _ := dbpkg.ResolveCarritoStationIdentity(&row)
					if currentStationID == estacionID {
						filtered = append(filtered, row)
					}
				}
				rows = filtered
			}
			if isStationBoardOnlyCarritoRequest(r) {
				filtered := make([]dbpkg.CarritoCompra, 0, len(rows))
				for _, row := range rows {
					currentStationID, _, _ := dbpkg.ResolveCarritoStationIdentity(&row)
					if currentStationID > 0 {
						filtered = append(filtered, row)
					}
				}
				rows = filtered
			}
			var accessErr error
			rows, accessErr = filterCarritosByStationAccess(dbEmp, empresaID, adminEmailFromRequest(r), rows)
			if accessErr != nil {
				writeCarritoStationAccessError(w, accessErr)
				return
			}
			attachCarritoStationRuntimeSummaries(dbEmp, rows)
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if isPorteroRestrictedCarritoRequest(r, action) {
				http.Error(w, "forbidden: el rol portero solo puede ver y activar estaciones", http.StatusForbidden)
				return
			}
			if isServicioLimpiezaRestrictedCarritoRequest(r, action) {
				http.Error(w, "forbidden: el rol Servicio de limpieza solo puede ver estaciones y reportar aseo", http.StatusForbidden)
				return
			}
			if action == "abono" || action == "registrar_abono" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				carritoID, errID := parseInt64Query(r, "carrito_id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				var payload struct {
					Monto          float64 `json:"monto"`
					MetodoPago     string  `json:"metodo_pago"`
					ReferenciaPago string  `json:"referencia_pago"`
					CierreCajaID   int64   `json:"cierre_caja_id"`
					CajaCodigo     string  `json:"caja_codigo"`
					CajaTurno      string  `json:"caja_turno"`
					CajaSucursalID int64   `json:"caja_sucursal_id"`
					Observaciones  string  `json:"observaciones"`
				}
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
						http.Error(w, "JSON invalido", http.StatusBadRequest)
						return
					}
				}
				monto := roundMoneyCarritoHandler(payload.Monto)
				if monto <= 0 {
					http.Error(w, "monto de abono invalido", http.StatusBadRequest)
					return
				}
				metodoPago := dbpkg.NormalizeMetodoPagoCarrito(payload.MetodoPago)
				if metodoPago == "" || metodoPago == "codigo_descuento" || metodoPago == "mixto" {
					http.Error(w, "metodo_pago invalido para abono. Use efectivo, tarjeta debito, tarjeta credito, transferencia Bre-B, Nequi u otra transferencia", http.StatusBadRequest)
					return
				}
				if carritoMetodoPagoRequiereReferencia(metodoPago) && len(strings.TrimSpace(payload.ReferenciaPago)) < 4 {
					http.Error(w, "referencia_pago es obligatoria para abonos con tarjeta o transferencia (minimo 4 caracteres)", http.StatusBadRequest)
					return
				}
				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, carritoID)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get for abono empresa_id=%d carrito_id=%d error: %v", empresaID, carritoID, errCarrito)
					http.Error(w, "No se pudo validar el carrito para abono", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				monto = roundMoneyCarritoForMoneda(payload.Monto, carrito.Moneda)
				if monto <= 0 {
					http.Error(w, "monto de abono invalido para la moneda del carrito", http.StatusBadRequest)
					return
				}
				if !isCarritoOperativoActivo(carrito) {
					http.Error(w, "solo se pueden registrar abonos en una cuenta activa de estacion", http.StatusConflict)
					return
				}
				rolOperacion := strings.TrimSpace(adminRoleFromRequest(r))
				cfgOperativa, errCfgOperativa := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, empresaID)
				if errCfgOperativa != nil {
					log.Printf("[carritos] get configuracion_operativa abono empresa_id=%d error: %v", empresaID, errCfgOperativa)
					cfgOperativa = nil
				}
				permisosOperativos := dbpkg.ResolveEmpresaConfiguracionOperativaParaRol(cfgOperativa, rolOperacion)
				if !permisosOperativos.IsMetodoPagoHabilitado(metodoPago) {
					http.Error(w, "metodo_pago no habilitado para la empresa/rol actual", http.StatusForbidden)
					return
				}
				if ok, err := carritoMetodoPagoHabilitadoPorEmpresa(dbEmp, empresaID, metodoPago); err != nil {
					log.Printf("[carritos] validar metodo pago empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo validar la configuracion de medios de pago", http.StatusInternalServerError)
					return
				} else if !ok {
					http.Error(w, "metodo_pago deshabilitado en la configuracion del carrito de esta empresa", http.StatusForbidden)
					return
				}
				usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))
				cierreCaja, errCierreCaja := dbpkg.GetEmpresaCierreCajaAbiertaUsuario(dbEmp, empresaID, payload.CierreCajaID, payload.CajaCodigo, payload.CajaTurno, payload.CajaSucursalID, usuarioOperacion)
				if errCierreCaja != nil {
					if errors.Is(errCierreCaja, sql.ErrNoRows) {
						http.Error(w, "debes seleccionar una caja abierta y activa antes de registrar un abono", http.StatusConflict)
						return
					}
					log.Printf("[carritos] resolver caja abierta abono empresa_id=%d carrito_id=%d error: %v", empresaID, carritoID, errCierreCaja)
					http.Error(w, "No se pudo validar la caja abierta para este abono", http.StatusInternalServerError)
					return
				}
				abonoID, errAbono := dbpkg.CreateCarritoCompraAbono(dbEmp, dbpkg.CarritoCompraAbono{
					EmpresaID:      empresaID,
					CarritoID:      carritoID,
					Monto:          monto,
					MetodoPago:     metodoPago,
					ReferenciaPago: strings.TrimSpace(payload.ReferenciaPago),
					CierreCajaID:   cierreCaja.ID,
					CajaCodigo:     cierreCaja.CajaCodigo,
					CajaTurno:      cierreCaja.Turno,
					CajaSucursalID: cierreCaja.SucursalID,
					UsuarioCreador: usuarioOperacion,
					Observaciones:  strings.TrimSpace(payload.Observaciones),
				})
				if errAbono != nil {
					log.Printf("[carritos] crear abono empresa_id=%d carrito_id=%d error: %v", empresaID, carritoID, errAbono)
					http.Error(w, errAbono.Error(), http.StatusBadRequest)
					return
				}
				if metodoPago == "efectivo" {
					if errCaja := dbpkg.RegistrarIngresoEfectivoCierreCaja(dbEmp, empresaID, cierreCaja.ID, monto); errCaja != nil {
						log.Printf("[carritos] sumar efectivo abono empresa_id=%d cierre_id=%d carrito_id=%d error: %v", empresaID, cierreCaja.ID, carritoID, errCaja)
					}
				}
				estacionID, estacionCodigo, estacionNombre := dbpkg.ResolveCarritoStationIdentity(carrito)
				if _, errMetric := dbpkg.RecordCarritoStationMetric(dbEmp, dbpkg.CarritoStationMetricInput{
					EmpresaID:           empresaID,
					CarritoID:           carritoID,
					EstacionID:          estacionID,
					EstacionCodigo:      estacionCodigo,
					EstacionNombre:      estacionNombre,
					EventoOperacion:     "abono",
					MetodoPago:          metodoPago,
					Moneda:              carrito.Moneda,
					MontoTotal:          carrito.Total,
					MontoPagado:         monto,
					ActivadoEn:          carrito.ActivadoEn,
					ReferenciaOperacion: strings.TrimSpace(payload.ReferenciaPago),
					CierreCajaID:        cierreCaja.ID,
					CajaCodigo:          cierreCaja.CajaCodigo,
					CajaTurno:           cierreCaja.Turno,
					CajaSucursalID:      cierreCaja.SucursalID,
					UsuarioCreador:      usuarioOperacion,
					Observaciones:       "abono registrado a cuenta de estacion",
				}); errMetric != nil {
					log.Printf("[carritos] metrica abono empresa_id=%d carrito_id=%d error: %v", empresaID, carritoID, errMetric)
				}
				rows, errRows := dbpkg.ListCarritoCompraAbonos(dbEmp, empresaID, carritoID, false)
				if errRows != nil {
					log.Printf("[carritos] listar abonos after create empresa_id=%d carrito_id=%d error: %v", empresaID, carritoID, errRows)
				}
				totalAbonos, errTotal := dbpkg.TotalCarritoCompraAbonos(dbEmp, empresaID, carritoID)
				if errTotal != nil {
					log.Printf("[carritos] total abonos after create empresa_id=%d carrito_id=%d error: %v", empresaID, carritoID, errTotal)
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{
					"ok":      true,
					"id":      abonoID,
					"abonos":  rows,
					"total":   totalAbonos,
					"metodo":  metodoPago,
					"caja_id": cierreCaja.ID,
				})
				return
			}
			var payload dbpkg.CarritoCompra
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if err := validateCarritoPayload(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := ensureCarritoStationAccessForCarrito(dbEmp, payload.EmpresaID, adminEmailFromRequest(r), &payload); err != nil {
				writeCarritoStationAccessError(w, err)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.CreateCarritoCompra(dbEmp, payload)
			if err != nil {
				log.Printf("[carritos] create empresa_id=%d nombre=%q error: %v", payload.EmpresaID, payload.Nombre, err)
				http.Error(w, "No se pudo crear el carrito (valide nombre/codigo no duplicados)", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if isPorteroRestrictedCarritoRequest(r, action) {
				http.Error(w, "forbidden: el rol portero solo puede ver y activar estaciones", http.StatusForbidden)
				return
			}
			if isServicioLimpiezaRestrictedCarritoRequest(r, action) {
				http.Error(w, "forbidden: el rol Servicio de limpieza solo puede ver estaciones y reportar aseo", http.StatusForbidden)
				return
			}
			if action == "abrir_caja_cobro" || action == "abrir_caja_para_cobro" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				var payload struct {
					CajaCodigo       string  `json:"caja_codigo"`
					Turno            string  `json:"turno"`
					SucursalID       int64   `json:"sucursal_id"`
					AperturaEfectivo float64 `json:"apertura_efectivo"`
					Moneda           string  `json:"moneda"`
				}
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
						http.Error(w, "payload invalido", http.StatusBadRequest)
						return
					}
				}
				cierre, created, errCaja := openCajaCobroForCarrito(dbEmp, dbSuper, empresaID, payload.CajaCodigo, payload.Turno, payload.SucursalID, payload.AperturaEfectivo, payload.Moneda, strings.TrimSpace(adminEmailFromRequest(r)))
				if errCaja != nil {
					log.Printf("[carritos] abrir caja cobro empresa_id=%d error: %v", empresaID, errCaja)
					http.Error(w, errCaja.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":      true,
					"created": created,
					"caja":    cierre,
				})
				return
			}
			if action == "cambiar_tarifa" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				carritoPrevio, errCarritoPrevio := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarritoPrevio != nil {
					if errors.Is(errCarritoPrevio, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo validar el carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carritoPrevio); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				var payload struct {
					TipoTarifa string `json:"tipo_tarifa"`
					Tipo       string `json:"tipo"`
					TarifaID   int64  `json:"tarifa_id"`
				}
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
						http.Error(w, "payload invalido", http.StatusBadRequest)
						return
					}
				}
				tipoTarifa := strings.TrimSpace(payload.TipoTarifa)
				if tipoTarifa == "" {
					tipoTarifa = strings.TrimSpace(payload.Tipo)
				}
				if tipoTarifa == "" {
					tipoTarifa = strings.TrimSpace(r.URL.Query().Get("tipo_tarifa"))
				}
				tarifaID := payload.TarifaID
				if tarifaID <= 0 {
					if raw := strings.TrimSpace(r.URL.Query().Get("tarifa_id")); raw != "" {
						if parsed, perr := strconv.ParseInt(raw, 10, 64); perr == nil {
							tarifaID = parsed
						}
					}
				}
				if err := dbpkg.SetCarritoTarifaTiempoManual(dbEmp, empresaID, id, tipoTarifa, tarifaID); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					http.Error(w, "No se pudo obtener carrito actualizado", http.StatusInternalServerError)
					return
				}
				if resumen, err := dbpkg.ResolveCarritoTarifaPorMinutosResumen(dbEmp, *carrito, time.Now()); err == nil {
					carrito.TarifaPorMinutos = resumen
				}
				if resumen, err := dbpkg.ResolveCarritoTarifaPorDiaResumen(dbEmp, *carrito, time.Now()); err == nil {
					carrito.TarifaPorDia = resumen
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                 true,
					"action":             "cambiar_tarifa",
					"carrito":            carrito,
					"tarifa_tiempo_tipo": carrito.TarifaTiempoTipo,
					"tarifa_tiempo_id":   carrito.TarifaTiempoID,
					"tarifa_por_minutos": carrito.TarifaPorMinutos,
					"tarifa_por_dia":     carrito.TarifaPorDia,
				})
				return
			}
			if action == "generar_codigo_vip" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}

				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo obtener carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				estacionID, _, _ := dbpkg.ResolveCarritoStationIdentity(carrito)
				if estacionID <= 0 {
					http.Error(w, "este carrito no es de estación", http.StatusBadRequest)
					return
				}
				if isCarritoVentaPagada(carrito) {
					http.Error(w, "carrito ya pagado", http.StatusConflict)
					return
				}
				ttl := 720
				if s := strings.TrimSpace(r.URL.Query().Get("ttl_minutos")); s != "" {
					if v, perr := strconv.Atoi(s); perr == nil && v > 0 && v <= 4320 {
						ttl = v
					}
				}
				vip, err := dbpkg.CreateEstacionVIPCodigo(dbEmp, empresaID, estacionID, carrito.ID, ttl, strings.TrimSpace(adminEmailFromRequest(r)), "codigo vip generado")
				if err != nil {
					http.Error(w, "No se pudo generar codigo vip", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":           true,
					"empresa_id":   empresaID,
					"estacion_id":  estacionID,
					"carrito_id":   carrito.ID,
					"codigo":       vip.Codigo,
					"expira_en":    vip.ExpiraEn,
					"public_route": "/productos_estacion_clientes_publico.html",
				})
				return
			}
			if action == "activar_estacion" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				resetItems := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("reset_items")), "1") ||
					strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("reset_items")), "true")
				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get for activar_estacion empresa_id=%d id=%d error: %v", empresaID, id, errCarrito)
					http.Error(w, "No se pudo validar estado del carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				if isCarritoVentaPagada(carrito) && !resetItems {
					http.Error(w, "venta pagada: use reset_items=1 para iniciar una nueva sesion", http.StatusConflict)
					return
				}
				if resetItems {
					if err := validateStationCancelMargin(dbEmp, empresaID, carrito); err != nil {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					if !isCarritoVentaPagada(carrito) {
						hasActiveItems := false
						items, errItems := dbpkg.GetCarritoCompraItems(dbEmp, empresaID, id, false)
						if errItems != nil {
							log.Printf("[carritos] get items for cancelar_estacion empresa_id=%d id=%d error: %v", empresaID, id, errItems)
							http.Error(w, "No se pudo validar si el carrito tiene productos o servicios", http.StatusInternalServerError)
							return
						}
						for _, item := range items {
							if strings.EqualFold(strings.TrimSpace(item.Estado), "activo") {
								hasActiveItems = true
								break
							}
						}
						if hasActiveItems || roundMoneyCarritoHandler(carrito.Total) > 0 || roundMoneyCarritoHandler(carrito.Subtotal) > 0 {
							http.Error(w, "no se puede cancelar este carrito porque tiene productos, servicios o valores cargados; primero devuelve los productos o servicios agregados", http.StatusConflict)
							return
						}
					}
				}
				if err := dbpkg.ActivateCarritoStationSession(dbEmp, empresaID, id, resetItems); err != nil {
					log.Printf("[carritos] activar_estacion empresa_id=%d id=%d reset_items=%v error: %v", empresaID, id, resetItems, err)
					http.Error(w, "No se pudo activar el carrito de estación", http.StatusInternalServerError)
					return
				}
				registrarEventoContableVentaCarrito(dbEmp, r, carrito, "venta_sesion_activada", carrito.Total, map[string]interface{}{
					"action":                "activar_estacion",
					"reset_items":           resetItems,
					"estado_venta_anterior": carrito.EstadoVenta,
					"estado_venta_nuevo":    "venta_abierta",
				}, "activacion de sesion de venta en estacion")
				dispatchControlElectricoEstacionAsync(dbEmp, carrito, true, strings.TrimSpace(adminEmailFromRequest(r)), "activar_estacion")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": "activo", "estado_carrito": "abierto", "estado_venta": "venta_abierta"})
				return
			}

			if action == "transferir_estacion" || action == "transferir_cuenta" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				origenID, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				if !carritoTransferenciaEstacionHabilitada(dbEmp, empresaID) {
					http.Error(w, "transferencia de cuenta desactivada para esta empresa", http.StatusForbidden)
					return
				}
				var payload struct {
					DestinoCarritoID  int64  `json:"destino_carrito_id"`
					TargetCarritoID   int64  `json:"target_carrito_id"`
					DestinoEstacionID int64  `json:"destino_estacion_id"`
					TargetEstacionID  int64  `json:"target_estacion_id"`
					Motivo            string `json:"motivo"`
				}
				if r.Body != nil {
					_ = json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&payload)
				}
				destinoCarritoID := payload.DestinoCarritoID
				if destinoCarritoID <= 0 {
					destinoCarritoID = payload.TargetCarritoID
				}
				destinoEstacionID := payload.DestinoEstacionID
				if destinoEstacionID <= 0 {
					destinoEstacionID = payload.TargetEstacionID
				}
				if destinoCarritoID <= 0 && destinoEstacionID > 0 {
					destinoCarrito, err := dbpkg.GetCarritoCompraByStation(dbEmp, empresaID, destinoEstacionID)
					if err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							http.Error(w, "carrito destino no encontrado para la estacion", http.StatusNotFound)
							return
						}
						log.Printf("[carritos] resolver destino transferencia empresa_id=%d estacion_id=%d error: %v", empresaID, destinoEstacionID, err)
						http.Error(w, "No se pudo resolver el carrito destino", http.StatusInternalServerError)
						return
					}
					destinoCarritoID = destinoCarrito.ID
				}
				if destinoCarritoID <= 0 {
					http.Error(w, "destino_carrito_id o destino_estacion_id es obligatorio", http.StatusBadRequest)
					return
				}
				origen, errOrigen := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, origenID)
				if errOrigen != nil {
					if errors.Is(errOrigen, sql.ErrNoRows) {
						http.Error(w, "carrito origen no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo validar el carrito origen", http.StatusInternalServerError)
					return
				}
				destino, errDestino := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, destinoCarritoID)
				if errDestino != nil {
					if errors.Is(errDestino, sql.ErrNoRows) {
						http.Error(w, "carrito destino no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo validar el carrito destino", http.StatusInternalServerError)
					return
				}
				usuarioActual := strings.TrimSpace(adminEmailFromRequest(r))
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, usuarioActual, origen); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, usuarioActual, destino); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				result, err := dbpkg.TransferCarritoStationCuenta(dbEmp, empresaID, origenID, destinoCarritoID, usuarioActual, payload.Motivo)
				if err != nil {
					lowerErr := strings.ToLower(err.Error())
					status := http.StatusConflict
					if strings.Contains(lowerErr, "obligatorio") || strings.Contains(lowerErr, "diferente") {
						status = http.StatusBadRequest
					}
					http.Error(w, err.Error(), status)
					return
				}
				registrarEventoContableVentaCarrito(dbEmp, r, origen, "cuenta_transferida", result.Total, map[string]interface{}{
					"action":                   action,
					"origen_carrito_id":        result.OrigenCarritoID,
					"destino_carrito_id":       result.DestinoCarritoID,
					"origen_estacion_id":       result.OrigenEstacionID,
					"destino_estacion_id":      result.DestinoEstacionID,
					"tarifa_tiempo_tipo":       result.TarifaTiempoTipo,
					"tarifa_tiempo_id":         result.TarifaTiempoID,
					"items_transferidos":       result.ItemsTransferidos,
					"abonos_transferidos":      result.AbonosTransferidos,
					"motivo":                   strings.TrimSpace(payload.Motivo),
					"validacion_tarifa":        "equivalente_o_sin_tarifa_temporal",
					"transferencia_habilitada": true,
				}, "transferencia de cuenta entre estaciones")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "transferencia": result})
				return
			}

			if action == "pagar_estacion" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				tarifasTiempo, errTarifasTiempo := dbpkg.RefreshCarritoTotalConTarifasTiempo(dbEmp, empresaID, id, time.Now())
				if errTarifasTiempo != nil {
					log.Printf("[carritos] refresh carrito tarifas_tiempo empresa_id=%d id=%d error: %v", empresaID, id, errTarifasTiempo)
					http.Error(w, "No se pudo recalcular tarifa temporal del carrito", http.StatusInternalServerError)
					return
				}
				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get for pagar_estacion empresa_id=%d id=%d error: %v", empresaID, id, errCarrito)
					http.Error(w, "No se pudo validar estado del carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				log.Printf("[carritos] debug intentar %s empresa_id=%d id=%d estado=%q estado_carrito=%q pagado_en=%q", action, empresaID, id, carrito.Estado, carrito.EstadoCarrito, carrito.PagadoEn)
				if err := validateCarritoTransitionForAction(carrito, action); err != nil {
					log.Printf("[carritos] validate failed action=%s empresa_id=%d id=%d estado=%q estado_carrito=%q pagado_en=%q err=%v", action, empresaID, id, carrito.Estado, carrito.EstadoCarrito, carrito.PagadoEn, err)
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				if carritoClienteObligatorioParaPago(dbEmp, empresaID) && carrito.ClienteID <= 0 {
					http.Error(w, "cliente obligatorio: registra o selecciona un cliente antes de pagar el carrito", http.StatusConflict)
					return
				}
				if prerequisite, err := validateCarritoPaymentPrerequisites(dbEmp, carrito); err != nil {
					log.Printf("[carritos] preflight pago empresa_id=%d id=%d error: %v", empresaID, id, err)
					http.Error(w, "No se pudieron validar las dependencias operativas del pago", http.StatusInternalServerError)
					return
				} else if prerequisite != nil {
					writeCarritoBusinessPrerequisite(w, http.StatusConflict, *prerequisite)
					return
				}

				var payload struct {
					MetodoPago      string                    `json:"metodo_pago"`
					ReferenciaPago  string                    `json:"referencia_pago"`
					PagosMixtos     []carritoPagoMixtoEntrada `json:"pagos_mixtos"`
					Pagos           []carritoPagoMixtoEntrada `json:"pagos"`
					DescuentoTipo   string                    `json:"descuento_tipo"`
					DescuentoCodigo string                    `json:"descuento_codigo"`
					CodigoDescuento string                    `json:"codigo_descuento"`
					DescuentoValor  float64                   `json:"descuento_valor"`
					DevolucionTotal float64                   `json:"devolucion_total"`
					AbonosTotal     float64                   `json:"abonos_total"`
					TotalPagado     float64                   `json:"total_pagado"`
					AplicarPropina  *bool                     `json:"aplicar_propina"`
					UsuarioLavador  string                    `json:"usuario_lavador"`
					CierreCajaID    int64                     `json:"cierre_caja_id"`
					CajaCodigo      string                    `json:"caja_codigo"`
					CajaTurno       string                    `json:"caja_turno"`
					CajaSucursalID  int64                     `json:"caja_sucursal_id"`
				}
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
						http.Error(w, "JSON invalido", http.StatusBadRequest)
						return
					}
				}
				if len(payload.PagosMixtos) == 0 && len(payload.Pagos) > 0 {
					payload.PagosMixtos = payload.Pagos
				}

				metodoPago := dbpkg.NormalizeMetodoPagoCarrito(payload.MetodoPago)
				if metodoPago == "" {
					http.Error(w, "metodo_pago invalido. Use: efectivo, tarjeta_credito, tarjeta_debito, transferencia_bre_b, transferencia_nequi, transferencia_otro, credito_cliente, codigo_descuento o mixto", http.StatusBadRequest)
					return
				}

				rolOperacion := strings.TrimSpace(adminRoleFromRequest(r))
				cfgOperativa, errCfgOperativa := dbpkg.GetEmpresaConfiguracionOperativa(dbEmp, empresaID)
				if errCfgOperativa != nil {
					log.Printf("[carritos] get configuracion_operativa empresa_id=%d error: %v", empresaID, errCfgOperativa)
					cfgOperativa = nil
				}
				permisosOperativos := dbpkg.ResolveEmpresaConfiguracionOperativaParaRol(cfgOperativa, rolOperacion)
				if !permisosOperativos.IsMetodoPagoHabilitado(metodoPago) {
					http.Error(w, "metodo_pago no habilitado para la empresa/rol actual", http.StatusForbidden)
					return
				}
				if ok, err := carritoMetodoPagoHabilitadoPorEmpresa(dbEmp, empresaID, metodoPago); err != nil {
					log.Printf("[carritos] validar metodo pago empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo validar la configuracion de medios de pago", http.StatusInternalServerError)
					return
				} else if !ok {
					http.Error(w, "metodo_pago deshabilitado en la configuracion del carrito de esta empresa", http.StatusForbidden)
					return
				}

				referenciaPago := strings.TrimSpace(payload.ReferenciaPago)
				pagosMixtos := make([]carritoPagoMixtoNormalizado, 0)
				totalPagadoMixto := 0.0
				if metodoPago == "mixto" {
					var err error
					pagosMixtos, totalPagadoMixto, err = normalizePagosMixtosCarrito(payload.PagosMixtos, carrito.Moneda)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					for _, tramo := range pagosMixtos {
						if !permisosOperativos.IsMetodoPagoHabilitado(tramo.Metodo) {
							http.Error(w, "uno o mas metodos del pago mixto no estan habilitados para la empresa/rol actual", http.StatusForbidden)
							return
						}
						if ok, err := carritoMetodoPagoHabilitadoPorEmpresa(dbEmp, empresaID, tramo.Metodo); err != nil {
							log.Printf("[carritos] validar tramo pago mixto empresa_id=%d error: %v", empresaID, err)
							http.Error(w, "No se pudo validar la configuracion de medios de pago", http.StatusInternalServerError)
							return
						} else if !ok {
							http.Error(w, "uno o mas metodos del pago mixto estan deshabilitados en la configuracion del carrito de esta empresa", http.StatusForbidden)
							return
						}
					}
					referenciaPago = buildReferenciaPagoMixto(pagosMixtos, carrito.Moneda)
				} else if carritoMetodoPagoRequiereReferencia(metodoPago) && len(referenciaPago) < 4 {
					http.Error(w, "referencia_pago es obligatoria para pagos con tarjeta o transferencia (minimo 4 caracteres)", http.StatusBadRequest)
					return
				}

				descuentoTipo := strings.TrimSpace(strings.ToLower(payload.DescuentoTipo))
				descuentoCodigo := strings.TrimSpace(payload.DescuentoCodigo)
				if descuentoCodigo == "" {
					descuentoCodigo = strings.TrimSpace(payload.CodigoDescuento)
				}
				descuentoValor := payload.DescuentoValor
				if descuentoValor < 0 {
					descuentoValor = 0
				}
				if descuentoValor > carrito.Total {
					descuentoValor = carrito.Total
				}

				codigoDescuentoID := int64(0)
				requiereCodigo := metodoPago == "codigo_descuento" || descuentoTipo == "code" || descuentoCodigo != ""
				if requiereCodigo {
					if strings.TrimSpace(descuentoCodigo) == "" {
						http.Error(w, "descuento_codigo es obligatorio cuando se usa codigo de descuento", http.StatusBadRequest)
						return
					}
					aplicado, err := dbpkg.ResolveCodigoDescuentoParaMontoConContexto(
						dbEmp,
						empresaID,
						descuentoCodigo,
						carrito.Total,
						dbpkg.CodigoDescuentoContexto{
							CarritoID:  id,
							ClienteID:  carrito.ClienteID,
							CanalVenta: carrito.CanalVenta,
						},
					)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					codigoDescuentoID = aplicado.ID
					descuentoTipo = "code"
					descuentoCodigo = aplicado.Codigo
					descuentoValor = aplicado.ValorAplicado
				}
				descuentoValor = roundMoneyCarritoForMoneda(descuentoValor, carrito.Moneda)

				devolucionTotal := payload.DevolucionTotal
				if devolucionTotal < 0 {
					devolucionTotal = 0
				}
				maxDevolucion := carrito.Total - descuentoValor
				if maxDevolucion < 0 {
					maxDevolucion = 0
				}
				if devolucionTotal > maxDevolucion {
					devolucionTotal = maxDevolucion
				}
				devolucionTotal = roundMoneyCarritoForMoneda(devolucionTotal, carrito.Moneda)

				totalEsperado := roundMoneyCarritoForMoneda(carrito.Total-descuentoValor-devolucionTotal, carrito.Moneda)
				if totalEsperado < 0 {
					totalEsperado = 0
				}
				abonosRegistrados, errAbonos := dbpkg.TotalCarritoCompraAbonos(dbEmp, empresaID, id)
				if errAbonos != nil {
					log.Printf("[carritos] total abonos empresa_id=%d carrito_id=%d error: %v", empresaID, id, errAbonos)
					http.Error(w, "No se pudieron validar los abonos del carrito", http.StatusInternalServerError)
					return
				}
				abonosAplicados := roundMoneyCarritoForMoneda(abonosRegistrados, carrito.Moneda)
				if abonosAplicados < 0 {
					abonosAplicados = 0
				}
				if abonosAplicados > totalEsperado {
					abonosAplicados = totalEsperado
				}
				if payload.AbonosTotal > 0 && math.Abs(roundMoneyCarritoForMoneda(payload.AbonosTotal, carrito.Moneda)-abonosAplicados) > carritoMoneyTolerance(carrito.Moneda) {
					http.Error(w, "los abonos del carrito cambiaron; actualiza el carrito antes de pagar", http.StatusConflict)
					return
				}
				saldoEsperado := roundMoneyCarritoForMoneda(totalEsperado-abonosAplicados, carrito.Moneda)
				if saldoEsperado < 0 {
					saldoEsperado = 0
				}

				propinaCfg, errCfgPropina := dbpkg.GetEmpresaPropinasConfiguracion(dbEmp, empresaID)
				if errCfgPropina != nil {
					log.Printf("[carritos] get propinas config empresa_id=%d error: %v", empresaID, errCfgPropina)
					propinaCfg = &dbpkg.EmpresaPropinasConfiguracion{
						EmpresaID:        empresaID,
						ModoDistribucion: dbpkg.EmpresaPropinaModoPorUsuario,
					}
				}

				aplicarPropina := true
				if payload.AplicarPropina != nil {
					aplicarPropina = *payload.AplicarPropina
				} else if propinaCfg != nil {
					aplicarPropina = propinaCfg.AplicarAutomaticamente
				}

				propinaHabilitada := propinaCfg != nil && propinaCfg.HabilitarPropina && permisosOperativos.HabilitarPropinas
				propinaPorcentaje := 0.0
				propinaModo := dbpkg.EmpresaPropinaModoPorUsuario
				if propinaCfg != nil {
					if propinaCfg.PorcentajePropina > 0 {
						propinaPorcentaje = propinaCfg.PorcentajePropina
					}
					if strings.TrimSpace(propinaCfg.ModoDistribucion) != "" {
						propinaModo = propinaCfg.ModoDistribucion
					}
				}

				propinaAplicada := propinaHabilitada && aplicarPropina && propinaPorcentaje > 0
				montoPropina := 0.0
				if propinaAplicada {
					montoPropina = roundMoneyCarritoForMoneda(totalEsperado*(propinaPorcentaje/100), carrito.Moneda)
					if montoPropina < 0 {
						montoPropina = 0
					}
				}
				totalDocumentoConPropina := roundMoneyCarritoForMoneda(totalEsperado+montoPropina, carrito.Moneda)
				totalEsperadoConPropina := roundMoneyCarritoForMoneda(saldoEsperado+montoPropina, carrito.Moneda)

				montoCreditoVenta := carritoCreditoAmount(metodoPago, pagosMixtos, totalEsperadoConPropina, carrito.Moneda)
				creditoPrevalidado, errCreditoPrevalidado := validarCupoCreditoClienteParaVenta(dbEmp, carrito, montoCreditoVenta)
				if errCreditoPrevalidado != nil {
					http.Error(w, errCreditoPrevalidado.Error(), http.StatusBadRequest)
					return
				}

				totalPagado := payload.TotalPagado
				if metodoPago == "mixto" {
					totalPagado = totalPagadoMixto
				} else {
					if totalPagado < 0 {
						totalPagado = 0
					}
					if totalPagado == 0 && metodoPago != "codigo_descuento" {
						totalPagado = totalEsperadoConPropina
					}
				}
				totalPagado = roundMoneyCarritoForMoneda(totalPagado, carrito.Moneda)

				if metodoPago == "codigo_descuento" {
					if totalEsperadoConPropina > 0 {
						http.Error(w, "el codigo de descuento no cubre el total del carrito; use efectivo o tarjeta para cubrir el saldo restante", http.StatusBadRequest)
						return
					}
					totalPagado = 0
				} else if metodoPago == "mixto" {
					if len(pagosMixtos) < 2 {
						http.Error(w, "pago mixto requiere al menos 2 metodos con monto mayor a cero", http.StatusBadRequest)
						return
					}
					if math.Abs(totalPagado-totalEsperadoConPropina) > carritoMoneyTolerance(carrito.Moneda) {
						http.Error(w, "la suma de pagos mixtos debe coincidir con el total esperado", http.StatusBadRequest)
						return
					}
				} else {
					if totalPagado+0.009 < totalEsperadoConPropina {
						http.Error(w, "total_pagado insuficiente para completar el pago", http.StatusBadRequest)
						return
					}
				}

				usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))
				usuarioOperacionID := int64(0)
				if usuarioOperacionItem, errUsuario := dbpkg.ResolveEmpresaUsuarioByReference(dbEmp, empresaID, usuarioOperacion); errUsuario == nil && usuarioOperacionItem != nil {
					usuarioOperacionID = usuarioOperacionItem.ID
				}
				cierreCaja, errCierreCaja := dbpkg.GetEmpresaCierreCajaAbiertaUsuario(dbEmp, empresaID, payload.CierreCajaID, payload.CajaCodigo, payload.CajaTurno, payload.CajaSucursalID, usuarioOperacion)
				if errCierreCaja != nil {
					if errors.Is(errCierreCaja, sql.ErrNoRows) {
						http.Error(w, "debes seleccionar una caja abierta y activa antes de pagar", http.StatusConflict)
						return
					}
					log.Printf("[carritos] resolver caja abierta empresa_id=%d carrito_id=%d error: %v", empresaID, id, errCierreCaja)
					http.Error(w, "No se pudo validar la caja abierta para este pago", http.StatusInternalServerError)
					return
				}
				montoEfectivoCaja := 0.0
				if metodoPago == "efectivo" {
					montoEfectivoCaja = totalEsperadoConPropina
				} else if metodoPago == "mixto" {
					for _, tramo := range pagosMixtos {
						if tramo.Metodo == "efectivo" {
							montoEfectivoCaja += tramo.Monto
						}
					}
				}

				totalPagadoCarrito := roundMoneyCarritoForMoneda(totalPagado+abonosAplicados, carrito.Moneda)
				if err := dbpkg.PayCarritoStationSession(
					dbEmp,
					empresaID,
					id,
					metodoPago,
					referenciaPago,
					descuentoTipo,
					descuentoCodigo,
					descuentoValor,
					devolucionTotal,
					totalPagadoCarrito,
					codigoDescuentoID,
					cierreCaja.ID,
					cierreCaja.CajaCodigo,
					cierreCaja.Turno,
					cierreCaja.SucursalID,
					usuarioOperacion,
				); err != nil {
					log.Printf("[carritos] pagar_estacion empresa_id=%d id=%d error: %v", empresaID, id, err)
					if errors.Is(err, dbpkg.ErrCarritoYaPagado) {
						carritoActual, errActual := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
						if errActual == nil && isCarritoVentaPagada(carritoActual) {
							writeJSON(w, http.StatusOK, map[string]interface{}{
								"ok":             true,
								"idempotente":    true,
								"estado":         "inactivo",
								"estado_carrito": "cerrado",
								"estado_venta":   "venta_pagada",
								"mensaje":        "El carrito ya habia sido pagado; no se duplico la venta.",
							})
							return
						}
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					lower := strings.ToLower(strings.TrimSpace(err.Error()))
					if strings.Contains(lower, "metodo_pago") ||
						strings.Contains(lower, "codigo de descuento") ||
						strings.Contains(lower, "sin usos") ||
						strings.Contains(lower, "vencido") {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					http.Error(w, "No se pudo cerrar el carrito por pago", http.StatusInternalServerError)
					return
				}
				if desactivadas, err := dbpkg.EnforceLicenciaDocumentosMensualesPorEmpresa(dbEmp, dbSuper, empresaID); err != nil {
					log.Printf("[licencias] limite mensual documentos empresa_id=%d error despues de pago carrito=%d: %v", empresaID, id, err)
				} else if desactivadas > 0 {
					log.Printf("[licencias] empresa_id=%d licencia desactivada por limite mensual despues de pago carrito=%d", empresaID, id)
				}
				montoEvento := totalPagadoCarrito
				if montoEvento <= 0 {
					montoEvento = totalDocumentoConPropina
				}

				propinaRegistroID := int64(0)
				propinaRegistrada := false
				propinaWarning := ""
				if propinaCfg != nil && propinaCfg.HabilitarPropina && !permisosOperativos.HabilitarPropinas {
					propinaWarning = "propinas deshabilitadas por configuracion operativa de empresa/rol"
				}
				if montoPropina > 0 {
					movimientoPropina := dbpkg.EmpresaPropinaMovimiento{
						EmpresaID:         empresaID,
						CarritoID:         id,
						VentaReferencia:   strings.TrimSpace(carrito.Codigo),
						UsuarioOrigen:     usuarioOperacion,
						UsuarioOrigenID:   usuarioOperacionID,
						UsuarioAsignado:   usuarioOperacion,
						UsuarioAsignadoID: usuarioOperacionID,
						ModoDistribucion:  propinaModo,
						Moneda:            strings.TrimSpace(carrito.Moneda),
						BaseCobro:         totalEsperado,
						PorcentajePropina: propinaPorcentaje,
						MontoPropina:      montoPropina,
						UsuarioCreador:    usuarioOperacion,
						Estado:            "activo",
						Observaciones:     "propina registrada al cerrar carrito en estacion",
					}
					if movimientoPropina.ModoDistribucion == dbpkg.EmpresaPropinaModoUniversal {
						movimientoPropina.UsuarioAsignado = ""
						movimientoPropina.UsuarioAsignadoID = 0
					}
					propinaRegistroIDTmp, errReg := dbpkg.CreateEmpresaPropinaMovimiento(dbEmp, movimientoPropina)
					if errReg != nil {
						propinaWarning = "no se pudo registrar movimiento de propina"
						log.Printf("[carritos] registrar propina empresa_id=%d carrito_id=%d error: %v", empresaID, id, errReg)
					} else {
						propinaRegistroID = propinaRegistroIDTmp
						propinaRegistrada = true
					}
				}

				comisionResultado := &dbpkg.EmpresaComisionServicioRegistroResultado{}
				if !permisosOperativos.HabilitarComisiones {
					comisionResultado.Warning = "comisiones deshabilitadas por configuracion operativa de empresa/rol"
				} else {
					if result, errComision := dbpkg.RegisterEmpresaComisionesServicioDesdeCarrito(
						dbEmp,
						empresaID,
						id,
						strings.TrimSpace(payload.UsuarioLavador),
						usuarioOperacion,
						rolOperacion,
					); errComision != nil {
						comisionResultado.Warning = "no se pudo registrar comisiones por servicio"
						log.Printf("[carritos] registrar comision servicio empresa_id=%d carrito_id=%d error: %v", empresaID, id, errComision)
					} else if result != nil {
						comisionResultado = result
					}
				}
				registrarEventoContableVentaCarrito(dbEmp, r, carrito, "venta_pagada", montoEvento, map[string]interface{}{
					"action":                "pagar_estacion",
					"rol_operacion":         rolOperacion,
					"forma_pago":            metodoPago,
					"metodo_pago":           metodoPago,
					"referencia_pago":       referenciaPago,
					"subtotal":              montoEvento,
					"base_gravable":         montoEvento,
					"total_neto":            montoEvento,
					"pagos_mixtos":          pagosMixtosToEventPayload(pagosMixtos),
					"cfg_metodo_efectivo":   permisosOperativos.MetodoPagoEfectivo,
					"cfg_metodo_tc":         permisosOperativos.MetodoPagoTarjetaCredito,
					"cfg_metodo_td":         permisosOperativos.MetodoPagoTarjetaDebito,
					"cfg_metodo_transfer":   permisosOperativos.MetodoPagoTransferenciaBancaria,
					"cfg_metodo_mixto":      permisosOperativos.MetodoPagoMixto,
					"cfg_metodo_codigo":     permisosOperativos.MetodoPagoCodigoDescuento,
					"cfg_propinas":          permisosOperativos.HabilitarPropinas,
					"cfg_comisiones":        permisosOperativos.HabilitarComisiones,
					"descuento_tipo":        descuentoTipo,
					"descuento_codigo":      descuentoCodigo,
					"descuento_valor":       descuentoValor,
					"codigo_descuento_id":   codigoDescuentoID,
					"devolucion_total":      devolucionTotal,
					"abonos_total":          abonosAplicados,
					"saldo_esperado":        saldoEsperado,
					"total_pagado":          totalPagado,
					"total_pagado_carrito":  totalPagadoCarrito,
					"total_esperado":        totalEsperado,
					"total_esperado_final":  totalEsperadoConPropina,
					"total_documento_final": totalDocumentoConPropina,
					"cierre_caja_id":        cierreCaja.ID,
					"caja_codigo":           cierreCaja.CajaCodigo,
					"caja_turno":            cierreCaja.Turno,
					"caja_sucursal_id":      cierreCaja.SucursalID,
					"propina_aplicada":      propinaAplicada,
					"propina_porcentaje":    propinaPorcentaje,
					"propina_monto":         montoPropina,
					"propina_modo":          propinaModo,
					"propina_registro_id":   propinaRegistroID,
					"propina_registrada":    propinaRegistrada,
					"propina_usuario_id":    usuarioOperacionID,
					"comision_aplicada":     comisionResultado.Aplicada,
					"comision_porcentaje":   comisionResultado.PorcentajeComision,
					"comision_filtro":       comisionResultado.FiltroServicio,
					"comision_lavador":      comisionResultado.UsuarioLavador,
					"comision_lavador_id":   comisionResultado.UsuarioLavadorID,
					"comision_base":         comisionResultado.BaseServicios,
					"comision_monto":        comisionResultado.MontoComision,
					"comision_movimientos":  comisionResultado.MovimientosRegistrados,
					"comision_warning":      comisionResultado.Warning,
					"credito_cliente":       montoCreditoVenta,
					"cuenta_por_cobrar":     "",
					"estado_venta_anterior": carrito.EstadoVenta,
					"estado_venta_nuevo":    "venta_pagada",
				}, "pago de venta en estacion")

				// invalidar códigos VIP públicos asociados al carrito pagado
				_ = dbpkg.InvalidateVIPCodesForCarrito(dbEmp, empresaID, id, "pago_estacion")

				carritoPagado, errCarritoPagado := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarritoPagado != nil {
					log.Printf("[carritos] get after pagar_estacion empresa_id=%d id=%d error: %v", empresaID, id, errCarritoPagado)
					carritoPagado = carrito
				}
				estacionID, estacionCodigo, estacionNombre := dbpkg.ResolveCarritoStationIdentity(carritoPagado)
				if _, errMetric := dbpkg.RecordCarritoStationMetric(dbEmp, dbpkg.CarritoStationMetricInput{
					EmpresaID:           empresaID,
					CarritoID:           id,
					EstacionID:          estacionID,
					EstacionCodigo:      estacionCodigo,
					EstacionNombre:      estacionNombre,
					EventoOperacion:     "venta_pagada",
					MetodoPago:          metodoPago,
					Moneda:              carritoPagado.Moneda,
					MontoTotal:          carritoPagado.Total,
					MontoPagado:         montoEvento,
					DevolucionTotal:     devolucionTotal,
					ActivadoEn:          carritoPagado.ActivadoEn,
					PagadoEn:            carritoPagado.PagadoEn,
					ReferenciaOperacion: referenciaPago,
					CierreCajaID:        cierreCaja.ID,
					CajaCodigo:          cierreCaja.CajaCodigo,
					CajaTurno:           cierreCaja.Turno,
					CajaSucursalID:      cierreCaja.SucursalID,
					UsuarioCreador:      usuarioOperacion,
					Observaciones:       "cierre de venta simple por estacion",
				}); errMetric != nil {
					log.Printf("[carritos] metrica venta_pagada empresa_id=%d carrito_id=%d error: %v", empresaID, id, errMetric)
				}
				if montoEfectivoCaja > 0 {
					if errCaja := dbpkg.RegistrarIngresoEfectivoCierreCaja(dbEmp, empresaID, cierreCaja.ID, montoEfectivoCaja); errCaja != nil {
						log.Printf("[carritos] actualizar efectivo cierre_caja empresa_id=%d cierre_id=%d carrito_id=%d error: %v", empresaID, cierreCaja.ID, id, errCaja)
					}
				}

				documentoVenta, errDocumentoVenta := registrarDocumentoVentaDesdeCarritoPagado(dbEmp, dbSuper, carritoPagado, totalDocumentoConPropina, usuarioOperacion)
				if errDocumentoVenta != nil {
					log.Printf("[carritos] documento_venta empresa_id=%d carrito_id=%d error: %v", empresaID, id, errDocumentoVenta)
				}
				creditoVenta := &carritoCreditoVentaResultado{Aplica: false}
				if montoCreditoVenta > 0 {
					var errCreditoVenta error
					creditoVenta, errCreditoVenta = registrarCreditoVentaDesdeCarrito(dbEmp, carritoPagado, montoCreditoVenta, usuarioOperacion, documentoVenta, creditoPrevalidado)
					if errCreditoVenta != nil {
						log.Printf("[carritos] registrar credito venta empresa_id=%d carrito_id=%d error: %v", empresaID, id, errCreditoVenta)
						creditoVenta = &carritoCreditoVentaResultado{
							Aplica:        true,
							ClienteID:     carrito.ClienteID,
							ClienteNombre: strings.TrimSpace(carrito.ClienteNombre),
							MontoCredito:  montoCreditoVenta,
							Warning:       "venta cerrada, pero no se pudo crear la cartera automaticamente; revisa Creditos",
						}
					}
				}
				dispatchControlElectricoEstacionAsync(dbEmp, carritoPagado, false, usuarioOperacion, "pagar_estacion")

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                          true,
					"estado":                      "inactivo",
					"estado_carrito":              "cerrado",
					"estado_venta":                "venta_pagada",
					"tarifa_por_dia":              tarifasTiempo.TarifaPorDia,
					"tarifa_por_minutos":          tarifasTiempo.TarifaPorMinutos,
					"total_esperado":              totalEsperado,
					"abonos_total":                abonosAplicados,
					"saldo_esperado":              saldoEsperado,
					"total_esperado_con_propina":  totalEsperadoConPropina,
					"total_documento_con_propina": totalDocumentoConPropina,
					"propina": map[string]interface{}{
						"aplicada":          propinaAplicada,
						"habilitada":        propinaHabilitada,
						"porcentaje":        propinaPorcentaje,
						"monto":             montoPropina,
						"modo_distribucion": propinaModo,
						"registrada":        propinaRegistrada,
						"registro_id":       propinaRegistroID,
						"usuario_id":        usuarioOperacionID,
						"warning":           propinaWarning,
					},
					"comision": map[string]interface{}{
						"aplicada":                comisionResultado.Aplicada,
						"habilitada":              comisionResultado.Habilitada,
						"aplicacion_automatica":   comisionResultado.AplicacionAutomatica,
						"porcentaje_comision":     comisionResultado.PorcentajeComision,
						"filtro_servicio":         comisionResultado.FiltroServicio,
						"usuario_lavador":         comisionResultado.UsuarioLavador,
						"usuario_lavador_id":      comisionResultado.UsuarioLavadorID,
						"base_servicios":          comisionResultado.BaseServicios,
						"monto_comision":          comisionResultado.MontoComision,
						"movimientos_registrados": comisionResultado.MovimientosRegistrados,
						"warning":                 comisionResultado.Warning,
					},
					"configuracion_operativa": map[string]interface{}{
						"rol":                                rolOperacion,
						"metodo_pago_efectivo":               permisosOperativos.MetodoPagoEfectivo,
						"metodo_pago_tarjeta_credito":        permisosOperativos.MetodoPagoTarjetaCredito,
						"metodo_pago_tarjeta_debito":         permisosOperativos.MetodoPagoTarjetaDebito,
						"metodo_pago_transferencia_bancaria": permisosOperativos.MetodoPagoTransferenciaBancaria,
						"metodo_pago_credito_cliente":        true,
						"metodo_pago_mixto":                  permisosOperativos.MetodoPagoMixto,
						"metodo_pago_codigo_descuento":       permisosOperativos.MetodoPagoCodigoDescuento,
						"habilitar_propinas":                 permisosOperativos.HabilitarPropinas,
						"habilitar_comisiones":               permisosOperativos.HabilitarComisiones,
					},
					"caja": map[string]interface{}{
						"cierre_caja_id":   cierreCaja.ID,
						"caja_codigo":      cierreCaja.CajaCodigo,
						"caja_turno":       cierreCaja.Turno,
						"caja_sucursal_id": cierreCaja.SucursalID,
						"efectivo_sumado":  montoEfectivoCaja,
					},
					"credito_venta":   creditoVenta,
					"documento_venta": documentoVenta,
				})
				return
			}

			if action == "recuperar_interrumpido" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}

				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get for recuperar_interrumpido empresa_id=%d id=%d error: %v", empresaID, id, errCarrito)
					http.Error(w, "No se pudo validar estado del carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				if isCarritoVentaPagada(carrito) {
					http.Error(w, "la venta ya fue pagada; para iniciar una nueva sesion use activar_estacion con reset_items=1", http.StatusConflict)
					return
				}

				if err := dbpkg.RecoverInterruptedCarritoSession(dbEmp, empresaID, id); err != nil {
					log.Printf("[carritos] recuperar_interrumpido empresa_id=%d id=%d error: %v", empresaID, id, err)
					http.Error(w, "No se pudo recuperar el carrito interrumpido", http.StatusInternalServerError)
					return
				}

				carritoActualizado, errUpdated := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errUpdated != nil {
					log.Printf("[carritos] get after recuperar_interrumpido empresa_id=%d id=%d error: %v", empresaID, id, errUpdated)
					carritoActualizado = carrito
				}

				registrarEventoContableVentaCarrito(dbEmp, r, carritoActualizado, "venta_interrumpida_recuperada", carritoActualizado.Total, map[string]interface{}{
					"action":                  "recuperar_interrumpido",
					"estado_registro_previo":  normalizeCarritoRegistroEstado(carrito.Estado),
					"estado_operativo_previo": normalizeCarritoOperativoEstado(carrito.EstadoCarrito),
					"estado_venta_anterior":   carrito.EstadoVenta,
					"estado_venta_nuevo":      carritoActualizado.EstadoVenta,
				}, "recuperacion de carrito interrumpido")

				registrarAuditoriaCarritoOperacionNoBloqueante(dbEmp, r, empresaID, id, "recuperar_interrumpido", http.StatusOK, map[string]interface{}{
					"estado_registro_previo":  normalizeCarritoRegistroEstado(carrito.Estado),
					"estado_operativo_previo": normalizeCarritoOperativoEstado(carrito.EstadoCarrito),
					"estado_registro_nuevo":   normalizeCarritoRegistroEstado(carritoActualizado.Estado),
					"estado_operativo_nuevo":  normalizeCarritoOperativoEstado(carritoActualizado.EstadoCarrito),
					"estado_venta_nuevo":      carritoActualizado.EstadoVenta,
					"pagado_en":               strings.TrimSpace(carritoActualizado.PagadoEn),
				}, "recuperacion manual de carrito interrumpido")

				estacionID, estacionCodigo, estacionNombre := dbpkg.ResolveCarritoStationIdentity(carritoActualizado)
				if _, errMetric := dbpkg.RecordCarritoStationMetric(dbEmp, dbpkg.CarritoStationMetricInput{
					EmpresaID:       empresaID,
					CarritoID:       id,
					EstacionID:      estacionID,
					EstacionCodigo:  estacionCodigo,
					EstacionNombre:  estacionNombre,
					EventoOperacion: "sesion_recuperada",
					MetodoPago:      carritoActualizado.MetodoPago,
					Moneda:          carritoActualizado.Moneda,
					MontoTotal:      carritoActualizado.Total,
					MontoPagado:     carritoActualizado.TotalPagado,
					DevolucionTotal: carritoActualizado.DevolucionTotal,
					ActivadoEn:      carritoActualizado.ActivadoEn,
					PagadoEn:        carritoActualizado.PagadoEn,
					UsuarioCreador:  strings.TrimSpace(adminEmailFromRequest(r)),
					Observaciones:   "recuperacion operativa de sesion interrumpida",
				}); errMetric != nil {
					log.Printf("[carritos] metrica sesion_recuperada empresa_id=%d carrito_id=%d error: %v", empresaID, id, errMetric)
				}
				dispatchControlElectricoEstacionAsync(dbEmp, carritoActualizado, true, strings.TrimSpace(adminEmailFromRequest(r)), "recuperar_interrumpido")

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":             true,
					"estado":         normalizeCarritoRegistroEstado(carritoActualizado.Estado),
					"estado_carrito": normalizeCarritoOperativoEstado(carritoActualizado.EstadoCarrito),
					"estado_venta":   carritoActualizado.EstadoVenta,
					"activado_en":    strings.TrimSpace(carritoActualizado.ActivadoEn),
				})
				return
			}

			if action == "anular_venta" || action == "anular_venta_total" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}

				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get for anular_venta empresa_id=%d id=%d error: %v", empresaID, id, errCarrito)
					http.Error(w, "No se pudo validar estado del carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				if strings.EqualFold(strings.TrimSpace(carrito.EstadoCarrito), "anulado") {
					http.Error(w, "la venta ya esta anulada", http.StatusConflict)
					return
				}
				if !isCarritoVentaPagada(carrito) {
					http.Error(w, "solo se puede anular una venta pagada", http.StatusConflict)
					return
				}

				var payload struct {
					Motivo       string `json:"motivo"`
					Confirmar    bool   `json:"confirmar"`
					Confirmacion bool   `json:"confirmacion"`
				}
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
						http.Error(w, "JSON invalido", http.StatusBadRequest)
						return
					}
				}
				motivo := strings.TrimSpace(payload.Motivo)
				if len(motivo) < 5 {
					http.Error(w, "motivo es obligatorio para anular una venta (minimo 5 caracteres)", http.StatusBadRequest)
					return
				}
				if !payload.Confirmar && !payload.Confirmacion {
					http.Error(w, "confirmar=true es obligatorio para anular completamente la venta", http.StatusBadRequest)
					return
				}

				usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))
				carritoActualizado, totalPagadoAnterior, devolucionTotalNueva, errCancel := dbpkg.CancelCarritoSale(dbEmp, empresaID, id, motivo, usuarioOperacion)
				if errCancel != nil {
					http.Error(w, errCancel.Error(), http.StatusBadRequest)
					return
				}
				if carritoActualizado == nil {
					carritoActualizado = carrito
					carritoActualizado.Estado = "inactivo"
					carritoActualizado.EstadoCarrito = "anulado"
					carritoActualizado.EstadoVenta = "venta_anulada"
					carritoActualizado.TotalPagado = 0
					carritoActualizado.DevolucionTotal = devolucionTotalNueva
				}
				documentosAnulados := anularDocumentosVentaDesdeCarrito(dbEmp, carrito, motivo, usuarioOperacion)

				registrarEventoContableVentaCarrito(dbEmp, r, carritoActualizado, "venta_anulada", totalPagadoAnterior, map[string]interface{}{
					"action":                    action,
					"motivo":                    motivo,
					"total_pagado_anterior":     totalPagadoAnterior,
					"devolucion_total_anterior": carrito.DevolucionTotal,
					"devolucion_total_nueva":    devolucionTotalNueva,
					"metodo_pago":               strings.TrimSpace(carrito.MetodoPago),
					"estado_venta_anterior":     carrito.EstadoVenta,
					"estado_venta_nuevo":        "venta_anulada",
					"inventario_liberado":       true,
					"documentos_anulados":       documentosAnulados,
				}, "anulacion total de venta pagada")

				registrarAuditoriaCarritoOperacionNoBloqueante(dbEmp, r, empresaID, id, "anular_venta", http.StatusOK, map[string]interface{}{
					"motivo":                    motivo,
					"total_pagado_anterior":     totalPagadoAnterior,
					"devolucion_total_anterior": carrito.DevolucionTotal,
					"devolucion_total_nueva":    devolucionTotalNueva,
					"estado_venta":              carritoActualizado.EstadoVenta,
					"inventario_liberado":       true,
					"documentos_anulados":       documentosAnulados,
				}, "anulacion total de carrito pagado")

				estacionID, estacionCodigo, estacionNombre := dbpkg.ResolveCarritoStationIdentity(carritoActualizado)
				if _, errMetric := dbpkg.RecordCarritoStationMetric(dbEmp, dbpkg.CarritoStationMetricInput{
					EmpresaID:           empresaID,
					CarritoID:           id,
					EstacionID:          estacionID,
					EstacionCodigo:      estacionCodigo,
					EstacionNombre:      estacionNombre,
					EventoOperacion:     "venta_anulada",
					MetodoPago:          carrito.MetodoPago,
					Moneda:              carritoActualizado.Moneda,
					MontoTotal:          carritoActualizado.Total,
					MontoPagado:         0,
					MontoAnulado:        totalPagadoAnterior,
					DevolucionTotal:     devolucionTotalNueva,
					ActivadoEn:          carritoActualizado.ActivadoEn,
					PagadoEn:            carritoActualizado.PagadoEn,
					ReferenciaOperacion: motivo,
					UsuarioCreador:      usuarioOperacion,
					Observaciones:       "anulacion total de venta pagada",
				}); errMetric != nil {
					log.Printf("[carritos] metrica venta_anulada empresa_id=%d carrito_id=%d error: %v", empresaID, id, errMetric)
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                    true,
					"estado":                normalizeCarritoRegistroEstado(carritoActualizado.Estado),
					"estado_carrito":        normalizeCarritoOperativoEstado(carritoActualizado.EstadoCarrito),
					"estado_venta":          carritoActualizado.EstadoVenta,
					"motivo":                motivo,
					"total_pagado_anterior": totalPagadoAnterior,
					"total_pagado_nuevo":    0,
					"devolucion_total":      devolucionTotalNueva,
					"inventario_liberado":   true,
					"documentos_anulados":   documentosAnulados,
				})
				return
			}

			if action == "anular_cierre_parcial" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}

				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get for anular_cierre_parcial empresa_id=%d id=%d error: %v", empresaID, id, errCarrito)
					http.Error(w, "No se pudo validar estado del carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				if !isCarritoVentaPagada(carrito) {
					http.Error(w, "solo se puede anular parcialmente una venta pagada", http.StatusConflict)
					return
				}

				var payload struct {
					MontoAnulado float64 `json:"monto_anulado"`
					Monto        float64 `json:"monto"`
					Motivo       string  `json:"motivo"`
				}
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
						http.Error(w, "JSON invalido", http.StatusBadRequest)
						return
					}
				}

				montoAnulado := roundMoneyCarritoHandler(payload.MontoAnulado)
				if montoAnulado <= 0 {
					montoAnulado = roundMoneyCarritoHandler(payload.Monto)
				}
				if montoAnulado <= 0 {
					http.Error(w, "monto_anulado debe ser mayor a cero", http.StatusBadRequest)
					return
				}
				motivo := strings.TrimSpace(payload.Motivo)
				if motivo == "" {
					motivo = "anulacion parcial de cierre"
				}

				totalPagadoNuevo, devolucionTotalNueva, errCancel := dbpkg.CancelCarritoPartialClosure(dbEmp, empresaID, id, montoAnulado)
				if errCancel != nil {
					http.Error(w, errCancel.Error(), http.StatusBadRequest)
					return
				}

				carritoActualizado, errUpdated := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errUpdated != nil {
					log.Printf("[carritos] get after anular_cierre_parcial empresa_id=%d id=%d error: %v", empresaID, id, errUpdated)
					carritoActualizado = carrito
					carritoActualizado.TotalPagado = totalPagadoNuevo
					carritoActualizado.DevolucionTotal = devolucionTotalNueva
				}

				registrarEventoContableVentaCarrito(dbEmp, r, carritoActualizado, "venta_cierre_parcial_anulada", montoAnulado, map[string]interface{}{
					"action":                    "anular_cierre_parcial",
					"motivo":                    motivo,
					"monto_anulado":             montoAnulado,
					"total_pagado_anterior":     carrito.TotalPagado,
					"total_pagado_nuevo":        totalPagadoNuevo,
					"devolucion_total_anterior": carrito.DevolucionTotal,
					"devolucion_total_nueva":    devolucionTotalNueva,
					"metodo_pago":               strings.TrimSpace(carrito.MetodoPago),
				}, "anulacion parcial de cierre de venta")

				registrarAuditoriaCarritoOperacionNoBloqueante(dbEmp, r, empresaID, id, "anular_cierre_parcial", http.StatusOK, map[string]interface{}{
					"motivo":                    motivo,
					"monto_anulado":             montoAnulado,
					"total_pagado_anterior":     carrito.TotalPagado,
					"total_pagado_nuevo":        totalPagadoNuevo,
					"devolucion_total_anterior": carrito.DevolucionTotal,
					"devolucion_total_nueva":    devolucionTotalNueva,
					"estado_venta":              carritoActualizado.EstadoVenta,
				}, "anulacion parcial de cierre en carrito pagado")

				estacionID, estacionCodigo, estacionNombre := dbpkg.ResolveCarritoStationIdentity(carritoActualizado)
				if _, errMetric := dbpkg.RecordCarritoStationMetric(dbEmp, dbpkg.CarritoStationMetricInput{
					EmpresaID:           empresaID,
					CarritoID:           id,
					EstacionID:          estacionID,
					EstacionCodigo:      estacionCodigo,
					EstacionNombre:      estacionNombre,
					EventoOperacion:     "cierre_parcial_anulado",
					MetodoPago:          carritoActualizado.MetodoPago,
					Moneda:              carritoActualizado.Moneda,
					MontoTotal:          carritoActualizado.Total,
					MontoPagado:         totalPagadoNuevo,
					MontoAnulado:        montoAnulado,
					DevolucionTotal:     devolucionTotalNueva,
					ActivadoEn:          carritoActualizado.ActivadoEn,
					PagadoEn:            carritoActualizado.PagadoEn,
					ReferenciaOperacion: motivo,
					UsuarioCreador:      strings.TrimSpace(adminEmailFromRequest(r)),
					Observaciones:       "correccion rapida post-cobro",
				}); errMetric != nil {
					log.Printf("[carritos] metrica cierre_parcial_anulado empresa_id=%d carrito_id=%d error: %v", empresaID, id, errMetric)
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                    true,
					"estado":                normalizeCarritoRegistroEstado(carritoActualizado.Estado),
					"estado_carrito":        normalizeCarritoOperativoEstado(carritoActualizado.EstadoCarrito),
					"estado_venta":          carritoActualizado.EstadoVenta,
					"monto_anulado":         montoAnulado,
					"total_pagado_anterior": carrito.TotalPagado,
					"total_pagado_nuevo":    totalPagadoNuevo,
					"devolucion_total":      devolucionTotalNueva,
				})
				return
			}

			if action == "activar" || action == "desactivar" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get for estado empresa_id=%d id=%d action=%s error: %v", empresaID, id, action, errCarrito)
					http.Error(w, "No se pudo validar estado del carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				if err := validateCarritoTransitionForAction(carrito, action); err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				estado := "activo"
				estadoVenta := "venta_abierta"
				if action == "desactivar" {
					estado = "inactivo"
					estadoVenta = "venta_suspendida"
				}
				if err := dbpkg.SetCarritoCompraEstado(dbEmp, empresaID, id, estado); err != nil {
					log.Printf("[carritos] set estado empresa_id=%d id=%d estado=%s error: %v", empresaID, id, estado, err)
					http.Error(w, "No se pudo actualizar estado del carrito", http.StatusInternalServerError)
					return
				}
				evento := "venta_suspendida"
				if action == "activar" {
					evento = "venta_activada"
				}
				registrarEventoContableVentaCarrito(dbEmp, r, carrito, evento, carrito.Total, map[string]interface{}{
					"action":                action,
					"estado_registro_nuevo": estado,
					"estado_venta_anterior": carrito.EstadoVenta,
					"estado_venta_nuevo":    estadoVenta,
				}, "actualizacion de estado de venta")
				dispatchControlElectricoEstacionAsync(dbEmp, carrito, action == "activar", strings.TrimSpace(adminEmailFromRequest(r)), action)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado, "estado_venta": estadoVenta})
				return
			}

			if action == "cerrar" || action == "reabrir" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
				if errCarrito != nil {
					if errors.Is(errCarrito, sql.ErrNoRows) {
						http.Error(w, "carrito no encontrado", http.StatusNotFound)
						return
					}
					log.Printf("[carritos] get for estado_operacion empresa_id=%d id=%d action=%s error: %v", empresaID, id, action, errCarrito)
					http.Error(w, "No se pudo validar estado del carrito", http.StatusInternalServerError)
					return
				}
				if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				if err := validateCarritoTransitionForAction(carrito, action); err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				estadoCarrito := "abierto"
				estadoVenta := "venta_abierta"
				if action == "cerrar" {
					estadoCarrito = "cerrado"
					estadoVenta = "venta_cerrada"
				}
				if err := dbpkg.SetCarritoOperacionEstado(dbEmp, empresaID, id, estadoCarrito); err != nil {
					log.Printf("[carritos] set estado_operacion empresa_id=%d id=%d estado=%s error: %v", empresaID, id, estadoCarrito, err)
					http.Error(w, "No se pudo actualizar estado operativo del carrito", http.StatusInternalServerError)
					return
				}
				evento := "venta_reabierta"
				if action == "cerrar" {
					evento = "venta_cerrada"
				}
				registrarEventoContableVentaCarrito(dbEmp, r, carrito, evento, carrito.Total, map[string]interface{}{
					"action":                 action,
					"estado_operativo_nuevo": estadoCarrito,
					"estado_venta_anterior":  carrito.EstadoVenta,
					"estado_venta_nuevo":     estadoVenta,
				}, "actualizacion de estado operativo de venta")
				dispatchControlElectricoEstacionAsync(dbEmp, carrito, action == "reabrir", strings.TrimSpace(adminEmailFromRequest(r)), action)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado_carrito": estadoCarrito, "estado_venta": estadoVenta})
				return
			}

			var payload dbpkg.CarritoCompra
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := validateCarritoPayload(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			carritoActual, errCarritoActual := dbpkg.GetCarritoCompraByID(dbEmp, payload.EmpresaID, payload.ID)
			if errCarritoActual != nil {
				if errors.Is(errCarritoActual, sql.ErrNoRows) {
					http.Error(w, "carrito no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo validar el carrito", http.StatusInternalServerError)
				return
			}
			if err := ensureCarritoStationAccessForCarrito(dbEmp, payload.EmpresaID, adminEmailFromRequest(r), carritoActual); err != nil {
				writeCarritoStationAccessError(w, err)
				return
			}
			if err := dbpkg.UpdateCarritoCompra(dbEmp, payload); err != nil {
				log.Printf("[carritos] update empresa_id=%d id=%d error: %v", payload.EmpresaID, payload.ID, err)
				http.Error(w, "No se pudo actualizar el carrito", http.StatusBadRequest)
				return
			}
			if err := dbpkg.RecalculateCarritoCompraTotals(dbEmp, payload.EmpresaID, payload.ID); err != nil {
				log.Printf("[carritos] recalculate empresa_id=%d id=%d error: %v", payload.EmpresaID, payload.ID, err)
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			if isPorteroRestrictedCarritoRequest(r, "") {
				http.Error(w, "forbidden: el rol portero solo puede ver y activar estaciones", http.StatusForbidden)
				return
			}
			if isServicioLimpiezaRestrictedCarritoRequest(r, "") {
				http.Error(w, "forbidden: el rol Servicio de limpieza solo puede ver estaciones y reportar aseo", http.StatusForbidden)
				return
			}
			empresaID, errEmp := parseEmpresaIDQuery(r)
			if errEmp != nil {
				http.Error(w, errEmp.Error(), http.StatusBadRequest)
				return
			}
			id, errID := parseInt64Query(r, "id")
			if errID != nil {
				http.Error(w, errID.Error(), http.StatusBadRequest)
				return
			}
			carrito, errCarrito := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, id)
			if errCarrito != nil {
				if errors.Is(errCarrito, sql.ErrNoRows) {
					http.Error(w, "carrito no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo validar el carrito", http.StatusInternalServerError)
				return
			}
			if err := ensureCarritoStationAccessForCarrito(dbEmp, empresaID, adminEmailFromRequest(r), carrito); err != nil {
				writeCarritoStationAccessError(w, err)
				return
			}
			if err := dbpkg.DeleteCarritoCompra(dbEmp, empresaID, id); err != nil {
				log.Printf("[carritos] delete empresa_id=%d id=%d error: %v", empresaID, id, err)
				http.Error(w, "No se pudo eliminar el carrito", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// EmpresaCarritoItemsHandler gestiona CRUD de items dentro de un carrito.
func EmpresaCarritoItemsHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isStationBoardOnlyCarritoRequest(r) {
			http.Error(w, "forbidden: este rol no puede consultar ni modificar items del carrito", http.StatusForbidden)
			return
		}
		if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
			log.Printf("[carritos_items] ensure schema error: %v", err)
			http.Error(w, "No se pudo preparar el modulo de carritos", http.StatusInternalServerError)
			return
		}
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			carritoID, err := parseInt64Query(r, "carrito_id")
			if err != nil {
				http.Error(w, "carrito_id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := ensureCarritoStationAccessByID(dbEmp, empresaID, carritoID, adminEmailFromRequest(r)); err != nil {
				writeCarritoStationAccessError(w, err)
				return
			}
			includeInactive := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "1") ||
				strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "true")
			rows, err := dbpkg.GetCarritoCompraItems(dbEmp, empresaID, carritoID, includeInactive)
			if err != nil {
				log.Printf("[carritos_items] list empresa_id=%d carrito_id=%d error: %v", empresaID, carritoID, err)
				http.Error(w, "No se pudieron listar los items", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.CarritoCompraItem
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if err := validateCarritoItemPayload(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := ensureCarritoStationAccessByID(dbEmp, payload.EmpresaID, payload.CarritoID, adminEmailFromRequest(r)); err != nil {
				writeCarritoStationAccessError(w, err)
				return
			}
			if err := normalizeCarritoItemMoneyFields(dbEmp, &payload); err != nil {
				log.Printf("[carritos_items] normalize money create empresa_id=%d carrito_id=%d error: %v", payload.EmpresaID, payload.CarritoID, err)
				http.Error(w, "No se pudo validar la moneda del carrito", http.StatusInternalServerError)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.CreateCarritoCompraItem(dbEmp, payload)
			if err != nil {
				log.Printf("[carritos_items] create empresa_id=%d carrito_id=%d error: %v", payload.EmpresaID, payload.CarritoID, err)
				// Errores de stock conocidos
				if errors.Is(err, dbpkg.ErrStockInsuficiente) {
					http.Error(w, "Stock insuficiente para agregar el item al carrito", http.StatusBadRequest)
					return
				}
				// Mensaje específico cuando falta bodega de inventario para el producto
				lowerErr := strings.ToLower(strings.TrimSpace(err.Error()))
				if strings.Contains(lowerErr, "sin bodega") || strings.Contains(lowerErr, "sin bodega de inventario") {
					userMsg := `No fue posible agregar el producto al carrito porque no se encontró una bodega de inventario asociada al producto.
Pasos sugeridos:
1) Cree al menos una bodega para la empresa (tabla 'bodegas').
2) Asigne existencia para el producto en la bodega (tabla 'inventario_existencias'), por ejemplo:
   INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado, fecha_creacion, fecha_actualizacion, usuario_creador) VALUES (<EMPRESA_ID>, <PRODUCTO_ID>, <BODEGA_ID>, 10, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, '<USUARIO>');
3) Alternativamente, establezca 'bodega_principal_id' en la tabla 'productos':
   UPDATE productos SET bodega_principal_id = <BODEGA_ID> WHERE empresa_id = <EMPRESA_ID> AND id = <PRODUCTO_ID>;
4) Reintente agregar el producto al carrito.

Si necesita ayuda, consulte la sección de Inventario o contacte al administrador.`
					http.Error(w, userMsg, http.StatusBadRequest)
					return
				}
				http.Error(w, "No se pudo crear el item del carrito", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
				empresaID, errEmp := parseEmpresaIDQuery(r)
				if errEmp != nil {
					http.Error(w, errEmp.Error(), http.StatusBadRequest)
					return
				}
				carritoID, errCarr := parseInt64Query(r, "carrito_id")
				if errCarr != nil {
					http.Error(w, errCarr.Error(), http.StatusBadRequest)
					return
				}
				id, errID := parseInt64Query(r, "id")
				if errID != nil {
					http.Error(w, errID.Error(), http.StatusBadRequest)
					return
				}
				if err := ensureCarritoStationAccessByID(dbEmp, empresaID, carritoID, adminEmailFromRequest(r)); err != nil {
					writeCarritoStationAccessError(w, err)
					return
				}
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetCarritoCompraItemEstado(dbEmp, empresaID, carritoID, id, estado); err != nil {
					log.Printf("[carritos_items] set estado empresa_id=%d carrito_id=%d id=%d estado=%s error: %v", empresaID, carritoID, id, estado, err)
					if errors.Is(err, dbpkg.ErrStockInsuficiente) {
						http.Error(w, "Stock insuficiente para activar el item en carrito", http.StatusBadRequest)
						return
					}
					http.Error(w, "No se pudo actualizar el estado del item", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload dbpkg.CarritoCompraItem
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			if err := validateCarritoItemPayload(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := ensureCarritoStationAccessByID(dbEmp, payload.EmpresaID, payload.CarritoID, adminEmailFromRequest(r)); err != nil {
				writeCarritoStationAccessError(w, err)
				return
			}
			if err := normalizeCarritoItemMoneyFields(dbEmp, &payload); err != nil {
				log.Printf("[carritos_items] normalize money update empresa_id=%d carrito_id=%d id=%d error: %v", payload.EmpresaID, payload.CarritoID, payload.ID, err)
				http.Error(w, "No se pudo validar la moneda del carrito", http.StatusInternalServerError)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if err := dbpkg.UpdateCarritoCompraItem(dbEmp, payload); err != nil {
				log.Printf("[carritos_items] update empresa_id=%d carrito_id=%d id=%d error: %v", payload.EmpresaID, payload.CarritoID, payload.ID, err)
				if errors.Is(err, dbpkg.ErrStockInsuficiente) {
					http.Error(w, "Stock insuficiente para actualizar el item del carrito", http.StatusBadRequest)
					return
				}
				http.Error(w, "No se pudo actualizar el item del carrito", http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		case http.MethodDelete:
			empresaID, errEmp := parseEmpresaIDQuery(r)
			if errEmp != nil {
				http.Error(w, errEmp.Error(), http.StatusBadRequest)
				return
			}
			carritoID, errCarr := parseInt64Query(r, "carrito_id")
			if errCarr != nil {
				http.Error(w, errCarr.Error(), http.StatusBadRequest)
				return
			}
			id, errID := parseInt64Query(r, "id")
			if errID != nil {
				http.Error(w, errID.Error(), http.StatusBadRequest)
				return
			}
			if err := ensureCarritoStationAccessByID(dbEmp, empresaID, carritoID, adminEmailFromRequest(r)); err != nil {
				writeCarritoStationAccessError(w, err)
				return
			}
			if err := dbpkg.DeleteCarritoCompraItem(dbEmp, empresaID, carritoID, id); err != nil {
				log.Printf("[carritos_items] delete empresa_id=%d carrito_id=%d id=%d error: %v", empresaID, carritoID, id, err)
				http.Error(w, "No se pudo eliminar el item del carrito", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func validateCarritoPayload(payload dbpkg.CarritoCompra) error {
	if payload.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(payload.Nombre) == "" {
		return fmt.Errorf("nombre es obligatorio")
	}
	if payload.ClienteID < 0 {
		return fmt.Errorf("cliente_id invalido")
	}
	return nil
}

func carritoBoolFromConfigValue(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		s := strings.TrimSpace(strings.ToLower(v))
		return s == "1" || s == "true" || s == "si" || s == "sí" || s == "yes" || s == "on"
	case float64:
		return v != 0
	case int:
		return v != 0
	case int64:
		return v != 0
	case json.Number:
		n, _ := v.Int64()
		return n != 0
	default:
		return false
	}
}

func carritoBoolFromConfigValueDefault(cfg map[string]interface{}, key string, fallback bool) bool {
	if cfg == nil {
		return fallback
	}
	value, ok := cfg[key]
	if !ok {
		return fallback
	}
	return carritoBoolFromConfigValue(value)
}

func carritoParseConfigJSON(raw string) map[string]interface{} {
	var current interface{} = strings.TrimSpace(raw)
	for i := 0; i < 3; i++ {
		text, ok := current.(string)
		if !ok {
			break
		}
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}
		var next interface{}
		if err := json.Unmarshal([]byte(text), &next); err != nil {
			return nil
		}
		current = next
	}
	if cfg, ok := current.(map[string]interface{}); ok {
		return cfg
	}
	return nil
}

func carritoGlobalUIConfig(dbEmp *sql.DB, empresaID int64) (map[string]interface{}, error) {
	if dbEmp == nil || empresaID <= 0 {
		return nil, nil
	}
	pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "estaciones_config")
	if err != nil || pref == nil || strings.TrimSpace(pref.Valor) == "" {
		return nil, err
	}
	root := carritoParseConfigJSON(pref.Valor)
	if root == nil {
		return nil, nil
	}
	for _, key := range []string{"carrito_ui_global", "carrito", "carrito_configuracion_global"} {
		if cfg := carritoMapFromConfig(root[key]); cfg != nil {
			return cfg, nil
		}
	}
	return nil, nil
}

func carritoMetodoPagoHabilitadoPorEmpresa(dbEmp *sql.DB, empresaID int64, metodoPago string) (bool, error) {
	cfg, err := carritoGlobalUIConfig(dbEmp, empresaID)
	if err != nil || cfg == nil {
		return err == nil, err
	}
	switch dbpkg.NormalizeMetodoPagoCarrito(metodoPago) {
	case "efectivo":
		return carritoBoolFromConfigValueDefault(cfg, "metodo_pago_efectivo", true), nil
	case "tarjeta_credito":
		return carritoBoolFromConfigValueDefault(cfg, "metodo_pago_tarjeta_credito", true), nil
	case "tarjeta_debito":
		return carritoBoolFromConfigValueDefault(cfg, "metodo_pago_tarjeta_debito", true), nil
	case "transferencia_bancaria":
		return carritoBoolFromConfigValueDefault(cfg, "metodo_pago_transferencia_bancaria", true), nil
	case "transferencia_bre_b":
		return carritoBoolFromConfigValueDefault(cfg, "metodo_pago_transferencia_bre_b", true), nil
	case "transferencia_nequi":
		return carritoBoolFromConfigValueDefault(cfg, "metodo_pago_nequi", true), nil
	case "transferencia_otro":
		return carritoBoolFromConfigValueDefault(cfg, "metodo_pago_otras_transferencias", true), nil
	case "credito_cliente":
		return carritoBoolFromConfigValueDefault(cfg, "metodo_pago_credito_cliente", true), nil
	case "mixto":
		return carritoBoolFromConfigValueDefault(cfg, "permitir_pago_mixto", true), nil
	case "codigo_descuento":
		return carritoBoolFromConfigValueDefault(cfg, "mostrar_descuentos", false), nil
	default:
		return false, nil
	}
}

func carritoCreditoAmount(metodoPago string, pagosMixtos []carritoPagoMixtoNormalizado, total float64, moneda string) float64 {
	metodo := dbpkg.NormalizeMetodoPagoCarrito(metodoPago)
	if metodo == "credito_cliente" {
		return roundMoneyCarritoForMoneda(total, moneda)
	}
	if metodo != "mixto" {
		return 0
	}
	monto := 0.0
	for _, tramo := range pagosMixtos {
		if tramo.Metodo == "credito_cliente" {
			monto = roundMoneyCarritoForMoneda(monto+tramo.Monto, moneda)
		}
	}
	return monto
}

func validarCupoCreditoClienteParaVenta(dbEmp *sql.DB, carrito *dbpkg.CarritoCompra, montoCredito float64) (*carritoCreditoVentaResultado, error) {
	if carrito == nil || montoCredito <= 0 {
		return &carritoCreditoVentaResultado{Aplica: false}, nil
	}
	if carrito.ClienteID <= 0 {
		return nil, fmt.Errorf("para vender a credito debes seleccionar un cliente registrado con cupo activo")
	}
	if _, err := dbpkg.GetClienteByID(dbEmp, carrito.EmpresaID, carrito.ClienteID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("cliente de credito no encontrado para esta empresa")
		}
		return nil, err
	}
	disponibilidad, err := dbpkg.GetEmpresaCreditoClienteDisponibilidad(dbEmp, carrito.EmpresaID, carrito.ClienteID)
	if err != nil {
		return nil, err
	}
	if disponibilidad == nil || disponibilidad.Limite == nil {
		return nil, fmt.Errorf("el cliente no tiene cupo de credito activo")
	}
	if disponibilidad.SaldoDisponible+carritoMoneyTolerance(carrito.Moneda) < montoCredito {
		return nil, fmt.Errorf("cupo de credito insuficiente: disponible %s, venta a credito %s", formatCarritoMoneyForReference(disponibilidad.SaldoDisponible, carrito.Moneda), formatCarritoMoneyForReference(montoCredito, carrito.Moneda))
	}
	if disponibilidad.Estado != "" && disponibilidad.Estado != "disponible" {
		if disponibilidad.Mensaje != "" {
			return nil, fmt.Errorf("cupo de credito no disponible: %s", disponibilidad.Mensaje)
		}
		return nil, fmt.Errorf("cupo de credito no disponible")
	}
	return &carritoCreditoVentaResultado{
		Aplica:          true,
		ClienteID:       carrito.ClienteID,
		ClienteNombre:   strings.TrimSpace(carrito.ClienteNombre),
		MontoCredito:    roundMoneyCarritoForMoneda(montoCredito, carrito.Moneda),
		CupoLimite:      disponibilidad.Limite.LimiteSaldoTotal,
		SaldoPrevio:     disponibilidad.SaldoActual,
		SaldoDisponible: disponibilidad.SaldoDisponible,
		SaldoDespues:    roundMoneyCarritoForMoneda(disponibilidad.SaldoActual+montoCredito, carrito.Moneda),
	}, nil
}

func registrarCreditoVentaDesdeCarrito(dbEmp *sql.DB, carrito *dbpkg.CarritoCompra, montoCredito float64, usuario string, documentoVenta map[string]interface{}, prevalidado *carritoCreditoVentaResultado) (*carritoCreditoVentaResultado, error) {
	if carrito == nil || montoCredito <= 0 {
		return &carritoCreditoVentaResultado{Aplica: false}, nil
	}
	if existing, err := dbpkg.GetEmpresaCreditoByVentaOrigenID(dbEmp, carrito.EmpresaID, carrito.ID); err == nil && existing != nil {
		return &carritoCreditoVentaResultado{
			Aplica:           true,
			CreditoID:        existing.ID,
			Codigo:           existing.Codigo,
			ClienteID:        existing.ClienteID,
			ClienteNombre:    existing.ClienteNombre,
			MontoCredito:     existing.MontoAprobado,
			SaldoDespues:     existing.SaldoActual,
			FechaVencimiento: existing.FechaVencimiento,
			Warning:          "credito ya existia para esta venta",
		}, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	cliente, err := dbpkg.GetClienteByID(dbEmp, carrito.EmpresaID, carrito.ClienteID)
	if err != nil {
		return nil, err
	}
	fechaInicio := time.Now().In(time.Local)
	documentoOrigen := strings.TrimSpace(carrito.Codigo)
	if documentoVenta != nil {
		if v, ok := documentoVenta["documento_codigo"].(string); ok && strings.TrimSpace(v) != "" {
			documentoOrigen = strings.TrimSpace(v)
		}
	}
	if documentoOrigen == "" {
		documentoOrigen = fmt.Sprintf("CAR-%d", carrito.ID)
	}
	creditoID, err := dbpkg.CreateEmpresaCredito(dbEmp, dbpkg.EmpresaCredito{
		EmpresaID:             carrito.EmpresaID,
		ClienteID:             carrito.ClienteID,
		ClienteNombre:         strings.TrimSpace(cliente.NombreRazonSocial),
		TipoCredito:           "fijo",
		MontoAprobado:         montoCredito,
		CupoCredito:           montoCredito,
		SaldoActual:           montoCredito,
		TasaInteres:           0,
		TasaMora:              0,
		PeriodicidadCuota:     "mensual",
		ValorCuotaPactada:     montoCredito,
		PlazoDias:             30,
		PlazoCuotas:           1,
		FechaInicio:           fechaInicio.Format("2006-01-02"),
		FechaVencimiento:      fechaInicio.AddDate(0, 0, 30).Format("2006-01-02"),
		BloqueoAutomaticoMora: true,
		VentaOrigenID:         carrito.ID,
		DocumentoOrigen:       documentoOrigen,
		EstadoCredito:         "activo",
		UsuarioCreador:        strings.TrimSpace(usuario),
		Estado:                "activo",
		Observaciones:         "credito creado automaticamente desde venta del carrito",
	})
	if err != nil {
		return nil, err
	}
	row, err := dbpkg.GetEmpresaCreditoByID(dbEmp, carrito.EmpresaID, creditoID)
	if err != nil {
		return nil, err
	}
	result := &carritoCreditoVentaResultado{
		Aplica:           true,
		CreditoID:        row.ID,
		Codigo:           row.Codigo,
		ClienteID:        row.ClienteID,
		ClienteNombre:    row.ClienteNombre,
		MontoCredito:     row.MontoAprobado,
		SaldoDespues:     row.SaldoActual,
		FechaVencimiento: row.FechaVencimiento,
	}
	if prevalidado != nil {
		result.CupoLimite = prevalidado.CupoLimite
		result.SaldoPrevio = prevalidado.SaldoPrevio
		result.SaldoDisponible = prevalidado.SaldoDisponible
	}
	return result, nil
}

var errCarritoStationAccessDenied = errors.New("usuario sin acceso a esta estacion")

type carritoStationAccessPolicy struct {
	Enabled       bool
	LimitStations bool
	AllowCaja     bool
	Stations      map[int64]bool
}

func loadCarritoStationAccessPolicy(dbEmp *sql.DB, empresaID int64, usuario string) (carritoStationAccessPolicy, error) {
	policy := carritoStationAccessPolicy{AllowCaja: true, Stations: map[int64]bool{}}
	if dbEmp == nil || empresaID <= 0 {
		return policy, nil
	}
	pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "estaciones_config")
	if err != nil {
		return policy, err
	}
	if pref == nil || strings.TrimSpace(pref.Valor) == "" {
		return policy, nil
	}
	root := carritoParseConfigJSON(pref.Valor)
	if root == nil {
		return policy, nil
	}
	access := carritoMapFromConfig(root["acceso_estaciones_cajeros"])
	if access == nil {
		access = carritoMapFromConfig(root["acceso_estaciones_por_cajero"])
	}
	if access == nil {
		return policy, nil
	}
	policy.Enabled = carritoBoolFromConfigValue(access["habilitado"]) || carritoBoolFromConfigValue(access["enabled"])
	if !policy.Enabled {
		return policy, nil
	}
	email := strings.ToLower(strings.TrimSpace(usuario))
	users := carritoMapFromConfig(access["usuarios"])
	if users == nil || email == "" {
		return policy, nil
	}
	entry := carritoMapFromConfig(users[email])
	if entry == nil {
		return policy, nil
	}
	policy.AllowCaja = true
	if _, exists := entry["ver_caja"]; exists {
		policy.AllowCaja = carritoBoolFromConfigValue(entry["ver_caja"])
	}
	if _, exists := entry["caja"]; exists {
		policy.AllowCaja = carritoBoolFromConfigValue(entry["caja"])
	}
	policy.LimitStations = carritoBoolFromConfigValue(entry["limitar_estaciones"])
	if _, exists := entry["restringir_estaciones"]; exists {
		policy.LimitStations = carritoBoolFromConfigValue(entry["restringir_estaciones"])
	}
	stationIDs := carritoInt64SliceFromConfig(entry["estaciones"])
	if len(stationIDs) == 0 {
		stationIDs = carritoInt64SliceFromConfig(entry["station_ids"])
	}
	if _, hasExplicitLimit := entry["limitar_estaciones"]; len(stationIDs) > 0 && !hasExplicitLimit {
		policy.LimitStations = true
	}
	for _, id := range stationIDs {
		if id > 0 {
			policy.Stations[id] = true
		}
	}
	return policy, nil
}

func carritoMapFromConfig(value interface{}) map[string]interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return v
	default:
		return nil
	}
}

func carritoInt64SliceFromConfig(value interface{}) []int64 {
	out := make([]int64, 0)
	switch v := value.(type) {
	case []interface{}:
		for _, item := range v {
			if n := carritoInt64FromConfig(item); n > 0 {
				out = append(out, n)
			}
		}
	case []int64:
		out = append(out, v...)
	case []int:
		for _, item := range v {
			if item > 0 {
				out = append(out, int64(item))
			}
		}
	case string:
		parts := strings.FieldsFunc(v, func(r rune) bool {
			return r == ',' || r == ';' || r == ' '
		})
		for _, part := range parts {
			if n, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64); err == nil && n > 0 {
				out = append(out, n)
			}
		}
	}
	return out
}

func carritoInt64FromConfig(value interface{}) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int:
		return int64(v)
	case int64:
		return v
	case json.Number:
		n, _ := v.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n
	default:
		return 0
	}
}

func ensureCarritoStationAccessForStation(dbEmp *sql.DB, empresaID int64, usuario string, estacionID int64) error {
	if estacionID <= 0 {
		return nil
	}
	policy, err := loadCarritoStationAccessPolicy(dbEmp, empresaID, usuario)
	if err != nil {
		return err
	}
	if !policy.Enabled || !policy.LimitStations {
		return nil
	}
	if policy.Stations[estacionID] {
		return nil
	}
	return errCarritoStationAccessDenied
}

func ensureCarritoStationAccessForCarrito(dbEmp *sql.DB, empresaID int64, usuario string, carrito *dbpkg.CarritoCompra) error {
	estacionID, _, _ := dbpkg.ResolveCarritoStationIdentity(carrito)
	return ensureCarritoStationAccessForStation(dbEmp, empresaID, usuario, estacionID)
}

func ensureCarritoStationCajaAccess(dbEmp *sql.DB, empresaID int64, usuario string) error {
	policy, err := loadCarritoStationAccessPolicy(dbEmp, empresaID, usuario)
	if err != nil {
		return err
	}
	if !policy.Enabled {
		return nil
	}
	if policy.AllowCaja {
		return nil
	}
	return errCarritoStationAccessDenied
}

func carritoTransferenciaEstacionHabilitada(dbEmp *sql.DB, empresaID int64) bool {
	if dbEmp == nil || empresaID <= 0 {
		return false
	}
	pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "estaciones_config")
	if err != nil || pref == nil || strings.TrimSpace(pref.Valor) == "" {
		return false
	}
	root := carritoParseConfigJSON(pref.Valor)
	if root == nil {
		return false
	}
	for _, key := range []string{"carrito_ui_global", "carrito", "carrito_configuracion_global"} {
		cfg := carritoMapFromConfig(root[key])
		if cfg == nil {
			continue
		}
		for _, flag := range []string{
			"permitir_transferir_cuenta_carrito",
			"permitir_transferencia_cuenta_carrito",
			"transferencia_cuenta_habilitada",
			"mostrar_boton_transferir_cuenta_carrito",
		} {
			if carritoBoolFromConfigValue(cfg[flag]) {
				return true
			}
		}
	}
	return false
}

func ensureCarritoStationAccessByID(dbEmp *sql.DB, empresaID, carritoID int64, usuario string) error {
	carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, carritoID)
	if err != nil {
		return err
	}
	return ensureCarritoStationAccessForCarrito(dbEmp, empresaID, usuario, carrito)
}

func filterCarritosByStationAccess(dbEmp *sql.DB, empresaID int64, usuario string, rows []dbpkg.CarritoCompra) ([]dbpkg.CarritoCompra, error) {
	policy, err := loadCarritoStationAccessPolicy(dbEmp, empresaID, usuario)
	if err != nil {
		return rows, err
	}
	if !policy.Enabled || !policy.LimitStations {
		return rows, nil
	}
	filtered := make([]dbpkg.CarritoCompra, 0, len(rows))
	for _, row := range rows {
		estacionID, _, _ := dbpkg.ResolveCarritoStationIdentity(&row)
		if estacionID <= 0 || policy.Stations[estacionID] {
			filtered = append(filtered, row)
		}
	}
	return filtered, nil
}

func writeCarritoStationAccessError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	if errors.Is(err, errCarritoStationAccessDenied) {
		http.Error(w, "No tienes acceso operativo a esta estacion", http.StatusForbidden)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Carrito no encontrado", http.StatusNotFound)
		return
	}
	http.Error(w, "No se pudo validar el acceso a estaciones", http.StatusInternalServerError)
}

func carritoClienteObligatorioParaPago(dbEmp *sql.DB, empresaID int64) bool {
	if dbEmp == nil || empresaID <= 0 {
		return false
	}
	pref, err := dbpkg.GetEmpresaEstacionPref(dbEmp, empresaID, 0, "estaciones_config")
	if err != nil || pref == nil || strings.TrimSpace(pref.Valor) == "" {
		return false
	}
	root := carritoParseConfigJSON(pref.Valor)
	if root == nil {
		return false
	}
	for _, key := range []string{"carrito_ui_global", "carrito", "carrito_configuracion_global"} {
		cfg, ok := root[key].(map[string]interface{})
		if !ok {
			continue
		}
		if carritoBoolFromConfigValue(cfg["cliente_obligatorio_pago"]) {
			return true
		}
	}
	return false
}

func normalizeCarritoCajaCode(value string) string {
	code := strings.ToUpper(strings.TrimSpace(value))
	code = strings.ReplaceAll(code, " ", "_")
	var b strings.Builder
	for _, r := range code {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	code = strings.Trim(b.String(), "_-")
	if code == "" {
		return "CAJA-1"
	}
	return code
}

func openCajaCobroForCarrito(dbEmp *sql.DB, dbSuper *sql.DB, empresaID int64, cajaCodigo, turno string, sucursalID int64, apertura float64, moneda, usuario string) (*dbpkg.EmpresaCierreCaja, bool, error) {
	if empresaID <= 0 {
		return nil, false, fmt.Errorf("empresa_id es obligatorio")
	}
	code := normalizeCarritoCajaCode(cajaCodigo)
	turno = strings.ToLower(strings.TrimSpace(turno))
	if turno == "" {
		turno = "general"
	}
	if sucursalID < 0 {
		sucursalID = 0
	}
	if apertura < 0 {
		apertura = 0
	}
	moneda = strings.ToUpper(strings.TrimSpace(moneda))
	if moneda == "" {
		moneda = "COP"
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	if existing, err := dbpkg.GetEmpresaCierreCajaAbiertaUsuario(dbEmp, empresaID, 0, code, turno, sucursalID, usuario); err == nil && existing != nil {
		return existing, false, nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}
	if _, _, err := validarCupoCajasLicencia(dbEmp, dbSuper, empresaID, 0); err != nil {
		return nil, false, err
	}
	now := time.Now()
	cierreID, err := dbpkg.CreateEmpresaCierreCaja(dbEmp, dbpkg.EmpresaCierreCaja{
		EmpresaID:        empresaID,
		SucursalID:       sucursalID,
		CajaCodigo:       code,
		Turno:            turno,
		FechaOperacion:   now.Format("2006-01-02"),
		FechaApertura:    now.Format("2006-01-02 15:04:05"),
		EstadoCierre:     "abierto",
		AperturaMonto:    apertura,
		CajaTeorica:      apertura,
		CajaFisica:       apertura,
		Moneda:           moneda,
		UsuarioCreador:   usuario,
		Estado:           "activo",
		Observaciones:    "Caja abierta automaticamente al pagar carrito",
		UmbralIncidencia: 0,
	})
	if err != nil {
		return nil, false, err
	}
	created, err := dbpkg.GetEmpresaCierreCajaAbiertaUsuario(dbEmp, empresaID, cierreID, "", "", 0, usuario)
	if err != nil {
		return nil, false, err
	}
	return created, true, nil
}

func attachCarritoStationRuntimeSummaries(dbEmp *sql.DB, rows []dbpkg.CarritoCompra) {
	if dbEmp == nil || len(rows) == 0 {
		return
	}
	now := time.Now()
	for i := range rows {
		estadoRegistro := normalizeCarritoRegistroEstado(rows[i].Estado)
		estadoOperativo := normalizeCarritoOperativoEstado(rows[i].EstadoCarrito)
		if estadoRegistro != "activo" || estadoOperativo != "abierto" || strings.TrimSpace(rows[i].PagadoEn) != "" {
			continue
		}
		estacionID, _, _ := dbpkg.ResolveCarritoStationIdentity(&rows[i])
		if estacionID <= 0 {
			continue
		}
		minutos, err := dbpkg.ResolveCarritoTarifaPorMinutosResumen(dbEmp, rows[i], now)
		if err != nil {
			log.Printf("[carritos] resumen tarifa minutos omitido empresa_id=%d carrito_id=%d error: %v", rows[i].EmpresaID, rows[i].ID, err)
		}
		if minutos != nil {
			rows[i].TarifaPorMinutos = minutos
			continue
		}
		dia, err := dbpkg.ResolveCarritoTarifaPorDiaResumen(dbEmp, rows[i], now)
		if err != nil {
			log.Printf("[carritos] resumen tarifa dia omitido empresa_id=%d carrito_id=%d error: %v", rows[i].EmpresaID, rows[i].ID, err)
			continue
		}
		if dia != nil {
			rows[i].TarifaPorDia = dia
		}
	}
}

func normalizeCarritoRegistroEstado(v string) string {
	trim := strings.TrimSpace(strings.ToLower(v))
	if trim == "" {
		return "activo"
	}
	return trim
}

func normalizeCarritoOperativoEstado(v string) string {
	trim := strings.TrimSpace(strings.ToLower(v))
	if trim == "" {
		return "abierto"
	}
	return trim
}

func isCarritoVentaPagada(carrito *dbpkg.CarritoCompra) bool {
	if carrito == nil {
		return false
	}
	estadoCarrito := strings.TrimSpace(strings.ToLower(carrito.EstadoCarrito))
	if estadoCarrito == "anulado" || estadoCarrito == "anulada" {
		return false
	}
	return strings.TrimSpace(carrito.PagadoEn) != ""
}

func isCarritoOperativoActivo(carrito *dbpkg.CarritoCompra) bool {
	if carrito == nil || isCarritoVentaPagada(carrito) {
		return false
	}
	estadoRegistro := normalizeCarritoRegistroEstado(carrito.Estado)
	estadoOperativo := normalizeCarritoOperativoEstado(carrito.EstadoCarrito)
	return estadoRegistro == "activo" && estadoOperativo == "abierto"
}

func validateCarritoTransitionForAction(carrito *dbpkg.CarritoCompra, action string) error {
	if carrito == nil {
		return fmt.Errorf("carrito no encontrado")
	}
	estadoRegistro := normalizeCarritoRegistroEstado(carrito.Estado)
	estadoOperativo := normalizeCarritoOperativoEstado(carrito.EstadoCarrito)
	pagada := isCarritoVentaPagada(carrito)

	switch strings.TrimSpace(strings.ToLower(action)) {
	case "pagar_estacion":
		if pagada {
			return fmt.Errorf("la venta ya fue pagada")
		}
		if estadoRegistro != "activo" {
			return fmt.Errorf("solo se puede pagar una venta activa")
		}

	case "cerrar":
		if pagada {
			return fmt.Errorf("la venta ya fue pagada")
		}
		// Permitir cerrar la operación aunque el registro esté marcado como inactivo.
		// Esto evita un conflicto 409 cuando el cliente primero desactiva el registro
		// y luego intenta cerrar la sesión operativa. Seguimos impidiendo cerrar
		// si la sesión operativa ya está cerrada.
		if estadoOperativo == "cerrado" {
			return fmt.Errorf("la venta ya se encuentra cerrada")
		}

	case "reabrir":
		if pagada {
			return fmt.Errorf("no se puede reabrir una venta pagada")
		}
		if estadoRegistro != "activo" {
			return fmt.Errorf("solo se puede reabrir una venta activa")
		}
		if estadoOperativo != "cerrado" {
			return fmt.Errorf("solo se puede reabrir una venta cerrada")
		}

	case "activar":
		if pagada {
			return fmt.Errorf("no se puede activar una venta pagada; use activar_estacion para iniciar una nueva sesion")
		}
		if estadoRegistro == "activo" {
			return fmt.Errorf("la venta ya se encuentra activa")
		}

	case "desactivar":
		if pagada {
			return fmt.Errorf("la venta pagada ya se encuentra inactiva")
		}
		if estadoRegistro != "activo" {
			return fmt.Errorf("la venta ya se encuentra inactiva")
		}
	}

	return nil
}

func registrarEventoContableVentaCarrito(dbEmp *sql.DB, r *http.Request, carrito *dbpkg.CarritoCompra, evento string, monto float64, payload map[string]interface{}, observaciones string) {
	if dbEmp == nil || carrito == nil || strings.TrimSpace(evento) == "" {
		return
	}
	if monto <= 0 {
		monto = carrito.Total
	}
	registrarEventoContableNoBloqueante(dbEmp, r, "carritos", dbpkg.EmpresaEventoContable{
		EmpresaID:       carrito.EmpresaID,
		Modulo:          "ventas",
		Evento:          evento,
		Entidad:         "carrito_compra",
		EntidadID:       carrito.ID,
		DocumentoTipo:   "carrito",
		DocumentoCodigo: strings.TrimSpace(carrito.Codigo),
		MontoTotal:      monto,
		Moneda:          strings.TrimSpace(carrito.Moneda),
		Origen:          "api_carritos_compra",
		Estado:          "activo",
		Observaciones:   strings.TrimSpace(observaciones),
	}, payload)
}

func registrarAuditoriaCarritoOperacionNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID, carritoID int64, accion string, statusCode int, metadata map[string]interface{}, observaciones string) {
	if dbEmp == nil || empresaID <= 0 || carritoID <= 0 {
		return
	}
	accion = strings.TrimSpace(strings.ToLower(accion))
	if accion == "" {
		return
	}

	metadataJSON := "{}"
	if metadata != nil {
		if raw, err := json.Marshal(metadata); err != nil {
			log.Printf("[carritos] auditoria metadata invalida empresa_id=%d carrito_id=%d accion=%s error=%v", empresaID, carritoID, accion, err)
		} else {
			metadataJSON = string(raw)
		}
	}

	auditoria := dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "ventas",
		Accion:         accion,
		Recurso:        "carritos_compra",
		RecursoID:      carritoID,
		MetodoHTTP:     strings.ToUpper(strings.TrimSpace(r.Method)),
		Endpoint:       strings.TrimSpace(r.URL.Path),
		Resultado:      resolveAuditoriaResultado(statusCode),
		CodigoHTTP:     int64(statusCode),
		RequestID:      resolveAuditoriaRequestID(r),
		IPOrigen:       resolveAuditoriaIP(r),
		UserAgent:      strings.TrimSpace(r.UserAgent()),
		MetadataJSON:   metadataJSON,
		RetencionDias:  normalizeRetencionDiasForHandler(0),
		UsuarioCreador: strings.TrimSpace(adminEmailFromRequest(r)),
		Estado:         "activo",
		Observaciones:  strings.TrimSpace(observaciones),
	}

	if _, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, auditoria); err != nil {
		log.Printf("[carritos] auditoria omitida empresa_id=%d carrito_id=%d accion=%s error=%v", empresaID, carritoID, accion, err)
	}
}

func validateCarritoItemPayload(payload dbpkg.CarritoCompraItem) error {
	if payload.EmpresaID <= 0 {
		return fmt.Errorf("empresa_id es obligatorio")
	}
	if payload.CarritoID <= 0 {
		return fmt.Errorf("carrito_id es obligatorio")
	}
	if strings.TrimSpace(payload.Descripcion) == "" {
		return fmt.Errorf("descripcion es obligatoria")
	}
	if !isNaturalCarritoQuantity(payload.Cantidad) {
		return fmt.Errorf("cantidad debe ser un numero natural positivo")
	}
	if payload.PrecioUnitario < 0 {
		return fmt.Errorf("precio_unitario invalido")
	}
	tipoItem := strings.TrimSpace(strings.ToLower(payload.TipoItem))
	if tipoItem == "receta" && payload.ReferenciaID <= 0 {
		return fmt.Errorf("referencia_id es obligatoria para tipo_item receta")
	}
	return nil
}

func normalizeCarritoItemMoneyFields(dbConn *sql.DB, payload *dbpkg.CarritoCompraItem) error {
	if dbConn == nil || payload == nil || payload.EmpresaID <= 0 || payload.CarritoID <= 0 {
		return nil
	}
	carrito, err := dbpkg.GetCarritoCompraByID(dbConn, payload.EmpresaID, payload.CarritoID)
	if err != nil {
		return err
	}
	payload.PrecioUnitario = roundMoneyCarritoForMoneda(payload.PrecioUnitario, carrito.Moneda)
	return nil
}

func isNaturalCarritoQuantity(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 1 && math.Trunc(value) == value
}

func roundMoneyCarritoHandler(v float64) float64 {
	return math.Round(v*100) / 100
}

func carritoUsesWholeMoney(moneda string) bool {
	return strings.EqualFold(strings.TrimSpace(moneda), "COP")
}

func roundMoneyCarritoForMoneda(v float64, moneda string) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	if carritoUsesWholeMoney(moneda) {
		return math.Round(v)
	}
	return roundMoneyCarritoHandler(v)
}

func carritoMetodoPagoRequiereReferencia(metodoPago string) bool {
	switch dbpkg.NormalizeMetodoPagoCarrito(metodoPago) {
	case "tarjeta_credito", "tarjeta_debito", "transferencia_bancaria", "transferencia_bre_b", "transferencia_nequi", "transferencia_otro":
		return true
	default:
		return false
	}
}

func carritoMoneyTolerance(moneda string) float64 {
	if carritoUsesWholeMoney(moneda) {
		return 0.01
	}
	return 0.01
}

func formatCarritoMoneyForReference(v float64, moneda string) string {
	if carritoUsesWholeMoney(moneda) {
		return fmt.Sprintf("%.0f", roundMoneyCarritoForMoneda(v, moneda))
	}
	return fmt.Sprintf("%.2f", roundMoneyCarritoForMoneda(v, moneda))
}

func normalizePagosMixtosCarrito(entries []carritoPagoMixtoEntrada, moneda string) ([]carritoPagoMixtoNormalizado, float64, error) {
	if len(entries) == 0 {
		return nil, 0, fmt.Errorf("pago mixto requiere detalle de pagos_mixtos")
	}

	normalized := make([]carritoPagoMixtoNormalizado, 0, len(entries))
	total := 0.0
	for _, item := range entries {
		metodo := dbpkg.NormalizeMetodoPagoCarrito(item.Metodo)
		if metodo == "" || metodo == "mixto" || metodo == "codigo_descuento" {
			return nil, 0, fmt.Errorf("pago mixto solo permite efectivo, tarjeta_credito, tarjeta_debito, transferencia_bre_b, transferencia_nequi, transferencia_otro y credito_cliente")
		}
		monto := roundMoneyCarritoForMoneda(item.Monto, moneda)
		if monto <= 0 {
			continue
		}
		referencia := strings.TrimSpace(item.Referencia)
		if carritoMetodoPagoRequiereReferencia(metodo) && len(referencia) < 4 {
			return nil, 0, fmt.Errorf("cada pago con tarjeta o transferencia en pago mixto requiere referencia minima de 4 caracteres")
		}

		normalized = append(normalized, carritoPagoMixtoNormalizado{
			Metodo:     metodo,
			Monto:      monto,
			Referencia: referencia,
		})
		total = roundMoneyCarritoForMoneda(total+monto, moneda)
	}

	if len(normalized) < 2 {
		return nil, 0, fmt.Errorf("pago mixto requiere al menos 2 metodos con monto mayor a cero")
	}

	return normalized, total, nil
}

func buildReferenciaPagoMixto(pagos []carritoPagoMixtoNormalizado, moneda string) string {
	if len(pagos) == 0 {
		return ""
	}
	parts := make([]string, 0, len(pagos))
	for _, item := range pagos {
		chunk := item.Metodo + ":" + formatCarritoMoneyForReference(item.Monto, moneda)
		if strings.TrimSpace(item.Referencia) != "" {
			chunk += "(ref:" + strings.TrimSpace(item.Referencia) + ")"
		}
		parts = append(parts, chunk)
	}
	return "mixto[" + strings.Join(parts, " | ") + "]"
}

func pagosMixtosToEventPayload(pagos []carritoPagoMixtoNormalizado) []map[string]interface{} {
	if len(pagos) == 0 {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(pagos))
	for _, item := range pagos {
		out = append(out, map[string]interface{}{
			"metodo":     item.Metodo,
			"monto":      roundMoneyCarritoHandler(item.Monto),
			"referencia": strings.TrimSpace(item.Referencia),
		})
	}
	return out
}

func normalizeVentaDocumentMode(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "factura_electronica" {
		return "factura_electronica"
	}
	return "comprobante_pago"
}

func extractVentaDocumentoBase(raw string) string {
	base := strings.ToUpper(strings.TrimSpace(raw))
	base = strings.ReplaceAll(base, " ", "")
	switch {
	case strings.HasPrefix(base, "CP-"), strings.HasPrefix(base, "FV-"):
		base = strings.TrimSpace(base[3:])
	}
	if base == "" {
		base = "VENTA"
	}
	return base
}

func buildVentaDocumentoCodigoFromBase(base, modo string) string {
	base = extractVentaDocumentoBase(base)
	if normalizeVentaDocumentMode(modo) == "factura_electronica" {
		return "FV-" + base
	}
	return "CP-" + base
}

func buildVentaDocumentoCodigo(carrito *dbpkg.CarritoCompra, modo string) string {
	base := ""
	if carrito != nil {
		base = strings.TrimSpace(carrito.Codigo)
		if base == "" && carrito.ID > 0 {
			base = fmt.Sprintf("CRT-%d", carrito.ID)
		}
	}
	return buildVentaDocumentoCodigoFromBase(base, modo)
}

func carritoAutoFacturaElectronicaActiva(cfg *dbpkg.EmpresaConfiguracionAvanzada) bool {
	if cfg == nil {
		return false
	}
	return normalizeVentaDocumentMode(cfg.ModoDocumentoVenta) == "factura_electronica"
}

func facturaElectronicaVentaRequiereAcuseFiscal(doc *dbpkg.EmpresaDocumentoFacturacion, resultado facturacionIntegracionResultado) bool {
	if doc == nil || !resultado.Aplica {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "factura_electronica") {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(doc.PaisCodigo), "CO") &&
		strings.EqualFold(strings.TrimSpace(doc.AmbienteFE), "produccion")
}

func facturaElectronicaVentaIntegracionConfirmada(resultado facturacionIntegracionResultado) bool {
	return normalizeFacturacionEstadoEnvio(resultado.EstadoEnvio) == "enviado"
}

func registrarFacturaElectronicaDesdeDocumentoVenta(dbEmp, dbSuper *sql.DB, ventaDoc *dbpkg.EmpresaDocumentoFacturacion, usuario, observaciones string) (map[string]interface{}, error) {
	if dbEmp == nil || ventaDoc == nil || ventaDoc.EmpresaID <= 0 || strings.TrimSpace(ventaDoc.DocumentoCodigo) == "" {
		return nil, nil
	}

	cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, ventaDoc.EmpresaID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &dbpkg.EmpresaConfiguracionAvanzada{EmpresaID: ventaDoc.EmpresaID}
	}

	documentoCodigo := buildVentaDocumentoCodigoFromBase(ventaDoc.DocumentoCodigo, "factura_electronica")
	existingDoc, existingErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, ventaDoc.EmpresaID, "factura_electronica", documentoCodigo)
	if existingErr != nil && !errors.Is(existingErr, sql.ErrNoRows) {
		return nil, existingErr
	}
	if existingDoc != nil {
		return map[string]interface{}{
			"ok":                true,
			"ya_existia":        true,
			"modo":              "factura_electronica",
			"requiere_dian":     true,
			"documento_id":      existingDoc.ID,
			"tipo_documento":    existingDoc.TipoDocumento,
			"documento_codigo":  existingDoc.DocumentoCodigo,
			"estado_documento":  existingDoc.EstadoDocumento,
			"numero_legal":      existingDoc.NumeroLegal,
			"codigo_validacion": existingDoc.CodigoValidacion,
			"pais_codigo":       existingDoc.PaisCodigo,
			"ambiente_fe":       existingDoc.AmbienteFE,
		}, nil
	}

	periodoContable := strings.TrimSpace(ventaDoc.PeriodoContable)
	if periodoContable == "" {
		periodoContable = time.Now().Format("2006-01")
	}
	fechaDocumento := strings.TrimSpace(ventaDoc.FechaDocumento)
	if fechaDocumento == "" {
		fechaDocumento = time.Now().Format("2006-01-02 15:04:05")
	}
	observacionBase := strings.TrimSpace(observaciones)
	if observacionBase == "" {
		observacionBase = "factura electronica generada desde la venta " + strings.TrimSpace(ventaDoc.DocumentoCodigo)
	}

	docPayload := dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:            ventaDoc.EmpresaID,
		TipoDocumento:        "factura_electronica",
		DocumentoCodigo:      documentoCodigo,
		EstadoDocumento:      "emitida",
		EstadoAnterior:       "borrador",
		EventoUltimo:         "factura_emitida",
		PeriodoContable:      periodoContable,
		MontoTotal:           ventaDoc.MontoTotal,
		Moneda:               strings.TrimSpace(ventaDoc.Moneda),
		FechaDocumento:       fechaDocumento,
		EntidadRelacionadaID: ventaDoc.EntidadRelacionadaID,
		UsuarioCreador:       strings.TrimSpace(usuario),
		Estado:               "activo",
		Observaciones:        observacionBase,
	}

	warning := ""
	legalDoc, legalErr := dbpkg.PrepareFacturacionDocumentoLegal(dbEmp, ventaDoc.EmpresaID, strings.TrimSpace(ventaDoc.PaisCodigo), documentoCodigo, ventaDoc.MontoTotal, ventaDoc.Moneda)
	if legalErr != nil {
		warning = legalErr.Error()
		docPayload.EstadoDocumento = "pendiente_emision"
		docPayload.EventoUltimo = "factura_pendiente_emision"
		docPayload.Observaciones = strings.TrimSpace(docPayload.Observaciones + ". Pendiente de emision legal: " + warning)
	} else if legalDoc != nil {
		docPayload.NumeroLegal = legalDoc.NumeroLegal
		docPayload.CodigoValidacion = legalDoc.CodigoValidacion
		docPayload.PaisCodigo = legalDoc.PaisCodigo
		docPayload.AmbienteFE = legalDoc.Ambiente
		docPayload.FechaDocumento = legalDoc.FechaEmisionLegal
	}

	docPersistido, err := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, docPayload)
	if err != nil {
		return nil, err
	}

	integracionFiscal := map[string]interface{}{}
	if strings.EqualFold(strings.TrimSpace(docPersistido.EstadoDocumento), "emitida") {
		payloadOperacion := facturacionOperacionPayload{
			EmpresaID:       ventaDoc.EmpresaID,
			EntidadID:       ventaDoc.EntidadRelacionadaID,
			ClienteID:       ventaDoc.EntidadRelacionadaID,
			TipoDocumento:   "factura_electronica",
			PaisCodigo:      strings.TrimSpace(docPersistido.PaisCodigo),
			DocumentoCodigo: strings.TrimSpace(docPersistido.DocumentoCodigo),
			EstadoActual:    strings.TrimSpace(docPersistido.EstadoDocumento),
			MontoTotal:      ventaDoc.MontoTotal,
			Moneda:          strings.TrimSpace(docPersistido.Moneda),
			PeriodoContable: periodoContable,
			Observaciones:   observacionBase,
		}
		resultadoIntegracion, retryItem, integErr := processFacturacionIntegracionForDocumento(
			dbEmp,
			payloadOperacion,
			*docPersistido,
			"emitir",
			strings.TrimSpace(usuario),
		)
		if integErr != nil {
			integracionFiscal["error"] = integErr.Error()
		}
		integracionFiscal["resultado"] = resultadoIntegracion
		if retryItem != nil {
			integracionFiscal["cola_reintentos"] = retryItem
		}
		if facturaElectronicaVentaRequiereAcuseFiscal(docPersistido, resultadoIntegracion) && !facturaElectronicaVentaIntegracionConfirmada(resultadoIntegracion) {
			motivo := strings.TrimSpace(resultadoIntegracion.Error)
			if motivo == "" {
				motivo = "integracion fiscal DIAN/proveedor no confirmada"
			}
			docPendiente := *docPersistido
			docPendiente.EstadoAnterior = strings.TrimSpace(docPersistido.EstadoDocumento)
			docPendiente.EstadoDocumento = "pendiente_emision"
			docPendiente.EventoUltimo = "factura_integracion_fallida"
			docPendiente.Observaciones = strings.TrimSpace(docPendiente.Observaciones + ". Pendiente de acuse fiscal: " + motivo)
			docActualizado, upErr := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, docPendiente)
			if upErr != nil {
				integracionFiscal["actualizacion_estado_error"] = upErr.Error()
			} else if docActualizado != nil {
				docPersistido = docActualizado
				warning = strings.TrimSpace(facturacionFirstNonBlank(warning, motivo))
				integracionFiscal["estado_documento_actualizado"] = docPersistido.EstadoDocumento
			}
		}
	}

	payloadCorreo := facturacionOperacionPayload{
		EmpresaID:       ventaDoc.EmpresaID,
		EntidadID:       ventaDoc.EntidadRelacionadaID,
		ClienteID:       ventaDoc.EntidadRelacionadaID,
		TipoDocumento:   "factura_electronica",
		PaisCodigo:      strings.TrimSpace(docPersistido.PaisCodigo),
		DocumentoCodigo: strings.TrimSpace(docPersistido.DocumentoCodigo),
		EstadoActual:    strings.TrimSpace(docPersistido.EstadoDocumento),
		MontoTotal:      ventaDoc.MontoTotal,
		Moneda:          strings.TrimSpace(docPersistido.Moneda),
		PeriodoContable: periodoContable,
		Observaciones:   observacionBase,
	}
	var envioCorreoFactura interface{} = map[string]interface{}{"intentado": false}
	if cfg.EnviarFacturaElectronicaVenta || facturacionAutoEmailClienteEnabled(dbEmp, ventaDoc.EmpresaID, strings.TrimSpace(docPersistido.PaisCodigo)) {
		envioCorreoFactura = enviarFacturaElectronicaAlCliente(dbEmp, dbSuper, payloadCorreo, *docPersistido)
	} else {
		envioCorreoFactura = facturaEmailAutoDisabledResultado(payloadCorreo)
	}

	return map[string]interface{}{
		"ok":                      true,
		"ya_existia":              false,
		"modo":                    "factura_electronica",
		"requiere_dian":           true,
		"documento_id":            docPersistido.ID,
		"tipo_documento":          docPersistido.TipoDocumento,
		"documento_codigo":        docPersistido.DocumentoCodigo,
		"estado_documento":        docPersistido.EstadoDocumento,
		"numero_legal":            docPersistido.NumeroLegal,
		"codigo_validacion":       docPersistido.CodigoValidacion,
		"pais_codigo":             docPersistido.PaisCodigo,
		"ambiente_fe":             docPersistido.AmbienteFE,
		"warning":                 warning,
		"integracion_fiscal":      integracionFiscal,
		"envio_correo_factura_fe": envioCorreoFactura,
		"documento_origen_codigo": ventaDoc.DocumentoCodigo,
		"documento_origen_tipo":   ventaDoc.TipoDocumento,
	}, nil
}

func registrarDocumentoVentaDesdeCarritoPagado(dbEmp, dbSuper *sql.DB, carrito *dbpkg.CarritoCompra, montoTotal float64, usuario string) (map[string]interface{}, error) {
	if dbEmp == nil || carrito == nil || carrito.EmpresaID <= 0 || carrito.ID <= 0 {
		return nil, nil
	}

	cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, carrito.EmpresaID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		empty := dbpkg.EmpresaConfiguracionAvanzada{EmpresaID: carrito.EmpresaID}
		cfg = &empty
	}

	emitirFacturaEnEstaVenta := carritoAutoFacturaElectronicaActiva(cfg)
	frecuenciaAplicada := false
	frecuenciaCadaNNo := int64(0)
	frecuenciaContadorAnterior := int64(0)
	frecuenciaContadorNuevo := int64(0)

	if emitirFacturaEnEstaVenta && cfg.FacturacionFrecuenciaAutomaticaActiva && cfg.FacturacionFrecuenciaCadaNNo > 0 {
		frecuenciaAplicada = true
		frecuenciaCadaNNo = cfg.FacturacionFrecuenciaCadaNNo
		ciclo := frecuenciaCadaNNo + 1
		frecuenciaContadorAnterior = cfg.FacturacionFrecuenciaContador % ciclo
		emitirFacturaEnEstaVenta = frecuenciaContadorAnterior == 0
		frecuenciaContadorNuevo = (frecuenciaContadorAnterior + 1) % ciclo
		cfg.FacturacionFrecuenciaContador = frecuenciaContadorNuevo
		cfg.UsuarioCreador = strings.TrimSpace(usuario)
		if _, upErr := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, *cfg); upErr != nil {
			log.Printf("[carritos] frecuencia_fe upsert empresa_id=%d carrito_id=%d error: %v", carrito.EmpresaID, carrito.ID, upErr)
		}
	}

	documentoCodigo := buildVentaDocumentoCodigo(carrito, "comprobante_pago")
	periodoContable := time.Now().Format("2006-01")
	fechaDocumento := strings.TrimSpace(carrito.PagadoEn)
	if fechaDocumento == "" {
		fechaDocumento = time.Now().Format("2006-01-02 15:04:05")
	}
	if montoTotal <= 0 {
		montoTotal = carrito.Total
	}

	docPayload := dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:            carrito.EmpresaID,
		TipoDocumento:        "comprobante_pago",
		DocumentoCodigo:      documentoCodigo,
		EstadoDocumento:      "emitida",
		EstadoAnterior:       "borrador",
		EventoUltimo:         "comprobante_pago_emitido",
		PeriodoContable:      periodoContable,
		MontoTotal:           montoTotal,
		Moneda:               strings.TrimSpace(carrito.Moneda),
		FechaDocumento:       fechaDocumento,
		EntidadRelacionadaID: carrito.ClienteID,
		UsuarioCreador:       strings.TrimSpace(usuario),
		Estado:               "activo",
		Observaciones:        "venta/comprobante generado automaticamente al cerrar la venta del carrito " + strings.TrimSpace(carrito.Codigo),
	}
	if docPayload.Moneda == "" {
		docPayload.Moneda = "COP"
	}
	if cfg != nil {
		docPayload.PaisCodigo = strings.TrimSpace(cfg.PaisCodigo)
	}
	docPayload.NumeroLegal = documentoCodigo
	docPayload.AmbienteFE = "no_aplica"

	docPersistido, err := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, docPayload)
	if err != nil {
		return nil, err
	}

	payloadCorreo := facturacionOperacionPayload{
		EmpresaID:       carrito.EmpresaID,
		EntidadID:       carrito.ClienteID,
		ClienteID:       carrito.ClienteID,
		TipoDocumento:   "comprobante_pago",
		PaisCodigo:      strings.TrimSpace(docPersistido.PaisCodigo),
		DocumentoCodigo: strings.TrimSpace(docPersistido.DocumentoCodigo),
		EstadoActual:    strings.TrimSpace(docPersistido.EstadoDocumento),
		MontoTotal:      montoTotal,
		Moneda:          strings.TrimSpace(docPersistido.Moneda),
		PeriodoContable: periodoContable,
		Observaciones:   "envio automatico desde pago de carrito",
	}
	var envioCorreoVenta interface{} = map[string]interface{}{"intentado": false}
	if cfg.EnviarEmailVenta {
		envioCorreoVenta = enviarFacturaElectronicaAlCliente(dbEmp, dbSuper, payloadCorreo, *docPersistido)
	}

	var facturaElectronica map[string]interface{}
	if emitirFacturaEnEstaVenta {
		facturaOut, facturaErr := registrarFacturaElectronicaDesdeDocumentoVenta(
			dbEmp,
			dbSuper,
			docPersistido,
			strings.TrimSpace(usuario),
			"factura electronica generada automaticamente desde la venta "+strings.TrimSpace(docPersistido.DocumentoCodigo),
		)
		if facturaErr != nil {
			log.Printf("[carritos] factura_electronica_automatica empresa_id=%d carrito_id=%d error: %v", carrito.EmpresaID, carrito.ID, facturaErr)
		} else if facturaOut != nil {
			facturaElectronica = facturaOut
		}
	}

	return map[string]interface{}{
		"ok":                true,
		"modo":              "comprobante_pago",
		"requiere_dian":     false,
		"documento_id":      docPersistido.ID,
		"tipo_documento":    docPersistido.TipoDocumento,
		"documento_codigo":  docPersistido.DocumentoCodigo,
		"estado_documento":  docPersistido.EstadoDocumento,
		"numero_legal":      docPersistido.NumeroLegal,
		"codigo_validacion": docPersistido.CodigoValidacion,
		"pais_codigo":       docPersistido.PaisCodigo,
		"ambiente_fe":       docPersistido.AmbienteFE,
		"warning":           "",
		"frecuencia": map[string]interface{}{
			"aplicada":             frecuenciaAplicada,
			"cada_n_no":            frecuenciaCadaNNo,
			"contador_anterior":    frecuenciaContadorAnterior,
			"contador_nuevo":       frecuenciaContadorNuevo,
			"emitio_en_esta_venta": emitirFacturaEnEstaVenta,
		},
		"facturacion_automatica_activa": emitirFacturaEnEstaVenta,
		"envio_correo_venta":            envioCorreoVenta,
		"factura_electronica":           facturaElectronica,
	}, nil
}

func anularDocumentosVentaDesdeCarrito(dbEmp *sql.DB, carrito *dbpkg.CarritoCompra, motivo, usuario string) []map[string]interface{} {
	out := []map[string]interface{}{}
	if dbEmp == nil || carrito == nil || carrito.EmpresaID <= 0 {
		return out
	}
	base := extractVentaDocumentoBase(strings.TrimSpace(carrito.Codigo))
	if base == "" && carrito.ID > 0 {
		base = fmt.Sprintf("CRT-%d", carrito.ID)
	}
	if base == "" {
		return out
	}
	for _, tipo := range []string{"comprobante_pago", "factura_electronica"} {
		codigo := buildVentaDocumentoCodigoFromBase(base, tipo)
		doc, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, carrito.EmpresaID, tipo, codigo)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("[carritos] anular documento venta empresa_id=%d carrito_id=%d tipo=%s codigo=%s error=%v", carrito.EmpresaID, carrito.ID, tipo, codigo, err)
			}
			continue
		}
		if doc == nil {
			continue
		}
		payload := *doc
		payload.EstadoAnterior = strings.TrimSpace(doc.EstadoDocumento)
		payload.EstadoDocumento = "anulada"
		payload.EventoUltimo = "venta_anulada"
		payload.UsuarioCreador = strings.TrimSpace(usuario)
		payload.Observaciones = strings.TrimSpace(doc.Observaciones)
		nota := fmt.Sprintf("Venta anulada desde carrito %s. Motivo: %s", strings.TrimSpace(carrito.Codigo), strings.TrimSpace(motivo))
		if payload.Observaciones != "" {
			payload.Observaciones += "\n"
		}
		payload.Observaciones += nota
		if _, upErr := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, payload); upErr != nil {
			log.Printf("[carritos] update documento anulada empresa_id=%d carrito_id=%d tipo=%s codigo=%s error=%v", carrito.EmpresaID, carrito.ID, tipo, codigo, upErr)
			continue
		}
		out = append(out, map[string]interface{}{
			"tipo_documento":   tipo,
			"documento_codigo": codigo,
			"estado_anterior":  doc.EstadoDocumento,
			"estado_nuevo":     "anulada",
		})
	}
	return out
}

func writeCarritoBusinessPrerequisite(w http.ResponseWriter, status int, prerequisite carritoBusinessPrerequisite) {
	prerequisite.OK = false
	if prerequisite.Scope == "" {
		prerequisite.Scope = "pagar_estacion"
	}
	writeJSON(w, status, prerequisite)
}

func validateCarritoPaymentPrerequisites(dbEmp *sql.DB, carrito *dbpkg.CarritoCompra) (*carritoBusinessPrerequisite, error) {
	if dbEmp == nil || carrito == nil || carrito.EmpresaID <= 0 || carrito.ID <= 0 {
		return nil, nil
	}
	if roundMoneyCarritoHandler(carrito.Total) <= 0 {
		return &carritoBusinessPrerequisite{
			Code:         "carrito_total_cero",
			Title:        "Carrito con cuenta en cero",
			Message:      "No se puede pagar ni cerrar este carrito porque la cuenta esta en cero. Agrega al menos un producto, servicio o tarifa antes de cerrar la venta.",
			RobotMessage: "No cierres todavia este carrito: la cuenta esta en cero. Primero agrega productos, servicios o configura una tarifa para esta estacion; revisa que el total sea mayor que cero y despues vuelve a presionar Pagar.",
			Scope:        "pagar_estacion",
			Steps: []string{
				"Agrega un producto, receta, servicio o tarifa al carrito.",
				"Verifica que el total del carrito sea mayor que cero.",
				"Vuelve a presionar Pagar cuando el detalle este completo.",
			},
			Actions: []map[string]interface{}{
				{"label": "Buscar productos", "url": fmt.Sprintf("/administrar_empresa/buscar_producto_botones.html?empresa_id=%d&carrito_id=%d", carrito.EmpresaID, carrito.ID)},
			},
		}, nil
	}

	cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, carrito.EmpresaID)
	if err != nil {
		return nil, err
	}
	if !carritoShouldEmitFacturaElectronica(cfg) {
		return nil, nil
	}

	paisCodigo := strings.TrimSpace(cfg.PaisCodigo)
	if paisCodigo == "" {
		paisDetectado, _, err := dbpkg.DetectFacturacionPais(dbEmp, carrito.EmpresaID, "", "")
		if err != nil {
			return nil, err
		}
		paisCodigo = paisDetectado.Codigo
	}
	feCfg, feErr := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, carrito.EmpresaID, paisCodigo)
	if feErr != nil && !errors.Is(feErr, sql.ErrNoRows) {
		return nil, feErr
	}

	missing := missingFacturacionElectronicaPaymentFields(cfg, feCfg)
	if feCfg != nil && strings.EqualFold(strings.TrimSpace(feCfg.Estado), "inactivo") {
		missing = append(missing, "perfil de facturacion por pais activo")
	}
	if reason := invalidFacturacionResolutionReason(cfg); reason != "" {
		missing = append(missing, reason)
	}
	if len(missing) == 0 {
		return nil, nil
	}

	configURL := fmt.Sprintf("/administrar_empresa/facturacion_electronica.html?empresa_id=%d", carrito.EmpresaID)
	message := "Antes de pagar debes completar el formato y la configuracion de facturacion electronica de la empresa. Faltan: " + strings.Join(missing, ", ") + "."
	return &carritoBusinessPrerequisite{
		Code:         "facturacion_configuracion_incompleta",
		Title:        "Configura la factura antes de pagar",
		Message:      message,
		RobotMessage: "Alto un momento: esta empresa tiene factura electronica automatica al cerrar la venta, pero la configuracion todavia no esta completa. Abre Facturacion electronica, completa datos fiscales, resolucion, numeracion y formato de impresion, guarda los cambios y vuelve al carrito para pagar.",
		Scope:        "pagar_estacion",
		Missing:      missing,
		Steps: []string{
			"Abre Administrar empresa > Configuracion > Facturacion electronica.",
			"Completa los datos fiscales del emisor: tipo de documento, NIT o identificador fiscal y razon social.",
			"Completa resolucion y numeracion: prefijo, numero de resolucion, fechas de vigencia y rango de consecutivos.",
			"Revisa Impresion y apariencia de factura: formato, logo, pie y notas legales si aplican.",
			"Guarda la configuracion y vuelve al carrito para pagar.",
		},
		Actions: []map[string]interface{}{
			{"label": "Abrir facturacion", "url": configURL},
		},
	}, nil
}

func carritoShouldEmitFacturaElectronica(cfg *dbpkg.EmpresaConfiguracionAvanzada) bool {
	if !carritoAutoFacturaElectronicaActiva(cfg) {
		return false
	}
	if cfg.FacturacionFrecuenciaAutomaticaActiva && cfg.FacturacionFrecuenciaCadaNNo > 0 {
		ciclo := cfg.FacturacionFrecuenciaCadaNNo + 1
		if ciclo <= 0 {
			return true
		}
		contador := cfg.FacturacionFrecuenciaContador % ciclo
		if contador < 0 {
			contador = 0
		}
		return contador == 0
	}
	return true
}

func missingFacturacionElectronicaPaymentFields(cfg *dbpkg.EmpresaConfiguracionAvanzada, feCfg *dbpkg.FacturacionElectronicaPaisConfig) []string {
	missing := make([]string, 0)
	value := func(advanced, country string) string {
		if strings.TrimSpace(country) != "" {
			return strings.TrimSpace(country)
		}
		return strings.TrimSpace(advanced)
	}
	if cfg == nil {
		return []string{"configuracion avanzada de facturacion"}
	}
	if value(cfg.TipoDocumentoEmisor, nonNilFacturacionPaisField(feCfg, "tipo_documento_emisor")) == "" {
		missing = append(missing, "tipo de documento del emisor")
	}
	if value(cfg.NIT, nonNilFacturacionPaisField(feCfg, "identificador_fiscal")) == "" {
		missing = append(missing, "NIT o identificador fiscal")
	}
	if value(cfg.RazonSocial, nonNilFacturacionPaisField(feCfg, "razon_social")) == "" {
		missing = append(missing, "razon social")
	}
	if value(cfg.PrefijoFactura, nonNilFacturacionPaisField(feCfg, "prefijo_factura")) == "" {
		missing = append(missing, "prefijo de factura")
	}
	if value(cfg.ResolucionNumero, nonNilFacturacionPaisField(feCfg, "resolucion_numero")) == "" {
		missing = append(missing, "numero de resolucion")
	}
	if strings.TrimSpace(cfg.FormatoImpresion) == "" {
		missing = append(missing, "formato de impresion de factura")
	}
	if cfg.ConsecutivoDesde <= 0 || cfg.ConsecutivoHasta < cfg.ConsecutivoDesde || cfg.ProximoConsecutivo < cfg.ConsecutivoDesde || cfg.ProximoConsecutivo > cfg.ConsecutivoHasta {
		missing = append(missing, "rango de consecutivos valido")
	}
	return missing
}

func nonNilFacturacionPaisField(cfg *dbpkg.FacturacionElectronicaPaisConfig, field string) string {
	if cfg == nil {
		return ""
	}
	switch field {
	case "tipo_documento_emisor":
		return cfg.TipoDocumentoEmisor
	case "identificador_fiscal":
		return cfg.IdentificadorFiscal
	case "razon_social":
		return cfg.RazonSocial
	case "prefijo_factura":
		return cfg.PrefijoFactura
	case "resolucion_numero":
		return cfg.ResolucionNumero
	default:
		return ""
	}
}

func validateStationCancelMargin(dbEmp *sql.DB, empresaID int64, carrito *dbpkg.CarritoCompra) error {
	if carrito == nil {
		return nil
	}
	if isCarritoVentaPagada(carrito) {
		return nil
	}
	if normalizeCarritoRegistroEstado(carrito.Estado) != "activo" || normalizeCarritoOperativoEstado(carrito.EstadoCarrito) != "abierto" {
		return nil
	}
	cfg, err := dbpkg.GetEmpresaTarifaPorMinutosConfiguracion(dbEmp, empresaID)
	if err != nil {
		log.Printf("[carritos] cancel margin config empresa_id=%d error: %v", empresaID, err)
		return nil
	}
	if cfg == nil || !cfg.MargenDesactivacionHabilitado || cfg.MargenDesactivacionMinutos <= 0 {
		return nil
	}
	activadoAt, ok := parseCarritoActivationTime(carrito.ActivadoEn)
	if !ok {
		return nil
	}
	elapsed := time.Since(activadoAt)
	if elapsed < 0 {
		return nil
	}
	if elapsed > time.Duration(cfg.MargenDesactivacionMinutos)*time.Minute {
		return fmt.Errorf("el margen para cancelar o desactivar la estacion vencio. Debes usar Pagar y cerrar carrito para finalizar esta sesion")
	}
	return nil
}

func parseCarritoActivationTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func invalidFacturacionResolutionReason(cfg *dbpkg.EmpresaConfiguracionAvanzada) string {
	if cfg == nil {
		return ""
	}
	nowDate := time.Now().Format("2006-01-02")
	if raw := strings.TrimSpace(cfg.ResolucionFechaDesde); raw != "" {
		desde, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return "fecha inicial de resolucion valida"
		}
		if nowDate < desde.Format("2006-01-02") {
			return "resolucion vigente: la fecha inicial aun no aplica"
		}
	}
	if raw := strings.TrimSpace(cfg.ResolucionFechaHasta); raw != "" {
		hasta, err := time.Parse("2006-01-02", raw)
		if err != nil {
			return "fecha final de resolucion valida"
		}
		if nowDate > hasta.Format("2006-01-02") {
			return "resolucion vigente: la resolucion esta vencida"
		}
	}
	return ""
}

func parseOptionalInt64CarritoQuery(r *http.Request, key string) (int64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s invalido", key)
	}
	return v, nil
}

func parseOptionalIntCarritoQuery(r *http.Request, key string) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s invalido", key)
	}
	return v, nil
}
