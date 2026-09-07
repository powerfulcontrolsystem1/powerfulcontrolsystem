package db

import "testing"

func TestNormalizeImportacionCosteo(t *testing.T) {
	row := normalizeImportacionCosteo(EmpresaImportacionCosteo{
		Codigo:               " imp-001 ",
		Incoterm:             " fob ",
		MonedaOrigen:         " usd ",
		FechaDocumento:       " 2026-05-08 ",
		FechaEstimadaLlegada: " 2026-05-20 ",
		DocumentoReferencia:  " bl-777 ",
		TRM:                  -10,
		Estado:               "COSTEADO",
	})
	if row.Codigo != "IMP-001" || row.Incoterm != "FOB" || row.MonedaOrigen != "USD" {
		t.Fatalf("normalizacion texto inesperada: %#v", row)
	}
	if row.TRM != 1 {
		t.Fatalf("trm normalizada = %v", row.TRM)
	}
	if row.Estado != "costeado" {
		t.Fatalf("estado normalizado = %q", row.Estado)
	}
	if row.FechaDocumento != "2026-05-08" || row.FechaEstimadaLlegada != "2026-05-20" || row.DocumentoReferencia != "bl-777" {
		t.Fatalf("fechas/referencia no saneadas: %#v", row)
	}
}

func TestNormalizeImportacionItem(t *testing.T) {
	item := normalizeImportacionItem(EmpresaImportacionItem{
		ProductoNombre:      "  Sensor  ",
		SKU:                 " sen-1 ",
		Cantidad:            10,
		CostoUnitarioOrigen: 2.5,
	}, 4000)
	if item.CostoOrigen != 25 || item.CostoBaseCOP != 100000 {
		t.Fatalf("costos calculados = origen %v base %v", item.CostoOrigen, item.CostoBaseCOP)
	}
	if item.CostoUnitarioFinalCOP != 10000 {
		t.Fatalf("unitario final = %v", item.CostoUnitarioFinalCOP)
	}
	if item.SKU != "SEN-1" {
		t.Fatalf("sku = %q", item.SKU)
	}
}

func TestImportacionBaseDistribucionItem(t *testing.T) {
	item := EmpresaImportacionItem{Cantidad: 5, PesoKG: 12, VolumenM3: 0.7, CostoBaseCOP: 240000}
	cases := map[string]float64{"cantidad": 5, "peso": 12, "volumen": 0.7, "valor": 240000, "x": 240000}
	for base, want := range cases {
		if got := importacionBaseDistribucionItem(item, base); got != want {
			t.Fatalf("base %s = %v, want %v", base, got, want)
		}
	}
}

func TestImportacionEstadoAbierto(t *testing.T) {
	for _, estado := range []string{"borrador", "en_transito", "costeado"} {
		if !importacionEstadoAbierto(estado) {
			t.Fatalf("estado %q debe ser abierto", estado)
		}
	}
	for _, estado := range []string{"cerrado", "contabilizado", "anulado"} {
		if importacionEstadoAbierto(estado) {
			t.Fatalf("estado %q no debe ser abierto", estado)
		}
	}
}
