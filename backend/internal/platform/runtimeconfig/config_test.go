package runtimeconfig

import "testing"

func TestLoadProductionDisablesCompatibilityBootstrapByDefault(t *testing.T) {
	t.Parallel()
	config, err := Load(func(key string) string {
		if key == "PCS_ENV" {
			return "production"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if config.Role != RoleAPI || config.LegacySchemaBootstrap {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestLoadProductionRejectsAPIBootstrapOverride(t *testing.T) {
	t.Parallel()
	config, err := Load(func(key string) string {
		switch key {
		case "PCS_ENV":
			return "production"
		case "PCS_RUNTIME_SCHEMA_BOOTSTRAP":
			return "1"
		}
		return ""
	})
	if err == nil {
		t.Fatalf("expected production API bootstrap override rejection: %#v", config)
	}
}

func TestLoadProductionRejectsWorkerBootstrapOverride(t *testing.T) {
	t.Parallel()
	config, err := Load(func(key string) string {
		switch key {
		case "PCS_ENV":
			return "production"
		case "PCS_RUNTIME_ROLE":
			return "worker"
		case "PCS_RUNTIME_SCHEMA_BOOTSTRAP":
			return "1"
		}
		return ""
	})
	if err == nil {
		t.Fatalf("expected production worker bootstrap override rejection: %#v", config)
	}
}

func TestLoadMigrationRoleEnablesSchemaBootstrap(t *testing.T) {
	t.Parallel()
	config, err := Load(func(key string) string {
		switch key {
		case "PCS_ENV":
			return "production"
		case "PCS_RUNTIME_ROLE":
			return "migrate"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !config.LegacySchemaBootstrap || config.Role != RoleMigrate {
		t.Fatalf("unexpected config: %#v", config)
	}
}

func TestLoadRejectsUnknownRole(t *testing.T) {
	t.Parallel()
	if _, err := Load(func(key string) string {
		if key == "PCS_RUNTIME_ROLE" {
			return "unknown"
		}
		return ""
	}); err == nil {
		t.Fatal("expected invalid role error")
	}
}
