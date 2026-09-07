package db

import "testing"

func TestCalculateEmpresaFinanzasRentaEvitaDobleConteo(t *testing.T) {
	got := CalculateEmpresaFinanzasRenta(EmpresaFinanzasRentaInputs{
		EmpresaID:                  12,
		TarifaRenta:                35,
		UsarVentasPOSComoIngreso:   true,
		UsarMovimientosComoIngreso: true,
		UsarComprasYNominaEgreso:   true,
		UsarMovimientosComoEgreso:  true,
		RetencionesAdicionales:     1000,
		DescuentosTributarios:      500,
	}, empresaFinanzasRentaBaseDatos{
		EmpresaID:               12,
		Moneda:                  "COP",
		VentasPOS:               100000,
		IngresosMovimientos:     90000,
		EgresosMovimientos:      20000,
		ComprasInventario:       12000,
		NominaDevengada:         3000,
		RetencionesIngresos:     1500,
		VentasPOSRegistros:      4,
		IngresosMovimientosRows: 3,
		EgresosMovimientosRows:  2,
	})

	if got.IngresosBase != 100000 {
		t.Fatalf("ingresos base debe usar el mayor para evitar doble conteo, got %.2f", got.IngresosBase)
	}
	if got.DeduccionesBase != 20000 {
		t.Fatalf("deducciones base debe usar egresos financieros al ser mayor, got %.2f", got.DeduccionesBase)
	}
	if got.RentaLiquidaGravable != 80000 || got.ImpuestoRentaEstimado != 28000 {
		t.Fatalf("renta/impuesto inesperados: %#v", got)
	}
	if got.SaldoEstimado != 25000 {
		t.Fatalf("saldo esperado 25000, got %.2f", got.SaldoEstimado)
	}
	if len(got.Alertas) < 2 {
		t.Fatalf("debe alertar doble conteo potencial, got %#v", got.Alertas)
	}
}

func TestNormalizeEmpresaFinanzasRentaInputsDefaults(t *testing.T) {
	got := NormalizeEmpresaFinanzasRentaInputs(EmpresaFinanzasRentaInputs{TarifaRenta: -1})
	if got.TarifaRenta != EmpresaRentaTarifaGeneralColombia {
		t.Fatalf("tarifa default inesperada %.2f", got.TarifaRenta)
	}
	if !got.UsarMovimientosComoIngreso || !got.UsarMovimientosComoEgreso {
		t.Fatalf("debe activar fuentes financieras por defecto: %#v", got)
	}
}
