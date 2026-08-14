package handlers

import (
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestValidateDomoticaSolarTelemetryAcceptsVictronReading(t *testing.T) {
	reading := dbpkg.EmpresaEnergiaSolarLectura{
		PotenciaSolarW: 210, ProduccionDiaKwh: 1.85,
		BateriaVoltaje: 13.37, BateriaCorrienteA: -0.12,
		BateriaCargaW: 0, BateriaDescargaW: 1.6044,
		EstadoPaneles: "produciendo", EstadoBateria: "descargando", EstadoInversor: "bulk",
		Raw: map[string]interface{}{"protocolo": "VE.Direct", "checksum_valido": true, "pid": "0xA07D"},
	}
	if err := validateDomoticaSolarTelemetry("BlueSolar MPPT 75/15 rev3", reading); err != nil {
		t.Fatal(err)
	}
}

func TestValidateDomoticaSolarTelemetryRejectsImpossibleOrOversizedData(t *testing.T) {
	tests := []dbpkg.EmpresaEnergiaSolarLectura{
		{PotenciaSolarW: -1},
		{BateriaVoltaje: 3000},
		{Raw: map[string]interface{}{"payload": strings.Repeat("x", 33*1024)}},
	}
	for i, reading := range tests {
		if err := validateDomoticaSolarTelemetry("BlueSolar", reading); err == nil {
			t.Fatalf("caso %d debio rechazarse", i)
		}
	}
}
