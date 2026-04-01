package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

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
				if err := dbpkg.ActivateCarritoStationSession(dbEmp, empresaID, id, resetItems); err != nil {
					log.Printf("[carritos] activar_estacion empresa_id=%d id=%d reset_items=%v error: %v", empresaID, id, resetItems, err)
					http.Error(w, "No se pudo activar el carrito de estación", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": "activo", "estado_carrito": "abierto"})
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

				var payload struct {
					DescuentoTipo   string  `json:"descuento_tipo"`
					DescuentoCodigo string  `json:"descuento_codigo"`
					DescuentoValor  float64 `json:"descuento_valor"`
					DevolucionTotal float64 `json:"devolucion_total"`
					TotalPagado     float64 `json:"total_pagado"`
				}
				if r.Body != nil {
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
						http.Error(w, "JSON invalido", http.StatusBadRequest)
						return
					}
				}

				if err := dbpkg.PayCarritoStationSession(
					dbEmp,
					empresaID,
					id,
					payload.DescuentoTipo,
					payload.DescuentoCodigo,
					payload.DescuentoValor,
					payload.DevolucionTotal,
					payload.TotalPagado,
				); err != nil {
					log.Printf("[carritos] pagar_estacion empresa_id=%d id=%d error: %v", empresaID, id, err)
					http.Error(w, "No se pudo cerrar el carrito por pago", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": "inactivo", "estado_carrito": "cerrado"})
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
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetCarritoCompraEstado(dbEmp, empresaID, id, estado); err != nil {
					log.Printf("[carritos] set estado empresa_id=%d id=%d estado=%s error: %v", empresaID, id, estado, err)
					http.Error(w, "No se pudo actualizar estado del carrito", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
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
				estadoCarrito := "abierto"
				if action == "cerrar" {
					estadoCarrito = "cerrado"
				}
				if err := dbpkg.SetCarritoOperacionEstado(dbEmp, empresaID, id, estadoCarrito); err != nil {
					log.Printf("[carritos] set estado_operacion empresa_id=%d id=%d estado=%s error: %v", empresaID, id, estadoCarrito, err)
					http.Error(w, "No se pudo actualizar estado operativo del carrito", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado_carrito": estadoCarrito})
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
			if err := dbpkg.UpdateCarritoCompraItem(dbEmp, payload); err != nil {
				log.Printf("[carritos_items] update empresa_id=%d carrito_id=%d id=%d error: %v", payload.EmpresaID, payload.CarritoID, payload.ID, err)
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
	return nil
}
