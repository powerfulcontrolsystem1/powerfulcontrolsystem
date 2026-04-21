package handlers

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func openTestPostgres(t *testing.T, envKey, schemaPrefix string) *sql.DB {
	t.Helper()
	t.Setenv("DB_DIALECT", "postgres")
	t.Setenv("DB_ENGINE", "postgres")
	t.Setenv("PCS_DB_DIALECT", "postgres")

	baseDSN := strings.TrimSpace(os.Getenv(envKey))
	if baseDSN == "" {
		baseDSN = readEnvLocalValue(t, envKey)
	}
	baseDSN = rewriteTestPostgresDSNForTunnel(baseDSN)
	if baseDSN == "" {
		t.Skipf("%s no definido para pruebas PostgreSQL", envKey)
	}

	adminDB, err := sql.Open(dbpkg.PostgresCompatDriverName(), baseDSN)
	if err != nil {
		t.Fatalf("open postgres admin connection: %v", err)
	}
	if err := adminDB.Ping(); err != nil {
		_ = adminDB.Close()
		t.Fatalf("ping postgres admin connection: %v", err)
	}

	schemaName := sanitizePostgresTestSchema(schemaPrefix + "_" + t.Name())
	if _, err := adminDB.Exec(`CREATE SCHEMA IF NOT EXISTS ` + quotePostgresTestIdentifier(schemaName)); err != nil {
		_ = adminDB.Close()
		t.Fatalf("create test schema %s: %v", schemaName, err)
	}

	dsnWithSchema := withPostgresSearchPath(t, baseDSN, schemaName)
	dbConn, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsnWithSchema)
	if err != nil {
		_, _ = adminDB.Exec(`DROP SCHEMA IF EXISTS ` + quotePostgresTestIdentifier(schemaName) + ` CASCADE`)
		_ = adminDB.Close()
		t.Fatalf("open postgres test connection: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	if err := dbConn.Ping(); err != nil {
		_ = dbConn.Close()
		_, _ = adminDB.Exec(`DROP SCHEMA IF EXISTS ` + quotePostgresTestIdentifier(schemaName) + ` CASCADE`)
		_ = adminDB.Close()
		t.Fatalf("ping postgres test connection: %v", err)
	}
	if err := dbpkg.EnsurePostgresRuntimeCompat(dbConn); err != nil {
		_ = dbConn.Close()
		_, _ = adminDB.Exec(`DROP SCHEMA IF EXISTS ` + quotePostgresTestIdentifier(schemaName) + ` CASCADE`)
		_ = adminDB.Close()
		t.Fatalf("ensure postgres runtime compat: %v", err)
	}

	t.Cleanup(func() {
		_ = dbConn.Close()
		_, _ = adminDB.Exec(`DROP SCHEMA IF EXISTS ` + quotePostgresTestIdentifier(schemaName) + ` CASCADE`)
		_ = adminDB.Close()
	})

	return dbConn
}

func readEnvLocalValue(t *testing.T, key string) string {
	if t != nil {
		t.Helper()
	}
	raw, err := os.ReadFile(".env.local")
	if err != nil {
		return ""
	}
	needle := key + "="
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.HasPrefix(trimmed, needle) {
			continue
		}
		return strings.TrimSpace(strings.TrimPrefix(trimmed, needle))
	}
	return ""
}

func withPostgresSearchPath(t *testing.T, baseDSN, schemaName string) string {
	t.Helper()
	parsed, err := url.Parse(baseDSN)
	if err != nil {
		t.Fatalf("parse postgres DSN: %v", err)
	}
	query := parsed.Query()
	query.Set("search_path", schemaName)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func rewriteTestPostgresDSNForTunnel(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}
	if strings.TrimSpace(os.Getenv("DB_VPS_TUNNEL_ENABLED")) != "1" {
		if strings.TrimSpace(readEnvLocalValue(nil, "DB_VPS_TUNNEL_ENABLED")) != "1" {
			return raw
		}
	}
	localPort := strings.TrimSpace(os.Getenv("DB_VPS_LOCAL_PORT"))
	if localPort == "" {
		localPort = strings.TrimSpace(readEnvLocalValue(nil, "DB_VPS_LOCAL_PORT"))
	}
	if localPort == "" {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	hostname := parsed.Hostname()
	if hostname == "" {
		hostname = "127.0.0.1"
	}
	if hostname != "127.0.0.1" && hostname != "localhost" {
		return raw
	}
	parsed.Host = net.JoinHostPort("127.0.0.1", localPort)
	return parsed.String()
}

func sanitizePostgresTestSchema(value string) string {
	var b strings.Builder
	b.Grow(len(value) + 16)
	b.WriteString("test_")
	for _, r := range strings.ToLower(value) {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	b.WriteString(fmt.Sprintf("_%d", time.Now().UnixNano()))
	result := b.String()
	if len(result) > 58 {
		result = result[:58]
	}
	return strings.TrimRight(result, "_")
}

func quotePostgresTestIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}