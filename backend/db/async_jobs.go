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

	defaultAsyncJobLease       = 5 * time.Minute
	asyncJobsSchemaFingerprint = "pcs_async_jobs:v2:lease-heartbeat-priority-idempotency-cancel-dead-letter"
)

// AsyncJob is the durable, tenant-aware contract shared by asynchronous
// integrations. Payloads contain only references and minimum operational
// context; credentials, raw user tokens and business secrets are invalid.
type AsyncJob struct {
	ID                 int64
	EmpresaID          int64
	Kind               string
	Version            int
	PayloadJSON        string
	Status             string
	Attempts           int
	MaxAttempts        int
	Priority           int
	AvailableAt        time.Time
	LeaseUntil         time.Time
	HeartbeatAt        time.Time
	CompletedAt        time.Time
	CancelledAt        time.Time
	DeadAt             time.Time
	IdempotencyKey     string
	IdempotencyKeyHash string
}

type AsyncJobMetrics struct {
	Pending    int64
	Processing int64
	Completed  int64
	Dead       int64
	Cancelled  int64
}

func ValidateAsyncJob(job AsyncJob) error {
	if job.EmpresaID < 0 {
		return fmt.Errorf("empresa_id invalid")
	}
	if strings.TrimSpace(job.Kind) == "" || len(job.Kind) > 120 {
		return fmt.Errorf("job kind invalid")
	}
	if job.Version < 0 || job.Version > 1000 {
		return fmt.Errorf("job version invalid")
	}
	if len(job.PayloadJSON) > 1<<20 {
		return fmt.Errorf("job payload exceeds 1 MiB")
	}
	if job.MaxAttempts != 0 && (job.MaxAttempts < 1 || job.MaxAttempts > 25) {
		return fmt.Errorf("max attempts must be between 1 and 25")
	}
	if job.Priority < 0 || job.Priority > 1000 {
		return fmt.Errorf("job priority invalid")
	}
	if len(strings.TrimSpace(job.IdempotencyKey)) > 512 {
		return fmt.Errorf("job idempotency key too long")
	}
	return nil
}

// EnsureAsyncJobsSchema is migration-role-only. API handlers and workers must
// fail closed when pcs-migrate has not applied this schema.
func EnsureAsyncJobsSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	for _, statement := range asyncJobsSchemaStatements() {
		if _, err := execSQLCompat(dbConn, statement); err != nil {
			return err
		}
	}
	return nil
}

// VerifyAsyncJobsSchema performs a read-only readiness check for API and
// worker roles. It deliberately does not attempt CREATE/ALTER recovery.
func VerifyAsyncJobsSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	var tableName sql.NullString
	if err := queryRowSQLCompat(dbConn, `SELECT to_regclass('pcs_async_jobs')`).Scan(&tableName); err != nil {
		return err
	}
	if !tableName.Valid || strings.TrimSpace(tableName.String) == "" {
		return fmt.Errorf("pcs_async_jobs schema missing; run pcs-migrate")
	}
	return nil
}

func applyAsyncJobsSchemaTx(tx *sql.Tx) error {
	if tx == nil {
		return fmt.Errorf("migration transaction is required")
	}
	for _, statement := range asyncJobsSchemaStatements() {
		if _, err := execTxSQLCompat(tx, statement); err != nil {
			return err
		}
	}
	return nil
}

