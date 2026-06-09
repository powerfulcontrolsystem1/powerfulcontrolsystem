package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func parseTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "si", "yes", "activo":
		return true
	default:
		return false
	}
}

type facturacionOperacionPayload struct {
	EmpresaID               int64   `json:"empresa_id"`
	EntidadID               int64   `json:"entidad_id"`
	ClienteID               int64   `json:"cliente_id"`
	TipoDocumento           string  `json:"tipo_documento"`
	ClienteEmail            string  `json:"cliente_email"`
	ClienteNombre           string  `json:"cliente_nombre"`
	PaisCodigo              string  `json:"pais_codigo"`
	DocumentoCodigo         string  `json:"documento_codigo"`
	EstadoActual            string  `json:"estado_actual"`
	MontoTotal              float64 `json:"monto_total"`
	Moneda                  string  `json:"moneda"`
	PeriodoContable         string  `json:"periodo_contable"`
	Observaciones           string  `json:"observaciones"`
	PermitirModoOffline     bool    `json:"permitir_modo_offline"`
	ConfirmarModoOffline    bool    `json:"confirmar_modo_offline"`
	OrigenModoOffline       string  `json:"origen_modo_offline"`
	MensajeConfirmacionDIAN string  `json:"mensaje_confirmacion_dian"`
}

type facturaEmailResultado struct {
	Intentado             bool   `json:"intentado"`
	Enviado               bool   `json:"enviado"`
	AutomaticoDesactivado bool   `json:"automatico_desactivado,omitempty"`
	Destinatario          string `json:"destinatario,omitempty"`
	ClienteID             int64  `json:"cliente_id,omitempty"`
	OrigenDestinatario    string `json:"origen_destinatario,omitempty"`
	Error                 string `json:"error,omitempty"`
}

type facturacionIntegracionResultado struct {
	Aplica                      bool   `json:"aplica"`
	Accion                      string `json:"accion"`
	PaisCodigo                  string `json:"pais_codigo,omitempty"`
	Proveedor                   string `json:"proveedor,omitempty"`
	Ambiente                    string `json:"ambiente,omitempty"`
	EstadoEnvio                 string `json:"estado_envio"`
	Intentos                    int64  `json:"intentos"`
	MaxIntentos                 int64  `json:"max_intentos"`
	ProximoIntento              string `json:"proximo_intento,omitempty"`
	ContingenciaActiva          bool   `json:"contingencia_activa"`
	ReferenciaExterna           string `json:"referencia_externa,omitempty"`
	Error                       string `json:"error,omitempty"`
	OfflineDisponible           bool   `json:"offline_disponible,omitempty"`
	OfflineConfirmado           bool   `json:"offline_confirmado,omitempty"`
	RequiereConfirmacionOffline bool   `json:"requiere_confirmacion_offline,omitempty"`
	ConexionEstado              string `json:"conexion_estado,omitempty"`
	ConexionMensaje             string `json:"conexion_mensaje,omitempty"`
	AccionRecomendada           string `json:"accion_recomendada,omitempty"`
}

type facturacionProveedorDispatchResult struct {
	Success             bool
	ReferenciaExterna   string
	RespuestaJSON       string
	Error               string
	ConnectivityFailure bool
	HTTPStatus          int
}

type facturacionDianOfflineSettings struct {
	Enabled           bool   `json:"modo_offline_dian_activo"`
	AskBeforeContinue bool   `json:"modo_offline_preguntar"`
	AutoRetry         bool   `json:"modo_offline_auto_reintentar"`
	ContingencyType   string `json:"dian_contingencia_tipo"`
}

func isISODateYYYYMMDD(v string) bool {
	v = strings.TrimSpace(v)
	if len(v) != 10 {
		return false
	}
	for i := 0; i < len(v); i += 1 {
		if i == 4 || i == 7 {
			if v[i] != '-' {
				return false
			}
			continue
		}
		if v[i] < '0' || v[i] > '9' {
			return false
		}
	}
	return true
}

