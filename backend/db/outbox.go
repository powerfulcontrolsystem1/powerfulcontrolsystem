package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	OutboxPending    = "pending"
	OutboxProcessing = "processing"
	OutboxPublished  = "published"
	OutboxDead       = "dead"

	defaultOutboxLease      = 5 * time.Minute
	outboxSchemaFingerprint = "pcs_outbox_events:v2:lease-attempts-deduplication-recovery"
)

// OutboxEvent is written in the same transaction as the business change. A
// dispatcher later creates idempotent durable jobs. It never contains provider
// credentials, user tokens or a client-controlled tenant authority value.
type OutboxEvent struct {
	ID                 int64
	EmpresaID          int64
	Topic              string
	Version            int
	PayloadJSON        string
	Status             string
	Attempts           int
	MaxAttempts        int
	AvailableAt        time.Time
	LeaseUntil         time.Time
	IdempotencyKey     string
	IdempotencyKeyHash string
}

func EnsureOutboxSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	for _, statement := range outboxSchemaStatements() {
		if _, err := execSQLCompat(dbConn, statement); err != nil {
			return err
		}
	}
	return nil
}

// VerifyOutboxSchema is read-only so API and worker startup cannot mutate a
// database that was deployed without its required migration.
func VerifyOutboxSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	var tableName sql.NullString
	if err := queryRowSQLCompat(dbConn, `SELECT to_regclass('pcs_outbox_events')`).Scan(&tableName); err != nil {
		return err
	}
	if !tableName.Valid || strings.TrimSpace(tableName.String) == "" {
		return fmt.Errorf("pcs_outbox_events schema missing; run pcs-migrate")
	}
	return nil
}

func applyOutboxSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range outboxSchemaStatements() {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}

func outboxSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS pcs_outbox_events (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL DEFAULT 0,
			topic TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			payload_json TEXT NOT NULL DEFAULT '{}',
			status TEXT NOT NULL DEFAULT 'pending',
			attempts INTEGER NOT NULL DEFAULT 0,
			max_attempts INTEGER NOT NULL DEFAULT 5,
			idempotency_key_hash TEXT,
			available_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			locked_at TIMESTAMPTZ,
			locked_by TEXT,
			lease_until TIMESTAMPTZ,
			published_at TIMESTAMPTZ,
			dead_at TIMESTAMPTZ,
			last_error TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (char_length(topic) BETWEEN 1 AND 160),
			CHECK (status IN ('pending','processing','published','dead')),
			CHECK (max_attempts BETWEEN 1 AND 25)
		)`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending'`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS max_attempts INTEGER NOT NULL DEFAULT 5`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS idempotency_key_hash TEXT`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS available_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS locked_by TEXT`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS dead_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS last_error TEXT`,
		`ALTER TABLE pcs_outbox_events ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP`,
		`UPDATE pcs_outbox_events SET status = CASE WHEN published_at IS NULL THEN 'pending' ELSE 'published' END
			WHERE status IS NULL OR btrim(status) = ''`,
		`CREATE INDEX IF NOT EXISTS ix_pcs_outbox_events_pending_v2
			ON pcs_outbox_events (available_at, id) WHERE status = 'pending'`,
		`CREATE INDEX IF NOT EXISTS ix_pcs_outbox_events_lease_v2
			ON pcs_outbox_events (lease_until, id) WHERE status = 'processing'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_pcs_outbox_events_idempotency_v2
			ON pcs_outbox_events (empresa_id, topic, idempotency_key_hash)
			WHERE idempotency_key_hash IS NOT NULL`,
	}
}

// InsertOutboxEvent performs an atomic, idempotent insert using the caller's
// business transaction. The raw idempotency key is hashed before persistence.
func InsertOutboxEvent(tx *sql.Tx, event OutboxEvent) error {
	_, _, err := InsertOutboxEventIdempotent(tx, event)
	return err
}

func InsertOutboxEventIdempotent(tx *sql.Tx, event OutboxEvent) (*OutboxEvent, bool, error) {
	if tx == nil {
		return nil, false, fmt.Errorf("transaction required")
	}
	if event.EmpresaID < 0 || strings.TrimSpace(event.Topic) == "" || len(event.Topic) > 160 {
		return nil, false, fmt.Errorf("outbox event invalid")
	}
	if len(event.PayloadJSON) > 1<<20 {
		return nil, false, fmt.Errorf("outbox payload exceeds 1 MiB")
	}
	if event.Version == 0 {
		event.Version = 1
	}
	if event.MaxAttempts == 0 {
		event.MaxAttempts = 5
	}
	if event.MaxAttempts < 1 || event.MaxAttempts > 25 {
		return nil, false, fmt.Errorf("outbox max attempts invalid")
	}
	if len(strings.TrimSpace(event.IdempotencyKey)) > 512 {
		return nil, false, fmt.Errorf("outbox idempotency key too long")
	}
	event.Topic = strings.TrimSpace(event.Topic)
	event.IdempotencyKeyHash = hashOutboxKey(event.IdempotencyKey)
	if event.AvailableAt.IsZero() {
		event.AvailableAt = time.Now().UTC()
	}
	result, err := execTxSQLCompat(tx, `INSERT INTO pcs_outbox_events
		(empresa_id, topic, version, payload_json, status, attempts, max_attempts, idempotency_key_hash, available_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'pending', 0, ?, NULLIF(?, ''), ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT DO NOTHING`, event.EmpresaID, event.Topic, event.Version, event.PayloadJSON, event.MaxAttempts,
		event.IdempotencyKeyHash, event.AvailableAt)
	if err != nil {
		return nil, false, err
	}
	if affected, _ := result.RowsAffected(); affected == 1 {
		return &event, true, nil
	}
	if event.IdempotencyKeyHash == "" {
		return nil, false, fmt.Errorf("outbox event was not created")
	}
	var existing OutboxEvent
	err = queryRowTxSQLCompat(tx, `SELECT id, empresa_id, topic, version, payload_json, status, attempts, max_attempts,
		COALESCE(idempotency_key_hash, ''), available_at
		FROM pcs_outbox_events WHERE empresa_id = ? AND topic = ? AND idempotency_key_hash = ?`,
		event.EmpresaID, event.Topic, event.IdempotencyKeyHash).
		Scan(&existing.ID, &existing.EmpresaID, &existing.Topic, &existing.Version, &existing.PayloadJSON, &existing.Status,
			&existing.Attempts, &existing.MaxAttempts, &existing.IdempotencyKeyHash, &existing.AvailableAt)
	if err != nil {
		return nil, false, err
	}
	return &existing, false, nil
}

