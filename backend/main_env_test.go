package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSelectedEnvDefaultsFromFileOnlyAllowedKeys(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("CONFIG_ENC_KEY", "")

	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env.platform")
	content := "OPENAI_API_KEY='sk-test-value'\nCONFIG_ENC_KEY=should-not-load\nPOSTGRES_PASSWORD=also-ignored\n"
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	added, err := loadSelectedEnvDefaultsFromFile(envPath, []string{"OPENAI_API_KEY"})
	if err != nil {
		t.Fatal(err)
	}
	if added != 1 {
		t.Fatalf("expected 1 loaded key, got %d", added)
	}
	if got := os.Getenv("OPENAI_API_KEY"); got != "sk-test-value" {
		t.Fatalf("unexpected OPENAI_API_KEY value %q", got)
	}
	if got := os.Getenv("CONFIG_ENC_KEY"); got != "" {
		t.Fatalf("CONFIG_ENC_KEY must not be loaded by AI fallback, got %q", got)
	}
}
