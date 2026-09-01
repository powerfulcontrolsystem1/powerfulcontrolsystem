package handlers

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"text/template"
)

func TestDomoticaTunnelPublicBaseURLRequiresHTTPS(t *testing.T) {
	t.Setenv("PCS_DOMOTICA_PUBLIC_BASE_URL", "http://example.test/pcs?secret=discard")
	t.Setenv("PCS_PUBLIC_BASE_URL", "")
	t.Setenv("PUBLIC_BASE_URL", "")
	if _, err := domoticaTunnelPublicBaseURL(); err == nil {
		t.Fatal("una URL publica HTTP debe rechazarse")
	}
}

func TestDomoticaTunnelPublicBaseURLRejectsEmbeddedCredentials(t *testing.T) {
	t.Setenv("PCS_DOMOTICA_PUBLIC_BASE_URL", "https://user:secret@pcs.example.test")
	if _, err := domoticaTunnelPublicBaseURL(); err == nil {
		t.Fatal("la URL publica no debe incorporar credenciales")
	}
}

func TestDomoticaTunnelPublicBaseURLNormalizesHTTPS(t *testing.T) {
	t.Setenv("PCS_DOMOTICA_PUBLIC_BASE_URL", "https://pcs.example.test/base/?secret=discard#fragment")
	got, err := domoticaTunnelPublicBaseURL()
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://pcs.example.test/base" {
		t.Fatalf("URL normalizada = %q", got)
	}
}

func TestDomoticaInstallerTemplateIsSelfContainedAndSecretSafe(t *testing.T) {
	baseJSON, _ := json.Marshal("https://pcs.example.test")
	deviceJSON, _ := json.Marshal("RPI-0123456789ABCDEF0123456789ABCDEF")
	tokenJSON, _ := json.Marshal("one-time-enrollment-token")
	tmpl, err := template.New("installer").Parse(domoticaRaspberryInstallerTemplate)
	if err != nil {
		t.Fatal(err)
	}
	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, domoticaInstallerTemplateData{
		BaseURLJSON:         string(baseJSON),
		DeviceUIDJSON:       string(deviceJSON),
		EnrollmentTokenJSON: string(tokenJSON),
	}); err != nil {
		t.Fatal(err)
	}
	body := rendered.String()
	for _, marker := range []string{
		"#!/bin/sh",
		"/api/public/domotica/tunnel",
		"pcs-domotica-agent.service",
		"systemctl enable --now",
		"chmod 0600 /etc/pcs-domotica/agent.json",
		"NoNewPrivileges=true",
		"StartLimitIntervalSec=0",
		"Restart=always",
		"backoff = min(30, backoff * 2)",
		"boot_id = os.urandom(16).hex()",
		"restore_delay_ms",
		`"agent_version":"1.3.0"`,
		"monitor_door_sensors",
		`request_json(config, "door_scan"`,
		"for pin in range(4, 20)",
		"gpio.set_output(output_pin, True, True)",
		"gpio.set_output(output_pin, False, True)",
		"one-time-enrollment-token",
	} {
		if !strings.Contains(body, marker) {
			t.Fatalf("instalador sin marcador requerido %q", marker)
		}
	}
	if strings.Contains(body, "{{.") || strings.Contains(body, "empresa_id") {
		t.Fatal("el instalador no debe conservar placeholders ni una empresa_id manipulable")
	}
}
