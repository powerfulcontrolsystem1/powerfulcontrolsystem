package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const SuperCorreoNotificacionTipoSistemaAlerta = "alerta_sistema_super"

type SuperAlertaConfig struct {
	ID                     int64   `json:"id"`
	Enabled                bool    `json:"enabled"`
	RecipientEmail         string  `json:"recipient_email"`
	DiskEnabled            bool    `json:"disk_enabled"`
	DiskThresholdPct       float64 `json:"disk_threshold_pct"`
	TrafficEnabled         bool    `json:"traffic_enabled"`
	TrafficThresholdPct    float64 `json:"traffic_threshold_pct"`
	TrafficThresholdGB     float64 `json:"traffic_threshold_gb"`
	SessionsEnabled        bool    `json:"sessions_enabled"`
	SessionsThreshold      int64   `json:"sessions_threshold"`
	DBConnectionsEnabled   bool    `json:"db_connections_enabled"`
	DBConnectionsThreshold int64   `json:"db_connections_threshold"`
	CooldownMinutes        int64   `json:"cooldown_minutes"`
	FechaCreacion          string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion     string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador         string  `json:"usuario_creador,omitempty"`
	Estado                 string  `json:"estado,omitempty"`
	Observaciones          string  `json:"observaciones,omitempty"`
}

