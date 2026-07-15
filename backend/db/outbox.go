package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	OutboxPending    = "pending"
	OutboxProcessing = "processing"
	OutboxPublished  = "published"
	OutboxDead       = "dead"
)

// OutboxEvent is persisted in the same transaction as its business mutation.
// Payloads contain references and safe metadata only, never credentials or raw
// payment/document content.
type OutboxEvent struct {
	ID             int64
	EmpresaID      int64
	Topic          string
	Version        int
	PayloadJSON    string
	Status         string
	Attempts       int
	MaxAttempts    int
	AvailableAt    time.Time
	CorrelationID  string
	IdempotencyKey string
}

func ValidateOutboxEvent(event OutboxEvent) error {
	if event.EmpresaID < 0 || strings.TrimSpace(event.Topic) == "" || len(event.Topic) > 160 {
		return fmt.Errorf("outbox event invalid")
	}
	if event.Version < 0 || event.Version > 1000 || len(event.PayloadJSON) > 1<<20 {
		return fmt.Errorf("outbox event payload invalid")
	}
	if event.MaxAttempts != 0 && (event.MaxAttempts < 1 || event.MaxAttempts > 25) {
		return fmt.Errorf("outbox retry policy invalid")
	}
	if len(strings.TrimSpace(event.CorrelationID)) > 120 {
		return fmt.Errorf("outbox correlation id too long")
	}
	if key := strings.TrimSpace(event.IdempotencyKey); key != "" && (len(key) < 16 || len(key) > 200) {
		return fmt.Errorf("outbox idempotency key invalid")
	}
	return nil
}

// EnsureOutboxSchema belongs exclusively to pcs-migrate. Request paths only
// insert events inside existing transactions after the schema is present.
func EnsureOutboxSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	_, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS pcs_outbox_events (
		id BIGSERIAL PRIMARY KEY,
		empresa_id BIGINT NOT NULL DEFAULT 0,
		topic TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		payload_json TEXT NOT NULL DEFAULT '{}',
		status TEXT NOT NULL DEFAULT 'pending',
		attempts INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 5,
		available_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		locked_at TIMESTAMPTZ,
		locked_by TEXT,
		heartbeat_at TIMESTAMPTZ,
		published_at TIMESTAMPTZ,
		last_error TEXT,
		correlation_id TEXT NOT NULL DEFAULT '',
		idempotency_key_hash TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CHECK (status IN ('pending','processing','published','dead')),
		CHECK (version BETWEEN 0 AND 1000),
		CHECK (max_attempts BETWEEN 1 AND 25)
	)`)
	if err != nil {
		return err
	}
	for _, statement := range []string{
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending'`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 5`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS available_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS locked_by TEXT`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS heartbeat_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS last_error TEXT`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS correlation_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS idempotency_key_hash TEXT`,
	} {
		if _, err := execSQLCompat(dbConn, statement); err != nil {
			return err
		}
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_pcs_outbox_events_ready
		ON pcs_outbox_events (status, available_at, id)`); err != nil {
		return err
	}
	_, err = execSQLCompat(dbConn, `CREATE UNIQUE INDEX IF NOT EXISTS ux_pcs_outbox_events_idempotency
		ON pcs_outbox_events (empresa_id, topic, idempotency_key_hash)
		WHERE idempotency_key_hash IS NOT NULL`)
	return err
}

func InsertOutboxEvent(tx *sql.Tx, event OutboxEvent) error {
	if tx == nil {
		return fmt.Errorf("transaction required")
	}
	if event.MaxAttempts == 0 {
		event.MaxAttempts = 5
	}
	if err := ValidateOutboxEvent(event); err != nil {
		return err
	}
	if event.AvailableAt.IsZero() {
		event.AvailableAt = time.Now().UTC()
	}
	var idempotencyHash interface{}
	if key := strings.TrimSpace(event.IdempotencyKey); key != "" {
		idempotencyHash = asyncJobIdempotencyHash(key)
	}
	_, err := execTxSQLCompat(tx, `INSERT INTO pcs_outbox_events (
		empresa_id, topic, version, payload_json, status, attempts, max_attempts,
		available_at, correlation_id, idempotency_key_hash, created_at, updated_at
	) VALUES (?, ?, ?, ?, 'pending', 0, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (empresa_id, topic, idempotency_key_hash) WHERE idempotency_key_hash IS NOT NULL DO NOTHING`,
		event.EmpresaID, strings.TrimSpace(event.Topic), normalizedOutboxVersion(event.Version), event.PayloadJSON,
		event.MaxAttempts, event.AvailableAt, strings.TrimSpace(event.CorrelationID), idempotencyHash)
	return err
}

func normalizedOutboxVersion(version int) int {
	if version == 0 {
		return 1
	}
	return version
}

// ClaimOutboxEvents uses SKIP LOCKED and is safe for concurrent dispatcher
// replicas. It intentionally does not publish anything by itself.
func ClaimOutboxEvents(dbConn *sql.DB, workerID string, limit int) ([]OutboxEvent, error) {
	if dbConn == nil || strings.TrimSpace(workerID) == "" || limit < 1 || limit > 100 {
		return nil, fmt.Errorf("outbox claim input invalid")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.Query(rebindCompatQuery(`SELECT id, empresa_id, topic, version, payload_json, status, attempts, max_attempts, available_at, correlation_id
		FROM pcs_outbox_events
		WHERE status = 'pending' AND available_at <= CURRENT_TIMESTAMP
		ORDER BY available_at, id FOR UPDATE SKIP LOCKED LIMIT ?`), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]OutboxEvent, 0, limit)
	for rows.Next() {
		var event OutboxEvent
		if err := rows.Scan(&event.ID, &event.EmpresaID, &event.Topic, &event.Version, &event.PayloadJSON, &event.Status, &event.Attempts, &event.MaxAttempts, &event.AvailableAt, &event.CorrelationID); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range events {
		event := &events[index]
		result, err := execTxSQLCompat(tx, `UPDATE pcs_outbox_events
			SET status = 'processing', attempts = attempts + 1, locked_at = CURRENT_TIMESTAMP,
				heartbeat_at = CURRENT_TIMESTAMP, locked_by = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND status = 'pending'`, strings.TrimSpace(workerID), event.ID)
		if err != nil {
			return nil, err
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return nil, fmt.Errorf("outbox claim lost")
		}
		event.Status = OutboxProcessing
		event.Attempts++
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return events, nil
}

func PublishOutboxEvent(dbConn *sql.DB, id int64, workerID string) error {
	result, err := execSQLCompat(dbConn, `UPDATE pcs_outbox_events
		SET status = 'published', published_at = CURRENT_TIMESTAMP, locked_at = NULL,
			locked_by = NULL, heartbeat_at = NULL, last_error = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, id, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("outbox event not owned by worker")
	}
	return nil
}

