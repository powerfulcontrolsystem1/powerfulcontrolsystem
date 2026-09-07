package handlers

import (
	"strings"
	"testing"
)

func TestBuildEmpresaUsuarioMultipartMessageEmbedsCorporateLogo(t *testing.T) {
	msg := buildEmpresaUsuarioMultipartMessage(
		nil,
		"https://powerfulcontrolsystem.com",
		"Powerful Control System - Soporte",
		"soporte@powerfulcontrolsystem.com",
		"cliente@example.com",
		"Prueba PCS",
		"Texto de prueba",
		"<p>Texto de prueba</p>",
	)

	required := []string{
		"Content-Type: multipart/related",
		"Content-Type: multipart/alternative",
		"Content-ID: <pcs-logo-",
		"Content-Disposition: inline; filename=\"pcs-logo\"",
		"src=\"cid:pcs-logo-",
	}
	for _, needle := range required {
		if !strings.Contains(msg, needle) {
			t.Fatalf("mensaje corporativo sin %q\n%s", needle, msg)
		}
	}
	if strings.Contains(msg, "src=\"https://powerfulcontrolsystem.com/img/Logo pcs 1.png\"") {
		t.Fatalf("el logo no debe depender de URL externa cuando existe archivo local")
	}
}