type SuperAlertaEvento struct {
	ID                 int64   `json:"id"`
	Tipo               string  `json:"tipo"`
	Severidad          string  `json:"severidad"`
	Titulo             string  `json:"titulo"`
	Detalle            string  `json:"detalle"`
	Valor              float64 `json:"valor"`
	Umbral             float64 `json:"umbral"`
	Unidad             string  `json:"unidad"`
	Destinatario       string  `json:"destinatario"`
	Asunto             string  `json:"asunto"`
	Cuerpo             string  `json:"cuerpo"`
	CorreoEnviado      bool    `json:"correo_enviado"`
	CorreoError        string  `json:"correo_error,omitempty"`
	MetadataJSON       string  `json:"metadata_json,omitempty"`
	FechaEvento        string  `json:"fecha_evento,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

func defaultSuperAlertaConfig() SuperAlertaConfig {
	return SuperAlertaConfig{
		ID:                     1,
		Enabled:                true,
		RecipientEmail:         "powerfulcontrolsystem@gmail.com",
		DiskEnabled:            true,
		DiskThresholdPct:       80,
		TrafficEnabled:         true,
		TrafficThresholdPct:    80,
		TrafficThresholdGB:     0,
		SessionsEnabled:        true,
		SessionsThreshold:      50,
		DBConnectionsEnabled:   true,
		DBConnectionsThreshold: 80,
		CooldownMinutes:        60,
		UsuarioCreador:         "sistema",
		Estado:                 "activo",
		Observaciones:          "configuracion_default_alertas_sistema",
	}
}

func EnsureSuperAlertasSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db connection is required")
	}

	var configCreate string
	var eventosCreate string
	if isPostgresDialect() {
		configCreate = `CREATE TABLE IF NOT EXISTS super_alertas_config (
			id INTEGER PRIMARY KEY,
			enabled INTEGER DEFAULT 1,
			recipient_email TEXT DEFAULT 'powerfulcontrolsystem@gmail.com',
			disk_enabled INTEGER DEFAULT 1,
			disk_threshold_pct DOUBLE PRECISION DEFAULT 80,
			traffic_enabled INTEGER DEFAULT 1,
			traffic_threshold_pct DOUBLE PRECISION DEFAULT 80,
			traffic_threshold_gb DOUBLE PRECISION DEFAULT 0,
			sessions_enabled INTEGER DEFAULT 1,
			sessions_threshold BIGINT DEFAULT 50,
			db_connections_enabled INTEGER DEFAULT 1,
			db_connections_threshold BIGINT DEFAULT 80,
			cooldown_minutes BIGINT DEFAULT 60,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`
		eventosCreate = `CREATE TABLE IF NOT EXISTS super_alertas_eventos (
			id BIGSERIAL PRIMARY KEY,
			tipo TEXT NOT NULL,
			severidad TEXT DEFAULT 'warning',
			titulo TEXT,
			detalle TEXT,
			valor DOUBLE PRECISION DEFAULT 0,
			umbral DOUBLE PRECISION DEFAULT 0,
			unidad TEXT,
			destinatario TEXT,
			asunto TEXT,
			cuerpo TEXT,
			correo_enviado INTEGER DEFAULT 0,
			correo_error TEXT,
			metadata_json TEXT,
			fecha_evento TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`
	} else {
		configCreate = `CREATE TABLE IF NOT EXISTS super_alertas_config (
			id INTEGER PRIMARY KEY,
			enabled INTEGER DEFAULT 1,
			recipient_email TEXT DEFAULT 'powerfulcontrolsystem@gmail.com',
			disk_enabled INTEGER DEFAULT 1,
			disk_threshold_pct REAL DEFAULT 80,
			traffic_enabled INTEGER DEFAULT 1,
			traffic_threshold_pct REAL DEFAULT 80,
			traffic_threshold_gb REAL DEFAULT 0,
			sessions_enabled INTEGER DEFAULT 1,
			sessions_threshold INTEGER DEFAULT 50,
			db_connections_enabled INTEGER DEFAULT 1,
			db_connections_threshold INTEGER DEFAULT 80,
			cooldown_minutes INTEGER DEFAULT 60,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`
		eventosCreate = `CREATE TABLE IF NOT EXISTS super_alertas_eventos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tipo TEXT NOT NULL,
			severidad TEXT DEFAULT 'warning',
			titulo TEXT,
			detalle TEXT,
			valor REAL DEFAULT 0,
			umbral REAL DEFAULT 0,
			unidad TEXT,
			destinatario TEXT,
			asunto TEXT,
			cuerpo TEXT,
			correo_enviado INTEGER DEFAULT 0,
			correo_error TEXT,
			metadata_json TEXT,
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`
	}

	if _, err := ExecCompat(dbConn, configCreate); err != nil {
		return err
	}
	if _, err := ExecCompat(dbConn, eventosCreate); err != nil {
		return err
	}

	configCols := []struct {
		name string
		def  string
	}{
		{"enabled", "INTEGER DEFAULT 1"},
		{"recipient_email", "TEXT DEFAULT 'powerfulcontrolsystem@gmail.com'"},
		{"disk_enabled", "INTEGER DEFAULT 1"},
		{"disk_threshold_pct", "DOUBLE PRECISION DEFAULT 80"},
		{"traffic_enabled", "INTEGER DEFAULT 1"},
		{"traffic_threshold_pct", "DOUBLE PRECISION DEFAULT 80"},
		{"traffic_threshold_gb", "DOUBLE PRECISION DEFAULT 0"},
		{"sessions_enabled", "INTEGER DEFAULT 1"},
		{"sessions_threshold", "BIGINT DEFAULT 50"},
		{"db_connections_enabled", "INTEGER DEFAULT 1"},
		{"db_connections_threshold", "BIGINT DEFAULT 80"},
		{"cooldown_minutes", "BIGINT DEFAULT 60"},
		{"fecha_creacion", "TEXT"},
		{"fecha_actualizacion", "TEXT"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	}
	for _, col := range configCols {
		if err := ensureColumnIfMissing(dbConn, "super_alertas_config", col.name, col.def); err != nil {
			return err
		}
	}

	eventCols := []struct {
		name string
		def  string
	}{
		{"tipo", "TEXT"},
		{"severidad", "TEXT DEFAULT 'warning'"},
		{"titulo", "TEXT"},
		{"detalle", "TEXT"},
		{"valor", "DOUBLE PRECISION DEFAULT 0"},
		{"umbral", "DOUBLE PRECISION DEFAULT 0"},
		{"unidad", "TEXT"},
		{"destinatario", "TEXT"},
		{"asunto", "TEXT"},
		{"cuerpo", "TEXT"},
		{"correo_enviado", "INTEGER DEFAULT 0"},
		{"correo_error", "TEXT"},
		{"metadata_json", "TEXT"},
		{"fecha_evento", "TEXT"},
		{"fecha_creacion", "TEXT"},
		{"fecha_actualizacion", "TEXT"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	}
	for _, col := range eventCols {
		if err := ensureColumnIfMissing(dbConn, "super_alertas_eventos", col.name, col.def); err != nil {
			return err
		}
	}

	_, _ = ExecCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_super_alertas_eventos_tipo_fecha ON super_alertas_eventos(tipo, fecha_evento DESC)`)
	_, _ = ExecCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_super_alertas_eventos_fecha ON super_alertas_eventos(fecha_evento DESC)`)
	return nil
}

func GetSuperAlertasConfig(dbConn *sql.DB) (SuperAlertaConfig, error) {
	if err := EnsureSuperAlertasSchema(dbConn); err != nil {
		return SuperAlertaConfig{}, err
	}
	cfg := SuperAlertaConfig{}
	row := QueryRowCompat(dbConn, `SELECT id, COALESCE(enabled,1), COALESCE(recipient_email,''), COALESCE(disk_enabled,1), COALESCE(disk_threshold_pct,80), COALESCE(traffic_enabled,1), COALESCE(traffic_threshold_pct,80), COALESCE(traffic_threshold_gb,0), COALESCE(sessions_enabled,1), COALESCE(sessions_threshold,50), COALESCE(db_connections_enabled,1), COALESCE(db_connections_threshold,80), COALESCE(cooldown_minutes,60), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,'activo'), COALESCE(observaciones,'') FROM super_alertas_config WHERE id = 1 LIMIT 1`)
	var enabled, diskEnabled, trafficEnabled, sessionsEnabled, dbConnectionsEnabled int
	err := row.Scan(&cfg.ID, &enabled, &cfg.RecipientEmail, &diskEnabled, &cfg.DiskThresholdPct, &trafficEnabled, &cfg.TrafficThresholdPct, &cfg.TrafficThresholdGB, &sessionsEnabled, &cfg.SessionsThreshold, &dbConnectionsEnabled, &cfg.DBConnectionsThreshold, &cfg.CooldownMinutes, &cfg.FechaCreacion, &cfg.FechaActualizacion, &cfg.UsuarioCreador, &cfg.Estado, &cfg.Observaciones)
	if err == sql.ErrNoRows {
		cfg = defaultSuperAlertaConfig()
		if saveErr := SaveSuperAlertasConfig(dbConn, cfg); saveErr != nil {
			return SuperAlertaConfig{}, saveErr
		}
		return cfg, nil
	}
	if err != nil {
		return SuperAlertaConfig{}, err
	}
	cfg.Enabled = enabled != 0
	cfg.DiskEnabled = diskEnabled != 0
	cfg.TrafficEnabled = trafficEnabled != 0
	cfg.SessionsEnabled = sessionsEnabled != 0
	cfg.DBConnectionsEnabled = dbConnectionsEnabled != 0
	cfg.normalize()
	return cfg, nil
}

func SaveSuperAlertasConfig(dbConn *sql.DB, cfg SuperAlertaConfig) error {
	if err := EnsureSuperAlertasSchema(dbConn); err != nil {
		return err
	}
	cfg.normalize()
	nowExpr := sqlNowExpr()
	res, err := ExecCompat(dbConn, `UPDATE super_alertas_config SET enabled=?, recipient_email=?, disk_enabled=?, disk_threshold_pct=?, traffic_enabled=?, traffic_threshold_pct=?, traffic_threshold_gb=?, sessions_enabled=?, sessions_threshold=?, db_connections_enabled=?, db_connections_threshold=?, cooldown_minutes=?, fecha_actualizacion=`+nowExpr+`, usuario_creador=?, estado=?, observaciones=? WHERE id=1`,
		superAlertBoolInt(cfg.Enabled), cfg.RecipientEmail, superAlertBoolInt(cfg.DiskEnabled), cfg.DiskThresholdPct, superAlertBoolInt(cfg.TrafficEnabled), cfg.TrafficThresholdPct, cfg.TrafficThresholdGB, superAlertBoolInt(cfg.SessionsEnabled), cfg.SessionsThreshold, superAlertBoolInt(cfg.DBConnectionsEnabled), cfg.DBConnectionsThreshold, cfg.CooldownMinutes, cfg.UsuarioCreador, cfg.Estado, cfg.Observaciones)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected > 0 {
		return nil
	}
	_, err = ExecCompat(dbConn, `INSERT INTO super_alertas_config (id, enabled, recipient_email, disk_enabled, disk_threshold_pct, traffic_enabled, traffic_threshold_pct, traffic_threshold_gb, sessions_enabled, sessions_threshold, db_connections_enabled, db_connections_threshold, cooldown_minutes, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones) VALUES (1,?,?,?,?,?,?,?,?,?,?,?,?,`+nowExpr+`,`+nowExpr+`,?,?,?)`,
		superAlertBoolInt(cfg.Enabled), cfg.RecipientEmail, superAlertBoolInt(cfg.DiskEnabled), cfg.DiskThresholdPct, superAlertBoolInt(cfg.TrafficEnabled), cfg.TrafficThresholdPct, cfg.TrafficThresholdGB, superAlertBoolInt(cfg.SessionsEnabled), cfg.SessionsThreshold, superAlertBoolInt(cfg.DBConnectionsEnabled), cfg.DBConnectionsThreshold, cfg.CooldownMinutes, cfg.UsuarioCreador, cfg.Estado, cfg.Observaciones)
	return err
}

func (cfg *SuperAlertaConfig) normalize() {
	cfg.ID = 1
	cfg.RecipientEmail = strings.TrimSpace(cfg.RecipientEmail)
	if cfg.RecipientEmail == "" {
		cfg.RecipientEmail = "powerfulcontrolsystem@gmail.com"
	}
	if cfg.DiskThresholdPct <= 0 {
		cfg.DiskThresholdPct = 80
	}
	if cfg.TrafficThresholdPct < 0 {
		cfg.TrafficThresholdPct = 0
	}
	if cfg.SessionsThreshold <= 0 {
		cfg.SessionsThreshold = 50
	}
	if cfg.DBConnectionsThreshold <= 0 {
		cfg.DBConnectionsThreshold = 80
	}
	if cfg.CooldownMinutes <= 0 {
		cfg.CooldownMinutes = 60
	}
	cfg.UsuarioCreador = strings.TrimSpace(cfg.UsuarioCreador)
	if cfg.UsuarioCreador == "" {
		cfg.UsuarioCreador = "sistema"
	}
	cfg.Estado = strings.TrimSpace(cfg.Estado)
	if cfg.Estado == "" {
		cfg.Estado = "activo"
	}
	cfg.Observaciones = strings.TrimSpace(cfg.Observaciones)
}

func superAlertBoolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func CreateSuperAlertaEvento(dbConn *sql.DB, item SuperAlertaEvento) (int64, error) {
	if err := EnsureSuperAlertasSchema(dbConn); err != nil {
		return 0, err
	}
	item.normalize()
	return insertSQLCompat(dbConn, `INSERT INTO super_alertas_eventos (tipo, severidad, titulo, detalle, valor, umbral, unidad, destinatario, asunto, cuerpo, correo_enviado, correo_error, metadata_json, fecha_evento, usuario_creador, estado, observaciones) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.Tipo, item.Severidad, item.Titulo, item.Detalle, item.Valor, item.Umbral, item.Unidad, item.Destinatario, item.Asunto, item.Cuerpo, superAlertBoolInt(item.CorreoEnviado), item.CorreoError, item.MetadataJSON, item.FechaEvento, item.UsuarioCreador, item.Estado, item.Observaciones)
}

