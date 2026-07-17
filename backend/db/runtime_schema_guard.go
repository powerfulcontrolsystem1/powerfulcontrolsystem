package db

import (
	"database/sql"
	"os"
	"strings"
)

// runtimeDDLBlocked is the final safety boundary between schema ownership and
// business traffic. Legacy Ensure* calls may remain temporarily in services,
// but production API and worker roles cannot execute DDL through the shared
// PostgreSQL compatibility layer.
func runtimeDDLBlocked(query string) bool {
	if !isRuntimeDDL(query) {
		return false
	}
	role := strings.ToLower(strings.TrimSpace(os.Getenv("PCS_RUNTIME_ROLE")))
	if role == "migrate" {
		return false
	}
	production := strings.EqualFold(strings.TrimSpace(os.Getenv("PCS_ENV")), "production") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("APP_ENV")), "production")
	if production && (role == "api" || role == "worker" || role == "") {
		return true
	}
	return isDisabledRuntimeFlag(os.Getenv("PCS_RUNTIME_SCHEMA_BOOTSTRAP"))
}

func isRuntimeDDL(query string) bool {
	value := strings.TrimSpace(query)
	for strings.HasPrefix(value, "--") {
		if index := strings.IndexByte(value, '\n'); index >= 0 {
			value = strings.TrimSpace(value[index+1:])
		} else {
			return false
		}
	}
	value = strings.ToUpper(value)
	for _, prefix := range []string{"CREATE ", "ALTER ", "DROP ", "TRUNCATE ", "COMMENT ", "GRANT ", "REVOKE ", "DO "} {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func isDisabledRuntimeFlag(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "false", "no", "off":
		return true
	default:
		return false
	}
}

type noOpSQLResult struct{}

func (noOpSQLResult) LastInsertId() (int64, error) { return 0, sql.ErrNoRows }
func (noOpSQLResult) RowsAffected() (int64, error) { return 0, nil }
