package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	QueueLanePrinting   = "printing"
	QueueLaneProductAdd = "product_add"
	QueueLaneFiscal     = "fiscal"

	queueCapacitySuperSchemaFingerprint    = "queue-capacity-super:v1:lane-config-alert-thresholds"
	queueCapacityBusinessSchemaFingerprint = "queue-capacity-business:v1:fair-tenant-service-and-operational-indexes"
)

// QueueCapacityConfig controls one independent operational lane. Rate limits
// are per tenant and minute; backlog thresholds are global observability
// limits. Fiscal work is never discarded merely because an alert threshold is
// crossed.
type QueueCapacityConfig struct {
	Lane                  string `json:"lane"`
	Label                 string `json:"label"`
	AlertsEnabled         bool   `json:"alerts_enabled"`
	RateLimitPerMinute    int64  `json:"rate_limit_per_minute"`
	PendingAlertThreshold int64  `json:"pending_alert_threshold"`
	OldestAlertSeconds    int64  `json:"oldest_alert_seconds"`
	MaxPendingPerTenant   int64  `json:"max_pending_per_tenant"`
	UpdatedAt             string `json:"updated_at,omitempty"`
	UpdatedBy             string `json:"updated_by,omitempty"`
}

type QueueCapacitySnapshot struct {
	Lane                  string  `json:"lane"`
	Label                 string  `json:"label"`
	Pending               int64   `json:"pending"`
	Processing            int64   `json:"processing"`
	Failed                int64   `json:"failed"`
	ActiveTenants         int64   `json:"active_tenants"`
	BusiestTenantID       int64   `json:"busiest_tenant_id,omitempty"`
	BusiestTenantPending  int64   `json:"busiest_tenant_pending"`
	OldestSeconds         float64 `json:"oldest_seconds"`
	RequestsCurrentMinute int64   `json:"requests_current_minute"`
	SaturationPercent     float64 `json:"saturation_percent"`
	QueryOK               bool    `json:"query_ok"`
	Error                 string  `json:"error,omitempty"`
}

func defaultQueueCapacityConfigs() []QueueCapacityConfig {
	return []QueueCapacityConfig{
		{Lane: QueueLanePrinting, Label: "Impresiones", AlertsEnabled: true, RateLimitPerMinute: 120, PendingAlertThreshold: 200, OldestAlertSeconds: 120, MaxPendingPerTenant: 100},
		{Lane: QueueLaneProductAdd, Label: "Agregar productos", AlertsEnabled: true, RateLimitPerMinute: 240, PendingAlertThreshold: 0, OldestAlertSeconds: 0, MaxPendingPerTenant: 0},
		{Lane: QueueLaneFiscal, Label: "Emision de facturas", AlertsEnabled: true, RateLimitPerMinute: 30, PendingAlertThreshold: 100, OldestAlertSeconds: 300, MaxPendingPerTenant: 25},
	}
}

func normalizeQueueLane(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case QueueLanePrinting:
		return QueueLanePrinting
	case QueueLaneProductAdd:
		return QueueLaneProductAdd
	case QueueLaneFiscal:
		return QueueLaneFiscal
	default:
		return ""
	}
}

func defaultQueueCapacityConfig(lane string) QueueCapacityConfig {
	lane = normalizeQueueLane(lane)
	for _, cfg := range defaultQueueCapacityConfigs() {
		if cfg.Lane == lane {
			return cfg
		}
	}
	return QueueCapacityConfig{}
}

