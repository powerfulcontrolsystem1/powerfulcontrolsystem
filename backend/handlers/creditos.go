package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaCreditoCreatePayload struct {
	EmpresaID             int64   `json:"empresa_id"`
	Codigo                string  `json:"codigo"`
	ClienteID             int64   `json:"cliente_id"`
	ClienteNombre         string  `json:"cliente_nombre"`
	TipoCredito           string  `json:"tipo_credito"`
	MontoAprobado         float64 `json:"monto_aprobado"`
	CupoCredito           float64 `json:"cupo_credito"`
	SaldoActual           float64 `json:"saldo_actual"`
	TasaInteres           float64 `json:"tasa_interes"`
	TasaMora              float64 `json:"tasa_mora"`
	PeriodicidadCuota     string  `json:"periodicidad_cuota"`
	ValorCuotaPactada     float64 `json:"valor_cuota_pactada"`
	OmitirDomingos        bool    `json:"omitir_domingos"`
	PlazoDias             int     `json:"plazo_dias"`
	PlazoCuotas           int     `json:"plazo_cuotas"`
	FechaInicio           string  `json:"fecha_inicio"`
	FechaVencimiento      string  `json:"fecha_vencimiento"`
	BloqueoAutomaticoMora bool    `json:"bloqueo_automatico_mora"`
	VentaOrigenID         int64   `json:"venta_origen_id"`
	DocumentoOrigen       string  `json:"documento_origen"`
	EstadoCredito         string  `json:"estado_credito"`
	Observaciones         string  `json:"observaciones"`
}

type empresaCreditoAbonoPayload struct {
	EmpresaID               int64   `json:"empresa_id"`
	CreditoID               int64   `json:"credito_id"`
	Monto                   float64 `json:"monto"`
	MetodoPago              string  `json:"metodo_pago"`
	ReferenciaPago          string  `json:"referencia_pago"`
	Comprobante             string  `json:"comprobante"`
	Observaciones           string  `json:"observaciones"`
	FechaMovimiento         string  `json:"fecha_movimiento"`
	RegistrarEventoContable *bool   `json:"registrar_evento_contable"`
	ProcesarAsientos        *bool   `json:"procesar_asientos"`
	AsientosLimit           *int    `json:"asientos_limit"`
	MaxReintentos           *int    `json:"max_reintentos"`
}

type empresaCreditoWorkflowSolicitudPayload struct {
	EmpresaID                int64                  `json:"empresa_id"`
	CreditoID                int64                  `json:"credito_id"`
	MovimientoOrigenID       int64                  `json:"movimiento_origen_id"`
	NivelAprobacionRequerido int                    `json:"nivel_aprobacion_requerido"`
	MotivoSolicitud          string                 `json:"motivo_solicitud"`
	Observaciones            string                 `json:"observaciones"`
	Payload                  map[string]interface{} `json:"payload"`
}

type empresaCreditoWorkflowDecisionPayload struct {
	EmpresaID        int64  `json:"empresa_id"`
	WorkflowID       int64  `json:"workflow_id"`
	AprobadoPor      string `json:"aprobado_por"`
	CodigoAprobacion string `json:"codigo_aprobacion"`
	MotivoAprobacion string `json:"motivo_aprobacion"`
	MotivoRechazo    string `json:"motivo_rechazo"`
	Observaciones    string `json:"observaciones"`
}

type empresaCreditoClienteLimitePayload struct {
	EmpresaID                int64   `json:"empresa_id"`
	ClienteID                int64   `json:"cliente_id"`
	LimiteSaldoTotal         float64 `json:"limite_saldo_total"`
	MaxCreditosActivos       int     `json:"max_creditos_activos"`
	RequiereAprobacionExceso bool    `json:"requiere_aprobacion_exceso"`
	Estado                   string  `json:"estado"`
	Observaciones            string  `json:"observaciones"`
}

func EmpresaCreditosHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "", "listar", "list":
				handleEmpresaCreditosList(w, r, dbEmp)
				return
			case "limites_cliente", "limite_cliente":
				handleEmpresaCreditosLimitesCliente(w, r, dbEmp)
				return
			case "workflows", "workflow":
				handleEmpresaCreditosWorkflows(w, r, dbEmp)
				return
			case "detalle":
				handleEmpresaCreditosDetail(w, r, dbEmp)
				return
			case "cuotas":
				handleEmpresaCreditosCuotas(w, r, dbEmp)
				return
			case "movimientos":
				handleEmpresaCreditosMovimientos(w, r, dbEmp)
				return
			case "estado_cuenta":
				handleEmpresaCreditosEstadoCuenta(w, r, dbEmp)
				return
			case "resumen", "resumen_cartera":
				handleEmpresaCreditosResumen(w, r, dbEmp)
				return
			case "alertas", "alertas_mora", "morosidad", "ranking_morosidad":
				handleEmpresaCreditosAlertasMora(w, r, dbEmp)
				return
			case "reporte", "export", "exportar":
				handleEmpresaCreditosReporte(w, r, dbEmp)
				return
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}

		case http.MethodPost:
			switch action {
			case "", "crear":
				handleEmpresaCreditosCreate(w, r, dbEmp)
				return
			case "limites_cliente", "limite_cliente", "upsert_limite_cliente":
				handleEmpresaCreditosUpsertLimiteCliente(w, r, dbEmp)
				return
			case "abono", "pago":
				handleEmpresaCreditosAbono(w, r, dbEmp)
				return
			case "solicitar_reverso", "solicitar_anulacion":
				handleEmpresaCreditosSolicitarReverso(w, r, dbEmp)
				return
			case "solicitar_refinanciacion":
				handleEmpresaCreditosSolicitarRefinanciacion(w, r, dbEmp)
				return
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}

		case http.MethodPut, http.MethodPatch:
			switch action {
			case "", "actualizar", "editar":
				handleEmpresaCreditosUpdate(w, r, dbEmp)
				return
			case "limites_cliente", "limite_cliente", "upsert_limite_cliente":
				handleEmpresaCreditosUpsertLimiteCliente(w, r, dbEmp)
				return
			case "aprobar_workflow", "aprobar_reverso", "aprobar_refinanciacion":
				handleEmpresaCreditosAprobarWorkflow(w, r, dbEmp)
				return
			case "rechazar_workflow", "rechazar_reverso", "rechazar_refinanciacion":
				handleEmpresaCreditosRechazarWorkflow(w, r, dbEmp)
				return
			case "estado":
				handleEmpresaCreditosEstado(w, r, dbEmp)
				return
			case "activar", "desactivar":
				handleEmpresaCreditosActivacion(w, r, dbEmp, action)
				return
			default:
				http.Error(w, "action invalida", http.StatusBadRequest)
				return
			}

		case http.MethodDelete:
			switch action {
			case "limites_cliente", "limite_cliente", "eliminar_limite_cliente":
				handleEmpresaCreditosDeleteLimiteCliente(w, r, dbEmp)
				return
			default:
				handleEmpresaCreditosActivacion(w, r, dbEmp, "desactivar")
				return
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func handleEmpresaCreditosCreate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload empresaCreditoCreatePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}

	if payload.ClienteID > 0 {
		cli, err := dbpkg.GetClienteByID(dbEmp, empresaID, payload.ClienteID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "cliente_id no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo validar cliente", http.StatusInternalServerError)
			return
		}
		if strings.TrimSpace(payload.ClienteNombre) == "" {
			payload.ClienteNombre = cli.NombreRazonSocial
		}
	}

	id, err := dbpkg.CreateEmpresaCredito(dbEmp, dbpkg.EmpresaCredito{
		EmpresaID:             empresaID,
		Codigo:                strings.TrimSpace(payload.Codigo),
		ClienteID:             payload.ClienteID,
		ClienteNombre:         strings.TrimSpace(payload.ClienteNombre),
		TipoCredito:           strings.TrimSpace(payload.TipoCredito),
		MontoAprobado:         payload.MontoAprobado,
		CupoCredito:           payload.CupoCredito,
		SaldoActual:           payload.SaldoActual,
		TasaInteres:           payload.TasaInteres,
		TasaMora:              payload.TasaMora,
		PeriodicidadCuota:     strings.TrimSpace(payload.PeriodicidadCuota),
		ValorCuotaPactada:     payload.ValorCuotaPactada,
		OmitirDomingos:        payload.OmitirDomingos,
		PlazoDias:             payload.PlazoDias,
		PlazoCuotas:           payload.PlazoCuotas,
		FechaInicio:           strings.TrimSpace(payload.FechaInicio),
		FechaVencimiento:      strings.TrimSpace(payload.FechaVencimiento),
		BloqueoAutomaticoMora: payload.BloqueoAutomaticoMora,
		VentaOrigenID:         payload.VentaOrigenID,
		DocumentoOrigen:       strings.TrimSpace(payload.DocumentoOrigen),
		EstadoCredito:         strings.TrimSpace(payload.EstadoCredito),
		UsuarioCreador:        strings.TrimSpace(adminEmailFromRequest(r)),
		Estado:                "activo",
		Observaciones:         strings.TrimSpace(payload.Observaciones),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	row, err := dbpkg.GetEmpresaCreditoByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "credito creado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"credito":    row,
	})
}

