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
