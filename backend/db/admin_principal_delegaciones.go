package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type AdminPrincipalDelegacion struct {
	ID                 int64  `json:"id"`
	AdminEmail         string `json:"admin_email"`
	AdminName          string `json:"admin_name,omitempty"`
	PrincipalEmail     string `json:"principal_email"`
	PrincipalName      string `json:"principal_name,omitempty"`
	InvitadoPorEmail   string `json:"invitado_por_email,omitempty"`
	TokenHash          string `json:"-"`
	ExpiraEn           string `json:"expira_en,omitempty"`
	FechaAceptada      string `json:"fecha_aceptada,omitempty"`
	FechaRevocada      string `json:"fecha_revocada,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

func normalizeAdminPrincipalDelegacionEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func EnsureAdminPrincipalDelegacionesSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS admin_principal_delegaciones (
			id BIGSERIAL PRIMARY KEY,
			admin_email TEXT NOT NULL,
			principal_email TEXT NOT NULL,
			invitado_por_email TEXT,
			token_hash TEXT,
			expira_en TEXT,
			fecha_aceptada TEXT,
			fecha_revocada TEXT,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'pendiente',
			observaciones TEXT
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_admin_principal_delegaciones_admin_principal ON admin_principal_delegaciones (admin_email, principal_email)`,
		`CREATE INDEX IF NOT EXISTS ix_admin_principal_delegaciones_admin_estado ON admin_principal_delegaciones (admin_email, estado)`,
		`CREATE INDEX IF NOT EXISTS ix_admin_principal_delegaciones_principal_estado ON admin_principal_delegaciones (principal_email, estado)`,
		`CREATE INDEX IF NOT EXISTS ix_admin_principal_delegaciones_token ON admin_principal_delegaciones (token_hash)`,
	}
	for _, stmt := range statements {
		if _, err := ExecCompat(dbConn, stmt); err != nil {
			return fmt.Errorf("ensure admin principal delegaciones schema: %w; stmt=%s", err, stmt)
		}
	}
	for _, col := range []struct {
		name string
		def  string
	}{
		{"invitado_por_email", "TEXT"},
		{"token_hash", "TEXT"},
		{"expira_en", "TEXT"},
		{"fecha_aceptada", "TEXT"},
		{"fecha_revocada", "TEXT"},
		{"fecha_actualizacion", "TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'pendiente'"},
		{"observaciones", "TEXT"},
	} {
		if err := ensureColumnIfMissing(dbConn, "admin_principal_delegaciones", col.name, col.def); err != nil {
			return fmt.Errorf("ensure admin principal delegaciones column %s: %w", col.name, err)
		}
	}
	return nil
}

func scanAdminPrincipalDelegacion(rows *sql.Rows) (AdminPrincipalDelegacion, error) {
	var item AdminPrincipalDelegacion
	if err := rows.Scan(
		&item.ID,
		&item.AdminEmail,
		&item.AdminName,
		&item.PrincipalEmail,
		&item.PrincipalName,
		&item.InvitadoPorEmail,
		&item.TokenHash,
		&item.ExpiraEn,
		&item.FechaAceptada,
		&item.FechaRevocada,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return AdminPrincipalDelegacion{}, err
	}
	return item, nil
}

