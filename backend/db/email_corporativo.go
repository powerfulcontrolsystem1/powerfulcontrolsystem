package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type EmpresaEmailCorporativo struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	EmpresaNombre      string `json:"empresa_nombre"`
	Email              string `json:"email"`
	LocalPart          string `json:"local_part"`
	Domain             string `json:"domain"`
	WebmailURL         string `json:"webmail_url"`
	EstadoProvision    string `json:"estado_provision"`
	ProvisionProvider  string `json:"provision_provider"`
	ProvisionAttempts  int    `json:"provision_attempts"`
	FechaProvision     string `json:"fecha_provision,omitempty"`
	UltimoError        string `json:"ultimo_error,omitempty"`
	InitialPasswordSet bool   `json:"initial_password_set"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

func EnsureEmpresaEmailCorporativoSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_email_corporativo (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			empresa_nombre TEXT,
			email TEXT NOT NULL,
			local_part TEXT,
			domain TEXT,
			webmail_url TEXT,
			estado_provision TEXT DEFAULT 'pendiente',
			provision_provider TEXT DEFAULT 'iredmail',
			provision_attempts INTEGER DEFAULT 0,
			fecha_provision TIMESTAMPTZ,
			ultimo_error TEXT,
			initial_password_enc TEXT,
			initial_password_encrypted INTEGER DEFAULT 1,
			fecha_creacion TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS empresa_nombre TEXT`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS local_part TEXT`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS domain TEXT`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS webmail_url TEXT`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS estado_provision TEXT DEFAULT 'pendiente'`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS provision_provider TEXT DEFAULT 'iredmail'`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS provision_attempts INTEGER DEFAULT 0`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS fecha_provision TIMESTAMPTZ`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS ultimo_error TEXT`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS initial_password_enc TEXT`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS initial_password_encrypted INTEGER DEFAULT 1`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS fecha_creacion TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS fecha_actualizacion TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'activo'`,
		`ALTER TABLE empresa_email_corporativo ADD COLUMN IF NOT EXISTS observaciones TEXT`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_email_corporativo_empresa ON empresa_email_corporativo(empresa_id) WHERE COALESCE(estado, 'activo') <> 'eliminado'`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_email_corporativo_email ON empresa_email_corporativo(lower(trim(email))) WHERE COALESCE(estado, 'activo') <> 'eliminado'`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_email_corporativo_estado ON empresa_email_corporativo(estado_provision, estado)`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

func NormalizeCorporateEmailLocalPart(nombre string) string {
	value := strings.ToLower(strings.TrimSpace(nombre))
	replacements := map[string]string{
		"\u00e1": "a", "\u00e0": "a", "\u00e4": "a", "\u00e2": "a", "\u00e3": "a",
		"\u00e9": "e", "\u00e8": "e", "\u00eb": "e", "\u00ea": "e",
		"\u00ed": "i", "\u00ec": "i", "\u00ef": "i", "\u00ee": "i",
		"\u00f3": "o", "\u00f2": "o", "\u00f6": "o", "\u00f4": "o", "\u00f5": "o",
		"\u00fa": "u", "\u00f9": "u", "\u00fc": "u", "\u00fb": "u",
		"\u00f1": "n", "\u00e7": "c",
		"&": " y ",
	}
	for old, next := range replacements {
		value = strings.ReplaceAll(value, old, next)
	}
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, ".")
	value = strings.Trim(value, ".-_")
	value = regexp.MustCompile(`[.]{2,}`).ReplaceAllString(value, ".")
	if value == "" {
		value = "empresa"
	}
	if len(value) > 48 {
		value = strings.Trim(value[:48], ".-_")
	}
	return value
}

func ResolveUniqueCorporateEmail(dbConn *sql.DB, empresaID int64, empresaNombre, domain string) (string, string, error) {
	if err := EnsureEmpresaEmailCorporativoSchema(dbConn); err != nil {
		return "", "", err
	}
	domain = strings.ToLower(strings.TrimSpace(domain))
	if domain == "" {
		domain = "powerfulcontrolsystem.com"
	}
	base := NormalizeCorporateEmailLocalPart(empresaNombre)
	for i := 0; i < 250; i++ {
		local := base
		if i > 0 {
			local = fmt.Sprintf("%s%d", base, i+1)
		}
		email := local + "@" + domain
		var count int
		if err := queryRowSQLCompat(dbConn, `SELECT COUNT(1)
			FROM empresa_email_corporativo
			WHERE lower(trim(email)) = lower(trim(?))
				AND empresa_id <> ?
				AND COALESCE(estado, 'activo') <> 'eliminado'`, email, empresaID).Scan(&count); err != nil {
			return "", "", err
		}
		if count == 0 {
			return email, local, nil
		}
	}
	return "", "", fmt.Errorf("no se pudo generar un email unico para %q", empresaNombre)
}

