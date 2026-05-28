package db

import "database/sql"

const DefaultUsuarioApariencia = "light"

// EnsureUsuarioConfiguracionSchema crea la tabla para preferencias por usuario (asociada por email)
func EnsureUsuarioConfiguracionSchema(dbConn *sql.DB) error {
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
