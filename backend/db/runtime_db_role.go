package db

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

var runtimeDatabaseRoleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,62}$`)
var runtimeDatabasePasswordPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{32,}$`)

// EnsureRuntimeDatabaseRole creates or rotates the login used by API and
// worker, grants only data-plane privileges, and removes DDL capability. It
// must run through the owner connection held exclusively by pcs-migrate.
func EnsureRuntimeDatabaseRole(ctx context.Context, dbConn *sql.DB, roleName, password string) error {
	if dbConn == nil {
		return fmt.Errorf("migration database is required")
	}
	roleName = strings.TrimSpace(roleName)
	if !runtimeDatabaseRoleNamePattern.MatchString(roleName) {
		return fmt.Errorf("PCS_RUNTIME_DB_USER must match %s", runtimeDatabaseRoleNamePattern.String())
	}
	if !runtimeDatabasePasswordPattern.MatchString(password) {
		return fmt.Errorf("PCS_RUNTIME_DB_PASSWORD must contain at least 32 URL-safe characters")
	}

	var owner, databaseName string
	if err := dbConn.QueryRowContext(ctx, `SELECT current_user, current_database()`).Scan(&owner, &databaseName); err != nil {
		return fmt.Errorf("inspect migration owner: %w", err)
	}
	if strings.EqualFold(strings.TrimSpace(owner), roleName) {
		return fmt.Errorf("runtime database role must differ from migration owner")
	}
	if !runtimeDatabaseRoleNamePattern.MatchString(owner) {
		return fmt.Errorf("migration owner %q is not a safe PostgreSQL identifier", owner)
	}

	var roleDDL string
	if err := dbConn.QueryRowContext(ctx, `
		SELECT CASE
			WHEN EXISTS (SELECT 1 FROM pg_roles WHERE rolname = $1::text)
			THEN format('ALTER ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS', $1::text, $2::text)
			ELSE format('CREATE ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS', $1::text, $2::text)
		END
	`, roleName, password).Scan(&roleDDL); err != nil {
		return fmt.Errorf("prepare runtime role: %w", err)
	}
	if _, err := dbConn.ExecContext(ctx, roleDDL); err != nil {
		return fmt.Errorf("create or rotate runtime role: %w", err)
	}

	quotedRole := `"` + roleName + `"`
	quotedOwner := `"` + owner + `"`
	quotedDatabase := `"` + strings.ReplaceAll(databaseName, `"`, `""`) + `"`
	statements := []string{
		`REVOKE CREATE ON SCHEMA public FROM PUBLIC`,
		`REVOKE CREATE ON SCHEMA public FROM ` + quotedRole,
		`GRANT CONNECT ON DATABASE ` + quotedDatabase + ` TO ` + quotedRole,
		`GRANT USAGE ON SCHEMA public TO ` + quotedRole,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO ` + quotedRole,
		`GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO ` + quotedRole,
		`GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO ` + quotedRole,
		`ALTER DEFAULT PRIVILEGES FOR ROLE ` + quotedOwner + ` IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ` + quotedRole,
		`ALTER DEFAULT PRIVILEGES FOR ROLE ` + quotedOwner + ` IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO ` + quotedRole,
		`ALTER DEFAULT PRIVILEGES FOR ROLE ` + quotedOwner + ` IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO ` + quotedRole,
	}
	for _, statement := range statements {
		if _, err := dbConn.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("grant runtime role privileges: %w", err)
		}
	}

	var superuser, createDB, createRole, bypassRLS, schemaCreate bool
	if err := dbConn.QueryRowContext(ctx, `
		SELECT r.rolsuper,
		       r.rolcreatedb,
		       r.rolcreaterole,
		       r.rolbypassrls,
		       has_schema_privilege(r.rolname, 'public', 'CREATE')
		FROM pg_roles r
		WHERE r.rolname = $1
	`, roleName).Scan(&superuser, &createDB, &createRole, &bypassRLS, &schemaCreate); err != nil {
		return fmt.Errorf("verify runtime role privileges: %w", err)
	}
	if superuser || createDB || createRole || bypassRLS || schemaCreate {
		return fmt.Errorf("runtime database role retains administrative or DDL privileges")
	}
	return nil
}