func (item *SuperAlertaEvento) normalize() {
	item.Tipo = strings.TrimSpace(item.Tipo)
	item.Severidad = strings.TrimSpace(item.Severidad)
	if item.Severidad == "" {
		item.Severidad = "warning"
	}
	item.Titulo = strings.TrimSpace(item.Titulo)
	item.Detalle = strings.TrimSpace(item.Detalle)
	item.Unidad = strings.TrimSpace(item.Unidad)
	item.Destinatario = strings.TrimSpace(item.Destinatario)
	item.Asunto = strings.TrimSpace(item.Asunto)
	item.Cuerpo = strings.TrimSpace(item.Cuerpo)
	item.CorreoError = strings.TrimSpace(item.CorreoError)
	item.MetadataJSON = strings.TrimSpace(item.MetadataJSON)
	item.UsuarioCreador = strings.TrimSpace(item.UsuarioCreador)
	if item.UsuarioCreador == "" {
		item.UsuarioCreador = "sistema"
	}
	item.Estado = strings.TrimSpace(item.Estado)
	if item.Estado == "" {
		item.Estado = "activo"
	}
	item.Observaciones = strings.TrimSpace(item.Observaciones)
	if item.FechaEvento == "" {
		item.FechaEvento = time.Now().Format("2006-01-02 15:04:05")
	}
}

