package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type datafonoProviderDescriptor struct {
	Proveedor         string `json:"proveedor"`
	Nombre            string `json:"nombre"`
	Tipo              string `json:"tipo"`
	RequiereContrato  bool   `json:"requiere_contrato"`
	ApiPublica        bool   `json:"api_publica"`
	ApiBaseSugerida   string `json:"api_base_sugerida,omitempty"`
	CrearPagoPath     string `json:"crear_pago_path_sugerido,omitempty"`
	ConsultarPagoPath string `json:"consultar_pago_path_sugerido,omitempty"`
	Nota              string `json:"nota"`
}

type datafonoPaymentPayload struct {
	ConfigID       int64                        `json:"config_id"`
	Proveedor      string                       `json:"proveedor"`
	CarritoID      int64                        `json:"carrito_id"`
	Monto          float64                      `json:"monto"`
	Moneda         string                       `json:"moneda"`
	Referencia     string                       `json:"referencia"`
	Cliente        dbpkg.EmpresaDatafonoCliente `json:"cliente"`
	AplicarAlPOS   bool                         `json:"aplicar_al_pos"`
	CierreCajaID   int64                        `json:"cierre_caja_id"`
	CajaCodigo     string                       `json:"caja_codigo"`
	CajaTurno      string                       `json:"caja_turno"`
	CajaSucursalID int64                        `json:"caja_sucursal_id"`
}

type datafonoConsultarPayload struct {
	ID           int64  `json:"id"`
	Referencia   string `json:"referencia"`
	Proveedor    string `json:"proveedor"`
	AplicarAlPOS bool   `json:"aplicar_al_pos"`
}

type datafonoProviderClient interface {
	InitiatePayment(ctx context.Context, cfg dbpkg.EmpresaDatafonoConfig, req dbpkg.EmpresaDatafonoPaymentRequest) (dbpkg.EmpresaDatafonoProviderResponse, error)
	QueryPayment(ctx context.Context, cfg dbpkg.EmpresaDatafonoConfig, tx dbpkg.EmpresaDatafonoTransaction) (dbpkg.EmpresaDatafonoProviderResponse, error)
}

type httpDatafonoProviderClient struct {
	httpClient *http.Client
}

var datafonoHTTPClient = &http.Client{Timeout: 15 * time.Second}

func EmpresaDatafonosHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	client := &httpDatafonoProviderClient{httpClient: datafonoHTTPClient}
	return empresaDatafonosHandlerWithClient(dbEmp, dbSuper, client)
}

