package db

import (
	"database/sql"
	"log"
	"strconv"
	"time"
)

// LicenciaEmpresaEstadoSyncResult resume la sincronizacion entre licencias y estado empresarial.
type LicenciaEmpresaEstadoSyncResult struct {
	LicenciasVencidas         int64 `json:"licencias_vencidas"`
	LicenciasLimiteDocumentos int64 `json:"licencias_limite_documentos"`
	EmpresasDesactivadas      int64 `json:"empresas_desactivadas"`
	EmpresasEvaluadas         int64 `json:"empresas_evaluadas"`
}

func licenciaActivePredicate(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	if isPostgresDialect() {
		return "COALESCE(" + prefix + "activo, 1) = 1 AND " + postgresLicenciaDatePredicate(prefix+"fecha_inicio", "<=") + " AND " + postgresLicenciaDatePredicate(prefix+"fecha_fin", ">=")
	}
	return "COALESCE(" + prefix + "activo, 1) = 1 AND (COALESCE(" + prefix + "fecha_inicio, '') = '' OR pcs_ts(" + prefix + "fecha_inicio) <= CURRENT_TIMESTAMP) AND (COALESCE(" + prefix + "fecha_fin, '') = '' OR pcs_ts(" + prefix + "fecha_fin) >= CURRENT_TIMESTAMP)"
}

func licenciaExpiredPredicate(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	if isPostgresDialect() {
		return postgresLicenciaHasExpiredPredicate(prefix + "fecha_fin")
	}
	return "COALESCE(" + prefix + "fecha_fin, '') <> '' AND pcs_ts(" + prefix + "fecha_fin) < CURRENT_TIMESTAMP"
}

func queryLicenciaEmpresaIDSet(dbConn *sql.DB, query string, args ...interface{}) (map[int64]struct{}, error) {
	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64]struct{})
	for rows.Next() {
		var id sql.NullInt64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if id.Valid && id.Int64 > 0 {
			out[id.Int64] = struct{}{}
		}
	}
	return out, rows.Err()
}

type licenciaDocumentoLimiteCandidate struct {
	ID        int64
	EmpresaID int64
	Limite    int64
}

func monthBoundsForLicenciaDocumentos(now time.Time) (string, string) {
	if now.IsZero() {
		now = time.Now()
	}
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, 0)
	return start.Format("2006-01-02 15:04:05"), end.Format("2006-01-02 15:04:05")
}