// EmpresaFacturacionElectronicaHandler gestiona configuración FE por empresa y país.
func EmpresaFacturacionElectronicaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "documentos" {
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				offset, err := parseIntQueryOptional(r, "offset")
				if err != nil {
					http.Error(w, "offset invalido", http.StatusBadRequest)
					return
				}

				fechaDesde := strings.TrimSpace(r.URL.Query().Get("fecha_desde"))
				if fechaDesde != "" && !isISODateYYYYMMDD(fechaDesde) {
					http.Error(w, "fecha_desde invalida (use YYYY-MM-DD)", http.StatusBadRequest)
					return
				}
				fechaHasta := strings.TrimSpace(r.URL.Query().Get("fecha_hasta"))
				if fechaHasta != "" && !isISODateYYYYMMDD(fechaHasta) {
					http.Error(w, "fecha_hasta invalida (use YYYY-MM-DD)", http.StatusBadRequest)
					return
				}

				items, err := dbpkg.ListEmpresaDocumentosFacturacionByEmpresa(dbEmp, dbpkg.EmpresaDocumentoFacturacionListFilter{
					EmpresaID:       empresaID,
					TipoDocumento:   strings.TrimSpace(r.URL.Query().Get("tipo_documento")),
					EstadoDocumento: strings.TrimSpace(r.URL.Query().Get("estado_documento")),
					IncludeInactive: parseTruthy(r.URL.Query().Get("include_inactive")) || parseTruthy(r.URL.Query().Get("incluir_inactivas")),
					ClienteQuery:    strings.TrimSpace(r.URL.Query().Get("cliente")),
					DocumentoQuery:  strings.TrimSpace(r.URL.Query().Get("documento")),
					FechaDesde:      fechaDesde,
					FechaHasta:      fechaHasta,
					Query:           strings.TrimSpace(r.URL.Query().Get("q")),
					Limit:           limit,
					Offset:          offset,
				})
				if err != nil {
					http.Error(w, "No se pudo listar documentos de facturacion", http.StatusInternalServerError)
					return
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"empresa_id": empresaID,
					"items":      items,
				})
				return
			}

			if action == "reintentos" {
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				offset, err := parseIntQueryOptional(r, "offset")
				if err != nil {
					http.Error(w, "offset invalido", http.StatusBadRequest)
					return
				}

				items, err := dbpkg.ListFacturacionElectronicaRetriesByEmpresa(dbEmp, empresaID, dbpkg.FacturacionElectronicaRetryFilter{
					TipoDocumento:   strings.TrimSpace(r.URL.Query().Get("tipo_documento")),
					EstadoEnvio:     strings.TrimSpace(r.URL.Query().Get("estado_envio")),
					DocumentoQuery:  strings.TrimSpace(comprasFirstNonBlank(r.URL.Query().Get("q"), r.URL.Query().Get("documento"))),
					SoloVencidos:    parseTruthy(comprasFirstNonBlank(r.URL.Query().Get("solo_vencidos"), r.URL.Query().Get("vencidos"))),
					IncludeInactive: parseTruthy(r.URL.Query().Get("include_inactive")) || parseTruthy(r.URL.Query().Get("incluir_inactivas")),
					Limit:           limit,
					Offset:          offset,
				})
				if err != nil {
					http.Error(w, "No se pudo listar cola de reintentos FE", http.StatusInternalServerError)
					return
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"empresa_id": empresaID,
					"items":      items,
				})
				return
			}

			if action == "reconciliacion" || action == "reconciliar_estados" {
				resumen, err := buildFacturacionReconciliacion(dbEmp, empresaID)
				if err != nil {
					http.Error(w, "No se pudo calcular reconciliacion FE", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return
			}

			if action == "estado_conexion_dian" || action == "estado_conexion" {
				paisCodigo := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("pais_codigo")))
				if paisCodigo == "" {
					paisCodigo = "CO"
				}
				cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, paisCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar conectividad DIAN", http.StatusInternalServerError)
					return
				}
				status := facturacionDIANConnectionStatus(dbEmp, empresaID, paisCodigo, cfg)
				if cfg != nil && parseTruthy(r.URL.Query().Get("procesar_reintentos")) {
					if online, _ := status["online"].(bool); online {
						settings := facturacionDianOfflineSettingsFromConfig(cfg)
						if settings.AutoRetry {
							processed, procErr := processFacturacionRetryQueue(dbEmp, empresaID, 100, strings.TrimSpace(adminEmailFromRequest(r)))
							if procErr != nil {
								status["retry_error"] = procErr.Error()
							} else {
								status["retry_procesado"] = processed
							}
						}
					}
				}
				writeJSON(w, http.StatusOK, status)
				return
			}

			if action == "catalogo_dian_colombia" || action == "documentos_dian_colombia" {
				cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, "CO")
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar catalogo DIAN Colombia", http.StatusInternalServerError)
					return
				}
				extra := map[string]interface{}{}
				if cfg != nil {
					extra = facturacionTryParseJSONMap(cfg.CamposPaisJSON)
				}
				documentosActivos := facturacionStringListFromAny(extra["documentos_soportados"])
				if len(documentosActivos) == 0 {
					documentosActivos = dbpkg.DefaultFacturacionDianDocumentosSoportados()
				}
				obligacionesActivas := facturacionStringListFromAny(extra["documentos_contadores_colombia"])
				if len(obligacionesActivas) == 0 {
					obligacionesActivas = []string{"declaraciones_tributarias", "informacion_exogena", "certificados_retencion", "conciliacion_fiscal"}
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                             true,
					"empresa_id":                     empresaID,
					"pais_codigo":                    "CO",
					"documentos":                     dbpkg.ListFacturacionDianDocumentosElectronicos(),
					"documentos_soportados":          documentosActivos,
					"obligaciones_contador":          dbpkg.ListFacturacionDianObligacionesContadores(),
					"documentos_contadores_colombia": obligacionesActivas,
					"fuentes":                        dbpkg.ListFacturacionDianFuentesNormativas(),
					"nota":                           "El catalogo separa documentos electronicos del SFE y obligaciones contables/tributarias que preparan contadores.",
				})
				return
			}

			paisCodigo := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("pais_codigo")))
			incluirInactivas := parseTruthy(r.URL.Query().Get("incluir_inactivas"))

			if paisCodigo != "" {
				cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, paisCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar la configuración de facturación electrónica", http.StatusInternalServerError)
					return
				}
				if cfg == nil {
					http.Error(w, "No se pudo resolver la configuración", http.StatusInternalServerError)
					return
				}
				if errors.Is(err, sql.ErrNoRows) {
					pais, source, derr := dbpkg.DetectFacturacionPais(dbEmp, empresaID, r.URL.Query().Get("tz"), r.URL.Query().Get("lang"))
					if derr == nil {
						cfg.PaisCodigo = pais.Codigo
						cfg.PaisNombre = pais.Nombre
						cfg.BanderaPais = pais.Bandera
						cfg.MonedaCodigo = pais.Moneda
						if cfg.Observaciones == "" {
							cfg.Observaciones = "Pais detectado por " + source
						}
					}
				}
				writeJSON(w, http.StatusOK, cfg)
				return
			}

			items, err := dbpkg.ListFacturacionElectronicaPaisConfigs(dbEmp, empresaID, incluirInactivas)
			if err != nil {
				http.Error(w, "No se pudo listar la configuración de facturación electrónica", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"empresa_id": empresaID,
				"items":      items,
			})
			return

		case http.MethodPost, http.MethodPut:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "procesar_reintentos" {
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				limit, err := parseIntQueryOptional(r, "limit")
				if err != nil {
					http.Error(w, "limit invalido", http.StatusBadRequest)
					return
				}
				processed, err := processFacturacionRetryQueue(dbEmp, empresaID, limit, strings.TrimSpace(adminEmailFromRequest(r)))
				if err != nil {
					http.Error(w, "No se pudo procesar cola de reintentos FE", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, processed)
				return
			}

			if action == "reconciliar_estados" {
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				aplicar := parseTruthy(comprasFirstNonBlank(r.URL.Query().Get("aplicar"), r.URL.Query().Get("sync"), r.URL.Query().Get("apply")))
				resumen, err := reconcileFacturacionEstados(dbEmp, empresaID, aplicar, strings.TrimSpace(adminEmailFromRequest(r)))
				if err != nil {
					http.Error(w, "No se pudo reconciliar estados FE", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, resumen)
				return
			}

			if action == "facturar_desde_venta" {
				var payload facturacionOperacionPayload
				if r.Body != nil {
					_ = json.NewDecoder(r.Body).Decode(&payload)
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.TipoDocumento) == "" {
					payload.TipoDocumento = strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
				}
				if strings.TrimSpace(payload.TipoDocumento) == "" {
					payload.TipoDocumento = "comprobante_pago"
				}

				ventaDoc, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, payload.EmpresaID, payload.TipoDocumento, payload.DocumentoCodigo)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "venta no encontrada", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar la venta", http.StatusInternalServerError)
					return
				}
				if strings.TrimSpace(strings.ToLower(ventaDoc.TipoDocumento)) != "comprobante_pago" {
					http.Error(w, "solo se puede facturar electronicamente una venta/comprobante", http.StatusConflict)
					return
				}
				if strings.TrimSpace(strings.ToLower(ventaDoc.EstadoDocumento)) != "emitida" {
					http.Error(w, "la venta debe estar emitida para generar la factura electronica", http.StatusConflict)
					return
				}

				resultado, err := registrarFacturaElectronicaDesdeDocumentoVenta(
					dbEmp,
					dbSuper,
					ventaDoc,
					strings.TrimSpace(adminEmailFromRequest(r)),
					"factura electronica generada manualmente desde la bandeja de ventas",
				)
				if err != nil {
					http.Error(w, "No se pudo generar la factura electronica desde la venta", http.StatusInternalServerError)
					return
				}

				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":               true,
					"accion":           "facturar_desde_venta",
					"empresa_id":       payload.EmpresaID,
					"venta_origen":     ventaDoc,
					"factura_generada": resultado,
				})
				return
			}

			if action == "reenviar_correo" || action == "enviar_correo" {
				var payload facturacionOperacionPayload
				if r.Body != nil {
					_ = json.NewDecoder(r.Body).Decode(&payload)
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.TipoDocumento) == "" {
					payload.TipoDocumento = strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
				}
				doc, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, payload.EmpresaID, payload.TipoDocumento, payload.DocumentoCodigo)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "documento no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar el documento", http.StatusInternalServerError)
					return
				}
				if payload.ClienteID <= 0 && payload.EntidadID <= 0 && doc.EntidadRelacionadaID > 0 {
					payload.ClienteID = doc.EntidadRelacionadaID
					payload.EntidadID = doc.EntidadRelacionadaID
				}

				resultado := enviarFacturaElectronicaAlCliente(dbEmp, dbSuper, payload, *doc)
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":               true,
					"accion":           "reenviar_correo",
					"empresa_id":       payload.EmpresaID,
					"tipo_documento":   doc.TipoDocumento,
					"documento_codigo": doc.DocumentoCodigo,
					"factura_email":    resultado,
				})
				return
			}
			if !facturacionActionIsPaisConfig(action) && facturacionActionRequiresFiscalIntegration(action) {
				var payload facturacionOperacionPayload
				if r.Body != nil {
					_ = json.NewDecoder(r.Body).Decode(&payload)
				}
				if payload.EmpresaID <= 0 {
					if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
						payload.EmpresaID = empresaID
					}
				}
				if payload.EmpresaID <= 0 {
					http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					payload.DocumentoCodigo = strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
				}
				if strings.TrimSpace(payload.DocumentoCodigo) == "" {
					http.Error(w, "documento_codigo es obligatorio para la accion", http.StatusBadRequest)
					return
				}

				if strings.TrimSpace(payload.EstadoActual) == "" {
					payload.EstadoActual = strings.TrimSpace(r.URL.Query().Get("estado_actual"))
				}

				if payload.ClienteID <= 0 && payload.EntidadID > 0 {
					payload.ClienteID = payload.EntidadID
				}
				if payload.EntidadID <= 0 && payload.ClienteID > 0 {
					payload.EntidadID = payload.ClienteID
				}

				documentoTipo := normalizeFacturacionDocumentoElectronicoTipo(payload.TipoDocumento)
				entidad := facturacionDocumentoEntidad(documentoTipo)
				actionNormalized := normalizeDocumentoState(action)
				if fromAction := facturacionDocumentoTipoFromAction(actionNormalized); fromAction != "" {
					documentoTipo = fromAction
					entidad = facturacionDocumentoEntidad(documentoTipo)
				}
				if !facturacionDocumentoElectronicoPermitido(documentoTipo) {
					http.Error(w, "tipo_documento electronico no soportado", http.StatusBadRequest)
					return
				}

				docExistente, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, payload.EmpresaID, documentoTipo, payload.DocumentoCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar el estado documental de facturacion", http.StatusInternalServerError)
					return
				}
				if docExistente != nil {
					payload.EstadoActual = docExistente.EstadoDocumento
				}

				transition, err := resolveFacturacionTransitionForDocument(action, payload.EstadoActual, documentoTipo)
				if err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}

				if block, preflightErr := facturacionOfflineDianPreflight(dbEmp, payload); preflightErr != nil {
					http.Error(w, "No se pudo validar conectividad DIAN", http.StatusInternalServerError)
					return
				} else if block != nil {
					writeJSON(w, http.StatusConflict, block)
					return
				}

				var legalDoc *dbpkg.FacturacionDocumentoLegal
				if transition.Accion == "emitir" && documentoTipo == "factura_electronica" {
					paisCodigo := strings.TrimSpace(payload.PaisCodigo)
					if paisCodigo == "" {
						paisCodigo = strings.TrimSpace(r.URL.Query().Get("pais_codigo"))
					}
					legalDoc, err = dbpkg.PrepareFacturacionDocumentoLegal(dbEmp, payload.EmpresaID, paisCodigo, payload.DocumentoCodigo, payload.MontoTotal, payload.Moneda)
					if err != nil {
						http.Error(w, "cumplimiento normativo: "+err.Error(), http.StatusUnprocessableEntity)
						return
					}
				}

				evento := transition.Evento
				docPayload := dbpkg.EmpresaDocumentoFacturacion{
					EmpresaID:            payload.EmpresaID,
					TipoDocumento:        documentoTipo,
					DocumentoCodigo:      payload.DocumentoCodigo,
					EstadoDocumento:      transition.EstadoNuevo,
					EstadoAnterior:       transition.EstadoAnterior,
					EventoUltimo:         evento,
					PeriodoContable:      payload.PeriodoContable,
					MontoTotal:           payload.MontoTotal,
					Moneda:               payload.Moneda,
					EntidadRelacionadaID: payload.EntidadID,
					UsuarioCreador:       strings.TrimSpace(adminEmailFromRequest(r)),
					Observaciones:        payload.Observaciones,
				}
				if legalDoc != nil {
					docPayload.NumeroLegal = legalDoc.NumeroLegal
					docPayload.CodigoValidacion = legalDoc.CodigoValidacion
					docPayload.PaisCodigo = legalDoc.PaisCodigo
					docPayload.AmbienteFE = legalDoc.Ambiente
					docPayload.FechaDocumento = legalDoc.FechaEmisionLegal
				}
				if docExistente != nil {
					if docPayload.NumeroLegal == "" {
						docPayload.NumeroLegal = docExistente.NumeroLegal
					}
					if docPayload.CodigoValidacion == "" {
						docPayload.CodigoValidacion = docExistente.CodigoValidacion
					}
					if docPayload.PaisCodigo == "" {
						docPayload.PaisCodigo = docExistente.PaisCodigo
					}
					if docPayload.AmbienteFE == "" {
						docPayload.AmbienteFE = docExistente.AmbienteFE
					}
					if docPayload.FechaDocumento == "" {
						docPayload.FechaDocumento = docExistente.FechaDocumento
					}
				}

				docPersistido, err := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, docPayload)
				if err != nil {
					http.Error(w, "No se pudo persistir el documento transaccional", http.StatusInternalServerError)
					return
				}

				registrarEventoContableNoBloqueante(dbEmp, r, "facturacion", dbpkg.EmpresaEventoContable{
					EmpresaID:       payload.EmpresaID,
					Modulo:          "facturacion",
					Evento:          evento,
					Entidad:         entidad,
					EntidadID:       docPersistido.ID,
					DocumentoTipo:   documentoTipo,
					DocumentoCodigo: strings.TrimSpace(payload.DocumentoCodigo),
					PeriodoContable: strings.TrimSpace(payload.PeriodoContable),
					MontoTotal:      payload.MontoTotal,
					Moneda:          strings.ToUpper(strings.TrimSpace(payload.Moneda)),
					Origen:          "api_facturacion_electronica",
					Observaciones:   strings.TrimSpace(payload.Observaciones),
				}, map[string]interface{}{
					"accion":            transition.Accion,
					"estado_anterior":   transition.EstadoAnterior,
					"estado_nuevo":      transition.EstadoNuevo,
					"entidad_id":        docPersistido.ID,
					"documento_codigo":  strings.TrimSpace(payload.DocumentoCodigo),
					"numero_legal":      docPersistido.NumeroLegal,
					"codigo_validacion": docPersistido.CodigoValidacion,
					"pais_codigo":       docPersistido.PaisCodigo,
					"ambiente_fe":       docPersistido.AmbienteFE,
					"periodo_contable":  strings.TrimSpace(payload.PeriodoContable),
					"empresa_id":        payload.EmpresaID,
				})

				integracionFiscal := facturacionIntegracionResultado{
					Aplica:             false,
					Accion:             transition.Accion,
					EstadoEnvio:        "no_aplica",
					ContingenciaActiva: false,
				}
				var retryRegistro *dbpkg.FacturacionElectronicaRetryItem

				if facturacionActionRequiresFiscalIntegration(transition.Accion) {
					resultadoIntegracion, retryItem, integErr := processFacturacionIntegracionForDocumento(
						dbEmp,
						payload,
						*docPersistido,
						transition.Accion,
						strings.TrimSpace(adminEmailFromRequest(r)),
					)
					if integErr != nil {
						log.Printf("[facturacion_electronica] error integracion fiscal empresa_id=%d documento=%s accion=%s err=%v", payload.EmpresaID, payload.DocumentoCodigo, transition.Accion, integErr)
					}
					integracionFiscal = resultadoIntegracion
					retryRegistro = retryItem

					eventoIntegracion := ""
					switch integracionFiscal.EstadoEnvio {
					case "enviado":
						eventoIntegracion = "factura_integracion_enviada"
					case "fallido":
						eventoIntegracion = "factura_integracion_fallida"
					case "contingencia":
						eventoIntegracion = "factura_contingencia_activada"
					}

					if eventoIntegracion != "" {
						registrarEventoContableNoBloqueante(dbEmp, r, "facturacion", dbpkg.EmpresaEventoContable{
							EmpresaID:       payload.EmpresaID,
							Modulo:          "facturacion",
							Evento:          eventoIntegracion,
							Entidad:         entidad,
							EntidadID:       docPersistido.ID,
							DocumentoTipo:   docPersistido.TipoDocumento,
							DocumentoCodigo: docPersistido.DocumentoCodigo,
							PeriodoContable: docPersistido.PeriodoContable,
							MontoTotal:      docPersistido.MontoTotal,
							Moneda:          docPersistido.Moneda,
							Origen:          "api_facturacion_electronica",
							Observaciones:   strings.TrimSpace(integracionFiscal.Error),
						}, map[string]interface{}{
							"accion":              transition.Accion,
							"estado_envio":        integracionFiscal.EstadoEnvio,
							"intentos":            integracionFiscal.Intentos,
							"max_intentos":        integracionFiscal.MaxIntentos,
							"contingencia_activa": integracionFiscal.ContingenciaActiva,
							"proximo_intento":     integracionFiscal.ProximoIntento,
							"referencia_externa":  integracionFiscal.ReferenciaExterna,
							"documento_codigo":    docPersistido.DocumentoCodigo,
							"codigo_validacion":   docPersistido.CodigoValidacion,
							"numero_legal":        docPersistido.NumeroLegal,
							"empresa_id":          payload.EmpresaID,
						})
					}
				}

				resp := map[string]interface{}{
					"ok":                 true,
					"accion":             transition.Accion,
					"evento":             evento,
					"estado_anterior":    transition.EstadoAnterior,
					"estado_nuevo":       transition.EstadoNuevo,
					"entidad_id":         docPersistido.ID,
					"documento_codigo":   strings.TrimSpace(payload.DocumentoCodigo),
					"numero_legal":       docPersistido.NumeroLegal,
					"codigo_validacion":  docPersistido.CodigoValidacion,
					"pais_codigo":        docPersistido.PaisCodigo,
					"ambiente_fe":        docPersistido.AmbienteFE,
					"integracion_fiscal": integracionFiscal,
				}
				if retryRegistro != nil {
					resp["cola_reintentos"] = retryRegistro
				}
				if legalDoc != nil {
					resp["cumplimiento_normativo"] = map[string]interface{}{
						"validado":               true,
						"prefijo_factura":        legalDoc.PrefijoFactura,
						"resolucion_numero":      legalDoc.ResolucionNumero,
						"consecutivo_asignado":   legalDoc.ConsecutivoAsignado,
						"fecha_emision_legal":    legalDoc.FechaEmisionLegal,
						"resolucion_fecha_desde": legalDoc.ResolucionFechaDesde,
						"resolucion_fecha_hasta": legalDoc.ResolucionFechaHasta,
					}
				}
				if transition.Accion == "emitir" && documentoTipo == "factura_electronica" {
					if facturacionAutoEmailClienteEnabled(dbEmp, payload.EmpresaID, payload.PaisCodigo) {
						resp["factura_email"] = enviarFacturaElectronicaAlCliente(dbEmp, dbSuper, payload, *docPersistido)
					} else {
						resp["factura_email"] = facturaEmailAutoDisabledResultado(payload)
					}
				}
				writeJSON(w, http.StatusOK, resp)
				return
			}

			var payload dbpkg.FacturacionElectronicaPaisConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}

			if payload.EmpresaID <= 0 {
				if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
					payload.EmpresaID = empresaID
				}
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.PaisCodigo) == "" {
				http.Error(w, "pais_codigo es obligatorio", http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

			id, err := dbpkg.UpsertFacturacionElectronicaPaisConfig(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar la configuración de facturación electrónica", http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, payload.EmpresaID, payload.PaisCodigo)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No se pudo recuperar la configuración guardada", http.StatusInternalServerError)
				return
			}
			monedaEvento := strings.ToUpper(strings.TrimSpace(payload.MonedaCodigo))
			if monedaEvento == "" && cfg != nil {
				monedaEvento = strings.ToUpper(strings.TrimSpace(cfg.MonedaCodigo))
			}
			registrarEventoContableNoBloqueante(dbEmp, r, "facturacion", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "facturacion",
				Evento:          "configuracion_facturacion_actualizada",
				Entidad:         "facturacion_electronica_pais",
				EntidadID:       id,
				DocumentoTipo:   "facturacion_pais",
				DocumentoCodigo: strings.ToUpper(strings.TrimSpace(payload.PaisCodigo)),
				Moneda:          monedaEvento,
				Origen:          "api_facturacion_electronica",
				Observaciones:   "configuracion de facturacion electronica actualizada",
			}, map[string]interface{}{
				"pais_codigo": strings.ToUpper(strings.TrimSpace(payload.PaisCodigo)),
				"ambiente":    strings.ToLower(strings.TrimSpace(payload.Ambiente)),
				"proveedor":   strings.TrimSpace(payload.Proveedor),
				"estado":      strings.ToLower(strings.TrimSpace(payload.Estado)),
			})
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"configuracion": cfg,
			})
			return
		}

		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func facturaEmailAutoDisabledResultado(payload facturacionOperacionPayload) facturaEmailResultado {
	clienteID := payload.ClienteID
	if clienteID <= 0 {
		clienteID = payload.EntidadID
	}
	return facturaEmailResultado{
		Intentado:             false,
		Enviado:               false,
		AutomaticoDesactivado: true,
		ClienteID:             clienteID,
		OrigenDestinatario:    "configuracion",
	}
}

