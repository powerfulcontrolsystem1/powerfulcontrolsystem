package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
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

// EnsurePostgresRuntimeCompat crea funciones auxiliares SQL en PostgreSQL para
// mantener compatibilidad con expresiones SQLite usadas por modulos legacy.
func EnsurePostgresRuntimeCompat(dbConn *sql.DB) error {
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}

	stmts := []string{
		`CREATE OR REPLACE FUNCTION datetime(base_value text DEFAULT 'now', VARIADIC modifiers text[] DEFAULT ARRAY[]::text[])
RETURNS text
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

	RETURN to_char(ts, 'YYYY-MM-DD HH24:MI:SS');
END;
$fn$;`,
		`CREATE OR REPLACE FUNCTION datetime(base_value timestamp, VARIADIC modifiers text[] DEFAULT ARRAY[]::text[])
RETURNS text
LANGUAGE sql
STABLE
AS $$
SELECT datetime(to_char(base_value, 'YYYY-MM-DD HH24:MI:SS'), VARIADIC modifiers);
$$;`,
		`CREATE OR REPLACE FUNCTION datetime(base_value timestamptz, VARIADIC modifiers text[] DEFAULT ARRAY[]::text[])
RETURNS text
LANGUAGE sql
STABLE
AS $$
SELECT datetime(to_char(base_value AT TIME ZONE current_setting('TIMEZONE'), 'YYYY-MM-DD HH24:MI:SS'), VARIADIC modifiers);
$$;`,
		`CREATE OR REPLACE FUNCTION date(base_value text, modifier text)
RETURNS text
LANGUAGE sql
STABLE
AS $$
SELECT split_part(datetime(base_value, modifier), ' ', 1);
$$;`,
		`CREATE OR REPLACE FUNCTION date(base_value text, modifier1 text, modifier2 text)
RETURNS text
LANGUAGE sql
STABLE
AS $$
SELECT split_part(datetime(base_value, modifier1, modifier2), ' ', 1);
$$;`,
		`CREATE OR REPLACE FUNCTION "time"(base_value text, modifier text)
RETURNS text
LANGUAGE sql
STABLE
AS $$
SELECT split_part(datetime(base_value, modifier), ' ', 2);
$$;`,
		`CREATE OR REPLACE FUNCTION "time"(base_value text, modifier1 text, modifier2 text)
RETURNS text
LANGUAGE sql
STABLE
AS $$
SELECT split_part(datetime(base_value, modifier1, modifier2), ' ', 2);
$$;`,
		`CREATE OR REPLACE FUNCTION printf(format_text text, value anyelement)
RETURNS text
LANGUAGE sql
IMMUTABLE
AS $$
SELECT replace(COALESCE(format_text, ''), '%d', COALESCE(value::text, ''));
$$;`,
	}

	for i, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return fmt.Errorf("postgres compat step %d failed: %w", i+1, err)
		}
	}

	return nil
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
	q = rewriteAutoIncrementForPostgres(q)
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

func rewriteAutoIncrementForPostgres(query string) string {
	q := replaceInsensitiveAll(query, "integer primary key autoincrement", "BIGSERIAL PRIMARY KEY")
	q = replaceInsensitiveAll(q, "int primary key autoincrement", "BIGSERIAL PRIMARY KEY")
	return q
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
	candidates := []string{
		strings.TrimSpace(os.Getenv("DB_DIALECT")),
		strings.TrimSpace(os.Getenv("DB_ENGINE")),
		strings.TrimSpace(os.Getenv("PCS_DB_DIALECT")),
	}
	for _, raw := range candidates {
		v := strings.ToLower(raw)
		if v == "" {
			continue
		}
		if strings.Contains(v, "postgres") {
			return "postgres"
		}
		if strings.Contains(v, "sqlite") {
			return "sqlite"
		}
	}
	return "sqlite"
}

func isPostgresDialect() bool {
	return currentSQLDialect() == "postgres"
}

// IsPostgresDialect expone el dialecto actual para capas fuera del paquete db.
func IsPostgresDialect() bool {
	return isPostgresDialect()
}

func sqlNowExpr() string {
	if isPostgresDialect() {
		return "CURRENT_TIMESTAMP"
	}
	return "datetime('now','localtime')"
}

func sqlPlusHoursExpr(hours int) string {
	if isPostgresDialect() {
		return "(CURRENT_TIMESTAMP + interval '" + strconv.Itoa(hours) + " hour')"
	}
	return "datetime('now','+" + strconv.Itoa(hours) + " hours','localtime')"
}

func sessionNotExpiredCondition(columnName string) string {
	if isPostgresDialect() {
		return "(COALESCE(CAST(" + columnName + " AS TEXT), '') = '' OR CAST(" + columnName + " AS TIMESTAMP) > CURRENT_TIMESTAMP)"
	}
	return "(COALESCE(" + columnName + ", '') = '' OR datetime(" + columnName + ") > datetime('now','localtime'))"
}

func rebindCompatQuery(query string) string {
	if !isPostgresDialect() {
		return query
	}
	return rebindQuestionPlaceholders(query)
}

func normalizeColumnDefForDialect(columnDef string) string {
	def := strings.TrimSpace(columnDef)
	if !isPostgresDialect() || def == "" {
		return def
	}

	replacements := []struct {
		old string
		new string
	}{
		{"TEXT DEFAULT (datetime('now','localtime'))", "TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))"},
		{"TEXT DEFAULT (datetime('now'))", "TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT))"},
		{"datetime('now','localtime')", "CURRENT_TIMESTAMP"},
		{"datetime('now')", "CURRENT_TIMESTAMP"},
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
	return dbConn.Exec(rebindCompatQuery(query), args...)
}

func execTxSQLCompat(tx *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	return tx.Exec(rebindCompatQuery(query), args...)
}

func querySQLCompat(dbConn *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return dbConn.Query(rebindCompatQuery(query), args...)
}

func queryTxSQLCompat(tx *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	return tx.Query(rebindCompatQuery(query), args...)
}

func queryRowSQLCompat(dbConn *sql.DB, query string, args ...interface{}) *sql.Row {
	return dbConn.QueryRow(rebindCompatQuery(query), args...)
}

func queryRowTxSQLCompat(tx *sql.Tx, query string, args ...interface{}) *sql.Row {
	return tx.QueryRow(rebindCompatQuery(query), args...)
}

func insertSQLCompat(dbConn *sql.DB, query string, args ...interface{}) (int64, error) {
	if isPostgresDialect() {
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

	res, err := execSQLCompat(dbConn, query, args...)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func insertTxSQLCompat(tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	if isPostgresDialect() {
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

	res, err := execTxSQLCompat(tx, query, args...)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}