func queryMonthlyCount(dbConn *sql.DB, query string, args ...interface{}) (int64, error) {
	var count sql.NullInt64
	if err := queryRowSQLCompat(dbConn, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	if count.Valid {
		return count.Int64, nil
	}
	return 0, nil
}

func countEmpresaVentasMes(dbConn *sql.DB, empresaID int64, desde, hasta string) (int64, error) {
	hasCarritos, err := tableExists(dbConn, "carritos_compras")
	if err != nil || !hasCarritos {
		return 0, err
	}
	return queryMonthlyCount(dbConn, `SELECT COALESCE(COUNT(1), 0)
		FROM carritos_compras
		WHERE empresa_id = ?
		  AND LOWER(COALESCE(estado_carrito, '')) IN ('cerrado', 'pagado', 'finalizado')
		  AND LOWER(COALESCE(estado, 'activo')) <> 'anulado'
		  AND COALESCE(NULLIF(pagado_en, ''), NULLIF(fecha_actualizacion, ''), NULLIF(fecha_creacion, ''), '') >= ?
		  AND COALESCE(NULLIF(pagado_en, ''), NULLIF(fecha_actualizacion, ''), NULLIF(fecha_creacion, ''), '') < ?`, empresaID, desde, hasta)
}

func countEmpresaFacturasElectronicasMes(dbConn *sql.DB, empresaID int64, desde, hasta string) (int64, error) {
	hasFE, err := tableExists(dbConn, "facturacion_electronica_reintentos")
	if err != nil || !hasFE {
		return 0, err
	}
	return queryMonthlyCount(dbConn, `SELECT COALESCE(COUNT(DISTINCT COALESCE(NULLIF(tipo_documento, '') || ':' || NULLIF(documento_codigo, ''), CAST(id AS TEXT))), 0)
		FROM facturacion_electronica_reintentos
		WHERE empresa_id = ?
		  AND LOWER(COALESCE(estado, 'activo')) <> 'anulado'
		  AND LOWER(COALESCE(estado_envio, '')) NOT IN ('anulado', 'cancelado')
		  AND COALESCE(NULLIF(fecha_emision_legal, ''), NULLIF(fecha_creacion, ''), '') >= ?
		  AND COALESCE(NULLIF(fecha_emision_legal, ''), NULLIF(fecha_creacion, ''), '') < ?`, empresaID, desde, hasta)
}

func countEmpresaAIUFacturasMes(dbConn *sql.DB, empresaID int64, desde, hasta string) (int64, error) {
	hasAIU, err := tableExists(dbConn, "empresa_aiu_facturas")
	if err != nil || !hasAIU {
		return 0, err
	}
	return queryMonthlyCount(dbConn, `SELECT COALESCE(COUNT(1), 0)
		FROM empresa_aiu_facturas
		WHERE empresa_id = ?
		  AND LOWER(COALESCE(estado, 'activo')) NOT IN ('anulado', 'cancelado')
		  AND COALESCE(NULLIF(fecha_documento, ''), NULLIF(fecha_creacion, ''), '') >= ?
		  AND COALESCE(NULLIF(fecha_documento, ''), NULLIF(fecha_creacion, ''), '') < ?`, empresaID, desde, hasta)
}

// CountEmpresaDocumentosLicenciaMes calcula el uso mensual que consume el limite de licencia.
// Se toma el mayor entre ventas cerradas y documentos electronicos para evitar doble conteo
// cuando una venta genera una factura electronica sobre el mismo evento comercial.
func CountEmpresaDocumentosLicenciaMes(dbConn *sql.DB, empresaID int64, now time.Time) (int64, error) {
	if dbConn == nil || empresaID <= 0 {
		return 0, nil
	}
	desde, hasta := monthBoundsForLicenciaDocumentos(now)
	ventas, err := countEmpresaVentasMes(dbConn, empresaID, desde, hasta)
	if err != nil {
		return 0, err
	}
	feDocs, err := countEmpresaFacturasElectronicasMes(dbConn, empresaID, desde, hasta)
	if err != nil {
		return 0, err
	}
	aiuDocs, err := countEmpresaAIUFacturasMes(dbConn, empresaID, desde, hasta)
	if err != nil {
		return 0, err
	}
	documentos := feDocs + aiuDocs
	if documentos > ventas {
		return documentos, nil
	}
	return ventas, nil
}

func listLicenciasConLimiteDocumentos(dbConn *sql.DB, empresaID int64) ([]licenciaDocumentoLimiteCandidate, error) {
	query := "SELECT id, empresa_id, COALESCE(max_documentos_mensuales, 0) FROM licencias WHERE empresa_id IS NOT NULL AND empresa_id > 0 AND COALESCE(max_documentos_mensuales, 0) > 0 AND " + licenciaActivePredicate("")
	args := []interface{}{}
	if empresaID > 0 {
		query += " AND empresa_id = ?"
		args = append(args, empresaID)
	}
	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]licenciaDocumentoLimiteCandidate, 0)
	for rows.Next() {
		var item licenciaDocumentoLimiteCandidate
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.Limite); err != nil {
			return nil, err
		}
		if item.ID > 0 && item.EmpresaID > 0 && item.Limite > 0 {
			out = append(out, item)
		}
	}
	return out, rows.Err()
}

func deactivateLicenciaPorLimiteDocumentos(dbConn *sql.DB, licenciaID, uso, limite int64) (int64, error) {
	nota := "Desactivada automaticamente por superar el limite mensual de documentos/ventas: " + strconv.FormatInt(uso, 10) + "/" + strconv.FormatInt(limite, 10)
	query := "UPDATE licencias SET activo = 0, estado = 'limite_documentos_mensuales', observaciones = TRIM(COALESCE(observaciones, '') || CASE WHEN COALESCE(observaciones, '') = '' THEN '' ELSE ' | ' END || ?), fecha_actualizacion = " + sqlNowExpr() + " WHERE id = ? AND COALESCE(activo, 1) = 1"
	res, err := execSQLCompat(dbConn, query, nota, licenciaID)
	if err != nil {
		return 0, err
	}
	if res == nil {
		return 0, nil
	}
	n, _ := res.RowsAffected()
	if n > 0 {
		invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn, licenciaID)
	}
	return n, nil
}

func enforceLicenciasDocumentosMensuales(dbEmp, dbSuper *sql.DB, empresaID int64) (int64, error) {
	candidates, err := listLicenciasConLimiteDocumentos(dbSuper, empresaID)
	if err != nil {
		return 0, err
	}
	var desactivadas int64
	now := time.Now()
	for _, candidate := range candidates {
		uso, err := CountEmpresaDocumentosLicenciaMes(dbEmp, candidate.EmpresaID, now)
		if err != nil {
			return desactivadas, err
		}
		if uso <= candidate.Limite {
			continue
		}
		n, err := deactivateLicenciaPorLimiteDocumentos(dbSuper, candidate.ID, uso, candidate.Limite)
		if err != nil {
			return desactivadas, err
		}
		desactivadas += n
	}
	return desactivadas, nil
}

