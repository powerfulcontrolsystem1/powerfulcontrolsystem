package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// RevokeEmpresaUsuarioSessions only revokes the operational principal in its
// company. The same email may own independent administrative and tenant sessions.
func RevokeEmpresaUsuarioSessions(dbConn *sql.DB, empresaID, usuarioID int64, email string) error {
	if dbConn == nil || empresaID <= 0 || usuarioID <= 0 || strings.TrimSpace(email) == "" {
		return fmt.Errorf("company, user and email required to revoke sessions")
	}
	_, err := execSQLCompat(dbConn, `UPDATE sesiones SET activo = 0, fecha_fin = `+sqlNowExpr()+`
		WHERE principal_type = 'empresa_usuario' AND empresa_id = ? AND principal_id = ?
		AND LOWER(TRIM(COALESCE(admin_email, ''))) = LOWER(?) AND activo = 1`,
		empresaID, usuarioID, strings.TrimSpace(email))
	return err
}
