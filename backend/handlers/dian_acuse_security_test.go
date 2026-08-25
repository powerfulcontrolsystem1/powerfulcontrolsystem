package handlers

import "testing"

func TestNormalizeDIANAcuseEstadoRejectsNegatedPositiveWords(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		"no validado",
		"No aceptado por DIAN",
		"validacion fallida",
		"documento invalidado",
		"not accepted",
	} {
		if got := normalizeDIANAcuseEstado(raw); got != "rechazado" {
			t.Fatalf("normalizeDIANAcuseEstado(%q) = %q; want rechazado", raw, got)
		}
	}
}

func TestResolveDIANAcuseDoesNotTreatGenericSuccessFlagsAsFiscalAcceptance(t *testing.T) {
	t.Parallel()
	for _, response := range []map[string]interface{}{
		{"ok": true},
		{"success": true},
		{"accepted": true},
	} {
		estado, _ := resolveDIANAcuseFromResponse(200, response)
		if estado != "enviado" {
			t.Fatalf("generic provider flag produced estado=%q; want enviado pending official acuse", estado)
		}
	}
}

func TestResolveDIANAcuseRequiresOfficialValidityForAcceptance(t *testing.T) {
	t.Parallel()
	estado, _ := resolveDIANAcuseFromResponse(200, map[string]interface{}{
		"is_valid":    true,
		"status_code": "00",
	})
	if estado != "aceptado" {
		t.Fatalf("official validity produced estado=%q; want aceptado", estado)
	}
}
