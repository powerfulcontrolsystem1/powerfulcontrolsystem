package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
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
	Revision       string                    `json:"revision"`
	PermisosModulo []rolPermisoModuloPayload `json:"permisos_modulo"`
	PermisosPagina []rolPermisoPaginaPayload `json:"permisos_pagina"`
}

// RolesDeUsuarioHandler maneja CRUD de roles configurables por tipo de empresa.
func RolesDeUsuarioHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if !requireRolesPermisosSchemaReady(w, dbSuper) {
			return
		}
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
			encodeJSONResponse(w, items)
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
			encodeJSONResponse(w, map[string]interface{}{"id": id})
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
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if !requireRolesPermisosSchemaReady(w, dbSuper) {
			return
		}
		switch r.Method {
		case http.MethodGet:
			rolID, err := parseRequiredInt64Query(r, "rol_id")
			if err != nil || rolID <= 0 || len(r.URL.Query()["rol_id"]) != 1 {
				http.Error(w, "rol_id required", http.StatusBadRequest)
				return
			}

			state, err := dbpkg.GetRolPermisosEstado(r.Context(), dbSuper, rolID)
			if err != nil {
				writeEmpresaRolPermissionError(w, err)
				return
			}
			rol := state.Rol
			modulos := buildRolPermissionEditorModuleRows(rol.Nombre, state.Modulos)
			pageOverrides := make(map[string]bool, len(state.Paginas))
			for _, item := range state.Paginas {
				pageOverrides[strings.TrimSpace(item.PaginaClave)] = item.Permitido
			}
			paginas := buildPermissionPagesCatalogFromModuleRows(modulos, pageOverrides)

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"revision":          state.Revision,
				"rol_id":            rol.ID,
				"rol_nombre":        rol.Nombre,
				"tipo_empresa_id":   rol.TipoEmpresaID,
				"acciones_catalogo": append([]string{}, permissionActionsCatalogOrdered...),
				"acciones_etiqueta": PermissionActionDisplayNameMap(),
				"modulos_catalogo":  append([]string{}, permissionModulesCatalogOrdered...),
				"modulos_etiqueta":  PermissionModuleDisplayNameMap(),
				"modulos":           modulos,
				"paginas":           paginas,
			})
			return

		case http.MethodPut:
			var payload rolPermisosUpsertPayload
			decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
			if err := decoder.Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			var extra interface{}
			qID, queryErr := parseOptionalInt64Query(r, "rol_id")
			if err := decoder.Decode(&extra); err != io.EOF || queryErr != nil || len(r.URL.Query()["rol_id"]) > 1 || (qID > 0 && payload.RolID > 0 && qID != payload.RolID) {
				http.Error(w, "payload o rol_id inconsistente", http.StatusBadRequest)
				return
			}
			if payload.RolID <= 0 && qID > 0 {
				payload.RolID = qID
			}
			if payload.RolID <= 0 {
				http.Error(w, "rol_id required", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Revision) == "" {
				writeEmpresaRolPermissionError(w, dbpkg.ErrRolPermisosRevisionRequired)
				return
			}
			moduleRows, pageRows, err := validateEmpresaRolPermissionPayload(payload.RolID, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := dbpkg.ReplaceRolPermisosDeUsuarioConRevision(r.Context(), dbSuper, payload.RolID, payload.Revision, moduleRows, pageRows, adminEmailFromRequest(r)); err != nil {
				writeEmpresaRolPermissionError(w, err)
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

// La matriz inicial del editor conserva las restricciones operativas; sólo una
// regla persistida explícita puede ampliar esos permisos.
func buildRolPermissionEditorModuleRows(role string, overrides []dbpkg.RolPermisoModulo) []permissionModuleMatrixRow {
	role = normalizePermissionRole(role)
	rows := restrictPermissionModuleRowsForOperationalRole(role, buildPermissionModuleMatrixForRole(role))
	byKey := make(map[string]bool, len(overrides))
	for _, item := range overrides {
		byKey[permissionModuleActionKey(item.Modulo, item.Accion)] = item.Permitido
	}
	for idx := range rows {
		for _, action := range permissionActionsCatalogOrdered {
			if allowed, ok := byKey[permissionModuleActionKey(rows[idx].Modulo, action)]; ok {
				setPermissionActionOnModuleRow(&rows[idx], action, allowed)
				if rows[idx].deniedActions == nil {
					rows[idx].deniedActions = map[string]bool{}
				}
				rows[idx].deniedActions[action] = !allowed
			}
		}
	}
	return rows
}

// EmpresaRolDeUsuarioPermisosHandler se invoca desde el wrapper de seguridad
// empresarial. El tenant validado es la autoridad y el ID sólo selecciona un rol propio.
func EmpresaRolDeUsuarioPermisosHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenant, ok := TenantContextFromRequest(r)
		if !ok || strings.TrimSpace(tenant.AdminEmail) == "" {
			http.Error(w, "sesion empresarial requerida", http.StatusUnauthorized)
			return
		}
		if !requireRolesPermisosSchemaReady(w, dbSuper) {
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodPut {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		rolID, err := parseRequiredInt64Query(r, "rol_id")
		if err != nil || rolID <= 0 || len(r.URL.Query()["rol_id"]) != 1 {
			http.Error(w, "rol_id invalido", http.StatusBadRequest)
			return
		}
		if r.Method == http.MethodGet {
			state, err := dbpkg.GetEmpresaRolPermisosEstado(r.Context(), dbSuper, tenant.EmpresaID, rolID)
			if err != nil {
				writeEmpresaRolPermissionError(w, err)
				return
			}
			inherited := buildRolPermissionEditorModuleRows(state.Base.Nombre, state.ModulosBase)
			combined := append(append([]dbpkg.RolPermisoModulo{}, state.ModulosBase...), state.Modulos...)
			modulos := buildRolPermissionEditorModuleRows(state.Base.Nombre, combined)
			inheritedPages, ownPages := map[string]bool{}, map[string]bool{}
			for _, item := range state.PaginasBase {
				inheritedPages[item.PaginaClave] = item.Permitido
				ownPages[item.PaginaClave] = item.Permitido
			}
			for _, item := range state.Paginas {
				ownPages[item.PaginaClave] = item.Permitido
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id": tenant.EmpresaID, "rol_id": state.Rol.ID, "rol_nombre": state.Rol.Nombre,
				"rol_base_id": state.Base.ID, "rol_base_nombre": state.Base.Nombre,
				"revision": state.Revision, "permisos_modulo": state.Modulos, "permisos_pagina": state.Paginas,
				"acciones_catalogo": append([]string{}, permissionActionsCatalogOrdered...),
				"acciones_etiqueta": PermissionActionDisplayNameMap(),
				"modulos_catalogo":  append([]string{}, permissionModulesCatalogOrdered...),
				"modulos_etiqueta":  PermissionModuleDisplayNameMap(),
				"modulos":           modulos, "paginas": buildPermissionPagesCatalogFromModuleRows(modulos, ownPages),
				"modulos_heredados": inherited, "paginas_heredadas": buildPermissionPagesCatalogFromModuleRows(inherited, inheritedPages),
			})
			return
		}
		var payload rolPermisosUpsertPayload
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		if err := decoder.Decode(&payload); err != nil {
			http.Error(w, "matriz de permisos invalida", http.StatusBadRequest)
			return
		}
		var extra interface{}
		if err := decoder.Decode(&extra); err != io.EOF || (payload.RolID != 0 && payload.RolID != rolID) {
			http.Error(w, "payload o rol_id inconsistente", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.Revision) == "" {
			writeEmpresaRolPermissionError(w, dbpkg.ErrRolPermisosRevisionRequired)
			return
		}
		modulos, paginas, err := validateEmpresaRolPermissionPayload(rolID, payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := dbpkg.ReplaceEmpresaRolPermisosDeUsuarioConRevision(r.Context(), dbSuper, tenant.EmpresaID, rolID, payload.Revision, modulos, paginas, tenant.AdminEmail); err != nil {
			writeEmpresaRolPermissionError(w, err)
			return
		}
		invalidateEmpresaPermissionCacheForEmpresa(tenant.EmpresaID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func requireRolesPermisosSchemaReady(w http.ResponseWriter, dbSuper *sql.DB) bool {
	if err := dbpkg.RolesPermisosSchemaReady(dbSuper); err != nil {
		http.Error(w, "el esquema migrado de roles y permisos no esta disponible", http.StatusInternalServerError)
		return false
	}
	return true
}

func writeEmpresaRolPermissionError(w http.ResponseWriter, err error) {
	if errors.Is(err, dbpkg.ErrRolPermisosRevisionRequired) {
		http.Error(w, "debe cargar la revision actual antes de guardar permisos", http.StatusPreconditionRequired)
		return
	}
	if errors.Is(err, dbpkg.ErrRolPermisosRevisionConflict) {
		http.Error(w, "los permisos propios o heredados cambiaron; recargue antes de guardar", http.StatusConflict)
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "rol no encontrado o no disponible", http.StatusNotFound)
		return
	}
	http.Error(w, "no se pudo procesar la matriz de permisos", http.StatusInternalServerError)
}

func validateEmpresaRolPermissionPayload(rolID int64, payload rolPermisosUpsertPayload) ([]dbpkg.RolPermisoModulo, []dbpkg.RolPermisoPagina, error) {
	if rolID <= 0 || payload.PermisosModulo == nil || payload.PermisosPagina == nil {
		return nil, nil, errors.New("debe enviar permisos_modulo y permisos_pagina como listas")
	}
	modules, actions, pages := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, key := range permissionModulesCatalogOrdered {
		modules[key] = true
	}
	for _, key := range permissionActionsCatalogOrdered {
		actions[key] = true
	}
	for _, rule := range permissionPagesCatalogOrdered {
		pages[rule.PaginaClave] = true
	}
	seen := map[string]bool{}
	modulos := make([]dbpkg.RolPermisoModulo, 0, len(payload.PermisosModulo))
	for _, item := range payload.PermisosModulo {
		modulo, accion := strings.ToLower(strings.TrimSpace(item.Modulo)), strings.ToUpper(strings.TrimSpace(item.Accion))
		key := "m:" + modulo + ":" + accion
		if !modules[modulo] || !actions[accion] || seen[key] {
			return nil, nil, errors.New("modulo o accion invalida o repetida")
		}
		seen[key] = true
		modulos = append(modulos, dbpkg.RolPermisoModulo{RolID: rolID, Modulo: modulo, Accion: accion, Permitido: item.Permitido})
	}
	paginas := make([]dbpkg.RolPermisoPagina, 0, len(payload.PermisosPagina))
	for _, item := range payload.PermisosPagina {
		pagina := strings.TrimSpace(item.PaginaClave)
		key := "p:" + pagina
		if !pages[pagina] || seen[key] {
			return nil, nil, errors.New("pagina invalida o repetida")
		}
		seen[key] = true
		paginas = append(paginas, dbpkg.RolPermisoPagina{RolID: rolID, PaginaClave: pagina, Permitido: item.Permitido})
	}
	return modulos, paginas, nil
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
