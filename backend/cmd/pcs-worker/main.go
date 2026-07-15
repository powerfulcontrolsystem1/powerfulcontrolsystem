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
	if err := dbConn.Ping(); err != nil {
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
	runner := &worker.Runner{DB: dbConn, WorkerID: workerID, Poll: 2 * time.Second, Batch: 20, Handlers: map[string]worker.Handler{}}
	if err := runner.Run(ctx); err != nil {
		log.Fatal(fmt.Errorf("worker stopped: %w", err))
	}
}
