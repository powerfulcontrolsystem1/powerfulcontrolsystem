package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestMailuRelayUsesPrivateNetwork(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "docker-compose.platform.yml"))
	if err != nil {
		t.Fatalf("read production compose: %v", err)
	}
	compose := string(raw)

	for _, service := range []string{"backend:", "worker:"} {
		start := strings.Index(compose, "\n  "+service)
		if start < 0 {
			t.Fatalf("service %s not found", service)
		}
		section := compose[start+1:]
		headerEnd := strings.Index(section, "\n")
		if headerEnd < 0 {
			t.Fatalf("service %s has no body", service)
		}
		section = section[headerEnd+1:]
		if next := regexp.MustCompile(`(?m)^  [a-zA-Z0-9_-]+:\s*$`).FindStringIndex(section); next != nil {
			section = section[:next[0]]
		}
		for _, marker := range []string{
			`PCS_MAILU_SMTP_ADDR: ${MAILU_SMTP_IP:-192.168.203.3}:25`,
			"mailu_internal:",
		} {
			if !strings.Contains(section, marker) {
				t.Fatalf("service %s missing private Mailu relay marker %q", service, marker)
			}
		}
	}
}

func TestBIMILogoDeclaresTinyPSProfile(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "web", "img", "bimi-pcs.svg"))
	if err != nil {
		t.Fatalf("read BIMI logo: %v", err)
	}
	svg := string(raw)
	for _, marker := range []string{
		`version="1.2"`,
		`baseProfile="tiny-ps"`,
		`<title`,
		`<desc`,
	} {
		if !strings.Contains(svg, marker) {
			t.Fatalf("BIMI logo missing %q", marker)
		}
	}
}
