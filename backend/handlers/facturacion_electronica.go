package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

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

// EmpresaFacturacionElectronicaHandler gestiona configuración FE por empresa y país.
func EmpresaFacturacionElectronicaHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaID, err := parseEmpresaIDQuery(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
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
			if action == "emitir" || action == "anular" || action == "nota_credito" || action == "emitir_nota_credito" {
				var payload struct {
					EmpresaID       int64   `json:"empresa_id"`
					EntidadID       int64   `json:"entidad_id"`
					PaisCodigo      string  `json:"pais_codigo"`
					DocumentoCodigo string  `json:"documento_codigo"`
					EstadoActual    string  `json:"estado_actual"`
					MontoTotal      float64 `json:"monto_total"`
					Moneda          string  `json:"moneda"`
					PeriodoContable string  `json:"periodo_contable"`
					Observaciones   string  `json:"observaciones"`
				}
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

				documentoTipo := "factura_electronica"
				entidad := "factura_electronica"
				actionNormalized := normalizeDocumentoState(action)
				if actionNormalized == "nota_credito" || actionNormalized == "emitir_nota_credito" {
					documentoTipo = "nota_credito"
					entidad = "nota_credito"
				}

				docExistente, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, payload.EmpresaID, documentoTipo, payload.DocumentoCodigo)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar el estado documental de facturacion", http.StatusInternalServerError)
					return
				}
				if docExistente != nil {
					payload.EstadoActual = docExistente.EstadoDocumento
				}

				transition, err := resolveFacturacionTransition(action, payload.EstadoActual)
				if err != nil {
					http.Error(w, err.Error(), http.StatusConflict)
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

				resp := map[string]interface{}{
					"ok":                true,
					"accion":            transition.Accion,
					"evento":            evento,
					"estado_anterior":   transition.EstadoAnterior,
					"estado_nuevo":      transition.EstadoNuevo,
					"entidad_id":        docPersistido.ID,
					"documento_codigo":  strings.TrimSpace(payload.DocumentoCodigo),
					"numero_legal":      docPersistido.NumeroLegal,
					"codigo_validacion": docPersistido.CodigoValidacion,
					"pais_codigo":       docPersistido.PaisCodigo,
					"ambiente_fe":       docPersistido.AmbienteFE,
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
			"items": dbpkg.ListPaisesFacturacionDisponibles(),
		})
	}
}
