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
