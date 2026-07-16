// Package outbox converts committed events into durable worker jobs. It does
// not call providers directly, so a request can commit safely before any
// external side effect begins.
package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type Dispatcher struct {
	DB           *sql.DB
	DispatcherID string
	Batch        int
	Lease        time.Duration
	AllowedKinds map[string]struct{}
}

// Dispatch claims committed events, creates one idempotent job for each one,
// then marks the event published. Unknown topics stay visible and eventually
// enter the outbox dead-letter state instead of being dropped silently.
func (d *Dispatcher) Dispatch(ctx context.Context) error {
	if d == nil || d.DB == nil || strings.TrimSpace(d.DispatcherID) == "" {
		return fmt.Errorf("outbox dispatcher is not configured")
	}
	if d.Batch < 1 || d.Batch > 100 {
		d.Batch = 20
	}
	if d.Lease < 30*time.Second || d.Lease > 30*time.Minute {
		d.Lease = 5 * time.Minute
	}
	if _, err := dbpkg.RecoverExpiredOutboxEvents(d.DB); err != nil {
		return err
	}
	events, err := dbpkg.ClaimOutboxEventsWithLease(d.DB, d.DispatcherID, d.Batch, d.Lease)
	if err != nil || len(events) == 0 {
		return err
	}
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		if _, allowed := d.AllowedKinds[event.Topic]; !allowed {
			if retryErr := dbpkg.RetryOutboxEvent(d.DB, event, d.DispatcherID, fmt.Errorf("outbox topic has no enabled worker handler"), time.Minute); retryErr != nil {
				return retryErr
			}
			continue
		}
		_, _, enqueueErr := dbpkg.EnqueueAsyncJobIdempotent(d.DB, dbpkg.AsyncJob{
			EmpresaID:      event.EmpresaID,
			Kind:           event.Topic,
			Version:        event.Version,
			PayloadJSON:    event.PayloadJSON,
			MaxAttempts:    event.MaxAttempts,
			IdempotencyKey: "outbox-event:" + strconv.FormatInt(event.ID, 10),
		})
		if enqueueErr != nil {
			if retryErr := dbpkg.RetryOutboxEvent(d.DB, event, d.DispatcherID, enqueueErr, time.Minute); retryErr != nil {
				return retryErr
			}
			continue
		}
		if err := dbpkg.CompleteOutboxEvent(d.DB, event.ID, d.DispatcherID); err != nil {
			return err
		}
	}
	return nil
}
