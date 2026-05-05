package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaParqueaderoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
		publicBaseURL := publicBaseURLFromRequest(r)

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "dashboard":
				out, err := dbpkg.BuildEmpresaParqueaderoDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo cargar el dashboard de parqueadero", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, out)
				return
			case "config":
				cfg, err := dbpkg.GetEmpresaParqueaderoConfig(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo cargar la configuracion de parqueadero", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, cfg)
				return
			case "tickets":
				estado := strings.TrimSpace(r.URL.Query().Get("estado"))
				rows, err := dbpkg.ListEmpresaParqueaderoTickets(dbEmp, empresaID, estado, 300)
				if err != nil {
					http.Error(w, "No se pudieron listar los tickets", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "validar_salida":
				token := strings.TrimSpace(r.URL.Query().Get("token"))
				if token == "" {
					http.Error(w, "token es obligatorio", http.StatusBadRequest)
					return
				}
				ticket, err := dbpkg.GetEmpresaParqueaderoTicketByToken(dbEmp, empresaID, token)
				if err != nil {
					http.Error(w, "Ticket no encontrado", http.StatusNotFound)
					return
				}
				cobro, _, err := dbpkg.CalcularEmpresaParqueaderoCobro(dbEmp, empresaID, ticket.ID, time.Now())
				if err != nil {
					http.Error(w, "No se pudo calcular el cobro", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "ticket": ticket, "cobro": cobro})
				return
			}

		case http.MethodPost, http.MethodPut:
			switch action {
			case "config":
				var payload dbpkg.EmpresaParqueaderoConfig
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				if err := dbpkg.UpsertEmpresaParqueaderoConfig(dbEmp, payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				cfg, _ := dbpkg.GetEmpresaParqueaderoConfig(dbEmp, empresaID)
				writeJSON(w, http.StatusOK, cfg)
				return
			case "entrada", "emitir_ticket":
				var payload dbpkg.EmpresaParqueaderoTicket
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = adminEmail
				item, err := dbpkg.CreateEmpresaParqueaderoTicket(dbEmp, payload, publicBaseURL)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, item)
				return
			case "calcular":
				var payload struct {
					TicketID int64  `json:"ticket_id"`
					Token    string `json:"token"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.TicketID <= 0 && strings.TrimSpace(payload.Token) != "" {
					ticket, err := dbpkg.GetEmpresaParqueaderoTicketByToken(dbEmp, empresaID, payload.Token)
					if err != nil {
						http.Error(w, "Ticket no encontrado", http.StatusNotFound)
						return
					}
					payload.TicketID = ticket.ID
				}
				if payload.TicketID <= 0 {
					http.Error(w, "ticket_id es obligatorio", http.StatusBadRequest)
					return
				}
				cobro, ticket, err := dbpkg.CalcularEmpresaParqueaderoCobro(dbEmp, empresaID, payload.TicketID, time.Now())
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "ticket": ticket, "cobro": cobro})
				return
			case "cobrar_salida", "salida":
				var payload struct {
					TicketID   int64  `json:"ticket_id"`
					Token      string `json:"token"`
					MetodoPago string `json:"metodo_pago"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.TicketID <= 0 && strings.TrimSpace(payload.Token) != "" {
					ticket, err := dbpkg.GetEmpresaParqueaderoTicketByToken(dbEmp, empresaID, payload.Token)
					if err != nil {
						http.Error(w, "Ticket no encontrado", http.StatusNotFound)
						return
					}
					payload.TicketID = ticket.ID
				}
				if payload.TicketID <= 0 {
					http.Error(w, "ticket_id es obligatorio", http.StatusBadRequest)
					return
				}
				ticket, cobro, err := dbpkg.CerrarEmpresaParqueaderoTicket(dbEmp, empresaID, payload.TicketID, payload.MetodoPago, adminEmail)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "ticket": ticket, "cobro": cobro})
				return
			case "anular":
				var payload struct {
					TicketID int64  `json:"ticket_id"`
					Motivo   string `json:"motivo"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.TicketID <= 0 {
					http.Error(w, "ticket_id es obligatorio", http.StatusBadRequest)
					return
				}
				if err := dbpkg.AnularEmpresaParqueaderoTicket(dbEmp, empresaID, payload.TicketID, adminEmail, payload.Motivo); err != nil {
					http.Error(w, "No se pudo anular el ticket", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			}
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func PublicParqueaderoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if r.Method != http.MethodGet || action != "validar_salida" {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		token := strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			http.Error(w, "token es obligatorio", http.StatusBadRequest)
			return
		}
		ticket, err := dbpkg.GetEmpresaParqueaderoTicketByToken(dbEmp, empresaID, token)
		if err != nil {
			http.Error(w, "Ticket no encontrado", http.StatusNotFound)
			return
		}
		cobro, _, err := dbpkg.CalcularEmpresaParqueaderoCobro(dbEmp, empresaID, ticket.ID, time.Now())
		if err != nil {
			http.Error(w, "No se pudo calcular el cobro", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "ticket": ticket, "cobro": cobro})
	}
}

func publicBaseURLFromRequest(r *http.Request) string {
	if r == nil {
		return "https://powerfulcontrolsystem.com/"
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwarded != "" {
		scheme := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto"))
		if scheme == "" {
			scheme = "https"
		}
		return scheme + "://" + forwarded
	}
	host := strings.TrimSpace(r.Host)
	if host == "" {
		return "https://powerfulcontrolsystem.com/"
	}
	scheme := "https"
	if r.TLS == nil && (strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1")) {
		scheme = "http"
	}
	return scheme + "://" + host
}
