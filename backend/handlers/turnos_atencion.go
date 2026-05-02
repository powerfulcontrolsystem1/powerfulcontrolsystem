package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaTurnosAtencionHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				fecha := strings.TrimSpace(r.URL.Query().Get("fecha"))
				row, err := dbpkg.BuildEmpresaTurnosAtencionDashboard(dbEmp, empresaID, fecha)
				if err != nil {
					http.Error(w, "No se pudo consultar el dashboard de turnos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "config":
				row, err := dbpkg.GetEmpresaTurnoAtencionConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar la configuracion", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "servicios":
				rows, err := dbpkg.ListEmpresaTurnosAtencionServicios(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo listar los servicios", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "puestos":
				rows, err := dbpkg.ListEmpresaTurnosAtencionPuestos(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo listar los puestos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "tickets":
				fecha := strings.TrimSpace(r.URL.Query().Get("fecha"))
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				rows, err := dbpkg.ListEmpresaTurnosAtencionTickets(dbEmp, empresaID, fecha, estado, 200)
				if err != nil {
					http.Error(w, "No se pudo listar los tickets", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			}

		case http.MethodPost:
			switch action {
			case "config":
				var payload dbpkg.EmpresaTurnoAtencionConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				if err := dbpkg.UpsertEmpresaTurnoAtencionConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "servicios":
				var payload dbpkg.EmpresaTurnoAtencionServicio
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaTurnoAtencionServicio(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "puestos":
				var payload dbpkg.EmpresaTurnoAtencionPuesto
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				id, err := dbpkg.CreateEmpresaTurnoAtencionPuesto(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "emitir_ticket":
				var payload dbpkg.EmpresaTurnoAtencionTicket
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				item, err := dbpkg.CreateEmpresaTurnoAtencionTicket(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, item)
				return
			case "llamar_siguiente":
				puestoID, _ := parseInt64QueryOptional(r, "puesto_id")
				if puestoID <= 0 {
					var payload struct {
						PuestoID int64 `json:"puesto_id"`
					}
					_ = json.NewDecoder(r.Body).Decode(&payload)
					puestoID = payload.PuestoID
				}
				if puestoID <= 0 {
					http.Error(w, "puesto_id es obligatorio", http.StatusBadRequest)
					return
				}
				item, err := dbpkg.LlamarSiguienteTurnoAtencion(dbEmp, empresaID, puestoID, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, item)
				return
			case "cambiar_estado":
				var payload struct {
					TicketID      int64  `json:"ticket_id"`
					PuestoID      int64  `json:"puesto_id"`
					Estado        string `json:"estado"`
					Observaciones string `json:"observaciones"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.TicketID <= 0 {
					http.Error(w, "ticket_id es obligatorio", http.StatusBadRequest)
					return
				}
				if err := dbpkg.CambiarEstadoTurnoAtencion(dbEmp, empresaID, payload.TicketID, payload.PuestoID, payload.Estado, adminEmail, payload.Observaciones); err != nil {
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

func PublicTurnosAtencionHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "servicios":
				rows, err := dbpkg.ListEmpresaTurnosAtencionServicios(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo listar los servicios", http.StatusInternalServerError)
					return
				}
				var activos []dbpkg.EmpresaTurnoAtencionServicio
				for _, row := range rows {
					if strings.ToLower(strings.TrimSpace(row.Estado)) == "activo" {
						activos = append(activos, row)
					}
				}
				writeJSON(w, http.StatusOK, activos)
				return
			case "display", "pantalla":
				fecha := strings.TrimSpace(r.URL.Query().Get("fecha"))
				row, err := dbpkg.BuildEmpresaTurnosAtencionDisplay(dbEmp, empresaID, fecha)
				if err != nil {
					http.Error(w, "No se pudo consultar la pantalla de turnos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			}
		case http.MethodPost:
			if action == "emitir_ticket" {
				cfg, err := dbpkg.GetEmpresaTurnoAtencionConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo validar la configuracion", http.StatusInternalServerError)
					return
				}
				if !cfg.PermitirEmisionPublica {
					http.Error(w, "La emision publica de turnos esta deshabilitada", http.StatusForbidden)
					return
				}
				var payload dbpkg.EmpresaTurnoAtencionTicket
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.CanalEmision = "publico"
				payload.UsuarioCreador = "portal_publico"
				item, err := dbpkg.CreateEmpresaTurnoAtencionTicket(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, item)
				return
			}
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}
