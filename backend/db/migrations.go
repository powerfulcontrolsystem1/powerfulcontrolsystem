package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Migration is an immutable, ordered schema change. Body is deliberately
// explicit: it participates in the checksum so an edited migration is never
// silently accepted after it has reached an environment.
type Migration struct {
	Version     string
	Description string
	Body        string
	Apply       func(context.Context, *sql.Tx) error
}

type MigrationReport struct {
	Scope        string
	Applied      []string
	AlreadyKnown []string
	LegacyMarked []string
}

var (
	ErrMigrationDrift   = errors.New("migration checksum drift detected")
	ErrMigrationInvalid = errors.New("migration catalog is invalid")
)

// EnsureSchemaMigrationsTable bootstraps only the migration ledger. It is
// called exclusively by pcs-migrate before it obtains the advisory lock; API
// and worker processes must never invoke it.
func EnsureSchemaMigrationsTable(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("migration database is required")
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS schema_migrations (
		id BIGSERIAL PRIMARY KEY,
		scope TEXT NOT NULL,
		version TEXT NOT NULL,
		description TEXT,
		checksum TEXT,
		applied_by TEXT,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
		finished_at TIMESTAMPTZ,
		state TEXT NOT NULL DEFAULT 'applied',
		UNIQUE(scope, version)
	)`); err != nil {
		return err
	}
	for _, statement := range []string{
		`ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS checksum TEXT`,
		`ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS applied_by TEXT`,
		`ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS finished_at TIMESTAMPTZ`,
		`ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS state TEXT NOT NULL DEFAULT 'applied'`,
		`ALTER TABLE schema_migrations ALTER COLUMN applied_at TYPE TIMESTAMPTZ USING applied_at::timestamptz`,
		`UPDATE schema_migrations SET state = 'applied' WHERE state IS NULL OR btrim(state) = ''`,
		`CREATE TABLE IF NOT EXISTS schema_migration_runs (
			id BIGSERIAL PRIMARY KEY,
			scope TEXT NOT NULL,
			version TEXT NOT NULL,
			checksum TEXT NOT NULL,
			started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
			finished_at TIMESTAMPTZ,
			state TEXT NOT NULL,
			error_code TEXT,
			error_detail TEXT,
			applied_by TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_schema_migration_runs_scope_version ON schema_migration_runs (scope, version, id DESC)`,
	} {
		if _, err := execSQLCompat(dbConn, statement); err != nil {
			return err
		}
	}
	return nil
}

// RegisterSchemaMigration is retained for compatibility with old maintenance
// utilities. New production code must use RunMigrations so checksums and the
// PostgreSQL advisory lock are always enforced.
func RegisterSchemaMigration(dbConn *sql.DB, scope, version, description string) error {
	if strings.TrimSpace(scope) == "" || strings.TrimSpace(version) == "" {
		return fmt.Errorf("scope y version son obligatorios")
	}
	if err := EnsureSchemaMigrationsTable(dbConn); err != nil {
		return err
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO schema_migrations (scope, version, description, state, finished_at)
		VALUES (?, ?, ?, 'applied', CURRENT_TIMESTAMP)
		ON CONFLICT(scope, version) DO NOTHING`, strings.TrimSpace(scope), strings.TrimSpace(version), strings.TrimSpace(description))
	return err
}

// ApplySchemaMigration retains the historical callback API for tools that are
// not yet catalogued. New migrations must be declared in a catalog and run by
// RunMigrations; this helper deliberately cannot claim transactional safety for
// a callback that receives *sql.DB.
func ApplySchemaMigration(dbConn *sql.DB, scope, version, description string, applyFn func(*sql.DB) error) error {
	if strings.TrimSpace(scope) == "" || strings.TrimSpace(version) == "" {
		return fmt.Errorf("scope y version son obligatorios")
	}
	if err := EnsureSchemaMigrationsTable(dbConn); err != nil {
		return err
	}
	var exists int
	err := queryRowSQLCompat(dbConn, `SELECT 1 FROM schema_migrations WHERE scope = ? AND version = ? LIMIT 1`, scope, version).Scan(&exists)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if applyFn != nil {
		if err := applyFn(dbConn); err != nil {
			return err
		}
	}
	return RegisterSchemaMigration(dbConn, scope, version, description)
}

