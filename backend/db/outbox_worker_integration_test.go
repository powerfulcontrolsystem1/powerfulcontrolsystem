package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/internal/platform/outbox"
)

// TestP108OutboxWorkerDurabilityIntegration covers the durable boundary used
// by the worker. It only runs against an explicitly marked disposable
// PostgreSQL database, never a configured application database.
func TestP108OutboxWorkerDurabilityIntegration(t *testing.T) {
	dbConn := openP108DisposablePostgres(t)
	if err := dbpkg.EnsureOutboxSchema(dbConn); err != nil {
		t.Fatalf("prepare outbox schema: %v", err)
	}
	if err := dbpkg.EnsureAsyncJobsSchema(dbConn); err != nil {
		t.Fatalf("prepare async jobs schema: %v", err)
	}

	tenantID := int64(9_000_000_000 + time.Now().UnixNano()%1_000_000)
	t.Cleanup(func() {
		_, _ = dbConn.Exec(`DELETE FROM pcs_async_jobs WHERE empresa_id = $1`, tenantID)
		_, _ = dbConn.Exec(`DELETE FROM pcs_outbox_events WHERE empresa_id = $1`, tenantID)
	})

	topic := "p108.outbox.integration"
	key := fmt.Sprintf("p108-outbox-%d", tenantID)
	_, created := insertP108OutboxEvent(t, dbConn, dbpkg.OutboxEvent{
		EmpresaID: tenantID, Topic: topic, PayloadJSON: `{"reference":"isolated"}`,
		IdempotencyKey: key,
	})
	if !created {
		t.Fatal("first logical event must be created")
	}
	second, created := insertP108OutboxEvent(t, dbConn, dbpkg.OutboxEvent{
		EmpresaID: tenantID, Topic: topic, PayloadJSON: `{"reference":"isolated"}`,
		IdempotencyKey: key,
	})
	if created || second.ID <= 0 {
		t.Fatalf("duplicate event must resolve the persisted row: created=%v replay_id=%d", created, second.ID)
	}
	if got := countP108Rows(t, dbConn, `SELECT COUNT(*) FROM pcs_outbox_events WHERE empresa_id = $1 AND topic = $2`, tenantID, topic); got != 1 {
		t.Fatalf("expected one outbox event after duplicate insertion, got %d", got)
	}

	dispatcher := &outbox.Dispatcher{
		SourceDB:     dbConn,
		QueueDB:      dbConn,
		DispatcherID: "p108-dispatcher",
		Batch:        10,
		Lease:        time.Minute,
		AllowedKinds: map[string]struct{}{topic: {}},
	}
	if err := dispatcher.Dispatch(context.Background()); err != nil {
		t.Fatalf("first dispatch: %v", err)
	}
	if err := dispatcher.Dispatch(context.Background()); err != nil {
		t.Fatalf("replayed dispatch: %v", err)
	}
	if got := countP108Rows(t, dbConn, `SELECT COUNT(*) FROM pcs_async_jobs WHERE empresa_id = $1 AND kind = $2`, tenantID, topic); got != 1 {
		t.Fatalf("expected one job after replayed dispatch, got %d", got)
	}

	// Two workers race to claim one job. SKIP LOCKED must give ownership to one.
	claims := make(chan []dbpkg.AsyncJob, 2)
	errs := make(chan error, 2)
	var start sync.WaitGroup
	start.Add(1)
	for _, workerID := range []string{"p108-worker-a", "p108-worker-b"} {
		go func(id string) {
			start.Wait()
			jobs, err := dbpkg.ClaimAsyncJobsWithLease(dbConn, id, 1, time.Minute)
			errs <- err
			claims <- jobs
		}(workerID)
	}
	start.Done()
	var claimed dbpkg.AsyncJob
	claimCount := 0
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatalf("concurrent claim: %v", err)
		}
		jobs := <-claims
		if len(jobs) > 1 {
			t.Fatalf("worker claimed unexpected batch: %d", len(jobs))
		}
		if len(jobs) == 1 {
			claimCount++
			claimed = jobs[0]
		}
	}
	if claimCount != 1 {
		t.Fatalf("one durable job must have one owner, claims=%d", claimCount)
	}
	if err := dbpkg.CompleteAsyncJob(dbConn, claimed.ID, claimedStatusOwner(t, dbConn, claimed.ID)); err != nil {
		t.Fatalf("complete claimed job: %v", err)
	}

	// A job abandoned by a crashed worker becomes claimable again after its lease.
	abandoned, _, err := dbpkg.EnqueueAsyncJobIdempotent(dbConn, dbpkg.AsyncJob{
		EmpresaID: tenantID, Kind: topic, PayloadJSON: `{"reference":"lease"}`,
		MaxAttempts: 2, IdempotencyKey: key + "-lease",
	})
	if err != nil {
		t.Fatalf("enqueue abandoned job: %v", err)
	}
	firstLease, err := dbpkg.ClaimAsyncJobsWithLease(dbConn, "p108-crashed-worker", 1, time.Second)
	if err != nil || len(firstLease) != 1 || firstLease[0].ID != abandoned.ID {
		t.Fatalf("claim abandoned job: jobs=%v err=%v", len(firstLease), err)
	}
	time.Sleep(1200 * time.Millisecond)
	if recovered, err := dbpkg.RecoverExpiredAsyncJobs(dbConn); err != nil || recovered != 1 {
		t.Fatalf("recover expired job: recovered=%d err=%v", recovered, err)
	}
	reclaimed, err := dbpkg.ClaimAsyncJobsWithLease(dbConn, "p108-recovery-worker", 1, time.Minute)
	if err != nil || len(reclaimed) != 1 || reclaimed[0].ID != abandoned.ID {
		t.Fatalf("reclaim expired job: jobs=%v err=%v", len(reclaimed), err)
	}
	if err := dbpkg.CompleteAsyncJob(dbConn, reclaimed[0].ID, "p108-recovery-worker"); err != nil {
		t.Fatalf("complete reclaimed job: %v", err)
	}

	dead, _, err := dbpkg.EnqueueAsyncJobIdempotent(dbConn, dbpkg.AsyncJob{
		EmpresaID: tenantID, Kind: topic, PayloadJSON: `{"reference":"dead-letter"}`,
		MaxAttempts: 1, IdempotencyKey: key + "-dead",
	})
	if err != nil {
		t.Fatalf("enqueue dead-letter job: %v", err)
	}
	deadClaim, err := dbpkg.ClaimAsyncJobsWithLease(dbConn, "p108-dead-worker", 1, time.Minute)
	if err != nil || len(deadClaim) != 1 || deadClaim[0].ID != dead.ID {
		t.Fatalf("claim dead-letter job: jobs=%v err=%v", len(deadClaim), err)
	}
	if err := dbpkg.FailAsyncJob(dbConn, deadClaim[0], "p108-dead-worker", fmt.Errorf("intentional isolated failure")); err != nil {
		t.Fatalf("dead-letter job: %v", err)
	}
	if got := countP108Rows(t, dbConn, `SELECT COUNT(*) FROM pcs_async_jobs WHERE empresa_id = $1 AND status = 'dead'`, tenantID); got != 1 {
		t.Fatalf("expected one durable dead-letter job, got %d", got)
	}
}

