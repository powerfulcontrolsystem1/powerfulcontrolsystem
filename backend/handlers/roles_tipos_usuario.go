package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type rolPermisoModuloPayload struct {
	Modulo    string `json:"modulo"`
	Accion    string `json:"accion"`
	Permitido bool   `json:"permitido"`
}

type rolPermisoPaginaPayload struct {
	PaginaClave string `json:"pagina_clave"`
	Permitido   bool   `json:"permitido"`
}

type rolPermisosUpsertPayload struct {
	RolID          int64                     `json:"rol_id"`
	PermisosModulo []rolPermisoModuloPayload `json:"permisos_modulo"`
	PermisosPagina []rolPermisoPaginaPayload `json:"permisos_pagina"`
}

// RolesDeUsuarioHandler maneja CRUD de roles configurables por tipo de empresa.
func RolesDeUsuarioHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tipoEmpresaID, err := parseOptionalInt64Query(r, "tipo_empresa_id")
			if err != nil {
				http.Error(w, "invalid tipo_empresa_id", http.StatusBadRequest)
				return
			}
			includeInactive := r.URL.Query().Get("include_inactive") == "1"
			items, err := dbpkg.GetRolesDeUsuario(dbSuper, tipoEmpresaID, includeInactive)
			if err != nil {
				http.Error(w, "failed to query roles_de_usuario: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(items)
			return
		case http.MethodPost:
			var payload struct {
				TipoEmpresaID int64  `json:"tipo_empresa_id"`
				Nombre        string `json:"nombre"`
				Descripcion   string `json:"descripcion"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.TipoEmpresaID <= 0 || payload.Nombre == "" {
				http.Error(w, "tipo_empresa_id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateRolDeUsuario(dbSuper, payload.TipoEmpresaID, payload.Nombre, payload.Descripcion, adminEmailFromRequest(r))
			if err != nil {
				http.Error(w, "failed to create rol_de_usuario: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
			return
		case http.MethodPut:
			id, err := parseRequiredInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if r.URL.Query().Get("action") == "activar" {
				estado := parseEstadoFromQuery(r)
				if err := dbpkg.SetRolDeUsuarioEstado(dbSuper, id, estado); err != nil {
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var payload struct {
				TipoEmpresaID int64  `json:"tipo_empresa_id"`
				Nombre        string `json:"nombre"`
				Descripcion   string `json:"descripcion"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.TipoEmpresaID <= 0 || payload.Nombre == "" {
				http.Error(w, "tipo_empresa_id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateRolDeUsuario(dbSuper, id, payload.TipoEmpresaID, payload.Nombre, payload.Descripcion); err != nil {
				http.Error(w, "failed to update rol_de_usuario: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		case http.MethodDelete:
			id, err := parseRequiredInt64Query(r, "id")
			if err != nil {
				http.Error(w, "id required", http.StatusBadRequest)
				return
			}
			if err := dbpkg.DeleteRolDeUsuario(dbSuper, id); err != nil {
				http.Error(w, "failed to delete rol_de_usuario: "+err.Error(), http.StatusInternalServerError)
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

// RolesDeUsuarioPermisosHandler gestiona permisos dinamicos por modulo/accion y por pagina para un rol.
func RolesDeUsuarioPermisosHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := dbpkg.EnsureRolesPermisosSchema(dbSuper); err != nil {
			http.Error(w, "failed to ensure roles permisos schema: "+err.Error(), http.StatusInternalServerError)
			return
		}

		switch r.Method {
		case http.MethodGet:
			rolID, err := parseRequiredInt64Query(r, "rol_id")
			if err != nil || rolID <= 0 {
				http.Error(w, "rol_id required", http.StatusBadRequest)
				return
			}

			rol, err := dbpkg.GetRolDeUsuarioByID(dbSuper, rolID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "rol no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "failed to load rol: "+err.Error(), http.StatusInternalServerError)
				return
			}

			modulos := buildPermissionModuleMatrixForRole(rol.Nombre)
			moduleItems, err := dbpkg.ListRolPermisosModuloByRolID(dbSuper, rol.ID)
			if err != nil {
				http.Error(w, "failed to load modulo permisos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			moduleOverrides := make(map[string]bool, len(moduleItems))
			for _, item := range moduleItems {
				moduleOverrides[permissionModuleActionKey(item.Modulo, item.Accion)] = item.Permitido
			}
			for idx := range modulos {
				row := &modulos[idx]
				for _, action := range permissionActionsCatalogOrdered {
					if permitido, ok := moduleOverrides[permissionModuleActionKey(row.Modulo, action)]; ok {
						setPermissionActionOnModuleRow(row, action, permitido)
					}
				}
			}

			pageItems, err := dbpkg.ListRolPermisosPaginaByRolID(dbSuper, rol.ID)
			if err != nil {
				http.Error(w, "failed to load page permisos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			pageOverrides := make(map[string]bool, len(pageItems))
			for _, item := range pageItems {
				pageOverrides[strings.TrimSpace(item.PaginaClave)] = item.Permitido
			}
			paginas := buildPermissionPagesCatalogFromModuleRows(modulos, pageOverrides)

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"rol_id":            rol.ID,
				"rol_nombre":        rol.Nombre,
				"tipo_empresa_id":   rol.TipoEmpresaID,
				"acciones_catalogo": append([]string{}, permissionActionsCatalogOrdered...),
				"modulos_catalogo":  append([]string{}, permissionModulesCatalogOrdered...),
				"modulos":           modulos,
				"paginas":           paginas,
			})
			return

		case http.MethodPut:
			var payload rolPermisosUpsertPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			if payload.RolID <= 0 {
				if qID, err := parseOptionalInt64Query(r, "rol_id"); err == nil && qID > 0 {
					payload.RolID = qID
				}
			}
			if payload.RolID <= 0 {
				http.Error(w, "rol_id required", http.StatusBadRequest)
				return
			}

			if _, err := dbpkg.GetRolDeUsuarioByID(dbSuper, payload.RolID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "rol no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "failed to load rol: "+err.Error(), http.StatusInternalServerError)
				return
			}

			moduleRows := make([]dbpkg.RolPermisoModulo, 0, len(payload.PermisosModulo))
			for _, item := range payload.PermisosModulo {
				moduleRows = append(moduleRows, dbpkg.RolPermisoModulo{
					RolID:     payload.RolID,
					Modulo:    strings.ToLower(strings.TrimSpace(item.Modulo)),
					Accion:    strings.ToUpper(strings.TrimSpace(item.Accion)),
					Permitido: item.Permitido,
				})
			}

			pageRows := make([]dbpkg.RolPermisoPagina, 0, len(payload.PermisosPagina))
			for _, item := range payload.PermisosPagina {
				pageRows = append(pageRows, dbpkg.RolPermisoPagina{
					RolID:       payload.RolID,
					PaginaClave: strings.TrimSpace(item.PaginaClave),
					Permitido:   item.Permitido,
				})
			}

			if err := dbpkg.ReplaceRolPermisosDeUsuario(dbSuper, payload.RolID, moduleRows, pageRows, adminEmailFromRequest(r)); err != nil {
				http.Error(w, "failed to save permisos: "+err.Error(), http.StatusInternalServerError)
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

func parseRequiredInt64Query(r *http.Request, key string) (int64, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(val, 10, 64)
}

func parseOptionalInt64Query(r *http.Request, key string) (int64, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return 0, nil
	}
	return strconv.ParseInt(val, 10, 64)
}

func parseEstadoFromQuery(r *http.Request) string {
	estado := r.URL.Query().Get("estado")
	if estado != "" {
		return estado
	}
	if r.URL.Query().Get("activo") == "1" {
		return "activo"
	}
	return "inactivo"
}