func empresaDatafonosHandlerWithClient(dbEmp, dbSuper *sql.DB, client datafonoProviderClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := dbpkg.EmpresaDatafonosSchemaReady(dbEmp); err != nil {
			log.Printf("[datafonos] schema unavailable empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo preparar el modulo de datafonos", http.StatusInternalServerError)
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch r.Method {
		case http.MethodGet:
			handleEmpresaDatafonosGET(w, r, dbEmp, empresaID, action)
		case http.MethodPost, http.MethodPut:
			handleEmpresaDatafonosWrite(w, r, dbEmp, dbSuper, client, empresaID, action)
		default:
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func handleEmpresaDatafonosGET(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	switch action {
	case "", "proveedores":
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": datafonoProviderCatalog()})
	case "config":
		includeInactive := strings.TrimSpace(r.URL.Query().Get("include_inactive")) == "1"
		items, err := dbpkg.ListEmpresaDatafonoConfigs(dbEmp, empresaID, includeInactive)
		if err != nil {
			log.Printf("[datafonos] list config empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudieron listar los datafonos configurados", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
	case "transacciones":
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		items, err := dbpkg.ListEmpresaDatafonoTransactions(dbEmp, empresaID, limit)
		if err != nil {
			log.Printf("[datafonos] list tx empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudieron listar las transacciones de datafono", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "items": items})
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func handleEmpresaDatafonosWrite(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, client datafonoProviderClient, empresaID int64, action string) {
	switch action {
	case "config":
		var cfg dbpkg.EmpresaDatafonoConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		cfg.EmpresaID = empresaID
		cfg.UsuarioCreador = adminEmailFromRequest(r)
		id, err := dbpkg.UpsertEmpresaDatafonoConfig(dbEmp, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "id": id})
	case "iniciar_pago":
		var payload datafonoPaymentPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		result, status := iniciarEmpresaDatafonoPago(r, dbEmp, dbSuper, client, empresaID, payload)
		writeJSON(w, status, result)
	case "consultar_pago":
		var payload datafonoConsultarPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		result, status := consultarEmpresaDatafonoPago(r, dbEmp, dbSuper, client, empresaID, payload)
		writeJSON(w, status, result)
	default:
		http.Error(w, "accion no soportada", http.StatusBadRequest)
	}
}

func iniciarEmpresaDatafonoPago(r *http.Request, dbEmp, dbSuper *sql.DB, client datafonoProviderClient, empresaID int64, payload datafonoPaymentPayload) (map[string]interface{}, int) {
	cfg, err := dbpkg.GetEmpresaDatafonoConfig(dbEmp, empresaID, payload.ConfigID, payload.Proveedor)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return map[string]interface{}{"ok": false, "error": "datafono no configurado para esta empresa/proveedor"}, http.StatusPreconditionFailed
		}
		log.Printf("[datafonos] get config empresa_id=%d error: %v", empresaID, err)
		return map[string]interface{}{"ok": false, "error": "No se pudo cargar la configuracion del datafono"}, http.StatusInternalServerError
	}
	req := dbpkg.EmpresaDatafonoPaymentRequest{
		EmpresaID:  empresaID,
		ConfigID:   cfg.ID,
		Proveedor:  cfg.Proveedor,
		CarritoID:  payload.CarritoID,
		Monto:      roundMoneyDatafono(payload.Monto),
		Moneda:     dbpkg.NormalizeDatafonoMoneda(firstNonEmptyDatafono(payload.Moneda, cfg.Moneda)),
		Referencia: strings.TrimSpace(payload.Referencia),
		Cliente:    payload.Cliente,
		Metadata: map[string]interface{}{
			"empresa_id": empresaID,
			"carrito_id": payload.CarritoID,
			"origen":     "pos_multiempresa",
		},
	}
	if req.Monto <= 0 {
		return map[string]interface{}{"ok": false, "error": "monto invalido"}, http.StatusBadRequest
	}
	if req.Referencia == "" {
		req.Referencia = fmt.Sprintf("POS-%d-%d", empresaID, time.Now().UnixNano())
	}
	txID, err := dbpkg.CreateEmpresaDatafonoTransaction(dbEmp, dbpkg.EmpresaDatafonoTransaction{
		EmpresaID:        empresaID,
		ConfigID:         cfg.ID,
		Proveedor:        cfg.Proveedor,
		CarritoID:        req.CarritoID,
		Referencia:       req.Referencia,
		Monto:            req.Monto,
		Moneda:           req.Moneda,
		ClienteNombre:    req.Cliente.Nombre,
		ClienteDocumento: req.Cliente.Documento,
		ClienteEmail:     req.Cliente.Email,
		ClienteTelefono:  req.Cliente.Telefono,
		EstadoPago:       dbpkg.DatafonoEstadoPendiente,
		RequestJSON:      dbpkg.DatafonoRequestJSON(req),
		UsuarioCreador:   adminEmailFromRequest(r),
		Estado:           "activo",
	})
	if err != nil {
		return map[string]interface{}{"ok": false, "error": err.Error()}, http.StatusBadRequest
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(cfg.TimeoutMs)*time.Millisecond)
	defer cancel()
	resp, err := client.InitiatePayment(ctx, cfg, req)
	if err != nil {
		_ = dbpkg.UpdateEmpresaDatafonoTransactionFromProvider(dbEmp, empresaID, txID, dbpkg.EmpresaDatafonoProviderResponse{EstadoPago: dbpkg.DatafonoEstadoError, MensajeRespuesta: err.Error()}, "", "", err.Error())
		return map[string]interface{}{"ok": false, "id": txID, "estado_pago": dbpkg.DatafonoEstadoError, "error": err.Error()}, http.StatusBadGateway
	}
	status := http.StatusOK
	errValidacion := dbpkg.ValidateDatafonoAmountAndReference(req, resp)
	errorMensaje := ""
	if errValidacion != nil {
		resp.EstadoPago = dbpkg.DatafonoEstadoError
		errorMensaje = errValidacion.Error()
		status = http.StatusConflict
	}
	if err := dbpkg.UpdateEmpresaDatafonoTransactionFromProvider(dbEmp, empresaID, txID, resp, dbpkg.DatafonoProviderResponseJSON(resp), "", errorMensaje); err != nil {
		log.Printf("[datafonos] update tx empresa_id=%d tx_id=%d error: %v", empresaID, txID, err)
		return map[string]interface{}{"ok": false, "id": txID, "error": "No se pudo guardar la respuesta del datafono"}, http.StatusInternalServerError
	}

	posAplicado, posWarning := false, ""
	if errorMensaje == "" && payload.AplicarAlPOS && dbpkg.NormalizeDatafonoEstadoPago(resp.EstadoPago) == dbpkg.DatafonoEstadoAprobado {
		posAplicado, posWarning = aplicarDatafonoAlPOS(r, dbEmp, dbSuper, empresaID, txID, cfg, req, resp, payload)
	}
	return map[string]interface{}{
		"ok":           errorMensaje == "",
		"id":           txID,
		"estado_pago":  dbpkg.NormalizeDatafonoEstadoPago(resp.EstadoPago),
		"respuesta":    resp,
		"pos_aplicado": posAplicado,
		"pos_warning":  posWarning,
		"error":        errorMensaje,
	}, status
}

func consultarEmpresaDatafonoPago(r *http.Request, dbEmp, dbSuper *sql.DB, client datafonoProviderClient, empresaID int64, payload datafonoConsultarPayload) (map[string]interface{}, int) {
	tx, err := dbpkg.GetEmpresaDatafonoTransaction(dbEmp, empresaID, payload.ID, payload.Proveedor, payload.Referencia)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return map[string]interface{}{"ok": false, "error": "transaccion de datafono no encontrada"}, http.StatusNotFound
		}
		log.Printf("[datafonos] get tx empresa_id=%d error: %v", empresaID, err)
		return map[string]interface{}{"ok": false, "error": "No se pudo cargar la transaccion del datafono"}, http.StatusInternalServerError
	}
	cfg, err := dbpkg.GetEmpresaDatafonoConfig(dbEmp, empresaID, tx.ConfigID, tx.Proveedor)
	if err != nil {
		return map[string]interface{}{"ok": false, "error": "datafono no configurado para consultar esta transaccion"}, http.StatusPreconditionFailed
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(cfg.TimeoutMs)*time.Millisecond)
	defer cancel()
	resp, err := client.QueryPayment(ctx, cfg, tx)
	if err != nil {
		_ = dbpkg.UpdateEmpresaDatafonoTransactionFromProvider(dbEmp, empresaID, tx.ID, dbpkg.EmpresaDatafonoProviderResponse{EstadoPago: dbpkg.DatafonoEstadoError, MensajeRespuesta: err.Error()}, "", "", err.Error())
		return map[string]interface{}{"ok": false, "id": tx.ID, "estado_pago": dbpkg.DatafonoEstadoError, "error": err.Error()}, http.StatusBadGateway
	}
	req := dbpkg.EmpresaDatafonoPaymentRequest{
		EmpresaID:  empresaID,
		ConfigID:   cfg.ID,
		Proveedor:  cfg.Proveedor,
		CarritoID:  tx.CarritoID,
		Monto:      tx.Monto,
		Moneda:     tx.Moneda,
		Referencia: tx.Referencia,
	}
	errValidacion := dbpkg.ValidateDatafonoAmountAndReference(req, resp)
	errorMensaje := ""
	status := http.StatusOK
	if errValidacion != nil {
		resp.EstadoPago = dbpkg.DatafonoEstadoError
		errorMensaje = errValidacion.Error()
		status = http.StatusConflict
	}
	if err := dbpkg.UpdateEmpresaDatafonoTransactionFromProvider(dbEmp, empresaID, tx.ID, resp, "", dbpkg.DatafonoProviderResponseJSON(resp), errorMensaje); err != nil {
		log.Printf("[datafonos] update query empresa_id=%d tx_id=%d error: %v", empresaID, tx.ID, err)
		return map[string]interface{}{"ok": false, "id": tx.ID, "error": "No se pudo guardar la consulta del datafono"}, http.StatusInternalServerError
	}

	posAplicado, posWarning := false, ""
	if errorMensaje == "" && payload.AplicarAlPOS && dbpkg.NormalizeDatafonoEstadoPago(resp.EstadoPago) == dbpkg.DatafonoEstadoAprobado {
		posAplicado, posWarning = aplicarDatafonoAlPOS(r, dbEmp, dbSuper, empresaID, tx.ID, cfg, req, resp, datafonoPaymentPayload{
			CarritoID: tx.CarritoID,
			Monto:     tx.Monto,
			Moneda:    tx.Moneda,
		})
	}
	return map[string]interface{}{
		"ok":           errorMensaje == "",
		"id":           tx.ID,
		"estado_pago":  dbpkg.NormalizeDatafonoEstadoPago(resp.EstadoPago),
		"respuesta":    resp,
		"pos_aplicado": posAplicado,
		"pos_warning":  posWarning,
		"error":        errorMensaje,
	}, status
}

func (c *httpDatafonoProviderClient) InitiatePayment(ctx context.Context, cfg dbpkg.EmpresaDatafonoConfig, req dbpkg.EmpresaDatafonoPaymentRequest) (dbpkg.EmpresaDatafonoProviderResponse, error) {
	path := strings.TrimSpace(cfg.CrearPagoPath)
	if path == "" {
		path = "/payments"
	}
	payload := map[string]interface{}{
		"amount":      req.Monto,
		"currency":    req.Moneda,
		"reference":   req.Referencia,
		"terminal_id": strings.TrimSpace(cfg.TerminalID),
		"merchant_id": strings.TrimSpace(cfg.ComercioID),
		"customer": map[string]interface{}{
			"name":     strings.TrimSpace(req.Cliente.Nombre),
			"document": strings.TrimSpace(req.Cliente.Documento),
			"email":    strings.TrimSpace(req.Cliente.Email),
			"phone":    strings.TrimSpace(req.Cliente.Telefono),
		},
		"metadata": req.Metadata,
	}
	return c.doProviderJSON(ctx, cfg, http.MethodPost, path, payload)
}

func (c *httpDatafonoProviderClient) QueryPayment(ctx context.Context, cfg dbpkg.EmpresaDatafonoConfig, tx dbpkg.EmpresaDatafonoTransaction) (dbpkg.EmpresaDatafonoProviderResponse, error) {
	path := strings.TrimSpace(cfg.ConsultarPagoPath)
	if path == "" {
		path = "/payments/{provider_transaction_id}"
	}
	replacer := strings.NewReplacer(
		"{id}", url.PathEscape(strings.TrimSpace(tx.ProviderTransactionID)),
		"{provider_transaction_id}", url.PathEscape(strings.TrimSpace(tx.ProviderTransactionID)),
		"{reference}", url.PathEscape(strings.TrimSpace(tx.Referencia)),
	)
	path = replacer.Replace(path)
	if strings.Contains(path, "{") {
		return dbpkg.EmpresaDatafonoProviderResponse{}, fmt.Errorf("ruta de consulta de datafono incompleta")
	}
	return c.doProviderJSON(ctx, cfg, http.MethodGet, path, nil)
}

func (c *httpDatafonoProviderClient) doProviderJSON(ctx context.Context, cfg dbpkg.EmpresaDatafonoConfig, method, path string, payload interface{}) (dbpkg.EmpresaDatafonoProviderResponse, error) {
	base := strings.TrimSpace(cfg.ApiBaseURL)
	if base == "" {
		return dbpkg.EmpresaDatafonoProviderResponse{}, fmt.Errorf("api_base_url del datafono no configurado")
	}
	endpoint, err := joinProviderURL(base, path)
	if err != nil {
		return dbpkg.EmpresaDatafonoProviderResponse{}, err
	}
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return dbpkg.EmpresaDatafonoProviderResponse{}, err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return dbpkg.EmpresaDatafonoProviderResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if err := applyDatafonoAuth(req, cfg); err != nil {
		return dbpkg.EmpresaDatafonoProviderResponse{}, err
	}
	httpClient := c.httpClient
	if httpClient == nil {
		httpClient = datafonoHTTPClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return dbpkg.EmpresaDatafonoProviderResponse{}, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	var decoded map[string]interface{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &decoded)
	}
	if decoded == nil {
		decoded = map[string]interface{}{}
	}
	out := normalizeDatafonoProviderHTTPResponse(cfg.Proveedor, decoded)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if out.MensajeRespuesta == "" {
			out.MensajeRespuesta = fmt.Sprintf("HTTP %d del proveedor", resp.StatusCode)
		}
		out.EstadoPago = dbpkg.DatafonoEstadoError
		return out, errors.New(out.MensajeRespuesta)
	}
	return out, nil
}

func applyDatafonoAuth(req *http.Request, cfg dbpkg.EmpresaDatafonoConfig) error {
	mode := strings.ToLower(strings.TrimSpace(cfg.AuthMode))
	if mode == "" || mode == "none" {
		return nil
	}
	secretRef := strings.TrimSpace(cfg.ApiKeyRef)
	if secretRef == "" {
		return fmt.Errorf("api_key_ref no configurado")
	}
	secret, err := resolveSecretReference(secretRef)
	if err != nil {
		return err
	}
	header := strings.TrimSpace(cfg.AuthHeader)
	switch mode {
	case "bearer":
		if header == "" {
			header = "Authorization"
		}
		req.Header.Set(header, "Bearer "+secret)
	case "api_key":
		if header == "" {
			header = "x-api-key"
		}
		req.Header.Set(header, secret)
	case "basic":
		if header == "" {
			header = "Authorization"
		}
		req.Header.Set(header, "Basic "+secret)
	default:
		return fmt.Errorf("auth_mode de datafono no soportado")
	}
	return nil
}

func resolveSecretReference(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if !strings.HasPrefix(strings.ToLower(ref), "env:") {
		return "", fmt.Errorf("referencia de secreto invalida")
	}
	key := strings.TrimSpace(ref[4:])
	if key == "" {
		return "", fmt.Errorf("variable de entorno de datafono vacia")
	}
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("variable de entorno de datafono no configurada")
	}
	return value, nil
}

