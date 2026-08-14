package db

import (
	"database/sql"
	"testing"
	"time"
)

func TestNormalizeListLimitOffset(t *testing.T) {
	tests := []struct {
		name               string
		limit, offset      int
		defaultLimit, max  int
		wantLimit, wantOff int
	}{
		{name: "defaults and clamps offset", limit: 0, offset: -4, defaultLimit: 100, max: 500, wantLimit: 100, wantOff: 0},
		{name: "caps limit", limit: 900, offset: 12, defaultLimit: 100, max: 500, wantLimit: 500, wantOff: 12},
		{name: "preserves valid values", limit: 25, offset: 7, defaultLimit: 100, max: 500, wantLimit: 25, wantOff: 7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotLimit, gotOffset := normalizeListLimitOffset(test.limit, test.offset, test.defaultLimit, test.max)
			if gotLimit != test.wantLimit || gotOffset != test.wantOff {
				t.Fatalf("normalizeListLimitOffset()=(%d,%d), want (%d,%d)", gotLimit, gotOffset, test.wantLimit, test.wantOff)
			}
		})
	}
}

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

func TestEmpresaPreconfigSchemaReadyRejectsNilDatabase(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*sql.DB) error
	}{
		{"productos", EmpresaProductosSchemaReady},
		{"usuarios", EmpresaUsuariosAuthSchemaReady},
		{"configuracion_operativa", EmpresaConfiguracionOperativaSchemaReady},
		{"comisiones", EmpresaComisionesServicioSchemaReady},
		{"tarifas_por_dia", EmpresaTarifasPorDiaSchemaReady},
		{"nextcloud", EmpresaNextcloudSchemaReady},
		{"eventos_contables", EmpresaEventosContablesSchemaReady},
		{"permisos_finos", EmpresaPermisosFinosSchemaReady},
		{"rappi", EmpresaRappiSchemaReady},
		{"roles_permisos", RolesPermisosSchemaReady},
		{"super_alertas", SuperAlertasSchemaReady},
		{"super_mantenimiento", SuperMantenimientoAgentesSchemaReady},
		{"email_corporativo", EmpresaEmailCorporativoSchemaReady},
		{"plantillas_licencias", PlantillasLicenciasSchemaReady},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.fn(nil); err == nil {
				t.Fatal("expected nil database to be rejected")
			}
		})
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

func TestSharedRepositoryValueHelpers(t *testing.T) {
	t.Parallel()
	if got := firstNonBlankValue("  ", " value ", "fallback"); got != "value" {
		t.Fatalf("first non-blank value = %q", got)
	}
	if got := firstNonBlankValue("", " \t "); got != "" {
		t.Fatalf("empty values returned %q", got)
	}
	if got := escapedContainsPattern(" 50%!_off "); got != "%50!%!!!_off%" {
		t.Fatalf("escaped LIKE pattern = %q", got)
	}
}

func TestRepositoryCoreCodeSharedFormatting(t *testing.T) {
	t.Parallel()
	if got := repositoryCoreCode("dom", "Room 1", "Relay_2"); got != "DOM-ROOM-1-RELAY-2" {
		t.Fatalf("core code = %q", got)
	}
	if got := repositoryCoreCode("dom", "   "); len(got) < len("DOM-") {
		t.Fatalf("fallback core code = %q", got)
	}
}

func TestNormalizeRepositoryPeriod(t *testing.T) {
	t.Parallel()
	if got := normalizeRepositoryPeriod(" 2026-08-14 ", ""); got != "2026-08" {
		t.Fatalf("normalized period = %q", got)
	}
	if got := normalizeRepositoryPeriod(" 2026 ", "fallback"); got != "fallback" {
		t.Fatalf("fallback period = %q", got)
	}
	if got := normalizeRepositoryPeriod(" 2026 ", ""); got != "2026" {
		t.Fatalf("short period = %q", got)
	}
}

func TestRepositorySharedScalarHelpers(t *testing.T) {
	t.Parallel()
	if got := boundedPositiveInt(0, 1, 999); got != 1 {
		t.Fatalf("bounded fallback = %d", got)
	}
	if got := boundedPositiveInt(1000, 1, 999); got != 999 {
		t.Fatalf("bounded maximum = %d", got)
	}
	if got := normalizeRepositoryCurrency(" usd-dollar ", "COP", 8); got != "USD-DOLL" {
		t.Fatalf("currency = %q", got)
	}
	if got := repositoryISOWeekday(time.Date(2026, time.August, 16, 0, 0, 0, 0, time.UTC)); got != 7 {
		t.Fatalf("ISO weekday = %d", got)
	}
	if hashRepositoryKey(" key ") != hashRepositoryKey("key") || hashRepositoryKey(" ") != "" {
		t.Fatal("repository hash normalization mismatch")
	}
}
