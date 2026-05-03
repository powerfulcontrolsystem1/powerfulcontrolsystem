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
	pais = strings.ToUpper(strings.TrimSpace(pais))
	switch pais {
	case "EC":
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "EC", Codigo: "IVA", Nombre: "IVA tarifa general", Tipo: "impuesto", TasaPorcentaje: 15, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "IVA_0", Nombre: "IVA 0% / exento", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "ICE", Nombre: "ICE consumos especiales", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "EC", Codigo: "RET_IVA", Nombre: "Retencion IVA segun SRI", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "EC", Codigo: "RET_IR", Nombre: "Retencion IR segun tabla SRI", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
		}
	case "PA":
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "PA", Codigo: "ITBMS_7", Nombre: "ITBMS 7% tasa general", Tipo: "impuesto", TasaPorcentaje: 7, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ITBMS_10", Nombre: "ITBMS 10% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 10, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ITBMS_15", Nombre: "ITBMS 15% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 15, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "ISC", Nombre: "ISC selectivo al consumo", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "PA", Codigo: "RET_ITBMS", Nombre: "Retencion ITBMS segun condicion", Tipo: "retencion", TasaPorcentaje: 50, Habilitado: 0, AplicaEn: "compras"},
		}
	case "CR":
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "CR", Codigo: "IVA_13", Nombre: "IVA 13% tarifa general", Tipo: "impuesto", TasaPorcentaje: 13, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_4", Nombre: "IVA 4% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 4, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_2", Nombre: "IVA 2% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 2, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "IVA_1", Nombre: "IVA 1% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 1, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CR", Codigo: "EXENTO", Nombre: "Exento / no sujeto", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
		}
	case "AR":
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "AR", Codigo: "IVA_21", Nombre: "IVA 21% tarifa general", Tipo: "impuesto", TasaPorcentaje: 21, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "IVA_105", Nombre: "IVA 10.5% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 10.5, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "IVA_27", Nombre: "IVA 27% tarifa diferencial", Tipo: "impuesto", TasaPorcentaje: 27, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "EXENTO", Nombre: "Exento / no gravado", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "AR", Codigo: "RET_GAN", Nombre: "Retencion ganancias segun regimen", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "AR", Codigo: "IIBB", Nombre: "Ingresos brutos jurisdiccional", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
		}
	case "VE":
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "VE", Codigo: "IVA_16", Nombre: "IVA 16% tarifa general", Tipo: "impuesto", TasaPorcentaje: 16, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IVA_8", Nombre: "IVA 8% tarifa reducida", Tipo: "impuesto", TasaPorcentaje: 8, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IVA_31", Nombre: "IVA adicional 31% rubros especiales", Tipo: "impuesto", TasaPorcentaje: 31, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "EXENTO", Nombre: "Exento / no sujeto", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "VE", Codigo: "IGTF", Nombre: "IGTF segun medio de pago", Tipo: "impuesto", TasaPorcentaje: 3, Habilitado: 0, AplicaEn: "ventas"},
		}
	default:
		return []dbpkg.EmpresaImpuestoConfig{
			{PaisCodigo: "CO", Codigo: "IVA", Nombre: "IVA tarifa general", Tipo: "impuesto", TasaPorcentaje: 19, Habilitado: 1, AplicaEn: "ventas"},
			{PaisCodigo: "CO", Codigo: "ICA", Nombre: "ICA municipal variable", Tipo: "impuesto", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "ventas"},
			{PaisCodigo: "CO", Codigo: "RETEFUENTE", Nombre: "Retencion en la fuente renta", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "CO", Codigo: "RETEIVA", Nombre: "Retencion a titulo de IVA", Tipo: "retencion", TasaPorcentaje: 15, Habilitado: 0, AplicaEn: "compras"},
			{PaisCodigo: "CO", Codigo: "RETEICA", Nombre: "Retencion a titulo de ICA", Tipo: "retencion", TasaPorcentaje: 0, Habilitado: 0, AplicaEn: "compras"},
		}
	}
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
