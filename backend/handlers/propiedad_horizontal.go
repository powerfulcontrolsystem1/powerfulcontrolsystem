package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaPropiedadHorizontalHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		usuario := strings.TrimSpace(adminEmailFromRequest(r))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "dashboard":
				row, err := dbpkg.BuildEmpresaPropiedadHorizontalDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar propiedad horizontal", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaPropiedadHorizontalConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar configuracion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "unidades":
				rows, err := dbpkg.ListEmpresaPropiedadHorizontalUnidades(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar unidades", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "personas":
				rows, err := dbpkg.ListEmpresaPropiedadHorizontalPersonas(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar personas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "cargos":
				rows, err := dbpkg.ListEmpresaPropiedadHorizontalCargos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar cargos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "recaudos":
				rows, err := dbpkg.ListEmpresaPropiedadHorizontalRecaudos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar recaudos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "pqrs":
				rows, err := dbpkg.ListEmpresaPropiedadHorizontalPQRs(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar PQR", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "asambleas":
				rows, err := dbpkg.ListEmpresaPropiedadHorizontalAsambleas(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar asambleas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "config":
				var payload dbpkg.EmpresaPropiedadHorizontalConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, usuario
				if err := dbpkg.UpsertEmpresaPropiedadHorizontalConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "unidad":
				var payload dbpkg.EmpresaPropiedadHorizontalUnidad
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, usuario
				id, err := dbpkg.UpsertEmpresaPropiedadHorizontalUnidad(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, statusForUpsert(r), map[string]interface{}{"ok": true, "id": id})
				return
			case "persona":
				var payload dbpkg.EmpresaPropiedadHorizontalPersona
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, usuario
				id, err := dbpkg.UpsertEmpresaPropiedadHorizontalPersona(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, statusForUpsert(r), map[string]interface{}{"ok": true, "id": id})
				return
			case "cargo":
				var payload dbpkg.EmpresaPropiedadHorizontalCargo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, usuario
				id, err := dbpkg.CreateEmpresaPropiedadHorizontalCargo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "recaudo":
				var payload dbpkg.EmpresaPropiedadHorizontalRecaudo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, usuario
				id, err := dbpkg.CreateEmpresaPropiedadHorizontalRecaudo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "pqr":
				var payload dbpkg.EmpresaPropiedadHorizontalPQR
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, usuario
				id, err := dbpkg.UpsertEmpresaPropiedadHorizontalPQR(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, statusForUpsert(r), map[string]interface{}{"ok": true, "id": id})
				return
			case "asamblea":
				var payload dbpkg.EmpresaPropiedadHorizontalAsamblea
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, usuario
				id, err := dbpkg.UpsertEmpresaPropiedadHorizontalAsamblea(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, statusForUpsert(r), map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaPropiedadHorizontalDemo(dbEmp, empresaID, usuario); err != nil {
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
