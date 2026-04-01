package db

import (
	"database/sql"
	"fmt"
)

// EnsureSchemaMigrationsTable crea la tabla de control de migraciones versionadas.
func EnsureSchemaMigrationsTable(dbConn *sql.DB) error {
	_, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		scope TEXT NOT NULL,
		version TEXT NOT NULL,
		description TEXT,
		applied_at TEXT DEFAULT (datetime('now','localtime')),
		UNIQUE(scope, version)
	);`)
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
	_, err := dbConn.Exec(`INSERT OR IGNORE INTO schema_migrations (scope, version, description) VALUES (?, ?, ?)`, scope, version, description)
	return err
}

// ApplySchemaMigration ejecuta la migracion solo si la version no existe y la registra al finalizar.
func ApplySchemaMigration(dbConn *sql.DB, scope, version, description string, applyFn func(*sql.DB) error) error {
	if scope == "" || version == "" {
		return fmt.Errorf("scope y version son obligatorios")
	}
	if err := EnsureSchemaMigrationsTable(dbConn); err != nil {
		return err
	}

	var exists int
	err := dbConn.QueryRow(`SELECT 1 FROM schema_migrations WHERE scope = ? AND version = ? LIMIT 1`, scope, version).Scan(&exists)
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

	_, err = dbConn.Exec(`INSERT INTO schema_migrations (scope, version, description) VALUES (?, ?, ?)`, scope, version, description)
	return err
}
