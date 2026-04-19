package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	db "github.com/you/pos-backend/db"
)

func EmpresaNewProveedoresHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			estado := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("estado")))
			list, err := db.GetEmpresaProveedores(dbEmp, empresaID, estado)
			if err != nil {
				log.Printf("Error GET EmpresaNewProveedores (empresa %d): %v", empresaID, err)
				http.Error(w, "Error al listar proveedores", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, list)
			return

		case http.MethodPost:
			var body db.EmpresaProveedor
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			body.EmpresaID = empresaID
			body.UsuarioCreador = adminEmailFromRequest(r)

			id, err := db.CreateEmpresaProveedor(dbEmp, body)
			if err != nil {
				log.Printf("Error POST EmpresaProveedor (empresa %d): %v", empresaID, err)
				http.Error(w, "Error al crear proveedor", http.StatusInternalServerError)
				return
			}
			body.ID = id
			body.Estado = "activo"
			writeJSON(w, http.StatusCreated, body)
			return

		case http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activar" || action == "desactivar" || action == "eliminar" {
				idStr := r.URL.Query().Get("id")
				id, err := strconv.ParseInt(idStr, 10, 64)
				if err != nil {
					http.Error(w, "invalid id", http.StatusBadRequest)
					return
				}

				nuevoEstado := action
				if action == "eliminar" {
					nuevoEstado = "eliminado"
				}

				if err := db.SetEstadoEmpresaProveedor(dbEmp, id, empresaID, nuevoEstado); err != nil {
					log.Printf("Error %s EmpresaProveedor ID %d (empresa %d): %v", action, id, empresaID, err)
					http.Error(w, "Error al cambiar estado", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]string{"message": "Estado modificado"})
				return
			}

			var body db.EmpresaProveedor
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			body.EmpresaID = empresaID

			if err := db.UpdateEmpresaProveedor(dbEmp, body); err != nil {
				log.Printf("Error PUT EmpresaProveedor ID %d (empresa %d): %v", body.ID, empresaID, err)
				http.Error(w, "Error al actualizar proveedor", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, body)
			return

		case http.MethodDelete:
			idStr := r.URL.Query().Get("id")
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			if err := db.SetEstadoEmpresaProveedor(dbEmp, id, empresaID, "inactivo"); err != nil {
				log.Printf("Error DELETE EmpresaProveedor ID %d (empresa %d): %v", id, empresaID, err)
				http.Error(w, "Error al desactivar proveedor", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"message": "Proveedor desactivado"})
			return

		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}
