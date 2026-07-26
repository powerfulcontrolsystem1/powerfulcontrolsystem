package handlers

import "testing"

func TestEmpresaAIChatAgentForModeIsServerOwned(t *testing.T) {
	if got := empresaAIChatAgentForMode(false); got != "general" {
		t.Fatalf("modo normal = %q, want general", got)
	}
	if got := empresaAIChatAgentForMode(true); got != "agente_configuracion_de_empresa" {
		t.Fatalf("modo agente = %q, want closed configuration agent", got)
	}
}
