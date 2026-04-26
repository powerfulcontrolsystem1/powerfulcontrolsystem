package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type EmpresaNextcloudAccount struct {
	ID               int64  `json:"id"`
	EmpresaID         int64  `json:"empresa_id"`
	NextcloudUser     string `json:"nextcloud_user"`
	PasswordEncrypted string `json:"password_encrypted,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
	CreatedAt         string `json:"created_at,omitempty"`
}

func EnsureEmpresaNextcloudSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_nextcloud_accounts (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL UNIQUE,
			nextcloud_user TEXT NOT NULL,
			password_encrypted TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_nextcloud_accounts_empresa ON empresa_nextcloud_accounts(empresa_id);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	// Backfill/migration support (older installs)
	if err := ensureColumnIfMissing(dbConn, "empresa_nextcloud_accounts", "password_encrypted", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nextcloud_accounts", "created_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_nextcloud_accounts", "updated_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP"); err != nil {
		return err
	}
	return nil
}

func normalizeNextcloudUser(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	v = strings.ToLower(v)
	// Nextcloud user: keep simple safe charset
	out := make([]rune, 0, len(v))
	for _, r := range v {
		switch {
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r >= '0' && r <= '9':
			out = append(out, r)
		case r == '_' || r == '-' || r == '.':
			out = append(out, r)
		}
	}
	return strings.Trim(outString(out), "._-")
}

func outString(rs []rune) string {
	return string(rs)
}

func GetEmpresaNextcloudAccount(dbConn *sql.DB, empresaID int64) (EmpresaNextcloudAccount, bool, error) {
	var out EmpresaNextcloudAccount
	if empresaID <= 0 {
		return out, false, fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaNextcloudSchema(dbConn); err != nil {
		return out, false, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, nextcloud_user, password_encrypted, created_at, updated_at
		FROM empresa_nextcloud_accounts WHERE empresa_id = ?`, empresaID)
	var created, updated time.Time
	if err := row.Scan(&out.ID, &out.EmpresaID, &out.NextcloudUser, &out.PasswordEncrypted, &created, &updated); err != nil {
		if err == sql.ErrNoRows {
			return EmpresaNextcloudAccount{}, false, nil
		}
		return out, false, err
	}
	out.CreatedAt = created.Format("2006-01-02 15:04:05")
	out.UpdatedAt = updated.Format("2006-01-02 15:04:05")
	return out, true, nil
}

func UpsertEmpresaNextcloudAccount(dbConn *sql.DB, empresaID int64, nextcloudUser string, passwordEncrypted string) (EmpresaNextcloudAccount, error) {
	var out EmpresaNextcloudAccount
	if empresaID <= 0 {
		return out, fmt.Errorf("empresa_id invalido")
	}
	if err := EnsureEmpresaNextcloudSchema(dbConn); err != nil {
		return out, err
	}
	nextcloudUser = normalizeNextcloudUser(nextcloudUser)
	if nextcloudUser == "" {
		return out, fmt.Errorf("nextcloud_user invalido")
	}
	passwordEncrypted = strings.TrimSpace(passwordEncrypted)
	if passwordEncrypted == "" {
		return out, fmt.Errorf("password_encrypted invalido")
	}

	// Upsert compat
	_, _ = execSQLCompat(dbConn, `INSERT INTO empresa_nextcloud_accounts (empresa_id, nextcloud_user, password_encrypted, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (empresa_id) DO UPDATE SET nextcloud_user = EXCLUDED.nextcloud_user, password_encrypted = EXCLUDED.password_encrypted, updated_at = CURRENT_TIMESTAMP`,
		empresaID, nextcloudUser, passwordEncrypted)

	acc, _, err := GetEmpresaNextcloudAccount(dbConn, empresaID)
	return acc, err
}

