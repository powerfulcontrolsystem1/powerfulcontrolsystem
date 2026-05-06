package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaCRMVentasAvanzadasHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "" {
				action = "dashboard"
			}
			periodo := strings.TrimSpace(r.URL.Query().Get("periodo"))
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaCRMVentasAvanzadasDashboard(dbEmp, empresaID, periodo)
				if err != nil {
					http.Error(w, "No se pudo cargar CRM avanzado", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
			case "metas":
				rows, err := dbpkg.ListEmpresaCRMMetasComerciales(dbEmp, empresaID, periodo)
				if err != nil {
					http.Error(w, "No se pudieron listar metas comerciales", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			case "scores":
				rows, err := dbpkg.ListEmpresaCRMLeadScores(dbEmp, empresaID, 30)
				if err != nil {
					http.Error(w, "No se pudo calcular scoring", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
			}
		case http.MethodPost:
			var payload struct {
				Action    string                        `json:"action"`
				EmpresaID int64                         `json:"empresa_id"`
				Meta      dbpkg.EmpresaCRMMetaComercial `json:"meta"`
				LeadID    int64                         `json:"lead_id"`
				Codigo    string                        `json:"codigo"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			if payload.EmpresaID <= 0 {
				payload.EmpresaID, _ = parseInt64QueryOptional(r, "empresa_id")
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			action := strings.ToLower(strings.TrimSpace(payload.Action))
			if action == "" {
				action = strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			}
			usuario := strings.TrimSpace(adminEmailFromRequest(r))
			switch action {
			case "seed_demo":
				id, err := dbpkg.SeedEmpresaCRMVentasAvanzadasDemo(dbEmp, payload.EmpresaID, usuario)
				if err != nil {
					http.Error(w, "No se pudo crear demo de CRM avanzado: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "meta":
				payload.Meta.EmpresaID = payload.EmpresaID
				if payload.Meta.UsuarioCreador == "" {
					payload.Meta.UsuarioCreador = usuario
				}
				id, err := dbpkg.UpsertEmpresaCRMMetaComercial(dbEmp, payload.Meta)
				if err != nil {
					http.Error(w, "No se pudo guardar meta comercial", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "cotizacion_desde_lead":
				if payload.LeadID <= 0 {
					payload.LeadID, _ = parseInt64QueryOptional(r, "lead_id")
				}
				if payload.LeadID <= 0 {
					http.Error(w, "lead_id es obligatorio", http.StatusBadRequest)
					return
				}
				id, err := dbpkg.CreateEmpresaCRMCotizacionDesdeLead(dbEmp, payload.EmpresaID, payload.LeadID, payload.Codigo, usuario)
				if err != nil {
					http.Error(w, "No se pudo convertir lead a cotizacion: "+err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}