func handleEmpresaCreditosLimitesCliente(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	clienteID, err := parseInt64QueryOptional(r, "cliente_id")
	if err != nil {
		http.Error(w, "cliente_id invalido", http.StatusBadRequest)
		return
	}
	includeInactive := queryBool(r, "include_inactive")

	if clienteID > 0 {
		row, err := dbpkg.GetEmpresaCreditoClienteLimite(dbEmp, empresaID, clienteID, includeInactive)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "limite de cliente no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo consultar limite de cliente", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"empresa_id": empresaID,
			"limite":     row,
		})
		return
	}

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

	rows, total, err := dbpkg.ListEmpresaCreditoClienteLimites(dbEmp, empresaID, dbpkg.EmpresaCreditoClienteLimiteFilter{
		ClienteID:       clienteID,
		IncludeInactive: includeInactive,
		Limit:           limit,
		Offset:          offset,
	})
	if err != nil {
		http.Error(w, "No se pudo listar limites de cliente", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"total":      total,
		"rows":       rows,
	})
}

func handleEmpresaCreditosUpsertLimiteCliente(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var payload empresaCreditoClienteLimitePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	if payload.ClienteID <= 0 {
		if v, qErr := parseInt64QueryOptional(r, "cliente_id"); qErr == nil && v > 0 {
			payload.ClienteID = v
		}
	}
	if payload.ClienteID <= 0 {
		http.Error(w, "cliente_id es obligatorio", http.StatusBadRequest)
		return
	}

	if ok := creditosValidateFineRolePermission(w, r, dbEmp, empresaID, 0, "credito_limite_cliente_upsert", ""); !ok {
		return
	}

	_, prevErr := dbpkg.GetEmpresaCreditoClienteLimite(dbEmp, empresaID, payload.ClienteID, true)
	if prevErr != nil && prevErr != sql.ErrNoRows {
		http.Error(w, "No se pudo validar limite actual del cliente", http.StatusInternalServerError)
		return
	}
	created := prevErr == sql.ErrNoRows

	id, err := dbpkg.UpsertEmpresaCreditoClienteLimite(dbEmp, dbpkg.EmpresaCreditoClienteLimite{
		EmpresaID:                empresaID,
		ClienteID:                payload.ClienteID,
		LimiteSaldoTotal:         payload.LimiteSaldoTotal,
		MaxCreditosActivos:       payload.MaxCreditosActivos,
		RequiereAprobacionExceso: payload.RequiereAprobacionExceso,
		UsuarioCreador:           strings.TrimSpace(adminEmailFromRequest(r)),
		Estado:                   strings.TrimSpace(payload.Estado),
		Observaciones:            strings.TrimSpace(payload.Observaciones),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	row, err := dbpkg.GetEmpresaCreditoClienteLimite(dbEmp, empresaID, payload.ClienteID, true)
	if err != nil {
		http.Error(w, "limite guardado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	if row != nil && row.ID <= 0 {
		row.ID = id
	}

	statusCode := http.StatusOK
	if created {
		statusCode = http.StatusCreated
	}
	registrarAuditoriaCreditosNoBloqueante(dbEmp, r, empresaID, "creditos_limites_clientes", payload.ClienteID, "credito_limite_cliente_upsert", statusCode, map[string]interface{}{
		"cliente_id":                 payload.ClienteID,
		"limite_saldo_total":         payload.LimiteSaldoTotal,
		"max_creditos_activos":       payload.MaxCreditosActivos,
		"requiere_aprobacion_exceso": payload.RequiereAprobacionExceso,
		"creado":                     created,
	}, payload.Observaciones)

	writeJSON(w, statusCode, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"limite":     row,
	})
}

func handleEmpresaCreditosDeleteLimiteCliente(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	clienteID, err := parseInt64QueryOptional(r, "cliente_id")
	if err != nil {
		http.Error(w, "cliente_id invalido", http.StatusBadRequest)
		return
	}
	if clienteID <= 0 {
		var payload empresaCreditoClienteLimitePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err == nil || err == io.EOF {
			clienteID = payload.ClienteID
		}
	}
	if clienteID <= 0 {
		http.Error(w, "cliente_id es obligatorio", http.StatusBadRequest)
		return
	}

	if ok := creditosValidateFineRolePermission(w, r, dbEmp, empresaID, 0, "credito_limite_cliente_delete", ""); !ok {
		return
	}

	if err := dbpkg.SetEmpresaCreditoClienteLimiteRowEstado(dbEmp, empresaID, clienteID, "inactivo"); err != nil {
		http.Error(w, "No se pudo eliminar limite de cliente", http.StatusInternalServerError)
		return
	}

	registrarAuditoriaCreditosNoBloqueante(dbEmp, r, empresaID, "creditos_limites_clientes", clienteID, "credito_limite_cliente_delete", http.StatusOK, map[string]interface{}{
		"cliente_id": clienteID,
		"estado":     "inactivo",
	}, "desactivacion de limite por cliente")

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"cliente_id": clienteID,
		"estado":     "inactivo",
	})
}

func buildEmpresaCreditosFilter(r *http.Request) (dbpkg.EmpresaCreditoFilter, error) {
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		return dbpkg.EmpresaCreditoFilter{}, fmt.Errorf("limit invalido")
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		return dbpkg.EmpresaCreditoFilter{}, fmt.Errorf("offset invalido")
	}
	clienteID, err := parseInt64QueryOptional(r, "cliente_id")
	if err != nil {
		return dbpkg.EmpresaCreditoFilter{}, fmt.Errorf("cliente_id invalido")
	}

	return dbpkg.EmpresaCreditoFilter{
		ClienteID:       clienteID,
		EstadoCredito:   strings.TrimSpace(r.URL.Query().Get("estado_credito")),
		Clasificacion:   strings.TrimSpace(r.URL.Query().Get("clasificacion")),
		SoloVencidos:    queryBool(r, "solo_vencidos") || queryBool(r, "vencidos"),
		IncludeInactive: queryBool(r, "include_inactive"),
		Desde:           strings.TrimSpace(r.URL.Query().Get("desde")),
		Hasta:           strings.TrimSpace(r.URL.Query().Get("hasta")),
		Q:               strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:           limit,
		Offset:          offset,
	}, nil
}

