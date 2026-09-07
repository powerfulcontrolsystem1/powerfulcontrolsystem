package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompareRejectsOnlyRegressions(t *testing.T) {
	limits := metrics{
		FunctionsOver100Lines:    10,
		FunctionsOver200Lines:    4,
		LargestFunctionLines:     500,
		ExactDuplicateBodyGroups: 3,
		DBCallsWithoutContext:    20,
		ExplicitIgnoredResults:   8,
	}
	if regressions := compare(limits, limits); len(regressions) != 0 {
		t.Fatalf("equal baseline must pass: %v", regressions)
	}
	current := limits
	current.DBCallsWithoutContext++
	regressions := compare(current, limits)
	if len(regressions) != 1 || regressions[0] != "db_calls_without_context: 21 > 20" {
		t.Fatalf("unexpected regressions: %v", regressions)
	}
}

func TestDBCallClassification(t *testing.T) {
	for _, name := range []string{"db.Query", "tx.Exec", "conn.QueryRow", "db.Prepare", "db.Begin"} {
		if !isDBCallWithoutContext(name) {
			t.Fatalf("expected %s to require context", name)
		}
	}
	for _, name := range []string{"db.QueryContext", "tx.ExecContext", "db.BeginTx", "rows.Scan", "r.URL.Query", "request.URL.Query"} {
		if isDBCallWithoutContext(name) {
			t.Fatalf("did not expect %s to be classified", name)
		}
	}
}

func TestMeasureReportsDebtByFile(t *testing.T) {
	root := t.TempDir()
	source := []byte(`package sample
func run(db interface{ Query(string, ...any) (any, error) }) {
	_, _ = db.Query("SELECT 1")
	_ = recover()
}
`)
	if err := os.WriteFile(filepath.Join(root, "sample.go"), source, 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := measure(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.TopDebtFiles) != 1 {
		t.Fatalf("expected one debt file, got %+v", got.TopDebtFiles)
	}
	item := got.TopDebtFiles[0]
	if item.File != "sample.go" || item.DBCallsWithoutContext != 1 || item.ExplicitIgnoredResult != 1 {
		t.Fatalf("unexpected debt summary: %+v", item)
	}
}
