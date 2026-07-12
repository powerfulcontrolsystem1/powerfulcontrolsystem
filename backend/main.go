package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/handlers"
	"github.com/you/pos-backend/metrics"
	"github.com/you/pos-backend/utils"
	"github.com/you/pos-backend/vpssecurity"
)

var (
	clientID      = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret  = os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL   = os.Getenv("GOOGLE_REDIRECT_URL") // e.g. http://localhost:8080/auth/google/callback
	dbEmpresasDSN = os.Getenv("DB_EMPRESAS_DSN")
	dbSuperDSN    = os.Getenv("DB_SUPERADMIN_DSN")
	dbEmpresas    *sql.DB
	dbSuper       *sql.DB
)

func resolveBackendRuntimeDir() string {
	candidates := []string{".", "backend"}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd, filepath.Join(wd, "backend"))
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			exeDir,
			filepath.Join(exeDir, "backend"),
			filepath.Join(exeDir, ".."),
			filepath.Join(exeDir, "..", "backend"),
		)
	}

	seen := map[string]bool{}
	for _, cand := range candidates {
		cand = strings.TrimSpace(cand)
		if cand == "" {
			continue
		}
		absCand, err := filepath.Abs(cand)
		if err != nil {
			absCand = cand
		}
		if seen[absCand] {
			continue
		}
		seen[absCand] = true

		goModPath := filepath.Join(absCand, "go.mod")
		if info, statErr := os.Stat(goModPath); statErr == nil && !info.IsDir() {
			return absCand
		}
	}

	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

func loadEnvDefaultsFromFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}

	added := 0
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for _, line := range lines {
		raw := strings.TrimSpace(line)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}

		idx := strings.Index(raw, "=")
		if idx <= 0 {
			continue
		}

		key := strings.TrimSpace(raw[:idx])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(raw[idx+1:])

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
			value = value[1 : len(value)-1]
		}
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") && len(value) >= 2 {
			value = value[1 : len(value)-1]
		}

		if os.Getenv(key) == "" && value != "" {
			if setErr := os.Setenv(key, value); setErr != nil {
				return added, setErr
			}
			added++
		}
	}

	return added, nil
}

func loadRuntimeEnvDefaults(backendDir string) {
	candidates := []string{
		filepath.Join(backendDir, ".env.local"),
		filepath.Join(backendDir, ".env"),
	}
	for _, candidate := range candidates {
		added, err := loadEnvDefaultsFromFile(candidate)
		if err != nil {
			log.Printf("warning: no se pudieron cargar variables desde %s: %v", candidate, err)
			continue
		}
		if added > 0 {
			log.Printf("INFO: variables de entorno cargadas desde %s (%d nuevas)", candidate, added)
		}
	}

	platformEnvCandidates := []string{
		filepath.Join(filepath.Dir(backendDir), "deploy", ".env.platform"),
	}
	if exportRoot := strings.TrimSpace(os.Getenv("PCS_PROJECT_EXPORT_ROOT")); exportRoot != "" {
		platformEnvCandidates = append(platformEnvCandidates, filepath.Join(exportRoot, "deploy", ".env.platform"))
	}
	loadedPlatformEnv := map[string]bool{}
	for _, platformEnv := range platformEnvCandidates {
		if strings.TrimSpace(platformEnv) == "" || loadedPlatformEnv[platformEnv] {
			continue
		}
		loadedPlatformEnv[platformEnv] = true
		if added, err := loadSelectedEnvDefaultsFromFile(platformEnv, []string{"OPENAI_API_KEY"}); err == nil && added > 0 {
			log.Printf("INFO: fallback IA cargado desde %s (%d variable sensible, valor oculto)", platformEnv, added)
		} else if err != nil {
			log.Printf("warning: no se pudo cargar fallback IA desde %s: %v", platformEnv, err)
		}
	}
}

func loadSelectedEnvDefaultsFromFile(path string, allowedKeys []string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}

	allowed := make(map[string]bool, len(allowedKeys))
	for _, key := range allowedKeys {
		key = strings.TrimSpace(key)
		if key != "" {
			allowed[key] = true
		}
	}
	if len(allowed) == 0 {
		return 0, nil
	}

	added := 0
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for _, line := range lines {
		raw := strings.TrimSpace(line)
		if raw == "" || strings.HasPrefix(raw, "#") {
			continue
		}
		idx := strings.Index(raw, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(raw[:idx])
		if !allowed[key] || os.Getenv(key) != "" {
			continue
		}
		value := strings.TrimSpace(raw[idx+1:])
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") && len(value) >= 2 {
			value = value[1 : len(value)-1]
		}
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") && len(value) >= 2 {
			value = value[1 : len(value)-1]
		}
		if value == "" {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return added, err
		}
		added++
	}
	return added, nil
}

func refreshRuntimeGlobalsFromEnv() {
	if v := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID")); v != "" {
		clientID = v
	}
	if v := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET")); v != "" {
		clientSecret = v
	}
	if v := strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL")); v != "" {
		redirectURL = v
	}
	if v := strings.TrimSpace(os.Getenv("DB_EMPRESAS_DSN")); v != "" {
		dbEmpresasDSN = v
	}
	if v := strings.TrimSpace(os.Getenv("DB_SUPERADMIN_DSN")); v != "" {
		dbSuperDSN = v
	}
}

func resolveRuntimeDBPath(rawPath, defaultFileName, backendDir string) string {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" {
		return filepath.Join(backendDir, "db", defaultFileName)
	}

	if filepath.IsAbs(trimmed) {
		return trimmed
	}

	return filepath.Join(backendDir, trimmed)
}

func normalizeRuntimeDBDialect(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return ""
	}
	if strings.Contains(v, "postgres") {
		return "postgres"
	}
	return ""
}

func resolveRuntimeDBDialect() string {
	candidates := []string{
		strings.TrimSpace(os.Getenv("DB_DIALECT")),
		strings.TrimSpace(os.Getenv("DB_ENGINE")),
		strings.TrimSpace(os.Getenv("PCS_DB_DIALECT")),
	}
	for _, candidate := range candidates {
		if dialect := normalizeRuntimeDBDialect(candidate); dialect != "" {
			return dialect
		}
	}
	return "postgres"
}

func resolveRuntimePostgresDSN(primary string, fallbackKeys ...string) string {
	if v := strings.TrimSpace(primary); v != "" {
		return rewriteRuntimePostgresDSNForTunnel(v)
	}
	for _, key := range fallbackKeys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return rewriteRuntimePostgresDSNForTunnel(v)
		}
	}
	return ""
}

func rewriteRuntimePostgresDSNForTunnel(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if strings.TrimSpace(os.Getenv("DB_VPS_TUNNEL_ENABLED")) != "1" {
		return raw
	}
	localPort := strings.TrimSpace(os.Getenv("DB_VPS_LOCAL_PORT"))
	if localPort == "" {
		return raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	hostname := u.Hostname()
	if hostname == "" {
		hostname = "127.0.0.1"
	}
	if hostname != "127.0.0.1" && hostname != "localhost" {
		return raw
	}
	u.Host = net.JoinHostPort("127.0.0.1", localPort)
	return u.String()
}

func openAndPingRuntimeDB(driverName, dsn, label string) (*sql.DB, error) {
	dbConn, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s db with driver %s: %w", label, driverName, err)
	}
	if err := dbConn.Ping(); err != nil {
		_ = dbConn.Close()
		return nil, fmt.Errorf("failed to ping %s db with driver %s: %w", label, driverName, err)
	}
	return dbConn, nil
}

func ensureRuntimeDBDir(dbPath string) error {
	dir := strings.TrimSpace(filepath.Dir(dbPath))
	if dir == "" || dir == "." {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func persistConfigEncKey(backendDir, value string) (string, error) {
	envLocalPath := filepath.Join(backendDir, ".env.local")
	prefix := "CONFIG_ENC_KEY="

	data, err := os.ReadFile(envLocalPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	if err == nil {
		lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
		replaced := false
		for i := range lines {
			trimmed := strings.TrimSpace(lines[i])
			if strings.HasPrefix(trimmed, prefix) {
				lines[i] = prefix + value
				replaced = true
				break
			}
		}
		if !replaced {
			if len(lines) == 1 && strings.TrimSpace(lines[0]) == "" {
				lines[0] = prefix + value
			} else {
				lines = append(lines, prefix+value)
			}
		}
		content := strings.Join(lines, "\n")
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		if writeErr := os.WriteFile(envLocalPath, []byte(content), 0600); writeErr != nil {
			return "", writeErr
		}
		return envLocalPath, nil
	}

	content := "# Archivo local de entorno (secrets de desarrollo; no versionar)\n" + prefix + value + "\n"
	if writeErr := os.WriteFile(envLocalPath, []byte(content), 0600); writeErr != nil {
		return "", writeErr
	}
	return envLocalPath, nil
}

func ensureRuntimeConfigEncKey(backendDir string) error {
	raw := strings.TrimSpace(os.Getenv("CONFIG_ENC_KEY"))
	if raw != "" {
		if !utils.EncryptionAvailable() {
			return fmt.Errorf("CONFIG_ENC_KEY invalid; use base64 valido o >=32 bytes")
		}
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("PCS_ENV")), "production") || strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production") {
		return fmt.Errorf("CONFIG_ENC_KEY es obligatoria en produccion y debe ser base64 de 32 bytes")
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("no se pudo generar CONFIG_ENC_KEY: %w", err)
	}

	generated := base64.StdEncoding.EncodeToString(key)
	if err := os.Setenv("CONFIG_ENC_KEY", generated); err != nil {
		return fmt.Errorf("no se pudo cargar CONFIG_ENC_KEY en entorno: %w", err)
	}

	envLocalPath, err := persistConfigEncKey(backendDir, generated)
	if err != nil {
		return fmt.Errorf("no se pudo persistir CONFIG_ENC_KEY en .env.local: %w", err)
	}

	if !utils.EncryptionAvailable() {
		return fmt.Errorf("CONFIG_ENC_KEY generada pero invalida")
	}

	log.Printf("INFO: CONFIG_ENC_KEY autogenerada para desarrollo y persistida en %s", envLocalPath)
	return nil
}

func runtimeProductionMode() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("PCS_ENV")), "production") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
}

func validateProductionSecurityConfig() error {
	if !runtimeProductionMode() {
		return nil
	}
	if strings.TrimSpace(os.Getenv("PCS_CSRF_ALLOWED_ORIGINS")) == "" && strings.TrimSpace(os.Getenv("CSRF_ALLOWED_ORIGINS")) != "" {
		if err := os.Setenv("PCS_CSRF_ALLOWED_ORIGINS", strings.TrimSpace(os.Getenv("CSRF_ALLOWED_ORIGINS"))); err != nil {
			return err
		}
	}
	required := []string{
		"PCS_TRUSTED_PROXY_CIDRS",
		"CONFIG_ENC_KEY_ID",
		"PCS_CSRF_ALLOWED_ORIGINS",
		"SESSION_TIMEOUT",
		"MAX_REQUEST_BODY_BYTES",
		"HTTP_READ_TIMEOUT",
		"HTTP_WRITE_TIMEOUT",
		"HTTP_IDLE_TIMEOUT",
	}
	for _, key := range required {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return fmt.Errorf("%s es obligatoria en produccion", key)
		}
	}
	for _, key := range []string{"SESSION_TIMEOUT", "HTTP_READ_TIMEOUT", "HTTP_WRITE_TIMEOUT", "HTTP_IDLE_TIMEOUT"} {
		if value, err := time.ParseDuration(strings.TrimSpace(os.Getenv(key))); err != nil || value <= 0 {
			return fmt.Errorf("%s debe ser una duracion positiva valida", key)
		}
	}
	maxBody, err := strconv.ParseInt(strings.TrimSpace(os.Getenv("MAX_REQUEST_BODY_BYTES")), 10, 64)
	if err != nil || maxBody < 1<<20 || maxBody > 512<<20 {
		return fmt.Errorf("MAX_REQUEST_BODY_BYTES debe estar entre 1 MiB y 512 MiB")
	}
	for _, raw := range strings.Split(os.Getenv("PCS_CSRF_ALLOWED_ORIGINS"), ",") {
		origin, err := url.Parse(strings.TrimSpace(raw))
		if err != nil || (origin.Scheme != "https" && origin.Scheme != "http") || origin.Host == "" || origin.User != nil || origin.Path != "" {
			return fmt.Errorf("PCS_CSRF_ALLOWED_ORIGINS contiene un origen invalido")
		}
	}
	return nil
}