func handleEmpresaCreditosList(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := parseInt64QueryOptional(r, "id")
	if err != nil {
		http.Error(w, "id invalido", http.StatusBadRequest)
		return
	}
	if id > 0 {
		handleEmpresaCreditosDetail(w, r, dbEmp)
		return
	}

	filter, err := buildEmpresaCreditosFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, total, err := dbpkg.ListEmpresaCreditos(dbEmp, empresaID, filter)
	if err != nil {
		http.Error(w, "No se pudo consultar creditos", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"total":      total,
		"rows":       rows,
		"filtros": map[string]interface{}{
			"cliente_id":       filter.ClienteID,
			"estado_credito":   filter.EstadoCredito,
			"clasificacion":    filter.Clasificacion,
			"solo_vencidos":    filter.SoloVencidos,
			"include_inactive": filter.IncludeInactive,
			"desde":            filter.Desde,
			"hasta":            filter.Hasta,
			"q":                filter.Q,
			"limit":            filter.Limit,
			"offset":           filter.Offset,
		},
	})
}

func resolveCreditoIDFromRequest(r *http.Request) (int64, error) {
	id, err := parseInt64QueryOptional(r, "id")
	if err != nil {
		return 0, fmt.Errorf("id invalido")
	}
	if id <= 0 {
		return 0, fmt.Errorf("id es obligatorio")
	}
	return id, nil
}

func handleEmpresaCreditosDetail(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := resolveCreditoIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	row, err := dbpkg.GetEmpresaCreditoByID(dbEmp, empresaID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "credito no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar credito", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"credito":    row,
	})
}

func handleEmpresaCreditosCuotas(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := resolveCreditoIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, err := dbpkg.ListEmpresaCreditoCuotas(dbEmp, empresaID, id, queryBool(r, "include_inactive"))
	if err != nil {
		http.Error(w, "No se pudo consultar cuotas", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"credito_id": id,
		"total":      len(rows),
		"rows":       rows,
	})
}

func handleEmpresaCreditosMovimientos(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := resolveCreditoIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	rows, err := dbpkg.ListEmpresaCreditoMovimientos(dbEmp, empresaID, id, queryBool(r, "include_inactive"), limit)
	if err != nil {
		http.Error(w, "No se pudo consultar movimientos", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"credito_id": id,
		"total":      len(rows),
		"rows":       rows,
	})
}

func handleEmpresaCreditosEstadoCuenta(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := resolveCreditoIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	credito, err := dbpkg.GetEmpresaCreditoByID(dbEmp, empresaID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "credito no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar credito", http.StatusInternalServerError)
		return
	}
	cuotas, err := dbpkg.ListEmpresaCreditoCuotas(dbEmp, empresaID, id, false)
	if err != nil {
		http.Error(w, "No se pudo consultar cuotas", http.StatusInternalServerError)
		return
	}
	movs, err := dbpkg.ListEmpresaCreditoMovimientos(dbEmp, empresaID, id, false, 200)
	if err != nil {
		http.Error(w, "No se pudo consultar movimientos", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"empresa_id":  empresaID,
		"credito":     credito,
		"cuotas":      cuotas,
		"movimientos": movs,
	})
}

func handleEmpresaCreditosResumen(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resumen, err := dbpkg.GetEmpresaCreditosCarteraResumen(dbEmp, empresaID, queryBool(r, "include_inactive"))
	if err != nil {
		http.Error(w, "No se pudo calcular resumen de cartera", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"resumen":    resumen,
	})
}

func parseCreditosMoraParams(r *http.Request) (int, int, error) {
	diasProximos, err := parseIntQueryOptional(r, "dias_proximos")
	if err != nil {
		return 0, 0, fmt.Errorf("dias_proximos invalido")
	}
	if diasProximos <= 0 {
		diasProximos = 7
	}
	if diasProximos > 365 {
		diasProximos = 365
	}

	top, err := parseIntQueryOptional(r, "top")
	if err != nil {
		return 0, 0, fmt.Errorf("top invalido")
	}
	if top <= 0 {
		top = 10
	}
	if top > 200 {
		top = 200
	}

	return diasProximos, top, nil
}

func handleEmpresaCreditosAlertasMora(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	diasProximos, top, err := parseCreditosMoraParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	alertas, err := dbpkg.GetEmpresaCreditosMoraDashboard(dbEmp, empresaID, diasProximos, top, queryBool(r, "include_inactive"))
	if err != nil {
		http.Error(w, "No se pudo consultar alertas de morosidad", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"alertas":    alertas,
	})
}

func buildEmpresaCreditosMorosidadDataset(empresaID int64, alertas *dbpkg.EmpresaCreditosMoraDashboard) empresaReporteDataset {
	rows := make([]map[string]interface{}, 0)
	addRows := func(grupo string, creditos []dbpkg.EmpresaCredito) {
		for _, row := range creditos {
			rows = append(rows, map[string]interface{}{
				"grupo":                 grupo,
				"id":                    row.ID,
				"codigo":                row.Codigo,
				"cliente_id":            row.ClienteID,
				"cliente_nombre":        row.ClienteNombre,
				"saldo_actual":          row.SaldoActual,
				"dias_mora":             row.DiasMora,
				"cuotas_vencidas":       row.CuotasVencidas,
				"dias_cuotas_vencidas":  row.DiasCuotasVencidas,
				"fecha_proxima_cuota":   row.FechaProximaCuota,
				"fecha_vencimiento":     row.FechaVencimiento,
				"estado_credito":        row.EstadoCredito,
				"clasificacion_cartera": row.ClasificacionCartera,
			})
		}
	}
	if alertas != nil {
		addRows("proximos_vencer", alertas.ProximosVencer)
		addRows("vencidos", alertas.Vencidos)
		addRows("ranking_morosidad", alertas.RankingMorosidad)
	}

	summary := map[string]interface{}{}
	if alertas != nil {
		summary = map[string]interface{}{
			"dias_proximos":           alertas.DiasProximos,
			"top":                     alertas.Top,
			"total_proximos_vencer":   alertas.TotalProximosVencer,
			"total_vencidos":          alertas.TotalVencidos,
			"total_ranking_morosidad": alertas.TotalRankingMorosidad,
			"monto_proximos_vencer":   alertas.MontoProximosVencer,
			"monto_vencidos":          alertas.MontoVencidos,
			"monto_ranking_morosidad": alertas.MontoRankingMorosidad,
			"generado_en_alertas":     alertas.GeneradoEn,
		}
	}

	return empresaReporteDataset{
		Key:         "finanzas_creditos_morosidad",
		Title:       "Alertas y Ranking de Morosidad de Creditos",
		Level:       "financiero",
		Description: "Alertas proactivas de vencimiento y ranking avanzado de morosidad por empresa.",
		EmpresaID:   empresaID,
		GeneratedAt: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Columns: []string{
			"grupo", "id", "codigo", "cliente_id", "cliente_nombre", "saldo_actual", "dias_mora", "cuotas_vencidas", "dias_cuotas_vencidas", "fecha_proxima_cuota", "fecha_vencimiento", "estado_credito", "clasificacion_cartera",
		},
		Rows:     rows,
		RowCount: len(rows),
		Summary:  summary,
	}
}

