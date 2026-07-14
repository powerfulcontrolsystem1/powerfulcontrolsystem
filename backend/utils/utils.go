package utils

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	secure "github.com/you/pos-backend/secure"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}
type apiCaptureResponseWriter struct {
	headers http.Header
	body    bytes.Buffer
	status  int
}

type requestContextKey string

const (
	ctxKeyRequestID         requestContextKey = "request_id"
	ctxKeyEmpresaID         requestContextKey = "empresa_id"
	publicAPIErrorHeader                      = "X-PCS-Public-API-Error"
	canonicalPublicApexHost                   = "powerfulcontrolsystem.com"
	canonicalPublicWWWHost                    = "www.powerfulcontrolsystem.com"
	csrfCookieName                            = "pcs_csrf"
	csrfHeaderName                            = "X-CSRF-Token"
)

var companyLogMu sync.Mutex

type authSessionCacheEntry struct {
	Session  *dbpkg.Session
	CachedAt time.Time
}

type authAdminCacheEntry struct {
	Admin    *dbpkg.Admin
	CachedAt time.Time
}

var (
	authMiddlewareCacheMu             sync.Mutex
	authMiddlewareSessionCache        = map[string]authSessionCacheEntry{}
	authMiddlewareAdminCache          = map[string]authAdminCacheEntry{}
	authMiddlewareMaintenanceActive   bool
	authMiddlewareMaintenanceLoadedAt time.Time
)

// InvalidateAuthCacheForToken removes any request-local auth lookup associated
// with a browser token. It is safe to call even while cache TTLs are disabled.
func InvalidateAuthCacheForToken(token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	authMiddlewareCacheMu.Lock()
	delete(authMiddlewareSessionCache, token)
	authMiddlewareCacheMu.Unlock()
}

// InvalidateAuthCacheForAdmin removes all cached authorization state that can
// preserve a previous role, account state or session after a security event.
func InvalidateAuthCacheForAdmin(email string) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return
	}
	authMiddlewareCacheMu.Lock()
	delete(authMiddlewareAdminCache, email)
	for token, entry := range authMiddlewareSessionCache {
		if entry.Session != nil && strings.EqualFold(strings.TrimSpace(entry.Session.AdminEmail), email) {
			delete(authMiddlewareSessionCache, token)
		}
	}
	authMiddlewareCacheMu.Unlock()
}

const (
	// Authentication and authorization are read on every request. A stale cache
	// can otherwise keep a revoked session or downgraded role effective.
	authMiddlewareSessionCacheTTL     = 0
	authMiddlewareAdminCacheTTL       = 0
	authMiddlewareMaintenanceCacheTTL = 30 * time.Second
	SuperControlRole                  = "control_super_administrador"
)

func AdminShouldUseSuperRole(email string) bool {
	// Roles are assigned through the audited administrative workflow, never by
	// matching an email address. Kept only as a compatibility shim while legacy
	// callers migrate to role/ID based checks.
	_ = email
	return false
}

func IsSuperAdministradorRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), "super_administrador")
}

func IsSuperControlRole(role string) bool {
	return strings.EqualFold(strings.TrimSpace(role), SuperControlRole)
}

func IsSuperPanelRole(role string) bool {
	return IsSuperAdministradorRole(role) || IsSuperControlRole(role)
}

func ManagedAdminRole(email, currentRole string) string {
	_ = email
	normalizedCurrent := strings.ToLower(strings.TrimSpace(currentRole))
	if normalizedCurrent != "" && normalizedCurrent != "administrador" && normalizedCurrent != "super_administrador" {
		return strings.TrimSpace(currentRole)
	}
	if normalizedCurrent == "super_administrador" {
		return "super_administrador"
	}
	return "administrador"
}

func newAPICaptureResponseWriter() *apiCaptureResponseWriter {
	return &apiCaptureResponseWriter{headers: make(http.Header), status: http.StatusOK}
}

func (rw *apiCaptureResponseWriter) Header() http.Header {
	return rw.headers
}

func (rw *apiCaptureResponseWriter) WriteHeader(code int) {
	rw.status = code
}

func (rw *apiCaptureResponseWriter) Write(p []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.body.Write(p)
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := lrw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response writer does not support hijacking")
	}
	return hijacker.Hijack()
}

