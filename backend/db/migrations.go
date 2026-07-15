package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// EnsureSchemaMigrationsTable crea la tabla de control de migraciones versionadas.
func EnsureSchemaMigrationsTable(dbConn *sql.DB) error {
	createStmt := `CREATE TABLE IF NOT EXISTS schema_migrations (
		id BIGSERIAL PRIMARY KEY,
		scope TEXT NOT NULL,
		version TEXT NOT NULL,
		description TEXT,
		applied_at TEXT DEFAULT (CURRENT_TIMESTAMP),
		UNIQUE(scope, version)
	);`
	if isPostgresDialect() {
		createStmt = `CREATE TABLE IF NOT EXISTS schema_migrations (
			id BIGSERIAL PRIMARY KEY,
			scope TEXT NOT NULL,
			version TEXT NOT NULL,
			description TEXT,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(scope, version)
		);`
	}
	if _, err := execSQLCompat(dbConn, createStmt); err != nil {
		return err
	}
	for _, statement := range []string{
		`ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS started_at TIMESTAMPTZ`,
		`ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS duration_ms BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS result TEXT NOT NULL DEFAULT 'applied'`,
		`ALTER TABLE schema_migrations ADD COLUMN IF NOT EXISTS applied_by TEXT NOT NULL DEFAULT ''`,
	} {
		if _, err := execSQLCompat(dbConn, statement); err != nil {
			return err
		}
	}
	return nil
}

// RegisterSchemaMigration registra una version de esquema de forma idempotente.
func RegisterSchemaMigration(dbConn *sql.DB, scope, version, description string) error {
	if scope == "" || version == "" {
		return fmt.Errorf("scope y version son obligatorios")
	}
	if err := EnsureSchemaMigrationsTable(dbConn); err != nil {
		return err
	}
	insertStmt := `INSERT OR IGNORE INTO schema_migrations (scope, version, description, started_at, duration_ms, result, applied_by) VALUES (?, ?, ?, CURRENT_TIMESTAMP, 0, 'applied', '')`
	if isPostgresDialect() {
		insertStmt = `INSERT INTO schema_migrations (scope, version, description, started_at, duration_ms, result, applied_by) VALUES (?, ?, ?, CURRENT_TIMESTAMP, 0, 'applied', '') ON CONFLICT(scope, version) DO NOTHING`
	}
	_, err := execSQLCompat(dbConn, insertStmt, scope, version, description)
	return err
}

// ApplySchemaMigration ejecuta la migracion solo si la version no existe y la registra al finalizar.
func ApplySchemaMigration(dbConn *sql.DB, scope, version, description string, applyFn func(*sql.DB) error) error {
	if scope == "" || version == "" {
		return fmt.Errorf("scope y version son obligatorios")
	}
	return WithMigrationAdvisoryLock(dbConn, scope+":"+version, func() error {
		if err := EnsureSchemaMigrationsTable(dbConn); err != nil {
			return err
		}

		var exists int
		err := queryRowSQLCompat(dbConn, `SELECT 1 FROM schema_migrations WHERE scope = ? AND version = ? LIMIT 1`, scope, version).Scan(&exists)
		if err == nil {
			return nil
		}
		if err != sql.ErrNoRows {
			return err
		}

		started := time.Now()
		if applyFn != nil {
			if err := applyFn(dbConn); err != nil {
				return err
			}
		}

		_, err = execSQLCompat(dbConn, `INSERT INTO schema_migrations (scope, version, description, started_at, duration_ms, result, applied_by)
			VALUES (?, ?, ?, ?, ?, 'applied', ?)`, scope, version, description, started.UTC(), time.Since(started).Milliseconds(), "pcs-migrate")
		return err
	})
}

// SchemaMigrationStatus is a safe operational view used by readiness and
// release diagnostics; it intentionally never contains SQL bodies or DSNs.
type SchemaMigrationStatus struct {
	Scope      string
	Version    string
	Result     string
	DurationMS int64
}

// ListSchemaMigrationStatus exposes only release metadata for controlled
// diagnostics. It intentionally omits SQL bodies, DSNs and runtime settings.
func ListSchemaMigrationStatus(dbConn *sql.DB, limit int) ([]SchemaMigrationStatus, error) {
	if dbConn == nil {
		return nil, fmt.Errorf("database not available")
	}
	if limit < 1 || limit > 500 {
		limit = 100
	}
	rows, err := querySQLCompat(dbConn, `SELECT scope, version, COALESCE(result, 'applied'), COALESCE(duration_ms, 0)
		FROM schema_migrations ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]SchemaMigrationStatus, 0, limit)
	for rows.Next() {
		var item SchemaMigrationStatus
		if err := rows.Scan(&item.Scope, &item.Version, &item.Result, &item.DurationMS); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func IsSchemaMigrationApplied(dbConn *sql.DB, scope, version string) (bool, error) {
	if dbConn == nil || strings.TrimSpace(scope) == "" || strings.TrimSpace(version) == "" {
		return false, fmt.Errorf("migration status input invalid")
	}
	var exists bool
	err := queryRowSQLCompat(dbConn, `SELECT EXISTS (
		SELECT 1 FROM schema_migrations WHERE scope = ? AND version = ? AND COALESCE(result, 'applied') = 'applied'
	)`, scope, version).Scan(&exists)
	return exists, err
}

func VerifyRequiredMigrations(dbConn *sql.DB, required ...SchemaMigrationStatus) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	for _, migration := range required {
		applied, err := IsSchemaMigrationApplied(dbConn, migration.Scope, migration.Version)
		if err != nil {
			return err
		}
		if !applied {
			return fmt.Errorf("required migration not applied")
		}
	}
	return nil
}

// WithMigrationAdvisoryLock serializes migration work across replicas. The
// lock is PostgreSQL-wide, so a second migrate container waits instead of
// racing a DDL change or recording a version before its work is complete.
func WithMigrationAdvisoryLock(dbConn *sql.DB, name string, fn func() error) error {
	if dbConn == nil {
		return fmt.Errorf("database not available")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("migration lock name is required")
	}
	if fn == nil {
		return fmt.Errorf("migration function is required")
	}
	if !isPostgresDialect() {
		return fn()
	}
	conn, err := dbConn.Conn(context.Background())
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	key := migrationAdvisoryLockKey(name)
	if _, err := conn.ExecContext(context.Background(), `SELECT pg_advisory_lock($1)`, key); err != nil {
		return err
	}
	defer func() { _, _ = conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, key) }()
	return fn()
}

func migrationAdvisoryLockKey(name string) int64 {
	sum := sha256.Sum256([]byte(strings.TrimSpace(name)))
	return int64(binary.BigEndian.Uint64(sum[:8]) & 0x7fffffffffffffff)
}