func handleEmpresaCreditosAbono(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var payload empresaCreditoAbonoPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	if payload.CreditoID <= 0 {
		if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
			payload.CreditoID = id
		}
	}
	if payload.CreditoID <= 0 {
		http.Error(w, "credito_id es obligatorio", http.StatusBadRequest)
		return
	}

	policy, err := creditosResolveContablePolicy(r, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	usuario := strings.TrimSpace(adminEmailFromRequest(r))

	movID, updated, err := dbpkg.RegisterEmpresaCreditoAbono(dbEmp, dbpkg.EmpresaCreditoAbonoInput{
		EmpresaID:       empresaID,
		CreditoID:       payload.CreditoID,
		Monto:           payload.Monto,
		MetodoPago:      strings.TrimSpace(payload.MetodoPago),
		ReferenciaPago:  strings.TrimSpace(payload.ReferenciaPago),
		Comprobante:     strings.TrimSpace(payload.Comprobante),
		Observaciones:   strings.TrimSpace(payload.Observaciones),
		FechaMovimiento: strings.TrimSpace(payload.FechaMovimiento),
		UsuarioCreador:  usuario,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	integracionContable := map[string]interface{}{
		"registrar_evento_contable": policy.RegistrarEventoContable,
		"procesar_asientos":         policy.ProcesarAsientos,
		"asientos_limit":            policy.AsientosLimit,
		"max_reintentos":            policy.MaxReintentos,
		"evento_registrado":         false,
		"asientos_procesados":       false,
	}

	if policy.RegistrarEventoContable {
		if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
			integracionContable["error_evento_contable"] = "no se pudo preparar esquema de eventos contables"
		} else {
			movimiento, okMov := creditosFindMovimientoByID(dbEmp, empresaID, payload.CreditoID, movID)
			montoEvento := payload.Monto
			capitalAplicado := 0.0
			interesAplicado := 0.0
			moraAplicada := 0.0
			metodoPago := strings.TrimSpace(payload.MetodoPago)
			referenciaPago := strings.TrimSpace(payload.ReferenciaPago)
			comprobante := strings.TrimSpace(payload.Comprobante)

			if okMov {
				montoEvento = movimiento.Monto
				capitalAplicado = movimiento.CapitalAplicado
				interesAplicado = movimiento.InteresAplicado
				moraAplicada = movimiento.MoraAplicada
				if metodoPago == "" {
					metodoPago = strings.TrimSpace(movimiento.MetodoPago)
				}
				if referenciaPago == "" {
					referenciaPago = strings.TrimSpace(movimiento.ReferenciaPago)
				}
				if comprobante == "" {
					comprobante = strings.TrimSpace(movimiento.Comprobante)
				}
			}

			periodoContable := creditosResolvePeriodoContableAbono(strings.TrimSpace(payload.FechaMovimiento))
			if periodoContable == "" {
				periodoContable = time.Now().In(time.Local).Format("2006-01")
			}

			codigoCredito := fmt.Sprintf("CR-%d", payload.CreditoID)
			if updated != nil && strings.TrimSpace(updated.Codigo) != "" {
				codigoCredito = strings.TrimSpace(updated.Codigo)
			}

			canalPago := creditosResolveCanalPago(metodoPago)
			eventoPayload := map[string]interface{}{
				"empresa_id":        empresaID,
				"credito_id":        payload.CreditoID,
				"credito_codigo":    codigoCredito,
				"movimiento_id":     movID,
				"tipo_movimiento":   "abono",
				"canal_pago":        canalPago,
				"metodo_pago":       metodoPago,
				"referencia_pago":   referenciaPago,
				"comprobante":       comprobante,
				"periodo_contable":  periodoContable,
				"monto_total":       montoEvento,
				"capital_aplicado":  capitalAplicado,
				"interes_aplicado":  interesAplicado,
				"mora_aplicada":     moraAplicada,
				"categoria":         "creditos",
				"categoria_interes": "intereses_credito",
				"categoria_mora":    "mora_credito",
			}
			eventoPayloadJSON := ""
			if b, mErr := json.Marshal(eventoPayload); mErr == nil {
				eventoPayloadJSON = string(b)
			}

			eventoID, evtErr := dbpkg.CreateEmpresaEventoContable(dbEmp, dbpkg.EmpresaEventoContable{
				EmpresaID:       empresaID,
				Modulo:          "creditos",
				Evento:          "credito_abono_registrado",
				Entidad:         "credito_movimiento",
				EntidadID:       movID,
				DocumentoTipo:   "credito",
				DocumentoCodigo: codigoCredito,
				PeriodoContable: periodoContable,
				MontoTotal:      montoEvento,
				Moneda:          "COP",
				PayloadJSON:     eventoPayloadJSON,
				Origen:          "api_creditos_abono",
				UsuarioCreador:  usuario,
				Estado:          "activo",
				Observaciones:   strings.TrimSpace(payload.Observaciones),
			})
			if evtErr != nil {
				integracionContable["error_evento_contable"] = evtErr.Error()
			} else {
				integracionContable["evento_registrado"] = true
				integracionContable["evento_contable_id"] = eventoID

				if policy.ProcesarAsientos {
					if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
						integracionContable["error_asientos"] = "no se pudo preparar esquema financiero"
					} else {
						resultadoAsientos, procErr := dbpkg.ProcessEmpresaEventosContablesPendientesConPolitica(dbEmp, empresaID, usuario, policy.AsientosLimit, policy.MaxReintentos)
						if procErr != nil {
							integracionContable["error_asientos"] = procErr.Error()
						} else {
							integracionContable["asientos_procesados"] = true
							integracionContable["procesamiento_asientos"] = map[string]interface{}{
								"eventos_revisados":   resultadoAsientos.EventosRevisados,
								"eventos_procesados":  resultadoAsientos.EventosProcesados,
								"asientos_creados":    resultadoAsientos.AsientosCreados,
								"asientos_existentes": resultadoAsientos.AsientosExistentes,
								"fallidos":            resultadoAsientos.Fallidos,
								"errores":             resultadoAsientos.Errores,
							}
						}
					}
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                   true,
		"empresa_id":           empresaID,
		"credito_id":           payload.CreditoID,
		"movimiento_id":        movID,
		"credito":              updated,
		"integracion_contable": integracionContable,
	})
}

func handleEmpresaCreditosEstado(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := resolveCreditoIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	estado := strings.TrimSpace(r.URL.Query().Get("estado_credito"))
	if estado == "" {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			estado = strings.TrimSpace(creditoAnyToString(body["estado_credito"]))
		}
	}
	if estado == "" {
		http.Error(w, "estado_credito es obligatorio", http.StatusBadRequest)
		return
	}
	if err := dbpkg.SetEmpresaCreditoEstado(dbEmp, empresaID, id, estado); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	row, err := dbpkg.GetEmpresaCreditoByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "estado actualizado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"credito":    row,
	})
}

func handleEmpresaCreditosActivacion(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, action string) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := resolveCreditoIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	estado := "activo"
	if action == "desactivar" {
		estado = "inactivo"
	}
	if err := dbpkg.SetEmpresaCreditoRowEstado(dbEmp, empresaID, id, estado); err != nil {
		http.Error(w, "No se pudo actualizar estado", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"id":         id,
		"estado":     estado,
	})
}

func handleEmpresaCreditosUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := resolveCreditoIDFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	current, err := dbpkg.GetEmpresaCreditoByID(dbEmp, empresaID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "credito no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo consultar credito", http.StatusInternalServerError)
		return
	}

	body := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	if v, ok := body["codigo"]; ok {
		current.Codigo = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["cliente_id"]; ok {
		current.ClienteID = creditoAnyToInt64(v)
	}
	if v, ok := body["cliente_nombre"]; ok {
		current.ClienteNombre = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["tipo_credito"]; ok {
		current.TipoCredito = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["monto_aprobado"]; ok {
		current.MontoAprobado = creditoAnyToFloat64(v)
	}
	if v, ok := body["cupo_credito"]; ok {
		current.CupoCredito = creditoAnyToFloat64(v)
	}
	if v, ok := body["saldo_actual"]; ok {
		current.SaldoActual = creditoAnyToFloat64(v)
	}
	if v, ok := body["tasa_interes"]; ok {
		current.TasaInteres = creditoAnyToFloat64(v)
	}
	if v, ok := body["tasa_mora"]; ok {
		current.TasaMora = creditoAnyToFloat64(v)
	}
	if v, ok := body["periodicidad_cuota"]; ok {
		current.PeriodicidadCuota = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["valor_cuota_pactada"]; ok {
		current.ValorCuotaPactada = creditoAnyToFloat64(v)
	}
	if v, ok := body["omitir_domingos"]; ok {
		current.OmitirDomingos = creditoAnyToBool(v)
	}
	if v, ok := body["plazo_dias"]; ok {
		current.PlazoDias = creditoAnyToInt(v)
	}
	if v, ok := body["plazo_cuotas"]; ok {
		current.PlazoCuotas = creditoAnyToInt(v)
	}
	if v, ok := body["fecha_inicio"]; ok {
		current.FechaInicio = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["fecha_vencimiento"]; ok {
		current.FechaVencimiento = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["bloqueo_automatico_mora"]; ok {
		current.BloqueoAutomaticoMora = creditoAnyToBool(v)
	}
	if v, ok := body["venta_origen_id"]; ok {
		current.VentaOrigenID = creditoAnyToInt64(v)
	}
	if v, ok := body["documento_origen"]; ok {
		current.DocumentoOrigen = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["estado_credito"]; ok {
		current.EstadoCredito = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["estado"]; ok {
		current.Estado = strings.TrimSpace(creditoAnyToString(v))
	}
	if v, ok := body["observaciones"]; ok {
		current.Observaciones = strings.TrimSpace(creditoAnyToString(v))
	}
	current.UsuarioCreador = strings.TrimSpace(adminEmailFromRequest(r))

	if err := dbpkg.UpdateEmpresaCredito(dbEmp, *current); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := dbpkg.GetEmpresaCreditoByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "actualizado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"credito":    updated,
	})
}

func handleEmpresaCreditosReporte(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tipoReporte := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("tipo")))
	tipoReporte = strings.ReplaceAll(tipoReporte, "-", "_")
	filter, err := buildEmpresaCreditosFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if filter.Limit <= 0 {
		filter.Limit = 2000
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if tipoReporte == "morosidad" || tipoReporte == "alertas_mora" || tipoReporte == "ranking_morosidad" {
		diasProximos, top, paramsErr := parseCreditosMoraParams(r)
		if paramsErr != nil {
			http.Error(w, paramsErr.Error(), http.StatusBadRequest)
			return
		}
		alertas, alertErr := dbpkg.GetEmpresaCreditosMoraDashboard(dbEmp, empresaID, diasProximos, top, filter.IncludeInactive)
		if alertErr != nil {
			http.Error(w, "No se pudo calcular reporte de morosidad", http.StatusInternalServerError)
			return
		}

		ds := buildEmpresaCreditosMorosidadDataset(empresaID, alertas)
		if filter.Desde != "" {
			ds.Desde = filter.Desde
		}
		if filter.Hasta != "" {
			ds.Hasta = filter.Hasta
		}

		format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
		if format == "" {
			format = "json"
		}
		if err := writeReportesDatasetExport(w, ds, format); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		return
	}

	rows, total, err := dbpkg.ListEmpresaCreditos(dbEmp, empresaID, filter)
	if err != nil {
		http.Error(w, "No se pudo consultar creditos para reporte", http.StatusInternalServerError)
		return
	}
	resumen, err := dbpkg.GetEmpresaCreditosCarteraResumen(dbEmp, empresaID, filter.IncludeInactive)
	if err != nil {
		http.Error(w, "No se pudo calcular resumen de cartera", http.StatusInternalServerError)
		return
	}

	datasetRows := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		datasetRows = append(datasetRows, map[string]interface{}{
			"id":                    row.ID,
			"codigo":                row.Codigo,
			"cliente_id":            row.ClienteID,
			"cliente_nombre":        row.ClienteNombre,
			"tipo_credito":          row.TipoCredito,
			"monto_aprobado":        row.MontoAprobado,
			"cupo_credito":          row.CupoCredito,
			"saldo_actual":          row.SaldoActual,
			"saldo_disponible":      row.SaldoDisponible,
			"tasa_interes":          row.TasaInteres,
			"tasa_mora":             row.TasaMora,
			"periodicidad_cuota":    row.PeriodicidadCuota,
			"valor_cuota_pactada":   row.ValorCuotaPactada,
			"omitir_domingos":       row.OmitirDomingos,
			"plazo_dias":            row.PlazoDias,
			"plazo_cuotas":          row.PlazoCuotas,
			"fecha_inicio":          row.FechaInicio,
			"fecha_vencimiento":     row.FechaVencimiento,
			"dias_mora":             row.DiasMora,
			"cuotas_vencidas":       row.CuotasVencidas,
			"dias_cuotas_vencidas":  row.DiasCuotasVencidas,
			"fecha_proxima_cuota":   row.FechaProximaCuota,
			"estado_credito":        row.EstadoCredito,
			"clasificacion_cartera": row.ClasificacionCartera,
			"estado":                row.Estado,
		})
	}

	ds := empresaReporteDataset{
		Key:         "finanzas_creditos_cartera",
		Title:       "Reporte de Creditos y Cartera",
		Level:       "financiero",
		Description: "Cartera de creditos por empresa con saldos, mora y clasificacion.",
		EmpresaID:   empresaID,
		Desde:       filter.Desde,
		Hasta:       filter.Hasta,
		GeneratedAt: time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
		Columns: []string{
			"id", "codigo", "cliente_id", "cliente_nombre", "tipo_credito", "monto_aprobado", "cupo_credito", "saldo_actual", "saldo_disponible",
			"tasa_interes", "tasa_mora", "periodicidad_cuota", "valor_cuota_pactada", "omitir_domingos", "plazo_dias", "plazo_cuotas", "fecha_inicio", "fecha_vencimiento", "dias_mora", "cuotas_vencidas", "dias_cuotas_vencidas", "fecha_proxima_cuota", "estado_credito", "clasificacion_cartera", "estado",
		},
		Rows:     datasetRows,
		RowCount: len(datasetRows),
		Summary: map[string]interface{}{
			"total_coincidencias": total,
			"resumen_cartera":     resumen,
		},
	}

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "" {
		format = "json"
	}
	if err := writeReportesDatasetExport(w, ds, format); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func buildEmpresaCreditoWorkflowFilter(r *http.Request) (dbpkg.EmpresaCreditoWorkflowFilter, error) {
	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		return dbpkg.EmpresaCreditoWorkflowFilter{}, fmt.Errorf("limit invalido")
	}
	offset, err := parseIntQueryOptional(r, "offset")
	if err != nil {
		return dbpkg.EmpresaCreditoWorkflowFilter{}, fmt.Errorf("offset invalido")
	}
	creditoID, err := parseInt64QueryOptional(r, "credito_id")
	if err != nil {
		return dbpkg.EmpresaCreditoWorkflowFilter{}, fmt.Errorf("credito_id invalido")
	}

	return dbpkg.EmpresaCreditoWorkflowFilter{
		CreditoID:       creditoID,
		TipoSolicitud:   strings.TrimSpace(r.URL.Query().Get("tipo_solicitud")),
		EstadoSolicitud: strings.TrimSpace(r.URL.Query().Get("estado_solicitud")),
		IncludeInactive: queryBool(r, "include_inactive"),
		Limit:           limit,
		Offset:          offset,
	}, nil
}

func handleEmpresaCreditosWorkflows(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	workflowID, err := parseInt64QueryOptional(r, "workflow_id")
	if err != nil {
		http.Error(w, "workflow_id invalido", http.StatusBadRequest)
		return
	}
	if workflowID <= 0 {
		workflowID, err = parseInt64QueryOptional(r, "id")
		if err != nil {
			http.Error(w, "id invalido", http.StatusBadRequest)
			return
		}
	}

	if workflowID > 0 {
		row, err := dbpkg.GetEmpresaCreditoWorkflowByID(dbEmp, empresaID, workflowID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "workflow no encontrado", http.StatusNotFound)
				return
			}
			http.Error(w, "No se pudo consultar workflow", http.StatusInternalServerError)
			return
		}
		credito, _ := dbpkg.GetEmpresaCreditoByID(dbEmp, empresaID, row.CreditoID)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"empresa_id": empresaID,
			"workflow":   row,
			"credito":    credito,
		})
		return
	}

	filter, err := buildEmpresaCreditoWorkflowFilter(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rows, total, err := dbpkg.ListEmpresaCreditoWorkflows(dbEmp, empresaID, filter)
	if err != nil {
		http.Error(w, "No se pudo consultar workflows", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"total":      total,
		"rows":       rows,
		"filtros": map[string]interface{}{
			"credito_id":       filter.CreditoID,
			"tipo_solicitud":   filter.TipoSolicitud,
			"estado_solicitud": filter.EstadoSolicitud,
			"include_inactive": filter.IncludeInactive,
			"limit":            filter.Limit,
			"offset":           filter.Offset,
		},
	})
}

