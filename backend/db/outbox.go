package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// OutboxEvent is written in the same transaction as the business change. A
// worker later converts it into external work, preventing lost notifications
// when a request commits but the provider is temporarily unavailable.
type OutboxEvent struct {
	ID          int64
	EmpresaID   int64
	Topic       string
	PayloadJSON string
}

func EnsureOutboxSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	_, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS pcs_outbox_events (
		id BIGSERIAL PRIMARY KEY,
		empresa_id BIGINT NOT NULL DEFAULT 0,
		topic TEXT NOT NULL,
		payload_json TEXT NOT NULL DEFAULT '{}',
		published_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CHECK (char_length(topic) BETWEEN 1 AND 160)
	)`)
	if err != nil {
		return err
	}
	_, err = execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_pcs_outbox_events_pending
		ON pcs_outbox_events (published_at, id)`)
	return err
}

func InsertOutboxEvent(tx *sql.Tx, event OutboxEvent) error {
	if tx == nil {
		return fmt.Errorf("transaction required")
	}
	if event.EmpresaID < 0 || strings.TrimSpace(event.Topic) == "" || len(event.Topic) > 160 {
		return fmt.Errorf("outbox event invalid")
	}
	if len(event.PayloadJSON) > 1<<20 {
		return fmt.Errorf("outbox payload exceeds 1 MiB")
	}
	_, err := tx.Exec(rebindCompatQuery(`INSERT INTO pcs_outbox_events (empresa_id, topic, payload_json)
		VALUES (?, ?, ?)`), event.EmpresaID, strings.TrimSpace(event.Topic), event.PayloadJSON)
	return err
}
