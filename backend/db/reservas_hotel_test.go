package db

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestReservaHotelFlowCRUDAndDisponibilidad(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := EnsureEmpresaReservasHotelSchema(dbConn); err != nil {
		t.Fatalf("ensure reservas schema: %v", err)
	}

	if _, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         1,
		Codigo:            "EST-1-1",
		Nombre:            "Habitacion 1",
		CanalVenta:        "estacion",
		Moneda:            "COP",
		ReferenciaExterna: "ESTACION_1",
		UsuarioCreador:    "test",
		Estado:            "activo",
	}); err != nil {
		t.Fatalf("create station 1: %v", err)
	}
	if _, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         1,
		Codigo:            "EST-1-2",
		Nombre:            "Habitacion 2",
		CanalVenta:        "estacion",
		Moneda:            "COP",
		ReferenciaExterna: "ESTACION_2",
		UsuarioCreador:    "test",
		Estado:            "activo",
	}); err != nil {
		t.Fatalf("create station 2: %v", err)
	}

	entrada := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	salida := time.Now().Add(6 * time.Hour).Format("2006-01-02 15:04:05")

	id, err := CreateReservaHotel(dbConn, ReservaHotel{
		EmpresaID:         1,
		EstacionID:        1,
		ClienteNombre:     "Cliente Uno",
		ClienteDocumento:  "1010",
		ClienteEmail:      "cliente1@test.com",
		CantidadHuespedes: 2,
		FechaEntrada:      entrada,
		FechaSalida:       salida,
		MontoTotal:        80000,
		Moneda:            "COP",
		UsuarioCreador:    "test",
	})
	if err != nil {
		t.Fatalf("create reserva: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	total, err := CountReservasHotelByEmpresa(dbConn, 1, ReservaHotelFilter{})
	if err != nil {
		t.Fatalf("count reservas: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}

	rows, err := ListReservasHotelByEmpresa(dbConn, 1, ReservaHotelFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list reservas: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].EstadoReserva != "pendiente_pago" {
		t.Fatalf("expected estado_reserva pendiente_pago, got %q", rows[0].EstadoReserva)
	}

	if _, err := CreateReservaHotel(dbConn, ReservaHotel{
		EmpresaID:         1,
		EstacionID:        1,
		ClienteNombre:     "Cliente Conflicto",
		FechaEntrada:      entrada,
		FechaSalida:       salida,
		CantidadHuespedes: 1,
		MontoTotal:        50000,
		Moneda:            "COP",
	}); !errors.Is(err, ErrReservaHotelConflicto) {
		t.Fatalf("expected ErrReservaHotelConflicto, got %v", err)
	}

	if err := UpdateReservaHotel(dbConn, ReservaHotel{
		ID:                id,
		EmpresaID:         1,
		EstacionID:        2,
		ClienteNombre:     "Cliente Uno Editado",
		CantidadHuespedes: 3,
		FechaEntrada:      entrada,
		FechaSalida:       salida,
		MontoTotal:        90000,
		Moneda:            "COP",
		Observaciones:     "actualizada",
	}); err != nil {
		t.Fatalf("update reserva: %v", err)
	}

	item, err := GetReservaHotelByID(dbConn, 1, id)
	if err != nil {
		t.Fatalf("get reserva by id: %v", err)
	}
	if item.EstacionID != 2 {
		t.Fatalf("expected estacion_id=2, got %d", item.EstacionID)
	}
	if item.ClienteNombre != "Cliente Uno Editado" {
		t.Fatalf("expected cliente editado, got %q", item.ClienteNombre)
	}

	avail, err := ListReservasHotelEstacionesDisponibles(dbConn, 1, entrada, salida)
	if err != nil {
		t.Fatalf("list disponibilidad: %v", err)
	}
	if len(avail) < 2 {
		t.Fatalf("expected at least 2 stations, got %d", len(avail))
	}
	var station2Found bool
	for _, st := range avail {
		if st.EstacionID == 2 {
			station2Found = true
			if st.Disponible {
				t.Fatalf("expected estacion 2 not disponible while pending reservation")
			}
		}
	}
	if !station2Found {
		t.Fatal("expected disponibilidad for estacion 2")
	}

	if err := ConfirmReservaHotelPago(dbConn, 1, id, "TRX-8899", "tesoreria", "ok"); err != nil {
		t.Fatalf("confirm pago: %v", err)
	}

	item, err = GetReservaHotelByID(dbConn, 1, id)
	if err != nil {
		t.Fatalf("get confirmed reserva: %v", err)
	}
	if item.EstadoReserva != "confirmada" || item.EstadoPago != "confirmado" {
		t.Fatalf("expected confirmada/confirmado, got %s/%s", item.EstadoReserva, item.EstadoPago)
	}

	if err := SetReservaHotelEstado(dbConn, 1, id, "inactivo"); err != nil {
		t.Fatalf("set estado inactivo: %v", err)
	}

	totalActivas, err := CountReservasHotelByEmpresa(dbConn, 1, ReservaHotelFilter{})
	if err != nil {
		t.Fatalf("count after inactivar: %v", err)
	}
	if totalActivas != 0 {
		t.Fatalf("expected total activas=0 after inactivar, got %d", totalActivas)
	}

	if err := DeleteReservaHotel(dbConn, 1, id); err != nil {
		t.Fatalf("delete reserva: %v", err)
	}

	if _, err := GetReservaHotelByID(dbConn, 1, id); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestReservaHotelMultiEstacionNoOverbookingYReconversion(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := EnsureEmpresaReservasHotelSchema(dbConn); err != nil {
		t.Fatalf("ensure reservas schema: %v", err)
	}

	stations := []CarritoCompra{
		{EmpresaID: 1, Codigo: "EST-1-1", Nombre: "Habitacion 1", CanalVenta: "estacion", Moneda: "COP", ReferenciaExterna: "ESTACION_1", UsuarioCreador: "test", Estado: "activo"},
		{EmpresaID: 1, Codigo: "EST-1-2", Nombre: "Habitacion 2", CanalVenta: "estacion", Moneda: "COP", ReferenciaExterna: "ESTACION_2", UsuarioCreador: "test", Estado: "activo"},
		{EmpresaID: 1, Codigo: "EST-1-3", Nombre: "Habitacion 3", CanalVenta: "estacion", Moneda: "COP", ReferenciaExterna: "ESTACION_3", UsuarioCreador: "test", Estado: "activo"},
	}
	for _, station := range stations {
		if _, err := CreateCarritoCompra(dbConn, station); err != nil {
			t.Fatalf("create station %s: %v", station.Codigo, err)
		}
	}

	entrada := time.Now().Add(4 * time.Hour).Format("2006-01-02 15:04:05")
	salida := time.Now().Add(8 * time.Hour).Format("2006-01-02 15:04:05")

	id1, err := CreateReservaHotel(dbConn, ReservaHotel{
		EmpresaID:         1,
		EstacionID:        1,
		ClienteNombre:     "Cliente A",
		CantidadHuespedes: 2,
		FechaEntrada:      entrada,
		FechaSalida:       salida,
		MontoTotal:        110000,
		Moneda:            "COP",
	})
	if err != nil {
		t.Fatalf("create reserva estacion 1: %v", err)
	}

	if _, err := CreateReservaHotel(dbConn, ReservaHotel{
		EmpresaID:         1,
		EstacionID:        2,
		ClienteNombre:     "Cliente B",
		CantidadHuespedes: 1,
		FechaEntrada:      entrada,
		FechaSalida:       salida,
		MontoTotal:        90000,
		Moneda:            "COP",
	}); err != nil {
		t.Fatalf("create reserva estacion 2: %v", err)
	}

	if _, err := CreateReservaHotel(dbConn, ReservaHotel{
		EmpresaID:         1,
		EstacionID:        1,
		ClienteNombre:     "Cliente Conflicto",
		CantidadHuespedes: 1,
		FechaEntrada:      entrada,
		FechaSalida:       salida,
		MontoTotal:        50000,
		Moneda:            "COP",
	}); !errors.Is(err, ErrReservaHotelConflicto) {
		t.Fatalf("expected conflict on estacion 1 overlapping window, got %v", err)
	}

	if err := ConfirmReservaHotelPago(dbConn, 1, id1, "TRX-RECON-1", "tesoreria", "confirmada para checkin"); err != nil {
		t.Fatalf("confirm reserva estacion 1: %v", err)
	}

	carritoID, err := ConvertReservaHotelToCarrito(dbConn, 1, id1, "recepcion")
	if err != nil {
		t.Fatalf("convert reserva a carrito: %v", err)
	}
	if carritoID <= 0 {
		t.Fatalf("expected carrito_id > 0, got %d", carritoID)
	}

	item, err := GetReservaHotelByID(dbConn, 1, id1)
	if err != nil {
		t.Fatalf("get reserva convertida: %v", err)
	}
	if item.EstadoReserva != "en_curso" {
		t.Fatalf("expected estado_reserva=en_curso after reconversion, got %q", item.EstadoReserva)
	}
	if item.CarritoID != carritoID {
		t.Fatalf("expected carrito_id=%d got=%d", carritoID, item.CarritoID)
	}

	if _, err := CreateReservaHotel(dbConn, ReservaHotel{
		EmpresaID:         1,
		EstacionID:        1,
		ClienteNombre:     "Cliente Conflicto 2",
		CantidadHuespedes: 1,
		FechaEntrada:      entrada,
		FechaSalida:       salida,
		MontoTotal:        60000,
		Moneda:            "COP",
	}); !errors.Is(err, ErrReservaHotelConflicto) {
		t.Fatalf("expected conflict on estacion 1 while en_curso, got %v", err)
	}

	avail, err := ListReservasHotelEstacionesDisponibles(dbConn, 1, entrada, salida)
	if err != nil {
		t.Fatalf("list disponibilidad: %v", err)
	}
	var s1, s2, s3 *ReservaHotelEstacion
	for i := range avail {
		switch avail[i].EstacionID {
		case 1:
			s1 = &avail[i]
		case 2:
			s2 = &avail[i]
		case 3:
			s3 = &avail[i]
		}
	}
	if s1 == nil || s2 == nil || s3 == nil {
		t.Fatalf("expected disponibilidad for estaciones 1, 2 y 3")
	}
	if s1.Disponible {
		t.Fatal("expected estacion 1 unavailable")
	}
	if s2.Disponible {
		t.Fatal("expected estacion 2 unavailable")
	}
	if !s3.Disponible {
		t.Fatal("expected estacion 3 available")
	}
}

func TestReservaHotelPoliticaNoShowYExpiracionAvanzada(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := EnsureEmpresaReservasHotelSchema(dbConn); err != nil {
		t.Fatalf("ensure reservas schema: %v", err)
	}

	if _, err := CreateCarritoCompra(dbConn, CarritoCompra{EmpresaID: 1, Codigo: "EST-1-1", Nombre: "Habitacion 1", CanalVenta: "estacion", Moneda: "COP", ReferenciaExterna: "ESTACION_1", UsuarioCreador: "test", Estado: "activo"}); err != nil {
		t.Fatalf("create station 1: %v", err)
	}
	if _, err := CreateCarritoCompra(dbConn, CarritoCompra{EmpresaID: 1, Codigo: "EST-1-2", Nombre: "Habitacion 2", CanalVenta: "estacion", Moneda: "COP", ReferenciaExterna: "ESTACION_2", UsuarioCreador: "test", Estado: "activo"}); err != nil {
		t.Fatalf("create station 2: %v", err)
	}

	noShowID, err := CreateReservaHotel(dbConn, ReservaHotel{
		EmpresaID:         1,
		EstacionID:        2,
		ClienteNombre:     "Cliente NoShow",
		CantidadHuespedes: 2,
		FechaEntrada:      time.Now().Add(-6 * time.Hour).Format("2006-01-02 15:04:05"),
		FechaSalida:       time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05"),
		FechaExpiracion:   time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05"),
		MontoTotal:        120000,
		Moneda:            "COP",
	})
	if err != nil {
		t.Fatalf("create reserva no_show: %v", err)
	}

	if err := ConfirmReservaHotelPago(dbConn, 1, noShowID, "TRX-NOSHOW", "tesoreria", "confirmada para prueba no_show"); err != nil {
		t.Fatalf("confirm reserva no_show: %v", err)
	}

	expiradaID, err := CreateReservaHotel(dbConn, ReservaHotel{
		EmpresaID:         1,
		EstacionID:        1,
		ClienteNombre:     "Cliente Expira",
		CantidadHuespedes: 1,
		FechaEntrada:      time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05"),
		FechaSalida:       time.Now().Add(3 * time.Hour).Format("2006-01-02 15:04:05"),
		FechaExpiracion:   time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05"),
		MontoTotal:        45000,
		Moneda:            "COP",
	})
	if err != nil {
		t.Fatalf("create reserva expirable: %v", err)
	}

	expiradas, noShow, err := ApplyReservasHotelOperationalPolicies(dbConn, 1)
	if err != nil {
		t.Fatalf("apply operational policies: %v", err)
	}
	if expiradas < 0 || noShow < 0 {
		t.Fatalf("invalid policy counters expiradas=%d no_show=%d", expiradas, noShow)
	}

	resExp, err := GetReservaHotelByID(dbConn, 1, expiradaID)
	if err != nil {
		t.Fatalf("get expirada: %v", err)
	}
	if resExp.EstadoReserva != "expirada" || resExp.EstadoPago != "expirado" {
		t.Fatalf("expected expirada/expirado, got %s/%s", resExp.EstadoReserva, resExp.EstadoPago)
	}

	resNoShow, err := GetReservaHotelByID(dbConn, 1, noShowID)
	if err != nil {
		t.Fatalf("get no_show: %v", err)
	}
	if resNoShow.EstadoReserva != "no_show" {
		t.Fatalf("expected estado_reserva=no_show, got %s", resNoShow.EstadoReserva)
	}
}
