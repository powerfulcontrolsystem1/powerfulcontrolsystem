package db

import "testing"

func TestHistoricalIdempotencyMigrationBodiesRemainFrozen(t *testing.T) {
	if mobileAPIIdempotencySchemaFingerprint != "empresa_mobile_api_idempotencia:v2:tenant-operation-key-hash-response-expiry" {
		t.Fatal("la migracion movil v2 no puede cambiar de cuerpo")
	}
	if empresaCxPAtomicSchemaFingerprint != "empresa_cxp_pagos:v1:tenant-lock-idempotency-finance-outbox" {
		t.Fatal("la migracion CxP v1 no puede cambiar de cuerpo")
	}
}

func TestEmpresaCatalogIncludesOperationalIdempotencyMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260826-003-operational-idempotency-v1" {
			continue
		}
		if migration.Apply == nil || migration.Body != empresaOperationalIdempotencyFingerprint {
			t.Fatal("la migracion operativa idempotente debe ser ejecutable e inmutable")
		}
		return
	}
	t.Fatal("falta la migracion operativa idempotente en el catalogo empresarial")
}
