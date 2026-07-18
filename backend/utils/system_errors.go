package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	ErrorLevelInfo     = "INFO"
	ErrorLevelWarning  = "WARNING"
	ErrorLevelError    = "ERROR"
	ErrorLevelCritical = "CRITICAL"

	internalErrorLoggedHeader = "X-PCS-Error-Logged"
	internalErrorIDHeader     = "X-PCS-Error-ID"
)

type SystemErrorRecord struct {
	Level         string
	ErrorType     string
	Message       string
	PublicMessage string
	Detail        string
	StackTrace    string
	EmpresaID     int64
	UserEmail     string
	Endpoint      string
	Module        string
	Method        string
	HTTPStatus    int
	RequestID     string
	Source        string
	IP            string
	UserAgent     string
	Metadata      map[string]interface{}
}

type errorMonitor struct {
	mu         sync.RWMutex
	fileMu     sync.Mutex
	dbSuper    *sql.DB
	backendDir string
}

var globalErrorMonitor = &errorMonitor{backendDir: "."}

func ConfigureErrorMonitor(dbSuper *sql.DB, backendDir string) {
	backendDir = strings.TrimSpace(backendDir)
	if backendDir == "" {
		backendDir = "."
	}
	if absBackendDir, err := filepath.Abs(backendDir); err == nil {
		backendDir = absBackendDir
	}

	globalErrorMonitor.mu.Lock()
	globalErrorMonitor.dbSuper = dbSuper
	globalErrorMonitor.backendDir = backendDir
	globalErrorMonitor.mu.Unlock()

	if dbSuper != nil {
		if err := dbpkg.EnsureSuperErroresSistemaSchema(dbSuper); err != nil {
			log.Printf("warning: no se pudo asegurar schema de errores del sistema: %v", err)
		}
	}
}

func normalizeSystemErrorLevel(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case ErrorLevelInfo:
		return ErrorLevelInfo
	case "WARN", ErrorLevelWarning:
		return ErrorLevelWarning
	case "ERR", ErrorLevelError:
		return ErrorLevelError
	case "FATAL", "PANIC", ErrorLevelCritical:
		return ErrorLevelCritical
	default:
		return ErrorLevelError
	}
}

func truncateSystemErrorText(raw string, max int) string {
	value := strings.TrimSpace(raw)
	if max <= 0 || len(value) <= max {
		return value
	}
	return value[:max]
}

func userEmailFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if v := r.Context().Value("adminEmail"); v != nil {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	for _, headerName := range []string{"X-Admin-Email", "X-User-Email"} {
		if headerValue := strings.TrimSpace(r.Header.Get(headerName)); headerValue != "" {
			return headerValue
		}
	}
	return ""
}

func inferModuleFromPath(path string) string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return "portal"
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "empresa" {
		return "empresa/" + parts[2]
	}
	if len(parts) >= 3 && parts[0] == "api" && parts[1] == "public" {
		return "public/" + parts[2]
	}
	if len(parts) >= 3 && parts[0] == "super" && parts[1] == "api" {
		return "super/" + parts[2]
	}
	if len(parts) >= 2 && parts[0] == "super" {
		return "super/" + parts[1]
	}
	return parts[0]
}

func isLikelyAPIPath(path string) bool {
	path = strings.TrimSpace(path)
	if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/super/api/") || strings.HasPrefix(path, "/epayco/") || strings.HasPrefix(path, "/wompi/") || strings.HasPrefix(path, "/licencias/") || strings.HasPrefix(path, "/accept/") {
		return true
	}
	switch path {
	case "/me", "/super/administradores", "/super/sesiones":
		return true
	default:
		return false
	}
}

func isLikelyAPIRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	if isLikelyAPIPath(r.URL.Path) {
		return true
	}
	accept := strings.ToLower(strings.TrimSpace(r.Header.Get("Accept")))
	if strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html") {
		return true
	}
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if contentType != "" && strings.Contains(contentType, "application/json") {
		return true
	}
	return false
}