func normalizeDatafonoProviderHTTPResponse(provider string, raw map[string]interface{}) dbpkg.EmpresaDatafonoProviderResponse {
	statusValue := firstStringFromMap(raw, "estado_pago", "payment_status", "status", "estado", "state", "result", "responseCode", "response_code", "approved")
	if v, ok := raw["approved"].(bool); ok {
		if v {
			statusValue = "approved"
		} else if statusValue == "" {
			statusValue = "rejected"
		}
	}
	resp := dbpkg.EmpresaDatafonoProviderResponse{
		ProviderTransactionID: firstStringFromMap(raw, "provider_transaction_id", "transaction_id", "transactionId", "id", "payment_id", "paymentId"),
		EstadoPago:            dbpkg.NormalizeDatafonoEstadoPago(statusValue),
		CodigoAutorizacion:    firstStringFromMap(raw, "codigo_autorizacion", "authorization_code", "authorizationCode", "approvalCode", "authCode"),
		MensajeRespuesta:      firstStringFromMap(raw, "mensaje_respuesta", "message", "description", "status_detail", "statusDetail", "error"),
		Referencia:            firstStringFromMap(raw, "reference", "referencia", "dev_reference", "external_reference", "externalReference"),
		Monto:                 firstFloatFromMap(raw, "amount", "monto", "total", "value"),
		Moneda:                dbpkg.NormalizeDatafonoMoneda(firstStringFromMap(raw, "currency", "moneda")),
		Raw:                   scrubDatafonoRaw(raw),
	}
	if resp.MensajeRespuesta == "" {
		resp.MensajeRespuesta = fmt.Sprintf("respuesta %s normalizada como %s", dbpkg.NormalizeDatafonoProvider(provider), resp.EstadoPago)
	}
	return resp
}

