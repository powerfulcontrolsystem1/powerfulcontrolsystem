package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// ResolveRolDeUsuarioIDByNombreEmpresaScope selects the administrative template
// for an active company's business type. A type-specific template takes priority
// over type zero; templates for other business types never participate.
// A missing applicable template returns zero so the caller may use its standard
// administrative policy. An inactive or ambiguous applicable template is an
// authorization error, never a reason to discard its restrictions.
func ResolveRolDeUsuarioIDByNombreEmpresaScope(dbConn *sql.DB, empresaID int64, nombreRol string) (int64, error) {
	nombreRol = strings.ToLower(strings.TrimSpace(nombreRol))
	if dbConn == nil || empresaID <= 0 || nombreRol == "" {
		return 0, fmt.Errorf("contexto de plantilla de rol no disponible")
	}
	var tipoID int64
	if err := queryRowSQLCompat(dbConn, `SELECT COALESCE(tipo_id, 0) FROM empresas
		WHERE COALESCE(empresa_id, id) = ? AND lower(trim(COALESCE(estado, 'activo'))) = 'activo'`, empresaID).Scan(&tipoID); err != nil {
		return 0, err
	}
	rows, err := dbConn.Query(`SELECT id, COALESCE(tipo_empresa_id, 0), lower(trim(COALESCE(estado, '')))
		FROM roles_de_usuario WHERE COALESCE(empresa_id, 0) = 0
		AND lower(trim(nombre)) = ? AND COALESCE(tipo_empresa_id, 0) IN (0, ?)
		ORDER BY CASE WHEN COALESCE(tipo_empresa_id, 0) = ? THEN 0 ELSE 1 END, id`, nombreRol, tipoID, tipoID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var selectedID, selectedType int64
	var applicable, active int
	for rows.Next() {
		var id, candidateType int64
		var state string
		if err := rows.Scan(&id, &candidateType, &state); err != nil {
			return 0, err
		}
		if applicable == 0 {
			selectedType = candidateType
		}
		if candidateType != selectedType {
			continue
		}
		applicable++
		if state == "activo" {
			selectedID = id
			active++
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if applicable == 0 {
		return 0, nil
	}
	if active != 1 || selectedID <= 0 {
		return 0, fmt.Errorf("plantilla de rol %q para tipo %d inactiva o ambigua; revise la configuracion", nombreRol, selectedType)
	}
	return selectedID, nil
}
