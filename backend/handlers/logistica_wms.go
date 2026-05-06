package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaWMSHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaWMSDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar logistica WMS", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "ubicaciones":
				rows, err := dbpkg.ListEmpresaWMSUbicaciones(dbEmp, empresaID, r.URL.Query().Get("estado"), 1000)
				if err != nil {
					http.Error(w, "No se pudieron listar ubicaciones WMS", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "ordenes":
				rows, err := dbpkg.ListEmpresaWMSOrdenes(dbEmp, empresaID, r.URL.Query().Get("tipo"), r.URL.Query().Get("estado"), 1000)
				if err != nil {
					http.Error(w, "No se pudieron listar ordenes WMS", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "detalle":
				row, err := dbpkg.GetEmpresaWMSOrden(dbEmp, empresaID, int64Query(r, "id"))
				if err != nil {
					http.Error(w, "Orden WMS no encontrada", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "items":
				rows, err := dbpkg.ListEmpresaWMSItems(dbEmp, empresaID, int64Query(r, "orden_id"))
				if err != nil {
					http.Error(w, "No se pudieron listar items WMS", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "despachos":
				rows, err := dbpkg.ListEmpresaWMSDespachos(dbEmp, empresaID, int64Query(r, "orden_id"), 1000)
				if err != nil {
					http.Error(w, "No se pudieron listar despachos WMS", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "eventos":
				rows, err := dbpkg.ListEmpresaWMSEventos(dbEmp, empresaID, 500)
				if err != nil {
					http.Error(w, "No se pudo consultar bitacora WMS", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "ubicacion":
				var payload dbpkg.EmpresaWMSUbicacion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaWMSUbicacion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "orden":
				var payload dbpkg.EmpresaWMSOrden
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaWMSOrden(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "item":
				var payload dbpkg.EmpresaWMSItem
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				id, err := dbpkg.CreateEmpresaWMSItem(dbEmp, payload, usuario)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "avance_item":
				var payload struct {
					ID               int64   `json:"id"`
					CantidadPickeada float64 `json:"cantidad_pickeada"`
					CantidadEmpacada float64 `json:"cantidad_empacada"`
					Estado           string  `json:"estado"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.ID <= 0 {
					payload.ID = int64Query(r, "id")
				}
				if err := dbpkg.ActualizarEmpresaWMSItemAvance(dbEmp, empresaID, payload.ID, payload.CantidadPickeada, payload.CantidadEmpacada, payload.Estado, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "despacho":
				var payload dbpkg.EmpresaWMSDespacho
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaWMSDespacho(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaWMSDemo(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, fmt.Sprintf("Metodo o accion WMS no permitida: %s", action), http.StatusMethodNotAllowed)
	}
}