func statusToErrorLevel(status int) string {
	if status >= http.StatusInternalServerError {
		return ErrorLevelError
	}
	if status >= http.StatusBadRequest {
		return ErrorLevelWarning
	}
	return ErrorLevelInfo
}

func friendlyAPIErrorMessage(status int) string {
	switch status {
	case http.StatusServiceUnavailable:
		return "El servicio no esta disponible en este momento. Intenta de nuevo en unos segundos."
	case http.StatusGatewayTimeout, http.StatusRequestTimeout:
		return "La operacion tardo demasiado. Intenta nuevamente."
	default:
		return "Ocurrio un problema interno. Intenta de nuevo en unos segundos."
	}
}

// friendlyClientAPIErrorMessage keeps validation failures useful without
// allowing database, provider, filesystem or credential details to escape in
// a client-controlled 4xx response.
func friendlyClientAPIErrorMessage(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "No fue posible validar tu sesion. Inicia sesion nuevamente."
	case http.StatusForbidden:
		return "No tienes permiso para realizar esta accion."
	case http.StatusNotFound:
		return "El recurso solicitado no esta disponible."
	case http.StatusConflict:
		return "La operacion no pudo completarse por un cambio reciente. Actualiza e intenta de nuevo."
	case http.StatusUnprocessableEntity:
		return "Hay datos pendientes o no validos para completar la operacion."
	case http.StatusTooManyRequests:
		return "Se alcanzo el limite temporal de solicitudes. Intenta de nuevo en unos segundos."
	default:
		return "La solicitud contiene datos no validos. Revisa la informacion e intenta de nuevo."
	}
}

func clientErrorMayExposeInternalDetail(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return false
	}
	for _, marker := range []string{
		"sql:", "pq:", "postgres", "database", "dsn", "password", "token", "secret",
		"certificate", "x509", "dial tcp", "connection refused", "no such file",
		"permission denied", "stack trace", "traceback", "panic", "smtp", "mailu",
		"nextcloud", "onlyoffice", "openai", "epayco", "wompi",
	} {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return strings.Contains(value, "/app/") || strings.Contains(value, "c:\\") || strings.Contains(value, "file:")
}

func friendlyHTMLServerError(requestID string, errorID int64) string {
	requestID = strings.TrimSpace(requestID)
	ref := requestID
	if errorID > 0 {
		ref = fmt.Sprintf("%s | error %d", requestID, errorID)
	}
	if ref == "" {
		ref = "sin referencia"
	}
	return "<!doctype html><html lang=\"es\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>Error interno</title><style>body{font-family:Segoe UI,Arial,sans-serif;background:#0f172a;color:#e2e8f0;display:grid;place-items:center;min-height:100vh;margin:0;padding:24px}.card{max-width:560px;background:#132038;border:1px solid rgba(148,163,184,.25);border-radius:18px;padding:28px;box-shadow:0 18px 48px rgba(0,0,0,.28)}h1{margin-top:0;font-size:1.8rem}p{line-height:1.6;color:#cbd5e1}.ref{margin-top:16px;font-size:.95rem;color:#93c5fd}</style></head><body><section class=\"card\"><h1>No fue posible completar la solicitud</h1><p>El sistema detecto un problema interno y lo registro para seguimiento. Puedes intentar nuevamente en unos segundos.</p><p class=\"ref\">Referencia: " + ref + "</p></section></body></html>"
}

func parseJSONErrorMessage(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !json.Valid([]byte(raw)) {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return ""
	}
	for _, key := range []string{"error", "message", "detalle", "detail"} {
		if value, ok := payload[key]; ok {
			text := strings.TrimSpace(fmt.Sprint(value))
			if text != "" {
				return text
			}
		}
	}
	return ""
}

func extractHTTPErrorLogValues(body, contentType string, status int) (string, string, string) {
	body = strings.TrimSpace(body)
	message := body
	if strings.Contains(strings.ToLower(contentType), "application/json") {
		if parsed := parseJSONErrorMessage(body); parsed != "" {
			message = parsed
		}
	}
	if message == "" {
		message = http.StatusText(status)
	}
	errorType := fmt.Sprintf("http_%d", status)
	if status >= http.StatusInternalServerError {
		errorType = "http_internal_error"
	}
	return truncateSystemErrorText(message, 4000), truncateSystemErrorText(body, 32000), errorType
}