func requestIDFromContext(ctx context.Context) string {
	v := ctx.Value(ctxKeyRequestID)
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

// RequestIDFromContext expone el request_id propagado por middleware para otros paquetes.
func RequestIDFromContext(ctx context.Context) string {
	return requestIDFromContext(ctx)
}

// MarkPublicAPIError permite que un handler marque un error JSON >=500 como
// seguro para mostrar al usuario. Los errores no marcados siguen protegidos.
func MarkPublicAPIError(w http.ResponseWriter) {
	if w == nil {
		return
	}
	w.Header().Set(publicAPIErrorHeader, "1")
}

func empresaIDFromContext(ctx context.Context) int64 {
	v := ctx.Value(ctxKeyEmpresaID)
	if n, ok := v.(int64); ok && n > 0 {
		return n
	}
	return 0
}

func makeRequestID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req-%d-%x", time.Now().UnixNano(), b)
}

func parsePositiveInt64(raw string) int64 {
	v := strings.TrimSpace(raw)
	if v == "" {
		return 0
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return 0
	}
	return n
}

func extractEmpresaIDFromBody(raw []byte) int64 {
	if len(raw) == 0 {
		return 0
	}
	if len(raw) > 2*1024*1024 {
		return 0
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0
	}
	toInt64 := func(v interface{}) int64 {
		switch n := v.(type) {
		case float64:
			if n <= 0 {
				return 0
			}
			return int64(n)
		case string:
			return parsePositiveInt64(n)
		case int:
			if n > 0 {
				return int64(n)
			}
		case int64:
			if n > 0 {
				return n
			}
		}
		return 0
	}

	if v, ok := payload["empresa_id"]; ok {
		if id := toInt64(v); id > 0 {
			return id
		}
	}
	if v, ok := payload["empresaId"]; ok {
		if id := toInt64(v); id > 0 {
			return id
		}
	}
	if empresaObj, ok := payload["empresa"].(map[string]interface{}); ok {
		if v, exists := empresaObj["id"]; exists {
			if id := toInt64(v); id > 0 {
				return id
			}
		}
	}

	return 0
}

func inferEmpresaIDFromRequest(r *http.Request) int64 {
	if id := empresaIDFromContext(r.Context()); id > 0 {
		return id
	}
	// A client-provided empresa_id is not a trusted tenant identifier. Only the
	// authorization layer may set it in request context after validation.
	return 0
}

