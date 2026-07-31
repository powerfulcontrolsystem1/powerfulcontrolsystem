package handlers

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReportesCatalogoIncluyeReportesColombianosAvanzados(t *testing.T) {
	required := []string{
		reporteDatasetVentasDiariasMetodoPago,
		reporteDatasetVentasRentabilidadProducto,
		reporteDatasetOperativoTurno,
		reporteDatasetInventarioKardexValorizado,
		reporteDatasetComprasProveedorDetalle,
		reporteDatasetContableBalancePrueba,
		reporteDatasetContableLibroAuxiliar,
		reporteDatasetContableLibroMayor,
		reporteDatasetFiscalImpuestosRetenciones,
		reporteDatasetFiscalInformacionExogena,
		reporteDatasetCarteraEdadesCobrar,
		reporteDatasetCarteraEdadesPagar,
	}

	seen := make(map[string]empresaReporteCatalogoItem)
	for _, item := range reportesCatalogo {
		if item.Key == "" {
			t.Fatalf("reporte con key vacio: %#v", item)
		}
		if _, exists := seen[item.Key]; exists {
			t.Fatalf("reporte duplicado en catalogo: %s", item.Key)
		}
		if len(item.Formats) < 5 {
			t.Fatalf("reporte %s no declara todos los formatos profesionales: %#v", item.Key, item.Formats)
		}
		seen[item.Key] = item
	}

	for _, key := range required {
		item, ok := seen[key]
		if !ok {
			t.Fatalf("falta reporte colombiano avanzado en catalogo: %s", key)
		}
		if item.Title == "" || item.Description == "" || item.Level == "" {
			t.Fatalf("reporte %s sin metadatos completos: %#v", key, item)
		}
	}
}

func TestReportesDatasetPDFAmplioNoRecortaColumnasNiFilas(t *testing.T) {
	columns := []string{
		"codigo", "tercero", "documento_tipo", "documento_codigo", "fecha_emision",
		"fecha_vencimiento", "dias_mora", "valor_original", "valor_pagado", "saldo", "estado_cartera",
	}
	rows := make([]map[string]interface{}, 60)
	for i := range rows {
		rows[i] = map[string]interface{}{
			"codigo":            fmt.Sprintf("CXP-%03d", i+1),
			"tercero":           "Proveedor de validacion con nombre suficientemente extenso",
			"documento_tipo":    "cuenta_por_pagar",
			"documento_codigo":  fmt.Sprintf("DOCUMENTO-%03d", i+1),
			"fecha_emision":     "2026-07-31",
			"fecha_vencimiento": "2026-08-31",
			"dias_mora":         0,
			"valor_original":    102.00,
			"valor_pagado":      27.01,
			"saldo":             74.99,
			"estado_cartera":    "pendiente",
		}
	}
	ds := empresaReporteDataset{
		Key: "contable_cuentas_por_pagar", Title: "Cuentas por Pagar", Level: "contable",
		EmpresaID: 12, Desde: "2026-07-01", Hasta: "2026-07-31", GeneratedAt: "2026-07-31 05:41:23",
		Columns: columns, Rows: rows,
	}

	lines := reportesDatasetPDFLines(ds)
	for _, line := range lines {
		if got := len([]rune(line)); got > 84 {
			t.Fatalf("linea PDF de %d caracteres excede el ancho: %q", got, line)
		}
	}
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Registro 1", "codigo: CXP-001", "estado cartera: pendiente", "Registro 60", "codigo: CXP-060"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("PDF amplio no conserva %q", want)
		}
	}
	if strings.Contains(joined, "Salida truncada") {
		t.Fatal("el PDF no debe truncar filas")
	}
	if strings.Contains(joined, "estado_cartera:") || strings.Contains(joined, "documento_codigo:") {
		t.Fatal("las etiquetas PDF no deben exponer nombres técnicos con guion bajo")
	}

	pdf := reportesDatasetPDFContent(ds)
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		t.Fatal("la exportacion no genero un PDF valido")
	}
	if !bytes.Contains(pdf, []byte("Pagina 2 de")) || !bytes.Contains(pdf, []byte("Registro 60")) {
		t.Fatal("el PDF debe paginar y conservar el ultimo registro")
	}
	if !bytes.Contains(pdf, []byte("/MediaBox [0 0 612 792]")) {
		t.Fatal("el PDF carta debe conservar dimensiones letter")
	}
}