func getenvDurationRange(key string, defaultValue, minValue, maxValue time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return defaultValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func getenvInt64Range(key string, defaultValue, minValue, maxValue int64) int64 {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return defaultValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func getenvIntRange(key string, defaultVal, minVal, maxVal int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("warning: %s invalido (%q), se usa valor por defecto %d", key, raw, defaultVal)
		return defaultVal
	}
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func resolveAsientosWorkerPolicy() (time.Duration, int, int) {
	intervalMinutes := getenvIntRange("ASIENTOS_WORKER_INTERVAL_MINUTES", 15, 1, 1440)
	batchSize := getenvIntRange("ASIENTOS_WORKER_BATCH_SIZE", 100, 1, 500)
	maxRetries := getenvIntRange("ASIENTOS_WORKER_MAX_RETRIES", 5, 1, 50)
	return time.Duration(intervalMinutes) * time.Minute, batchSize, maxRetries
}

func readConfigValueFromDB(dbConn *sql.DB, keys []string) (string, string, error) {
	for _, key := range keys {
		val, enc, err := dbpkg.GetConfigValue(dbConn, key)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return "", "", err
		}

		clean := strings.TrimSpace(val)
		if clean == "" {
			continue
		}

		if enc {
			dec, derr := utils.DecryptString(clean)
			if derr != nil {
				log.Printf("warning: no se pudo descifrar la configuración %s: %v", key, derr)
				continue
			}
			clean = strings.TrimSpace(dec)
			if clean == "" {
				continue
			}
		}

		return clean, key, nil
	}

	return "", "", nil
}

func loadGoogleOAuthFromDB(dbConn *sql.DB) {
	clientIDKeys := []string{
		"google.client_id",
		"oauth.google.client_id",
		"auth.google.client_id",
		"google_oauth.client_id",
		"GOOGLE_CLIENT_ID",
	}
	clientSecretKeys := []string{
		"google.client_secret",
		"oauth.google.client_secret",
		"auth.google.client_secret",
		"google_oauth.client_secret",
		"GOOGLE_CLIENT_SECRET",
	}
	redirectURLKeys := []string{
		"google.redirect_url",
		"oauth.google.redirect_url",
		"auth.google.redirect_url",
		"google_oauth.redirect_url",
		"GOOGLE_REDIRECT_URL",
	}

	dbClientID, clientIDKey, err := readConfigValueFromDB(dbConn, clientIDKeys)
	if err != nil {
		log.Printf("warning: no se pudo leer GOOGLE_CLIENT_ID desde DB: %v", err)
	}
	dbClientSecret, clientSecretKey, err := readConfigValueFromDB(dbConn, clientSecretKeys)
	if err != nil {
		log.Printf("warning: no se pudo leer GOOGLE_CLIENT_SECRET desde DB: %v", err)
	}
	dbRedirectURL, redirectURLKey, err := readConfigValueFromDB(dbConn, redirectURLKeys)
	if err != nil {
		log.Printf("warning: no se pudo leer GOOGLE_REDIRECT_URL desde DB: %v", err)
	}

	// Prioridad: variables de entorno > configuración en DB.
	// La DB solo completa faltantes para evitar sobreescrituras inesperadas en VPS.
	if clientID == "" && dbClientID != "" {
		clientID = dbClientID
		log.Printf("INFO: GOOGLE_CLIENT_ID completado desde DB (%s)", clientIDKey)
	}
	if clientSecret == "" && dbClientSecret != "" {
		clientSecret = dbClientSecret
		log.Printf("INFO: GOOGLE_CLIENT_SECRET completado desde DB (%s)", clientSecretKey)
	}
	if redirectURL == "" && dbRedirectURL != "" {
		redirectURL = dbRedirectURL
		log.Printf("INFO: GOOGLE_REDIRECT_URL completado desde DB (%s)", redirectURLKey)
	}

	if clientID != "" && clientSecret != "" {
		log.Printf("INFO: OAuth Google listo (client_id/config secret activos)")
	}
}

func resolveWebDir() string {
	candidates := []string{
		"web",
		"../web",
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "web"),
			filepath.Join(exeDir, "..", "web"),
			filepath.Join(exeDir, "..", "..", "web"),
			filepath.Join(exeDir, "..", "..", "..", "web"),
		)
	}

	seen := map[string]bool{}
	fallback := ""
	for _, cand := range candidates {
		cand = strings.TrimSpace(cand)
		if cand == "" {
			continue
		}

		absCand, err := filepath.Abs(cand)
		if err != nil {
			absCand = cand
		}
		if seen[absCand] {
			continue
		}
		seen[absCand] = true

		info, statErr := os.Stat(absCand)
		if statErr != nil || !info.IsDir() {
			continue
		}

		if fallback == "" {
			fallback = absCand
		}

		indexPath := filepath.Join(absCand, "index.html")
		if idxInfo, idxErr := os.Stat(indexPath); idxErr == nil && !idxInfo.IsDir() {
			return absCand
		}
	}

	if fallback != "" {
		return fallback
	}
	return "web"
}

func noCacheAdminStaticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.ToLower(strings.TrimSpace(r.URL.Path))
		if strings.HasPrefix(path, "/administrar_empresa/") ||
			strings.HasPrefix(path, "/super/") ||
			path == "/administrar_empresa.html" ||
			path == "/seleccionar_empresa.html" ||
			path == "/js/administrar_empresa.js" ||
			path == "/js/plantillas_nuevas_catalogo.js" ||
			path == "/menu.js" ||
			strings.HasSuffix(path, ".html") {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		next.ServeHTTP(w, r)
	})
}

type contextualHelpCaptureWriter struct {
	header http.Header
	body   []byte
	status int
}

func (w *contextualHelpCaptureWriter) Header() http.Header {
	return w.header
}

func (w *contextualHelpCaptureWriter) WriteHeader(status int) {
	if w.status == 0 {
		w.status = status
	}
}

func (w *contextualHelpCaptureWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.body = append(w.body, data...)
	return len(data), nil
}

func contextualHelpStaticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

const buttonIconsScriptTag = "\n<script src=\"/js/button_icons.js?v=20260521-global-button-icons\" defer></script>\n"

func buttonIconsStaticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}
		if !shouldCaptureButtonIconsRequest(r) {
			next.ServeHTTP(w, r)
			return
		}

		capture := &contextualHelpCaptureWriter{header: make(http.Header)}
		next.ServeHTTP(capture, r)
		status := capture.status
		if status == 0 {
			status = http.StatusOK
		}

		body := capture.body
		if shouldInjectButtonIcons(r, capture.Header(), status, body) {
			body = injectButtonIconsScript(body)
			capture.Header().Set("Content-Length", strconv.Itoa(len(body)))
		}

		for key, values := range capture.Header() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(status)
		_, _ = w.Write(body)
	})
}

func shouldCaptureButtonIconsRequest(r *http.Request) bool {
	path := strings.ToLower(strings.TrimSpace(r.URL.Path))
	if path == "" || path == "/" {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	return ext == "" || ext == ".html" || ext == ".htm"
}

func shouldInjectButtonIcons(r *http.Request, header http.Header, status int, body []byte) bool {
	if status < 200 || status >= 300 || len(body) == 0 {
		return false
	}
	path := strings.ToLower(strings.TrimSpace(r.URL.Path))
	contentType := strings.ToLower(header.Get("Content-Type"))
	if !strings.Contains(contentType, "text/html") && !strings.HasSuffix(path, ".html") && path != "/" {
		return false
	}
	text := strings.ToLower(string(body))
	if strings.Contains(text, "/js/button_icons.js") {
		return false
	}
	return strings.Contains(text, "<body") || strings.Contains(text, "</body>")
}

func injectButtonIconsScript(body []byte) []byte {
	text := string(body)
	lower := strings.ToLower(text)
	insertAt := strings.LastIndex(lower, "</body>")
	if insertAt < 0 {
		return append(body, []byte(buttonIconsScriptTag)...)
	}
	return []byte(text[:insertAt] + buttonIconsScriptTag + text[insertAt:])
}

func resolveDownloadsDir() string {
	candidates := []string{
		"descargas",
		"../descargas",
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "descargas"),
			filepath.Join(exeDir, "..", "descargas"),
			filepath.Join(exeDir, "..", "..", "descargas"),
			filepath.Join(exeDir, "..", "..", "..", "descargas"),
		)
	}

	seen := map[string]bool{}
	for _, cand := range candidates {
		cand = strings.TrimSpace(cand)
		if cand == "" {
			continue
		}

		absCand, err := filepath.Abs(cand)
		if err != nil {
			absCand = cand
		}
		if seen[absCand] {
			continue
		}
		seen[absCand] = true

		info, statErr := os.Stat(absCand)
		if statErr != nil || !info.IsDir() {
			continue
		}
		return absCand
	}

	if absCand, err := filepath.Abs("../descargas"); err == nil {
		return absCand
	}
	return "../descargas"
}

