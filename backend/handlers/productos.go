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

func jsonPayloadKeys(body []byte) map[string]bool {
	keys := make(map[string]bool)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return keys
	}
	for key, value := range raw {
		if len(value) == 0 || string(value) == "null" {
			continue
		}
		keys[key] = true
	}
	return keys
}

func hasJSONPayloadValue(keys map[string]bool, key string) bool {
	return keys != nil && keys[key]
}

func validateProductoCamposObligatorios(dbEmp *sql.DB, p dbpkg.Producto, stockInicial float64, isCreate bool, keys map[string]bool) error {
	conf, err := dbpkg.GetEmpresaInventarioConfiguracion(dbEmp, p.EmpresaID)
	if err != nil {
		return fmt.Errorf("no se pudo cargar la configuracion de campos obligatorios: %w", err)
	}
	required := conf.ProductoCamposObligatorios
	missing := make([]string, 0)
	addMissing := func(label string) {
		missing = append(missing, label)
	}
	if required.SKU && strings.TrimSpace(p.SKU) == "" {
		addMissing("SKU")
	}
	if required.CodigoBarras && strings.TrimSpace(p.CodigoBarras) == "" {
		addMissing("codigo de barras")
	}
	if required.CategoriaID && p.CategoriaID <= 0 {
		addMissing("categoria")
	}
	if required.Marca && strings.TrimSpace(p.Marca) == "" {
		addMissing("marca")
	}
	if required.UnidadMedida && strings.TrimSpace(p.UnidadMedida) == "" {
		addMissing("unidad de medida")
	}
	if required.Costo && !hasJSONPayloadValue(keys, "costo") {
		addMissing("costo")
	}
	if required.Precio && !hasJSONPayloadValue(keys, "precio") {
		addMissing("precio")
	}
	if required.ImpuestoPorcentaje && !hasJSONPayloadValue(keys, "impuesto_porcentaje") {
		addMissing("impuesto")
	}
	if required.StockMinimo && !hasJSONPayloadValue(keys, "stock_minimo") {
		addMissing("stock minimo")
	}
	if required.StockMaximo && !hasJSONPayloadValue(keys, "stock_maximo") {
		addMissing("stock maximo")
	}
	if required.StockInicial && isCreate && (!hasJSONPayloadValue(keys, "stock_inicial") || stockInicial < 0) {
		addMissing("stock inicial")
	}
	if required.BodegaPrincipalID && p.BodegaPrincipalID <= 0 {
		addMissing("bodega principal")
	}
	if required.ProveedorPrincipalID && p.ProveedorPrincipalID <= 0 {
		addMissing("proveedor principal")
	}
	if required.ImagenURL && strings.TrimSpace(p.ImagenURL) == "" {
		addMissing("URL de imagen")
	}
	if required.Descripcion && strings.TrimSpace(p.Descripcion) == "" {
		addMissing("descripcion")
	}
	if required.Observaciones && strings.TrimSpace(p.Observaciones) == "" {
		addMissing("observaciones")
	}
	if required.ManejaVencimiento && !p.ManejaVencimiento {
		addMissing("control de vencimiento")
	}
	if required.FechaVencimiento && strings.TrimSpace(p.FechaVencimiento) == "" {
		addMissing("fecha de vencimiento")
	}
	if required.DiasAlertaVencimiento && (!hasJSONPayloadValue(keys, "dias_alerta_vencimiento") || p.DiasAlertaVencimiento <= 0) {
		addMissing("dias de alerta de vencimiento")
	}
	if required.LoteCodigo && strings.TrimSpace(p.LoteCodigo) == "" {
		addMissing("lote")
	}
	if len(missing) > 0 {
		return fmt.Errorf("campos obligatorios por configuracion de la empresa: %s", strings.Join(missing, ", "))
	}
	return nil
}

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
				writeBodegaPersistenceError(w, err)
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
				writeBodegaPersistenceError(w, err)
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

