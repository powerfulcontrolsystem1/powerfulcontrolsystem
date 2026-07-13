package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaRappiConfigPayload struct {
	EmpresaID          int64  `json:"empresa_id"`
	Activo             bool   `json:"activo"`
	Ambiente           string `json:"ambiente"`
	CountryDomain      string `json:"country_domain"`
	NewDomain          string `json:"new_domain"`
	ClientID           string `json:"client_id"`
	ClientSecretRef    string `json:"client_secret_ref"`
	WebhookSecretRef   string `json:"webhook_secret_ref"`
	StoreIntegrationID string `json:"store_integration_id"`
	RappiStoreID       string `json:"rappi_store_id"`
	AutoTomarOrdenes   bool   `json:"auto_tomar_ordenes"`
	CookingTimeMinutes int    `json:"cooking_time_minutes"`
	CrearVentaInterna  bool   `json:"crear_venta_interna"`
	Observaciones      string `json:"observaciones"`
}

type rappiActionPayload struct {
	OrderID            string `json:"order_id"`
	CookingTimeMinutes int    `json:"cooking_time_minutes"`
	Reason             string `json:"reason"`
}

func EmpresaRappiHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := dbpkg.EnsureEmpresaRappiSchema(dbEmp); err != nil {
			log.Printf("[rappi] ensure schema empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo preparar la integracion Rappi", http.StatusInternalServerError)
			return
		}
		switch r.Method {
		case http.MethodGet:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			switch action {
			case "", "dashboard", "config":
				handleEmpresaRappiDashboard(w, r, dbEmp, empresaID)
			case "stores", "tiendas":
				handleEmpresaRappiProxy(w, r, dbEmp, empresaID, http.MethodGet, "/api/v2/restaurants-integrations-public-api/stores-pa", nil, true)
			case "orders", "ordenes":
				handleEmpresaRappiOrders(w, r, dbEmp, empresaID, false)
			case "orders_sent", "ordenes_sent", "sent":
				handleEmpresaRappiOrders(w, r, dbEmp, empresaID, true)
			case "events", "eventos":
				orderID := strings.TrimSpace(r.URL.Query().Get("order_id"))
				if orderID == "" {
					http.Error(w, "order_id es obligatorio", http.StatusBadRequest)
					return
				}
				handleEmpresaRappiProxy(w, r, dbEmp, empresaID, http.MethodGet, "/api/v2/restaurants-integrations-public-api/orders/"+url.PathEscape(orderID)+"/events", nil, true)
			default:
				http.Error(w, "accion no permitida", http.StatusBadRequest)
			}
		case http.MethodPut:
			var payload empresaRappiConfigPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON invalido", http.StatusBadRequest)
				return
			}
			payload.EmpresaID = empresaID
			current, _ := dbpkg.GetEmpresaRappiConfig(dbEmp, empresaID)
			if strings.TrimSpace(payload.ClientSecretRef) == "" || strings.EqualFold(strings.TrimSpace(payload.ClientSecretRef), "configurado") {
				payload.ClientSecretRef = current.ClientSecretRef
			}
			if strings.TrimSpace(payload.WebhookSecretRef) == "" || strings.EqualFold(strings.TrimSpace(payload.WebhookSecretRef), "configurado") {
				payload.WebhookSecretRef = current.WebhookSecretRef
			}
			cfg := dbpkg.EmpresaRappiConfig{
				EmpresaID:          empresaID,
				Activo:             payload.Activo,
				Ambiente:           payload.Ambiente,
				CountryDomain:      payload.CountryDomain,
				NewDomain:          payload.NewDomain,
				ClientID:           strings.TrimSpace(payload.ClientID),
				ClientSecretRef:    strings.TrimSpace(payload.ClientSecretRef),
				WebhookSecretRef:   strings.TrimSpace(payload.WebhookSecretRef),
				StoreIntegrationID: strings.TrimSpace(payload.StoreIntegrationID),
				RappiStoreID:       strings.TrimSpace(payload.RappiStoreID),
				AutoTomarOrdenes:   payload.AutoTomarOrdenes,
				CookingTimeMinutes: payload.CookingTimeMinutes,
				CrearVentaInterna:  payload.CrearVentaInterna,
				Observaciones:      strings.TrimSpace(payload.Observaciones),
				UsuarioCreador:     adminEmailFromRequest(r),
			}
			if err := dbpkg.SaveEmpresaRappiConfig(dbEmp, cfg); err != nil {
				log.Printf("[rappi] save config empresa_id=%d error: %v", empresaID, err)
				http.Error(w, "No se pudo guardar configuracion Rappi", http.StatusInternalServerError)
				return
			}
			registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "rappi", "configuracion_guardada", "empresa_rappi_configuracion", 0, http.StatusOK, map[string]interface{}{
				"activo":               payload.Activo,
				"ambiente":             dbpkg.NormalizeRappiAmbiente(payload.Ambiente),
				"store_integration_id": strings.TrimSpace(payload.StoreIntegrationID),
				"rappi_store_id":       strings.TrimSpace(payload.RappiStoreID),
				"auto_tomar_ordenes":   payload.AutoTomarOrdenes,
				"crear_venta_interna":  payload.CrearVentaInterna,
			}, "configuracion Rappi actualizada por empresa")
			cfg, _ = dbpkg.GetEmpresaRappiConfig(dbEmp, empresaID)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": sanitizeRappiConfig(cfg)})
		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			switch action {
			case "test", "probar":
				handleEmpresaRappiProxy(w, r, dbEmp, empresaID, http.MethodGet, "/api/v2/restaurants-integrations-public-api/stores-pa", nil, true)
			case "take", "tomar":
				handleEmpresaRappiOrderAction(w, r, dbEmp, empresaID, "take")
			case "reject", "rechazar":
				handleEmpresaRappiOrderAction(w, r, dbEmp, empresaID, "reject")
			case "ready", "ready_for_pickup", "listo":
				handleEmpresaRappiOrderAction(w, r, dbEmp, empresaID, "ready")
			default:
				http.Error(w, "accion no permitida", http.StatusBadRequest)
			}
		default:
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		}
	}
}

