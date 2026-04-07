package metrics

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	dbpkg "github.com/you/pos-backend/db"
)

func openMetricsTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "metrics_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	if err := dbpkg.InitMetricsTable(dbConn); err != nil {
		t.Fatalf("init metrics table: %v", err)
	}
	return dbConn
}

func TestDefaultIntervalSeconds(t *testing.T) {
	t.Setenv("METRICS_INTERVAL_SECONDS", "")
	if got := DefaultIntervalSeconds(); got != 10 {
		t.Fatalf("expected default interval 10, got %d", got)
	}

	t.Setenv("METRICS_INTERVAL_SECONDS", "25")
	if got := DefaultIntervalSeconds(); got != 25 {
		t.Fatalf("expected interval 25, got %d", got)
	}

	t.Setenv("METRICS_INTERVAL_SECONDS", "-4")
	if got := DefaultIntervalSeconds(); got != 10 {
		t.Fatalf("expected fallback interval 10 for negative value, got %d", got)
	}

	t.Setenv("METRICS_INTERVAL_SECONDS", "abc")
	if got := DefaultIntervalSeconds(); got != 10 {
		t.Fatalf("expected fallback interval 10 for invalid value, got %d", got)
	}
}

func TestCollectAndStoreInsertsMetric(t *testing.T) {
	dbConn := openMetricsTestDB(t)

	collectAndStore(dbConn)

	m, err := dbpkg.GetLatestMetric(dbConn)
	if err != nil {
		t.Fatalf("get latest metric: %v", err)
	}
	if m == nil {
		t.Fatal("expected a stored metric, got nil")
	}
}

func TestStartCollectorStopsOnSignal(t *testing.T) {
	dbConn := openMetricsTestDB(t)
	stopCh := make(chan struct{})
	done := make(chan struct{})

	go func() {
		StartCollector(dbConn, 1, stopCh)
		close(done)
	}()

	close(stopCh)

	select {
	case <-done:
		// ok
	case <-time.After(3 * time.Second):
		t.Fatal("collector did not stop after stop signal")
	}
}