func requestClientIP(r *http.Request) string {
	if RequestFromTrustedProxy(r) && strings.TrimSpace(r.Header.Get("X-Forwarded-For")) != "" {
		v := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
		parts := strings.Split(v, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if RequestFromTrustedProxy(r) && strings.TrimSpace(r.Header.Get("X-Real-IP")) != "" {
		v := strings.TrimSpace(r.Header.Get("X-Real-IP"))
		return v
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

// RequestFromTrustedProxy only accepts forwarding headers from explicit CIDRs.
// With no configuration, only loopback proxies are trusted for local development.
func RequestFromTrustedProxy(r *http.Request) bool {
	if r == nil {
		return false
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	configured := strings.TrimSpace(os.Getenv("PCS_TRUSTED_PROXY_CIDRS"))
	if configured == "" {
		return ip.IsLoopback()
	}
	for _, rawCIDR := range strings.Split(configured, ",") {
		_, network, parseErr := net.ParseCIDR(strings.TrimSpace(rawCIDR))
		if parseErr == nil && network.Contains(ip) {
			return true
		}
	}
	return false
}

func firstForwardedHeaderValue(raw string) string {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func requestHostWithoutPort(rawHost string) string {
	trimmed := strings.TrimSpace(rawHost)
	if trimmed == "" {
		return ""
	}
	hostOnly, _, err := net.SplitHostPort(trimmed)
	if err == nil {
		return strings.TrimSpace(hostOnly)
	}
	if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		return strings.Trim(strings.TrimSpace(trimmed), "[]")
	}
	return trimmed
}

func resolveRequestHost(r *http.Request) string {
	if r == nil {
		return ""
	}
	if RequestFromTrustedProxy(r) {
		if host := firstForwardedHeaderValue(r.Header.Get("X-Forwarded-Host")); host != "" {
			return host
		}
	}
	return strings.TrimSpace(r.Host)
}

func resolveRequestScheme(r *http.Request) string {
	if r == nil {
		return ""
	}
	if RequestFromTrustedProxy(r) {
		if scheme := strings.ToLower(firstForwardedHeaderValue(r.Header.Get("X-Forwarded-Proto"))); scheme == "http" || scheme == "https" {
			return scheme
		}
	}
	if r.TLS != nil {
		return "https"
	}
	if scheme := strings.ToLower(strings.TrimSpace(r.URL.Scheme)); scheme == "http" || scheme == "https" {
		return scheme
	}
	return "http"
}

// IsSameOriginRequest validates browser Origin/Referer against the effective
// host. It is used for authenticated cookie mutations and WebSocket upgrades.
func IsSameOriginRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	requestHost := strings.TrimSpace(resolveRequestHost(r))
	requestScheme := resolveRequestScheme(r)
	if requestHost == "" || requestScheme == "" {
		return false
	}
	for _, header := range []string{"Origin", "Referer"} {
		raw := strings.TrimSpace(r.Header.Get(header))
		if raw == "" {
			continue
		}
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Host == "" || parsed.User != nil {
			return false
		}
		if strings.EqualFold(parsed.Scheme, requestScheme) && strings.EqualFold(parsed.Host, requestHost) {
			return true
		}
		allowedOrigins := strings.TrimSpace(os.Getenv("PCS_CSRF_ALLOWED_ORIGINS"))
		if allowedOrigins == "" {
			allowedOrigins = strings.TrimSpace(os.Getenv("CSRF_ALLOWED_ORIGINS"))
		}
		for _, allowed := range strings.Split(allowedOrigins, ",") {
			origin, parseErr := url.Parse(strings.TrimSpace(allowed))
			if parseErr == nil && origin.Scheme != "" && origin.Host != "" && strings.EqualFold(parsed.Scheme, origin.Scheme) && strings.EqualFold(parsed.Host, origin.Host) {
				return true
			}
		}
		return false
	}
	return false
}

// SessionCookieMaxAge centraliza la duracion de cookies de autenticacion.
// SESSION_TIMEOUT usa el formato de time.ParseDuration, por ejemplo 12h.
func SessionCookieMaxAge() int {
	duration := 24 * time.Hour
	if raw := strings.TrimSpace(os.Getenv("SESSION_TIMEOUT")); raw != "" {
		if parsed, err := time.ParseDuration(raw); err == nil {
			duration = parsed
		}
	}
	if duration < 5*time.Minute {
		duration = 5 * time.Minute
	}
	if duration > 7*24*time.Hour {
		duration = 7 * 24 * time.Hour
	}
	return int(duration.Seconds())
}

// RequestBodyLimitMiddleware aplica un techo global antes de los limites mas
// especificos de cada handler. Los upgrades WebSocket no transportan body HTTP.
func RequestBodyLimitMiddleware(next http.Handler, maxBytes int64) http.Handler {
	if maxBytes <= 0 {
		maxBytes = 64 << 20
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil && !isWebSocketUpgrade(r) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	})
}

func isCookieAuthenticatedMutation(r *http.Request) bool {
	if r == nil || isWebSocketUpgrade(r) {
		return false
	}
	// Login, registration and recovery endpoints are intentionally public. A
	// stale cookie from an earlier browser session must not convert them into an
	// authenticated mutation and block access before the handler can rotate it.
	switch r.URL.Path {
	case "/accept/complete",
		"/super/api/administradores/register",
		"/super/api/administradores/login",
		"/super/api/administradores/solicitar_recuperacion",
		"/super/api/administradores/restablecer_password",
		"/api/empresa/usuarios/login",
		"/api/empresa/usuarios/establecer_password",
		"/api/empresa/usuarios/recuperar_invitacion",
		"/api/empresa/usuarios/solicitar_recuperacion_password",
		"/api/empresa/usuarios/restablecer_password":
		return false
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(r.Header.Get("Authorization"))), "bearer ") {
			return false
		}
		_, err := r.Cookie("session_token")
		return err == nil
	default:
		return false
	}
}

func newCSRFToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func setCSRFCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		Secure:   resolveRequestScheme(r) == "https",
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int((12 * time.Hour).Seconds()),
	})
}

func requestCSRFTokenValid(r *http.Request) bool {
	cookie, err := r.Cookie(csrfCookieName)
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		return false
	}
	headerToken := strings.TrimSpace(r.Header.Get(csrfHeaderName))
	if headerToken == "" || len(headerToken) != len(cookie.Value) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookie.Value)) == 1
}

func csrfShouldRotate(r *http.Request) bool {
	if r == nil || r.Method != http.MethodPost {
		return false
	}
	switch r.URL.Path {
	case "/super/api/administradores/login",
		"/api/empresa/usuarios/login",
		"/api/empresa/usuarios/establecer_password",
		"/api/account/change_password",
		"/api/account/set_google_password",
		"/super/api/administradores/2fa":
		return true
	default:
		return false
	}
}

