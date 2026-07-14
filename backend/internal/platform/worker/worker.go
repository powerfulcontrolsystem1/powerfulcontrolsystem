// Package worker runs durable background work. It intentionally owns no HTTP
// state; jobs are leased from PostgreSQL so replicas can scale independently.
package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type Handler func(context.Context, dbpkg.AsyncJob) error

type Runner struct {
	DB       *sql.DB
	WorkerID string
	Poll     time.Duration
	Batch    int
	Handlers map[string]Handler
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
	if err := dbpkg.EnsureAsyncJobsSchema(r.DB); err != nil {
		return fmt.Errorf("ensure async jobs schema: %w", err)
	}
	ticker := time.NewTicker(r.Poll)
	defer ticker.Stop()
	for {
		if err := r.runBatch(ctx); err != nil {
			log.Printf("async worker batch failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (r *Runner) runBatch(ctx context.Context) error {
	jobs, err := dbpkg.ClaimAsyncJobs(r.DB, r.WorkerID, r.Batch)
	if err != nil || len(jobs) == 0 {
		return err
	}
	for _, job := range jobs {
		handler := r.Handlers[job.Kind]
		if handler == nil {
			_ = dbpkg.RetryAsyncJob(r.DB, job, r.WorkerID, fmt.Errorf("unsupported async job kind"), time.Minute)
			continue
		}
		if err := handler(ctx, job); err != nil {
			backoff := time.Duration(job.Attempts*job.Attempts) * time.Minute
			_ = dbpkg.RetryAsyncJob(r.DB, job, r.WorkerID, err, backoff)
			continue
		}
		if err := dbpkg.CompleteAsyncJob(r.DB, job.ID, r.WorkerID); err != nil {
			return err
		}
	}
	return nil
}
