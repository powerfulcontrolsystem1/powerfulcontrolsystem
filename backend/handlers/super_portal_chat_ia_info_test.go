package handlers

import (
	"strings"
	"testing"
)

func TestDefaultContextoIALogicaNegocioIncluyeVerticales2026(t *testing.T) {
	ctx := defaultContextoIALogicaNegocioText()

	required := []string{
		"20 verticales 2026",
		"agencia de viajes",
		"transporte de carga/TMS",
		"cooperativa/fondo de empleados",
		"empresa_modulos_colombia",
		"/api/public/verticales_nuevos/catalogo",
		"250 documentos mensuales",
		"4000 documentos mensuales",
	}
	for _, want := range required {
		if !strings.Contains(ctx, want) {
			t.Fatalf("default IA context is missing %q", want)
		}
	}
}
