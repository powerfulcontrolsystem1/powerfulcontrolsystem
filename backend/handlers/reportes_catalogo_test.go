package handlers

import (
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
		".reports-print-page strong{min-width:0;overflow-wrap:anywhere;word-break:break-word}",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("contrato de impresion ancha ausente: %q", want)
		}
	}
}
