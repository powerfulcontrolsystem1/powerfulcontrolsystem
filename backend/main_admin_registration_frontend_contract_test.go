package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdminRegistrationKeepsFailedEmailRetryVisible(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "web", "js", "registrar_nuevo_usuario_administrador.js"))
	if err != nil {
		t.Fatalf("read admin registration frontend: %v", err)
	}
	script := string(raw)
	for _, marker := range []string{
		"response.json.email_sent === false",
		"if (!emailWasSent && !invitationToken)",
		"return;",
	} {
		if !strings.Contains(script, marker) {
			t.Fatalf("admin registration frontend missing retry marker %q", marker)
		}
	}
}
