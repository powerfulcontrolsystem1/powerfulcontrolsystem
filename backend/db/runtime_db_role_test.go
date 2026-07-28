package db

import "testing"

func TestRuntimeDatabaseRoleCredentialsAreURLSafe(t *testing.T) {
	for _, valid := range []string{"pcs_runtime", "runtime_api_2026"} {
		if !runtimeDatabaseRoleNamePattern.MatchString(valid) {
			t.Fatalf("expected valid runtime role %q", valid)
		}
	}
	for _, invalid := range []string{"PC", "PCS_RUNTIME", "pcs-runtime", `pcs";DROP ROLE pcs;--`} {
		if runtimeDatabaseRoleNamePattern.MatchString(invalid) {
			t.Fatalf("expected invalid runtime role %q", invalid)
		}
	}
	if !runtimeDatabasePasswordPattern.MatchString("0123456789abcdef0123456789abcdef") {
		t.Fatal("expected 32-character URL-safe password")
	}
	for _, invalid := range []string{
		"short",
		"0123456789abcdef0123456789abcde#",
		"0123456789abcdef 123456789abcdef",
	} {
		if runtimeDatabasePasswordPattern.MatchString(invalid) {
			t.Fatalf("expected invalid runtime database password %q", invalid)
		}
	}
}
