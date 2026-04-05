package db

import (
	"testing"
)

func TestEmpresaVehiculoRegistroCRUDFlow(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaVehiculosRegistroSchema(dbConn); err != nil {
		t.Fatalf("ensure vehiculos registro schema: %v", err)
	}

	id, err := CreateEmpresaVehiculoRegistro(dbConn, EmpresaVehiculoRegistro{
		EmpresaID:          44,
		Patente:            "abc 123",
		TipoVehiculo:       "automovil",
		Marca:              "Toyota",
		Modelo:             "Corolla",
		ConductorNombre:    "Juan Perez",
		ConductorDocumento: "10101010",
		MotivoIngreso:      "Visita comercial",
		ReferenciaExterna:  "VIS-001",
		UsuarioCreador:     "porteria@empresa.com",
	})
	if err != nil {
		t.Fatalf("create vehiculo registro: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	rows, err := ListEmpresaVehiculosRegistros(dbConn, 44, false, "", "", "", "ABC", "", 50)
	if err != nil {
		t.Fatalf("list vehiculos registros: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].Patente != "ABC123" {
		t.Fatalf("expected patente ABC123, got %q", rows[0].Patente)
	}
	if rows[0].EstadoRegistro != "en_empresa" {
		t.Fatalf("expected estado_registro en_empresa, got %q", rows[0].EstadoRegistro)
	}

	if err := MarkEmpresaVehiculoSalida(dbConn, 44, id, "", "porteria@empresa.com", "Salida registrada"); err != nil {
		t.Fatalf("mark salida: %v", err)
	}

	rowsSalida, err := ListEmpresaVehiculosRegistros(dbConn, 44, false, "", "", "retirado", "", "", 50)
	if err != nil {
		t.Fatalf("list vehiculos retirados: %v", err)
	}
	if len(rowsSalida) != 1 {
		t.Fatalf("expected 1 retirado row, got %d", len(rowsSalida))
	}
	if rowsSalida[0].FechaSalida == "" {
		t.Fatal("expected fecha_salida not empty")
	}

	if err := UpdateEmpresaVehiculoRegistro(dbConn, EmpresaVehiculoRegistro{
		ID:              id,
		EmpresaID:       44,
		Patente:         "XYZ-999",
		TipoVehiculo:    "camioneta",
		ConductorNombre: "Carlos Ruiz",
		MotivoIngreso:   "Entrega de mercancia",
		EstadoRegistro:  "retirado",
		FechaIngreso:    rowsSalida[0].FechaIngreso,
		FechaSalida:     rowsSalida[0].FechaSalida,
		Observaciones:   "Actualizado",
		UsuarioSalida:   "porteria@empresa.com",
	}); err != nil {
		t.Fatalf("update vehiculo registro: %v", err)
	}

	rowsUpdated, err := ListEmpresaVehiculosRegistros(dbConn, 44, false, "", "", "", "XYZ", "", 50)
	if err != nil {
		t.Fatalf("list updated vehiculos registros: %v", err)
	}
	if len(rowsUpdated) != 1 {
		t.Fatalf("expected 1 updated row, got %d", len(rowsUpdated))
	}
	if rowsUpdated[0].TipoVehiculo != "camioneta" {
		t.Fatalf("expected tipo_vehiculo camioneta, got %q", rowsUpdated[0].TipoVehiculo)
	}

	if err := SetEmpresaVehiculoRegistroEstado(dbConn, 44, id, "inactivo"); err != nil {
		t.Fatalf("set estado inactivo: %v", err)
	}
	rowsActive, err := ListEmpresaVehiculosRegistros(dbConn, 44, false, "", "", "", "", "", 50)
	if err != nil {
		t.Fatalf("list active vehiculos registros: %v", err)
	}
	if len(rowsActive) != 0 {
		t.Fatalf("expected 0 active rows after inactivar, got %d", len(rowsActive))
	}

	if err := DeleteEmpresaVehiculoRegistro(dbConn, 44, id); err != nil {
		t.Fatalf("delete vehiculo registro: %v", err)
	}
	rowsAll, err := ListEmpresaVehiculosRegistros(dbConn, 44, true, "", "", "", "", "", 50)
	if err != nil {
		t.Fatalf("list all after delete: %v", err)
	}
	if len(rowsAll) != 0 {
		t.Fatalf("expected 0 rows after delete, got %d", len(rowsAll))
	}
}
