package db

import "testing"

func TestEmpresaModulosFaltantesSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaModulosFaltantesSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaModulosFaltantesRequiredTablesAreAllowed(t *testing.T) {
	if len(empresaModulosFaltantesRequiredTables) == 0 {
		t.Fatal("required ERP table catalog must not be empty")
	}
	seen := make(map[string]struct{}, len(empresaModulosFaltantesRequiredTables))
	for _, table := range empresaModulosFaltantesRequiredTables {
		if _, ok := empresaGenericAllowedTables[table]; !ok {
			t.Fatalf("required table %q is not allowed by the ERP repository", table)
		}
		if _, duplicate := seen[table]; duplicate {
			t.Fatalf("required table %q appears more than once", table)
		}
		seen[table] = struct{}{}
	}
}

func TestEmpresaDocumentosTransaccionalesSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaDocumentosTransaccionalesSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaFacturacionElectronicaSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaFacturacionElectronicaSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaConfiguracionAvanzadaSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaConfiguracionAvanzadaSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaCarritosSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaCarritosSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaEstacionPrefsSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaEstacionPrefsSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaUbicacionGPSSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaUbicacionGPSSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaGrafologiaSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaGrafologiaSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaEnergiaSolarSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaEnergiaSolarSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaHojaVidaOperativaSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaHojaVidaOperativaSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaReservasHotelSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaReservasHotelSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaReportesProgramacionSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaReportesProgramacionSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaChatTareasSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaChatTareasSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaControlElectricoSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaControlElectricoSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaTarifasMotelSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaTarifasMotelSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaTarifasPorMinutosSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaTarifasPorMinutosSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaSensorPuertasSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaSensorPuertasSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaFinanzasSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaFinanzasSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestSuperContractSchemaReadyRejectsNilDatabase(t *testing.T) {
	t.Parallel()
	if err := SuperContractSchemaReady(nil); err == nil {
		t.Fatal("super contract readiness accepted nil database")
	}
}