func (monitor *errorMonitor) report(payload SystemErrorRecord) int64 {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("warning: error monitor panic recuperado: %v", recovered)
		}
	}()

	payload.Level = normalizeSystemErrorLevel(payload.Level)
	payload.ErrorType = truncateSystemErrorText(payload.ErrorType, 180)
	payload.Message = truncateSystemErrorText(payload.Message, 4000)
	payload.PublicMessage = truncateSystemErrorText(payload.PublicMessage, 1000)
	payload.Detail = truncateSystemErrorText(payload.Detail, 32000)
	payload.StackTrace = truncateSystemErrorText(payload.StackTrace, 64000)
	payload.UserEmail = truncateSystemErrorText(payload.UserEmail, 320)
	payload.Endpoint = truncateSystemErrorText(payload.Endpoint, 500)
	payload.Module = truncateSystemErrorText(payload.Module, 180)
	payload.Method = truncateSystemErrorText(payload.Method, 16)
	payload.RequestID = truncateSystemErrorText(payload.RequestID, 180)
	payload.Source = truncateSystemErrorText(payload.Source, 80)
	payload.IP = truncateSystemErrorText(payload.IP, 120)
	payload.UserAgent = truncateSystemErrorText(payload.UserAgent, 1000)
	if payload.Message == "" {
		payload.Message = "Error del sistema sin detalle adicional"
	}
	if payload.Module == "" {
		payload.Module = inferModuleFromPath(payload.Endpoint)
	}
	if payload.Source == "" {
		payload.Source = "backend"
	}

	metadataJSON := ""
	if len(payload.Metadata) > 0 {
		if raw, err := json.Marshal(payload.Metadata); err == nil {
			metadataJSON = truncateSystemErrorText(string(raw), 32000)
		}
	}

	monitor.mu.RLock()
	dbSuper := monitor.dbSuper
	backendDir := monitor.backendDir
	monitor.mu.RUnlock()

	record := dbpkg.SuperErrorSistema{
		Nivel:          payload.Level,
		TipoError:      payload.ErrorType,
		Mensaje:        payload.Message,
		MensajePublico: payload.PublicMessage,
		Detalle:        payload.Detail,
		StackTrace:     payload.StackTrace,
		EmpresaID:      payload.EmpresaID,
		UsuarioEmail:   payload.UserEmail,
		Endpoint:       payload.Endpoint,
		Modulo:         payload.Module,
		MetodoHTTP:     payload.Method,
		CodigoHTTP:     payload.HTTPStatus,
		RequestID:      payload.RequestID,
		Origen:         payload.Source,
		IP:             payload.IP,
		UserAgent:      payload.UserAgent,
		MetadataJSON:   metadataJSON,
		FechaError:     time.Now().Format("2006-01-02 15:04:05"),
		UsuarioCreador: "sistema",
		Estado:         "activo",
	}

	storedID := int64(0)
	if dbSuper != nil {
		id, err := dbpkg.CreateSuperErrorSistema(dbSuper, record)
		if err != nil {
			log.Printf("warning: no se pudo persistir error del sistema en DB: %v", err)
		} else {
			storedID = id
		}
	}

	linePayload := map[string]interface{}{
		"id":             storedID,
		"timestamp":      record.FechaError,
		"level":          record.Nivel,
		"error_type":     record.TipoError,
		"message":        record.Mensaje,
		"public_message": record.MensajePublico,
		"detail":         record.Detalle,
		"stack_trace":    record.StackTrace,
		"empresa_id":     record.EmpresaID,
		"user_email":     record.UsuarioEmail,
		"endpoint":       record.Endpoint,
		"module":         record.Modulo,
		"method":         record.MetodoHTTP,
		"http_status":    record.CodigoHTTP,
		"request_id":     record.RequestID,
		"source":         record.Origen,
		"ip":             record.IP,
		"user_agent":     record.UserAgent,
		"metadata_json":  metadataJSON,
	}
	monitor.appendFileLog(backendDir, linePayload)
	return storedID
}

