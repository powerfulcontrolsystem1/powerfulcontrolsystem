package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const (
	MigrationTargetEmpresas = "empresas"
	MigrationTargetSuper    = "superadministrador"
	platformMigrationScope  = "platform"
)

// PlatformMigrations owns the runtime foundation and records the one-time
// reviewed legacy baseline. API and worker processes only verify this ledger.
func PlatformMigrations(target string) ([]Migration, error) {
	switch target {
	case MigrationTargetEmpresas:
		return []Migration{
			{
				Version:     "20260714-runtime-foundation",
				Description: "runtime roles, durable queue and outbox",
				Body:        "legacy ledger marker; no implicit replay",
			},
			{
				Version:     legacySchemaBaselineVersion,
				Description: "legacy business schema baseline executed by migration role",
				Body:        "legacy-schema-bootstrap:v1:empresas:migration-role-only",
			},
			{
				Version:     "20260716-001-mobile-idempotency-v2",
				Description: "durable mobile idempotency schema",
				Body:        mobileAPIIdempotencySchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyMobileAPIIdempotencySchemaTx(tx)
				},
			},
			{
				Version:     "20260716-002-nextcloud-accounts-v1",
				Description: "enterprise Nextcloud account assignments",
				Body:        empresaNextcloudSchemaFingerprint,
				Apply: func(ctx context.Context, tx *sql.Tx) error {
					return applyEmpresaNextcloudSchemaTx(ctx, tx)
				},
			},
			{
				Version:     "20260716-003-durable-outbox-v2",
				Description: "tenant transactional outbox source",
				Body:        outboxSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyOutboxSchemaTx(tx)
				},
			},
		}, nil
	case MigrationTargetSuper:
		return []Migration{
			{
				Version:     "20260714-runtime-foundation",
				Description: "runtime roles, durable queue and outbox",
				Body:        "legacy ledger marker; no implicit replay",
			},
			{
				Version:     legacySchemaBaselineVersion,
				Description: "legacy administrative schema baseline executed by migration role",
				Body:        "legacy-schema-bootstrap:v1:superadministrador:migration-role-only",
			},
			{
				Version:     "20260716-001-durable-async-jobs-v2",
				Description: "leased async jobs, recovery and idempotency",
				Body:        asyncJobsSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyAsyncJobsSchemaTx(tx)
				},
			},
			{
				Version:     "20260716-002-durable-outbox-v2",
				Description: "leased transactional outbox",
				Body:        outboxSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyOutboxSchemaTx(tx)
				},
			},
			{
				Version:     "20260716-004-system-metrics-v1",
				Description: "durable system metrics schema owned by migration role",
				Body:        metricsSchemaFingerprint,
				Apply: func(ctx context.Context, tx *sql.Tx) error {
					return applyMetricsSchemaTx(ctx, tx)
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown platform migration target %q", target)
	}
}

// VerifyPlatformMigrations is read-only and suitable for API/worker readiness.
func VerifyPlatformMigrations(ctx context.Context, dbConn *sql.DB, target string) error {
	if dbConn == nil {
		return fmt.Errorf("migration database is required")
	}
	migrations, err := PlatformMigrations(target)
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		var checksum sql.NullString
		err := dbConn.QueryRowContext(ctx, `SELECT checksum FROM schema_migrations WHERE scope = $1 AND version = $2 AND state = 'applied'`, platformMigrationScope, migration.Version).Scan(&checksum)
		if err != nil {
			return fmt.Errorf("required migration %s/%s is not applied", target, migration.Version)
		}
		if !checksum.Valid || strings.TrimSpace(checksum.String) != MigrationChecksum(platformMigrationScope, migration) {
			return fmt.Errorf("required migration %s/%s has invalid checksum", target, migration.Version)
		}
	}
	return nil
}

func ApplyPlatformMigrations(ctx context.Context, dbConn *sql.DB, target, appliedBy string) (MigrationReport, error) {
	migrations, err := PlatformMigrations(target)
	if err != nil {
		return MigrationReport{Scope: target}, err
	}
	return RunMigrations(ctx, dbConn, platformMigrationScope, appliedBy, migrations)
}
