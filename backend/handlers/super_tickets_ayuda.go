package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaAyudaTicketsHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := parseEmpresaIDFromContext(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			items, err := dbpkg.ListAyudaTickets(dbSuper, dbpkg.AyudaTicketFilter{
				EmpresaID: empresaID,
				Estado:    r.URL.Query().Get("estado"),
				Limit:     parseAyudaTicketLimit(r.URL.Query().Get("limit"), 80),
			})
			if err != nil {
				log.Printf("[tickets_ayuda] list empresa_id=%d error: %v", empresaID, err)
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudieron consultar los tickets"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "tickets": items})
		case http.MethodPost:
			var payload struct {
				Asunto    string `json:"asunto"`
				Categoria string `json:"categoria"`
				Prioridad string `json:"prioridad"`
				Mensaje   string `json:"mensaje"`
				Modulo    string `json:"modulo"`
				Ruta      string `json:"ruta"`
				Origen    string `json:"origen"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			empresaNombre := ""
			if empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID); err == nil && empresa != nil {
				empresaNombre = empresa.Nombre
			}
			solicitanteNombre := ""
			if admin, err := dbpkg.GetAdminByEmailFull(dbSuper, adminEmail); err == nil && admin != nil {
				solicitanteNombre = admin.Name
			}
			ticket, err := dbpkg.CreateAyudaTicket(dbSuper, dbpkg.AyudaTicketCreateRequest{
				EmpresaID:         empresaID,
				EmpresaNombre:     empresaNombre,
				SolicitanteNombre: solicitanteNombre,
				SolicitanteEmail:  adminEmail,
				Origen:            firstNonEmptyString(payload.Origen, "administrar_empresa"),
				Modulo:            payload.Modulo,
				Ruta:              payload.Ruta,
				Asunto:            payload.Asunto,
				Categoria:         payload.Categoria,
				Prioridad:         payload.Prioridad,
				Mensaje:           payload.Mensaje,
				UsuarioCreador:    adminEmail,
			})
			if err != nil {
				status := http.StatusInternalServerError
				message := "No se pudo crear el ticket"
				if strings.Contains(strings.ToLower(err.Error()), "obligatorio") {
					status = http.StatusBadRequest
					message = err.Error()
				}
				log.Printf("[tickets_ayuda] create empresa_id=%d error: %v", empresaID, err)
				writeJSON(w, status, map[string]interface{}{"ok": false, "error": message})
				return
			}
			writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "ticket": ticket})
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func SuperAyudaTicketsHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}

		switch r.Method {
		case http.MethodGet:
			id := parsePositiveInt64(strings.TrimSpace(r.URL.Query().Get("id")))
			if id > 0 {
				detail, err := dbpkg.GetAyudaTicketDetalle(dbSuper, id)
				if err != nil {
					if err == sql.ErrNoRows {
						writeJSON(w, http.StatusNotFound, map[string]interface{}{"ok": false, "error": "ticket no encontrado"})
						return
					}
					log.Printf("[tickets_ayuda] detail id=%d error: %v", id, err)
					writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo consultar el ticket"})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "ticket": detail.Ticket, "mensajes": detail.Mensajes})
				return
			}
			items, err := dbpkg.ListAyudaTickets(dbSuper, dbpkg.AyudaTicketFilter{
				Estado:    r.URL.Query().Get("estado"),
				Prioridad: r.URL.Query().Get("prioridad"),
				Query:     r.URL.Query().Get("q"),
				Limit:     parseAyudaTicketLimit(r.URL.Query().Get("limit"), 150),
			})
			if err != nil {
				log.Printf("[tickets_ayuda] super list error: %v", err)
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudieron consultar los tickets"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "tickets": items})
		case http.MethodPost, http.MethodPatch:
			var payload struct {
				TicketID  int64  `json:"ticket_id"`
				ID        int64  `json:"id"`
				Estado    string `json:"estado"`
				Prioridad string `json:"prioridad"`
				AsignadoA string `json:"asignado_a"`
				Mensaje   string `json:"mensaje"`
				Interno   bool   `json:"interno"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			ticketID := payload.TicketID
			if ticketID <= 0 {
				ticketID = payload.ID
			}
			if ticketID <= 0 {
				http.Error(w, "ticket_id es obligatorio", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Mensaje) != "" {
				interno := 0
				if payload.Interno {
					interno = 1
				}
				if err := dbpkg.AddAyudaTicketMensaje(dbSuper, ticketID, dbpkg.AyudaTicketMensaje{
					AutorTipo:      "super",
					AutorNombre:    "Super administrador",
					AutorEmail:     adminEmail,
					Mensaje:        payload.Mensaje,
					Interno:        interno,
					UsuarioCreador: adminEmail,
				}); err != nil {
					log.Printf("[tickets_ayuda] super message ticket=%d error: %v", ticketID, err)
					writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo guardar la respuesta"})
					return
				}
			}
			ticket, err := dbpkg.UpdateAyudaTicketEstado(dbSuper, ticketID, payload.Estado, payload.Prioridad, payload.AsignadoA, adminEmail)
			if err != nil {
				log.Printf("[tickets_ayuda] super update ticket=%d error: %v", ticketID, err)
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "No se pudo actualizar el ticket"})
				return
			}
			detail, _ := dbpkg.GetAyudaTicketDetalle(dbSuper, ticket.ID)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "ticket": ticket, "mensajes": detail.Mensajes})
		default:
			w.Header().Set("Allow", "GET, POST, PATCH")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func parseAyudaTicketLimit(raw string, fallback int) int {
	limit, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || limit <= 0 {
		return fallback
	}
	if limit > 300 {
		return 300
	}
	return limit
}