func asyncJobsSchemaStatements() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS pcs_async_jobs (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL DEFAULT 0,
			kind TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			payload_json TEXT NOT NULL DEFAULT '{}',
			status TEXT NOT NULL DEFAULT 'pending',
			attempts INTEGER NOT NULL DEFAULT 0,
			max_attempts INTEGER NOT NULL DEFAULT 5,
			priority INTEGER NOT NULL DEFAULT 100,
			idempotency_key_hash TEXT,
			available_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			locked_at TIMESTAMPTZ,
			locked_by TEXT,
			lease_until TIMESTAMPTZ,
			heartbeat_at TIMESTAMPTZ,
			cancelled_at TIMESTAMPTZ,
			completed_at TIMESTAMPTZ,
			dead_at TIMESTAMPTZ,
			last_error TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CHECK (status IN ('pending','processing','completed','dead')),
			CHECK (max_attempts BETWEEN 1 AND 25),
			CHECK (priority BETWEEN 0 AND 1000)
		)`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 100`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS idempotency_key_hash TEXT`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS lease_until TIMESTAMPTZ`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS heartbeat_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS completed_at TIMESTAMPTZ`,
		`ALTER TABLE pcs_async_jobs ADD COLUMN IF NOT EXISTS dead_at TIMESTAMPTZ`,
		`CREATE INDEX IF NOT EXISTS ix_pcs_async_jobs_ready_v2
			ON pcs_async_jobs (priority DESC, available_at, id)
			WHERE status = 'pending' AND cancelled_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS ix_pcs_async_jobs_lease_v2
			ON pcs_async_jobs (lease_until, id)
			WHERE status = 'processing'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_pcs_async_jobs_idempotency_v2
			ON pcs_async_jobs (empresa_id, kind, idempotency_key_hash)
			WHERE idempotency_key_hash IS NOT NULL`,
	}
}

// EnqueueAsyncJob preserves the historical fire-and-forget signature. Callers
// that can retry must set IdempotencyKey and use EnqueueAsyncJobIdempotent.
func EnqueueAsyncJob(dbConn *sql.DB, job AsyncJob) error {
	_, _, err := EnqueueAsyncJobIdempotent(dbConn, job)
	return err
}

// EnqueueAsyncJobIdempotent returns the stable row for repeated logical work.
// The raw key is hashed before persistence and never included in logs.
func EnqueueAsyncJobIdempotent(dbConn *sql.DB, job AsyncJob) (*AsyncJob, bool, error) {
	if dbConn == nil {
		return nil, false, fmt.Errorf("database not available")
	}
	if job.MaxAttempts == 0 {
		job.MaxAttempts = 5
	}
	if job.Version == 0 {
		job.Version = 1
	}
	if job.Priority == 0 {
		job.Priority = 100
	}
	if err := ValidateAsyncJob(job); err != nil {
		return nil, false, err
	}
	if job.AvailableAt.IsZero() {
		job.AvailableAt = time.Now().UTC()
	}
	job.Kind = strings.TrimSpace(job.Kind)
	job.IdempotencyKeyHash = hashAsyncJobKey(job.IdempotencyKey)
	result, err := execSQLCompat(dbConn, `INSERT INTO pcs_async_jobs (
		empresa_id, kind, version, payload_json, status, attempts, max_attempts, priority,
		idempotency_key_hash, available_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, 'pending', 0, ?, ?, NULLIF(?, ''), ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT DO NOTHING`, job.EmpresaID, job.Kind, job.Version, job.PayloadJSON, job.MaxAttempts, job.Priority,
		job.IdempotencyKeyHash, job.AvailableAt)
	if err != nil {
		return nil, false, err
	}
	if affected, _ := result.RowsAffected(); affected == 1 {
		if job.IdempotencyKeyHash != "" {
			stored, getErr := getAsyncJobByIdempotencyKey(dbConn, job.EmpresaID, job.Kind, job.IdempotencyKeyHash)
			if getErr == nil {
				return stored, true, nil
			}
		}
		return &job, true, nil
	}
	if job.IdempotencyKeyHash == "" {
		return nil, false, fmt.Errorf("async job was not created")
	}
	stored, err := getAsyncJobByIdempotencyKey(dbConn, job.EmpresaID, job.Kind, job.IdempotencyKeyHash)
	if err != nil {
		return nil, false, err
	}
	return stored, false, nil
}

// ClaimAsyncJobs preserves the old API while adding a bounded lease to every
// claim. Workers must finish or renew the lease before it expires.
func ClaimAsyncJobs(dbConn *sql.DB, workerID string, limit int) ([]AsyncJob, error) {
	return ClaimAsyncJobsWithLease(dbConn, workerID, limit, defaultAsyncJobLease)
}

func ClaimAsyncJobsWithLease(dbConn *sql.DB, workerID string, limit int, lease time.Duration) ([]AsyncJob, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("database not available")
	}
	if strings.TrimSpace(workerID) == "" {
		return nil, fmt.Errorf("worker id required")
	}
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("claim limit must be between 1 and 100")
	}
	if lease < time.Second || lease > 30*time.Minute {
		return nil, fmt.Errorf("job lease must be between one second and thirty minutes")
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := execTxSQLCompat(tx, `UPDATE pcs_async_jobs
		SET status = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'pending' END,
			locked_at = NULL, locked_by = NULL, lease_until = NULL, heartbeat_at = NULL,
			dead_at = CASE WHEN attempts >= max_attempts THEN CURRENT_TIMESTAMP ELSE dead_at END,
			available_at = CASE WHEN attempts >= max_attempts THEN available_at ELSE CURRENT_TIMESTAMP END,
			last_error = CASE WHEN last_error IS NULL OR last_error = '' THEN 'worker lease expired' ELSE last_error END,
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'processing' AND lease_until IS NOT NULL AND lease_until < CURRENT_TIMESTAMP`); err != nil {
		return nil, err
	}
	rows, err := queryTxSQLCompat(tx, `SELECT id, empresa_id, kind, version, payload_json, status, attempts, max_attempts,
		priority, available_at, COALESCE(lease_until, CURRENT_TIMESTAMP), COALESCE(heartbeat_at, CURRENT_TIMESTAMP),
		COALESCE(completed_at, CURRENT_TIMESTAMP), COALESCE(cancelled_at, CURRENT_TIMESTAMP), COALESCE(dead_at, CURRENT_TIMESTAMP),
		COALESCE(idempotency_key_hash, '')
		FROM pcs_async_jobs
		WHERE status = 'pending' AND cancelled_at IS NULL AND available_at <= CURRENT_TIMESTAMP
		ORDER BY priority DESC, available_at, id
		FOR UPDATE SKIP LOCKED
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]AsyncJob, 0, limit)
	for rows.Next() {
		var job AsyncJob
		if err := rows.Scan(&job.ID, &job.EmpresaID, &job.Kind, &job.Version, &job.PayloadJSON, &job.Status,
			&job.Attempts, &job.MaxAttempts, &job.Priority, &job.AvailableAt, &job.LeaseUntil, &job.HeartbeatAt,
			&job.CompletedAt, &job.CancelledAt, &job.DeadAt, &job.IdempotencyKeyHash); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range jobs {
		job := &jobs[index]
		if _, err := execTxSQLCompat(tx, `UPDATE pcs_async_jobs
			SET status = 'processing', attempts = attempts + 1, locked_at = CURRENT_TIMESTAMP, locked_by = ?,
				lease_until = CURRENT_TIMESTAMP + (? * interval '1 millisecond'), heartbeat_at = CURRENT_TIMESTAMP,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND status = 'pending' AND cancelled_at IS NULL`, strings.TrimSpace(workerID), lease.Milliseconds(), job.ID); err != nil {
			return nil, err
		}
		job.Status = AsyncJobProcessing
		job.Attempts++
		job.LeaseUntil = time.Now().UTC().Add(lease)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func RenewAsyncJobLease(dbConn *sql.DB, id int64, workerID string, lease time.Duration) error {
	if id <= 0 || strings.TrimSpace(workerID) == "" || lease < time.Second || lease > 30*time.Minute {
		return fmt.Errorf("async job lease input invalid")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET lease_until = CURRENT_TIMESTAMP + (? * interval '1 millisecond'), heartbeat_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ? AND cancelled_at IS NULL`, lease.Milliseconds(), id, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job not owned by worker")
	}
	return nil
}

func CompleteAsyncJob(dbConn *sql.DB, id int64, workerID string) error {
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = 'completed', locked_at = NULL, locked_by = NULL, lease_until = NULL, heartbeat_at = NULL,
			last_error = NULL, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ? AND cancelled_at IS NULL`, id, strings.TrimSpace(workerID))
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
	availableAt := time.Now().UTC().Add(retryAfter)
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = ?, available_at = ?, locked_at = NULL, locked_by = NULL, lease_until = NULL, heartbeat_at = NULL,
			last_error = ?, dead_at = CASE WHEN ? = 'dead' THEN CURRENT_TIMESTAMP ELSE dead_at END,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ? AND cancelled_at IS NULL`,
		status, availableAt, redactAsyncJobError(cause), status, job.ID, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job not owned by worker")
	}
	return nil
}

// FailAsyncJob moves a claimed job straight to the durable dead-letter state.
// It is used for an unsupported kind or version, which cannot become valid by
// repeating the same payload without an explicit operator or deployment change.
func FailAsyncJob(dbConn *sql.DB, job AsyncJob, workerID string, cause error) error {
	if job.ID <= 0 || strings.TrimSpace(workerID) == "" {
		return fmt.Errorf("async job ownership invalid")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = 'dead', locked_at = NULL, locked_by = NULL, lease_until = NULL, heartbeat_at = NULL,
			dead_at = CURRENT_TIMESTAMP, last_error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, redactAsyncJobError(cause), job.ID, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job not owned by worker")
	}
	return nil
}

func CancelAsyncJob(dbConn *sql.DB, id, empresaID int64) error {
	if id <= 0 || empresaID < 0 {
		return fmt.Errorf("async job cancel input invalid")
	}
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET cancelled_at = CURRENT_TIMESTAMP, locked_at = NULL, locked_by = NULL, lease_until = NULL,
			heartbeat_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND empresa_id = ? AND status IN ('pending', 'processing') AND cancelled_at IS NULL`, id, empresaID)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job cannot be cancelled")
	}
	return nil
}

