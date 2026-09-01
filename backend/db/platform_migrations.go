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

	// legacySchemaCatalogManifestV1SourceFingerprint is the source fingerprint
	// recorded when version 20260717-001 was first released. Migration bodies
	// are immutable: using the current generated catalog fingerprint here would
	// make a later catalog change invalidate an already-applied migration.
	legacySchemaCatalogManifestV1SourceFingerprint = "2a:db:98:91:89:0f:14:d5:ce:6b:94:3c:ac:56:b3:d6:25:35:bd:57:3a:0b:6e:60:5a:cc:4e:db:b3:7e:11:d3"
)

// PlatformMigrations owns the runtime foundation and records the one-time
// reviewed legacy baseline. API and worker processes only verify this ledger.
func PlatformMigrations(target string) ([]Migration, error) {
	if err := ValidateLegacySchemaCatalogManifest(); err != nil {
		return nil, err
	}
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
			{
				Version:     "20260717-001-legacy-schema-manifest-v1",
				Description: "frozen source fingerprint for legacy enterprise schema catalog",
				Body:        legacySchemaCatalogManifestV1SourceFingerprint + ":empresas",
			},
			{
				Version:     "20260724-001-cxp-atomic-payments-v1",
				Description: "atomic supplier payable payment allocations",
				Body:        empresaCxPAtomicSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyEmpresaCxPAtomicSchemaTx(tx)
				},
			},
			{
				Version:     "20260725-001-ai-user-isolation-v1",
				Description: "tenant and user isolation for enterprise AI preferences and history",
				Body:        empresaAIUserIsolationSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyEmpresaAIUserIsolationSchemaTx(tx)
				},
			},
			{
				Version:     "20260730-001-nextcloud-accounts-v2",
				Description: "repair provisioned flag on legacy Nextcloud assignments",
				Body:        empresaNextcloudSchemaRepairFingerprint,
				Apply: func(ctx context.Context, tx *sql.Tx) error {
					return applyEmpresaNextcloudSchemaRepairTx(ctx, tx)
				},
			},
			{
				Version:     "20260730-002-nextcloud-accounts-v3",
				Description: "complete required columns on legacy Nextcloud assignments",
				Body:        empresaNextcloudSchemaCompleteRepairFingerprint,
				Apply: func(ctx context.Context, tx *sql.Tx) error {
					return applyEmpresaNextcloudSchemaCompleteRepairTx(ctx, tx)
				},
			},
			{
				Version:     "20260730-003-nextcloud-accounts-v4",
				Description: "remove obsolete locally stored Nextcloud credential column",
				Body:        empresaNextcloudSchemaCredentialCleanupFingerprint,
				Apply: func(ctx context.Context, tx *sql.Tx) error {
					return applyEmpresaNextcloudSchemaCredentialCleanupTx(ctx, tx)
				},
			},
			{
				Version:     "20260730-004-outbox-recovery-audit-v1",
				Description: "audited tenant-scoped recovery for dead outbox events",
				Body:        outboxRecoveryAuditSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyOutboxRecoveryAuditSchemaTx(tx)
				},
			},
			{
				Version:     "20260731-001-ai-usage-unique-v1",
				Description: "unique daily enterprise AI usage per tenant, provider and model",
				Body:        empresaAIUsoDiarioUniqueSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyEmpresaAIUsoDiarioUniqueSchemaTx(tx)
				},
			},
			{
				Version:     "20260801-001-cartera-money-precision-v1",
				Description: "exact two-decimal balances for accounts receivable and payable",
				Body:        empresaCarteraMoneyPrecisionFingerprint,
				Apply: func(ctx context.Context, tx *sql.Tx) error {
					return applyEmpresaCarteraMoneyPrecisionTx(ctx, tx)
				},
			},
			{
				Version:     "20260809-001-dian-local-production-flag-v1",
				Description: "dedicated local production activation gate for each DIAN tenant",
				Body:        empresaDIANLocalProductionFlagFingerprint,
				Apply: func(ctx context.Context, tx *sql.Tx) error {
					return applyEmpresaDIANLocalProductionFlagTx(ctx, tx)
				},
			},
			{
				Version:     "20260809-002-domotica-raspberry-tunnel-v1",
				Description: "secure outbound Raspberry tunnel, GPIO inputs and daily transfer accounting",
				Body:        empresaControlElectricoTunnelSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyEmpresaControlElectricoTunnelSchemaTx(tx)
				},
			},
			{
				Version:     "20260810-001-domotica-station-automation-v1",
				Description: "independent relay policies for station activation and deactivation",
				Body:        empresaControlElectricoStationAutomationSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyEmpresaControlElectricoStationAutomationSchemaTx(tx)
				},
			},
			{
				Version:     "20260811-001-domotica-datetime-schedule-v1",
				Description: "dated start and end scheduling for electronic equipment",
				Body:        empresaControlElectricoScheduleSchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyEmpresaControlElectricoScheduleSchemaTx(tx)
				},
			},
			{
				Version:     "20260811-002-domotica-restart-category-v1",
				Description: "idempotent Raspberry restart recovery and electronic equipment categories",
				Body:        empresaControlElectricoRestartCategorySchemaFingerprint,
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return applyEmpresaControlElectricoRestartCategorySchemaTx(tx)
				},
			},
			{Version: "20260811-003-domotica-timer-v1", Description: "timer duration for electronic equipment and sensor rules", Body: empresaControlElectricoTimerSchemaFingerprint, Apply: func(_ context.Context, tx *sql.Tx) error { return applyEmpresaControlElectricoTimerSchemaTx(tx) }},
			{Version: "20260812-001-domotica-activation-queue-v1", Description: "company-wide serialized relay activation delay", Body: empresaControlElectricoActivationQueueSchemaFingerprint, Apply: func(_ context.Context, tx *sql.Tx) error {
				return applyEmpresaControlElectricoActivationQueueSchemaTx(tx)
			}},
			{Version: "20260813-001-domotica-governance-v1", Description: "Raspberry disconnect alerts and per-company monthly tunnel limits", Body: empresaControlElectricoGovernanceSchemaFingerprint, Apply: func(_ context.Context, tx *sql.Tx) error { return applyEmpresaControlElectricoGovernanceSchemaTx(tx) }},
			{Version: "20260823-001-retire-grafologia-v1", Description: "remove retired handwriting-analysis module data", Body: empresaGrafologiaDecommissionFingerprint, Apply: applyEmpresaGrafologiaDecommissionTx},
			{Version: "20260823-002-retire-vertical-modules-v1", Description: "remove retired gym taxi apartments horizontal property dentistry and pharmacy module data", Body: empresaVerticalModulesDecommissionFingerprint, Apply: applyEmpresaVerticalModulesDecommissionTx},
			{Version: "20260826-001-sale-accounting-idempotency-v1", Description: "one active paid-sale accounting event per tenant cart", Body: empresaSaleAccountingIdempotencyFingerprint, Apply: applyEmpresaSaleAccountingIdempotencyTx},
			{Version: "20260826-002-cart-sale-history-v1", Description: "immutable cart sale history for cash reports", Body: empresaCarritoSaleHistorySchemaFingerprint, Apply: func(_ context.Context, tx *sql.Tx) error {
				return applyEmpresaCarritoSaleHistorySchemaTx(tx)
			}},
			{Version: "20260826-003-operational-idempotency-v1", Description: "request fingerprints and leased claims for CxP offline sales and DIAN retries", Body: empresaOperationalIdempotencyFingerprint, Apply: applyEmpresaOperationalIdempotencyTx},
			{Version: "20260831-001-vida-personal-v1", Description: "personal expenses receipts subscriptions and reminders isolated by tenant and user", Body: empresaVidaSchemaFingerprint, Apply: applyEmpresaVidaSchemaTx},
			{Version: "20260831-002-raspberry-door-sensor-v1", Description: "multiplexed four-input door sensors for enrolled Raspberry controllers", Body: empresaControlElectricoDoorSensorSchemaFingerprint, Apply: func(_ context.Context, tx *sql.Tx) error {
				return applyEmpresaControlElectricoDoorSensorSchemaTx(tx)
			}},
			{Version: "20260831-003-vida-price-history-ai-v1", Description: "personal price history barcode captures and AI invoice line items", Body: empresaVidaPriceHistorySchemaFingerprint, Apply: applyEmpresaVidaPriceHistorySchemaTx},
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
			{
				Version:     "20260717-001-legacy-schema-manifest-v1",
				Description: "frozen source fingerprint for legacy administrative schema catalog",
				Body:        legacySchemaCatalogManifestV1SourceFingerprint + ":superadministrador",
			},
			{
				Version:     "20260722-001-session-token-hashes-v1",
				Description: "hashed administrator session tokens owned by migration role",
				Body:        "ALTER TABLE sesiones ADD COLUMN token_hash; migrate legacy tokens transactionally; migration-role-only",
				Apply: func(_ context.Context, tx *sql.Tx) error {
					return migrateSessionTokensToHashesTx(tx)
				},
			},
			{
				Version:     "20260728-001-portal-visitas-v1",
				Description: "public portal visit counter owned by migration role",
				Body:        portalVisitasSchemaFingerprint,
				Apply: func(ctx context.Context, tx *sql.Tx) error {
					return applyPortalVisitasSchemaTx(ctx, tx)
				},
			},
			{
				Version:     "20260826-001-payment-idempotency-v1",
				Description: "durable payment checkout and post-effect idempotency",
				Body:        paymentIdempotencySchemaFingerprint,
				Apply:       applyPaymentIdempotencySchemaTx,
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
