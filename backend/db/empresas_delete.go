package db

import (
	"database/sql"
	"fmt"
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

func empresaDeleteListCandidateTables(dbConn *sql.DB, excludeTables map[string]bool) ([]string, error) {
	if dbConn == nil {
		return nil, nil
	}

	tablesQuery := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	if isPostgresDialect() {
		tablesQuery = `
			SELECT table_name AS name
			FROM information_schema.tables
			WHERE table_schema = ANY (current_schemas(false))
			  AND table_type = 'BASE TABLE'
			ORDER BY table_name
		`
	}

	rows, err := dbConn.Query(tablesQuery)
	if err != nil {
		return nil, err
	}
	tableNames := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return nil, err
		}
		tableNames = append(tableNames, name)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	out := make([]string, 0)
	for _, name := range tableNames {
		table := strings.ToLower(strings.TrimSpace(name))
		if table == "" || !isSafeSQLIdentifier(table) {
			continue
		}
		if excludeTables != nil && excludeTables[table] {
			continue
		}
		cols, colErr := empresaBackupGetTableColumns(dbConn, table)
		if colErr != nil || !empresaBackupHasColumn(cols, "empresa_id") {
			continue
		}
		out = append(out, table)
	}
	return out, nil
}

func empresaDeleteRowsByEmpresa(tx *sql.Tx, scope string, empresaID int64, tables []string) ([]EmpresaDeleteTableResult, int64, error) {
	results := make([]EmpresaDeleteTableResult, 0, len(tables))
	var totalDeleted int64
	for _, table := range tables {
		if table == "" || !isSafeSQLIdentifier(table) {
			continue
		}
		res, err := execTxSQLCompat(tx, "DELETE FROM "+table+" WHERE empresa_id = ?", empresaID)
		if err != nil {
			return nil, 0, err
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
		superResults, deletedSuper, err := empresaDeleteRowsByEmpresa(txSuper, "super", empresaID, tablesSuper)
		if err != nil {
			return nil, err
		}
		detail = append(detail, superResults...)
		totalDeleted += deletedSuper
	}

	empResults, deletedEmp, err := empresaDeleteRowsByEmpresa(txEmp, "operativa", empresaID, tablesEmp)
	if err != nil {
		return nil, err
	}
	detail = append(detail, empResults...)
	totalDeleted += deletedEmp

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

	return &EmpresaDeleteCascadeResult{
		EmpresaID:           empresaID,
		EmpresaNombre:       strings.TrimSpace(empresa.Nombre),
		TablasAfectadas:     len(detail),
		RegistrosEliminados: totalDeleted,
		Detalle:             detail,
	}, nil
}
