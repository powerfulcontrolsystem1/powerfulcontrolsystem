package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEmpresaDIANDocumentoSoporteMigrationIsCatalogued(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260826-002-dian-documento-soporte-v1" {
			continue
		}
		if migration.Body != empresaDIANDocumentoSoporteFingerprint || migration.Apply == nil {
			t.Fatal("la migracion de documento soporte debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("no se encontro la migracion empresarial de documento soporte")
}

func TestEmpresaDIANDocumentoSoporteMigrationPostgres(t *testing.T) {
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
	schema := fmt.Sprintf("dian_documento_soporte_test_%d", time.Now().UnixNano())
	if _, err := tx.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`SET LOCAL search_path TO ` + schema); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`CREATE TABLE empresa_contabilidad_documentos_soporte (
		id BIGSERIAL PRIMARY KEY, empresa_id BIGINT NOT NULL, subtotal DOUBLE PRECISION,
		iva DOUBLE PRECISION, retenciones DOUBLE PRECISION, total DOUBLE PRECISION,
		estado_dian TEXT NOT NULL DEFAULT 'borrador'
	)`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO empresa_contabilidad_documentos_soporte
		(empresa_id, subtotal, iva, retenciones, total) VALUES (12, 100, 19, 13.85, 119)`); err != nil {
		t.Fatal(err)
	}
	if err := applyEmpresaDIANDocumentoSoporteTx(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	var dataType string
	var precision, scale int
	var nullable string
	if err := tx.QueryRow(`SELECT data_type, numeric_precision, numeric_scale, is_nullable
		FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'empresa_contabilidad_documentos_soporte' AND column_name = 'total'`).Scan(&dataType, &precision, &scale, &nullable); err != nil {
		t.Fatal(err)
	}
	if dataType != "numeric" || precision != 18 || scale != 2 || nullable != "NO" {
		t.Fatalf("contrato monetario inesperado: type=%s precision=%d scale=%d nullable=%s", dataType, precision, scale, nullable)
	}
	if _, err := tx.Exec(`UPDATE empresa_contabilidad_documentos_soporte SET numero_legal='DS1' WHERE id=1`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`INSERT INTO empresa_contabilidad_documentos_soporte
		(empresa_id, subtotal, iva, retenciones, total, numero_legal)
		VALUES (12, 10, 0, 0, 10, 'DS1')`); err == nil {
		t.Fatal("se esperaba rechazo del numero legal duplicado dentro de la empresa")
	}
}

func TestNormalizeDocumentoSoporteRecalculaTotalesYRetencionesEnServidor(t *testing.T) {
	document := documentoSoporteDraftFixtureDB()
	document.Subtotal = 999999
	document.IVA = 999999
	document.Retenciones = 999999
	document.Total = 1
	document.TotalNetoContable = 1
	if err := normalizeEmpresaDocumentoSoporteFiscalDraft(&document); err != nil {
		t.Fatal(err)
	}
	if !closeDocumentoSoporteTest(document.Subtotal, 100) || !closeDocumentoSoporteTest(document.IVA, 19) ||
		!closeDocumentoSoporteTest(document.Retenciones, 13.85) || !closeDocumentoSoporteTest(document.Total, 119) ||
		!closeDocumentoSoporteTest(document.TotalNetoContable, 105.15) {
		t.Fatalf("totales recalculados inesperados: %+v", document)
	}
	var lines []EmpresaDocumentoSoporteLinea
	if err := json.Unmarshal([]byte(document.LineasJSON), &lines); err != nil || len(lines) != 1 {
		t.Fatalf("lineas normalizadas invalidas: %v %s", err, document.LineasJSON)
	}
	line := lines[0]
	if line.Numero != 1 || !closeDocumentoSoporteTest(line.ReteIVAValor, 2.85) || !closeDocumentoSoporteTest(line.ReteRentaValor, 11) ||
		!closeDocumentoSoporteTest(line.TotalLinea, 119) || !closeDocumentoSoporteTest(line.TotalNetoContableLinea, 105.15) {
		t.Fatalf("linea recalculada inesperada: %+v", line)
	}
}

func TestNormalizeDocumentoSoporteRechazaTarifasYMediosFueraDelAnexo(t *testing.T) {
	document := documentoSoporteDraftFixtureDB()
	document.LineasJSON = strings.Replace(document.LineasJSON, `"iva_porcentaje":19`, `"iva_porcentaje":12`, 1)
	if err := normalizeEmpresaDocumentoSoporteFiscalDraft(&document); err == nil || !strings.Contains(err.Error(), "tarifa") {
		t.Fatalf("IVA fuera de lista debe fallar: %v", err)
	}

	document = documentoSoporteDraftFixtureDB()
	document.MedioPagoCodigo = "8"
	if err := normalizeEmpresaDocumentoSoporteFiscalDraft(&document); err == nil || !strings.Contains(err.Error(), "medio_pago_codigo") {
		t.Fatalf("medio de pago fuera de lista debe fallar: %v", err)
	}

	document = documentoSoporteDraftFixtureDB()
	document.FormaPagoCodigo = "2"
	if err := normalizeEmpresaDocumentoSoporteFiscalDraft(&document); err == nil || !strings.Contains(err.Error(), "fecha_vencimiento") {
		t.Fatalf("credito sin vencimiento debe fallar: %v", err)
	}
}

func TestNormalizeDocumentoSoporteRechazaCamposDesconocidosYJSONConcatenado(t *testing.T) {
	document := documentoSoporteDraftFixtureDB()
	document.LineasJSON = strings.Replace(document.LineasJSON, `"codigo":"SERV-1"`, `"codigo":"SERV-1","estado_dian":"aceptado"`, 1)
	if err := normalizeEmpresaDocumentoSoporteFiscalDraft(&document); err == nil || !strings.Contains(err.Error(), "lineas_json") {
		t.Fatalf("campo desconocido dentro de lineas_json debe fallar: %v", err)
	}

	document = documentoSoporteDraftFixtureDB()
	document.LineasJSON += `[]`
	if err := normalizeEmpresaDocumentoSoporteFiscalDraft(&document); err == nil || !strings.Contains(err.Error(), "datos adicionales") {
		t.Fatalf("JSON concatenado dentro de lineas_json debe fallar: %v", err)
	}
}

func TestDocumentoSoporteCUDSDebeSerSHA384Hex(t *testing.T) {
	if !empresaDocumentoSoporteCUDSValido(strings.Repeat("a", 96)) {
		t.Fatal("un CUDS SHA-384 hexadecimal debe ser valido")
	}
	for _, invalid := range []string{strings.Repeat("a", 95), strings.Repeat("g", 96), ""} {
		if empresaDocumentoSoporteCUDSValido(invalid) {
			t.Fatalf("CUDS invalido aceptado: %q", invalid)
		}
	}
}

func TestDocumentoSoporteUnidadesUsanCatalogoLimpioDIANRevision4(t *testing.T) {
	codes := strings.Fields(documentoSoporteUnitCodesRevision4)
	if len(codes) != 1093 {
		t.Fatalf("catalogo DIAN Revision 4 tiene %d codigos, se esperaban 1093", len(codes))
	}
	for _, code := range []string{"04", "94", "1A", "2A", "4A", "5A", "ARE", "AS", "AY", "BE", "EA", "GK", "HE", "HUR", "KGM", "LTR", "MTR", "NIU", "NT", "P0", "PI", "ZZ"} {
		if !EmpresaDocumentoSoporteUnitCodeValid(code) {
			t.Fatalf("codigo oficial DIAN rechazado: %s", code)
		}
	}
	for _, code := range []string{"", "1ª", "COMO", "SÍ", "G K", "ÉL", "BAD", "AAAA"} {
		if EmpresaDocumentoSoporteUnitCodeValid(code) {
			t.Fatalf("codigo corrupto o inexistente aceptado: %q", code)
		}
	}
}

func TestValidateDocumentoSoporteConfigAdmitePrefijoVacioYExigeAutorizacionCompleta(t *testing.T) {
	emissionTime := time.Date(2026, 8, 26, 10, 0, 0, 0, time.FixedZone("America/Bogota", -5*60*60))
	valid := EmpresaDIANDocumentoConfiguracion{
		EmpresaID: 12, TipoDocumento: "documento_soporte", Estado: "habilitacion", TipoAmbiente: "habilitacion",
		Prefijo: "", ResolucionNumero: "18764000000001", ResolucionFechaDesde: "2026-01-01", ResolucionFechaHasta: "2026-12-31",
		RangoDesde: 1, RangoHasta: 999999999, ConsecutivoActual: 1,
	}
	if err := ValidateEmpresaDocumentoSoporteConfigForEmission(valid, emissionTime); err != nil {
		t.Fatalf("el anexo DIAN permite numeracion sin prefijo: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*EmpresaDIANDocumentoConfiguracion)
		want   string
	}{
		{"autorizacion corta", func(item *EmpresaDIANDocumentoConfiguracion) { item.ResolucionNumero = "123" }, "14 digitos"},
		{"autorizacion no numerica", func(item *EmpresaDIANDocumentoConfiguracion) { item.ResolucionNumero = "1876400000000A" }, "14 digitos"},
		{"fecha inicial ausente", func(item *EmpresaDIANDocumentoConfiguracion) { item.ResolucionFechaDesde = "" }, "vigencia completa"},
		{"fecha final ausente", func(item *EmpresaDIANDocumentoConfiguracion) { item.ResolucionFechaHasta = "" }, "vigencia completa"},
		{"prefijo largo", func(item *EmpresaDIANDocumentoConfiguracion) { item.Prefijo = "ABCDE" }, "maximo 4"},
		{"prefijo con simbolo", func(item *EmpresaDIANDocumentoConfiguracion) { item.Prefijo = "DS-" }, "alfanumerico"},
		{"rango excedido", func(item *EmpresaDIANDocumentoConfiguracion) { item.RangoHasta = 1000000000 }, "999999999"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := valid
			tc.mutate(&item)
			err := ValidateEmpresaDocumentoSoporteConfigForEmission(item, emissionTime)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, se esperaba texto %q", err, tc.want)
			}
		})
	}
}

func documentoSoporteDraftFixtureDB() EmpresaDocumentoSoporteElectronico {
	return EmpresaDocumentoSoporteElectronico{
		EmpresaID: 12, TipoDocumento: "NIT", Documento: "1020", NombreProveedor: "Proveedor",
		VendedorResidencia: "residente", VendedorPaisCodigo: "CO", VendedorTipoPersona: "juridica",
		FechaDocumento: "2026-08-26", Periodo: "2026-08", Concepto: "Servicio", Moneda: "COP",
		FormaPagoCodigo: "1", MedioPagoCodigo: "10",
		LineasJSON: `[{
			"numero":99,"codigo":"SERV-1","descripcion":"Servicio","unidad_medida":"94",
			"cantidad":1,"precio_unitario":100,"descuento_porcentaje":0,
			"iva_porcentaje":19,"reteiva_porcentaje":15,"reterenta_porcentaje":11,
			"valor_descuento":999,"base_gravable":999,"iva_valor":999,"reteiva_valor":999,
			"reterenta_valor":999,"subtotal_linea":999,"total_linea":999,"total_neto_contable_linea":999
		}]`,
	}
}

func closeDocumentoSoporteTest(a, b float64) bool {
	return math.Abs(a-b) <= 0.001
}
