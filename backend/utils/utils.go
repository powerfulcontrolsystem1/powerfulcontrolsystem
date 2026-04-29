package utils

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
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
	canonicalPublicApexHost                   = "powerfulcontrolsystem.com"
	canonicalPublicWWWHost                    = "www.powerfulcontrolsystem.com"
	reservedSuperAdminEmail                   = "powerfulcontrolsystem@gmail.com"
)

var companyLogMu sync.Mutex

func AdminShouldUseSuperRole(email string) bool {
	return strings.EqualFold(strings.TrimSpace(email), reservedSuperAdminEmail)
}

func ManagedAdminRole(email, currentRole string) string {
	normalizedCurrent := strings.ToLower(strings.TrimSpace(currentRole))
	if normalizedCurrent != "" && normalizedCurrent != "administrador" && normalizedCurrent != "super_administrador" {
		return strings.TrimSpace(currentRole)
	}
	if AdminShouldUseSuperRole(email) {
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

func requestIDFromContext(ctx context.Context) string {
	v := ctx.Value(ctxKeyRequestID)
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

// RequestIDFromContext expone el request_id propagado por middleware para otros paquetes.
func RequestIDFromContext(ctx context.Context) string {
	return requestIDFromContext(ctx)
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
	if id := parsePositiveInt64(r.URL.Query().Get("empresa_id")); id > 0 {
		return id
	}
	if id := parsePositiveInt64(r.Header.Get("X-Empresa-ID")); id > 0 {
		return id
	}
	if id := empresaIDFromContext(r.Context()); id > 0 {
		return id
	}

	if r.Body == nil {
		return 0
	}
	method := strings.ToUpper(strings.TrimSpace(r.Method))
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
		return 0
	}
	contentType := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Type")))
	if !strings.Contains(contentType, "application/json") {
		return 0
	}

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewReader(raw))
		return 0
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	return extractEmpresaIDFromBody(raw)
}

func requestClientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return v
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
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
	if host := firstForwardedHeaderValue(r.Header.Get("X-Forwarded-Host")); host != "" {
		return host
	}
	return strings.TrimSpace(r.Host)
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
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		log.Printf("warning: no se pudo crear carpeta de logs %s: %v", logDir, err)
		return
	}

	fileName := "empresa_global.log"
	if empresaID > 0 {
		fileName = fmt.Sprintf("empresa_%d.log", empresaID)
	}
	filePath := filepath.Join(logDir, fileName)
	line := fmt.Sprintf("%s [%s] %s\n", time.Now().Format(time.RFC3339), strings.ToUpper(strings.TrimSpace(level)), strings.TrimSpace(msg))

	companyLogMu.Lock()
	defer companyLogMu.Unlock()

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("warning: no se pudo abrir log de empresa %s: %v", filePath, err)
		return
	}
	defer f.Close()

	if _, err := f.WriteString(line); err != nil {
		log.Printf("warning: no se pudo escribir log de empresa %s: %v", filePath, err)
	}
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
		log.Printf("-> req_id=%s empresa_id=%d %s %s from %s", requestID, empresaID, r.Method, r.URL.RequestURI(), clientIP)
		writeCompanyLogEntry(empresaID, "INFO", fmt.Sprintf("req_id=%s event=request_start method=%s path=%s ip=%s ua=%q", requestID, r.Method, r.URL.RequestURI(), clientIP, r.UserAgent()))

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
		writeCompanyLogEntry(finalEmpresaID, level, fmt.Sprintf("req_id=%s event=request_end method=%s path=%s status=%d dur_ms=%d", requestID, r.Method, r.URL.RequestURI(), lrw.status, elapsedMs))
	})
}

