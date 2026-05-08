package db

import "testing"

func TestNormalizeCRMMetaComercial(t *testing.T) {
	meta := normalizeCRMMetaComercial(EmpresaCRMMetaComercial{
		Canal:             " Web ",
		MetaValor:         -100,
		MetaLeads:         -3,
		MetaConversionPct: 140,
		Estado:            "bad",
	})
	if meta.Periodo == "" || meta.Canal != "web" || meta.MetaValor != 0 || meta.MetaLeads != 0 || meta.MetaConversionPct != 100 || meta.Estado != "activo" {
		t.Fatalf("meta normalizada inesperada: %+v", meta)
	}
}

func TestCRMLeadScoreRecommendation(t *testing.T) {
	lead := EmpresaCRMLeadScore{EstadoLead: "propuesta", ValorPotencial: 6000000, Probabilidad: 55, Interacciones: 4}
	lead.Score = crmLeadScore(lead)
	if lead.Score != 95 {
		t.Fatalf("score inesperado: %v", lead.Score)
	}
	if got := crmLeadRecommendation(lead); got != "priorizar_cierre" {
		t.Fatalf("recomendacion inesperada: %s", got)
	}
}

func TestCRMConversionPctAndAlerts(t *testing.T) {
	if got := crmConversionPct(3, 1); got != 75 {
		t.Fatalf("conversion inesperada: %v", got)
	}
	alertas := buildEmpresaCRMAlertas(EmpresaCRMVentasAvanzadasDashboard{
		LeadsActivos:        0,
		LeadsVencidos:       2,
		MetaValor:           100,
		CumplimientoMetaPct: 50,
		CampanasActivas:     0,
	})
	if len(alertas) != 4 {
		t.Fatalf("alertas inesperadas: %+v", alertas)
	}
}
