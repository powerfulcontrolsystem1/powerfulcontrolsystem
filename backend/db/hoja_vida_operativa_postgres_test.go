package db

import (
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
}