func validateQueueCapacityConfig(cfg QueueCapacityConfig) error {
	if normalizeQueueLane(cfg.Lane) == "" {
		return fmt.Errorf("carril de cola no soportado")
	}
	if cfg.RateLimitPerMinute < 1 || cfg.RateLimitPerMinute > 100000 {
		return fmt.Errorf("rate_limit_per_minute debe estar entre 1 y 100000")
	}
	if cfg.PendingAlertThreshold < 0 || cfg.PendingAlertThreshold > 1000000 {
		return fmt.Errorf("pending_alert_threshold fuera de rango")
	}
	if cfg.OldestAlertSeconds < 0 || cfg.OldestAlertSeconds > 604800 {
		return fmt.Errorf("oldest_alert_seconds fuera de rango")
	}
	if cfg.MaxPendingPerTenant < 0 || cfg.MaxPendingPerTenant > 100000 {
		return fmt.Errorf("max_pending_per_tenant fuera de rango")
	}
	return nil
}

func applyQueueCapacitySuperSchemaTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS super_queue_capacity_config (
		lane TEXT PRIMARY KEY,
		label TEXT NOT NULL,
		alerts_enabled BOOLEAN NOT NULL DEFAULT TRUE,
		rate_limit_per_minute BIGINT NOT NULL,
		pending_alert_threshold BIGINT NOT NULL DEFAULT 0,
		oldest_alert_seconds BIGINT NOT NULL DEFAULT 0,
		max_pending_per_tenant BIGINT NOT NULL DEFAULT 0,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_by TEXT NOT NULL DEFAULT 'sistema',
		CHECK (lane IN ('printing','product_add','fiscal')),
		CHECK (rate_limit_per_minute BETWEEN 1 AND 100000),
		CHECK (pending_alert_threshold BETWEEN 0 AND 1000000),
		CHECK (oldest_alert_seconds BETWEEN 0 AND 604800),
		CHECK (max_pending_per_tenant BETWEEN 0 AND 100000)
	)`); err != nil {
		return err
	}
	for _, cfg := range defaultQueueCapacityConfigs() {
		if _, err := tx.ExecContext(ctx, `INSERT INTO super_queue_capacity_config
			(lane, label, alerts_enabled, rate_limit_per_minute, pending_alert_threshold, oldest_alert_seconds, max_pending_per_tenant)
			VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (lane) DO NOTHING`, cfg.Lane, cfg.Label, cfg.AlertsEnabled,
			cfg.RateLimitPerMinute, cfg.PendingAlertThreshold, cfg.OldestAlertSeconds, cfg.MaxPendingPerTenant); err != nil {
			return err
		}
	}
	return nil
}

func applyQueueCapacityBusinessSchemaTx(ctx context.Context, tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS pcs_queue_tenant_state (
			queue_name TEXT NOT NULL,
			empresa_id BIGINT NOT NULL,
			last_served_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (queue_name, empresa_id)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_impresoras_cola_operational_v1 ON empresa_impresoras_cola (estado, fecha_creacion, empresa_id)`,
		`CREATE INDEX IF NOT EXISTS ix_fe_reintentos_operational_v1 ON facturacion_electronica_reintentos (estado_envio, estado, empresa_id, id)`,
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}

func QueueCapacitySuperSchemaReady(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("base super administrador no disponible")
	}
	var marker int
	err := dbConn.QueryRow(`SELECT 1 FROM super_queue_capacity_config WHERE 1=0`).Scan(&marker)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("configuracion de capacidad no disponible: %w", err)
	}
	return nil
}

