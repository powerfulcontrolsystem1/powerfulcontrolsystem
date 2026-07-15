// Package worker runs durable background work. It intentionally owns no HTTP
// state; jobs are leased from PostgreSQL so replicas can scale independently.
package worker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type Handler func(context.Context, dbpkg.AsyncJob) error

// HandlerSpec is the central registry contract for durable jobs. New work must
// be declared here instead of being launched from an HTTP request goroutine.
type HandlerSpec struct {
	Kind        string
	Version     int
	Enabled     bool
	Timeout     time.Duration
	MaxAttempts int
	Backoff     func(attempt int) time.Duration
	Handler     Handler
}

type Runner struct {
	DB       *sql.DB
	WorkerID string
	Poll     time.Duration
	Batch    int
	Lease    time.Duration
	Handlers map[string]Handler
	Registry []HandlerSpec
}

func (r *Runner) Run(ctx context.Context) error {
	if r.DB == nil {
		return fmt.Errorf("worker database unavailable")
	}
	if strings.TrimSpace(r.WorkerID) == "" {
		return fmt.Errorf("worker id required")
	}
	r.normalize()
	if err := r.validateRegistry(); err != nil {
		return err
	}
	defer func() {
		if _, err := dbpkg.ReleaseAsyncJobsForWorker(r.DB, r.WorkerID); err != nil {
			log.Printf("async worker release failed: %v", err)
		}
	}()
	ticker := time.NewTicker(r.Poll)
	defer ticker.Stop()
	for {
		if err := r.runBatch(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Printf("async worker batch failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (r *Runner) normalize() {
	if r.Poll < time.Second {
		r.Poll = 2 * time.Second
	}
	if r.Batch < 1 || r.Batch > 100 {
		r.Batch = 20
	}
	if r.Lease < time.Minute {
		r.Lease = 5 * time.Minute
	}
}

func (r *Runner) validateRegistry() error {
	known := make(map[string]struct{}, len(r.Registry))
	for _, spec := range r.Registry {
		kind := strings.TrimSpace(spec.Kind)
		if kind == "" || spec.Version < 1 || spec.Handler == nil {
			return fmt.Errorf("invalid async handler registration")
		}
		if _, exists := known[kind]; exists {
			return fmt.Errorf("duplicate async handler %s", kind)
		}
		known[kind] = struct{}{}
		if spec.Enabled {
			if r.Handlers == nil {
				r.Handlers = make(map[string]Handler)
			}
			r.Handlers[kind] = spec.Handler
		}
	}
	return nil
}

func (r *Runner) runBatch(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := dbpkg.ExpireAsyncJobs(r.DB); err != nil {
		return err
	}
	if _, err := dbpkg.RecoverExpiredAsyncJobs(r.DB, r.Lease); err != nil {
		return err
	}
	jobs, err := dbpkg.ClaimAsyncJobs(r.DB, r.WorkerID, r.Batch)
	if err != nil || len(jobs) == 0 {
		return err
	}
	for _, job := range jobs {
		if err := ctx.Err(); err != nil {
			return err
		}
		spec, handler := r.handlerFor(job)
		if handler == nil {
			_ = dbpkg.RetryAsyncJob(r.DB, job, r.WorkerID, fmt.Errorf("unsupported async job kind"), time.Minute)
			continue
		}
		job = jobWithRegisteredMaxAttempts(job, spec)
		if err := r.runWithHeartbeat(ctx, job, spec, handler); err != nil {
			backoff := retryBackoff(spec, job.Attempts)
			_ = dbpkg.RetryAsyncJob(r.DB, job, r.WorkerID, err, backoff)
			continue
		}
		if err := dbpkg.CompleteAsyncJob(r.DB, job.ID, r.WorkerID); err != nil {
			return err
		}
	}
	return nil
}

// jobWithRegisteredMaxAttempts keeps the durable queue policy aligned with the
// active handler registry. The database value remains the source of history,
// while an active handler can only make its own retry budget stricter.
func jobWithRegisteredMaxAttempts(job dbpkg.AsyncJob, spec HandlerSpec) dbpkg.AsyncJob {
	if spec.MaxAttempts >= 1 && spec.MaxAttempts <= 25 {
		job.MaxAttempts = spec.MaxAttempts
	}
	return job
}

func (r *Runner) handlerFor(job dbpkg.AsyncJob) (HandlerSpec, Handler) {
	for _, spec := range r.Registry {
		if spec.Enabled && spec.Kind == job.Kind && spec.Version == job.PayloadVersion {
			return spec, spec.Handler
		}
	}
	if handler := r.Handlers[job.Kind]; handler != nil {
		return HandlerSpec{Kind: job.Kind, Version: job.PayloadVersion, Timeout: r.Lease}, handler
	}
	return HandlerSpec{}, nil
}

func (r *Runner) runWithHeartbeat(parent context.Context, job dbpkg.AsyncJob, spec HandlerSpec, handler Handler) error {
	timeout := spec.Timeout
	if timeout <= 0 || timeout > r.Lease {
		timeout = r.Lease
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	done := make(chan struct{})
	heartbeatStopped := make(chan struct{})
	interval := r.Lease / 3
	if interval < time.Second {
		interval = time.Second
	}
	go func() {
		defer close(heartbeatStopped)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := dbpkg.HeartbeatAsyncJob(r.DB, job.ID, r.WorkerID); err != nil {
					log.Printf("async worker heartbeat failed for job %d: %v", job.ID, err)
				}
			}
		}
	}()
	err := handler(ctx, job)
	close(done)
	<-heartbeatStopped
	return err
}

func retryBackoff(spec HandlerSpec, attempt int) time.Duration {
	if spec.Backoff != nil {
		if duration := spec.Backoff(attempt); duration >= time.Second {
			return duration
		}
	}
	if attempt < 1 {
		attempt = 1
	}
	return time.Duration(attempt*attempt) * time.Minute
}
