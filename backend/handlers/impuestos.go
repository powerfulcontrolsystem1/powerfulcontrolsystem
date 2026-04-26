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
	EmpresaID      int64                        `json:"empresa_id"`
	PaisCodigo     string                       `json:"pais_codigo"`
	PaisNombre     string                       `json:"pais_nombre"`
	Bandera        string                       `json:"bandera"`
	Moneda         string                       `json:"moneda"`
	CatalogoBase   []dbpkg.EmpresaImpuestoConfig `json:"catalogo_base"`
	Configurados   []dbpkg.EmpresaImpuestoConfig `json:"configurados"`
	GeneradoEn     string                       `json:"generado_en"`
}

func impuestosCatalogoBase(pais string) []dbpkg.EmpresaImpuestoConfig {
	pais = strings.ToUpper(strings.TrimSpace(pais))
	switch pais {
	case "EC":
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "EC", Codigo: "IVA", Nombre: "IVA (tarifa general)", Tipo: "impuesto", TasaPorcentaje: 15, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "ICE", Nombre: "ICE (consumos especiales)", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "RET_IVA", Nombre: "Retención IVA (según calificación SRI)", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "EC", Codigo: "RET_IR", Nombre: "Retención IR (según tabla SRI)", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
		}
	case "PA":
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "PA", Codigo: "ITBMS_7", Nombre: "ITBMS 7% (tasa general)", Tipo: "impuesto", TasaPorcentaje: 7, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ITBMS_10", Nombre: "ITBMS 10% (alcohol/hospedaje)", Tipo: "impuesto", TasaPorcentaje: 10, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ITBMS_15", Nombre: "ITBMS 15% (tabaco)", Tipo: "impuesto", TasaPorcentaje: 15, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ISC", Nombre: "ISC (selectivo al consumo)", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "RET_ITBMS", Nombre: "Retención ITBMS (gran comprador)", Tipo: "retencion", TasaPorcentaje: 50, Habilitado: 0, AplicaEn: "compras"},
		}
	default: // CO
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "CO", Codigo: "IVA", Nombre: "IVA (tarifa general)", Tipo: "impuesto", TasaPorcentaje: 19, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "CO", Codigo: "ICA", Nombre: "ICA (municipal, variable)", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CO", Codigo: "RETEFUENTE", Nombre: "Retención en la fuente (renta)", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "CO", Codigo: "RETEIVA", Nombre: "Retención a título de IVA (ReteIVA)", Tipo: "retencion", TasaPorcentaje: 15, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "CO", Codigo: "RETEICA", Nombre: "Retención a título de ICA (ReteICA)", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
		}
	}
}

// EmpresaImpuestosHandler gestiona catálogo y configuración de impuestos por empresa.
// GET  /api/empresa/impuestos?action=context
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
				// fallback seguro
				pais = dbpkg.PaisFacturacion{Codigo: "CO", Nombre: "Colombia", Bandera: "🇨🇴", Moneda: "COP"}
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
			http.Error(w, "action invalida (use context o upsert)", http.StatusBadRequest)
			return
		}
	}
}

