package db

import "testing"

func TestEmpresaCarteraCXPEdadRango(t *testing.T) {
	cases := []struct {
		name      string
		venc      string
		corte     string
		wantRango string
	}{
		{name: "por vencer", venc: "2026-06-20", corte: "2026-06-12", wantRango: "por_vencer"},
		{name: "vencido 30", venc: "2026-05-20", corte: "2026-06-12", wantRango: "0_30"},
		{name: "vencido 60", venc: "2026-04-20", corte: "2026-06-12", wantRango: "31_60"},
		{name: "vencido 90", venc: "2026-03-20", corte: "2026-06-12", wantRango: "61_90"},
		{name: "vencido 180", venc: "2026-01-20", corte: "2026-06-12", wantRango: "91_180"},
		{name: "vencido mayor", venc: "2025-01-20", corte: "2026-06-12", wantRango: "181_mas"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := empresaCarteraCXPEdadRango(tc.venc, tc.corte)
			if got != tc.wantRango {
				t.Fatalf("rango inesperado got=%s want=%s", got, tc.wantRango)
			}
		})
	}
}

func TestNormalizeEmpresaCarteraCXPEstado(t *testing.T) {
	if got := normalizeEmpresaCarteraCXPEstado("pendiente", 0, "2026-06-12"); got != "pagado" {
		t.Fatalf("saldo cero debe quedar pagado, got=%s", got)
	}
	if got := normalizeEmpresaCarteraCXPEstado("anulado", 100, "2026-06-12"); got != "anulado" {
		t.Fatalf("estado anulado debe conservarse, got=%s", got)
	}
}