// CSRFMiddleware protects mutable requests authenticated with the HttpOnly
// session cookie using strict origin validation and a synchronizer token. The
// token is deliberately readable by same-origin JavaScript and is compared in
// constant time against its independent cookie. Bearer and signed-webhook
// clients do not share this cookie-CSRF threat model.
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		csrfCookie, cookieErr := r.Cookie(csrfCookieName)
		if cookieErr != nil || strings.TrimSpace(csrfCookie.Value) == "" || csrfShouldRotate(r) {
			token, err := newCSRFToken()
			if err != nil {
				http.Error(w, "csrf token generation failed", http.StatusInternalServerError)
				return
			}
			setCSRFCookie(w, r, token)
		}
		if isCookieAuthenticatedMutation(r) {
			if !IsSameOriginRequest(r) {
				http.Error(w, "csrf origin validation failed", http.StatusForbidden)
				return
			}
			if !requestCSRFTokenValid(r) {
				http.Error(w, "csrf token validation failed", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func sensitiveNoStorePath(path string) bool {
	path = strings.TrimSpace(path)
	return path == "/login.html" || strings.HasPrefix(path, "/auth/") || strings.Contains(path, "recuperacion") || strings.Contains(path, "password") || strings.Contains(path, "totp")
}

// SecurityHeadersMiddleware provides conservative browser protections while
// retaining the currently integrated payment, Google and OnlyOffice origins.
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(self), geolocation=(self), payment=(self)")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' https://accounts.google.com https://checkout.epayco.co https://checkout.wompi.co; connect-src 'self' https: wss:; frame-src 'self' https://accounts.google.com https://checkout.epayco.co https://checkout.wompi.co https://*.google.com")
		// Report-only starts the controlled CSP transition without breaking the
		// legacy static frontend. Once reports show no unexpected dependencies,
		// this policy can replace the compatibility policy above.
		w.Header().Set("Content-Security-Policy-Report-Only", "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'self'; form-action 'self'; img-src 'self' data: blob:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline' https://accounts.google.com https://checkout.epayco.co https://checkout.wompi.co; connect-src 'self' wss://powerfulcontrolsystem.com https://api.openai.com https://checkout.wompi.co https://secure.epayco.co; frame-src 'self' https://accounts.google.com https://checkout.epayco.co https://checkout.wompi.co")
		if r.TLS != nil || strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") && RequestFromTrustedProxy(r) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if sensitiveNoStorePath(r.URL.Path) {
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("Pragma", "no-cache")
		}
		next.ServeHTTP(w, r)
	})
}

func CanonicalPublicHostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(requestHostWithoutPort(resolveRequestHost(r)))
		if host == canonicalPublicWWWHost {
			target := &url.URL{
				Scheme:   "https",
				Host:     canonicalPublicApexHost,
				Path:     r.URL.Path,
				RawQuery: r.URL.RawQuery,
			}
			http.Redirect(w, r, target.String(), http.StatusPermanentRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeCompanyLogEntry(empresaID int64, level, msg string) {
	logDir := filepath.Join(".", "logs")
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		log.Printf("warning: no se pudo crear carpeta de logs %s: %v", logDir, err)
		return
	}

	fileName := "empresa_global.log"
	if empresaID > 0 {
		fileName = fmt.Sprintf("empresa_%d.log", empresaID)
	}
	filePath := filepath.Join(logDir, fileName)
	line := fmt.Sprintf("%s [%s] %s\n", time.Now().Format(time.RFC3339), strings.ToUpper(strings.TrimSpace(level)), sanitizeLogValue(msg))

	companyLogMu.Lock()
	defer companyLogMu.Unlock()

	// #nosec G304 -- path is normalized and constrained to a server-controlled root before this operation.
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		log.Printf("warning: no se pudo abrir log de empresa %s: %v", filePath, err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(line); err != nil {
		log.Printf("warning: no se pudo escribir log de empresa %s: %v", filePath, err)
	}
}

func sanitizeLogValue(value string) string {
	value = strings.NewReplacer("\r", " ", "\n", " ", "\x00", " ").Replace(value)
	return strings.TrimSpace(value)
}

func truncateLogMessage(s string, max int) string {
	v := strings.TrimSpace(s)
	if max <= 0 || len(v) <= max {
		return v
	}
	return v[:max]
}

// LoggingMiddleware registra trazabilidad profesional por request y por empresa.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := makeRequestID()
		empresaID := inferEmpresaIDFromRequest(r)

		ctx := context.WithValue(r.Context(), ctxKeyRequestID, requestID)
		if empresaID > 0 {
			ctx = context.WithValue(ctx, ctxKeyEmpresaID, empresaID)
		}
		r = r.WithContext(ctx)

		w.Header().Set("X-Request-ID", requestID)
		if empresaID > 0 {
			w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		}

		start := time.Now()
		clientIP := requestClientIP(r)
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		log.Printf("-> req_id=%s empresa_id=%d %s %s from %s", requestID, empresaID, r.Method, r.URL.Path, clientIP)
		writeCompanyLogEntry(empresaID, "INFO", fmt.Sprintf("req_id=%s event=request_start method=%s path=%s ip=%s", requestID, r.Method, r.URL.Path, clientIP))

		next.ServeHTTP(lrw, r)

		finalEmpresaID := empresaID
		if headerEmpresaID := parsePositiveInt64(lrw.Header().Get("X-Empresa-ID")); headerEmpresaID > 0 {
			finalEmpresaID = headerEmpresaID
		}

		elapsedMs := time.Since(start).Milliseconds()
		level := ErrorLevelInfo
		if lrw.status >= 500 {
			level = ErrorLevelError
		} else if lrw.status >= 400 {
			level = ErrorLevelWarning
		}

		log.Printf("<- req_id=%s empresa_id=%d status=%d %s %s dur_ms=%d", requestID, finalEmpresaID, lrw.status, r.Method, r.URL.Path, elapsedMs)
		writeCompanyLogEntry(finalEmpresaID, level, fmt.Sprintf("req_id=%s event=request_end method=%s path=%s status=%d dur_ms=%d", requestID, r.Method, r.URL.Path, lrw.status, elapsedMs))
	})
}

