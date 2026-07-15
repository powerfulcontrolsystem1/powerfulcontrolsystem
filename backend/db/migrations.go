package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"strings"
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
	_, err := execSQLCompat(dbConn, createStmt)
	return err
}

// RegisterSchemaMigration registra una version de esquema de forma idempotente.
func RegisterSchemaMigration(dbConn *sql.DB, scope, version, description string) error {
	if scope == "" || version == "" {
		return fmt.Errorf("scope y version son obligatorios")
	}
	if err := EnsureSchemaMigrationsTable(dbConn); err != nil {
		return err
	}
	insertStmt := `INSERT OR IGNORE INTO schema_migrations (scope, version, description) VALUES (?, ?, ?)`
	if isPostgresDialect() {
		insertStmt = `INSERT INTO schema_migrations (scope, version, description) VALUES (?, ?, ?) ON CONFLICT(scope, version) DO NOTHING`
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

		if applyFn != nil {
			if err := applyFn(dbConn); err != nil {
				return err
			}
		}

		_, err = execSQLCompat(dbConn, `INSERT INTO schema_migrations (scope, version, description) VALUES (?, ?, ?)`, scope, version, description)
		return err
	})
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
