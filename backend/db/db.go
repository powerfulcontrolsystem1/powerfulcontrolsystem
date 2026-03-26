package db

import "database/sql"

// UpsertUser inserta o actualiza un usuario por email
func UpsertUser(dbConn *sql.DB, email, name string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT OR IGNORE INTO users (email, name) VALUES (?, ?)", email, name); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE users SET name = ? WHERE email = ?", name, email); err != nil {
		return err
	}
	return tx.Commit()
}

// EnsureUserEmpresa crea una empresa por defecto para el usuario si no tiene una asociada
func EnsureUserEmpresa(dbConn *sql.DB, email, empresaNombre string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64
	var empresaID sql.NullInt64
	row := tx.QueryRow("SELECT id, empresa_id FROM users WHERE email = ?", email)
	if err := row.Scan(&userID, &empresaID); err != nil {
		return err
	}

	if empresaID.Valid {
		// ya tiene empresa asociada
		return tx.Commit()
	}

	res, err := tx.Exec("INSERT INTO empresas (nombre, usuario_creador) VALUES (?, ?)", empresaNombre, email)
	if err != nil {
		return err
	}
	newEmpresaID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("UPDATE users SET empresa_id = ? WHERE id = ?", newEmpresaID, userID); err != nil {
		return err
	}

	return tx.Commit()
}
