package db

import "testing"

func TestEmpresaRetencionEstadoNoOperativo(t *testing.T) {
	activeStates := []string{"", "activo", "activa", "operando"}
	for _, estado := range activeStates {
		if empresaRetencionEstadoNoOperativo(estado) {
			t.Fatalf("estado %q no debe ser candidato de retencion", estado)
		}
	}

	inactiveStates := []string{"inactivo", "suspendida", "bloqueado", "vencida", "deshabilitado"}
	for _, estado := range inactiveStates {
		if !empresaRetencionEstadoNoOperativo(estado) {
			t.Fatalf("estado %q debe ser candidato de retencion", estado)
		}
	}
}
