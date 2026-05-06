package handlers

import "testing"

func TestNormalizeFacturacionDocumentoElectronicoTipoIncluyeDocumentosSiigoDian(t *testing.T) {
	cases := map[string]string{
		"factura":                         "factura_electronica",
		"nota debito ventas":              "nota_debito",
		"documento soporte adquisiciones": "documento_soporte",
		"documento soporte de pago nomina electronica": "nomina_electronica",
		"tiquete maquina registradora pos":             "documento_equivalente_pos",
		"documento equivalente electronico POS":        "documento_equivalente_pos",
		"nota credito":                                 "nota_credito",
	}
	for raw, want := range cases {
		if got := normalizeFacturacionDocumentoElectronicoTipo(raw); got != want {
			t.Fatalf("normalizeFacturacionDocumentoElectronicoTipo(%q)=%q, want %q", raw, got, want)
		}
	}
}

func TestResolveFacturacionTransitionForDocumentosElectronicosNuevos(t *testing.T) {
	cases := []struct {
		name          string
		action        string
		tipoDocumento string
		wantAccion    string
		wantEstado    string
		wantEvento    string
	}{
		{name: "nota debito", action: "nota_debito", tipoDocumento: "nota_debito", wantAccion: "nota_debito", wantEstado: "emitida", wantEvento: "nota_debito_emitida"},
		{name: "documento soporte", action: "documento_soporte", tipoDocumento: "documento_soporte", wantAccion: "documento_soporte", wantEstado: "emitida", wantEvento: "documento_soporte_emitido"},
		{name: "nomina electronica", action: "nomina_electronica", tipoDocumento: "nomina_electronica", wantAccion: "nomina_electronica", wantEstado: "emitida", wantEvento: "nomina_electronica_emitida"},
		{name: "pos electronico", action: "documento_equivalente_pos", tipoDocumento: "documento_equivalente_pos", wantAccion: "documento_equivalente_pos", wantEstado: "emitida", wantEvento: "documento_equivalente_pos_emitido"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveFacturacionTransitionForDocument(tc.action, "borrador", tc.tipoDocumento)
			if err != nil {
				t.Fatalf("resolve transition returned error: %v", err)
			}
			if got.Accion != tc.wantAccion || got.EstadoNuevo != tc.wantEstado || got.Evento != tc.wantEvento {
				t.Fatalf("unexpected transition: %#v", got)
			}
		})
	}
}
