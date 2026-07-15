package db

import (
	"database/sql"
	"errors"
	"strings"
)

var ErrLicenciaGratisYaUsada = errors.New("licencia gratis ya usada para esta empresa")

func EnsureLicenciasGratisActivacionesSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return errors.New("dbConn is nil")
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS licencias_activaciones_gratis (
			id BIGSERIAL PRIMARY KEY,
			licencia_id BIGINT NOT NULL,
			empresa_id BIGINT NOT NULL,
			discount_code TEXT,
			asesor_id TEXT,
			motivo TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`ALTER TABLE licencias_activaciones_gratis ADD COLUMN IF NOT EXISTS asesor_id TEXT`,
		`UPDATE licencias_activaciones_gratis
			SET estado = 'historico_duplicado',
				observaciones = TRIM(COALESCE(observaciones, '') || ' normalizado_por_prueba_unica_empresa')
			WHERE COALESCE(estado, 'activo') = 'activo'
				AND id NOT IN (
					SELECT MIN(id)
					FROM licencias_activaciones_gratis
					WHERE COALESCE(estado, 'activo') = 'activo'
					GROUP BY empresa_id
				)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_licencias_gratis_licencia_empresa ON licencias_activaciones_gratis(licencia_id, empresa_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_licencias_gratis_empresa_unica ON licencias_activaciones_gratis(empresa_id) WHERE COALESCE(estado, 'activo') = 'activo'`,
		`CREATE INDEX IF NOT EXISTS ix_licencias_gratis_empresa_fecha ON licencias_activaciones_gratis(empresa_id, fecha_creacion DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_licencias_gratis_empresa_estado ON licencias_activaciones_gratis(empresa_id, estado)`,
		`CREATE INDEX IF NOT EXISTS ix_licencias_gratis_asesor ON licencias_activaciones_gratis(upper(trim(COALESCE(asesor_id, ''))))`,
	}
	for _, stmt := range statements {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	if err := EnsurePostgresPrimaryKeySequences(dbConn); err != nil {
		return err
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
		WHERE empresa_id = ?`, empresaID).Scan(&count); err != nil {
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
			AND COALESCE(valor, 0) <= 0
			AND trim(COALESCE(fecha_inicio, '')) <> ''
			AND (
				COALESCE(duracion_dias, 0) = 15
				OR LOWER(COALESCE(nombre, '')) LIKE '%prueba%'
				OR LOWER(COALESCE(nombre, '')) LIKE '%gratis%'
				OR LOWER(COALESCE(nombre, '')) LIKE '%trial%'
				OR LOWER(COALESCE(descripcion, '')) LIKE '%prueba%'
				OR LOWER(COALESCE(descripcion, '')) LIKE '%gratis%'
				OR LOWER(COALESCE(descripcion, '')) LIKE '%trial%'
				OR LOWER(COALESCE(codigo_funcion, '')) LIKE '%trial%'
			)`, empresaID).Scan(&count); err != nil {
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

func HasLicenciaAdvisorCodeUsedByEmpresa(dbConn *sql.DB, empresaID int64) (bool, error) {
	if empresaID <= 0 {
		return false, nil
	}
	if err := EnsurePaymentGatewaySchema(dbConn); err != nil {
		return false, err
	}

	for _, tableName := range []string{"pagos_epayco", "pagos_wompi"} {
		rows, err := querySQLCompat(dbConn, "SELECT COALESCE(status, '') FROM "+tableName+" WHERE empresa_id = ? AND trim(COALESCE(asesor_id, '')) <> ''", empresaID)
		if err != nil {
			if isMissingTableError(err) || isMissingColumnError(err) {
				continue
			}
			return false, err
		}
		for rows.Next() {
			var status string
			if err := rows.Scan(&status); err != nil {
				_ = rows.Close()
				return false, err
			}
			if isApprovedLicenciaPaymentStatus(status) {
				_ = rows.Close()
				return true, nil
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return false, err
		}
		_ = rows.Close()
	}

	if err := EnsureLicenciasGratisActivacionesSchema(dbConn); err != nil {
		return false, err
	}
	var count int
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1)
		FROM licencias_activaciones_gratis
		WHERE empresa_id = ?
			AND trim(COALESCE(asesor_id, '')) <> ''
			AND COALESCE(estado, 'activo') = 'activo'`, empresaID).Scan(&count); err != nil {
		if isMissingTableError(err) || isMissingColumnError(err) {
			return false, nil
		}
		return false, err
	}
	if count > 0 {
		return true, nil
	}

	if err := EnsureAsesorComercialSchema(dbConn); err != nil {
		return false, err
	}
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1)
		FROM asesor_comercial_comisiones
		WHERE empresa_id = ?
			AND trim(COALESCE(asesor_codigo, '')) <> ''
			AND COALESCE(estado, 'activo') = 'activo'`, empresaID).Scan(&count); err != nil {
		if isMissingTableError(err) || isMissingColumnError(err) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func hasLicenciaDiscountCodeUsedByEmpresaExceptPayment(dbConn *sql.DB, empresaID int64, discountCode, provider, transactionID, reference string) (bool, error) {
	code := strings.ToUpper(strings.TrimSpace(discountCode))
	if empresaID <= 0 || code == "" {
		return false, nil
	}
	if err := EnsurePaymentGatewaySchema(dbConn); err != nil {
		return false, err
	}

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
		rows, err := querySQLCompat(dbConn, "SELECT COALESCE(status, ''), COALESCE(transaction_id, ''), COALESCE(reference, '') FROM "+tableName+" WHERE empresa_id = ? AND upper(trim(COALESCE(discount_code, ''))) = ?", empresaID, code)
		if err != nil {
			if isMissingTableError(err) || isMissingColumnError(err) {
				continue
			}
			return false, err
		}
		for rows.Next() {
			var status, rowTransactionID, rowReference string
			if err := rows.Scan(&status, &rowTransactionID, &rowReference); err != nil {
				_ = rows.Close()
				return false, err
			}
			if tableName == excludeTable && licenciaPaymentRecordMatchesExclusion(rowTransactionID, rowReference, transactionID, reference) {
				continue
			}
			if isApprovedLicenciaPaymentStatus(status) {
				_ = rows.Close()
				return true, nil
			}
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return false, err
		}
		_ = rows.Close()
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

func licenciaPaymentRecordMatchesExclusion(rowTransactionID, rowReference, transactionID, reference string) bool {
	rowTransactionID = strings.TrimSpace(rowTransactionID)
	rowReference = strings.TrimSpace(rowReference)
	transactionID = strings.TrimSpace(transactionID)
	reference = strings.TrimSpace(reference)
	return (transactionID != "" && rowTransactionID == transactionID) || (reference != "" && rowReference == reference)
}

func isApprovedLicenciaPaymentStatus(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved", "accepted", "accredited", "aprobado", "aprobada", "aceptado", "aceptada", "acreditado", "acreditada", "success", "successful", "ok", "1", "manual":
		return true
	default:
		return false
	}
}

func ActivateLicenciaGratisForEmpresa(dbConn *sql.DB, licenciaID, empresaID int64, fechaInicio, fechaFin, discountCode, motivo, asesorID string) error {
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
	code := strings.ToUpper(strings.TrimSpace(discountCode))
	licenciaValor := 0.0
	if err := queryRowTxSQLCompat(tx, `SELECT COALESCE(valor, 0) FROM licencias WHERE id = ?`, licenciaID).Scan(&licenciaValor); err != nil {
		if !errors.Is(err, sql.ErrNoRows) && !isMissingTableError(err) && !isMissingColumnError(err) {
			return err
		}
	}
	isZeroValueLicense := licenciaValor <= 0
	if code != "" && !isZeroValueLicense {
		if err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)
			FROM licencias_activaciones_gratis
			WHERE empresa_id = ?
				AND upper(trim(COALESCE(discount_code, ''))) = ?
				AND COALESCE(estado, 'activo') = 'activo'`, empresaID, code).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return ErrLicenciaGratisYaUsada
		}
	} else {
		if err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)
			FROM licencias_activaciones_gratis
			WHERE empresa_id = ?`, empresaID).Scan(&count); err != nil {
			return err
		}
		if count > 0 {
			return ErrLicenciaGratisYaUsada
		}
		if err := queryRowTxSQLCompat(tx, `SELECT COUNT(1)
			FROM licencias
			WHERE empresa_id = ?
				AND COALESCE(valor, 0) <= 0
				AND trim(COALESCE(fecha_inicio, '')) <> ''
				AND (
					COALESCE(duracion_dias, 0) = 15
					OR LOWER(COALESCE(nombre, '')) LIKE '%prueba%'
					OR LOWER(COALESCE(nombre, '')) LIKE '%gratis%'
					OR LOWER(COALESCE(nombre, '')) LIKE '%trial%'
					OR LOWER(COALESCE(descripcion, '')) LIKE '%prueba%'
					OR LOWER(COALESCE(descripcion, '')) LIKE '%gratis%'
					OR LOWER(COALESCE(descripcion, '')) LIKE '%trial%'
					OR LOWER(COALESCE(codigo_funcion, '')) LIKE '%trial%'
				)`, empresaID).Scan(&count); err != nil {
			if !isMissingTableError(err) && !isMissingColumnError(err) {
				return err
			}
			count = 0
		}
		if count > 0 {
			return ErrLicenciaGratisYaUsada
		}
	}

	nowExpr := sqlNowExpr()
	asesorID = strings.ToUpper(strings.TrimSpace(asesorID))
	assignedLicenciaID, err := activateLicenciaForEmpresaTx(tx, licenciaID, empresaID, fechaInicio, fechaFin)
	if err != nil {
		return err
	}
	if assignedLicenciaID <= 0 {
		assignedLicenciaID = licenciaID
	}

	if _, err := execTxSQLCompat(tx, "INSERT INTO licencias_activaciones_gratis (licencia_id, empresa_id, discount_code, asesor_id, motivo, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, ?, ?, "+nowExpr+", "+nowExpr+", 'activo')", assignedLicenciaID, empresaID, strings.TrimSpace(discountCode), asesorID, strings.TrimSpace(motivo)); err != nil {
		if isLicenciaGratisUniqueConstraintErr(err) {
			return ErrLicenciaGratisYaUsada
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
	return nil
}

func isLicenciaGratisUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "unique constraint failed") || strings.Contains(lower, "duplicate key value violates unique constraint")
}
