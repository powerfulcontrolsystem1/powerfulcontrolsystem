package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/jackc/pgx/v5/stdlib"
)

const pgxCompatDriverName = "pgx_compat"

var registerCompatDriverOnce sync.Once

// PostgresCompatDriverName registra (una sola vez) el driver de compatibilidad
// para PostgreSQL y devuelve el nombre utilizable por sql.Open.
func PostgresCompatDriverName() string {
	registerCompatDriverOnce.Do(func() {
		sql.Register(pgxCompatDriverName, &pgxCompatDriver{base: stdlib.GetDefaultDriver()})
	})
	return pgxCompatDriverName
}

// EnsurePostgresRuntimeCompat crea funciones auxiliares internas PCS en
// PostgreSQL para normalizar fechas antiguas sin depender de funciones de otros
// motores.
func EnsurePostgresRuntimeCompat(dbConn *sql.DB) error {
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}

	stmts := []string{
		`CREATE OR REPLACE FUNCTION pcs_ts(base_value text DEFAULT 'now', VARIADIC modifiers text[] DEFAULT ARRAY[]::text[])
RETURNS timestamp
LANGUAGE plpgsql
STABLE
AS $fn$
DECLARE
	ts timestamp;
	mod text;
BEGIN
	IF base_value IS NULL OR btrim(base_value) = '' OR lower(btrim(base_value)) = 'now' THEN
		ts := CURRENT_TIMESTAMP;
	ELSE
		BEGIN
			ts := base_value::timestamp;
		EXCEPTION WHEN others THEN
			ts := CURRENT_TIMESTAMP;
		END;
	END IF;

	IF modifiers IS NOT NULL THEN
		FOREACH mod IN ARRAY modifiers LOOP
			IF mod IS NULL OR btrim(mod) = '' THEN
				CONTINUE;
			END IF;
			IF lower(btrim(mod)) = 'localtime' THEN
				CONTINUE;
			END IF;
			BEGIN
				ts := ts + mod::interval;
			EXCEPTION WHEN others THEN
				CONTINUE;
			END;
		END LOOP;
	END IF;

	RETURN ts;
END;
$fn$;`,
		`CREATE OR REPLACE FUNCTION pcs_ts(base_value timestamp, VARIADIC modifiers text[] DEFAULT ARRAY[]::text[])
RETURNS timestamp
LANGUAGE sql
STABLE
AS $$
SELECT pcs_ts(to_char(base_value, 'YYYY-MM-DD HH24:MI:SS'), VARIADIC modifiers);
$$;`,
		`CREATE OR REPLACE FUNCTION pcs_ts(base_value timestamptz, VARIADIC modifiers text[] DEFAULT ARRAY[]::text[])
RETURNS timestamp
LANGUAGE sql
STABLE
AS $$
SELECT pcs_ts(to_char(base_value AT TIME ZONE current_setting('TIMEZONE'), 'YYYY-MM-DD HH24:MI:SS'), VARIADIC modifiers);
$$;`,
		`CREATE OR REPLACE FUNCTION pcs_date(base_value text)
RETURNS date
LANGUAGE sql
STABLE
AS $$
SELECT pcs_ts(base_value)::date;
$$;`,
		`CREATE OR REPLACE FUNCTION pcs_date(base_value timestamp)
RETURNS date
LANGUAGE sql
STABLE
AS $$
SELECT base_value::date;
$$;`,
		`CREATE OR REPLACE FUNCTION pcs_julian_day(base_value text)
RETURNS double precision
LANGUAGE plpgsql
STABLE
AS $fn$
DECLARE
	ts timestamp;
BEGIN
	IF base_value IS NULL OR btrim(base_value) = '' OR lower(btrim(base_value)) = 'now' THEN
		ts := CURRENT_TIMESTAMP;
	ELSE
		BEGIN
			ts := base_value::timestamp;
		EXCEPTION WHEN others THEN
			BEGIN
				ts := (base_value::date)::timestamp;
			EXCEPTION WHEN others THEN
				ts := CURRENT_TIMESTAMP;
			END;
		END;
	END IF;

	RETURN EXTRACT(EPOCH FROM ts) / 86400.0;
END;
$fn$;`,
		`CREATE OR REPLACE FUNCTION pcs_julian_day(base_value timestamp)
RETURNS double precision
LANGUAGE sql
STABLE
AS $$
SELECT EXTRACT(EPOCH FROM base_value) / 86400.0;
$$;`,
		`CREATE OR REPLACE FUNCTION pcs_julian_day(base_value date)
RETURNS double precision
LANGUAGE sql
STABLE
AS $$
SELECT EXTRACT(EPOCH FROM (base_value::timestamp)) / 86400.0;
$$;`,
		`CREATE OR REPLACE FUNCTION pcs_julian_day(base_value timestamptz)
RETURNS double precision
LANGUAGE sql
STABLE
AS $$
SELECT EXTRACT(EPOCH FROM (base_value AT TIME ZONE current_setting('TIMEZONE'))) / 86400.0;
$$;`,
	}

	for i, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return fmt.Errorf("postgres compat step %d failed: %w", i+1, err)
		}
	}

	return nil
}

