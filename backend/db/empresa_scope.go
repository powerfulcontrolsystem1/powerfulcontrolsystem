package db

import (
	"database/sql"
)

// EnsureEmpresasScopeReferences asegura que las tablas base de empresas.db
// tengan una referencia de alcance por empresa donde aplique.
func EnsureEmpresasScopeReferences(dbConn *sql.DB) error {
	if err := ensureColumnIfMissing(dbConn, "empresas", "empresa_id", "INTEGER"); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, "UPDATE empresas SET empresa_id = id WHERE empresa_id IS NULL OR empresa_id <= 0"); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, "CREATE UNIQUE INDEX IF NOT EXISTS ux_empresas_empresa_id ON empresas(empresa_id)"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "schema_migrations", "empresa_id", "INTEGER"); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, "UPDATE schema_migrations SET empresa_id = 0 WHERE empresa_id IS NULL"); err != nil {
		return err
	}

	if hasTipos, err := tableExists(dbConn, "tipos_de_empresas"); err != nil {
		return err
	} else if hasTipos {
		if err := ensureColumnIfMissing(dbConn, "tipos_de_empresas", "empresa_id", "INTEGER"); err != nil {
			return err
		}
		if _, err := execSQLCompat(dbConn, "UPDATE tipos_de_empresas SET empresa_id = 0 WHERE empresa_id IS NULL"); err != nil {
			return err
		}
	}

	return nil
}

func tableExists(dbConn *sql.DB, tableName string) (bool, error) {
	var exists bool
	err := queryRowSQLCompat(dbConn, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = ANY (current_schemas(false))
			  AND table_name = ?
		)
	`, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
