package db

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeAyudaContextoJSONKeepsOnlySafeKeys(t *testing.T) {
	raw := map[string]interface{}{
		"titulo":       "Caja principal",
		"modulo":       "estaciones",
		"cookie":       "session=secret",
		"localStorage": "token",
		"password":     "clave",
	}

	normalized := normalizeAyudaContextoJSON(raw)
	if normalized == "" {
		t.Fatal("expected normalized context")
	}
	if strings.Contains(normalized, "secret") || strings.Contains(normalized, "token") || strings.Contains(normalized, "clave") {
		t.Fatalf("unsafe values leaked in normalized context: %s", normalized)
	}

	var decoded map[string]string
	if err := json.Unmarshal([]byte(normalized), &decoded); err != nil {
		t.Fatalf("normalized context is not valid JSON: %v", err)
	}
	if decoded["titulo"] != "Caja principal" || decoded["modulo"] != "estaciones" {
		t.Fatalf("safe context keys were not preserved: %#v", decoded)
	}
	if _, ok := decoded["cookie"]; ok {
		t.Fatalf("unsafe key cookie was preserved: %#v", decoded)
	}
}

func TestNormalizeAyudaContactoPreferidoDefaultsToEmail(t *testing.T) {
	if got := normalizeAyudaContactoPreferido("whatsapp"); got != "whatsapp" {
		t.Fatalf("expected whatsapp, got %q", got)
	}
	if got := normalizeAyudaContactoPreferido("sms"); got != "email" {
		t.Fatalf("expected invalid preference to default to email, got %q", got)
	}
}
