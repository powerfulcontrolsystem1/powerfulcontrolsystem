package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/internal/platform/runtimeconfig"
	"github.com/you/pos-backend/internal/platform/worker"
)

func main() {
	if err := os.Setenv("DB_DIALECT", "postgres"); err != nil {
		log.Fatal(err)
	}
	config, err := runtimeconfig.Load(os.Getenv)
	if err != nil {
		log.Fatal(fmt.Errorf("worker runtime configuration: %w", err))
	}
	superDSN := firstNonEmptyEnv("DB_SUPERADMIN_DSN", "DATABASE_SUPERADMIN_URL")
	empresasDSN := firstNonEmptyEnv("DB_EMPRESAS_DSN", "DATABASE_EMPRESAS_URL")
	superDB, err := config.Database.OpenAndPing(context.Background(), dbpkg.PostgresCompatDriverName(), superDSN, "superadministrador")
	if err != nil {
		log.Fatal(err)
	}
	defer superDB.Close()
	empresasDB, err := config.Database.OpenAndPing(context.Background(), dbpkg.PostgresCompatDriverName(), empresasDSN, "empresas")
	if err != nil {
		log.Fatal(err)
	}
	defer empresasDB.Close()

	workerID := resolveWorkerID()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := startWorkerHeartbeat(ctx); err != nil {
		log.Fatal(err)
	}
	startPeriodicWorkers(ctx, empresasDB, superDB)

	asyncRunner := &worker.Runner{
		DB:       superDB,
		WorkerID: workerID,
		Poll:     workerPollInterval(),
		Batch:    workerBatchSize(),
		Lease:    workerLease(),
		Registry: productionAsyncJobRegistry(empresasDB, superDB),
	}
	outbox := &worker.OutboxDispatcher{
		DB:       superDB,
		WorkerID: workerID,
		Poll:     workerPollInterval(),
		Batch:    workerBatchSize(),
		Lease:    workerLease(),
		Handlers: productionOutboxHandlers(superDB),
	}
	errs := make(chan error, 2)
	go func() { errs <- asyncRunner.Run(ctx) }()
	go func() { errs <- outbox.Run(ctx) }()
	select {
	case <-ctx.Done():
		return
	case err := <-errs:
		if err != nil {
			log.Fatal(fmt.Errorf("worker stopped: %w", err))
		}
	}
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func resolveWorkerID() string {
	if workerID := strings.TrimSpace(os.Getenv("PCS_WORKER_ID")); workerID != "" {
		return workerID
	}
	host, err := os.Hostname()
	if err != nil || strings.TrimSpace(host) == "" {
		host = "unknown"
	}
	return "pcs-worker-" + host
}

func workerPollInterval() time.Duration {
	if value, err := time.ParseDuration(strings.TrimSpace(os.Getenv("PCS_WORKER_POLL_INTERVAL"))); err == nil && value >= time.Second && value <= time.Minute {
		return value
	}
	return 2 * time.Second
}

func workerBatchSize() int {
	value := strings.TrimSpace(os.Getenv("PCS_WORKER_BATCH_SIZE"))
	if value == "" {
		return 20
	}
	var parsed int
	if _, err := fmt.Sscan(value, &parsed); err == nil && parsed >= 1 && parsed <= 100 {
		return parsed
	}
	return 20
}

func workerLease() time.Duration {
	if value, err := time.ParseDuration(strings.TrimSpace(os.Getenv("PCS_WORKER_JOB_LEASE"))); err == nil && value >= time.Minute && value <= 15*time.Minute {
		return value
	}
	return 5 * time.Minute
}

// startWorkerHeartbeat creates a local liveness signal for the container
// healthcheck. It only permits a tmpfs path so runtime configuration cannot
// turn this helper into an arbitrary host-mounted file writer.
func startWorkerHeartbeat(ctx context.Context) error {
	path := strings.TrimSpace(os.Getenv("PCS_WORKER_HEARTBEAT_FILE"))
	if path == "" {
		return nil
	}
	path = filepath.Clean(path)
	if path != "/tmp/pcs-worker.heartbeat" {
		return fmt.Errorf("worker heartbeat path is not allowed")
	}
	write := func() error {
		return os.WriteFile(path, []byte(time.Now().UTC().Format(time.RFC3339Nano)+"\n"), 0o600)
	}
	if err := write(); err != nil {
		return fmt.Errorf("worker heartbeat: %w", err)
	}
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := write(); err != nil {
					log.Printf("worker heartbeat write failed: %v", err)
				}
			}
		}
	}()
	return nil
}