func (monitor *errorMonitor) appendFileLog(backendDir string, payload map[string]interface{}) {
	logDir := filepath.Join(strings.TrimSpace(backendDir), "logs")
	if logDir == "" {
		logDir = filepath.Join(".", "logs")
	}
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		log.Printf("warning: no se pudo crear directorio de logs del sistema: %v", err)
		return
	}
	line, err := json.Marshal(payload)
	if err != nil {
		log.Printf("warning: no se pudo serializar log de error del sistema: %v", err)
		return
	}
	monitor.fileMu.Lock()
	defer monitor.fileMu.Unlock()
	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	f, err := os.OpenFile(filepath.Join(logDir, "system_errors.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		log.Printf("warning: no se pudo abrir system_errors.log: %v", err)
		return
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		log.Printf("warning: no se pudo escribir system_errors.log: %v", err)
	}
}

func ReportError(payload SystemErrorRecord) int64 {
	return globalErrorMonitor.report(payload)
}

func ReportProcessError(module, errorType, message string, err error, level string, metadata map[string]interface{}) int64 {
	detail := ""
	if err != nil {
		detail = err.Error()
		if strings.TrimSpace(message) == "" {
			message = err.Error()
		}
	}
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["reported_at"] = time.Now().Format(time.RFC3339)
	return ReportError(SystemErrorRecord{
		Level:     level,
		ErrorType: errorType,
		Message:   message,
		Detail:    detail,
		Module:    module,
		Source:    "process",
		Metadata:  metadata,
	})
}

func ReportExternalAPIError(module, endpoint, message string, err error, metadata map[string]interface{}) int64 {
	detail := ""
	if err != nil {
		detail = err.Error()
		if strings.TrimSpace(message) == "" {
			message = err.Error()
		}
	}
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["reported_at"] = time.Now().Format(time.RFC3339)
	return ReportError(SystemErrorRecord{
		Level:     ErrorLevelError,
		ErrorType: "external_api_error",
		Message:   message,
		Detail:    detail,
		Endpoint:  endpoint,
		Module:    module,
		Source:    "external_api",
		Metadata:  metadata,
	})
}

func RunProtectedProcess(module string, metadata map[string]interface{}, fn func()) {
	defer func() {
		if recovered := recover(); recovered != nil {
			panicMessage := fmt.Sprintf("%v", recovered)
			if metadata == nil {
				metadata = map[string]interface{}{}
			}
			metadata["panic"] = panicMessage
			metadata["stack_trace"] = string(debug.Stack())
			ReportError(SystemErrorRecord{
				Level:      ErrorLevelCritical,
				ErrorType:  "process_panic",
				Message:    "Se recupero un panic en proceso interno",
				Detail:     panicMessage,
				StackTrace: string(debug.Stack()),
				Module:     module,
				Source:     "process",
				Metadata:   metadata,
			})
		}
	}()
	fn()
}

func reportRequestError(r *http.Request, status int, level, errorType, message, publicMessage, detail, stack string, metadata map[string]interface{}) int64 {
	requestID := ""
	if r != nil {
		requestID = requestIDFromContext(r.Context())
	}
	if requestID == "" {
		requestID = makeRequestID()
	}
	if metadata == nil {
		metadata = map[string]interface{}{}
	}
	metadata["captured_at"] = time.Now().Format(time.RFC3339)

	record := SystemErrorRecord{
		Level:         level,
		ErrorType:     errorType,
		Message:       message,
		PublicMessage: publicMessage,
		Detail:        detail,
		StackTrace:    stack,
		HTTPStatus:    status,
		RequestID:     requestID,
		Source:        "http",
		Metadata:      metadata,
	}
	if r != nil {
		record.EmpresaID = inferEmpresaIDFromRequest(r)
		record.UserEmail = userEmailFromRequest(r)
		record.Endpoint = r.URL.Path
		record.Module = inferModuleFromPath(r.URL.Path)
		record.Method = r.Method
		record.IP = requestClientIP(r)
		record.UserAgent = r.UserAgent()
	}
	return ReportError(record)
}

