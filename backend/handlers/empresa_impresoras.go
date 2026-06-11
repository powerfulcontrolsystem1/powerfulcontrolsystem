package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func parsePrinterIncludeInactive(r *http.Request) bool {
	raw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("include_inactive")))
	return raw == "1" || raw == "true" || raw == "si" || raw == "yes"
}

func parsePrinterIDQuery(r *http.Request) (int64, error) {
	if id, err := parseInt64QueryOptional(r, "impresora_id"); err == nil && id > 0 {
		return id, nil
	}
	if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
		return id, nil
	}
	return 0, fmt.Errorf("impresora_id requerido")
}

// EmpresaImpresorasHandler administra impresoras de empresa, asignación por funcionalidad y por producto.
func EmpresaImpresorasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureEmpresaImpresorasSchema(dbEmp); err != nil {
			log.Printf("[empresa_impresoras] ensure schema error: %v", err)
			http.Error(w, "No se pudo preparar configuracion de impresoras", http.StatusInternalServerError)
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
			case "", "impresoras":
				rows, err := dbpkg.ListEmpresaImpresorasByEmpresa(dbEmp, empresaID, parsePrinterIncludeInactive(r))
				if err != nil {
					log.Printf("[empresa_impresoras] list impresoras empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar las impresoras", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "funcionalidades":
				rows, err := dbpkg.ListEmpresaImpresoraFuncionalidadesByEmpresa(dbEmp, empresaID)
				if err != nil {
					log.Printf("[empresa_impresoras] list funcionalidades empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar las funcionalidades", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "productos":
				rows, err := dbpkg.ListEmpresaImpresoraProductosByEmpresa(dbEmp, empresaID)
				if err != nil {
					log.Printf("[empresa_impresoras] list productos empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar las asignaciones por producto", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "producto_reglas":
				rows, err := dbpkg.ListEmpresaImpresoraProductoReglasByEmpresa(dbEmp, empresaID)
				if err != nil {
					log.Printf("[empresa_impresoras] list producto reglas empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar las reglas masivas por producto", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "recetas":
				rows, err := dbpkg.ListEmpresaImpresoraRecetasByEmpresa(dbEmp, empresaID)
				if err != nil {
					log.Printf("[empresa_impresoras] list recetas empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar las asignaciones por receta", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return

			case "cola":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				rows, err := dbpkg.ListEmpresaImpresoraCola(dbEmp, empresaID, estado, int64(limit))
				if err != nil {
					log.Printf("[empresa_impresoras] list cola empresa_id=%d estado=%q error: %v", empresaID, estado, err)
					http.Error(w, "No se pudo cargar la cola de impresion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":         true,
					"empresa_id": empresaID,
					"trabajos":   rows,
				})
				return

			case "catalogo_productos":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				if limit <= 0 {
					limit = 500
				}
				if limit > 1500 {
					limit = 1500
				}
				filtro := strings.TrimSpace(r.URL.Query().Get("filtro"))
				productos, err := dbpkg.GetProductosByEmpresa(dbEmp, empresaID, filtro, "activo", 0, 0, limit, 0)
				if err != nil {
					log.Printf("[empresa_impresoras] catalogo productos empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar productos", http.StatusInternalServerError)
					return
				}
				items := make([]map[string]interface{}, 0, len(productos))
				for _, p := range productos {
					items = append(items, map[string]interface{}{
						"id":            p.ID,
						"empresa_id":    p.EmpresaID,
						"categoria_id":  p.CategoriaID,
						"categoria":     p.Categoria,
						"nombre":        p.Nombre,
						"sku":           p.SKU,
						"codigo_barras": p.CodigoBarras,
						"estado":        p.Estado,
					})
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"productos": items,
					"total":     len(items),
				})
				return

			case "catalogo_categorias":
				filtro := strings.TrimSpace(r.URL.Query().Get("filtro"))
				categorias, err := dbpkg.GetCategoriasProductoByEmpresa(dbEmp, empresaID, false, filtro)
				if err != nil {
					log.Printf("[empresa_impresoras] catalogo categorias empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar categorias", http.StatusInternalServerError)
					return
				}
				items := make([]map[string]interface{}, 0, len(categorias))
				for _, c := range categorias {
					items = append(items, map[string]interface{}{
						"id":         c.ID,
						"empresa_id": c.EmpresaID,
						"nombre":     c.Nombre,
						"codigo":     c.Codigo,
						"estado":     c.Estado,
					})
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"categorias": items,
					"total":      len(items),
				})
				return

			case "catalogo_recetas":
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				if limit <= 0 {
					limit = 500
				}
				if limit > 1500 {
					limit = 1500
				}
				filtro := strings.TrimSpace(r.URL.Query().Get("filtro"))
				recetas, err := dbpkg.GetRecetasProductosByEmpresa(dbEmp, empresaID, filtro, "activo", true, limit, 0)
				if err != nil {
					log.Printf("[empresa_impresoras] catalogo recetas empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudieron cargar recetas", http.StatusInternalServerError)
					return
				}
				items := make([]map[string]interface{}, 0, len(recetas))
				for _, c := range recetas {
					items = append(items, map[string]interface{}{
						"id":         c.ID,
						"empresa_id": c.EmpresaID,
						"nombre":     c.Nombre,
						"codigo":     c.Codigo,
						"estado":     c.Estado,
					})
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"recetas": items,
					"total":   len(items),
				})
				return

			case "resolver":
				funcionalidad := strings.TrimSpace(r.URL.Query().Get("funcionalidad"))
				productoID, err := parseInt64QueryOptional(r, "producto_id")
				if err != nil {
					http.Error(w, "producto_id invalido", http.StatusBadRequest)
					return
				}
				recetaID, err := parseInt64QueryOptional(r, "receta_id")
				if err != nil {
					http.Error(w, "receta_id invalido", http.StatusBadRequest)
					return
				}
				tipoItem := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tipo_item")))
				referenciaID := productoID
				if tipoItem == "receta" || recetaID > 0 {
					tipoItem = "receta"
					referenciaID = recetaID
				} else if productoID > 0 {
					tipoItem = "producto"
				}
				resolved, err := dbpkg.ResolveEmpresaImpresoraOperacion(dbEmp, empresaID, funcionalidad, tipoItem, referenciaID)
				if err != nil {
					log.Printf("[empresa_impresoras] resolver empresa_id=%d funcionalidad=%q tipo_item=%q referencia_id=%d error: %v", empresaID, funcionalidad, tipoItem, referenciaID, err)
					http.Error(w, "No se pudo resolver impresora", http.StatusInternalServerError)
					return
				}
				if resolved == nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"ok":            false,
						"empresa_id":    empresaID,
						"funcionalidad": funcionalidad,
						"producto_id":   productoID,
						"receta_id":     recetaID,
						"tipo_item":     tipoItem,
						"message":       "No hay impresora configurada para el contexto solicitado",
					})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":         true,
					"resolucion": resolved,
					"impresora":  resolved.Impresora,
					"fuente":     resolved.Fuente,
				})
				return

			case "resumen":
				impresoras, err := dbpkg.ListEmpresaImpresorasByEmpresa(dbEmp, empresaID, true)
				if err != nil {
					log.Printf("[empresa_impresoras] resumen impresoras empresa_id=%d error: %v", empresaID, err)
					http.Error(w, "No se pudo cargar resumen de impresoras", http.StatusInternalServerError)
					return
				}
				warnings := make([]string, 0)
				funcionalidades, err := dbpkg.ListEmpresaImpresoraFuncionalidadesByEmpresa(dbEmp, empresaID)
				if err != nil {
					log.Printf("[empresa_impresoras] resumen funcionalidades empresa_id=%d error: %v", empresaID, err)
					funcionalidades = []dbpkg.EmpresaImpresoraFuncionalidad{}
					warnings = append(warnings, "No se pudieron cargar asignaciones por funcionalidad")
				}
				productos, err := dbpkg.ListEmpresaImpresoraProductosByEmpresa(dbEmp, empresaID)
				if err != nil {
					log.Printf("[empresa_impresoras] resumen productos empresa_id=%d error: %v", empresaID, err)
					productos = []dbpkg.EmpresaImpresoraProducto{}
					warnings = append(warnings, "No se pudieron cargar asignaciones por producto")
				}
				productoReglas, err := dbpkg.ListEmpresaImpresoraProductoReglasByEmpresa(dbEmp, empresaID)
				if err != nil {
					log.Printf("[empresa_impresoras] resumen producto reglas empresa_id=%d error: %v", empresaID, err)
					productoReglas = []dbpkg.EmpresaImpresoraProductoRegla{}
					warnings = append(warnings, "No se pudieron cargar reglas masivas por producto")
				}
				recetas, err := dbpkg.ListEmpresaImpresoraRecetasByEmpresa(dbEmp, empresaID)
				if err != nil {
					log.Printf("[empresa_impresoras] resumen recetas empresa_id=%d error: %v", empresaID, err)
					recetas = []dbpkg.EmpresaImpresoraReceta{}
					warnings = append(warnings, "No se pudieron cargar asignaciones por receta")
				}
				cola, err := dbpkg.ListEmpresaImpresoraCola(dbEmp, empresaID, "", 50)
				if err != nil {
					log.Printf("[empresa_impresoras] resumen cola empresa_id=%d error: %v", empresaID, err)
					cola = []dbpkg.EmpresaImpresoraTrabajo{}
					warnings = append(warnings, "No se pudo cargar la cola de impresion")
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"impresoras":      impresoras,
					"funcionalidades": funcionalidades,
					"productos":       productos,
					"producto_reglas": productoReglas,
					"recetas":         recetas,
					"cola":            cola,
					"warnings":        warnings,
				})
				return

			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodPost, http.MethodPut:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			switch action {
			case "", "impresora":
				var payload dbpkg.EmpresaImpresora
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaImpresora(dbEmp, payload)
				if err != nil {
					log.Printf("[empresa_impresoras] upsert impresora empresa_id=%d error: %v", empresaID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				item, err := dbpkg.GetEmpresaImpresoraByID(dbEmp, empresaID, id)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "impresora_guardada", "empresa_impresoras", id, http.StatusOK, map[string]interface{}{
					"codigo":            item.Codigo,
					"nombre":            item.Nombre,
					"tipo_conexion":     item.TipoConexion,
					"es_predeterminada": item.EsPredeterminada,
				}, "impresora empresarial creada o actualizada")
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":        true,
					"id":        id,
					"impresora": item,
				})
				return

			case "activar", "desactivar", "inactivar":
				impresoraID, err := parsePrinterIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				estado := "activo"
				if action == "desactivar" || action == "inactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetEmpresaImpresoraEstado(dbEmp, empresaID, impresoraID, estado, strings.TrimSpace(adminEmailFromRequest(r))); err != nil {
					log.Printf("[empresa_impresoras] set estado empresa_id=%d impresora_id=%d estado=%s error: %v", empresaID, impresoraID, estado, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "estado_impresora_actualizado", "empresa_impresoras", impresoraID, http.StatusOK, map[string]interface{}{
					"estado": estado,
				}, "estado de impresora actualizado")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "impresora_id": impresoraID, "estado": estado})
				return

			case "predeterminada", "default":
				impresoraID, err := parsePrinterIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaImpresoraPredeterminada(dbEmp, empresaID, impresoraID, strings.TrimSpace(adminEmailFromRequest(r))); err != nil {
					log.Printf("[empresa_impresoras] set default empresa_id=%d impresora_id=%d error: %v", empresaID, impresoraID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "predeterminada_actualizada", "empresa_impresoras", impresoraID, http.StatusOK, nil, "impresora predeterminada actualizada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "impresora_id": impresoraID})
				return

			case "funcionalidad":
				var payload dbpkg.EmpresaImpresoraFuncionalidad
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaImpresoraFuncionalidad(dbEmp, payload)
				if err != nil {
					log.Printf("[empresa_impresoras] upsert funcionalidad empresa_id=%d funcionalidad=%q error: %v", empresaID, payload.Funcionalidad, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "funcionalidad_asignada", "empresa_impresoras_funcionalidades", id, http.StatusOK, map[string]interface{}{
					"funcionalidad": payload.Funcionalidad,
					"impresora_id":  payload.ImpresoraID,
				}, "regla de impresion por funcionalidad guardada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return

			case "producto":
				var payload dbpkg.EmpresaImpresoraProducto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaImpresoraProducto(dbEmp, payload)
				if err != nil {
					log.Printf("[empresa_impresoras] upsert producto empresa_id=%d producto_id=%d error: %v", empresaID, payload.ProductoID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "producto_asignado", "empresa_impresoras_productos", id, http.StatusOK, map[string]interface{}{
					"producto_id":  payload.ProductoID,
					"impresora_id": payload.ImpresoraID,
				}, "regla de impresion por producto guardada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return

			case "producto_regla":
				var payload dbpkg.EmpresaImpresoraProductoRegla
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaImpresoraProductoRegla(dbEmp, payload)
				if err != nil {
					log.Printf("[empresa_impresoras] upsert producto regla empresa_id=%d alcance=%q categoria_id=%d error: %v", empresaID, payload.Alcance, payload.CategoriaID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "regla_producto_asignada", "empresa_impresoras_productos_reglas", id, http.StatusOK, map[string]interface{}{
					"alcance":      payload.Alcance,
					"categoria_id": payload.CategoriaID,
					"impresora_id": payload.ImpresoraID,
				}, "regla masiva de impresion por producto guardada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return

			case "receta":
				var payload dbpkg.EmpresaImpresoraReceta
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.UpsertEmpresaImpresoraReceta(dbEmp, payload)
				if err != nil {
					log.Printf("[empresa_impresoras] upsert receta empresa_id=%d receta_id=%d error: %v", empresaID, payload.RecetaID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "receta_asignada", "empresa_impresoras_recetas", id, http.StatusOK, map[string]interface{}{
					"receta_id":    payload.RecetaID,
					"impresora_id": payload.ImpresoraID,
				}, "regla de impresion por receta guardada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return

			case "cola_trabajo", "crear_trabajo", "trabajo":
				var payload dbpkg.EmpresaImpresoraTrabajo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
				id, err := dbpkg.CrearEmpresaImpresoraTrabajo(dbEmp, payload)
				if err != nil {
					log.Printf("[empresa_impresoras] crear trabajo empresa_id=%d error: %v", empresaID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				item, err := dbpkg.GetEmpresaImpresoraTrabajoByID(dbEmp, empresaID, id)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "trabajo_cola_creado", "empresa_impresoras_cola", id, http.StatusOK, map[string]interface{}{
					"funcionalidad":  item.Funcionalidad,
					"tipo_documento": item.TipoDocumento,
					"estacion_id":    item.EstacionID,
					"impresora_id":   item.ImpresoraID,
				}, "trabajo de impresion creado en cola")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "trabajo": item})
				return

			case "agente_tomar", "tomar_trabajos":
				var payload struct {
					AgenteID   string `json:"agente_id"`
					EstacionID int64  `json:"estacion_id"`
					Limit      int64  `json:"limit"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if strings.TrimSpace(payload.AgenteID) == "" {
					payload.AgenteID = strings.TrimSpace(r.URL.Query().Get("agente_id"))
				}
				if payload.EstacionID <= 0 {
					if id, parseErr := parseInt64QueryOptional(r, "estacion_id"); parseErr == nil {
						payload.EstacionID = id
					}
				}
				if payload.Limit <= 0 {
					if limit, parseErr := parseIntQueryOptional(r, "limit"); parseErr == nil {
						payload.Limit = int64(limit)
					}
				}
				trabajos, err := dbpkg.TomarEmpresaImpresoraTrabajos(dbEmp, empresaID, payload.AgenteID, payload.EstacionID, payload.Limit)
				if err != nil {
					log.Printf("[empresa_impresoras] tomar trabajos empresa_id=%d agente=%q error: %v", empresaID, payload.AgenteID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "trabajos": trabajos})
				return

			case "cola_estado", "trabajo_estado":
				var payload struct {
					ID          int64  `json:"id"`
					TrabajoID   int64  `json:"trabajo_id"`
					Estado      string `json:"estado"`
					AgenteID    string `json:"agente_id"`
					UltimoError string `json:"ultimo_error"`
					Error       string `json:"error"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.TrabajoID <= 0 {
					payload.TrabajoID = payload.ID
				}
				if strings.TrimSpace(payload.UltimoError) == "" {
					payload.UltimoError = payload.Error
				}
				if err := dbpkg.ActualizarEmpresaImpresoraTrabajoEstado(dbEmp, empresaID, payload.TrabajoID, payload.Estado, payload.AgenteID, payload.UltimoError); err != nil {
					log.Printf("[empresa_impresoras] actualizar trabajo empresa_id=%d trabajo_id=%d estado=%q error: %v", empresaID, payload.TrabajoID, payload.Estado, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				item, _ := dbpkg.GetEmpresaImpresoraTrabajoByID(dbEmp, empresaID, payload.TrabajoID)
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "trabajo_cola_estado", "empresa_impresoras_cola", payload.TrabajoID, http.StatusOK, map[string]interface{}{
					"estado":    payload.Estado,
					"agente_id": payload.AgenteID,
					"con_error": strings.TrimSpace(payload.UltimoError) != "",
				}, "estado de trabajo de impresion actualizado")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "trabajo": item})
				return

			case "cola_reintentar", "trabajo_reintentar":
				trabajoID, err := parseInt64QueryOptional(r, "trabajo_id")
				if err != nil {
					http.Error(w, "trabajo_id invalido", http.StatusBadRequest)
					return
				}
				if trabajoID <= 0 {
					var payload struct {
						ID        int64 `json:"id"`
						TrabajoID int64 `json:"trabajo_id"`
					}
					_ = json.NewDecoder(r.Body).Decode(&payload)
					trabajoID = payload.TrabajoID
					if trabajoID <= 0 {
						trabajoID = payload.ID
					}
				}
				if err := dbpkg.ReintentarEmpresaImpresoraTrabajo(dbEmp, empresaID, trabajoID, strings.TrimSpace(adminEmailFromRequest(r))); err != nil {
					log.Printf("[empresa_impresoras] reintentar trabajo empresa_id=%d trabajo_id=%d error: %v", empresaID, trabajoID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				item, _ := dbpkg.GetEmpresaImpresoraTrabajoByID(dbEmp, empresaID, trabajoID)
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "trabajo_cola_reintentado", "empresa_impresoras_cola", trabajoID, http.StatusOK, nil, "trabajo de impresion marcado para reintento")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "trabajo": item})
				return

			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			switch action {
			case "", "impresora":
				impresoraID, err := parsePrinterIDQuery(r)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetEmpresaImpresoraEstado(dbEmp, empresaID, impresoraID, "inactivo", strings.TrimSpace(adminEmailFromRequest(r))); err != nil {
					log.Printf("[empresa_impresoras] delete impresora empresa_id=%d impresora_id=%d error: %v", empresaID, impresoraID, err)
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "impresora_desactivada", "empresa_impresoras", impresoraID, http.StatusOK, nil, "impresora desactivada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "impresora_id": impresoraID})
				return

			case "funcionalidad":
				funcionalidad := strings.TrimSpace(r.URL.Query().Get("funcionalidad"))
				if funcionalidad == "" {
					http.Error(w, "funcionalidad requerida", http.StatusBadRequest)
					return
				}
				if err := dbpkg.DeleteEmpresaImpresoraFuncionalidad(dbEmp, empresaID, funcionalidad); err != nil {
					log.Printf("[empresa_impresoras] delete funcionalidad empresa_id=%d funcionalidad=%q error: %v", empresaID, funcionalidad, err)
					http.Error(w, "No se pudo eliminar la asignación", http.StatusInternalServerError)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "funcionalidad_eliminada", "empresa_impresoras_funcionalidades", 0, http.StatusOK, map[string]interface{}{
					"funcionalidad": funcionalidad,
				}, "regla de impresion por funcionalidad eliminada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "funcionalidad": funcionalidad})
				return

			case "producto":
				productoID, err := parseInt64QueryOptional(r, "producto_id")
				if err != nil || productoID <= 0 {
					http.Error(w, "producto_id requerido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.DeleteEmpresaImpresoraProducto(dbEmp, empresaID, productoID); err != nil {
					log.Printf("[empresa_impresoras] delete producto empresa_id=%d producto_id=%d error: %v", empresaID, productoID, err)
					http.Error(w, "No se pudo eliminar la asignación", http.StatusInternalServerError)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "producto_asignacion_eliminada", "empresa_impresoras_productos", productoID, http.StatusOK, map[string]interface{}{
					"producto_id": productoID,
				}, "regla de impresion por producto eliminada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "producto_id": productoID})
				return
			case "producto_regla":
				alcance := strings.TrimSpace(r.URL.Query().Get("alcance"))
				categoriaID, err := parseInt64QueryOptional(r, "categoria_id")
				if err != nil {
					http.Error(w, "categoria_id invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.DeleteEmpresaImpresoraProductoRegla(dbEmp, empresaID, alcance, categoriaID); err != nil {
					log.Printf("[empresa_impresoras] delete producto regla empresa_id=%d alcance=%q categoria_id=%d error: %v", empresaID, alcance, categoriaID, err)
					http.Error(w, "No se pudo eliminar la regla", http.StatusInternalServerError)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "regla_producto_eliminada", "empresa_impresoras_productos_reglas", categoriaID, http.StatusOK, map[string]interface{}{
					"alcance":      alcance,
					"categoria_id": categoriaID,
				}, "regla masiva de impresion eliminada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "alcance": alcance, "categoria_id": categoriaID})
				return
			case "receta":
				recetaID, err := parseInt64QueryOptional(r, "receta_id")
				if err != nil || recetaID <= 0 {
					http.Error(w, "receta_id requerido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.DeleteEmpresaImpresoraReceta(dbEmp, empresaID, recetaID); err != nil {
					log.Printf("[empresa_impresoras] delete receta empresa_id=%d receta_id=%d error: %v", empresaID, recetaID, err)
					http.Error(w, "No se pudo eliminar la asignación", http.StatusInternalServerError)
					return
				}
				registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "impresoras", "receta_asignacion_eliminada", "empresa_impresoras_recetas", recetaID, http.StatusOK, map[string]interface{}{
					"receta_id": recetaID,
				}, "regla de impresion por receta eliminada")
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "receta_id": recetaID})
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

// EmpresaImpresorasResolverHandler expone resolución de impresora para flujos operativos (ventas/impresión).
func EmpresaImpresorasResolverHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if err := dbpkg.EnsureEmpresaImpresorasSchema(dbEmp); err != nil {
			log.Printf("[empresa_impresoras] ensure resolver schema error: %v", err)
			http.Error(w, "No se pudo preparar configuracion de impresoras", http.StatusInternalServerError)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		funcionalidad := strings.TrimSpace(r.URL.Query().Get("funcionalidad"))
		productoID, err := parseInt64QueryOptional(r, "producto_id")
		if err != nil {
			http.Error(w, "producto_id invalido", http.StatusBadRequest)
			return
		}
		recetaID, err := parseInt64QueryOptional(r, "receta_id")
		if err != nil {
			http.Error(w, "receta_id invalido", http.StatusBadRequest)
			return
		}
		tipoItem := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tipo_item")))
		referenciaID := productoID
		if tipoItem == "receta" || recetaID > 0 {
			tipoItem = "receta"
			referenciaID = recetaID
		} else if productoID > 0 {
			tipoItem = "producto"
		}
		resolved, err := dbpkg.ResolveEmpresaImpresoraOperacion(dbEmp, empresaID, funcionalidad, tipoItem, referenciaID)
		if err != nil {
			log.Printf("[empresa_impresoras] resolver publico empresa_id=%d funcionalidad=%q tipo_item=%q referencia_id=%d error: %v", empresaID, funcionalidad, tipoItem, referenciaID, err)
			http.Error(w, "No se pudo resolver impresora", http.StatusInternalServerError)
			return
		}
		if resolved == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            false,
				"empresa_id":    empresaID,
				"funcionalidad": funcionalidad,
				"producto_id":   productoID,
				"receta_id":     recetaID,
				"tipo_item":     tipoItem,
				"message":       "No hay impresora configurada",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"resolucion": resolved,
			"impresora":  resolved.Impresora,
			"fuente":     resolved.Fuente,
		})
	}
}

// EmpresaImpresorasAgenteHandler expone solo el contrato operativo del agente
// local de impresion. No permite crear ni editar impresoras.
func EmpresaImpresorasAgenteHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if err := dbpkg.EnsureEmpresaImpresorasSchema(dbEmp); err != nil {
			log.Printf("[empresa_impresoras_agente] ensure schema error: %v", err)
			http.Error(w, "No se pudo preparar cola de impresion", http.StatusInternalServerError)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "", "tomar", "agente_tomar":
			var payload struct {
				AgenteID   string `json:"agente_id"`
				EstacionID int64  `json:"estacion_id"`
				Limit      int64  `json:"limit"`
			}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			if strings.TrimSpace(payload.AgenteID) == "" {
				payload.AgenteID = strings.TrimSpace(r.URL.Query().Get("agente_id"))
			}
			if strings.TrimSpace(payload.AgenteID) == "" {
				http.Error(w, "agente_id requerido", http.StatusBadRequest)
				return
			}
			if payload.EstacionID <= 0 {
				if id, parseErr := parseInt64QueryOptional(r, "estacion_id"); parseErr == nil {
					payload.EstacionID = id
				}
			}
			if payload.Limit <= 0 {
				if limit, parseErr := parseIntQueryOptional(r, "limit"); parseErr == nil {
					payload.Limit = int64(limit)
				}
			}
			trabajos, err := dbpkg.TomarEmpresaImpresoraTrabajos(dbEmp, empresaID, payload.AgenteID, payload.EstacionID, payload.Limit)
			if err != nil {
				log.Printf("[empresa_impresoras_agente] tomar empresa_id=%d agente=%q error: %v", empresaID, payload.AgenteID, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "empresa_id": empresaID, "trabajos": trabajos})
			return

		case "estado", "cola_estado":
			var payload struct {
				ID          int64  `json:"id"`
				TrabajoID   int64  `json:"trabajo_id"`
				Estado      string `json:"estado"`
				AgenteID    string `json:"agente_id"`
				UltimoError string `json:"ultimo_error"`
				Error       string `json:"error"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.AgenteID) == "" {
				http.Error(w, "agente_id requerido", http.StatusBadRequest)
				return
			}
			if payload.TrabajoID <= 0 {
				payload.TrabajoID = payload.ID
			}
			if strings.TrimSpace(payload.UltimoError) == "" {
				payload.UltimoError = payload.Error
			}
			if err := dbpkg.ActualizarEmpresaImpresoraTrabajoEstado(dbEmp, empresaID, payload.TrabajoID, payload.Estado, payload.AgenteID, payload.UltimoError); err != nil {
				log.Printf("[empresa_impresoras_agente] estado empresa_id=%d trabajo_id=%d error: %v", empresaID, payload.TrabajoID, err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "trabajo_id": payload.TrabajoID})
			return

		default:
			http.Error(w, "action no soportada", http.StatusBadRequest)
			return
		}
	}
}
