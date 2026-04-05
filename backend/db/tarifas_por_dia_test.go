package db

import (
	"database/sql"
	"errors"
	"math"
	"testing"
)

func TestEmpresaTarifasPorDiaCRUDYCalculo(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		t.Fatalf("ensure tarifas por dia schema: %v", err)
	}

	id, err := CreateEmpresaTarifaPorDia(dbConn, EmpresaTarifaPorDia{
		EmpresaID:              1,
		EstacionID:             21,
		EstacionCodigo:         "EST-1-21",
		EstacionNombre:         "Habitacion 21",
		ServicioNombre:         "hotel",
		ValorDia:               85000,
		HoraCheckIn:            "15:00",
		HoraCheckOut:           "12:00",
		Moneda:                 "COP",
		Prioridad:              1,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa@empresa.com",
		Estado:                 "activo",
	})
	if err != nil {
		t.Fatalf("create tarifa por dia: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	rows, err := ListEmpresaTarifasPorDia(dbConn, 1, EmpresaTarifaPorDiaFilter{EstacionID: 21, Limit: 20})
	if err != nil {
		t.Fatalf("list tarifas por dia: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	activa, err := GetEmpresaTarifaPorDiaActiva(dbConn, 1, 21)
	if err != nil {
		t.Fatalf("get activa: %v", err)
	}
	if activa == nil {
		t.Fatal("expected active tariff")
	}

	aplicable, err := GetEmpresaTarifaPorDiaAplicable(dbConn, 1, 21)
	if err != nil {
		t.Fatalf("get aplicable: %v", err)
	}
	if aplicable == nil {
		t.Fatal("expected applicable tariff")
	}

	inicio, _ := parseTarifaPorDiaDateTime("2026-04-01 16:00:00")
	corte, _ := parseTarifaPorDiaDateTime("2026-04-03 13:00:00")
	dias, monto := CalcularMontoTarifaPorDia(*aplicable, inicio, corte)
	if dias != 3 {
		t.Fatalf("expected dias 3, got %d", dias)
	}
	if math.Abs(monto-255000) > 0.001 {
		t.Fatalf("expected monto 255000, got %.2f", monto)
	}

	if err := UpdateEmpresaTarifaPorDia(dbConn, EmpresaTarifaPorDia{
		ID:                     id,
		EmpresaID:              1,
		EstacionID:             21,
		EstacionCodigo:         "EST-1-21",
		EstacionNombre:         "Habitacion 21",
		ServicioNombre:         "hotel",
		ValorDia:               90000,
		HoraCheckIn:            "14:00",
		HoraCheckOut:           "11:00",
		Moneda:                 "COP",
		Prioridad:              1,
		AplicarAutomaticamente: false,
		UsuarioCreador:         "qa2@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("update tarifa por dia: %v", err)
	}

	aplicable, err = GetEmpresaTarifaPorDiaAplicable(dbConn, 1, 21)
	if err != nil {
		t.Fatalf("get aplicable after disable auto: %v", err)
	}
	if aplicable != nil {
		t.Fatal("expected nil applicable tariff when aplicar_automaticamente=false")
	}

	activa, err = GetEmpresaTarifaPorDiaActiva(dbConn, 1, 21)
	if err != nil {
		t.Fatalf("get activa after disable auto: %v", err)
	}
	if activa == nil {
		t.Fatal("expected active tariff even when automatic is disabled")
	}

	if err := SetEmpresaTarifaPorDiaEstado(dbConn, 1, id, "inactivo"); err != nil {
		t.Fatalf("set estado inactivo: %v", err)
	}
	activa, err = GetEmpresaTarifaPorDiaActiva(dbConn, 1, 21)
	if err != nil {
		t.Fatalf("get activa after inactivar: %v", err)
	}
	if activa != nil {
		t.Fatal("expected nil active tariff after inactivar")
	}

	if err := DeleteEmpresaTarifaPorDia(dbConn, 1, id); err != nil {
		t.Fatalf("delete tarifa por dia: %v", err)
	}
	if _, err := GetEmpresaTarifaPorDiaByID(dbConn, 1, id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestRefreshCarritoTotalConTarifaPorDia(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := EnsureEmpresaTarifasPorDiaSchema(dbConn); err != nil {
		t.Fatalf("ensure tarifas por dia schema: %v", err)
	}

	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         1,
		Codigo:            "EST-1-9",
		Nombre:            "Habitacion 9",
		CanalVenta:        "estacion",
		Moneda:            "COP",
		ReferenciaExterna: "ESTACION_9",
		UsuarioCreador:    "qa",
		Estado:            "activo",
	})
	if err != nil {
		t.Fatalf("create carrito estacion: %v", err)
	}

	if _, err := dbConn.Exec(`UPDATE carritos_compras SET
		estado = 'activo',
		estado_carrito = 'abierto',
		activado_en = ?,
		pagado_en = NULL
	WHERE empresa_id = ? AND id = ?`, "2026-04-01 16:00:00", 1, carritoID); err != nil {
		t.Fatalf("seed activado_en: %v", err)
	}

	_, err = CreateEmpresaTarifaPorDia(dbConn, EmpresaTarifaPorDia{
		EmpresaID:              1,
		EstacionID:             9,
		EstacionCodigo:         "EST-1-9",
		EstacionNombre:         "Habitacion 9",
		ServicioNombre:         "hotel",
		ValorDia:               100000,
		HoraCheckIn:            "15:00",
		HoraCheckOut:           "12:00",
		Moneda:                 "COP",
		Prioridad:              1,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa",
		Estado:                 "activo",
	})
	if err != nil {
		t.Fatalf("create tarifa diaria estacion: %v", err)
	}

	fechaCorte, _ := parseTarifaPorDiaDateTime("2026-04-03 13:00:00")
	calc, err := RefreshCarritoTotalConTarifaPorDia(dbConn, 1, carritoID, fechaCorte)
	if err != nil {
		t.Fatalf("refresh carrito tarifa por dia: %v", err)
	}
	if calc == nil {
		t.Fatal("expected calc not nil")
	}
	if !calc.Aplicada {
		t.Fatal("expected applied=true")
	}
	if calc.DiasCobrados != 3 {
		t.Fatalf("expected dias 3, got %d", calc.DiasCobrados)
	}
	if math.Abs(calc.MontoTarifa-300000) > 0.001 {
		t.Fatalf("expected monto 300000, got %.2f", calc.MontoTarifa)
	}
	if math.Abs(calc.TotalFinal-300000) > 0.001 {
		t.Fatalf("expected total_final 300000, got %.2f", calc.TotalFinal)
	}

	carrito, err := GetCarritoCompraByID(dbConn, 1, carritoID)
	if err != nil {
		t.Fatalf("get carrito: %v", err)
	}
	if math.Abs(carrito.Total-300000) > 0.001 {
		t.Fatalf("expected carrito total 300000, got %.2f", carrito.Total)
	}
	if math.Abs(carrito.Subtotal-300000) > 0.001 {
		t.Fatalf("expected carrito subtotal 300000, got %.2f", carrito.Subtotal)
	}

	if _, err := dbConn.Exec(`UPDATE empresa_tarifas_por_dia SET aplicar_automaticamente = 0 WHERE empresa_id = 1 AND estacion_id = 9`); err != nil {
		t.Fatalf("disable auto apply: %v", err)
	}
	calc, err = RefreshCarritoTotalConTarifaPorDia(dbConn, 1, carritoID, fechaCorte)
	if err != nil {
		t.Fatalf("refresh carrito after disable auto: %v", err)
	}
	if calc == nil {
		t.Fatal("expected calc not nil after disable auto")
	}
	if calc.Aplicada {
		t.Fatal("expected applied=false after disable auto")
	}
	if math.Abs(calc.TotalFinal-0) > 0.001 {
		t.Fatalf("expected total_final 0 after disable auto, got %.2f", calc.TotalFinal)
	}
}
