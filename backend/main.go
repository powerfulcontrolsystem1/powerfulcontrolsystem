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
				log.Printf("warning: no se pudo descifrar la configuraciÃ³n %s: %v", key, derr)
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

	// Prioridad: variables de entorno > configuraciÃ³n en DB.
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
			path == "/menu.js" ||
			strings.HasSuffix(path, ".html") {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		next.ServeHTTP(w, r)
	})
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
		if err := dbpkg.EnsurePaymentGatewaySchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure payment gateway schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_payment_gateway_schema")
		if err := dbpkg.EnsureLicenciasSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure licencias schema in superadministrador db: %v", err)
		}
		startupTrace("after_ensure_licencias_schema")
		if err := dbpkg.EnsureUsuarioConfiguracionSchema(dbSuper); err != nil {
			log.Fatalf("failed to ensure usuario configuracion schema in superadministrador db: %v", err)
		}
		startupTrace("after_usuario_config_schema")
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
	startupTrace("after_empresa_usuarios_auth_schema")
	if err := dbpkg.EnsureEmpresaNextcloudSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure nextcloud schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_nextcloud_schema")
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure carritos schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_carritos_schema")
	if err := dbpkg.EnsureEmpresaReservasHotelSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure reservas hotel schema in empresas db: %v", err)
	}
	startupTrace("after_empresa_reservas_schema")
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

	// Inicializar tabla de mÃ©tricas y arrancar collector periÃ³dico
	if err := dbpkg.InitMetricsTable(dbSuper); err != nil {
		log.Printf("warning: failed to init metrics table: %v", err)
		utils.ReportProcessError("metrics.collector", "metrics_schema_init", "No se pudo inicializar la tabla de metricas", err, utils.ErrorLevelError, nil)
	}
	metricsInterval := metrics.DefaultIntervalSeconds()
	stopMetrics := make(chan struct{})
	go utils.RunProtectedProcess("metrics.collector", map[string]interface{}{"interval_seconds": metricsInterval}, func() {
		metrics.StartCollector(dbSuper, metricsInterval, stopMetrics)
	})

	stopAuditRetention := make(chan struct{})
	go utils.RunProtectedProcess("auditoria.retention_worker", map[string]interface{}{"interval_hours": 12}, func() {
		dbpkg.StartEmpresaAuditoriaRetentionWorker(dbEmpresas, 12*time.Hour, stopAuditRetention)
	})

	stopLicenciasEstado := make(chan struct{})
	go utils.RunProtectedProcess("licencias.estado_empresas_worker", map[string]interface{}{"interval_hours": 1}, func() {
		dbpkg.StartLicenciaEmpresaEstadoWorker(dbEmpresas, dbSuper, time.Hour, stopLicenciasEstado)
	})

	asientosInterval, asientosBatchSize, asientosMaxRetries := resolveAsientosWorkerPolicy()
	log.Printf("[asientos_worker] policy interval=%s batch=%d max_reintentos=%d", asientosInterval, asientosBatchSize, asientosMaxRetries)
	stopAsientosWorker := make(chan struct{})
	go utils.RunProtectedProcess("finanzas.asientos_worker", map[string]interface{}{"interval": asientosInterval.String(), "batch_size": asientosBatchSize, "max_retries": asientosMaxRetries}, func() {
		dbpkg.StartEmpresaAsientosContablesWorker(dbEmpresas, asientosInterval, asientosBatchSize, asientosMaxRetries, stopAsientosWorker)
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
	// Pasar la conexiÃ³n de la base `empresas` al callback para persistir usuarios y empresas
	// Pasar tanto la conexiÃ³n de empresas como la de superadministrador al callback
	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback(dbEmpresas, dbSuper, clientID, clientSecret, redirectURL))

	// Endpoint que expone configuraciÃ³n pÃºblica simple en JS.
	http.HandleFunc("/config.js", handlers.PublicConfigJSHandler(dbSuper))
	// Endpoint para procesar la aceptaciÃ³n del contrato desde la pÃ¡gina /accept.html
	http.HandleFunc("/accept/complete", handlers.AcceptCompleteHandler(dbSuper))

	// Endpoints para administraciÃ³n y auditorÃ­a (listar administradores y sesiones)
	http.HandleFunc("/super/administradores", handlers.ListAdministradoresHandler(dbSuper))
	http.HandleFunc("/super/sesiones", handlers.ListSesionesHandler(dbSuper))
	http.HandleFunc("/api/user/configuracion", handlers.UserConfiguracionHandler(dbSuper))

	// Endpoints CRUD para tipos de empresas
	http.HandleFunc("/super/api/tipos_empresas", handlers.TiposEmpresasHandler(dbSuper))
	http.HandleFunc("/super/api/tipos_empresas/preconfiguracion", handlers.SuperTipoEmpresaPreconfiguracionHandler(dbSuper))
	http.HandleFunc("/super/api/servidores", handlers.SuperServidoresListHandler(dbSuper))
	http.HandleFunc("/super/api/servidores/toggle", handlers.SuperServidoresToggleHandler(dbSuper))
	http.HandleFunc("/super/api/servidores/probar", handlers.SuperServidoresProbeHandler(dbSuper))
	http.HandleFunc("/super/api/vps/procesos", handlers.SuperVPSProcessesHandler(dbSuper))
	http.HandleFunc("/super/api/roles_de_usuario", handlers.RolesDeUsuarioHandler(dbSuper))
	http.HandleFunc("/super/api/roles_de_usuario/permisos", handlers.RolesDeUsuarioPermisosHandler(dbSuper))
	// Endpoint CRUD para empresas (persistidas en pcs_empresas PostgreSQL)
	http.HandleFunc("/super/api/empresas", handlers.EmpresasHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/super/api/empresas/compartidos", handlers.EmpresaCompartidaHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/super/api/empresas/compartidos/aceptar", handlers.EmpresaCompartidaAcceptHandler(dbEmpresas, dbSuper))
	// Endpoints para asesores comerciales y comisiones de licencias.
	http.HandleFunc("/super/api/asesor_comercial", handlers.AsesorComercialSuperHandler(dbSuper))
	http.HandleFunc("/api/asesor_comercial/aceptar", handlers.AsesorComercialAcceptHandler(dbSuper))
	http.HandleFunc("/api/asesor_comercial/mis_clientes", handlers.AsesorComercialMisClientesHandler(dbSuper))
	http.HandleFunc("/super/api/soporte_remoto", handlers.SuperSoporteRemotoHandler(dbEmpresas))
	startupTrace("after_super_and_core_routes")
	// MÃ³dulo de productos por empresa (persistido en pcs_empresas PostgreSQL)
	http.HandleFunc("/api/empresa/bodegas", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaBodegasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/categorias_productos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaCategoriasProductosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/productos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaProductosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/combos_productos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaCombosProductosHandler(dbEmpresas)))
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
	http.HandleFunc("/api/empresa/productos/precios_historial", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaProductoPrecioHistorialHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras/plan_reposicion/emitir_orden", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasPlanReposicionEmitirOrdenHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras/plan_reposicion/actualizar_estado", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasPlanReposicionActualizarEstadoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras/documentos", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasDocumentosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/compras/documentos/comprobante", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComprasDocumentoComprobanteUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/proveedores", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaProveedoresHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/servicios", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaServiciosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/usuarios/login", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioLoginHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/establecer_password", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioSetPasswordHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/solicitar_recuperacion_password", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioRequestPasswordRecoveryHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/restablecer_password", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioResetPasswordHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios/cambiar_password", handlers.WithEmpresaPublicScope(handlers.EmpresaUsuarioChangePasswordHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/usuarios", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaUsuariosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/asistencia_empleados", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaAsistenciaEmpleadosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/nomina", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaNominaSueldosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/vehiculos_registro", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaVehiculosRegistroHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/publicaciones", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaPublicacionesRedSocialHandler(dbEmpresas))) // Protegido
	http.HandleFunc("/api/empresa/publicaciones/", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaPublicacionesRedSocialHandler(dbEmpresas)))
	http.HandleFunc("/api/public/publicaciones", handlers.PublicacionesRedSocialHandler(dbEmpresas)) // Publico
	http.HandleFunc("/api/public/market_symbol", handlers.PublicMarketSymbolHandler())
	http.HandleFunc("/api/public/publicaciones/", handlers.PublicRedSocialInteraccionesHandler(dbEmpresas))
	http.HandleFunc("/api/empresa/clientes", handlers.WithEmpresaClientesPermissions(dbEmpresas, dbSuper, handlers.EmpresaClientesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/carritos_compra", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCarritosCompraHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/carritos_compra/items", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCarritoItemsHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/venta_publica", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaVentaPublicaHandler(dbEmpresas)))
	http.HandleFunc("/api/public/venta_publica", handlers.PublicVentaPublicaHandler(dbEmpresas))
	http.HandleFunc("/api/public/estacion_vip", handlers.PublicEstacionVIPHandler(dbEmpresas))
	http.HandleFunc("/api/public/chat_portal", handlers.PublicPortalCompanyChatHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/public/chat_portal_stream", handlers.PublicPortalCompanyChatStreamHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/public/mensajes_privados", handlers.PublicMensajesPrivadosHandler(dbEmpresas))
	http.HandleFunc("/api/public/soporte_remoto", handlers.PublicEmpresaSoporteRemotoAgentHandler(dbEmpresas))
	http.HandleFunc("/api/public/venta_digital", handlers.PublicVentaDigitalHandler(dbSuper))
	http.HandleFunc("/api/public/pagina_principal", handlers.PublicPaginaPrincipalHandler(dbSuper))
	http.HandleFunc("/api/public/contrato", handlers.PublicContratoHandler(dbSuper))
	http.HandleFunc("/api/public/geo", handlers.PublicGeoHandler())
	http.HandleFunc("/api/empresa/reservas_hotel", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaReservasHotelHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/tarifas_por_minutos", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaTarifasPorMinutosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/tarifas_por_dia", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaTarifasPorDiaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/codigos_de_descuento", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCodigosDescuentoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/propinas", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaPropinasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/comisiones", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaComisionesServicioHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_general", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionGeneralHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_operativa", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionOperativaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_avanzada", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionAvanzadaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/db_admin", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaDBAdminHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/impresoras", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaImpresorasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/impresoras/resolver", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaImpresorasResolverHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/estacion_prefs", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaEstacionPrefsHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/facturacion_electronica/pais_detectado", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaPaisDetectadoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica/paises_disponibles", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaPaisesDisponiblesHandler()))
	http.HandleFunc("/api/empresa/impuestos", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaImpuestosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/chat_tareas/conversaciones", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasConversacionesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/participantes", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasParticipantesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/mensajes", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasMensajesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/mensajes/adjunto", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasAdjuntoUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/tareas", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasTareasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/citas", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasCitasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/tareas/nota_voz", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasTareaNotaVozUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/papelera", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasPapeleraHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/ubicacion_gps/dispositivos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaUbicacionGPSDispositivosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/ubicacion_gps/recorridos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaUbicacionGPSRecorridosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/hoja_vida_operativa", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaHojaVidaOperativaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/movimientos", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasMovimientosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/movimientos/comprobante", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasMovimientoComprobanteUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/corte_caja", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCorteCajaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/configuracion", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasConfiguracionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/periodos", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasPeriodosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/asientos_contables", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasAsientosContablesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/cierres_caja", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasCierresCajaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/calculadora", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCalculadoraHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/creditos", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCreditosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/backups", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaBackupsHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/documentos", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.OnlyOfficeDocumentosHandler(dbSuper)))
	http.HandleFunc("/api/empresa/nextcloud", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaNextcloudHandler(dbEmpresas, dbSuper)))
	startupTrace("after_empresa_routes")

	// OnlyOffice public endpoints (token temporal)
	http.HandleFunc("/api/onlyoffice/file", handlers.OnlyOfficeFilePublicHandler(dbSuper))
	http.HandleFunc("/api/onlyoffice/callback", handlers.OnlyOfficeCallbackPublicHandler(dbSuper))
	http.HandleFunc("/api/empresa/soporte_remoto", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaSoporteRemotoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/reportes", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaReportesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/reportes_ia_chat", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaReportesIAChatHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/graficos_estadisticas", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaGraficosEstadisticasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/auditoria/eventos", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaAuditoriaEventosHandler(dbEmpresas)))
	// Endpoint empresa: verificar acceso a la página frecuencia_fe.html (alias legacy frecuencia_fp.html)
	http.HandleFunc("/api/empresa/frecuencia_fp/permitido", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaFrecuenciaFPAllowedHandler(dbSuper)))
	handlers.RegisterEmpresaChatIARoutes(dbEmpresas, dbSuper)
	handlers.RegisterEmpresaModulosFaltantesRoutes(dbEmpresas, dbSuper)
	// Rutas del mÃ³dulo sensor de puertas: configuraciÃ³n protegida y endpoint pÃºblico para heartbeats
	http.HandleFunc("/api/empresa/sensor_puertas", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaSensorConfigHandler(dbEmpresas)))
	http.HandleFunc("/api/public/sensor_puertas", handlers.PublicSensorPuertasHandler(dbEmpresas))
	http.HandleFunc("/api/empresa/sensor_puertas/messages", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaSensorMessagesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/roles_de_usuario", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaRolesDeUsuarioHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/permisos_contexto", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaPermisosContextoHandler(dbSuper)))
	http.HandleFunc("/api/empresa/permisos_empresa", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaPermisosFinosHandler(dbSuper)))
	// Endpoint para obtener admin actual desde la cookie de sesiÃ³n
	http.HandleFunc("/me", handlers.MeHandler(dbSuper))
	// Endpoint para obtener perfil/cuenta enriquecida (admin + usuario de empresa)
	http.HandleFunc("/api/account", handlers.AccountHandler(dbEmpresas, dbSuper))
	// Endpoints para actualizar perfil y cambiar contraseÃ±a (usuario autenticado)
	http.HandleFunc("/api/account/update_profile", handlers.AccountUpdateProfileHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/account/change_password", handlers.AccountChangePasswordHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/account/set_google_password", handlers.AccountSetGooglePasswordHandler(dbEmpresas, dbSuper))
	// Endpoint CRUD para administradores (API)
	http.HandleFunc("/super/api/administradores", handlers.AdministradoresHandler(dbSuper))
	// Endpoints adicionales para flujo de autenticaciÃ³n de administradores (registro, login, confirmaciÃ³n, recuperaciÃ³n)
	http.HandleFunc("/super/api/administradores/register", handlers.AdminRegisterHandler(dbSuper))
	http.HandleFunc("/super/api/administradores/login", handlers.AdminLoginHandler(dbSuper))
	http.HandleFunc("/auth/confirmar_admin", handlers.ConfirmarAdminHandler(dbSuper))
	http.HandleFunc("/super/api/administradores/solicitar_recuperacion", handlers.AdminRequestPasswordRecoveryHandler(dbSuper))
	http.HandleFunc("/super/api/administradores/restablecer_password", handlers.AdminResetPasswordHandler(dbSuper))
	// Endpoint CRUD para licencias (nuevo)
	http.HandleFunc("/super/api/licencias", handlers.LicenciasHandler(dbSuper))
	// Endpoint super: lista de administradores autorizados (Frecuencia FE/FP)
	http.HandleFunc("/super/api/administradores_frecuencia_fe", handlers.SuperAdministradoresFrecuenciaFEHandler(dbSuper))
	// Endpoint publico para exponer metodos de pago activos del checkout de licencias
	http.HandleFunc("/api/public/licencias/payment_methods", handlers.PublicLicenciasPaymentMethodsHandler(dbSuper))
	// Endpoint publico para calcular total, descuentos y activacion sin pago del checkout de licencias
	http.HandleFunc("/api/public/licencias/checkout_summary", handlers.LicenciaCheckoutSummaryHandler(dbSuper))
	// Endpoint para gestionar credenciales de Wompi (GET/PUT)
	http.HandleFunc("/super/api/config/wompi", handlers.WompiConfigHandler(dbSuper))
	// Endpoint para gestionar credenciales de Epayco (GET/PUT)
	http.HandleFunc("/super/api/config/epayco", handlers.EpaycoConfigHandler(dbSuper))
	// Endpoint para gestionar SMTP Gmail (GET/PUT)
	http.HandleFunc("/super/api/config/gmail", handlers.GmailConfigHandler(dbSuper))
	// Endpoint para activar o desactivar Google reCAPTCHA (GET/PUT)
	http.HandleFunc("/super/api/config/recaptcha", handlers.RecaptchaConfigHandler(dbSuper))
	// Endpoint para administrar plantillas de correo del panel super
	http.HandleFunc("/super/api/config/email_templates", handlers.SuperEmailTemplatesHandler(dbSuper))
	// Endpoint super para administrar venta digital global
	http.HandleFunc("/super/api/venta_digital", handlers.SuperVentaDigitalHandler(dbSuper))
	// Endpoint para gestionar credenciales IA de modelos populares (GET/PUT)
	http.HandleFunc("/super/api/config/ai", handlers.AIModelsConfigHandler(dbSuper))
	// Endpoint para configurar limitaciones por empresa (RustDesk e IA)
	http.HandleFunc("/super/api/config/limitaciones_empresa", handlers.SuperEmpresaLimitacionesConfigHandler(dbSuper))
	// Endpoint para configurar la lógica del chat con IA (empresas y super)
	http.HandleFunc("/super/api/config/chat_ia_logica", handlers.SuperChatIALogicaConfigHandler(dbEmpresas, dbSuper))
	// Endpoint para configurar gestion RustDesk en el VPS (GET/PUT)
	http.HandleFunc("/super/api/config/rustdesk", handlers.RustDeskConfigHandler(dbSuper))
	// Endpoint para configurar voz natural por streaming en VPS (GET/PUT)
	http.HandleFunc("/super/api/config/voice_stream", handlers.SuperVoiceStreamConfigHandler(dbSuper))
	http.HandleFunc("/api/voice_stream/status", handlers.VoiceStreamStatusHandler(dbSuper))
	http.HandleFunc("/api/voice_stream/tts", handlers.VoiceStreamTTSProxyHandler(dbSuper))
	http.HandleFunc("/api/chat_flotante/preferencias", handlers.ChatFlotantePreferenciasHandler(dbSuper))
	// Endpoints para generar y descargar documentos dinamicos asistidos por IA.
	http.HandleFunc("/generate", handlers.DynamicDocumentGenerateHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/download", handlers.DynamicDocumentDownloadHandler(dbSuper))
	// Endpoint para configurar Nextcloud en el VPS (GET/PUT)
	http.HandleFunc("/super/api/config/nextcloud", handlers.NextcloudConfigHandler(dbSuper))
	superAIChatController := handlers.NewSuperAIChatController(dbEmpresas, dbSuper)
	http.HandleFunc("/super/api/chat_con_ia_global/modelos", superAIChatController.ModelosHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/modelo_preferido", superAIChatController.ModeloPreferidoHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/consultar", superAIChatController.ConsultarHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/consultar_con_adjunto", superAIChatController.ConsultarConAdjuntoHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/consultar_stream", superAIChatController.ConsultarStreamHandler)
	http.HandleFunc("/super/api/chat_con_ia_global/historial", superAIChatController.HistorialHandler)
	// Endpoint para respaldo/restauracion de configuracion critica del panel super
	http.HandleFunc("/super/api/config/backup", handlers.SuperConfigBackupHandler(dbSuper))
	// Endpoint para configuración de modo mantenimiento global
	http.HandleFunc("/super/api/config/mantenimiento", handlers.SuperMantenimientoConfigHandler(dbSuper))
	http.HandleFunc("/super/api/config/onlyoffice", handlers.OnlyOfficeConfigHandler(dbSuper))
	// Endpoint super para administrar contrato versionado y su historial
	http.HandleFunc("/super/api/contrato", handlers.SuperContratoHandler(dbSuper))
	// Endpoint super para monitoreo centralizado de errores del sistema
	http.HandleFunc("/super/api/errores", handlers.SuperErroresSistemaHandler(dbSuper))
	// Endpoint super para consumos (OpenAI/Hostinger/Cursor) y contador de errores
	http.HandleFunc("/super/api/consumos", handlers.SuperConsumosHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/super/api/config/portal_chat_ia_info", handlers.SuperPortalChatIAInfoHandler(dbSuper))
	http.HandleFunc("/super/api/config/contexto_ia_logica_negocio", handlers.SuperContextoIALogicaNegocioHandler(dbSuper))
	// Endpoint super para administrar tarjetas dinamicas de la pagina principal (index)
	http.HandleFunc("/super/api/pagina_principal", handlers.SuperPaginaPrincipalHandler(dbSuper, webDir))
	// Endpoints Wompi (Nequi): crear transacciÃ³n y consultar estado
	http.HandleFunc("/wompi/terms", handlers.WompiTermsHandler(dbSuper))
	http.HandleFunc("/wompi/create_transaction_nequi", handlers.WompiCreateNequiTransactionHandler(dbSuper))
	http.HandleFunc("/wompi/transaction_status", handlers.WompiTransactionStatusHandler(dbSuper))
	http.HandleFunc("/wompi/webhook", handlers.WompiWebhookHandler(dbSuper, dbEmpresas))
	// Endpoints Epayco: crear transacciÃ³n y consultar estado
	http.HandleFunc("/epayco/create_transaction", handlers.EpaycoCreateTransactionHandler(dbSuper))
	http.HandleFunc("/epayco/transaction_status", handlers.EpaycoTransactionStatusHandler(dbSuper))
	http.HandleFunc("/epayco/webhook", handlers.EpaycoWebhookHandler(dbSuper, dbEmpresas))
	// ActivaciÃ³n manual de licencia sin pago (uso interno de avance/prototipo)
	http.HandleFunc("/licencias/activar_sin_pago", handlers.ActivateLicenciaSinPagoHandler(dbSuper, dbEmpresas))
	// ConfirmaciÃ³n de correo para usuarios de empresa.
	http.HandleFunc("/auth/confirmar_correo", handlers.ConfirmarCorreoUsuarioHandler(dbEmpresas))

	// Endpoints de mÃ©tricas (actual y histÃ³rico)
	http.HandleFunc("/super/api/metrics/current", handlers.MetricsCurrentHandler(dbSuper))
	http.HandleFunc("/super/api/metrics/history", handlers.MetricsHistoryHandler(dbSuper))
	http.HandleFunc("/super/api/reportes_globales", handlers.SuperReportesGlobalesHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/super/api/postgres/performance", handlers.PostgresPerformanceHandler(dbEmpresas, dbSuper))
	// Endpoint de seguridad: escaneo de puertos
	http.HandleFunc("/super/api/security/ports", handlers.SecurityPortsHandler(dbSuper))
	// Endpoint de seguridad: listado de procesos en memoria RAM
	http.HandleFunc("/super/api/security/processes", handlers.SecurityProcessesHandler(dbSuper))
	http.HandleFunc("/super/api/security/vps/config", handlers.SecurityVPSConfigHandler(dbSuper, vpsSecurityService))
	http.HandleFunc("/super/api/security/vps/run", handlers.SecurityVPSRunHandler(dbSuper, vpsSecurityService))
	http.HandleFunc("/super/api/security/vps/status", handlers.SecurityVPSStatusHandler(dbSuper, vpsSecurityService))
	http.HandleFunc("/super/api/security/vps/history", handlers.SecurityVPSHistoryHandler(dbSuper, vpsSecurityService))
	http.HandleFunc("/super/api/security/vps/report", handlers.SecurityVPSReportHandler(dbSuper, vpsSecurityService))
	http.HandleFunc("/super/api/security/vps/compare", handlers.SecurityVPSCompareHandler(dbSuper, vpsSecurityService))
	startupTrace("after_super_config_routes")

	// Logout handler: limpiar cookie de sesiÃ³n (si existe) y redirigir a la pÃ¡gina de login
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
	registerLocalEmulatorRoutes(backendDir, webDir)
	startupTrace("after_static_helper_routes")

	// Servir pÃ¡ginas estÃ¡ticas desde la carpeta `web` detectada
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
	staticFS := noCacheAdminStaticHandler(http.FileServer(http.Dir(webDir)))
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

		trimmed := strings.Trim(path, "/")
		parts := strings.Split(trimmed, "/")
		if len(parts) == 2 && strings.EqualFold(parts[1], "venta_publica.html") && strings.TrimSpace(parts[0]) != "" {
			r2 := r.Clone(r.Context())
			r2.URL.Path = "/venta_publica.html"
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
	handler := utils.LoggingMiddleware(utils.CanonicalPublicHostMiddleware(utils.JSONErrorMiddleware(utils.RecoveryMiddleware(utils.AuthMiddleware(dbSuper, http.DefaultServeMux)))))
	startupTrace("after_handler_wrap")

	// Respetar la variable de entorno PORT si estÃ¡ definida; por defecto usar 8080
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
		Addr:    addr,
		Handler: handler,
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
