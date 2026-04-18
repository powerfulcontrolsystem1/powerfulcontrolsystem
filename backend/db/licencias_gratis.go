package db

import (
	"database/sql"
	"errors"
	"strings"
)

var ErrLicenciaGratisYaUsada = errors.New("licencia gratis ya usada para esta empresa")

func EnsureLicenciasGratisActivacionesSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("dbConn is nil")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS licencias_activaciones_gratis (
			id INTEGER PRIMARY KEY,
			licencia_id BIGINT NOT NULL,
			empresa_id BIGINT NOT NULL,
			discount_code TEXT,
			motivo TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_licencias_gratis_licencia_empresa ON licencias_activaciones_gratis(licencia_id, empresa_id)`,
		`CREATE INDEX IF NOT EXISTS ix_licencias_gratis_empresa_fecha ON licencias_activaciones_gratis(empresa_id, fecha_creacion DESC)`,
	}
	for _, stmt := range statements {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func HasLicenciaGratisActivation(dbConn *sql.DB, licenciaID, empresaID int64) (bool, error) {
	if licenciaID <= 0 || empresaID <= 0 {
		return false, nil
	}
	if err := EnsureLicenciasGratisActivacionesSchema(dbConn); err != nil {
		return false, err
	}
	var count int
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1) FROM licencias_activaciones_gratis WHERE licencia_id = ? AND empresa_id = ?`, licenciaID, empresaID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func ActivateLicenciaGratisForEmpresa(dbConn *sql.DB, licenciaID, empresaID int64, fechaInicio, fechaFin, discountCode, motivo string) error {
	if licenciaID <= 0 || empresaID <= 0 {
		return errors.New("licencia_id y empresa_id son obligatorios")
	}
	if err := EnsureLicenciasGratisActivacionesSchema(dbConn); err != nil {
		return err
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	if err := queryRowTxSQLCompat(tx, `SELECT COUNT(1) FROM licencias_activaciones_gratis WHERE licencia_id = ? AND empresa_id = ?`, licenciaID, empresaID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrLicenciaGratisYaUsada
	}

	nowExpr := sqlNowExpr()
	if _, err := execTxSQLCompat(tx, "UPDATE licencias SET empresa_id = ?, activo = 1, fecha_inicio = ?, fecha_fin = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", empresaID, fechaInicio, fechaFin, licenciaID); err != nil {
		if _, fallbackErr := execTxSQLCompat(tx, "UPDATE licencias SET empresa_id = ?, activo = 1, fecha_inicio = ?, fecha_fin = ? WHERE id = ?", empresaID, fechaInicio, fechaFin, licenciaID); fallbackErr != nil {
			return fallbackErr
		}
	}

	if _, err := execTxSQLCompat(tx, "INSERT INTO licencias_activaciones_gratis (licencia_id, empresa_id, discount_code, motivo, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, ?, "+nowExpr+", "+nowExpr+", 'activo')", licenciaID, empresaID, strings.TrimSpace(discountCode), strings.TrimSpace(motivo)); err != nil {
		if isLicenciaGratisUniqueConstraintErr(err) {
			return ErrLicenciaGratisYaUsada
		}
		return err
	}

	return tx.Commit()
}

func isLicenciaGratisUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "unique constraint failed") || strings.Contains(lower, "duplicate key value violates unique constraint")
}