package handlers

import "testing"

func TestFacturacionDIANLegacySignedXMLNeedsManualRegeneration(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		tipo string
		want bool
	}{
		{
			name: "invoice abbreviated profile",
			xml:  `<Invoice><cbc:ProfileID>DIAN 2.1</cbc:ProfileID><cac:StandardItemIdentification><cbc:ID>SKU-1</cbc:ID></cac:StandardItemIdentification></Invoice>`,
			tipo: "factura_electronica",
			want: true,
		},
		{
			name: "invoice missing standard item identification",
			xml:  `<Invoice><cbc:ProfileID>DIAN 2.1: Factura Electrónica de Venta</cbc:ProfileID></Invoice>`,
			tipo: "factura_electronica",
			want: true,
		},
		{
			name: "current invoice",
			xml:  `<Invoice><cbc:ProfileID>DIAN 2.1: Factura Electrónica de Venta</cbc:ProfileID><cac:StandardItemIdentification><cbc:ID>SKU-1</cbc:ID></cac:StandardItemIdentification></Invoice>`,
			tipo: "factura_electronica",
			want: false,
		},
		{
			name: "malformed xml stays fail closed",
			xml:  `<Invoice><cbc:ProfileID>DIAN 2.1</Invoice>`,
			tipo: "factura_electronica",
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := facturacionDIANLegacySignedXMLNeedsManualRegeneration(tc.xml, tc.tipo); got != tc.want {
				t.Fatalf("got %v; want %v", got, tc.want)
			}
		})
	}
}