func decodeEmpresaCreditoWorkflowSolicitud(r *http.Request) (empresaCreditoWorkflowSolicitudPayload, map[string]interface{}, error) {
	rawMap, _ := extractJSONBodyMap(r)
	var payload empresaCreditoWorkflowSolicitudPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return empresaCreditoWorkflowSolicitudPayload{}, nil, err
	}
	if rawMap == nil {
		rawMap = map[string]interface{}{}
	}
	return payload, rawMap, nil
}

func handleEmpresaCreditosSolicitarReverso(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload, _, err := decodeEmpresaCreditoWorkflowSolicitud(r)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	if payload.CreditoID <= 0 {
		if v, qErr := parseInt64QueryOptional(r, "credito_id"); qErr == nil && v > 0 {
			payload.CreditoID = v
		}
	}
	if payload.MovimientoOrigenID <= 0 {
		if v, qErr := parseInt64QueryOptional(r, "movimiento_origen_id"); qErr == nil && v > 0 {
			payload.MovimientoOrigenID = v
		}
	}
	if payload.CreditoID <= 0 {
		http.Error(w, "credito_id es obligatorio", http.StatusBadRequest)
		return
	}
	if payload.MovimientoOrigenID <= 0 {
		http.Error(w, "movimiento_origen_id es obligatorio", http.StatusBadRequest)
		return
	}

	payloadJSON := "{}"
	if payload.Payload != nil {
		if raw, mErr := json.Marshal(payload.Payload); mErr == nil {
			payloadJSON = string(raw)
		}
	}

	id, err := dbpkg.CreateEmpresaCreditoWorkflowSolicitud(dbEmp, dbpkg.EmpresaCreditoWorkflowSolicitudInput{
		EmpresaID:                empresaID,
		CreditoID:                payload.CreditoID,
		TipoSolicitud:            "reverso_abono",
		MovimientoOrigenID:       payload.MovimientoOrigenID,
		NivelAprobacionRequerido: payload.NivelAprobacionRequerido,
		MotivoSolicitud:          strings.TrimSpace(payload.MotivoSolicitud),
		PayloadJSON:              payloadJSON,
		UsuarioCreador:           strings.TrimSpace(adminEmailFromRequest(r)),
		Observaciones:            strings.TrimSpace(payload.Observaciones),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	workflow, err := dbpkg.GetEmpresaCreditoWorkflowByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "workflow creado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	registrarAuditoriaCreditosNoBloqueante(dbEmp, r, empresaID, "empresa_creditos_workflow", id, "credito_workflow_solicitado_reverso", http.StatusCreated, map[string]interface{}{
		"workflow_id":          id,
		"credito_id":           payload.CreditoID,
		"tipo_solicitud":       "reverso_abono",
		"movimiento_origen_id": payload.MovimientoOrigenID,
	}, payload.Observaciones)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"workflow":   workflow,
	})
}

func handleEmpresaCreditosSolicitarRefinanciacion(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload, rawMap, err := decodeEmpresaCreditoWorkflowSolicitud(r)
	if err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	if payload.CreditoID <= 0 {
		if v, qErr := parseInt64QueryOptional(r, "credito_id"); qErr == nil && v > 0 {
			payload.CreditoID = v
		}
	}
	if payload.CreditoID <= 0 {
		http.Error(w, "credito_id es obligatorio", http.StatusBadRequest)
		return
	}

	payloadMap := map[string]interface{}{}
	for k, v := range rawMap {
		if k == "empresa_id" || k == "credito_id" || k == "movimiento_origen_id" || k == "nivel_aprobacion_requerido" || k == "motivo_solicitud" || k == "observaciones" || k == "payload" {
			continue
		}
		payloadMap[k] = v
	}
	for k, v := range payload.Payload {
		payloadMap[k] = v
	}
	if len(payloadMap) == 0 {
		http.Error(w, "se requiere payload de refinanciacion (ej: nuevo_plazo_cuotas, nueva_tasa_interes)", http.StatusBadRequest)
		return
	}
	payloadJSON := "{}"
	if raw, mErr := json.Marshal(payloadMap); mErr == nil {
		payloadJSON = string(raw)
	}

	id, err := dbpkg.CreateEmpresaCreditoWorkflowSolicitud(dbEmp, dbpkg.EmpresaCreditoWorkflowSolicitudInput{
		EmpresaID:                empresaID,
		CreditoID:                payload.CreditoID,
		TipoSolicitud:            "refinanciacion",
		NivelAprobacionRequerido: payload.NivelAprobacionRequerido,
		MotivoSolicitud:          strings.TrimSpace(payload.MotivoSolicitud),
		PayloadJSON:              payloadJSON,
		UsuarioCreador:           strings.TrimSpace(adminEmailFromRequest(r)),
		Observaciones:            strings.TrimSpace(payload.Observaciones),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	workflow, err := dbpkg.GetEmpresaCreditoWorkflowByID(dbEmp, empresaID, id)
	if err != nil {
		http.Error(w, "workflow creado pero no se pudo consultar", http.StatusInternalServerError)
		return
	}
	registrarAuditoriaCreditosNoBloqueante(dbEmp, r, empresaID, "empresa_creditos_workflow", id, "credito_workflow_solicitado_refinanciacion", http.StatusCreated, map[string]interface{}{
		"workflow_id":    id,
		"credito_id":     payload.CreditoID,
		"tipo_solicitud": "refinanciacion",
		"payload":        payloadMap,
	}, payload.Observaciones)
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"workflow":   workflow,
	})
}

