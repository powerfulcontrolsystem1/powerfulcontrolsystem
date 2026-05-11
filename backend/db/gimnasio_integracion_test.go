package db

import "testing"

func TestGimnasioCoreCodeEsEstableParaServicios(t *testing.T) {
	got := gymCoreCode("GYM-PLAN", "12", "Mensual estándar")
	if got != "GYM-PLAN-12-MENSUAL-EST-NDAR" {
		t.Fatalf("gymCoreCode() = %q", got)
	}
	if len(got) > 51 {
		t.Fatalf("codigo demasiado largo: %q", got)
	}
}

func TestNormalizeGymPagoUsaMetodoPagoCentral(t *testing.T) {
	row, err := normalizeGymPago(EmpresaGimnasioPago{
		EmpresaID:  7,
		SocioID:    3,
		Concepto:   "Mensualidad",
		Monto:      120000,
		MetodoPago: "transferencia",
	})
	if err != nil {
		t.Fatalf("normalizeGymPago() error = %v", err)
	}
	if row.MetodoPago != "transferencia_bancaria" {
		t.Fatalf("MetodoPago = %q", row.MetodoPago)
	}
	if row.Moneda != "COP" || row.Estado != "pagado" {
		t.Fatalf("defaults invalidos: %+v", row)
	}
}