func openP108DisposablePostgres(t *testing.T) *sql.DB {
	t.Helper()
	if os.Getenv("P108_OUTBOX_TEST_DATABASE") != "isolated" {
		t.Skip("requires P108_OUTBOX_TEST_DATABASE=isolated and a disposable PostgreSQL DSN")
	}
	dsn := strings.TrimSpace(os.Getenv("P108_OUTBOX_TEST_DSN"))
	if dsn == "" || !strings.Contains(strings.ToLower(dsn), "p108_") {
		t.Fatal("P108_OUTBOX_TEST_DSN must name an isolated p108_ PostgreSQL database")
	}
	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open isolated PostgreSQL: %v", err)
	}
	t.Cleanup(func() { _ = dbConn.Close() })
	if err := dbConn.Ping(); err != nil {
		t.Fatalf("ping isolated PostgreSQL: %v", err)
	}
	return dbConn
}

func insertP108OutboxEvent(t *testing.T, dbConn *sql.DB, event dbpkg.OutboxEvent) (*dbpkg.OutboxEvent, bool) {
	t.Helper()
	tx, err := dbConn.Begin()
	if err != nil {
		t.Fatalf("begin outbox transaction: %v", err)
	}
	defer func() { _ = tx.Rollback() }()
	stored, created, err := dbpkg.InsertOutboxEventIdempotent(tx, event)
	if err != nil {
		t.Fatalf("insert outbox event: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit outbox event: %v", err)
	}
	return stored, created
}

func countP108Rows(t *testing.T, dbConn *sql.DB, query string, args ...any) int64 {
	t.Helper()
	var value int64
	if err := dbConn.QueryRow(query, args...).Scan(&value); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	return value
}

func claimedStatusOwner(t *testing.T, dbConn *sql.DB, id int64) string {
	t.Helper()
	var owner string
	if err := dbConn.QueryRow(`SELECT locked_by FROM pcs_async_jobs WHERE id = $1`, id).Scan(&owner); err != nil {
		t.Fatalf("read claimed owner: %v", err)
	}
	return owner
}
