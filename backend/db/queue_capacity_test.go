package db

import "testing"

func TestDefaultQueueCapacityConfigsCoverIndependentLanes(t *testing.T) {
	configs := defaultQueueCapacityConfigs()
	if len(configs) != 3 {
		t.Fatalf("se esperaban tres carriles, got %d", len(configs))
	}
	seen := map[string]bool{}
	for _, cfg := range configs {
		if err := validateQueueCapacityConfig(cfg); err != nil {
			t.Fatalf("configuracion default invalida para %s: %v", cfg.Lane, err)
		}
		if seen[cfg.Lane] {
			t.Fatalf("carril duplicado: %s", cfg.Lane)
		}
		seen[cfg.Lane] = true
	}
	for _, lane := range []string{QueueLanePrinting, QueueLaneProductAdd, QueueLaneFiscal} {
		if !seen[lane] {
			t.Fatalf("falta carril %s", lane)
		}
	}
}

func TestQueueLaneForRateScope(t *testing.T) {
	cases := map[string]string{
		"queue.printing":    QueueLanePrinting,
		"queue.product_add": QueueLaneProductAdd,
		"queue.fiscal":      QueueLaneFiscal,
		"api":               "",
	}
	for scope, want := range cases {
		if got := QueueLaneForRateScope(scope); got != want {
			t.Fatalf("scope %s: got %s want %s", scope, got, want)
		}
	}
}

func TestQueueSaturationUsesMostRestrictiveSignal(t *testing.T) {
	cfg := QueueCapacityConfig{PendingAlertThreshold: 100, OldestAlertSeconds: 60, MaxPendingPerTenant: 20}
	snapshot := QueueCapacitySnapshot{Pending: 40, OldestSeconds: 90, BusiestTenantPending: 10}
	if got := queueSaturationPercent(snapshot, cfg); got != 150 {
		t.Fatalf("saturacion inesperada: got %.2f want 150", got)
	}
}

func TestQueueCapacityMigrationsAreRegistered(t *testing.T) {
	targets := []struct {
		target      string
		version     string
		fingerprint string
	}{
		{MigrationTargetEmpresas, "20260906-001-queue-capacity-business-v1", queueCapacityBusinessSchemaFingerprint},
		{MigrationTargetSuper, "20260906-001-queue-capacity-super-v1", queueCapacitySuperSchemaFingerprint},
	}
	for _, target := range targets {
		migrations, err := PlatformMigrations(target.target)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, migration := range migrations {
			if migration.Version != target.version {
				continue
			}
			found = migration.Body == target.fingerprint && migration.Apply != nil
		}
		if !found {
			t.Fatalf("migracion de capacidad ausente o incompleta: %s", target.version)
		}
	}
}
