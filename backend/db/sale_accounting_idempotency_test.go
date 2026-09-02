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
		"ON CONFLICT (empresa_id, clave_idempotencia)",
		"clave de idempotencia obligatoria para venta pagada",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("falta garantia contable idempotente %q", required)
		}
	}
}

func TestSaleAccountingOperationKeyPrecedesImmutableCartMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	operationKeyIndex := -1
	legacyIndex := -1
	for i, migration := range migrations {
		switch migration.Version {
		case "20260826-000-sale-accounting-operation-key-v1":
			operationKeyIndex = i
			if migration.Apply == nil || migration.Body != empresaSaleAccountingOperationKeyFingerprint {
				t.Fatal("la migracion de clave operacional debe ser ejecutable e inmutable")
			}
		case "20260826-001-sale-accounting-idempotency-v1":
			legacyIndex = i
		}
	}
	if operationKeyIndex < 0 || legacyIndex < 0 || operationKeyIndex >= legacyIndex {
		t.Fatalf("la preparacion operacional debe anteceder la migracion historica: operation=%d legacy=%d", operationKeyIndex, legacyIndex)
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
