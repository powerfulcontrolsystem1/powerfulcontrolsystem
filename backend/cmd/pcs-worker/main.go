// pcs-worker performs only durable background work. Schema ownership belongs
// to pcs-migrate; HTTP ownership belongs to the API binary.
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/internal/platform/outbox"
	"github.com/you/pos-backend/internal/platform/worker"
)

func main() {
	dsn := strings.TrimSpace(os.Getenv("DB_SUPERADMIN_DSN"))
	if dsn == "" {
		dsn = strings.TrimSpace(os.Getenv("DATABASE_SUPERADMIN_URL"))
	}
	if dsn == "" {
		log.Fatal("DB_SUPERADMIN_DSN is required for pcs-worker")
	}
	if err := os.Setenv("DB_DIALECT", "postgres"); err != nil {
		log.Fatal(err)
	}
	dbConn, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()
	pool, err := dbpkg.LoadPostgresPoolConfig(os.Getenv, "worker")
	if err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.ConfigurePostgresPool(dbConn, pool); err != nil {
		log.Fatal(err)
	}
	if err := dbConn.Ping(); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.VerifyAsyncJobsSchema(dbConn); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.VerifyOutboxSchema(dbConn); err != nil {
		log.Fatal(err)
	}
	workerID := strings.TrimSpace(os.Getenv("PCS_WORKER_ID"))
	if workerID == "" {
		host, hostErr := os.Hostname()
		if hostErr != nil || host == "" {
			host = "unknown"
		}
		workerID = "pcs-worker-" + host
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	registry := worker.DefaultRegistry()
	health := &worker.HealthState{}
	healthAddr := strings.TrimSpace(os.Getenv("PCS_WORKER_HEALTH_ADDR"))
	if healthAddr == "" {
		healthAddr = "127.0.0.1:8082"
	}
	healthErrors, err := worker.StartHealthServer(ctx, healthAddr, dbConn, health)
	if err != nil {
		log.Fatal(err)
	}
	dispatcher := &outbox.Dispatcher{DB: dbConn, DispatcherID: workerID + "-outbox", Batch: 20, Lease: 5 * time.Minute, AllowedKinds: worker.Kinds(registry)}
	runner := &worker.Runner{
		DB:       dbConn,
		WorkerID: workerID,
		Poll:     2 * time.Second,
		Batch:    20,
		Lease:    5 * time.Minute,
		Handlers: registry,
		Health:   health,
		BeforeBatch: func(ctx context.Context) error {
			return dispatcher.Dispatch(ctx)
		},
	}
	runnerErrors := make(chan error, 1)
	go func() { runnerErrors <- runner.Run(ctx) }()
	select {
	case err := <-healthErrors:
		if ctx.Err() != nil {
			if runErr := <-runnerErrors; runErr != nil {
				log.Fatal(fmt.Errorf("worker stopped: %w", runErr))
			}
			break
		}
		if err != nil {
			log.Fatal(fmt.Errorf("worker health server stopped: %w", err))
		}
		log.Fatal("worker health server stopped unexpectedly")
	case err := <-runnerErrors:
		if err != nil {
			log.Fatal(fmt.Errorf("worker stopped: %w", err))
		}
	}
	log.Print("worker stopped gracefully")
}