func main() {
	backendDir := resolveBackendRuntimeDir()
	startupTraceEnabled := strings.TrimSpace(os.Getenv("PCS_STARTUP_TRACE")) == "1"
	startupTrace := func(step string) {
		if startupTraceEnabled {
			log.Printf("STARTUP TRACE: %s", strings.TrimSpace(step))
		}
	}

	startupTrace("main_enter")
	loadRuntimeEnvDefaults(backendDir)
	startupTrace("after_load_runtime_env_defaults")
	refreshRuntimeGlobalsFromEnv()
	startupTrace("after_refresh_runtime_globals")
	if err := ensureRuntimeConfigEncKey(backendDir); err != nil {
		log.Fatalf("failed to ensure CONFIG_ENC_KEY: %v", err)
	}
	if err := validateProductionSecurityConfig(); err != nil {
		log.Fatalf("invalid production security configuration: %v", err)
	}
	startupTrace("after_ensure_runtime_config_enc_key")
	runtimeDBDialect := resolveRuntimeDBDialect()
	startupTrace("after_resolve_runtime_db_dialect")
	runtimePostgres := runtimeDBDialect == "postgres"
	if !runtimePostgres {
		log.Fatalf("DB_DIALECT=%q no soportado. La migracion fue cerrada a PostgreSQL-only", runtimeDBDialect)
	}

	if redirectURL == "" {
		log.Println("INFO: GOOGLE_REDIRECT_URL no configurado; se resolvera dinamicamente segun host de la solicitud")
	}

	var err error
	if runtimePostgres {
		if strings.TrimSpace(os.Getenv("DB_DIALECT")) == "" {
			_ = os.Setenv("DB_DIALECT", "postgres")
		}

		dbEmpresasDSN = resolveRuntimePostgresDSN(
			dbEmpresasDSN,
			"DATABASE_EMPRESAS_URL",
			"DB_EMPRESAS_URL",
			"PCS_DB_EMPRESAS_DSN",
		)
		dbSuperDSN = resolveRuntimePostgresDSN(
			dbSuperDSN,
			"DATABASE_SUPERADMIN_URL",
			"DB_SUPERADMIN_URL",
			"PCS_DB_SUPERADMIN_DSN",
		)

		if strings.TrimSpace(dbEmpresasDSN) == "" || strings.TrimSpace(dbSuperDSN) == "" {
			log.Fatalf("modo postgres activo pero faltan DSN: define DB_EMPRESAS_DSN y DB_SUPERADMIN_DSN en backend/.env.local del VPS")
		}

		postgresDriverName := dbpkg.PostgresCompatDriverName()
		startupTrace("before_open_db_empresas")
		dbEmpresas, err = openAndPingRuntimeDB(postgresDriverName, dbEmpresasDSN, "empresas")
		if err != nil {
			log.Fatal(err)
		}
		startupTrace("after_open_db_empresas")
		// Registrar la conexión principal de empresas en el paquete db para wrappers
		dbpkg.SetDefaultDB(dbEmpresas)
		startupTrace("before_open_db_super")
		dbSuper, err = openAndPingRuntimeDB(postgresDriverName, dbSuperDSN, "superadministrador")
		if err != nil {
			log.Fatal(err)
		}
		startupTrace("after_open_db_super")
		if err := dbpkg.EnsurePostgresRuntimeCompat(dbEmpresas); err != nil {
			log.Fatalf("failed to ensure postgres compat functions in empresas db: %v", err)
		}
		startupTrace("after_ensure_pg_compat_empresas")
		if err := dbpkg.EnsurePostgresRuntimeCompat(dbSuper); err != nil {
			log.Fatalf("failed to ensure postgres compat functions in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_pg_compat_super")
		if err := dbpkg.EnsureAdministradoresAuthSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure administradores auth schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_administradores_auth_schema")
		if err := dbpkg.MigrateSessionTokensToHashes(dbSuper); err != nil {
			log.Fatalf("failed to protect existing session tokens: %v", err)
		}
		startupTrace("after_migrate_session_tokens_to_hashes")
		totpMigrationDryRun := strings.EqualFold(strings.TrimSpace(os.Getenv("PCS_TOTP_MIGRATION_DRY_RUN")), "1") || strings.EqualFold(strings.TrimSpace(os.Getenv("PCS_TOTP_MIGRATION_DRY_RUN")), "true")
		migratedTOTP, err := dbpkg.MigrateAdministradorTOTPSecrets(dbSuper, totpMigrationDryRun)
		if err != nil {
			log.Fatalf("failed to protect existing TOTP secrets: %v", err)
		}
		if totpMigrationDryRun {
			log.Printf("INFO: TOTP secret migration dry-run found %d legacy secret(s)", migratedTOTP)
		} else if migratedTOTP > 0 {
			log.Printf("INFO: encrypted %d legacy TOTP secret(s)", migratedTOTP)
		}
		startupTrace("after_migrate_totp_secrets")
		migratedResetTokens, err := dbpkg.MigrateAdministradorPasswordResetTokens(dbSuper, totpMigrationDryRun)
		if err != nil {
			log.Fatalf("failed to protect existing password reset tokens: %v", err)
		}
		if totpMigrationDryRun {
			log.Printf("INFO: password reset token migration dry-run found %d legacy token(s)", migratedResetTokens)
		} else if migratedResetTokens > 0 {
			log.Printf("INFO: protected %d legacy password reset token(s)", migratedResetTokens)
		}
		startupTrace("after_migrate_password_reset_tokens")
		migratedConfirmTokens, err := dbpkg.MigrateAdministradorEmailConfirmTokens(dbSuper, totpMigrationDryRun)
		if err != nil {
			log.Fatalf("failed to protect existing email confirmation tokens: %v", err)
		}
		if totpMigrationDryRun {
			log.Printf("INFO: email confirmation token migration dry-run found %d legacy token(s)", migratedConfirmTokens)
		} else if migratedConfirmTokens > 0 {
			log.Printf("INFO: protected %d legacy email confirmation token(s)", migratedConfirmTokens)
		}
		startupTrace("after_migrate_email_confirmation_tokens")
		if err := dbpkg.EnsurePaymentGatewaySchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure payment gateway schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_payment_gateway_schema")
		if err := dbpkg.EnsureLicenciasSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure licencias schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_licencias_schema")
		if ensured, err := dbpkg.EnsureLicenciasCatalogoGlobal(dbSuper, "sistema.arranque"); err != nil {
			log.Printf("warning: no se pudo asegurar catalogo global de licencias: %v", err)
		} else {
			log.Printf("INFO: catalogo global de licencias verificado: planes=%d", ensured)
		}
		startupTrace("after_ensure_licencias_catalogo_global")
		if empresaSistema, err := dbpkg.EnsurePowerfulSystemEmpresa(dbEmpresas, dbSuper); err != nil {
			log.Printf("warning: no se pudo asegurar empresa interna Powerful Control System: %v", err)
		} else if empresaSistema != nil {
			log.Printf("INFO: empresa interna Powerful Control System verificada: empresa_id=%d", empresaSistema.EmpresaID)
		}
		startupTrace("after_ensure_powerful_system_empresa")
		if err := dbpkg.EnsureSuperAuditoriaSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure super auditoria schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_super_auditoria_schema")
		if err := dbpkg.EnsureSuperVPSSnapshotSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure super vps snapshots schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_super_vps_snapshots_schema")
		if err := dbpkg.EnsureLicenciaVencimientoNotificacionesSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure licencia vencimiento notificaciones schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_licencia_vencimiento_notificaciones_schema")
		if err := dbpkg.EnsureLicenciaEmpresaRetencionSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure licencia empresa retencion schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_licencia_empresa_retencion_schema")
		if err := dbpkg.EnsureUsuarioConfiguracionSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure usuario configuracion schema in superadministrador db: %v", err)
		}
		startupTrace("after_usuario_config_schema")
		if err := dbpkg.EnsureEmpresaEmailCorporativoSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure empresa email corporativo schema in superadministrador db: %v", err)
		}
		startupTrace("after_empresa_email_corporativo_schema")
		if err := handlers.EnsureCorporateEmailConfigFromEnv(dbSuper); err != nil {
			log.Printf("warning: no se pudo registrar configuracion Mailu desde entorno: %v", err)
		}
		startupTrace("after_empresa_email_corporativo_env")
		if strings.TrimSpace(os.Getenv("PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC")) == "1" {
			log.Printf("INFO: sincronizacion inicial de emails corporativos omitida por PCS_SKIP_CORPORATE_EMAIL_STARTUP_SYNC=1")
		} else if created, err := handlers.EnsureCorporateEmailRowsForExistingCompanies(dbSuper, dbEmpresas, "sistema.arranque"); err != nil {
			log.Printf("warning: no se pudieron generar emails corporativos para empresas existentes: %v", err)
		} else if created > 0 {
			log.Printf("INFO: emails corporativos generados para empresas existentes: %d", created)
		}
		startupTrace("after_empresa_email_corporativo_existing_companies")
		if err := dbpkg.DecommissionNextcloudArtifacts(dbEmpresas, dbSuper); err != nil {
			log.Printf("warning: no se pudieron retirar artefactos Nextcloud obsoletos: %v", err)
		}
		startupTrace("after_nextcloud_decommission")
		if err := dbpkg.DecommissionRemovedEntertainmentArtifacts(dbSuper); err != nil {
			log.Printf("warning: no se pudieron retirar artefactos de juegos y emulador: %v", err)
		}
		startupTrace("after_entertainment_decommission")
		if err := dbpkg.EnsureAsesorComercialSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure asesor comercial schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_asesor_schema")
		if seedResult, err := dbpkg.SeedDefaultTipoEmpresaPreconfiguraciones(dbSuper, "sistema.arranque", false); err != nil {
			log.Printf("warning: no se pudieron registrar preconfiguraciones por tipo de empresa: %v", err)
		} else {
			log.Printf("INFO: preconfiguraciones por tipo verificadas: tipos=%d creadas=%d omitidas=%d errores=%d", seedResult.TotalTipos, seedResult.Creadas, seedResult.Omitidas, seedResult.Errores)
		}
		startupTrace("after_seed_default_tipo_empresa")
		if tipoID, licencias, err := dbpkg.EnsureConstructoraTipoEmpresaYLicencias(dbSuper, "sistema.arranque"); err != nil {
			log.Printf("warning: no se pudo asegurar constructora/licencias: %v", err)
		} else {
			log.Printf("INFO: tipo constructora verificado: tipo_id=%d licencias=%d", tipoID, licencias)
		}
		startupTrace("after_ensure_constructora_tipo_licencias")
		if tipoID, licencias, err := dbpkg.EnsureDrogueriaFarmaciaTipoEmpresaYLicencias(dbSuper, "sistema.arranque"); err != nil {
			log.Printf("warning: no se pudo asegurar drogueria/farmacia/licencias: %v", err)
		} else {
			log.Printf("INFO: tipo drogueria/farmacia verificado: tipo_id=%d licencias=%d", tipoID, licencias)
		}
		startupTrace("after_ensure_drogueria_farmacia_tipo_licencias")
		if tipoID, licencias, err := dbpkg.EnsureAlquileresTipoEmpresaYLicencias(dbSuper, "sistema.arranque"); err != nil {
			log.Printf("warning: no se pudo asegurar alquileres/licencias: %v", err)
		} else {
			log.Printf("INFO: tipo alquileres verificado: tipo_id=%d licencias=%d", tipoID, licencias)
		}
		startupTrace("after_ensure_alquileres_tipo_licencias")
		if tipos, licencias, err := dbpkg.EnsureNuevasPlantillasTipoEmpresaYLicencias(dbSuper, "sistema.arranque"); err != nil {
			log.Printf("warning: no se pudieron asegurar nuevas plantillas/licencias: %v", err)
		} else {
			log.Printf("INFO: nuevas plantillas verificados: tipos=%d licencias=%d", tipos, licencias)
		}
		startupTrace("after_ensure_plantillas_nuevas_tipo_licencias")
		if err := dbpkg.DisableRobotRadioInTipoEmpresaPreconfiguraciones(dbSuper); err != nil {
			log.Printf("warning: no se pudieron apagar robot/emisora en preconfiguraciones: %v", err)
		}
		startupTrace("after_preconfig_robot_radio_defaults")
		if err := dbpkg.EnsureEnergiaSolarInTipoEmpresaPreconfiguraciones(dbSuper); err != nil {
			log.Printf("warning: no se pudo agregar energia solar a preconfiguraciones: %v", err)
		}
		startupTrace("after_preconfig_energia_solar")
		if err := dbpkg.DropTiposDeUsuarioTable(dbSuper); err != nil {
			log.Printf("warning: no se pudo eliminar tabla legada tipos_de_usuario: %v", err)
		}
		log.Println("INFO: runtime DB dialect=postgres (VPS)")
	} else {
		log.Fatalf("Runtime DB no soportada: configure DB_DIALECT=postgres y DSN de PostgreSQL")
	}

	if err := dbpkg.EnsureEmpresaUsuariosAuthSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure users auth schema in empresas db: %v", err)
	}
	if _, err := dbpkg.MigrateEmpresaUsuarioTemporaryTokens(dbEmpresas, strings.EqualFold(strings.TrimSpace(os.Getenv("PCS_TOTP_MIGRATION_DRY_RUN")), "1") || strings.EqualFold(strings.TrimSpace(os.Getenv("PCS_TOTP_MIGRATION_DRY_RUN")), "true")); err != nil {
		log.Fatalf("failed to protect enterprise user temporary tokens: %v", err)
	}
	if err := dbpkg.EnsureEmpresaBuzonSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure empresa buzon schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_usuarios_auth_schema")
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure carritos schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_carritos_schema")
	if err := dbpkg.EnsureEmpresaDatafonosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure datafonos schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_datafonos_schema")
	if err := dbpkg.DisableLegacyFloatingRobotAndRadioPrefs(dbEmpresas); err != nil {
		log.Fatalf("failed to disable legacy floating robot/radio prefs: %v", err)
	}
	startupTrace("after_empresa_chat_robot_radio_defaults")
	if err := dbpkg.DisableFloatingChatVoicePrefs(dbEmpresas); err != nil {
		log.Fatalf("failed to disable floating chat voice prefs: %v", err)
	}
	startupTrace("after_empresa_chat_voice_defaults")
	if err := handlers.ApplyDefaultCarritoUIToExistingEmpresaPrefs(dbEmpresas); err != nil {
		log.Fatalf("failed to apply default cart UI to existing empresas: %v", err)
	}
	startupTrace("after_empresa_carrito_ui_defaults")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure finanzas schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_finanzas_schema")
	if err := dbpkg.EnsureEmpresaImpuestosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure impuestos schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_impuestos_schema")
	if err := dbpkg.EnsureEmpresaNominaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure nomina schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_nomina_schema")
	if result, err := dbpkg.ApplyColombiaDefaultsToExistingEmpresas(dbEmpresas); err != nil {
		log.Printf("warning: no se pudo aplicar preconfiguracion Colombia a empresas existentes: %v", err)
	} else if result != nil {
		log.Printf("INFO: preconfiguracion Colombia %s verificada: empresas=%d aplicadas=%d errores=%d", result.Version, result.Empresas, result.Aplicadas, len(result.Errores))
	}
	if err := dbpkg.SeedCatalogoLegalColombiaBase(dbEmpresas, "sistema.arranque"); err != nil {
		log.Printf("warning: no se pudo registrar catalogo legal Colombia: %v", err)
	}
	if result, err := dbpkg.ApplyParametrosLegalesToExistingEmpresas(dbEmpresas); err != nil {
		log.Printf("warning: no se pudo registrar parametros legales por empresa: %v", err)
	} else if result != nil {
		log.Printf("INFO: parametros legales Colombia %s registrados: empresas=%d aplicadas=%d errores=%d", result.Version, result.Empresas, result.Aplicadas, len(result.Errores))
	}
	startupTrace("after_empresa_colombia_defaults")
	if err := dbpkg.EnsureEmpresaCreditosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure creditos schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_creditos_schema")
	if err := dbpkg.EnsureEmpresaContabilidadColombiaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure contabilidad colombia schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_contabilidad_colombia_schema")
	if err := dbpkg.EnsureEmpresaContabilidadColombiaAvanzadaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure contabilidad colombia avanzada schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_contabilidad_colombia_avanzada_schema")
	if err := dbpkg.EnsureEmpresaCentrosCostoSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure centros costo schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_centros_costo_schema")
	if err := dbpkg.EnsureEmpresaCierreFiscalSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure cierre fiscal schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_cierre_fiscal_schema")
	if err := dbpkg.EnsureEmpresaDeclaracionesTributariasSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure declaraciones tributarias schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_declaraciones_tributarias_schema")
	if err := dbpkg.EnsureEmpresaTesoreriaPresupuestoSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure tesoreria presupuesto schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaImportacionesCosteoSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure importaciones costeo schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaAIUConstruccionSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure AIU construccion schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCobranzaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure cobranza schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaPortalContadorSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure portal contador schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaPortalTercerosCertificadosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure portal terceros certificados schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaSoportesComprasIASchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure soportes compras IA schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaModulosColombiaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure modulos empresariales colombia schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaComprasAvanzadasSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure compras avanzadas schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaReservasHotelSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure reservas hotel schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_reservas_schema")
	if err := dbpkg.EnsureEmpresaTarifasMotelSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure tarifas motel schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_tarifas_motel_schema")
	if err := dbpkg.EnsureEmpresaSensorPuertasSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure sensor puertas schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaControlElectricoSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure control electrico schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEnergiaSolarSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure energia solar schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCamarasSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure camaras schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaGrafologiaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure grafologia schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCarnetsSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure carnets empresa schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaParqueaderoSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure parqueadero empresa schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaApartamentosTuristicosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure apartamentos turisticos empresa schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaPropiedadHorizontalSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure propiedad horizontal empresa schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaProduccionMRPSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure produccion mrp empresa schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaWMSSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure logistica WMS empresa schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureHotelTarjetasAccesoSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure hotel tarjetas acceso schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure empresa productos schema in empresas db: %v", err)
	}
	if result, err := dbpkg.ApplyDefaultBodega1ToExistingEmpresas(dbEmpresas); err != nil {
		log.Printf("warning: no se pudo crear Bodega 1 por defecto en empresas existentes: %v", err)
	} else if result != nil {
		log.Printf("INFO: Bodega 1 por defecto verificada: empresas=%d aplicadas=%d errores=%d", result.Empresas, result.Aplicadas, len(result.Errores))
	}
	if err := dbpkg.EnsureEmpresaInventarioAvanzadoSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure inventario avanzado schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCRMVentasAvanzadasSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure crm ventas avanzadas schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaSoporteRemotoSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure soporte remoto schema in empresas db: %v", err)
	}
	if created, updated, err := dbpkg.SeedEmpresaSoporteRemotoDefaults(dbEmpresas, dbSuper); err != nil {
		log.Printf("warning: no se pudieron preparar configuraciones RustDesk por empresa: %v", err)
	} else {
		log.Printf("INFO: soporte remoto RustDesk preparado: configuraciones creadas=%d actualizadas=%d", created, updated)
	}
	startupTrace("after_empresa_sensor_puertas_schema")
	if runtimePostgres {
		if err := handlers.EnsureSensitiveSuperConfigEncrypted(dbSuper); err != nil {
			log.Fatalf("failed to enforce sensitive config encryption in super db: %v", err)
		}
		startupTrace("after_sensitive_super_config")
		if err := dbpkg.EnsurePostgresPrimaryKeySequences(dbEmpresas); err != nil {
			log.Fatalf("failed to ensure postgres primary key sequences in empresas db: %v", err)
		}
		startupTrace("after_empresas_pk_sequences")
		if err := dbpkg.EnsurePostgresPrimaryKeySequences(dbSuper); err != nil {
			log.Fatalf("failed to ensure postgres primary key sequences in super db: %v", err)
		}
		startupTrace("after_super_pk_sequences")
		loadGoogleOAuthFromDB(dbSuper)
		startupTrace("after_load_google_oauth_from_db")
		if err := handlers.EnsureSuperContextoIALogicaNegocio(dbSuper); err != nil {
			log.Printf("warning: no se pudo registrar contexto IA de logica de negocio: %v", err)
		}
		startupTrace("after_super_contexto_ia")
		if clientID == "" || clientSecret == "" {
			log.Println("Warning: GOOGLE_CLIENT_ID o GOOGLE_CLIENT_SECRET no configurados (entorno/DB)")
		}
		log.Println("INFO: modo PostgreSQL activo; bootstrap legacy desactivado.")
	}
	utils.ConfigureErrorMonitor(dbSuper, backendDir)
	startupTrace("after_error_monitor")

	// Inicializar tabla de métricas y arrancar collector periódico
	if err := dbpkg.InitMetricsTable(dbSuper); err != nil {
		log.Printf("warning: failed to init metrics table: %v", err)
		utils.ReportProcessError("metrics.collector", "metrics_schema_init", "No se pudo inicializar la tabla de metricas", err, utils.ErrorLevelError, nil)
	}
	metricsInterval := metrics.DefaultIntervalSeconds()
	stopMetrics := make(chan struct{})
	go utils.RunProtectedProcess("metrics.collector", map[string]interface{}{"interval_seconds": metricsInterval}, func() {
		metrics.StartCollector(dbSuper, metricsInterval, stopMetrics)
	})

	stopSuperAlertas := make(chan struct{})
	go utils.RunProtectedProcess("super.alertas_worker", map[string]interface{}{"interval_minutes": 1}, func() {
		handlers.StartSuperAlertasWorker(dbSuper, time.Minute, stopSuperAlertas)
	})

	stopAuditRetention := make(chan struct{})
	go utils.RunProtectedProcess("auditoria.retention_worker", map[string]interface{}{"interval_hours": 12}, func() {
		dbpkg.StartEmpresaAuditoriaRetentionWorker(dbEmpresas, 12*time.Hour, stopAuditRetention)
	})

	stopLicenciasEstado := make(chan struct{})
	go utils.RunProtectedProcess("licencias.estado_empresas_worker", map[string]interface{}{"interval_hours": 1}, func() {
		dbpkg.StartLicenciaEmpresaEstadoWorker(dbEmpresas, dbSuper, time.Hour, stopLicenciasEstado)
	})

	stopLicenciasVencimiento := make(chan struct{})
	go utils.RunProtectedProcess("licencias.vencimiento_alertas_worker", map[string]interface{}{"interval_hours": 12}, func() {
		handlers.StartLicenciaVencimientoAlertasWorker(dbSuper, dbEmpresas, 12*time.Hour, stopLicenciasVencimiento)
	})

	stopVPSSnapshotWorker := make(chan struct{})
	go utils.RunProtectedProcess("super.vps_snapshot_worker", map[string]interface{}{"interval_hours": 1}, func() {
		handlers.StartSuperVPSSnapshotWorker(dbSuper, time.Hour, stopVPSSnapshotWorker)
	})

	stopMantenimientoAgentesWorker := make(chan struct{})
	go utils.RunProtectedProcess("super.mantenimiento_agentes_worker", map[string]interface{}{"interval_minutes": 1}, func() {
		handlers.StartSuperMantenimientoAgentesWorker(dbSuper, time.Minute, stopMantenimientoAgentesWorker)
	})

	stopParametrosLegales := make(chan struct{})
	go utils.RunProtectedProcess("parametros_legales.worker", map[string]interface{}{"interval_hours": 24}, func() {
		dbpkg.StartEmpresaParametrosLegalesWorker(dbEmpresas, 24*time.Hour, stopParametrosLegales)
	})

	stopCobranzaRecordatorios := make(chan struct{})
	go utils.RunProtectedProcess("cobranza.recordatorios_worker", map[string]interface{}{"interval_hours": 1}, func() {
		handlers.StartEmpresaCobranzaRecordatoriosWorker(dbEmpresas, dbSuper, time.Hour, stopCobranzaRecordatorios)
	})

	asientosInterval, asientosBatchSize, asientosMaxRetries := resolveAsientosWorkerPolicy()
	log.Printf("[asientos_worker] policy interval=%s batch=%d max_reintentos=%d", asientosInterval, asientosBatchSize, asientosMaxRetries)
	stopAsientosWorker := make(chan struct{})
	go utils.RunProtectedProcess("finanzas.asientos_worker", map[string]interface{}{"interval": asientosInterval.String(), "batch_size": asientosBatchSize, "max_retries": asientosMaxRetries}, func() {
		dbpkg.StartEmpresaAsientosContablesWorker(dbEmpresas, asientosInterval, asientosBatchSize, asientosMaxRetries, stopAsientosWorker)
	})

	stopControlElectricoWorker := make(chan struct{})
	go utils.RunProtectedProcess("control_electrico.programacion_worker", map[string]interface{}{"interval_minutes": 1}, func() {
		handlers.StartControlElectricoProgramacionWorker(dbEmpresas, time.Minute, stopControlElectricoWorker)
	})
	startupTrace("after_workers")

	// Determinar carpeta web una sola vez para rutas estaticas y handlers que listan recursos.
	webDir := resolveWebDir()
	downloadsDir := resolveDownloadsDir()
	startupTrace("after_resolve_dirs")
	vpsSecurityService, err := vpssecurity.NewService(nil, nil, nil)
	if err != nil {
		log.Fatalf("failed to initialize VPS security service: %v", err)
	}
	startupTrace("after_vps_security_service")

	http.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin(clientID, redirectURL))
	http.HandleFunc("/auth/google/usuario/login", handlers.HandleGoogleUsuarioLogin(clientID, redirectURL))
	// Pasar la conexión de la base `empresas` al callback para persistir usuarios y empresas
	// Pasar tanto la conexión de empresas como la de superadministrador al callback
	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback(dbEmpresas, dbSuper, clientID, clientSecret, redirectURL))

	// Endpoint que expone configuración pública simple en JS.
	http.HandleFunc("/config.js", handlers.PublicConfigJSHandler(dbSuper))
	// Endpoint para procesar la aceptación del contrato desde la página /accept.html
	http.HandleFunc("/accept/complete", handlers.AcceptCompleteHandler(dbSuper))

	// Endpoints para administración y auditoría (listar administradores y sesiones)
	http.HandleFunc("/super/administradores", handlers.ListAdministradoresHandler(dbSuper))
	http.HandleFunc("/super/sesiones", handlers.ListSesionesHandler(dbSuper))
	http.HandleFunc("/api/user/configuracion", handlers.UserConfiguracionHandler(dbSuper))

	// Endpoints CRUD para tipos de empresas
	http.HandleFunc("/super/api/tipos_empresas", handlers.WithSuperAuditoria(dbSuper, "tipos_empresas", handlers.TiposEmpresasHandler(dbSuper)))
	http.HandleFunc("/super/api/tipos_empresas/preconfiguracion", handlers.SuperTipoEmpresaPreconfiguracionHandler(dbSuper))
	http.HandleFunc("/super/api/servidores", handlers.WithSuperAuditoria(dbSuper, "super_servidores", handlers.SuperServidoresListHandler(dbSuper)))
	http.HandleFunc("/super/api/servidores/toggle", handlers.WithSuperAuditoria(dbSuper, "super_servidores", handlers.SuperServidoresToggleHandler(dbSuper)))
	http.HandleFunc("/super/api/servidores/probar", handlers.WithSuperAuditoria(dbSuper, "super_servidores", handlers.SuperServidoresProbeHandler(dbSuper)))
	http.HandleFunc("/super/api/vps2", handlers.WithSuperAuditoria(dbSuper, "super_vps2", handlers.SuperVPS2Handler(dbSuper)))
	http.HandleFunc("/super/api/vps/procesos", handlers.SuperVPSProcessesHandler(dbSuper))
	http.HandleFunc("/super/api/plantillas_nuevas/catalogo", handlers.SuperPlantillasNuevosCatalogoHandler(dbSuper))
	http.HandleFunc("/super/api/plantillas_integracion/catalogo", handlers.SuperPlantillasIntegracionCatalogoHandler(dbSuper))
	http.HandleFunc("/super/api/roles_de_usuario", handlers.RolesDeUsuarioHandler(dbSuper))
	http.HandleFunc("/super/api/roles_de_usuario/permisos", handlers.RolesDeUsuarioPermisosHandler(dbSuper))
	// Endpoint CRUD para empresas (persistidas en pcs_empresas PostgreSQL)
	http.HandleFunc("/super/api/empresas", handlers.WithSuperAuditoria(dbSuper, "selector_empresas", handlers.EmpresasHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/super/api/empresas_estado", handlers.WithSuperAuditoria(dbSuper, "super_empresas_estado", handlers.SuperEmpresasEstadoHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/super/api/empresas/compartidos", handlers.WithSuperAuditoria(dbSuper, "empresas_compartidas", handlers.EmpresaCompartidaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/super/api/empresas/compartidos/aceptar", handlers.WithSuperAuditoria(dbSuper, "empresas_compartidas", handlers.EmpresaCompartidaAcceptHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/super/api/email_corporativo", handlers.WithSuperAuditoria(dbSuper, "super_email_corporativo", handlers.SuperEmailCorporativoHandler(dbSuper, dbEmpresas)))
	http.HandleFunc("/api/internal/email_corporativo/autologin", handlers.EmpresaEmailCorporativoAutologinHandler(dbSuper))
	// Endpoints para asesores comerciales y comisiones de licencias.
	http.HandleFunc("/super/api/asesor_comercial", handlers.AsesorComercialSuperHandler(dbSuper))
	http.HandleFunc("/api/asesor_comercial/aceptar", handlers.AsesorComercialAcceptHandler(dbSuper))
	http.HandleFunc("/api/asesor_comercial/mis_clientes", handlers.AsesorComercialMisClientesHandler(dbSuper))
	http.HandleFunc("/super/api/soporte_remoto", handlers.WithSuperAuditoria(dbSuper, "super_soporte_remoto", handlers.SuperSoporteRemotoHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/super/api/tickets_ayuda", handlers.SuperAyudaTicketsHandler(dbSuper))
	http.HandleFunc("/super/api/correos_masivos", handlers.SuperCorreosMasivosHandler(dbEmpresas, dbSuper))
	startupTrace("after_super_and_core_routes")
	// Módulo de productos por empresa (persistido en pcs_empresas PostgreSQL)
	http.HandleFunc("/api/empresa/bodegas", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaBodegasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/categorias_productos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaCategoriasProductosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/productos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaProductosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/recetas_productos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaRecetasProductosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/productos/imagen", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaProductoImagenUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/existencias", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioExistenciasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/configuracion", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioConfiguracionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/alertas", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioAlertasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/conteo_ciclico", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioConteoCiclicoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/resumen", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioResumenHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/tendencia", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioTendenciaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/balance_bodegas", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioBalanceBodegasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/proyeccion_quiebre", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioProyeccionQuiebreHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/plan_reposicion", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioPlanReposicionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/plan_reposicion_resumen", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioPlanReposicionResumenHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/plan_reposicion_borrador", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioPlanReposicionBorradorHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/movimientos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioMovimientosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/transferir", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioTransferHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/ajustar", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioAjusteHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/cambiar_producto", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioCambioProductoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario_avanzado", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioAvanzadoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/productos/precios_historial", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaProductoPrecioHistorialHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras/plan_reposicion/emitir_orden", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasPlanReposicionEmitirOrdenHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras/plan_reposicion/actualizar_estado", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasPlanReposicionActualizarEstadoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras/documentos", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasDocumentosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras/documentos/comprobante", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasDocumentoComprobanteUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras_avanzadas", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasAvanzadasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/soportes_compras_ia", handlers.WithEmpresaSoportesComprasIAPermissions(dbEmpresas, dbSuper, handlers.EmpresaSoportesComprasIAHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/gestion_documental", handlers.WithEmpresaGestionDocumentalPermissions(dbEmpresas, dbSuper, handlers.EmpresaModuloColombiaHandler(dbEmpresas, "gestion_documental")))
	http.HandleFunc("/api/empresa/contratos_obligaciones", handlers.WithEmpresaContratosObligacionesPermissions(dbEmpresas, dbSuper, handlers.EmpresaModuloColombiaHandler(dbEmpresas, "contratos_obligaciones")))
	http.HandleFunc("/api/empresa/tickets_ayuda", handlers.WithEmpresaSelfServicePermissions(dbEmpresas, dbSuper, handlers.EmpresaAyudaTicketsHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/buzon", handlers.WithEmpresaSelfServicePermissions(dbEmpresas, dbSuper, handlers.EmpresaBuzonHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/noticias", handlers.WithEmpresaSelfServicePermissions(dbEmpresas, dbSuper, handlers.EmpresaNoticiasPortalHandler(dbSuper)))
	http.HandleFunc("/api/empresa/drogueria_farmacia", handlers.WithEmpresaDrogueriaFarmaciaPermissions(dbEmpresas, dbSuper, handlers.EmpresaModuloColombiaHandler(dbEmpresas, "drogueria_farmacia")))
	http.HandleFunc("/api/empresa/proveedores", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaProveedoresHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/importaciones_costeo", handlers.WithEmpresaImportacionesCosteoPermissions(dbEmpresas, dbSuper, handlers.EmpresaImportacionesCosteoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/aiu_construccion", handlers.WithEmpresaAIUConstruccionPermissions(dbEmpresas, dbSuper, handlers.EmpresaAIUConstruccionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/produccion_mrp", handlers.WithEmpresaProduccionMRPPermissions(dbEmpresas, dbSuper, handlers.EmpresaProduccionMRPHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/logistica_wms", handlers.WithEmpresaWMSPermissions(dbEmpresas, dbSuper, handlers.EmpresaWMSHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/servicios", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaServiciosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/usuarios/login", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioLoginHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/establecer_password", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioSetPasswordHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/recuperar_invitacion", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioRequestInvitationRecoveryHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/solicitar_recuperacion_password", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioRequestPasswordRecoveryHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/restablecer_password", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioResetPasswordHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/cambiar_password", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioChangePasswordHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaUsuariosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/horarios_trabajadores", handlers.WithEmpresaHorariosTrabajadoresPermissions(dbEmpresas, dbSuper, handlers.EmpresaHorariosTrabajadoresHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/mi_horario", handlers.WithEmpresaSelfServicePermissions(dbEmpresas, dbSuper, handlers.EmpresaMiHorarioUsuarioHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/asistencia_empleados", handlers.WithEmpresaAsistenciaEmpleadosPermissions(dbEmpresas, dbSuper, handlers.EmpresaAsistenciaEmpleadosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/nomina", handlers.WithEmpresaNominaSueldosPermissions(dbEmpresas, dbSuper, handlers.EmpresaNominaSueldosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/nomina/agente_internet", handlers.WithEmpresaNominaSueldosPermissions(dbEmpresas, dbSuper, handlers.EmpresaAgenteInternetNominaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/vehiculos_registro", handlers.WithEmpresaVehiculosRegistroPermissions(dbEmpresas, dbSuper, handlers.EmpresaVehiculosRegistroHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/carnets", handlers.WithEmpresaCarnetsPermissions(dbEmpresas, dbSuper, handlers.EmpresaCarnetsHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/gimnasio", handlers.WithEmpresaGimnasioPermissions(dbEmpresas, dbSuper, handlers.EmpresaGimnasioHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/taxi_system", handlers.WithEmpresaTaxiSystemPermissions(dbEmpresas, dbSuper, handlers.EmpresaTaxiSystemHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/domicilios", handlers.WithEmpresaDomiciliosPermissions(dbEmpresas, dbSuper, handlers.EmpresaDomiciliosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/parqueadero", handlers.WithEmpresaParqueaderoPermissions(dbEmpresas, dbSuper, handlers.EmpresaParqueaderoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/apartamentos_turisticos", handlers.WithEmpresaApartamentosTuristicosPermissions(dbEmpresas, dbSuper, handlers.EmpresaApartamentosTuristicosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/propiedad_horizontal", handlers.WithEmpresaPropiedadHorizontalPermissions(dbEmpresas, dbSuper, handlers.EmpresaPropiedadHorizontalHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/alquileres", handlers.WithEmpresaAlquileresPermissions(dbEmpresas, dbSuper, handlers.EmpresaAlquileresHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/odontologia", handlers.WithEmpresaOdontologiaPermissions(dbEmpresas, dbSuper, handlers.EmpresaOdontologiaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/turnos_atencion", handlers.WithEmpresaTurnosAtencionPermissions(dbEmpresas, dbSuper, handlers.EmpresaTurnosAtencionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/publicaciones", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaPublicacionesRedSocialHandler(dbEmpresas))) // Protegido
	http.HandleFunc("/api/empresa/publicaciones/", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaPublicacionesRedSocialHandler(dbEmpresas)))
	http.HandleFunc("/api/public/publicaciones", handlers.PublicacionesRedSocialHandler(dbEmpresas)) // Publico
	http.HandleFunc("/api/public/market_symbol", handlers.PublicMarketSymbolHandler())
	http.HandleFunc("/api/public/publicaciones/", handlers.PublicRedSocialInteraccionesHandler(dbEmpresas))
	http.HandleFunc("/api/empresa/clientes", handlers.WithEmpresaClientesPermissions(dbEmpresas, dbSuper, handlers.EmpresaClientesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/crm_avanzado", handlers.WithEmpresaCRMUnificadoPermissions(dbEmpresas, dbSuper, handlers.EmpresaCRMVentasAvanzadasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/carritos_compra", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCarritosCompraHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/carritos_compra/items", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCarritoItemsHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/carritos_compra/historial_productos", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCarritoProductoHistorialHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/offline_ventas", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaOfflineVentasHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/datafonos", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaDatafonosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/venta_publica", handlers.WithEmpresaVentaPublicaPermissions(dbEmpresas, dbSuper, handlers.EmpresaVentaPublicaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/rappi", handlers.WithEmpresaVentaPublicaPermissions(dbEmpresas, dbSuper, handlers.EmpresaRappiHandler(dbEmpresas)))
	http.HandleFunc("/api/public/venta_publica", handlers.PublicVentaPublicaHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/public/rappi/webhook", handlers.PublicRappiWebhookHandler(dbEmpresas))
	http.HandleFunc("/api/public/turnos_atencion", handlers.PublicTurnosAtencionHandler(dbEmpresas))
	http.HandleFunc("/api/public/taxi_system", handlers.PublicTaxiSystemHandler(dbEmpresas))
	http.HandleFunc("/api/public/domicilios", handlers.PublicDomiciliosHandler(dbEmpresas))
	http.HandleFunc("/api/public/parqueadero", handlers.PublicParqueaderoHandler(dbEmpresas))
	http.HandleFunc("/api/public/estacion_vip", handlers.PublicEstacionVIPHandler(dbEmpresas))
	http.HandleFunc("/api/public/chat_portal", handlers.PublicPortalCompanyChatHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/public/chat_portal_stream", handlers.PublicPortalCompanyChatStreamHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/public/mensajes_privados", handlers.PublicMensajesPrivadosHandler(dbEmpresas))
	http.HandleFunc("/api/public/soporte_remoto", handlers.PublicEmpresaSoporteRemotoAgentHandler(dbEmpresas))
	http.HandleFunc("/api/public/venta_digital", handlers.PublicVentaDigitalHandler(dbSuper))
	http.HandleFunc("/api/public/pagina_principal", handlers.PublicPaginaPrincipalHandler(dbSuper))
	http.HandleFunc("/api/public/informacion_de_modulos", handlers.PublicInformacionModulosHandler(dbSuper))
	http.HandleFunc("/api/public/noticias", handlers.PublicNoticiasPortalHandler(dbSuper))
	http.HandleFunc("/api/public/portal_visitas", handlers.PublicPortalVisitasHandler(dbSuper))
	http.HandleFunc("/api/public/plantillas_nuevas/catalogo", handlers.PublicPlantillasNuevosCatalogoHandler())
	http.HandleFunc("/api/public/plantillas_integracion/catalogo", handlers.PublicPlantillasIntegracionCatalogoHandler())
	http.HandleFunc("/api/public/contrato", handlers.PublicContratoHandler(dbSuper))
	http.HandleFunc("/api/public/geo", handlers.PublicGeoHandler())
	http.HandleFunc("/api/empresa/reservas_hotel", handlers.WithEmpresaReservasHotelPermissions(dbEmpresas, dbSuper, handlers.EmpresaReservasHotelHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/tarifas_por_minutos", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaTarifasPorMinutosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/tarifas_por_dia", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaTarifasPorDiaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/tarifas_motel", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaTarifasMotelHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/hotel_tarjetas_acceso", handlers.WithEmpresaReservasHotelPermissions(dbEmpresas, dbSuper, handlers.EmpresaHotelTarjetasAccesoHandler(dbEmpresas)))
	http.HandleFunc("/api/public/hotel_tarjetas_acceso", handlers.PublicHotelTarjetasAccesoHandler(dbEmpresas))
	http.HandleFunc("/api/empresa/codigos_de_descuento", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCodigosDescuentoHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/propinas", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaPropinasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/comisiones", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComisionesServicioHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_general", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionGeneralHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_operativa", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionOperativaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_avanzada", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionAvanzadaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_avanzada/logo", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionAvanzadaLogoUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_guiada", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionGuiadaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/panel_configuracion", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaPanelConfiguracionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/licencia_sistema/pdf", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaLicenciaSistemaPDFHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/email_corporativo", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaEmailCorporativoHandler(dbSuper, dbEmpresas)))
	http.HandleFunc("/api/empresa/db_admin", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaDBAdminHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/impresoras", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaImpresorasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/impresoras/agente", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaImpresorasAgenteHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/impresoras/resolver", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaImpresorasResolverHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/estacion_prefs", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaEstacionPrefsHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/estacion_aseo", handlers.WithEmpresaSelfServicePermissions(dbEmpresas, dbSuper, handlers.EmpresaEstacionAseoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/facturacion_electronica/ecuador", handlers.WithEmpresaFacturacionEcuadorPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaEcuadorHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica/panama", handlers.WithEmpresaFacturacionPanamaPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaPanamaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica/pais_detectado", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaPaisDetectadoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica/paises_disponibles", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaPaisesDisponiblesHandler()))
	http.HandleFunc("/api/empresa/impuestos", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaImpuestosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/impuestos/agente_internet", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaAgenteInternetImpuestosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/energia_solar", handlers.WithEmpresaEnergiaSolarPermissions(dbEmpresas, dbSuper, handlers.EmpresaEnergiaSolarHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/camaras", handlers.WithEmpresaCamarasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCamarasHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/grafologia", handlers.WithEmpresaGrafologiaPermissions(dbEmpresas, dbSuper, handlers.EmpresaGrafologiaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/bolsa", handlers.WithEmpresaBolsaPermissions(dbEmpresas, dbSuper, handlers.EmpresaBolsaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/ia_empresarial", handlers.WithEmpresaReportesPermissions(dbEmpresas, dbSuper, handlers.EmpresaIAEmpresarialHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/chat_tareas/conversaciones", handlers.WithEmpresaChatTareasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasConversacionesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/participantes", handlers.WithEmpresaChatTareasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasParticipantesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/mensajes", handlers.WithEmpresaChatTareasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasMensajesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/mensajes/adjunto", handlers.WithEmpresaChatTareasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasAdjuntoUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/tareas", handlers.WithEmpresaChatTareasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasTareasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/citas", handlers.WithEmpresaChatTareasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasCitasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/tareas/nota_voz", handlers.WithEmpresaChatTareasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasTareaNotaVozUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/papelera", handlers.WithEmpresaChatTareasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasPapeleraHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/ubicacion_gps/dispositivos", handlers.WithEmpresaUbicacionGPSPermissions(dbEmpresas, dbSuper, handlers.EmpresaUbicacionGPSDispositivosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/ubicacion_gps/recorridos", handlers.WithEmpresaUbicacionGPSPermissions(dbEmpresas, dbSuper, handlers.EmpresaUbicacionGPSRecorridosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/hoja_vida_operativa", handlers.WithEmpresaHojaVidaOperativaPermissions(dbEmpresas, dbSuper, handlers.EmpresaHojaVidaOperativaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/movimientos", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasMovimientosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/movimientos/comprobante", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasMovimientoComprobanteUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/corte_caja", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCorteCajaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/corte_caja/configuracion", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCorteCajaConfiguracionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/configuracion", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasConfiguracionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/periodos", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasPeriodosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/breb_qr", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasBrebQRHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/renta_ia", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasRentaIAHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/finanzas/asientos_contables", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasAsientosContablesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/cierres_caja", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasCierresCajaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/contabilidad_colombia", handlers.WithEmpresaContabilidadColombiaPermissions(dbEmpresas, dbSuper, handlers.EmpresaContabilidadColombiaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/contabilidad_colombia_avanzada", handlers.WithEmpresaContabilidadColombiaAvanzadaPermissions(dbEmpresas, dbSuper, handlers.EmpresaContabilidadColombiaAvanzadaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/activos_fijos_niif_fiscal", handlers.WithEmpresaActivosFijosNIIFPermissions(dbEmpresas, dbSuper, handlers.EmpresaActivosFijosNIIFiscalHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/centros_costo", handlers.WithEmpresaCentrosCostoPermissions(dbEmpresas, dbSuper, handlers.EmpresaCentrosCostoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/cierre_fiscal", handlers.WithEmpresaCierreFiscalPermissions(dbEmpresas, dbSuper, handlers.EmpresaCierreFiscalHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/declaraciones_tributarias", handlers.WithEmpresaDeclaracionesTributariasPermissions(dbEmpresas, dbSuper, handlers.EmpresaDeclaracionesTributariasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/tesoreria_presupuesto", handlers.WithEmpresaTesoreriaPresupuestoPermissions(dbEmpresas, dbSuper, handlers.EmpresaTesoreriaPresupuestoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/plantillas_nuevas/catalogo", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaPlantillasNuevosCatalogoHandler()))
	http.HandleFunc("/api/empresa/plantillas_integracion/catalogo", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaPlantillasIntegracionCatalogoHandler()))
	http.HandleFunc("/api/empresa/bancos_pagos", handlers.WithEmpresaBancosPagosPermissions(dbEmpresas, dbSuper, handlers.EmpresaModuloColombiaHandler(dbEmpresas, "bancos_pagos")))
	http.HandleFunc("/api/empresa/cumplimiento_kyc", handlers.WithEmpresaCumplimientoKYCPermissions(dbEmpresas, dbSuper, handlers.EmpresaModuloColombiaHandler(dbEmpresas, "cumplimiento_kyc")))
	http.HandleFunc("/api/empresa/calidad_procesos", handlers.WithEmpresaCalidadProcesosPermissions(dbEmpresas, dbSuper, handlers.EmpresaModuloColombiaHandler(dbEmpresas, "calidad_procesos")))
	for _, moduloVertical := range handlers.NuevasPlantillasEmpresaModules() {
		moduloVertical := moduloVertical
		http.HandleFunc("/api/empresa/"+moduloVertical, handlers.WithEmpresaModuloVerticalPermissions(dbEmpresas, dbSuper, moduloVertical, handlers.EmpresaModuloColombiaHandler(dbEmpresas, moduloVertical)))
	}
	http.HandleFunc("/api/empresa/calculadora", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCalculadoraHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/creditos", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCreditosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/cobranza", handlers.WithEmpresaCobranzaPermissions(dbEmpresas, dbSuper, handlers.EmpresaCobranzaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/portal_contador", handlers.WithEmpresaPortalContadorPermissions(dbEmpresas, dbSuper, handlers.EmpresaPortalContadorHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/portal_terceros_certificados", handlers.WithEmpresaPortalTercerosPermissions(dbEmpresas, dbSuper, handlers.EmpresaPortalTercerosCertificadosHandler(dbEmpresas)))
	http.HandleFunc("/api/public/certificados_tributarios", handlers.PublicCertificadosTributariosHandler(dbEmpresas))
	http.HandleFunc("/api/empresa/backups", handlers.WithEmpresaBackupsPermissions(dbEmpresas, dbSuper, handlers.EmpresaBackupsHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/documentos", handlers.WithEmpresaDocumentosOnlyOfficePermissions(dbEmpresas, dbSuper, handlers.OnlyOfficeDocumentosHandler(dbSuper)))
	startupTrace("after_empresa_routes")

	// OnlyOffice public endpoints (token temporal)
	http.HandleFunc("/api/onlyoffice/file", handlers.OnlyOfficeFilePublicHandler(dbSuper))
	http.HandleFunc("/api/onlyoffice/callback", handlers.OnlyOfficeCallbackPublicHandler(dbSuper))
	http.HandleFunc("/api/empresa/soporte_remoto", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaSoporteRemotoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/reportes", handlers.WithEmpresaReportesPermissions(dbEmpresas, dbSuper, handlers.EmpresaReportesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/reportes_ia_chat", handlers.WithEmpresaReportesPermissions(dbEmpresas, dbSuper, handlers.EmpresaReportesIAChatHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/auditoria/eventos", handlers.WithEmpresaAuditoriaPermissions(dbEmpresas, dbSuper, handlers.EmpresaAuditoriaEventosHandler(dbEmpresas)))
	// Endpoint empresa: verificar acceso a la página frecuencia_fe.html (alias legacy frecuencia_fp.html)
	http.HandleFunc("/api/empresa/frecuencia_fp/permitido", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaFrecuenciaFPAllowedHandler(dbSuper)))
	handlers.RegisterEmpresaChatIARoutes(dbEmpresas, dbSuper)
	handlers.RegisterEmpresaModulosFaltantesRoutes(dbEmpresas, dbSuper)
	// Rutas del módulo sensor de puertas: configuración protegida y endpoint público para heartbeats
	http.HandleFunc("/api/empresa/sensor_puertas", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaSensorConfigHandler(dbEmpresas)))
	http.HandleFunc("/api/public/sensor_puertas", handlers.PublicSensorPuertasHandler(dbEmpresas))
	http.HandleFunc("/api/public/webrtc/signaling", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.SoporteRemotoSignalingHandler()))
	http.HandleFunc("/api/empresa/sensor_puertas/messages", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaSensorMessagesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/control_electrico", handlers.WithEmpresaControlElectricoPermissions(dbEmpresas, dbSuper, handlers.EmpresaControlElectricoHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/roles_de_usuario", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaRolesDeUsuarioHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/permisos_contexto", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaPermisosContextoHandler(dbSuper)))
	http.HandleFunc("/api/empresa/permisos_empresa", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaPermisosFinosHandler(dbSuper)))
	// Endpoint para obtener admin actual desde la cookie de sesión
	http.HandleFunc("/me", handlers.MeHandler(dbSuper))
	// Endpoint para obtener perfil/cuenta enriquecida (admin + usuario de empresa)
	http.HandleFunc("/api/account", handlers.AccountHandler(dbEmpresas, dbSuper))
	// Endpoints para actualizar perfil y cambiar contraseña (usuario autenticado)
	http.HandleFunc("/api/account/update_profile", handlers.AccountUpdateProfileHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/account/change_password", handlers.AccountChangePasswordHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/account/set_google_password", handlers.AccountSetGooglePasswordHandler(dbEmpresas, dbSuper))
	// Endpoint CRUD para administradores (API)
	http.HandleFunc("/super/api/administradores", handlers.WithSuperAuditoria(dbSuper, "administradores", handlers.AdministradoresHandler(dbSuper)))
	http.HandleFunc("/super/api/auditoria", handlers.SuperAuditoriaHandler(dbEmpresas, dbSuper))
	// Endpoints adicionales para flujo de autenticación de administradores (registro, login, confirmación, recuperación)
	http.HandleFunc("/super/api/administradores/register", handlers.AdminRegisterHandler(dbSuper))
	http.HandleFunc("/super/api/administradores/login", handlers.AdminLoginHandler(dbSuper))
	http.HandleFunc("/super/api/administradores/2fa", handlers.AdminTwoFactorHandler(dbSuper))
	http.HandleFunc("/super/api/config/admin_2fa", handlers.WithSuperAuditoria(dbSuper, "super_config_admin_2fa", handlers.AdminTwoFactorGlobalConfigHandler(dbSuper)))
	http.HandleFunc("/super/api/config/admin_page_urls", handlers.WithSuperAuditoria(dbSuper, "super_config_admin_page_urls", handlers.AdminPageURLsGlobalConfigHandler(dbSuper)))
	http.HandleFunc("/auth/confirmar_admin", handlers.ConfirmarAdminHandler(dbSuper))
	http.HandleFunc("/super/api/administradores/solicitar_recuperacion", handlers.AdminRequestPasswordRecoveryHandler(dbSuper))
	http.HandleFunc("/super/api/administradores/restablecer_password", handlers.AdminResetPasswordHandler(dbSuper))
	// Endpoint CRUD para licencias (nuevo)
	http.HandleFunc("/super/api/licencias", handlers.WithSuperAuditoria(dbSuper, "licencias", handlers.LicenciasHandler(dbSuper)))
	http.HandleFunc("/super/api/licencias/ventas_resumen", handlers.WithSuperAuditoria(dbSuper, "super_licencias_ventas_resumen", handlers.SuperLicenciasVentasResumenHandler(dbSuper)))
	http.HandleFunc("/super/api/licencias/configuracion", handlers.WithSuperAuditoria(dbSuper, "super_config_licencias", handlers.SuperLicenciasConfiguracionHandler(dbSuper)))
	http.HandleFunc("/super/api/licencias/codigos_descuento", handlers.WithSuperAuditoria(dbSuper, "licencias_codigos_descuento", handlers.SuperLicenciasCodigosDescuentoHandler(dbSuper)))
	http.HandleFunc("/super/api/empresa_licencias_adicionales", handlers.EmpresaLicenciasAdicionalesHandler(dbSuper))
	http.HandleFunc("/super/api/licencias/vencimiento_alertas", handlers.WithSuperAuditoria(dbSuper, "super_config_alertas_licencia", handlers.SuperLicenciaVencimientoAlertasHandler(dbSuper, dbEmpresas)))
	// Endpoint super: lista de administradores autorizados (Frecuencia FE/FP)
	http.HandleFunc("/super/api/administradores_frecuencia_fe", handlers.SuperAdministradoresFrecuenciaFEHandler(dbSuper))
	// Endpoint publico para exponer metodos de pago activos del checkout de licencias
	http.HandleFunc("/api/public/licencias/payment_methods", handlers.PublicLicenciasPaymentMethodsHandler(dbSuper, dbEmpresas))
	// Endpoint publico para calcular total, descuentos y activacion sin pago del checkout de licencias
	http.HandleFunc("/api/public/licencias/checkout_summary", handlers.LicenciaCheckoutSummaryHandler(dbSuper))
	// Endpoint para gestionar credenciales de Wompi (GET/PUT)
	http.HandleFunc("/super/api/config/wompi", handlers.WithSuperAuditoria(dbSuper, "super_config_wompi", handlers.WompiConfigHandler(dbSuper)))
	// Endpoint para gestionar credenciales de Epayco (GET/PUT)
	http.HandleFunc("/super/api/config/epayco", handlers.WithSuperAuditoria(dbSuper, "super_config_epayco", handlers.EpaycoConfigHandler(dbSuper)))
	// Endpoint para gestionar SMTP Gmail (GET/PUT)
	http.HandleFunc("/super/api/config/gmail", handlers.WithSuperAuditoria(dbSuper, "super_config_gmail", handlers.GmailConfigHandler(dbSuper)))
	http.HandleFunc("/super/api/config/whatsapp_notificaciones", handlers.WithSuperAuditoria(dbSuper, "super_config_whatsapp_notificaciones", handlers.SuperWhatsAppNotificationsHandler(dbSuper)))
	http.HandleFunc("/super/api/recordatorios_infraestructura", handlers.WithSuperAuditoria(dbSuper, "super_recordatorios_infraestructura", handlers.SuperRecordatoriosInfraestructuraHandler(dbSuper)))
	// Endpoint para activar o desactivar Google reCAPTCHA (GET/PUT)
	http.HandleFunc("/super/api/config/recaptcha", handlers.WithSuperAuditoria(dbSuper, "super_config_recaptcha", handlers.RecaptchaConfigHandler(dbSuper)))
	// Endpoint para administrar plantillas de correo del panel super
	http.HandleFunc("/super/api/config/email_templates", handlers.WithSuperAuditoria(dbSuper, "super_config_email_templates", handlers.SuperEmailTemplatesHandler(dbSuper)))
	// Endpoint super para administrar venta digital global
	http.HandleFunc("/super/api/venta_digital", handlers.SuperVentaDigitalHandler(dbSuper))
	// Endpoint para gestionar credenciales IA de modelos populares (GET/PUT)
	http.HandleFunc("/super/api/config/ai", handlers.WithSuperAuditoria(dbSuper, "super_config_ia_global", handlers.AIModelsConfigHandler(dbSuper)))
	// Endpoint para configurar limitaciones por empresa (RustDesk e IA)
	http.HandleFunc("/super/api/config/limitaciones_empresa", handlers.WithSuperAuditoria(dbSuper, "super_config_limitaciones", handlers.SuperEmpresaLimitacionesConfigHandler(dbSuper)))
	// Endpoint para configurar la lógica del chat con IA (empresas y super)
	http.HandleFunc("/super/api/config/chat_ia_logica", handlers.WithSuperAuditoria(dbSuper, "super_config_chat_ia", handlers.SuperChatIALogicaConfigHandler(dbEmpresas, dbSuper)))
	// Endpoint para configurar gestion RustDesk en el VPS (GET/PUT)
	http.HandleFunc("/super/api/config/rustdesk", handlers.WithSuperAuditoria(dbSuper, "super_config_rustdesk", handlers.RustDeskConfigHandler(dbSuper)))
	// Endpoint para configurar voz natural por streaming en VPS (GET/PUT)
	http.HandleFunc("/super/api/config/voice_stream", handlers.WithSuperAuditoria(dbSuper, "super_config_voz_ia", handlers.SuperVoiceStreamConfigHandler(dbSuper)))
	http.HandleFunc("/api/voice_stream/status", handlers.VoiceStreamStatusHandler(dbSuper))
	http.HandleFunc("/api/voice_stream/tts", handlers.VoiceStreamTTSProxyHandler(dbSuper))
	http.HandleFunc("/api/chat_flotante/preferencias", handlers.ChatFlotantePreferenciasHandler(dbSuper, dbEmpresas))
	// Endpoints para generar y descargar documentos dinamicos asistidos por IA.
	http.HandleFunc("/generate", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.DynamicDocumentGenerateHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/download", handlers.DynamicDocumentDownloadHandler(dbEmpresas, dbSuper))
	superAIChatController := handlers.NewSuperAIChatController(dbEmpresas, dbSuper)
	http.HandleFunc("/super/api/chat_con_ia_global/modelos", superAIChatController.ModelosHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/modelo_preferido", superAIChatController.ModeloPreferidoHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/consultar", superAIChatController.ConsultarHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/consultar_con_adjunto", superAIChatController.ConsultarConAdjuntoHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/consultar_stream", superAIChatController.ConsultarStreamHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/historial", superAIChatController.HistorialHandler)
	// Endpoint para respaldo/restauracion de configuracion critica del panel super
	http.HandleFunc("/super/api/config/backup", handlers.WithSuperAuditoria(dbSuper, "super_config_respaldo", handlers.SuperConfigBackupHandler(dbSuper)))
	// Endpoint para configuración de modo mantenimiento global
	http.HandleFunc("/super/api/config/mantenimiento", handlers.WithSuperAuditoria(dbSuper, "super_config_mantenimiento", handlers.SuperMantenimientoConfigHandler(dbSuper)))
	http.HandleFunc("/api/empresa/mantenimiento_programado", handlers.WithEmpresaSelfServicePermissions(dbEmpresas, dbSuper, handlers.EmpresaMantenimientoProgramadoHandler(dbSuper)))
	http.HandleFunc("/super/api/config/onlyoffice", handlers.WithSuperAuditoria(dbSuper, "super_config_onlyoffice", handlers.OnlyOfficeConfigHandler(dbSuper)))
	http.HandleFunc("/super/api/config/empresa_storage", handlers.WithSuperAuditoria(dbSuper, "super_config_empresa_storage", handlers.SuperEmpresaStorageConfigHandler(dbSuper, dbEmpresas)))
	// Endpoint super para administrar contrato versionado y su historial
	http.HandleFunc("/super/api/contrato", handlers.SuperContratoHandler(dbSuper))
	// Endpoint super para monitoreo centralizado de errores del sistema
	http.HandleFunc("/super/api/errores", handlers.SuperErroresSistemaHandler(dbSuper))
	// Endpoint super para consumos (OpenAI/Hostinger/Cursor) y contador de errores
	http.HandleFunc("/super/api/consumos", handlers.WithSuperAuditoria(dbSuper, "super_config_consumos", handlers.SuperConsumosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/super/api/panel_control/reset", handlers.WithSuperAuditoria(dbSuper, "super_panel_control_reset", handlers.SuperPanelControlResetHandler(dbSuper)))
	http.HandleFunc("/super/api/agentes_mantenimiento", handlers.WithSuperAuditoria(dbSuper, "super_agentes_mantenimiento", handlers.SuperMantenimientoAgentesHandler(dbSuper)))
	http.HandleFunc("/super/api/alertas_sistema", handlers.WithSuperAuditoria(dbSuper, "super_alertas_sistema", handlers.SuperAlertasSistemaHandler(dbSuper)))
	http.HandleFunc("/super/api/config/portal_chat_ia_info", handlers.WithSuperAuditoria(dbSuper, "super_config_chat_flotante", handlers.SuperPortalChatIAInfoHandler(dbSuper)))
	http.HandleFunc("/super/api/config/contexto_ia_logica_negocio", handlers.WithSuperAuditoria(dbSuper, "super_config_contexto_ia", handlers.SuperContextoIALogicaNegocioHandler(dbSuper)))
	// Endpoint super para administrar tarjetas dinamicas de la pagina principal (index)
	http.HandleFunc("/super/api/pagina_principal", handlers.WithSuperAuditoria(dbSuper, "super_pagina_principal", handlers.SuperPaginaPrincipalHandler(dbSuper, webDir)))
	http.HandleFunc("/super/api/informacion_de_modulos", handlers.WithSuperAuditoria(dbSuper, "super_informacion_modulos", handlers.SuperInformacionModulosHandler(dbSuper, webDir)))
	http.HandleFunc("/super/api/noticias", handlers.WithSuperAuditoria(dbSuper, "super_noticias", handlers.SuperNoticiasPortalHandler(dbSuper, webDir)))
	// Endpoints Wompi (Nequi): crear transacción y consultar estado
	http.HandleFunc("/wompi/terms", handlers.WompiTermsHandler(dbSuper))
	http.HandleFunc("/wompi/create_checkout", handlers.WompiCreateCheckoutHandler(dbSuper))
	http.HandleFunc("/wompi/create_transaction_nequi", handlers.WompiCreateNequiTransactionHandler(dbSuper))
	http.HandleFunc("/wompi/transaction_status", handlers.WompiTransactionStatusHandler(dbSuper))
	http.HandleFunc("/wompi/webhook", handlers.WompiWebhookHandler(dbSuper, dbEmpresas))
	// Endpoints Epayco: crear transacción y consultar estado
	http.HandleFunc("/epayco/create_transaction", handlers.EpaycoCreateTransactionHandler(dbSuper))
	http.HandleFunc("/epayco/transaction_status", handlers.EpaycoTransactionStatusHandler(dbSuper))
	http.HandleFunc("/epayco/webhook", handlers.EpaycoWebhookHandler(dbSuper, dbEmpresas))
	// Activación manual de licencia sin pago (uso interno de avance/prototipo)
	http.HandleFunc("/licencias/activar_sin_pago", handlers.ActivateLicenciaSinPagoHandler(dbSuper, dbEmpresas))
	// Confirmación de correo para usuarios de empresa.
	http.HandleFunc("/auth/confirmar_correo", handlers.ConfirmarCorreoUsuarioHandler(dbEmpresas))

	// Endpoints de métricas (actual y histórico)
	http.HandleFunc("/super/api/metrics/current", handlers.MetricsCurrentHandler(dbSuper))
	http.HandleFunc("/super/api/metrics/history", handlers.MetricsHistoryHandler(dbSuper))
	http.HandleFunc("/super/api/reportes_globales", handlers.WithSuperAuditoria(dbSuper, "reportes_globales", handlers.SuperReportesGlobalesHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/super/api/postgres/performance", handlers.WithSuperAuditoria(dbSuper, "super_postgresql", handlers.PostgresPerformanceHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/super/api/explorador_archivos", handlers.WithSuperAuditoria(dbSuper, "super_explorador_archivos", handlers.SuperFileExplorerHandler(dbSuper)))
	http.HandleFunc("/super/api/docker_portabilidad", handlers.WithSuperAuditoria(dbSuper, "super_docker_portabilidad", handlers.SuperDockerPortabilidadHandler(dbSuper)))
	http.HandleFunc("/super/api/vps_snapshots", handlers.WithSuperAuditoria(dbSuper, "super_vps_snapshots", handlers.SuperVPSSnapshotsHandler(dbSuper)))
	http.HandleFunc("/super/api/domotica_storage", handlers.WithSuperAuditoria(dbSuper, "super_domotica_storage", handlers.SuperDomoticaStorageHandler(dbSuper, dbEmpresas)))
	// Endpoint de seguridad: escaneo de puertos
	http.HandleFunc("/super/api/security/ports", handlers.WithSuperAuditoria(dbSuper, "super_seguridad_vps", handlers.SecurityPortsHandler(dbSuper)))
	// Endpoint de seguridad: listado de procesos en memoria RAM
	http.HandleFunc("/super/api/security/processes", handlers.WithSuperAuditoria(dbSuper, "super_seguridad_vps", handlers.SecurityProcessesHandler(dbSuper)))
	http.HandleFunc("/super/api/security/vps/config", handlers.WithSuperAuditoria(dbSuper, "super_seguridad_vps", handlers.SecurityVPSConfigHandler(dbSuper, vpsSecurityService)))
	http.HandleFunc("/super/api/security/vps/run", handlers.WithSuperAuditoria(dbSuper, "super_seguridad_vps", handlers.SecurityVPSRunHandler(dbSuper, vpsSecurityService)))
	http.HandleFunc("/super/api/security/vps/status", handlers.WithSuperAuditoria(dbSuper, "super_seguridad_vps", handlers.SecurityVPSStatusHandler(dbSuper, vpsSecurityService)))
	http.HandleFunc("/super/api/security/vps/history", handlers.WithSuperAuditoria(dbSuper, "super_seguridad_vps", handlers.SecurityVPSHistoryHandler(dbSuper, vpsSecurityService)))
	http.HandleFunc("/super/api/security/vps/report", handlers.WithSuperAuditoria(dbSuper, "super_seguridad_vps", handlers.SecurityVPSReportHandler(dbSuper, vpsSecurityService)))
	http.HandleFunc("/super/api/security/vps/compare", handlers.WithSuperAuditoria(dbSuper, "super_seguridad_vps", handlers.SecurityVPSCompareHandler(dbSuper, vpsSecurityService)))
	startupTrace("after_super_config_routes")

	// Logout handler: limpiar cookie de sesión (si existe) y redirigir a la página de login
	http.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		token := ""
		if c, err := r.Cookie("session_token"); err == nil {
			token = strings.TrimSpace(c.Value)
		}
		if token == "" {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				token = strings.TrimSpace(authHeader[len("Bearer "):])
			}
		}
		if token != "" {
			if err := dbpkg.RevokeSessionByToken(dbSuper, token); err != nil {
				log.Printf("warning: failed to revoke session token on logout: %v", err)
			}
			utils.InvalidateAuthCacheForToken(token)
		}

		// Invalidate common session cookie names
		cookies := []string{"session", "sid", "auth"}
		for _, name := range cookies {
			// set cookie expired
			http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1})
		}
		// also clear our session_token cookie with same attributes
		http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: handlers.SessionCookieSecure(r), SameSite: http.SameSiteLaxMode})
		handlers.SetBrowserSessionStateCookie(w, r, false)
		// Redirigir al login
		http.Redirect(w, r, "/login.html", http.StatusFound)
	})

	// Carpeta web determinada previamente para servir estaticos y handlers de recursos.

	// Servir assets centralizados (CSS, JS)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(webDir))))
	http.Handle("/descargas/", http.StripPrefix("/descargas/", http.FileServer(http.Dir(downloadsDir))))
	startupTrace("after_static_helper_routes")

	// Servir páginas estáticas desde la carpeta `web` detectada
	// Verificar existencia de index.html y loguear la ruta usada
	indexPath := filepath.Join(webDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		log.Printf("Warning: index.html no encontrado en %s\n", indexPath)
	} else if err != nil {
		log.Printf("Warning: error comprobando index.html en %s: %v\n", indexPath, err)
	} else {
		log.Printf("index.html encontrado en %s\n", indexPath)
	}
	faviconPath := filepath.Join(webDir, "favicon.ico")
	fallbackFaviconPath := filepath.Join(webDir, "img", "punto_venta.png")
	staticFS := noCacheAdminStaticHandler(buttonIconsStaticHandler(contextualHelpStaticHandler(http.FileServer(http.Dir(webDir)))))
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		if info, err := os.Stat(faviconPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, faviconPath)
			return
		}
		if info, err := os.Stat(fallbackFaviconPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, fallbackFaviconPath)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	http.HandleFunc("/pantalla", func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(backendDir, "..", "web", "pantalla_publica.html")
		http.ServeFile(w, r, path)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSpace(r.URL.Path)
		if (path == "/" || path == "") && handlers.IsEmpresaUsuarioLoginSubdomainRequest(r) {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/login_usuario.html"
			staticFS.ServeHTTP(w, r2)
			return
		}
		if path == "/descripcion_de_los_sistemas.ht" {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/descripcion_de_los_sistemas.html"
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			staticFS.ServeHTTP(w, r2)
			return
		}

		trimmed := strings.Trim(path, "/")
		parts := strings.Split(trimmed, "/")
		if len(parts) == 2 && strings.EqualFold(parts[1], "venta_publica.html") && strings.TrimSpace(parts[0]) != "" {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/venta_publica.html"
			staticFS.ServeHTTP(w, r2)
			return
		}
		if len(parts) == 2 && strings.EqualFold(parts[1], "visualizar_productos_y_precios_publico.html") && strings.TrimSpace(parts[0]) != "" {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/visualizar_productos_y_precios_publico.html"
			staticFS.ServeHTTP(w, r2)
			return
		}
		if len(parts) == 2 && strings.EqualFold(parts[1], "pagar_productos_de_venta_publica.html") && strings.TrimSpace(parts[0]) != "" {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/pagar_productos_de_venta_publica.html"
			staticFS.ServeHTTP(w, r2)
			return
		}
		if (path == "/" || path == "") && handlers.ResolveVentaPublicaSlugFromHost(r) != "" {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/venta_publica.html"
			staticFS.ServeHTTP(w, r2)
			return
		}
		if path == "/descargar_informacion_de_la_empresa" || path == "/descargar_informacion_de_la_empresa/" {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/descargar_informacion_de_la_empresa.html"
			staticFS.ServeHTTP(w, r2)
			return
		}
		staticFS.ServeHTTP(w, r)
	})
	startupTrace("after_root_handler")

	// Wrap DefaultServeMux with authentication, JSON error normalization and logging middleware
	maxRequestBodyBytes := getenvInt64Range("MAX_REQUEST_BODY_BYTES", 64<<20, 1<<20, 512<<20)
	handler := utils.LoggingMiddleware(utils.SecurityHeadersMiddleware(utils.CanonicalPublicHostMiddleware(utils.JSONErrorMiddleware(utils.RecoveryMiddleware(utils.RequestBodyLimitMiddleware(utils.AuthMiddleware(dbSuper, utils.CSRFMiddleware(http.DefaultServeMux)), maxRequestBodyBytes))))))
	startupTrace("after_handler_wrap")

	// Respetar la variable de entorno PORT si está definida; por defecto usar 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	markServerStopped, startupEventErr := handlers.RegisterServerStartupEvent(dbSuper, handlers.ServerStartupRegistration{
		BackendDir:  backendDir,
		ListenAddr:  addr,
		StartReason: strings.TrimSpace(os.Getenv("PCS_SERVER_START_REASON")),
	})
	if startupEventErr != nil {
		log.Printf("warning: no se pudo registrar evento de inicio de servidor: %v", startupEventErr)
		utils.ReportProcessError("server.runtime_notifications", "startup_event_registration", "No se pudo registrar el evento de inicio del servidor", startupEventErr, utils.ErrorLevelError, map[string]interface{}{"listen_addr": addr})
	}
	startupTrace("after_startup_event_registration")

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       getenvDurationRange("HTTP_READ_TIMEOUT", 30*time.Second, 5*time.Second, 5*time.Minute),
		WriteTimeout:      getenvDurationRange("HTTP_WRITE_TIMEOUT", 60*time.Second, 5*time.Second, 10*time.Minute),
		IdleTimeout:       getenvDurationRange("HTTP_IDLE_TIMEOUT", 120*time.Second, 15*time.Second, 15*time.Minute),
		MaxHeaderBytes:    1 << 20,
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signalCh)

	go utils.RunProtectedProcess("server.shutdown_signal", nil, func() {
		sig := <-signalCh
		reason := "signal_" + strings.ToLower(strings.TrimSpace(sig.String()))
		if markServerStopped != nil {
			markServerStopped(reason)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("warning: shutdown con error: %v", err)
			utils.ReportProcessError("server.shutdown", "shutdown_error", "Error durante el apagado controlado del servidor", err, utils.ErrorLevelError, map[string]interface{}{"reason": reason})
		}
	})
	startupTrace("before_listen")

	log.Println("Servidor arrancado en", addr)
	err = server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		if markServerStopped != nil {
			markServerStopped("listen_and_serve_error: " + err.Error())
		}
		utils.ReportProcessError("server.listen", "listen_and_serve_error", "El servidor HTTP termino con error en ListenAndServe", err, utils.ErrorLevelCritical, map[string]interface{}{"addr": addr})
		log.Fatal(err)
	}
	if markServerStopped != nil {
		markServerStopped("apagado_controlado")
	}
}