func TestReportesDatasetPDFRespetaPapelPOS80mm(t *testing.T) {
	ds := empresaReporteDataset{
		Key: "contable_cuentas_por_pagar", Title: "Cuentas por Pagar", Level: "contable",
		EmpresaID: 12, Desde: "2026-07-01", Hasta: "2026-07-31", GeneratedAt: "2026-07-31 06:27:29",
		Paper: "pos",
		Columns: []string{
			"codigo", "tercero", "documento_tipo", "documento_codigo", "fecha_emision",
			"fecha_vencimiento", "valor_original", "saldo", "estado_cartera",
		},
		Rows: []map[string]interface{}{{
			"codigo": "CXP-001", "tercero": "Proveedor con nombre suficientemente extenso para papel termico",
			"documento_tipo": "cuenta_por_pagar", "documento_codigo": "DOCUMENTO-POS-001",
			"fecha_emision": "2026-07-31", "fecha_vencimiento": "2026-08-31",
			"valor_original": 102.00, "saldo": 74.99, "estado_cartera": "pendiente",
		}},
	}

	for _, line := range reportesDatasetPDFLines(ds) {
		if got := len([]rune(line)); got > 42 {
			t.Fatalf("linea POS de %d caracteres excede el ancho de 80 mm: %q", got, line)
		}
	}
	lines := reportesDatasetPDFLines(ds)
	expectedHeight := len(lines)*11 + 48
	if expectedHeight < 144 {
		expectedHeight = 144
	}
	if expectedHeight > 792 {
		expectedHeight = 792
	}
	pdf := reportesDatasetPDFContent(ds)
	if !bytes.Contains(pdf, []byte(fmt.Sprintf("/MediaBox [0 0 227 %d]", expectedHeight))) {
		t.Fatal("paper=pos debe generar un PDF de 80 mm de ancho")
	}
	if bytes.Contains(pdf, []byte("/MediaBox [0 0 612 792]")) {
		t.Fatal("paper=pos no debe conservar dimensiones carta")
	}
}

func TestReportesDatasetPDFHumanizaEncabezadoTabular(t *testing.T) {
	ds := empresaReporteDataset{
		Title:       "Reporte IA",
		Level:       "contable",
		EmpresaID:   12,
		GeneratedAt: "2026-07-31 16:08:11",
		Columns:     []string{"tercero", "documento_codigo", "fecha_vencimiento", "estado_cartera"},
		Rows: []map[string]interface{}{{
			"tercero": "Proveedor", "documento_codigo": "CXP-001",
			"fecha_vencimiento": "2026-08-31", "estado_cartera": "parcial",
		}},
	}
	joined := strings.Join(reportesDatasetPDFLines(ds), "\n")
	if !strings.Contains(joined, "tercero | documento codigo | fecha vencimiento | estado cartera") {
		t.Fatalf("encabezado PDF no humanizado: %q", joined)
	}
	if strings.Contains(joined, "documento_codigo") || strings.Contains(joined, "fecha_vencimiento") || strings.Contains(joined, "estado_cartera") {
		t.Fatalf("el encabezado PDF expone nombres tecnicos: %q", joined)
	}
}

func TestReportesHelpersFiscalesYEdades(t *testing.T) {
	if !reportesCuentaEsFiscal("240805") {
		t.Fatalf("cuenta IVA 240805 debe clasificarse como fiscal")
	}
	if got := reportesConceptoFiscalCuenta("236805", "Rete ICA por pagar"); got != "retencion_ica" {
		t.Fatalf("concepto fiscal inesperado: %s", got)
	}
	if reportesInventarioTipoEsSalida("entrada_compra") {
		t.Fatalf("entrada_compra no debe clasificarse como salida de inventario")
	}
	if !reportesInventarioTipoEsSalida("salida_venta") {
		t.Fatalf("salida_venta debe clasificarse como salida de inventario")
	}
}

func TestReportesImprimiblesAmpliosUsanRegistrosEnLugarDeTablaRecortada(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "reportes_ejecutivos.html"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, want := range []string{
		".reports-print-page.reports-print-wide .reports-print-table{display:none}",
		".reports-print-page.reports-print-wide .reports-print-records{display:block}",
		"var useRecordLayout = columns.length > 8;",
		"sheet.classList.toggle('reports-print-wide', useRecordLayout);",
		".reports-print-meta{grid-template-columns:1fr}",
		".reports-print-page h2,.reports-print-page strong{color:#111827 !important}",
		".reports-paper-pos .reports-print-meta{grid-template-columns:1fr;gap:0}",
		".reports-paper-pos .reports-print-page strong{min-width:0;overflow-wrap:anywhere;word-break:break-word}",
		"function formatPreviewLabel(value)",
		"escapeHtml(formatPreviewLabel(col))",
		".reports-format-buttons{width:100%;min-width:0}",
		".reports-print-page.reports-print-wide .reports-print-row{grid-template-columns:minmax(112px,.45fr) minmax(0,1fr);gap:8px;align-items:start}",
		".reports-print-page strong{min-width:0;overflow-wrap:anywhere;word-break:break-word}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("contrato de impresion ancha ausente: %q", want)
		}
	}
}