func PublicRappiWebhookHandler(dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseInt64QueryOptional(r, "empresa_id")
		if err != nil || empresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		cfg, err := dbpkg.GetEmpresaRappiConfig(dbEmp, empresaID)
		if err != nil {
			http.Error(w, "No se pudo leer configuracion Rappi", http.StatusInternalServerError)
			return
		}
		if !cfg.Activo {
			http.Error(w, "Rappi no esta activo para esta empresa", http.StatusForbidden)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4<<20)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "webhook invalido", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(cfg.WebhookSecretRef) == "" {
			http.Error(w, "webhook invalido", http.StatusUnauthorized)
			return
		}
		secret, err := resolveDIANSecretValue(cfg.WebhookSecretRef)
		if err != nil || strings.TrimSpace(secret) == "" || !verifyRappiSignature(r.Header.Get("Rappi-Signature"), body, secret) {
			http.Error(w, "webhook invalido", http.StatusUnauthorized)
			return
		}
		orderLog := rappiOrderLogFromPayload(empresaID, "webhook", body)
		if _, err := dbpkg.UpsertEmpresaRappiOrderLog(dbEmp, orderLog); err != nil {
			log.Printf("[rappi] webhook persist empresa_id=%d error: %v", empresaID, err)
			http.Error(w, "No se pudo registrar webhook", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func handleEmpresaRappiDashboard(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64) {
	cfg, err := dbpkg.GetEmpresaRappiConfig(dbEmp, empresaID)
	if err != nil {
		http.Error(w, "No se pudo leer configuracion Rappi", http.StatusInternalServerError)
		return
	}
	orders, err := dbpkg.ListEmpresaRappiOrderLogs(dbEmp, empresaID, 80)
	if err != nil {
		http.Error(w, "No se pudo leer ordenes Rappi", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                   true,
		"empresa_id":           empresaID,
		"config":               sanitizeRappiConfig(cfg),
		"ordenes":              orders,
		"webhook_url_sugerida": fmt.Sprintf("/api/public/rappi/webhook?empresa_id=%d", empresaID),
	})
}

func handleEmpresaRappiOrders(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, sent bool) {
	path := "/api/v2/restaurants-integrations-public-api/orders"
	if sent {
		path = "/api/v2/restaurants-integrations-public-api/orders/status/sent"
	}
	if storeID := strings.TrimSpace(r.URL.Query().Get("store_id")); storeID != "" {
		path += "?storeId=" + url.QueryEscape(storeID)
	}
	respBody, status, err := empresaRappiRequest(dbEmp, empresaID, http.MethodGet, path, nil)
	if err != nil {
		http.Error(w, "No se pudo comunicar con Rappi", http.StatusBadGateway)
		return
	}
	if status >= 400 {
		writeJSON(w, status, map[string]interface{}{"ok": false, "status": status, "rappi_response": jsonRawOrString(respBody)})
		return
	}
	persistRappiOrdersResponse(dbEmp, empresaID, map[bool]string{true: "api_sent", false: "api_ready"}[sent], respBody)
	orders, _ := dbpkg.ListEmpresaRappiOrderLogs(dbEmp, empresaID, 80)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":             true,
		"empresa_id":     empresaID,
		"status":         status,
		"rappi_response": jsonRawOrString(respBody),
		"ordenes":        orders,
	})
}

func handleEmpresaRappiOrderAction(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, action string) {
	var payload rappiActionPayload
	_ = json.NewDecoder(r.Body).Decode(&payload)
	orderID := strings.TrimSpace(payload.OrderID)
	if orderID == "" {
		orderID = strings.TrimSpace(r.URL.Query().Get("order_id"))
	}
	if orderID == "" {
		http.Error(w, "order_id es obligatorio", http.StatusBadRequest)
		return
	}
	cfg, _ := dbpkg.GetEmpresaRappiConfig(dbEmp, empresaID)
	cookingTime := payload.CookingTimeMinutes
	if cookingTime <= 0 {
		cookingTime = cfg.CookingTimeMinutes
	}
	if cookingTime <= 0 {
		cookingTime = 15
	}
	method := http.MethodPut
	path := ""
	var body []byte
	switch action {
	case "take":
		path = "/api/v2/restaurants-integrations-public-api/orders/" + url.PathEscape(orderID) + "/take/" + strconv.Itoa(cookingTime)
	case "reject":
		path = "/api/v2/restaurants-integrations-public-api/orders/" + url.PathEscape(orderID) + "/reject"
		reason := strings.TrimSpace(payload.Reason)
		if reason != "" {
			body, _ = json.Marshal(map[string]string{"reason": reason})
		}
	case "ready":
		method = http.MethodPost
		path = "/api/v2/restaurants-integrations-public-api/orders/" + url.PathEscape(orderID) + "/ready-for-pickup"
	default:
		http.Error(w, "accion no permitida", http.StatusBadRequest)
		return
	}
	respBody, status, err := empresaRappiRequest(dbEmp, empresaID, method, path, body)
	if err != nil {
		http.Error(w, "No se pudo comunicar con Rappi", http.StatusBadGateway)
		return
	}
	localState := map[string]string{"take": "tomada", "reject": "rechazada", "ready": "lista"}[action]
	_, _ = dbpkg.UpsertEmpresaRappiOrderLog(dbEmp, dbpkg.EmpresaRappiOrderLog{
		EmpresaID:      empresaID,
		RappiOrderID:   orderID,
		EstadoRappi:    strings.ToUpper(action),
		EstadoLocal:    localState,
		RawPayloadJSON: string(respBody),
		Origen:         "accion_" + action,
		UsuarioCreador: adminEmailFromRequest(r),
	})
	registrarAuditoriaModuloEmpresaNoBloqueante(dbEmp, r, empresaID, "rappi", "orden_"+localState, "empresa_rappi_ordenes", 0, status, map[string]interface{}{
		"rappi_order_id": orderID,
		"accion":         action,
	}, "accion operativa enviada a Rappi")
	writeJSON(w, status, map[string]interface{}{"ok": status < 400, "status": status, "rappi_response": jsonRawOrString(respBody)})
}

func handleEmpresaRappiProxy(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, empresaID int64, method, path string, body []byte, includeConfig bool) {
	respBody, status, err := empresaRappiRequest(dbEmp, empresaID, method, path, body)
	if err != nil {
		http.Error(w, "No se pudo comunicar con Rappi", http.StatusBadGateway)
		return
	}
	resp := map[string]interface{}{"ok": status < 400, "status": status, "rappi_response": jsonRawOrString(respBody)}
	if includeConfig {
		if cfg, err := dbpkg.GetEmpresaRappiConfig(dbEmp, empresaID); err == nil {
			resp["config"] = sanitizeRappiConfig(cfg)
		}
	}
	writeJSON(w, status, resp)
}

func empresaRappiRequest(dbEmp *sql.DB, empresaID int64, method, path string, body []byte) ([]byte, int, error) {
	cfg, err := dbpkg.GetEmpresaRappiConfig(dbEmp, empresaID)
	if err != nil {
		return nil, 0, fmt.Errorf("no se pudo leer configuracion Rappi")
	}
	if !cfg.Activo {
		return nil, 0, fmt.Errorf("Rappi no esta activo para esta empresa")
	}
	token, err := empresaRappiAccessToken(cfg)
	if err != nil {
		return nil, 0, err
	}
	base := strings.TrimRight(cfg.CountryDomain, "/")
	if err := validateRappiBaseURL(base); err != nil {
		return nil, 0, err
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	req, err := http.NewRequest(method, base+path, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("x-authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("Rappi API no respondio: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	return respBody, resp.StatusCode, nil
}

func empresaRappiAccessToken(cfg dbpkg.EmpresaRappiConfig) (string, error) {
	clientID := strings.TrimSpace(cfg.ClientID)
	clientSecret, err := resolveDIANSecretValue(cfg.ClientSecretRef)
	if err != nil {
		return "", fmt.Errorf("no se pudo resolver client_secret de Rappi")
	}
	clientSecret = strings.TrimSpace(clientSecret)
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("credenciales Rappi incompletas")
	}
	loginBase := strings.TrimRight(cfg.NewDomain, "/")
	if err := validateRappiBaseURL(loginBase); err != nil {
		return "", err
	}
	loginURL := loginBase + "/restaurants/auth/v1/token/login/integrations"
	body, _ := json.Marshal(map[string]string{"client_id": clientID, "client_secret": clientSecret})
	req, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login Rappi no respondio: %w", err)
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("login Rappi fallo con HTTP %d", resp.StatusCode)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("respuesta de login Rappi invalida")
	}
	token := firstNonEmptyString(
		strings.TrimSpace(fmt.Sprint(parsed["access_token"])),
		strings.TrimSpace(fmt.Sprint(parsed["token"])),
	)
	if token == "" || token == "<nil>" {
		return "", fmt.Errorf("Rappi no devolvio access_token")
	}
	return token, nil
}

func validateRappiBaseURL(raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Hostname() == "" {
		return fmt.Errorf("dominio Rappi invalido")
	}
	if !strings.EqualFold(parsed.Scheme, "https") {
		return fmt.Errorf("dominio Rappi debe usar HTTPS")
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || !isAllowedRappiHost(host) {
		return fmt.Errorf("dominio Rappi no permitido")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("dominio Rappi no permitido")
		}
	}
	return nil
}

func isAllowedRappiHost(host string) bool {
	allowedHosts := []string{
		"rappi.com",
		"rappi.com.ar",
		"rappi.com.br",
		"rappi.cl",
		"rappi.com.co",
		"rappi.com.cr",
		"rappi.com.ec",
		"rappi.com.mx",
		"rappi.com.pe",
		"rappi.pe",
		"rappi.com.uy",
	}
	for _, allowed := range allowedHosts {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}

func sanitizeRappiConfig(cfg dbpkg.EmpresaRappiConfig) map[string]interface{} {
	return map[string]interface{}{
		"id":                     cfg.ID,
		"empresa_id":             cfg.EmpresaID,
		"activo":                 cfg.Activo,
		"ambiente":               cfg.Ambiente,
		"country_domain":         cfg.CountryDomain,
		"new_domain":             cfg.NewDomain,
		"client_id":              cfg.ClientID,
		"client_secret_ref":      secretRefStatus(cfg.ClientSecretRef),
		"webhook_secret_ref":     secretRefStatus(cfg.WebhookSecretRef),
		"store_integration_id":   cfg.StoreIntegrationID,
		"rappi_store_id":         cfg.RappiStoreID,
		"auto_tomar_ordenes":     cfg.AutoTomarOrdenes,
		"cooking_time_minutes":   cfg.CookingTimeMinutes,
		"crear_venta_interna":    cfg.CrearVentaInterna,
		"observaciones":          cfg.Observaciones,
		"fecha_actualizacion":    cfg.FechaActualizacion,
		"credenciales_completas": strings.TrimSpace(cfg.ClientID) != "" && strings.TrimSpace(cfg.ClientSecretRef) != "",
	}
}

func secretRefStatus(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return "configurado"
}

func verifyRappiSignature(header string, payload []byte, secret string) bool {
	header = strings.TrimSpace(header)
	if header == "" || strings.TrimSpace(secret) == "" {
		return false
	}
	var ts, signature string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch strings.TrimSpace(kv[0]) {
		case "t":
			ts = strings.TrimSpace(kv[1])
		case "sign":
			signature = strings.TrimSpace(kv[1])
		}
	}
	if ts == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(signature)), []byte(strings.ToLower(expected)))
}

