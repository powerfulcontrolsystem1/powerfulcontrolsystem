// Package worker runs durable background work. It owns no HTTP state and never
// creates schema; pcs-migrate must have applied the required tables first.
package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type Handler func(context.Context, dbpkg.AsyncJob) error

type HandlerSpec struct {
	Kind        string
	Version     int
	Timeout     time.Duration
	MaxAttempts int
	Enabled     bool
	Handle      Handler
}

type Runner struct {
	DB          *sql.DB
	WorkerID    string
	Poll        time.Duration
	Batch       int
	Lease       time.Duration
	Handlers    map[string]HandlerSpec
	BeforeBatch func(context.Context) error
	Health      *HealthState
}

func (r *Runner) Run(ctx context.Context) error {
	if r.DB == nil {
		return fmt.Errorf("worker database unavailable")
	}
	if strings.TrimSpace(r.WorkerID) == "" {
		return fmt.Errorf("worker id required")
	}
	if r.Poll < time.Second {
		r.Poll = 2 * time.Second
	}
	if r.Batch < 1 || r.Batch > 100 {
		r.Batch = 20
	}
	if r.Lease < 30*time.Second || r.Lease > 30*time.Minute {
		r.Lease = 5 * time.Minute
	}
	if err := validateHandlerRegistry(r.Handlers); err != nil {
		return err
	}
	if err := r.runBatch(ctx); err != nil {
		r.markBatchFailure()
		log.Printf("async worker initial batch failed: %v", err)
	} else {
		r.markBatchSuccess()
	}
	ticker := time.NewTicker(r.Poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := r.runBatch(ctx); err != nil {
				r.markBatchFailure()
				log.Printf("async worker batch failed: %v", err)
			} else {
				r.markBatchSuccess()
			}
		}
	}
}

func (r *Runner) runBatch(ctx context.Context) error {
	if _, err := dbpkg.RecoverExpiredAsyncJobs(r.DB); err != nil {
		return fmt.Errorf("recover expired jobs: %w", err)
	}
	if r.BeforeBatch != nil {
		if err := r.BeforeBatch(ctx); err != nil {
			return fmt.Errorf("dispatch durable events: %w", err)
		}
	}
	jobs, err := dbpkg.ClaimAsyncJobsWithLease(r.DB, r.WorkerID, r.Batch, r.Lease)
	if err != nil || len(jobs) == 0 {
		return err
	}
	for _, job := range jobs {
		if err := r.runJob(ctx, job); err != nil {
			log.Printf("async worker job retry scheduled: kind=%s version=%d job_id=%d err=%v", job.Kind, job.Version, job.ID, err)
		}
	}
	return nil
}

func (r *Runner) runJob(parent context.Context, job dbpkg.AsyncJob) error {
	spec, found := r.Handlers[job.Kind]
	if !found {
		return dbpkg.FailAsyncJob(r.DB, job, r.WorkerID, fmt.Errorf("unsupported async job kind"))
	}
	if spec.Version != job.Version {
		return dbpkg.FailAsyncJob(r.DB, job, r.WorkerID, fmt.Errorf("unsupported async job version"))
	}
	if !spec.Enabled {
		return dbpkg.RetryAsyncJob(r.DB, job, r.WorkerID, fmt.Errorf("async job kind is temporarily disabled"), time.Minute)
	}
	jobCtx, cancel := context.WithTimeout(parent, spec.Timeout)
	defer cancel()
	stopLease := make(chan struct{})
	var leaseWG sync.WaitGroup
	leaseWG.Add(1)
	go func() {
		defer leaseWG.Done()
		r.maintainLease(jobCtx, stopLease, job)
	}()
	err := spec.Handle(jobCtx, job)
	close(stopLease)
	leaseWG.Wait()
	if err != nil {
		backoff := retryBackoff(job.Attempts)
		if jobCtx.Err() != nil {
			err = fmt.Errorf("job deadline exceeded")
		}
		if retryErr := dbpkg.RetryAsyncJob(r.DB, job, r.WorkerID, err, backoff); retryErr != nil {
			return retryErr
		}
		return err
	}
	return dbpkg.CompleteAsyncJob(r.DB, job.ID, r.WorkerID)
}

func (r *Runner) maintainLease(ctx context.Context, stop <-chan struct{}, job dbpkg.AsyncJob) {
	interval := r.Lease / 3
	if interval < 10*time.Second {
		interval = 10 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-stop:
			return
		case <-ticker.C:
			if err := dbpkg.RenewAsyncJobLease(r.DB, job.ID, r.WorkerID, r.Lease); err != nil {
				log.Printf("async worker lease renewal failed: job_id=%d kind=%s err=%v", job.ID, job.Kind, err)
				return
			}
		}
	}
}

func retryBackoff(attempt int) time.Duration {
	backoffs := [...]time.Duration{
		time.Minute,
		2 * time.Minute,
		4 * time.Minute,
		8 * time.Minute,
		16 * time.Minute,
		32 * time.Minute,
		64 * time.Minute,
		128 * time.Minute,
	}
	if attempt < 1 {
		return backoffs[0]
	}
	if attempt >= len(backoffs) {
		return backoffs[len(backoffs)-1]
	}
	return backoffs[attempt-1]
}

func (r *Runner) markBatchSuccess() {
	if r.Health != nil {
		r.Health.MarkBatchSuccess(time.Now().UTC())
	}
}

func (r *Runner) markBatchFailure() {
	if r.Health != nil {
		r.Health.MarkBatchFailure(time.Now().UTC())
	}
}

func validateHandlerRegistry(registry map[string]HandlerSpec) error {
	if len(registry) == 0 {
		return fmt.Errorf("worker handler registry is empty")
	}
	for key, spec := range registry {
		if strings.TrimSpace(key) == "" || strings.TrimSpace(spec.Kind) != key || spec.Version < 1 ||
			spec.Timeout < time.Second || spec.Timeout > 30*time.Minute || spec.MaxAttempts < 1 || spec.MaxAttempts > 25 || spec.Handle == nil {
			return fmt.Errorf("invalid worker handler registration for %q", key)
		}
	}
	return nil
}
