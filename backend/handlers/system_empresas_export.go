package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type empresaInfoExportTable struct {
	Source    string                   `json:"source"`
	Table     string                   `json:"table"`
	Columns   []string                 `json:"columns"`
	RowCount  int                      `json:"row_count"`
	Rows      []map[string]interface{} `json:"rows,omitempty"`
	Truncated bool                     `json:"truncated,omitempty"`
}

type empresaInfoExportSnapshot struct {
	Empresa      *dbpkg.Empresa           `json:"empresa"`
	GeneratedAt  string                   `json:"generated_at"`
	Formats      []string                 `json:"formats"`
	TotalTables  int                      `json:"total_tables"`
	TotalRows    int                      `json:"total_rows"`
	PreviewLimit int                      `json:"preview_limit,omitempty"`
	SourceTotals map[string]int           `json:"source_totals,omitempty"`
	Tables       []empresaInfoExportTable `json:"tables"`
}

func superSQLCompat(query string) string {
	if !dbpkg.IsPostgresDialect() {
		return query
	}
	var builder strings.Builder
	index := 1
	for _, ch := range query {
		if ch == '?' {
			builder.WriteString("$")
			builder.WriteString(strconv.Itoa(index))
			index++
			continue
		}
		builder.WriteRune(ch)
	}
	return builder.String()
}

func superQuoteIdentifier(identifier string) string {
	trimmed := strings.TrimSpace(identifier)
	return `"` + strings.ReplaceAll(trimmed, `"`, `""`) + `"`
}

func superListTables(dbConn *sql.DB) ([]string, error) {
	if dbConn == nil {
		return nil, nil
	}
	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ANY(current_schemas(false))
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name`
	rows, err := dbConn.Query(superSQLCompat(query))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		table = strings.TrimSpace(table)
		if table == "" {
			continue
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tables, nil
}

func superListColumns(dbConn *sql.DB, table string) ([]string, error) {
	if dbConn == nil || strings.TrimSpace(table) == "" {
		return nil, nil
	}
	columns := make([]string, 0)
	rows, err := dbConn.Query(superSQLCompat(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = ANY(current_schemas(false))
		  AND table_name = ?
		ORDER BY ordinal_position`), strings.TrimSpace(table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return nil, err
		}
		columns = append(columns, strings.TrimSpace(column))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return columns, nil
}

func superHasColumn(columns []string, name string) bool {
	needle := strings.TrimSpace(strings.ToLower(name))
	for _, column := range columns {
		if strings.TrimSpace(strings.ToLower(column)) == needle {
			return true
		}
	}
	return false
}