func rappiOrderLogFromPayload(empresaID int64, origen string, payload []byte) dbpkg.EmpresaRappiOrderLog {
	var parsed map[string]interface{}
	_ = json.Unmarshal(payload, &parsed)
	orderID := firstNonEmptyString(
		valueAtPath(parsed, "order_id"),
		valueAtPath(parsed, "id"),
		valueAtPath(parsed, "order.id"),
		valueAtPath(parsed, "data.order_id"),
		valueAtPath(parsed, "data.id"),
	)
	if orderID == "" {
		sum := sha256.Sum256(payload)
		orderID = "webhook-" + hex.EncodeToString(sum[:])[:24]
	}
	storeID := firstNonEmptyString(valueAtPath(parsed, "store_id"), valueAtPath(parsed, "store.id"), valueAtPath(parsed, "data.store_id"))
	storeIntegrationID := firstNonEmptyString(valueAtPath(parsed, "store_integration_id"), valueAtPath(parsed, "integration_id"), valueAtPath(parsed, "data.store_integration_id"))
	status := firstNonEmptyString(valueAtPath(parsed, "status"), valueAtPath(parsed, "state"), valueAtPath(parsed, "data.status"), valueAtPath(parsed, "order.status"))
	total := floatFromAny(firstNonEmptyInterface(parsed, "total", "total_order", "billing.total_order", "data.total"))
	itemsJSON := ""
	if items, ok := firstNonNilInterface(parsed, "items", "products", "order.items", "data.items"); ok {
		if blob, err := json.Marshal(items); err == nil {
			itemsJSON = string(blob)
		}
	}
	return dbpkg.EmpresaRappiOrderLog{
		EmpresaID:          empresaID,
		RappiOrderID:       orderID,
		RappiStoreID:       storeID,
		StoreIntegrationID: storeIntegrationID,
		EstadoRappi:        status,
		EstadoLocal:        "recibida",
		Total:              total,
		Moneda:             "COP",
		ItemsJSON:          itemsJSON,
		RawPayloadJSON:     string(payload),
		Origen:             origen,
		UsuarioCreador:     "rappi",
		Observaciones:      "Orden registrada desde Rappi; convertir a venta interna requiere mapeo operativo de productos y caja.",
	}
}

