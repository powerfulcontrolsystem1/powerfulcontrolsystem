package db

import (
	"database/sql"
	"fmt"
)

// ResolveRolDeUsuarioIDByNombreEmpresaScope selects the administrative template
// for an active company's business type. A type-specific template takes priority
// over type zero; templates for other business types never participate.
// A missing applicable template returns zero so the caller may use its standard
// administrative policy. An inactive or ambiguous applicable template is an
// authorization error, never a reason to discard its restrictions.
func ResolveRolDeUsuarioIDByNombreEmpresaScope(dbEmp, dbSuper *sql.DB, empresaID int64, nombreRol string) (int64, error) {
	nombreRol = normalizeRolCatalogKey(nombreRol)
	if dbEmp == nil || dbSuper == nil || empresaID <= 0 || nombreRol == "" {
		return 0, fmt.Errorf("contexto de plantilla de rol no disponible")
	}
	var tipoID int64
	if err := queryRowSQLCompat(dbEmp, `SELECT COALESCE(tipo_id, 0) FROM empresas
		WHERE COALESCE(empresa_id, id) = ? AND lower(trim(COALESCE(estado, 'activo'))) = 'activo'`, empresaID).Scan(&tipoID); err != nil {
		return 0, err
	}
	// Normalize names with the same alias rules as the assignment catalog. The
	// query is limited to the two applicable types; it never loads other types
	// and does not perform a query per candidate or per permission.
	rows, err := dbSuper.Query(`SELECT id, COALESCE(tipo_empresa_id, 0), COALESCE(nombre, ''), lower(trim(COALESCE(estado, '')))
		FROM roles_de_usuario WHERE COALESCE(empresa_id, 0) = 0
		AND COALESCE(tipo_empresa_id, 0) IN (0, ?)
		ORDER BY CASE WHEN COALESCE(tipo_empresa_id, 0) = ? THEN 0 ELSE 1 END, id`, tipoID, tipoID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var selectedID, selectedType int64
	var applicable, active int
	for rows.Next() {
		var id, candidateType int64
		var name, state string
		if err := rows.Scan(&id, &candidateType, &name, &state); err != nil {
			return 0, err
		}
		if normalizeRolCatalogKey(name) != nombreRol {
			continue
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