// EmpresaFacturacionElectronicaPanamaHandler gestiona el perfil independiente Panama/DGI.
func EmpresaFacturacionElectronicaPanamaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, "PA")
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No se pudo cargar facturacion electronica Panama", http.StatusInternalServerError)
				return
			}
			if cfg == nil {
				tmp := dbpkg.FacturacionElectronicaPaisConfig{EmpresaID: empresaID, PaisCodigo: "PA"}
				_, _ = dbpkg.UpsertFacturacionElectronicaPaisConfig(dbEmp, tmp)
				cfg, _ = dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, "PA")
			}
			checklist := dbpkg.BuildFacturacionPanamaChecklist(cfg)
			if action == "checklist" || action == "validar" {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            checklist.Ok,
					"empresa_id":    empresaID,
					"pais_codigo":   "PA",
					"checklist":     checklist,
					"configuracion": cfg,
				})
				return
			}
			if action == "guia_onboarding" {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":          true,
					"empresa_id":  empresaID,
					"pais_codigo": "PA",
					"pasos": []map[string]string{
						{"clave": "registro_sfep", "titulo": "Registrarse en SFEP/e-Tax2.0", "detalle": "Completar la solicitud de factura electronica ante DGI Panama."},
						{"clave": "modalidad", "titulo": "Elegir modalidad", "detalle": "Seleccionar Facturador Gratuito o Proveedor Autorizado Calificado (PAC)."},
						{"clave": "firma", "titulo": "Configurar firma electronica", "detalle": "Registrar certificado o referencia segura para firmar documentos electronicos."},
						{"clave": "pruebas", "titulo": "Validar ambiente de pruebas", "detalle": "Probar emision, CAFE/CUFE/QR y respuesta del PAC o facturador antes de produccion."},
					},
					"fuentes": checklist.Fuentes,
				})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"empresa_id":    empresaID,
				"pais_codigo":   "PA",
				"configuracion": cfg,
				"checklist":     checklist,
				"vista":         dbpkg.FacturacionPaisVistaFor("PA"),
			})
			return
		case http.MethodPost, http.MethodPut:
			var payload dbpkg.FacturacionElectronicaPaisConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			payload.PaisCodigo = "PA"
			payload.PaisNombre = "Panama"
			payload.BanderaPais = "PA"
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if strings.TrimSpace(payload.MonedaCodigo) == "" {
				payload.MonedaCodigo = "PAB"
			}
			id, err := dbpkg.UpsertFacturacionElectronicaPaisConfig(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar facturacion electronica Panama", http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, "PA")
			if err != nil {
				http.Error(w, "No se pudo recuperar facturacion electronica Panama", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"empresa_id":    empresaID,
				"pais_codigo":   "PA",
				"configuracion": cfg,
				"checklist":     dbpkg.BuildFacturacionPanamaChecklist(cfg),
			})
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaFacturacionElectronicaEcuadorHandler gestiona el perfil independiente Ecuador/SRI.
func EmpresaFacturacionElectronicaEcuadorHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, "EC")
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "No se pudo cargar facturacion electronica Ecuador", http.StatusInternalServerError)
				return
			}
			if cfg == nil {
				tmp := dbpkg.FacturacionElectronicaPaisConfig{EmpresaID: empresaID, PaisCodigo: "EC"}
				_, _ = dbpkg.UpsertFacturacionElectronicaPaisConfig(dbEmp, tmp)
				cfg, _ = dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, "EC")
			}
			checklist := dbpkg.BuildFacturacionEcuadorChecklist(cfg)
			if action == "checklist" || action == "validar" {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":            checklist.Ok,
					"empresa_id":    empresaID,
					"pais_codigo":   "EC",
					"checklist":     checklist,
					"configuracion": cfg,
				})
				return
			}
			if action == "guia_onboarding" {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":          true,
					"empresa_id":  empresaID,
					"pais_codigo": "EC",
					"pasos": []map[string]string{
						{"clave": "firma", "titulo": "Adquirir firma electronica", "detalle": "Mantener vigente un certificado de firma electronica tipo archivo para firmar XML."},
						{"clave": "ambiente_pruebas", "titulo": "Preparar ambiente de pruebas", "detalle": "Configurar ambiente SRI 1 para validar XML, firma y secuencias antes de produccion."},
						{"clave": "autorizacion", "titulo": "Confirmar autorizacion SRI", "detalle": "Verificar autorizacion de emision de comprobantes electronicos en SRI en Linea para produccion."},
						{"clave": "ride", "titulo": "Generar RIDE y notificar", "detalle": "Emitir XML autorizado, generar representacion impresa RIDE y notificar por correo al destinatario."},
					},
					"fuentes": checklist.Fuentes,
				})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"empresa_id":    empresaID,
				"pais_codigo":   "EC",
				"configuracion": cfg,
				"checklist":     checklist,
				"vista":         dbpkg.FacturacionPaisVistaFor("EC"),
			})
			return
		case http.MethodPost, http.MethodPut:
			var payload dbpkg.FacturacionElectronicaPaisConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			payload.PaisCodigo = "EC"
			payload.PaisNombre = "Ecuador"
			payload.BanderaPais = "EC"
			payload.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))
			if strings.TrimSpace(payload.MonedaCodigo) == "" {
				payload.MonedaCodigo = "USD"
			}
			id, err := dbpkg.UpsertFacturacionElectronicaPaisConfig(dbEmp, payload)
			if err != nil {
				http.Error(w, "No se pudo guardar facturacion electronica Ecuador", http.StatusBadRequest)
				return
			}
			cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, "EC")
			if err != nil {
				http.Error(w, "No se pudo recuperar facturacion electronica Ecuador", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":            true,
				"id":            id,
				"empresa_id":    empresaID,
				"pais_codigo":   "EC",
				"configuracion": cfg,
				"checklist":     dbpkg.BuildFacturacionEcuadorChecklist(cfg),
			})
			return
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
	}
}