// JSONErrorMiddleware unifica errores no-JSON para endpoints de API.
func JSONErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLikelyAPIRequest(r) || isWebSocketUpgrade(r) {
			next.ServeHTTP(w, r)
			return
		}

		capture := newAPICaptureResponseWriter()
		next.ServeHTTP(capture, r)
		copyResponseHeaders(w.Header(), capture.Header())

		contentType := strings.ToLower(capture.Header().Get("Content-Type"))
		isJSON := strings.Contains(contentType, "application/json")
		path := r.URL.Path
		loggedAlready := strings.TrimSpace(capture.Header().Get(internalErrorLoggedHeader)) == "1"
		errorID := parsePositiveInt64(capture.Header().Get(internalErrorIDHeader))
		reqID := requestIDFromContext(r.Context())
		if reqID == "" {
			reqID = makeRequestID()
		}
		empresaID := inferEmpresaIDFromRequest(r)
		if reqID != "" {
			w.Header().Set("X-Request-ID", reqID)
		}
		if empresaID > 0 {
			w.Header().Set("X-Empresa-ID", strconv.FormatInt(empresaID, 10))
		}

		if capture.status >= 400 {
			msg, detail, errorType := extractHTTPErrorLogValues(capture.body.String(), contentType, capture.status)
			level := statusToErrorLevel(capture.status)
			if !loggedAlready {
				errorID = reportRequestError(r, capture.status, level, errorType, msg, friendlyAPIErrorMessage(capture.status), detail, "", map[string]interface{}{
					"content_type": contentType,
					"path":         path,
				})
			}
			writeCompanyLogEntry(empresaID, level, fmt.Sprintf("req_id=%s event=api_error method=%s path=%s status=%d error=%q", reqID, r.Method, path, capture.status, truncateLogMessage(msg, 500)))

			publicError := strings.TrimSpace(capture.Header().Get(publicAPIErrorHeader)) == "1"
			if capture.status >= http.StatusInternalServerError {
				if publicError {
					writePublicAPIErrorResponse(w, capture.status, reqID, empresaID, errorID, path, r.Method, capture.body.Bytes())
					return
				}
				writeFriendlyAPIErrorResponse(w, capture.status, reqID, empresaID, errorID)
				return
			}

			if !isJSON {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(capture.status)
				payload := map[string]interface{}{
					"ok":         false,
					"status":     capture.status,
					"error":      msg,
					"path":       path,
					"method":     r.Method,
					"request_id": reqID,
				}
				if empresaID > 0 {
					payload["empresa_id"] = empresaID
				}
				if errorID > 0 {
					payload["error_id"] = errorID
				}
				_ = json.NewEncoder(w).Encode(payload)
				return
			}
		}

		w.WriteHeader(capture.status)
		_, _ = w.Write(capture.body.Bytes())
	})
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(r.Header.Get("Upgrade")), "websocket")
}

