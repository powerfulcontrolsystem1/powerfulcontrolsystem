package db

import "testing"

func TestNormalizeComprasAvanzadas(t *testing.T) {
	req := normalizeCompraRequisicion(EmpresaCompraRequisicion{Codigo: " req-1 ", Prioridad: "URGENTE", EstadoFlujo: "solicitada"})
	if req.Codigo != "REQ-1" || req.Prioridad != "urgente" || req.EstadoFlujo != "solicitada" {
		t.Fatalf("normalizacion requisicion inesperada: %+v", req)
	}
	cot := normalizeCompraCotizacion(EmpresaCompraCotizacion{ProveedorNombre: " Proveedor ", Numero: " c1 ", Subtotal: 1000, Impuestos: 190})
	if cot.Numero != "C1" || cot.Total != 1190 || cot.Estado != "recibida" {
		t.Fatalf("normalizacion cotizacion inesperada: %+v", cot)
	}
	rec := normalizeCompraRecepcionItem(EmpresaCompraRecepcionItem{CantidadOrdenada: 10, CantidadRecibida: 4, CostoUnitario: 123.456})
	if rec.CantidadPendiente != 6 || rec.CostoUnitario != 123.46 {
		t.Fatalf("normalizacion recepcion inesperada: %+v", rec)
	}
}
