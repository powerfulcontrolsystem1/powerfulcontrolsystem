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
	"strings"

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
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
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
				if err := validateCarritoTransitionForAction(carrito, action); err != nil {
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
					http.Error(w, "metodo_pago invalido. Use: efectivo, tarjeta_credito, tarjeta_debito, codigo_descuento o mixto", http.StatusBadRequest)
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
					referenciaPago = buildReferenciaPagoMixto(pagosMixtos)
				} else if (metodoPago == "tarjeta_credito" || metodoPago == "tarjeta_debito") && len(referenciaPago) < 4 {
					http.Error(w, "referencia_pago es obligatoria para pagos con tarjeta (minimo 4 caracteres)", http.StatusBadRequest)
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
					aplicado, err := dbpkg.ResolveCodigoDescuentoParaMonto(dbEmp, empresaID, descuentoCodigo, carrito.Total)
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

				totalPagado := payload.TotalPagado
				if metodoPago == "mixto" {
					totalPagado = totalPagadoMixto
				} else {
					if totalPagado < 0 {
						totalPagado = 0
					}
					if totalPagado == 0 && metodoPago != "codigo_descuento" {
						totalPagado = totalEsperado
					}
				}
				totalPagado = roundMoneyCarritoHandler(totalPagado)

				if metodoPago == "codigo_descuento" {
					if totalEsperado > 0 {
						http.Error(w, "el codigo de descuento no cubre el total del carrito; use efectivo o tarjeta para cubrir el saldo restante", http.StatusBadRequest)
						return
					}
					totalPagado = 0
				} else if metodoPago == "mixto" {
					if len(pagosMixtos) < 2 {
						http.Error(w, "pago mixto requiere al menos 2 metodos con monto mayor a cero", http.StatusBadRequest)
						return
					}
					if math.Abs(totalPagado-totalEsperado) > 0.01 {
						http.Error(w, "la suma de pagos mixtos debe coincidir con el total esperado", http.StatusBadRequest)
						return
					}
				} else {
					if totalPagado+0.009 < totalEsperado {
						http.Error(w, "total_pagado insuficiente para completar el pago", http.StatusBadRequest)
						return
					}
				}

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
					montoEvento = totalEsperado
				}
				registrarEventoContableVentaCarrito(dbEmp, r, carrito, "venta_pagada", montoEvento, map[string]interface{}{
					"action":                "pagar_estacion",
					"metodo_pago":           metodoPago,
					"referencia_pago":       referenciaPago,
					"pagos_mixtos":          pagosMixtosToEventPayload(pagosMixtos),
					"descuento_tipo":        descuentoTipo,
					"descuento_codigo":      descuentoCodigo,
					"descuento_valor":       descuentoValor,
					"codigo_descuento_id":   codigoDescuentoID,
					"devolucion_total":      devolucionTotal,
					"total_pagado":          totalPagado,
					"total_esperado":        totalEsperado,
					"estado_venta_anterior": carrito.EstadoVenta,
					"estado_venta_nuevo":    "venta_pagada",
				}, "pago de venta en estacion")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": "inactivo", "estado_carrito": "cerrado", "estado_venta": "venta_pagada"})
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
				if errors.Is(err, dbpkg.ErrStockInsuficiente) {
					http.Error(w, "Stock insuficiente para agregar el item al carrito", http.StatusBadRequest)
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
		if estadoRegistro != "activo" {
			return fmt.Errorf("solo se puede cerrar una venta activa")
		}
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
			return nil, 0, fmt.Errorf("pago mixto solo permite efectivo, tarjeta_credito y tarjeta_debito")
		}
		monto := roundMoneyCarritoHandler(item.Monto)
		if monto <= 0 {
			continue
		}
		referencia := strings.TrimSpace(item.Referencia)
		if (metodo == "tarjeta_credito" || metodo == "tarjeta_debito") && len(referencia) < 4 {
			return nil, 0, fmt.Errorf("cada pago con tarjeta en pago mixto requiere referencia minima de 4 caracteres")
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
