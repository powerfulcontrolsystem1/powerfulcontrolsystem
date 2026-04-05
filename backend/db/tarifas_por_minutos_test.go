package db

import (
	"database/sql"
	"errors"
	"math"
	"testing"
)

func TestEmpresaTarifasPorMinutosCRUDYResolucionPorDia(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaTarifasPorMinutosSchema(dbConn); err != nil {
		t.Fatalf("ensure tarifas por minutos schema: %v", err)
	}

	idSemana, err := CreateEmpresaTarifaPorMinutos(dbConn, EmpresaTarifaPorMinutos{
		EmpresaID:      1,
		EstacionID:     10,
		EstacionCodigo: "EST-1-10",
		EstacionNombre: "Habitacion 10",
		DiaSemanaDesde: 1,
		DiaSemanaHasta: 4,
		MinutosBase:    120,
		ValorBase:      30000,
		MinutosExtra:   60,
		ValorExtra:     15000,
		Moneda:         "COP",
		Prioridad:      1,
		UsuarioCreador: "qa@empresa.com",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("create tarifa lunes-jueves: %v", err)
	}
	if idSemana <= 0 {
		t.Fatalf("expected id > 0, got %d", idSemana)
	}

	idFinSemana, err := CreateEmpresaTarifaPorMinutos(dbConn, EmpresaTarifaPorMinutos{
		EmpresaID:      1,
		EstacionID:     10,
		EstacionCodigo: "EST-1-10",
		EstacionNombre: "Habitacion 10",
		DiaSemanaDesde: 5,
		DiaSemanaHasta: 7,
		MinutosBase:    120,
		ValorBase:      40000,
		MinutosExtra:   60,
		ValorExtra:     20000,
		Moneda:         "COP",
		Prioridad:      1,
		UsuarioCreador: "qa@empresa.com",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("create tarifa viernes-domingo: %v", err)
	}

	rowsJueves, err := ListEmpresaTarifasPorMinutos(dbConn, 1, EmpresaTarifaPorMinutosFilter{
		EstacionID: 10,
		DiaSemana:  4,
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("list tarifas dia 4: %v", err)
	}
	if len(rowsJueves) != 1 {
		t.Fatalf("expected 1 tarifa for dia 4, got %d", len(rowsJueves))
	}
	if rowsJueves[0].ID != idSemana {
		t.Fatalf("expected id %d for dia 4, got %d", idSemana, rowsJueves[0].ID)
	}

	rowsSabado, err := ListEmpresaTarifasPorMinutos(dbConn, 1, EmpresaTarifaPorMinutosFilter{
		EstacionID: 10,
		DiaSemana:  6,
		Limit:      20,
	})
	if err != nil {
		t.Fatalf("list tarifas dia 6: %v", err)
	}
	if len(rowsSabado) != 1 {
		t.Fatalf("expected 1 tarifa for dia 6, got %d", len(rowsSabado))
	}
	if rowsSabado[0].ID != idFinSemana {
		t.Fatalf("expected id %d for dia 6, got %d", idFinSemana, rowsSabado[0].ID)
	}

	aplicableSabado, err := GetEmpresaTarifaPorMinutosAplicable(dbConn, 1, 10, 6)
	if err != nil {
		t.Fatalf("get aplicable dia 6: %v", err)
	}
	if aplicableSabado == nil {
		t.Fatal("expected tarifa aplicable for dia 6")
	}
	if math.Abs(aplicableSabado.ValorBase-40000) > 0.001 {
		t.Fatalf("expected valor_base 40000, got %.2f", aplicableSabado.ValorBase)
	}

	if err := UpdateEmpresaTarifaPorMinutos(dbConn, EmpresaTarifaPorMinutos{
		ID:             idFinSemana,
		EmpresaID:      1,
		EstacionID:     10,
		EstacionCodigo: "EST-1-10",
		EstacionNombre: "Habitacion 10",
		DiaSemanaDesde: 5,
		DiaSemanaHasta: 7,
		MinutosBase:    120,
		ValorBase:      45000,
		MinutosExtra:   60,
		ValorExtra:     22000,
		Moneda:         "COP",
		Prioridad:      1,
		UsuarioCreador: "qa2@empresa.com",
		Estado:         "activo",
	}); err != nil {
		t.Fatalf("update tarifa fin de semana: %v", err)
	}

	item, err := GetEmpresaTarifaPorMinutosByID(dbConn, 1, idFinSemana)
	if err != nil {
		t.Fatalf("get tarifa by id: %v", err)
	}
	if item == nil {
		t.Fatal("expected item not nil")
	}
	if math.Abs(item.ValorExtra-22000) > 0.001 {
		t.Fatalf("expected valor_extra 22000, got %.2f", item.ValorExtra)
	}

	total, bloques := CalcularMontoTarifaPorMinutos(*item, 190)
	if bloques != 2 {
		t.Fatalf("expected 2 bloques extra, got %d", bloques)
	}
	if math.Abs(total-89000) > 0.001 {
		t.Fatalf("expected total 89000, got %.2f", total)
	}

	if err := SetEmpresaTarifaPorMinutosEstado(dbConn, 1, idFinSemana, "inactivo"); err != nil {
		t.Fatalf("set estado inactivo: %v", err)
	}
	aplicableSabado, err = GetEmpresaTarifaPorMinutosAplicable(dbConn, 1, 10, 6)
	if err != nil {
		t.Fatalf("get aplicable after inactivar: %v", err)
	}
	if aplicableSabado != nil {
		t.Fatal("expected nil tarifa aplicable after inactivar fin de semana")
	}

	if err := DeleteEmpresaTarifaPorMinutos(dbConn, 1, idSemana); err != nil {
		t.Fatalf("delete tarifa semana: %v", err)
	}
	if _, err := GetEmpresaTarifaPorMinutosByID(dbConn, 1, idSemana); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after delete, got %v", err)
	}
}
