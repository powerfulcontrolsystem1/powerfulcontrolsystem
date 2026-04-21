package db

import (
	"database/sql"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
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

func TestEmpresaTarifasPorMinutosConfiguracionYCalculoAvanzado(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaTarifasPorMinutosSchema(dbConn); err != nil {
		t.Fatalf("ensure tarifas por minutos schema: %v", err)
	}

	defaultCfg, err := GetEmpresaTarifaPorMinutosConfiguracion(dbConn, 2)
	if err != nil {
		t.Fatalf("get default cfg: %v", err)
	}
	if defaultCfg.RedondeoModo != "ninguno" {
		t.Fatalf("expected redondeo_modo ninguno, got %q", defaultCfg.RedondeoModo)
	}

	cfg, err := UpsertEmpresaTarifaPorMinutosConfiguracion(dbConn, EmpresaTarifaPorMinutosConfiguracion{
		EmpresaID:         2,
		RedondeoModo:      "arriba",
		RedondeoUnidad:    100,
		MontoMinimoDiario: 50000,
		MontoMaximoDiario: 70000,
		UsuarioCreador:    "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("upsert cfg: %v", err)
	}

	tarifa := EmpresaTarifaPorMinutos{
		ID:           99,
		EmpresaID:    2,
		EstacionID:   30,
		MinutosBase:  120,
		ValorBase:    30000,
		MinutosExtra: 60,
		ValorExtra:   15000,
		Moneda:       "COP",
	}

	detalleFraccion := CalcularDetalleTarifaPorMinutos(tarifa, 120.25, *cfg)
	if detalleFraccion.BloquesExtra != 1 {
		t.Fatalf("expected bloques_extra=1 for fraction jump, got %d", detalleFraccion.BloquesExtra)
	}
	if math.Abs(detalleFraccion.MontoSubtotal-45000) > 0.001 {
		t.Fatalf("expected subtotal 45000, got %.2f", detalleFraccion.MontoSubtotal)
	}
	if !detalleFraccion.MontoMinimoAplicado {
		t.Fatalf("expected monto_minimo_aplicado=true")
	}
	if math.Abs(detalleFraccion.MontoTotal-50000) > 0.001 {
		t.Fatalf("expected total 50000 by minimum, got %.2f", detalleFraccion.MontoTotal)
	}

	detalleMaximo := CalcularDetalleTarifaPorMinutos(tarifa, 300, *cfg)
	if detalleMaximo.BloquesExtra != 3 {
		t.Fatalf("expected bloques_extra=3, got %d", detalleMaximo.BloquesExtra)
	}
	if !detalleMaximo.MontoMaximoAplicado {
		t.Fatalf("expected monto_maximo_aplicado=true")
	}
	if math.Abs(detalleMaximo.MontoTotal-70000) > 0.001 {
		t.Fatalf("expected total 70000 by maximum, got %.2f", detalleMaximo.MontoTotal)
	}
}

func TestApplyEmpresaTarifaPorMinutosToAllStations(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaTarifasPorMinutosSchema(dbConn); err != nil {
		t.Fatalf("ensure tarifas por minutos schema: %v", err)
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	_, err := dbConn.Exec(`INSERT INTO carritos_compras (empresa_id, codigo, nombre, referencia_externa, estado, estado_carrito, moneda)
	VALUES
		(5, 'EST-5-11', 'Estacion 11', 'ESTACION_11', 'activo', 'abierto', 'COP'),
		(5, 'EST-5-12', 'Estacion 12', 'ESTACION_12', 'activo', 'abierto', 'COP')`)
	if err != nil {
		t.Fatalf("insert estaciones carritos: %v", err)
	}

	result, err := ApplyEmpresaTarifaPorMinutosToAllStations(dbConn, EmpresaTarifaPorMinutos{
		EmpresaID:      5,
		DiaSemanaDesde: 1,
		DiaSemanaHasta: 7,
		MinutosBase:    120,
		ValorBase:      25000,
		MinutosExtra:   60,
		ValorExtra:     12000,
		Moneda:         "COP",
		Prioridad:      1,
		Estado:         "activo",
		UsuarioCreador: "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("apply all stations create: %v", err)
	}
	if result.EstacionesObjetivo < 2 {
		t.Fatalf("expected at least 2 stations, got %d", result.EstacionesObjetivo)
	}
	if result.TarifasCreadas < 2 {
		t.Fatalf("expected at least 2 created tariffs, got %d", result.TarifasCreadas)
	}

	resultUpdate, err := ApplyEmpresaTarifaPorMinutosToAllStations(dbConn, EmpresaTarifaPorMinutos{
		EmpresaID:      5,
		DiaSemanaDesde: 1,
		DiaSemanaHasta: 7,
		MinutosBase:    120,
		ValorBase:      27000,
		MinutosExtra:   60,
		ValorExtra:     13000,
		Moneda:         "COP",
		Prioridad:      1,
		Estado:         "activo",
		UsuarioCreador: "qa2@empresa.com",
	})
	if err != nil {
		t.Fatalf("apply all stations update: %v", err)
	}
	if resultUpdate.TarifasActualizadas < 2 {
		t.Fatalf("expected at least 2 updated tariffs, got %d", resultUpdate.TarifasActualizadas)
	}

	rows, err := ListEmpresaTarifasPorMinutos(dbConn, 5, EmpresaTarifaPorMinutosFilter{EstacionID: 11, DiaSemana: 2, Limit: 10})
	if err != nil {
		t.Fatalf("list tarifas estacion 11: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 tarifa for station 11, got %d", len(rows))
	}
	if math.Abs(rows[0].ValorBase-27000) > 0.001 {
		t.Fatalf("expected valor_base updated to 27000, got %.2f", rows[0].ValorBase)
	}
}

func TestRegisterTarifaPorMinutosCalculoContable(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaTarifasPorMinutosSchema(dbConn); err != nil {
		t.Fatalf("ensure tarifas por minutos schema: %v", err)
	}
	if err := EnsureEmpresaEventosContablesSchema(dbConn); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	id, err := CreateEmpresaTarifaPorMinutos(dbConn, EmpresaTarifaPorMinutos{
		EmpresaID:      8,
		EstacionID:     20,
		DiaSemanaDesde: 1,
		DiaSemanaHasta: 7,
		MinutosBase:    120,
		ValorBase:      30000,
		MinutosExtra:   60,
		ValorExtra:     15000,
		Moneda:         "COP",
		Prioridad:      1,
		Estado:         "activo",
		UsuarioCreador: "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("create tarifa: %v", err)
	}
	tarifa, err := GetEmpresaTarifaPorMinutosByID(dbConn, 8, id)
	if err != nil {
		t.Fatalf("get tarifa by id: %v", err)
	}
	cfg, err := UpsertEmpresaTarifaPorMinutosConfiguracion(dbConn, EmpresaTarifaPorMinutosConfiguracion{
		EmpresaID:         8,
		RedondeoModo:      "matematico",
		RedondeoUnidad:    100,
		MontoMinimoDiario: 0,
		MontoMaximoDiario: 90000,
	})
	if err != nil {
		t.Fatalf("upsert cfg: %v", err)
	}

	detalle := CalcularDetalleTarifaPorMinutos(*tarifa, 190, *cfg)
	eventoID, documentoCodigo, periodo, err := RegisterTarifaPorMinutosCalculoContable(
		dbConn,
		8,
		*tarifa,
		*cfg,
		2,
		190,
		detalle,
		"qa@empresa.com",
		"REQ-TPM-1",
	)
	if err != nil {
		t.Fatalf("register contable trace: %v", err)
	}
	if eventoID <= 0 {
		t.Fatalf("expected evento_id > 0, got %d", eventoID)
	}
	if !strings.HasPrefix(documentoCodigo, "TPM-8-") {
		t.Fatalf("unexpected documento_codigo: %q", documentoCodigo)
	}
	if len(periodo) != 7 {
		t.Fatalf("expected periodo_contable length 7, got %q", periodo)
	}

	rows, err := ListEmpresaEventosContables(dbConn, 8, EmpresaEventoContableFilter{Evento: "tarifa_por_minutos_calculada", Limit: 10})
	if err != nil {
		t.Fatalf("list eventos contables: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 evento contable, got %d", len(rows))
	}
	if rows[0].DocumentoCodigo != documentoCodigo {
		t.Fatalf("expected documento_codigo %q, got %q", documentoCodigo, rows[0].DocumentoCodigo)
	}
	if !rows[0].Procesado {
		t.Fatalf("expected evento procesado=true")
	}
}

func TestRefreshCarritoTotalConTarifaPorMinutos(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := EnsureEmpresaTarifasPorMinutosSchema(dbConn); err != nil {
		t.Fatalf("ensure tarifas por minutos schema: %v", err)
	}

	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:         77,
		Codigo:            "EST-77-1",
		Nombre:            "Habitacion 1",
		CanalVenta:        "estacion",
		Moneda:            "COP",
		ReferenciaExterna: "ESTACION_1",
		UsuarioCreador:    "qa@empresa.com",
		Estado:            "activo",
	})
	if err != nil {
		t.Fatalf("create carrito: %v", err)
	}
	if _, err := dbConn.Exec(`UPDATE carritos_compras SET estado='activo', estado_carrito='abierto', activado_en=?, pagado_en='' WHERE empresa_id=? AND id=?`, time.Now().Add(-150*time.Minute).Format("2006-01-02 15:04:05"), 77, carritoID); err != nil {
		t.Fatalf("seed carrito activo: %v", err)
	}
	if _, err := CreateEmpresaTarifaPorMinutos(dbConn, EmpresaTarifaPorMinutos{
		EmpresaID:      77,
		EstacionID:     1,
		EstacionCodigo: "EST-77-1",
		EstacionNombre: "Habitacion 1",
		DiaSemanaDesde: 1,
		DiaSemanaHasta: 7,
		MinutosBase:    120,
		ValorBase:      55000,
		MinutosExtra:   60,
		ValorExtra:     30000,
		Moneda:         "COP",
		Prioridad:      1,
		UsuarioCreador: "qa@empresa.com",
		Estado:         "activo",
	}); err != nil {
		t.Fatalf("create tarifa por minutos: %v", err)
	}

	calc, err := RefreshCarritoTotalConTarifaPorMinutos(dbConn, 77, carritoID, time.Now())
	if err != nil {
		t.Fatalf("refresh carrito tarifa por minutos: %v", err)
	}
	if calc == nil || calc.TarifaID <= 0 {
		t.Fatalf("expected tarifa por minutos applied, got %#v", calc)
	}
	if calc.BloquesExtra != 1 {
		t.Fatalf("expected 1 bloque extra, got %d", calc.BloquesExtra)
	}
	if math.Abs(calc.MontoTarifa-85000) > 0.001 {
		t.Fatalf("expected monto_tarifa 85000, got %.2f", calc.MontoTarifa)
	}
	carrito, err := GetCarritoCompraByID(dbConn, 77, carritoID)
	if err != nil {
		t.Fatalf("get carrito after refresh: %v", err)
	}
	if math.Abs(carrito.Total-85000) > 0.001 {
		t.Fatalf("expected carrito total 85000, got %.2f", carrito.Total)
	}
	resumen, err := ResolveCarritoTarifaPorMinutosResumen(dbConn, *carrito, time.Now())
	if err != nil {
		t.Fatalf("resolve resumen tarifa por minutos: %v", err)
	}
	if resumen == nil || strings.TrimSpace(resumen.FechaFinTarifaActual) == "" {
		t.Fatalf("expected fecha_fin_tarifa_actual in resumen, got %#v", resumen)
	}
}
