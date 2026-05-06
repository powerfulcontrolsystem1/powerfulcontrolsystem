package db

import "testing"

func TestNormalizeInventarioAvanzado(t *testing.T) {
	lote := normalizeInventarioLoteAvanzado(EmpresaInventarioLoteAvanzado{
		LoteCodigo:      " lote-1 ",
		CantidadInicial: 10,
		CostoUnitario:   123.456,
		EstadoCalidad:   "CUARENTENA",
		Estado:          "x",
	})
	if lote.LoteCodigo != "LOTE-1" || lote.CantidadDisponible != 10 || lote.CostoUnitario != 123.46 || lote.EstadoCalidad != "cuarentena" || lote.Estado != "activo" {
		t.Fatalf("lote normalizado inesperado: %+v", lote)
	}
	serial := normalizeInventarioSerialAvanzado(EmpresaInventarioSerialAvanzado{Serial: " ser-1 ", EstadoInventario: "mantenimiento"})
	if serial.Serial != "SER-1" || serial.EstadoInventario != "mantenimiento" || serial.EstadoOperativo != "operativo" {
		t.Fatalf("serial normalizado inesperado: %+v", serial)
	}
	reserva := normalizeInventarioReservaAvanzada(EmpresaInventarioReservaAvanzada{Cantidad: 2.345, OrigenModulo: " Venta ", Estado: "bad"})
	if reserva.Cantidad != 2.35 || reserva.OrigenModulo != "venta" || reserva.Estado != "activa" {
		t.Fatalf("reserva normalizada inesperada: %+v", reserva)
	}
}

func TestApplyInventarioLoteRuntime(t *testing.T) {
	lote := EmpresaInventarioLoteAvanzado{CantidadDisponible: 10, CantidadReservada: 3, CostoUnitario: 2500}
	applyInventarioLoteRuntime(&lote)
	if lote.CantidadLibre != 7 || lote.ValorDisponible != 25000 || lote.EstadoVencimiento != "no_aplica" {
		t.Fatalf("runtime lote inesperado: %+v", lote)
	}
}
