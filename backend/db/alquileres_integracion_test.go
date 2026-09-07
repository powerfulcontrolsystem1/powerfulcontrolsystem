package db

import "testing"

func TestAlquilerCoreCodeNormalizaAcentosYPrefijo(t *testing.T) {
	got := alquilerCoreCode("alq contrato", "Alquiler cami\u00f3n 350", "Cliente \u00d1")
	want := "ALQ-CONTRATO-ALQUILER-CAMION-350-CLIENTE-N"
	if got != want {
		t.Fatalf("codigo alquiler = %q, want %q", got, want)
	}
}

func TestAlquilerTarifaPrecioRespetaModalidad(t *testing.T) {
	cases := []struct {
		name string
		in   EmpresaAlquilerTarifa
		want float64
	}{
		{name: "hora", in: EmpresaAlquilerTarifa{ModalidadCobro: "hora", PrecioHora: 12000, PrecioDia: 50000}, want: 12000},
		{name: "semana", in: EmpresaAlquilerTarifa{ModalidadCobro: "semana", PrecioSemana: 280000, PrecioDia: 50000}, want: 280000},
		{name: "evento", in: EmpresaAlquilerTarifa{ModalidadCobro: "evento", PrecioBase: 180000, PrecioDia: 50000}, want: 180000},
		{name: "fallback dia", in: EmpresaAlquilerTarifa{ModalidadCobro: "mes", PrecioDia: 50000}, want: 50000},
	}
	for _, tc := range cases {
		if got := alquilerTarifaPrecio(tc.in); got != tc.want {
			t.Fatalf("%s: precio = %.2f, want %.2f", tc.name, got, tc.want)
		}
	}
}
