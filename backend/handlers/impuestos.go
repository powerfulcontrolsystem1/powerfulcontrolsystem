package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type impuestosContextResponse struct {
	EmpresaID    int64                         `json:"empresa_id"`
	PaisCodigo   string                        `json:"pais_codigo"`
	PaisNombre   string                        `json:"pais_nombre"`
	Bandera      string                        `json:"bandera"`
	Moneda       string                        `json:"moneda"`
	CatalogoBase []dbpkg.EmpresaImpuestoConfig `json:"catalogo_base"`
	Configurados []dbpkg.EmpresaImpuestoConfig `json:"configurados"`
	GeneradoEn   string                        `json:"generado_en"`
}

func impuestosCatalogoBase(pais string) []dbpkg.EmpresaImpuestoConfig {
	return dbpkg.EmpresaImpuestosCatalogoBase(pais)
}

// EmpresaImpuestosHandler gestiona catálogo, configuración y reportes fiscales por empresa.
// GET  /api/empresa/impuestos?action=context
// GET  /api/empresa/impuestos?action=dashboard
// POST /api/empresa/impuestos?action=upsert   body: EmpresaImpuestoConfig
func EmpresaImpuestosHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID := extractEmpresaIDForPermissions(r)
		if empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "context"
		}

		switch action {
		case "context":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			pais, _, err := dbpkg.DetectFacturacionPais(dbEmp, empresaID, "", "")
			if err != nil {
				pais = dbpkg.PaisFacturacion{Codigo: "CO", Nombre: "Colombia", Bandera: "CO", Moneda: "COP"}
			}
			config, err := dbpkg.ListEmpresaImpuestos(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "no se pudo cargar impuestos", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, impuestosContextResponse{
				EmpresaID:    empresaID,
				PaisCodigo:   pais.Codigo,
				PaisNombre:   pais.Nombre,
				Bandera:      pais.Bandera,
				Moneda:       pais.Moneda,
				CatalogoBase: impuestosCatalogoBase(pais.Codigo),
				Configurados: config,
				GeneradoEn:   time.Now().Format("2006-01-02 15:04:05"),
			})
			return
		case "dashboard":
			if r.Method != http.MethodGet {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			desde := strings.TrimSpace(r.URL.Query().Get("desde"))
			hasta := strings.TrimSpace(r.URL.Query().Get("hasta"))
			dashboard, err := dbpkg.EmpresaImpuestosDashboardData(dbEmp, empresaID, desde, hasta)
			if err != nil {
				http.Error(w, "no se pudo construir dashboard fiscal", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, dashboard)
			return
		case "upsert":
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var payload dbpkg.EmpresaImpuestoConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "json invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if payload.Estado == "" {
				payload.Estado = "activo"
			}
			id, err := dbpkg.UpsertEmpresaImpuesto(dbEmp, payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"id":         id,
				"empresa_id": empresaID,
			})
			return
		default:
			http.Error(w, "action invalida (use context, dashboard o upsert)", http.StatusBadRequest)
			return
		}
	}
}
