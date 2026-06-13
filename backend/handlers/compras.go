package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// EmpresaComprasDocumentosHandler gestiona ciclo documental general de compras (orden, recepcion y contabilizacion).
func EmpresaComprasDocumentosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tipoDocumento := strings.TrimSpace(r.URL.Query().Get("tipo_documento"))
			estadoDocumento := strings.TrimSpace(r.URL.Query().Get("estado_documento"))
			proveedorID, err := parseInt64QueryOptional(r, "proveedor_id")
			if err != nil {
				http.Error(w, "proveedor_id invalido", http.StatusBadRequest)
				return
			}
			includeInactive := parseBoolQuery(r, "include_inactive")
			q := strings.TrimSpace(r.URL.Query().Get("q"))
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

			rows, err := dbpkg.ListEmpresaDocumentosCompraByEmpresa(dbEmp, empresaID, tipoDocumento, proveedorID, estadoDocumento, includeInactive, q, limit, offset)
			if err != nil {
				http.Error(w, "No se pudo listar documentos de compras", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(rows)
			return

		case http.MethodPost:
			var payload struct {
				EmpresaID            int64   `json:"empresa_id"`
				ProveedorID          int64   `json:"proveedor_id"`
				TipoDocumento        string  `json:"tipo_documento"`
				DocumentoCodigo      string  `json:"documento_codigo"`
				EstadoDocumento      string  `json:"estado_documento"`
				EstadoActual         string  `json:"estado_actual"`
				Accion               string  `json:"accion"`
				PeriodoContable      string  `json:"periodo_contable"`
				FormaPago            string  `json:"forma_pago"`
				MetodoPago           string  `json:"metodo_pago"`
				Subtotal             float64 `json:"subtotal"`
				BaseGravable         float64 `json:"base_gravable"`
				IVA                  float64 `json:"iva"`
				Impuestos            float64 `json:"impuestos"`
				RetencionFuente      float64 `json:"retencion_fuente"`
				RetencionICA         float64 `json:"retencion_ica"`
				RetencionIVA         float64 `json:"retencion_iva"`
				TotalRetenciones     float64 `json:"total_retenciones"`
				TotalNeto            float64 `json:"total_neto"`
				MontoTotal           float64 `json:"monto_total"`
				Moneda               string  `json:"moneda"`
				FechaDocumento       string  `json:"fecha_documento"`
				RequiereAprobacion   bool    `json:"requiere_aprobacion"`
				NivelesAprobacion    int     `json:"niveles_aprobacion_requeridos"`
				ProveedorDocRef      string  `json:"proveedor_documento_ref"`
				FacturaDocRef        string  `json:"factura_documento_ref"`
				EntradaDocRef        string  `json:"entrada_documento_ref"`
				ValidacionDocumental string  `json:"validacion_documental_estado"`
				Observaciones        string  `json:"observaciones"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			if payload.EmpresaID <= 0 {
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
			}
			if payload.EmpresaID <= 0 {
				http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
				return
			}
			if payload.ProveedorID <= 0 {
				http.Error(w, "proveedor_id es obligatorio", http.StatusBadRequest)
				return
			}

			if strings.TrimSpace(payload.DocumentoCodigo) == "" {
				payload.DocumentoCodigo = fmt.Sprintf("OC-%d-%s", payload.ProveedorID, time.Now().Format("20060102150405"))
			}
			if strings.TrimSpace(payload.PeriodoContable) == "" {
				payload.PeriodoContable = time.Now().Format("2006-01")
			}
			if strings.TrimSpace(payload.FechaDocumento) == "" {
				payload.FechaDocumento = time.Now().Format("2006-01-02")
			}

			action := normalizeComprasAction(payload.Accion)
			if action == "" {
				action = normalizeComprasAction(r.URL.Query().Get("action"))
			}
			if action == "" {
				action = "crear"
			}

			requiereAprobacion := payload.RequiereAprobacion || payload.NivelesAprobacion > 1
			nivelesAprobacion := payload.NivelesAprobacion
			if requiereAprobacion {
				if nivelesAprobacion <= 1 {
					nivelesAprobacion = 2
				}
			} else {
				nivelesAprobacion = 1
			}
			nivelAprobacion := 0
			aprobadoresJSON := "[]"
			recepcionDetalleJSON := ""
			recepcionResumenJSON := ""
			validacionDocumental := comprasFirstNonBlank(payload.ValidacionDocumental, "no_aplica")
			proveedorDocRef := normalizeComprasDocumentoRef(payload.ProveedorDocRef)
			facturaDocRef := normalizeComprasDocumentoRef(payload.FacturaDocRef)
			entradaDocRef := normalizeComprasDocumentoRef(payload.EntradaDocRef)

			estadoDocumento := strings.TrimSpace(payload.EstadoDocumento)
			if estadoDocumento == "" {
				estadoDocumento = "borrador"
			}
			estadoAnterior := ""
			evento := "orden_compra_creada"
			accionResp := "crear"

			switch action {
			case "crear", "guardar":
				// Sin transicion automatica adicional.
			case "solicitar_aprobacion":
				requiereAprobacion = true
				if nivelesAprobacion <= 1 {
					nivelesAprobacion = 2
				}
				estadoAnterior = comprasFirstNonBlank(payload.EstadoActual, estadoDocumento, "borrador")
				estadoDocumento = "pendiente_aprobacion"
				evento = "orden_compra_pendiente_aprobacion"
				accionResp = "solicitar_aprobacion"
			case "emitir", "emitir_orden":
				if requiereAprobacion {
					estadoAnterior = comprasFirstNonBlank(payload.EstadoActual, estadoDocumento, "borrador")
					estadoDocumento = "pendiente_aprobacion"
					evento = "orden_compra_pendiente_aprobacion"
					accionResp = "solicitar_aprobacion"
					break
				}
				transition, err := resolveComprasTransition(action, comprasFirstNonBlank(payload.EstadoActual, estadoDocumento, "borrador"))
				if err != nil {
					errLower := strings.ToLower(err.Error())
					if strings.Contains(errLower, "transicion invalida") {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				estadoDocumento = transition.EstadoNuevo
				estadoAnterior = transition.EstadoAnterior
				evento = transition.Evento
				accionResp = transition.Accion
			case "aprobar_compra", "rechazar_compra", "validar_documentos", "recepcionar_parcial_compra":
				http.Error(w, "action invalida para crear documento; use PUT para aprobar, validar o recepcionar", http.StatusBadRequest)
				return
			default:
				transition, err := resolveComprasTransition(action, comprasFirstNonBlank(payload.EstadoActual, estadoDocumento, "borrador"))
				if err != nil {
					errLower := strings.ToLower(err.Error())
					if strings.Contains(errLower, "transicion invalida") {
						http.Error(w, err.Error(), http.StatusConflict)
						return
					}
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				estadoDocumento = transition.EstadoNuevo
				estadoAnterior = transition.EstadoAnterior
				evento = transition.Evento
				accionResp = transition.Accion
			}

			docPersistido, err := dbpkg.UpsertEmpresaDocumentoCompra(dbEmp, dbpkg.EmpresaDocumentoCompra{
				EmpresaID:            payload.EmpresaID,
				ProveedorID:          payload.ProveedorID,
				TipoDocumento:        comprasFirstNonBlank(payload.TipoDocumento, "orden_compra"),
				DocumentoCodigo:      payload.DocumentoCodigo,
				EstadoDocumento:      estadoDocumento,
				EstadoAnterior:       estadoAnterior,
				EventoUltimo:         evento,
				PeriodoContable:      payload.PeriodoContable,
				MontoTotal:           payload.MontoTotal,
				Moneda:               payload.Moneda,
				FechaDocumento:       payload.FechaDocumento,
				EntidadRelacionadaID: payload.ProveedorID,
				RequiereAprobacion:   requiereAprobacion,
				NivelesAprobacion:    nivelesAprobacion,
				NivelAprobacion:      nivelAprobacion,
				AprobadoresJSON:      aprobadoresJSON,
				RecepcionDetalleJSON: recepcionDetalleJSON,
				RecepcionResumenJSON: recepcionResumenJSON,
				ValidacionEstado:     validacionDocumental,
				ProveedorDocRef:      proveedorDocRef,
				FacturaDocRef:        facturaDocRef,
				EntradaDocRef:        entradaDocRef,
				UsuarioCreador:       adminEmailFromRequest(r),
				Estado:               "activo",
				Observaciones:        payload.Observaciones,
			})
			if err != nil {
				http.Error(w, "No se pudo guardar el documento de compras", http.StatusInternalServerError)
				return
			}

			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "compras",
				Evento:          evento,
				Entidad:         "orden_compra",
				EntidadID:       docPersistido.ID,
				DocumentoTipo:   docPersistido.TipoDocumento,
				DocumentoCodigo: docPersistido.DocumentoCodigo,
				PeriodoContable: docPersistido.PeriodoContable,
				MontoTotal:      docPersistido.MontoTotal,
				Moneda:          docPersistido.Moneda,
				Origen:          "api_compras_documentos",
				Observaciones:   strings.TrimSpace(payload.Observaciones),
			}, map[string]interface{}{
				"accion":           accionResp,
				"estado_anterior":  docPersistido.EstadoAnterior,
				"estado_nuevo":     docPersistido.EstadoDocumento,
				"entidad_id":       docPersistido.ID,
				"documento_codigo": docPersistido.DocumentoCodigo,
				"proveedor_id":     docPersistido.ProveedorID,
				"forma_pago":       comprasFirstNonBlank(payload.FormaPago, "credito"),
				"metodo_pago":      strings.TrimSpace(payload.MetodoPago),
				"subtotal":         comprasFirstPositive(payload.Subtotal, payload.BaseGravable, payload.MontoTotal),
				"base_gravable":    comprasFirstPositive(payload.BaseGravable, payload.Subtotal, payload.MontoTotal),
				"iva":              comprasFirstPositive(payload.IVA, payload.Impuestos),
				"impuestos":        comprasFirstPositive(payload.Impuestos, payload.IVA),
				"retencion_fuente": payload.RetencionFuente,
				"retencion_ica":    payload.RetencionICA,
				"retencion_iva":    payload.RetencionIVA,
				"total_retenciones": comprasFirstPositive(
					payload.TotalRetenciones,
					payload.RetencionFuente+payload.RetencionICA+payload.RetencionIVA,
				),
				"total_neto": payload.TotalNeto,
				"empresa_id": docPersistido.EmpresaID,
			})

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"ok":        true,
				"accion":    accionResp,
				"evento":    evento,
				"resultado": docPersistido,
			})
			return

		case http.MethodPut:
			var payload struct {
				EmpresaID          int64                  `json:"empresa_id"`
				ProveedorID        int64                  `json:"proveedor_id"`
				TipoDocumento      string                 `json:"tipo_documento"`
				DocumentoCodigo    string                 `json:"documento_codigo"`
				EstadoActual       string                 `json:"estado_actual"`
				EstadoDocumento    string                 `json:"estado_documento"`
				Accion             string                 `json:"accion"`
				PeriodoContable    string                 `json:"periodo_contable"`
				FormaPago          string                 `json:"forma_pago"`
				MetodoPago         string                 `json:"metodo_pago"`
				Subtotal           float64                `json:"subtotal"`
				BaseGravable       float64                `json:"base_gravable"`
				IVA                float64                `json:"iva"`
				Impuestos          float64                `json:"impuestos"`
				RetencionFuente    float64                `json:"retencion_fuente"`
				RetencionICA       float64                `json:"retencion_ica"`
				RetencionIVA       float64                `json:"retencion_iva"`
				TotalRetenciones   float64                `json:"total_retenciones"`
				TotalNeto          float64                `json:"total_neto"`
				MontoTotal         float64                `json:"monto_total"`
				Moneda             string                 `json:"moneda"`
				FechaDocumento     string                 `json:"fecha_documento"`
				RequiereAprobacion *bool                  `json:"requiere_aprobacion"`
				NivelesAprobacion  *int                   `json:"niveles_aprobacion_requeridos"`
				RecepcionItems     []comprasRecepcionItem `json:"recepcion_items"`
				ProveedorDocRef    string                 `json:"proveedor_documento_ref"`
				FacturaDocRef      string                 `json:"factura_documento_ref"`
				EntradaDocRef      string                 `json:"entrada_documento_ref"`
				Observaciones      string                 `json:"observaciones"`
				Activo             *bool                  `json:"activo"`
			}
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&payload)
			}

			if payload.EmpresaID <= 0 {
				empresaID, err := parseInt64QueryOptional(r, "empresa_id")
				if err != nil {
					http.Error(w, "empresa_id invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
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

			action := normalizeComprasAction(payload.Accion)
			if action == "" {
				action = normalizeComprasAction(r.URL.Query().Get("action"))
			}
			if action == "" {
				action = "actualizar"
			}

			tipoDocumento := comprasFirstNonBlank(payload.TipoDocumento, "orden_compra")

			if action == "activar" {
				estado := "activo"
				if payload.Activo != nil {
					if !*payload.Activo {
						estado = "inactivo"
					}
				} else {
					activoRaw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("activo")))
					if activoRaw == "0" || activoRaw == "false" || activoRaw == "no" {
						estado = "inactivo"
					}
				}

				if err := dbpkg.SetEmpresaDocumentoCompraEstadoByCodigo(dbEmp, payload.EmpresaID, tipoDocumento, payload.DocumentoCodigo, estado); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "documento no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo actualizar estado activo del documento", http.StatusInternalServerError)
					return
				}

				evento := "orden_compra_activada"
				if estado == "inactivo" {
					evento = "orden_compra_desactivada"
				}
				registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
					EmpresaID:       payload.EmpresaID,
					Modulo:          "compras",
					Evento:          evento,
					Entidad:         "orden_compra",
					DocumentoTipo:   tipoDocumento,
					DocumentoCodigo: strings.ToUpper(strings.TrimSpace(payload.DocumentoCodigo)),
					Origen:          "api_compras_documentos",
					Observaciones:   strings.TrimSpace(payload.Observaciones),
				}, map[string]interface{}{
					"accion":           "activar",
					"estado":           estado,
					"documento_codigo": strings.ToUpper(strings.TrimSpace(payload.DocumentoCodigo)),
					"empresa_id":       payload.EmpresaID,
				})

				w.WriteHeader(http.StatusNoContent)
				return
			}

			docActual, err := dbpkg.GetEmpresaDocumentoCompraByCodigo(dbEmp, payload.EmpresaID, tipoDocumento, payload.DocumentoCodigo)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "documento no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo consultar el documento de compras", http.StatusInternalServerError)
				return
			}

			estadoActualPayload := normalizeComprasAction(payload.EstadoActual)
			estadoActualReal := normalizeComprasAction(docActual.EstadoDocumento)
			if estadoActualPayload != "" && estadoActualReal != "" && estadoActualPayload != estadoActualReal {
				http.Error(w, fmt.Sprintf("estado_actual no coincide con el estado real del documento: esperado=%s recibido=%s", estadoActualReal, estadoActualPayload), http.StatusConflict)
				return
			}
			payload.EstadoActual = comprasFirstNonBlank(docActual.EstadoDocumento, payload.EstadoActual)
			payload.ProveedorID = comprasFirstNonBlankInt64(payload.ProveedorID, docActual.ProveedorID)

			estadoDocumento := docActual.EstadoDocumento
			estadoAnterior := docActual.EstadoAnterior
			evento := "orden_compra_actualizada"
			accionResp := "actualizar"
			usuarioActual := strings.TrimSpace(adminEmailFromRequest(r))
			if usuarioActual == "" {
				usuarioActual = "sistema"
			}

			requiereAprobacion := docActual.RequiereAprobacion
			if payload.RequiereAprobacion != nil {
				requiereAprobacion = *payload.RequiereAprobacion
			}
			nivelesAprobacion := docActual.NivelesAprobacion
			if payload.NivelesAprobacion != nil && *payload.NivelesAprobacion > 0 {
				nivelesAprobacion = *payload.NivelesAprobacion
			}
			if nivelesAprobacion <= 0 {
				nivelesAprobacion = 1
			}
			if !requiereAprobacion {
				nivelesAprobacion = 1
			}
			nivelAprobacion := docActual.NivelAprobacion
			if nivelAprobacion < 0 {
				nivelAprobacion = 0
			}
			if nivelAprobacion > nivelesAprobacion {
				nivelAprobacion = nivelesAprobacion
			}

			aprobadoresJSON := strings.TrimSpace(docActual.AprobadoresJSON)
			if aprobadoresJSON == "" {
				aprobadoresJSON = "[]"
			}
			recepcionDetalleJSON := strings.TrimSpace(docActual.RecepcionDetalleJSON)
			recepcionResumenJSON := strings.TrimSpace(docActual.RecepcionResumenJSON)
			validacionEstado := comprasFirstNonBlank(docActual.ValidacionEstado, "no_aplica")
			proveedorDocRef := normalizeComprasDocumentoRef(comprasFirstNonBlank(payload.ProveedorDocRef, docActual.ProveedorDocRef))
			facturaDocRef := normalizeComprasDocumentoRef(comprasFirstNonBlank(payload.FacturaDocRef, docActual.FacturaDocRef))
			entradaDocRef := normalizeComprasDocumentoRef(comprasFirstNonBlank(payload.EntradaDocRef, docActual.EntradaDocRef))

			validationStatus := 0
			var validationErr error
			hasRecepcionResumen := false
			recepcionResumen := comprasRecepcionResumen{}

			estadoActualDocumento := normalizeComprasAction(docActual.EstadoDocumento)

			switch action {
			case "actualizar":
				// Actualizacion simple sin transicion.

			case "solicitar_aprobacion":
				estadoActualNorm := estadoActualDocumento
				if estadoActualNorm != "borrador" && estadoActualNorm != "pendiente_emision" && estadoActualNorm != "rechazada" {
					http.Error(w, "transicion invalida: solicitar_aprobacion requiere estado borrador, pendiente_emision o rechazada", http.StatusConflict)
					return
				}
				requiereAprobacion = true
				if nivelesAprobacion <= 1 {
					nivelesAprobacion = 2
				}
				nivelAprobacion = 0
				aprobadoresJSON = "[]"
				estadoDocumento = "pendiente_aprobacion"
				estadoAnterior = payload.EstadoActual
				evento = "orden_compra_pendiente_aprobacion"
				accionResp = "solicitar_aprobacion"

			case "aprobar_compra":
				estadoActualNorm := estadoActualDocumento
				if estadoActualNorm != "pendiente_aprobacion" {
					http.Error(w, "transicion invalida: aprobar_compra requiere estado pendiente_aprobacion", http.StatusConflict)
					return
				}
				requiereAprobacion = true
				if nivelesAprobacion <= 1 {
					nivelesAprobacion = 2
				}
				if nivelAprobacion >= nivelesAprobacion {
					http.Error(w, "la orden ya completo todos los niveles de aprobacion", http.StatusConflict)
					return
				}
				nivelAprobacion++
				aprobadoresJSON = appendComprasAprobador(aprobadoresJSON, usuarioActual, payload.Observaciones, nivelAprobacion)
				estadoAnterior = payload.EstadoActual
				if nivelAprobacion >= nivelesAprobacion {
					estadoDocumento = "emitida"
					evento = "orden_compra_aprobada"
				} else {
					estadoDocumento = "pendiente_aprobacion"
					evento = "orden_compra_aprobacion_parcial"
				}
				accionResp = "aprobar_compra"

			case "rechazar_compra":
				estadoActualNorm := estadoActualDocumento
				if estadoActualNorm != "pendiente_aprobacion" {
					http.Error(w, "transicion invalida: rechazar_compra requiere estado pendiente_aprobacion", http.StatusConflict)
					return
				}
				estadoDocumento = "rechazada"
				estadoAnterior = payload.EstadoActual
				evento = "orden_compra_rechazada"
				accionResp = "rechazar_compra"

			case "recepcionar_parcial_compra":
				if estadoActualDocumento != "emitida" && estadoActualDocumento != "recepcion_parcial" {
					http.Error(w, "transicion invalida: recepcionar_parcial_compra requiere estado emitida o recepcion_parcial", http.StatusConflict)
					return
				}
				resumen, err := buildComprasRecepcionResumen(payload.RecepcionItems)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				detalleBytes, _ := json.Marshal(resumen.Items)
				resumenBytes, _ := json.Marshal(resumen)
				recepcionDetalleJSON = string(detalleBytes)
				recepcionResumenJSON = string(resumenBytes)
				hasRecepcionResumen = true
				recepcionResumen = resumen
				estadoDocumento = "recepcion_parcial"
				estadoAnterior = payload.EstadoActual
				evento = "compra_recepcion_parcial"
				accionResp = "recepcionar_parcial_compra"

			case "recepcionar_compra":
				if estadoActualDocumento != "emitida" && estadoActualDocumento != "recepcion_parcial" {
					http.Error(w, "transicion invalida: recepcionar_compra requiere estado emitida o recepcion_parcial", http.StatusConflict)
					return
				}
				if len(payload.RecepcionItems) > 0 {
					resumen, err := buildComprasRecepcionResumen(payload.RecepcionItems)
					if err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
					detalleBytes, _ := json.Marshal(resumen.Items)
					resumenBytes, _ := json.Marshal(resumen)
					recepcionDetalleJSON = string(detalleBytes)
					recepcionResumenJSON = string(resumenBytes)
					hasRecepcionResumen = true
					recepcionResumen = resumen
					if resumen.ItemsPendientes > 0 {
						estadoDocumento = "recepcion_parcial"
						estadoAnterior = payload.EstadoActual
						evento = "compra_recepcion_parcial"
						accionResp = "recepcionar_parcial_compra"
						break
					}
				} else if resumenExistente, ok := parseComprasRecepcionResumen(recepcionResumenJSON); ok {
					if resumenExistente.ItemsPendientes > 0 {
						http.Error(w, "la recepcion aun tiene items pendientes; use recepcionar_parcial_compra con diferencias por item", http.StatusConflict)
						return
					}
				}

				transition, err := resolveComprasTransition(action, payload.EstadoActual)
				if err != nil {
					errLower := strings.ToLower(err.Error())
					switch {
					case strings.Contains(errLower, "transicion invalida"):
						http.Error(w, err.Error(), http.StatusConflict)
						return
					case strings.Contains(errLower, "accion no soportada"):
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					default:
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
				}
				estadoDocumento = transition.EstadoNuevo
				estadoAnterior = transition.EstadoAnterior
				evento = transition.Evento
				accionResp = transition.Accion

			case "validar_documentos":
				normalizadoProveedorRef, normalizadoFacturaRef, normalizadoEntradaRef, statusErr, err := validarComprasDocumentos(
					dbEmp,
					payload.EmpresaID,
					payload.ProveedorID,
					proveedorDocRef,
					facturaDocRef,
					entradaDocRef,
				)
				proveedorDocRef = normalizadoProveedorRef
				facturaDocRef = normalizadoFacturaRef
				entradaDocRef = normalizadoEntradaRef
				estadoAnterior = payload.EstadoActual
				accionResp = "validar_documentos"
				if err != nil {
					if statusErr == http.StatusInternalServerError {
						http.Error(w, "No se pudo validar documentos de compra", http.StatusInternalServerError)
						return
					}
					validacionEstado = "inconsistente"
					evento = "compra_documentos_inconsistentes"
					validationStatus = statusErr
					validationErr = err
				} else {
					validacionEstado = "validada"
					evento = "compra_documentos_validados"
				}

			default:
				if action == "contabilizar" || action == "contabilizar_compra" {
					if estadoActualDocumento != "recepcionada" {
						http.Error(w, "transicion invalida: contabilizar_compra requiere estado real recepcionada", http.StatusConflict)
						return
					}
				}
				if (action == "emitir" || action == "emitir_orden") && requiereAprobacion && nivelAprobacion < nivelesAprobacion {
					http.Error(w, "la orden requiere aprobacion multinivel antes de emitir", http.StatusConflict)
					return
				}

				transition, err := resolveComprasTransition(action, payload.EstadoActual)
				if err != nil {
					errLower := strings.ToLower(err.Error())
					switch {
					case strings.Contains(errLower, "transicion invalida"):
						http.Error(w, err.Error(), http.StatusConflict)
						return
					case strings.Contains(errLower, "accion no soportada"):
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					default:
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
				}
				estadoDocumento = transition.EstadoNuevo
				estadoAnterior = transition.EstadoAnterior
				evento = transition.Evento
				accionResp = transition.Accion
			}

			docPersistido, err := dbpkg.UpsertEmpresaDocumentoCompra(dbEmp, dbpkg.EmpresaDocumentoCompra{
				EmpresaID:            payload.EmpresaID,
				ProveedorID:          payload.ProveedorID,
				TipoDocumento:        docActual.TipoDocumento,
				DocumentoCodigo:      docActual.DocumentoCodigo,
				EstadoDocumento:      estadoDocumento,
				EstadoAnterior:       estadoAnterior,
				EventoUltimo:         evento,
				PeriodoContable:      comprasFirstNonBlank(payload.PeriodoContable, docActual.PeriodoContable),
				MontoTotal:           comprasFirstPositive(payload.MontoTotal, docActual.MontoTotal),
				Moneda:               comprasFirstNonBlank(payload.Moneda, docActual.Moneda),
				FechaDocumento:       comprasFirstNonBlank(payload.FechaDocumento, docActual.FechaDocumento),
				EntidadRelacionadaID: payload.ProveedorID,
				RequiereAprobacion:   requiereAprobacion,
				NivelesAprobacion:    nivelesAprobacion,
				NivelAprobacion:      nivelAprobacion,
				AprobadoresJSON:      aprobadoresJSON,
				RecepcionDetalleJSON: recepcionDetalleJSON,
				RecepcionResumenJSON: recepcionResumenJSON,
				ValidacionEstado:     validacionEstado,
				ProveedorDocRef:      proveedorDocRef,
				FacturaDocRef:        facturaDocRef,
				EntradaDocRef:        entradaDocRef,
				UsuarioCreador:       usuarioActual,
				Estado:               docActual.Estado,
				Observaciones:        comprasFirstNonBlank(payload.Observaciones, docActual.Observaciones),
			})
			if err != nil {
				http.Error(w, "No se pudo actualizar el documento de compras", http.StatusInternalServerError)
				return
			}

			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       payload.EmpresaID,
				Modulo:          "compras",
				Evento:          evento,
				Entidad:         "orden_compra",
				EntidadID:       docPersistido.ID,
				DocumentoTipo:   docPersistido.TipoDocumento,
				DocumentoCodigo: docPersistido.DocumentoCodigo,
				PeriodoContable: docPersistido.PeriodoContable,
				MontoTotal:      docPersistido.MontoTotal,
				Moneda:          docPersistido.Moneda,
				Origen:          "api_compras_documentos",
				Observaciones:   strings.TrimSpace(payload.Observaciones),
			}, map[string]interface{}{
				"accion":           accionResp,
				"estado_anterior":  docPersistido.EstadoAnterior,
				"estado_nuevo":     docPersistido.EstadoDocumento,
				"entidad_id":       docPersistido.ID,
				"documento_codigo": docPersistido.DocumentoCodigo,
				"proveedor_id":     docPersistido.ProveedorID,
				"empresa_id":       docPersistido.EmpresaID,
				"nivel_aprobacion": docPersistido.NivelAprobacion,
				"forma_pago":       comprasFirstNonBlank(payload.FormaPago, "credito"),
				"metodo_pago":      strings.TrimSpace(payload.MetodoPago),
				"subtotal":         comprasFirstPositive(payload.Subtotal, payload.BaseGravable, docPersistido.MontoTotal),
				"base_gravable":    comprasFirstPositive(payload.BaseGravable, payload.Subtotal, docPersistido.MontoTotal),
				"iva":              comprasFirstPositive(payload.IVA, payload.Impuestos),
				"impuestos":        comprasFirstPositive(payload.Impuestos, payload.IVA),
				"retencion_fuente": payload.RetencionFuente,
				"retencion_ica":    payload.RetencionICA,
				"retencion_iva":    payload.RetencionIVA,
				"total_retenciones": comprasFirstPositive(
					payload.TotalRetenciones,
					payload.RetencionFuente+payload.RetencionICA+payload.RetencionIVA,
				),
				"total_neto": payload.TotalNeto,
			})

			if validationErr != nil {
				http.Error(w, validationErr.Error(), validationStatus)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			response := map[string]interface{}{
				"ok":        true,
				"accion":    accionResp,
				"evento":    evento,
				"resultado": docPersistido,
			}
			if hasRecepcionResumen {
				response["recepcion_resumen"] = recepcionResumen
			}
			json.NewEncoder(w).Encode(response)
			return

		case http.MethodDelete:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			tipoDocumento := comprasFirstNonBlank(r.URL.Query().Get("tipo_documento"), "orden_compra")
			documentoCodigo := strings.TrimSpace(r.URL.Query().Get("documento_codigo"))
			if documentoCodigo == "" {
				http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
				return
			}

			if err := dbpkg.SetEmpresaDocumentoCompraEstadoByCodigo(dbEmp, empresaID, tipoDocumento, documentoCodigo, "inactivo"); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "documento no encontrado", http.StatusNotFound)
					return
				}
				http.Error(w, "No se pudo eliminar documento de compras", http.StatusInternalServerError)
				return
			}

			registrarEventoContableNoBloqueante(dbEmp, r, "compras", dbpkg.EmpresaEventoContable{
				EmpresaID:       empresaID,
				Modulo:          "compras",
				Evento:          "orden_compra_eliminada",
				Entidad:         "orden_compra",
				DocumentoTipo:   tipoDocumento,
				DocumentoCodigo: strings.ToUpper(strings.TrimSpace(documentoCodigo)),
				Origen:          "api_compras_documentos",
				Observaciones:   "eliminacion logica de orden de compra",
			}, map[string]interface{}{
				"accion":           "eliminar",
				"documento_codigo": strings.ToUpper(strings.TrimSpace(documentoCodigo)),
				"empresa_id":       empresaID,
			})

			w.WriteHeader(http.StatusNoContent)
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// EmpresaComprasDocumentoComprobanteUploadHandler carga un recibo o comprobante físico para un documento de compras.
func EmpresaComprasDocumentoComprobanteUploadHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "payload multipart invalido", http.StatusBadRequest)
			return
		}

		empresaID, err := parseInt64Form(r, "empresa_id")
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		documentoCodigo := strings.TrimSpace(r.FormValue("documento_codigo"))
		if documentoCodigo == "" {
			http.Error(w, "documento_codigo es obligatorio", http.StatusBadRequest)
			return
		}
		tipoDocumento := strings.TrimSpace(r.FormValue("tipo_documento"))
		if tipoDocumento == "" {
			tipoDocumento = "orden_compra"
		}

		file, header, err := r.FormFile("archivo")
		if err != nil {
			file, header, err = r.FormFile("comprobante")
		}
		if err != nil {
			http.Error(w, "archivo es obligatorio", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileURL, fileName, absPath, err := saveEmpresaComprobanteUpload(file, header.Filename, empresaID, "compras", documentoCodigo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := dbpkg.UpdateEmpresaDocumentoCompraComprobante(dbEmp, empresaID, tipoDocumento, documentoCodigo, fileURL, fileName); err != nil {
			_ = os.Remove(absPath)
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "documento de compras no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo guardar el comprobante", http.StatusInternalServerError)
			return
		}

		item, err := dbpkg.GetEmpresaDocumentoCompraByCodigo(dbEmp, empresaID, tipoDocumento, documentoCodigo)
		if err != nil {
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"ok":                         true,
				"empresa_id":                 empresaID,
				"documento_codigo":           documentoCodigo,
				"tipo_documento":             tipoDocumento,
				"comprobante_url":            fileURL,
				"comprobante_nombre_archivo": fileName,
			})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"ok":                         true,
			"empresa_id":                 empresaID,
			"documento_codigo":           documentoCodigo,
			"tipo_documento":             tipoDocumento,
			"comprobante_url":            fileURL,
			"comprobante_nombre_archivo": fileName,
			"resultado":                  item,
		})
	}
}

func parseBoolQuery(r *http.Request, key string) bool {
	raw := strings.ToLower(strings.TrimSpace(r.URL.Query().Get(key)))
	switch raw {
	case "1", "true", "si", "yes":
		return true
	default:
		return false
	}
}

type comprasRecepcionItem struct {
	ProductoID       int64   `json:"producto_id"`
	CantidadOrdenada float64 `json:"cantidad_ordenada"`
	CantidadRecibida float64 `json:"cantidad_recibida"`
	CostoUnitario    float64 `json:"costo_unitario"`
	Diferencia       float64 `json:"diferencia"`
	DiferenciaTipo   string  `json:"diferencia_tipo"`
	DiferenciaMotivo string  `json:"diferencia_motivo,omitempty"`
}

type comprasRecepcionResumen struct {
	TotalItems         int                    `json:"total_items"`
	ItemsConDiferencia int                    `json:"items_con_diferencia"`
	ItemsPendientes    int                    `json:"items_pendientes"`
	CantidadOrdenada   float64                `json:"cantidad_ordenada"`
	CantidadRecibida   float64                `json:"cantidad_recibida"`
	MontoOrdenado      float64                `json:"monto_ordenado"`
	MontoRecibido      float64                `json:"monto_recibido"`
	EstadoRecepcion    string                 `json:"estado_recepcion"`
	Items              []comprasRecepcionItem `json:"items"`
}

type comprasAprobacionPaso struct {
	Nivel         int    `json:"nivel"`
	Usuario       string `json:"usuario"`
	FechaAprobado string `json:"fecha_aprobado"`
	Observaciones string `json:"observaciones,omitempty"`
}

func normalizeComprasAction(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	return v
}

func normalizeComprasDocumentoRef(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

func roundCompras(v float64) float64 {
	return math.Round(v*100) / 100
}

func buildComprasRecepcionResumen(items []comprasRecepcionItem) (comprasRecepcionResumen, error) {
	if len(items) == 0 {
		return comprasRecepcionResumen{}, fmt.Errorf("recepcion_items es obligatorio")
	}

	resumen := comprasRecepcionResumen{
		EstadoRecepcion: "recepcionada",
		Items:           make([]comprasRecepcionItem, 0, len(items)),
	}

	for _, item := range items {
		if item.ProductoID <= 0 {
			return comprasRecepcionResumen{}, fmt.Errorf("producto_id invalido en recepcion_items")
		}
		if item.CantidadOrdenada < 0 || item.CantidadRecibida < 0 {
			return comprasRecepcionResumen{}, fmt.Errorf("cantidades invalidas en recepcion_items")
		}
		if item.CostoUnitario < 0 {
			return comprasRecepcionResumen{}, fmt.Errorf("costo_unitario invalido en recepcion_items")
		}

		item.CantidadOrdenada = roundCompras(item.CantidadOrdenada)
		item.CantidadRecibida = roundCompras(item.CantidadRecibida)
		item.CostoUnitario = roundCompras(item.CostoUnitario)
		item.Diferencia = roundCompras(item.CantidadRecibida - item.CantidadOrdenada)

		if strings.TrimSpace(item.DiferenciaTipo) == "" {
			switch {
			case item.Diferencia == 0:
				item.DiferenciaTipo = "sin_diferencia"
			case item.Diferencia < 0:
				item.DiferenciaTipo = "faltante"
			default:
				item.DiferenciaTipo = "excedente"
			}
		} else {
			item.DiferenciaTipo = normalizeComprasAction(item.DiferenciaTipo)
		}

		resumen.TotalItems++
		if item.Diferencia != 0 {
			resumen.ItemsConDiferencia++
		}
		if item.CantidadRecibida < item.CantidadOrdenada {
			resumen.ItemsPendientes++
		}
		resumen.CantidadOrdenada = roundCompras(resumen.CantidadOrdenada + item.CantidadOrdenada)
		resumen.CantidadRecibida = roundCompras(resumen.CantidadRecibida + item.CantidadRecibida)
		resumen.MontoOrdenado = roundCompras(resumen.MontoOrdenado + (item.CantidadOrdenada * item.CostoUnitario))
		resumen.MontoRecibido = roundCompras(resumen.MontoRecibido + (item.CantidadRecibida * item.CostoUnitario))
		resumen.Items = append(resumen.Items, item)
	}

	if resumen.ItemsPendientes > 0 {
		resumen.EstadoRecepcion = "recepcion_parcial"
	}

	return resumen, nil
}

func parseComprasRecepcionResumen(raw string) (comprasRecepcionResumen, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return comprasRecepcionResumen{}, false
	}
	var resumen comprasRecepcionResumen
	if err := json.Unmarshal([]byte(raw), &resumen); err != nil {
		return comprasRecepcionResumen{}, false
	}
	return resumen, true
}

func appendComprasAprobador(aprobadoresJSON, usuario, observaciones string, nivel int) string {
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	steps := make([]comprasAprobacionPaso, 0)
	if strings.TrimSpace(aprobadoresJSON) != "" {
		_ = json.Unmarshal([]byte(aprobadoresJSON), &steps)
	}
	steps = append(steps, comprasAprobacionPaso{
		Nivel:         nivel,
		Usuario:       usuario,
		FechaAprobado: time.Now().Format("2006-01-02 15:04:05"),
		Observaciones: strings.TrimSpace(observaciones),
	})
	encoded, err := json.Marshal(steps)
	if err != nil {
		return "[]"
	}
	return string(encoded)
}

func validarComprasDocumentos(dbEmp *sql.DB, empresaID, proveedorID int64, proveedorRefRaw, facturaRefRaw, entradaRefRaw string) (string, string, string, int, error) {
	proveedorRef := normalizeComprasDocumentoRef(proveedorRefRaw)
	facturaRef := normalizeComprasDocumentoRef(facturaRefRaw)
	entradaRef := normalizeComprasDocumentoRef(entradaRefRaw)

	if proveedorID <= 0 {
		return proveedorRef, facturaRef, entradaRef, http.StatusBadRequest, fmt.Errorf("proveedor_id es obligatorio")
	}
	if facturaRef == "" {
		return proveedorRef, facturaRef, entradaRef, http.StatusBadRequest, fmt.Errorf("factura_documento_ref es obligatorio")
	}
	if entradaRef == "" {
		return proveedorRef, facturaRef, entradaRef, http.StatusBadRequest, fmt.Errorf("entrada_documento_ref es obligatorio")
	}
	if facturaRef == entradaRef {
		return proveedorRef, facturaRef, entradaRef, http.StatusConflict, fmt.Errorf("factura_documento_ref y entrada_documento_ref no pueden ser iguales")
	}

	var proveedorDocumento string
	err := dbEmp.QueryRow(`SELECT COALESCE(documento, '') FROM proveedores WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, proveedorID).Scan(&proveedorDocumento)
	if errors.Is(err, sql.ErrNoRows) {
		return proveedorRef, facturaRef, entradaRef, http.StatusConflict, fmt.Errorf("proveedor no encontrado para validacion documental")
	}
	if err != nil {
		return proveedorRef, facturaRef, entradaRef, http.StatusInternalServerError, err
	}

	proveedorDocumento = normalizeComprasDocumentoRef(proveedorDocumento)
	if proveedorRef == "" {
		proveedorRef = proveedorDocumento
	}
	if proveedorRef == "" {
		return proveedorRef, facturaRef, entradaRef, http.StatusBadRequest, fmt.Errorf("proveedor_documento_ref es obligatorio")
	}
	if proveedorDocumento != "" && proveedorRef != proveedorDocumento {
		return proveedorRef, facturaRef, entradaRef, http.StatusConflict, fmt.Errorf("proveedor_documento_ref no coincide con el documento registrado del proveedor")
	}

	return proveedorRef, facturaRef, entradaRef, 0, nil
}

func comprasFirstNonBlank(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func comprasFirstPositive(values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func comprasFirstNonBlankInt64(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
