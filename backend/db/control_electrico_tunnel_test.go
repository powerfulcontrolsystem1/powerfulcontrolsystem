package db

import (
	"database/sql"
	"strings"
	"testing"
)

func TestControlElectricoTunnelSecretsAreRandomAndHashed(t *testing.T) {
	first, err := generateControlElectricoTunnelSecret(32)
	if err != nil {
		t.Fatal(err)
	}
	second, err := generateControlElectricoTunnelSecret(32)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) < 40 {
		t.Fatalf("secreto aleatorio invalido: len=%d", len(first))
	}
	hash := controlElectricoTunnelTokenHash(first)
	if len(hash) != 64 || hash == first || strings.Contains(hash, first) {
		t.Fatal("el token debe persistirse unicamente como SHA-256 hexadecimal")
	}
}

func TestControlElectricoStationQueriesRequireTenantAndStation(t *testing.T) {
	var conn *sql.DB
	if _, err := ListEmpresaControlElectricoReglasByEstacion(conn, 1, 0, false); err == nil {
		t.Fatal("las reglas por estacion deben rechazar estacion_id invalido")
	}
	if _, err := ListEmpresaControlElectricoReglasByEstacion(conn, 0, 1, false); err == nil {
		t.Fatal("las reglas por estacion deben rechazar empresa_id invalido")
	}
	if _, err := ListEmpresaControlElectricoEventosByEstacion(conn, 1, 0, 20); err == nil {
		t.Fatal("los eventos por estacion deben rechazar estacion_id invalido")
	}
	if _, err := ListEmpresaControlElectricoEventosByEstacion(conn, 0, 1, 20); err == nil {
		t.Fatal("los eventos por estacion deben rechazar empresa_id invalido")
	}
}

func TestControlElectricoTunnelDeviceUIDIsUniqueAndOpaque(t *testing.T) {
	first, err := generateControlElectricoTunnelDeviceUID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := generateControlElectricoTunnelDeviceUID()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.HasPrefix(first, "RPI-") || len(first) != 36 {
		t.Fatalf("device UID invalido: %q", first)
	}
}

func TestEmpresaCatalogIncludesControlElectricoTunnelMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260809-002-domotica-raspberry-tunnel-v1" {
			continue
		}
		if migration.Body != empresaControlElectricoTunnelSchemaFingerprint || migration.Apply == nil {
			t.Fatal("la migracion de tunel debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("falta la migracion versionada del tunel de domotica")
}

func TestEmpresaCatalogIncludesControlElectricoScheduleMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260811-001-domotica-datetime-schedule-v1" {
			continue
		}
		if migration.Body != empresaControlElectricoScheduleSchemaFingerprint || migration.Apply == nil {
			t.Fatal("la migracion de programacion fechada debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("falta la migracion versionada de programacion fechada de domotica")
}

func TestEmpresaCatalogIncludesControlElectricoRestartCategoryMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260811-002-domotica-restart-category-v1" {
			continue
		}
		if migration.Body != empresaControlElectricoRestartCategorySchemaFingerprint || migration.Apply == nil {
			t.Fatal("la migracion de reinicio y categoria debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("falta la migracion versionada de reinicio y categoria de domotica")
}