func facturacionAutoEmailClienteEnabled(dbEmp *sql.DB, empresaID int64, paisCodigo string) bool {
	if dbEmp == nil || empresaID <= 0 {
		return false
	}
	code := strings.TrimSpace(paisCodigo)
	if code == "" {
		code = "CO"
	}
	if cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, empresaID, code); err == nil && cfg != nil && cfg.EnviarFacturaEmailClienteAuto {
		return true
	}
	if cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, empresaID); err == nil && cfg != nil && cfg.EnviarFacturaElectronicaVenta {
		return true
	}
	return false
}

func enviarFacturaElectronicaAlCliente(dbEmp, dbSuper *sql.DB, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion) facturaEmailResultado {
	emailCliente, nombreCliente, clienteID, origen, err := resolverDestinoCorreoFactura(dbEmp, payload)
	resultado := facturaEmailResultado{
		Intentado:          false,
		Enviado:            false,
		ClienteID:          clienteID,
		OrigenDestinatario: origen,
	}
	if err != nil {
		resultado.Error = err.Error()
		return resultado
	}
	if strings.TrimSpace(emailCliente) == "" {
		resultado.Error = "sin destinatario de cliente para envio automatico"
		return resultado
	}

	resultado.Intentado = true
	resultado.Destinatario = emailCliente
	if err := sendFacturaElectronicaEmail(dbSuper, emailCliente, nombreCliente, doc, payload); err != nil {
		resultado.Error = err.Error()
		log.Printf("[facturacion_electronica] envio correo fallido empresa_id=%d documento=%s destinatario=%s error=%v", payload.EmpresaID, payload.DocumentoCodigo, emailCliente, err)
		return resultado
	}

	resultado.Enviado = true
	resultado.Error = ""
	return resultado
}

func resolverDestinoCorreoFactura(dbEmp *sql.DB, payload facturacionOperacionPayload) (string, string, int64, string, error) {
	clienteID := payload.ClienteID
	if clienteID <= 0 {
		clienteID = payload.EntidadID
	}

	emailCliente := strings.TrimSpace(payload.ClienteEmail)
	nombreCliente := strings.TrimSpace(payload.ClienteNombre)
	if emailCliente != "" {
		if _, err := mail.ParseAddress(emailCliente); err != nil {
			return "", nombreCliente, clienteID, "payload", fmt.Errorf("cliente_email invalido: %w", err)
		}
		if nombreCliente == "" {
			nombreCliente = "cliente"
		}
		return emailCliente, nombreCliente, clienteID, "payload", nil
	}

	if clienteID <= 0 {
		return "", nombreCliente, 0, "sin_cliente", nil
	}

	cliente, err := dbpkg.GetClienteByID(dbEmp, payload.EmpresaID, clienteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nombreCliente, clienteID, "cliente_id", nil
		}
		return "", nombreCliente, clienteID, "cliente_id", err
	}

	emailCliente = strings.TrimSpace(cliente.Email)
	if nombreCliente == "" {
		nombreCliente = strings.TrimSpace(cliente.NombreRazonSocial)
	}
	if emailCliente == "" {
		return "", nombreCliente, clienteID, "cliente_id", nil
	}
	if _, err := mail.ParseAddress(emailCliente); err != nil {
		return "", nombreCliente, clienteID, "cliente_id", fmt.Errorf("email de cliente invalido en registro: %w", err)
	}
	if nombreCliente == "" {
		nombreCliente = "cliente"
	}

	return emailCliente, nombreCliente, clienteID, "cliente_id", nil
}

func sendFacturaElectronicaEmail(dbSuper *sql.DB, toEmail, toName string, doc dbpkg.EmpresaDocumentoFacturacion, payload facturacionOperacionPayload) error {
	if dbSuper == nil {
		return fmt.Errorf("configuracion SMTP no disponible")
	}

	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return fmt.Errorf("gmail.smtp_email no configurado")
	}

	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return err
	}
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" {
		return fmt.Errorf("gmail.smtp_app_password no configurado")
	}

	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")

	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" {
		smtpPort = "587"
	}
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Powerful Control System"
	}

	mailHostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, err := net.SplitHostPort(smtpHost); err == nil {
			mailHostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}

	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	safeName := strings.TrimSpace(toName)
	if safeName == "" {
		safeName = "cliente"
	}
	numeroLegal := strings.TrimSpace(doc.NumeroLegal)
	if numeroLegal == "" {
		numeroLegal = strings.TrimSpace(doc.DocumentoCodigo)
	}
	codigoValidacion := strings.TrimSpace(doc.CodigoValidacion)
	monto := doc.MontoTotal
	if monto <= 0 {
		monto = payload.MontoTotal
	}
	moneda := strings.ToUpper(strings.TrimSpace(doc.Moneda))
	if moneda == "" {
		moneda = strings.ToUpper(strings.TrimSpace(payload.Moneda))
	}
	if moneda == "" {
		moneda = "COP"
	}

	documentLabel := "Factura electronica"
	introLine := "Tu factura electronica fue emitida correctamente."
	feDetail := "Pais FE: " + strings.ToUpper(strings.TrimSpace(doc.PaisCodigo)) + "\r\n" +
		"Ambiente FE: " + strings.TrimSpace(doc.AmbienteFE) + "\r\n"
	if strings.EqualFold(strings.TrimSpace(doc.TipoDocumento), "comprobante_pago") {
		documentLabel = "Comprobante de pago"
		introLine = "Tu comprobante de pago fue generado correctamente."
		feDetail = ""
	}

	subject := documentLabel + " emitido " + numeroLegal
	body := "Hola " + safeName + ",\r\n\r\n" +
		introLine + "\r\n" +
		"Documento: " + strings.TrimSpace(doc.DocumentoCodigo) + "\r\n" +
		"Numero legal: " + numeroLegal + "\r\n" +
		"Codigo de validacion: " + codigoValidacion + "\r\n" +
		"Total: " + fmt.Sprintf("%.2f", monto) + " " + moneda + "\r\n" +
		feDetail + "\r\n" +
		"Gracias por tu compra.\r\n"

	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	return smtp.SendMail(addr, auth, smtpEmail, []string{toEmail}, []byte(msg))
}

// EmpresaFacturacionElectronicaPaisDetectadoHandler detecta automáticamente país FE.
func EmpresaFacturacionElectronicaPaisDetectadoHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}

		empresaID, err := parseInt64QueryOptional(r, "empresa_id")
		if err != nil {
			http.Error(w, "empresa_id inválido", http.StatusBadRequest)
			return
		}

		tz := strings.TrimSpace(r.URL.Query().Get("tz"))
		if tz == "" {
			tz = strings.TrimSpace(r.URL.Query().Get("timezone"))
		}
		lang := strings.TrimSpace(r.URL.Query().Get("lang"))
		if lang == "" {
			acceptLang := strings.TrimSpace(r.Header.Get("Accept-Language"))
			if idx := strings.Index(acceptLang, ","); idx > 0 {
				lang = strings.TrimSpace(acceptLang[:idx])
			} else {
				lang = acceptLang
			}
		}

		pais, source, err := dbpkg.DetectFacturacionPais(dbEmp, empresaID, tz, lang)
		if err != nil {
			http.Error(w, "No se pudo detectar el país", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"empresa_id":  empresaID,
			"pais_codigo": pais.Codigo,
			"pais_nombre": pais.Nombre,
			"bandera":     pais.Bandera,
			"moneda":      pais.Moneda,
			"source":      source,
			"vista":       dbpkg.FacturacionPaisVistaFor(pais.Codigo),
		})
	}
}

// EmpresaFacturacionElectronicaPaisesDisponiblesHandler retorna catálogo de países FE soportados.
func EmpresaFacturacionElectronicaPaisesDisponiblesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": dbpkg.ListFacturacionPaisesConVista(),
		})
	}
}

func normalizeFacturacionEstadoEnvio(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pendiente", "fallido", "enviado", "reconciliado", "contingencia", "no_aplica":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "pendiente"
	}
}

func facturacionNowLocal() string {
	return time.Now().In(time.Local).Format("2006-01-02 15:04:05")
}

func facturacionNextRetryAt(intentos int64) string {
	if intentos < 0 {
		intentos = 0
	}
	minutes := int64(1)
	if intentos > 0 {
		minutes = 1 << intentos
	}
	if minutes > 120 {
		minutes = 120
	}
	return time.Now().In(time.Local).Add(time.Duration(minutes) * time.Minute).Format("2006-01-02 15:04:05")
}

