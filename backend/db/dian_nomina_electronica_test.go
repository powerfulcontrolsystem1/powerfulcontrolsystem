package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEmpresaDIANNominaElectronicaMigrationIsCatalogued(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260826-005-dian-nomina-electronica-v2" {
			continue
		}
		if migration.Body != empresaDIANNominaElectronicaFingerprint || migration.Apply == nil {
			t.Fatal("la migración de nómina electrónica debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("no se encontró la migración empresarial de nómina electrónica")
}

func TestCalcularTiempoLaboradoDIANUsaConvencion360Y30(t *testing.T) {
	start := time.Date(2020, 1, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 4, 28, 0, 0, 0, 0, time.UTC)
	days, err := CalcularTiempoLaboradoDIAN(start, end)
	if err != nil {
		t.Fatal(err)
	}
	if days != 1908 {
		t.Fatalf("tiempo laborado=%d, se esperaba el vector oficial 1908", days)
	}
	sameDay, err := CalcularTiempoLaboradoDIAN(start, start)
	if err != nil || sameDay != 1 {
		t.Fatalf("la regla NIE006 exige mínimo un día: days=%d err=%v", sameDay, err)
	}
	if _, err := CalcularTiempoLaboradoDIAN(end, start); err == nil {
		t.Fatal("una fecha de corte anterior al ingreso debe fallar")
	}
}

func TestValidateEmpresaNominaElectronicaConfigNoExigeResolucionNiRango(t *testing.T) {
	config := EmpresaDIANDocumentoConfiguracion{
		EmpresaID: 12, TipoDocumento: "nomina_electronica", Estado: "habilitacion",
		TipoAmbiente: "habilitacion", TestSetID: "fa5561a0-b872-482d-8dc3-5ef58902408b", Prefijo: "NE", ConsecutivoActual: 1,
	}
	if err := ValidateEmpresaNominaElectronicaConfigForEmission(config); err != nil {
		t.Fatalf("la numeración interna de nómina no usa resolución/rango de factura: %v", err)
	}
	config.Prefijo = "NE-"
	if err := ValidateEmpresaNominaElectronicaConfigForEmission(config); err == nil || !strings.Contains(err.Error(), "alfanumérico") {
		t.Fatalf("prefijo con guion debe fallar: %v", err)
	}
	config.Prefijo = "NE"
	config.ConsecutivoActual = 0
	if err := ValidateEmpresaNominaElectronicaConfigForEmission(config); err == nil || !strings.Contains(err.Error(), "consecutivo") {
		t.Fatalf("consecutivo cero debe fallar: %v", err)
	}
}

func TestValidateEmpresaNominaElectronicaConfigExigeTestSetEnHabilitacion(t *testing.T) {
	config := EmpresaDIANDocumentoConfiguracion{
		EmpresaID: 12, TipoDocumento: "nomina_electronica", Estado: "habilitacion",
		TipoAmbiente: "habilitacion", Prefijo: "NE", ConsecutivoActual: 1,
	}
	if err := ValidateEmpresaNominaElectronicaConfigForEmission(config); err == nil || !strings.Contains(err.Error(), "TestSetId") {
		t.Fatalf("habilitación sin TestSetId debe bloquearse: %v", err)
	}
	config.Estado = "activo"
	config.TipoAmbiente = "produccion"
	if err := ValidateEmpresaNominaElectronicaConfigForEmission(config); err != nil {
		t.Fatalf("producción no debe reutilizar el TestSetId como requisito: %v", err)
	}
}

func TestValidateEmpresaNominaDIANFuenteBloqueaHorasSinIntervalos(t *testing.T) {
	source := validEmpresaNominaDIANFuenteFixture()
	if blockers := ValidateEmpresaNominaDIANFuente(source); len(blockers) != 0 {
		t.Fatalf("fuente válida bloqueada: %v", blockers)
	}
	source.Devengados.TieneHorasSinTrazabilidad = true
	blockers := strings.Join(ValidateEmpresaNominaDIANFuente(source), " ")
	if !strings.Contains(blockers, "intervalos") {
		t.Fatalf("las horas agregadas sin inicio/fin deben bloquearse: %s", blockers)
	}
}

func TestValidateEmpresaNominaDIANFuenteExigeDVUnicoAntesDeReservar(t *testing.T) {
	for name, mutate := range map[string]func(*EmpresaNominaDIANFuente){
		"empleador no numerico": func(source *EmpresaNominaDIANFuente) { source.Empleador.DV = "X" },
		"empleador multiple":    func(source *EmpresaNominaDIANFuente) { source.Empleador.DV = "12" },
		"proveedor vacio":       func(source *EmpresaNominaDIANFuente) { source.ProveedorXML.DV = "" },
		"proveedor multiple":    func(source *EmpresaNominaDIANFuente) { source.ProveedorXML.DV = "98" },
	} {
		t.Run(name, func(t *testing.T) {
			source := validEmpresaNominaDIANFuenteFixture()
			mutate(source)
			blockers := strings.Join(ValidateEmpresaNominaDIANFuente(source), " ")
			if !strings.Contains(blockers, "identidad") {
				t.Fatalf("DV inválido no bloqueó la fuente antes de reservar: %s", blockers)
			}
		})
	}
}

func TestEmpresaNominaDIANPeriodoReporteEsMensualYBloqueaCruces(t *testing.T) {
	report, start, end, err := empresaNominaDIANPeriodoReporte("2026-08-16", "2026-08-31")
	if err != nil || report != "2026-08" || start != "2026-08-01" || end != "2026-08-31" {
		t.Fatalf("periodo mensual inesperado report=%q start=%q end=%q err=%v", report, start, end, err)
	}
	if _, _, _, err := empresaNominaDIANPeriodoReporte("2026-08-16", "2026-09-15"); err == nil {
		t.Fatal("una liquidación que cruza meses debe bloquearse")
	}
}

func TestEmpresaNominaDIANPeriodoCerradoBloqueaMesActualYFuturo(t *testing.T) {
	reference := time.Date(2026, time.August, 26, 12, 0, 0, 0, time.FixedZone("COT", -5*60*60))
	if !EmpresaNominaDIANPeriodoCerrado("2026-07", reference) {
		t.Fatal("previous reporting month must be closed")
	}
	if EmpresaNominaDIANPeriodoCerrado("2026-08", reference) {
		t.Fatal("current reporting month must remain open")
	}
	if EmpresaNominaDIANPeriodoCerrado("2026-09", reference) {
		t.Fatal("future reporting month must remain open")
	}
	if EmpresaNominaDIANPeriodoCerrado("invalid", reference) {
		t.Fatal("invalid reporting month must fail closed")
	}
}

func TestEmpresaNominaDIANFuenteOperacionalCoincideDetectaCambios(t *testing.T) {
	stored := validEmpresaNominaDIANFuenteFixture()
	stored.NominaID = 88
	stored.Prefijo = "NE"
	stored.Consecutivo = 7
	stored.NumeroLegal = "NE7"
	stored.FechaEmisionLegal = "2026-08-31T10:00:00-05:00"
	stored.TipoAmbiente = "produccion"
	current := validEmpresaNominaDIANFuenteFixture()
	if !EmpresaNominaDIANFuenteOperacionalCoincide(stored, current) {
		t.Fatal("la metadata de reserva no debe alterar la comparación operacional")
	}
	current.PagoIDs = []int64{46}
	current.PagoID = 46
	if EmpresaNominaDIANFuenteOperacionalCoincide(stored, current) {
		t.Fatal("un pago reemplazado debe invalidar la fuente reservada")
	}
}

func TestDecodeEmpresaNominaDIANSnapshotsEsEstricto(t *testing.T) {
	source := validEmpresaNominaDIANFuenteFixture()
	source.Prefijo = "NE"
	source.Consecutivo = 1
	source.NumeroLegal = "NE1"
	source.FechaEmisionLegal = "2026-08-31T10:00:00-05:00"
	source.TipoAmbiente = "produccion"
	sourceJSON, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	snapshotJSON, err := json.Marshal(EmpresaNominaDIANConfiguracionSnapshot{TipoDocumento: "nomina_electronica", TipoAmbiente: "produccion", Prefijo: "NE", ConsecutivoAsignado: 1})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := DecodeEmpresaNominaDIANSnapshots(string(sourceJSON), string(snapshotJSON)); err != nil {
		t.Fatalf("snapshots válidos fueron rechazados: %v", err)
	}
	withUnknown := strings.TrimSuffix(string(sourceJSON), "}") + `,"campo_ajeno":true}`
	if _, _, err := DecodeEmpresaNominaDIANSnapshots(withUnknown, string(snapshotJSON)); err == nil {
		t.Fatal("un campo fiscal desconocido debe fallar cerrado")
	}
	if _, _, err := DecodeEmpresaNominaDIANSnapshots(string(sourceJSON)+` {}`, string(snapshotJSON)); err == nil {
		t.Fatal("JSON fiscal concatenado debe fallar cerrado")
	}
}

func TestEmpresaDIANNominaElectronicaMigrationPostgresDoesNotInventFiscalData(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("PCS_TEST_POSTGRES_DSN"))
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	dbConn, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	tx, err := dbConn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	schema := fmt.Sprintf("dian_nomina_test_%d", time.Now().UnixNano())
	if _, err := tx.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`SET LOCAL search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE empresa_contabilidad_nomina_electronica (
			id BIGSERIAL PRIMARY KEY, empresa_id INTEGER NOT NULL, empleado_id INTEGER DEFAULT 0,
			tipo_documento TEXT DEFAULT 'CC', documento TEXT NOT NULL, nombre TEXT NOT NULL,
			periodo TEXT NOT NULL, fecha_pago TEXT NOT NULL, salario_base DOUBLE PRECISION DEFAULT 0,
			devengados DOUBLE PRECISION DEFAULT 0, deducciones DOUBLE PRECISION DEFAULT 0,
			total DOUBLE PRECISION DEFAULT 0, cune TEXT, estado_dian TEXT DEFAULT 'borrador',
			respuesta_dian TEXT, json_payload TEXT, fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP,
			fecha_actualizacion TEXT DEFAULT CURRENT_TIMESTAMP, usuario_creador TEXT,
			UNIQUE(empresa_id, documento, periodo)
		)`,
		`CREATE TABLE empresa_nomina_configuracion (id BIGSERIAL PRIMARY KEY, empresa_id INTEGER NOT NULL UNIQUE)`,
		`CREATE TABLE empresa_dian_configuracion (id BIGSERIAL PRIMARY KEY, empresa_id INTEGER NOT NULL)`,
		`INSERT INTO empresa_contabilidad_nomina_electronica
			(empresa_id, documento, nombre, periodo, fecha_pago, salario_base, devengados, deducciones, total)
			VALUES (12, '1001', 'Empleado legado', '2026-07', '2026-07-31', 1000, 1000, 80, 920)`,
	} {
		if _, err := tx.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
	if err := applyEmpresaDIANNominaElectronicaTx(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	var numero, cune, source string
	var sealed bool
	if err := tx.QueryRow(`SELECT numero_legal, COALESCE(cune, ''), fuente_fiscal_json::TEXT, fuente_fiscal_sellada
		FROM empresa_contabilidad_nomina_electronica WHERE empresa_id=12`).Scan(&numero, &cune, &source, &sealed); err != nil {
		t.Fatal(err)
	}
	if numero != "" || cune != "" || source != "{}" || sealed {
		t.Fatalf("la migración inventó estado fiscal: numero=%q cune=%q source=%q sealed=%v", numero, cune, source, sealed)
	}
}

func validEmpresaNominaDIANFuenteFixture() *EmpresaNominaDIANFuente {
	return &EmpresaNominaDIANFuente{
		Esquema: empresaNominaDIANFuenteEsquema, Version: empresaNominaDIANFuenteVersion,
		EmpresaID: 12, LiquidacionID: 90, LiquidacionIDs: []int64{90}, EmpleadoNominaID: 30,
		PagoID: 45, PagoIDs: []int64{45}, FechasPago: []string{"2026-08-30"}, PeriodoReporte: "2026-08",
		PeriodoDesde: "2026-08-01", PeriodoHasta: "2026-08-30", FechaIngreso: "2020-01-10",
		TiempoLaborado: 2390, PeriodoNomina: 5, SoftwareID: "software-id",
		Empleador:    EmpresaNominaDIANParte{RazonSocial: "Empresa SAS", NIT: "900373913", DV: "4", Pais: "CO", Departamento: "25", Municipio: "25175", Direccion: "Calle 1"},
		ProveedorXML: EmpresaNominaDIANParte{RazonSocial: "Proveedor SAS", NIT: "900373913", DV: "4"},
		Trabajador: EmpresaNominaDIANTrabajador{
			TipoTrabajador: "01", SubTipoTrabajador: "00", TipoDocumento: "13", NumeroDocumento: "1001001001",
			PrimerApellido: "Perez", SegundoApellido: "Gomez", PrimerNombre: "Ana",
			LugarTrabajoPais: "CO", LugarTrabajoDepartamento: "25", LugarTrabajoMunicipio: "25175",
			LugarTrabajoDireccion: "Calle 2", TipoContrato: "2", Sueldo: 3000000, CodigoTrabajador: "EMP-1",
		},
		Pago:             EmpresaNominaDIANPagoFuente{FechaPago: "2026-08-30", Forma: "1", Metodo: "30", NetoPagado: 2760000},
		Devengados:       EmpresaNominaDIANDevengados{DiasTrabajados: 30, SueldoTrabajado: 3000000, Total: 3000000},
		Deducciones:      EmpresaNominaDIANDeducciones{PorcentajeSalud: 4, Salud: 120000, PorcentajePension: 4, Pension: 120000, Total: 240000},
		ComprobanteTotal: 2760000,
	}
}
