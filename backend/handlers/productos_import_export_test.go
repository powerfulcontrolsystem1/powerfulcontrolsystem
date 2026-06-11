package handlers

import (
	"bytes"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestProductosImportHeaderNormalization(t *testing.T) {
	headers := normalizeProductoImportHeaders([]string{
		"SKU",
		"Codigo de barras",
		"Nombre",
		"Unidad medida",
		"Impuesto %",
		"Stock inicial",
		"Bodega principal id",
	})
	expected := map[string]int{
		"sku":                 0,
		"codigo_barras":       1,
		"nombre":              2,
		"unidad_medida":       3,
		"impuesto_porcentaje": 4,
		"stock_inicial":       5,
		"bodega_principal_id": 6,
	}
	for key, want := range expected {
		if got, ok := headers[key]; !ok || got != want {
			t.Fatalf("header %s got (%d,%v), want %d", key, got, ok, want)
		}
	}
}

func TestProductosImportFloatParser(t *testing.T) {
	cases := map[string]float64{
		"1000":     1000,
		"1,25":     1.25,
		"$1,200.5": 1200.5,
		"1.200,5":  1200.5,
		"":         0,
	}
	for raw, want := range cases {
		if got := parseProductoImportFloat(raw); got != want {
			t.Fatalf("parseProductoImportFloat(%q) = %v, want %v", raw, got, want)
		}
	}
}

func TestProductosExportWriters(t *testing.T) {
	rows := []dbpkg.Producto{{
		ID:            12,
		EmpresaID:     7,
		SKU:           "SKU-1",
		Nombre:        "Menta",
		Categoria:     "Dulceria",
		Precio:        100,
		StockTotal:    3,
		UnidadMedida:  "unidad",
		CodigoBarras:  "7701",
		Observaciones: "ok",
	}}
	var csvOut bytes.Buffer
	if err := writeProductosCSV(&csvOut, rows); err != nil {
		t.Fatalf("writeProductosCSV: %v", err)
	}
	if body := csvOut.String(); !strings.Contains(body, "SKU-1") || !strings.Contains(body, "stock_total") {
		t.Fatalf("csv export incompleto: %s", body)
	}
	var htmlOut bytes.Buffer
	if err := writeProductosPrintHTML(&htmlOut, rows, "pos"); err != nil {
		t.Fatalf("writeProductosPrintHTML: %v", err)
	}
	if body := htmlOut.String(); !strings.Contains(body, "@page{size:80mm auto") || !strings.Contains(body, "Menta") {
		t.Fatalf("html POS incompleto: %s", body)
	}
}
