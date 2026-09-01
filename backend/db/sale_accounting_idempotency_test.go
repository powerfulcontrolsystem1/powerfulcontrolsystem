package db

import (
	"os"
	"strings"
	"testing"
)

func TestPaidCartAccountingUsesDatabaseUniqueness(t *testing.T) {
	raw, err := os.ReadFile("eventos_contables.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, required := range []string{
		"ux_empresa_eventos_contables_venta_pagada_carrito",
		"ON CONFLICT (empresa_id, modulo, evento, entidad, entidad_id)",
		"e.Modulo == \"ventas\" && e.Evento == \"venta_pagada\"",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("falta garantia contable idempotente %q", required)
		}
	}
}

func TestEmpresaCatalogIncludesSaleAccountingIdempotencyMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260826-001-sale-accounting-idempotency-v1" {
			continue
		}
		if migration.Apply == nil || migration.Body != empresaSaleAccountingIdempotencyFingerprint {
			t.Fatal("la migracion contable idempotente debe ser ejecutable e inmutable")
		}
		return
	}
	t.Fatal("falta la migracion contable idempotente en el catalogo empresarial")
}
