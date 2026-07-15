package db

import (
	"database/sql"
	"encoding/json"
	"strings"
)

const DefaultUsuarioApariencia = "light"

// EnsureUsuarioConfiguracionSchema crea la tabla para preferencias por usuario (asociada por email)
func EnsureUsuarioConfiguracionSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	q := `CREATE TABLE IF NOT EXISTS usuario_configuracion (
		email TEXT PRIMARY KEY,
		apariencia TEXT DEFAULT 'light',
		fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP
	);`
	if isPostgresDialect() {
		q = `CREATE TABLE IF NOT EXISTS usuario_configuracion (
			email TEXT PRIMARY KEY,
			apariencia TEXT DEFAULT 'light',
			fecha_actualizacion TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`
	}
	if _, err := execSQLCompat(dbConn, q); err != nil {
		return err
	}
	if isPostgresDialect() {
		if _, err := execSQLCompat(dbConn, `ALTER TABLE usuario_configuracion ADD COLUMN IF NOT EXISTS selector_empresas_orden_json TEXT DEFAULT '[]'`); err != nil {
			return err
		}
	} else {
		_, _ = execSQLCompat(dbConn, `ALTER TABLE usuario_configuracion ADD COLUMN selector_empresas_orden_json TEXT DEFAULT '[]'`)
	}
	if isPostgresDialect() {
		_, err := execSQLCompat(dbConn, `ALTER TABLE usuario_configuracion ALTER COLUMN apariencia SET DEFAULT 'light'`)
		return err
	}
	return nil
}

func GetUsuarioApariencia(dbConn *sql.DB, email string) (string, error) {
	var ap string
	err := dbConn.QueryRow("SELECT apariencia FROM usuario_configuracion WHERE email = $1", email).Scan(&ap)
	if err == sql.ErrNoRows {
		return DefaultUsuarioApariencia, nil
	}
	if err != nil {
		return "", err
	}
	if ap == "" {
		return DefaultUsuarioApariencia, nil
	}
	return ap, nil
}
func SetUsuarioApariencia(dbConn *sql.DB, email, apariencia string) error {
	q := `INSERT INTO usuario_configuracion (email, apariencia, fecha_actualizacion) 
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT(email) DO UPDATE SET apariencia = $2, fecha_actualizacion = CURRENT_TIMESTAMP`
	if isPostgresDialect() {
		q = `INSERT INTO usuario_configuracion (email, apariencia, fecha_actualizacion) 
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT(email) DO UPDATE SET apariencia = EXCLUDED.apariencia, fecha_actualizacion = CURRENT_TIMESTAMP`
	}
	_, err := execSQLCompat(dbConn, q, email, apariencia)
	return err
}

func GetUsuarioSelectorEmpresasOrden(dbConn *sql.DB, email string) ([]int64, error) {
	var raw string
	err := dbConn.QueryRow("SELECT COALESCE(selector_empresas_orden_json, '[]') FROM usuario_configuracion WHERE email = $1", email).Scan(&raw)
	if err == sql.ErrNoRows {
		return []int64{}, nil
	}
	if err != nil {
		return nil, err
	}
	var values []int64
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &values); err != nil {
		return []int64{}, nil
	}
	return normalizeUsuarioSelectorEmpresaIDs(values), nil
}

func SetUsuarioSelectorEmpresasOrden(dbConn *sql.DB, email string, empresaIDs []int64) error {
	ids := normalizeUsuarioSelectorEmpresaIDs(empresaIDs)
	if len(ids) > 500 {
		ids = ids[:500]
	}
	raw, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	q := `INSERT INTO usuario_configuracion (email, selector_empresas_orden_json, fecha_actualizacion)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT(email) DO UPDATE SET selector_empresas_orden_json = $2, fecha_actualizacion = CURRENT_TIMESTAMP`
	if isPostgresDialect() {
		q = `INSERT INTO usuario_configuracion (email, selector_empresas_orden_json, fecha_actualizacion)
			VALUES ($1, $2, CURRENT_TIMESTAMP)
			ON CONFLICT(email) DO UPDATE SET selector_empresas_orden_json = EXCLUDED.selector_empresas_orden_json, fecha_actualizacion = CURRENT_TIMESTAMP`
	}
	_, err = execSQLCompat(dbConn, q, email, string(raw))
	return err
}

func removeEmpresaIDFromSelectorOrderRaw(raw string, empresaID int64) (string, bool) {
	if empresaID <= 0 {
		return "[]", false
	}
	var values []int64
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &values); err != nil {
		return "[]", false
	}
	ids := normalizeUsuarioSelectorEmpresaIDs(values)
	filtered := make([]int64, 0, len(ids))
	changed := false
	for _, id := range ids {
		if id == empresaID {
			changed = true
			continue
		}
		filtered = append(filtered, id)
	}
	if !changed {
		return strings.TrimSpace(raw), false
	}
	if len(filtered) > 500 {
		filtered = filtered[:500]
	}
	encoded, err := json.Marshal(filtered)
	if err != nil {
		return "[]", true
	}
	return string(encoded), true
}

func RemoveEmpresaFromAllUsuarioSelectorEmpresasOrdenTx(tx *sql.Tx, empresaID int64) (int64, error) {
	if tx == nil || empresaID <= 0 {
		return 0, nil
	}
	rows, err := queryTxSQLCompat(tx, `SELECT email, COALESCE(selector_empresas_orden_json, '[]') FROM usuario_configuracion WHERE COALESCE(selector_empresas_orden_json, '') <> ''`)
	if err != nil {
		if isMissingTableError(err) || isMissingColumnError(err) {
			return 0, nil
		}
		return 0, err
	}
	defer rows.Close()

	type update struct {
		email string
		raw   string
	}
	updates := make([]update, 0)
	for rows.Next() {
		var email, raw string
		if err := rows.Scan(&email, &raw); err != nil {
			return 0, err
		}
		nextRaw, changed := removeEmpresaIDFromSelectorOrderRaw(raw, empresaID)
		if changed {
			updates = append(updates, update{email: strings.TrimSpace(email), raw: nextRaw})
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	var affected int64
	for _, item := range updates {
		if item.email == "" {
			continue
		}
		res, err := execTxSQLCompat(tx, `UPDATE usuario_configuracion SET selector_empresas_orden_json = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE email = ?`, item.raw, item.email)
		if err != nil {
			return affected, err
		}
		n, _ := res.RowsAffected()
		affected += n
	}
	return affected, nil
}

func normalizeUsuarioSelectorEmpresaIDs(values []int64) []int64 {
	seen := make(map[int64]bool, len(values))
	out := make([]int64, 0, len(values))
	for _, id := range values {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}
