package db

import (
	"errors"
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

func TestEmpresaVehiculoRegistroConfigValidacionDuplicidadYPermanencia(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaVehiculosRegistroSchema(dbConn); err != nil {
		t.Fatalf("ensure vehiculos registro schema: %v", err)
	}

	defaultCfg, err := GetEmpresaVehiculosRegistroConfiguracion(dbConn, 80)
	if err != nil {
		t.Fatalf("get default config: %v", err)
	}
	if defaultCfg.PaisCodigo != "CO" {
		t.Fatalf("expected default pais_codigo CO, got %q", defaultCfg.PaisCodigo)
	}

	if _, err := UpsertEmpresaVehiculosRegistroConfiguracion(dbConn, EmpresaVehiculosRegistroConfiguracion{
		EmpresaID:             80,
		PaisCodigo:            "MX",
		PatenteRegex:          "^[A-Z0-9]{6,7}$",
		PatenteDescripcion:    "MX test",
		EvitarDuplicadoActivo: true,
		Estado:                "activo",
		UsuarioCreador:        "qa@empresa.com",
	}); err != nil {
		t.Fatalf("upsert vehiculos config: %v", err)
	}

	id, err := CreateEmpresaVehiculoRegistro(dbConn, EmpresaVehiculoRegistro{
		EmpresaID:       80,
		Patente:         "ABC1234",
		TipoVehiculo:    "automovil",
		FechaIngreso:    "2026-04-02 08:00:00",
		EstadoRegistro:  "en_empresa",
		Estado:          "activo",
		UsuarioCreador:  "qa@empresa.com",
		ConductorNombre: "Driver 1",
	})
	if err != nil {
		t.Fatalf("create vehiculo valido: %v", err)
	}

	if _, err := CreateEmpresaVehiculoRegistro(dbConn, EmpresaVehiculoRegistro{
		EmpresaID:      80,
		Patente:        "AB@1234",
		TipoVehiculo:   "automovil",
		FechaIngreso:   "2026-04-02 09:00:00",
		EstadoRegistro: "en_empresa",
		Estado:         "activo",
	}); err == nil {
		t.Fatalf("expected error for patente invalida")
	}

	if _, err := CreateEmpresaVehiculoRegistro(dbConn, EmpresaVehiculoRegistro{
		EmpresaID:      80,
		Patente:        "ABC-1234",
		TipoVehiculo:   "automovil",
		FechaIngreso:   "2026-04-02 09:00:00",
		EstadoRegistro: "en_empresa",
		Estado:         "activo",
	}); !errors.Is(err, ErrEmpresaVehiculoDuplicadoActivo) {
		t.Fatalf("expected ErrEmpresaVehiculoDuplicadoActivo, got %v", err)
	}

	if err := MarkEmpresaVehiculoSalida(dbConn, 80, id, "2026-04-02 10:00:00", "qa@empresa.com", "salida controlada"); err != nil {
		t.Fatalf("mark salida: %v", err)
	}

	if _, err := CreateEmpresaVehiculoRegistro(dbConn, EmpresaVehiculoRegistro{
		EmpresaID:      80,
		Patente:        "ABC1234",
		TipoVehiculo:   "camioneta",
		FechaIngreso:   "2026-04-02 12:00:00",
		EstadoRegistro: "en_empresa",
		Estado:         "activo",
	}); err != nil {
		t.Fatalf("create vehiculo luego de salida: %v", err)
	}

	reporte, err := ListEmpresaVehiculosPermanenciaReporte(dbConn, 80, true, "2026-04-01", "2026-04-30", "", "", 100)
	if err != nil {
		t.Fatalf("list permanencia reporte: %v", err)
	}
	if len(reporte) != 2 {
		t.Fatalf("expected 2 rows in permanencia reporte, got %d", len(reporte))
	}

	foundRetirado := false
	for _, row := range reporte {
		if row.EstadoRegistro == "retirado" {
			foundRetirado = true
			if row.MinutosEstadia != 120 {
				t.Fatalf("expected retirado minutos_estadia=120, got %d", row.MinutosEstadia)
			}
		}
	}
	if !foundRetirado {
		t.Fatalf("expected at least one retirado row in permanencia reporte")
	}
}