func aplicarDatafonoAlPOS(r *http.Request, dbEmp, dbSuper *sql.DB, empresaID, txID int64, cfg dbpkg.EmpresaDatafonoConfig, req dbpkg.EmpresaDatafonoPaymentRequest, resp dbpkg.EmpresaDatafonoProviderResponse, payload datafonoPaymentPayload) (bool, string) {
	if req.CarritoID <= 0 {
		return false, "transaccion aprobada sin carrito asociado"
	}
	carrito, err := dbpkg.GetCarritoCompraByID(dbEmp, empresaID, req.CarritoID)
	if err != nil {
		return false, "no se pudo cargar el carrito para aplicar el pago"
	}
	if strings.EqualFold(strings.TrimSpace(carrito.EstadoCarrito), "cerrado") || strings.TrimSpace(carrito.PagadoEn) != "" {
		_ = dbpkg.MarkEmpresaDatafonoTransactionPOSApplied(dbEmp, empresaID, txID, req.CarritoID)
		return true, "el carrito ya estaba cerrado"
	}
	usuarioOperacion := strings.TrimSpace(adminEmailFromRequest(r))
	cierreCaja, err := dbpkg.GetEmpresaCierreCajaAbiertaUsuario(dbEmp, empresaID, payload.CierreCajaID, payload.CajaCodigo, payload.CajaTurno, payload.CajaSucursalID, usuarioOperacion)
	if err != nil {
		return false, "pago aprobado, pero no hay caja abierta para aplicar el cierre en POS"
	}
	metodoPago := dbpkg.NormalizeMetodoPagoCarrito(cfg.MetodoPagoPOS)
	if metodoPago == "" || metodoPago == "efectivo" {
		metodoPago = "tarjeta_debito"
	}
	referenciaPago := strings.TrimSpace(resp.CodigoAutorizacion)
	if referenciaPago == "" {
		referenciaPago = strings.TrimSpace(resp.ProviderTransactionID)
	}
	if referenciaPago == "" {
		referenciaPago = req.Referencia
	}
	if err := dbpkg.PayCarritoStationSession(dbEmp, empresaID, req.CarritoID, metodoPago, referenciaPago, "", "", 0, 0, req.Monto, 0, cierreCaja.ID, cierreCaja.CajaCodigo, cierreCaja.Turno, cierreCaja.SucursalID, usuarioOperacion); err != nil {
		return false, "pago aprobado, pero no se pudo cerrar el carrito en POS: " + err.Error()
	}
	_ = dbpkg.MarkEmpresaDatafonoTransactionPOSApplied(dbEmp, empresaID, txID, req.CarritoID)
	_, _ = dbpkg.EnforceLicenciaDocumentosMensualesPorEmpresa(dbEmp, dbSuper, empresaID)
	return true, ""
}

