package db

import (
	"os"
	"strings"
	"testing"
)

func TestNormalizeEmpresaAITarifaMinutosPlanSemanal(t *testing.T) {
	plan := EmpresaAITarifaMinutosPlan{
		EstacionID: 4, NombreEstacion: "Habitacion 4",
		Tarifas: []EmpresaAITarifaMinutosRegla{
			{DiaSemanaDesde: 1, DiaSemanaHasta: 4, MinutosBase: 120, ValorBase: 50000, MinutosExtra: 60, ValorExtra: 25000, CobrarPorFraccion: true, Prioridad: 1},
			{DiaSemanaDesde: 5, DiaSemanaHasta: 7, MinutosBase: 180, ValorBase: 60000, MinutosExtra: 60, ValorExtra: 25000, CobrarPorFraccion: true, Prioridad: 1},
		},
	}
	if err := NormalizeEmpresaAITarifaMinutosPlan(&plan); err != nil {
		t.Fatalf("plan semanal valido rechazado: %v", err)
	}
}

func TestNormalizeEmpresaAITarifaMinutosPlanRechazaExtraIncompleto(t *testing.T) {
	plan := EmpresaAITarifaMinutosPlan{EstacionID: 2, Tarifas: []EmpresaAITarifaMinutosRegla{{DiaSemanaDesde: 1, DiaSemanaHasta: 7, MinutosBase: 60, ValorBase: 30000, MinutosExtra: 15, ValorExtra: 0}}}
	if err := NormalizeEmpresaAITarifaMinutosPlan(&plan); err == nil {
		t.Fatal("se acepto una fraccion extra sin valor")
	}
}

func TestNormalizeEmpresaAITarifaMinutosPlanRechazaDiasSuperpuestos(t *testing.T) {
	plan := EmpresaAITarifaMinutosPlan{EstacionID: 4, Tarifas: []EmpresaAITarifaMinutosRegla{
		{DiaSemanaDesde: 1, DiaSemanaHasta: 5, MinutosBase: 120, ValorBase: 50000, MinutosExtra: 60, ValorExtra: 25000},
		{DiaSemanaDesde: 5, DiaSemanaHasta: 7, MinutosBase: 180, ValorBase: 60000, MinutosExtra: 60, ValorExtra: 25000},
	}}
	if err := NormalizeEmpresaAITarifaMinutosPlan(&plan); err == nil {
		t.Fatal("se aceptaron rangos de dias superpuestos")
	}
}

func TestAITariffWritesRequireMigratedSchemaAndLockStationConfig(t *testing.T) {
	for _, path := range []string{"ai_tarifas_estacion.go", "ai_enterprise.go"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source := string(raw)
		if !strings.Contains(source, "SchemaReady(dbConn)") || !strings.Contains(source, "FOR UPDATE") {
			t.Fatalf("%s debe validar migraciones y bloquear estaciones_config", path)
		}
	}
	if raw, err := os.ReadFile("ai_enterprise.go"); err != nil {
		t.Fatal(err)
	} else if strings.Contains(string(raw), "EnsureEmpresaTarifasPorDiaSchema(dbConn)") {
		t.Fatal("la escritura IA de tarifas no debe ejecutar DDL en runtime")
	}
}
