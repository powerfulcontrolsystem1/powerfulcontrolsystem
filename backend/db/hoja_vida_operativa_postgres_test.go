package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmpresaHojaVidaReporteUsesIntegerCompatibleRecurrentePredicate(t *testing.T) {
	t.Parallel()
	if strings.Contains(empresaHojaVidaReporteQuery, "recurrente = 1") {
		t.Fatal("PostgreSQL report must not compare a legacy integer flag through a boolean-only predicate")
	}
	if !strings.Contains(empresaHojaVidaReporteQuery, "COALESCE(recurrente, 0) <> 0") {
		t.Fatal("report must preserve the legacy integer recurrente contract")
	}
	if strings.Contains(empresaHojaVidaReporteQuery, "fecha_programada < CURRENT_TIMESTAMP") {
		t.Fatal("PostgreSQL report must not compare a legacy text date directly with CURRENT_TIMESTAMP")
	}
	if !strings.Contains(empresaHojaVidaReporteQuery, "pcs_ts(fecha_programada) < CURRENT_TIMESTAMP") {
		t.Fatal("report must normalize the legacy text date before comparing it")
	}
}

func TestVehiculosPermanenciaNormalizaTimestampParaPostgreSQL(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("vehiculos_registro.go"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, want := range []string{
		"COALESCE(CAST(fecha_ingreso AS TEXT), '')",
		"COALESCE(CAST(fecha_salida AS TEXT), '')",
		"NULLIF(CAST(fecha_salida AS TEXT), '')",
		"CAST(CURRENT_TIMESTAMP AS TEXT)",
		"COALESCE(CAST(fecha_ingreso AS TEXT), CAST(fecha_creacion AS TEXT), '')",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("contrato PostgreSQL de permanencia ausente: %q", want)
		}
	}
}
