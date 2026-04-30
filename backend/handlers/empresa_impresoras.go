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

			case "resolver":
				funcionalidad := strings.TrimSpace(r.URL.Query().Get("funcionalidad"))
				productoID, err := parseInt64QueryOptional(r, "producto_id")
				if err != nil {
					http.Error(w, "producto_id invalido", http.StatusBadRequest)
					return
				}
				resolved, err := dbpkg.ResolveEmpresaImpresora(dbEmp, empresaID, funcionalidad, productoID)
				if err != nil {
					log.Printf("[empresa_impresoras] resolver empresa_id=%d funcionalidad=%q producto_id=%d error: %v", empresaID, funcionalidad, productoID, err)
					http.Error(w, "No se pudo resolver impresora", http.StatusInternalServerError)
					return
				}
				if resolved == nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"ok":            false,
						"empresa_id":    empresaID,
						"funcionalidad": funcionalidad,
						"producto_id":   productoID,
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
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"impresoras":      impresoras,
					"funcionalidades": funcionalidades,
					"productos":       productos,
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
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
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
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "producto_id": productoID})
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
		resolved, err := dbpkg.ResolveEmpresaImpresora(dbEmp, empresaID, funcionalidad, productoID)
		if err != nil {
			log.Printf("[empresa_impresoras] resolver publico empresa_id=%d funcionalidad=%q producto_id=%d error: %v", empresaID, funcionalidad, productoID, err)
			http.Error(w, "No se pudo resolver impresora", http.StatusInternalServerError)
			return
		}
		if resolved == nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            false,
				"empresa_id":    empresaID,
				"funcionalidad": funcionalidad,
				"producto_id":   productoID,
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
