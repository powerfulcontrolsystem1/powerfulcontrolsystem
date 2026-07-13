package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultManagerStoresConfigInWritableRuntimeData(t *testing.T) {
	backendDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(backendDir, "handlers"), 0o755); err != nil {
		t.Fatalf("create handlers directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backendDir, "go.mod"), []byte("module test\n"), 0o600); err != nil {
		t.Fatalf("create go.mod marker: %v", err)
	}
	t.Setenv("PCS_BACKEND_DIR", backendDir)

	manager := NewManager("")
	settings, err := manager.Load()
	if err != nil {
		t.Fatalf("load default settings: %v", err)
	}
	if got, want := settings.ConfigPath, filepath.Join(backendDir, "logs", "vps_security", "config.json"); got != want {
		t.Fatalf("config path = %q, want %q", got, want)
	}
	if _, err := os.Stat(settings.ConfigPath); err != nil {
		t.Fatalf("runtime config was not created: %v", err)
	}
}

func TestDefaultManagerMigratesLegacyConfig(t *testing.T) {
	backendDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(backendDir, "handlers"), 0o755); err != nil {
		t.Fatalf("create handlers directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backendDir, "go.mod"), []byte("module test\n"), 0o600); err != nil {
		t.Fatalf("create go.mod marker: %v", err)
	}
	legacyPath := filepath.Join(backendDir, "secure", "vps_security_config.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o700); err != nil {
		t.Fatalf("create legacy directory: %v", err)
	}
	legacy := []byte(`{"target_host":"scanner.internal","profile":"quick","data_dir":""}`)
	if err := os.WriteFile(legacyPath, legacy, 0o600); err != nil {
		t.Fatalf("write legacy config: %v", err)
	}
	t.Setenv("PCS_BACKEND_DIR", backendDir)

	settings, err := NewManager("").Load()
	if err != nil {
		t.Fatalf("migrate legacy settings: %v", err)
	}
	if settings.TargetHost != "scanner.internal" || settings.Profile != "quick" {
		t.Fatalf("legacy settings were not preserved: %#v", settings)
	}
	if _, err := os.Stat(filepath.Join(backendDir, "logs", "vps_security", "config.json")); err != nil {
		t.Fatalf("migrated config was not created: %v", err)
	}
}
