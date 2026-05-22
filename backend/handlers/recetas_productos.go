package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaRecetasProductosHandler gestiona CRUD de recetas de productos por empresa.
func EmpresaRecetasProductosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if id, _ := parseInt64QueryOptional(r, "id"); id > 0 {
				receta, err := dbpkg.GetRecetaProductoByID(dbEmp, empresaID, id)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "receta no encontrada", http.StatusNotFound)
						return
					}
					log.Printf("[recetas] get empresa_id=%d id=%d error: %v", empresaID, id, err)
					http.Error(w, "No se pudo consultar la receta", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, receta)
				return
			}

			q := strings.TrimSpace(r.URL.Query().Get("q"))
			estado := strings.TrimSpace(r.URL.Query().Get("estado"))
			includeInactive := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "1") ||
				strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("include_inactive")), "true")
			limit, _ := parseIntQueryOptional(r, "limit")
			offset, _ := parseIntQueryOptional(r, "offset")

			rows, err := dbpkg.GetRecetasProductosByEmpresa(dbEmp, empresaID, q, estado, includeInactive, limit, offset)
			if err != nil {
				log.Printf("[recetas] list empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudieron listar las recetas", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload struct {
				dbpkg.RecetaProducto
				Ingredientes []dbpkg.RecetaProductoDetalle `json:"ingredientes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.CreateRecetaProducto(dbEmp, payload.RecetaProducto, payload.Ingredientes)
			if err != nil {
				status := recetaWriteStatus(err)
				log.Printf("[recetas] create empresa_id=%d nombre=%q error: %v", payload.EmpresaID, payload.Nombre, err)
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" {
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
				estado := "inactivo"
				if action == "activar" {
					estado = "activo"
				}
				if err := dbpkg.SetRecetaProductoEstado(dbEmp, empresaID, id, estado); err != nil {
					status := recetaWriteStatus(err)
					log.Printf("[recetas] set estado empresa_id=%d id=%d estado=%s error: %v", empresaID, id, estado, err)
					http.Error(w, err.Error(), status)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return
			}

			var payload struct {
				dbpkg.RecetaProducto
				Ingredientes []dbpkg.RecetaProductoDetalle `json:"ingredientes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.ID <= 0 {
				http.Error(w, "id es obligatorio", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			if err := dbpkg.UpdateRecetaProducto(dbEmp, payload.RecetaProducto, payload.Ingredientes); err != nil {
				status := recetaWriteStatus(err)
				log.Printf("[recetas] update empresa_id=%d id=%d error: %v", payload.EmpresaID, payload.ID, err)
				http.Error(w, err.Error(), status)
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
			if err := dbpkg.DeleteRecetaProducto(dbEmp, empresaID, id); err != nil {
				status := recetaWriteStatus(err)
				log.Printf("[recetas] delete empresa_id=%d id=%d error: %v", empresaID, id, err)
				http.Error(w, err.Error(), status)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func recetaWriteStatus(err error) int {
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
		strings.Contains(lower, "no existe") ||
		strings.Contains(lower, "inactivo") {
		return http.StatusBadRequest
	}
	if strings.Contains(lower, "no se puede") || strings.Contains(lower, "duplic") || strings.Contains(lower, "conflict") {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}
