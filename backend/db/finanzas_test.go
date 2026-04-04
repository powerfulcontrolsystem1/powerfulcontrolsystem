package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openFinanzasTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "finanzas_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	return dbConn
}

func TestEmpresaFinanzasConfiguracionUpsertAndGet(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	cfgDefault, err := GetEmpresaFinanzasConfiguracion(dbConn, 77)
	if err != nil {
		t.Fatalf("get default config: %v", err)
	}
	if !cfgDefault.HabilitarIngresos || !cfgDefault.HabilitarEgresos {
		t.Fatalf("expected default ingresos/egresos enabled")
	}

	_, err = UpsertEmpresaFinanzasConfiguracion(dbConn, EmpresaFinanzasConfiguracion{
		EmpresaID:                  77,
		HabilitarIngresos:          true,
		HabilitarEgresos:           true,
		Moneda:                     "COP",
		CategoriasIngreso:          "ventas\nservicios",
		CategoriasEgreso:           "compras\nnomina",
		PrefijoIngreso:             "ING",
		PrefijoEgreso:              "EGR",
		FormatoImpresion:           "pos",
		RequiereAprobacion:         true,
		IntegracionContableDestino: "siigo",
		CuentaCajaBancos:           "110505",
		CuentaIngresos:             "413510",
		CuentaIVAGenerado:          "240801",
		CuentaGastos:               "519510",
		CuentaIVADescontable:       "240805",
		CuentasIngresoCategoria:    "ventas=413510",
		CuentasEgresoCategoria:     "compras=613510",
		UsuarioCreador:             "tester",
	})
	if err != nil {
		t.Fatalf("upsert config: %v", err)
	}

	cfg, err := GetEmpresaFinanzasConfiguracion(dbConn, 77)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if cfg.EmpresaID != 77 {
		t.Fatalf("expected empresa_id=77, got %d", cfg.EmpresaID)
	}
	if cfg.FormatoImpresion != "pos" {
		t.Fatalf("expected formato pos, got %s", cfg.FormatoImpresion)
	}
	if !cfg.RequiereAprobacion {
		t.Fatalf("expected requiere_aprobacion=true")
	}
	if cfg.IntegracionContableDestino != "siigo" {
		t.Fatalf("expected integracion siigo, got %s", cfg.IntegracionContableDestino)
	}
	if cfg.CuentaIngresos != "413510" {
		t.Fatalf("expected cuenta ingresos 413510, got %s", cfg.CuentaIngresos)
	}
	if cfg.CuentasIngresoCategoria != "ventas=413510" {
		t.Fatalf("expected cuentas ingreso map ventas=413510, got %s", cfg.CuentasIngresoCategoria)
	}
}

func TestEmpresaFinanzasMovimientosCRUDFlow(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	id, err := CreateEmpresaFinanzasMovimiento(dbConn, EmpresaFinanzasMovimiento{
		EmpresaID:       1,
		TipoMovimiento:  "ingreso",
		Concepto:        "Ingreso test",
		Categoria:       "ventas",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           100000,
		Impuesto:        0,
		Total:           100000,
		TipoComprobante: "recibo_interno",
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create movimiento: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected id > 0")
	}

	rows, err := ListEmpresaFinanzasMovimientos(dbConn, 1, EmpresaFinanzasMovimientoFilter{Limit: 50})
	if err != nil {
		t.Fatalf("list movimientos: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 movimiento, got %d", len(rows))
	}

	mov := rows[0]
	mov.Concepto = "Ingreso test actualizado"
	mov.Total = 110000
	if err := UpdateEmpresaFinanzasMovimiento(dbConn, mov); err != nil {
		t.Fatalf("update movimiento: %v", err)
	}

	if err := SetEmpresaFinanzasMovimientoEstado(dbConn, 1, mov.ID, "inactivo"); err != nil {
		t.Fatalf("set estado: %v", err)
	}

	activos, err := ListEmpresaFinanzasMovimientos(dbConn, 1, EmpresaFinanzasMovimientoFilter{Limit: 50})
	if err != nil {
		t.Fatalf("list activos: %v", err)
	}
	if len(activos) != 0 {
		t.Fatalf("expected 0 activos after inactivar, got %d", len(activos))
	}

	incluyendoInactivos, err := ListEmpresaFinanzasMovimientos(dbConn, 1, EmpresaFinanzasMovimientoFilter{IncludeInactive: true, Limit: 50})
	if err != nil {
		t.Fatalf("list include inactive: %v", err)
	}
	if len(incluyendoInactivos) != 1 {
		t.Fatalf("expected 1 movimiento including inactive, got %d", len(incluyendoInactivos))
	}

	if err := DeleteEmpresaFinanzasMovimiento(dbConn, 1, mov.ID); err != nil {
		t.Fatalf("delete movimiento: %v", err)
	}

	finalRows, err := ListEmpresaFinanzasMovimientos(dbConn, 1, EmpresaFinanzasMovimientoFilter{IncludeInactive: true, Limit: 50})
	if err != nil {
		t.Fatalf("list final: %v", err)
	}
	if len(finalRows) != 0 {
		t.Fatalf("expected 0 movimientos after delete, got %d", len(finalRows))
	}
}

func TestEmpresaFinanzasMovimientoMontoInvalido(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	_, err := CreateEmpresaFinanzasMovimiento(dbConn, EmpresaFinanzasMovimiento{
		EmpresaID:      1,
		TipoMovimiento: "egreso",
		Concepto:       "Pago invalido",
		Monto:          0,
	})
	if err == nil {
		t.Fatalf("expected error for monto <= 0")
	}
}