func datafonoProviderCatalog() []datafonoProviderDescriptor {
	return []datafonoProviderDescriptor{
		{Proveedor: dbpkg.DatafonoProviderRedeban, Nombre: "Redeban", Tipo: "REST / datáfono / QR Redeban", RequiereContrato: true, ApiPublica: true, Nota: "Configura el endpoint y credenciales entregados por Redeban para comercio/terminal."},
		{Proveedor: dbpkg.DatafonoProviderCredibanco, Nombre: "CredibanCo", Tipo: "TEF Cloud / certificacion", RequiereContrato: true, ApiPublica: false, Nota: "La integracion tecnica se habilita bajo certificacion TEF Cloud o proveedor homologado."},
		{Proveedor: dbpkg.DatafonoProviderBold, Nombre: "Bold", Tipo: "Pagos presenciales/en linea API", RequiereContrato: true, ApiPublica: true, ApiBaseSugerida: "https://api.online.payments.bold.co", Nota: "Bold publica API de pagos; ajusta paths segun producto contratado y ambiente."},
		{Proveedor: dbpkg.DatafonoProviderBBVA, Nombre: "BBVA", Tipo: "QR/recaudo/reporting/API contractual", RequiereContrato: true, ApiPublica: true, Nota: "BBVA Colombia publica QR y APIs de reporte/recaudo; el endpoint de cobro depende del convenio."},
	}
}

