package db

import "testing"

func TestCalculateEmpresaAIUContratoModeloNoSumado(t *testing.T) {
	got := CalculateEmpresaAIUContrato(EmpresaAIUContrato{
		CostoDirecto:          10000000,
		ModeloAIU:             "base_aiu_no_sumada",
		BaseIVAModo:           "utilidad",
		PorcentajeAdmin:       10,
		PorcentajeImprevistos: 5,
		PorcentajeUtilidad:    10,
		PorcentajeIVA:         19,
	})
	if got.ValorAdministracion != 1000000 || got.ValorImprevistos != 500000 || got.ValorUtilidad != 1000000 {
		t.Fatalf("AIU inesperado: %#v", got)
	}
	if got.BaseIVA != 1000000 || got.ValorIVA != 190000 || got.TotalFactura != 10190000 {
		t.Fatalf("totales inesperados no sumado: %#v", got)
	}
}

func TestCalculateEmpresaAIUContratoModeloSumado(t *testing.T) {
	got := CalculateEmpresaAIUContrato(EmpresaAIUContrato{
		CostoDirecto:          10000000,
		ModeloAIU:             "base_aiu_sumada",
		BaseIVAModo:           "aiu_total",
		PorcentajeAdmin:       10,
		PorcentajeImprevistos: 5,
		PorcentajeUtilidad:    10,
		PorcentajeIVA:         19,
	})
	if got.AIUTotal != 2500000 || got.BaseIVA != 2500000 || got.ValorIVA != 475000 || got.TotalFactura != 12975000 {
		t.Fatalf("totales inesperados sumado: %#v", got)
	}
}

func TestCalculateEmpresaAIUContratoRetencionesYNeto(t *testing.T) {
	got := CalculateEmpresaAIUContrato(EmpresaAIUContrato{
		CostoDirecto:          2500000,
		ModeloAIU:             "base_aiu_sumada",
		BaseIVAModo:           "utilidad",
		PorcentajeAdmin:       10,
		PorcentajeImprevistos: 5,
		PorcentajeUtilidad:    8,
		PorcentajeIVA:         19,
		PorcentajeRetFuente:   2,
		PorcentajeRetICA:      1,
		PorcentajeRetIVA:      15,
		PorcentajeAnticipo:    10,
		PorcentajeGarantia:    5,
	})
	if got.TotalFactura != 3113000 || got.ValorRetFuente != 4000 || got.ValorRetICA != 31130 || got.ValorRetIVA != 5700 {
		t.Fatalf("retenciones AIU inesperadas: %#v", got)
	}
	if got.ValorAnticipo != 311300 || got.ValorGarantia != 155650 || got.NetoCobrar != 2605220 {
		t.Fatalf("neto AIU inesperado: %#v", got)
	}
}

func TestNormalizeEmpresaAIUContrato(t *testing.T) {
	got := NormalizeEmpresaAIUContrato(EmpresaAIUContrato{Codigo: " obra-1 ", CentroCosto: " cc-9 ", ModalidadContrato: "administracion delegada", RiesgoNivel: "ALTO", ModeloAIU: "sumada", BaseIVAModo: "aiu", Estado: "EN EJECUCION", PorcentajeAdmin: 120, PorcentajeIVA: -1, PorcentajeRetICA: 101})
	if got.Codigo != "OBRA-1" || got.ModeloAIU != "base_aiu_sumada" || got.BaseIVAModo != "aiu_total" || got.Estado != "en_ejecucion" {
		t.Fatalf("normalizacion texto inesperada: %#v", got)
	}
	if got.CentroCosto != "CC-9" || got.ModalidadContrato != "administracion_delegada" || got.RiesgoNivel != "alto" {
		t.Fatalf("normalizacion profesional inesperada: %#v", got)
	}
	if got.PorcentajeAdmin != 100 || got.PorcentajeIVA != 0 || got.PorcentajeRetICA != 100 {
		t.Fatalf("normalizacion porcentajes inesperada: %#v", got)
	}
}

func TestAIUEstadoTransitionAllowed(t *testing.T) {
	if !aiuEstadoTransitionAllowed("aprobado", "en_ejecucion") {
		t.Fatal("aprobado debe poder pasar a en_ejecucion")
	}
	if aiuEstadoTransitionAllowed("facturado", "anulado") {
		t.Fatal("facturado no debe poder anularse desde AIU")
	}
	if aiuEstadoTransitionAllowed("cerrado", "en_ejecucion") {
		t.Fatal("cerrado no debe reabrirse desde AIU")
	}
}
