package db

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

const (
	outboxRecoveryAuditSchemaFingerprint = "pcs_outbox_recovery_audit:v1:tenant-topic-explicit-ids-actor-reason"
	MaxOutboxRecoveryEvents              = 20
)

// OutboxRecoveryEvent is the bounded operational view of one dead outbox
// event. PayloadJSON is returned only to trusted callers so they can build a
// topic-specific, non-secret summary; it must not be emitted raw by handlers.
type OutboxRecoveryEvent struct {
	ID             int64
	EmpresaID      int64
	Topic          string
	Version        int
	PayloadJSON    string
	Status         string
	Attempts       int
	MaxAttempts    int
	AvailableAt    string
	DeadAt         string
	LastError      string
	CreatedAt      string
	RecoveryAction string
}

func applyOutboxRecoveryAuditSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range outboxRecoveryAuditSchemaStatements() {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}

func outboxRecoveryAuditSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS pcs_outbox_recovery_audit (
			id BIGSERIAL PRIMARY KEY,
			outbox_event_id BIGINT NOT NULL,
			empresa_id BIGINT NOT NULL,
			topic TEXT NOT NULL,
			previous_attempts INTEGER NOT NULL,
			reason TEXT NOT NULL,
			requested_by TEXT NOT NULL,
			requested_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (empresa_id > 0),
			CHECK (char_length(topic) BETWEEN 1 AND 160),
			CHECK (char_length(reason) BETWEEN 10 AND 500),
			CHECK (char_length(requested_by) BETWEEN 1 AND 120)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_pcs_outbox_recovery_audit_tenant_event
			ON pcs_outbox_recovery_audit (empresa_id, outbox_event_id, requested_at DESC)`,
	}
}

// VerifyOutboxRecoveryAuditSchema is read-only. Runtime roles must never create
// this table implicitly; pcs-migrate owns the schema.
func VerifyOutboxRecoveryAuditSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	var tableName sql.NullString
	if err := queryRowSQLCompat(dbConn, `SELECT to_regclass('pcs_outbox_recovery_audit')`).Scan(&tableName); err != nil {
		return err
	}
	if !tableName.Valid || strings.TrimSpace(tableName.String) == "" {
		return fmt.Errorf("pcs_outbox_recovery_audit schema missing; run pcs-migrate")
	}
	return nil
}

// ListDeadOutboxEventsForRecovery previews only dead events in one tenant and
// one topic. Optional IDs further narrow the result; they never broaden scope.
func ListDeadOutboxEventsForRecovery(dbConn *sql.DB, empresaID int64, topic string, eventIDs []int64, limit int) ([]OutboxRecoveryEvent, error) {
	ids, topic, limit, err := normalizeOutboxRecoveryInput(empresaID, topic, eventIDs, limit, false)
	if err != nil {
		return nil, err
	}
	if dbConn == nil {
		return nil, fmt.Errorf("database not available")
	}

	whereIDs, args := outboxRecoveryIDClause(ids)
	queryArgs := make([]interface{}, 0, 3+len(args))
	queryArgs = append(queryArgs, empresaID, topic)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, limit)
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, topic, version, payload_json, status,
			attempts, max_attempts, COALESCE(CAST(available_at AS TEXT), ''),
			COALESCE(CAST(dead_at AS TEXT), ''), COALESCE(last_error, ''),
			COALESCE(CAST(created_at AS TEXT), '')
		FROM pcs_outbox_events
		WHERE empresa_id = ? AND topic = ? AND status = 'dead'`+whereIDs+`
		ORDER BY dead_at DESC NULLS LAST, id DESC LIMIT ?`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]OutboxRecoveryEvent, 0, limit)
	for rows.Next() {
		var event OutboxRecoveryEvent
		if err := rows.Scan(
			&event.ID, &event.EmpresaID, &event.Topic, &event.Version, &event.PayloadJSON,
			&event.Status, &event.Attempts, &event.MaxAttempts, &event.AvailableAt,
			&event.DeadAt, &event.LastError, &event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

// RequeueDeadOutboxEvents performs one atomic, explicitly-scoped recovery. It
// refuses partial matches so a stale, cross-tenant or already-recovered ID
// cannot be silently ignored.
func RequeueDeadOutboxEvents(dbConn *sql.DB, empresaID int64, topic string, eventIDs []int64, reason, requestedBy string) ([]OutboxRecoveryEvent, error) {
	ids, topic, _, err := normalizeOutboxRecoveryInput(empresaID, topic, eventIDs, len(eventIDs), true)
	if err != nil {
		return nil, err
	}
	reason = strings.Join(strings.Fields(strings.TrimSpace(reason)), " ")
	requestedBy = strings.TrimSpace(requestedBy)
	if len(reason) < 10 || len(reason) > 500 {
		return nil, fmt.Errorf("la razon debe tener entre 10 y 500 caracteres")
	}
	if requestedBy == "" || len(requestedBy) > 120 {
		return nil, fmt.Errorf("el actor de recuperacion es obligatorio")
	}
	if dbConn == nil {
		return nil, fmt.Errorf("database not available")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	whereIDs, idArgs := outboxRecoveryIDClause(ids)
	queryArgs := make([]interface{}, 0, 2+len(idArgs))
	queryArgs = append(queryArgs, empresaID, topic)
	queryArgs = append(queryArgs, idArgs...)
	rows, err := queryTxSQLCompat(tx, `SELECT id, empresa_id, topic, version, payload_json, status,
			attempts, max_attempts, COALESCE(CAST(available_at AS TEXT), ''),
			COALESCE(CAST(dead_at AS TEXT), ''), COALESCE(last_error, ''),
			COALESCE(CAST(created_at AS TEXT), '')
		FROM pcs_outbox_events
		WHERE empresa_id = ? AND topic = ? AND status = 'dead'`+whereIDs+`
		ORDER BY id FOR UPDATE`, queryArgs...)
	if err != nil {
		return nil, err
	}

	events := make([]OutboxRecoveryEvent, 0, len(ids))
	for rows.Next() {
		var event OutboxRecoveryEvent
		if err := rows.Scan(
			&event.ID, &event.EmpresaID, &event.Topic, &event.Version, &event.PayloadJSON,
			&event.Status, &event.Attempts, &event.MaxAttempts, &event.AvailableAt,
			&event.DeadAt, &event.LastError, &event.CreatedAt,
		); err != nil {
			_ = rows.Close()
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if len(events) != len(ids) {
		return nil, fmt.Errorf("uno o mas eventos no estan disponibles como dead dentro de la empresa y tema solicitados")
	}

	for index := range events {
		event := &events[index]
		if _, err := execTxSQLCompat(tx, `INSERT INTO pcs_outbox_recovery_audit
			(outbox_event_id, empresa_id, topic, previous_attempts, reason, requested_by, requested_at)
			VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
			event.ID, empresaID, topic, event.Attempts, reason, requestedBy); err != nil {
			return nil, err
		}
		result, err := execTxSQLCompat(tx, `UPDATE pcs_outbox_events
			SET status = 'pending', attempts = 0, available_at = CURRENT_TIMESTAMP,
				locked_at = NULL, locked_by = NULL, lease_until = NULL,
				published_at = NULL, dead_at = NULL,
				last_error = 'reactivado manualmente con auditoria',
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND empresa_id = ? AND topic = ? AND status = 'dead'`,
			event.ID, empresaID, topic)
		if err != nil {
			return nil, err
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return nil, fmt.Errorf("el evento %d cambio durante la recuperacion", event.ID)
		}
		event.Status = OutboxPending
		event.Attempts = 0
		event.DeadAt = ""
		event.RecoveryAction = "requeued"
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return events, nil
}

func normalizeOutboxRecoveryInput(empresaID int64, topic string, eventIDs []int64, limit int, requireIDs bool) ([]int64, string, int, error) {
	topic = strings.TrimSpace(topic)
	if empresaID <= 0 {
		return nil, "", 0, fmt.Errorf("empresa_id es obligatorio")
	}
	if topic == "" || len(topic) > 160 {
		return nil, "", 0, fmt.Errorf("topic es obligatorio")
	}

	seen := make(map[int64]struct{}, len(eventIDs))
	ids := make([]int64, 0, len(eventIDs))
	for _, id := range eventIDs {
		if id <= 0 {
			return nil, "", 0, fmt.Errorf("los ids de evento deben ser positivos")
		}
		if _, exists := seen[id]; exists {
			if requireIDs {
				return nil, "", 0, fmt.Errorf("event_ids no puede contener duplicados")
			}
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	if requireIDs && len(ids) == 0 {
		return nil, "", 0, fmt.Errorf("event_ids es obligatorio")
	}
	if len(ids) > MaxOutboxRecoveryEvents {
		return nil, "", 0, fmt.Errorf("maximo %d eventos por recuperacion", MaxOutboxRecoveryEvents)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	if limit <= 0 {
		limit = MaxOutboxRecoveryEvents
	}
	if limit > MaxOutboxRecoveryEvents {
		limit = MaxOutboxRecoveryEvents
	}
	if len(ids) > 0 && limit < len(ids) {
		limit = len(ids)
	}
	return ids, topic, limit, nil
}

func outboxRecoveryIDClause(ids []int64) (string, []interface{}) {
	if len(ids) == 0 {
		return "", nil
	}
	placeholders := make([]string, 0, len(ids))
	args := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	return " AND id IN (" + strings.Join(placeholders, ",") + ")", args
}