// GenerateSecureToken devuelve un token seguro en hex de `n` bytes.
func GenerateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func getCachedMaintenanceActive(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return false
	}
	authMiddlewareCacheMu.Lock()
	if time.Since(authMiddlewareMaintenanceLoadedAt) < authMiddlewareMaintenanceCacheTTL {
		value := authMiddlewareMaintenanceActive
		authMiddlewareCacheMu.Unlock()
		return value
	}
	authMiddlewareCacheMu.Unlock()

	active := false
	if val, _, err := dbpkg.GetConfigValue(dbSuper, "mantenimiento_activo"); err == nil && val == "true" {
		active = true
	}

	authMiddlewareCacheMu.Lock()
	authMiddlewareMaintenanceActive = active
	authMiddlewareMaintenanceLoadedAt = time.Now()
	authMiddlewareCacheMu.Unlock()
	return active
}

func getCachedSessionByToken(dbSuper *sql.DB, token string) (*dbpkg.Session, error) {
	token = strings.TrimSpace(token)
	if dbSuper == nil || token == "" {
		return nil, sql.ErrNoRows
	}
	authMiddlewareCacheMu.Lock()
	if cached, ok := authMiddlewareSessionCache[token]; ok && time.Since(cached.CachedAt) < authMiddlewareSessionCacheTTL {
		authMiddlewareCacheMu.Unlock()
		return cached.Session, nil
	}
	authMiddlewareCacheMu.Unlock()

	session, err := dbpkg.GetSessionByToken(dbSuper, token)
	if err != nil {
		return nil, err
	}
	authMiddlewareCacheMu.Lock()
	authMiddlewareSessionCache[token] = authSessionCacheEntry{
		Session:  session,
		CachedAt: time.Now(),
	}
	authMiddlewareCacheMu.Unlock()
	return session, nil
}

func getCachedAdminByEmailFull(dbSuper *sql.DB, email string) (*dbpkg.Admin, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if dbSuper == nil || email == "" {
		return nil, sql.ErrNoRows
	}
	authMiddlewareCacheMu.Lock()
	if cached, ok := authMiddlewareAdminCache[email]; ok && time.Since(cached.CachedAt) < authMiddlewareAdminCacheTTL {
		authMiddlewareCacheMu.Unlock()
		return cached.Admin, nil
	}
	authMiddlewareCacheMu.Unlock()

	admin, err := dbpkg.GetAdminByEmailFull(dbSuper, email)
	if err != nil {
		return nil, err
	}
	authMiddlewareCacheMu.Lock()
	authMiddlewareAdminCache[email] = authAdminCacheEntry{
		Admin:    admin,
		CachedAt: time.Now(),
	}
	authMiddlewareCacheMu.Unlock()
	return admin, nil
}

func allowAdminLimitedSuperRoute(path, method, role string) bool {
	if !strings.EqualFold(strings.TrimSpace(role), "administrador") {
		return false
	}
	switch strings.TrimSpace(path) {
	case "/super/administradores.html",
		"/super/api/administradores",
		"/super/administradores":
		return method == http.MethodGet || method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
	case "/super/licencias.html":
		return method == http.MethodGet
	case "/super/api/empresas":
		return method == http.MethodGet || method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
	case "/super/api/empresas/compartidos":
		return method == http.MethodGet || method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete
	case "/super/api/empresas/compartidos/aceptar":
		return method == http.MethodPost
	case "/super/api/tipos_empresas", "/super/api/licencias":
		return method == http.MethodGet
	default:
		return false
	}
}

func allowSuperControlRoute(path, method, role string) bool {
	_ = method
	if !IsSuperControlRole(role) {
		return false
	}
	switch strings.TrimSpace(path) {
	case "/super/licencias_resumen.html",
		"/super/empresas.html",
		"/super/administradores.html",
		"/super/seguridad.html",
		"/super/api/administradores",
		"/super/administradores",
		"/super/sesiones",
		"/super/api/errores",
		"/super/api/metrics/current",
		"/super/api/metrics/history",
		"/super/api/empresas_estado",
		"/super/api/vps2",
		"/super/api/security/ports",
		"/super/api/security/processes",
		"/super/api/security/vps/config",
		"/super/api/security/vps/run",
		"/super/api/security/vps/status",
		"/super/api/security/vps/history",
		"/super/api/security/vps/report",
		"/super/api/security/vps/compare":
		return true
	default:
		return false
	}
}

