package db

import "testing"

func TestNormalizeTesoreriaCuentaDefaults(t *testing.T) {
	got := normalizeTesoreriaCuenta(EmpresaTesoreriaCuenta{
		Codigo:       " banco 01 ",
		Nombre:       "  Cuenta principal ",
		Tipo:         "Cuenta Bancaria",
		SaldoInicial: 100,
		Estado:       "raro",
	})
	if got.Codigo != "BANCO 01" || got.Nombre != "Cuenta principal" {
		t.Fatalf("campos base no normalizados: %#v", got)
	}
	if got.Tipo != "banco" || got.Moneda != "COP" || got.SaldoActual != 100 || got.Estado != "activo" {
		t.Fatalf("defaults de cuenta inesperados: %#v", got)
	}
}

func TestNormalizeTesoreriaPresupuesto(t *testing.T) {
	got := normalizeTesoreriaPresupuesto(EmpresaTesoreriaPresupuesto{
		Codigo:       " pres-01 ",
		Nombre:       "  Presupuesto base ",
		Escenario:    "OPTIMISTA",
		Estado:       "APROBADO",
		IngresosMeta: -1,
		EgresosMeta:  500,
	})
	if got.Codigo != "PRES-01" || got.Nombre != "Presupuesto base" {
		t.Fatalf("presupuesto base no normalizado: %#v", got)
	}
	if got.Escenario != "optimista" || got.Estado != "aprobado" || got.IngresosMeta != 0 || got.EgresosMeta != 500 {
		t.Fatalf("catalogos/valores inesperados: %#v", got)
	}
}

func TestNormalizeTesoreriaPartidaYFlujo(t *testing.T) {
	partida := normalizeTesoreriaPartida(EmpresaTesoreriaPartida{Tipo: "Ingreso", Categoria: "Ventas Netas", Concepto: "  Venta mensual ", ValorPresupuestado: -10, Periodicidad: "Semanal"})
	if partida.Tipo != "ingreso" || partida.Categoria != "ventas_netas" || partida.Concepto != "Venta mensual" || partida.ValorPresupuestado != 0 || partida.Periodicidad != "semanal" {
		t.Fatalf("partida no normalizada: %#v", partida)
	}
	flujo := normalizeTesoreriaFlujo(EmpresaTesoreriaFlujo{FechaFlujo: "2026-05-20", Tipo: "otro", Categoria: "", Concepto: " Pago ", Valor: -5, Estado: "programado"})
	if flujo.Periodo != "2026-05" || flujo.Tipo != "egreso" || flujo.Categoria != "general" || flujo.Valor != 0 || flujo.Estado != "programado" {
		t.Fatalf("flujo no normalizado: %#v", flujo)
	}
}
