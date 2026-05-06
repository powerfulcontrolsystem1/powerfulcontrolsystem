package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaCierreFiscalHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaCierreFiscalDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar cierre fiscal", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "politicas":
				rows, err := dbpkg.ListEmpresaCierreFiscalPoliticas(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar politicas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "periodos":
				rows, err := dbpkg.ListEmpresaCierreFiscalPeriodos(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("estado")), 180)
				if err != nil {
					http.Error(w, "No se pudieron listar periodos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "excepciones":
				rows, err := dbpkg.ListEmpresaCierreFiscalExcepciones(dbEmp, empresaID, strings.TrimSpace(r.URL.Query().Get("periodo")), strings.TrimSpace(r.URL.Query().Get("estado")), 180)
				if err != nil {
					http.Error(w, "No se pudieron listar excepciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "eventos":
				rows, err := dbpkg.ListEmpresaCierreFiscalEventos(dbEmp, empresaID, 180)
				if err != nil {
					http.Error(w, "No se pudieron listar eventos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "validar":
				documentoID, _ := parseInt64QueryOptional(r, "documento_id")
				row, err := dbpkg.ValidarEmpresaCierreFiscalOperacion(
					dbEmp,
					empresaID,
					strings.TrimSpace(r.URL.Query().Get("fecha_operacion")),
					strings.TrimSpace(r.URL.Query().Get("modulo")),
					strings.TrimSpace(r.URL.Query().Get("accion")),
					strings.TrimSpace(r.URL.Query().Get("documento_tipo")),
					documentoID,
					adminEmail,
					true,
				)
				if err != nil {
					http.Error(w, "No se pudo validar cierre fiscal", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "politica":
				var payload dbpkg.EmpresaCierreFiscalPolitica
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaCierreFiscalPolitica(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, statusForUpsert(r), map[string]interface{}{"ok": true, "id": id})
				return
			case "periodo":
				var payload dbpkg.EmpresaCierreFiscalPeriodo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				id, err := dbpkg.UpsertEmpresaCierreFiscalPeriodo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, statusForUpsert(r), map[string]interface{}{"ok": true, "id": id})
				return
			case "estado_periodo":
				var payload struct {
					ID     int64  `json:"id"`
					Estado string `json:"estado"`
					Motivo string `json:"motivo"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.CambiarEstadoEmpresaCierreFiscalPeriodo(dbEmp, empresaID, payload.ID, payload.Estado, adminEmail, payload.Motivo)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "excepcion":
				var payload dbpkg.EmpresaCierreFiscalExcepcion
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID, payload.UsuarioCreador = empresaID, adminEmail
				if payload.AprobadoPor == "" {
					payload.AprobadoPor = adminEmail
				}
				id, err := dbpkg.CreateEmpresaCierreFiscalExcepcion(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaCierreFiscalDemo(dbEmp, empresaID, adminEmail); err != nil {
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

func statusForUpsert(r *http.Request) int {
	if r != nil && strings.EqualFold(strings.TrimSpace(r.Method), http.MethodPut) {
		return http.StatusOK
	}
	return http.StatusCreated
}
