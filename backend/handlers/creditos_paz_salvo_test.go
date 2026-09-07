package handlers

import (
	"bytes"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestBuildEmpresaCreditoPazYSalvoPDF(t *testing.T) {
	pdf := buildEmpresaCreditoPazYSalvoPDF(
		dbpkg.Empresa{ID: 12, EmpresaID: 12, Nombre: "Powerful Control System", Nit: "84456779"},
		dbpkg.EmpresaCredito{ID: 7, EmpresaID: 12, Codigo: "CRED-0007", ClienteNombre: "Cliente prueba", MontoAprobado: 100000, SaldoActual: 0},
		"10000001",
		100000,
		"PCS-PS-ABC123",
		time.Date(2026, 7, 9, 10, 0, 0, 0, time.Local),
	)
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		preview := pdf
		if len(preview) > 16 {
			preview = preview[:16]
		}
		t.Fatalf("se esperaba PDF valido, encabezado=%q", string(preview))
	}
	text := string(pdf)
	for _, expected := range []string{"CERTIFICADO DE PAZ Y SALVO", "CRED-0007", "PCS-PS-ABC123"} {
		if !strings.Contains(text, expected) {
			t.Fatalf("PDF no contiene %q", expected)
		}
	}
}

func TestBuildAgenteInternetNominaProposalsUsesOfficial2026Values(t *testing.T) {
	cfg := &dbpkg.EmpresaNominaConfiguracion{SalarioMinimoMensual: 1423500, HorasOrdinariasSemana: 44, DivisorHoraOrdinaria: 220, HoraNocturnaDesde: "21:00:00", RecargoDominicalDiurnoPorcentaje: 80}
	rows := buildAgenteInternetFiscalProposals("nomina", "CO", cfg)
	wanted := map[string]string{
		"salario_minimo_mensual":              "1750905",
		"horas_ordinarias_semana":             "42",
		"divisor_hora_ordinaria":              "210",
		"hora_nocturna_desde":                 "19:00:00",
		"recargo_dominical_diurno_porcentaje": "90",
	}
	for _, row := range rows {
		if value, ok := wanted[row.CampoConfig]; ok {
			if row.Sugerido != value || row.FuenteURL == "" || !row.RequiereOK {
				t.Fatalf("propuesta invalida para %s: %+v", row.CampoConfig, row)
			}
			delete(wanted, row.CampoConfig)
		}
	}
	if len(wanted) != 0 {
		t.Fatalf("faltan propuestas: %+v", wanted)
	}
}
