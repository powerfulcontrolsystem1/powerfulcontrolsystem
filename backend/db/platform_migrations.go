package db

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	MigrationTargetEmpresas = "empresas"
	MigrationTargetSuper    = "superadministrador"
	platformMigrationScope  = "platform"
)

// PlatformMigrations owns only the new runtime foundation. The historical
// bootstrap remains opt-in until every legacy Ensure* mutation has been moved
// into a reviewed catalog; this prevents an unsafe all-at-once schema cutover.
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
				Version:     "20260716-001-mobile-idempotency-v2",
				Description: "durable mobile idempotency schema",
				Body:        mobileAPIIdempotencySchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyMobileAPIIdempotencySchemaTx(tx)
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
		}, nil
	default:
		return nil, fmt.Errorf("unknown platform migration target %q", target)
	}
}

func ApplyPlatformMigrations(ctx context.Context, dbConn *sql.DB, target, appliedBy string) (MigrationReport, error) {
	migrations, err := PlatformMigrations(target)
	if err != nil {
		return MigrationReport{Scope: target}, err
	}
	return RunMigrations(ctx, dbConn, platformMigrationScope, appliedBy, migrations)
}
