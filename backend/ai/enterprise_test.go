package ai

import (
	"testing"
	"time"
)

func TestEnterpriseContextRejectsClientlessTenant(t *testing.T) {
	if err := (ExecutionContext{UserID: "user", EmpresaID: 0, Mode: ModeAssisted}).Validate(); err == nil {
		t.Fatal("expected invalid tenant context")
	}
}

func TestOpaqueProposalIDsAreUnpredictableAndDistinct(t *testing.T) {
	a, err := NewOpaqueID("proposal")
	if err != nil {
		t.Fatal(err)
	}
	b, err := NewOpaqueID("proposal")
	if err != nil {
		t.Fatal(err)
	}
	if a == b || len(a) < 70 {
		t.Fatalf("invalid opaque IDs: %q %q", a, b)
	}
}

func TestPlanHashAndRedaction(t *testing.T) {
	h, err := CanonicalPlanHash(map[string]interface{}{"tool": ToolHotelConfigureRoomStation, "station": 1})
	if err != nil || len(h) != 64 {
		t.Fatalf("hash=%q err=%v", h, err)
	}
	data := RedactProviderFields(map[string]string{"nombre": "Habitacion 1", "api_token": "private"})
	if data["nombre"] == "" || data["api_token"] != "" {
		t.Fatalf("redaction failed: %#v", data)
	}
	if !IsProposalExpired(time.Now().Add(-time.Second), time.Now()) {
		t.Fatal("expired proposal accepted")
	}
}

func TestAgentModeRequiresExplicitServerEnablementAndLimits(t *testing.T) {
	ctx := ExecutionContext{UserID: "user", EmpresaID: 1, Mode: ModeAgent, AuthorizedScope: []string{"current_company"}, MaxOperations: 1}
	if AllowsAgentMode(false, ctx) {
		t.Fatal("agent mode must fail closed")
	}
	if !AllowsAgentMode(true, ctx) {
		t.Fatal("enabled agent mode with a bounded scope was rejected")
	}
}

func TestHotelReadToolIsClosedAndLowRisk(t *testing.T) {
	tool, ok := Registry()[ToolHotelInspectRoomStation]
	if !ok {
		t.Fatal("hotel read tool missing from closed registry")
	}
	if tool.Confirmation != "none" || tool.RiskLevel != "low" || tool.TenantScope != "current_company" || len(tool.RequiredPermissions) != 1 {
		t.Fatalf("unexpected hotel read policy: %#v", tool)
	}
}

func TestToolPermissionIsServerScoped(t *testing.T) {
	tool := Registry()[ToolCatalogCreateProduct]
	if ToolAllowed(tool, []string{"inventario:R"}) {
		t.Fatal("read permission must not create catalog records")
	}
	if !ToolAllowed(tool, []string{"inventario:C", "ventas:R"}) {
		t.Fatal("required inventory create permission was rejected")
	}
}
