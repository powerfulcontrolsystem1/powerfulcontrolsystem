package db

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PostgresPoolConfig is process-local. Values are intentionally bounded so a
// replica cannot exhaust PostgreSQL merely because an environment variable was
// typoed or copied from an unrelated service.
type PostgresPoolConfig struct {
	MaxOpen     int
	MaxIdle     int
	MaxLifetime time.Duration
	MaxIdleTime time.Duration
}

func LoadPostgresPoolConfig(getenv func(string) string, role string) (PostgresPoolConfig, error) {
	if getenv == nil {
		return PostgresPoolConfig{}, fmt.Errorf("environment accessor is required")
	}
	config := defaultPostgresPoolConfig(role)
	var err error
	if config.MaxOpen, err = poolEnvInt(getenv, "PCS_DB_MAX_OPEN_CONNS", config.MaxOpen, 1, 200); err != nil {
		return PostgresPoolConfig{}, err
	}
	if config.MaxIdle, err = poolEnvInt(getenv, "PCS_DB_MAX_IDLE_CONNS", config.MaxIdle, 0, config.MaxOpen); err != nil {
		return PostgresPoolConfig{}, err
	}
	if config.MaxLifetime, err = poolEnvDuration(getenv, "PCS_DB_CONN_MAX_LIFETIME", config.MaxLifetime, time.Minute, 24*time.Hour); err != nil {
		return PostgresPoolConfig{}, err
	}
	if config.MaxIdleTime, err = poolEnvDuration(getenv, "PCS_DB_CONN_MAX_IDLE_TIME", config.MaxIdleTime, time.Second, 12*time.Hour); err != nil {
		return PostgresPoolConfig{}, err
	}
	if config.MaxIdle > config.MaxOpen {
		return PostgresPoolConfig{}, fmt.Errorf("PCS_DB_MAX_IDLE_CONNS cannot exceed PCS_DB_MAX_OPEN_CONNS")
	}
	return config, nil
}

func ConfigurePostgresPool(dbConn *sql.DB, config PostgresPoolConfig) error {
	if dbConn == nil {
		return fmt.Errorf("database pool is required")
	}
	if config.MaxOpen < 1 || config.MaxOpen > 200 || config.MaxIdle < 0 || config.MaxIdle > config.MaxOpen ||
		config.MaxLifetime < time.Minute || config.MaxIdleTime < time.Second {
		return fmt.Errorf("postgres pool configuration is invalid")
	}
	dbConn.SetMaxOpenConns(config.MaxOpen)
	dbConn.SetMaxIdleConns(config.MaxIdle)
	dbConn.SetConnMaxLifetime(config.MaxLifetime)
	dbConn.SetConnMaxIdleTime(config.MaxIdleTime)
	return nil
}

func defaultPostgresPoolConfig(role string) PostgresPoolConfig {
	config := PostgresPoolConfig{MaxOpen: 24, MaxIdle: 12, MaxLifetime: 30 * time.Minute, MaxIdleTime: 5 * time.Minute}
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "worker":
		config = PostgresPoolConfig{MaxOpen: 8, MaxIdle: 4, MaxLifetime: 30 * time.Minute, MaxIdleTime: 5 * time.Minute}
	case "migrate":
		config = PostgresPoolConfig{MaxOpen: 3, MaxIdle: 1, MaxLifetime: 15 * time.Minute, MaxIdleTime: time.Minute}
	}
	return config
}

func poolEnvInt(getenv func(string) string, name string, fallback, min, max int) (int, error) {
	raw := strings.TrimSpace(getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be an integer between %d and %d", name, min, max)
	}
	return value, nil
}

func poolEnvDuration(getenv func(string) string, name string, fallback, min, max time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(getenv(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value < min || value > max {
		return 0, fmt.Errorf("%s must be a duration between %s and %s", name, min, max)
	}
	return value, nil
}