func resolveCreditoWorkflowID(r *http.Request, payloadID int64) (int64, error) {
	if payloadID > 0 {
		return payloadID, nil
	}
	if id, err := parseInt64QueryOptional(r, "workflow_id"); err == nil && id > 0 {
		return id, nil
	}
	if id, err := parseInt64QueryOptional(r, "id"); err == nil && id > 0 {
		return id, nil
	}
	return 0, fmt.Errorf("workflow_id es obligatorio")
}

func handleEmpresaCreditosAprobarWorkflow(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	evidence, err := extractPermissionApprovalEvidence(r)
	if err != nil {
		http.Error(w, "no se pudo leer evidencia de aprobacion", http.StatusBadRequest)
		return
	}

	var payload empresaCreditoWorkflowDecisionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	workflowID, err := resolveCreditoWorkflowID(r, payload.WorkflowID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	workflowActual, err := dbpkg.GetEmpresaCreditoWorkflowByID(dbEmp, empresaID, workflowID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "workflow no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo validar workflow", http.StatusInternalServerError)
		return
	}
	if ok := creditosValidateFineRolePermission(w, r, dbEmp, empresaID, workflowID, creditosWorkflowFineAction("aprobar", workflowActual.TipoSolicitud), workflowActual.TipoSolicitud); !ok {
		return
	}

	aprobadoPor := strings.TrimSpace(payload.AprobadoPor)
	if aprobadoPor == "" {
		aprobadoPor = strings.TrimSpace(evidence.ApprovedBy)
	}
	codigoAprobacion := strings.TrimSpace(payload.CodigoAprobacion)
	if codigoAprobacion == "" {
		codigoAprobacion = strings.TrimSpace(evidence.ApprovalCode)
	}
	motivoAprobacion := strings.TrimSpace(payload.MotivoAprobacion)
	if motivoAprobacion == "" {
		motivoAprobacion = strings.TrimSpace(evidence.Reason)
	}

	workflow, err := dbpkg.AprobarEmpresaCreditoWorkflow(dbEmp, dbpkg.EmpresaCreditoWorkflowAprobacionInput{
		EmpresaID:        empresaID,
		WorkflowID:       workflowID,
		AprobadoPor:      aprobadoPor,
		CodigoAprobacion: codigoAprobacion,
		MotivoAprobacion: motivoAprobacion,
		EjecutadoPor:     strings.TrimSpace(adminEmailFromRequest(r)),
		UsuarioCreador:   strings.TrimSpace(adminEmailFromRequest(r)),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	registrarAuditoriaCreditosNoBloqueante(dbEmp, r, empresaID, "empresa_creditos_workflow", workflowID, "credito_workflow_aprobado", http.StatusOK, map[string]interface{}{
		"workflow_id":             workflowID,
		"credito_id":              workflow.CreditoID,
		"tipo_solicitud":          workflow.TipoSolicitud,
		"estado_solicitud":        workflow.EstadoSolicitud,
		"movimiento_resultado_id": workflow.MovimientoResultadoID,
		"aprobado_por":            aprobadoPor,
	}, payload.Observaciones)

	credito, _ := dbpkg.GetEmpresaCreditoByID(dbEmp, empresaID, workflow.CreditoID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"workflow":   workflow,
		"credito":    credito,
	})
}

func handleEmpresaCreditosRechazarWorkflow(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB) {
	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	evidence, err := extractPermissionApprovalEvidence(r)
	if err != nil {
		http.Error(w, "no se pudo leer evidencia de aprobacion", http.StatusBadRequest)
		return
	}

	var payload empresaCreditoWorkflowDecisionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID > 0 && payload.EmpresaID != empresaID {
		http.Error(w, "empresa_id no coincide con contexto", http.StatusBadRequest)
		return
	}
	workflowID, err := resolveCreditoWorkflowID(r, payload.WorkflowID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	workflowActual, err := dbpkg.GetEmpresaCreditoWorkflowByID(dbEmp, empresaID, workflowID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "workflow no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "No se pudo validar workflow", http.StatusInternalServerError)
		return
	}
	if ok := creditosValidateFineRolePermission(w, r, dbEmp, empresaID, workflowID, creditosWorkflowFineAction("rechazar", workflowActual.TipoSolicitud), workflowActual.TipoSolicitud); !ok {
		return
	}

	aprobadoPor := strings.TrimSpace(payload.AprobadoPor)
	if aprobadoPor == "" {
		aprobadoPor = strings.TrimSpace(evidence.ApprovedBy)
	}
	codigoAprobacion := strings.TrimSpace(payload.CodigoAprobacion)
	if codigoAprobacion == "" {
		codigoAprobacion = strings.TrimSpace(evidence.ApprovalCode)
	}
	motivoRechazo := strings.TrimSpace(payload.MotivoRechazo)
	if motivoRechazo == "" {
		motivoRechazo = strings.TrimSpace(evidence.Reason)
	}

	workflow, err := dbpkg.RechazarEmpresaCreditoWorkflow(dbEmp, dbpkg.EmpresaCreditoWorkflowAprobacionInput{
		EmpresaID:        empresaID,
		WorkflowID:       workflowID,
		AprobadoPor:      aprobadoPor,
		CodigoAprobacion: codigoAprobacion,
		MotivoRechazo:    motivoRechazo,
		EjecutadoPor:     strings.TrimSpace(adminEmailFromRequest(r)),
		UsuarioCreador:   strings.TrimSpace(adminEmailFromRequest(r)),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	registrarAuditoriaCreditosNoBloqueante(dbEmp, r, empresaID, "empresa_creditos_workflow", workflowID, "credito_workflow_rechazado", http.StatusOK, map[string]interface{}{
		"workflow_id":      workflowID,
		"credito_id":       workflow.CreditoID,
		"tipo_solicitud":   workflow.TipoSolicitud,
		"estado_solicitud": workflow.EstadoSolicitud,
		"aprobado_por":     aprobadoPor,
		"motivo_rechazo":   motivoRechazo,
	}, payload.Observaciones)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"workflow":   workflow,
	})
}

func creditoAnyToString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func creditoAnyToFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return 0
	}
	s = strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func creditoAnyToInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", v))
	if s == "" {
		return 0
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return int64(f)
	}
	return 0
}

func creditoAnyToInt(v interface{}) int {
	return int(creditoAnyToInt64(v))
}

func creditoAnyToBool(v interface{}) bool {
	s := strings.ToLower(strings.TrimSpace(creditoAnyToString(v)))
	switch s {
	case "1", "true", "si", "yes", "on":
		return true
	default:
		return false
	}
}

func creditosWorkflowFineAction(decision, tipoSolicitud string) string {
	decision = strings.ToLower(strings.TrimSpace(decision))
	tipoSolicitud = strings.ToLower(strings.TrimSpace(tipoSolicitud))
	if tipoSolicitud == "" {
		tipoSolicitud = "workflow"
	}
	if decision == "" {
		decision = "operar"
	}
	return "credito_workflow_" + decision + "_" + tipoSolicitud
}

