package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func EmpresaContabilidadColombiaAvanzadaHandler(dbEmp *sql.DB) http.HandlerFunc {
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
				row, err := dbpkg.BuildEmpresaContabilidadAvanzadaDashboard(dbEmp, empresaID)
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
				rows, err := dbpkg.ListEmpresaNominaElectronica(dbEmp, empresaID, r.URL.Query().Get("periodo"))
				if err != nil {
					http.Error(w, "No se pudo listar nomina electronica", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
				return
			case "documentos_soporte":
				rows, err := dbpkg.ListEmpresaDocumentosSoporte(dbEmp, empresaID, r.URL.Query().Get("periodo"))
				if err != nil {
					http.Error(w, "No se pudieron listar documentos soporte", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, rows)
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
				row, err := dbpkg.BuildEmpresaContabilidadAvanzadaDashboard(dbEmp, empresaID)
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
				var payload dbpkg.EmpresaNominaElectronica
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
					return
				}
				payload.EmpresaID = empresaID
				payload.UsuarioCreador = usuario
				id, err := dbpkg.CreateEmpresaNominaElectronica(dbEmp, payload)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				writeJSON(w, http.StatusCreated, map[string]interface{}{"ok": true, "id": id, "referencia": dbpkg.FormatEmpresaDocumentoElectronicoRef("NE", empresaID, id)})
				return
			case "documentos_soporte":
				var payload dbpkg.EmpresaDocumentoSoporteElectronico
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					http.Error(w, "JSON invalido", http.StatusBadRequest)
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
