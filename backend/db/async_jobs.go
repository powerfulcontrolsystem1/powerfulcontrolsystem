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
	AsyncJobPending    = "pending"
	AsyncJobProcessing = "processing"
	AsyncJobCompleted  = "completed"
	AsyncJobDead       = "dead"
	AsyncJobCanceled   = "canceled"
)

// AsyncJob is the durable, tenant-aware contract shared by asynchronous
// integrations. Payloads must contain only the minimum data required by the
// worker; credentials and raw user tokens are never valid payload fields.
type AsyncJob struct {
	ID             int64
	EmpresaID      int64
	OriginUserID   int64
	Kind           string
	PayloadVersion int
	PayloadJSON    string
	Status         string
	Priority       int
	Attempts       int
	MaxAttempts    int
	AvailableAt    time.Time
	ExpiresAt      *time.Time
	StartedAt      *time.Time
	FinishedAt     *time.Time
	HeartbeatAt    *time.Time
	CorrelationID  string
	IdempotencyKey string
	ResultJSON     string
}

func ValidateAsyncJob(job AsyncJob) error {
	if job.EmpresaID < 0 {
		return fmt.Errorf("empresa_id invalid")
	}
	if strings.TrimSpace(job.Kind) == "" || len(job.Kind) > 120 {
		return fmt.Errorf("job kind invalid")
	}
	if job.PayloadVersion < 0 || job.PayloadVersion > 1000 {
		return fmt.Errorf("job payload version invalid")
	}
	if job.Priority < -100 || job.Priority > 100 {
		return fmt.Errorf("job priority invalid")
	}
	if len(job.ResultJSON) > 1<<20 {
		return fmt.Errorf("job result exceeds 1 MiB")
	}
	if len(job.PayloadJSON) > 1<<20 {
		return fmt.Errorf("job payload exceeds 1 MiB")
	}
	if job.MaxAttempts < 1 || job.MaxAttempts > 25 {
		return fmt.Errorf("max attempts must be between 1 and 25")
	}
	if job.OriginUserID < 0 {
		return fmt.Errorf("origin user id invalid")
	}
	if len(strings.TrimSpace(job.CorrelationID)) > 120 {
		return fmt.Errorf("correlation id too long")
	}
	if key := strings.TrimSpace(job.IdempotencyKey); key != "" && (len(key) < 16 || len(key) > 200) {
		return fmt.Errorf("job idempotency key must be between 16 and 200 characters")
	}
	return nil
}

// EnsureAsyncJobsSchema is deliberately invoked by the migration role. API
// handlers only enqueue work after deployment has applied the schema.
func EnsureAsyncJobsSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	statement := `CREATE TABLE IF NOT EXISTS pcs_async_jobs (
		id BIGSERIAL PRIMARY KEY,
		empresa_id BIGINT NOT NULL DEFAULT 0,
		kind TEXT NOT NULL,
		payload_version INTEGER NOT NULL DEFAULT 1,
		payload_json TEXT NOT NULL DEFAULT '{}',
		status TEXT NOT NULL DEFAULT 'pending',
		priority INTEGER NOT NULL DEFAULT 0,
		attempts INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 5,
		available_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		locked_at TIMESTAMPTZ,
		locked_by TEXT,
		heartbeat_at TIMESTAMPTZ,
		started_at TIMESTAMPTZ,
		finished_at TIMESTAMPTZ,
		last_error TEXT,
		result_json TEXT NOT NULL DEFAULT '',
		origin_user_id BIGINT NOT NULL DEFAULT 0,
		correlation_id TEXT NOT NULL DEFAULT '',
		idempotency_key_hash TEXT,
		expires_at TIMESTAMPTZ,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CHECK (status IN ('pending','processing','completed','dead','canceled')),
		CHECK (max_attempts BETWEEN 1 AND 25),
		CHECK (priority BETWEEN -100 AND 100),
		CHECK (payload_version BETWEEN 0 AND 1000)
	)`
	if _, err := execSQLCompat(dbConn, statement); err != nil {
		return err
	}
	for _, statement := range []string{
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS origin_user_id BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS payload_version INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS correlation_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS idempotency_key_hash TEXT`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS heartbeat_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS finished_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS result_json TEXT NOT NULL DEFAULT ''`,
	} {
		if _, err := execSQLCompat(dbConn, statement); err != nil {
			return err
		}
	}
	_, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_pcs_async_jobs_ready
		ON pcs_async_jobs (status, priority DESC, available_at, id)`)
	if err != nil {
		return err
	}
	_, err = execSQLCompat(dbConn, `CREATE UNIQUE INDEX IF NOT EXISTS ux_pcs_async_jobs_idempotency
		ON pcs_async_jobs (empresa_id, kind, idempotency_key_hash)
		WHERE idempotency_key_hash IS NOT NULL`)
	return err
}

func EnqueueAsyncJob(dbConn *sql.DB, job AsyncJob) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	if job.MaxAttempts == 0 {
		job.MaxAttempts = 5
	}
	if err := ValidateAsyncJob(job); err != nil {
		return err
	}
	if job.AvailableAt.IsZero() {
		job.AvailableAt = time.Now().UTC()
	}
	var idempotencyHash interface{}
	if key := strings.TrimSpace(job.IdempotencyKey); key != "" {
		idempotencyHash = asyncJobIdempotencyHash(key)
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO pcs_async_jobs (
		empresa_id, origin_user_id, kind, payload_version, payload_json, status, priority, attempts, max_attempts,
		available_at, correlation_id, idempotency_key_hash, expires_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, 'pending', ?, 0, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT (empresa_id, kind, idempotency_key_hash) WHERE idempotency_key_hash IS NOT NULL DO NOTHING`,
		job.EmpresaID, job.OriginUserID, strings.TrimSpace(job.Kind), normalizedPayloadVersion(job.PayloadVersion), job.PayloadJSON, job.Priority, job.MaxAttempts,
		job.AvailableAt, strings.TrimSpace(job.CorrelationID), idempotencyHash, job.ExpiresAt)
	return err
}

