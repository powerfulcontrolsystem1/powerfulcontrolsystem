package db

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/you/pos-backend/secure"
)

var ErrSoporteRemotoPlanLimit = errors.New("limite de soporte remoto alcanzado")
var ErrSoporteRemotoSignalingCredential = errors.New("credencial de senalizacion invalida")

// EmpresaSoporteRemotoConfig define configuracion operativa de soporte remoto por empresa.
type EmpresaSoporteRemotoConfig struct {
	ID                         int64  `json:"id"`
	EmpresaID                  int64  `json:"empresa_id"`
	Habilitado                 bool   `json:"habilitado"`
	ProveedorPreferido         string `json:"proveedor_preferido"`
	ModoOperacion              string `json:"modo_operacion"`
	RequiereAprobacionOperador bool   `json:"requiere_aprobacion_operador"`
	AutoCerrarMinutos          int    `json:"auto_cerrar_minutos"`
	MaxConexionesMes           int    `json:"max_conexiones_mes"`
	MaxMinutosMes              int    `json:"max_minutos_mes"`
	MaxMinutosDiaRustDesk      int    `json:"max_minutos_dia_rustdesk"`
	MaxDispositivos            int    `json:"max_dispositivos"`
	PortalPublicoHabilitado    bool   `json:"portal_publico_habilitado"`
	RustDeskServerHost         string `json:"rustdesk_server_host,omitempty"`
	RustDeskServerKey          string `json:"rustdesk_server_key,omitempty"`
	ClienteWindowsURL          string `json:"cliente_windows_url,omitempty"`
	ClienteLinuxURL            string `json:"cliente_linux_url,omitempty"`
	ClienteMacURL              string `json:"cliente_mac_url,omitempty"`
	ServidorWindowsURL         string `json:"servidor_windows_url,omitempty"`
	ServidorLinuxURL           string `json:"servidor_linux_url,omitempty"`
	ServidorMacURL             string `json:"servidor_mac_url,omitempty"`
	CarpetaTransferencia       string `json:"carpeta_transferencia,omitempty"`
	InstruccionesPublicas      string `json:"instrucciones_publicas,omitempty"`
	FechaCreacion              string `json:"fecha_creacion,omitempty"`
	FechaActualizacion         string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador             string `json:"usuario_creador,omitempty"`
	Estado                     string `json:"estado,omitempty"`
	Observaciones              string `json:"observaciones,omitempty"`
}

// EmpresaSoporteRemotoDispositivo representa un equipo remoto registrado por empresa.
type EmpresaSoporteRemotoDispositivo struct {
	ID                      int64  `json:"id"`
	EmpresaID               int64  `json:"empresa_id"`
	CodigoDispositivo       string `json:"codigo_dispositivo"`
	NombreEquipo            string `json:"nombre_equipo"`
	AliasOperativo          string `json:"alias_operativo,omitempty"`
	Ubicacion               string `json:"ubicacion,omitempty"`
	SistemaOperativo        string `json:"sistema_operativo,omitempty"`
	AgenteVersion           string `json:"agente_version,omitempty"`
	StreamURL               string `json:"stream_url"`
	RustDeskDeviceID        string `json:"rustdesk_device_id,omitempty"`
	CarpetaTransferencia    string `json:"carpeta_transferencia,omitempty"`
	AccesoPublicoHabilitado bool   `json:"acceso_publico_habilitado"`
	EstadoConexion          string `json:"estado_conexion"`
	UltimoHeartbeat         string `json:"ultimo_heartbeat,omitempty"`
	AccesoPINHash           string `json:"-"`
	RustDeskPasswordEnc     string `json:"-"`
	FechaCreacion           string `json:"fecha_creacion,omitempty"`
	FechaActualizacion      string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador          string `json:"usuario_creador,omitempty"`
	Estado                  string `json:"estado,omitempty"`
	Observaciones           string `json:"observaciones,omitempty"`
}

// EmpresaSoporteRemotoSession representa una sesion de visualizacion remota por empresa.
type EmpresaSoporteRemotoSession struct {
	ID                    int64  `json:"id"`
	EmpresaID             int64  `json:"empresa_id"`
	DispositivoID         int64  `json:"dispositivo_id"`
	DispositivoCodigo     string `json:"dispositivo_codigo,omitempty"`
	DispositivoNombre     string `json:"dispositivo_nombre,omitempty"`
	CodigoSesion          string `json:"codigo_sesion"`
	SolicitadaPor         string `json:"solicitada_por,omitempty"`
	OperadorNombre        string `json:"operador_nombre,omitempty"`
	OperadorEmail         string `json:"operador_email,omitempty"`
	Motivo                string `json:"motivo,omitempty"`
	EstadoSesion          string `json:"estado_sesion"`
	DuracionMinSolicitada int    `json:"duracion_min_solicitada,omitempty"`
	DuracionMinConsumida  int    `json:"duracion_min_consumida,omitempty"`
	BloqueadaPorLimite    bool   `json:"bloqueada_por_limite,omitempty"`
	URLVisualizacion      string `json:"url_visualizacion,omitempty"`
	TokenVisualizacionRaw string `json:"token_visualizacion,omitempty"`
	IniciadaEn            string `json:"iniciada_en,omitempty"`
	ExpiraEn              string `json:"expira_en,omitempty"`
	FinalizadaEn          string `json:"finalizada_en,omitempty"`
	FechaCreacion         string `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string `json:"usuario_creador,omitempty"`
	Estado                string `json:"estado,omitempty"`
	Observaciones         string `json:"observaciones,omitempty"`

	tokenVisualizacionHash string
}

// EmpresaSoporteRemotoSignalingCredential solo expone el secreto al crearlo.
// La base de datos conserva exclusivamente sus verificadores SHA-256.
type EmpresaSoporteRemotoSignalingCredential struct {
	TokenRaw  string `json:"token"`
	NonceRaw  string `json:"nonce"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
}

// EmpresaSoporteRemotoDispositivoFilter define filtros para listar dispositivos.
type EmpresaSoporteRemotoDispositivoFilter struct {
	IncludeInactive bool
	Q               string
	Limit           int
	Offset          int
}

// EmpresaSoporteRemotoSessionFilter define filtros para listar sesiones.
type EmpresaSoporteRemotoSessionFilter struct {
	IncludeInactive bool
	EstadoSesion    string
	Q               string
	Limit           int
	Offset          int
}

type EmpresaSoporteRemotoUso struct {
	EmpresaID                     int64  `json:"empresa_id"`
	MesReferencia                 string `json:"mes_referencia"`
	DiaReferencia                 string `json:"dia_referencia"`
	DispositivosActivos           int64  `json:"dispositivos_activos"`
	DispositivosOnline            int64  `json:"dispositivos_online"`
	SesionesMes                   int64  `json:"sesiones_mes"`
	IntentosBloqueadosMes         int64  `json:"intentos_bloqueados_mes"`
	MinutosConsumidosMes          int64  `json:"minutos_consumidos_mes"`
	MinutosConsumidosDiaRustDesk  int64  `json:"minutos_consumidos_dia_rustdesk"`
	MaxConexionesMes              int64  `json:"max_conexiones_mes"`
	MaxMinutosMes                 int64  `json:"max_minutos_mes"`
	MaxMinutosDiaRustDesk         int64  `json:"max_minutos_dia_rustdesk"`
	MinutosDisponiblesDiaRustDesk int64  `json:"minutos_disponibles_dia_rustdesk"`
	MaxDispositivos               int64  `json:"max_dispositivos"`
	PuedeCrearDispositivo         bool   `json:"puede_crear_dispositivo"`
	PuedeCrearSesion              bool   `json:"puede_crear_sesion"`
	BloqueoMotivo                 string `json:"bloqueo_motivo,omitempty"`
}

func soporteRemotoNormalizeEstado(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func soporteRemotoNormalizeConexion(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "online", "offline", "intermitente":
		return value
	default:
		return "offline"
	}
}

func soporteRemotoNormalizeSesionEstado(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "pendiente", "aprobada", "activa", "finalizada", "rechazada", "expirada":
		return value
	default:
		return "pendiente"
	}
}

func soporteRemotoNormalizeProveedor(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "novnc", "rustdesk_web", "rustdesk_oss", "guacamole", "custom_url":
		return value
	default:
		return "novnc"
	}
}

func soporteRemotoNormalizeModo(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "agente_web", "agente_local", "cliente_local", "hibrido":
		return value
	default:
		return "agente_web"
	}
}

func soporteRemotoNormalizeURL(raw string) string {
	return strings.TrimSpace(raw)
}

func soporteRemotoEncryptSecret(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	return secure.EncryptString(value)
}

func soporteRemotoDecryptSecret(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", nil
	}
	return secure.DecryptString(value)
}

// ResolveEmpresaSoporteRemotoRustDeskPassword descifra el secreto operativo del dispositivo cuando existe.
func ResolveEmpresaSoporteRemotoRustDeskPassword(item EmpresaSoporteRemotoDispositivo) (string, error) {
	return soporteRemotoDecryptSecret(item.RustDeskPasswordEnc)
}

func soporteRemotoNormalizeAutoCerrar(raw int) int {
	if raw <= 0 {
		return 30
	}
	if raw > 480 {
		return 480
	}
	return raw
}

func soporteRemotoNormalizePlanLimit(raw int) int {
	if raw < 0 {
		return 0
	}
	if raw > 100000 {
		return 100000
	}
	return raw
}

func soporteRemotoNormalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func soporteRemotoLikePattern(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "!", "!!")
	value = strings.ReplaceAll(value, "%", "!%")
	value = strings.ReplaceAll(value, "_", "!_")
	return "%" + value + "%"
}

func soporteRemotoBoolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func soporteRemotoHash(v string) string {
	raw := strings.TrimSpace(v)
	if raw == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func soporteRemotoHashEqual(raw, expectedHash string) bool {
	actual := soporteRemotoHash(raw)
	expected := strings.TrimSpace(expectedHash)
	if actual == "" || len(actual) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(actual), []byte(expected)) == 1
}

func soporteRemotoGenerateSecureSecret(prefix string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("no se pudo generar secreto criptografico: %w", err)
	}
	base := strings.ToUpper(strings.TrimSpace(prefix))
	if base == "" {
		base = "SR"
	}
	return base + "-" + hex.EncodeToString(buf), nil
}

