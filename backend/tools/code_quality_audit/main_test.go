package main

import "testing"

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
	for _, name := range []string{"db.QueryContext", "tx.ExecContext", "db.BeginTx", "rows.Scan"} {
		if isDBCallWithoutContext(name) {
			t.Fatalf("did not expect %s to be classified", name)
		}
	}
}