func ListSuperAlertaEventos(dbConn *sql.DB, limit int) ([]SuperAlertaEvento, error) {
	if err := EnsureSuperAlertasSchema(dbConn); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT id, COALESCE(tipo,''), COALESCE(severidad,''), COALESCE(titulo,''), COALESCE(detalle,''), COALESCE(valor,0), COALESCE(umbral,0), COALESCE(unidad,''), COALESCE(destinatario,''), COALESCE(asunto,''), COALESCE(cuerpo,''), COALESCE(correo_enviado,0), COALESCE(correo_error,''), COALESCE(metadata_json,''), COALESCE(fecha_evento,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(estado,''), COALESCE(observaciones,'') FROM super_alertas_eventos ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SuperAlertaEvento{}
	for rows.Next() {
		var item SuperAlertaEvento
		var sent int
		if err := rows.Scan(&item.ID, &item.Tipo, &item.Severidad, &item.Titulo, &item.Detalle, &item.Valor, &item.Umbral, &item.Unidad, &item.Destinatario, &item.Asunto, &item.Cuerpo, &sent, &item.CorreoError, &item.MetadataJSON, &item.FechaEvento, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		item.CorreoEnviado = sent != 0
		out = append(out, item)
	}
	return out, rows.Err()
}

func SuperAlertaRecentlySent(dbConn *sql.DB, tipo string, cooldownMinutes int64) (bool, error) {
	if cooldownMinutes <= 0 {
		return false, nil
	}
	if err := EnsureSuperAlertasSchema(dbConn); err != nil {
		return false, err
	}
	var raw string
	err := QueryRowCompat(dbConn, `SELECT COALESCE(fecha_evento,'') FROM super_alertas_eventos WHERE tipo=? ORDER BY id DESC LIMIT 1`, strings.TrimSpace(tipo)).Scan(&raw)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	last, ok := parseSuperAlertaTime(raw)
	if !ok {
		return false, nil
	}
	return time.Since(last) < time.Duration(cooldownMinutes)*time.Minute, nil
}

func parseSuperAlertaTime(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05.999999-07",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func CountActiveAdminSessions(dbConn *sql.DB) (int64, error) {
	condition := sessionNotExpiredCondition("fecha_fin")
	var total int64
	err := QueryRowCompat(dbConn, "SELECT COUNT(1) FROM sesiones WHERE COALESCE(activo,0)=1 AND "+condition).Scan(&total)
	if err != nil {
		if isMissingTableError(err) || isMissingColumnError(err) {
			return 0, nil
		}
		return 0, err
	}
	return total, nil
}

func CountDatabaseConnections(dbConn *sql.DB) (int64, error) {
	if !isPostgresDialect() {
		return 0, nil
	}
	var total int64
	if err := QueryRowCompat(dbConn, "SELECT COUNT(1) FROM pg_stat_activity").Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func GetFloatConfig(dbConn *sql.DB, key string) float64 {
	raw, _, err := GetConfigValue(dbConn, key)
	if err != nil {
		return 0
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}
	return n
}
