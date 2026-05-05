package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaApartamentosTuristicosHandler(dbEmp *sql.DB) http.HandlerFunc {
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
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaApartamentoTuristicoDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar el dashboard de apartamentos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaApartamentoTuristicoConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "apartamentos":
				rows, err := dbpkg.ListEmpresaApartamentosTuristicosUnidades(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar apartamentos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "tarifas":
				rows, err := dbpkg.ListEmpresaApartamentosTuristicosTarifas(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudieron listar tarifas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "reservas":
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				rows, err := dbpkg.ListEmpresaApartamentosTuristicosReservas(dbEmp, empresaID, estado, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar reservas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "tareas":
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				rows, err := dbpkg.ListEmpresaApartamentoTuristicoTareas(dbEmp, empresaID, estado, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar tareas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "config":
				var payload dbpkg.EmpresaApartamentoTuristicoConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				if err := dbpkg.UpsertEmpresaApartamentoTuristicoConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "apartamentos":
				var payload dbpkg.EmpresaApartamentoTuristicoUnidad
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaApartamentoTuristicoUnidad(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "tarifas":
				var payload dbpkg.EmpresaApartamentoTuristicoTarifa
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaApartamentoTuristicoTarifa(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "reservas":
				var payload dbpkg.EmpresaApartamentoTuristicoReserva
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaApartamentoTuristicoReserva(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "checkin", "checkout", "cancelar":
				var payload struct {
					ReservaID int64 `json:"reserva_id"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.ReservaID <= 0 {
					http.Error(w, "reserva_id es obligatorio", http.StatusBadRequest)
					return
				}
				estado := action
				if action == "cancelar" {
					estado = "cancelada"
				}
				if err := dbpkg.CambiarEstadoApartamentoTuristicoReserva(dbEmp, empresaID, payload.ReservaID, estado, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "estado_apartamento":
				var payload struct {
					ApartamentoID   int64  `json:"apartamento_id"`
					EstadoOperativo string `json:"estado_operativo"`
					EstadoOcupacion string `json:"estado_ocupacion"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.CambiarEstadoApartamentoTuristicoUnidad(dbEmp, empresaID, payload.ApartamentoID, payload.EstadoOperativo, payload.EstadoOcupacion, adminEmail); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "tareas":
				var payload dbpkg.EmpresaApartamentoTuristicoTarea
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaApartamentoTuristicoTarea(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "estado_tarea":
				var payload struct {
					TareaID   int64   `json:"tarea_id"`
					Estado    string  `json:"estado"`
					CostoReal float64 `json:"costo_real"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if err := dbpkg.CambiarEstadoApartamentoTuristicoTarea(dbEmp, empresaID, payload.TareaID, payload.Estado, adminEmail, payload.CostoReal); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
