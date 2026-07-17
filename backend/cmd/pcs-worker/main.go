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

func openWorkerDB(name string, dsn string) (*sql.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("%s is required for pcs-worker", name)
	}
	dbConn, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		return nil, err
	}
	pool, err := dbpkg.LoadPostgresPoolConfig(os.Getenv, "worker")
	if err != nil {
		_ = dbConn.Close()
		return nil, err
	}
	if err := dbpkg.ConfigurePostgresPool(dbConn, pool); err != nil {
		_ = dbConn.Close()
		return nil, err
	}
	if err := dbConn.Ping(); err != nil {
		_ = dbConn.Close()
		return nil, err
	}
	return dbConn, nil
}

func main() {
	if err := os.Setenv("DB_DIALECT", "postgres"); err != nil {
		log.Fatal(err)
	}
	dbSuper, err := openWorkerDB("DB_SUPERADMIN_DSN", firstNonEmptyEnv("DB_SUPERADMIN_DSN", "DATABASE_SUPERADMIN_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer dbSuper.Close()
	dbEmp, err := openWorkerDB("DB_EMPRESAS_DSN", firstNonEmptyEnv("DB_EMPRESAS_DSN", "DATABASE_EMPRESAS_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer dbEmp.Close()
	if err := dbpkg.VerifyAsyncJobsSchema(dbSuper); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.VerifyOutboxSchema(dbSuper); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.VerifyMetricsSchema(dbSuper); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.VerifyOutboxSchema(dbEmp); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.VerifyPlatformMigrations(context.Background(), dbEmp, dbpkg.MigrationTargetEmpresas); err != nil {
		log.Fatal(err)
	}
	if err := dbpkg.VerifyPlatformMigrations(context.Background(), dbSuper, dbpkg.MigrationTargetSuper); err != nil {
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
	registry := businessRegistry(dbEmp, dbSuper)
	health := &worker.HealthState{}
	healthAddr := strings.TrimSpace(os.Getenv("PCS_WORKER_HEALTH_ADDR"))
	if healthAddr == "" {
		healthAddr = "127.0.0.1:8082"
	}
	healthErrors, err := worker.StartHealthServer(ctx, healthAddr, dbSuper, health)
	if err != nil {
		log.Fatal(err)
	}
	dispatcherSuper := &outbox.Dispatcher{SourceDB: dbSuper, QueueDB: dbSuper, DispatcherID: workerID + "-outbox-super", Batch: 20, Lease: 5 * time.Minute, AllowedKinds: worker.Kinds(registry)}
	dispatcherEmp := &outbox.Dispatcher{SourceDB: dbEmp, QueueDB: dbSuper, DispatcherID: workerID + "-outbox-empresas", Batch: 20, Lease: 5 * time.Minute, AllowedKinds: worker.Kinds(registry)}
	scheduler := &worker.Scheduler{DB: dbSuper, Specs: businessSchedules()}
	runner := &worker.Runner{
		DB:       dbSuper,
		WorkerID: workerID,
		Poll:     2 * time.Second,
		Batch:    20,
		Lease:    5 * time.Minute,
		Handlers: registry,
		Health:   health,
		BeforeBatch: func(ctx context.Context) error {
			if err := scheduler.EnqueueDue(ctx); err != nil {
				return err
			}
			if err := dispatcherSuper.Dispatch(ctx); err != nil {
				return err
			}
			return dispatcherEmp.Dispatch(ctx)
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

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}