// EnsurePostgresPrimaryKeySequences repara tablas legacy de PostgreSQL cuyo
// campo id primario quedo sin secuencia/default autogenerado tras migraciones antiguas.
func EnsurePostgresPrimaryKeySequences(dbConn *sql.DB) error {
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}

	rows, err := dbConn.Query(`SELECT c.table_name, COALESCE(c.column_default, '')
		FROM information_schema.columns c
		JOIN information_schema.tables t
		  ON t.table_schema = c.table_schema
		 AND t.table_name = c.table_name
		JOIN information_schema.table_constraints tc
		  ON tc.table_schema = c.table_schema
		 AND tc.table_name = c.table_name
		 AND tc.constraint_type = 'PRIMARY KEY'
		JOIN information_schema.key_column_usage kcu
		  ON kcu.table_schema = tc.table_schema
		 AND kcu.table_name = tc.table_name
		 AND kcu.constraint_name = tc.constraint_name
		 AND kcu.column_name = c.column_name
		WHERE c.table_schema = current_schema()
		  AND t.table_type = 'BASE TABLE'
		  AND c.column_name = 'id'
		  AND (c.data_type IN ('integer', 'bigint') OR c.udt_name IN ('int4', 'int8'))
		ORDER BY c.table_name`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		var columnDefault string
		if err := rows.Scan(&tableName, &columnDefault); err != nil {
			return err
		}
		if strings.Contains(strings.ToLower(columnDefault), "nextval(") {
			continue
		}
		if err := ensurePostgresTableIDSequence(dbConn, tableName); err != nil {
			return fmt.Errorf("ensure postgres id sequence for %s: %w", tableName, err)
		}
	}

	return rows.Err()
}

