package db

import (
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