func ClaimOutboxEvents(dbConn *sql.DB, dispatcherID string, limit int) ([]OutboxEvent, error) {
	return ClaimOutboxEventsWithLease(dbConn, dispatcherID, limit, defaultOutboxLease)
}

func ClaimOutboxEventsWithLease(dbConn *sql.DB, dispatcherID string, limit int, lease time.Duration) ([]OutboxEvent, error) {
	if dbConn == nil || strings.TrimSpace(dispatcherID) == "" || limit < 1 || limit > 100 {
		return nil, fmt.Errorf("outbox claim input invalid")
	}
	if lease < time.Second || lease > 30*time.Minute {
		return nil, fmt.Errorf("outbox lease invalid")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := execTxSQLCompat(tx, `UPDATE pcs_outbox_events
		SET status = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'pending' END,
			locked_at = NULL, locked_by = NULL, lease_until = NULL,
			dead_at = CASE WHEN attempts >= max_attempts THEN CURRENT_TIMESTAMP ELSE dead_at END,
			available_at = CASE WHEN attempts >= max_attempts THEN available_at ELSE CURRENT_TIMESTAMP END,
			last_error = CASE WHEN last_error IS NULL OR last_error = '' THEN 'dispatcher lease expired' ELSE last_error END,
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'processing' AND lease_until IS NOT NULL AND lease_until < CURRENT_TIMESTAMP`); err != nil {
		return nil, err
	}
	rows, err := queryTxSQLCompat(tx, `SELECT id, empresa_id, topic, version, payload_json, status, attempts, max_attempts,
		COALESCE(idempotency_key_hash, ''), available_at
		FROM pcs_outbox_events WHERE status = 'pending' AND available_at <= CURRENT_TIMESTAMP
		ORDER BY available_at, id FOR UPDATE SKIP LOCKED LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]OutboxEvent, 0, limit)
	for rows.Next() {
		var event OutboxEvent
		if err := rows.Scan(&event.ID, &event.EmpresaID, &event.Topic, &event.Version, &event.PayloadJSON, &event.Status,
			&event.Attempts, &event.MaxAttempts, &event.IdempotencyKeyHash, &event.AvailableAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range events {
		event := &events[index]
		if _, err := execTxSQLCompat(tx, `UPDATE pcs_outbox_events
			SET status = 'processing', attempts = attempts + 1, locked_at = CURRENT_TIMESTAMP, locked_by = ?,
				lease_until = CURRENT_TIMESTAMP + (? * interval '1 millisecond'), updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND status = 'pending'`, strings.TrimSpace(dispatcherID), lease.Milliseconds(), event.ID); err != nil {
			return nil, err
		}
		event.Status = OutboxProcessing
		event.Attempts++
		event.LeaseUntil = time.Now().UTC().Add(lease)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return events, nil
}

func CompleteOutboxEvent(dbConn *sql.DB, id int64, dispatcherID string) error {
	result, err := execSQLCompat(dbConn, `UPDATE pcs_outbox_events
		SET status = 'published', locked_at = NULL, locked_by = NULL, lease_until = NULL,
			published_at = CURRENT_TIMESTAMP, last_error = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, id, strings.TrimSpace(dispatcherID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("outbox event not owned by dispatcher")
	}
	return nil
}

func RetryOutboxEvent(dbConn *sql.DB, event OutboxEvent, dispatcherID string, cause error, retryAfter time.Duration) error {
	if event.ID <= 0 || strings.TrimSpace(dispatcherID) == "" {
		return fmt.Errorf("outbox ownership invalid")
	}
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	status := OutboxPending
	if event.Attempts >= event.MaxAttempts {
		status = OutboxDead
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_outbox_events
		SET status = ?, available_at = ?, locked_at = NULL, locked_by = NULL, lease_until = NULL,
			dead_at = CASE WHEN ? = 'dead' THEN CURRENT_TIMESTAMP ELSE dead_at END,
			last_error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, status, time.Now().UTC().Add(retryAfter), status,
		redactAsyncJobError(cause), event.ID, strings.TrimSpace(dispatcherID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("outbox event not owned by dispatcher")
	}
	return nil
}

func RecoverExpiredOutboxEvents(dbConn *sql.DB) (int64, error) {
	result, err := execSQLCompat(dbConn, `UPDATE pcs_outbox_events
		SET status = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'pending' END,
			locked_at = NULL, locked_by = NULL, lease_until = NULL,
			dead_at = CASE WHEN attempts >= max_attempts THEN CURRENT_TIMESTAMP ELSE dead_at END,
			available_at = CASE WHEN attempts >= max_attempts THEN available_at ELSE CURRENT_TIMESTAMP END,
			last_error = CASE WHEN last_error IS NULL OR last_error = '' THEN 'dispatcher lease expired' ELSE last_error END,
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'processing' AND lease_until IS NOT NULL AND lease_until < CURRENT_TIMESTAMP`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func hashOutboxKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