// AuthMiddleware protege rutas usando la tabla sesiones y administradores en la BD superadministrador.
// Permite un conjunto público de rutas (login/callback/activos). Para rutas que comienzan con /super/
// exige rol 'super_administrador'. Añade `adminEmail` en el contexto de la petición.
func AuthMiddleware(dbSuper *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Verificación de Modo Mantenimiento
		if getCachedMaintenanceActive(dbSuper) {
			// Rutas permitidas durante el mantenimiento (acceso de administradores y assets estáticos)
			isMntAllowed := path == "/mantenimiento.html" || strings.HasPrefix(path, "/super/") || strings.HasPrefix(path, "/auth/") || path == "/login.html" || path == "/registrar_nuevo_usuario_administrador.html"
			isStatic := strings.HasPrefix(path, "/img/") || strings.HasPrefix(path, "/css/") || strings.HasPrefix(path, "/js/") || path == "/estilos.css" || path == "/menu.js" || path == "/manifest.webmanifest" || path == "/sw.js"
			if !isMntAllowed && !isStatic {
				http.Redirect(w, r, "/mantenimiento.html", http.StatusTemporaryRedirect)
				return
			}
		}

		// Rutas públicas exactas (no usar prefijo "/" porque abriría todo el sistema).
		publicExact := map[string]struct{}{
			"/":                                                     {},
			"/index.html":                                           {},
			"/mantenimiento.html":                                   {},
			"/descripcion_de_los_sistemas.ht":                       {},
			"/descripcion_de_los_sistemas.html":                     {},
			"/Informacion_de_contacto.html":                         {},
			"/soporte_remoto_acceso.html":                           {},
			"/red_social_comercial.html":                            {},
			"/perfil_red_social.html":                               {},
			"/venta_publica.html":                                   {},
			"/visualizar_productos_y_precios_publico.html":          {},
			"/pagar_productos_de_venta_publica.html":                {},
			"/pagar_licencia.html":                                  {},
			"/venta_digital.html":                                   {},
			"/elegir_licencia.html":                                 {},
			"/login.html":                                           {},
			"/registrar_nuevo_usuario_administrador.html":           {},
			"/login_usuario.html":                                   {},
			"/contrato.html":                                        {},
			"/super/api/administradores/register":                   {},
			"/super/api/administradores/login":                      {},
			"/super/api/administradores/solicitar_recuperacion":     {},
			"/super/api/administradores/restablecer_password":       {},
			"/super/api/empresas/compartidos/aceptar":               {},
			"/api/asesor_comercial/aceptar":                         {},
			"/auth/google/login":                                    {},
			"/auth/google/usuario/login":                            {},
			"/auth/google/callback":                                 {},
			"/auth/confirmar_correo":                                {},
			"/auth/confirmar_admin":                                 {},
			"/auth/logout":                                          {},
			"/api/v1/auth/login":                                    {},
			"/api/public/venta_publica":                             {},
			"/api/public/taxi_system":                               {},
			"/api/public/estacion_vip":                              {},
			"/api/public/chat_portal":                               {},
			"/api/public/chat_portal_stream":                        {},
			"/api/public/mensajes_privados":                         {},
			"/api/public/publicaciones":                             {},
			"/api/public/market_symbol":                             {},
			"/api/public/soporte_remoto":                            {},
			"/api/public/venta_digital":                             {},
			"/api/public/pagina_principal":                          {},
			"/api/public/informacion_de_modulos":                    {},
			"/api/public/plantillas_nuevas/catalogo":                {},
			"/api/public/plantillas_integracion/catalogo":           {},
			"/api/internal/email_corporativo/autologin":             {},
			"/api/public/contrato":                                  {},
			"/api/public/turnos_atencion":                           {},
			"/api/public/licencias/payment_methods":                 {},
			"/api/public/licencias/checkout_summary":                {},
			"/licencias/activar_sin_pago":                           {},
			"/api/empresa/usuarios/login":                           {},
			"/api/empresa/usuarios/establecer_password":             {},
			"/api/empresa/usuarios/solicitar_recuperacion_password": {},
			"/api/empresa/usuarios/restablecer_password":            {},
			"/api/empresa/usuarios/cambiar_password":                {},
			"/api/onlyoffice/file":                                  {},
			"/api/onlyoffice/callback":                              {},
			"/config.js":                                            {},
			"/accept.html":                                          {},
			"/accept/complete":                                      {},
			"/pantalla_publica.html":                                {},
			"/pantalla_turnos.html":                                 {},
			"/turnos_publicos.html":                                 {},
			"/taxi_system.html":                                     {},
			"/taxi_system_conductor.html":                           {},
			"/calculadora.html":                                     {},
			"/productos_estacion_clientes_publico.html":             {},
			"/estilos.css":                                          {},
			"/menu.js":                                              {},
			"/favicon.ico":                                          {},
			"/manifest.webmanifest":                                 {},
			"/sw.js":                                                {},
		}
		if _, ok := publicExact[path]; ok {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(path, "/wompi/") || strings.HasPrefix(path, "/epayco/") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasSuffix(strings.ToLower(path), "/venta_publica.html") {
			trimmed := strings.Trim(path, "/")
			parts := strings.Split(trimmed, "/")
			if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
				next.ServeHTTP(w, r)
				return
			}
		}
		if strings.HasSuffix(strings.ToLower(path), "/visualizar_productos_y_precios_publico.html") {
			trimmed := strings.Trim(path, "/")
			parts := strings.Split(trimmed, "/")
			if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
				next.ServeHTTP(w, r)
				return
			}
		}
		if strings.HasSuffix(strings.ToLower(path), "/pagar_productos_de_venta_publica.html") {
			trimmed := strings.Trim(path, "/")
			parts := strings.Split(trimmed, "/")
			if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Recursos estáticos públicos
		publicPrefixes := []string{"/assets/", "/img/", "/uploads/"}
		publicPrefixes = append(publicPrefixes, "/js/")
		if strings.HasPrefix(path, "/ayuda/") && path != "/ayuda/ayuda.html" {
			next.ServeHTTP(w, r)
			return
		}
		for _, p := range publicPrefixes {
			if strings.HasPrefix(path, p) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Obtener token desde cookie o header Authorization
		var token string
		if c, err := r.Cookie("session_token"); err == nil {
			token = c.Value
		} else if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
		if token == "" {
			writeAuthFailure(w, r, http.StatusUnauthorized)
			return
		}

		sess, err := getCachedSessionByToken(dbSuper, token)
		if err != nil {
			writeAuthFailure(w, r, http.StatusUnauthorized)
			return
		}

		admin, err := getCachedAdminByEmailFull(dbSuper, sess.AdminEmail)
		if err != nil {
			writeAuthFailure(w, r, http.StatusUnauthorized)
			return
		}

		if path == "/ayuda/ayuda.html" && !IsSuperAdministradorRole(admin.Role) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if path == "/super_administrador.html" && !IsSuperPanelRole(admin.Role) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Rutas /super/ requieren rol super_administrador, excepto lecturas puntuales
		// necesarias para el selector de empresas del rol administrador.
		if strings.HasPrefix(path, "/super/") && !IsSuperAdministradorRole(admin.Role) && !allowSuperControlRoute(path, r.Method, admin.Role) && !allowAdminLimitedSuperRoute(path, r.Method, admin.Role) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Propagar información del admin en el contexto
		ctx := context.WithValue(r.Context(), "adminEmail", admin.Email)
		r = r.WithContext(ctx)
		// Añadir cabecera informativa
		r.Header.Set("X-Admin-Email", admin.Email)

		next.ServeHTTP(w, r)
	})
}

