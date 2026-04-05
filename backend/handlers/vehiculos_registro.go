package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaVehiculosRegistroHandler gestiona el modulo de registro de vehiculos por empresa.
func EmpresaVehiculosRegistroHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			includeInactive := queryBool(r, "include_inactive")
			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			estadoRegistro := strings.TrimSpace(r.URL.Query().Get("estado_registro"))
			patente := strings.TrimSpace(r.URL.Query().Get("patente"))
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			limit, err := parseIntQueryOptional(r, "limit")
			if err != nil {
				http.Error(w, "limit invalido", http.StatusBadRequest)
				return
			}

			rows, err := dbpkg.ListEmpresaVehiculosRegistros(dbEmp, empresaID, includeInactive, desde, hasta, estadoRegistro, patente, q, limit)
			if err != nil {
				http.Error(w, "No se pudo listar el registro de vehiculos", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, rows)
			return

		case http.MethodPost:
			var payload dbpkg.EmpresaVehiculoRegistro
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
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
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			id, err := dbpkg.CreateEmpresaVehiculoRegistro(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			switch action {
			case "activar", "desactivar":
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
				estado := "activo"
				if action == "desactivar" {
					estado = "inactivo"
				}
				if err := dbpkg.SetEmpresaVehiculoRegistroEstado(dbEmp, empresaID, id, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "registro de vehiculo no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo actualizar el estado del registro", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "estado": estado})
				return

			case "marcar_salida":
				var payload dbpkg.EmpresaVehiculoRegistro
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.ID <= 0 {
					if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
						payload.ID = id
					}
				}
				if payload.EmpresaID <= 0 || payload.ID <= 0 {
					http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.UsuarioSalida) == "" {
					payload.UsuarioSalida = strings.TrimSpace(adminEmailFromRequest(r))
				}
				if err := dbpkg.MarkEmpresaVehiculoSalida(dbEmp, payload.EmpresaID, payload.ID, payload.FechaSalida, payload.UsuarioSalida, payload.Observaciones); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "registro de vehiculo no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "accion": "marcar_salida"})
				return
			}

			var payload dbpkg.EmpresaVehiculoRegistro
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 || payload.ID <= 0 {
				http.Error(w, "empresa_id e id son obligatorios", http.StatusBadRequest)
				return
			}
			if err := dbpkg.UpdateEmpresaVehiculoRegistro(dbEmp, payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "registro de vehiculo no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
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
			if err := dbpkg.DeleteEmpresaVehiculoRegistro(dbEmp, empresaID, id); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "registro de vehiculo no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo eliminar el registro de vehiculo", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