func ensurePostgresTableIDSequence(dbConn *sql.DB, tableName string) error {
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}

	quotedTable := quotePostgresIdentifier(tableName)
	seqName := tableName + "_id_seq"
	quotedSeq := quotePostgresIdentifier(seqName)

	if _, err := dbConn.Exec(fmt.Sprintf(`CREATE SEQUENCE IF NOT EXISTS %s`, quotedSeq)); err != nil {
		return err
	}
	if _, err := dbConn.Exec(fmt.Sprintf(`ALTER SEQUENCE %s OWNED BY %s.id`, quotedSeq, quotedTable)); err != nil {
		return err
	}
	if _, err := dbConn.Exec(fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN id SET DEFAULT nextval('%s')`, quotedTable, seqName)); err != nil {
		return err
	}

	var maxID int64
	if err := dbConn.QueryRow(fmt.Sprintf(`SELECT COALESCE(MAX(id), 0) FROM %s`, quotedTable)).Scan(&maxID); err != nil {
		return err
	}
	if maxID > 0 {
		if _, err := dbConn.Exec(`SELECT setval($1, $2, true)`, seqName, maxID); err != nil {
			return err
		}
		return nil
	}
	if _, err := dbConn.Exec(`SELECT setval($1, 1, false)`, seqName); err != nil {
		return err
	}
	return nil
}

func quotePostgresIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

type pgxCompatDriver struct {
	base driver.Driver
}

func (d *pgxCompatDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.base.Open(name)
	if err != nil {
		return nil, err
	}
	return &pgxCompatConn{Conn: conn}, nil
}

type pgxCompatConn struct {
	driver.Conn
}

func (c *pgxCompatConn) Prepare(query string) (driver.Stmt, error) {
	return c.Conn.Prepare(rewritePostgresCompatQuery(query))
}

func (c *pgxCompatConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if withPrepareContext, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return withPrepareContext.PrepareContext(ctx, rewritePostgresCompatQuery(query))
	}
	return c.Prepare(query)
}

func (c *pgxCompatConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if withExecContext, ok := c.Conn.(driver.ExecerContext); ok {
		return withExecContext.ExecContext(ctx, rewritePostgresCompatQuery(query), args)
	}
	return nil, driver.ErrSkip
}

func (c *pgxCompatConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if withQueryContext, ok := c.Conn.(driver.QueryerContext); ok {
		return withQueryContext.QueryContext(ctx, rewritePostgresCompatQuery(query), args)
	}
	return nil, driver.ErrSkip
}

func (c *pgxCompatConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if withBeginTx, ok := c.Conn.(driver.ConnBeginTx); ok {
		return withBeginTx.BeginTx(ctx, opts)
	}
	if opts.ReadOnly || opts.Isolation != driver.IsolationLevel(sql.LevelDefault) {
		return nil, driver.ErrSkip
	}
	return c.Conn.Begin()
}

func (c *pgxCompatConn) Ping(ctx context.Context) error {
	if withPing, ok := c.Conn.(driver.Pinger); ok {
		return withPing.Ping(ctx)
	}
	return nil
}

func (c *pgxCompatConn) CheckNamedValue(nv *driver.NamedValue) error {
	if checker, ok := c.Conn.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(nv)
	}
	return driver.ErrSkip
}

func (c *pgxCompatConn) ResetSession(ctx context.Context) error {
	if resetter, ok := c.Conn.(driver.SessionResetter); ok {
		return resetter.ResetSession(ctx)
	}
	return nil
}

func (c *pgxCompatConn) IsValid() bool {
	if validator, ok := c.Conn.(driver.Validator); ok {
		return validator.IsValid()
	}
	return true
}

func rewritePostgresCompatQuery(query string) string {
	q := rebindQuestionPlaceholders(query)
	q = rewriteInsertOrIgnoreQuery(q)
	return q
}

func rewriteInsertOrIgnoreQuery(query string) string {
	lower := strings.ToLower(query)
	if !strings.Contains(lower, "insert or ignore into") {
		return query
	}

	rewritten := replaceInsensitiveFirst(query, "insert or ignore into", "INSERT INTO")
	if strings.Contains(strings.ToLower(rewritten), " on conflict ") {
		return rewritten
	}

	rewrittenLower := strings.ToLower(rewritten)
	if returningIdx := strings.LastIndex(rewrittenLower, " returning "); returningIdx >= 0 {
		return strings.TrimRight(rewritten[:returningIdx], " \t\r\n;") + " ON CONFLICT DO NOTHING " + rewritten[returningIdx:]
	}

	trimmed := strings.TrimSpace(rewritten)
	if strings.HasSuffix(trimmed, ";") {
		core := strings.TrimSpace(strings.TrimSuffix(trimmed, ";"))
		return core + " ON CONFLICT DO NOTHING;"
	}

	return rewritten + " ON CONFLICT DO NOTHING"
}

func replaceInsensitiveFirst(value, target, replacement string) string {
	lowerValue := strings.ToLower(value)
	lowerTarget := strings.ToLower(target)
	idx := strings.Index(lowerValue, lowerTarget)
	if idx < 0 {
		return value
	}
	return value[:idx] + replacement + value[idx+len(target):]
}

func replaceInsensitiveAll(value, target, replacement string) string {
	if target == "" {
		return value
	}

	lowerValue := strings.ToLower(value)
	lowerTarget := strings.ToLower(target)

	start := 0
	var b strings.Builder
	b.Grow(len(value))

	for {
		relIdx := strings.Index(lowerValue[start:], lowerTarget)
		if relIdx < 0 {
			b.WriteString(value[start:])
			break
		}
		idx := start + relIdx
		b.WriteString(value[start:idx])
		b.WriteString(replacement)
		start = idx + len(target)
	}

	return b.String()
}

func rebindQuestionPlaceholders(query string) string {
	var b strings.Builder
	b.Grow(len(query) + 8)
	inSingle := false
	arg := 1
	for i := 0; i < len(query); i++ {
		ch := query[i]
		if ch == '\'' {
			b.WriteByte(ch)
			if inSingle {
				if i+1 < len(query) && query[i+1] == '\'' {
					b.WriteByte(query[i+1])
					i++
				} else {
					inSingle = false
				}
			} else {
				inSingle = true
			}
			continue
		}
		if ch == '?' && !inSingle {
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(arg))
			arg++
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

func currentSQLDialect() string {
	// El proyecto es PostgreSQL-only. Se conserva la función para compatibilidad
	// de API interna, pero ya no soporta dialectos alternos.
	return "postgres"
}

func isPostgresDialect() bool {
	return currentSQLDialect() == "postgres"
}

func shouldUsePostgresCompat(dbConn *sql.DB) bool {
	if dbConn == nil || !isPostgresDialect() {
		return false
	}
	return isPostgresConnection(dbConn)
}

func isPostgresConnection(dbConn *sql.DB) bool {
	if dbConn == nil {
		return false
	}
	conn, err := dbConn.Conn(context.Background())
	if err != nil {
		return false
	}
	defer conn.Close()

	isPostgres := false
	err = conn.Raw(func(driverConn interface{}) error {
		typeName := strings.ToLower(fmt.Sprintf("%T", driverConn))
		if strings.Contains(typeName, "pgx") || strings.Contains(typeName, "postgres") || strings.Contains(typeName, "stdlib") {
			isPostgres = true
		}
		return nil
	})
	if err != nil {
		return false
	}
	return isPostgres
}

// IsPostgresDialect expone el dialecto actual para capas fuera del paquete db.
func IsPostgresDialect() bool {
	return isPostgresDialect()
}

func sqlNowExpr() string {
	return "CURRENT_TIMESTAMP"
}

func sqlNowTextExpr() string {
	return "CAST(CURRENT_TIMESTAMP AS TEXT)"
}

func sqlPlusHoursExpr(hours int) string {
	return "(CURRENT_TIMESTAMP + interval '" + strconv.Itoa(hours) + " hour')"
}

func sessionNotExpiredCondition(columnName string) string {
	return "(COALESCE(CAST(" + columnName + " AS TEXT), '') = '' OR CAST(" + columnName + " AS TIMESTAMP) > CURRENT_TIMESTAMP)"
}

func rebindCompatQuery(query string) string {
	return rebindQuestionPlaceholders(query)
}

func normalizeColumnDefForDialect(columnDef string) string {
	def := strings.TrimSpace(columnDef)
	if def == "" {
		return def
	}

	replacements := []struct {
		old string
		new string
	}{
		{"TEXT DEFAULT (CURRENT_TIMESTAMP)", "TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))"},
		{"TEXT DEFAULT (CURRENT_TIMESTAMP)", "TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))"},
		{"CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP"},
		{"CURRENT_TIMESTAMP", "CURRENT_TIMESTAMP"},
	}
	for _, item := range replacements {
		def = strings.ReplaceAll(def, item.old, item.new)
	}

	return def
}

func isMissingColumnError(err error) bool {
	if err == nil {
		return false
	}
	low := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(low, "no such column") ||
		(strings.Contains(low, "column") && strings.Contains(low, "does not exist")) ||
		strings.Contains(low, "has no column named")
}

func isMissingTableError(err error) bool {
	if err == nil {
		return false
	}
	low := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(low, "no such table") || strings.Contains(low, "relation") && strings.Contains(low, "does not exist")
}

func execSQLCompat(dbConn *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	if dbConn == nil {
		dbConn = GetDB()
	}
	if runtimeDDLBlocked(query) {
		return noOpSQLResult{}, nil
	}
	return dbConn.Exec(rebindCompatQuery(query), args...)
}

func execTxSQLCompat(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	if runtimeDDLBlocked(query) {
		return noOpSQLResult{}, nil
	}
	return tx.Exec(rebindCompatQuery(query), args...)
}

func querySQLCompat(dbConn *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	if dbConn == nil {
		dbConn = GetDB()
	}
	return dbConn.Query(rebindCompatQuery(query), args...)
}

func queryTxSQLCompat(tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	return tx.Query(rebindCompatQuery(query), args...)
}

func queryRowSQLCompat(dbConn *sql.DB, query string, args ...interface{}) *sql.Row {
	if dbConn == nil {
		dbConn = GetDB()
	}
	return dbConn.QueryRow(rebindCompatQuery(query), args...)
}

// QueryRowCompat expone QueryRow con rebind ? -> $n para PostgreSQL.
// Se usa desde handlers legacy que aún construyen SQL con placeholders '?'.
func QueryRowCompat(dbConn *sql.DB, query string, args ...interface{}) *sql.Row {
	return queryRowSQLCompat(dbConn, query, args...)
}

func queryRowTxSQLCompat(tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	return tx.QueryRow(rebindCompatQuery(query), args...)
}

func insertSQLCompat(dbConn *sql.DB, query string, args ...interface{}) (int64, error) {
	insertQuery := strings.TrimSpace(query)
	if !strings.Contains(strings.ToLower(insertQuery), "returning") {
		insertQuery += " RETURNING id"
	}
	var id int64
	if err := queryRowSQLCompat(dbConn, insertQuery, args...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func insertTxSQLCompat(tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	insertQuery := strings.TrimSpace(query)
	if !strings.Contains(strings.ToLower(insertQuery), "returning") {
		insertQuery += " RETURNING id"
	}
	var id int64
	if err := queryRowTxSQLCompat(tx, insertQuery, args...).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}
