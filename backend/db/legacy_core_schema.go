package db

import (
	"database/sql"
	"fmt"
)

// BootstrapLegacyCoreEmpresasSchema creates only the root table required by the
// reviewed legacy migration catalog. It is called exclusively by the migrate
// role when an empty PostgreSQL database explicitly enables the legacy
// bootstrap. API and worker roles remain protected by runtime_schema_guard.
func BootstrapLegacyCoreEmpresasSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("empresas migration database is required")
	}
	_, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS empresas (
		id BIGSERIAL PRIMARY KEY,
		empresa_id BIGINT,
		tipo_id BIGINT DEFAULT 0,
		tipo_nombre TEXT,
		nombre TEXT NOT NULL DEFAULT '',
		nit TEXT,
		fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
		fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`)
	return err
}

// BootstrapLegacyCoreSuperSchema creates the two roots that older administrative
// migrations extend before the rest of the catalog is applied.
func BootstrapLegacyCoreSuperSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("superadministrador migration database is required")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS administradores (
			id BIGSERIAL PRIMARY KEY,
			email TEXT UNIQUE,
			name TEXT,
			role TEXT DEFAULT 'administrador',
			photo TEXT,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS sesiones (
			id BIGSERIAL PRIMARY KEY,
			admin_email TEXT,
			token TEXT,
			token_hash VARCHAR(64),
			ip TEXT,
			user_agent TEXT,
			fecha_inicio TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_fin TEXT,
			activo INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)
		)`,
		`CREATE TABLE IF NOT EXISTS configuraciones (
			id BIGSERIAL PRIMARY KEY,
			config_key TEXT NOT NULL UNIQUE,
			value TEXT,
			encrypted INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)
		)`,
		`CREATE TABLE IF NOT EXISTS tipos_de_empresas (
			id BIGSERIAL PRIMARY KEY,
			nombre TEXT NOT NULL UNIQUE,
			observaciones TEXT,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo'
		)`,
	}
	for _, statement := range statements {
		if _, err := execSQLCompat(dbConn, statement); err != nil {
			return err
		}
	}
	return nil
}
