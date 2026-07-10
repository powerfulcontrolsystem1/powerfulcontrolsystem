package handlers

import "testing"

func TestDockerInspectStatus(t *testing.T) {
	tests := map[string]string{
		"running|healthy":   "active",
		"running|":          "active",
		"running|starting":  "degraded",
		"restarting|":       "degraded",
		"exited|":           "inactive",
		"dead|":             "error",
		"running|unhealthy": "error",
		"unknown|":          "unavailable",
	}
	for input, want := range tests {
		if got := dockerInspectStatus(input); got != want {
			t.Errorf("dockerInspectStatus(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPreferServiceStatus(t *testing.T) {
	if got := preferServiceStatus("inactive", "active"); got != "active" {
		t.Fatalf("preferServiceStatus() = %q, want active", got)
	}
	if got := preferServiceStatus("error", "unavailable"); got != "unavailable" {
		t.Fatalf("preferServiceStatus() = %q, want unavailable", got)
	}
}
