package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaDeclaracionesTributariasHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaDeclaracionesTributariasDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar declaraciones tributarias", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "declaraciones":
				rows, err := dbpkg.ListEmpresaDeclaracionesTributarias(dbEmp, empresaID, r.URL.Query().Get("tipo"), r.URL.Query().Get("estado"), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar declaraciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "calendario":
				rows, err := dbpkg.ListEmpresaCalendarioTributario(dbEmp, empresaID, intQuery(r, "anio"), 500)
				if err != nil {
					http.Error(w, "No se pudo listar calendario tributario", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "movimientos":
				rows, err := dbpkg.ListEmpresaDeclaracionesTributariasMovimientos(dbEmp, empresaID, r.URL.Query().Get("tipo"), r.URL.Query().Get("periodo"), 1000)
				if err != nil {
					http.Error(w, "No se pudieron listar movimientos tributarios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "declaracion":
				var payload dbpkg.EmpresaDeclaracionTributaria
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaDeclaracionTributaria(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "preliquidar":
				var payload struct {
					Tipo    string `json:"tipo_declaracion"`
					Periodo string `json:"periodo"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.Periodo == "" {
					payload.Periodo = strings.TrimSpace(r.URL.Query().Get("periodo"))
				}
				if payload.Periodo == "" {
					payload.Periodo = time.Now().Format("2006-01")
				}
				if payload.Tipo == "" {
					payload.Tipo = strings.TrimSpace(r.URL.Query().Get("tipo"))
				}
				row, err := dbpkg.PreliquidarEmpresaDeclaracionTributaria(dbEmp, empresaID, payload.Tipo, payload.Periodo, usuario)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "declaracion": row})
				return
			case "calendario":
				var payload dbpkg.EmpresaCalendarioTributario
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaCalendarioTributario(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaDeclaracionesTributariasDemo(dbEmp, empresaID, usuario); err != nil {
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
