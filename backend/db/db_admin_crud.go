package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var dbAdminIdentRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func isSafeIdent(s string) bool {
	return dbAdminIdentRe.MatchString(strings.TrimSpace(s))
}

type DBAdminColumn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type DBAdminTable struct {
	Name    string         `json:"name"`
	Columns []DBAdminColumn`json:"columns,omitempty"`
}

func DBAdminListEmpresaTables(dbConn *sql.DB) ([]string, error) {
	rows, err := querySQLCompat(dbConn, `
		SELECT table_name
		FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name = 'empresa_id'
		GROUP BY table_name
		ORDER BY table_name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func DBAdminGetTableColumns(dbConn *sql.DB, table string) ([]DBAdminColumn, error) {
	table = strings.TrimSpace(table)
	if !isSafeIdent(table) {
		return nil, fmt.Errorf("tabla invalida")
	}
	rows, err := querySQLCompat(dbConn, `
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_schema='public' AND table_name = ?
		ORDER BY ordinal_position ASC`, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := make([]DBAdminColumn, 0)
	for rows.Next() {
		var name, typ string
		if err := rows.Scan(&name, &typ); err != nil {
			continue
		}
		name = strings.TrimSpace(name)
		typ = strings.TrimSpace(typ)
		if name == "" {
			continue
		}
		cols = append(cols, DBAdminColumn{Name: name, Type: typ})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func dbAdminHasEmpresaIDColumn(cols []DBAdminColumn) bool {
	for _, c := range cols {
		if strings.EqualFold(strings.TrimSpace(c.Name), "empresa_id") {
			return true
		}
	}
	return false
}

func dbAdminColumnsSet(cols []DBAdminColumn) map[string]struct{} {
	m := make(map[string]struct{}, len(cols))
	for _, c := range cols {
		m[strings.ToLower(strings.TrimSpace(c.Name))] = struct{}{}
	}
	return m
}

type DBAdminSelectRequest struct {
	Table   string                 `json:"table"`
	Columns []string               `json:"columns,omitempty"`
	Where   map[string]interface{} `json:"where,omitempty"`
	Limit   int                    `json:"limit,omitempty"`
	OrderBy string                 `json:"order_by,omitempty"`
}

type DBAdminMutationRequest struct {
	Table  string                 `json:"table"`
	Values map[string]interface{} `json:"values,omitempty"`
	Where  map[string]interface{} `json:"where,omitempty"`
}

func DBAdminSelect(dbConn *sql.DB, empresaID int64, req DBAdminSelectRequest) ([]map[string]interface{}, error) {
	if empresaID <= 0 {
		return nil, errors.New("empresa_id es obligatorio")
	}
	req.Table = strings.TrimSpace(req.Table)
	if !isSafeIdent(req.Table) {
		return nil, errors.New("tabla invalida")
	}
	cols, err := DBAdminGetTableColumns(dbConn, req.Table)
	if err != nil {
		return nil, err
	}
	if !dbAdminHasEmpresaIDColumn(cols) {
		return nil, errors.New("tabla no soportada (no tiene empresa_id)")
	}
	colset := dbAdminColumnsSet(cols)

	selectCols := make([]string, 0)
	if len(req.Columns) == 0 {
		// por defecto, limitar a 50 primeras columnas para no exceder payloads gigantes
		for i, c := range cols {
			if i >= 50 {
				break
			}
			selectCols = append(selectCols, c.Name)
		}
	} else {
		for _, c := range req.Columns {
			c = strings.TrimSpace(c)
			if !isSafeIdent(c) {
				return nil, errors.New("columna invalida")
			}
			if _, ok := colset[strings.ToLower(c)]; !ok {
				return nil, fmt.Errorf("columna no existe: %s", c)
			}
			selectCols = append(selectCols, c)
		}
	}
	if len(selectCols) == 0 {
		return nil, errors.New("sin columnas para seleccionar")
	}

	whereParts := []string{"empresa_id = ?"}
	args := []interface{}{empresaID}

	keys := make([]string, 0, len(req.Where))
	for k := range req.Where {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		col := strings.TrimSpace(k)
		if !isSafeIdent(col) {
			return nil, errors.New("where invalido")
		}
		if _, ok := colset[strings.ToLower(col)]; !ok {
			return nil, fmt.Errorf("columna where no existe: %s", col)
		}
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", col))
		args = append(args, req.Where[k])
	}

	limit := req.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	orderBy := strings.TrimSpace(req.OrderBy)
	if orderBy != "" {
		if !isSafeIdent(orderBy) {
			return nil, errors.New("order_by invalido")
		}
		if _, ok := colset[strings.ToLower(orderBy)]; !ok {
			return nil, fmt.Errorf("columna order_by no existe: %s", orderBy)
		}
	}

	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s",
		strings.Join(selectCols, ", "),
		req.Table,
		strings.Join(whereParts, " AND "),
	)
	if orderBy != "" {
		q += " ORDER BY " + orderBy + " DESC"
	}
	q += " LIMIT ?"
	args = append(args, limit)

	rows, err := querySQLCompat(dbConn, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colNames, _ := rows.Columns()
	out := make([]map[string]interface{}, 0)
	for rows.Next() {
		raw := make([]interface{}, len(colNames))
		ptrs := make([]interface{}, len(colNames))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(map[string]interface{}, len(colNames))
		for i, name := range colNames {
			row[name] = raw[i]
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func DBAdminInsert(dbConn *sql.DB, empresaID int64, req DBAdminMutationRequest) (int64, error) {
	if empresaID <= 0 {
		return 0, errors.New("empresa_id es obligatorio")
	}
	req.Table = strings.TrimSpace(req.Table)
	if !isSafeIdent(req.Table) {
		return 0, errors.New("tabla invalida")
	}
	cols, err := DBAdminGetTableColumns(dbConn, req.Table)
	if err != nil {
		return 0, err
	}
	if !dbAdminHasEmpresaIDColumn(cols) {
		return 0, errors.New("tabla no soportada (no tiene empresa_id)")
	}
	colset := dbAdminColumnsSet(cols)

	values := map[string]interface{}{}
	for k, v := range req.Values {
		values[strings.TrimSpace(k)] = v
	}
	values["empresa_id"] = empresaID

	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	colsOut := make([]string, 0, len(keys))
	args := make([]interface{}, 0, len(keys))
	placeholders := make([]string, 0, len(keys))
	for _, k := range keys {
		if !isSafeIdent(k) {
			return 0, errors.New("columna invalida")
		}
		if _, ok := colset[strings.ToLower(k)]; !ok {
			return 0, fmt.Errorf("columna no existe: %s", k)
		}
		colsOut = append(colsOut, k)
		placeholders = append(placeholders, "?")
		args = append(args, values[k])
	}
	if len(colsOut) == 0 {
		return 0, errors.New("sin valores para insertar")
	}

	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", req.Table, strings.Join(colsOut, ", "), strings.Join(placeholders, ", "))
	return insertSQLCompat(dbConn, q, args...)
}

func DBAdminUpdate(dbConn *sql.DB, empresaID int64, req DBAdminMutationRequest) (int64, error) {
	if empresaID <= 0 {
		return 0, errors.New("empresa_id es obligatorio")
	}
	req.Table = strings.TrimSpace(req.Table)
	if !isSafeIdent(req.Table) {
		return 0, errors.New("tabla invalida")
	}
	cols, err := DBAdminGetTableColumns(dbConn, req.Table)
	if err != nil {
		return 0, err
	}
	if !dbAdminHasEmpresaIDColumn(cols) {
		return 0, errors.New("tabla no soportada (no tiene empresa_id)")
	}
	colset := dbAdminColumnsSet(cols)

	if len(req.Values) == 0 {
		return 0, errors.New("values es obligatorio")
	}
	if len(req.Where) == 0 {
		return 0, errors.New("where es obligatorio")
	}

	setKeys := make([]string, 0, len(req.Values))
	for k := range req.Values {
		setKeys = append(setKeys, k)
	}
	sort.Strings(setKeys)
	setParts := make([]string, 0, len(setKeys))
	args := make([]interface{}, 0)
	for _, k := range setKeys {
		col := strings.TrimSpace(k)
		if strings.EqualFold(col, "empresa_id") {
			continue
		}
		if !isSafeIdent(col) {
			return 0, errors.New("columna invalida")
		}
		if _, ok := colset[strings.ToLower(col)]; !ok {
			return 0, fmt.Errorf("columna no existe: %s", col)
		}
		setParts = append(setParts, fmt.Sprintf("%s = ?", col))
		args = append(args, req.Values[k])
	}
	if len(setParts) == 0 {
		return 0, errors.New("no hay columnas actualizables")
	}

	whereParts := []string{"empresa_id = ?"}
	args = append(args, empresaID)
	whereKeys := make([]string, 0, len(req.Where))
	for k := range req.Where {
		whereKeys = append(whereKeys, k)
	}
	sort.Strings(whereKeys)
	for _, k := range whereKeys {
		col := strings.TrimSpace(k)
		if !isSafeIdent(col) {
			return 0, errors.New("where invalido")
		}
		if _, ok := colset[strings.ToLower(col)]; !ok {
			return 0, fmt.Errorf("columna where no existe: %s", col)
		}
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", col))
		args = append(args, req.Where[k])
	}

	q := fmt.Sprintf("UPDATE %s SET %s WHERE %s", req.Table, strings.Join(setParts, ", "), strings.Join(whereParts, " AND "))
	res, err := execSQLCompat(dbConn, q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func DBAdminDelete(dbConn *sql.DB, empresaID int64, req DBAdminMutationRequest) (int64, error) {
	if empresaID <= 0 {
		return 0, errors.New("empresa_id es obligatorio")
	}
	req.Table = strings.TrimSpace(req.Table)
	if !isSafeIdent(req.Table) {
		return 0, errors.New("tabla invalida")
	}
	cols, err := DBAdminGetTableColumns(dbConn, req.Table)
	if err != nil {
		return 0, err
	}
	if !dbAdminHasEmpresaIDColumn(cols) {
		return 0, errors.New("tabla no soportada (no tiene empresa_id)")
	}
	colset := dbAdminColumnsSet(cols)

	if len(req.Where) == 0 {
		return 0, errors.New("where es obligatorio")
	}
	// política: para borrar exige "id" en where (reduce riesgo de borrar masivo)
	if _, ok := req.Where["id"]; !ok {
		if _, ok2 := req.Where["ID"]; !ok2 {
			return 0, errors.New("para DELETE debes indicar where.id")
		}
	}

	whereParts := []string{"empresa_id = ?"}
	args := []interface{}{empresaID}
	whereKeys := make([]string, 0, len(req.Where))
	for k := range req.Where {
		whereKeys = append(whereKeys, k)
	}
	sort.Strings(whereKeys)
	for _, k := range whereKeys {
		col := strings.TrimSpace(k)
		if !isSafeIdent(col) {
			return 0, errors.New("where invalido")
		}
		if _, ok := colset[strings.ToLower(col)]; !ok {
			return 0, fmt.Errorf("columna where no existe: %s", col)
		}
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", col))
		args = append(args, req.Where[k])
	}

	q := fmt.Sprintf("DELETE FROM %s WHERE %s", req.Table, strings.Join(whereParts, " AND "))
	res, err := execSQLCompat(dbConn, q, args...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func DBAdminMarshalSafeMetadata(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return `{}`
	}
	if len(b) > 4000 {
		return `{"warning":"metadata_truncated"}`
	}
	return string(b)
}

