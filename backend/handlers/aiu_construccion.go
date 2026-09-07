package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaAIUConstruccionHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaAIUDashboard(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo consultar AIU", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "contratos":
				rows, err := dbpkg.ListEmpresaAIUContratosFiltrados(dbEmp, empresaID, dbpkg.EmpresaAIUContratoFiltro{
					Estado: r.URL.Query().Get("estado"),
					Query:  r.URL.Query().Get("q"),
					Limit:  300,
				})
				if err != nil {
					http.Error(w, "No se pudieron listar contratos AIU", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "facturas":
				rows, err := dbpkg.ListEmpresaAIUFacturas(dbEmp, empresaID, int64Query(r, "contrato_id"), 300)
				if err != nil {
					http.Error(w, "No se pudieron listar facturas AIU", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "eventos":
				rows, err := dbpkg.ListEmpresaAIUEventos(dbEmp, empresaID, int64Query(r, "contrato_id"), 300)
				if err != nil {
					http.Error(w, "No se pudo listar historial AIU", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "detalle":
				row, err := dbpkg.GetEmpresaAIUContrato(dbEmp, empresaID, int64Query(r, "id"))
				if err != nil {
					http.Error(w, "Contrato AIU no encontrado", http.StatusNotFound)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "reporte":
				contratos, err := dbpkg.ListEmpresaAIUContratosFiltrados(dbEmp, empresaID, dbpkg.EmpresaAIUContratoFiltro{
					Estado: r.URL.Query().Get("estado"),
					Query:  r.URL.Query().Get("q"),
					Limit:  500,
				})
				if err != nil {
					http.Error(w, "No se pudo generar reporte AIU", http.StatusInternalServerError)
					return
				}
				facturas, err := dbpkg.ListEmpresaAIUFacturas(dbEmp, empresaID, 0, 500)
				if err != nil {
					http.Error(w, "No se pudo generar reporte AIU", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "contratos": contratos, "facturas": facturas})
				return
			}

		case http.MethodPost, http.MethodPut:
			switch action {
			case "calcular":
				var payload dbpkg.EmpresaAIUContrato
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload = dbpkg.NormalizeEmpresaAIUContrato(payload)
				if err := dbpkg.ValidateEmpresaAIUContrato(payload); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, dbpkg.CalculateEmpresaAIUContrato(payload))
				return
			case "contrato":
				var payload dbpkg.EmpresaAIUContrato
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.UpsertEmpresaAIUContrato(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				row, _ := dbpkg.GetEmpresaAIUContrato(dbEmp, empresaID, id)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id, "contrato": row})
				return
			case "item":
				var payload dbpkg.EmpresaAIUItem
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				id, err := dbpkg.CreateEmpresaAIUItem(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				row, _ := dbpkg.GetEmpresaAIUContrato(dbEmp, empresaID, payload.ContratoID)
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id, "contrato": row})
				return
			case "generar_factura":
				var payload struct {
					ContratoID      int64  `json:"contrato_id"`
					DocumentoCodigo string `json:"documento_codigo"`
					PeriodoContable string `json:"periodo_contable"`
					ClienteID       int64  `json:"cliente_id"`
					ClienteNombre   string `json:"cliente_nombre"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.ContratoID <= 0 {
					payload.ContratoID = int64Query(r, "contrato_id")
				}
				factura, err := dbpkg.RegistrarEmpresaAIUFactura(dbEmp, empresaID, payload.ContratoID, payload.DocumentoCodigo, payload.PeriodoContable, usuario)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				contrato, _ := dbpkg.GetEmpresaAIUContrato(dbEmp, empresaID, payload.ContratoID)
				entidadID := payload.ClienteID
				if entidadID <= 0 {
					entidadID = contrato.ClienteID
				}
				obs := factura.Observaciones
				if strings.TrimSpace(payload.ClienteNombre) != "" {
					obs = strings.TrimSpace(obs + " Cliente: " + payload.ClienteNombre + ".")
				}
				obs = strings.TrimSpace(fmt.Sprintf("%s Neto a cobrar: %.2f.", obs, factura.NetoCobrar))
				doc, err := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, dbpkg.EmpresaDocumentoFacturacion{
					EmpresaID:            empresaID,
					TipoDocumento:        "factura_electronica",
					DocumentoCodigo:      factura.DocumentoCodigo,
					EstadoDocumento:      "emitida",
					EstadoAnterior:       "borrador",
					EventoUltimo:         "factura_aiu_emitida",
					PeriodoContable:      factura.PeriodoContable,
					MontoTotal:           factura.TotalFactura,
					Moneda:               "COP",
					PaisCodigo:           "CO",
					EntidadRelacionadaID: entidadID,
					UsuarioCreador:       usuario,
					Observaciones:        obs,
				})
				if err != nil {
					http.Error(w, "Factura AIU registrada, pero no se pudo crear documento electronico", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "factura_aiu": factura, "documento_facturacion": doc})
				return
			case "estado":
				var payload struct {
					ContratoID  int64  `json:"contrato_id"`
					Estado      string `json:"estado"`
					Observacion string `json:"observacion"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if payload.ContratoID <= 0 {
					payload.ContratoID = int64Query(r, "contrato_id")
				}
				if strings.TrimSpace(payload.Estado) == "" {
					payload.Estado = r.URL.Query().Get("estado")
				}
				row, err := dbpkg.UpdateEmpresaAIUContratoEstado(dbEmp, empresaID, payload.ContratoID, payload.Estado, usuario, payload.Observacion)
				if err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "contrato": row})
				return
			case "seed_demo":
				if err := dbpkg.SeedEmpresaAIUDemo(dbEmp, empresaID, usuario); err != nil {
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