func facturacionFirstNonBlank(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func normalizeFacturacionDocumentoElectronicoTipo(raw string) string {
	v := normalizeDocumentoState(raw)
	switch v {
	case "", "factura", "factura_venta", "factura_de_venta", "factura_electronica_venta", "factura_electronica":
		return "factura_electronica"
	case "nota_credito", "nota_credito_ventas", "nota_credito_venta", "credit_note":
		return "nota_credito"
	case "nota_debito", "nota_debito_ventas", "nota_debito_venta", "debit_note":
		return "nota_debito"
	case "documento_soporte", "documento_soporte_electronico", "documento_soporte_adquisicion", "documento_soporte_adquisiciones", "soporte_compras":
		return "documento_soporte"
	case "nota_ajuste_documento_soporte", "nota_ajuste_soporte", "ajuste_documento_soporte":
		return "nota_ajuste_documento_soporte"
	case "nomina", "nomina_electronica", "documento_soporte_nomina", "documento_soporte_pago_nomina", "documento_soporte_de_pago_nomina", "documento_soporte_pago_nomina_electronica", "documento_soporte_de_pago_nomina_electronica":
		return "nomina_electronica"
	case "nota_ajuste_nomina", "nota_ajuste_nomina_electronica", "ajuste_nomina_electronica":
		return "nota_ajuste_nomina_electronica"
	case "pos", "pos_electronico", "tiquete_pos", "tiquete_maquina_registradora_pos", "tiquete_de_maquina_registradora_pos", "documento_equivalente", "documento_equivalente_pos", "documento_equivalente_electronico_pos":
		return "documento_equivalente_pos"
	case "nota_ajuste_documento_equivalente", "nota_ajuste_equivalente", "ajuste_documento_equivalente":
		return "nota_ajuste_documento_equivalente"
	case "factura_talonario", "factura_papel_contingencia", "talonario_contingencia", "factura_talonario_contingencia":
		return "factura_talonario_contingencia"
	case "eventos_radian", "evento_radian", "radian", "eventos_radian_recepcion":
		return "eventos_radian_recepcion"
	default:
		return v
	}
}

func facturacionDocumentoElectronicoPermitido(tipo string) bool {
	normalized := normalizeFacturacionDocumentoElectronicoTipo(tipo)
	for _, item := range dbpkg.ListFacturacionDianDocumentosElectronicos() {
		if item.Codigo == normalized {
			return true
		}
	}
	return false
}

func facturacionDocumentoTipoFromAction(actionRaw string) string {
	action := normalizeDocumentoState(actionRaw)
	switch action {
	case "nota_credito", "emitir_nota_credito":
		return "nota_credito"
	case "nota_debito", "emitir_nota_debito":
		return "nota_debito"
	case "documento_soporte", "emitir_documento_soporte":
		return "documento_soporte"
	case "nomina_electronica", "emitir_nomina_electronica":
		return "nomina_electronica"
	case "documento_equivalente_pos", "emitir_documento_equivalente_pos":
		return "documento_equivalente_pos"
	default:
		if strings.HasPrefix(action, "emitir_") {
			action = strings.TrimPrefix(action, "emitir_")
		}
		docType := normalizeFacturacionDocumentoElectronicoTipo(action)
		if facturacionDocumentoElectronicoPermitido(docType) {
			return docType
		}
		return ""
	}
}

func facturacionDocumentoEntidad(tipo string) string {
	switch normalizeFacturacionDocumentoElectronicoTipo(tipo) {
	case "factura_electronica":
		return "factura_electronica"
	case "nota_credito":
		return "nota_credito"
	case "nota_debito":
		return "nota_debito"
	case "documento_soporte":
		return "documento_soporte"
	case "nomina_electronica":
		return "nomina_electronica"
	case "documento_equivalente_pos":
		return "documento_equivalente_pos"
	default:
		normalized := normalizeFacturacionDocumentoElectronicoTipo(tipo)
		if facturacionDocumentoElectronicoPermitido(normalized) {
			return normalized
		}
		return "documento_electronico"
	}
}

func facturacionActionRequiresFiscalIntegration(action string) bool {
	actionNormalized := normalizeDocumentoState(action)
	switch actionNormalized {
	case "emitir", "anular", "nota_credito", "emitir_nota_credito", "nota_debito", "emitir_nota_debito", "documento_soporte", "emitir_documento_soporte", "nomina_electronica", "emitir_nomina_electronica", "documento_equivalente_pos", "emitir_documento_equivalente_pos":
		return true
	default:
		if strings.HasPrefix(actionNormalized, "emitir_") {
			actionNormalized = strings.TrimPrefix(actionNormalized, "emitir_")
		}
		return facturacionDocumentoElectronicoPermitido(actionNormalized)
	}
}

func facturacionActionIsPaisConfig(action string) bool {
	switch normalizeDocumentoState(action) {
	case "", "config_pais", "guardar_config_pais", "configuracion_pais":
		return true
	default:
		return false
	}
}

func facturacionTryParseJSONMap(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]interface{}{}
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return map[string]interface{}{}
	}
	if out == nil {
		return map[string]interface{}{}
	}
	return out
}

func facturacionAnyToBool(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t > 0
	case int:
		return t > 0
	case int64:
		return t > 0
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "1" || s == "true" || s == "si" || s == "yes" || s == "on"
	default:
		return false
	}
}

