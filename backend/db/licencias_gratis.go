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
		`CREATE INDEX IF NOT EXISTS ix_licencias_gratis_empresa_estado ON licencias_activaciones_gratis(empresa_id, estado)`,
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

func HasAnyLicenciaGratisActivationForEmpresa(dbConn *sql.DB, empresaID int64) (bool, error) {
	if empresaID <= 0 {
		return false, nil
	}
	if err := EnsureLicenciasGratisActivacionesSchema(dbConn); err != nil {
		return false, err
	}
	var count int
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1)
		FROM licencias_activaciones_gratis
		WHERE empresa_id = ?
			AND COALESCE(estado, 'activo') = 'activo'`, empresaID).Scan(&count); err != nil {
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	if err := EnsureLicenciasSchema(dbConn); err != nil {
		return false, err
	}
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1)
		FROM licencias
		WHERE empresa_id = ?
			AND COALESCE(activo, 1) = 1
			AND COALESCE(valor, 0) <= 0
			AND trim(COALESCE(fecha_inicio, '')) <> ''`, empresaID).Scan(&count); err != nil {
		if isMissingTableError(err) || isMissingColumnError(err) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func HasLicenciaDiscountCodeUsedByEmpresa(dbConn *sql.DB, empresaID int64, discountCode string) (bool, error) {
	return hasLicenciaDiscountCodeUsedByEmpresaExceptPayment(dbConn, empresaID, discountCode, "", "", "")
}

func HasLicenciaDiscountCodeUsedByEmpresaExceptPayment(dbConn *sql.DB, empresaID int64, discountCode, provider, transactionID, reference string) (bool, error) {
	return hasLicenciaDiscountCodeUsedByEmpresaExceptPayment(dbConn, empresaID, discountCode, provider, transactionID, reference)
}

func hasLicenciaDiscountCodeUsedByEmpresaExceptPayment(dbConn *sql.DB, empresaID int64, discountCode, provider, transactionID, reference string) (bool, error) {
	code := strings.ToUpper(strings.TrimSpace(discountCode))
	if empresaID <= 0 || code == "" {
		return false, nil
	}
	if err := EnsurePaymentGatewaySchema(dbConn); err != nil {
		return false, err
	}

	approvedStatuses := []string{"approved", "accredited", "manual"}
	excludeTable := ""
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "epayco":
		excludeTable = "pagos_epayco"
	case "wompi":
		excludeTable = "pagos_wompi"
	}
	transactionID = strings.TrimSpace(transactionID)
	reference = strings.TrimSpace(reference)
	for _, tableName := range []string{"pagos_epayco", "pagos_wompi"} {
		var count int
		query := "SELECT COUNT(1) FROM " + tableName + " WHERE empresa_id = ? AND upper(trim(COALESCE(discount_code, ''))) = ? AND lower(trim(COALESCE(status, ''))) IN (?, ?, ?)"
		args := []interface{}{empresaID, code, approvedStatuses[0], approvedStatuses[1], approvedStatuses[2]}
		if tableName == excludeTable {
			query += " AND NOT ((? <> '' AND COALESCE(transaction_id, '') = ?) OR (? <> '' AND COALESCE(reference, '') = ?))"
			args = append(args, transactionID, transactionID, reference, reference)
		}
		if err := queryRowSQLCompat(dbConn, query, args...).Scan(&count); err != nil {
			if isMissingTableError(err) || isMissingColumnError(err) {
				continue
			}
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}

	var activationCount int
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1)
		FROM licencias_activaciones_gratis
		WHERE empresa_id = ?
			AND upper(trim(COALESCE(discount_code, ''))) = ?
			AND COALESCE(estado, 'activo') = 'activo'`, empresaID, code).Scan(&activationCount); err != nil {
		if isMissingTableError(err) || isMissingColumnError(err) {
			return false, nil
		}
		return false, err
	}
	return activationCount > 0, nil
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
	if err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)
		FROM licencias_activaciones_gratis
		WHERE empresa_id = ?
			AND COALESCE(estado, 'activo') = 'activo'`, empresaID).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return ErrLicenciaGratisYaUsada
	}
	if err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)
		FROM licencias
		WHERE empresa_id = ?
			AND COALESCE(activo, 1) = 1
			AND COALESCE(valor, 0) <= 0
			AND trim(COALESCE(fecha_inicio, '')) <> ''`, empresaID).Scan(&count); err != nil {
		if !isMissingTableError(err) && !isMissingColumnError(err) {
			return err
		}
		count = 0
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
