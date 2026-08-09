package db

import "testing"

func TestRefundEmpresaAgenteUsoDiarioRejectsInvalidInputBeforeDatabase(t *testing.T) {
	if err := RefundEmpresaAgenteUsoDiario(nil, EmpresaAgenteUsoDiario{}); err == nil {
		t.Fatal("refund without empresa_id must fail")
	}
	if err := RefundEmpresaAgenteUsoDiario(nil, EmpresaAgenteUsoDiario{EmpresaID: 12, ConsultasAvanzadas: -1}); err == nil {
		t.Fatal("negative advanced refund must fail")
	}
	if err := RefundEmpresaAgenteUsoDiario(nil, EmpresaAgenteUsoDiario{EmpresaID: 12, SegundosUsados: -1}); err == nil {
		t.Fatal("negative seconds refund must fail")
	}
}
