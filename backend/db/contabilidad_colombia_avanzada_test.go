package db

import "testing"

func TestDefaultExogenaFormatosCubreFormatosBaseColombia(t *testing.T) {
	rows := defaultExogenaFormatos(7, "tester", 2026)
	want := map[string]bool{"1001": false, "1003": false, "1005": false, "1006": false, "1007": false, "1008": false, "1009": false}
	for _, row := range rows {
		if row.EmpresaID != 7 {
			t.Fatalf("empresa_id incorrecto: got %d", row.EmpresaID)
		}
		if row.AnioGravable != 2026 {
			t.Fatalf("anio gravable incorrecto para %s: got %d", row.Formato, row.AnioGravable)
		}
		if _, ok := want[row.Formato]; ok {
			want[row.Formato] = true
		}
	}
	for formato, found := range want {
		if !found {
			t.Fatalf("falta formato base %s", formato)
		}
	}
}

func TestValidateExogenaRegistro(t *testing.T) {
	ok := validateExogenaRegistro(EmpresaExogenaRegistro{Documento: "900123456", RazonSocial: "Proveedor SAS", Total: 100})
	if ok != "OK" {
		t.Fatalf("validacion esperada OK, got %q", ok)
	}
	bad := validateExogenaRegistro(EmpresaExogenaRegistro{})
	if bad == "OK" || bad == "" {
		t.Fatalf("validacion incompleta no fue reportada: %q", bad)
	}
}

func TestFormatEmpresaDocumentoElectronicoRef(t *testing.T) {
	got := FormatEmpresaDocumentoElectronicoRef("ne", 7, 42)
	if got != "NE-7-000042" {
		t.Fatalf("referencia incorrecta: got %q", got)
	}
}

func TestCalcularEmpresaActivoDepreciacionPeriodoLineaRecta(t *testing.T) {
	activo := EmpresaActivoFijo{
		ID:                      11,
		EmpresaID:               7,
		Codigo:                  "AF-001",
		Nombre:                  "Equipo QA",
		FechaCompra:             "2026-01-15",
		FechaInicioDepreciacion: "2026-01-15",
		Costo:                   1200000,
		ValorResidual:           120000,
		VidaUtilMeses:           12,
		MetodoDepreciacion:      "linea_recta",
	}
	row := calcularEmpresaActivoDepreciacionPeriodo(activo, "2026-03", "2026-03-28", "qa")
	if row.DepreciacionPeriodo != 90000 {
		t.Fatalf("depreciacion periodo = %v", row.DepreciacionPeriodo)
	}
	if row.DepreciacionAcumulada != 270000 {
		t.Fatalf("depreciacion acumulada = %v", row.DepreciacionAcumulada)
	}
	if row.ValorLibros != 930000 {
		t.Fatalf("valor libros = %v", row.ValorLibros)
	}
}

func TestNormalizeActivoEventoTipo(t *testing.T) {
	if got := normalizeActivoEventoTipo(" TRASLADO "); got != "traslado" {
		t.Fatalf("tipo evento = %q", got)
	}
	if got := normalizeActivoEventoTipo("desconocido"); got != "mantenimiento" {
		t.Fatalf("tipo evento por defecto = %q", got)
	}
}