func creditosFineRoleAllowsAction(role, fineAction, tipoSolicitud string) bool {
	role = normalizePermissionRole(role)
	if role == "" || role == "super_administrador" {
		return true
	}

	fineAction = strings.ToLower(strings.TrimSpace(fineAction))
	tipoSolicitud = strings.ToLower(strings.TrimSpace(tipoSolicitud))

	if fineAction == "credito_limite_cliente_delete" {
		return roleIn(role, "contabilidad")
	}
	if strings.HasPrefix(fineAction, "credito_limite_cliente_") {
		return roleIn(role, "admin_empresa", "contabilidad")
	}

	if strings.HasPrefix(fineAction, "credito_workflow_aprobar_") || strings.HasPrefix(fineAction, "credito_workflow_rechazar_") {
		switch tipoSolicitud {
		case "refinanciacion":
			return roleIn(role, "admin_empresa")
		case "reverso_abono":
			return roleIn(role, "admin_empresa", "contabilidad")
		default:
			return roleIn(role, "admin_empresa")
		}
	}

	return true
}

func creditosValidateFineRolePermission(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID, recursoID int64, fineAction, tipoSolicitud string) bool {
	role := normalizePermissionRole(adminRoleFromRequest(r))
	if creditosFineRoleAllowsAction(role, fineAction, tipoSolicitud) {
		return true
	}
	recurso := "empresa_creditos_workflow"
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(fineAction)), "credito_limite_cliente_") {
		recurso = "creditos_limites_clientes"
	}

	metadata := map[string]interface{}{
		"fine_action":    fineAction,
		"tipo_solicitud": tipoSolicitud,
		"role":           role,
	}
	registrarAuditoriaCreditosNoBloqueante(dbEmp, r, empresaID, recurso, recursoID, "credito_permiso_fino_denegado", http.StatusForbidden, metadata, "permiso fino insuficiente para operacion de creditos")
	http.Error(w, "forbidden: rol sin permiso fino para la accion solicitada", http.StatusForbidden)
	return false
}

func registrarAuditoriaCreditosNoBloqueante(dbEmp *sql.DB, r *http.Request, empresaID int64, recurso string, recursoID int64, accion string, statusCode int, metadata map[string]interface{}, observaciones string) {
	if dbEmp == nil || empresaID <= 0 {
		return
	}
	accion = strings.TrimSpace(strings.ToLower(accion))
	if accion == "" {
		return
	}
	recurso = strings.TrimSpace(strings.ToLower(recurso))
	if recurso == "" {
		recurso = "empresa_creditos"
	}

	metadataJSON := "{}"
	if metadata != nil {
		if raw, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(raw)
		}
	}

	auditoria := dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "finanzas",
		Accion:         accion,
		Recurso:        recurso,
		RecursoID:      recursoID,
		MetodoHTTP:     strings.ToUpper(strings.TrimSpace(r.Method)),
		Endpoint:       strings.TrimSpace(r.URL.Path),
		Resultado:      resolveAuditoriaResultado(statusCode),
		CodigoHTTP:     int64(statusCode),
		RequestID:      resolveAuditoriaRequestID(r),
		IPOrigen:       resolveAuditoriaIP(r),
		UserAgent:      strings.TrimSpace(r.UserAgent()),
		MetadataJSON:   metadataJSON,
		RetencionDias:  normalizeRetencionDiasForHandler(0),
		UsuarioCreador: strings.TrimSpace(adminEmailFromRequest(r)),
		Estado:         "activo",
		Observaciones:  strings.TrimSpace(observaciones),
	}

	_, _ = dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, auditoria)
}

type creditosContablePolicy struct {
	RegistrarEventoContable bool
	ProcesarAsientos        bool
	AsientosLimit           int
	MaxReintentos           int
}

func creditosResolveContablePolicy(r *http.Request, payload empresaCreditoAbonoPayload) (creditosContablePolicy, error) {
	policy := creditosContablePolicy{
		RegistrarEventoContable: true,
		ProcesarAsientos:        true,
		AsientosLimit:           20,
		MaxReintentos:           0,
	}

	if payload.RegistrarEventoContable != nil {
		policy.RegistrarEventoContable = *payload.RegistrarEventoContable
	}
	if payload.ProcesarAsientos != nil {
		policy.ProcesarAsientos = *payload.ProcesarAsientos
	}
	if payload.AsientosLimit != nil {
		policy.AsientosLimit = *payload.AsientosLimit
	}
	if payload.MaxReintentos != nil {
		policy.MaxReintentos = *payload.MaxReintentos
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("registrar_evento_contable")); raw != "" {
		v, _, err := creditosParseOptionalBool(raw)
		if err != nil {
			return policy, fmt.Errorf("registrar_evento_contable invalido")
		}
		policy.RegistrarEventoContable = v
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("procesar_asientos")); raw != "" {
		v, _, err := creditosParseOptionalBool(raw)
		if err != nil {
			return policy, fmt.Errorf("procesar_asientos invalido")
		}
		policy.ProcesarAsientos = v
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("asientos_limit")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return policy, fmt.Errorf("asientos_limit invalido")
		}
		policy.AsientosLimit = v
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("max_reintentos")); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return policy, fmt.Errorf("max_reintentos invalido")
		}
		policy.MaxReintentos = v
	}

	policy.AsientosLimit = creditosNormalizeAsientosLimit(policy.AsientosLimit)
	policy.MaxReintentos = creditosNormalizeMaxReintentos(policy.MaxReintentos)
	return policy, nil
}

func creditosParseOptionalBool(raw string) (bool, bool, error) {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if raw == "" {
		return false, false, nil
	}
	switch raw {
	case "1", "true", "si", "yes", "on":
		return true, true, nil
	case "0", "false", "no", "off":
		return false, true, nil
	default:
		return false, true, fmt.Errorf("invalid bool")
	}
}

func creditosNormalizeAsientosLimit(v int) int {
	if v <= 0 {
		return 20
	}
	if v > 500 {
		return 500
	}
	return v
}

func creditosNormalizeMaxReintentos(v int) int {
	if v < 0 {
		return 0
	}
	if v > 50 {
		return 50
	}
	return v
}

func creditosFindMovimientoByID(dbEmp *sql.DB, empresaID, creditoID, movimientoID int64) (dbpkg.EmpresaCreditoMovimiento, bool) {
	rows, err := dbpkg.ListEmpresaCreditoMovimientos(dbEmp, empresaID, creditoID, true, 100)
	if err != nil {
		return dbpkg.EmpresaCreditoMovimiento{}, false
	}
	for i := range rows {
		if rows[i].ID == movimientoID {
			return rows[i], true
		}
	}
	return dbpkg.EmpresaCreditoMovimiento{}, false
}

func creditosResolveCanalPago(metodoPago string) string {
	m := strings.ToLower(strings.TrimSpace(metodoPago))
	if m == "" {
		return "caja"
	}
	if strings.Contains(m, "pasarela") || strings.Contains(m, "wompi") || strings.Contains(m, "stripe") || strings.Contains(m, "paypal") || strings.Contains(m, "mercadopago") {
		return "pasarela"
	}
	if strings.Contains(m, "transferencia") || strings.Contains(m, "banco") || strings.Contains(m, "pse") || strings.Contains(m, "nequi") || strings.Contains(m, "daviplata") {
		return "bancos"
	}
	return "caja"
}

func creditosResolvePeriodoContableAbono(fechaMovimiento string) string {
	fechaMovimiento = strings.TrimSpace(fechaMovimiento)
	if len(fechaMovimiento) >= 7 && fechaMovimiento[4] == '-' {
		return fechaMovimiento[:7]
	}
	if fechaMovimiento == "" {
		return ""
	}
	layouts := []string{"2006-01-02 15:04:05", time.RFC3339, "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, fechaMovimiento); err == nil {
			return parsed.In(time.Local).Format("2006-01")
		}
	}
	return ""
}
