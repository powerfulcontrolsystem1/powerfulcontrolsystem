package handlers

import "testing"

func TestEmpresaAIChatAgentForModeIsServerOwned(t *testing.T) {
	if got := empresaAIChatAgentForMode(false); got != "agente_pcs" {
		t.Fatalf("crafted normal mode = %q, want permanent PCS agent", got)
	}
	if got := empresaAIChatAgentForMode(true); got != "agente_pcs" {
		t.Fatalf("agent mode = %q, want permanent PCS agent", got)
	}
}