func writeInternalErrorHeaders(h http.Header, requestID string, errorID int64) {
	h.Set(internalErrorLoggedHeader, "1")
	if requestID != "" {
		h.Set("X-Request-ID", requestID)
	}
	if errorID > 0 {
		h.Set(internalErrorIDHeader, strconv.FormatInt(errorID, 10))
	}
}

func copyResponseHeaders(dst, src http.Header) {
	for k, vals := range src {
		if strings.EqualFold(k, internalErrorLoggedHeader) ||
			strings.EqualFold(k, internalErrorIDHeader) ||
			strings.EqualFold(k, publicAPIErrorHeader) {
			continue
		}
		for _, v := range vals {
			dst.Add(k, v)
		}
	}
}

func writePublicAPIErrorResponse(w http.ResponseWriter, status int, requestID string, empresaID int64, errorID int64, path string, method string, raw []byte) {
	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	if empresaID > 0 {
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
	}
	w.WriteHeader(status)

	payload := map[string]interface{}{}
	if err := json.Unmarshal(raw, &payload); err != nil || payload == nil {
		message := strings.TrimSpace(string(raw))
		if message == "" {
			message = friendlyAPIErrorMessage(status)
		}
		payload = map[string]interface{}{
			"error": message,
		}
	}
	if _, ok := payload["ok"]; !ok {
		payload["ok"] = false
	}
	payload["status"] = status
	if _, ok := payload["request_id"]; !ok {
		payload["request_id"] = requestID
	}
	if _, ok := payload["path"]; !ok && path != "" {
		payload["path"] = path
	}
	if _, ok := payload["method"]; !ok && method != "" {
		payload["method"] = method
	}
	if empresaID > 0 {
		if _, ok := payload["empresa_id"]; !ok {
			payload["empresa_id"] = empresaID
		}
	}
	if errorID > 0 {
		if _, ok := payload["error_id"]; !ok {
			payload["error_id"] = errorID
		}
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func writeFriendlyAPIErrorResponse(w http.ResponseWriter, status int, requestID string, empresaID int64, errorID int64) {
	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	if empresaID > 0 {
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
	}
	w.WriteHeader(status)
	payload := map[string]interface{}{
		"ok":         false,
		"status":     status,
		"error":      friendlyAPIErrorMessage(status),
		"request_id": requestID,
	}
	if empresaID > 0 {
		payload["empresa_id"] = empresaID
	}
	if errorID > 0 {
		payload["error_id"] = errorID
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func writeFriendlyClientAPIErrorResponse(w http.ResponseWriter, status int, requestID string, empresaID int64, errorID int64) {
	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	if empresaID > 0 {
		w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
	}
	w.WriteHeader(status)
	payload := map[string]interface{}{
		"ok":         false,
		"status":     status,
		"error":      friendlyClientAPIErrorMessage(status),
		"request_id": requestID,
	}
	if empresaID > 0 {
		payload["empresa_id"] = empresaID
	}
	if errorID > 0 {
		payload["error_id"] = errorID
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID := requestIDFromContext(r.Context())
				if requestID == "" {
					requestID = makeRequestID()
				}
				stack := string(debug.Stack())
				panicMessage := fmt.Sprintf("%v", recovered)
				errorID := reportRequestError(r, http.StatusInternalServerError, ErrorLevelCritical, "panic_recovered", "Se recupero un panic en la solicitud", friendlyAPIErrorMessage(http.StatusInternalServerError), panicMessage, stack, map[string]interface{}{
					"panic": recovered,
				})
				writeInternalErrorHeaders(w.Header(), requestID, errorID)
				if isLikelyAPIRequest(r) {
					writeFriendlyAPIErrorResponse(w, http.StatusInternalServerError, requestID, inferEmpresaIDFromRequest(r), errorID)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(friendlyHTMLServerError(requestID, errorID)))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RequestContextWithModule(ctx context.Context, module string) context.Context {
	if strings.TrimSpace(module) == "" {
		return ctx
	}
	return context.WithValue(ctx, requestContextKey("module"), strings.TrimSpace(module))
}
