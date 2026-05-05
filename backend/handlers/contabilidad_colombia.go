package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaContabilidadColombiaHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaContabilidadDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar contabilidad Colombia", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaContabilidadConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar configuracion contable", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "cuentas":
				rows, err := dbpkg.ListEmpresaContabilidadCuentas(dbEmp, empresaID, r.URL.Query().Get("q"))
				if err != nil {
					http.Error(w, "No se pudo listar el PUC", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "terceros":
				rows, err := dbpkg.ListEmpresaContabilidadTerceros(dbEmp, empresaID, r.URL.Query().Get("q"))
				if err != nil {
					http.Error(w, "No se pudieron listar terceros", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "impuestos":
				rows, err := dbpkg.ListEmpresaContabilidadImpuestos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar impuestos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "comprobantes":
				rows, err := dbpkg.ListEmpresaContabilidadComprobantes(dbEmp, empresaID, r.URL.Query().Get("periodo"), r.URL.Query().Get("estado"), 250)
				if err != nil {
					http.Error(w, "No se pudieron listar comprobantes", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "comprobante":
				id, err := parseInt64Query(r, "id")
				if err != nil {
					http.Error(w, "id invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.GetEmpresaContabilidadComprobante(dbEmp, empresaID, id)
				if err != nil {
					http.Error(w, "No se pudo consultar el comprobante", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "periodos":
				rows, err := dbpkg.ListEmpresaContabilidadPeriodos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar periodos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "seed":
				if err := dbpkg.SeedEmpresaContabilidadColombiaBase(dbEmp, empresaID, usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "config":
				var payload dbpkg.EmpresaContabilidadConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				if err := dbpkg.UpsertEmpresaContabilidadConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "cuentas":
				var payload dbpkg.EmpresaContabilidadCuenta
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaContabilidadCuenta(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "terceros":
				var payload dbpkg.EmpresaContabilidadTercero
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaContabilidadTercero(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "impuestos":
				var payload dbpkg.EmpresaContabilidadImpuesto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaContabilidadImpuesto(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "comprobantes":
				var payload dbpkg.EmpresaContabilidadComprobante
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaContabilidadComprobante(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "anular_comprobante":
				var payload struct {
					ID int64 `json:"id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.CambiarEstadoEmpresaContabilidadComprobante(dbEmp, empresaID, payload.ID, "anulado", usuario); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "cerrar_periodo", "reabrir_periodo":
				var payload struct {
					Periodo       string `json:"periodo"`
					Observaciones string `json:"observaciones"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if action == "cerrar_periodo" {
					err = dbpkg.CerrarEmpresaContabilidadPeriodo(dbEmp, empresaID, payload.Periodo, usuario, payload.Observaciones)
				} else {
					err = dbpkg.ReabrirEmpresaContabilidadPeriodo(dbEmp, empresaID, payload.Periodo, usuario, payload.Observaciones)
				}
				if err != nil {
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