// writeAuthFailure preserves the stable v1 envelope when global authentication
// stops a mobile request before its route-specific handler can run.
func writeAuthFailure(w http.ResponseWriter, r *http.Request, status int) {
	if strings.HasPrefix(strings.TrimSpace(r.URL.Path), "/api/v1/") {
		requestID := requestIDFromContext(r.Context())
		if requestID == "" {
			requestID = makeRequestID()
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok": false,
			"error": map[string]string{
				"code":    "unauthenticated",
				"message": "No fue posible completar la solicitud.",
			},
			"request_id": requestID,
		})
		return
	}
	http.Error(w, "unauthorized", status)
}

// getEncKeyFromEnv intenta obtener la clave de cifrado desde la variable `CONFIG_ENC_KEY`.
// Se admite una cadena Base64 (preferida) o una cadena de al menos 32 bytes.
// EncryptString cifra texto plano usando AES-GCM delegando al paquete secure.
func EncryptString(plain string) (string, error) {
	return secure.EncryptString(plain)
}

// DecryptString descifra un payload cifrado delegando al paquete secure.
func DecryptString(payload string) (string, error) {
	return secure.DecryptString(payload)
}

// EncryptionAvailable devuelve true si la variable de entorno CONFIG_ENC_KEY está disponible y válida.
func EncryptionAvailable() bool {
	return secure.EncryptionAvailable()
}
