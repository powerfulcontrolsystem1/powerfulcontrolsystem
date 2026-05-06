package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaComprasAvanzadasHandler(dbEmp *sql.DB) http.HandlerFunc {
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
			switch action {
			case "dashboard":
				row, err := dbpkg.BuildEmpresaComprasAvanzadasDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo cargar dashboard de compras avanzadas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
			case "requisiciones":
				rows, err := dbpkg.ListEmpresaCompraRequisiciones(dbEmp, empresaID, r.URL.Query().Get("estado"), 200)
				if err != nil {
					http.Error(w, "No se pudieron listar requisiciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			case "detalle":
				id, err := parseInt64QueryOptional(r, "id")
				if err != nil || id <= 0 {
					http.Error(w, "id invalido", http.StatusBadRequest)
					return
				}
				row, err := dbpkg.GetEmpresaCompraRequisicion(dbEmp, empresaID, id)
				if err != nil {
					http.Error(w, "No se encontro la requisicion", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
			case "cotizaciones":
				id, err := parseInt64QueryOptional(r, "requisicion_id")
				if err != nil || id <= 0 {
					http.Error(w, "requisicion_id invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaCompraCotizaciones(dbEmp, empresaID, id)
				if err != nil {
					http.Error(w, "No se pudieron listar cotizaciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			case "aprobaciones":
				id, err := parseInt64QueryOptional(r, "requisicion_id")
				if err != nil || id <= 0 {
					http.Error(w, "requisicion_id invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaCompraAprobaciones(dbEmp, empresaID, id)
				if err != nil {
					http.Error(w, "No se pudieron listar aprobaciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			case "recepciones":
				id, err := parseInt64QueryOptional(r, "requisicion_id")
				if err != nil || id <= 0 {
					http.Error(w, "requisicion_id invalido", http.StatusBadRequest)
					return
				}
				rows, err := dbpkg.ListEmpresaCompraRecepciones(dbEmp, empresaID, id)
				if err != nil {
					http.Error(w, "No se pudieron listar recepciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
			}
		case http.MethodPost:
			var payload struct {
				Action      string                         `json:"action"`
				EmpresaID   int64                          `json:"empresa_id"`
				Requisicion dbpkg.EmpresaCompraRequisicion `json:"requisicion"`
				Cotizacion  dbpkg.EmpresaCompraCotizacion  `json:"cotizacion"`
				Aprobacion  dbpkg.EmpresaCompraAprobacion  `json:"aprobacion"`
				Recepcion   dbpkg.EmpresaCompraRecepcion   `json:"recepcion"`
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
				id, err := dbpkg.SeedEmpresaComprasAvanzadasDemo(dbEmp, payload.EmpresaID, usuario)
				if err != nil {
					http.Error(w, "No se pudo crear demo de compras avanzadas", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "requisicion":
				payload.Requisicion.EmpresaID = payload.EmpresaID
				if payload.Requisicion.UsuarioCreador == "" {
					payload.Requisicion.UsuarioCreador = usuario
				}
				id, err := dbpkg.CreateEmpresaCompraRequisicion(dbEmp, payload.Requisicion)
				if err != nil {
					http.Error(w, "No se pudo guardar la requisicion", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "cotizacion":
				payload.Cotizacion.EmpresaID = payload.EmpresaID
				if payload.Cotizacion.UsuarioCreador == "" {
					payload.Cotizacion.UsuarioCreador = usuario
				}
				id, err := dbpkg.CreateEmpresaCompraCotizacion(dbEmp, payload.Cotizacion)
				if err != nil {
					http.Error(w, "No se pudo guardar la cotizacion", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "aprobar":
				payload.Aprobacion.EmpresaID = payload.EmpresaID
				if payload.Aprobacion.Aprobador == "" {
					payload.Aprobacion.Aprobador = usuario
				}
				id, err := dbpkg.ResolverEmpresaCompraAprobacion(dbEmp, payload.Aprobacion)
				if err != nil {
					http.Error(w, "No se pudo registrar la aprobacion", http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
			case "recepcion":
				payload.Recepcion.EmpresaID = payload.EmpresaID
				if payload.Recepcion.UsuarioCreador == "" {
					payload.Recepcion.UsuarioCreador = usuario
				}
				if payload.Recepcion.Responsable == "" {
					payload.Recepcion.Responsable = usuario
				}
				id, err := dbpkg.CreateEmpresaCompraRecepcion(dbEmp, payload.Recepcion)
				if err != nil {
					http.Error(w, "No se pudo guardar la recepcion", http.StatusBadRequest)
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
