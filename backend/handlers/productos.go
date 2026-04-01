package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaBodegasHandler maneja CRUD de bodegas por empresa.
func EmpresaBodegasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			incluirInactivas := r.URL.Query().Get("include_inactive") == "1"
			rows, err := dbpkg.GetBodegasByEmpresa(dbEmp, empresaID, incluirInactivas)
			if err != nil {
				http.Error(w, "failed to list bodegas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return
		case http.MethodPost:
			var payload dbpkg.Bodega
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id required", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = adminEmailFromRequest(r)
			id, err := dbpkg.CreateBodega(dbEmp, payload)
			if err != nil {
				http.Error(w, "failed to create bodega: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			if q.Get("action") == "activar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id required", http.StatusBadRequest)
					return
				}
				estado := "inactivo"
				if q.Get("activo") == "1" || strings.EqualFold(q.Get("estado"), "activo") {
					estado = "activo"
				}
				if err := dbpkg.SetBodegaEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			var payload dbpkg.Bodega
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "empresa_id and id required", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateBodega(dbEmp, payload); err != nil {
				http.Error(w, "failed to update bodega: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteBodega(dbEmp, empresaID, id); err != nil {
				http.Error(w, "failed to delete bodega: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaProductosHandler maneja CRUD de productos por empresa.
func EmpresaProductosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if id, _ := parseInt64QueryOptional(r, "id"); id > 0 {
				p, err := dbpkg.GetProductoByID(dbEmp, empresaID, id)
				if err != nil {
					if err == sql.ErrNoRows {
						http.Error(w, "producto not found", http.StatusNotFound)
						return
					}
					http.Error(w, "failed to get producto: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(p)
				return
			}

			q := r.URL.Query().Get("q")
			estado := r.URL.Query().Get("estado")
			bodegaID, _ := parseInt64QueryOptional(r, "bodega_id")
			limit, _ := parseIntQueryOptional(r, "limit")
			offset, _ := parseIntQueryOptional(r, "offset")
			rows, err := dbpkg.GetProductosByEmpresa(dbEmp, empresaID, q, estado, bodegaID, limit, offset)
			if err != nil {
				http.Error(w, "failed to list productos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return
		case http.MethodPost:
			var payload struct {
				dbpkg.Producto
				StockInicial      float64 `json:"stock_inicial"`
				ReferenciaInicial string  `json:"referencia_inicial"`
				MotivoPrecio      string  `json:"motivo_precio"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id required", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = adminEmailFromRequest(r)
			id, err := dbpkg.CreateProducto(dbEmp, payload.Producto, payload.StockInicial, payload.ReferenciaInicial)
			if err != nil {
				http.Error(w, "failed to create producto: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			if q.Get("action") == "activar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id required", http.StatusBadRequest)
					return
				}
				estado := "inactivo"
				if q.Get("activo") == "1" || strings.EqualFold(q.Get("estado"), "activo") {
					estado = "activo"
				}
				if err := dbpkg.SetProductoEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			var payload dbpkg.Producto
			var motivoPrecio string
			var referenciaPrecio string
			if err := json.NewDecoder(r.Body).Decode(&struct {
				*dbpkg.Producto
				MotivoPrecio     *string `json:"motivo_precio"`
				ReferenciaPrecio *string `json:"referencia_precio"`
			}{Producto: &payload, MotivoPrecio: &motivoPrecio, ReferenciaPrecio: &referenciaPrecio}); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "empresa_id and id required", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "nombre required", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = adminEmailFromRequest(r)
			if err := dbpkg.UpdateProducto(dbEmp, payload, motivoPrecio, referenciaPrecio); err != nil {
				http.Error(w, "failed to update producto: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteProducto(dbEmp, empresaID, id); err != nil {
				http.Error(w, "failed to delete producto: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaInventarioExistenciasHandler lista existencias por empresa.
func EmpresaInventarioExistenciasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		productoID, _ := parseInt64QueryOptional(r, "producto_id")
		bodegaID, _ := parseInt64QueryOptional(r, "bodega_id")
		limit, _ := parseIntQueryOptional(r, "limit")
		offset, _ := parseIntQueryOptional(r, "offset")
		rows, err := dbpkg.GetExistenciasByEmpresa(dbEmp, empresaID, productoID, bodegaID, limit, offset)
		if err != nil {
			http.Error(w, "failed to list existencias: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioMovimientosHandler lista movimientos de inventario por empresa.
func EmpresaInventarioMovimientosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		productoID, _ := parseInt64QueryOptional(r, "producto_id")
		limit, _ := parseIntQueryOptional(r, "limit")
		offset, _ := parseIntQueryOptional(r, "offset")
		rows, err := dbpkg.GetMovimientosByEmpresa(dbEmp, empresaID, productoID, limit, offset)
		if err != nil {
			http.Error(w, "failed to list movimientos: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioTransferHandler traslada inventario entre bodegas.
func EmpresaInventarioTransferHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			EmpresaID       int64   `json:"empresa_id"`
			ProductoID      int64   `json:"producto_id"`
			BodegaOrigenID  int64   `json:"bodega_origen_id"`
			BodegaDestinoID int64   `json:"bodega_destino_id"`
			Cantidad        float64 `json:"cantidad"`
			Referencia      string  `json:"referencia"`
			Observaciones   string  `json:"observaciones"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 || payload.ProductoID <= 0 || payload.BodegaOrigenID <= 0 || payload.BodegaDestinoID <= 0 {
			http.Error(w, "empresa_id, producto_id, bodega_origen_id y bodega_destino_id son obligatorios", http.StatusBadRequest)
			return
		}
		if payload.Cantidad <= 0 {
			http.Error(w, "cantidad debe ser mayor a 0", http.StatusBadRequest)
			return
		}

		err := dbpkg.TransferirProductoEntreBodegas(
			dbEmp,
			payload.EmpresaID,
			payload.ProductoID,
			payload.BodegaOrigenID,
			payload.BodegaDestinoID,
			payload.Cantidad,
			payload.Referencia,
			adminEmailFromRequest(r),
			payload.Observaciones,
		)
		if err != nil {
			if errors.Is(err, dbpkg.ErrStockInsuficiente) {
				http.Error(w, "stock insuficiente en la bodega origen", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to transfer stock: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"moved": true})
	}
}

// EmpresaInventarioAjusteHandler registra entradas/salidas/devoluciones/pérdidas.
func EmpresaInventarioAjusteHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			EmpresaID     int64   `json:"empresa_id"`
			ProductoID    int64   `json:"producto_id"`
			BodegaID      int64   `json:"bodega_id"`
			Tipo          string  `json:"tipo"`
			Cantidad      float64 `json:"cantidad"`
			Referencia    string  `json:"referencia"`
			Observaciones string  `json:"observaciones"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 || payload.ProductoID <= 0 || payload.BodegaID <= 0 {
			http.Error(w, "empresa_id, producto_id y bodega_id son obligatorios", http.StatusBadRequest)
			return
		}
		if payload.Cantidad <= 0 {
			http.Error(w, "cantidad debe ser mayor a 0", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.Tipo) == "" {
			http.Error(w, "tipo requerido", http.StatusBadRequest)
			return
		}

		err := dbpkg.RegistrarMovimientoInventario(
			dbEmp,
			payload.EmpresaID,
			payload.ProductoID,
			payload.BodegaID,
			payload.Tipo,
			payload.Cantidad,
			payload.Referencia,
			adminEmailFromRequest(r),
			payload.Observaciones,
		)
		if err != nil {
			if errors.Is(err, dbpkg.ErrStockInsuficiente) {
				http.Error(w, "stock insuficiente para la operación", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to adjust inventario: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"adjusted": true})
	}
}

// EmpresaInventarioCambioProductoHandler registra cambio de un producto por otro.
func EmpresaInventarioCambioProductoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var payload struct {
			EmpresaID         int64   `json:"empresa_id"`
			ProductoOrigenID  int64   `json:"producto_origen_id"`
			ProductoDestinoID int64   `json:"producto_destino_id"`
			BodegaID          int64   `json:"bodega_id"`
			Cantidad          float64 `json:"cantidad"`
			Referencia        string  `json:"referencia"`
			Observaciones     string  `json:"observaciones"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 || payload.ProductoOrigenID <= 0 || payload.ProductoDestinoID <= 0 || payload.BodegaID <= 0 {
			http.Error(w, "empresa_id, producto_origen_id, producto_destino_id y bodega_id son obligatorios", http.StatusBadRequest)
			return
		}
		if payload.Cantidad <= 0 {
			http.Error(w, "cantidad debe ser mayor a 0", http.StatusBadRequest)
			return
		}

		err := dbpkg.RegistrarCambioProducto(
			dbEmp,
			payload.EmpresaID,
			payload.ProductoOrigenID,
			payload.ProductoDestinoID,
			payload.BodegaID,
			payload.Cantidad,
			payload.Referencia,
			adminEmailFromRequest(r),
			payload.Observaciones,
		)
		if err != nil {
			if errors.Is(err, dbpkg.ErrStockInsuficiente) {
				http.Error(w, "stock insuficiente para cambio de producto", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to register cambio de producto: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"changed": true})
	}
}

// EmpresaProductoPrecioHistorialHandler lista historial de cambios de precio.
func EmpresaProductoPrecioHistorialHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		productoID, _ := parseInt64QueryOptional(r, "producto_id")
		limit, _ := parseIntQueryOptional(r, "limit")
		offset, _ := parseIntQueryOptional(r, "offset")
		rows, err := dbpkg.GetProductoPrecioHistorialByEmpresa(dbEmp, empresaID, productoID, limit, offset)
		if err != nil {
			http.Error(w, "failed to list historial de precios: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaProveedoresHandler maneja CRUD de proveedores por empresa.
func EmpresaProveedoresHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := r.URL.Query().Get("include_inactive") == "1"
			rows, err := dbpkg.GetProveedoresByEmpresa(dbEmp, empresaID, includeInactive)
			if err != nil {
				http.Error(w, "failed to list proveedores: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return
		case http.MethodPost:
			var payload dbpkg.Proveedor
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "empresa_id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = adminEmailFromRequest(r)
			id, err := dbpkg.CreateProveedor(dbEmp, payload)
			if err != nil {
				http.Error(w, "failed to create proveedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			if q.Get("action") == "activar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id required", http.StatusBadRequest)
					return
				}
				estado := "inactivo"
				if q.Get("activo") == "1" || strings.EqualFold(q.Get("estado"), "activo") {
					estado = "activo"
				}
				if err := dbpkg.SetProveedorEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "failed to set proveedor estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			var payload dbpkg.Proveedor
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "empresa_id, id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateProveedor(dbEmp, payload); err != nil {
				http.Error(w, "failed to update proveedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteProveedor(dbEmp, empresaID, id); err != nil {
				http.Error(w, "failed to delete proveedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaServiciosHandler maneja CRUD de servicios por empresa.
func EmpresaServiciosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			q := r.URL.Query().Get("q")
			estado := r.URL.Query().Get("estado")
			limit, _ := parseIntQueryOptional(r, "limit")
			offset, _ := parseIntQueryOptional(r, "offset")
			rows, err := dbpkg.GetServiciosByEmpresa(dbEmp, empresaID, q, estado, limit, offset)
			if err != nil {
				http.Error(w, "failed to list servicios: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return
		case http.MethodPost:
			var payload dbpkg.Servicio
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "empresa_id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = adminEmailFromRequest(r)
			id, err := dbpkg.CreateServicio(dbEmp, payload)
			if err != nil {
				http.Error(w, "failed to create servicio: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			if q.Get("action") == "activar" {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id required", http.StatusBadRequest)
					return
				}
				estado := "inactivo"
				if q.Get("activo") == "1" || strings.EqualFold(q.Get("estado"), "activo") {
					estado = "activo"
				}
				if err := dbpkg.SetServicioEstado(dbEmp, empresaID, id, estado); err != nil {
					http.Error(w, "failed to set servicio estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			var payload dbpkg.Servicio
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "empresa_id, id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateServicio(dbEmp, payload); err != nil {
				http.Error(w, "failed to update servicio: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			id, err := parseInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteServicio(dbEmp, empresaID, id); err != nil {
				http.Error(w, "failed to delete servicio: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaProductoImagenUploadHandler permite subir imagen/logo para un producto.
func EmpresaProductoImagenUploadHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseMultipartForm(12 << 20); err != nil { // 12MB
			http.Error(w, "invalid multipart payload", http.StatusBadRequest)
			return
		}

		empresaID, err := parseInt64Form(r, "empresa_id")
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id required", http.StatusBadRequest)
			return
		}
		productoID, err := parseInt64Form(r, "producto_id")
		if err != nil || productoID <= 0 {
			http.Error(w, "producto_id required", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("imagen")
		if err != nil {
			http.Error(w, "imagen required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".svg": true}
		if !allowed[ext] {
			http.Error(w, "image extension not allowed", http.StatusBadRequest)
			return
		}

		webRoot := resolveWebRootDir()
		dir := filepath.Join(webRoot, "uploads", "productos", fmt.Sprintf("empresa_%d", empresaID))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			http.Error(w, "failed to prepare upload directory", http.StatusInternalServerError)
			return
		}

		fileName := fmt.Sprintf("producto_%d_%d%s", productoID, time.Now().UnixNano(), ext)
		absPath := filepath.Join(dir, fileName)
		out, err := os.Create(absPath)
		if err != nil {
			http.Error(w, "failed to create image file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, "failed to save image file", http.StatusInternalServerError)
			return
		}

		imageURL := "/uploads/productos/empresa_" + strconv.FormatInt(empresaID, 10) + "/" + fileName
		if err := dbpkg.UpdateProductoImagen(dbEmp, empresaID, productoID, imageURL); err != nil {
			http.Error(w, "failed to update image url in producto: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"saved":       true,
			"empresa_id":  empresaID,
			"producto_id": productoID,
			"image_url":   imageURL,
		})
	}
}

func resolveWebRootDir() string {
	candidates := []string{"../web", "web"}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "..", "web"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return "../web"
}

func parseEmpresaIDQuery(r *http.Request) (int64, error) {
	empresaID, err := parseInt64Query(r, "empresa_id")
	if err != nil || empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id required")
	}
	return empresaID, nil
}

func parseInt64Query(r *http.Request, key string) (int64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, fmt.Errorf("%s required", key)
	}
	return strconv.ParseInt(raw, 10, 64)
}

func parseInt64QueryOptional(r *http.Request, key string) (int64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func parseIntQueryOptional(r *http.Request, key string) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}
	return strconv.Atoi(raw)
}

func parseInt64Form(r *http.Request, key string) (int64, error) {
	raw := strings.TrimSpace(r.FormValue(key))
	if raw == "" {
		return 0, fmt.Errorf("%s required", key)
	}
	return strconv.ParseInt(raw, 10, 64)
}

func adminEmailFromRequest(r *http.Request) string {
	if v := r.Context().Value("adminEmail"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	if h := strings.TrimSpace(r.Header.Get("X-Admin-Email")); h != "" {
		return h
	}
	return "sistema"
}
