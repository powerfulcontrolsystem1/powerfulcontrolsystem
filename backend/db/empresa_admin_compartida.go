package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	adminEmpresaCompartidaAccessCacheMu  sync.Mutex
	adminEmpresaCompartidaAccessCache    = map[string]cachedAdminEmpresaCompartidaAccess{}
	adminEmpresaCompartidaAccessCacheTTL = 60 * time.Second
)

type cachedAdminEmpresaCompartidaAccess struct {
	Item     *AdminEmpresaCompartidaAcceso
	LoadedAt time.Time
}

type AdminEmpresaCompartidaAcceso struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	AdminEmail         string `json:"admin_email"`
	AdminName          string `json:"admin_name,omitempty"`
	CompartidoPorEmail string `json:"compartido_por_email,omitempty"`
	CompartidoPorName  string `json:"compartido_por_name,omitempty"`
	InvitacionID       int64  `json:"invitacion_id,omitempty"`
	NivelAcceso        string `json:"nivel_acceso,omitempty"`
	ModulosPermitidos  string `json:"modulos_permitidos,omitempty"`
	FechaAceptada      string `json:"fecha_aceptada,omitempty"`
	FechaRevocada      string `json:"fecha_revocada,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

type AdminEmpresaCompartidaInvitacion struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	AdminEmail         string `json:"admin_email"`
	AdminName          string `json:"admin_name,omitempty"`
	InvitadoPorEmail   string `json:"invitado_por_email,omitempty"`
	InvitadoPorName    string `json:"invitado_por_name,omitempty"`
	TokenHash          string `json:"-"`
	NivelAcceso        string `json:"nivel_acceso,omitempty"`
	ModulosPermitidos  string `json:"modulos_permitidos,omitempty"`
	Mensaje            string `json:"mensaje,omitempty"`
	ExpiraEn           string `json:"expira_en,omitempty"`
	AceptadaEn         string `json:"aceptada_en,omitempty"`
	RechazadaEn        string `json:"rechazada_en,omitempty"`
	RevocadaEn         string `json:"revocada_en,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

func normalizeAdminEmpresaCompartidaEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func adminEmpresaCompartidaAccessCacheKey(empresaID int64, adminEmail string) string {
	return fmt.Sprintf("%d|%s", empresaID, normalizeAdminEmpresaCompartidaEmail(adminEmail))
}

func InvalidateAdminEmpresaCompartidaAccessCache(empresaID int64, adminEmail string) {
	adminEmail = normalizeAdminEmpresaCompartidaEmail(adminEmail)
	if empresaID <= 0 || adminEmail == "" {
		return
	}
	adminEmpresaCompartidaAccessCacheMu.Lock()
	delete(adminEmpresaCompartidaAccessCache, adminEmpresaCompartidaAccessCacheKey(empresaID, adminEmail))
	adminEmpresaCompartidaAccessCacheMu.Unlock()
}

func EnsureAdminEmpresaCompartidaSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}

	if shouldUsePostgresCompat(dbConn) {
		statements := []string{
			`CREATE TABLE IF NOT EXISTS admin_empresa_compartida (
				id BIGSERIAL PRIMARY KEY,
				empresa_id BIGINT NOT NULL,
				admin_email TEXT NOT NULL,
				compartido_por_email TEXT,
				invitacion_id BIGINT,
				nivel_acceso TEXT DEFAULT 'acceso_total',
				modulos_permitidos TEXT,
				fecha_aceptada TEXT,
				fecha_revocada TEXT,
				fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
				fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
				usuario_creador TEXT,
				estado TEXT DEFAULT 'activo',
				observaciones TEXT
			)`,
			`CREATE INDEX IF NOT EXISTS ix_admin_empresa_compartida_empresa_admin ON admin_empresa_compartida (empresa_id, admin_email)`,
			`CREATE INDEX IF NOT EXISTS ix_admin_empresa_compartida_admin_estado ON admin_empresa_compartida (admin_email, estado)`,
			`CREATE TABLE IF NOT EXISTS admin_empresa_compartida_invitaciones (
				id BIGSERIAL PRIMARY KEY,
				empresa_id BIGINT NOT NULL,
				admin_email TEXT NOT NULL,
				invitado_por_email TEXT,
				token_hash TEXT NOT NULL,
				nivel_acceso TEXT DEFAULT 'acceso_total',
				modulos_permitidos TEXT,
				mensaje TEXT,
				expira_en TEXT,
				aceptada_en TEXT,
				rechazada_en TEXT,
				revocada_en TEXT,
				fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
				fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
				usuario_creador TEXT,
				estado TEXT DEFAULT 'pendiente',
				observaciones TEXT
			)`,
			`CREATE INDEX IF NOT EXISTS ix_admin_empresa_compartida_invitaciones_empresa_admin ON admin_empresa_compartida_invitaciones (empresa_id, admin_email)`,
			`CREATE INDEX IF NOT EXISTS ix_admin_empresa_compartida_invitaciones_token ON admin_empresa_compartida_invitaciones (token_hash)`,
		}
		for _, stmt := range statements {
			if _, err := dbConn.Exec(stmt); err != nil {
				return fmt.Errorf("ensure admin empresa compartida postgres schema: %w; stmt=%s", err, stmt)
			}
		}
		for _, col := range []struct {
			table string
			name  string
			def   string
		}{
			{"admin_empresa_compartida", "nivel_acceso", "TEXT DEFAULT 'acceso_total'"},
			{"admin_empresa_compartida", "modulos_permitidos", "TEXT"},
			{"admin_empresa_compartida_invitaciones", "nivel_acceso", "TEXT DEFAULT 'acceso_total'"},
			{"admin_empresa_compartida_invitaciones", "modulos_permitidos", "TEXT"},
		} {
			if err := ensureColumnIfMissing(dbConn, col.table, col.name, col.def); err != nil {
				return fmt.Errorf("ensure admin empresa compartida column %s.%s: %w", col.table, col.name, err)
			}
		}
		return nil
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS admin_empresa_compartida (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			admin_email TEXT NOT NULL,
			compartido_por_email TEXT,
			invitacion_id INTEGER,
			nivel_acceso TEXT DEFAULT 'acceso_total',
			modulos_permitidos TEXT,
			fecha_aceptada TEXT,
			fecha_revocada TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_admin_empresa_compartida_empresa_admin ON admin_empresa_compartida (empresa_id, admin_email)`,
		`CREATE INDEX IF NOT EXISTS ix_admin_empresa_compartida_admin_estado ON admin_empresa_compartida (admin_email, estado)`,
		`CREATE TABLE IF NOT EXISTS admin_empresa_compartida_invitaciones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			admin_email TEXT NOT NULL,
			invitado_por_email TEXT,
			token_hash TEXT NOT NULL,
			nivel_acceso TEXT DEFAULT 'acceso_total',
			modulos_permitidos TEXT,
			mensaje TEXT,
			expira_en TEXT,
			aceptada_en TEXT,
			rechazada_en TEXT,
			revocada_en TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'pendiente',
			observaciones TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS ix_admin_empresa_compartida_invitaciones_empresa_admin ON admin_empresa_compartida_invitaciones (empresa_id, admin_email)`,
		`CREATE INDEX IF NOT EXISTS ix_admin_empresa_compartida_invitaciones_token ON admin_empresa_compartida_invitaciones (token_hash)`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return fmt.Errorf("ensure admin empresa compartida schema: %w; stmt=%s", err, stmt)
		}
	}
	for _, col := range []struct {
		table string
		name  string
		def   string
	}{
		{"admin_empresa_compartida", "nivel_acceso", "TEXT DEFAULT 'acceso_total'"},
		{"admin_empresa_compartida", "modulos_permitidos", "TEXT"},
		{"admin_empresa_compartida_invitaciones", "nivel_acceso", "TEXT DEFAULT 'acceso_total'"},
		{"admin_empresa_compartida_invitaciones", "modulos_permitidos", "TEXT"},
	} {
		if err := ensureColumnIfMissing(dbConn, col.table, col.name, col.def); err != nil {
			return fmt.Errorf("ensure admin empresa compartida column %s.%s: %w", col.table, col.name, err)
		}
	}
	return nil
}

func scanAdminEmpresaCompartidaAcceso(rows *sql.Rows) (AdminEmpresaCompartidaAcceso, error) {
	var item AdminEmpresaCompartidaAcceso
	var invitacionID sql.NullInt64
	if err := rows.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.AdminEmail,
		&item.AdminName,
		&item.CompartidoPorEmail,
		&item.CompartidoPorName,
		&invitacionID,
		&item.NivelAcceso,
		&item.ModulosPermitidos,
		&item.FechaAceptada,
		&item.FechaRevocada,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return AdminEmpresaCompartidaAcceso{}, err
	}
	if invitacionID.Valid {
		item.InvitacionID = invitacionID.Int64
	}
	return item, nil
}

func scanAdminEmpresaCompartidaInvitacion(rows *sql.Rows) (AdminEmpresaCompartidaInvitacion, error) {
	var item AdminEmpresaCompartidaInvitacion
	if err := rows.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.AdminEmail,
		&item.AdminName,
		&item.InvitadoPorEmail,
		&item.InvitadoPorName,
		&item.TokenHash,
		&item.NivelAcceso,
		&item.ModulosPermitidos,
		&item.Mensaje,
		&item.ExpiraEn,
		&item.AceptadaEn,
		&item.RechazadaEn,
		&item.RevocadaEn,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return AdminEmpresaCompartidaInvitacion{}, err
	}
	return item, nil
}

func ListAdminEmpresaCompartidaAccesosByEmpresa(dbConn *sql.DB, empresaID int64) ([]AdminEmpresaCompartidaAcceso, error) {
	if dbConn == nil || empresaID <= 0 {
		return []AdminEmpresaCompartidaAcceso{}, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		a.id,
		a.empresa_id,
		COALESCE(a.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(a.compartido_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(a.invitacion_id, 0),
		COALESCE(a.nivel_acceso, 'acceso_total'),
		COALESCE(a.modulos_permitidos, ''),
		COALESCE(a.fecha_aceptada, ''),
		COALESCE(a.fecha_revocada, ''),
		COALESCE(a.fecha_creacion, ''),
		COALESCE(a.fecha_actualizacion, ''),
		COALESCE(a.usuario_creador, ''),
		COALESCE(a.estado, 'activo'),
		COALESCE(a.observaciones, '')
	FROM admin_empresa_compartida a
	LEFT JOIN administradores adm ON lower(adm.email) = lower(a.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(a.compartido_por_email)
	WHERE a.empresa_id = ?
	ORDER BY CASE WHEN lower(COALESCE(a.estado, 'activo')) = 'activo' THEN 0 ELSE 1 END, a.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AdminEmpresaCompartidaAcceso, 0)
	for rows.Next() {
		item, scanErr := scanAdminEmpresaCompartidaAcceso(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func ListAdminEmpresaCompartidaInvitacionesByEmpresa(dbConn *sql.DB, empresaID int64) ([]AdminEmpresaCompartidaInvitacion, error) {
	if dbConn == nil || empresaID <= 0 {
		return []AdminEmpresaCompartidaInvitacion{}, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(i.invitado_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(i.token_hash, ''),
		COALESCE(i.nivel_acceso, 'acceso_total'),
		COALESCE(i.modulos_permitidos, ''),
		COALESCE(i.mensaje, ''),
		COALESCE(i.expira_en, ''),
		COALESCE(i.aceptada_en, ''),
		COALESCE(i.rechazada_en, ''),
		COALESCE(i.revocada_en, ''),
		COALESCE(i.fecha_creacion, ''),
		COALESCE(i.fecha_actualizacion, ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'pendiente'),
		COALESCE(i.observaciones, '')
	FROM admin_empresa_compartida_invitaciones i
	LEFT JOIN administradores adm ON lower(adm.email) = lower(i.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(i.invitado_por_email)
	WHERE i.empresa_id = ?
	ORDER BY i.id DESC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AdminEmpresaCompartidaInvitacion, 0)
	for rows.Next() {
		item, scanErr := scanAdminEmpresaCompartidaInvitacion(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func ListActiveAdminEmpresaCompartidaAccesosByAdmin(dbConn *sql.DB, adminEmail string) ([]AdminEmpresaCompartidaAcceso, error) {
	adminEmail = normalizeAdminEmpresaCompartidaEmail(adminEmail)
	if dbConn == nil || adminEmail == "" {
		return []AdminEmpresaCompartidaAcceso{}, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		a.id,
		a.empresa_id,
		COALESCE(a.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(a.compartido_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(a.invitacion_id, 0),
		COALESCE(a.nivel_acceso, 'acceso_total'),
		COALESCE(a.modulos_permitidos, ''),
		COALESCE(a.fecha_aceptada, ''),
		COALESCE(a.fecha_revocada, ''),
		COALESCE(a.fecha_creacion, ''),
		COALESCE(a.fecha_actualizacion, ''),
		COALESCE(a.usuario_creador, ''),
		COALESCE(a.estado, 'activo'),
		COALESCE(a.observaciones, '')
	FROM admin_empresa_compartida a
	LEFT JOIN administradores adm ON lower(adm.email) = lower(a.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(a.compartido_por_email)
	WHERE lower(COALESCE(a.admin_email, '')) = lower(?)
	  AND lower(COALESCE(a.estado, 'activo')) = 'activo'
	  AND COALESCE(a.fecha_revocada, '') = ''
	ORDER BY a.id DESC`, adminEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AdminEmpresaCompartidaAcceso, 0)
	for rows.Next() {
		item, scanErr := scanAdminEmpresaCompartidaAcceso(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func ListActiveAdminEmpresaCompartidaAccesosBySharer(dbConn *sql.DB, sharedByEmail string) ([]AdminEmpresaCompartidaAcceso, error) {
	sharedByEmail = normalizeAdminEmpresaCompartidaEmail(sharedByEmail)
	if dbConn == nil || sharedByEmail == "" {
		return []AdminEmpresaCompartidaAcceso{}, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		a.id,
		a.empresa_id,
		COALESCE(a.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(a.compartido_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(a.invitacion_id, 0),
		COALESCE(a.nivel_acceso, 'acceso_total'),
		COALESCE(a.modulos_permitidos, ''),
		COALESCE(a.fecha_aceptada, ''),
		COALESCE(a.fecha_revocada, ''),
		COALESCE(a.fecha_creacion, ''),
		COALESCE(a.fecha_actualizacion, ''),
		COALESCE(a.usuario_creador, ''),
		COALESCE(a.estado, 'activo'),
		COALESCE(a.observaciones, '')
	FROM admin_empresa_compartida a
	LEFT JOIN administradores adm ON lower(adm.email) = lower(a.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(a.compartido_por_email)
	WHERE lower(COALESCE(a.compartido_por_email, '')) = lower(?)
	  AND lower(COALESCE(a.estado, 'activo')) = 'activo'
	  AND COALESCE(a.fecha_revocada, '') = ''
	ORDER BY a.id DESC`, sharedByEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AdminEmpresaCompartidaAcceso, 0)
	for rows.Next() {
		item, scanErr := scanAdminEmpresaCompartidaAcceso(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func HasActiveAdminEmpresaCompartidaAccesoBySharer(dbConn *sql.DB, empresaID int64, sharedByEmail string) (bool, error) {
	sharedByEmail = normalizeAdminEmpresaCompartidaEmail(sharedByEmail)
	if dbConn == nil || empresaID <= 0 || sharedByEmail == "" {
		return false, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return false, err
	}
	var total int
	if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1)
		FROM admin_empresa_compartida
		WHERE empresa_id = ?
		  AND lower(COALESCE(compartido_por_email, '')) = lower(?)
		  AND lower(COALESCE(estado, 'activo')) = 'activo'
		  AND COALESCE(fecha_revocada, '') = ''`, empresaID, sharedByEmail).Scan(&total); err != nil {
		return false, err
	}
	return total > 0, nil
}

func ListPendingAdminEmpresaCompartidaInvitacionesByAdmin(dbConn *sql.DB, adminEmail string) ([]AdminEmpresaCompartidaInvitacion, error) {
	adminEmail = normalizeAdminEmpresaCompartidaEmail(adminEmail)
	if dbConn == nil || adminEmail == "" {
		return []AdminEmpresaCompartidaInvitacion{}, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(i.invitado_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(i.token_hash, ''),
		COALESCE(i.nivel_acceso, 'acceso_total'),
		COALESCE(i.modulos_permitidos, ''),
		COALESCE(i.mensaje, ''),
		COALESCE(i.expira_en, ''),
		COALESCE(i.aceptada_en, ''),
		COALESCE(i.rechazada_en, ''),
		COALESCE(i.revocada_en, ''),
		COALESCE(i.fecha_creacion, ''),
		COALESCE(i.fecha_actualizacion, ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'pendiente'),
		COALESCE(i.observaciones, '')
	FROM admin_empresa_compartida_invitaciones i
	LEFT JOIN administradores adm ON lower(adm.email) = lower(i.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(i.invitado_por_email)
	WHERE lower(COALESCE(i.admin_email, '')) = lower(?)
	  AND lower(COALESCE(i.estado, 'pendiente')) = 'pendiente'
	  AND COALESCE(i.aceptada_en, '') = ''
	  AND COALESCE(i.rechazada_en, '') = ''
	  AND COALESCE(i.revocada_en, '') = ''
	ORDER BY i.id DESC`, adminEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]AdminEmpresaCompartidaInvitacion, 0)
	for rows.Next() {
		item, scanErr := scanAdminEmpresaCompartidaInvitacion(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func GetActiveAdminEmpresaCompartidaAcceso(dbConn *sql.DB, empresaID int64, adminEmail string) (*AdminEmpresaCompartidaAcceso, error) {
	adminEmail = normalizeAdminEmpresaCompartidaEmail(adminEmail)
	if dbConn == nil || empresaID <= 0 || adminEmail == "" {
		return nil, nil
	}
	cacheKey := adminEmpresaCompartidaAccessCacheKey(empresaID, adminEmail)
	adminEmpresaCompartidaAccessCacheMu.Lock()
	if cached, ok := adminEmpresaCompartidaAccessCache[cacheKey]; ok && time.Since(cached.LoadedAt) < adminEmpresaCompartidaAccessCacheTTL {
		adminEmpresaCompartidaAccessCacheMu.Unlock()
		if cached.Item == nil {
			return nil, nil
		}
		copyItem := *cached.Item
		return &copyItem, nil
	}
	adminEmpresaCompartidaAccessCacheMu.Unlock()
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		a.id,
		a.empresa_id,
		COALESCE(a.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(a.compartido_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(a.invitacion_id, 0),
		COALESCE(a.nivel_acceso, 'acceso_total'),
		COALESCE(a.modulos_permitidos, ''),
		COALESCE(a.fecha_aceptada, ''),
		COALESCE(a.fecha_revocada, ''),
		COALESCE(a.fecha_creacion, ''),
		COALESCE(a.fecha_actualizacion, ''),
		COALESCE(a.usuario_creador, ''),
		COALESCE(a.estado, 'activo'),
		COALESCE(a.observaciones, '')
	FROM admin_empresa_compartida a
	LEFT JOIN administradores adm ON lower(adm.email) = lower(a.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(a.compartido_por_email)
	WHERE a.empresa_id = ?
	  AND lower(COALESCE(a.admin_email, '')) = lower(?)
	  AND lower(COALESCE(a.estado, 'activo')) = 'activo'
	  AND COALESCE(a.fecha_revocada, '') = ''
	ORDER BY a.id DESC`, empresaID, adminEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		adminEmpresaCompartidaAccessCacheMu.Lock()
		adminEmpresaCompartidaAccessCache[cacheKey] = cachedAdminEmpresaCompartidaAccess{Item: nil, LoadedAt: time.Now()}
		adminEmpresaCompartidaAccessCacheMu.Unlock()
		return nil, nil
	}
	item, scanErr := scanAdminEmpresaCompartidaAcceso(rows)
	if scanErr != nil {
		return nil, scanErr
	}
	adminEmpresaCompartidaAccessCacheMu.Lock()
	adminEmpresaCompartidaAccessCache[cacheKey] = cachedAdminEmpresaCompartidaAccess{Item: &item, LoadedAt: time.Now()}
	adminEmpresaCompartidaAccessCacheMu.Unlock()
	copyItem := item
	return &copyItem, nil
}

func GetPendingAdminEmpresaCompartidaInvitacion(dbConn *sql.DB, empresaID int64, adminEmail string) (*AdminEmpresaCompartidaInvitacion, error) {
	adminEmail = normalizeAdminEmpresaCompartidaEmail(adminEmail)
	if dbConn == nil || empresaID <= 0 || adminEmail == "" {
		return nil, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(i.invitado_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(i.token_hash, ''),
		COALESCE(i.nivel_acceso, 'acceso_total'),
		COALESCE(i.modulos_permitidos, ''),
		COALESCE(i.mensaje, ''),
		COALESCE(i.expira_en, ''),
		COALESCE(i.aceptada_en, ''),
		COALESCE(i.rechazada_en, ''),
		COALESCE(i.revocada_en, ''),
		COALESCE(i.fecha_creacion, ''),
		COALESCE(i.fecha_actualizacion, ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'pendiente'),
		COALESCE(i.observaciones, '')
	FROM admin_empresa_compartida_invitaciones i
	LEFT JOIN administradores adm ON lower(adm.email) = lower(i.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(i.invitado_por_email)
	WHERE i.empresa_id = ?
	  AND lower(COALESCE(i.admin_email, '')) = lower(?)
	  AND lower(COALESCE(i.estado, 'pendiente')) = 'pendiente'
	  AND COALESCE(i.aceptada_en, '') = ''
	  AND COALESCE(i.rechazada_en, '') = ''
	  AND COALESCE(i.revocada_en, '') = ''
	ORDER BY i.id DESC`, empresaID, adminEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	item, scanErr := scanAdminEmpresaCompartidaInvitacion(rows)
	if scanErr != nil {
		return nil, scanErr
	}
	return &item, nil
}

func GetAdminEmpresaCompartidaInvitacionByID(dbConn *sql.DB, id int64) (*AdminEmpresaCompartidaInvitacion, error) {
	if dbConn == nil || id <= 0 {
		return nil, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(i.invitado_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(i.token_hash, ''),
		COALESCE(i.nivel_acceso, 'acceso_total'),
		COALESCE(i.modulos_permitidos, ''),
		COALESCE(i.mensaje, ''),
		COALESCE(i.expira_en, ''),
		COALESCE(i.aceptada_en, ''),
		COALESCE(i.rechazada_en, ''),
		COALESCE(i.revocada_en, ''),
		COALESCE(i.fecha_creacion, ''),
		COALESCE(i.fecha_actualizacion, ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'pendiente'),
		COALESCE(i.observaciones, '')
	FROM admin_empresa_compartida_invitaciones i
	LEFT JOIN administradores adm ON lower(adm.email) = lower(i.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(i.invitado_por_email)
	WHERE i.id = ?
	LIMIT 1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	item, scanErr := scanAdminEmpresaCompartidaInvitacion(rows)
	if scanErr != nil {
		return nil, scanErr
	}
	return &item, nil
}

func GetAdminEmpresaCompartidaInvitacionByTokenHash(dbConn *sql.DB, tokenHash string) (*AdminEmpresaCompartidaInvitacion, error) {
	tokenHash = strings.TrimSpace(tokenHash)
	if dbConn == nil || tokenHash == "" {
		return nil, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		i.id,
		i.empresa_id,
		COALESCE(i.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(i.invitado_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(i.token_hash, ''),
		COALESCE(i.nivel_acceso, 'acceso_total'),
		COALESCE(i.modulos_permitidos, ''),
		COALESCE(i.mensaje, ''),
		COALESCE(i.expira_en, ''),
		COALESCE(i.aceptada_en, ''),
		COALESCE(i.rechazada_en, ''),
		COALESCE(i.revocada_en, ''),
		COALESCE(i.fecha_creacion, ''),
		COALESCE(i.fecha_actualizacion, ''),
		COALESCE(i.usuario_creador, ''),
		COALESCE(i.estado, 'pendiente'),
		COALESCE(i.observaciones, '')
	FROM admin_empresa_compartida_invitaciones i
	LEFT JOIN administradores adm ON lower(adm.email) = lower(i.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(i.invitado_por_email)
	WHERE i.token_hash = ?
	LIMIT 1`, tokenHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	item, scanErr := scanAdminEmpresaCompartidaInvitacion(rows)
	if scanErr != nil {
		return nil, scanErr
	}
	return &item, nil
}

func CreateAdminEmpresaCompartidaInvitacion(dbConn *sql.DB, payload AdminEmpresaCompartidaInvitacion) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is required")
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return 0, err
	}
	payload.AdminEmail = normalizeAdminEmpresaCompartidaEmail(payload.AdminEmail)
	payload.InvitadoPorEmail = normalizeAdminEmpresaCompartidaEmail(payload.InvitadoPorEmail)
	payload.TokenHash = strings.TrimSpace(payload.TokenHash)
	payload.NivelAcceso = strings.ToLower(strings.TrimSpace(payload.NivelAcceso))
	payload.ModulosPermitidos = strings.TrimSpace(payload.ModulosPermitidos)
	payload.Mensaje = strings.TrimSpace(payload.Mensaje)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = strings.TrimSpace(payload.Estado)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	if payload.NivelAcceso == "" {
		payload.NivelAcceso = "acceso_total"
	}
	if payload.Estado == "" {
		payload.Estado = "pendiente"
	}
	if payload.EmpresaID <= 0 || payload.AdminEmail == "" || payload.TokenHash == "" {
		return 0, fmt.Errorf("empresa_id, admin_email y token_hash son obligatorios")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO admin_empresa_compartida_invitaciones (
		empresa_id,
		admin_email,
		invitado_por_email,
		token_hash,
		nivel_acceso,
		modulos_permitidos,
		mensaje,
		expira_en,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+sqlNowExpr()+`, `+sqlNowExpr()+`)`,
		payload.EmpresaID,
		payload.AdminEmail,
		payload.InvitadoPorEmail,
		payload.TokenHash,
		payload.NivelAcceso,
		payload.ModulosPermitidos,
		payload.Mensaje,
		payload.ExpiraEn,
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func RefreshAdminEmpresaCompartidaInvitacion(dbConn *sql.DB, id int64, tokenHash, mensaje, expiraEn, usuario, nivelAcceso, modulosPermitidos string) error {
	if dbConn == nil || id <= 0 {
		return fmt.Errorf("id invalido")
	}
	nivelAcceso = strings.ToLower(strings.TrimSpace(nivelAcceso))
	if nivelAcceso == "" {
		nivelAcceso = "acceso_total"
	}
	_, err := execSQLCompat(dbConn, `UPDATE admin_empresa_compartida_invitaciones
	SET token_hash = ?,
	    nivel_acceso = ?,
	    modulos_permitidos = ?,
	    mensaje = ?,
	    expira_en = ?,
	    fecha_actualizacion = `+sqlNowExpr()+`,
	    usuario_creador = ?,
	    aceptada_en = '',
	    rechazada_en = '',
	    revocada_en = '',
	    estado = 'pendiente'
	WHERE id = ?`, strings.TrimSpace(tokenHash), nivelAcceso, strings.TrimSpace(modulosPermitidos), strings.TrimSpace(mensaje), strings.TrimSpace(expiraEn), strings.TrimSpace(usuario), id)
	return err
}

func MarkAdminEmpresaCompartidaInvitacionAccepted(dbConn *sql.DB, id int64, acceptedAt, usuario string) error {
	if dbConn == nil || id <= 0 {
		return fmt.Errorf("id invalido")
	}
	_, err := execSQLCompat(dbConn, `UPDATE admin_empresa_compartida_invitaciones
	SET aceptada_en = ?,
	    token_hash = '',
	    fecha_actualizacion = `+sqlNowExpr()+`,
	    usuario_creador = ?,
	    estado = 'aceptada'
	WHERE id = ?`, strings.TrimSpace(acceptedAt), strings.TrimSpace(usuario), id)
	return err
}

func SetAdminEmpresaCompartidaInvitacionEstado(dbConn *sql.DB, id int64, estado, usuario string) error {
	if dbConn == nil || id <= 0 {
		return fmt.Errorf("id invalido")
	}
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado == "" {
		estado = "revocada"
	}
	column := "revocada_en"
	switch estado {
	case "rechazada":
		column = "rechazada_en"
	case "aceptada":
		column = "aceptada_en"
	case "expirada", "revocada":
		column = "revocada_en"
	}
	_, err := execSQLCompat(dbConn, `UPDATE admin_empresa_compartida_invitaciones
	SET estado = ?,
	    `+column+` = ?,
	    fecha_actualizacion = `+sqlNowExpr()+`,
	    usuario_creador = ?
	WHERE id = ?`, estado, sqlNowValue(), strings.TrimSpace(usuario), id)
	return err
}

func UpsertAdminEmpresaCompartidaAcceso(dbConn *sql.DB, payload AdminEmpresaCompartidaAcceso) (int64, error) {
	if dbConn == nil {
		return 0, fmt.Errorf("db connection is required")
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return 0, err
	}
	payload.AdminEmail = normalizeAdminEmpresaCompartidaEmail(payload.AdminEmail)
	payload.CompartidoPorEmail = normalizeAdminEmpresaCompartidaEmail(payload.CompartidoPorEmail)
	payload.NivelAcceso = strings.ToLower(strings.TrimSpace(payload.NivelAcceso))
	payload.ModulosPermitidos = strings.TrimSpace(payload.ModulosPermitidos)
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Estado = strings.TrimSpace(payload.Estado)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
	if payload.NivelAcceso == "" {
		payload.NivelAcceso = "acceso_total"
	}
	if payload.Estado == "" {
		payload.Estado = "activo"
	}
	if payload.EmpresaID <= 0 || payload.AdminEmail == "" {
		return 0, fmt.Errorf("empresa_id y admin_email son obligatorios")
	}
	existing, err := GetActiveAdminEmpresaCompartidaAcceso(dbConn, payload.EmpresaID, payload.AdminEmail)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		_, err = execSQLCompat(dbConn, `UPDATE admin_empresa_compartida
		SET compartido_por_email = ?,
		    invitacion_id = ?,
		    nivel_acceso = ?,
		    modulos_permitidos = ?,
		    fecha_aceptada = ?,
		    fecha_revocada = '',
		    fecha_actualizacion = `+sqlNowExpr()+`,
		    usuario_creador = ?,
		    estado = 'activo',
		    observaciones = ?
		WHERE id = ?`, payload.CompartidoPorEmail, nullableInt64Arg(payload.InvitacionID), payload.NivelAcceso, payload.ModulosPermitidos, strings.TrimSpace(payload.FechaAceptada), payload.UsuarioCreador, payload.Observaciones, existing.ID)
		if err != nil {
			return 0, err
		}
		InvalidateAdminEmpresaCompartidaAccessCache(payload.EmpresaID, payload.AdminEmail)
		InvalidateCanAdminAccessEmpresaIACache(payload.EmpresaID, payload.AdminEmail)
		InvalidateCanAdminAccessEmpresaIACache(payload.EmpresaID, payload.CompartidoPorEmail)
		return existing.ID, nil
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO admin_empresa_compartida (
		empresa_id,
		admin_email,
		compartido_por_email,
		invitacion_id,
		nivel_acceso,
		modulos_permitidos,
		fecha_aceptada,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+sqlNowExpr()+`, `+sqlNowExpr()+`)`,
		payload.EmpresaID,
		payload.AdminEmail,
		payload.CompartidoPorEmail,
		nullableInt64Arg(payload.InvitacionID),
		payload.NivelAcceso,
		payload.ModulosPermitidos,
		strings.TrimSpace(payload.FechaAceptada),
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	InvalidateAdminEmpresaCompartidaAccessCache(payload.EmpresaID, payload.AdminEmail)
	InvalidateCanAdminAccessEmpresaIACache(payload.EmpresaID, payload.AdminEmail)
	InvalidateCanAdminAccessEmpresaIACache(payload.EmpresaID, payload.CompartidoPorEmail)
	return id, nil
}

func GetAdminEmpresaCompartidaAccesoByID(dbConn *sql.DB, id int64) (*AdminEmpresaCompartidaAcceso, error) {
	if dbConn == nil || id <= 0 {
		return nil, nil
	}
	if err := EnsureAdminEmpresaCompartidaSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		a.id,
		a.empresa_id,
		COALESCE(a.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(a.compartido_por_email, ''),
		COALESCE(inv.name, ''),
		COALESCE(a.invitacion_id, 0),
		COALESCE(a.nivel_acceso, 'acceso_total'),
		COALESCE(a.modulos_permitidos, ''),
		COALESCE(a.fecha_aceptada, ''),
		COALESCE(a.fecha_revocada, ''),
		COALESCE(a.fecha_creacion, ''),
		COALESCE(a.fecha_actualizacion, ''),
		COALESCE(a.usuario_creador, ''),
		COALESCE(a.estado, 'activo'),
		COALESCE(a.observaciones, '')
	FROM admin_empresa_compartida a
	LEFT JOIN administradores adm ON lower(adm.email) = lower(a.admin_email)
	LEFT JOIN administradores inv ON lower(inv.email) = lower(a.compartido_por_email)
	WHERE a.id = ?
	LIMIT 1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	item, scanErr := scanAdminEmpresaCompartidaAcceso(rows)
	if scanErr != nil {
		return nil, scanErr
	}
	return &item, nil
}

func RevokeAdminEmpresaCompartidaAcceso(dbConn *sql.DB, id int64, usuario string) error {
	if dbConn == nil || id <= 0 {
		return fmt.Errorf("id invalido")
	}
	existing, _ := GetAdminEmpresaCompartidaAccesoByID(dbConn, id)
	_, err := execSQLCompat(dbConn, `UPDATE admin_empresa_compartida
	SET estado = 'revocada',
	    fecha_revocada = ?,
	    fecha_actualizacion = `+sqlNowExpr()+`,
	    usuario_creador = ?
	WHERE id = ?`, sqlNowValue(), strings.TrimSpace(usuario), id)
	if err == nil && existing != nil {
		InvalidateAdminEmpresaCompartidaAccessCache(existing.EmpresaID, existing.AdminEmail)
		InvalidateCanAdminAccessEmpresaIACache(existing.EmpresaID, existing.AdminEmail)
		InvalidateCanAdminAccessEmpresaIACache(existing.EmpresaID, existing.CompartidoPorEmail)
	}
	return err
}

func nullableInt64Arg(value int64) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func sqlNowValue() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
