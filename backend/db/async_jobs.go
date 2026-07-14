package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const (
	AsyncJobPending    = "pending"
	AsyncJobProcessing = "processing"
	AsyncJobCompleted  = "completed"
	AsyncJobDead       = "dead"
)

// AsyncJob is the durable, tenant-aware contract shared by asynchronous
// integrations. Payloads must contain only the minimum data required by the
// worker; credentials and raw user tokens are never valid payload fields.
type AsyncJob struct {
	ID          int64
	EmpresaID   int64
	Kind        string
	PayloadJSON string
	Status      string
	Attempts    int
	MaxAttempts int
	AvailableAt time.Time
}

func ValidateAsyncJob(job AsyncJob) error {
	if job.EmpresaID < 0 {
		return fmt.Errorf("empresa_id invalid")
	}
	if strings.TrimSpace(job.Kind) == "" || len(job.Kind) > 120 {
		return fmt.Errorf("job kind invalid")
	}
	if len(job.PayloadJSON) > 1<<20 {
		return fmt.Errorf("job payload exceeds 1 MiB")
	}
	if job.MaxAttempts < 1 || job.MaxAttempts > 25 {
		return fmt.Errorf("max attempts must be between 1 and 25")
	}
	return nil
}

// EnsureAsyncJobsSchema is deliberately invoked by the migration role. API
// handlers only enqueue work after deployment has applied the schema.
func EnsureAsyncJobsSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	statement := `CREATE TABLE IF NOT EXISTS pcs_async_jobs (
		id BIGSERIAL PRIMARY KEY,
		empresa_id BIGINT NOT NULL DEFAULT 0,
		kind TEXT NOT NULL,
		payload_json TEXT NOT NULL DEFAULT '{}',
		status TEXT NOT NULL DEFAULT 'pending',
		attempts INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 5,
		available_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		locked_at TIMESTAMPTZ,
		locked_by TEXT,
		last_error TEXT,
		created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		CHECK (status IN ('pending','processing','completed','dead')),
		CHECK (max_attempts BETWEEN 1 AND 25)
	)`
	if _, err := execSQLCompat(dbConn, statement); err != nil {
		return err
	}
	_, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_pcs_async_jobs_ready
		ON pcs_async_jobs (status, available_at, id)`)
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
	_, err := execSQLCompat(dbConn, `INSERT INTO pcs_async_jobs (
		empresa_id, kind, payload_json, status, attempts, max_attempts, available_at, created_at, updated_at
	) VALUES (?, ?, ?, 'pending', 0, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		job.EmpresaID, strings.TrimSpace(job.Kind), job.PayloadJSON, job.MaxAttempts, job.AvailableAt)
	return err
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
	rows, err := tx.Query(rebindCompatQuery(`SELECT id, empresa_id, kind, payload_json, status, attempts, max_attempts, available_at
		FROM pcs_async_jobs
		WHERE status = 'pending' AND available_at <= CURRENT_TIMESTAMP
		ORDER BY available_at, id
		FOR UPDATE SKIP LOCKED
		LIMIT ?`), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]AsyncJob, 0, limit)
	for rows.Next() {
		var job AsyncJob
		if err := rows.Scan(&job.ID, &job.EmpresaID, &job.Kind, &job.PayloadJSON, &job.Status, &job.Attempts, &job.MaxAttempts, &job.AvailableAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for index := range jobs {
		job := &jobs[index]
		if _, err := tx.Exec(rebindCompatQuery(`UPDATE pcs_async_jobs
			SET status = 'processing', attempts = attempts + 1, locked_at = CURRENT_TIMESTAMP, locked_by = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ? AND status = 'pending'`), strings.TrimSpace(workerID), job.ID); err != nil {
			return nil, err
		}
		job.Status = AsyncJobProcessing
		job.Attempts++
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func CompleteAsyncJob(dbConn *sql.DB, id int64, workerID string) error {
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = 'completed', locked_at = NULL, locked_by = NULL, last_error = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, id, strings.TrimSpace(workerID))
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
	message := "worker failed"
	if cause != nil {
		message = cause.Error()
	}
	if len(message) > 1000 {
		message = message[:1000]
	}
	availableAt := time.Now().UTC().Add(retryAfter)
	result, err := execSQLCompat(dbConn, `UPDATE pcs_async_jobs
		SET status = ?, available_at = ?, locked_at = NULL, locked_by = NULL, last_error = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'processing' AND locked_by = ?`, status, availableAt, message, job.ID, strings.TrimSpace(workerID))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return fmt.Errorf("async job not owned by worker")
	}
	return nil
}
