package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaImportacionesCosteoHandler(dbEmp *sql.DB) http.HandlerFunc {
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

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaImportacionesCosteoDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar importaciones y costeo", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "importaciones":
				rows, err := dbpkg.ListEmpresaImportacionesCosteo(dbEmp, empresaID, r.URL.Query().Get("estado"), 300)
				if err != nil {
					http.Error(w, "No se pudieron listar importaciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "detalle":
				row, err := dbpkg.GetEmpresaImportacionCosteo(dbEmp, empresaID, int64Query(r, "id"))
				if err != nil {
					http.Error(w, "No se pudo consultar la importacion", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "importacion":
				var payload dbpkg.EmpresaImportacionCosteo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaImportacionCosteo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "item":
				var payload dbpkg.EmpresaImportacionItem
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				imp, err := dbpkg.GetEmpresaImportacionCosteo(dbEmp, empresaID, payload.ImportacionID)
				if err != nil {
					http.Error(w, "importacion no encontrada", http.StatusBadRequest)
					return
				}
				id, err := dbpkg.CreateEmpresaImportacionItem(dbEmp, payload, imp.TRM)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "costo":
				var payload dbpkg.EmpresaImportacionCosto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaImportacionCosto(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "distribuir":
				var payload struct {
					ImportacionID int64 `json:"importacion_id"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.ImportacionID <= 0 {
					payload.ImportacionID = int64Query(r, "importacion_id")
				}
				row, err := dbpkg.DistribuirEmpresaImportacionCostos(dbEmp, empresaID, payload.ImportacionID, usuario)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaImportacionesCosteoDemo(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, "Metodo o accion no permitida", http.StatusMethodNotAllowed)
	}
}
