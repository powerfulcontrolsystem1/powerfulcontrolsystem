package handlers

import (
	"strings"
	"testing"
)

func TestValidateEmpresaReportePersonalizadoSpecBloqueaCamposYFuentesNoAutorizadas(t *testing.T) {
	invalidSource := empresaReportePersonalizadoSpec{SourceDataset: "ventas; drop table empresas", Limit: 10}
	if err := validateEmpresaReportePersonalizadoSpec(&invalidSource); err == nil {
		t.Fatal("una fuente no autorizada debe ser rechazada")
	}

	invalidField := empresaReportePersonalizadoSpec{
		SourceDataset: reporteDatasetContableCxC,
		Columns:       []string{"saldo", "empresa_id"},
		Limit:         10,
	}
	if err := validateEmpresaReportePersonalizadoSpec(&invalidField); err == nil {
		t.Fatal("un campo fuera del contrato debe ser rechazado")
	}
}

func TestReportesPersonalizadosAgrupaCalculaYOrdenaSinSQL(t *testing.T) {
	spec := empresaReportePersonalizadoSpec{
		SourceDataset: reporteDatasetContableCxC,
		GroupBy:       []string{"tercero"},
		Metrics: []empresaReportePersonalizadoMetric{
			{Alias: "documentos", Operation: "count"},
			{Alias: "saldo_total", Operation: "sum", Field: "saldo"},
		},
		Formulas: []empresaReportePersonalizadoFormula{
			{Alias: "saldo_promedio", Operation: "divide", Left: "saldo_total", Right: "documentos"},
		},
		OrderBy: []empresaReportePersonalizadoOrder{{Field: "saldo_total", Direction: "desc"}},
		Limit:   10,
	}
	if err := validateEmpresaReportePersonalizadoSpec(&spec); err != nil {
		t.Fatalf("spec valido rechazado: %v", err)
	}
	columns, rows := reportesPersonalizadoAggregate([]map[string]interface{}{
		{"tercero": "Cliente A", "saldo": 100.0},
		{"tercero": "Cliente A", "saldo": 50.0},
		{"tercero": "Cliente B", "saldo": 300.0},
	}, spec)
	reportesPersonalizadoSortRows(rows, spec.OrderBy)
	if len(columns) != 4 || len(rows) != 2 {
		t.Fatalf("resultado inesperado: columnas=%v filas=%v", columns, rows)
	}
	if rows[0]["tercero"] != "Cliente B" || rows[0]["saldo_total"] != 300.0 || rows[0]["saldo_promedio"] != 300.0 {
		t.Fatalf("primera fila agregada inesperada: %#v", rows[0])
	}
	if rows[1]["documentos"] != 2 || rows[1]["saldo_total"] != 150.0 || rows[1]["saldo_promedio"] != 75.0 {
		t.Fatalf("segunda fila agregada inesperada: %#v", rows[1])
	}
}

func TestEstadosFinancierosSinAsientosQuedanMarcadosPreliminares(t *testing.T) {
	dataset := empresaReporteDataset{Summary: map[string]interface{}{}}
	reportesMarkFinancialStatementStatus(&dataset, 0)
	if dataset.Summary["estado_contable"] != "preliminar_no_oficial" {
		t.Fatalf("estado contable inesperado: %#v", dataset.Summary)
	}
	if dataset.Summary["apto_para_presentacion_oficial"] != false {
		t.Fatalf("un reporte sin asientos no puede ser oficial: %#v", dataset.Summary)
	}
	if dataset.Description == "" {
		t.Fatal("el reporte preliminar debe mostrar una advertencia visible")
	}

	reportesMarkFinancialStatementStatus(&dataset, 3)
	if dataset.Summary["estado_contable"] != "basado_en_asientos_contables" || dataset.Summary["apto_para_revision_contable"] != true || dataset.Summary["apto_para_presentacion_oficial"] != false {
		t.Fatalf("estado con asientos inesperado: %#v", dataset.Summary)
	}
}

func TestReporteIAPersonalizadoAuditaHuellaDeLaEspecificacion(t *testing.T) {
	spec := empresaReportePersonalizadoSpec{SourceDataset: reporteDatasetContableCxC, Columns: []string{"tercero", "saldo"}, Limit: 10}
	first := reportesIACustomAuditObservations(spec)
	second := reportesIACustomAuditObservations(spec)
	if first != second || !strings.HasPrefix(first, "reportes_ia_custom:contable_cuentas_por_cobrar:spec_sha256:") {
		t.Fatalf("huella de auditoria inesperada: %q", first)
	}
}

func TestRubrosAuxiliaresDeEstadosFinancierosUsanNaturalezaPUC(t *testing.T) {
	tests := []struct {
		dataset string
		cuenta  string
		want    string
		include bool
	}{
		{reporteDatasetContableResultadoDetallado, "4135", "Ingresos", true},
		{reporteDatasetContableResultadoDetallado, "5135", "Costos y gastos", true},
		{reporteDatasetContableResultadoDetallado, "1105", "", false},
		{reporteDatasetContableSituacionDetallada, "1105", "Activos", true},
		{reporteDatasetContableSituacionDetallada, "2205", "Pasivos", true},
		{reporteDatasetContableSituacionDetallada, "3605", "Patrimonio", true},
		{reporteDatasetContablePatrimonioDetallado, "3605", "Patrimonio", true},
		{reporteDatasetContablePatrimonioDetallado, "4135", "", false},
	}
	for _, tc := range tests {
		got, included := reportesRubroEstadoFinanciero(tc.dataset, tc.cuenta)
		if got != tc.want || included != tc.include {
			t.Fatalf("dataset=%s cuenta=%s => (%q,%v), quiere (%q,%v)", tc.dataset, tc.cuenta, got, included, tc.want, tc.include)
		}
	}
	if got := reportesSaldoRubroEstadoFinanciero(reporteDatasetContableSituacionDetallada, "1105", 100, 20); got != 80 {
		t.Fatalf("saldo activo = %v, quiere 80", got)
	}
	if got := reportesSaldoRubroEstadoFinanciero(reporteDatasetContableResultadoDetallado, "4135", 10, 100); got != 90 {
		t.Fatalf("saldo ingreso = %v, quiere 90", got)
	}
}

