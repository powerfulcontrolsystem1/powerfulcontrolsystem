package db

import (
	"database/sql"
	"math"
	"strconv"
	"strings"
	"time"
)

type LicenciaVencimientoCandidate struct {
	LicenciaTipo    string `json:"licencia_tipo"`
	LicenciaID      int64  `json:"licencia_id"`
	LicenciaPlanID  int64  `json:"licencia_plan_id,omitempty"`
	EmpresaID       int64  `json:"empresa_id"`
	EmpresaNombre   string `json:"empresa_nombre"`
	AdminEmail      string `json:"admin_email"`
	AdminNombre     string `json:"admin_nombre,omitempty"`
	LicenciaNombre  string `json:"licencia_nombre"`
	CodigoFuncion   string `json:"codigo_funcion,omitempty"`
	FechaFin        string `json:"fecha_fin"`
	DiasRestantes   int    `json:"dias_restantes"`
	DiasAviso       int    `json:"dias_aviso,omitempty"`
	AutoRenovar     int    `json:"auto_renovar,omitempty"`
	NotificacionKey string `json:"notificacion_key,omitempty"`
}

type LicenciaVencimientoNotificationLog struct {
	ID                 int64  `json:"id"`
	LicenciaTipo       string `json:"licencia_tipo"`
	LicenciaID         int64  `json:"licencia_id"`
	EmpresaID          int64  `json:"empresa_id"`
	DestinatarioEmail  string `json:"destinatario_email"`
	DiasAntes          int    `json:"dias_antes"`
	FechaFin           string `json:"fecha_fin"`
	EnviadoEn          string `json:"enviado_en,omitempty"`
	Estado             string `json:"estado"`
	Error              string `json:"error,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
}

func EnsureLicenciaVencimientoNotificacionesSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	if dbConn == nil {
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS licencia_vencimiento_notificaciones (
			id BIGSERIAL PRIMARY KEY,
			licencia_tipo TEXT NOT NULL,
			licencia_id BIGINT NOT NULL,
			empresa_id BIGINT NOT NULL,
			destinatario_email TEXT NOT NULL,
			dias_antes INTEGER DEFAULT 0,
			fecha_fin TEXT,
			enviado_en TEXT,
			estado TEXT DEFAULT 'pendiente',
			error TEXT,
			usuario_creador TEXT,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)
		)`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS licencia_tipo TEXT`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS licencia_id BIGINT`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS empresa_id BIGINT`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS destinatario_email TEXT`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS dias_antes INTEGER DEFAULT 0`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS fecha_fin TEXT`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS enviado_en TEXT`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'pendiente'`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS error TEXT`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`ALTER TABLE licencia_vencimiento_notificaciones ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_licencia_vencimiento_notificaciones_unica ON licencia_vencimiento_notificaciones(licencia_tipo, licencia_id, empresa_id, destinatario_email, dias_antes, fecha_fin)`,
		`CREATE INDEX IF NOT EXISTS ix_licencia_vencimiento_notificaciones_empresa ON licencia_vencimiento_notificaciones(empresa_id, estado, fecha_actualizacion DESC)`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func ListLicenciaVencimientoCandidates(dbSuper, dbEmp *sql.DB, maxDays int, now time.Time) ([]LicenciaVencimientoCandidate, error) {
	if dbSuper == nil {
		return []LicenciaVencimientoCandidate{}, nil
	}
	if maxDays <= 0 {
		return []LicenciaVencimientoCandidate{}, nil
	}
	if now.IsZero() {
		now = time.Now()
	}
	if err := EnsureLicenciasSchema(dbSuper); err != nil {
		return nil, err
	}
	if err := EnsureEmpresaLicenciasAdicionalesSchema(dbSuper); err != nil {
		return nil, err
	}

	out := make([]LicenciaVencimientoCandidate, 0)
	baseRows, err := querySQLCompat(dbSuper, `SELECT
		l.id,
		COALESCE(l.empresa_id, 0),
		COALESCE(l.nombre, ''),
		COALESCE(l.codigo_funcion, ''),
		COALESCE(l.fecha_fin, '')
	FROM licencias l
	WHERE COALESCE(l.empresa_id, 0) > 0
		AND COALESCE(l.es_adicional, 0) = 0
		AND COALESCE(l.activo, 1) = 1
		AND COALESCE(l.fecha_fin, '') <> ''`)
	if err != nil {
		return nil, err
	}
	defer baseRows.Close()
	for baseRows.Next() {
		var c LicenciaVencimientoCandidate
		c.LicenciaTipo = "base"
		if err := baseRows.Scan(&c.LicenciaID, &c.EmpresaID, &c.LicenciaNombre, &c.CodigoFuncion, &c.FechaFin); err != nil {
			return nil, err
		}
		if candidateInVencimientoWindow(&c, maxDays, now) {
			enrichLicenciaVencimientoCandidate(dbSuper, dbEmp, &c)
			out = append(out, normalizeLicenciaVencimientoCandidate(c))
		}
	}
	if err := baseRows.Err(); err != nil {
		return nil, err
	}

	addonRows, err := querySQLCompat(dbSuper, `SELECT
		a.id,
		a.licencia_id,
		a.empresa_id,
		COALESCE(l.nombre, ''),
		COALESCE(l.codigo_funcion, ''),
		COALESCE(a.fecha_fin, ''),
		COALESCE(a.auto_renovar, 1)
	FROM empresa_licencias_adicionales a
	LEFT JOIN licencias l ON l.id = a.licencia_id
	WHERE COALESCE(a.empresa_id, 0) > 0
		AND COALESCE(a.activo, 1) = 1
		AND COALESCE(a.fecha_fin, '') <> ''`)
	if err != nil {
		return nil, err
	}
	defer addonRows.Close()
	for addonRows.Next() {
		var c LicenciaVencimientoCandidate
		c.LicenciaTipo = "adicional"
		if err := addonRows.Scan(&c.LicenciaID, &c.LicenciaPlanID, &c.EmpresaID, &c.LicenciaNombre, &c.CodigoFuncion, &c.FechaFin, &c.AutoRenovar); err != nil {
			return nil, err
		}
		if candidateInVencimientoWindow(&c, maxDays, now) {
			enrichLicenciaVencimientoCandidate(dbSuper, dbEmp, &c)
			out = append(out, normalizeLicenciaVencimientoCandidate(c))
		}
	}
	if err := addonRows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func enrichLicenciaVencimientoCandidate(dbSuper, dbEmp *sql.DB, c *LicenciaVencimientoCandidate) {
	if c == nil || c.EmpresaID <= 0 {
		return
	}
	if dbEmp != nil {
		if empresa, err := GetEmpresaByScopeID(dbEmp, c.EmpresaID); err == nil && empresa != nil {
			c.EmpresaNombre = strings.TrimSpace(empresa.Nombre)
			c.AdminEmail = strings.TrimSpace(empresa.UsuarioCreador)
		}
	}
	if dbSuper != nil && strings.TrimSpace(c.AdminEmail) != "" {
		if adm, err := GetAdminByEmail(dbSuper, c.AdminEmail); err == nil && adm != nil {
			c.AdminNombre = strings.TrimSpace(adm.Name)
		}
	}
	if dbSuper != nil && strings.TrimSpace(c.AdminEmail) == "" {
		if accesos, err := ListAdminEmpresaCompartidaAccesosByEmpresa(dbSuper, c.EmpresaID); err == nil {
			for _, access := range accesos {
				if !strings.EqualFold(strings.TrimSpace(access.Estado), "activo") {
					continue
				}
				if strings.TrimSpace(access.AdminEmail) == "" {
					continue
				}
				c.AdminEmail = strings.TrimSpace(access.AdminEmail)
				c.AdminNombre = strings.TrimSpace(access.AdminName)
				break
			}
		}
	}
}

func candidateInVencimientoWindow(c *LicenciaVencimientoCandidate, maxDays int, now time.Time) bool {
	if c == nil {
		return false
	}
	endTime, ok := parseMaybeTime(c.FechaFin)
	if !ok {
		return false
	}
	diff := endTime.Sub(now)
	if diff < 0 {
		return false
	}
	days := int(math.Ceil(diff.Hours() / 24))
	if days < 0 || days > maxDays {
		return false
	}
	c.DiasRestantes = days
	return true
}

func normalizeLicenciaVencimientoCandidate(c LicenciaVencimientoCandidate) LicenciaVencimientoCandidate {
	c.LicenciaTipo = strings.ToLower(strings.TrimSpace(c.LicenciaTipo))
	c.AdminEmail = strings.ToLower(strings.TrimSpace(c.AdminEmail))
	c.AdminNombre = strings.TrimSpace(c.AdminNombre)
	c.EmpresaNombre = strings.TrimSpace(c.EmpresaNombre)
	if c.EmpresaNombre == "" {
		c.EmpresaNombre = "Empresa " + strconv.FormatInt(c.EmpresaID, 10)
	}
	c.LicenciaNombre = strings.TrimSpace(c.LicenciaNombre)
	if c.LicenciaNombre == "" {
		c.LicenciaNombre = "Licencia"
	}
	c.FechaFin = strings.TrimSpace(c.FechaFin)
	c.NotificacionKey = c.LicenciaTipo + ":" + strconv.FormatInt(c.LicenciaID, 10) + ":" + strconv.FormatInt(c.EmpresaID, 10) + ":" + c.AdminEmail + ":" + strconv.Itoa(c.DiasAviso) + ":" + c.FechaFin
	return c
}

func LicenciaVencimientoNotificationSent(dbConn *sql.DB, licenciaTipo string, licenciaID, empresaID int64, destinatario string, diasAntes int, fechaFin string) (bool, error) {
	if dbConn == nil {
		return false, nil
	}
	if err := EnsureLicenciaVencimientoNotificacionesSchema(dbConn); err != nil {
		return false, err
	}
	licenciaTipo = strings.ToLower(strings.TrimSpace(licenciaTipo))
	destinatario = strings.ToLower(strings.TrimSpace(destinatario))
	fechaFin = strings.TrimSpace(fechaFin)
	var count int
	err := queryRowSQLCompat(dbConn, `SELECT COUNT(1)
		FROM licencia_vencimiento_notificaciones
		WHERE licencia_tipo = ?
			AND licencia_id = ?
			AND empresa_id = ?
			AND destinatario_email = ?
			AND dias_antes = ?
			AND COALESCE(fecha_fin, '') = ?
			AND lower(COALESCE(estado, '')) IN ('enviado', 'capturado')`,
		licenciaTipo, licenciaID, empresaID, destinatario, diasAntes, fechaFin).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func UpsertLicenciaVencimientoNotificationResult(dbConn *sql.DB, c LicenciaVencimientoCandidate, estado, errorText, usuario string) error {
	if dbConn == nil {
		return nil
	}
	if err := EnsureLicenciaVencimientoNotificacionesSchema(dbConn); err != nil {
		return err
	}
	c = normalizeLicenciaVencimientoCandidate(c)
	estado = strings.TrimSpace(estado)
	if estado == "" {
		estado = "pendiente"
	}
	usuario = strings.TrimSpace(usuario)
	if usuario == "" {
		usuario = "sistema"
	}
	errorText = strings.TrimSpace(errorText)
	if len(errorText) > 800 {
		errorText = errorText[:800]
	}
	nowExpr := sqlNowExpr()
	enviadoExpr := "NULL"
	if strings.EqualFold(estado, "enviado") || strings.EqualFold(estado, "capturado") {
		enviadoExpr = nowExpr
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO licencia_vencimiento_notificaciones (
		licencia_tipo, licencia_id, empresa_id, destinatario_email, dias_antes, fecha_fin, enviado_en, estado, error, usuario_creador, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, `+enviadoExpr+`, ?, ?, ?, `+nowExpr+`, `+nowExpr+`)
	ON CONFLICT (licencia_tipo, licencia_id, empresa_id, destinatario_email, dias_antes, fecha_fin)
	DO UPDATE SET enviado_en = EXCLUDED.enviado_en, estado = EXCLUDED.estado, error = EXCLUDED.error, usuario_creador = EXCLUDED.usuario_creador, fecha_actualizacion = `+nowExpr,
		c.LicenciaTipo, c.LicenciaID, c.EmpresaID, c.AdminEmail, c.DiasAviso, c.FechaFin, estado, errorText, usuario)
	return err
}

func ListLicenciaVencimientoNotificationLogs(dbConn *sql.DB, limit int) ([]LicenciaVencimientoNotificationLog, error) {
	if dbConn == nil {
		return []LicenciaVencimientoNotificationLog{}, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if err := EnsureLicenciaVencimientoNotificacionesSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, licencia_tipo, licencia_id, empresa_id, destinatario_email, dias_antes, COALESCE(fecha_fin, ''), COALESCE(enviado_en, ''), COALESCE(estado, ''), COALESCE(error, ''), COALESCE(usuario_creador, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, '')
		FROM licencia_vencimiento_notificaciones
		ORDER BY id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]LicenciaVencimientoNotificationLog, 0)
	for rows.Next() {
		var item LicenciaVencimientoNotificationLog
		if err := rows.Scan(&item.ID, &item.LicenciaTipo, &item.LicenciaID, &item.EmpresaID, &item.DestinatarioEmail, &item.DiasAntes, &item.FechaFin, &item.EnviadoEn, &item.Estado, &item.Error, &item.UsuarioCreador, &item.FechaCreacion, &item.FechaActualizacion); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
