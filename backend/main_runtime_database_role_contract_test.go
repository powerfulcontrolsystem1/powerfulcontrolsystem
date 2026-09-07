package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlatformComposeSeparatesMigrationAndRuntimeDatabaseUsers(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "docker-compose.platform.yml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	for _, required := range []string{
		"PCS_RUNTIME_DB_USER: ${PCS_RUNTIME_DB_USER:?Defina PCS_RUNTIME_DB_USER}",
		"PCS_RUNTIME_DB_PASSWORD: ${PCS_RUNTIME_DB_PASSWORD:?Defina PCS_RUNTIME_DB_PASSWORD}",
		"PCS_BACKUP_DB_USER: ${PCS_BACKUP_DB_USER:?Defina PCS_BACKUP_DB_USER}",
		"PCS_BACKUP_DB_PASSWORD: ${PCS_BACKUP_DB_PASSWORD:?Defina PCS_BACKUP_DB_PASSWORD}",
		"postgres://${PCS_RUNTIME_DB_USER:?Defina PCS_RUNTIME_DB_USER}:${PCS_RUNTIME_DB_PASSWORD:?Defina PCS_RUNTIME_DB_PASSWORD}@postgres:5432/pcs_empresas",
		"postgres://${PCS_RUNTIME_DB_USER:?Defina PCS_RUNTIME_DB_USER}:${PCS_RUNTIME_DB_PASSWORD:?Defina PCS_RUNTIME_DB_PASSWORD}@postgres:5432/pcs_superadministrador",
		"PGUSER: ${PCS_BACKUP_DB_USER:?Defina PCS_BACKUP_DB_USER}",
		"PGPASSWORD: ${PCS_BACKUP_DB_PASSWORD:?Defina PCS_BACKUP_DB_PASSWORD}",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("runtime database separation is missing %q", required)
		}
	}
	if got := strings.Count(content, "postgres://${POSTGRES_USER:-pcs}:${POSTGRES_PASSWORD}@postgres:5432/"); got != 2 {
		t.Fatalf("only migrate may use owner DSNs; found %d owner references, want 2", got)
	}
}

func TestProductionBootstrapBackfillsBackupRoleWithoutLoggingSecret(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "scripts", "sync_to_vps.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(raw)
	for _, required := range []string{
		"current_backup_db_user=pcs_backup",
		"openssl rand -hex 24",
		"PCS_BACKUP_DB_PASSWORD generado y guardado sin exponer su valor",
		"BACKUP_DB_ROLE_COLLISION",
		"INVALID_BACKUP_DB_PASSWORD",
		"unset current_backup_db_password",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("backup role bootstrap contract is missing %q", required)
		}
	}
	if strings.Contains(content, `ok "PCS_BACKUP_DB_PASSWORD=$current_backup_db_password"`) {
		t.Fatal("backup role bootstrap must not log the generated password")
	}
}

func TestFrontendWaitsForHealthyBackendAndProbesDynamicRoutes(t *testing.T) {
	composeRaw, err := os.ReadFile(filepath.Join("..", "deploy", "docker-compose.platform.yml"))
	if err != nil {
		t.Fatal(err)
	}
	compose := string(composeRaw)
	for _, required := range []string{
		"condition: service_healthy",
		`test: ["CMD", "wget", "-q", "--spider", "http://127.0.0.1:8080/ready"]`,
	} {
		if !strings.Contains(compose, required) {
			t.Fatalf("frontend health contract is missing %q", required)
		}
	}

	sidecarRaw, err := os.ReadFile(filepath.Join("..", "deploy", "scripts", "vps-compose-sidecar-up.sh"))
	if err != nil {
		t.Fatal(err)
	}
	sidecar := string(sidecarRaw)
	for _, required := range []string{
		"restart frontend",
		`$http_port/health`,
		`$http_port/ready`,
	} {
		if !strings.Contains(sidecar, required) {
			t.Fatalf("sidecar dynamic health contract is missing %q", required)
		}
	}
}