func TestPlantillaReporteIAValidaEspecificacionYFuente(t *testing.T) {
	config := map[string]interface{}{
		"report_spec": map[string]interface{}{
			"source_dataset": reporteDatasetContableCxC,
			"columns":        []interface{}{"tercero", "saldo"},
			"limit":          25,
		},
	}
	if err := reportesValidateCustomSpecTemplate(reporteDatasetContableCxC, config); err != nil {
		t.Fatalf("plantilla IA valida rechazada: %v", err)
	}
	if _, ok := config["report_spec"].(empresaReportePersonalizadoSpec); !ok {
		t.Fatalf("spec no quedo normalizado: %#v", config["report_spec"])
	}
	if err := reportesValidateCustomSpecTemplate(reporteDatasetContableCxP, config); err == nil {
		t.Fatal("plantilla IA con fuente diferente debe rechazarse")
	}
}

func TestCatalogoContadorConservaReportesYFormatosObligatorios(t *testing.T) {
	required := map[string]bool{
		reporteDatasetContableEstadoResultados:    false,
		reporteDatasetContableBalanceGeneral:      false,
		reporteDatasetContableBalancePrueba:       false,
		reporteDatasetContableLibroAuxiliar:       false,
		reporteDatasetContableLibroMayor:          false,
		reporteDatasetContableCxC:                 false,
		reporteDatasetContableCxP:                 false,
		reporteDatasetCarteraEdadesCobrar:         false,
		reporteDatasetCarteraEdadesPagar:          false,
		reporteDatasetContableConciliacionBanco:   false,
		reporteDatasetFiscalImpuestosRetenciones:  false,
		reporteDatasetFiscalInformacionExogena:    false,
		reporteDatasetContableResultadoDetallado:  false,
		reporteDatasetContableSituacionDetallada:  false,
		reporteDatasetContablePatrimonioDetallado: false,
	}
	seen := make(map[string]struct{})
	for _, item := range reportesCatalogo {
		if _, duplicate := seen[item.Key]; duplicate {
			t.Fatalf("dataset duplicado en catalogo: %s", item.Key)
		}
		seen[item.Key] = struct{}{}
		if _, requiredForAccountant := required[item.Key]; !requiredForAccountant {
			continue
		}
		formats := make(map[string]bool, len(item.Formats))
		for _, format := range item.Formats {
			formats[format] = true
		}
		for _, format := range []string{"json", "csv", "xls", "pdf"} {
			if !formats[format] {
				t.Fatalf("dataset contable %s sin formato obligatorio %s", item.Key, format)
			}
		}
		required[item.Key] = true
	}
	for dataset, found := range required {
		if !found {
			t.Fatalf("falta dataset obligatorio para contador: %s", dataset)
		}
	}
}

func TestFuentesDeReportesIAPersonalizadosPertenecenAlCatalogo(t *testing.T) {
	catalog := make(map[string]struct{}, len(reportesCatalogo))
	for _, item := range reportesCatalogo {
		catalog[item.Key] = struct{}{}
	}
	for source, columns := range reportesPersonalizadosSources {
		if _, ok := catalog[source]; !ok {
			t.Fatalf("fuente IA no publicada en el catalogo: %s", source)
		}
		if len(columns) == 0 {
			t.Fatalf("fuente IA sin campos permitidos: %s", source)
		}
		seen := make(map[string]struct{}, len(columns))
		for _, column := range columns {
			if column == "" {
				t.Fatalf("fuente IA %s tiene campo vacio", source)
			}
			if _, duplicate := seen[column]; duplicate {
				t.Fatalf("fuente IA %s repite el campo %s", source, column)
			}
			seen[column] = struct{}{}
		}
	}
}

func TestReportSpecIARechazaInyeccionesYReferenciasFueraDelContrato(t *testing.T) {
	tests := []empresaReportePersonalizadoSpec{
		{SourceDataset: reporteDatasetContableCxC + " UNION SELECT *", Limit: 10},
		{SourceDataset: reporteDatasetContableCxC, Columns: []string{"tercero; DROP TABLE empresas"}, Limit: 10},
		{SourceDataset: reporteDatasetContableCxC, Metrics: []empresaReportePersonalizadoMetric{{Alias: "saldo total", Operation: "sum", Field: "saldo"}}, Limit: 10},
		{SourceDataset: reporteDatasetContableCxC, Metrics: []empresaReportePersonalizadoMetric{{Alias: "total", Operation: "sum", Field: "saldo"}}, Formulas: []empresaReportePersonalizadoFormula{{Alias: "fuga", Operation: "add", Left: "total", Right: "empresa_id"}}, Limit: 10},
		{SourceDataset: reporteDatasetContableCxC, Filters: []empresaReportePersonalizadoFilter{{Field: "saldo", Operator: "exec", Value: "1"}}, Limit: 10},
		{SourceDataset: reporteDatasetContableCxC, Limit: 1001},
	}
	for _, spec := range tests {
		if err := validateEmpresaReportePersonalizadoSpec(&spec); err == nil {
			t.Fatalf("ReportSpec fuera de contrato fue aceptado: %#v", spec)
		}
	}
}