// EnforceLicenciaDocumentosMensualesPorEmpresa aplica inmediatamente el limite mensual de una empresa.
func EnforceLicenciaDocumentosMensualesPorEmpresa(dbEmp, dbSuper *sql.DB, empresaID int64) (int64, error) {
	if dbEmp == nil || dbSuper == nil || empresaID <= 0 {
		return 0, nil
	}
	if err := EnsureLicenciasSchema(dbSuper); err != nil {
		return 0, err
	}
	desactivadas, err := enforceLicenciasDocumentosMensuales(dbEmp, dbSuper, empresaID)
	if err != nil || desactivadas == 0 {
		return desactivadas, err
	}
	conLicenciaVigente, err := queryLicenciaEmpresaIDSet(dbSuper, "SELECT DISTINCT empresa_id FROM licencias WHERE empresa_id = ? AND "+licenciaActivePredicate(""), empresaID)
	if err != nil {
		return desactivadas, err
	}
	if _, ok := conLicenciaVigente[empresaID]; ok {
		return desactivadas, nil
	}
	updateQuery := "UPDATE empresas SET estado = 'inactivo', fecha_actualizacion = " + sqlNowExpr() + " WHERE (id = ? OR COALESCE(empresa_id, id) = ?) AND LOWER(COALESCE(estado, 'activo')) <> 'inactivo'"
	if _, err := execSQLCompat(dbEmp, updateQuery, empresaID, empresaID); err != nil {
		return desactivadas, err
	}
	return desactivadas, nil
}

// SyncEmpresasEstadoPorLicencia desactiva licencias vencidas y empresas que ya no tienen licencia vigente.
func SyncEmpresasEstadoPorLicencia(dbEmp, dbSuper *sql.DB) (LicenciaEmpresaEstadoSyncResult, error) {
	var result LicenciaEmpresaEstadoSyncResult
	if dbEmp == nil || dbSuper == nil {
		return result, nil
	}
	if err := EnsureLicenciasSchema(dbSuper); err != nil {
		return result, err
	}

	nowExpr := sqlNowExpr()
	expireQuery := "UPDATE licencias SET activo = 0, estado = 'vencida', fecha_actualizacion = " + nowExpr + " WHERE COALESCE(activo, 1) = 1 AND " + licenciaExpiredPredicate("")
	if res, err := execSQLCompat(dbSuper, expireQuery); err != nil {
		return result, err
	} else if res != nil {
		if n, nerr := res.RowsAffected(); nerr == nil {
			result.LicenciasVencidas = n
		}
	}

	porLimite, err := enforceLicenciasDocumentosMensuales(dbEmp, dbSuper, 0)
	if err != nil {
		return result, err
	}
	result.LicenciasLimiteDocumentos = porLimite

	conLicencia, err := queryLicenciaEmpresaIDSet(dbSuper, "SELECT DISTINCT empresa_id FROM licencias WHERE empresa_id IS NOT NULL AND empresa_id > 0")
	if err != nil {
		return result, err
	}
	conLicenciaVigente, err := queryLicenciaEmpresaIDSet(dbSuper, "SELECT DISTINCT empresa_id FROM licencias WHERE empresa_id IS NOT NULL AND empresa_id > 0 AND "+licenciaActivePredicate(""))
	if err != nil {
		return result, err
	}

	updateQuery := "UPDATE empresas SET estado = 'inactivo', fecha_actualizacion = " + nowExpr + " WHERE (id = ? OR COALESCE(empresa_id, id) = ?) AND LOWER(COALESCE(estado, 'activo')) <> 'inactivo'"
	for empresaID := range conLicencia {
		result.EmpresasEvaluadas++
		if _, ok := conLicenciaVigente[empresaID]; ok {
			continue
		}
		res, err := execSQLCompat(dbEmp, updateQuery, empresaID, empresaID)
		if err != nil {
			return result, err
		}
		if res != nil {
			if n, nerr := res.RowsAffected(); nerr == nil {
				result.EmpresasDesactivadas += n
			}
		}
	}
	return result, nil
}

// StartLicenciaEmpresaEstadoWorker ejecuta la sincronizacion de vencimientos de forma periodica.
func StartLicenciaEmpresaEstadoWorker(dbEmp, dbSuper *sql.DB, interval time.Duration, stop <-chan struct{}) {
	if dbEmp == nil || dbSuper == nil {
		return
	}
	if interval <= 0 {
		interval = time.Hour
	}
	run := func(origin string) {
		result, err := SyncEmpresasEstadoPorLicencia(dbEmp, dbSuper)
		if err != nil {
			log.Printf("[licencias] sincronizacion estado empresas (%s) error: %v", origin, err)
			return
		}
		if result.LicenciasVencidas > 0 || result.LicenciasLimiteDocumentos > 0 || result.EmpresasDesactivadas > 0 {
			log.Printf("[licencias] sincronizacion estado empresas (%s): licencias_vencidas=%d licencias_limite_documentos=%d empresas_desactivadas=%d evaluadas=%d", origin, result.LicenciasVencidas, result.LicenciasLimiteDocumentos, result.EmpresasDesactivadas, result.EmpresasEvaluadas)
		}
	}
	run("startup")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			run("periodic")
		case <-stop:
			return
		}
	}
}
