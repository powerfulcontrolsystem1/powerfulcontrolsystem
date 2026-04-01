package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	dbpkg "github.com/you/pos-backend/db"
)

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

// TiposDeUsuarioHandler maneja CRUD de tipos de usuario por tipo de empresa.
func TiposDeUsuarioHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tipoEmpresaID, err := parseOptionalInt64Query(r, "tipo_empresa_id")
			if err != nil {
				http.Error(w, "invalid tipo_empresa_id", http.StatusBadRequest)
				return
			}
			includeInactive := r.URL.Query().Get("include_inactive") == "1"
			items, err := dbpkg.GetTiposDeUsuario(dbSuper, tipoEmpresaID, includeInactive)
			if err != nil {
				http.Error(w, "failed to query tipos_de_usuario: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(items)
			return
		case http.MethodPost:
			var payload struct {
				TipoEmpresaID int64  `json:"tipo_empresa_id"`
				RolID         int64  `json:"rol_id"`
				Nombre        string `json:"nombre"`
				Descripcion   string `json:"descripcion"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.TipoEmpresaID <= 0 || payload.RolID <= 0 || payload.Nombre == "" {
				http.Error(w, "tipo_empresa_id, rol_id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			id, err := dbpkg.CreateTipoDeUsuario(dbSuper, payload.TipoEmpresaID, payload.RolID, payload.Nombre, payload.Descripcion, adminEmailFromRequest(r))
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "rol no valido para ese tipo de empresa", http.StatusBadRequest)
					return
				}
				http.Error(w, "failed to create tipo_de_usuario: "+err.Error(), http.StatusInternalServerError)
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
				if err := dbpkg.SetTipoDeUsuarioEstado(dbSuper, id, estado); err != nil {
					http.Error(w, "failed to set estado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var payload struct {
				TipoEmpresaID int64  `json:"tipo_empresa_id"`
				RolID         int64  `json:"rol_id"`
				Nombre        string `json:"nombre"`
				Descripcion   string `json:"descripcion"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}
			if payload.TipoEmpresaID <= 0 || payload.RolID <= 0 || payload.Nombre == "" {
				http.Error(w, "tipo_empresa_id, rol_id y nombre son obligatorios", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateTipoDeUsuario(dbSuper, id, payload.TipoEmpresaID, payload.RolID, payload.Nombre, payload.Descripcion); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "rol no valido para ese tipo de empresa", http.StatusBadRequest)
					return
				}
				http.Error(w, "failed to update tipo_de_usuario: "+err.Error(), http.StatusInternalServerError)
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
			if err := dbpkg.DeleteTipoDeUsuario(dbSuper, id); err != nil {
				http.Error(w, "failed to delete tipo_de_usuario: "+err.Error(), http.StatusInternalServerError)
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
