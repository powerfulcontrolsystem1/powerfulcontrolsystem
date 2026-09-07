package db

import "testing"

func TestLoadPostgresPoolConfigUsesBoundedRoleDefaults(t *testing.T) {
	t.Parallel()
	config, err := LoadPostgresPoolConfig(func(string) string { return "" }, "worker")
	if err != nil {
		t.Fatalf("LoadPostgresPoolConfig: %v", err)
	}
	if config.MaxOpen != 8 || config.MaxIdle != 4 {
		t.Fatalf("unexpected worker defaults: %#v", config)
	}
}

func TestLoadPostgresPoolConfigRejectsUnsafeValues(t *testing.T) {
	t.Parallel()
	_, err := LoadPostgresPoolConfig(func(key string) string {
		if key == "PCS_DB_MAX_OPEN_CONNS" {
			return "999"
		}
		return ""
	}, "api")
	if err == nil {
		t.Fatal("expected bounded pool validation error")
	}
}
