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
		case "PCS_RUNTIME_SCHEMA_BOOTSTRAP":
			return "1"
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

func TestLoadProductionMigrationRequiresExplicitBootstrapSetting(t *testing.T) {
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
	if err == nil {
		t.Fatalf("expected explicit production migration bootstrap setting, got %#v", config)
	}
}

func TestLoadMigrationRoleAllowsExplicitLegacyBootstrapDisable(t *testing.T) {
	t.Parallel()
	config, err := Load(func(key string) string {
		switch key {
		case "PCS_ENV":
			return "production"
		case "PCS_RUNTIME_ROLE":
			return "migrate"
		case "PCS_RUNTIME_SCHEMA_BOOTSTRAP":
			return "0"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if config.LegacySchemaBootstrap {
		t.Fatalf("explicit migration bootstrap disable was ignored: %#v", config)
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

func TestRuntimeEnvironmentHelpers(t *testing.T) {
	t.Parallel()
	env := map[string]string{
		"SECOND":                " value ",
		"DB_VPS_TUNNEL_ENABLED": "1",
		"DB_VPS_LOCAL_PORT":     "55432",
	}
	getenv := func(key string) string { return env[key] }
	if got := FirstNonEmptyEnv(getenv, "FIRST", "SECOND"); got != "value" {
		t.Fatalf("FirstNonEmptyEnv() = %q", got)
	}
	got := RewritePostgresDSNForTunnel("postgres://user:pass@localhost:5432/pcs", getenv)
	want := "postgres://user:pass@127.0.0.1:55432/pcs"
	if got != want {
		t.Fatalf("RewritePostgresDSNForTunnel() = %q, want %q", got, want)
	}
	remote := "postgres://user:pass@db.example.com:5432/pcs"
	if got := RewritePostgresDSNForTunnel(remote, getenv); got != remote {
		t.Fatalf("remote DSN must remain unchanged: %q", got)
	}
}
