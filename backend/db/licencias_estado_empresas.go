package db

import (
	"database/sql"
	"log"
	"time"
)

// LicenciaEmpresaEstadoSyncResult resume la sincronizacion entre licencias y estado empresarial.
type LicenciaEmpresaEstadoSyncResult struct {
	LicenciasVencidas    int64 `json:"licencias_vencidas"`
	EmpresasDesactivadas int64 `json:"empresas_desactivadas"`
	EmpresasEvaluadas    int64 `json:"empresas_evaluadas"`
}

func licenciaActivePredicate(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	if isPostgresDialect() {
		return "COALESCE(" + prefix + "activo, 1) = 1 AND (COALESCE(CAST(" + prefix + "fecha_inicio AS TEXT), '') = '' OR CAST(" + prefix + "fecha_inicio AS TIMESTAMP) <= CURRENT_TIMESTAMP) AND (COALESCE(CAST(" + prefix + "fecha_fin AS TEXT), '') = '' OR CAST(" + prefix + "fecha_fin AS TIMESTAMP) >= CURRENT_TIMESTAMP)"
	}
	return "COALESCE(" + prefix + "activo, 1) = 1 AND (COALESCE(" + prefix + "fecha_inicio, '') = '' OR datetime(" + prefix + "fecha_inicio) <= datetime('now','localtime')) AND (COALESCE(" + prefix + "fecha_fin, '') = '' OR datetime(" + prefix + "fecha_fin) >= datetime('now','localtime'))"
}

func licenciaExpiredPredicate(alias string) string {
	prefix := ""
	if alias != "" {
		prefix = alias + "."
	}
	if isPostgresDialect() {
		return "COALESCE(CAST(" + prefix + "fecha_fin AS TEXT), '') <> '' AND CAST(" + prefix + "fecha_fin AS TIMESTAMP) < CURRENT_TIMESTAMP"
	}
	return "COALESCE(" + prefix + "fecha_fin, '') <> '' AND datetime(" + prefix + "fecha_fin) < datetime('now','localtime')"
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
		if result.LicenciasVencidas > 0 || result.EmpresasDesactivadas > 0 {
			log.Printf("[licencias] sincronizacion estado empresas (%s): licencias_vencidas=%d empresas_desactivadas=%d evaluadas=%d", origin, result.LicenciasVencidas, result.EmpresasDesactivadas, result.EmpresasEvaluadas)
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