func superNormalizeValue(value interface{}) interface{} {
	switch typed := value.(type) {
	case nil:
		return nil
	case []byte:
		return string(typed)
	case time.Time:
		return typed.Format(time.RFC3339)
	case bool, string, int, int8, int16, int32, int64, float32, float64:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func superFetchTableRows(dbConn *sql.DB, source, table string, columns []string, empresaID int64, previewLimit int) ([]map[string]interface{}, int, error) {
	if dbConn == nil || len(columns) == 0 {
		return nil, 0, nil
	}
	whereClause := ""
	args := []interface{}{}
	if source == "empresas" && strings.EqualFold(table, "empresas") {
		whereClause = " WHERE (id = ? OR COALESCE(empresa_id, id) = ?)"
		args = append(args, empresaID, empresaID)
	} else if superHasColumn(columns, "empresa_id") {
		whereClause = " WHERE empresa_id = ?"
		args = append(args, empresaID)
	} else {
		return nil, 0, nil
	}

	countQuery := "SELECT COUNT(1) FROM " + superQuoteIdentifier(table) + whereClause
	var total int
	if err := dbConn.QueryRow(superSQLCompat(countQuery), args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	selectColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		selectColumns = append(selectColumns, superQuoteIdentifier(column))
	}
	query := "SELECT " + strings.Join(selectColumns, ", ") + " FROM " + superQuoteIdentifier(table) + whereClause
	if superHasColumn(columns, "id") {
		query += " ORDER BY id"
	}
	queryArgs := append([]interface{}{}, args...)
	if previewLimit > 0 {
		query += " LIMIT ?"
		queryArgs = append(queryArgs, previewLimit)
	}

	rows, err := dbConn.Query(superSQLCompat(query), queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		values := make([]interface{}, len(columns))
		scans := make([]interface{}, len(columns))
		for i := range values {
			scans[i] = &values[i]
		}
		if err := rows.Scan(scans...); err != nil {
			return nil, 0, err
		}
		item := make(map[string]interface{}, len(columns))
		for i, column := range columns {
			item[column] = superNormalizeValue(values[i])
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func buildEmpresaInfoExportSnapshot(dbEmp, dbSuper *sql.DB, empresaID int64, previewLimit int) (*empresaInfoExportSnapshot, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id invalido")
	}
	empresa, err := dbpkg.GetEmpresaByID(dbEmp, empresaID)
	if err != nil {
		return nil, err
	}

	snapshot := &empresaInfoExportSnapshot{
		Empresa:      empresa,
		GeneratedAt:  time.Now().Format(time.RFC3339),
		Formats:      []string{"pdf", "xls", "csv", "json", "txt"},
		PreviewLimit: previewLimit,
		SourceTotals: map[string]int{"empresas": 0, "super": 0},
		Tables:       make([]empresaInfoExportTable, 0),
	}

	sources := []struct {
		name string
		db   *sql.DB
	}{
		{name: "empresas", db: dbEmp},
		{name: "super", db: dbSuper},
	}

	for _, source := range sources {
		tables, err := superListTables(source.db)
		if err != nil {
			return nil, err
		}
		for _, table := range tables {
			columns, err := superListColumns(source.db, table)
			if err != nil {
				return nil, err
			}
			rows, rowCount, err := superFetchTableRows(source.db, source.name, table, columns, empresaID, previewLimit)
			if err != nil {
				return nil, err
			}
			if rowCount == 0 {
				continue
			}
			snapshot.Tables = append(snapshot.Tables, empresaInfoExportTable{
				Source:    source.name,
				Table:     table,
				Columns:   columns,
				RowCount:  rowCount,
				Rows:      rows,
				Truncated: previewLimit > 0 && rowCount > len(rows),
			})
			snapshot.TotalRows += rowCount
			snapshot.SourceTotals[source.name] += rowCount
		}
	}

	snapshot.TotalTables = len(snapshot.Tables)
	return snapshot, nil
}

func buildEmpresaInfoExportKey(row map[string]interface{}) string {
	candidates := []string{"id", "empresa_id", "nombre", "codigo", "email", "documento_codigo", "reference"}
	parts := make([]string, 0, 3)
	for _, key := range candidates {
		value := strings.TrimSpace(reportesStringValue(row[key]))
		if value == "" {
			continue
		}
		parts = append(parts, key+":"+value)
		if len(parts) == 3 {
			break
		}
	}
	if len(parts) == 0 {
		return "registro"
	}
	return strings.Join(parts, " | ")
}

func buildEmpresaInfoExportDataset(snapshot *empresaInfoExportSnapshot) (empresaReporteDataset, error) {
	if snapshot == nil || snapshot.Empresa == nil {
		return empresaReporteDataset{}, fmt.Errorf("snapshot invalido")
	}
	ds := empresaReporteDataset{
		Key:         "super_empresa_informacion_integral",
		Title:       "Informacion integral de empresa",
		Level:       "super",
		Description: "Consolidado exportable de la informacion empresarial y sus registros asociados por empresa_id.",
		EmpresaID:   snapshot.Empresa.ID,
		GeneratedAt: snapshot.GeneratedAt,
		Columns:     []string{"empresa_id", "empresa_nombre", "fuente", "tabla", "registro", "clave", "datos_json"},
		Rows:        make([]map[string]interface{}, 0),
		Summary: map[string]interface{}{
			"empresa_id":        snapshot.Empresa.ID,
			"empresa_nombre":    strings.TrimSpace(snapshot.Empresa.Nombre),
			"tablas_incluidas":  snapshot.TotalTables,
			"registros_totales": snapshot.TotalRows,
			"fuente_empresas":   snapshot.SourceTotals["empresas"],
			"fuente_super":      snapshot.SourceTotals["super"],
		},
	}

	for _, table := range snapshot.Tables {
		for index, row := range table.Rows {
			rawJSON, err := json.Marshal(row)
			if err != nil {
				return empresaReporteDataset{}, err
			}
			ds.Rows = append(ds.Rows, map[string]interface{}{
				"empresa_id":     snapshot.Empresa.ID,
				"empresa_nombre": strings.TrimSpace(snapshot.Empresa.Nombre),
				"fuente":         table.Source,
				"tabla":          table.Table,
				"registro":       index + 1,
				"clave":          buildEmpresaInfoExportKey(row),
				"datos_json":     string(rawJSON),
			})
		}
	}

	ds.RowCount = len(ds.Rows)
	return ds, nil
}

func writeEmpresaInfoExport(w http.ResponseWriter, snapshot *empresaInfoExportSnapshot, format string) error {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		return fmt.Errorf("format requerido")
	}
	if format == "json" {
		fileName := reportesBuildFileName("super_empresa_informacion_integral", snapshot.Empresa.ID, "json")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
		writeJSON(w, http.StatusOK, snapshot)
		return nil
	}
	ds, err := buildEmpresaInfoExportDataset(snapshot)
	if err != nil {
		return err
	}
	return writeReportesDatasetExport(w, ds, format)
}