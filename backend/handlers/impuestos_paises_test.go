package handlers

import "testing"

func TestImpuestosCatalogoBaseIncluyePaisesFacturacionRegional(t *testing.T) {
	cases := []struct {
		pais     string
		codigo   string
		tasa     float64
		enabled  int
		minItems int
	}{
		{pais: "CR", codigo: "IVA_13", tasa: 13, enabled: 1, minItems: 5},
		{pais: "AR", codigo: "IVA_21", tasa: 21, enabled: 1, minItems: 6},
		{pais: "VE", codigo: "IVA_16", tasa: 16, enabled: 1, minItems: 5},
		{pais: "EC", codigo: "IVA", tasa: 15, enabled: 1, minItems: 5},
		{pais: "PA", codigo: "ITBMS_7", tasa: 7, enabled: 1, minItems: 5},
	}
	for _, tc := range cases {
		t.Run(tc.pais, func(t *testing.T) {
			got := impuestosCatalogoBase(tc.pais)
			if len(got) < tc.minItems {
				t.Fatalf("expected at least %d taxes for %s, got %d", tc.minItems, tc.pais, len(got))
			}
			for _, item := range got {
				if item.Codigo == tc.codigo {
					if item.PaisCodigo != tc.pais || item.TasaPorcentaje != tc.tasa || item.Habilitado != tc.enabled {
						t.Fatalf("unexpected tax item for %s: %#v", tc.pais, item)
					}
					return
				}
			}
			t.Fatalf("tax %s not found in %#v", tc.codigo, got)
		})
	}
}
