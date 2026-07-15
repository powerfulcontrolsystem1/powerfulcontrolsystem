package db

import (
	"database/sql"
	"errors"
)

// EnsurePortalVisitasSchema belongs to the migration role. Public traffic may
// increment a counter but must not perform DDL on the API process.
func EnsurePortalVisitasSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return errors.New("database not available")
	}
	_, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS portal_visitas_paises (
		pais_codigo TEXT NOT NULL,
		fecha DATE NOT NULL DEFAULT CURRENT_DATE,
		visitas BIGINT NOT NULL DEFAULT 0,
		actualizado_en TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		PRIMARY KEY (pais_codigo, fecha)
	)`)
	return err
}

func VerifyPortalVisitasSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("database not available")
	}
	var exists int
	err := queryRowSQLCompat(dbConn, `SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = ? LIMIT 1`, "portal_visitas_paises").Scan(&exists)
	if err == sql.ErrNoRows {
		return errors.New("contador de visitas no migrado")
	}
	return err
}
