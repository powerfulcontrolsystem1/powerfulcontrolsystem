package handlers

import (
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestValidateCarritoItemPayloadRequiresNaturalPositiveQuantity(t *testing.T) {
	base := dbpkg.CarritoCompraItem{
		EmpresaID:      7,
		CarritoID:      3,
		Descripcion:    "Producto prueba",
		Cantidad:       1,
		PrecioUnitario: 1000,
		TipoItem:       "producto",
	}

	valid := base
	valid.Cantidad = 2
	if err := validateCarritoItemPayload(valid); err != nil {
		t.Fatalf("expected natural quantity to pass, got %v", err)
	}

	for _, tc := range []struct {
		name     string
		cantidad float64
	}{
		{name: "zero", cantidad: 0},
		{name: "negative", cantidad: -1},
		{name: "decimal", cantidad: 1.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := base
			payload.Cantidad = tc.cantidad
			err := validateCarritoItemPayload(payload)
			if err == nil {
				t.Fatalf("expected quantity %v to fail", tc.cantidad)
			}
			if !strings.Contains(err.Error(), "numero natural positivo") {
				t.Fatalf("expected natural quantity error, got %q", err.Error())
			}
		})
	}
}

func TestCarritoDocumentoYCobroKeepsTipOutsideDocumentTotal(t *testing.T) {
	documento, cobro := carritoDocumentoYCobro(100000, 10000, "COP")
	if documento != 100000 {
		t.Fatalf("document total must exclude tip, got %v", documento)
	}
	if cobro != 110000 {
		t.Fatalf("collection total must include tip, got %v", cobro)
	}

	documento, cobro = carritoDocumentoYCobro(100000, -5000, "COP")
	if documento != 100000 || cobro != 100000 {
		t.Fatalf("negative tip must be ignored, got document=%v collection=%v", documento, cobro)
	}
}

func TestBuildVentaDocumentoCodigoUsesImmutableCarritoID(t *testing.T) {
	carritoA := &dbpkg.CarritoCompra{ID: 101, Codigo: "VENTA-DIRECTA-12"}
	carritoB := &dbpkg.CarritoCompra{ID: 102, Codigo: "VENTA-DIRECTA-12"}

	codigoA := buildVentaDocumentoCodigo(carritoA, "comprobante_pago")
	codigoB := buildVentaDocumentoCodigo(carritoB, "comprobante_pago")
	if codigoA == codigoB {
		t.Fatalf("dos carritos distintos no pueden reutilizar codigo de documento: %q", codigoA)
	}
	if codigoA != "CP-VENTA-DIRECTA-12-CRT-101" || codigoB != "CP-VENTA-DIRECTA-12-CRT-102" {
		t.Fatalf("codigos inesperados: %q y %q", codigoA, codigoB)
	}
	if got := buildVentaDocumentoCodigo(carritoA, "factura_electronica"); got != "FV-VENTA-DIRECTA-12-CRT-101" {
		t.Fatalf("codigo de factura electronica=%q", got)
	}
}