func UpsertAdminPrincipalDelegacionInvitacion(dbConn *sql.DB, adminEmail, principalEmail, invitedByEmail, tokenHash, expiraEn string) (int64, error) {
	adminEmail = normalizeAdminPrincipalDelegacionEmail(adminEmail)
	principalEmail = normalizeAdminPrincipalDelegacionEmail(principalEmail)
	invitedByEmail = normalizeAdminPrincipalDelegacionEmail(invitedByEmail)
	tokenHash = strings.TrimSpace(tokenHash)
	expiraEn = strings.TrimSpace(expiraEn)
	if dbConn == nil || adminEmail == "" || principalEmail == "" || tokenHash == "" {
		return 0, fmt.Errorf("delegacion principal invalida")
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return 0, err
	}
	_, err := ExecCompat(dbConn, `INSERT INTO admin_principal_delegaciones
		(admin_email, principal_email, invitado_por_email, token_hash, expira_en, fecha_aceptada, fecha_revocada, estado, fecha_creacion, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, ?, ?, '', '', 'pendiente', CAST(CURRENT_TIMESTAMP AS TEXT), CAST(CURRENT_TIMESTAMP AS TEXT), ?)
		ON CONFLICT (admin_email, principal_email) DO UPDATE SET
			invitado_por_email = EXCLUDED.invitado_por_email,
			token_hash = EXCLUDED.token_hash,
			expira_en = EXCLUDED.expira_en,
			fecha_aceptada = '',
			fecha_revocada = '',
			estado = 'pendiente',
			fecha_actualizacion = CAST(CURRENT_TIMESTAMP AS TEXT),
			usuario_creador = EXCLUDED.usuario_creador`,
		adminEmail, principalEmail, invitedByEmail, tokenHash, expiraEn, invitedByEmail)
	if err != nil {
		return 0, err
	}
	InvalidateCanAdminAccessEmpresaIAAdminCache(adminEmail)
	var id int64
	if err := QueryRowCompat(dbConn, `SELECT id FROM admin_principal_delegaciones WHERE admin_email = ? AND principal_email = ? LIMIT 1`, adminEmail, principalEmail).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func UpsertAdminPrincipalDelegacionActiva(dbConn *sql.DB, adminEmail, principalEmail, invitedByEmail string) (int64, error) {
	adminEmail = normalizeAdminPrincipalDelegacionEmail(adminEmail)
	principalEmail = normalizeAdminPrincipalDelegacionEmail(principalEmail)
	invitedByEmail = normalizeAdminPrincipalDelegacionEmail(invitedByEmail)
	if dbConn == nil || adminEmail == "" || principalEmail == "" {
		return 0, fmt.Errorf("delegacion principal invalida")
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return 0, err
	}
	_, err := ExecCompat(dbConn, `INSERT INTO admin_principal_delegaciones
		(admin_email, principal_email, invitado_por_email, token_hash, expira_en, fecha_aceptada, fecha_revocada, estado, fecha_creacion, fecha_actualizacion, usuario_creador)
		VALUES (?, ?, ?, '', '', CAST(CURRENT_TIMESTAMP AS TEXT), '', 'activo', CAST(CURRENT_TIMESTAMP AS TEXT), CAST(CURRENT_TIMESTAMP AS TEXT), ?)
		ON CONFLICT (admin_email, principal_email) DO UPDATE SET
			invitado_por_email = EXCLUDED.invitado_por_email,
			token_hash = '',
			expira_en = '',
			fecha_aceptada = CASE WHEN COALESCE(admin_principal_delegaciones.fecha_aceptada, '') = '' THEN CAST(CURRENT_TIMESTAMP AS TEXT) ELSE admin_principal_delegaciones.fecha_aceptada END,
			fecha_revocada = '',
			estado = 'activo',
			fecha_actualizacion = CAST(CURRENT_TIMESTAMP AS TEXT),
			usuario_creador = EXCLUDED.usuario_creador`,
		adminEmail, principalEmail, invitedByEmail, invitedByEmail)
	if err != nil {
		return 0, err
	}
	InvalidateCanAdminAccessEmpresaIAAdminCache(adminEmail)
	var id int64
	if err := QueryRowCompat(dbConn, `SELECT id FROM admin_principal_delegaciones WHERE admin_email = ? AND principal_email = ? LIMIT 1`, adminEmail, principalEmail).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func GetAdminPrincipalDelegacionByTokenHash(dbConn *sql.DB, tokenHash string) (*AdminPrincipalDelegacion, error) {
	tokenHash = strings.TrimSpace(tokenHash)
	if dbConn == nil || tokenHash == "" {
		return nil, nil
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT
		d.id,
		COALESCE(d.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(d.principal_email, ''),
		COALESCE(principal.name, ''),
		COALESCE(d.invitado_por_email, ''),
		COALESCE(d.token_hash, ''),
		COALESCE(d.expira_en, ''),
		COALESCE(d.fecha_aceptada, ''),
		COALESCE(d.fecha_revocada, ''),
		COALESCE(d.fecha_creacion, ''),
		COALESCE(d.fecha_actualizacion, ''),
		COALESCE(d.usuario_creador, ''),
		COALESCE(d.estado, 'pendiente'),
		COALESCE(d.observaciones, '')
	FROM admin_principal_delegaciones d
	LEFT JOIN administradores adm ON lower(adm.email) = lower(d.admin_email)
	LEFT JOIN administradores principal ON lower(principal.email) = lower(d.principal_email)
	WHERE d.token_hash = ?
	ORDER BY d.id DESC LIMIT 1`, tokenHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	item, err := scanAdminPrincipalDelegacion(rows)
	if err != nil {
		return nil, err
	}
	return &item, rows.Err()
}

func MarkAdminPrincipalDelegacionAccepted(dbConn *sql.DB, id int64, acceptedAt, actorEmail string) error {
	if dbConn == nil || id <= 0 {
		return nil
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return err
	}
	_, err := ExecCompat(dbConn, `UPDATE admin_principal_delegaciones
		SET estado = 'activo',
			fecha_aceptada = ?,
			token_hash = '',
			fecha_actualizacion = CAST(CURRENT_TIMESTAMP AS TEXT),
			usuario_creador = ?
		WHERE id = ?`, strings.TrimSpace(acceptedAt), normalizeAdminPrincipalDelegacionEmail(actorEmail), id)
	if err == nil {
		InvalidateCanAdminAccessEmpresaIAAdminCache(actorEmail)
	}
	return err
}

func SetAdminPrincipalDelegacionEstado(dbConn *sql.DB, id int64, estado, actorEmail string) error {
	if dbConn == nil || id <= 0 {
		return nil
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return err
	}
	estado = strings.ToLower(strings.TrimSpace(estado))
	if estado == "" {
		estado = "pendiente"
	}
	fechaRevocada := ""
	if estado == "revocada" || estado == "rechazada" || estado == "expirada" {
		fechaRevocada = "CURRENT_TIMESTAMP"
	}
	if fechaRevocada == "" {
		_, err := ExecCompat(dbConn, `UPDATE admin_principal_delegaciones SET estado = ?, fecha_actualizacion = CAST(CURRENT_TIMESTAMP AS TEXT), usuario_creador = ? WHERE id = ?`,
			estado, normalizeAdminPrincipalDelegacionEmail(actorEmail), id)
		if err == nil {
			InvalidateCanAdminAccessEmpresaIAAdminCache(actorEmail)
		}
		return err
	}
	_, err := ExecCompat(dbConn, `UPDATE admin_principal_delegaciones SET estado = ?, fecha_revocada = CAST(CURRENT_TIMESTAMP AS TEXT), fecha_actualizacion = CAST(CURRENT_TIMESTAMP AS TEXT), usuario_creador = ? WHERE id = ?`,
		estado, normalizeAdminPrincipalDelegacionEmail(actorEmail), id)
	if err == nil {
		InvalidateCanAdminAccessEmpresaIAAdminCache(actorEmail)
	}
	return err
}

func RevokeAdminPrincipalDelegacion(dbConn *sql.DB, principalEmail, adminEmail, actorEmail string) error {
	principalEmail = normalizeAdminPrincipalDelegacionEmail(principalEmail)
	adminEmail = normalizeAdminPrincipalDelegacionEmail(adminEmail)
	if dbConn == nil || principalEmail == "" || adminEmail == "" {
		return nil
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return err
	}
	_, err := ExecCompat(dbConn, `UPDATE admin_principal_delegaciones
		SET estado = 'revocada',
			fecha_revocada = CAST(CURRENT_TIMESTAMP AS TEXT),
			token_hash = '',
			fecha_actualizacion = CAST(CURRENT_TIMESTAMP AS TEXT),
			usuario_creador = ?
		WHERE admin_email = ? AND principal_email = ? AND lower(COALESCE(estado, 'pendiente')) <> 'revocada'`,
		normalizeAdminPrincipalDelegacionEmail(actorEmail), adminEmail, principalEmail)
	if err == nil {
		InvalidateCanAdminAccessEmpresaIAAdminCache(adminEmail)
	}
	return err
}

func ListActiveAdminPrincipalDelegacionPrincipals(dbConn *sql.DB, adminEmail string) ([]string, error) {
	adminEmail = normalizeAdminPrincipalDelegacionEmail(adminEmail)
	if dbConn == nil || adminEmail == "" {
		return []string{}, nil
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT DISTINCT COALESCE(principal_email, '')
		FROM admin_principal_delegaciones
		WHERE admin_email = ?
		  AND lower(COALESCE(estado, 'pendiente')) = 'activo'
		  AND COALESCE(fecha_revocada, '') = ''`, adminEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		email = normalizeAdminPrincipalDelegacionEmail(email)
		if email != "" {
			out = append(out, email)
		}
	}
	return out, rows.Err()
}

func ListAdminPrincipalDelegacionesByPrincipal(dbConn *sql.DB, principalEmail string) ([]AdminPrincipalDelegacion, error) {
	principalEmail = normalizeAdminPrincipalDelegacionEmail(principalEmail)
	if dbConn == nil || principalEmail == "" {
		return []AdminPrincipalDelegacion{}, nil
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT
		d.id,
		COALESCE(d.admin_email, ''),
		COALESCE(adm.name, ''),
		COALESCE(d.principal_email, ''),
		COALESCE(principal.name, ''),
		COALESCE(d.invitado_por_email, ''),
		COALESCE(d.token_hash, ''),
		COALESCE(d.expira_en, ''),
		COALESCE(d.fecha_aceptada, ''),
		COALESCE(d.fecha_revocada, ''),
		COALESCE(d.fecha_creacion, ''),
		COALESCE(d.fecha_actualizacion, ''),
		COALESCE(d.usuario_creador, ''),
		COALESCE(d.estado, 'pendiente'),
		COALESCE(d.observaciones, '')
	FROM admin_principal_delegaciones d
	LEFT JOIN administradores adm ON lower(adm.email) = lower(d.admin_email)
	LEFT JOIN administradores principal ON lower(principal.email) = lower(d.principal_email)
	WHERE d.principal_email = ?
	  AND lower(COALESCE(d.estado, 'pendiente')) IN ('pendiente', 'activo')
	  AND COALESCE(d.fecha_revocada, '') = ''
	ORDER BY d.id DESC`, principalEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AdminPrincipalDelegacion{}
	for rows.Next() {
		item, scanErr := scanAdminPrincipalDelegacion(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func AdminHasPrincipalDelegacionInScope(dbConn *sql.DB, principalEmail, adminEmail string) (bool, error) {
	principalEmail = normalizeAdminPrincipalDelegacionEmail(principalEmail)
	adminEmail = normalizeAdminPrincipalDelegacionEmail(adminEmail)
	if dbConn == nil || principalEmail == "" || adminEmail == "" {
		return false, nil
	}
	if err := EnsureAdminPrincipalDelegacionesSchema(dbConn); err != nil {
		return false, err
	}
	var total int
	if err := QueryRowCompat(dbConn, `SELECT COUNT(1)
		FROM admin_principal_delegaciones
		WHERE principal_email = ?
		  AND admin_email = ?
		  AND lower(COALESCE(estado, 'pendiente')) IN ('pendiente', 'activo')
		  AND COALESCE(fecha_revocada, '') = ''`, principalEmail, adminEmail).Scan(&total); err != nil {
		return false, err
	}
	return total > 0, nil
}
