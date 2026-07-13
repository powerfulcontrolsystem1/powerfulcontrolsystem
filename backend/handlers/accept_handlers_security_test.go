package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAcceptContractNeverFallsBackToPredictableSessionToken(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("accept_handlers.go"))
	if err != nil {
		t.Fatalf("read accept handler: %v", err)
	}
	source := string(raw)
	for _, forbidden := range []string{"token = data.Email"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("accept contract flow must not contain insecure session fallback: %q", forbidden)
		}
	}
	if !strings.Contains(source, "http.MaxBytesReader(w, r.Body, 64<<10)") {
		t.Fatal("accept contract flow must bound request bodies")
	}
}