func RecoverExpiredAsyncJobs(dbConn *sql.DB) (int64, error) {
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = CASE WHEN attempts >= max_attempts THEN 'dead' ELSE 'pending' END,
			locked_at = NULL, locked_by = NULL, lease_until = NULL, heartbeat_at = NULL,
			dead_at = CASE WHEN attempts >= max_attempts THEN CURRENT_TIMESTAMP ELSE dead_at END,
			available_at = CASE WHEN attempts >= max_attempts THEN available_at ELSE CURRENT_TIMESTAMP END,
			last_error = CASE WHEN last_error IS NULL OR last_error = '' THEN 'worker lease expired' ELSE last_error END,
			updated_at = CURRENT_TIMESTAMP
		WHERE status = 'processing' AND cancelled_at IS NULL
			AND lease_until IS NOT NULL AND lease_until < CURRENT_TIMESTAMP`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetAsyncJobMetrics(dbConn *sql.DB) (AsyncJobMetrics, error) {
	var result AsyncJobMetrics
	err := queryRowSQLCompat(dbConn, `SELECT
		COUNT(*) FILTER (WHERE status = 'pending' AND cancelled_at IS NULL),
		COUNT(*) FILTER (WHERE status = 'processing' AND cancelled_at IS NULL),
		COUNT(*) FILTER (WHERE status = 'completed'),
		COUNT(*) FILTER (WHERE status = 'dead'),
		COUNT(*) FILTER (WHERE cancelled_at IS NOT NULL)
		FROM pcs_async_jobs`).Scan(&result.Pending, &result.Processing, &result.Completed, &result.Dead, &result.Cancelled)
	return result, err
}

func getAsyncJobByIdempotencyKey(dbConn *sql.DB, empresaID int64, kind, keyHash string) (*AsyncJob, error) {
	var job AsyncJob
	err := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, kind, version, payload_json, status, attempts, max_attempts,
		priority, available_at, COALESCE(idempotency_key_hash, '')
		FROM pcs_async_jobs WHERE empresa_id = ? AND kind = ? AND idempotency_key_hash = ?`, empresaID, kind, keyHash).
		Scan(&job.ID, &job.EmpresaID, &job.Kind, &job.Version, &job.PayloadJSON, &job.Status, &job.Attempts,
			&job.MaxAttempts, &job.Priority, &job.AvailableAt, &job.IdempotencyKeyHash)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func hashAsyncJobKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func redactAsyncJobError(cause error) string {
	if cause == nil {
		return "worker failed"
	}
	message := strings.TrimSpace(cause.Error())
	if message == "" {
		return "worker failed"
	}
	if len(message) > 500 {
		message = message[:500]
	}
	lower := strings.ToLower(message)
	for _, marker := range []string{"password=", "token=", "secret=", "authorization:", "postgres://", "postgresql://"} {
		if index := strings.Index(lower, marker); index >= 0 {
			return strings.TrimSpace(message[:index]) + "[redacted]"
		}
	}
	return message
}