func persistRappiOrdersResponse(dbEmp *sql.DB, empresaID int64, origen string, payload []byte) {
	var parsed interface{}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return
	}
	candidates := extractRappiOrderCandidates(parsed)
	for _, item := range candidates {
		blob, _ := json.Marshal(item)
		_, _ = dbpkg.UpsertEmpresaRappiOrderLog(dbEmp, rappiOrderLogFromPayload(empresaID, origen, blob))
	}
}

func extractRappiOrderCandidates(value interface{}) []interface{} {
	switch v := value.(type) {
	case []interface{}:
		return v
	case map[string]interface{}:
		for _, key := range []string{"orders", "data", "items", "results"} {
			if nested, ok := v[key]; ok {
				if out := extractRappiOrderCandidates(nested); len(out) > 0 {
					return out
				}
			}
		}
		return []interface{}{v}
	default:
		return nil
	}
}

func jsonRawOrString(raw []byte) interface{} {
	var out interface{}
	if err := json.Unmarshal(raw, &out); err == nil {
		return out
	}
	return string(raw)
}

func valueAtPath(root map[string]interface{}, path string) string {
	if root == nil || path == "" {
		return ""
	}
	var current interface{} = root
	for _, part := range strings.Split(path, ".") {
		m, ok := current.(map[string]interface{})
		if !ok {
			return ""
		}
		current = m[part]
	}
	value := strings.TrimSpace(fmt.Sprint(current))
	if value == "" || value == "<nil>" {
		return ""
	}
	return value
}

func firstNonNilInterface(root map[string]interface{}, paths ...string) (interface{}, bool) {
	for _, path := range paths {
		var current interface{} = root
		ok := true
		for _, part := range strings.Split(path, ".") {
			m, mapOK := current.(map[string]interface{})
			if !mapOK {
				ok = false
				break
			}
			current, ok = m[part]
			if !ok {
				break
			}
		}
		if ok && current != nil {
			return current, true
		}
	}
	return nil, false
}

func firstNonEmptyInterface(root map[string]interface{}, paths ...string) interface{} {
	value, ok := firstNonNilInterface(root, paths...)
	if !ok {
		return nil
	}
	return value
}

func floatFromAny(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		n, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return n
	default:
		return 0
	}
}
