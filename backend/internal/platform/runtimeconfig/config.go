// Package runtimeconfig centralizes the process role and critical runtime
// switches. It deliberately has no dependency on HTTP or database packages so
// API, migration and worker binaries can validate their startup consistently.
package runtimeconfig

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Role string

const (
	RoleAPI     Role = "api"
	RoleWorker  Role = "worker"
	RoleMigrate Role = "migrate"
)

type Config struct {
	Role                  Role
	Production            bool
	LegacySchemaBootstrap bool
	Database              DatabaseConfig
}

// DatabaseConfig centralizes PostgreSQL pool and deadline policy. Containers
// receive the values explicitly, while conservative defaults keep local
// development usable without creating an unbounded pool.
type DatabaseConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	ConnectTimeout  time.Duration
	QueryTimeout    time.Duration
	TxTimeout       time.Duration
}

// Load accepts an environment accessor to keep production startup logic unit
// testable. A production API must never mutate schemas as a side effect of
// serving traffic. Historical bootstrapping can be enabled explicitly only as
// a temporary compatibility measure while a deployment is migrated.
func Load(getenv func(string) string) (Config, error) {
	if getenv == nil {
		return Config{}, fmt.Errorf("environment accessor is required")
	}
	production := isProduction(getenv("PCS_ENV")) || isProduction(getenv("APP_ENV"))
	rawRole := strings.ToLower(strings.TrimSpace(getenv("PCS_RUNTIME_ROLE")))
	if rawRole == "" {
		rawRole = string(RoleAPI)
	}
	role := Role(rawRole)
	switch role {
	case RoleAPI, RoleWorker, RoleMigrate:
	default:
		return Config{}, fmt.Errorf("PCS_RUNTIME_ROLE must be api, worker or migrate")
	}
	legacyBootstrap := !production
	if role == RoleMigrate {
		legacyBootstrap = true
	}
	if production && role == RoleAPI {
		legacyBootstrap = isEnabled(getenv("PCS_RUNTIME_SCHEMA_BOOTSTRAP"))
	}
	database, err := loadDatabaseConfig(getenv)
	if err != nil {
		return Config{}, err
	}
	return Config{Role: role, Production: production, LegacySchemaBootstrap: legacyBootstrap, Database: database}, nil
}

func loadDatabaseConfig(getenv func(string) string) (DatabaseConfig, error) {
	maxOpen, err := positiveInt(getenv("PCS_DB_MAX_OPEN_CONNS"), 32, 1, 500)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("PCS_DB_MAX_OPEN_CONNS: %w", err)
	}
	maxIdle, err := positiveInt(getenv("PCS_DB_MAX_IDLE_CONNS"), 16, 0, maxOpen)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("PCS_DB_MAX_IDLE_CONNS: %w", err)
	}
	lifetime, err := boundedDuration(getenv("PCS_DB_CONN_MAX_LIFETIME"), 30*time.Minute, time.Minute, 24*time.Hour)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("PCS_DB_CONN_MAX_LIFETIME: %w", err)
	}
	idleTime, err := boundedDuration(getenv("PCS_DB_CONN_MAX_IDLE_TIME"), 10*time.Minute, time.Second, lifetime)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("PCS_DB_CONN_MAX_IDLE_TIME: %w", err)
	}
	connectTimeout, err := boundedDuration(getenv("PCS_DB_CONNECT_TIMEOUT"), 10*time.Second, time.Second, time.Minute)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("PCS_DB_CONNECT_TIMEOUT: %w", err)
	}
	queryTimeout, err := boundedDuration(getenv("PCS_DB_QUERY_TIMEOUT"), 30*time.Second, time.Second, 10*time.Minute)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("PCS_DB_QUERY_TIMEOUT: %w", err)
	}
	txTimeout, err := boundedDuration(getenv("PCS_DB_TX_TIMEOUT"), time.Minute, time.Second, 15*time.Minute)
	if err != nil {
		return DatabaseConfig{}, fmt.Errorf("PCS_DB_TX_TIMEOUT: %w", err)
	}
	return DatabaseConfig{MaxOpenConns: maxOpen, MaxIdleConns: maxIdle, ConnMaxLifetime: lifetime, ConnMaxIdleTime: idleTime, ConnectTimeout: connectTimeout, QueryTimeout: queryTimeout, TxTimeout: txTimeout}, nil
}

// OpenAndPing configures a pool before its first use. Callers keep ownership
// of Close; this helper never writes credentials or DSNs to errors.
func (c DatabaseConfig) OpenAndPing(ctx context.Context, driverName, dsn, label string) (*sql.DB, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, fmt.Errorf("%s database DSN is required", strings.TrimSpace(label))
	}
	dbConn, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s database: %w", strings.TrimSpace(label), err)
	}
	dbConn.SetMaxOpenConns(c.MaxOpenConns)
	dbConn.SetMaxIdleConns(c.MaxIdleConns)
	dbConn.SetConnMaxLifetime(c.ConnMaxLifetime)
	dbConn.SetConnMaxIdleTime(c.ConnMaxIdleTime)
	if ctx == nil {
		ctx = context.Background()
	}
	pingCtx, cancel := context.WithTimeout(ctx, c.ConnectTimeout)
	defer cancel()
	if err := dbConn.PingContext(pingCtx); err != nil {
		_ = dbConn.Close()
		return nil, fmt.Errorf("failed to connect %s database: %w", strings.TrimSpace(label), err)
	}
	return dbConn, nil
}

func positiveInt(raw string, fallback, minimum, maximum int) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < minimum || value > maximum {
		return 0, fmt.Errorf("must be between %d and %d", minimum, maximum)
	}
	return value, nil
}

func boundedDuration(raw string, fallback, minimum, maximum time.Duration) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value < minimum || value > maximum {
		return 0, fmt.Errorf("must be a duration between %s and %s", minimum, maximum)
	}
	return value, nil
}

func isProduction(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "production")
}

func isEnabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func isDisabled(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "0", "false", "no", "off":
		return true
	default:
		return false
	}
}