// JSONErrorMiddleware unifica errores no-JSON para endpoints de API.
func JSONErrorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isLikelyAPIRequest(r) {
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

			if capture.status >= http.StatusInternalServerError {
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

// GenerateSecureToken devuelve un token seguro en hex de `n` bytes.
func GenerateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func allowAdminLimitedSuperRoute(path, method, role string) bool {
	if !strings.EqualFold(strings.TrimSpace(role), "administrador") {
		return false
	}
	switch strings.TrimSpace(path) {
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

// AuthMiddleware protege rutas usando la tabla sesiones y administradores en la BD superadministrador.
// Permite un conjunto público de rutas (login/callback/activos). Para rutas que comienzan con /super/
// exige rol 'super_administrador'. Añade `adminEmail` en el contexto de la petición.
func AuthMiddleware(dbSuper *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Verificación de Modo Mantenimiento
		if dbSuper != nil {
			if val, _, err := dbpkg.GetConfigValue(dbSuper, "mantenimiento_activo"); err == nil && val == "true" {
				// Rutas permitidas durante el mantenimiento (acceso de administradores y assets estáticos)
				isMntAllowed := path == "/mantenimiento.html" || strings.HasPrefix(path, "/super/") || strings.HasPrefix(path, "/auth/") || path == "/login.html" || path == "/registrar_nuevo_usuario_administrador.html"
				isStatic := strings.HasPrefix(path, "/img/") || strings.HasPrefix(path, "/css/") || strings.HasPrefix(path, "/js/") || strings.HasPrefix(path, "/descargas/") || path == "/estilos.css" || path == "/menu.js"
				if !isMntAllowed && !isStatic {
					http.Redirect(w, r, "/mantenimiento.html", http.StatusTemporaryRedirect)
					return
				}
			}
		}

		// Rutas públicas exactas (no usar prefijo "/" porque abriría todo el sistema).
		publicExact := map[string]struct{}{
			"/":                                                     {},
			"/index.html":                                           {},
			"/descripcion_de_los_sistemas.ht":                       {},
			"/Informacion_de_contacto.html":                         {},
			"/soporte_remoto_acceso.html":                           {},
			"/venta_publica.html":                                   {},
			"/pagar_productos_de_venta_publica.html":                {},
			"/venta_digital.html":                                   {},
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
			"/auth/google/callback":                                 {},
			"/auth/confirmar_correo":                                {},
			"/auth/confirmar_admin":                                 {},
			"/auth/logout":                                          {},
			"/api/public/venta_publica":                             {},
			"/api/public/estacion_vip":                              {},
			"/api/public/chat_portal":                               {},
			"/api/public/chat_portal_stream":                        {},
			"/api/public/mensajes_privados":                          {},
			"/api/public/publicaciones":                             {},
			"/api/public/soporte_remoto":                            {},
			"/api/public/venta_digital":                             {},
			"/api/public/pagina_principal":                          {},
			"/api/public/contrato":                                  {},
			"/api/public/licencias/payment_methods":                 {},
			"/api/empresa/usuarios/login":                           {},
			"/api/empresa/usuarios/establecer_password":             {},
			"/api/empresa/usuarios/solicitar_recuperacion_password": {},
			"/api/empresa/usuarios/restablecer_password":            {},
			"/api/empresa/usuarios/cambiar_password":                {},
			"/config.js":                                            {},
			"/accept.html":                                          {},
			"/accept/complete":                                      {},
			"/productos_estacion_clientes_publico.html":             {},
			"/estilos.css":                                          {},
			"/menu.js":                                              {},
			"/favicon.ico":                                          {},
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
		if strings.HasSuffix(strings.ToLower(path), "/pagar_productos_de_venta_publica.html") {
			trimmed := strings.Trim(path, "/")
			parts := strings.Split(trimmed, "/")
			if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Recursos estáticos públicos
		publicPrefixes := []string{"/assets/", "/img/", "/ayuda/", "/uploads/", "/descargas/"}
		publicPrefixes = append(publicPrefixes, "/js/")
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
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		sess, err := dbpkg.GetSessionByToken(dbSuper, token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		admin, err := dbpkg.GetAdminByEmailFull(dbSuper, sess.AdminEmail)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		desiredRole := ManagedAdminRole(admin.Email, admin.Role)
		if admin.ID > 0 && !strings.EqualFold(strings.TrimSpace(admin.Role), desiredRole) {
			if err := dbpkg.UpdateAdministrador(dbSuper, admin.ID, admin.Name, desiredRole); err == nil {
				admin.Role = desiredRole
			}
		}

		// Rutas /super/ requieren rol super_administrador, excepto lecturas puntuales
		// necesarias para el selector de empresas del rol administrador.
		if strings.HasPrefix(path, "/super/") && !strings.EqualFold(strings.TrimSpace(admin.Role), "super_administrador") && !allowAdminLimitedSuperRoute(path, r.Method, admin.Role) {
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
