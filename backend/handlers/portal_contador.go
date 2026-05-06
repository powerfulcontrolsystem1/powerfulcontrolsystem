package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaPortalContadorHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "dashboard"
		}
		usuario := strings.TrimSpace(adminEmailFromRequest(r))
		if usuario == "" {
			usuario = "sistema"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaPortalContadorDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar portal contador", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "clientes":
				rows, err := dbpkg.ListEmpresaPortalContadorClientes(dbEmp, empresaID, r.URL.Query().Get("q"), 300)
				if err != nil {
					http.Error(w, "No se pudieron listar clientes contables", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "obligaciones":
				rows, err := dbpkg.ListEmpresaPortalContadorObligaciones(dbEmp, empresaID, r.URL.Query().Get("estado"), 300)
				if err != nil {
					http.Error(w, "No se pudieron listar obligaciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "solicitudes":
				rows, err := dbpkg.ListEmpresaPortalContadorSolicitudes(dbEmp, empresaID, r.URL.Query().Get("estado"), 300)
				if err != nil {
					http.Error(w, "No se pudieron listar solicitudes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "comunicaciones":
				rows, err := dbpkg.ListEmpresaPortalContadorComunicaciones(dbEmp, empresaID, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar comunicaciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "cliente":
				var payload dbpkg.EmpresaPortalContadorCliente
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaPortalContadorCliente(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "obligacion":
				var payload dbpkg.EmpresaPortalContadorObligacion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaPortalContadorObligacion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "solicitud":
				var payload dbpkg.EmpresaPortalContadorSolicitud
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.UpsertEmpresaPortalContadorSolicitud(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "comunicacion":
				var payload dbpkg.EmpresaPortalContadorComunicacion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.Usuario = usuario
				id, err := dbpkg.CreateEmpresaPortalContadorComunicacion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaPortalContadorDemo(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, fmt.Sprintf("Metodo o accion no permitida: %s", action), http.StatusMethodNotAllowed)
	}
}