func RetryOutboxEvent(dbConn *sql.DB, event OutboxEvent, workerID string, retryAfter time.Duration) error {
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	status := OutboxPending
	if event.Attempts >= event.MaxAttempts {
		status = OutboxDead
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_outbox_events
		SET status = ?, available_at = ?, locked_at = NULL, locked_by = NULL,
			heartbeat_at = NULL, last_error = 'outbox dispatch failed', updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, status, time.Now().UTC().Add(retryAfter), event.ID, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("outbox event not owned by worker")
	}
	return nil
}

func RecoverExpiredOutboxEvents(dbConn *sql.DB, staleAfter time.Duration) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("database not available")
	}
	if staleAfter < time.Minute {
		staleAfter = 5 * time.Minute
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_outbox_events
		SET status = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'pending' END,
			available_at = CURRENT_TIMESTAMP, locked_at = NULL, locked_by = NULL, heartbeat_at = NULL,
			last_error = 'outbox lease expired', updated_at = CURRENT_TIMESTAMP
		WHERE status = 'processing' AND COALESCE(heartbeat_at, locked_at) IS NOT NULL
			AND COALESCE(heartbeat_at, locked_at) < ?`, time.Now().UTC().Add(-staleAfter))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// HeartbeatOutboxEvent extends a claimed event lease. This prevents an active
// dispatcher from being recovered by another replica during a slow provider
// call while preserving ownership checks.
func HeartbeatOutboxEvent(dbConn *sql.DB, id int64, workerID string) error {
	if dbConn == nil || id <= 0 || strings.TrimSpace(workerID) == "" {
		return fmt.Errorf("outbox heartbeat input invalid")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_outbox_events
		SET locked_at = CURRENT_TIMESTAMP, heartbeat_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, id, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("outbox event not owned by worker")
	}
	return nil
}

type OutboxStats struct {
	Pending    int64
	Processing int64
	Published  int64
	Dead       int64
}

func GetOutboxStats(dbConn *sql.DB) (OutboxStats, error) {
	if dbConn == nil {
		return OutboxStats{}, fmt.Errorf("database not available")
	}
	var stats OutboxStats
	err := queryRowSQLCompat(dbConn, `SELECT
		COUNT(*) FILTER (WHERE status = 'pending'),
		COUNT(*) FILTER (WHERE status = 'processing'),
		COUNT(*) FILTER (WHERE status = 'published'),
		COUNT(*) FILTER (WHERE status = 'dead')
		FROM pcs_outbox_events`).Scan(&stats.Pending, &stats.Processing, &stats.Published, &stats.Dead)
	return stats, err
}
