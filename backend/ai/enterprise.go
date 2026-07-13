// Package ai contiene contratos puros para el orquestador empresarial.
// No llama proveedores ni ejecuta SQL: esas decisiones siguen siendo del backend.
package ai

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	ModeConsultive = "consultive"
	ModeAssisted   = "assisted"
	ModeAgent      = "agent"

	ToolHotelConfigureRoomStation = "hotel.configure_room_station"
)

// ExecutionContext is server-derived and must never be populated from model output.
type ExecutionContext struct {
	UserID          string   `json:"user_id"`
	EmpresaID       int64    `json:"empresa_id"`
	SedeID          int64    `json:"sede_id,omitempty"`
	Role            string   `json:"role"`
	Permissions     []string `json:"permissions,omitempty"`
	SessionID       string   `json:"session_id,omitempty"`
	ConversationID  string   `json:"conversation_id"`
	RequestID       string   `json:"request_id"`
	Mode            string   `json:"mode"`
	AuthorizedScope []string `json:"authorized_scope,omitempty"`
	MaxOperations   int      `json:"max_operations,omitempty"`
	MaxValue        int64    `json:"max_value,omitempty"`
}

func (c ExecutionContext) Validate() error {
	if strings.TrimSpace(c.UserID) == "" || c.EmpresaID <= 0 {
		return fmt.Errorf("execution context incompleto")
	}
	switch c.Mode {
	case ModeConsultive, ModeAssisted, ModeAgent:
	default:
		return fmt.Errorf("modo de IA invalido")
	}
	return nil
}

// ToolDefinition is a closed, server-owned registry entry.
type ToolDefinition struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	RiskLevel           string   `json:"risk_level"`
	RequiredPermissions []string `json:"required_permissions"`
	TenantScope         string   `json:"tenant_scope"`
	Confirmation        string   `json:"confirmation_policy"`
	Idempotency         string   `json:"idempotency_policy"`
	TimeoutSeconds      int      `json:"timeout_seconds"`
	RateLimitPerMinute  int      `json:"rate_limit_per_minute"`
	AuditCategory       string   `json:"audit_category"`
	Rollback            string   `json:"rollback_strategy"`
}

func Registry() map[string]ToolDefinition {
	return map[string]ToolDefinition{
		ToolHotelConfigureRoomStation: {
			Name:        ToolHotelConfigureRoomStation,
			Description: "Configura una estación como habitación y registra tarifas de hospedaje.",
			RiskLevel:   "medium", RequiredPermissions: []string{"ventas:U"},
			TenantScope: "current_company", Confirmation: "required", Idempotency: "required",
			TimeoutSeconds: 20, RateLimitPerMinute: 6, AuditCategory: "ai_hotel_configuration", Rollback: "transactional_before_commit",
		},
	}
}

func NewOpaqueID(prefix string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return strings.TrimSpace(prefix) + "_" + hex.EncodeToString(buf), nil
}

func CanonicalPlanHash(v interface{}) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func IsProposalExpired(expiresAt time.Time, now time.Time) bool { return !expiresAt.After(now) }

// RedactProviderFields uses a whitelist at the call site; it is a final guard for common sensitive names.
func RedactProviderFields(values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for k, v := range values {
		key := strings.ToLower(strings.TrimSpace(k))
		if strings.Contains(key, "password") || strings.Contains(key, "secret") || strings.Contains(key, "token") || strings.Contains(key, "key") || strings.Contains(key, "certificate") {
			continue
		}
		out[k] = v
	}
	return out
}
