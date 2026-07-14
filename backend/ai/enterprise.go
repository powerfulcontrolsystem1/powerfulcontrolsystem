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
	ToolHotelInspectRoomStation   = "hotel.inspect_room_station"
	ToolCatalogSearchProducts     = "catalog.search_products"
	ToolCatalogCreateProduct      = "catalog.create_product"
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

// AllowsAgentMode is intentionally separate from the selected UI mode. A client
// can ask for agent mode, but the server must still enable it explicitly.
func AllowsAgentMode(enabled bool, c ExecutionContext) bool {
	if c.Mode != ModeAgent {
		return true
	}
	return enabled && c.MaxOperations > 0 && len(c.AuthorizedScope) > 0
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
	Module              string   `json:"module"`
	EnabledByDefault    bool     `json:"enabled_by_default"`
}

func Registry() map[string]ToolDefinition {
	return map[string]ToolDefinition{
		ToolHotelInspectRoomStation: {
			Name:        ToolHotelInspectRoomStation,
			Description: "Consulta la configuracion y tarifas actuales de una estacion hotelera.",
			RiskLevel:   "low", RequiredPermissions: []string{"reservas_hotel:R"},
			TenantScope: "current_company", Confirmation: "none", Idempotency: "not_applicable",
			TimeoutSeconds: 10, RateLimitPerMinute: 20, AuditCategory: "ai_hotel_read", Rollback: "not_applicable", Module: "reservas_hotel", EnabledByDefault: true,
		},
		ToolHotelConfigureRoomStation: {
			Name:        ToolHotelConfigureRoomStation,
			Description: "Configura una estación como habitación y registra tarifas de hospedaje.",
			RiskLevel:   "medium", RequiredPermissions: []string{"ventas:U"},
			TenantScope: "current_company", Confirmation: "required", Idempotency: "required",
			TimeoutSeconds: 20, RateLimitPerMinute: 6, AuditCategory: "ai_hotel_configuration", Rollback: "transactional_before_commit", Module: "reservas_hotel", EnabledByDefault: false,
		},
		ToolCatalogSearchProducts: {
			Name:        ToolCatalogSearchProducts,
			Description: "Consulta productos, categorias y bodegas de la empresa activa.",
			RiskLevel:   "read", RequiredPermissions: []string{"inventario:R"},
			TenantScope: "current_company", Confirmation: "none", Idempotency: "not_applicable",
			TimeoutSeconds: 10, RateLimitPerMinute: 20, AuditCategory: "ai_catalog_read", Rollback: "not_applicable", Module: "inventario", EnabledByDefault: true,
		},
		ToolCatalogCreateProduct: {
			Name:        ToolCatalogCreateProduct,
			Description: "Crea un producto con categoria, bodega y stock inicial previamente validados.",
			RiskLevel:   "medium", RequiredPermissions: []string{"inventario:C"},
			TenantScope: "current_company", Confirmation: "required", Idempotency: "required",
			TimeoutSeconds: 20, RateLimitPerMinute: 6, AuditCategory: "ai_catalog_create", Rollback: "transactional_before_commit", Module: "inventario", EnabledByDefault: false,
		},
	}
}

// ToolAllowed verifies the server-derived permission snapshot. The model and
// browser never supply permissions; they only receive the resulting catalog.
func ToolAllowed(def ToolDefinition, granted []string) bool {
	if len(def.RequiredPermissions) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(granted))
	for _, permission := range granted {
		set[strings.TrimSpace(permission)] = struct{}{}
	}
	for _, required := range def.RequiredPermissions {
		if _, ok := set[strings.TrimSpace(required)]; !ok {
			return false
		}
	}
	return true
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
