package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLicenciasCodigosDescuentoEnvianTokenCSRFAutenticado(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "web", "super", "licencias_codigos_descuento.html"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(page)
	for _, required := range []string{"readCookieValue", "pcs_csrf", "X-CSRF-Token", "credentials:\"same-origin\""} {
		if !strings.Contains(content, required) {
			t.Fatalf("el administrador de codigos de licencia debe enviar CSRF; falta %q", required)
		}
	}
}

func TestActivacionLicenciaSinPagoEnviaTokenCSRFAutenticado(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "web", "pagar_licencia.html"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(page)
	for _, required := range []string{"/licencias/activar_sin_pago", "readCookieValue('pcs_csrf')", "X-CSRF-Token", "headers: freeHeaders"} {
		if !strings.Contains(content, required) {
			t.Fatalf("la activacion sin pago debe enviar CSRF; falta %q", required)
		}
	}
}