func MigrationChecksum(scope string, migration Migration) string {
	value := strings.Join([]string{
		strings.TrimSpace(scope),
		strings.TrimSpace(migration.Version),
		strings.TrimSpace(migration.Description),
		strings.TrimSpace(migration.Body),
	}, "\n")
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func ValidateMigrationCatalog(migrations []Migration) error {
	if len(migrations) == 0 {
		return fmt.Errorf("%w: no migrations declared", ErrMigrationInvalid)
	}
	previous := ""
	seen := make(map[string]struct{}, len(migrations))
	for _, migration := range migrations {
		version := strings.TrimSpace(migration.Version)
		if version == "" || strings.TrimSpace(migration.Description) == "" || strings.TrimSpace(migration.Body) == "" {
			return fmt.Errorf("%w: version, description and body are required", ErrMigrationInvalid)
		}
		if _, found := seen[version]; found {
			return fmt.Errorf("%w: repeated version %q", ErrMigrationInvalid, version)
		}
		if previous != "" && version <= previous {
			return fmt.Errorf("%w: version %q is not ordered after %q", ErrMigrationInvalid, version, previous)
		}
		seen[version] = struct{}{}
		previous = version
	}
	return nil
}

// RunMigrations applies one ordered catalog under a transaction-scoped
// PostgreSQL advisory lock. Every new migration and its ledger record commit
// atomically. Existing pre-checksum rows are marked as legacy only; they are
// never re-executed implicitly against live data.
func RunMigrations(ctx context.Context, dbConn *sql.DB, scope, appliedBy string, migrations []Migration) (MigrationReport, error) {
	report := MigrationReport{Scope: strings.TrimSpace(scope)}
	if dbConn == nil {
		return report, fmt.Errorf("migration database is required")
	}
	if report.Scope == "" {
		return report, fmt.Errorf("migration scope is required")
	}
	if !isPostgresDialect() {
		return report, fmt.Errorf("migrations require PostgreSQL")
	}
	if err := ValidateMigrationCatalog(migrations); err != nil {
		return report, err
	}
	if err := EnsureSchemaMigrationsTable(dbConn); err != nil {
		return report, fmt.Errorf("ensure migration ledger: %w", err)
	}

	for _, migration := range migrations {
		checksum := MigrationChecksum(report.Scope, migration)
		runID, err := startMigrationRun(ctx, dbConn, report.Scope, migration.Version, checksum, appliedBy)
		if err != nil {
			return report, err
		}
		outcome, err := runOneMigration(ctx, dbConn, report.Scope, appliedBy, migration, checksum)
		if err != nil {
			_ = finishMigrationRun(dbConn, runID, "failed", "migration_failed", redactMigrationError(err))
			return report, fmt.Errorf("migration %s/%s: %w", report.Scope, migration.Version, err)
		}
		if err := finishMigrationRun(dbConn, runID, outcome, "", ""); err != nil {
			return report, err
		}
		switch outcome {
		case "applied":
			report.Applied = append(report.Applied, migration.Version)
		case "legacy_marked":
			report.LegacyMarked = append(report.LegacyMarked, migration.Version)
		default:
			report.AlreadyKnown = append(report.AlreadyKnown, migration.Version)
		}
	}
	return report, nil
}

func runOneMigration(ctx context.Context, dbConn *sql.DB, scope, appliedBy string, migration Migration, checksum string) (string, error) {
	tx, err := dbConn.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := execTxSQLCompat(tx, `SELECT pg_advisory_xact_lock(hashtext(?))`, "pcs-schema-migrations:"+scope); err != nil {
		return "", fmt.Errorf("acquire advisory lock: %w", err)
	}

	var storedChecksum sql.NullString
	err = queryRowTxSQLCompat(tx, `SELECT checksum FROM schema_migrations WHERE scope = ? AND version = ? FOR UPDATE`, scope, migration.Version).Scan(&storedChecksum)
	if err == nil {
		if strings.TrimSpace(storedChecksum.String) == "" {
			if _, updateErr := execTxSQLCompat(tx, `UPDATE schema_migrations
				SET checksum = ?, description = ?, applied_by = COALESCE(NULLIF(applied_by, ''), ?),
					finished_at = COALESCE(finished_at, CURRENT_TIMESTAMP), state = 'applied'
				WHERE scope = ? AND version = ?`, checksum, migration.Description, appliedBy, scope, migration.Version); updateErr != nil {
				return "", updateErr
			}
			if err := tx.Commit(); err != nil {
				return "", err
			}
			return "legacy_marked", nil
		}
		if storedChecksum.String != checksum {
			return "", fmt.Errorf("%w for %s/%s", ErrMigrationDrift, scope, migration.Version)
		}
		if err := tx.Commit(); err != nil {
			return "", err
		}
		return "already_applied", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if migration.Apply != nil {
		if err := migration.Apply(ctx, tx); err != nil {
			return "", err
		}
	}
	if _, err := execTxSQLCompat(tx, `INSERT INTO schema_migrations
		(scope, version, description, checksum, applied_by, applied_at, finished_at, state)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'applied')`,
		scope, migration.Version, migration.Description, checksum, strings.TrimSpace(appliedBy)); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return "applied", nil
}

func startMigrationRun(ctx context.Context, dbConn *sql.DB, scope, version, checksum, appliedBy string) (int64, error) {
	var id int64
	err := queryRowSQLCompat(dbConn, `INSERT INTO schema_migration_runs
		(scope, version, checksum, state, applied_by) VALUES (?, ?, ?, 'running', ?) RETURNING id`,
		scope, version, checksum, strings.TrimSpace(appliedBy)).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func finishMigrationRun(dbConn *sql.DB, id int64, state, code, detail string) error {
	if id <= 0 {
		return nil
	}
	_, err := execSQLCompat(dbConn, `UPDATE schema_migration_runs
		SET state = ?, error_code = NULLIF(?, ''), error_detail = NULLIF(?, ''), finished_at = CURRENT_TIMESTAMP
		WHERE id = ?`, state, code, detail, id)
	return err
}

func redactMigrationError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	if len(message) > 500 {
		message = message[:500]
	}
	for _, marker := range []string{"password=", "token=", "secret=", "postgres://", "postgresql://"} {
		if index := strings.Index(strings.ToLower(message), marker); index >= 0 {
			return strings.TrimSpace(message[:index]) + "[redacted]"
		}
	}
	return message
}

// SortMigrationsByVersion is used only by tests and tooling that build a
// catalog dynamically. Production catalogs remain declared in source order.
func SortMigrationsByVersion(migrations []Migration) {
	sort.SliceStable(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
}

// MigrationRunAge is intentionally small and pure so operations can expose
// safe migration latency metrics without querying business tables.
func MigrationRunAge(startedAt time.Time) time.Duration {
	if startedAt.IsZero() {
		return 0
	}
	return time.Since(startedAt)
}
