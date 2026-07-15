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

func TestLoadProductionCanExplicitlyEnableCompatibilityBootstrap(t *testing.T) {
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
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !config.LegacySchemaBootstrap {
		t.Fatalf("expected explicitly enabled bootstrap: %#v", config)
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

func TestLoadConfiguresBoundedDatabasePool(t *testing.T) {
	t.Parallel()
	values := map[string]string{
		"PCS_DB_MAX_OPEN_CONNS":     "40",
		"PCS_DB_MAX_IDLE_CONNS":     "12",
		"PCS_DB_CONN_MAX_LIFETIME":  "20m",
		"PCS_DB_CONN_MAX_IDLE_TIME": "5m",
		"PCS_DB_CONNECT_TIMEOUT":    "8s",
		"PCS_DB_QUERY_TIMEOUT":      "25s",
		"PCS_DB_TX_TIMEOUT":         "45s",
	}
	config, err := Load(func(key string) string { return values[key] })
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if config.Database.MaxOpenConns != 40 || config.Database.MaxIdleConns != 12 {
		t.Fatalf("unexpected pool: %#v", config.Database)
	}
	if config.Database.QueryTimeout.String() != "25s" || config.Database.TxTimeout.String() != "45s" {
		t.Fatalf("unexpected timeouts: %#v", config.Database)
	}
}

func TestLoadRejectsUnsafeDatabasePool(t *testing.T) {
	t.Parallel()
	_, err := Load(func(key string) string {
		if key == "PCS_DB_MAX_OPEN_CONNS" {
			return "0"
		}
		return ""
	})
	if err == nil {
		t.Fatal("expected invalid pool configuration")
	}
}