func normalizedPayloadVersion(version int) int {
	if version == 0 {
		return 1
	}
	return version
}

func asyncJobIdempotencyHash(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

// ClaimAsyncJobs uses SKIP LOCKED so multiple worker replicas never process
// the same row concurrently. Only PostgreSQL is supported by PCS runtime.
func ClaimAsyncJobs(dbConn *sql.DB, workerID string, limit int) ([]AsyncJob, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("database not available")
	}
	if strings.TrimSpace(workerID) == "" {
		return nil, fmt.Errorf("worker id required")
	}
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("claim limit must be between 1 and 100")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	rows, err := tx.Query(rebindCompatQuery(`SELECT id, empresa_id, origin_user_id, kind, payload_version, payload_json, status, priority, attempts, max_attempts, available_at, expires_at, correlation_id
		FROM pcs_async_jobs
		WHERE status = 'pending' AND available_at <= CURRENT_TIMESTAMP
			AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
		ORDER BY priority DESC, available_at, id
		FOR UPDATE SKIP LOCKED
		LIMIT ?`), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]AsyncJob, 0, limit)
	for rows.Next() {
		var job AsyncJob
		if err := rows.Scan(&job.ID, &job.EmpresaID, &job.OriginUserID, &job.Kind, &job.PayloadVersion, &job.PayloadJSON, &job.Status, &job.Priority, &job.Attempts, &job.MaxAttempts, &job.AvailableAt, &job.ExpiresAt, &job.CorrelationID); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range jobs {
		job := &jobs[index]
		result, err := tx.Exec(rebindCompatQuery(`UPDATE pcs_async_jobs
			SET status = 'processing', attempts = attempts + 1, locked_at = CURRENT_TIMESTAMP, heartbeat_at = CURRENT_TIMESTAMP,
				started_at = COALESCE(started_at, CURRENT_TIMESTAMP), locked_by = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND status = 'pending'`), strings.TrimSpace(workerID), job.ID)
		if err != nil {
			return nil, err
		}
		if affected, _ := result.RowsAffected(); affected != 1 {
			return nil, fmt.Errorf("async job claim lost")
		}
		job.Status = AsyncJobProcessing
		job.Attempts++
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

// ExpireAsyncJobs keeps the queue from retrying work whose business deadline
// has passed. A dead job remains auditable and can be reviewed explicitly.
func ExpireAsyncJobs(dbConn *sql.DB) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("database not available")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = 'dead', locked_at = NULL, locked_by = NULL, finished_at = CURRENT_TIMESTAMP,
			last_error = 'job expired', updated_at = CURRENT_TIMESTAMP
		WHERE status IN ('pending', 'processing') AND expires_at IS NOT NULL AND expires_at <= CURRENT_TIMESTAMP`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// RecoverExpiredAsyncJobs returns jobs that were leased by a worker that did
// not finish. The job becomes pending again unless it has exhausted its retry
// budget, in which case it is marked dead for explicit operational review.
func RecoverExpiredAsyncJobs(dbConn *sql.DB, staleAfter time.Duration) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("database not available")
	}
	if staleAfter < time.Minute {
		staleAfter = 5 * time.Minute
	}
	cutoff := time.Now().UTC().Add(-staleAfter)
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'pending' END,
			available_at = CURRENT_TIMESTAMP,
			locked_at = NULL,
			locked_by = NULL,
			heartbeat_at = NULL,
			finished_at = CASE WHEN attempts >= max_attempts THEN CURRENT_TIMESTAMP ELSE NULL END,
			last_error = 'processing lease expired',
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'processing' AND COALESCE(heartbeat_at, locked_at) IS NOT NULL
			AND COALESCE(heartbeat_at, locked_at) < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

// ReleaseAsyncJobsForWorker makes a graceful worker shutdown immediately
// recoverable instead of waiting for the lease timeout.
func ReleaseAsyncJobsForWorker(dbConn *sql.DB, workerID string) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("database not available")
	}
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return 0, fmt.Errorf("worker id required")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = 'pending',
			available_at = CURRENT_TIMESTAMP,
			locked_at = NULL,
			locked_by = NULL,
			heartbeat_at = NULL,
			last_error = 'worker shutdown',
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'processing' AND locked_by = ?`, workerID)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func CompleteAsyncJob(dbConn *sql.DB, id int64, workerID string) error {
	return CompleteAsyncJobWithResult(dbConn, id, workerID, "")
}

// CompleteAsyncJobWithResult stores a bounded, non-sensitive result summary.
// Provider responses, tokens and document contents must never be persisted here.
func CompleteAsyncJobWithResult(dbConn *sql.DB, id int64, workerID, resultJSON string) error {
	if len(resultJSON) > 1<<20 {
		return fmt.Errorf("job result exceeds 1 MiB")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = 'completed', locked_at = NULL, locked_by = NULL, heartbeat_at = NULL,
			finished_at = CURRENT_TIMESTAMP, result_json = ?, last_error = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, resultJSON, id, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job not owned by worker")
	}
	return nil
}

func RetryAsyncJob(dbConn *sql.DB, job AsyncJob, workerID string, cause error, retryAfter time.Duration) error {
	if job.ID <= 0 || strings.TrimSpace(workerID) == "" {
		return fmt.Errorf("async job ownership invalid")
	}
	if retryAfter < time.Second {
		retryAfter = time.Second
	}
	status := AsyncJobPending
	if job.Attempts >= job.MaxAttempts {
		status = AsyncJobDead
	}
	// Provider errors may include request data or credentials. The durable job
	// ledger is operationally visible, so retain only a stable, non-sensitive
	// state message here. The handler remains responsible for secure logging.
	message := "worker failed"
	if cause != nil {
		message = "worker failed; retry scheduled"
	}
	availableAt := time.Now().UTC().Add(retryAfter)
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = ?, available_at = ?, locked_at = NULL, locked_by = NULL, heartbeat_at = NULL,
			finished_at = CASE WHEN ? = 'dead' THEN CURRENT_TIMESTAMP ELSE NULL END,
			last_error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, status, availableAt, status, message, job.ID, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job not owned by worker")
	}
	return nil
}

