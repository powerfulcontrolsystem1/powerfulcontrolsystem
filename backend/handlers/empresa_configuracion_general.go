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

type empresaConfiguracionGeneralCajaPayload struct {
	CajaNombre                          string `json:"caja_nombre"`
	CajaCodigo                          string `json:"caja_codigo"`
	CajaActiva                          bool   `json:"caja_activa"`
	CajonMonederoHabilitado             bool   `json:"cajon_monedero_habilitado"`
	AbrirCajonAlPagarCarrito            bool   `json:"abrir_cajon_al_pagar_carrito"`
	AbrirCajonAlCerrarTransaccion       bool   `json:"abrir_cajon_al_cerrar_transaccion"`
	CajonMonederoMetodo                 string `json:"cajon_monedero_metodo"`
	CajonMonederoImpresoraFuncionalidad string `json:"cajon_monedero_impresora_funcionalidad"`
	CajonMonederoComando                string `json:"cajon_monedero_comando"`
	CajaObservaciones                   string `json:"caja_observaciones"`
}

type empresaConfiguracionGeneralPayload struct {
	EmpresaID     int64                                        `json:"empresa_id"`
	Productos     *empresaConfiguracionGeneralProductosPayload `json:"productos,omitempty"`
	Caja          *empresaConfiguracionGeneralCajaPayload      `json:"caja,omitempty"`
	Estado        string                                       `json:"estado,omitempty"`
	Observaciones string                                       `json:"observaciones,omitempty"`
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

			cfg, err := dbpkg.GetEmpresaConfiguracionGeneral(dbEmp, payload.EmpresaID)
			if err != nil {
				log.Printf("[empresa_config_general] preload empresa_id=%d error: %v", payload.EmpresaID, err)
				http.Error(w, "No se pudo cargar la configuracion actual", http.StatusInternalServerError)
				return
			}
			if cfg == nil {
				cfg = &dbpkg.EmpresaConfiguracionGeneral{EmpresaID: payload.EmpresaID}
			}
			cfg.EmpresaID = payload.EmpresaID
			cfg.Estado = strings.TrimSpace(payload.Estado)
			cfg.Observaciones = strings.TrimSpace(payload.Observaciones)
			cfg.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if payload.Productos != nil {
				cfg.ImprimirOrdenServicio = payload.Productos.ImprimirOrdenServicio
				cfg.AreaDespacho = payload.Productos.AreaDespacho
				cfg.CopiasOrdenServicio = payload.Productos.CopiasOrdenServicio
				cfg.NotaOrdenServicio = payload.Productos.NotaOrdenServicio
				cfg.DescuentosHabilitados = payload.Productos.DescuentosHabilitados
				cfg.PermitirDescuentoPorcentaje = payload.Productos.PermitirDescuentoPorcentaje
				cfg.PermitirDescuentoCodigo = payload.Productos.PermitirDescuentoCodigo
				cfg.PermitirDescuentoValor = payload.Productos.PermitirDescuentoValor
				cfg.CodigosDescuento = payload.Productos.CodigosDescuento
				cfg.LectorCodigoBarrasHabilitado = payload.Productos.LectorCodigoBarrasHabilitado
				cfg.LectorCodigoBarrasAutofoco = payload.Productos.LectorCodigoBarrasAutofoco
				cfg.LectorCodigoBarrasAcumular = payload.Productos.LectorCodigoBarrasAcumular
			}
			if payload.Caja != nil {
				cfg.CajaNombre = payload.Caja.CajaNombre
				cfg.CajaCodigo = payload.Caja.CajaCodigo
				cfg.CajaActiva = payload.Caja.CajaActiva
				cfg.CajonMonederoHabilitado = payload.Caja.CajonMonederoHabilitado
				cfg.AbrirCajonAlPagarCarrito = payload.Caja.AbrirCajonAlPagarCarrito
				cfg.AbrirCajonAlCerrarTransaccion = payload.Caja.AbrirCajonAlCerrarTransaccion
				cfg.CajonMonederoMetodo = payload.Caja.CajonMonederoMetodo
				cfg.CajonMonederoImpresoraFuncionalidad = payload.Caja.CajonMonederoImpresoraFuncionalidad
				cfg.CajonMonederoComando = payload.Caja.CajonMonederoComando
				cfg.CajaObservaciones = payload.Caja.CajaObservaciones
			}
			id, err := dbpkg.UpsertEmpresaConfiguracionGeneral(dbEmp, *cfg)
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
			"caja":       map[string]interface{}{},
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
		"caja": map[string]interface{}{
			"caja_nombre":                            cfg.CajaNombre,
			"caja_codigo":                            cfg.CajaCodigo,
			"caja_activa":                            cfg.CajaActiva,
			"cajon_monedero_habilitado":              cfg.CajonMonederoHabilitado,
			"abrir_cajon_al_pagar_carrito":           cfg.AbrirCajonAlPagarCarrito,
			"abrir_cajon_al_cerrar_transaccion":      cfg.AbrirCajonAlCerrarTransaccion,
			"cajon_monedero_metodo":                  cfg.CajonMonederoMetodo,
			"cajon_monedero_impresora_funcionalidad": cfg.CajonMonederoImpresoraFuncionalidad,
			"cajon_monedero_comando":                 cfg.CajonMonederoComando,
			"caja_observaciones":                     cfg.CajaObservaciones,
		},
	}
}
