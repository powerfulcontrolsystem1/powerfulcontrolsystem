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

// OutboxHandler converts a committed domain event into durable external work.
// It receives the event after claim; effects must be idempotent by event ID.
type OutboxHandler func(context.Context, dbpkg.OutboxEvent) error

type OutboxDispatcher struct {
	DB       *sql.DB
	WorkerID string
	Poll     time.Duration
	Batch    int
	Lease    time.Duration
	Handlers map[string]OutboxHandler
}

func (d *OutboxDispatcher) Run(ctx context.Context) error {
	if d.DB == nil || strings.TrimSpace(d.WorkerID) == "" {
		return fmt.Errorf("outbox dispatcher requires database and worker id")
	}
	d.normalize()
	ticker := time.NewTicker(d.Poll)
	defer ticker.Stop()
	for {
		if err := d.runBatch(ctx); err != nil && ctx.Err() == nil {
			log.Printf("outbox dispatcher batch failed: %v", err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (d *OutboxDispatcher) normalize() {
	if d.Poll < time.Second {
		d.Poll = 2 * time.Second
	}
	if d.Batch < 1 || d.Batch > 100 {
		d.Batch = 20
	}
	if d.Lease < time.Minute {
		d.Lease = 5 * time.Minute
	}
}

func (d *OutboxDispatcher) runBatch(ctx context.Context) error {
	if _, err := dbpkg.RecoverExpiredOutboxEvents(d.DB, d.Lease); err != nil {
		return err
	}
	events, err := dbpkg.ClaimOutboxEvents(d.DB, d.WorkerID, d.Batch)
	if err != nil || len(events) == 0 {
		return err
	}
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		handler := d.Handlers[event.Topic]
		if handler == nil {
			if err := dbpkg.RetryOutboxEvent(d.DB, event, d.WorkerID, time.Minute); err != nil {
				return err
			}
			continue
		}
		err := d.runWithHeartbeat(ctx, event, handler)
		if err != nil {
			if retryErr := dbpkg.RetryOutboxEvent(d.DB, event, d.WorkerID, time.Duration(event.Attempts*event.Attempts)*time.Minute); retryErr != nil {
				return retryErr
			}
			continue
		}
		if err := dbpkg.PublishOutboxEvent(d.DB, event.ID, d.WorkerID); err != nil {
			return err
		}
	}
	return nil
}

func (d *OutboxDispatcher) runWithHeartbeat(parent context.Context, event dbpkg.OutboxEvent, handler OutboxHandler) error {
	ctx, cancel := context.WithTimeout(parent, d.Lease)
	defer cancel()
	done := make(chan struct{})
	stopped := make(chan struct{})
	interval := d.Lease / 3
	if interval < time.Second {
		interval = time.Second
	}
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := dbpkg.HeartbeatOutboxEvent(d.DB, event.ID, d.WorkerID); err != nil {
					log.Printf("outbox heartbeat failed for event %d: %v", event.ID, err)
				}
			}
		}
	}()
	err := handler(ctx, event)
	close(done)
	<-stopped
	return err
}
