package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaCentrosCostoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "dashboard":
				row, err := dbpkg.BuildEmpresaCentrosCostoDashboard(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")))
				if err != nil {
					http.Error(w, "No se pudo consultar centros de costo", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "centros":
				rows, err := dbpkg.ListEmpresaCentrosCosto(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("estado")), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar centros de costo", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "reglas":
				rows, err := dbpkg.ListEmpresaCentroCostoReglas(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("origen")), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar reglas de imputacion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "presupuestos":
				rows, err := dbpkg.ListEmpresaCentroCostoPresupuestos(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")), strings.TrimSpace(r.URL.Query().Get("escenario")), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar presupuestos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "movimientos":
				rows, err := dbpkg.ListEmpresaCentroCostoMovimientos(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar movimientos por centro de costo", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "centro":
				var payload dbpkg.EmpresaCentroCosto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaCentroCosto(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				status := http.StatusCreated
				if r.Method == http.MethodPut {
					status = http.StatusOK
				}
				writeJSON(w, status, map[string]interface{}{"ok": true, "id": id})
				return
			case "regla":
				var payload dbpkg.EmpresaCentroCostoRegla
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaCentroCostoRegla(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				status := http.StatusCreated
				if r.Method == http.MethodPut {
					status = http.StatusOK
				}
				writeJSON(w, status, map[string]interface{}{"ok": true, "id": id})
				return
			case "presupuesto":
				var payload dbpkg.EmpresaCentroCostoPresupuesto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaCentroCostoPresupuesto(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				status := http.StatusCreated
				if r.Method == http.MethodPut {
					status = http.StatusOK
				}
				writeJSON(w, status, map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaCentrosCostoDemo(dbEmp, empresaID, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, "accion o metodo no soportado", http.StatusMethodNotAllowed)
	}
}
