package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type ScheduleSpec struct {
	Kind        string
	Version     int
	Interval    time.Duration
	MaxAttempts int
	Priority    int
}

type Scheduler struct {
	DB    *sql.DB
	Specs []ScheduleSpec
	Now   func() time.Time
}

// EnqueueDue materializes one durable job per time bucket. The queue's unique
// idempotency hash makes this safe with multiple worker replicas.
func (s *Scheduler) EnqueueDue(ctx context.Context) error {
	if s == nil || s.DB == nil {
		return fmt.Errorf("worker scheduler database unavailable")
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	for _, spec := range s.Specs {
		if err := validateScheduleSpec(spec); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		bucket := now.UnixNano() / spec.Interval.Nanoseconds()
		payload, _ := json.Marshal(map[string]string{"scheduled_at": now.Format(time.RFC3339)})
		_, _, err := dbpkg.EnqueueAsyncJobIdempotent(s.DB, dbpkg.AsyncJob{
			Kind:           spec.Kind,
			Version:        spec.Version,
			PayloadJSON:    string(payload),
			MaxAttempts:    spec.MaxAttempts,
			Priority:       spec.Priority,
			IdempotencyKey: fmt.Sprintf("schedule:%s:%d", spec.Kind, bucket),
		})
		if err != nil {
			return fmt.Errorf("enqueue scheduled job %s: %w", spec.Kind, err)
		}
	}
	return nil
}

func validateScheduleSpec(spec ScheduleSpec) error {
	if strings.TrimSpace(spec.Kind) == "" || spec.Version < 1 || spec.Interval < time.Minute ||
		spec.Interval > 31*24*time.Hour || spec.MaxAttempts < 1 || spec.MaxAttempts > 25 ||
		spec.Priority < 0 || spec.Priority > 1000 {
		return fmt.Errorf("invalid worker schedule for %q", spec.Kind)
	}
	return nil
}
