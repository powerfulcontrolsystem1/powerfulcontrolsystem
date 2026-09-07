package db

import "testing"

func TestWMSOrderTransitionUsesLockedTotals(t *testing.T) {
	if err := validateWMSOrdenTransitionWithTotals("en_packing", "lista_despacho", wmsOrdenTotals{Items: 2, Total: 10, Empacada: 9}); err == nil {
		t.Fatal("una orden incompleta pudo pasar a lista de despacho")
	}
	if err := validateWMSOrdenTransitionWithTotals("en_packing", "lista_despacho", wmsOrdenTotals{Items: 2, Total: 10, Empacada: 10}); err != nil {
		t.Fatalf("una orden completamente empacada fue rechazada: %v", err)
	}
}

func TestPlatformSuperMigrationsIncludeSharedTenantRateLimit(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetSuper)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260905-002-empresa-rate-limit-v1" {
			return
		}
	}
	t.Fatal("falta migracion compartida de limite por empresa")
}