func GetEmpresaEmailCorporativoByEmpresa(dbConn *sql.DB, empresaID int64) (*EmpresaEmailCorporativo, error) {
	if err := EnsureEmpresaEmailCorporativoSchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(empresa_nombre, ''), COALESCE(email, ''), COALESCE(local_part, ''), COALESCE(domain, ''), COALESCE(webmail_url, ''),
		COALESCE(estado_provision, 'pendiente'), COALESCE(provision_provider, 'iredmail'), COALESCE(provision_attempts, 0), COALESCE(CAST(fecha_provision AS TEXT), ''),
		COALESCE(ultimo_error, ''), CASE WHEN trim(COALESCE(initial_password_enc, '')) <> '' THEN 1 ELSE 0 END,
		COALESCE(CAST(fecha_creacion AS TEXT), ''), COALESCE(CAST(fecha_actualizacion AS TEXT), ''), COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'), COALESCE(observaciones, '')
		FROM empresa_email_corporativo
		WHERE empresa_id = ? AND COALESCE(estado, 'activo') <> 'eliminado'
		ORDER BY id DESC LIMIT 1`, empresaID)
	var item EmpresaEmailCorporativo
	var passSet int
	if err := row.Scan(&item.ID, &item.EmpresaID, &item.EmpresaNombre, &item.Email, &item.LocalPart, &item.Domain, &item.WebmailURL, &item.EstadoProvision, &item.ProvisionProvider, &item.ProvisionAttempts, &item.FechaProvision, &item.UltimoError, &passSet, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
		return nil, err
	}
	item.InitialPasswordSet = passSet == 1
	return &item, nil
}

func GetEmpresaEmailCorporativoInitialPasswordEncrypted(dbConn *sql.DB, empresaID int64) (string, error) {
	if err := EnsureEmpresaEmailCorporativoSchema(dbConn); err != nil {
		return "", err
	}
	var encrypted string
	err := queryRowSQLCompat(dbConn, `SELECT COALESCE(initial_password_enc, '')
		FROM empresa_email_corporativo
		WHERE empresa_id = ? AND COALESCE(estado, 'activo') <> 'eliminado'
		ORDER BY id DESC LIMIT 1`, empresaID).Scan(&encrypted)
	return encrypted, err
}

func ListEmpresaEmailCorporativo(dbConn *sql.DB) ([]EmpresaEmailCorporativo, error) {
	if err := EnsureEmpresaEmailCorporativoSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(empresa_nombre, ''), COALESCE(email, ''), COALESCE(local_part, ''), COALESCE(domain, ''), COALESCE(webmail_url, ''),
		COALESCE(estado_provision, 'pendiente'), COALESCE(provision_provider, 'iredmail'), COALESCE(provision_attempts, 0), COALESCE(CAST(fecha_provision AS TEXT), ''),
		COALESCE(ultimo_error, ''), CASE WHEN trim(COALESCE(initial_password_enc, '')) <> '' THEN 1 ELSE 0 END,
		COALESCE(CAST(fecha_creacion AS TEXT), ''), COALESCE(CAST(fecha_actualizacion AS TEXT), ''), COALESCE(usuario_creador, ''), COALESCE(estado, 'activo'), COALESCE(observaciones, '')
		FROM empresa_email_corporativo
		WHERE COALESCE(estado, 'activo') <> 'eliminado'
		ORDER BY empresa_nombre, empresa_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmpresaEmailCorporativo
	for rows.Next() {
		var item EmpresaEmailCorporativo
		var passSet int
		if err := rows.Scan(&item.ID, &item.EmpresaID, &item.EmpresaNombre, &item.Email, &item.LocalPart, &item.Domain, &item.WebmailURL, &item.EstadoProvision, &item.ProvisionProvider, &item.ProvisionAttempts, &item.FechaProvision, &item.UltimoError, &passSet, &item.FechaCreacion, &item.FechaActualizacion, &item.UsuarioCreador, &item.Estado, &item.Observaciones); err != nil {
			return nil, err
		}
		item.InitialPasswordSet = passSet == 1
		out = append(out, item)
	}
	return out, rows.Err()
}

func UpsertEmpresaEmailCorporativo(dbConn *sql.DB, item EmpresaEmailCorporativo, encryptedPassword string) (*EmpresaEmailCorporativo, error) {
	if item.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaEmailCorporativoSchema(dbConn); err != nil {
		return nil, err
	}
	item.Email = strings.ToLower(strings.TrimSpace(item.Email))
	item.Domain = strings.ToLower(strings.TrimSpace(item.Domain))
	item.LocalPart = strings.ToLower(strings.TrimSpace(item.LocalPart))
	item.EstadoProvision = strings.TrimSpace(item.EstadoProvision)
	if item.EstadoProvision == "" {
		item.EstadoProvision = "pendiente"
	}
	if item.ProvisionProvider == "" {
		item.ProvisionProvider = "iredmail"
	}
	nowExpr := sqlNowExpr()
	if encryptedPassword != "" {
		_, err := execSQLCompat(dbConn, `INSERT INTO empresa_email_corporativo
			(empresa_id, empresa_nombre, email, local_part, domain, webmail_url, estado_provision, provision_provider, initial_password_enc, initial_password_encrypted, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, 'activo', ?, `+nowExpr+`, `+nowExpr+`)
			ON CONFLICT (empresa_id) WHERE COALESCE(estado, 'activo') <> 'eliminado'
			DO UPDATE SET empresa_nombre = EXCLUDED.empresa_nombre, email = EXCLUDED.email, local_part = EXCLUDED.local_part, domain = EXCLUDED.domain,
				webmail_url = EXCLUDED.webmail_url, estado_provision = EXCLUDED.estado_provision, provision_provider = EXCLUDED.provision_provider,
				initial_password_enc = EXCLUDED.initial_password_enc, initial_password_encrypted = 1, usuario_creador = EXCLUDED.usuario_creador,
				observaciones = EXCLUDED.observaciones, fecha_actualizacion = `+nowExpr, item.EmpresaID, item.EmpresaNombre, item.Email, item.LocalPart, item.Domain, item.WebmailURL, item.EstadoProvision, item.ProvisionProvider, encryptedPassword, item.UsuarioCreador, item.Observaciones)
		if err != nil {
			return nil, err
		}
		return GetEmpresaEmailCorporativoByEmpresa(dbConn, item.EmpresaID)
	}
	_, err := execSQLCompat(dbConn, `INSERT INTO empresa_email_corporativo
		(empresa_id, empresa_nombre, email, local_part, domain, webmail_url, estado_provision, provision_provider, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'activo', ?, `+nowExpr+`, `+nowExpr+`)
		ON CONFLICT (empresa_id) WHERE COALESCE(estado, 'activo') <> 'eliminado'
		DO UPDATE SET empresa_nombre = EXCLUDED.empresa_nombre, email = EXCLUDED.email, local_part = EXCLUDED.local_part, domain = EXCLUDED.domain,
			webmail_url = EXCLUDED.webmail_url, estado_provision = EXCLUDED.estado_provision, provision_provider = EXCLUDED.provision_provider,
			usuario_creador = EXCLUDED.usuario_creador, observaciones = EXCLUDED.observaciones, fecha_actualizacion = `+nowExpr, item.EmpresaID, item.EmpresaNombre, item.Email, item.LocalPart, item.Domain, item.WebmailURL, item.EstadoProvision, item.ProvisionProvider, item.UsuarioCreador, item.Observaciones)
	if err != nil {
		return nil, err
	}
	return GetEmpresaEmailCorporativoByEmpresa(dbConn, item.EmpresaID)
}

func MarkEmpresaEmailProvisionResult(dbConn *sql.DB, empresaID int64, status, msg string, success bool) error {
	if err := EnsureEmpresaEmailCorporativoSchema(dbConn); err != nil {
		return err
	}
	status = strings.TrimSpace(status)
	if status == "" {
		if success {
			status = "provisionado"
		} else {
			status = "error"
		}
	}
	nowExpr := sqlNowExpr()
	fechaExpr := "fecha_provision"
	if success {
		fechaExpr = nowExpr
	}
	_, err := execSQLCompat(dbConn, `UPDATE empresa_email_corporativo
		SET estado_provision = ?, ultimo_error = ?, provision_attempts = COALESCE(provision_attempts, 0) + 1,
			fecha_provision = `+fechaExpr+`, fecha_actualizacion = `+nowExpr+`
		WHERE empresa_id = ? AND COALESCE(estado, 'activo') <> 'eliminado'`, status, strings.TrimSpace(msg), empresaID)
	return err
}

func EnsureEmpresaEmailRowsForExistingEmpresas(dbSuper, dbEmp *sql.DB, domain, webmailURL, usuario string) (int, error) {
	if err := EnsureEmpresaEmailCorporativoSchema(dbSuper); err != nil {
		return 0, err
	}
	empresas, err := GetEmpresas(dbEmp)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, empresa := range empresas {
		empresaID := empresa.EmpresaID
		if empresaID <= 0 {
			empresaID = empresa.ID
		}
		if empresaID <= 0 {
			continue
		}
		if _, err := GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID); err == nil {
			continue
		} else if err != sql.ErrNoRows {
			return total, err
		}
		email, local, err := ResolveUniqueCorporateEmail(dbSuper, empresaID, empresa.Nombre, domain)
		if err != nil {
			return total, err
		}
		_, err = UpsertEmpresaEmailCorporativo(dbSuper, EmpresaEmailCorporativo{
			EmpresaID:         empresaID,
			EmpresaNombre:     empresa.Nombre,
			Email:             email,
			LocalPart:         local,
			Domain:            strings.ToLower(strings.TrimSpace(domain)),
			WebmailURL:        strings.TrimSpace(webmailURL),
			EstadoProvision:   "pendiente",
			ProvisionProvider: "iredmail",
			UsuarioCreador:    usuario,
			Observaciones:     "Generado por sincronizacion de email corporativo",
		}, "")
		if err != nil {
			return total, err
		}
		total++
	}
	return total, nil
}

func CorporateEmailNowLabel() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
