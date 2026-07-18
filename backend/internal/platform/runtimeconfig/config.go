// Package runtimeconfig centralizes the process role and critical runtime
// switches. It deliberately has no dependency on HTTP or database packages so
// API, migration and worker binaries can validate their startup consistently.
package runtimeconfig

import (
	"fmt"
	"strings"
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
}

// Load accepts an environment accessor to keep production startup logic unit
// testable. Legacy schema bootstrap remains available only as an explicit
// migration role in production; development keeps the historical behavior.
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
	// Production schema ownership is exclusive to the migration role. While the
	// historical catalog is being extracted, the migration role may still run
	// that compatibility bridge, but production must declare the choice
	// explicitly. A missing value must never silently enable legacy DDL.
	legacyBootstrap := !production
	if role == RoleMigrate {
		if production && strings.TrimSpace(getenv("PCS_RUNTIME_SCHEMA_BOOTSTRAP")) == "" {
			return Config{}, fmt.Errorf("PCS_RUNTIME_SCHEMA_BOOTSTRAP is required for the production migrate process")
		}
		legacyBootstrap = !isDisabled(getenv("PCS_RUNTIME_SCHEMA_BOOTSTRAP"))
	} else if !production && isDisabled(getenv("PCS_RUNTIME_SCHEMA_BOOTSTRAP")) {
		legacyBootstrap = false
	}
	if production && role != RoleMigrate && isEnabled(getenv("PCS_RUNTIME_SCHEMA_BOOTSTRAP")) {
		return Config{}, fmt.Errorf("PCS_RUNTIME_SCHEMA_BOOTSTRAP cannot enable DDL in the production %s process", role)
	}
	return Config{Role: role, Production: production, LegacySchemaBootstrap: legacyBootstrap}, nil
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
