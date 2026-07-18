package db

import (
	"os"
	"strings"
	"testing"
)

func TestConfiguracionOperativaWritesUsePostgresReturningID(t *testing.T) {
	raw, err := os.ReadFile("configuracion_operativa.go")
	if err != nil {
		t.Fatalf("read configuracion_operativa.go: %v", err)
	}
	src := string(raw)

	for _, fn := range []string{
		"func UpsertEmpresaConfiguracionOperativa(",
		"func UpsertEmpresaConfiguracionOperativaRol(",
		"func UpsertEmpresaConfiguracionOperativaPolitica(",
		"func CreateEmpresaConfiguracionOperativaHistorialSnapshot(",
	} {
		body := extractConfiguracionOperativaFunctionForTest(t, src, fn)
		if strings.Contains(body, "LastInsertId(") {
			t.Fatalf("%s no debe depender de LastInsertId en PostgreSQL: %s", fn, body)
		}
		if !strings.Contains(body, "QueryRowCompat") {
			t.Fatalf("%s debe usar QueryRowCompat para rebind PostgreSQL: %s", fn, body)
		}
		if !strings.Contains(body, "RETURNING id") {
			t.Fatalf("%s debe retornar id con RETURNING id: %s", fn, body)
		}
		if !strings.Contains(body, "sqlNowExpr()") {
			t.Fatalf("%s debe usar sqlNowExpr para fechas runtime: %s", fn, body)
		}
	}
}

func TestConfiguracionOperativaRolPermiteIngresosEgresosManuales(t *testing.T) {
	cfg := defaultEmpresaConfiguracionOperativa(12)
	cfg.Roles = []EmpresaConfiguracionOperativaRol{
		{
			EmpresaID:                12,
			Rol:                      "cajero",
			MetodoPagoEfectivo:       true,
			HabilitarPropinas:        true,
			HabilitarComisiones:      true,
			PermitirIngresosManuales: true,
			PermitirEgresosManuales:  false,
			Estado:                   "activo",
		},
	}

	permisos := ResolveEmpresaConfiguracionOperativaParaRol(&cfg, "cajero")
	if !permisos.PermiteMovimientoFinancieroManual("ingreso") {
		t.Fatal("cajero debe poder registrar ingresos manuales cuando el override del rol lo habilita")
	}
	if permisos.PermiteMovimientoFinancieroManual("egreso") {
		t.Fatal("cajero no debe poder registrar egresos manuales si el override del rol no lo habilita")
	}
	if permisos.PermiteMovimientoFinancieroManual("transferencia") {
		t.Fatal("tipo de movimiento desconocido no debe quedar habilitado")
	}
}

func TestGetEmpresaConfiguracionOperativaKeepsSelectAndScanCompatible(t *testing.T) {
	raw, err := os.ReadFile("configuracion_operativa.go")
	if err != nil {
		t.Fatalf("read configuracion_operativa.go: %v", err)
	}
	body := extractConfiguracionOperativaFunctionForTest(t, string(raw), "func GetEmpresaConfiguracionOperativa(")
	for _, unexpected := range []string{
		"COALESCE(permitir_ingresos_manuales, 0)",
		"COALESCE(permitir_egresos_manuales, 0)",
	} {
		if strings.Contains(body, unexpected) {
			t.Fatalf("GetEmpresaConfiguracionOperativa no expone %q sin un destino en EmpresaConfiguracionOperativa", unexpected)
		}
	}
	// El Scan recibe 15 destinos: id y empresa_id se leen directamente y las
	// trece columnas restantes se normalizan con COALESCE.
	if got := strings.Count(body, "COALESCE("); got != 13 {
		t.Fatalf("GetEmpresaConfiguracionOperativa selecciona %d columnas COALESCE, want 13 mas id y empresa_id compatibles con Scan", got)
	}
}

func extractConfiguracionOperativaFunctionForTest(t *testing.T, src, startMarker string) string {
	t.Helper()

	start := strings.Index(src, startMarker)
	if start < 0 {
		t.Fatalf("no se encontro %s", startMarker)
	}
	rest := src[start+len(startMarker):]
	next := strings.Index(rest, "\nfunc ")
	if next < 0 {
		return src[start:]
	}
	return src[start : start+len(startMarker)+next]
}
