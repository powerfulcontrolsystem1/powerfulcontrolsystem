package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type EmpresaDeleteTableResult struct {
	Scope   string `json:"scope"`
	Table   string `json:"table"`
	Deleted int64  `json:"deleted"`
}

type EmpresaDeleteCascadeResult struct {
	EmpresaID           int64                      `json:"empresa_id"`
	EmpresaNombre       string                     `json:"empresa_nombre,omitempty"`
	TablasAfectadas     int                        `json:"tablas_afectadas"`
	RegistrosEliminados int64                      `json:"registros_eliminados"`
	Detalle             []EmpresaDeleteTableResult `json:"detalle,omitempty"`
}

type empresaDeleteCandidateTable struct {
	Name              string
	EmpresaIDDataType string
	EmpresaIDUDTName  string
}

func empresaDeleteListCandidateTables(dbConn *sql.DB, excludeTables map[string]bool) ([]empresaDeleteCandidateTable, error) {
	if dbConn == nil {
		return nil, nil
	}

	tablesQuery := `
		SELECT t.table_name AS name,
		       COALESCE(c.data_type, '') AS empresa_id_data_type,
		       COALESCE(c.udt_name, '') AS empresa_id_udt_name
		FROM information_schema.tables t
		INNER JOIN information_schema.columns c
		   ON c.table_schema = t.table_schema
		  AND c.table_name = t.table_name
		  AND c.column_name = 'empresa_id'
		WHERE t.table_schema = ANY (current_schemas(false))
		  AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name
	`

	rows, err := dbConn.Query(tablesQuery)
	if err != nil {
		return nil, err
	}
	out := make([]empresaDeleteCandidateTable, 0)
	for rows.Next() {
		var item empresaDeleteCandidateTable
		if err := rows.Scan(&item.Name, &item.EmpresaIDDataType, &item.EmpresaIDUDTName); err != nil {
			rows.Close()
			return nil, err
		}
		table := strings.ToLower(strings.TrimSpace(item.Name))
		if table == "" || !isSafeSQLIdentifier(table) {
			continue
		}
		if excludeTables != nil && excludeTables[table] {
			continue
		}
		item.Name = table
		item.EmpresaIDDataType = strings.ToLower(strings.TrimSpace(item.EmpresaIDDataType))
		item.EmpresaIDUDTName = strings.ToLower(strings.TrimSpace(item.EmpresaIDUDTName))
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	return out, nil
}

func empresaDeleteEmpresaIDIsTextColumn(table empresaDeleteCandidateTable) bool {
	dataType := strings.ToLower(strings.TrimSpace(table.EmpresaIDDataType))
	udtName := strings.ToLower(strings.TrimSpace(table.EmpresaIDUDTName))
	switch dataType {
	case "text", "character varying", "character", "varchar", "char", "citext":
		return true
	}
	switch udtName {
	case "text", "varchar", "bpchar", "char", "citext":
		return true
	}
	return false
}

func empresaDeleteBuildEmpresaIDPredicate(table empresaDeleteCandidateTable, empresaID int64) (string, interface{}) {
	if empresaDeleteEmpresaIDIsTextColumn(table) {
		return "TRIM(empresa_id) = ?", strconv.FormatInt(empresaID, 10)
	}
	return "empresa_id = ?", empresaID
}

func empresaDeleteRowsByEmpresa(tx *sql.Tx, scope string, empresaID int64, tables []empresaDeleteCandidateTable) ([]EmpresaDeleteTableResult, int64, error) {
	results := make([]EmpresaDeleteTableResult, 0, len(tables))
	var totalDeleted int64
	for _, candidate := range tables {
		table := strings.ToLower(strings.TrimSpace(candidate.Name))
		if table == "" || !isSafeSQLIdentifier(table) {
			continue
		}
		predicate, arg := empresaDeleteBuildEmpresaIDPredicate(candidate, empresaID)
		res, err := execTxSQLCompat(tx, "DELETE FROM "+table+" WHERE "+predicate, arg)
		if err != nil {
			return nil, 0, fmt.Errorf("%s.%s delete empresa_id=%d: %w", scope, table, empresaID, err)
		}
		deleted, _ := res.RowsAffected()
		results = append(results, EmpresaDeleteTableResult{
			Scope:   scope,
			Table:   table,
			Deleted: deleted,
		})
		totalDeleted += deleted
	}
	return results, totalDeleted, nil
}

func DeleteEmpresaCascade(dbEmp, dbSuper *sql.DB, empresaID int64) (*EmpresaDeleteCascadeResult, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	if dbEmp == nil {
		return nil, fmt.Errorf("db empresas is nil")
	}

	empresa, err := GetEmpresaByID(dbEmp, empresaID)
	if err != nil {
		return nil, err
	}
	targetEmpresaIDs := []int64{empresaID}
	if empresa.EmpresaID > 0 && empresa.EmpresaID != empresaID {
		if _, lookupErr := GetEmpresaByID(dbEmp, empresa.EmpresaID); lookupErr == sql.ErrNoRows {
			targetEmpresaIDs = append(targetEmpresaIDs, empresa.EmpresaID)
		} else if lookupErr != nil {
			return nil, lookupErr
		}
	}

	tablesEmp, err := empresaDeleteListCandidateTables(dbEmp, map[string]bool{"empresas": true})
	if err != nil {
		return nil, err
	}
	tablesSuper, err := empresaDeleteListCandidateTables(dbSuper, nil)
	if err != nil {
		return nil, err
	}

	txEmp, err := dbEmp.Begin()
	if err != nil {
		return nil, err
	}
	defer txEmp.Rollback()

	var txSuper *sql.Tx
	if dbSuper != nil {
		txSuper, err = dbSuper.Begin()
		if err != nil {
			return nil, err
		}
		defer txSuper.Rollback()
	}

	detail := make([]EmpresaDeleteTableResult, 0, len(tablesEmp)+len(tablesSuper)+1)
	var totalDeleted int64

	if txSuper != nil {
		for _, targetEmpresaID := range targetEmpresaIDs {
			superResults, deletedSuper, err := empresaDeleteRowsByEmpresa(txSuper, "super", targetEmpresaID, tablesSuper)
			if err != nil {
				return nil, err
			}
			detail = append(detail, superResults...)
			totalDeleted += deletedSuper

			selectorRows, err := RemoveEmpresaFromAllUsuarioSelectorEmpresasOrdenTx(txSuper, targetEmpresaID)
			if err != nil {
				return nil, fmt.Errorf("super.usuario_configuracion limpiar orden selector empresa_id=%d: %w", targetEmpresaID, err)
			}
			if selectorRows > 0 {
				detail = append(detail, EmpresaDeleteTableResult{
					Scope:   "super",
					Table:   "usuario_configuracion.selector_empresas_orden_json",
					Deleted: selectorRows,
				})
			}
		}
	}

	for _, targetEmpresaID := range targetEmpresaIDs {
		empResults, deletedEmp, err := empresaDeleteRowsByEmpresa(txEmp, "operativa", targetEmpresaID, tablesEmp)
		if err != nil {
			return nil, err
		}
		detail = append(detail, empResults...)
		totalDeleted += deletedEmp
	}

	deleteEmpresaRes, err := execTxSQLCompat(txEmp, "DELETE FROM empresas WHERE id = ?", empresaID)
	if err != nil {
		return nil, err
	}
	deletedEmpresaRows, _ := deleteEmpresaRes.RowsAffected()
	detail = append(detail, EmpresaDeleteTableResult{
		Scope:   "operativa",
		Table:   "empresas",
		Deleted: deletedEmpresaRows,
	})
	totalDeleted += deletedEmpresaRows

	if txSuper != nil {
		if err := txSuper.Commit(); err != nil {
			return nil, err
		}
		txSuper = nil
	}
	if err := txEmp.Commit(); err != nil {
		return nil, err
	}
	InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
	InvalidateEmpresaByScopeCacheForEmpresa(empresaID, empresa.EmpresaID)
	InvalidateAdminEmpresaCompartidaAccessCacheForEmpresa(empresaID)
	if empresa.EmpresaID != empresaID {
		InvalidateAdminEmpresaCompartidaAccessCacheForEmpresa(empresa.EmpresaID)
	}

	return &EmpresaDeleteCascadeResult{
		EmpresaID:           empresaID,
		EmpresaNombre:       strings.TrimSpace(empresa.Nombre),
		TablasAfectadas:     len(detail),
		RegistrosEliminados: totalDeleted,
		Detalle:             detail,
	}, nil
}