func facturacionStringListFromAny(v interface{}) []string {
	out := []string{}
	seen := map[string]struct{}{}
	appendOne := func(raw string) {
		item := strings.TrimSpace(raw)
		if item == "" {
			return
		}
		if _, ok := seen[item]; ok {
			return
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	switch t := v.(type) {
	case []string:
		for _, item := range t {
			appendOne(item)
		}
	case []interface{}:
		for _, item := range t {
			appendOne(fmt.Sprintf("%v", item))
		}
	case string:
		for _, item := range strings.Split(t, ",") {
			appendOne(item)
		}
	}
	return out
}

func facturacionDianOfflineSettingsFromConfig(cfg *dbpkg.FacturacionElectronicaPaisConfig) facturacionDianOfflineSettings {
	settings := facturacionDianOfflineSettings{
		Enabled:           false,
		AskBeforeContinue: false,
		AutoRetry:         false,
		ContingencyType:   "servicio_dian",
	}
	if cfg == nil {
		return settings
	}
	extra := facturacionTryParseJSONMap(cfg.CamposPaisJSON)
	if raw := strings.TrimSpace(fmt.Sprintf("%v", extra["dian_contingencia_tipo"])); raw != "" && raw != "<nil>" {
		settings.ContingencyType = strings.ToLower(raw)
	}
	if settings.ContingencyType == "" {
		settings.ContingencyType = "servicio_dian"
	}
	return settings
}

func facturacionIsConnectivityHTTPStatus(status int) bool {
	switch status {
	case http.StatusRequestTimeout, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func facturacionConnectivityMessage(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "no hay internet o no se detecta el servidor de la DIAN/proveedor"
	}
	return base
}

func facturacionTruncate(raw string, max int) string {
	raw = strings.TrimSpace(raw)
	if max <= 0 || len(raw) <= max {
		return raw
	}
	return strings.TrimSpace(raw[:max])
}

func facturacionExtractReferenciaExterna(raw string) string {
	m := facturacionTryParseJSONMap(raw)
	keys := []string{"referencia_externa", "external_reference", "reference", "id", "uuid", "codigo", "tracking_id"}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

func dispatchFacturacionProveedorHTTP(url string, payload map[string]interface{}) facturacionProveedorDispatchResult {
	body, err := json.Marshal(payload)
	if err != nil {
		return facturacionProveedorDispatchResult{Success: false, Error: "no se pudo serializar request de integracion"}
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return facturacionProveedorDispatchResult{Success: false, Error: "no se pudo construir request de integracion"}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
			return facturacionProveedorDispatchResult{Success: false, Error: "timeout de comunicacion con proveedor fiscal", ConnectivityFailure: true}
		}
		return facturacionProveedorDispatchResult{Success: false, Error: "fallo de comunicacion con proveedor fiscal", ConnectivityFailure: true}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	rawResp := strings.TrimSpace(string(respBody))
	if rawResp == "" {
		rawResp = "{}"
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ref := facturacionExtractReferenciaExterna(rawResp)
		if ref == "" {
			ref = strings.TrimSpace(resp.Header.Get("X-Referencia-Externa"))
		}
		if ref == "" {
			ref = fmt.Sprintf("EXT-%d", time.Now().UnixNano())
		}
		return facturacionProveedorDispatchResult{
			Success:           true,
			ReferenciaExterna: ref,
			RespuestaJSON:     rawResp,
		}
	}

	statusMsg := fmt.Sprintf("proveedor fiscal respondio HTTP %d", resp.StatusCode)
	if rawResp != "" && rawResp != "{}" {
		statusMsg += ": " + facturacionTruncate(rawResp, 280)
	}
	return facturacionProveedorDispatchResult{
		Success:             false,
		RespuestaJSON:       rawResp,
		Error:               statusMsg,
		ConnectivityFailure: facturacionIsConnectivityHTTPStatus(resp.StatusCode),
		HTTPStatus:          resp.StatusCode,
	}
}

func dispatchFacturacionProveedor(cfg *dbpkg.FacturacionElectronicaPaisConfig, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion, accion string) facturacionProveedorDispatchResult {
	proveedor := "manual"
	ambiente := "sandbox"
	apiBaseURL := ""
	camposPaisJSON := "{}"
	paisCodigo := ""

	if cfg != nil {
		if strings.TrimSpace(cfg.Proveedor) != "" {
			proveedor = strings.ToLower(strings.TrimSpace(cfg.Proveedor))
		}
		if strings.TrimSpace(cfg.Ambiente) != "" {
			ambiente = strings.ToLower(strings.TrimSpace(cfg.Ambiente))
		}
		apiBaseURL = strings.TrimSpace(cfg.APIBaseURL)
		camposPaisJSON = strings.TrimSpace(cfg.CamposPaisJSON)
		paisCodigo = strings.ToUpper(strings.TrimSpace(cfg.PaisCodigo))
	}

	if ambiente != "produccion" {
		return facturacionProveedorDispatchResult{Success: false, Error: "integracion fiscal no aplica fuera de produccion"}
	}

	camposPais := facturacionTryParseJSONMap(camposPaisJSON)
	if facturacionAnyToBool(camposPais["force_fail"]) || facturacionAnyToBool(camposPais["simular_error"]) {
		return facturacionProveedorDispatchResult{Success: false, Error: "simulacion de fallo de proveedor fiscal"}
	}

	referenciaLocal := fmt.Sprintf("%s-%d-%s", strings.ToUpper(proveedor), doc.EmpresaID, strings.ToUpper(strings.TrimSpace(doc.DocumentoCodigo)))
	if proveedor == "manual" || proveedor == "interno" || proveedor == "local" {
		if paisCodigo == "CO" {
			return facturacionProveedorDispatchResult{
				Success: false,
				Error:   "proveedor DIAN real no configurado para Colombia en produccion",
			}
		}
		respuesta := map[string]interface{}{
			"ok":                 true,
			"provider":           proveedor,
			"ambiente":           ambiente,
			"referencia_externa": referenciaLocal,
			"accion":             strings.ToLower(strings.TrimSpace(accion)),
			"documento_codigo":   strings.TrimSpace(doc.DocumentoCodigo),
			"numero_legal":       strings.TrimSpace(doc.NumeroLegal),
			"codigo_validacion":  strings.TrimSpace(doc.CodigoValidacion),
			"timestamp":          facturacionNowLocal(),
		}
		raw, _ := json.Marshal(respuesta)
		return facturacionProveedorDispatchResult{
			Success:           true,
			ReferenciaExterna: referenciaLocal,
			RespuestaJSON:     string(raw),
		}
	}

	if strings.HasPrefix(strings.ToLower(apiBaseURL), "mock://") {
		if strings.Contains(strings.ToLower(apiBaseURL), "ok") {
			respuesta := map[string]interface{}{
				"ok":                 true,
				"provider":           proveedor,
				"referencia_externa": referenciaLocal,
				"modo":               "mock",
			}
			raw, _ := json.Marshal(respuesta)
			return facturacionProveedorDispatchResult{Success: true, ReferenciaExterna: referenciaLocal, RespuestaJSON: string(raw)}
		}
		return facturacionProveedorDispatchResult{Success: false, Error: "proveedor fiscal mock en estado de error", ConnectivityFailure: true}
	}

	if apiBaseURL == "" {
		return facturacionProveedorDispatchResult{Success: false, Error: "api_base_url no configurado para proveedor fiscal"}
	}

	endpoint := strings.TrimRight(apiBaseURL, "/")
	payloadReq := map[string]interface{}{
		"empresa_id":        doc.EmpresaID,
		"accion":            strings.ToLower(strings.TrimSpace(accion)),
		"tipo_documento":    strings.TrimSpace(doc.TipoDocumento),
		"documento_codigo":  strings.TrimSpace(doc.DocumentoCodigo),
		"numero_legal":      strings.TrimSpace(doc.NumeroLegal),
		"codigo_validacion": strings.TrimSpace(doc.CodigoValidacion),
		"pais_codigo":       strings.ToUpper(strings.TrimSpace(facturacionFirstNonBlank(doc.PaisCodigo, payload.PaisCodigo))),
		"ambiente":          ambiente,
		"monto_total":       doc.MontoTotal,
		"moneda":            strings.ToUpper(strings.TrimSpace(facturacionFirstNonBlank(doc.Moneda, payload.Moneda))),
		"periodo_contable":  strings.TrimSpace(facturacionFirstNonBlank(doc.PeriodoContable, payload.PeriodoContable)),
		"campos_pais":       camposPais,
	}
	if m, _ := payloadReq["moneda"].(string); strings.TrimSpace(m) == "" {
		def := "COP"
		if cfg != nil {
			if mc := strings.TrimSpace(cfg.MonedaCodigo); mc != "" {
				def = strings.ToUpper(mc)
			} else {
				switch strings.ToUpper(strings.TrimSpace(cfg.PaisCodigo)) {
				case "EC":
					def = "USD"
				case "PA":
					def = "PAB"
				case "CO":
					def = "COP"
				}
			}
		}
		payloadReq["moneda"] = def
	}

	return dispatchFacturacionProveedorHTTP(endpoint, payloadReq)
}

func facturacionProveedorConnectionStatus(cfg *dbpkg.FacturacionElectronicaPaisConfig) map[string]interface{} {
	settings := facturacionDianOfflineSettingsFromConfig(cfg)
	out := map[string]interface{}{
		"ok":                            true,
		"online":                        false,
		"estado_conexion":               "sin_configuracion",
		"mensaje":                       "configuracion FE no disponible",
		"modo_offline_dian_activo":      settings.Enabled,
		"modo_offline_preguntar":        settings.AskBeforeContinue,
		"modo_offline_auto_reintentar":  settings.AutoRetry,
		"dian_contingencia_tipo":        settings.ContingencyType,
		"accion_recomendada":            "bloquear_facturacion_electronica",
		"requiere_confirmacion_offline": false,
	}
	if cfg == nil {
		return out
	}

	paisCodigo := strings.ToUpper(strings.TrimSpace(cfg.PaisCodigo))
	ambiente := strings.ToLower(strings.TrimSpace(cfg.Ambiente))
	proveedor := strings.ToLower(strings.TrimSpace(cfg.Proveedor))
	apiBaseURL := strings.TrimSpace(cfg.APIBaseURL)
	out["pais_codigo"] = paisCodigo
	out["proveedor"] = strings.TrimSpace(cfg.Proveedor)
	out["ambiente"] = ambiente

	if paisCodigo != "CO" {
		out["online"] = true
		out["estado_conexion"] = "no_aplica"
		out["mensaje"] = "validacion de conexion DIAN solo aplica para Colombia"
		out["accion_recomendada"] = "continuar_online"
		return out
	}
	if ambiente != "produccion" || strings.ToLower(strings.TrimSpace(cfg.Estado)) == "inactivo" {
		out["online"] = true
		out["estado_conexion"] = "no_aplica"
		out["mensaje"] = "la integracion DIAN no aplica fuera de produccion o esta inactiva"
		out["accion_recomendada"] = "continuar_online"
		return out
	}
	if proveedor == "" || proveedor == "manual" || proveedor == "interno" || proveedor == "local" {
		out["online"] = false
		out["estado_conexion"] = "sin_proveedor_dian"
		out["mensaje"] = "proveedor DIAN real no configurado para Colombia en produccion"
		out["accion_recomendada"] = "bloquear_facturacion_electronica"
		return out
	}
	if strings.HasPrefix(strings.ToLower(apiBaseURL), "mock://") {
		if strings.Contains(strings.ToLower(apiBaseURL), "ok") {
			out["online"] = true
			out["estado_conexion"] = "online"
			out["mensaje"] = "proveedor mock disponible"
			out["accion_recomendada"] = "continuar_online"
			return out
		}
		out["estado_conexion"] = "offline"
		out["mensaje"] = "proveedor mock en estado de error"
	} else if apiBaseURL == "" {
		out["estado_conexion"] = "sin_endpoint"
		out["mensaje"] = "api_base_url no configurado para proveedor DIAN"
	} else {
		endpoint := strings.TrimRight(apiBaseURL, "/")
		client := &http.Client{Timeout: 4 * time.Second}
		req, err := http.NewRequest(http.MethodHead, endpoint, nil)
		if err != nil {
			out["estado_conexion"] = "sin_endpoint"
			out["mensaje"] = "api_base_url invalido para proveedor DIAN"
		} else {
			req.Header.Set("Accept", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					out["mensaje"] = "timeout al detectar servidor DIAN/proveedor"
				} else {
					out["mensaje"] = "no se detecta internet o servidor DIAN/proveedor"
				}
				out["estado_conexion"] = "offline"
			} else {
				defer resp.Body.Close()
				out["http_status"] = resp.StatusCode
				if resp.StatusCode < 500 || resp.StatusCode == http.StatusMethodNotAllowed {
					out["online"] = true
					out["estado_conexion"] = "online"
					out["mensaje"] = "servidor DIAN/proveedor detectado"
					out["accion_recomendada"] = "continuar_online"
					return out
				}
				out["estado_conexion"] = "offline"
				out["mensaje"] = fmt.Sprintf("servidor DIAN/proveedor respondio HTTP %d", resp.StatusCode)
			}
		}
	}

	out["accion_recomendada"] = "bloquear_facturacion_electronica"
	return out
}

func facturacionDIANConnectionStatus(dbEmp *sql.DB, empresaID int64, paisCodigo string, cfg *dbpkg.FacturacionElectronicaPaisConfig) map[string]interface{} {
	status := facturacionProveedorConnectionStatus(cfg)
	if strings.ToUpper(strings.TrimSpace(paisCodigo)) != "CO" || empresaID <= 0 || dbEmp == nil {
		return status
	}

	dianCfg, err := getEmpresaDIANConfig(dbEmp, empresaID)
	if err != nil || len(dianCfg) == 0 {
		return status
	}
	endpoint := normalizeDIANSOAPEndpoint(genericStringValue(dianCfg["url_dian"]))
	if endpoint == "" {
		return status
	}

	httpStatus, reachable, latencyMS, message := runIntegracionProbe(endpoint)
	status["ok"] = true
	status["online"] = reachable
	status["estado_conexion"] = map[bool]string{true: "online", false: "offline"}[reachable]
	status["mensaje"] = message
	status["endpoint"] = endpoint
	status["http_status"] = httpStatus
	status["latency_ms"] = latencyMS
	status["proveedor"] = "DIAN"
	status["transporte"] = "soap_dian"
	status["ambiente"] = chooseDIANAmbiente(dianCfg)
	status["estado_dian"] = genericStringValue(dianCfg["estado_dian"])
	status["test_set_id_configurado"] = strings.TrimSpace(genericStringValue(dianCfg["test_set_id"])) != ""
	if reachable {
		status["accion_recomendada"] = "continuar_online"
	} else {
		status["accion_recomendada"] = "revisar_endpoint_dian"
	}
	return status
}

func facturacionOfflineDianPreflight(dbEmp *sql.DB, payload facturacionOperacionPayload) (map[string]interface{}, error) {
	if dbEmp == nil || payload.EmpresaID <= 0 {
		return nil, nil
	}
	paisCodigo := strings.ToUpper(strings.TrimSpace(payload.PaisCodigo))
	if paisCodigo == "" {
		paisDetectado, _, detectErr := dbpkg.DetectFacturacionPais(dbEmp, payload.EmpresaID, "", "")
		if detectErr == nil {
			paisCodigo = strings.ToUpper(strings.TrimSpace(paisDetectado.Codigo))
		}
	}
	if paisCodigo == "" {
		paisCodigo = "CO"
	}
	if paisCodigo != "CO" {
		return nil, nil
	}
	cfg, err := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, payload.EmpresaID, paisCodigo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	status := facturacionProveedorConnectionStatus(cfg)
	online, _ := status["online"].(bool)
	if online {
		return nil, nil
	}
	status["ok"] = false
	status["bloqueado"] = true
	status["requiere_confirmacion_offline"] = false
	status["modo_offline_dian_activo"] = false
	status["error"] = "DIAN/proveedor no disponible; se requiere conexion activa para facturar"
	return status, nil
}

func processFacturacionIntegracionForDocumento(dbEmp *sql.DB, payload facturacionOperacionPayload, doc dbpkg.EmpresaDocumentoFacturacion, accion, usuario string) (facturacionIntegracionResultado, *dbpkg.FacturacionElectronicaRetryItem, error) {
	resultado := facturacionIntegracionResultado{
		Aplica:             false,
		Accion:             strings.ToLower(strings.TrimSpace(accion)),
		EstadoEnvio:        "no_aplica",
		ContingenciaActiva: false,
		MaxIntentos:        5,
	}

	if dbEmp == nil {
		resultado.Error = "conexion de base de datos no disponible"
		return resultado, nil, fmt.Errorf("base de datos de empresa no disponible")
	}

	if doc.EmpresaID <= 0 {
		doc.EmpresaID = payload.EmpresaID
	}
	if doc.EmpresaID <= 0 {
		resultado.Error = "empresa_id es obligatorio"
		return resultado, nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(doc.DocumentoCodigo) == "" {
		doc.DocumentoCodigo = strings.TrimSpace(payload.DocumentoCodigo)
	}
	if strings.TrimSpace(doc.DocumentoCodigo) == "" {
		resultado.Error = "documento_codigo es obligatorio"
		return resultado, nil, fmt.Errorf("documento_codigo es obligatorio")
	}
	if strings.TrimSpace(doc.TipoDocumento) == "" {
		doc.TipoDocumento = strings.TrimSpace(payload.TipoDocumento)
	}
	if strings.TrimSpace(doc.TipoDocumento) == "" {
		doc.TipoDocumento = "factura_electronica"
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema_facturacion"
	}

	paisCodigo := strings.ToUpper(strings.TrimSpace(facturacionFirstNonBlank(doc.PaisCodigo, payload.PaisCodigo)))
	if paisCodigo == "" {
		paisDetectado, _, detectErr := dbpkg.DetectFacturacionPais(dbEmp, doc.EmpresaID, "", "")
		if detectErr == nil {
			paisCodigo = strings.ToUpper(strings.TrimSpace(paisDetectado.Codigo))
		}
	}
	if paisCodigo == "" {
		paisCodigo = "CO"
	}
	resultado.PaisCodigo = paisCodigo

	cfg, cfgErr := dbpkg.GetFacturacionElectronicaPaisConfig(dbEmp, doc.EmpresaID, paisCodigo)
	if cfgErr != nil && !errors.Is(cfgErr, sql.ErrNoRows) {
		resultado.Error = "no se pudo cargar configuracion FE"
		return resultado, nil, cfgErr
	}
	if cfg == nil {
		resultado.Error = "configuracion FE no disponible"
		return resultado, nil, nil
	}
	offlineSettings := facturacionDianOfflineSettingsFromConfig(cfg)
	offlineAplicaDIAN := paisCodigo == "CO"
	if offlineAplicaDIAN {
		resultado.OfflineDisponible = offlineSettings.Enabled
		resultado.OfflineConfirmado = false
		resultado.ConexionEstado = "online"
	}

	resultado.Proveedor = strings.TrimSpace(cfg.Proveedor)
	if resultado.Proveedor == "" {
		resultado.Proveedor = "manual"
	}
	resultado.Ambiente = strings.ToLower(strings.TrimSpace(cfg.Ambiente))
	if resultado.Ambiente != "produccion" {
		resultado.Ambiente = "sandbox"
	}

	retryActual, retryErr := dbpkg.GetFacturacionElectronicaRetryByDocumento(dbEmp, doc.EmpresaID, doc.TipoDocumento, doc.DocumentoCodigo)
	if retryErr != nil && !errors.Is(retryErr, sql.ErrNoRows) {
		resultado.Error = "no se pudo consultar cola de reintentos FE"
		return resultado, nil, retryErr
	}

	retryPayload := dbpkg.FacturacionElectronicaRetryItem{
		EmpresaID:         doc.EmpresaID,
		TipoDocumento:     doc.TipoDocumento,
		DocumentoCodigo:   doc.DocumentoCodigo,
		PaisCodigo:        paisCodigo,
		Proveedor:         resultado.Proveedor,
		Ambiente:          resultado.Ambiente,
		MaxIntentos:       5,
		NumeroLegal:       strings.TrimSpace(doc.NumeroLegal),
		CodigoValidacion:  strings.TrimSpace(doc.CodigoValidacion),
		FechaEmisionLegal: strings.TrimSpace(doc.FechaDocumento),
		UsuarioCreador:    strings.TrimSpace(usuario),
		Estado:            "activo",
		Observaciones:     strings.TrimSpace(doc.Observaciones),
	}
	if retryActual != nil {
		retryPayload.ID = retryActual.ID
		retryPayload.Intentos = retryActual.Intentos
		retryPayload.MaxIntentos = retryActual.MaxIntentos
		retryPayload.ProximoIntento = strings.TrimSpace(retryActual.ProximoIntento)
		retryPayload.ReferenciaExterna = strings.TrimSpace(retryActual.ReferenciaExterna)
		retryPayload.FechaContingencia = strings.TrimSpace(retryActual.FechaContingencia)
		retryPayload.ContingenciaActiva = retryActual.ContingenciaActiva
	}
	if retryPayload.MaxIntentos <= 0 {
		retryPayload.MaxIntentos = 5
	}
	resultado.Intentos = retryPayload.Intentos
	resultado.MaxIntentos = retryPayload.MaxIntentos

	aplicaIntegracion := resultado.Ambiente == "produccion" && strings.ToLower(strings.TrimSpace(cfg.Estado)) != "inactivo"
	if !aplicaIntegracion {
		retryPayload.EstadoEnvio = "no_aplica"
		retryPayload.Estado = "inactivo"
		retryPayload.ProximoIntento = ""
		retryPayload.FechaUltimoIntento = facturacionNowLocal()
		retryPayload.UltimoError = ""
		retryPayload.RespuestaProveedor = ""
		retryPayload.ContingenciaActiva = false
		retryPayload.FechaContingencia = ""
		retryPayload.Observaciones = strings.TrimSpace(facturacionFirstNonBlank(retryPayload.Observaciones, "integracion no aplica para ambiente/configuracion actual"))

		persistido, err := dbpkg.UpsertFacturacionElectronicaRetry(dbEmp, retryPayload)
		if err != nil {
			resultado.Error = "no se pudo actualizar cola FE no_aplica"
			return resultado, nil, err
		}
		resultado.EstadoEnvio = "no_aplica"
		resultado.Intentos = persistido.Intentos
		resultado.MaxIntentos = persistido.MaxIntentos
		resultado.ProximoIntento = strings.TrimSpace(persistido.ProximoIntento)
		resultado.ContingenciaActiva = persistido.ContingenciaActiva
		resultado.ReferenciaExterna = strings.TrimSpace(persistido.ReferenciaExterna)
		return resultado, persistido, nil
	}

	resultado.Aplica = true
	dispatch := dispatchFacturacionProveedor(cfg, payload, doc, accion)
	now := facturacionNowLocal()
	retryPayload.Intentos = retryPayload.Intentos + 1
	retryPayload.FechaUltimoIntento = now
	retryPayload.RespuestaProveedor = strings.TrimSpace(dispatch.RespuestaJSON)
	retryPayload.UsuarioCreador = strings.TrimSpace(usuario)
	retryPayload.Estado = "activo"
	resultado.Intentos = retryPayload.Intentos
	resultado.MaxIntentos = retryPayload.MaxIntentos

	if dispatch.Success {
		retryPayload.EstadoEnvio = "enviado"
		retryPayload.ProximoIntento = ""
		retryPayload.UltimoError = ""
		retryPayload.ContingenciaActiva = false
		retryPayload.FechaContingencia = ""
		retryPayload.ReferenciaExterna = strings.TrimSpace(dispatch.ReferenciaExterna)
		if retryPayload.ReferenciaExterna == "" {
			retryPayload.ReferenciaExterna = fmt.Sprintf("EXT-%d", time.Now().UnixNano())
		}
		resultado.EstadoEnvio = "enviado"
		resultado.ReferenciaExterna = retryPayload.ReferenciaExterna
		resultado.Error = ""
		resultado.ConexionEstado = "online"
		resultado.ConexionMensaje = "servidor DIAN/proveedor disponible"
	} else {
		retryPayload.UltimoError = strings.TrimSpace(dispatch.Error)
		if retryPayload.UltimoError == "" {
			retryPayload.UltimoError = "fallo de integracion fiscal"
		}
		if offlineAplicaDIAN && dispatch.ConnectivityFailure {
			resultado.ConexionEstado = "offline"
			resultado.ConexionMensaje = facturacionConnectivityMessage(dispatch.Error)
			retryPayload.EstadoEnvio = "fallido"
			retryPayload.ContingenciaActiva = false
			retryPayload.FechaContingencia = ""
			retryPayload.ProximoIntento = facturacionNextRetryAt(retryPayload.Intentos)
			retryPayload.UltimoError = "No hay conexion activa con DIAN/proveedor; la facturacion electronica no puede continuar"
			resultado.EstadoEnvio = "fallido"
			resultado.ProximoIntento = retryPayload.ProximoIntento
			resultado.RequiereConfirmacionOffline = false
			resultado.AccionRecomendada = "bloquear_facturacion_electronica"
		} else if retryPayload.Intentos >= retryPayload.MaxIntentos {
			retryPayload.EstadoEnvio = "fallido"
			retryPayload.ContingenciaActiva = false
			retryPayload.FechaContingencia = ""
			retryPayload.ProximoIntento = ""
			resultado.EstadoEnvio = "fallido"
			resultado.ContingenciaActiva = false
		} else {
			retryPayload.EstadoEnvio = "fallido"
			retryPayload.ContingenciaActiva = false
			retryPayload.FechaContingencia = ""
			retryPayload.ProximoIntento = facturacionNextRetryAt(retryPayload.Intentos)
			resultado.EstadoEnvio = "fallido"
			resultado.ProximoIntento = retryPayload.ProximoIntento
		}
		resultado.Error = retryPayload.UltimoError
	}

	persistido, err := dbpkg.UpsertFacturacionElectronicaRetry(dbEmp, retryPayload)
	if err != nil {
		resultado.Error = "no se pudo persistir estado de integracion FE"
		return resultado, nil, err
	}

	resultado.EstadoEnvio = normalizeFacturacionEstadoEnvio(persistido.EstadoEnvio)
	resultado.Intentos = persistido.Intentos
	resultado.MaxIntentos = persistido.MaxIntentos
	resultado.ProximoIntento = strings.TrimSpace(persistido.ProximoIntento)
	resultado.ContingenciaActiva = persistido.ContingenciaActiva
	if strings.TrimSpace(resultado.ReferenciaExterna) == "" {
		resultado.ReferenciaExterna = strings.TrimSpace(persistido.ReferenciaExterna)
	}
	if strings.TrimSpace(resultado.Error) == "" {
		resultado.Error = strings.TrimSpace(persistido.UltimoError)
	}

	return resultado, persistido, nil
}

func facturacionBuildOperacionPayloadFromDocumento(doc dbpkg.EmpresaDocumentoFacturacion) facturacionOperacionPayload {
	return facturacionOperacionPayload{
		EmpresaID:       doc.EmpresaID,
		EntidadID:       doc.EntidadRelacionadaID,
		ClienteID:       doc.EntidadRelacionadaID,
		TipoDocumento:   strings.TrimSpace(doc.TipoDocumento),
		PaisCodigo:      strings.TrimSpace(doc.PaisCodigo),
		DocumentoCodigo: strings.TrimSpace(doc.DocumentoCodigo),
		EstadoActual:    strings.TrimSpace(doc.EstadoDocumento),
		MontoTotal:      doc.MontoTotal,
		Moneda:          strings.TrimSpace(doc.Moneda),
		PeriodoContable: strings.TrimSpace(doc.PeriodoContable),
		Observaciones:   strings.TrimSpace(doc.Observaciones),
	}
}

func facturacionDeriveAccionByDocumento(doc dbpkg.EmpresaDocumentoFacturacion) string {
	tipo := strings.ToLower(strings.TrimSpace(doc.TipoDocumento))
	estado := strings.ToLower(strings.TrimSpace(doc.EstadoDocumento))
	switch normalizeFacturacionDocumentoElectronicoTipo(tipo) {
	case "nota_credito":
		return "nota_credito"
	case "nota_debito":
		return "nota_debito"
	case "documento_soporte":
		return "documento_soporte"
	case "nomina_electronica":
		return "nomina_electronica"
	case "documento_equivalente_pos":
		return "documento_equivalente_pos"
	default:
		if facturacionDocumentoElectronicoPermitido(tipo) {
			return normalizeFacturacionDocumentoElectronicoTipo(tipo)
		}
	}
	if estado == "anulada" {
		return "anular"
	}
	return "emitir"
}

func processFacturacionRetryQueue(dbEmp *sql.DB, empresaID int64, limit int, usuario string) (map[string]interface{}, error) {
	if dbEmp == nil {
		return nil, fmt.Errorf("base de datos de empresa no disponible")
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema_facturacion"
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	items, err := dbpkg.ListFacturacionElectronicaRetriesByEmpresa(dbEmp, empresaID, dbpkg.FacturacionElectronicaRetryFilter{
		SoloVencidos:    true,
		IncludeInactive: false,
		Limit:           limit,
		Offset:          0,
	})
	if err != nil {
		return nil, err
	}

	resumenItems := make([]map[string]interface{}, 0, len(items))
	procesados := 0
	enviados := 0
	fallidos := 0
	contingencia := 0
	noAplica := 0
	erroresInternos := 0

	for _, retryItem := range items {
		detail := map[string]interface{}{
			"tipo_documento":   retryItem.TipoDocumento,
			"documento_codigo": retryItem.DocumentoCodigo,
			"estado_anterior":  retryItem.EstadoEnvio,
		}

		doc, docErr := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, empresaID, retryItem.TipoDocumento, retryItem.DocumentoCodigo)
		if docErr != nil {
			if errors.Is(docErr, sql.ErrNoRows) {
				actualizado := retryItem
				if actualizado.MaxIntentos <= 0 {
					actualizado.MaxIntentos = 5
				}
				actualizado.Intentos = actualizado.Intentos + 1
				actualizado.FechaUltimoIntento = facturacionNowLocal()
				actualizado.UltimoError = "documento transaccional no encontrado para reintento"
				actualizado.UsuarioCreador = usuario
				actualizado.Estado = "activo"
				if actualizado.Intentos >= actualizado.MaxIntentos {
					actualizado.EstadoEnvio = "fallido"
					actualizado.ContingenciaActiva = false
					actualizado.FechaContingencia = ""
					actualizado.ProximoIntento = ""
				} else {
					actualizado.EstadoEnvio = "fallido"
					actualizado.ContingenciaActiva = false
					actualizado.FechaContingencia = ""
					actualizado.ProximoIntento = facturacionNextRetryAt(actualizado.Intentos)
				}
				persistido, upErr := dbpkg.UpsertFacturacionElectronicaRetry(dbEmp, actualizado)
				if upErr != nil {
					erroresInternos += 1
					detail["error"] = "no se pudo actualizar retry para documento inexistente"
				} else {
					detail["estado_nuevo"] = persistido.EstadoEnvio
					detail["intentos"] = persistido.Intentos
					detail["ultimo_error"] = persistido.UltimoError
					fallidos += 1
					procesados += 1
				}
				resumenItems = append(resumenItems, detail)
				continue
			}
			erroresInternos += 1
			detail["error"] = "no se pudo consultar documento para reintento"
			resumenItems = append(resumenItems, detail)
			continue
		}

		payload := facturacionBuildOperacionPayloadFromDocumento(*doc)
		accion := facturacionDeriveAccionByDocumento(*doc)
		resultado, persistido, procErr := processFacturacionIntegracionForDocumento(dbEmp, payload, *doc, accion, usuario)
		if procErr != nil {
			erroresInternos += 1
			detail["error"] = procErr.Error()
		} else {
			detail["estado_nuevo"] = resultado.EstadoEnvio
			detail["intentos"] = resultado.Intentos
			detail["max_intentos"] = resultado.MaxIntentos
			detail["referencia_externa"] = resultado.ReferenciaExterna
			detail["proximo_intento"] = resultado.ProximoIntento
			detail["contingencia_activa"] = resultado.ContingenciaActiva
			detail["error_integracion"] = resultado.Error
			if persistido != nil {
				detail["cola_reintentos"] = persistido
			}

			switch resultado.EstadoEnvio {
			case "enviado", "reconciliado":
				enviados += 1
			case "contingencia":
				contingencia += 1
			case "no_aplica":
				noAplica += 1
			default:
				fallidos += 1
			}
			procesados += 1
		}
		resumenItems = append(resumenItems, detail)
	}

	return map[string]interface{}{
		"ok":               true,
		"empresa_id":       empresaID,
		"limit":            limit,
		"en_cola":          len(items),
		"procesados":       procesados,
		"enviados":         enviados,
		"fallidos":         fallidos,
		"contingencia":     contingencia,
		"no_aplica":        noAplica,
		"errores_internos": erroresInternos,
		"items":            resumenItems,
	}, nil
}

func buildFacturacionReconciliacion(dbEmp *sql.DB, empresaID int64) (map[string]interface{}, error) {
	return reconcileFacturacionEstados(dbEmp, empresaID, false, "")
}

func listFacturacionDocumentosForReconciliacion(dbEmp *sql.DB, empresaID int64) ([]dbpkg.EmpresaDocumentoFacturacionListado, error) {
	documentos, err := dbpkg.ListEmpresaDocumentosFacturacionByEmpresa(dbEmp, dbpkg.EmpresaDocumentoFacturacionListFilter{
		EmpresaID:       empresaID,
		IncludeInactive: false,
		Limit:           1000,
		Offset:          0,
	})
	if err == nil {
		return documentos, nil
	}

	errMsg := strings.ToLower(strings.TrimSpace(err.Error()))
	if !strings.Contains(errMsg, "no such table: clientes") {
		return nil, err
	}

	rows, qErr := dbEmp.Query(`SELECT
		id,
		empresa_id,
		COALESCE(tipo_documento, 'factura_electronica'),
		COALESCE(documento_codigo, ''),
		COALESCE(numero_legal, ''),
		COALESCE(codigo_validacion, ''),
		COALESCE(pais_codigo, ''),
		COALESCE(ambiente_fe, ''),
		COALESCE(estado_documento, 'borrador'),
		COALESCE(estado_anterior, ''),
		COALESCE(evento_ultimo, ''),
		COALESCE(periodo_contable, ''),
		COALESCE(monto_total, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(fecha_documento, ''),
		COALESCE(entidad_relacionada_id, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_facturacion_documentos
	WHERE empresa_id = ? AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY pcs_ts(COALESCE(NULLIF(fecha_documento, ''), fecha_creacion)) DESC, id DESC
	LIMIT 1000`, empresaID)
	if qErr != nil {
		return nil, qErr
	}
	defer rows.Close()

	out := make([]dbpkg.EmpresaDocumentoFacturacionListado, 0)
	for rows.Next() {
		var item dbpkg.EmpresaDocumentoFacturacionListado
		if scanErr := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.TipoDocumento,
			&item.DocumentoCodigo,
			&item.NumeroLegal,
			&item.CodigoValidacion,
			&item.PaisCodigo,
			&item.AmbienteFE,
			&item.EstadoDocumento,
			&item.EstadoAnterior,
			&item.EventoUltimo,
			&item.PeriodoContable,
			&item.MontoTotal,
			&item.Moneda,
			&item.FechaDocumento,
			&item.EntidadRelacionadaID,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); scanErr != nil {
			return nil, scanErr
		}
		out = append(out, item)
	}

	return out, rows.Err()
}

func reconcileFacturacionEstados(dbEmp *sql.DB, empresaID int64, aplicar bool, usuario string) (map[string]interface{}, error) {
	if dbEmp == nil {
		return nil, fmt.Errorf("base de datos de empresa no disponible")
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	if strings.TrimSpace(usuario) == "" {
		usuario = "sistema_facturacion"
	}

	documentos, err := listFacturacionDocumentosForReconciliacion(dbEmp, empresaID)
	if err != nil {
		return nil, err
	}

	inconsistencias := make([]map[string]interface{}, 0)
	documentosEvaluados := 0
	conciliados := 0
	pendientes := 0
	noAplica := 0
	procesados := 0
	enviados := 0
	fallidos := 0
	contingencia := 0
	erroresInternos := 0

	for _, doc := range documentos {
		tipo := strings.ToLower(strings.TrimSpace(doc.TipoDocumento))
		estadoDocumento := strings.ToLower(strings.TrimSpace(doc.EstadoDocumento))
		if !facturacionDocumentoElectronicoPermitido(tipo) {
			continue
		}
		if estadoDocumento != "emitida" && estadoDocumento != "anulada" {
			continue
		}

		documentosEvaluados += 1
		retryItem, retryErr := dbpkg.GetFacturacionElectronicaRetryByDocumento(dbEmp, empresaID, doc.TipoDocumento, doc.DocumentoCodigo)
		if retryErr != nil && !errors.Is(retryErr, sql.ErrNoRows) {
			erroresInternos += 1
			inconsistencias = append(inconsistencias, map[string]interface{}{
				"tipo_documento":   doc.TipoDocumento,
				"documento_codigo": doc.DocumentoCodigo,
				"problema":         "error_consulta_retry",
				"detalle":          retryErr.Error(),
			})
			continue
		}

		estadoRetry := "sin_cola"
		if retryItem != nil {
			estadoRetry = normalizeFacturacionEstadoEnvio(retryItem.EstadoEnvio)
		}

		if estadoRetry == "enviado" || estadoRetry == "reconciliado" {
			conciliados += 1
			continue
		}
		if estadoRetry == "no_aplica" {
			noAplica += 1
			continue
		}

		if strings.ToLower(strings.TrimSpace(doc.AmbienteFE)) == "sandbox" && retryItem == nil {
			noAplica += 1
			continue
		}

		pendientes += 1
		item := map[string]interface{}{
			"tipo_documento":   doc.TipoDocumento,
			"documento_codigo": doc.DocumentoCodigo,
			"estado_documento": doc.EstadoDocumento,
			"estado_retry":     estadoRetry,
			"pais_codigo":      doc.PaisCodigo,
			"ambiente_fe":      doc.AmbienteFE,
		}

		if aplicar {
			payload := facturacionBuildOperacionPayloadFromDocumento(doc.EmpresaDocumentoFacturacion)
			accion := facturacionDeriveAccionByDocumento(doc.EmpresaDocumentoFacturacion)
			resultado, persistido, procErr := processFacturacionIntegracionForDocumento(dbEmp, payload, doc.EmpresaDocumentoFacturacion, accion, usuario)
			if procErr != nil {
				erroresInternos += 1
				item["error"] = procErr.Error()
			} else {
				item["estado_reconciliado"] = resultado.EstadoEnvio
				item["intentos"] = resultado.Intentos
				item["max_intentos"] = resultado.MaxIntentos
				item["proximo_intento"] = resultado.ProximoIntento
				item["contingencia_activa"] = resultado.ContingenciaActiva
				item["referencia_externa"] = resultado.ReferenciaExterna
				item["error_integracion"] = resultado.Error
				if persistido != nil {
					item["cola_reintentos"] = persistido
				}

				switch resultado.EstadoEnvio {
				case "enviado", "reconciliado":
					enviados += 1
				case "contingencia":
					contingencia += 1
				case "no_aplica":
					noAplica += 1
				default:
					fallidos += 1
				}
				procesados += 1
			}
		}

		inconsistencias = append(inconsistencias, item)
	}

	return map[string]interface{}{
		"ok":                        true,
		"empresa_id":                empresaID,
		"aplicar":                   aplicar,
		"timestamp":                 facturacionNowLocal(),
		"documentos_evaluados":      documentosEvaluados,
		"documentos_conciliados":    conciliados,
		"pendientes_reconciliacion": pendientes,
		"documentos_no_aplica":      noAplica,
		"procesados":                procesados,
		"enviados":                  enviados,
		"fallidos":                  fallidos,
		"contingencia":              contingencia,
		"errores_internos":          erroresInternos,
		"inconsistencias":           inconsistencias,
	}, nil
}