func joinProviderURL(base, path string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(base))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("api_base_url invalida")
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("ruta del proveedor vacia")
	}
	basePath := strings.TrimRight(u.Path, "/")
	u.Path = basePath + "/" + strings.TrimLeft(path, "/")
	return u.String(), nil
}

func firstStringFromMap(raw map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := raw[key]; ok {
			switch typed := v.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return strings.TrimSpace(typed)
				}
			case float64:
				return strconv.FormatFloat(typed, 'f', -1, 64)
			case bool:
				if typed {
					return "true"
				}
				return "false"
			}
		}
	}
	return ""
}

func firstFloatFromMap(raw map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		if v, ok := raw[key]; ok {
			switch typed := v.(type) {
			case float64:
				return roundMoneyDatafono(typed)
			case int:
				return roundMoneyDatafono(float64(typed))
			case json.Number:
				f, _ := typed.Float64()
				return roundMoneyDatafono(f)
			case string:
				f, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(typed), ",", "."), 64)
				if err == nil {
					return roundMoneyDatafono(f)
				}
			}
		}
	}
	return 0
}

func scrubDatafonoRaw(raw map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(raw))
	for key, value := range raw {
		lower := strings.ToLower(key)
		if strings.Contains(lower, "card") || strings.Contains(lower, "pan") || strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "key") {
			out[key] = "[redacted]"
			continue
		}
		out[key] = value
	}
	return out
}

func firstNonEmptyDatafono(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func roundMoneyDatafono(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return math.Round(v*100) / 100
}
