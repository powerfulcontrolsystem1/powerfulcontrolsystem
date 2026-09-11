package db

import (
	"database/sql"
	"fmt"
)

// GetEmpresaUsuarioRoleIDs reads only distinct assignments, never user profiles
// or credentials, so catalogue construction does not grow with the user count.
func GetEmpresaUsuarioRoleIDs(conn *sql.DB, empresaID int64) (map[int64]bool, error) {
	if conn == nil || empresaID <= 0 {
		return nil, fmt.Errorf("company required")
	}
	rows, err := querySQLCompat(conn, `SELECT DISTINCT rol_usuario_id FROM users
		WHERE empresa_id = ? AND rol_usuario_id > 0`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make(map[int64]bool)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids[id] = true
	}
	return ids, rows.Err()
}