func GetQueueCapacityConfigs(dbConn *sql.DB) ([]QueueCapacityConfig, error) {
	if err := QueueCapacitySuperSchemaReady(dbConn); err != nil {
		return nil, err
	}
	rows, err := dbConn.Query(`SELECT lane, label, alerts_enabled, rate_limit_per_minute,
		pending_alert_threshold, oldest_alert_seconds, max_pending_per_tenant,
		CAST(updated_at AS TEXT), updated_by
		FROM super_queue_capacity_config ORDER BY CASE lane WHEN 'printing' THEN 1 WHEN 'product_add' THEN 2 ELSE 3 END`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	configs := make([]QueueCapacityConfig, 0, 3)
	for rows.Next() {
		var cfg QueueCapacityConfig
		if err := rows.Scan(&cfg.Lane, &cfg.Label, &cfg.AlertsEnabled, &cfg.RateLimitPerMinute,
			&cfg.PendingAlertThreshold, &cfg.OldestAlertSeconds, &cfg.MaxPendingPerTenant,
			&cfg.UpdatedAt, &cfg.UpdatedBy); err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, rows.Err()
}

func GetQueueCapacityConfig(dbConn *sql.DB, lane string) (QueueCapacityConfig, error) {
	lane = normalizeQueueLane(lane)
	if lane == "" {
		return QueueCapacityConfig{}, fmt.Errorf("carril de cola no soportado")
	}
	if err := QueueCapacitySuperSchemaReady(dbConn); err != nil {
		return QueueCapacityConfig{}, err
	}
	var cfg QueueCapacityConfig
	err := dbConn.QueryRow(`SELECT lane, label, alerts_enabled, rate_limit_per_minute,
		pending_alert_threshold, oldest_alert_seconds, max_pending_per_tenant,
		CAST(updated_at AS TEXT), updated_by FROM super_queue_capacity_config WHERE lane=$1`, lane).
		Scan(&cfg.Lane, &cfg.Label, &cfg.AlertsEnabled, &cfg.RateLimitPerMinute,
			&cfg.PendingAlertThreshold, &cfg.OldestAlertSeconds, &cfg.MaxPendingPerTenant,
			&cfg.UpdatedAt, &cfg.UpdatedBy)
	if err == sql.ErrNoRows {
		return defaultQueueCapacityConfig(lane), nil
	}
	return cfg, err
}

func SaveQueueCapacityConfigs(ctx context.Context, dbConn *sql.DB, configs []QueueCapacityConfig, updatedBy string) error {
	if dbConn == nil {
		return fmt.Errorf("base super administrador no disponible")
	}
	if len(configs) != 3 {
		return fmt.Errorf("se requieren los tres carriles de capacidad")
	}
	seen := map[string]bool{}
	for i := range configs {
		configs[i].Lane = normalizeQueueLane(configs[i].Lane)
		if err := validateQueueCapacityConfig(configs[i]); err != nil {
			return err
		}
		if seen[configs[i].Lane] {
			return fmt.Errorf("carril de cola duplicado")
		}
		seen[configs[i].Lane] = true
	}
	updatedBy = strings.TrimSpace(updatedBy)
	if updatedBy == "" {
		updatedBy = "sistema"
	}
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	for _, cfg := range configs {
		label := strings.TrimSpace(cfg.Label)
		if label == "" {
			label = defaultQueueCapacityConfig(cfg.Lane).Label
		}
		if _, err := tx.ExecContext(ctx, `UPDATE super_queue_capacity_config SET label=$2,
			alerts_enabled=$3, rate_limit_per_minute=$4, pending_alert_threshold=$5,
			oldest_alert_seconds=$6, max_pending_per_tenant=$7,
			updated_at=CURRENT_TIMESTAMP, updated_by=$8 WHERE lane=$1`, cfg.Lane, label,
			cfg.AlertsEnabled, cfg.RateLimitPerMinute, cfg.PendingAlertThreshold,
			cfg.OldestAlertSeconds, cfg.MaxPendingPerTenant, updatedBy); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func QueueLaneForRateScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "queue.printing":
		return QueueLanePrinting
	case "queue.product_add":
		return QueueLaneProductAdd
	case "queue.fiscal":
		return QueueLaneFiscal
	default:
		return ""
	}
}

func GetQueueCapacitySnapshots(ctx context.Context, businessDB, superDB *sql.DB, configs []QueueCapacityConfig) []QueueCapacitySnapshot {
	byLane := map[string]QueueCapacityConfig{}
	for _, cfg := range configs {
		byLane[normalizeQueueLane(cfg.Lane)] = cfg
	}
	result := []QueueCapacitySnapshot{
		collectPrintingQueueSnapshot(ctx, businessDB, byLane[QueueLanePrinting]),
		collectProductAddQueueSnapshot(ctx, superDB, byLane[QueueLaneProductAdd]),
		collectFiscalQueueSnapshot(ctx, businessDB, byLane[QueueLaneFiscal]),
	}
	return result
}

func collectPrintingQueueSnapshot(ctx context.Context, dbConn *sql.DB, cfg QueueCapacityConfig) QueueCapacitySnapshot {
	s := QueueCapacitySnapshot{Lane: QueueLanePrinting, Label: defaultQueueCapacityConfig(QueueLanePrinting).Label}
	if dbConn == nil {
		s.Error = "base empresarial no disponible"
		return s
	}
	err := dbConn.QueryRowContext(ctx, `WITH base AS (
		SELECT empresa_id, COALESCE(estado,'pendiente') AS estado,
			CASE WHEN COALESCE(fecha_creacion,'') ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}' THEN CAST(fecha_creacion AS TIMESTAMPTZ) END AS created_at
		FROM empresa_impresoras_cola
	), per_tenant AS (
		SELECT empresa_id, COUNT(*) FILTER (WHERE estado IN ('pendiente','tomado')) AS active_count
		FROM base GROUP BY empresa_id
	), top_tenant AS (
		SELECT empresa_id, active_count FROM per_tenant ORDER BY active_count DESC, empresa_id LIMIT 1
	)
	SELECT COUNT(*) FILTER (WHERE estado='pendiente'), COUNT(*) FILTER (WHERE estado='tomado'),
		COUNT(*) FILTER (WHERE estado='error'), COUNT(DISTINCT empresa_id) FILTER (WHERE estado IN ('pendiente','tomado')),
		COALESCE(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP-MIN(created_at) FILTER (WHERE estado='pendiente'))),0),
		COALESCE((SELECT empresa_id FROM top_tenant),0), COALESCE((SELECT active_count FROM top_tenant),0)
	FROM base`).Scan(&s.Pending, &s.Processing, &s.Failed, &s.ActiveTenants, &s.OldestSeconds, &s.BusiestTenantID, &s.BusiestTenantPending)
	if err != nil {
		s.Error = "no se pudo medir la cola de impresion"
		return s
	}
	s.QueryOK = true
	s.SaturationPercent = queueSaturationPercent(s, cfg)
	return s
}

func collectProductAddQueueSnapshot(ctx context.Context, dbConn *sql.DB, cfg QueueCapacityConfig) QueueCapacitySnapshot {
	s := QueueCapacitySnapshot{Lane: QueueLaneProductAdd, Label: defaultQueueCapacityConfig(QueueLaneProductAdd).Label}
	if dbConn == nil {
		s.Error = "base super administrador no disponible"
		return s
	}
	err := dbConn.QueryRowContext(ctx, `WITH active AS (
		SELECT empresa_id, request_count FROM empresa_api_rate_limits
		WHERE scope='queue.product_add' AND window_start=DATE_TRUNC('minute',CURRENT_TIMESTAMP)
	), top_tenant AS (
		SELECT empresa_id, request_count FROM active ORDER BY request_count DESC, empresa_id LIMIT 1
	)
	SELECT COALESCE(SUM(request_count),0), COUNT(*),
		COALESCE((SELECT empresa_id FROM top_tenant),0), COALESCE((SELECT request_count FROM top_tenant),0)
	FROM active`).Scan(&s.RequestsCurrentMinute, &s.ActiveTenants, &s.BusiestTenantID, &s.BusiestTenantPending)
	if err != nil {
		s.Error = "no se pudo medir el carril de productos"
		return s
	}
	s.QueryOK = true
	if cfg.RateLimitPerMinute > 0 {
		s.SaturationPercent = float64(s.BusiestTenantPending) * 100 / float64(cfg.RateLimitPerMinute)
	}
	return s
}

func collectFiscalQueueSnapshot(ctx context.Context, dbConn *sql.DB, cfg QueueCapacityConfig) QueueCapacitySnapshot {
	s := QueueCapacitySnapshot{Lane: QueueLaneFiscal, Label: defaultQueueCapacityConfig(QueueLaneFiscal).Label}
	if dbConn == nil {
		s.Error = "base empresarial no disponible"
		return s
	}
	err := dbConn.QueryRowContext(ctx, `WITH base AS (
		SELECT empresa_id, COALESCE(estado_envio,'pendiente') AS estado_envio, lease_until,
			CASE WHEN COALESCE(fecha_creacion,'') ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}' THEN CAST(fecha_creacion AS TIMESTAMPTZ) END AS created_at
		FROM facturacion_electronica_reintentos WHERE estado='activo'
	), per_tenant AS (
		SELECT empresa_id, COUNT(*) FILTER (WHERE estado_envio IN ('pendiente','fallido','contingencia')) AS active_count
		FROM base GROUP BY empresa_id
	), top_tenant AS (
		SELECT empresa_id, active_count FROM per_tenant ORDER BY active_count DESC, empresa_id LIMIT 1
	)
	SELECT COUNT(*) FILTER (WHERE estado_envio IN ('pendiente','fallido','contingencia') AND (lease_until IS NULL OR lease_until<CURRENT_TIMESTAMP)),
		COUNT(*) FILTER (WHERE estado_envio IN ('pendiente','fallido','contingencia') AND lease_until>=CURRENT_TIMESTAMP),
		COUNT(*) FILTER (WHERE estado_envio='fallido_terminal'),
		COUNT(DISTINCT empresa_id) FILTER (WHERE estado_envio IN ('pendiente','fallido','contingencia')),
		COALESCE(EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP-MIN(created_at) FILTER (WHERE estado_envio IN ('pendiente','fallido','contingencia')))),0),
		COALESCE((SELECT empresa_id FROM top_tenant),0), COALESCE((SELECT active_count FROM top_tenant),0)
	FROM base`).Scan(&s.Pending, &s.Processing, &s.Failed, &s.ActiveTenants, &s.OldestSeconds, &s.BusiestTenantID, &s.BusiestTenantPending)
	if err != nil {
		s.Error = "no se pudo medir la cola fiscal"
		return s
	}
	s.QueryOK = true
	s.SaturationPercent = queueSaturationPercent(s, cfg)
	return s
}

func queueSaturationPercent(snapshot QueueCapacitySnapshot, cfg QueueCapacityConfig) float64 {
	percent := float64(0)
	if cfg.PendingAlertThreshold > 0 {
		percent = float64(snapshot.Pending) * 100 / float64(cfg.PendingAlertThreshold)
	}
	if cfg.OldestAlertSeconds > 0 {
		agePercent := snapshot.OldestSeconds * 100 / float64(cfg.OldestAlertSeconds)
		if agePercent > percent {
			percent = agePercent
		}
	}
	if cfg.MaxPendingPerTenant > 0 {
		tenantPercent := float64(snapshot.BusiestTenantPending) * 100 / float64(cfg.MaxPendingPerTenant)
		if tenantPercent > percent {
			percent = tenantPercent
		}
	}
	return percent
}

func MarkQueueTenantServedContext(ctx context.Context, dbConn *sql.DB, queueName string, empresaID int64) error {
	if dbConn == nil || empresaID <= 0 || strings.TrimSpace(queueName) == "" {
		return fmt.Errorf("estado de servicio de cola invalido")
	}
	_, err := dbConn.ExecContext(ctx, `INSERT INTO pcs_queue_tenant_state (queue_name, empresa_id, last_served_at)
		VALUES ($1,$2,CURRENT_TIMESTAMP) ON CONFLICT (queue_name,empresa_id)
		DO UPDATE SET last_served_at=EXCLUDED.last_served_at`, strings.TrimSpace(queueName), empresaID)
	return err
}

func QueueCapacityConfigCacheTTL() time.Duration { return 5 * time.Second }
