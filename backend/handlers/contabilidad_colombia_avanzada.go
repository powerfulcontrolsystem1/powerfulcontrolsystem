package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaContabilidadColombiaAvanzadaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
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
				includeNomina, status, message, permissionErr := inspectEmpresaAdditionalModulePermission(r, dbEmp, dbSuper, permModuleNominaSueldos, permActionRead, "linkNominaSueldos")
				if permissionErr != nil {
					http.Error(w, "No se pudo validar el acceso a nómina", http.StatusInternalServerError)
					return
				}
				if !includeNomina && status != http.StatusForbidden {
					http.Error(w, message, status)
					return
				}
				row, err := dbpkg.BuildEmpresaContabilidadAvanzadaDashboardScoped(dbEmp, empresaID, includeNomina)
				if err != nil {
					http.Error(w, "No se pudo consultar la suite contable Colombia", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "exogena_formatos":
				rows, err := dbpkg.ListEmpresaExogenaFormatos(dbEmp, empresaID, intQuery(r, "anio"))
				if err != nil {
					http.Error(w, "No se pudieron listar formatos de informacion exogena", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "exogena_registros":
				rows, err := dbpkg.ListEmpresaExogenaRegistros(dbEmp, empresaID, int64Query(r, "formato_id"))
				if err != nil {
					http.Error(w, "No se pudieron listar registros de informacion exogena", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "nomina_electronica":
				if !requireEmpresaAdditionalModulePermission(w, r, dbEmp, dbSuper, permModuleNominaSueldos, permActionRead, "linkNominaSueldos") {
					return
				}
				rows, err := dbpkg.ListEmpresaNominaElectronica(dbEmp, empresaID, r.URL.Query().Get("periodo"))
				if err != nil {
					http.Error(w, "No se pudo listar nomina electronica", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "nomina_electronica_preflight":
				if !requireEmpresaAdditionalModulePermission(w, r, dbEmp, dbSuper, permModuleNominaSueldos, permActionRead, "linkNominaSueldos") {
					return
				}
				nominaID, err := parseInt64QueryOptional(r, "nomina_id")
				if err != nil || nominaID <= 0 {
					http.Error(w, "nomina_id es obligatorio", http.StatusBadRequest)
					return
				}
				nomina, err := dbpkg.GetEmpresaNominaElectronicaByIDContext(r.Context(), dbEmp, empresaID, nominaID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "nómina electrónica no encontrada", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar la nómina electrónica", http.StatusInternalServerError)
					return
				}
				if nomina.LiquidacionID <= 0 {
					writeJSON(w, http.StatusOK, map[string]interface{}{
						"ok": false, "bloqueado": true, "empresa_id": empresaID, "nomina_id": nomina.ID,
						"error": "el registro histórico no tiene una liquidación fuente; no puede reconstruirse ni emitirse automáticamente",
					})
					return
				}
				preflightContext, preflightErr := loadNominaElectronicaPreflightContext(r.Context(), dbEmp, empresaID, nomina.LiquidacionID, false)
				resultado := map[string]interface{}{
					"ok": preflightErr == nil, "bloqueado": preflightErr != nil,
					"empresa_id": empresaID, "nomina_id": nomina.ID, "liquidacion_id": nomina.LiquidacionID,
				}
				if preflightContext != nil && preflightContext.preflight != nil {
					for key, value := range preflightContext.preflight {
						resultado[key] = value
					}
				}
				if preflightErr != nil {
					resultado["error"] = preflightErr.Error()
				}
				writeJSON(w, http.StatusOK, resultado)
				return
			case "documentos_soporte":
				rows, err := dbpkg.ListEmpresaDocumentosSoporte(dbEmp, empresaID, r.URL.Query().Get("periodo"))
				if err != nil {
					http.Error(w, "No se pudieron listar documentos soporte", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "documento_soporte_preflight":
				documentoSoporteID, err := parseInt64QueryOptional(r, "documento_soporte_id")
				if err != nil || documentoSoporteID <= 0 {
					http.Error(w, "documento_soporte_id es obligatorio", http.StatusBadRequest)
					return
				}
				documento, err := dbpkg.GetEmpresaDocumentoSoporteByIDContext(r.Context(), dbEmp, empresaID, documentoSoporteID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "documento soporte no encontrado", http.StatusNotFound)
						return
					}
					http.Error(w, "No se pudo consultar el documento soporte", http.StatusInternalServerError)
					return
				}
				var configuracion *dbpkg.EmpresaDIANDocumentoConfiguracion
				configuracionPendiente := ""
				if err := dbpkg.EmpresaDIANDocumentosConfiguracionSchemaReady(dbEmp); err != nil {
					configuracionPendiente = "La tabla de configuración DIAN por documento no está disponible; ejecute pcs-migrate."
				} else {
					configuracion, err = dbpkg.GetEmpresaDIANDocumentoConfiguracionContext(r.Context(), dbEmp, empresaID, "documento_soporte")
					if err != nil && !errors.Is(err, sql.ErrNoRows) {
						http.Error(w, "No se pudo consultar la configuración DIAN de documento soporte", http.StatusInternalServerError)
						return
					}
					if errors.Is(err, sql.ErrNoRows) {
						configuracionPendiente = "No existe configuración DIAN separada para documento soporte."
					}
				}
				var empresa *dbpkg.EmpresaConfiguracionAvanzada
				empresa, err = dbpkg.GetEmpresaConfiguracionAvanzada(dbEmp, empresaID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar la configuración fiscal empresarial", http.StatusInternalServerError)
					return
				}
				configuracionPrincipal, configErr := getEmpresaDIANConfig(dbEmp, empresaID)
				if configErr != nil && !errors.Is(configErr, sql.ErrNoRows) {
					http.Error(w, "No se pudo consultar la configuración DIAN principal", http.StatusInternalServerError)
					return
				}
				resultado := buildDocumentoSoporteDIANPreflight(documento, configuracion, empresa, configuracionPrincipal)
				if configuracionPendiente != "" {
					resultado.Bloqueos = append(resultado.Bloqueos, configuracionPendiente)
					resultado.PuedeEmitir = false
				}
				writeJSON(w, http.StatusOK, resultado)
				return
			case "activos_fijos":
				rows, err := dbpkg.ListEmpresaActivosFijos(dbEmp, empresaID, r.URL.Query().Get("estado"))
				if err != nil {
					http.Error(w, "No se pudieron listar activos fijos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "activos_resumen":
				row, err := dbpkg.BuildEmpresaActivosFijosAvanzadoResumen(dbEmp, empresaID, r.URL.Query().Get("periodo"))
				if err != nil {
					http.Error(w, "No se pudo consultar resumen avanzado de activos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "activos_depreciaciones":
				rows, err := dbpkg.ListEmpresaActivosDepreciacion(dbEmp, empresaID, r.URL.Query().Get("periodo"), 1000)
				if err != nil {
					http.Error(w, "No se pudieron listar depreciaciones", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "activos_eventos":
				rows, err := dbpkg.ListEmpresaActivosEventos(dbEmp, empresaID, int64Query(r, "activo_id"), 500)
				if err != nil {
					http.Error(w, "No se pudieron listar eventos de activos", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "cartera_cxp":
				rows, err := dbpkg.ListEmpresaCarteraCXP(dbEmp, empresaID, r.URL.Query().Get("tipo"), r.URL.Query().Get("estado"))
				if err != nil {
					http.Error(w, "No se pudo listar cartera y cuentas por pagar", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "cartera_cxp_edades", "edades_cartera":
				row, err := dbpkg.BuildEmpresaCarteraCXPEdades(dbEmp, empresaID, r.URL.Query().Get("tipo"), r.URL.Query().Get("fecha_corte"))
				if err != nil {
					http.Error(w, "No se pudo calcular edades de cartera", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row)
				return
			case "libros":
				rows, err := dbpkg.ListEmpresaLibroOficial(dbEmp, empresaID, r.URL.Query().Get("tipo"), r.URL.Query().Get("periodo"))
				if err != nil {
					http.Error(w, "No se pudieron generar libros oficiales", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "libros_resumen":
				row, err := dbpkg.BuildEmpresaContabilidadAvanzadaDashboardScoped(dbEmp, empresaID, false)
				if err != nil {
					http.Error(w, "No se pudo generar resumen de libros", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, row.LibrosDisponibles)
				return
			}
		case http.MethodPost, http.MethodPut:
			switch action {
			case "seed":
				anio := intQuery(r, "anio")
				if anio <= 0 {
					anio = time.Now().Year()
				}
				if err := dbpkg.SeedEmpresaContabilidadAvanzadaBase(dbEmp, empresaID, usuario, anio); err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
				return
			case "exogena_formatos":
				var payload dbpkg.EmpresaExogenaFormato
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaExogenaFormato(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "exogena_registros":
				var payload dbpkg.EmpresaExogenaRegistro
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaExogenaRegistro(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "generar_exogena":
				formatoID := int64Query(r, "formato_id")
				if formatoID <= 0 {
					var payload struct {
						FormatoID int64 `json:"formato_id"`
					}
					_ = json.NewDecoder(r.Body).Decode(&payload)
					formatoID = payload.FormatoID
				}
				created, err := dbpkg.GenerateEmpresaExogenaFromAccounting(dbEmp, empresaID, formatoID, usuario)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "creados": created})
				return
			case "nomina_electronica":
				if !requireEmpresaAdditionalModulePermission(w, r, dbEmp, dbSuper, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos") {
					return
				}
				writeJSON(w, http.StatusConflict, map[string]interface{}{
					"ok": false, "bloqueado": true, "codigo": "nomina_manual_no_permitida",
					"error": "la nómina electrónica no se crea manualmente; use Nómina y sueldos con liquidaciones y pagos reales para construir la fuente mensual",
				})
				return
			case "documentos_soporte":
				var payload dbpkg.EmpresaDocumentoSoporteElectronico
				decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
				decoder.DisallowUnknownFields()
				if err := decoder.Decode(&payload); err != nil {
					http.Error(w, "JSON de documento soporte invalido", http.StatusBadRequest)
					return
				}
				if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
					http.Error(w, "JSON de documento soporte invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaDocumentoSoporte(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id, "referencia": dbpkg.FormatEmpresaDocumentoElectronicoRef("DS", empresaID, id)})
				return
			case "activos_fijos":
				var payload dbpkg.EmpresaActivoFijo
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaActivoFijo(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "generar_depreciacion_activos":
				var payload struct {
					Periodo string `json:"periodo"`
				}
				_ = json.NewDecoder(r.Body).Decode(&payload)
				periodo := strings.TrimSpace(payload.Periodo)
				if periodo == "" {
					periodo = strings.TrimSpace(r.URL.Query().Get("periodo"))
				}
				rows, err := dbpkg.GenerarEmpresaActivosDepreciacion(dbEmp, empresaID, periodo, usuario)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "depreciaciones": rows})
				return
			case "activo_evento":
				var payload dbpkg.EmpresaActivoEvento
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.RegistrarEmpresaActivoEvento(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "cartera_cxp":
				var payload dbpkg.EmpresaCarteraCXP
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				// P106: the advanced accounting table remains a legacy read surface
				// for historical CxP while it is reconciled. New supplier obligations
				// must use the canonical Finanzas CxP module; accepting new writes here
				// would recreate the two-source-of-truth defect.
				if strings.EqualFold(strings.TrimSpace(payload.Tipo), "cxp") {
					http.Error(w, "Las nuevas cuentas por pagar se registran en Finanzas > Cartera de proveedores; la cartera contable historica esta en conciliacion P106", http.StatusConflict)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaCarteraCXP(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				evento := "cuenta_por_cobrar_generada"
				if strings.EqualFold(payload.Tipo, "cxp") {
					evento = "cuenta_por_pagar_generada"
				}
				registrarEventoContableNoBloqueante(dbEmp, r, "contabilidad_avanzada", dbpkg.EmpresaEventoContable{
					EmpresaID:       empresaID,
					Modulo:          "cartera",
					Evento:          evento,
					Entidad:         "empresa_contabilidad_cartera_cxp",
					EntidadID:       id,
					DocumentoTipo:   strings.ToLower(strings.TrimSpace(payload.Tipo)),
					DocumentoCodigo: strings.TrimSpace(payload.Documento),
					MontoTotal:      payload.Saldo,
					Moneda:          "COP",
					Origen:          "api_contabilidad_avanzada_cartera",
					UsuarioCreador:  usuario,
					Estado:          "activo",
					Observaciones:   strings.TrimSpace(payload.Concepto),
				}, map[string]interface{}{
					"tipo":               strings.ToLower(strings.TrimSpace(payload.Tipo)),
					"tercero_id":         payload.TerceroID,
					"tercero_nombre":     strings.TrimSpace(payload.TerceroNombre),
					"documento":          strings.TrimSpace(payload.Documento),
					"fecha_emision":      strings.TrimSpace(payload.FechaEmision),
					"fecha_vencimiento":  strings.TrimSpace(payload.FechaVencimiento),
					"cuenta_cartera":     strings.TrimSpace(payload.CuentaCodigo),
					"cuenta_cxp":         strings.TrimSpace(payload.CuentaCodigo),
					"subtotal":           payload.ValorOriginal,
					"base_gravable":      payload.ValorOriginal,
					"total_neto":         payload.Saldo,
					"saldo":              payload.Saldo,
					"origen_modulo":      strings.TrimSpace(payload.OrigenModulo),
					"referencia_externa": strings.TrimSpace(payload.ReferenciaExterna),
				})
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id})
				return
			case "cartera_cxp_abono", "abono_cartera":
				var payload struct {
					ID              int64   `json:"id"`
					Monto           float64 `json:"monto"`
					FechaAplicacion string  `json:"fecha_aplicacion"`
					ReferenciaPago  string  `json:"referencia_pago"`
					MetodoPago      string  `json:"metodo_pago"`
					Observaciones   string  `json:"observaciones"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				if payload.ID <= 0 {
					payload.ID = int64Query(r, "id")
				}
				if payload.Monto <= 0 {
					payload.Monto = floatQuery(r, "monto")
				}
				carteraActual, err := dbpkg.GetEmpresaCarteraCXPByID(dbEmp, empresaID, payload.ID)
				if err != nil {
					http.Error(w, "Cuenta de cartera no encontrada", http.StatusNotFound)
					return
				}
				if strings.EqualFold(carteraActual.Tipo, "cxp") {
					http.Error(w, "Los abonos a CxP historica estan bloqueados hasta completar la conciliacion P106; use la cuenta canonica en Finanzas", http.StatusConflict)
					return
				}
				result, err := dbpkg.AplicarEmpresaCarteraCXPAbono(dbEmp, empresaID, payload.ID, payload.Monto, payload.FechaAplicacion, payload.ReferenciaPago, usuario)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				moduloEvento := "cartera"
				documentoTipo := strings.ToLower(strings.TrimSpace(result.Cartera.Tipo))
				registrarEventoContableNoBloqueante(dbEmp, r, "contabilidad_avanzada", dbpkg.EmpresaEventoContable{
					EmpresaID:       empresaID,
					Modulo:          moduloEvento,
					Evento:          result.EventoContable,
					Entidad:         "empresa_contabilidad_cartera_cxp",
					EntidadID:       result.Cartera.ID,
					DocumentoTipo:   documentoTipo,
					DocumentoCodigo: result.DocumentoContable,
					MontoTotal:      result.MontoAplicado,
					Moneda:          "COP",
					Origen:          "api_contabilidad_avanzada_cartera_abono",
					UsuarioCreador:  usuario,
					Estado:          "activo",
					Observaciones:   strings.TrimSpace(payload.Observaciones),
				}, map[string]interface{}{
					"tipo":             documentoTipo,
					"tercero_id":       result.Cartera.TerceroID,
					"tercero_nombre":   result.Cartera.TerceroNombre,
					"documento":        result.Cartera.Documento,
					"metodo_pago":      strings.TrimSpace(payload.MetodoPago),
					"referencia_pago":  strings.TrimSpace(payload.ReferenciaPago),
					"monto":            result.MontoAplicado,
					"total_neto":       result.MontoAplicado,
					"saldo_anterior":   result.SaldoAnterior,
					"saldo_nuevo":      result.SaldoNuevo,
					"estado_anterior":  result.EstadoAnterior,
					"estado_nuevo":     result.EstadoNuevo,
					"cuenta_cartera":   result.Cartera.CuentaCodigo,
					"cuenta_cxp":       result.Cartera.CuentaCodigo,
					"fecha_aplicacion": result.FechaAplicacion,
				})
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "resultado": result})
				return
			}
		}
		http.Error(w, "Metodo o accion no permitida", http.StatusMethodNotAllowed)
	}
}

type documentoSoporteDIANPreflight struct {
	EmpresaID           int64    `json:"empresa_id"`
	DocumentoSoporteID  int64    `json:"documento_soporte_id"`
	TipoDocumento       string   `json:"tipo_documento"`
	Estado              string   `json:"estado"`
	PuedeEmitir         bool     `json:"puede_emitir"`
	Bloqueos            []string `json:"bloqueos"`
	Advertencias        []string `json:"advertencias"`
	EstadoConfiguracion string   `json:"estado_configuracion,omitempty"`
	Ambiente            string   `json:"ambiente,omitempty"`
}

// buildDocumentoSoporteDIANPreflight validates every server-side prerequisite
// without reserving a legal number, generating XML, signing, or transmitting.
func buildDocumentoSoporteDIANPreflight(documento *dbpkg.EmpresaDocumentoSoporteElectronico, configuracion *dbpkg.EmpresaDIANDocumentoConfiguracion, empresa *dbpkg.EmpresaConfiguracionAvanzada, configuracionPrincipal map[string]interface{}) documentoSoporteDIANPreflight {
	resultado := documentoSoporteDIANPreflight{
		TipoDocumento: "documento_soporte",
		Estado:        "bloqueado_preflight",
		Bloqueos:      make([]string, 0),
		Advertencias:  make([]string, 0),
	}
	addBlock := func(message string) {
		message = strings.TrimSpace(message)
		if message == "" {
			return
		}
		for _, existing := range resultado.Bloqueos {
			if existing == message {
				return
			}
		}
		resultado.Bloqueos = append(resultado.Bloqueos, message)
	}
	if documento == nil {
		addBlock("No se encontró el borrador contable de documento soporte.")
		return resultado
	}
	resultado.EmpresaID = documento.EmpresaID
	resultado.DocumentoSoporteID = documento.ID
	switch strings.ToLower(strings.TrimSpace(documento.EstadoDIAN)) {
	case "", "borrador", "preparado", "pendiente", "fallido", "contingencia":
	case "aceptado":
		addBlock("El documento soporte ya fue aceptado por DIAN y no puede emitirse de nuevo.")
	default:
		addBlock("El estado actual del documento soporte no permite una nueva emisión.")
	}
	if strings.TrimSpace(documento.Documento) == "" {
		addBlock("Falta el documento del vendedor no obligado.")
	}
	if strings.TrimSpace(documento.NombreProveedor) == "" {
		addBlock("Falta el nombre del vendedor no obligado.")
	}
	if strings.TrimSpace(documento.FechaDocumento) == "" {
		addBlock("Falta la fecha de emisión del documento soporte.")
	}
	if strings.TrimSpace(documento.Concepto) == "" || documento.Subtotal <= 0 || documento.Total <= 0 {
		addBlock("Falta concepto o importes positivos del documento soporte.")
	}
	if math.Abs((documento.Subtotal+documento.IVA)-documento.Total) > 0.01 || math.Abs((documento.Total-documento.Retenciones)-documento.TotalNetoContable) > 0.01 {
		addBlock("Los importes no cuadran: total DIAN debe ser subtotal + IVA y el neto contable debe restar las retenciones.")
	}
	if documento.ProveedorID <= 0 {
		resultado.Advertencias = append(resultado.Advertencias, "El borrador no está vinculado a un proveedor empresarial; revise identidad y datos de contacto antes de emitir.")
	}
	if empresa == nil {
		addBlock("No existe configuración fiscal empresarial completa para identificar al adquirente.")
	} else if empresa.EmpresaID != documento.EmpresaID {
		addBlock("La configuración fiscal empresarial no pertenece a la empresa del documento soporte.")
	} else if _, err := buildDocumentoSoporteFuenteFiscal(documento, empresa); err != nil {
		addBlock(err.Error())
	}
	if configuracion == nil {
		addBlock("No existe configuración DIAN separada para documento soporte.")
		return resultado
	}
	resultado.EstadoConfiguracion = configuracion.Estado
	resultado.Ambiente = configuracion.TipoAmbiente
	if configuracion.EmpresaID > 0 && configuracion.EmpresaID != documento.EmpresaID {
		addBlock("La configuración DIAN de documento soporte no pertenece a la empresa.")
	} else if err := dbpkg.ValidateEmpresaDocumentoSoporteConfigForEmission(*configuracion, time.Now()); err != nil {
		addBlock(err.Error())
	}
	if len(configuracionPrincipal) == 0 {
		addBlock("No existe configuración DIAN principal para firma y transporte.")
	} else {
		merged := documentoSoporteMergeDIANConfig(configuracionPrincipal, documentoSoporteConfigSnapshotFromRow(configuracion))
		for _, field := range missingDIANFieldsForDocument(merged, "documento_soporte", documento.EmpresaID) {
			addBlock("Configuración DIAN incompleta: " + field + ".")
		}
		if empresa != nil {
			nitConfig := dianOnlyDigits(genericStringValue(merged["nit"]))
			nitEmpresa := dianOnlyDigits(empresa.NIT)
			if nitConfig != "" && nitEmpresa != "" && nitConfig != nitEmpresa {
				addBlock("El NIT de la configuración DIAN principal no coincide con el NIT fiscal de la empresa.")
			}
		}
	}
	if len(resultado.Bloqueos) == 0 {
		resultado.Estado = "listo_para_emision"
		resultado.PuedeEmitir = true
	}
	return resultado
}

func intQuery(r *http.Request, key string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get(key)))
	return v
}

func int64Query(r *http.Request, key string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get(key)), 10, 64)
	return v
}

func floatQuery(r *http.Request, key string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(r.URL.Query().Get(key)), 64)
	return v
}