// EmpresaCategoriasProductosHandler maneja CRUD de categorías de productos por empresa.
func EmpresaCategoriasProductosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := r.URL.Query().Get("include_inactive") == "1"
			q := r.URL.Query().Get("q")
			rows, err := dbpkg.GetCategoriasProductoByEmpresa(dbEmp, empresaID, includeInactive, q)
			if err != nil {
				http.Error(w, "failed to list categorias de productos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return
		case http.MethodPost:
			var payload dbpkg.CategoriaProducto
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
			id, err := dbpkg.CreateCategoriaProducto(dbEmp, payload)
			if err != nil {
				http.Error(w, "failed to create categoria de producto: "+err.Error(), http.StatusInternalServerError)
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
				if err := dbpkg.SetCategoriaProductoEstado(dbEmp, empresaID, id, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "categoria de producto no encontrada", http.StatusNotFound)
						return
					}
					http.Error(w, "failed to set categoria estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			var payload dbpkg.CategoriaProducto
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
			if err := dbpkg.UpdateCategoriaProducto(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "categoria de producto no encontrada", http.StatusNotFound)
					return
				}
				http.Error(w, "failed to update categoria de producto: "+err.Error(), http.StatusInternalServerError)
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
			if err := dbpkg.DeleteCategoriaProducto(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "categoria de producto no encontrada", http.StatusNotFound)
					return
				}
				if strings.Contains(strings.ToLower(err.Error()), "asociada") {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				http.Error(w, "failed to delete categoria de producto: "+err.Error(), http.StatusInternalServerError)
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
			qParams := r.URL.Query()
			action := strings.ToLower(strings.TrimSpace(qParams.Get("action")))
			if action == "vencimientos" || action == "alertas_vencimiento" || action == "por_vencer" {
				diasVentana, _ := parseIntQueryOptional(r, "dias")
				limit, _ := parseIntQueryOptional(r, "limit")
				offset, _ := parseIntQueryOptional(r, "offset")
				rows, err := dbpkg.GetProductosVencimientoByEmpresa(dbEmp, empresaID, qParams.Get("estado_vencimiento"), diasVentana, limit, offset)
				if err != nil {
					http.Error(w, "failed to list productos por vencimiento: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(rows)
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
			categoriaID, _ := parseInt64QueryOptional(r, "categoria_id")
			limit, _ := parseIntQueryOptional(r, "limit")
			offset, _ := parseIntQueryOptional(r, "offset")
			rows, err := dbpkg.GetProductosByEmpresa(dbEmp, empresaID, q, estado, bodegaID, categoriaID, limit, offset)
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
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			payloadKeys := jsonPayloadKeys(body)
			if err := json.Unmarshal(body, &payload); err != nil {
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
			if payload.StockInicial > 0 && payload.BodegaPrincipalID <= 0 {
				http.Error(w, "bodega_principal_id required when stock_inicial is greater than zero", http.StatusBadRequest)
				return
			}
			if err := validateProductoCamposObligatorios(dbEmp, payload.Producto, payload.StockInicial, true, payloadKeys); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
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
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "producto no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			var payload dbpkg.Producto
			var motivoPrecio string
			var referenciaPrecio string
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			payloadKeys := jsonPayloadKeys(body)
			if err := json.Unmarshal(body, &struct {
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
			if err := validateProductoCamposObligatorios(dbEmp, payload, 0, false, payloadKeys); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = adminEmailFromRequest(r)
			if err := dbpkg.UpdateProducto(dbEmp, payload, motivoPrecio, referenciaPrecio); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "producto no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "failed to update producto: "+err.Error(), http.StatusInternalServerError)
				return
			}
			updated, err := dbpkg.GetProductoByID(dbEmp, payload.EmpresaID, payload.ID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "producto no encontrado despues de actualizar", http.StatusNotFound)
					return
				}
				http.Error(w, "producto actualizado, pero no se pudo recargar: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updated)
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
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "producto no encontrado", http.StatusNotFound)
					return
				}
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
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

		if action == "alertas" || action == "alertas_quiebre" || action == "quiebre" {
			rows, err := dbpkg.GetAlertasQuiebreByEmpresa(dbEmp, empresaID, productoID, bodegaID, limit, offset)
			if err != nil {
				http.Error(w, "failed to list alertas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return
		}

		rows, err := dbpkg.GetExistenciasByEmpresa(dbEmp, empresaID, productoID, bodegaID, limit, offset)
		if err != nil {
			http.Error(w, "failed to list existencias: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioAlertasHandler lista alertas de quiebre/bajo minimo por bodega.
func EmpresaInventarioAlertasHandler(dbEmp *sql.DB) http.HandlerFunc {
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
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		modo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("modo")))
		proactivo := action == "proactivas" || action == "operativas" || action == "sobrestock" || modo == "proactivo" || r.URL.Query().Get("proactivo") == "1"

		if proactivo {
			rows, err := dbpkg.GetAlertasOperativasByEmpresa(dbEmp, empresaID, productoID, bodegaID, limit, offset)
			if err != nil {
				http.Error(w, "failed to list alertas operativas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return
		}

		rows, err := dbpkg.GetAlertasQuiebreByEmpresa(dbEmp, empresaID, productoID, bodegaID, limit, offset)
		if err != nil {
			http.Error(w, "failed to list alertas: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioConfiguracionHandler gestiona configuracion operativa de inventario por empresa.
func EmpresaInventarioConfiguracionHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			conf, err := dbpkg.GetEmpresaInventarioConfiguracion(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "failed to get inventario config: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(conf)
			return
		case http.MethodPut:
			var payload struct {
				EmpresaID                  int64                             `json:"empresa_id"`
				PoliticaCosto              string                            `json:"politica_costo"`
				ProductoCamposObligatorios *dbpkg.ProductoCamposObligatorios `json:"producto_campos_obligatorios"`
				Observaciones              string                            `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id requerido", http.StatusBadRequest)
				return
			}
			politica := strings.ToLower(strings.TrimSpace(payload.PoliticaCosto))
			switch politica {
			case "", "promedio":
				politica = "promedio"
			case "peps", "fifo":
				politica = "peps"
			default:
				http.Error(w, "politica_costo invalida (valores permitidos: promedio, peps)", http.StatusBadRequest)
				return
			}

			productoCampos := dbpkg.ProductoCamposObligatorios{}
			if payload.ProductoCamposObligatorios != nil {
				productoCampos = *payload.ProductoCamposObligatorios
			} else {
				actual, err := dbpkg.GetEmpresaInventarioConfiguracion(dbEmp, payload.EmpresaID)
				if err != nil {
					http.Error(w, "failed to get inventario config actual: "+err.Error(), http.StatusInternalServerError)
					return
				}
				productoCampos = actual.ProductoCamposObligatorios
			}

			conf, err := dbpkg.UpsertEmpresaInventarioConfiguracion(dbEmp, dbpkg.EmpresaInventarioConfiguracion{
				EmpresaID:                  payload.EmpresaID,
				PoliticaCosto:              politica,
				ProductoCamposObligatorios: productoCampos,
				UsuarioCreador:             adminEmailFromRequest(r),
				Estado:                     "activo",
				Observaciones:              strings.TrimSpace(payload.Observaciones),
			})
			if err != nil {
				http.Error(w, "failed to save inventario config: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(conf)
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaInventarioConteoCiclicoHandler gestiona conteos ciclicos con ajuste auditado.
func EmpresaInventarioConteoCiclicoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			productoID, err := parseInt64QueryOptional(r, "producto_id")
			if err != nil {
				http.Error(w, "producto_id invalido", http.StatusBadRequest)
				return
			}
			bodegaID, err := parseInt64QueryOptional(r, "bodega_id")
			if err != nil {
				http.Error(w, "bodega_id invalido", http.StatusBadRequest)
				return
			}
			estadoConteo := strings.TrimSpace(r.URL.Query().Get("estado_conteo"))
			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			if desde != "" && !isISODate(desde) {
				http.Error(w, "desde debe usar formato YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			if hasta != "" && !isISODate(hasta) {
				http.Error(w, "hasta debe usar formato YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			limit, _ := parseIntQueryOptional(r, "limit")
			offset, _ := parseIntQueryOptional(r, "offset")

			rows, err := dbpkg.GetInventarioConteosCiclicosByEmpresa(dbEmp, empresaID, productoID, bodegaID, estadoConteo, desde, hasta, limit, offset)
			if err != nil {
				http.Error(w, "failed to list conteos ciclicos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return
		case http.MethodPost:
			var payload struct {
				EmpresaID       int64   `json:"empresa_id"`
				ProductoID      int64   `json:"producto_id"`
				BodegaID        int64   `json:"bodega_id"`
				CantidadContada float64 `json:"cantidad_contada"`
				Referencia      string  `json:"referencia"`
				Observaciones   string  `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ProductoID <= 0 || payload.BodegaID <= 0 {
				http.Error(w, "empresa_id, producto_id y bodega_id son obligatorios", http.StatusBadRequest)
				return
			}
			if payload.CantidadContada < 0 {
				http.Error(w, "cantidad_contada no puede ser negativa", http.StatusBadRequest)
				return
			}

			conteo, err := dbpkg.RegistrarConteoCiclicoInventario(dbEmp, dbpkg.InventarioConteoCiclico{
				EmpresaID:       payload.EmpresaID,
				ProductoID:      payload.ProductoID,
				BodegaID:        payload.BodegaID,
				CantidadContada: payload.CantidadContada,
				Referencia:      strings.TrimSpace(payload.Referencia),
				UsuarioRevisor:  adminEmailFromRequest(r),
				UsuarioCreador:  adminEmailFromRequest(r),
				Observaciones:   strings.TrimSpace(payload.Observaciones),
			})
			if err != nil {
				if errors.Is(err, dbpkg.ErrStockInsuficiente) {
					http.Error(w, "stock insuficiente para ajuste de conteo", http.StatusConflict)
					return
				}
				http.Error(w, "failed to register conteo ciclico: "+err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": conteo.ID, "resultado": conteo})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaInventarioResumenHandler devuelve KPI operativos de inventario por empresa.
func EmpresaInventarioResumenHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		desde := strings.TrimSpace(r.URL.Query().Get("desde"))
		hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
		if desde != "" && !isISODate(desde) {
			http.Error(w, "desde debe usar formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		if hasta != "" && !isISODate(hasta) {
			http.Error(w, "hasta debe usar formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		resumen, err := dbpkg.GetInventarioResumenByEmpresa(dbEmp, empresaID, desde, hasta)
		if err != nil {
			http.Error(w, "failed to build inventario resumen: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resumen)
	}
}

// EmpresaInventarioTendenciaHandler devuelve tendencia diaria de inventario por empresa.
func EmpresaInventarioTendenciaHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		bodegaID, err := parseInt64QueryOptional(r, "bodega_id")
		if err != nil {
			http.Error(w, "bodega_id invalido", http.StatusBadRequest)
			return
		}
		dias, err := parseIntQueryOptional(r, "dias")
		if err != nil {
			http.Error(w, "dias invalido", http.StatusBadRequest)
			return
		}
		if dias <= 0 {
			dias = 7
		}
		if dias > 120 {
			dias = 120
		}

		desde := strings.TrimSpace(r.URL.Query().Get("desde"))
		hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
		if desde != "" && !isISODate(desde) {
			http.Error(w, "desde debe usar formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		if hasta != "" && !isISODate(hasta) {
			http.Error(w, "hasta debe usar formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		rows, err := dbpkg.GetInventarioTendenciaByEmpresa(dbEmp, empresaID, bodegaID, desde, hasta, dias)
		if err != nil {
			http.Error(w, "failed to build inventario tendencia: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioBalanceBodegasHandler devuelve el balance operativo por bodega.
func EmpresaInventarioBalanceBodegasHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		bodegaID, err := parseInt64QueryOptional(r, "bodega_id")
		if err != nil {
			http.Error(w, "bodega_id invalido", http.StatusBadRequest)
			return
		}
		dias, err := parseIntQueryOptional(r, "dias")
		if err != nil {
			http.Error(w, "dias invalido", http.StatusBadRequest)
			return
		}
		if dias <= 0 {
			dias = 7
		}
		if dias > 120 {
			dias = 120
		}

		desde := strings.TrimSpace(r.URL.Query().Get("desde"))
		hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
		if desde != "" && !isISODate(desde) {
			http.Error(w, "desde debe usar formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		if hasta != "" && !isISODate(hasta) {
			http.Error(w, "hasta debe usar formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		rows, err := dbpkg.GetInventarioBalanceBodegasByEmpresa(dbEmp, empresaID, bodegaID, desde, hasta, dias)
		if err != nil {
			http.Error(w, "failed to build inventario balance por bodega: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioProyeccionQuiebreHandler estima quiebre por producto/bodega.
func EmpresaInventarioProyeccionQuiebreHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		bodegaID, err := parseInt64QueryOptional(r, "bodega_id")
		if err != nil {
			http.Error(w, "bodega_id invalido", http.StatusBadRequest)
			return
		}
		diasVentana, err := parseIntQueryOptional(r, "dias_ventana")
		if err != nil {
			http.Error(w, "dias_ventana invalido", http.StatusBadRequest)
			return
		}
		if diasVentana <= 0 {
			diasVentana = 30
		}
		if diasVentana > 180 {
			diasVentana = 180
		}
		limit, err := parseIntQueryOptional(r, "limit")
		if err != nil {
			http.Error(w, "limit invalido", http.StatusBadRequest)
			return
		}
		offset, err := parseIntQueryOptional(r, "offset")
		if err != nil {
			http.Error(w, "offset invalido", http.StatusBadRequest)
			return
		}

		rows, err := dbpkg.GetInventarioProyeccionQuiebreByEmpresa(dbEmp, empresaID, bodegaID, diasVentana, limit, offset)
		if err != nil {
			http.Error(w, "failed to build inventario proyeccion quiebre: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioPlanReposicionHandler consolida sugerencias de reposicion por proveedor.
func EmpresaInventarioPlanReposicionHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		bodegaID, err := parseInt64QueryOptional(r, "bodega_id")
		if err != nil {
			http.Error(w, "bodega_id invalido", http.StatusBadRequest)
			return
		}
		diasVentana, err := parseIntQueryOptional(r, "dias_ventana")
		if err != nil {
			http.Error(w, "dias_ventana invalido", http.StatusBadRequest)
			return
		}
		if diasVentana <= 0 {
			diasVentana = 30
		}
		if diasVentana > 180 {
			diasVentana = 180
		}

		soloRiesgo := true
		rawSoloRiesgo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("solo_riesgo")))
		if rawSoloRiesgo != "" {
			switch rawSoloRiesgo {
			case "1", "true", "si", "yes":
				soloRiesgo = true
			case "0", "false", "no":
				soloRiesgo = false
			default:
				http.Error(w, "solo_riesgo invalido", http.StatusBadRequest)
				return
			}
		}

		limit, err := parseIntQueryOptional(r, "limit")
		if err != nil {
			http.Error(w, "limit invalido", http.StatusBadRequest)
			return
		}
		offset, err := parseIntQueryOptional(r, "offset")
		if err != nil {
			http.Error(w, "offset invalido", http.StatusBadRequest)
			return
		}

		rows, err := dbpkg.GetInventarioPlanReposicionByEmpresa(dbEmp, empresaID, bodegaID, diasVentana, soloRiesgo, limit, offset)
		if err != nil {
			http.Error(w, "failed to build inventario plan reposicion: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioPlanReposicionResumenHandler devuelve el consolidado de compra por proveedor.
func EmpresaInventarioPlanReposicionResumenHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		bodegaID, err := parseInt64QueryOptional(r, "bodega_id")
		if err != nil {
			http.Error(w, "bodega_id invalido", http.StatusBadRequest)
			return
		}
		diasVentana, err := parseIntQueryOptional(r, "dias_ventana")
		if err != nil {
			http.Error(w, "dias_ventana invalido", http.StatusBadRequest)
			return
		}
		if diasVentana <= 0 {
			diasVentana = 30
		}
		if diasVentana > 180 {
			diasVentana = 180
		}

		soloRiesgo := true
		rawSoloRiesgo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("solo_riesgo")))
		if rawSoloRiesgo != "" {
			switch rawSoloRiesgo {
			case "1", "true", "si", "yes":
				soloRiesgo = true
			case "0", "false", "no":
				soloRiesgo = false
			default:
				http.Error(w, "solo_riesgo invalido", http.StatusBadRequest)
				return
			}
		}

		limit, err := parseIntQueryOptional(r, "limit")
		if err != nil {
			http.Error(w, "limit invalido", http.StatusBadRequest)
			return
		}
		offset, err := parseIntQueryOptional(r, "offset")
		if err != nil {
			http.Error(w, "offset invalido", http.StatusBadRequest)
			return
		}

		rows, err := dbpkg.GetInventarioPlanReposicionResumenByEmpresa(dbEmp, empresaID, bodegaID, diasVentana, soloRiesgo, limit, offset)
		if err != nil {
			http.Error(w, "failed to build inventario plan reposicion resumen: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rows)
	}
}

// EmpresaInventarioPlanReposicionBorradorHandler devuelve un borrador de orden de compra para un proveedor.
func EmpresaInventarioPlanReposicionBorradorHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		proveedorID, err := parseInt64QueryOptional(r, "proveedor_id")
		if err != nil {
			http.Error(w, "proveedor_id invalido", http.StatusBadRequest)
			return
		}
		if proveedorID <= 0 {
			http.Error(w, "proveedor_id es obligatorio", http.StatusBadRequest)
			return
		}

		bodegaID, err := parseInt64QueryOptional(r, "bodega_id")
		if err != nil {
			http.Error(w, "bodega_id invalido", http.StatusBadRequest)
			return
		}
		diasVentana, err := parseIntQueryOptional(r, "dias_ventana")
		if err != nil {
			http.Error(w, "dias_ventana invalido", http.StatusBadRequest)
			return
		}
		if diasVentana <= 0 {
			diasVentana = 30
		}
		if diasVentana > 180 {
			diasVentana = 180
		}

		soloRiesgo := true
		rawSoloRiesgo := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("solo_riesgo")))
		if rawSoloRiesgo != "" {
			switch rawSoloRiesgo {
			case "1", "true", "si", "yes":
				soloRiesgo = true
			case "0", "false", "no":
				soloRiesgo = false
			default:
				http.Error(w, "solo_riesgo invalido", http.StatusBadRequest)
				return
			}
		}

		row, err := dbpkg.GetInventarioPlanReposicionBorradorByEmpresa(dbEmp, empresaID, proveedorID, bodegaID, diasVentana, soloRiesgo)
		if err != nil {
			http.Error(w, "failed to build inventario plan reposicion borrador: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(row)
	}
}

// EmpresaComprasPlanReposicionEmitirOrdenHandler emite una OC desde un borrador de reposicion por proveedor.
func EmpresaComprasPlanReposicionEmitirOrdenHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		q := r.URL.Query()
		var payload struct {
			EmpresaID       int64  `json:"empresa_id"`
			ProveedorID     int64  `json:"proveedor_id"`
			BodegaID        int64  `json:"bodega_id"`
			DiasVentana     int    `json:"dias_ventana"`
			SoloRiesgo      *bool  `json:"solo_riesgo"`
			DocumentoCodigo string `json:"documento_codigo"`
			PeriodoContable string `json:"periodo_contable"`
			Moneda          string `json:"moneda"`
			Observaciones   string `json:"observaciones"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if payload.EmpresaID <= 0 {
			empresaID, err := parseInt64QueryOptional(r, "empresa_id")
			if err != nil {
				http.Error(w, "empresa_id invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		if payload.ProveedorID <= 0 {
			proveedorID, err := parseInt64QueryOptional(r, "proveedor_id")
			if err != nil {
				http.Error(w, "proveedor_id invalido", http.StatusBadRequest)
				return
			}
			if proveedorID <= 0 {
				proveedorID, err = parseInt64QueryOptional(r, "id")
				if err != nil {
					http.Error(w, "id/proveedor_id invalido", http.StatusBadRequest)
					return
				}
			}
			payload.ProveedorID = proveedorID
		}
		if payload.ProveedorID <= 0 {
			http.Error(w, "proveedor_id es obligatorio", http.StatusBadRequest)
			return
		}

		if payload.BodegaID <= 0 {
			bodegaID, err := parseInt64QueryOptional(r, "bodega_id")
			if err != nil {
				http.Error(w, "bodega_id invalido", http.StatusBadRequest)
				return
			}
			payload.BodegaID = bodegaID
		}

		if payload.DiasVentana <= 0 {
			diasVentana, err := parseIntQueryOptional(r, "dias_ventana")
			if err != nil {
				http.Error(w, "dias_ventana invalido", http.StatusBadRequest)
				return
			}
			payload.DiasVentana = diasVentana
		}
		if payload.DiasVentana <= 0 {
			payload.DiasVentana = 30
		}
		if payload.DiasVentana > 180 {
			payload.DiasVentana = 180
		}

		soloRiesgo := true
		if payload.SoloRiesgo != nil {
			soloRiesgo = *payload.SoloRiesgo
		} else {
			rawSoloRiesgo := strings.ToLower(strings.TrimSpace(q.Get("solo_riesgo")))
			if rawSoloRiesgo != "" {
				switch rawSoloRiesgo {
				case "1", "true", "si", "yes":
					soloRiesgo = true
				case "0", "false", "no":
					soloRiesgo = false
				default:
					http.Error(w, "solo_riesgo invalido", http.StatusBadRequest)
					return
				}
			}
		}

		if strings.TrimSpace(payload.DocumentoCodigo) == "" {
			payload.DocumentoCodigo = strings.TrimSpace(q.Get("documento_codigo"))
		}
		if strings.TrimSpace(payload.PeriodoContable) == "" {
			payload.PeriodoContable = strings.TrimSpace(q.Get("periodo_contable"))
		}
		if strings.TrimSpace(payload.Moneda) == "" {
			payload.Moneda = strings.TrimSpace(q.Get("moneda"))
		}

		resultado, err := dbpkg.EmitirOrdenCompraDesdePlanReposicionBorrador(
			dbEmp,
			payload.EmpresaID,
			payload.ProveedorID,
			payload.BodegaID,
			payload.DiasVentana,
			soloRiesgo,
			payload.DocumentoCodigo,
			payload.PeriodoContable,
			payload.Moneda,
			adminEmailFromRequest(r),
			payload.Observaciones,
		)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "no hay items sugeridos") {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			http.Error(w, "failed to emitir orden de compra desde borrador: "+err.Error(), http.StatusInternalServerError)
			return
		}

		registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
			EmpresaID:       resultado.EmpresaID,
			Modulo:          "compras",
			Evento:          resultado.Evento,
			Entidad:         "orden_compra",
			EntidadID:       resultado.EntidadID,
			DocumentoTipo:   "orden_compra",
			DocumentoCodigo: strings.TrimSpace(resultado.DocumentoCodigo),
			PeriodoContable: strings.TrimSpace(resultado.PeriodoContable),
			MontoTotal:      resultado.CostoTotal,
			Moneda:          strings.ToUpper(strings.TrimSpace(resultado.Moneda)),
			Origen:          "api_compras_plan_reposicion",
			Observaciones:   strings.TrimSpace(payload.Observaciones),
		}, map[string]interface{}{
			"accion":           "emitir_orden",
			"estado_anterior":  resultado.EstadoAnterior,
			"estado_nuevo":     resultado.EstadoNuevo,
			"entidad_id":       resultado.EntidadID,
			"documento_codigo": strings.TrimSpace(resultado.DocumentoCodigo),
			"proveedor_id":     resultado.ProveedorID,
			"empresa_id":       resultado.EmpresaID,
			"total_items":      resultado.TotalItems,
			"costo_total":      resultado.CostoTotal,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":        true,
			"resultado": resultado,
		})
	}
}

// EmpresaComprasPlanReposicionActualizarEstadoHandler gestiona recepcion/contabilizacion de OC emitidas por reposicion.
func EmpresaComprasPlanReposicionActualizarEstadoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		q := r.URL.Query()
		var payload struct {
			EmpresaID       int64  `json:"empresa_id"`
			ProveedorID     int64  `json:"proveedor_id"`
			DocumentoCodigo string `json:"documento_codigo"`
			Accion          string `json:"accion"`
			EstadoActual    string `json:"estado_actual"`
			PeriodoContable string `json:"periodo_contable"`
			Observaciones   string `json:"observaciones"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && !errors.Is(err, io.EOF) {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if payload.EmpresaID <= 0 {
			empresaID, err := parseInt64QueryOptional(r, "empresa_id")
			if err != nil {
				http.Error(w, "empresa_id invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		if payload.ProveedorID <= 0 {
			proveedorID, err := parseInt64QueryOptional(r, "proveedor_id")
			if err != nil {
				http.Error(w, "proveedor_id invalido", http.StatusBadRequest)
				return
			}
			if proveedorID <= 0 {
				proveedorID, err = parseInt64QueryOptional(r, "id")
				if err != nil {
					http.Error(w, "id/proveedor_id invalido", http.StatusBadRequest)
					return
				}
			}
			payload.ProveedorID = proveedorID
		}
		if payload.ProveedorID <= 0 {
			http.Error(w, "proveedor_id es obligatorio", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(payload.DocumentoCodigo) == "" {
			payload.DocumentoCodigo = strings.TrimSpace(q.Get("documento_codigo"))
		}
		if strings.TrimSpace(payload.DocumentoCodigo) == "" {
			http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(payload.Accion) == "" {
			payload.Accion = strings.TrimSpace(q.Get("accion"))
		}
		if strings.TrimSpace(payload.Accion) == "" {
			http.Error(w, "accion es obligatoria", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(payload.EstadoActual) == "" {
			payload.EstadoActual = strings.TrimSpace(q.Get("estado_actual"))
		}
		if strings.TrimSpace(payload.PeriodoContable) == "" {
			payload.PeriodoContable = strings.TrimSpace(q.Get("periodo_contable"))
		}

		resultado, err := dbpkg.ActualizarEstadoOrdenCompraDesdeReposicion(
			dbEmp,
			payload.EmpresaID,
			payload.ProveedorID,
			payload.DocumentoCodigo,
			payload.Accion,
			payload.EstadoActual,
			payload.PeriodoContable,
			payload.Observaciones,
			adminEmailFromRequest(r),
		)
		if err != nil {
			errLower := strings.ToLower(err.Error())
			switch {
			case strings.Contains(errLower, "documento no encontrado"):
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			case strings.Contains(errLower, "transicion invalida"):
				http.Error(w, err.Error(), http.StatusConflict)
				return
			case strings.Contains(errLower, "accion no soportada"), strings.Contains(errLower, "obligatori"), strings.Contains(errLower, "invalido"):
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, "failed to actualizar estado de orden de compra de reposicion: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
			EmpresaID:       resultado.EmpresaID,
			Modulo:          "compras",
			Evento:          resultado.Evento,
			Entidad:         "orden_compra",
			EntidadID:       resultado.EntidadID,
			DocumentoTipo:   "orden_compra",
			DocumentoCodigo: strings.TrimSpace(resultado.DocumentoCodigo),
			PeriodoContable: strings.TrimSpace(resultado.PeriodoContable),
			MontoTotal:      resultado.MontoTotal,
			Moneda:          strings.ToUpper(strings.TrimSpace(resultado.Moneda)),
			Origen:          "api_compras_plan_reposicion",
			Observaciones:   strings.TrimSpace(payload.Observaciones),
		}, map[string]interface{}{
			"accion":           resultado.Accion,
			"estado_anterior":  resultado.EstadoAnterior,
			"estado_nuevo":     resultado.EstadoNuevo,
			"entidad_id":       resultado.EntidadID,
			"documento_codigo": strings.TrimSpace(resultado.DocumentoCodigo),
			"proveedor_id":     resultado.ProveedorID,
			"empresa_id":       resultado.EmpresaID,
			"monto_total":      resultado.MontoTotal,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":        true,
			"resultado": resultado,
		})
	}
}

func isISODate(raw string) bool {
	_, err := time.Parse("2006-01-02", strings.TrimSpace(raw))
	return err == nil
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
		bodegaID, _ := parseInt64QueryOptional(r, "bodega_id")
		tipo := strings.TrimSpace(r.URL.Query().Get("tipo"))
		desde := strings.TrimSpace(r.URL.Query().Get("desde"))
		hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
		if desde != "" && !isISODate(desde) {
			http.Error(w, "desde debe usar formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		if hasta != "" && !isISODate(hasta) {
			http.Error(w, "hasta debe usar formato YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		limit, _ := parseIntQueryOptional(r, "limit")
		offset, _ := parseIntQueryOptional(r, "offset")
		rows, err := dbpkg.GetMovimientosByEmpresa(dbEmp, empresaID, productoID, bodegaID, tipo, desde, hasta, limit, offset)
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
			var payload struct {
				dbpkg.Proveedor
				NombreComercial string `json:"nombre_comercial"`
				RazonSocial     string `json:"razon_social"`
				NIT             string `json:"nit"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				empresaID, err := parseEmpresaIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
			}
			if strings.TrimSpace(payload.Nombre) == "" {
				payload.Nombre = strings.TrimSpace(payload.NombreComercial)
			}
			if strings.TrimSpace(payload.Nombre) == "" {
				payload.Nombre = strings.TrimSpace(payload.RazonSocial)
			}
			if strings.TrimSpace(payload.Documento) == "" {
				payload.Documento = strings.TrimSpace(payload.NIT)
			}
			if payload.EmpresaID <= 0 || strings.TrimSpace(payload.Nombre) == "" {
				http.Error(w, "empresa_id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			if err := validateProveedorComercialPayload(payload.Proveedor); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = adminEmailFromRequest(r)
			id, err := dbpkg.CreateProveedor(dbEmp, payload.Proveedor)
			if err != nil {
				http.Error(w, "failed to create proveedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "compras",
				Evento:          "proveedor_registrado",
				Entidad:         "proveedor",
				EntidadID:       id,
				DocumentoTipo:   "proveedor",
				DocumentoCodigo: strings.TrimSpace(payload.Codigo),
				Origen:          "api_proveedores",
				Observaciones:   "alta de proveedor en modulo de compras",
			}, map[string]interface{}{
				"nombre":                  strings.TrimSpace(payload.Nombre),
				"documento":               strings.TrimSpace(payload.Documento),
				"contacto":                strings.TrimSpace(payload.Contacto),
				"catalogo_referencia":     strings.TrimSpace(payload.CatalogoReferencia),
				"precio_base_referencial": payload.PrecioBaseReferencial,
				"descuento_porcentaje":    payload.DescuentoPorcentaje,
				"plazo_pago_dias":         payload.PlazoPagoDias,
				"condicion_entrega":       strings.TrimSpace(payload.CondicionEntrega),
				"estado":                  "activo",
				"empresa_id":              payload.EmpresaID,
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			q := r.URL.Query()
			action := strings.ToLower(strings.TrimSpace(q.Get("action")))
			if action == "emitir" || action == "emitir_orden" || action == "recepcionar" || action == "recepcionar_compra" || action == "contabilizar" || action == "contabilizar_compra" {
				var payload struct {
					EmpresaID       int64   `json:"empresa_id"`
					ProveedorID     int64   `json:"proveedor_id"`
					DocumentoCodigo string  `json:"documento_codigo"`
					EstadoActual    string  `json:"estado_actual"`
					MontoTotal      float64 `json:"monto_total"`
					Moneda          string  `json:"moneda"`
					PeriodoContable string  `json:"periodo_contable"`
					Observaciones   string  `json:"observaciones"`
				}
				if r.Body != nil {
					_ = json.NewDecoder(r.Body).Decode(&payload)
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
				if payload.ProveedorID <= 0 {
					if proveedorID, err := parseInt64QueryOptional(r, "id"); err == nil && proveedorID > 0 {
						payload.ProveedorID = proveedorID
					}
				}
				if payload.ProveedorID <= 0 {
					http.Error(w, "id/proveedor_id es obligatorio para la accion", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(q.Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio para la accion", http.StatusBadRequest)
					return
				}

				if strings.TrimSpace(payload.EstadoActual) == "" {
					payload.EstadoActual = strings.TrimSpace(q.Get("estado_actual"))
				}

				docExistente, err := dbpkg.GetEmpresaDocumentoCompraByCodigo(dbEmp, payload.EmpresaID, "orden_compra", payload.DocumentoCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar el estado documental de compras", http.StatusInternalServerError)
					return
				}
				if docExistente != nil {
					payload.EstadoActual = docExistente.EstadoDocumento
				}

				transition, err := resolveComprasTransition(action, payload.EstadoActual)
				if err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}

				evento := transition.Evento
				docPersistido, err := dbpkg.UpsertEmpresaDocumentoCompra(dbEmp, dbpkg.EmpresaDocumentoCompra{
					EmpresaID:            payload.EmpresaID,
					ProveedorID:          payload.ProveedorID,
					TipoDocumento:        "orden_compra",
					DocumentoCodigo:      payload.DocumentoCodigo,
					EstadoDocumento:      transition.EstadoNuevo,
					EstadoAnterior:       transition.EstadoAnterior,
					EventoUltimo:         evento,
					PeriodoContable:      payload.PeriodoContable,
					MontoTotal:           payload.MontoTotal,
					Moneda:               payload.Moneda,
					EntidadRelacionadaID: payload.ProveedorID,
					UsuarioCreador:       strings.TrimSpace(adminEmailFromRequest(r)),
					Observaciones:        payload.Observaciones,
				})
				if err != nil {
					http.Error(w, "No se pudo persistir el documento transaccional", http.StatusInternalServerError)
					return
				}

				registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
					EmpresaID:       payload.EmpresaID,
					Modulo:          "compras",
					Evento:          evento,
					Entidad:         "orden_compra",
					EntidadID:       docPersistido.ID,
					DocumentoTipo:   "orden_compra",
					DocumentoCodigo: strings.TrimSpace(payload.DocumentoCodigo),
					PeriodoContable: strings.TrimSpace(payload.PeriodoContable),
					MontoTotal:      payload.MontoTotal,
					Moneda:          strings.ToUpper(strings.TrimSpace(payload.Moneda)),
					Origen:          "api_proveedores",
					Observaciones:   strings.TrimSpace(payload.Observaciones),
				}, map[string]interface{}{
					"accion":           transition.Accion,
					"estado_anterior":  transition.EstadoAnterior,
					"estado_nuevo":     transition.EstadoNuevo,
					"entidad_id":       docPersistido.ID,
					"documento_codigo": strings.TrimSpace(payload.DocumentoCodigo),
					"proveedor_id":     payload.ProveedorID,
					"empresa_id":       payload.EmpresaID,
				})

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"ok":               true,
					"accion":           transition.Accion,
					"evento":           evento,
					"estado_anterior":  transition.EstadoAnterior,
					"estado_nuevo":     transition.EstadoNuevo,
					"entidad_id":       docPersistido.ID,
					"documento_codigo": strings.TrimSpace(payload.DocumentoCodigo),
				})
				return
			}
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
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "proveedor no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "failed to set proveedor estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				evento := "proveedor_desactivado"
				if estado == "activo" {
					evento = "proveedor_activado"
				}
				registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
					EmpresaID:       empresaID,
					Modulo:          "compras",
					Evento:          evento,
					Entidad:         "proveedor",
					EntidadID:       id,
					DocumentoTipo:   "proveedor",
					DocumentoCodigo: strconv.FormatInt(id, 10),
					Origen:          "api_proveedores",
					Observaciones:   "actualizacion de estado del proveedor",
				}, map[string]interface{}{
					"estado":     estado,
					"empresa_id": empresaID,
					"id":         id,
				})
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
			if err := validateProveedorComercialPayload(payload); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateProveedor(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "proveedor no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "failed to update proveedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "compras",
				Evento:          "proveedor_actualizado",
				Entidad:         "proveedor",
				EntidadID:       payload.ID,
				DocumentoTipo:   "proveedor",
				DocumentoCodigo: strings.TrimSpace(payload.Codigo),
				Origen:          "api_proveedores",
				Observaciones:   "actualizacion de proveedor en modulo de compras",
			}, map[string]interface{}{
				"nombre":                  strings.TrimSpace(payload.Nombre),
				"documento":               strings.TrimSpace(payload.Documento),
				"contacto":                strings.TrimSpace(payload.Contacto),
				"catalogo_referencia":     strings.TrimSpace(payload.CatalogoReferencia),
				"precio_base_referencial": payload.PrecioBaseReferencial,
				"descuento_porcentaje":    payload.DescuentoPorcentaje,
				"plazo_pago_dias":         payload.PlazoPagoDias,
				"condicion_entrega":       strings.TrimSpace(payload.CondicionEntrega),
				"empresa_id":              payload.EmpresaID,
				"id":                      payload.ID,
			})
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
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "proveedor no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "failed to delete proveedor: "+err.Error(), http.StatusInternalServerError)
				return
			}
			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       empresaID,
				Modulo:          "compras",
				Evento:          "proveedor_eliminado",
				Entidad:         "proveedor",
				EntidadID:       id,
				DocumentoTipo:   "proveedor",
				DocumentoCodigo: strconv.FormatInt(id, 10),
				Origen:          "api_proveedores",
				Observaciones:   "eliminacion de proveedor en modulo de compras",
			}, map[string]interface{}{
				"empresa_id": empresaID,
				"id":         id,
			})
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
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		productoID, err := parseInt64Form(r, "producto_id")
		if err != nil || productoID <= 0 {
			http.Error(w, "producto_id required", http.StatusBadRequest)
			return
		}
		existing, err := dbpkg.GetProductoByID(dbEmp, empresaID, productoID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "producto no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to query producto: "+err.Error(), http.StatusInternalServerError)
			return
		}
		oldImageURL := strings.TrimSpace(existing.ImagenURL)

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
		const maxProductImageBytes = 10 << 20
		if header.Size > maxProductImageBytes {
			http.Error(w, "la imagen supera 10 MB", http.StatusBadRequest)
			return
		}

		dir, publicDir, _ := empresaUploadsSubdir(dbEmp, empresaID, "imagenes", "productos")
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

		written, copyErr := io.Copy(out, io.LimitReader(file, maxProductImageBytes+1))
		closeErr := out.Close()
		if copyErr != nil {
			_ = os.Remove(absPath)
			http.Error(w, "failed to save image file", http.StatusInternalServerError)
			return
		}
		if written > maxProductImageBytes {
			_ = os.Remove(absPath)
			http.Error(w, "la imagen supera 10 MB", http.StatusBadRequest)
			return
		}
		if closeErr != nil {
			_ = os.Remove(absPath)
			http.Error(w, "failed to close image file", http.StatusInternalServerError)
			return
		}

		imageURL := publicDir + "/" + fileName
		if err := dbpkg.UpdateProductoImagen(dbEmp, empresaID, productoID, imageURL); err != nil {
			_ = os.Remove(absPath)
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "producto no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to update image url in producto: "+err.Error(), http.StatusInternalServerError)
			return
		}
		deletedPrevious := false
		if oldImageURL != "" && oldImageURL != imageURL {
			deletedPrevious = deleteEmpresaProductoUploadedPublicURL(dbEmp, empresaID, oldImageURL)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"saved":        true,
			"empresa_id":   empresaID,
			"producto_id":  productoID,
			"image_url":    imageURL,
			"replaced_old": deletedPrevious,
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
	if empresaID := parseEmpresaIDFromContext(r); empresaID > 0 {
		return empresaID, nil
	}
	empresaID, err := parseInt64Query(r, "empresa_id")
	if err != nil || empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id required")
	}
	return empresaID, nil
}

func writeBodegaPersistenceError(w http.ResponseWriter, err error) {
	errText := strings.ToLower(err.Error())
	if strings.Contains(errText, "ux_bodegas_empresa_nombre") ||
		strings.Contains(errText, "ux_bodegas_empresa_codigo") ||
		(strings.Contains(errText, "bodegas") && (strings.Contains(errText, "duplicate") || strings.Contains(errText, "duplic"))) {
		http.Error(w, "Ya existe una bodega con ese nombre o codigo para esta empresa.", http.StatusConflict)
		return
	}
	http.Error(w, "No se pudo guardar la bodega. Intenta de nuevo en unos segundos.", http.StatusInternalServerError)
}

func parseInt64Query(r *http.Request, key string) (int64, error) {
	if isEmpresaIDKey(key) {
		if empresaID := parseEmpresaIDFromContext(r); empresaID > 0 {
			return empresaID, nil
		}
	}
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, fmt.Errorf("%s required", key)
	}
	return strconv.ParseInt(raw, 10, 64)
}

func parseInt64QueryOptional(r *http.Request, key string) (int64, error) {
	if isEmpresaIDKey(key) {
		if empresaID := parseEmpresaIDFromContext(r); empresaID > 0 {
			return empresaID, nil
		}
	}
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func isEmpresaIDKey(key string) bool {
	return strings.EqualFold(strings.TrimSpace(key), "empresa_id")
}

func parseEmpresaIDFromContext(r *http.Request) int64 {
	if r == nil {
		return 0
	}
	v := r.Context().Value("empresaID")
	if v == nil {
		return 0
	}
	switch id := v.(type) {
	case int64:
		if id > 0 {
			return id
		}
	case int:
		if id > 0 {
			return int64(id)
		}
	case float64:
		if id > 0 {
			return int64(id)
		}
	case string:
		if parsed, err := strconv.ParseInt(strings.TrimSpace(id), 10, 64); err == nil && parsed > 0 {
			return parsed
		}
	}
	return 0
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

func validateProveedorComercialPayload(payload dbpkg.Proveedor) error {
	if payload.PrecioBaseReferencial < 0 {
		return fmt.Errorf("precio_base_referencial no puede ser negativo")
	}
	if payload.DescuentoPorcentaje < 0 || payload.DescuentoPorcentaje > 100 {
		return fmt.Errorf("descuento_porcentaje debe estar entre 0 y 100")
	}
	if payload.PlazoPagoDias < 0 {
		return fmt.Errorf("plazo_pago_dias no puede ser negativo")
	}
	return nil
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

func adminRoleFromRequest(r *http.Request) string {
	if v := r.Context().Value("adminRole"); v != nil {
		if s, ok := v.(string); ok {
			trim := strings.TrimSpace(s)
			if trim != "" {
				return trim
			}
		}
	}
	if h := strings.TrimSpace(r.Header.Get("X-Admin-Role")); h != "" {
		return h
	}
	return ""
}
