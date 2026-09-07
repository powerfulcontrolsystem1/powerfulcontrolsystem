package handlers

import (
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
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

func TestAutomaticDoorChannelRequiresAuthenticatedRaspberryTunnel(t *testing.T) {
	if !sensorRequiresRaspberryTunnel(&dbpkg.EmpresaSensorDevice{SourceRaspberryID: 12}) {
		t.Fatal("un canal automatico debe rechazar la API publica heredada")
	}
	if sensorRequiresRaspberryTunnel(&dbpkg.EmpresaSensorDevice{}) {
		t.Fatal("un sensor heredado no debe cambiar su contrato por este guard")
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
