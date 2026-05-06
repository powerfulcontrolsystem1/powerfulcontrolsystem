package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaTesoreriaPresupuestoHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaTesoreriaDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar tesoreria", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaTesoreriaConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar configuracion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "cuentas":
				rows, err := dbpkg.ListEmpresaTesoreriaCuentas(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("estado")), 250)
				if err != nil {
					http.Error(w, "No se pudieron listar cuentas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "presupuestos":
				rows, err := dbpkg.ListEmpresaTesoreriaPresupuestos(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")), 250)
				if err != nil {
					http.Error(w, "No se pudieron listar presupuestos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "partidas":
				presupuestoID, _ := parseInt64QueryOptional(r, "presupuesto_id")
				rows, err := dbpkg.ListEmpresaTesoreriaPartidas(dbEmp, empresaID, presupuestoID, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar partidas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "flujo":
				rows, err := dbpkg.ListEmpresaTesoreriaFlujo(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")), 300)
				if err != nil {
					http.Error(w, "No se pudo listar flujo de caja", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost:
			switch action {
			case "config":
				var payload dbpkg.EmpresaTesoreriaConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				if err := dbpkg.UpsertEmpresaTesoreriaConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "cuentas":
				var payload dbpkg.EmpresaTesoreriaCuenta
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaTesoreriaCuenta(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "presupuestos":
				var payload dbpkg.EmpresaTesoreriaPresupuesto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaTesoreriaPresupuesto(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "partidas":
				var payload dbpkg.EmpresaTesoreriaPartida
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaTesoreriaPartida(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "flujo":
				var payload dbpkg.EmpresaTesoreriaFlujo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.CreateEmpresaTesoreriaFlujo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "generar_flujo":
				var payload struct {
					PresupuestoID int64 `json:"presupuesto_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.GenerarEmpresaTesoreriaFlujoDesdePresupuesto(dbEmp, empresaID, payload.PresupuestoID, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaTesoreriaDemo(dbEmp, empresaID, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		case http.MethodPut:
			switch action {
			case "cuentas":
				var payload dbpkg.EmpresaTesoreriaCuenta
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				id, err := dbpkg.UpsertEmpresaTesoreriaCuenta(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "presupuestos":
				var payload dbpkg.EmpresaTesoreriaPresupuesto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				id, err := dbpkg.UpsertEmpresaTesoreriaPresupuesto(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			case "partidas":
				var payload dbpkg.EmpresaTesoreriaPartida
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				id, err := dbpkg.UpsertEmpresaTesoreriaPartida(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			}
		}
		http.Error(w, "accion o metodo no soportado", http.StatusMethodNotAllowed)
	}
}