func soporteRemotoGenerateToken(prefix string) string {
	buf := make([]byte, 18)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", strings.ToUpper(strings.TrimSpace(prefix)), time.Now().UnixNano())
	}
	base := strings.ToUpper(strings.TrimSpace(prefix))
	if base == "" {
		base = "SR"
	}
	return base + "-" + hex.EncodeToString(buf)
}

func soporteRemotoGenerateDeviceCode(nombre string) string {
	base := NormalizeEmpresaPublicSlug(nombre)
	if base == "" || base == "empresa" {
		base = "dispositivo"
	}
	if len(base) > 20 {
		base = base[:20]
	}
	return strings.ToUpper("SRD-" + base + "-" + time.Now().In(time.Local).Format("20060102150405"))
}

func soporteRemotoGenerateSessionCode() string {
	return strings.ToUpper("SRS-" + time.Now().In(time.Local).Format("20060102150405") + "-" + soporteRemotoGenerateToken(""))
}

func soporteRemotoParseDateTime(raw string) (time.Time, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func soporteRemotoIsCurrentMonth(raw string, now time.Time) bool {
	parsed, ok := soporteRemotoParseDateTime(raw)
	if !ok {
		return false
	}
	return parsed.Year() == now.Year() && parsed.Month() == now.Month()
}

func soporteRemotoComputeSessionMinutes(iniciadaEn, finalizadaEn string, fallback int) int {
	start, ok := soporteRemotoParseDateTime(iniciadaEn)
	if !ok {
		if fallback > 0 {
			return fallback
		}
		return 0
	}
	end, ok := soporteRemotoParseDateTime(finalizadaEn)
	if !ok {
		end = time.Now().In(time.Local)
	}
	if end.Before(start) {
		end = start
	}
	minutes := int(math.Ceil(end.Sub(start).Minutes()))
	if minutes <= 0 {
		minutes = 1
	}
	if fallback > 0 && minutes > fallback {
		return fallback
	}
	return minutes
}

func soporteRemotoComputeSessionMinutesInWindow(iniciadaEn, finalizadaEn string, windowStart, windowEnd time.Time) int {
	start, ok := soporteRemotoParseDateTime(iniciadaEn)
	if !ok {
		return 0
	}
	end, ok := soporteRemotoParseDateTime(finalizadaEn)
	if !ok {
		end = time.Now().In(time.Local)
	}
	if end.Before(start) {
		end = start
	}
	if !end.After(windowStart) || !start.Before(windowEnd) {
		return 0
	}
	if start.Before(windowStart) {
		start = windowStart
	}
	if end.After(windowEnd) {
		end = windowEnd
	}
	minutes := int(math.Ceil(end.Sub(start).Minutes()))
	if minutes <= 0 {
		return 1
	}
	return minutes
}

func soporteRemotoUsesRustDesk(cfg EmpresaSoporteRemotoConfig, rustDeskDeviceID string) bool {
	if strings.TrimSpace(rustDeskDeviceID) != "" {
		return true
	}
	provider := soporteRemotoNormalizeProveedor(cfg.ProveedorPreferido)
	return provider == "rustdesk_web" || provider == "rustdesk_oss"
}

// EnsureEmpresaSoporteRemotoSchema crea/migra tablas de soporte remoto por empresa.
func EnsureEmpresaSoporteRemotoSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_soporte_remoto_configuracion (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL UNIQUE,
			habilitado INTEGER DEFAULT 1,
			proveedor_preferido TEXT DEFAULT 'novnc',
			modo_operacion TEXT DEFAULT 'agente_web',
			requiere_aprobacion_operador INTEGER DEFAULT 1,
			auto_cerrar_minutos INTEGER DEFAULT 30,
			max_conexiones_mes INTEGER DEFAULT 0,
			max_minutos_mes INTEGER DEFAULT 0,
			max_minutos_dia_rustdesk INTEGER DEFAULT 0,
			max_dispositivos INTEGER DEFAULT 0,
			portal_publico_habilitado INTEGER DEFAULT 1,
			rustdesk_server_host TEXT,
			rustdesk_server_key TEXT,
			cliente_windows_url TEXT,
			cliente_linux_url TEXT,
			cliente_mac_url TEXT,
			servidor_windows_url TEXT,
			servidor_linux_url TEXT,
			servidor_mac_url TEXT,
			carpeta_transferencia TEXT,
			instrucciones_publicas TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_soporte_remoto_configuracion_estado
		ON empresa_soporte_remoto_configuracion(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_soporte_remoto_dispositivos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo_dispositivo TEXT NOT NULL,
			nombre_equipo TEXT NOT NULL,
			alias_operativo TEXT,
			ubicacion TEXT,
			sistema_operativo TEXT,
			agente_version TEXT,
			stream_url TEXT NOT NULL,
			rustdesk_device_id TEXT,
			rustdesk_password_enc TEXT,
			carpeta_transferencia TEXT,
			acceso_publico_habilitado INTEGER DEFAULT 1,
			estado_conexion TEXT DEFAULT 'offline',
			ultimo_heartbeat TEXT,
			acceso_pin_hash TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_soporte_remoto_dispositivos_codigo
		ON empresa_soporte_remoto_dispositivos(empresa_id, codigo_dispositivo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_soporte_remoto_dispositivos_lookup
		ON empresa_soporte_remoto_dispositivos(empresa_id, estado, estado_conexion, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_soporte_remoto_sesiones (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			dispositivo_id INTEGER NOT NULL,
			codigo_sesion TEXT NOT NULL,
			solicitada_por TEXT,
			operador_nombre TEXT,
			operador_email TEXT,
			motivo TEXT,
			estado_sesion TEXT DEFAULT 'pendiente',
			duracion_minutos_solicitada INTEGER DEFAULT 0,
			duracion_minutos_consumida INTEGER DEFAULT 0,
			bloqueada_por_limite INTEGER DEFAULT 0,
			token_visualizacion_hash TEXT,
			token_visualizacion_usado_en TEXT,
			url_visualizacion TEXT,
			iniciada_en TEXT,
			expira_en TEXT,
			finalizada_en TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_soporte_remoto_sesiones_codigo
		ON empresa_soporte_remoto_sesiones(empresa_id, codigo_sesion);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_soporte_remoto_sesiones_lookup
		ON empresa_soporte_remoto_sesiones(empresa_id, estado_sesion, fecha_creacion DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_soporte_remoto_signaling_tokens (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			sesion_id INTEGER NOT NULL,
			role TEXT NOT NULL,
			token_hash TEXT NOT NULL,
			nonce_hash TEXT NOT NULL,
			expira_en TEXT NOT NULL,
			usado_en TEXT,
			revocado_en TEXT,
			creado_por TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP)
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_soporte_remoto_signaling_token
		ON empresa_soporte_remoto_signaling_tokens(token_hash);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_soporte_remoto_signaling_lookup
		ON empresa_soporte_remoto_signaling_tokens(empresa_id, sesion_id, role, fecha_creacion DESC);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "proveedor_preferido", "TEXT DEFAULT 'novnc'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "modo_operacion", "TEXT DEFAULT 'agente_web'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "max_conexiones_mes", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "max_minutos_mes", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "max_minutos_dia_rustdesk", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "max_dispositivos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "portal_publico_habilitado", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "rustdesk_server_host", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "rustdesk_server_key", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "cliente_windows_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "cliente_linux_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "cliente_mac_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "servidor_windows_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "servidor_linux_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "servidor_mac_url", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "carpeta_transferencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_configuracion", "instrucciones_publicas", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_dispositivos", "acceso_pin_hash", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_dispositivos", "rustdesk_device_id", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_dispositivos", "rustdesk_password_enc", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_dispositivos", "carpeta_transferencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_dispositivos", "acceso_publico_habilitado", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_sesiones", "token_visualizacion_hash", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_sesiones", "token_visualizacion_usado_en", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_sesiones", "url_visualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_sesiones", "duracion_minutos_solicitada", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_sesiones", "duracion_minutos_consumida", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_soporte_remoto_sesiones", "bloqueada_por_limite", "INTEGER DEFAULT 0"); err != nil {
		return err
	}

	return nil
}

// SeedEmpresaSoporteRemotoDefaults deja una configuracion RustDesk persistida
// para cada empresa sin reemplazar decisiones particulares ya guardadas.
func SeedEmpresaSoporteRemotoDefaults(dbEmp, dbSuper *sql.DB) (created, updated int64, err error) {
	if dbEmp == nil || dbSuper == nil {
		return 0, 0, errors.New("db connection is nil")
	}
	if err := EnsureEmpresaSoporteRemotoSchema(dbEmp); err != nil {
		return 0, 0, err
	}

	host, _, _, _, err := GetConfigEntry(dbSuper, "rustdesk.server_host")
	if err != nil {
		return 0, 0, err
	}
	publicKey, _, _, _, err := GetConfigEntry(dbSuper, "rustdesk.server_key")
	if err != nil {
		return 0, 0, err
	}
	host = strings.TrimSpace(host)
	publicKey = strings.TrimSpace(publicKey)

	result, err := execSQLCompat(dbEmp, `
		INSERT INTO empresa_soporte_remoto_configuracion (
			empresa_id, habilitado, proveedor_preferido, modo_operacion,
			requiere_aprobacion_operador, auto_cerrar_minutos,
			max_conexiones_mes, max_minutos_mes, max_minutos_dia_rustdesk,
			max_dispositivos, portal_publico_habilitado,
			rustdesk_server_host, rustdesk_server_key, usuario_creador, estado
		)
		SELECT COALESCE(NULLIF(empresa_id, 0), id), 1, 'rustdesk_oss', 'cliente_local',
			1, 30, 0, 0, 30, 0, 1, ?, ?, 'sistema.rustdesk', 'activo'
		FROM empresas
		WHERE COALESCE(NULLIF(empresa_id, 0), id) > 0
		ON CONFLICT (empresa_id) DO NOTHING`, host, publicKey)
	if err != nil {
		return 0, 0, err
	}
	if result != nil {
		created, _ = result.RowsAffected()
	}

	if host == "" && publicKey == "" {
		return created, 0, nil
	}
	result, err = execSQLCompat(dbEmp, `
		UPDATE empresa_soporte_remoto_configuracion
		SET rustdesk_server_host = CASE WHEN ? <> '' AND BTRIM(COALESCE(rustdesk_server_host, '')) = '' THEN ? ELSE rustdesk_server_host END,
			rustdesk_server_key = CASE WHEN ? <> '' AND BTRIM(COALESCE(rustdesk_server_key, '')) = '' THEN ? ELSE rustdesk_server_key END,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE BTRIM(COALESCE(rustdesk_server_host, '')) = ''
			OR BTRIM(COALESCE(rustdesk_server_key, '')) = ''`, host, host, publicKey, publicKey)
	if err != nil {
		return created, 0, err
	}
	if result != nil {
		updated, _ = result.RowsAffected()
	}
	return created, updated, nil
}

// GetEmpresaSoporteRemotoConfig consulta configuracion de soporte remoto por empresa.
func GetEmpresaSoporteRemotoConfig(dbConn *sql.DB, empresaID int64) (EmpresaSoporteRemotoConfig, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoConfig{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaSoporteRemotoConfig{}, errors.New("empresa_id invalido")
	}

	var out EmpresaSoporteRemotoConfig
	var habilitado sql.NullInt64
	var requiereAprobacion sql.NullInt64
	var portalPublico sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(habilitado, 1),
		COALESCE(proveedor_preferido, 'novnc'),
		COALESCE(modo_operacion, 'agente_web'),
		COALESCE(requiere_aprobacion_operador, 1),
		COALESCE(auto_cerrar_minutos, 30),
		COALESCE(max_conexiones_mes, 0),
		COALESCE(max_minutos_mes, 0),
		COALESCE(max_minutos_dia_rustdesk, 0),
		COALESCE(max_dispositivos, 0),
		COALESCE(portal_publico_habilitado, 1),
		COALESCE(rustdesk_server_host, ''),
		COALESCE(rustdesk_server_key, ''),
		COALESCE(cliente_windows_url, ''),
		COALESCE(cliente_linux_url, ''),
		COALESCE(cliente_mac_url, ''),
		COALESCE(servidor_windows_url, ''),
		COALESCE(servidor_linux_url, ''),
		COALESCE(servidor_mac_url, ''),
		COALESCE(carpeta_transferencia, ''),
		COALESCE(instrucciones_publicas, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_soporte_remoto_configuracion
	WHERE empresa_id = ?
	LIMIT 1`, empresaID).Scan(
		&out.ID,
		&out.EmpresaID,
		&habilitado,
		&out.ProveedorPreferido,
		&out.ModoOperacion,
		&requiereAprobacion,
		&out.AutoCerrarMinutos,
		&out.MaxConexionesMes,
		&out.MaxMinutosMes,
		&out.MaxMinutosDiaRustDesk,
		&out.MaxDispositivos,
		&portalPublico,
		&out.RustDeskServerHost,
		&out.RustDeskServerKey,
		&out.ClienteWindowsURL,
		&out.ClienteLinuxURL,
		&out.ClienteMacURL,
		&out.ServidorWindowsURL,
		&out.ServidorLinuxURL,
		&out.ServidorMacURL,
		&out.CarpetaTransferencia,
		&out.InstruccionesPublicas,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EmpresaSoporteRemotoConfig{
				EmpresaID:                  empresaID,
				Habilitado:                 true,
				ProveedorPreferido:         "rustdesk_oss",
				ModoOperacion:              "cliente_local",
				RequiereAprobacionOperador: true,
				AutoCerrarMinutos:          30,
				MaxConexionesMes:           0,
				MaxMinutosMes:              0,
				MaxMinutosDiaRustDesk:      0,
				MaxDispositivos:            0,
				PortalPublicoHabilitado:    true,
				Estado:                     "activo",
			}, nil
		}
		return EmpresaSoporteRemotoConfig{}, err
	}

	out.Habilitado = habilitado.Valid && habilitado.Int64 > 0
	out.RequiereAprobacionOperador = requiereAprobacion.Valid && requiereAprobacion.Int64 > 0
	out.PortalPublicoHabilitado = !portalPublico.Valid || portalPublico.Int64 > 0
	out.ProveedorPreferido = soporteRemotoNormalizeProveedor(out.ProveedorPreferido)
	out.ModoOperacion = soporteRemotoNormalizeModo(out.ModoOperacion)
	out.AutoCerrarMinutos = soporteRemotoNormalizeAutoCerrar(out.AutoCerrarMinutos)
	out.MaxConexionesMes = soporteRemotoNormalizePlanLimit(out.MaxConexionesMes)
	out.MaxMinutosMes = soporteRemotoNormalizePlanLimit(out.MaxMinutosMes)
	out.MaxMinutosDiaRustDesk = soporteRemotoNormalizePlanLimit(out.MaxMinutosDiaRustDesk)
	out.MaxDispositivos = soporteRemotoNormalizePlanLimit(out.MaxDispositivos)
	out.RustDeskServerHost = strings.TrimSpace(out.RustDeskServerHost)
	out.RustDeskServerKey = strings.TrimSpace(out.RustDeskServerKey)
	out.ClienteWindowsURL = soporteRemotoNormalizeURL(out.ClienteWindowsURL)
	out.ClienteLinuxURL = soporteRemotoNormalizeURL(out.ClienteLinuxURL)
	out.ClienteMacURL = soporteRemotoNormalizeURL(out.ClienteMacURL)
	out.ServidorWindowsURL = soporteRemotoNormalizeURL(out.ServidorWindowsURL)
	out.ServidorLinuxURL = soporteRemotoNormalizeURL(out.ServidorLinuxURL)
	out.ServidorMacURL = soporteRemotoNormalizeURL(out.ServidorMacURL)
	out.CarpetaTransferencia = strings.TrimSpace(out.CarpetaTransferencia)
	out.InstruccionesPublicas = strings.TrimSpace(out.InstruccionesPublicas)
	out.Estado = soporteRemotoNormalizeEstado(out.Estado)
	return out, nil
}

// UpsertEmpresaSoporteRemotoConfig crea o actualiza la configuracion de soporte remoto por empresa.
func UpsertEmpresaSoporteRemotoConfig(dbConn *sql.DB, cfg EmpresaSoporteRemotoConfig) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if cfg.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}

	cfg.ProveedorPreferido = soporteRemotoNormalizeProveedor(cfg.ProveedorPreferido)
	cfg.ModoOperacion = soporteRemotoNormalizeModo(cfg.ModoOperacion)
	cfg.AutoCerrarMinutos = soporteRemotoNormalizeAutoCerrar(cfg.AutoCerrarMinutos)
	cfg.MaxConexionesMes = soporteRemotoNormalizePlanLimit(cfg.MaxConexionesMes)
	cfg.MaxMinutosMes = soporteRemotoNormalizePlanLimit(cfg.MaxMinutosMes)
	cfg.MaxMinutosDiaRustDesk = soporteRemotoNormalizePlanLimit(cfg.MaxMinutosDiaRustDesk)
	cfg.MaxDispositivos = soporteRemotoNormalizePlanLimit(cfg.MaxDispositivos)
	cfg.RustDeskServerHost = strings.TrimSpace(cfg.RustDeskServerHost)
	cfg.RustDeskServerKey = strings.TrimSpace(cfg.RustDeskServerKey)
	cfg.ClienteWindowsURL = soporteRemotoNormalizeURL(cfg.ClienteWindowsURL)
	cfg.ClienteLinuxURL = soporteRemotoNormalizeURL(cfg.ClienteLinuxURL)
	cfg.ClienteMacURL = soporteRemotoNormalizeURL(cfg.ClienteMacURL)
	cfg.ServidorWindowsURL = soporteRemotoNormalizeURL(cfg.ServidorWindowsURL)
	cfg.ServidorLinuxURL = soporteRemotoNormalizeURL(cfg.ServidorLinuxURL)
	cfg.ServidorMacURL = soporteRemotoNormalizeURL(cfg.ServidorMacURL)
	cfg.CarpetaTransferencia = strings.TrimSpace(cfg.CarpetaTransferencia)
	cfg.InstruccionesPublicas = strings.TrimSpace(cfg.InstruccionesPublicas)
	cfg.Estado = soporteRemotoNormalizeEstado(cfg.Estado)
	if cfg.Estado == "" {
		cfg.Estado = "activo"
	}

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_soporte_remoto_configuracion WHERE empresa_id = ? LIMIT 1`, cfg.EmpresaID).Scan(&existingID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if existingID > 0 {
		_, err = dbConn.Exec(`UPDATE empresa_soporte_remoto_configuracion
			SET habilitado = ?,
				proveedor_preferido = ?,
				modo_operacion = ?,
				requiere_aprobacion_operador = ?,
				auto_cerrar_minutos = ?,
				max_conexiones_mes = ?,
				max_minutos_mes = ?,
				max_minutos_dia_rustdesk = ?,
				max_dispositivos = ?,
				portal_publico_habilitado = ?,
				rustdesk_server_host = ?,
				rustdesk_server_key = ?,
				cliente_windows_url = ?,
				cliente_linux_url = ?,
				cliente_mac_url = ?,
				servidor_windows_url = ?,
				servidor_linux_url = ?,
				servidor_mac_url = ?,
				carpeta_transferencia = ?,
				instrucciones_publicas = ?,
				usuario_creador = ?,
				estado = ?,
				observaciones = ?,
				fecha_actualizacion = CURRENT_TIMESTAMP
			WHERE id = ?`,
			soporteRemotoBoolToInt(cfg.Habilitado),
			cfg.ProveedorPreferido,
			cfg.ModoOperacion,
			soporteRemotoBoolToInt(cfg.RequiereAprobacionOperador),
			cfg.AutoCerrarMinutos,
			cfg.MaxConexionesMes,
			cfg.MaxMinutosMes,
			cfg.MaxMinutosDiaRustDesk,
			cfg.MaxDispositivos,
			soporteRemotoBoolToInt(cfg.PortalPublicoHabilitado),
			cfg.RustDeskServerHost,
			cfg.RustDeskServerKey,
			cfg.ClienteWindowsURL,
			cfg.ClienteLinuxURL,
			cfg.ClienteMacURL,
			cfg.ServidorWindowsURL,
			cfg.ServidorLinuxURL,
			cfg.ServidorMacURL,
			cfg.CarpetaTransferencia,
			cfg.InstruccionesPublicas,
			strings.TrimSpace(cfg.UsuarioCreador),
			cfg.Estado,
			strings.TrimSpace(cfg.Observaciones),
			existingID,
		)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_soporte_remoto_configuracion (
		empresa_id,
		habilitado,
		proveedor_preferido,
		modo_operacion,
		requiere_aprobacion_operador,
		auto_cerrar_minutos,
		max_conexiones_mes,
		max_minutos_mes,
		max_minutos_dia_rustdesk,
		max_dispositivos,
		portal_publico_habilitado,
		rustdesk_server_host,
		rustdesk_server_key,
		cliente_windows_url,
		cliente_linux_url,
		cliente_mac_url,
		servidor_windows_url,
		servidor_linux_url,
		servidor_mac_url,
		carpeta_transferencia,
		instrucciones_publicas,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		cfg.EmpresaID,
		soporteRemotoBoolToInt(cfg.Habilitado),
		cfg.ProveedorPreferido,
		cfg.ModoOperacion,
		soporteRemotoBoolToInt(cfg.RequiereAprobacionOperador),
		cfg.AutoCerrarMinutos,
		cfg.MaxConexionesMes,
		cfg.MaxMinutosMes,
		cfg.MaxMinutosDiaRustDesk,
		cfg.MaxDispositivos,
		soporteRemotoBoolToInt(cfg.PortalPublicoHabilitado),
		cfg.RustDeskServerHost,
		cfg.RustDeskServerKey,
		cfg.ClienteWindowsURL,
		cfg.ClienteLinuxURL,
		cfg.ClienteMacURL,
		cfg.ServidorWindowsURL,
		cfg.ServidorLinuxURL,
		cfg.ServidorMacURL,
		cfg.CarpetaTransferencia,
		cfg.InstruccionesPublicas,
		strings.TrimSpace(cfg.UsuarioCreador),
		cfg.Estado,
		strings.TrimSpace(cfg.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// CreateEmpresaSoporteRemotoDispositivo registra un dispositivo remoto por empresa.
func CreateEmpresaSoporteRemotoDispositivo(dbConn *sql.DB, item EmpresaSoporteRemotoDispositivo, accesoPINPlano string) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if item.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	item.NombreEquipo = strings.TrimSpace(item.NombreEquipo)
	if item.NombreEquipo == "" {
		return 0, errors.New("nombre_equipo es obligatorio")
	}
	item.StreamURL = strings.TrimSpace(item.StreamURL)
	item.RustDeskDeviceID = strings.TrimSpace(item.RustDeskDeviceID)
	item.CarpetaTransferencia = strings.TrimSpace(item.CarpetaTransferencia)
	if item.StreamURL == "" && item.RustDeskDeviceID == "" {
		return 0, errors.New("stream_url o rustdesk_device_id es obligatorio")
	}
	item.CodigoDispositivo = strings.TrimSpace(item.CodigoDispositivo)
	if item.CodigoDispositivo == "" {
		item.CodigoDispositivo = soporteRemotoGenerateDeviceCode(item.NombreEquipo)
	}
	item.EstadoConexion = soporteRemotoNormalizeConexion(item.EstadoConexion)
	item.Estado = soporteRemotoNormalizeEstado(item.Estado)
	if item.Estado == "" {
		item.Estado = "activo"
	}

	if cfg, err := GetEmpresaSoporteRemotoConfig(dbConn, item.EmpresaID); err == nil && cfg.MaxDispositivos > 0 {
		activeDevices, countErr := soporteRemotoCountActiveDevices(dbConn, item.EmpresaID)
		if countErr != nil {
			return 0, countErr
		}
		if activeDevices >= int64(cfg.MaxDispositivos) {
			return 0, fmt.Errorf("%w: maximo de dispositivos (%d) alcanzado", ErrSoporteRemotoPlanLimit, cfg.MaxDispositivos)
		}
	}

	accesoPINHash := strings.TrimSpace(item.AccesoPINHash)
	if accesoPINHash == "" {
		accesoPINHash = soporteRemotoHash(accesoPINPlano)
	}
	rustDeskPasswordEnc := strings.TrimSpace(item.RustDeskPasswordEnc)
	if rustDeskPasswordEnc != "" {
		var encErr error
		rustDeskPasswordEnc, encErr = soporteRemotoEncryptSecret(rustDeskPasswordEnc)
		if encErr != nil {
			return 0, encErr
		}
	}

	res, err := dbConn.Exec(`INSERT INTO empresa_soporte_remoto_dispositivos (
		empresa_id,
		codigo_dispositivo,
		nombre_equipo,
		alias_operativo,
		ubicacion,
		sistema_operativo,
		agente_version,
		stream_url,
		rustdesk_device_id,
		rustdesk_password_enc,
		carpeta_transferencia,
		acceso_publico_habilitado,
		estado_conexion,
		ultimo_heartbeat,
		acceso_pin_hash,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.EmpresaID,
		item.CodigoDispositivo,
		item.NombreEquipo,
		strings.TrimSpace(item.AliasOperativo),
		strings.TrimSpace(item.Ubicacion),
		strings.TrimSpace(item.SistemaOperativo),
		strings.TrimSpace(item.AgenteVersion),
		item.StreamURL,
		item.RustDeskDeviceID,
		rustDeskPasswordEnc,
		item.CarpetaTransferencia,
		soporteRemotoBoolToInt(item.AccesoPublicoHabilitado),
		item.EstadoConexion,
		strings.TrimSpace(item.UltimoHeartbeat),
		accesoPINHash,
		strings.TrimSpace(item.UsuarioCreador),
		item.Estado,
		strings.TrimSpace(item.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func soporteRemotoCountActiveDevices(dbConn *sql.DB, empresaID int64) (int64, error) {
	var total int64
	err := dbConn.QueryRow(`SELECT COUNT(1)
		FROM empresa_soporte_remoto_dispositivos
		WHERE empresa_id = ? AND COALESCE(estado, 'activo') <> 'inactivo'`, empresaID).Scan(&total)
	return total, err
}

func GetEmpresaSoporteRemotoUso(dbConn *sql.DB, empresaID int64) (EmpresaSoporteRemotoUso, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoUso{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaSoporteRemotoUso{}, errors.New("empresa_id invalido")
	}

	cfg, err := GetEmpresaSoporteRemotoConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaSoporteRemotoUso{}, err
	}

	uso := EmpresaSoporteRemotoUso{
		EmpresaID:             empresaID,
		MesReferencia:         time.Now().In(time.Local).Format("2006-01"),
		DiaReferencia:         time.Now().In(time.Local).Format("2006-01-02"),
		MaxConexionesMes:      int64(cfg.MaxConexionesMes),
		MaxMinutosMes:         int64(cfg.MaxMinutosMes),
		MaxMinutosDiaRustDesk: int64(cfg.MaxMinutosDiaRustDesk),
		MaxDispositivos:       int64(cfg.MaxDispositivos),
	}

	err = dbConn.QueryRow(`SELECT
		COALESCE(SUM(CASE WHEN COALESCE(estado, 'activo') <> 'inactivo' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(estado, 'activo') <> 'inactivo' AND COALESCE(estado_conexion, 'offline') = 'online' THEN 1 ELSE 0 END), 0)
		FROM empresa_soporte_remoto_dispositivos
		WHERE empresa_id = ?`, empresaID).Scan(&uso.DispositivosActivos, &uso.DispositivosOnline)
	if err != nil {
		return EmpresaSoporteRemotoUso{}, err
	}

	rows, err := dbConn.Query(`SELECT
		COALESCE(s.fecha_creacion, ''),
		COALESCE(s.estado_sesion, 'pendiente'),
		COALESCE(s.iniciada_en, ''),
		COALESCE(s.finalizada_en, ''),
		COALESCE(s.duracion_minutos_solicitada, 0),
		COALESCE(s.duracion_minutos_consumida, 0),
		COALESCE(s.bloqueada_por_limite, 0),
		COALESCE(d.rustdesk_device_id, '')
		FROM empresa_soporte_remoto_sesiones s
		LEFT JOIN empresa_soporte_remoto_dispositivos d ON d.empresa_id = s.empresa_id AND d.id = s.dispositivo_id
		WHERE s.empresa_id = ? AND COALESCE(s.estado, 'activo') <> 'inactivo'`, empresaID)
	if err != nil {
		return EmpresaSoporteRemotoUso{}, err
	}
	defer rows.Close()

	now := time.Now().In(time.Local)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dayEnd := dayStart.Add(24 * time.Hour)
	for rows.Next() {
		var fechaCreacion string
		var estadoSesion string
		var iniciadaEn string
		var finalizadaEn string
		var durSolicitada int
		var durConsumida int
		var bloqueada int64
		var rustDeskDeviceID string
		if err := rows.Scan(&fechaCreacion, &estadoSesion, &iniciadaEn, &finalizadaEn, &durSolicitada, &durConsumida, &bloqueada, &rustDeskDeviceID); err != nil {
			return EmpresaSoporteRemotoUso{}, err
		}
		if !soporteRemotoIsCurrentMonth(fechaCreacion, now) {
			continue
		}
		if bloqueada > 0 {
			uso.IntentosBloqueadosMes++
			continue
		}
		estadoSesion = soporteRemotoNormalizeSesionEstado(estadoSesion)
		if estadoSesion == "rechazada" {
			continue
		}
		uso.SesionesMes++
		minutes := durConsumida
		if minutes <= 0 {
			minutes = soporteRemotoComputeSessionMinutes(iniciadaEn, finalizadaEn, durSolicitada)
		}
		uso.MinutosConsumidosMes += int64(minutes)
		if soporteRemotoUsesRustDesk(cfg, rustDeskDeviceID) {
			uso.MinutosConsumidosDiaRustDesk += int64(soporteRemotoComputeSessionMinutesInWindow(iniciadaEn, finalizadaEn, dayStart, dayEnd))
		}
	}
	if err := rows.Err(); err != nil {
		return EmpresaSoporteRemotoUso{}, err
	}

	uso.PuedeCrearDispositivo = uso.MaxDispositivos <= 0 || uso.DispositivosActivos < uso.MaxDispositivos
	if uso.MaxMinutosDiaRustDesk > 0 {
		uso.MinutosDisponiblesDiaRustDesk = uso.MaxMinutosDiaRustDesk - uso.MinutosConsumidosDiaRustDesk
		if uso.MinutosDisponiblesDiaRustDesk < 0 {
			uso.MinutosDisponiblesDiaRustDesk = 0
		}
	}
	uso.PuedeCrearSesion = true
	if uso.MaxConexionesMes > 0 && uso.SesionesMes >= uso.MaxConexionesMes {
		uso.PuedeCrearSesion = false
		uso.BloqueoMotivo = fmt.Sprintf("maximo de conexiones del mes alcanzado (%d)", uso.MaxConexionesMes)
	}
	if uso.PuedeCrearSesion && uso.MaxMinutosMes > 0 && uso.MinutosConsumidosMes >= uso.MaxMinutosMes {
		uso.PuedeCrearSesion = false
		uso.BloqueoMotivo = fmt.Sprintf("maximo de minutos del mes alcanzado (%d)", uso.MaxMinutosMes)
	}
	if uso.PuedeCrearSesion && soporteRemotoUsesRustDesk(cfg, "") && uso.MaxMinutosDiaRustDesk > 0 && uso.MinutosConsumidosDiaRustDesk >= uso.MaxMinutosDiaRustDesk {
		uso.PuedeCrearSesion = false
		uso.BloqueoMotivo = fmt.Sprintf("maximo de minutos RustDesk del dia alcanzado (%d)", uso.MaxMinutosDiaRustDesk)
	}
	if !uso.PuedeCrearDispositivo && strings.TrimSpace(uso.BloqueoMotivo) == "" {
		uso.BloqueoMotivo = fmt.Sprintf("maximo de dispositivos alcanzado (%d)", uso.MaxDispositivos)
	}

	return uso, nil
}

func validateEmpresaSoporteRemotoSessionMonthlyPlan(dbConn *sql.DB, empresaID int64, duracionMinutos int) (EmpresaSoporteRemotoUso, EmpresaSoporteRemotoConfig, error) {
	cfg, err := GetEmpresaSoporteRemotoConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaSoporteRemotoUso{}, EmpresaSoporteRemotoConfig{}, err
	}
	uso, err := GetEmpresaSoporteRemotoUso(dbConn, empresaID)
	if err != nil {
		return EmpresaSoporteRemotoUso{}, EmpresaSoporteRemotoConfig{}, err
	}
	if cfg.MaxConexionesMes > 0 && uso.SesionesMes >= int64(cfg.MaxConexionesMes) {
		return uso, cfg, fmt.Errorf("%w: maximo de conexiones del mes alcanzado (%d)", ErrSoporteRemotoPlanLimit, cfg.MaxConexionesMes)
	}
	if cfg.MaxMinutosMes > 0 && uso.MinutosConsumidosMes+int64(duracionMinutos) > int64(cfg.MaxMinutosMes) {
		return uso, cfg, fmt.Errorf("%w: maximo de minutos del mes excedido (%d)", ErrSoporteRemotoPlanLimit, cfg.MaxMinutosMes)
	}
	return uso, cfg, nil
}

func validateEmpresaSoporteRemotoSessionDailyRustDeskLimit(dbConn *sql.DB, empresaID, dispositivoID int64, duracionMinutos int) (EmpresaSoporteRemotoUso, EmpresaSoporteRemotoConfig, error) {
	cfg, err := GetEmpresaSoporteRemotoConfig(dbConn, empresaID)
	if err != nil {
		return EmpresaSoporteRemotoUso{}, EmpresaSoporteRemotoConfig{}, err
	}
	device, err := GetEmpresaSoporteRemotoDispositivoByID(dbConn, empresaID, dispositivoID)
	if err != nil {
		return EmpresaSoporteRemotoUso{}, EmpresaSoporteRemotoConfig{}, err
	}
	uso, err := GetEmpresaSoporteRemotoUso(dbConn, empresaID)
	if err != nil {
		return EmpresaSoporteRemotoUso{}, EmpresaSoporteRemotoConfig{}, err
	}
	if !soporteRemotoUsesRustDesk(cfg, device.RustDeskDeviceID) || cfg.MaxMinutosDiaRustDesk <= 0 {
		return uso, cfg, nil
	}
	if uso.MinutosConsumidosDiaRustDesk+int64(duracionMinutos) > int64(cfg.MaxMinutosDiaRustDesk) {
		return uso, cfg, fmt.Errorf("%w: maximo de minutos RustDesk del dia excedido (%d)", ErrSoporteRemotoPlanLimit, cfg.MaxMinutosDiaRustDesk)
	}
	return uso, cfg, nil
}

func createEmpresaSoporteRemotoBlockedAttempt(dbConn *sql.DB, empresaID, dispositivoID int64, solicitadaPor, operadorNombre, operadorEmail, motivo string, duracionMinutos int, observaciones string) {
	_, _ = dbConn.Exec(`INSERT INTO empresa_soporte_remoto_sesiones (
		empresa_id,
		dispositivo_id,
		codigo_sesion,
		solicitada_por,
		operador_nombre,
		operador_email,
		motivo,
		estado_sesion,
		duracion_minutos_solicitada,
		duracion_minutos_consumida,
		bloqueada_por_limite,
		finalizada_en,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, 'rechazada', ?, 0, 1, CURRENT_TIMESTAMP, ?, 'activo', ?)`,
		empresaID,
		dispositivoID,
		soporteRemotoGenerateSessionCode(),
		strings.TrimSpace(solicitadaPor),
		strings.TrimSpace(operadorNombre),
		strings.TrimSpace(operadorEmail),
		strings.TrimSpace(motivo),
		duracionMinutos,
		strings.TrimSpace(solicitadaPor),
		strings.TrimSpace(observaciones),
	)
}

// UpdateEmpresaSoporteRemotoDispositivo actualiza metadata de un dispositivo.
func UpdateEmpresaSoporteRemotoDispositivo(dbConn *sql.DB, item EmpresaSoporteRemotoDispositivo, accesoPINPlano string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if item.ID <= 0 || item.EmpresaID <= 0 {
		return errors.New("id/empresa_id invalidos")
	}
	item.NombreEquipo = strings.TrimSpace(item.NombreEquipo)
	if item.NombreEquipo == "" {
		return errors.New("nombre_equipo es obligatorio")
	}
	item.StreamURL = strings.TrimSpace(item.StreamURL)
	item.RustDeskDeviceID = strings.TrimSpace(item.RustDeskDeviceID)
	item.CarpetaTransferencia = strings.TrimSpace(item.CarpetaTransferencia)
	if item.StreamURL == "" && item.RustDeskDeviceID == "" {
		return errors.New("stream_url o rustdesk_device_id es obligatorio")
	}
	item.CodigoDispositivo = strings.TrimSpace(item.CodigoDispositivo)
	if item.CodigoDispositivo == "" {
		return errors.New("codigo_dispositivo es obligatorio")
	}
	item.EstadoConexion = soporteRemotoNormalizeConexion(item.EstadoConexion)
	item.Estado = soporteRemotoNormalizeEstado(item.Estado)
	if item.Estado == "" {
		item.Estado = "activo"
	}

	accesoPINHash := strings.TrimSpace(item.AccesoPINHash)
	if accesoPINHash == "" {
		accesoPINHash = soporteRemotoHash(accesoPINPlano)
	}
	rustDeskPasswordEnc := strings.TrimSpace(item.RustDeskPasswordEnc)
	if rustDeskPasswordEnc != "" {
		var encErr error
		rustDeskPasswordEnc, encErr = soporteRemotoEncryptSecret(rustDeskPasswordEnc)
		if encErr != nil {
			return encErr
		}
	}

	res, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_dispositivos
		SET codigo_dispositivo = ?,
			nombre_equipo = ?,
			alias_operativo = ?,
			ubicacion = ?,
			sistema_operativo = ?,
			agente_version = ?,
			stream_url = ?,
			rustdesk_device_id = ?,
			rustdesk_password_enc = CASE WHEN ? = '' THEN rustdesk_password_enc ELSE ? END,
			carpeta_transferencia = ?,
			acceso_publico_habilitado = ?,
			estado_conexion = ?,
			ultimo_heartbeat = ?,
			acceso_pin_hash = CASE WHEN ? = '' THEN acceso_pin_hash ELSE ? END,
			usuario_creador = ?,
			estado = ?,
			observaciones = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ? AND empresa_id = ?`,
		item.CodigoDispositivo,
		item.NombreEquipo,
		strings.TrimSpace(item.AliasOperativo),
		strings.TrimSpace(item.Ubicacion),
		strings.TrimSpace(item.SistemaOperativo),
		strings.TrimSpace(item.AgenteVersion),
		item.StreamURL,
		item.RustDeskDeviceID,
		rustDeskPasswordEnc, rustDeskPasswordEnc,
		item.CarpetaTransferencia,
		soporteRemotoBoolToInt(item.AccesoPublicoHabilitado),
		item.EstadoConexion,
		strings.TrimSpace(item.UltimoHeartbeat),
		accesoPINHash,
		accesoPINHash,
		strings.TrimSpace(item.UsuarioCreador),
		item.Estado,
		strings.TrimSpace(item.Observaciones),
		item.ID,
		item.EmpresaID,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected <= 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetEmpresaSoporteRemotoDispositivoEstadoByID activa/desactiva un dispositivo por empresa.
func SetEmpresaSoporteRemotoDispositivoEstadoByID(dbConn *sql.DB, empresaID, deviceID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || deviceID <= 0 {
		return errors.New("empresa_id/id invalidos")
	}
	res, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_dispositivos
		SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ? AND empresa_id = ?`, soporteRemotoNormalizeEstado(estado), deviceID, empresaID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected <= 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetEmpresaSoporteRemotoDispositivoByID obtiene un dispositivo por empresa e ID.
func GetEmpresaSoporteRemotoDispositivoByID(dbConn *sql.DB, empresaID, deviceID int64) (EmpresaSoporteRemotoDispositivo, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 || deviceID <= 0 {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("empresa_id/id invalidos")
	}

	var out EmpresaSoporteRemotoDispositivo
	var publico sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo_dispositivo, ''),
		COALESCE(nombre_equipo, ''),
		COALESCE(alias_operativo, ''),
		COALESCE(ubicacion, ''),
		COALESCE(sistema_operativo, ''),
		COALESCE(agente_version, ''),
		COALESCE(stream_url, ''),
		COALESCE(rustdesk_device_id, ''),
		COALESCE(rustdesk_password_enc, ''),
		COALESCE(carpeta_transferencia, ''),
		COALESCE(acceso_publico_habilitado, 1),
		COALESCE(estado_conexion, 'offline'),
		COALESCE(ultimo_heartbeat, ''),
		COALESCE(acceso_pin_hash, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_soporte_remoto_dispositivos
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, deviceID).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.CodigoDispositivo,
		&out.NombreEquipo,
		&out.AliasOperativo,
		&out.Ubicacion,
		&out.SistemaOperativo,
		&out.AgenteVersion,
		&out.StreamURL,
		&out.RustDeskDeviceID,
		&out.RustDeskPasswordEnc,
		&out.CarpetaTransferencia,
		&publico,
		&out.EstadoConexion,
		&out.UltimoHeartbeat,
		&out.AccesoPINHash,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}
	out.AccesoPublicoHabilitado = !publico.Valid || publico.Int64 > 0
	out.EstadoConexion = soporteRemotoNormalizeConexion(out.EstadoConexion)
	out.Estado = soporteRemotoNormalizeEstado(out.Estado)
	return out, nil
}

// GetEmpresaSoporteRemotoDispositivoByCodigo obtiene un dispositivo por codigo y empresa.
func GetEmpresaSoporteRemotoDispositivoByCodigo(dbConn *sql.DB, empresaID int64, codigoDispositivo string) (EmpresaSoporteRemotoDispositivo, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("empresa_id invalido")
	}
	codigoDispositivo = strings.TrimSpace(codigoDispositivo)
	if codigoDispositivo == "" {
		return EmpresaSoporteRemotoDispositivo{}, errors.New("codigo_dispositivo invalido")
	}

	var out EmpresaSoporteRemotoDispositivo
	var publico sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo_dispositivo, ''),
		COALESCE(nombre_equipo, ''),
		COALESCE(alias_operativo, ''),
		COALESCE(ubicacion, ''),
		COALESCE(sistema_operativo, ''),
		COALESCE(agente_version, ''),
		COALESCE(stream_url, ''),
		COALESCE(rustdesk_device_id, ''),
		COALESCE(rustdesk_password_enc, ''),
		COALESCE(carpeta_transferencia, ''),
		COALESCE(acceso_publico_habilitado, 1),
		COALESCE(estado_conexion, 'offline'),
		COALESCE(ultimo_heartbeat, ''),
		COALESCE(acceso_pin_hash, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_soporte_remoto_dispositivos
	WHERE empresa_id = ? AND trim(codigo_dispositivo) = trim(?)
	LIMIT 1`, empresaID, codigoDispositivo).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.CodigoDispositivo,
		&out.NombreEquipo,
		&out.AliasOperativo,
		&out.Ubicacion,
		&out.SistemaOperativo,
		&out.AgenteVersion,
		&out.StreamURL,
		&out.RustDeskDeviceID,
		&out.RustDeskPasswordEnc,
		&out.CarpetaTransferencia,
		&publico,
		&out.EstadoConexion,
		&out.UltimoHeartbeat,
		&out.AccesoPINHash,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}
	out.AccesoPublicoHabilitado = !publico.Valid || publico.Int64 > 0
	out.EstadoConexion = soporteRemotoNormalizeConexion(out.EstadoConexion)
	out.Estado = soporteRemotoNormalizeEstado(out.Estado)
	return out, nil
}

// ValidateEmpresaSoporteRemotoDispositivoAccess valida PIN de acceso del dispositivo.
func ValidateEmpresaSoporteRemotoDispositivoAccess(dbConn *sql.DB, empresaID int64, codigoDispositivo, accesoPIN string) (EmpresaSoporteRemotoDispositivo, error) {
	item, err := GetEmpresaSoporteRemotoDispositivoByCodigo(dbConn, empresaID, codigoDispositivo)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}
	if strings.EqualFold(item.Estado, "inactivo") {
		return EmpresaSoporteRemotoDispositivo{}, sql.ErrNoRows
	}
	hash := strings.TrimSpace(item.AccesoPINHash)
	if hash != "" {
		if soporteRemotoHash(accesoPIN) != hash {
			return EmpresaSoporteRemotoDispositivo{}, sql.ErrNoRows
		}
	}
	return item, nil
}

// RegisterEmpresaSoporteRemotoDispositivoHeartbeat actualiza latido/estado de conexion para un dispositivo.
func RegisterEmpresaSoporteRemotoDispositivoHeartbeat(dbConn *sql.DB, empresaID int64, codigoDispositivo, accesoPIN, streamURL, sistemaOperativo, agenteVersion string) (EmpresaSoporteRemotoDispositivo, error) {
	device, err := ValidateEmpresaSoporteRemotoDispositivoAccess(dbConn, empresaID, codigoDispositivo, accesoPIN)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}

	stream := strings.TrimSpace(streamURL)
	if stream == "" {
		stream = device.StreamURL
	}

	_, err = dbConn.Exec(`UPDATE empresa_soporte_remoto_dispositivos
		SET stream_url = ?,
			sistema_operativo = CASE WHEN ? = '' THEN sistema_operativo ELSE ? END,
			agente_version = CASE WHEN ? = '' THEN agente_version ELSE ? END,
			estado_conexion = 'online',
			ultimo_heartbeat = CURRENT_TIMESTAMP,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE id = ? AND empresa_id = ?`,
		stream,
		strings.TrimSpace(sistemaOperativo), strings.TrimSpace(sistemaOperativo),
		strings.TrimSpace(agenteVersion), strings.TrimSpace(agenteVersion),
		device.ID,
		empresaID,
	)
	if err != nil {
		return EmpresaSoporteRemotoDispositivo{}, err
	}

	return GetEmpresaSoporteRemotoDispositivoByID(dbConn, empresaID, device.ID)
}

// ListEmpresaSoporteRemotoDispositivos lista dispositivos remotos por empresa.
func ListEmpresaSoporteRemotoDispositivos(dbConn *sql.DB, empresaID int64, filter EmpresaSoporteRemotoDispositivoFilter) ([]EmpresaSoporteRemotoDispositivo, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	limit, offset := soporteRemotoNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := "WHERE empresa_id = ?"
	args := make([]interface{}, 0)
	args = append(args, empresaID)

	if !filter.IncludeInactive {
		where += " AND COALESCE(estado, 'activo') <> 'inactivo'"
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := soporteRemotoLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(codigo_dispositivo, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(nombre_equipo, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(alias_operativo, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(ubicacion, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern, pattern)
	}

	var total int64
	if err := dbConn.QueryRow("SELECT COUNT(1) FROM empresa_soporte_remoto_dispositivos "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := dbConn.Query(`SELECT
		id,
		empresa_id,
		COALESCE(codigo_dispositivo, ''),
		COALESCE(nombre_equipo, ''),
		COALESCE(alias_operativo, ''),
		COALESCE(ubicacion, ''),
		COALESCE(sistema_operativo, ''),
		COALESCE(agente_version, ''),
		COALESCE(stream_url, ''),
		COALESCE(rustdesk_device_id, ''),
		COALESCE(rustdesk_password_enc, ''),
		COALESCE(carpeta_transferencia, ''),
		COALESCE(acceso_publico_habilitado, 1),
		COALESCE(estado_conexion, 'offline'),
		COALESCE(ultimo_heartbeat, ''),
		COALESCE(acceso_pin_hash, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_soporte_remoto_dispositivos `+where+`
	ORDER BY id DESC
	LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaSoporteRemotoDispositivo, 0)
	for rows.Next() {
		var item EmpresaSoporteRemotoDispositivo
		var publico sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CodigoDispositivo,
			&item.NombreEquipo,
			&item.AliasOperativo,
			&item.Ubicacion,
			&item.SistemaOperativo,
			&item.AgenteVersion,
			&item.StreamURL,
			&item.RustDeskDeviceID,
			&item.RustDeskPasswordEnc,
			&item.CarpetaTransferencia,
			&publico,
			&item.EstadoConexion,
			&item.UltimoHeartbeat,
			&item.AccesoPINHash,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, err
		}
		item.AccesoPublicoHabilitado = !publico.Valid || publico.Int64 > 0
		item.EstadoConexion = soporteRemotoNormalizeConexion(item.EstadoConexion)
		item.Estado = soporteRemotoNormalizeEstado(item.Estado)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// CreateEmpresaSoporteRemotoSession crea una sesion de visualizacion remota y retorna token de vista.
func CreateEmpresaSoporteRemotoSession(dbConn *sql.DB, empresaID, dispositivoID int64, solicitadaPor, operadorNombre, operadorEmail, motivo string, duracionMinutos int, requiereAprobacion bool) (EmpresaSoporteRemotoSession, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoSession{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 || dispositivoID <= 0 {
		return EmpresaSoporteRemotoSession{}, errors.New("empresa_id/dispositivo_id invalidos")
	}

	device, err := GetEmpresaSoporteRemotoDispositivoByID(dbConn, empresaID, dispositivoID)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	if strings.EqualFold(device.Estado, "inactivo") {
		return EmpresaSoporteRemotoSession{}, errors.New("dispositivo inactivo")
	}
	if strings.TrimSpace(device.StreamURL) == "" && strings.TrimSpace(device.RustDeskDeviceID) == "" {
		return EmpresaSoporteRemotoSession{}, errors.New("dispositivo sin stream_url ni rustdesk_device_id")
	}

	duracion := soporteRemotoNormalizeAutoCerrar(duracionMinutos)
	if _, _, err := validateEmpresaSoporteRemotoSessionMonthlyPlan(dbConn, empresaID, duracion); err != nil {
		if errors.Is(err, ErrSoporteRemotoPlanLimit) {
			createEmpresaSoporteRemotoBlockedAttempt(dbConn, empresaID, dispositivoID, solicitadaPor, operadorNombre, operadorEmail, motivo, duracion, err.Error())
		}
		return EmpresaSoporteRemotoSession{}, err
	}
	if !requiereAprobacion {
		if _, _, err := validateEmpresaSoporteRemotoSessionDailyRustDeskLimit(dbConn, empresaID, dispositivoID, duracion); err != nil {
			if errors.Is(err, ErrSoporteRemotoPlanLimit) {
				createEmpresaSoporteRemotoBlockedAttempt(dbConn, empresaID, dispositivoID, solicitadaPor, operadorNombre, operadorEmail, motivo, duracion, err.Error())
			}
			return EmpresaSoporteRemotoSession{}, err
		}
	}
	estadoSesion := "activa"
	iniciadaEn := time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	if requiereAprobacion {
		estadoSesion = "pendiente"
		iniciadaEn = ""
	}
	expiraEn := time.Now().In(time.Local).Add(time.Duration(duracion) * time.Minute).Format("2006-01-02 15:04:05")
	codigoSesion := soporteRemotoGenerateSessionCode()
	tokenRaw, err := soporteRemotoGenerateSecureSecret("SRV")
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	tokenHash := soporteRemotoHash(tokenRaw)

	res, err := dbConn.Exec(`INSERT INTO empresa_soporte_remoto_sesiones (
		empresa_id,
		dispositivo_id,
		codigo_sesion,
		solicitada_por,
		operador_nombre,
		operador_email,
		motivo,
		estado_sesion,
		duracion_minutos_solicitada,
		duracion_minutos_consumida,
		bloqueada_por_limite,
		token_visualizacion_hash,
		url_visualizacion,
		iniciada_en,
		expira_en,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, ?, ?, ?, ?, ?, 'activo', ?)`,
		empresaID,
		dispositivoID,
		codigoSesion,
		strings.TrimSpace(solicitadaPor),
		strings.TrimSpace(operadorNombre),
		strings.TrimSpace(operadorEmail),
		strings.TrimSpace(motivo),
		estadoSesion,
		duracion,
		tokenHash,
		strings.TrimSpace(device.StreamURL),
		iniciadaEn,
		expiraEn,
		strings.TrimSpace(solicitadaPor),
		"sesion creada por panel",
	)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}

	session, err := GetEmpresaSoporteRemotoSessionByID(dbConn, empresaID, id)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	session.TokenVisualizacionRaw = tokenRaw
	return session, nil
}

// GetEmpresaSoporteRemotoSessionByID consulta una sesion por ID.
func GetEmpresaSoporteRemotoSessionByID(dbConn *sql.DB, empresaID, sessionID int64) (EmpresaSoporteRemotoSession, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoSession{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 || sessionID <= 0 {
		return EmpresaSoporteRemotoSession{}, errors.New("empresa_id/id invalidos")
	}

	var out EmpresaSoporteRemotoSession
	var blocked sql.NullInt64
	err := dbConn.QueryRow(`SELECT
		s.id,
		s.empresa_id,
		s.dispositivo_id,
		COALESCE(d.codigo_dispositivo, ''),
		COALESCE(d.nombre_equipo, ''),
		COALESCE(s.codigo_sesion, ''),
		COALESCE(s.solicitada_por, ''),
		COALESCE(s.operador_nombre, ''),
		COALESCE(s.operador_email, ''),
		COALESCE(s.motivo, ''),
		COALESCE(s.estado_sesion, 'pendiente'),
		COALESCE(s.duracion_minutos_solicitada, 0),
		COALESCE(s.duracion_minutos_consumida, 0),
		COALESCE(s.bloqueada_por_limite, 0),
		COALESCE(s.token_visualizacion_hash, ''),
		COALESCE(s.url_visualizacion, ''),
		COALESCE(s.iniciada_en, ''),
		COALESCE(s.expira_en, ''),
		COALESCE(s.finalizada_en, ''),
		COALESCE(s.fecha_creacion, ''),
		COALESCE(s.fecha_actualizacion, ''),
		COALESCE(s.usuario_creador, ''),
		COALESCE(s.estado, 'activo'),
		COALESCE(s.observaciones, '')
	FROM empresa_soporte_remoto_sesiones s
	LEFT JOIN empresa_soporte_remoto_dispositivos d ON d.id = s.dispositivo_id
	WHERE s.empresa_id = ? AND s.id = ?
	LIMIT 1`, empresaID, sessionID).Scan(
		&out.ID,
		&out.EmpresaID,
		&out.DispositivoID,
		&out.DispositivoCodigo,
		&out.DispositivoNombre,
		&out.CodigoSesion,
		&out.SolicitadaPor,
		&out.OperadorNombre,
		&out.OperadorEmail,
		&out.Motivo,
		&out.EstadoSesion,
		&out.DuracionMinSolicitada,
		&out.DuracionMinConsumida,
		&blocked,
		&out.tokenVisualizacionHash,
		&out.URLVisualizacion,
		&out.IniciadaEn,
		&out.ExpiraEn,
		&out.FinalizadaEn,
		&out.FechaCreacion,
		&out.FechaActualizacion,
		&out.UsuarioCreador,
		&out.Estado,
		&out.Observaciones,
	)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	out.BloqueadaPorLimite = blocked.Valid && blocked.Int64 > 0
	out.EstadoSesion = soporteRemotoNormalizeSesionEstado(out.EstadoSesion)
	out.Estado = soporteRemotoNormalizeEstado(out.Estado)
	return out, nil
}

// GetEmpresaSoporteRemotoSessionByCodigo consulta una sesion por codigo.
func GetEmpresaSoporteRemotoSessionByCodigo(dbConn *sql.DB, empresaID int64, codigoSesion string) (EmpresaSoporteRemotoSession, error) {
	if dbConn == nil {
		return EmpresaSoporteRemotoSession{}, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return EmpresaSoporteRemotoSession{}, errors.New("empresa_id invalido")
	}
	codigoSesion = strings.TrimSpace(codigoSesion)
	if codigoSesion == "" {
		return EmpresaSoporteRemotoSession{}, errors.New("codigo_sesion invalido")
	}

	var sessionID int64
	if err := dbConn.QueryRow(`SELECT id FROM empresa_soporte_remoto_sesiones WHERE empresa_id = ? AND trim(codigo_sesion) = trim(?) LIMIT 1`, empresaID, codigoSesion).Scan(&sessionID); err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	return GetEmpresaSoporteRemotoSessionByID(dbConn, empresaID, sessionID)
}

// ResolveEmpresaSoporteRemotoViewerSession valida token de vista para una sesion.
func ResolveEmpresaSoporteRemotoViewerSession(dbConn *sql.DB, empresaID int64, codigoSesion, tokenVisualizacion string) (EmpresaSoporteRemotoSession, error) {
	session, err := GetEmpresaSoporteRemotoSessionByCodigo(dbConn, empresaID, codigoSesion)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	if strings.EqualFold(session.Estado, "inactivo") {
		return EmpresaSoporteRemotoSession{}, sql.ErrNoRows
	}
	if strings.TrimSpace(session.tokenVisualizacionHash) == "" {
		return EmpresaSoporteRemotoSession{}, sql.ErrNoRows
	}
	if !soporteRemotoHashEqual(tokenVisualizacion, session.tokenVisualizacionHash) {
		return EmpresaSoporteRemotoSession{}, sql.ErrNoRows
	}
	if expiry, ok := soporteRemotoParseDateTime(session.ExpiraEn); ok && !time.Now().Before(expiry) {
		_ = SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbConn, empresaID, session.CodigoSesion, "expirada", "sesion expirada automaticamente")
		return EmpresaSoporteRemotoSession{}, sql.ErrNoRows
	}
	usedAt := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_sesiones
		SET token_visualizacion_usado_en = ?, fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ? AND COALESCE(token_visualizacion_usado_en, '') = ''`,
		usedAt, empresaID, session.ID)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return EmpresaSoporteRemotoSession{}, sql.ErrNoRows
	}

	return session, nil
}

func soporteRemotoNormalizeSignalingRole(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "host":
		return "host"
	case "viewer":
		return "viewer"
	default:
		return ""
	}
}

// CreateEmpresaSoporteRemotoSignalingCredential crea una credencial corta y
// de un solo uso para abrir un WebSocket de senalizacion. Invalida cualquier
// credencial anterior no consumida para la misma sesion y rol.
func CreateEmpresaSoporteRemotoSignalingCredential(dbConn *sql.DB, empresaID int64, codigoSesion, role, actor string) (EmpresaSoporteRemotoSignalingCredential, error) {
	if dbConn == nil || empresaID <= 0 {
		return EmpresaSoporteRemotoSignalingCredential{}, ErrSoporteRemotoSignalingCredential
	}
	role = soporteRemotoNormalizeSignalingRole(role)
	if role == "" {
		return EmpresaSoporteRemotoSignalingCredential{}, ErrSoporteRemotoSignalingCredential
	}
	session, err := GetEmpresaSoporteRemotoSessionByCodigo(dbConn, empresaID, codigoSesion)
	if err != nil || (session.EstadoSesion != "activa" && session.EstadoSesion != "aprobada") || strings.EqualFold(session.Estado, "inactivo") {
		return EmpresaSoporteRemotoSignalingCredential{}, ErrSoporteRemotoSignalingCredential
	}
	if expiry, ok := soporteRemotoParseDateTime(session.ExpiraEn); ok && !time.Now().Before(expiry) {
		return EmpresaSoporteRemotoSignalingCredential{}, ErrSoporteRemotoSignalingCredential
	}

	tokenRaw, err := soporteRemotoGenerateSecureSecret("SRSIG")
	if err != nil {
		return EmpresaSoporteRemotoSignalingCredential{}, err
	}
	nonceRaw, err := soporteRemotoGenerateSecureSecret("NONCE")
	if err != nil {
		return EmpresaSoporteRemotoSignalingCredential{}, err
	}
	expiresAt := time.Now().UTC().Add(2 * time.Minute).Format(time.RFC3339Nano)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	tx, err := dbConn.Begin()
	if err != nil {
		return EmpresaSoporteRemotoSignalingCredential{}, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`UPDATE empresa_soporte_remoto_signaling_tokens
		SET revocado_en = ?
		WHERE empresa_id = ? AND sesion_id = ? AND role = ?
			AND COALESCE(usado_en, '') = '' AND COALESCE(revocado_en, '') = ''`,
		now, empresaID, session.ID, role); err != nil {
		return EmpresaSoporteRemotoSignalingCredential{}, err
	}
	if _, err := tx.Exec(`INSERT INTO empresa_soporte_remoto_signaling_tokens (
		empresa_id, sesion_id, role, token_hash, nonce_hash, expira_en, creado_por
	) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		empresaID, session.ID, role, soporteRemotoHash(tokenRaw), soporteRemotoHash(nonceRaw), expiresAt, strings.TrimSpace(actor)); err != nil {
		return EmpresaSoporteRemotoSignalingCredential{}, err
	}
	if err := tx.Commit(); err != nil {
		return EmpresaSoporteRemotoSignalingCredential{}, err
	}
	return EmpresaSoporteRemotoSignalingCredential{TokenRaw: tokenRaw, NonceRaw: nonceRaw, Role: role, ExpiresAt: expiresAt}, nil
}

// ConsumeEmpresaSoporteRemotoSignalingCredential valida y marca la credencial
// como usada de forma atomica. Un segundo intento siempre falla.
func ConsumeEmpresaSoporteRemotoSignalingCredential(dbConn *sql.DB, empresaID int64, codigoSesion, role, tokenRaw, nonceRaw string) (EmpresaSoporteRemotoSession, error) {
	if dbConn == nil || empresaID <= 0 {
		return EmpresaSoporteRemotoSession{}, ErrSoporteRemotoSignalingCredential
	}
	role = soporteRemotoNormalizeSignalingRole(role)
	if role == "" || strings.TrimSpace(tokenRaw) == "" || strings.TrimSpace(nonceRaw) == "" {
		return EmpresaSoporteRemotoSession{}, ErrSoporteRemotoSignalingCredential
	}
	session, err := GetEmpresaSoporteRemotoSessionByCodigo(dbConn, empresaID, codigoSesion)
	if err != nil || (session.EstadoSesion != "activa" && session.EstadoSesion != "aprobada") || strings.EqualFold(session.Estado, "inactivo") {
		return EmpresaSoporteRemotoSession{}, ErrSoporteRemotoSignalingCredential
	}

	var credentialID int64
	var tokenHash, nonceHash, expiresAt string
	err = dbConn.QueryRow(`SELECT id, token_hash, nonce_hash, expira_en
		FROM empresa_soporte_remoto_signaling_tokens
		WHERE empresa_id = ? AND sesion_id = ? AND role = ?
			AND COALESCE(usado_en, '') = '' AND COALESCE(revocado_en, '') = ''
		ORDER BY id DESC LIMIT 1`, empresaID, session.ID, role).Scan(&credentialID, &tokenHash, &nonceHash, &expiresAt)
	if err != nil || !soporteRemotoHashEqual(tokenRaw, tokenHash) || !soporteRemotoHashEqual(nonceRaw, nonceHash) {
		return EmpresaSoporteRemotoSession{}, ErrSoporteRemotoSignalingCredential
	}
	expires, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(expiresAt))
	if err != nil || !time.Now().UTC().Before(expires) {
		return EmpresaSoporteRemotoSession{}, ErrSoporteRemotoSignalingCredential
	}
	usedAt := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_signaling_tokens
		SET usado_en = ?
		WHERE id = ? AND empresa_id = ? AND COALESCE(usado_en, '') = '' AND COALESCE(revocado_en, '') = ''`,
		usedAt, credentialID, empresaID)
	if err != nil {
		return EmpresaSoporteRemotoSession{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return EmpresaSoporteRemotoSession{}, ErrSoporteRemotoSignalingCredential
	}
	return session, nil
}

func IsEmpresaSoporteRemotoSessionActive(dbConn *sql.DB, empresaID, sessionID int64) bool {
	if dbConn == nil || empresaID <= 0 || sessionID <= 0 {
		return false
	}
	var state, status, expiresAt string
	if err := dbConn.QueryRow(`SELECT COALESCE(estado_sesion, ''), COALESCE(estado, ''), COALESCE(expira_en, '')
		FROM empresa_soporte_remoto_sesiones WHERE empresa_id = ? AND id = ? LIMIT 1`, empresaID, sessionID).Scan(&state, &status, &expiresAt); err != nil {
		return false
	}
	if strings.EqualFold(status, "inactivo") || (state != "activa" && state != "aprobada") {
		return false
	}
	if expiry, ok := soporteRemotoParseDateTime(expiresAt); ok && !time.Now().Before(expiry) {
		return false
	}
	return true
}

func RevokeEmpresaSoporteRemotoSignalingCredentials(dbConn *sql.DB, empresaID, sessionID int64) error {
	if dbConn == nil || empresaID <= 0 || sessionID <= 0 {
		return nil
	}
	_, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_signaling_tokens
		SET revocado_en = ?
		WHERE empresa_id = ? AND sesion_id = ? AND COALESCE(revocado_en, '') = ''`,
		time.Now().UTC().Format(time.RFC3339Nano), empresaID, sessionID)
	return err
}

// SetEmpresaSoporteRemotoSessionEstadoByCodigo actualiza estado de una sesion por codigo.
func SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbConn *sql.DB, empresaID int64, codigoSesion, estadoSesion, observaciones string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return errors.New("empresa_id invalido")
	}
	codigoSesion = strings.TrimSpace(codigoSesion)
	if codigoSesion == "" {
		return errors.New("codigo_sesion invalido")
	}

	estado := soporteRemotoNormalizeSesionEstado(estadoSesion)
	current, err := GetEmpresaSoporteRemotoSessionByCodigo(dbConn, empresaID, codigoSesion)
	if err != nil {
		return err
	}
	if (estado == "activa" || estado == "aprobada") && strings.TrimSpace(current.IniciadaEn) == "" && !current.BloqueadaPorLimite {
		if _, _, err := validateEmpresaSoporteRemotoSessionDailyRustDeskLimit(dbConn, empresaID, current.DispositivoID, current.DuracionMinSolicitada); err != nil {
			return err
		}
	}
	iniciadaEn := ""
	finalizadaEn := ""
	duracionConsumida := current.DuracionMinConsumida
	if (estado == "activa" || estado == "aprobada") && strings.TrimSpace(current.IniciadaEn) == "" {
		iniciadaEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	}
	if estado == "finalizada" || estado == "rechazada" || estado == "expirada" {
		finalizadaEn = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
		if !current.BloqueadaPorLimite {
			baseInicio := current.IniciadaEn
			if strings.TrimSpace(baseInicio) == "" {
				baseInicio = iniciadaEn
			}
			duracionConsumida = soporteRemotoComputeSessionMinutes(baseInicio, finalizadaEn, current.DuracionMinSolicitada)
		}
	}

	res, err := dbConn.Exec(`UPDATE empresa_soporte_remoto_sesiones
		SET estado_sesion = ?,
			iniciada_en = CASE WHEN ? = '' THEN iniciada_en ELSE ? END,
			finalizada_en = CASE WHEN ? = '' THEN finalizada_en ELSE ? END,
			duracion_minutos_consumida = CASE WHEN ? <= 0 THEN duracion_minutos_consumida ELSE ? END,
			observaciones = CASE WHEN ? = '' THEN observaciones ELSE ? END,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND trim(codigo_sesion) = trim(?)`,
		estado,
		iniciadaEn, iniciadaEn,
		finalizadaEn, finalizadaEn,
		duracionConsumida, duracionConsumida,
		strings.TrimSpace(observaciones), strings.TrimSpace(observaciones),
		empresaID,
		codigoSesion,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected <= 0 {
		return sql.ErrNoRows
	}
	if estado == "finalizada" || estado == "rechazada" || estado == "expirada" {
		if err := RevokeEmpresaSoporteRemotoSignalingCredentials(dbConn, empresaID, current.ID); err != nil {
			return err
		}
	}
	return nil
}

// ListEmpresaSoporteRemotoSesiones lista sesiones de soporte remoto por empresa.
func ListEmpresaSoporteRemotoSesiones(dbConn *sql.DB, empresaID int64, filter EmpresaSoporteRemotoSessionFilter) ([]EmpresaSoporteRemotoSession, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	limit, offset := soporteRemotoNormalizeLimitOffset(filter.Limit, filter.Offset)
	where := "WHERE s.empresa_id = ?"
	args := make([]interface{}, 0)
	args = append(args, empresaID)

	if !filter.IncludeInactive {
		where += " AND COALESCE(s.estado, 'activo') <> 'inactivo'"
	}
	if estado := strings.TrimSpace(strings.ToLower(filter.EstadoSesion)); estado != "" {
		where += " AND LOWER(COALESCE(s.estado_sesion, '')) = ?"
		args = append(args, estado)
	}
	if q := strings.TrimSpace(filter.Q); q != "" {
		pattern := soporteRemotoLikePattern(q)
		where += ` AND (
			LOWER(COALESCE(s.codigo_sesion, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(s.solicitada_por, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(s.operador_nombre, '')) LIKE LOWER(?) ESCAPE '!' OR
			LOWER(COALESCE(d.nombre_equipo, '')) LIKE LOWER(?) ESCAPE '!'
		)`
		args = append(args, pattern, pattern, pattern, pattern)
	}

	var total int64
	if err := dbConn.QueryRow(`SELECT COUNT(1)
		FROM empresa_soporte_remoto_sesiones s
		LEFT JOIN empresa_soporte_remoto_dispositivos d ON d.id = s.dispositivo_id `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := dbConn.Query(`SELECT
		s.id,
		s.empresa_id,
		s.dispositivo_id,
		COALESCE(d.codigo_dispositivo, ''),
		COALESCE(d.nombre_equipo, ''),
		COALESCE(s.codigo_sesion, ''),
		COALESCE(s.solicitada_por, ''),
		COALESCE(s.operador_nombre, ''),
		COALESCE(s.operador_email, ''),
		COALESCE(s.motivo, ''),
		COALESCE(s.estado_sesion, 'pendiente'),
		COALESCE(s.duracion_minutos_solicitada, 0),
		COALESCE(s.duracion_minutos_consumida, 0),
		COALESCE(s.bloqueada_por_limite, 0),
		COALESCE(s.token_visualizacion_hash, ''),
		COALESCE(s.url_visualizacion, ''),
		COALESCE(s.iniciada_en, ''),
		COALESCE(s.expira_en, ''),
		COALESCE(s.finalizada_en, ''),
		COALESCE(s.fecha_creacion, ''),
		COALESCE(s.fecha_actualizacion, ''),
		COALESCE(s.usuario_creador, ''),
		COALESCE(s.estado, 'activo'),
		COALESCE(s.observaciones, '')
	FROM empresa_soporte_remoto_sesiones s
	LEFT JOIN empresa_soporte_remoto_dispositivos d ON d.id = s.dispositivo_id `+where+`
	ORDER BY s.id DESC
	LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaSoporteRemotoSession, 0)
	for rows.Next() {
		var item EmpresaSoporteRemotoSession
		var blocked sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.DispositivoID,
			&item.DispositivoCodigo,
			&item.DispositivoNombre,
			&item.CodigoSesion,
			&item.SolicitadaPor,
			&item.OperadorNombre,
			&item.OperadorEmail,
			&item.Motivo,
			&item.EstadoSesion,
			&item.DuracionMinSolicitada,
			&item.DuracionMinConsumida,
			&blocked,
			&item.tokenVisualizacionHash,
			&item.URLVisualizacion,
			&item.IniciadaEn,
			&item.ExpiraEn,
			&item.FinalizadaEn,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, 0, err
		}
		item.BloqueadaPorLimite = blocked.Valid && blocked.Int64 > 0
		item.EstadoSesion = soporteRemotoNormalizeSesionEstado(item.EstadoSesion)
		item.Estado = soporteRemotoNormalizeEstado(item.Estado)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
