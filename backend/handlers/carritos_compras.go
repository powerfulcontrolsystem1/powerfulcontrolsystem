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

// EmpresaCarritosCompraHandler gestiona CRUD de carritos por empresa.
func EmpresaCarritosCompraHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "totales_pago" {
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

				query := `SELECT COALESCE(LOWER(TRIM(metodo_pago)),'efectivo') AS metodo_pago, ROUND(SUM(COALESCE(monto_pagado,0)),2) AS total_pagado
					FROM empresa_ventas_estacion_metricas
					WHERE empresa_id = ?
						AND COALESCE(estado,'activo') = 'activo'
						AND evento_operacion = 'venta_pagada'
						AND datetime(COALESCE(fecha_evento, fecha_creacion, datetime('now','localtime'))) >= datetime('now','localtime', ?)`
				args := []interface{}{empresaID, fmt.Sprintf("-%d day", days)}
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
					"tarjeta_debito":        0.0,
					"tarjeta_credito":       0.0,
					"transferencia_bancaria": 0.0,
				}
				for rows.Next() {
					var metodo string
					var total float64
					if err := rows.Scan(&metodo, &total); err != nil {
						continue
					}
					metodo = strings.TrimSpace(strings.ToLower(metodo))
					totals[metodo] = total
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

			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := dbpkg.RefreshCarritosActivosConTarifaPorDia(dbEmp, empresaID, time.Now()); err != nil {
				log.Printf("[carritos] refresh tarifas_por_dia empresa_id=%d error: %v", empresaID, err)
			}
			includeInactive := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "1") ||
				strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "true")
			q := strings.TrimSpace(r.URL.Query().Get("q"))

			rows, err := dbpkg.GetCarritosCompraByEmpresa(dbEmp, empresaID, includeInactive, q)
			if err != nil {
				log.Printf("[carritos] list empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudieron listar los carritos", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.CarritoCompra
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if err := validateCarritoPayload(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
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
				if isCarritoVentaPagada(carrito) && !resetItems {
					http.Error(w, "venta pagada: use reset_items=1 para iniciar una nueva sesion", http.StatusConflict)
					return
				}
				if !resetItems && normalizeCarritoRegistroEstado(carrito.Estado) == "activo" && normalizeCarritoOperativoEstado(carrito.EstadoCarrito) == "abierto" {
					http.Error(w, "la venta ya se encuentra activa y abierta", http.StatusConflict)
					return
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
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": "activo", "estado_carrito": "abierto", "estado_venta": "venta_abierta"})
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
				tarifaPorDiaCalculo, errTarifaDia := dbpkg.RefreshCarritoTotalConTarifaPorDia(dbEmp, empresaID, id, time.Now())
				if errTarifaDia != nil {
					log.Printf("[carritos] refresh carrito tarifa_por_dia empresa_id=%d id=%d error: %v", empresaID, id, errTarifaDia)
					http.Error(w, "No se pudo recalcular tarifa diaria del carrito", http.StatusInternalServerError)
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
				log.Printf("[carritos] debug intentar %s empresa_id=%d id=%d estado=%q estado_carrito=%q pagado_en=%q", action, empresaID, id, carrito.Estado, carrito.EstadoCarrito, carrito.PagadoEn)
				if err := validateCarritoTransitionForAction(carrito, action); err != nil {
					log.Printf("[carritos] validate failed action=%s empresa_id=%d id=%d estado=%q estado_carrito=%q pagado_en=%q err=%v", action, empresaID, id, carrito.Estado, carrito.EstadoCarrito, carrito.PagadoEn, err)
					http.Error(w, err.Error(), http.StatusConflict)
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
					TotalPagado     float64                   `json:"total_pagado"`
					AplicarPropina  *bool                     `json:"aplicar_propina"`
					UsuarioLavador  string                    `json:"usuario_lavador"`
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
					http.Error(w, "metodo_pago invalido. Use: efectivo, tarjeta_credito, tarjeta_debito, transferencia_bancaria, codigo_descuento o mixto", http.StatusBadRequest)
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

				referenciaPago := strings.TrimSpace(payload.ReferenciaPago)
				pagosMixtos := make([]carritoPagoMixtoNormalizado, 0)
				totalPagadoMixto := 0.0
				if metodoPago == "mixto" {
					var err error
					pagosMixtos, totalPagadoMixto, err = normalizePagosMixtosCarrito(payload.PagosMixtos)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					for _, tramo := range pagosMixtos {
						if !permisosOperativos.IsMetodoPagoHabilitado(tramo.Metodo) {
							http.Error(w, "uno o mas metodos del pago mixto no estan habilitados para la empresa/rol actual", http.StatusForbidden)
							return
						}
					}
					referenciaPago = buildReferenciaPagoMixto(pagosMixtos)
				} else if (metodoPago == "tarjeta_credito" || metodoPago == "tarjeta_debito" || metodoPago == "transferencia_bancaria") && len(referenciaPago) < 4 {
					http.Error(w, "referencia_pago es obligatoria para pagos con tarjeta o transferencia bancaria (minimo 4 caracteres)", http.StatusBadRequest)
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

				totalEsperado := roundMoneyCarritoHandler(carrito.Total - descuentoValor - devolucionTotal)
				if totalEsperado < 0 {
					totalEsperado = 0
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
					montoPropina = roundMoneyCarritoHandler(totalEsperado * (propinaPorcentaje / 100))
					if montoPropina < 0 {
						montoPropina = 0
					}
				}
				totalEsperadoConPropina := roundMoneyCarritoHandler(totalEsperado + montoPropina)

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
				totalPagado = roundMoneyCarritoHandler(totalPagado)

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
					if math.Abs(totalPagado-totalEsperadoConPropina) > 0.01 {
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
					totalPagado,
					codigoDescuentoID,
					usuarioOperacion,
				); err != nil {
					log.Printf("[carritos] pagar_estacion empresa_id=%d id=%d error: %v", empresaID, id, err)
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
				montoEvento := totalPagado
				if montoEvento <= 0 {
					montoEvento = totalEsperadoConPropina
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
						UsuarioAsignado:   usuarioOperacion,
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
					"metodo_pago":           metodoPago,
					"referencia_pago":       referenciaPago,
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
					"total_pagado":          totalPagado,
					"total_esperado":        totalEsperado,
					"total_esperado_final":  totalEsperadoConPropina,
					"propina_aplicada":      propinaAplicada,
					"propina_porcentaje":    propinaPorcentaje,
					"propina_monto":         montoPropina,
					"propina_modo":          propinaModo,
					"propina_registro_id":   propinaRegistroID,
					"propina_registrada":    propinaRegistrada,
					"comision_aplicada":     comisionResultado.Aplicada,
					"comision_porcentaje":   comisionResultado.PorcentajeComision,
					"comision_filtro":       comisionResultado.FiltroServicio,
					"comision_lavador":      comisionResultado.UsuarioLavador,
					"comision_base":         comisionResultado.BaseServicios,
					"comision_monto":        comisionResultado.MontoComision,
					"comision_movimientos":  comisionResultado.MovimientosRegistrados,
					"comision_warning":      comisionResultado.Warning,
					"estado_venta_anterior": carrito.EstadoVenta,
					"estado_venta_nuevo":    "venta_pagada",
				}, "pago de venta en estacion")

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
					UsuarioCreador:      usuarioOperacion,
					Observaciones:       "cierre de venta simple por estacion",
				}); errMetric != nil {
					log.Printf("[carritos] metrica venta_pagada empresa_id=%d carrito_id=%d error: %v", empresaID, id, errMetric)
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                         true,
					"estado":                     "inactivo",
					"estado_carrito":             "cerrado",
					"estado_venta":               "venta_pagada",
					"tarifa_por_dia":             tarifaPorDiaCalculo,
					"total_esperado":             totalEsperado,
					"total_esperado_con_propina": totalEsperadoConPropina,
					"propina": map[string]interface{}{
						"aplicada":          propinaAplicada,
						"habilitada":        propinaHabilitada,
						"porcentaje":        propinaPorcentaje,
						"monto":             montoPropina,
						"modo_distribucion": propinaModo,
						"registrada":        propinaRegistrada,
						"registro_id":       propinaRegistroID,
						"warning":           propinaWarning,
					},
					"comision": map[string]interface{}{
						"aplicada":                comisionResultado.Aplicada,
						"habilitada":              comisionResultado.Habilitada,
						"aplicacion_automatica":   comisionResultado.AplicacionAutomatica,
						"porcentaje_comision":     comisionResultado.PorcentajeComision,
						"filtro_servicio":         comisionResultado.FiltroServicio,
						"usuario_lavador":         comisionResultado.UsuarioLavador,
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
						"metodo_pago_mixto":                  permisosOperativos.MetodoPagoMixto,
						"metodo_pago_codigo_descuento":       permisosOperativos.MetodoPagoCodigoDescuento,
						"habilitar_propinas":                 permisosOperativos.HabilitarPropinas,
						"habilitar_comisiones":               permisosOperativos.HabilitarComisiones,
					},
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

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":             true,
					"estado":         normalizeCarritoRegistroEstado(carritoActualizado.Estado),
					"estado_carrito": normalizeCarritoOperativoEstado(carritoActualizado.EstadoCarrito),
					"estado_venta":   carritoActualizado.EstadoVenta,
					"activado_en":    strings.TrimSpace(carritoActualizado.ActivadoEn),
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
   INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado, fecha_creacion, fecha_actualizacion, usuario_creador) VALUES (6, 2, <BODEGA_ID>, 10, 'activo', datetime('now','localtime'), datetime('now','localtime'), 'admin@example.com');
3) Alternativamente, establezca 'bodega_principal_id' en la tabla 'productos':
   UPDATE productos SET bodega_principal_id = <BODEGA_ID> WHERE empresa_id = 6 AND id = 2;
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
	return strings.TrimSpace(carrito.PagadoEn) != ""
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
	if payload.Cantidad <= 0 {
		return fmt.Errorf("cantidad debe ser mayor a cero")
	}
	if payload.PrecioUnitario < 0 {
		return fmt.Errorf("precio_unitario invalido")
	}
	tipoItem := strings.TrimSpace(strings.ToLower(payload.TipoItem))
	if tipoItem == "combo" && payload.ReferenciaID <= 0 {
		return fmt.Errorf("referencia_id es obligatoria para tipo_item combo")
	}
	return nil
}

func roundMoneyCarritoHandler(v float64) float64 {
	return math.Round(v*100) / 100
}

func normalizePagosMixtosCarrito(entries []carritoPagoMixtoEntrada) ([]carritoPagoMixtoNormalizado, float64, error) {
	if len(entries) == 0 {
		return nil, 0, fmt.Errorf("pago mixto requiere detalle de pagos_mixtos")
	}

	normalized := make([]carritoPagoMixtoNormalizado, 0, len(entries))
	total := 0.0
	for _, item := range entries {
		metodo := dbpkg.NormalizeMetodoPagoCarrito(item.Metodo)
		if metodo == "" || metodo == "mixto" || metodo == "codigo_descuento" {
			return nil, 0, fmt.Errorf("pago mixto solo permite efectivo, tarjeta_credito, tarjeta_debito y transferencia_bancaria")
		}
		monto := roundMoneyCarritoHandler(item.Monto)
		if monto <= 0 {
			continue
		}
		referencia := strings.TrimSpace(item.Referencia)
		if (metodo == "tarjeta_credito" || metodo == "tarjeta_debito" || metodo == "transferencia_bancaria") && len(referencia) < 4 {
			return nil, 0, fmt.Errorf("cada pago con tarjeta o transferencia bancaria en pago mixto requiere referencia minima de 4 caracteres")
		}

		normalized = append(normalized, carritoPagoMixtoNormalizado{
			Metodo:     metodo,
			Monto:      monto,
			Referencia: referencia,
		})
		total = roundMoneyCarritoHandler(total + monto)
	}

	if len(normalized) < 2 {
		return nil, 0, fmt.Errorf("pago mixto requiere al menos 2 metodos con monto mayor a cero")
	}

	return normalized, total, nil
}

func buildReferenciaPagoMixto(pagos []carritoPagoMixtoNormalizado) string {
	if len(pagos) == 0 {
		return ""
	}
	parts := make([]string, 0, len(pagos))
	for _, item := range pagos {
		chunk := item.Metodo + ":" + fmt.Sprintf("%.2f", item.Monto)
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
