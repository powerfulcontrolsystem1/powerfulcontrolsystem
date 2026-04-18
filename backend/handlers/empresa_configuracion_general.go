package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaConfiguracionGeneralProductosPayload struct {
	ImprimirOrdenServicio        bool   `json:"imprimir_orden_servicio"`
	AreaDespacho                 string `json:"area_despacho"`
	CopiasOrdenServicio          int64  `json:"copias_orden_servicio"`
	NotaOrdenServicio            string `json:"nota_orden_servicio"`
	DescuentosHabilitados        bool   `json:"descuentos_habilitados"`
	PermitirDescuentoPorcentaje  bool   `json:"permitir_descuento_porcentaje"`
	PermitirDescuentoCodigo      bool   `json:"permitir_descuento_codigo"`
	PermitirDescuentoValor       bool   `json:"permitir_descuento_valor"`
	CodigosDescuento             string `json:"codigos_descuento"`
	LectorCodigoBarrasHabilitado bool   `json:"lector_codigo_barras_habilitado"`
	LectorCodigoBarrasAutofoco   bool   `json:"lector_codigo_barras_autofoco"`
	LectorCodigoBarrasAcumular   bool   `json:"lector_codigo_barras_acumular"`
}

type empresaConfiguracionGeneralPayload struct {
	EmpresaID     int64                                    `json:"empresa_id"`
	Productos     empresaConfiguracionGeneralProductosPayload `json:"productos"`
	Estado        string                                   `json:"estado,omitempty"`
	Observaciones string                                   `json:"observaciones,omitempty"`
}

func EmpresaConfiguracionGeneralHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetEmpresaConfiguracionGeneral(dbEmp, empresaID)
			if err != nil {
				log.Printf("[empresa_config_general] get empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo cargar la configuracion general", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, empresaConfiguracionGeneralResponse(cfg))
			return

		case http.MethodPost, http.MethodPut:
			body, err := io.ReadAll(r.Body)
			if err != nil || len(body) == 0 {
				http.Error(w, "JSON invalido o cuerpo vacio", http.StatusBadRequest)
				return
			}

			var payload empresaConfiguracionGeneralPayload
			if err := json.Unmarshal(body, &payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}

			if payload.EmpresaID <= 0 {
				if empresaID, qErr := parseInt64QueryOptional(r, "empresa_id"); qErr == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}

			cfg := dbpkg.EmpresaConfiguracionGeneral{
				EmpresaID:                    payload.EmpresaID,
				ImprimirOrdenServicio:        payload.Productos.ImprimirOrdenServicio,
				AreaDespacho:                 payload.Productos.AreaDespacho,
				CopiasOrdenServicio:          payload.Productos.CopiasOrdenServicio,
				NotaOrdenServicio:            payload.Productos.NotaOrdenServicio,
				DescuentosHabilitados:        payload.Productos.DescuentosHabilitados,
				PermitirDescuentoPorcentaje:  payload.Productos.PermitirDescuentoPorcentaje,
				PermitirDescuentoCodigo:      payload.Productos.PermitirDescuentoCodigo,
				PermitirDescuentoValor:       payload.Productos.PermitirDescuentoValor,
				CodigosDescuento:             payload.Productos.CodigosDescuento,
				LectorCodigoBarrasHabilitado: payload.Productos.LectorCodigoBarrasHabilitado,
				LectorCodigoBarrasAutofoco:   payload.Productos.LectorCodigoBarrasAutofoco,
				LectorCodigoBarrasAcumular:   payload.Productos.LectorCodigoBarrasAcumular,
				Estado:                       strings.TrimSpace(payload.Estado),
				Observaciones:                strings.TrimSpace(payload.Observaciones),
				UsuarioCreador:               strings.TrimSpace(adminEmailFromRequest(r)),
			}

			id, err := dbpkg.UpsertEmpresaConfiguracionGeneral(dbEmp, cfg)
			if err != nil {
				log.Printf("[empresa_config_general] upsert empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, "No se pudo guardar la configuracion general", http.StatusInternalServerError)
				return
			}

			stored, err := dbpkg.GetEmpresaConfiguracionGeneral(dbEmp, payload.EmpresaID)
			if err != nil {
				log.Printf("[empresa_config_general] get after upsert empresa_id=%d error: %v", payload.EmpresaID, err)
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"configuracion": empresaConfiguracionGeneralResponse(stored),
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func empresaConfiguracionGeneralResponse(cfg *dbpkg.EmpresaConfiguracionGeneral) map[string]interface{} {
	if cfg == nil {
		return map[string]interface{}{
			"empresa_id": 0,
			"productos":  map[string]interface{}{},
		}
	}
	return map[string]interface{}{
		"id":            cfg.ID,
		"empresa_id":    cfg.EmpresaID,
		"estado":        cfg.Estado,
		"observaciones": cfg.Observaciones,
		"productos": map[string]interface{}{
			"imprimir_orden_servicio":         cfg.ImprimirOrdenServicio,
			"area_despacho":                   cfg.AreaDespacho,
			"copias_orden_servicio":           cfg.CopiasOrdenServicio,
			"nota_orden_servicio":             cfg.NotaOrdenServicio,
			"descuentos_habilitados":          cfg.DescuentosHabilitados,
			"permitir_descuento_porcentaje":   cfg.PermitirDescuentoPorcentaje,
			"permitir_descuento_codigo":       cfg.PermitirDescuentoCodigo,
			"permitir_descuento_valor":        cfg.PermitirDescuentoValor,
			"codigos_descuento":               cfg.CodigosDescuento,
			"lector_codigo_barras_habilitado": cfg.LectorCodigoBarrasHabilitado,
			"lector_codigo_barras_autofoco":   cfg.LectorCodigoBarrasAutofoco,
			"lector_codigo_barras_acumular":   cfg.LectorCodigoBarrasAcumular,
		},
	}
}