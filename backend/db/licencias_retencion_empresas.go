package db

import (
	"database/sql"
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"
)

type LicenciaEmpresaRetencionCandidate struct {
	EmpresaID                  int64  `json:"empresa_id"`
	EmpresaNombre              string `json:"empresa_nombre"`
	EstadoEmpresa              string `json:"estado_empresa"`
	AdminEmail                 string `json:"admin_email,omitempty"`
	AdminNombre                string `json:"admin_nombre,omitempty"`
	UltimaLicenciaFin          string `json:"ultima_licencia_fin"`
	DiasVencida                int    `json:"dias_vencida"`
	RetencionDias              int    `json:"retencion_dias"`
	PreavisoDias               int    `json:"preaviso_dias"`
	FechaProgramadaEliminacion string `json:"fecha_programada_eliminacion"`
	FechaPreaviso              string `json:"fecha_preaviso"`
	PreavisoEnviadoEn          string `json:"preaviso_enviado_en,omitempty"`
	EliminadoEn                string `json:"eliminado_en,omitempty"`
	EstadoReporte              string `json:"estado_reporte,omitempty"`
	Error                      string `json:"error,omitempty"`
	DebePreavisar              bool   `json:"debe_preavisar"`
	DebeEliminar               bool   `json:"debe_eliminar"`
	ReportID                   int64  `json:"report_id,omitempty"`
}

type LicenciaEmpresaRetencionLog struct {
	ID                         int64  `json:"id"`
	EmpresaRefID               int64  `json:"empresa_ref_id"`
	EmpresaNombre              string `json:"empresa_nombre"`
	AdminEmail                 string `json:"admin_email,omitempty"`
	AdminNombre                string `json:"admin_nombre,omitempty"`
	EstadoEmpresa              string `json:"estado_empresa,omitempty"`
	UltimaLicenciaFin          string `json:"ultima_licencia_fin,omitempty"`
	RetencionDias              int    `json:"retencion_dias"`
	PreavisoDias               int    `json:"preaviso_dias"`
	FechaProgramadaEliminacion string `json:"fecha_programada_eliminacion,omitempty"`
	FechaPreaviso              string `json:"fecha_preaviso,omitempty"`
	PreavisoEnviadoEn          string `json:"preaviso_enviado_en,omitempty"`
	EliminadoEn                string `json:"eliminado_en,omitempty"`
	Estado                     string `json:"estado"`
	Error                      string `json:"error,omitempty"`
	DeleteResultJSON           string `json:"delete_result_json,omitempty"`
	UsuarioCreador             string `json:"usuario_creador,omitempty"`
	FechaCreacion              string `json:"fecha_creacion,omitempty"`
	FechaActualizacion         string `json:"fecha_actualizacion,omitempty"`
}

func EnsureLicenciaEmpresaRetencionSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS licencia_empresa_retencion_log (
			id BIGSERIAL PRIMARY KEY,
			empresa_ref_id BIGINT NOT NULL,
			empresa_nombre TEXT,
			admin_email TEXT,
			admin_nombre TEXT,
			estado_empresa TEXT,
			ultima_licencia_fin TEXT,
			retencion_dias INTEGER DEFAULT 365,
			preaviso_dias INTEGER DEFAULT 1,
			fecha_programada_eliminacion TEXT,
			fecha_preaviso TEXT,
			preaviso_enviado_en TEXT,
			eliminado_en TEXT,
			estado TEXT DEFAULT 'pendiente',
			error TEXT,
			delete_result_json TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)
		)`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS empresa_ref_id BIGINT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS empresa_nombre TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS admin_email TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS admin_nombre TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS estado_empresa TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS ultima_licencia_fin TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS retencion_dias INTEGER DEFAULT 365`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS preaviso_dias INTEGER DEFAULT 1`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS fecha_programada_eliminacion TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS fecha_preaviso TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS preaviso_enviado_en TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS eliminado_en TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'pendiente'`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS error TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS delete_result_json TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`ALTER TABLE licencia_empresa_retencion_log ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_licencia_empresa_retencion_unica ON licencia_empresa_retencion_log(empresa_ref_id, ultima_licencia_fin, fecha_programada_eliminacion)`,
		`CREATE INDEX IF NOT EXISTS ix_licencia_empresa_retencion_estado ON licencia_empresa_retencion_log(estado, fecha_actualizacion DESC)`,
		`CREATE INDEX IF NOT EXISTS ix_licencia_empresa_retencion_empresa ON licencia_empresa_retencion_log(empresa_ref_id, id DESC)`,
	}
	for _, stmt := range statements {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func ListLicenciaEmpresaRetencionCandidates(dbSuper, dbEmp *sql.DB, retentionDays, noticeDays int, now time.Time, limit int) ([]LicenciaEmpresaRetencionCandidate, error) {
	if dbSuper == nil || dbEmp == nil {
		return []LicenciaEmpresaRetencionCandidate{}, nil
	}
	if retentionDays <= 0 {
		retentionDays = 365
	}
	if retentionDays > 3650 {
		retentionDays = 3650
	}
	if noticeDays <= 0 {
		noticeDays = 1
	}
	if noticeDays > retentionDays {
		noticeDays = retentionDays
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if now.IsZero() {
		now = time.Now()
	}
	if err := EnsureLicenciasSchema(dbSuper); err != nil {
		return nil, err
	}
	if err := EnsureLicenciaEmpresaRetencionSchema(dbSuper); err != nil {
		return nil, err
	}
	empresas, err := GetEmpresas(dbEmp)
	if err != nil {
		return nil, err
	}
	out := make([]LicenciaEmpresaRetencionCandidate, 0)
	for _, emp := range empresas {
		if len(out) >= limit {
			break
		}
		if emp.ID <= 0 || !empresaRetencionEstadoNoOperativo(emp.Estado) {
			continue
		}
		scopeIDs := []int64{emp.ID}
		if emp.EmpresaID > 0 && emp.EmpresaID != emp.ID {
			scopeIDs = append(scopeIDs, emp.EmpresaID)
		}
		hasActive, lastExpired, lastExpiredRaw, err := licenciaEmpresaRetencionLicenseState(dbSuper, scopeIDs, now)
		if err != nil {
			return nil, err
		}
		if hasActive || lastExpired.IsZero() {
			continue
		}
		deletionAt := lastExpired.AddDate(0, 0, retentionDays)
		preNoticeAt := deletionAt.AddDate(0, 0, -noticeDays)
		if now.Before(preNoticeAt) {
			continue
		}
		c := LicenciaEmpresaRetencionCandidate{
			EmpresaID:                  emp.ID,
			EmpresaNombre:              strings.TrimSpace(emp.Nombre),
			EstadoEmpresa:              strings.TrimSpace(emp.Estado),
			AdminEmail:                 strings.ToLower(strings.TrimSpace(emp.UsuarioCreador)),
			UltimaLicenciaFin:          strings.TrimSpace(lastExpiredRaw),
			DiasVencida:                int(math.Floor(now.Sub(lastExpired).Hours() / 24)),
			RetencionDias:              retentionDays,
			PreavisoDias:               noticeDays,
			FechaProgramadaEliminacion: deletionAt.Format("2006-01-02 15:04:05"),
			FechaPreaviso:              preNoticeAt.Format("2006-01-02 15:04:05"),
		}
		enrichLicenciaEmpresaRetencionCandidate(dbSuper, &c)
		logItem, found, err := GetLicenciaEmpresaRetencionLog(dbSuper, c.EmpresaID, c.UltimaLicenciaFin, c.FechaProgramadaEliminacion)
		if err != nil {
			return nil, err
		}
		if found {
			c.ReportID = logItem.ID
			c.PreavisoEnviadoEn = logItem.PreavisoEnviadoEn
			c.EliminadoEn = logItem.EliminadoEn
			c.EstadoReporte = logItem.Estado
			c.Error = logItem.Error
		}
		c.DebePreavisar = strings.TrimSpace(c.PreavisoEnviadoEn) == "" && strings.TrimSpace(c.EliminadoEn) == ""
		c.DebeEliminar = strings.TrimSpace(c.PreavisoEnviadoEn) != "" && strings.TrimSpace(c.EliminadoEn) == "" && !now.Before(deletionAt)
		out = append(out, normalizeLicenciaEmpresaRetencionCandidate(c))
	}
	return out, nil
}

func empresaRetencionEstadoNoOperativo(estado string) bool {
	switch strings.ToLower(strings.TrimSpace(estado)) {
	case "inactivo", "inactiva", "suspendido", "suspendida", "bloqueado", "bloqueada", "vencido", "vencida", "deshabilitado", "deshabilitada":
		return true
	default:
		return false
	}
}

func licenciaEmpresaRetencionLicenseState(dbSuper *sql.DB, empresaIDs []int64, now time.Time) (bool, time.Time, string, error) {
	cleanIDs := make([]int64, 0, len(empresaIDs))
	seen := map[int64]bool{}
	for _, id := range empresaIDs {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		cleanIDs = append(cleanIDs, id)
	}
	if len(cleanIDs) == 0 {
		return false, time.Time{}, "", nil
	}
	placeholders := make([]string, 0, len(cleanIDs))
	args := make([]interface{}, 0, len(cleanIDs))
	for _, id := range cleanIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rows, err := querySQLCompat(dbSuper, `SELECT COALESCE(fecha_fin, ''), COALESCE(activo, 1)
		FROM licencias
		WHERE COALESCE(empresa_id, 0) IN (`+strings.Join(placeholders, ",")+`)
			AND COALESCE(es_adicional, 0) = 0`, args...)
	if err != nil {
		return false, time.Time{}, "", err
	}
	defer rows.Close()
	hasActive := false
	var lastExpired time.Time
	lastExpiredRaw := ""
	for rows.Next() {
		var raw string
		var active int
		if err := rows.Scan(&raw, &active); err != nil {
			return false, time.Time{}, "", err
		}
		raw = strings.TrimSpace(raw)
		end, ok := parseMaybeTime(raw)
		if active == 1 && (!ok || !end.Before(now)) {
			hasActive = true
			continue
		}
		if !ok || !end.Before(now) {
			continue
		}
		if lastExpired.IsZero() || end.After(lastExpired) {
			lastExpired = end
			lastExpiredRaw = raw
		}
	}
	if err := rows.Err(); err != nil {
		return false, time.Time{}, "", err
	}
	return hasActive, lastExpired, lastExpiredRaw, nil
}

func enrichLicenciaEmpresaRetencionCandidate(dbSuper *sql.DB, c *LicenciaEmpresaRetencionCandidate) {
	if c == nil || dbSuper == nil {
		return
	}
	if strings.TrimSpace(c.AdminEmail) != "" {
		if adm, err := GetAdminByEmail(dbSuper, c.AdminEmail); err == nil && adm != nil {
			c.AdminNombre = strings.TrimSpace(adm.Name)
		}
	}
	if strings.TrimSpace(c.AdminEmail) == "" {
		if accesos, err := ListAdminEmpresaCompartidaAccesosByEmpresa(dbSuper, c.EmpresaID); err == nil {
			for _, access := range accesos {
				if !strings.EqualFold(strings.TrimSpace(access.Estado), "activo") || strings.TrimSpace(access.AdminEmail) == "" {
					continue
				}
				c.AdminEmail = strings.ToLower(strings.TrimSpace(access.AdminEmail))
				c.AdminNombre = strings.TrimSpace(access.AdminName)
				break
			}
		}
	}
}

func normalizeLicenciaEmpresaRetencionCandidate(c LicenciaEmpresaRetencionCandidate) LicenciaEmpresaRetencionCandidate {
	c.EmpresaNombre = strings.TrimSpace(c.EmpresaNombre)
	if c.EmpresaNombre == "" {
		c.EmpresaNombre = "Empresa " + strconv.FormatInt(c.EmpresaID, 10)
	}
	c.EstadoEmpresa = strings.TrimSpace(c.EstadoEmpresa)
	c.AdminEmail = strings.ToLower(strings.TrimSpace(c.AdminEmail))
	c.AdminNombre = strings.TrimSpace(c.AdminNombre)
	c.UltimaLicenciaFin = strings.TrimSpace(c.UltimaLicenciaFin)
	c.FechaProgramadaEliminacion = strings.TrimSpace(c.FechaProgramadaEliminacion)
	c.FechaPreaviso = strings.TrimSpace(c.FechaPreaviso)
	c.EstadoReporte = strings.TrimSpace(c.EstadoReporte)
	c.Error = strings.TrimSpace(c.Error)
	return c
}

func GetLicenciaEmpresaRetencionLog(dbConn *sql.DB, empresaRefID int64, ultimaLicenciaFin, fechaProgramada string) (LicenciaEmpresaRetencionLog, bool, error) {
	if dbConn == nil || empresaRefID <= 0 {
		return LicenciaEmpresaRetencionLog{}, false, nil
	}
	if err := EnsureLicenciaEmpresaRetencionSchema(dbConn); err != nil {
		return LicenciaEmpresaRetencionLog{}, false, err
	}
	var item LicenciaEmpresaRetencionLog
	err := queryRowSQLCompat(dbConn, `SELECT id, empresa_ref_id, COALESCE(empresa_nombre, ''), COALESCE(admin_email, ''), COALESCE(admin_nombre, ''), COALESCE(estado_empresa, ''), COALESCE(ultima_licencia_fin, ''), COALESCE(retencion_dias, 365), COALESCE(preaviso_dias, 1), COALESCE(fecha_programada_eliminacion, ''), COALESCE(fecha_preaviso, ''), COALESCE(preaviso_enviado_en, ''), COALESCE(eliminado_en, ''), COALESCE(estado, ''), COALESCE(error, ''), COALESCE(delete_result_json, ''), COALESCE(usuario_creador, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, '')
		FROM licencia_empresa_retencion_log
		WHERE empresa_ref_id = ?
			AND COALESCE(ultima_licencia_fin, '') = ?
			AND COALESCE(fecha_programada_eliminacion, '') = ?
		LIMIT 1`, empresaRefID, strings.TrimSpace(ultimaLicenciaFin), strings.TrimSpace(fechaProgramada)).Scan(&item.ID, &item.EmpresaRefID, &item.EmpresaNombre, &item.AdminEmail, &item.AdminNombre, &item.EstadoEmpresa, &item.UltimaLicenciaFin, &item.RetencionDias, &item.PreavisoDias, &item.FechaProgramadaEliminacion, &item.FechaPreaviso, &item.PreavisoEnviadoEn, &item.EliminadoEn, &item.Estado, &item.Error, &item.DeleteResultJSON, &item.UsuarioCreador, &item.FechaCreacion, &item.FechaActualizacion)
	if err == sql.ErrNoRows {
		return LicenciaEmpresaRetencionLog{}, false, nil
	}
	if err != nil {
		return LicenciaEmpresaRetencionLog{}, false, err
	}
	return item, true, nil
}

func UpsertLicenciaEmpresaRetencionLog(dbConn *sql.DB, c LicenciaEmpresaRetencionCandidate, estado, errorText, deleteResultJSON, usuario string) error {
	if dbConn == nil {
		return nil
	}
	if err := EnsureLicenciaEmpresaRetencionSchema(dbConn); err != nil {
		return err
	}
	c = normalizeLicenciaEmpresaRetencionCandidate(c)
	estado = strings.TrimSpace(estado)
	if estado == "" {
		estado = "pendiente"
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	errorText = strings.TrimSpace(errorText)
	if len(errorText) > 1200 {
		errorText = errorText[:1200]
	}
	deleteResultJSON = strings.TrimSpace(deleteResultJSON)
	nowExpr := sqlNowExpr()
	preavisoExpr := "preaviso_enviado_en"
	if strings.EqualFold(estado, "preaviso_enviado") || strings.EqualFold(estado, "preaviso_capturado") {
		preavisoExpr = nowExpr
	}
	eliminadoExpr := "eliminado_en"
	if strings.EqualFold(estado, "eliminado") {
		eliminadoExpr = nowExpr
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO licencia_empresa_retencion_log (
		empresa_ref_id, empresa_nombre, admin_email, admin_nombre, estado_empresa, ultima_licencia_fin, retencion_dias, preaviso_dias, fecha_programada_eliminacion, fecha_preaviso, preaviso_enviado_en, eliminado_en, estado, error, delete_result_json, usuario_creador, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CASE WHEN ? IN ('preaviso_enviado', 'preaviso_capturado') THEN `+nowExpr+` ELSE NULL END, CASE WHEN ? = 'eliminado' THEN `+nowExpr+` ELSE NULL END, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`)
	ON CONFLICT (empresa_ref_id, ultima_licencia_fin, fecha_programada_eliminacion)
	DO UPDATE SET empresa_nombre = EXCLUDED.empresa_nombre,
		admin_email = EXCLUDED.admin_email,
		admin_nombre = EXCLUDED.admin_nombre,
		estado_empresa = EXCLUDED.estado_empresa,
		retencion_dias = EXCLUDED.retencion_dias,
		preaviso_dias = EXCLUDED.preaviso_dias,
		fecha_preaviso = EXCLUDED.fecha_preaviso,
		preaviso_enviado_en = CASE WHEN EXCLUDED.estado IN ('preaviso_enviado', 'preaviso_capturado') THEN `+preavisoExpr+` ELSE licencia_empresa_retencion_log.preaviso_enviado_en END,
		eliminado_en = CASE WHEN EXCLUDED.estado = 'eliminado' THEN `+eliminadoExpr+` ELSE licencia_empresa_retencion_log.eliminado_en END,
		estado = EXCLUDED.estado,
		error = EXCLUDED.error,
		delete_result_json = CASE WHEN COALESCE(EXCLUDED.delete_result_json, '') <> '' THEN EXCLUDED.delete_result_json ELSE licencia_empresa_retencion_log.delete_result_json END,
		usuario_creador = EXCLUDED.usuario_creador,
		fecha_actualizacion = `+nowExpr,
		c.EmpresaID, c.EmpresaNombre, c.AdminEmail, c.AdminNombre, c.EstadoEmpresa, c.UltimaLicenciaFin, c.RetencionDias, c.PreavisoDias, c.FechaProgramadaEliminacion, c.FechaPreaviso, estado, estado, estado, errorText, deleteResultJSON, usuario)
	return err
}

func UpsertLicenciaEmpresaRetencionDeleted(dbConn *sql.DB, c LicenciaEmpresaRetencionCandidate, deleteResult interface{}, usuario string) error {
	data := ""
	if deleteResult != nil {
		if b, err := json.Marshal(deleteResult); err == nil {
			data = string(b)
		}
	}
	return UpsertLicenciaEmpresaRetencionLog(dbConn, c, "eliminado", "", data, usuario)
}

func ListLicenciaEmpresaRetencionLogs(dbConn *sql.DB, limit int, estado string) ([]LicenciaEmpresaRetencionLog, error) {
	if dbConn == nil {
		return []LicenciaEmpresaRetencionLog{}, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if err := EnsureLicenciaEmpresaRetencionSchema(dbConn); err != nil {
		return nil, err
	}
	args := []interface{}{}
	where := ""
	if strings.TrimSpace(estado) != "" {
		where = "WHERE lower(COALESCE(estado, '')) = ?"
		args = append(args, strings.ToLower(strings.TrimSpace(estado)))
	}
	args = append(args, limit)
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_ref_id, COALESCE(empresa_nombre, ''), COALESCE(admin_email, ''), COALESCE(admin_nombre, ''), COALESCE(estado_empresa, ''), COALESCE(ultima_licencia_fin, ''), COALESCE(retencion_dias, 365), COALESCE(preaviso_dias, 1), COALESCE(fecha_programada_eliminacion, ''), COALESCE(fecha_preaviso, ''), COALESCE(preaviso_enviado_en, ''), COALESCE(eliminado_en, ''), COALESCE(estado, ''), COALESCE(error, ''), COALESCE(delete_result_json, ''), COALESCE(usuario_creador, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, '')
		FROM licencia_empresa_retencion_log
		`+where+`
		ORDER BY id DESC
		LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LicenciaEmpresaRetencionLog, 0)
	for rows.Next() {
		var item LicenciaEmpresaRetencionLog
		if err := rows.Scan(&item.ID, &item.EmpresaRefID, &item.EmpresaNombre, &item.AdminEmail, &item.AdminNombre, &item.EstadoEmpresa, &item.UltimaLicenciaFin, &item.RetencionDias, &item.PreavisoDias, &item.FechaProgramadaEliminacion, &item.FechaPreaviso, &item.PreavisoEnviadoEn, &item.EliminadoEn, &item.Estado, &item.Error, &item.DeleteResultJSON, &item.UsuarioCreador, &item.FechaCreacion, &item.FechaActualizacion); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
