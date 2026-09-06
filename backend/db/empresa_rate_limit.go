package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const empresaRateLimitSchemaFingerprint = "empresa-api-rate-limit:v1:tenant-scope-window-counter"

func applyEmpresaRateLimitSchemaTx(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS empresa_api_rate_limits (
		empresa_id BIGINT NOT NULL,
		scope TEXT NOT NULL,
		window_start TIMESTAMPTZ NOT NULL,
		request_count BIGINT NOT NULL DEFAULT 0,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (empresa_id, scope)
	)`)
	return err
}

// ConsumeEmpresaRateLimit increments a tenant quota atomically in PostgreSQL.
// A single row per tenant/scope keeps memory bounded and coordinates replicas.
func ConsumeEmpresaRateLimit(ctx context.Context, dbConn *sql.DB, empresaID int64, scope string, limit int64, now time.Time) (allowed bool, remaining, retryAfter, current int64, err error) {
	if dbConn == nil || empresaID <= 0 || limit <= 0 {
		return true, 0, 0, 0, nil
	}
	scope = strings.TrimSpace(strings.ToLower(scope))
	if scope == "" {
		scope = "api"
	}
	if now.IsZero() {
		now = time.Now()
	}
	windowStart := now.UTC().Truncate(time.Minute)
	var storedWindow time.Time
	err = dbConn.QueryRowContext(ctx, `
		INSERT INTO empresa_api_rate_limits (empresa_id, scope, window_start, request_count, updated_at)
		VALUES ($1, $2, $3, 1, CURRENT_TIMESTAMP)
		ON CONFLICT (empresa_id, scope) DO UPDATE SET
			window_start = EXCLUDED.window_start,
			request_count = CASE
				WHEN empresa_api_rate_limits.window_start = EXCLUDED.window_start
				THEN empresa_api_rate_limits.request_count + 1
				ELSE 1
			END,
			updated_at = CURRENT_TIMESTAMP
		RETURNING request_count, window_start`, empresaID, scope, windowStart).Scan(&current, &storedWindow)
	if err != nil {
		return false, 0, 0, 0, fmt.Errorf("consume tenant rate limit: %w", err)
	}
	remaining = limit - current
	if remaining < 0 {
		remaining = 0
	}
	if current > limit {
		retryAfter = int64(storedWindow.Add(time.Minute).Sub(now.UTC()).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, remaining, retryAfter, current, nil
	}
	return true, remaining, 0, current, nil
}
