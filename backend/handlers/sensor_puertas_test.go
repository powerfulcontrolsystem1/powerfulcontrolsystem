package handlers

import (
	"strings"
	"testing"
)

func TestBuildEmpresaSensorProvisioningPayload(t *testing.T) {
	payload := buildEmpresaSensorProvisioningPayload(" RPI Mesa 1 ", "tok_123")
	if payload["device_id"] != "rpi-mesa-1" {
		t.Fatalf("device_id normalizado inesperado: %v", payload["device_id"])
	}
	curl, _ := payload["curl"].(string)
	if curl == "" || !containsAll(curl, []string{"X-Device-Token: tok_123", "rpi-mesa-1", "heartbeat"}) {
		t.Fatalf("curl de provisionamiento incompleto: %q", curl)
	}
	python, _ := payload["python"].(string)
	if python == "" || !containsAll(python, []string{"tok_123", "rpi-mesa-1", "requests.post"}) {
		t.Fatalf("python de provisionamiento incompleto: %q", python)
	}
}

func containsAll(value string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