// HeartbeatAsyncJob extends the lease only for its owning worker. It is safe to
// call periodically while a provider request or document render is running.
func HeartbeatAsyncJob(dbConn *sql.DB, id int64, workerID string) error {
	if id <= 0 || strings.TrimSpace(workerID) == "" {
		return fmt.Errorf("async job ownership invalid")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET locked_at = CURRENT_TIMESTAMP, heartbeat_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, id, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job not owned by worker")
	}
	return nil
}

// CancelAsyncJob only cancels work that has not started. An active provider
// call must finish or be explicitly compensated by its domain handler.
func CancelAsyncJob(dbConn *sql.DB, empresaID, id int64) error {
	if empresaID < 0 || id <= 0 {
		return fmt.Errorf("async job cancellation input invalid")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = 'canceled', finished_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND empresa_id = ? AND status = 'pending'`, id, empresaID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job cannot be canceled")
	}
	return nil
}

// AsyncJobStats supplies low-cardinality operational counts for health and
// metrics without exposing payloads or tenant information.
type AsyncJobStats struct {
	Pending    int64
	Processing int64
	Completed  int64
	Dead       int64
	Canceled   int64
}

func GetAsyncJobStats(dbConn *sql.DB) (AsyncJobStats, error) {
	if dbConn == nil {
		return AsyncJobStats{}, fmt.Errorf("database not available")
	}
	var stats AsyncJobStats
	err := queryRowSQLCompat(dbConn, `SELECT
		COUNT(*) FILTER (WHERE status = 'pending'),
		COUNT(*) FILTER (WHERE status = 'processing'),
		COUNT(*) FILTER (WHERE status = 'completed'),
		COUNT(*) FILTER (WHERE status = 'dead'),
		COUNT(*) FILTER (WHERE status = 'canceled')
		FROM pcs_async_jobs`).Scan(&stats.Pending, &stats.Processing, &stats.Completed, &stats.Dead, &stats.Canceled)
	return stats, err
}
