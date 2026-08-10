package handlers

import (
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestFacturaElectronicaPendienteDebeSincronizarCliente(t *testing.T) {
	tests := []struct {
		name     string
		existing *dbpkg.EmpresaDocumentoFacturacion
		venta    *dbpkg.EmpresaDocumentoFacturacion
		want     bool
	}{
		{
			name:     "pendiente sin cliente hereda cliente de venta",
			existing: &dbpkg.EmpresaDocumentoFacturacion{EstadoDocumento: "pendiente_emision"},
			venta:    &dbpkg.EmpresaDocumentoFacturacion{EntidadRelacionadaID: 22},
			want:     true,
		},
		{
			name:     "borrador sin cliente hereda cliente de venta",
			existing: &dbpkg.EmpresaDocumentoFacturacion{EstadoDocumento: "borrador"},
			venta:    &dbpkg.EmpresaDocumentoFacturacion{EntidadRelacionadaID: 22},
			want:     true,
		},
		{
			name:     "factura emitida queda inmutable",
			existing: &dbpkg.EmpresaDocumentoFacturacion{EstadoDocumento: "emitida"},
			venta:    &dbpkg.EmpresaDocumentoFacturacion{EntidadRelacionadaID: 22},
			want:     false,
		},
		{
			name:     "cliente existente no se reemplaza",
			existing: &dbpkg.EmpresaDocumentoFacturacion{EstadoDocumento: "pendiente_emision", EntidadRelacionadaID: 29},
			venta:    &dbpkg.EmpresaDocumentoFacturacion{EntidadRelacionadaID: 22},
			want:     false,
		},
		{
			name:     "venta sin cliente no cambia factura",
			existing: &dbpkg.EmpresaDocumentoFacturacion{EstadoDocumento: "pendiente_emision"},
			venta:    &dbpkg.EmpresaDocumentoFacturacion{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := facturaElectronicaPendienteDebeSincronizarCliente(tt.existing, tt.venta); got != tt.want {
				t.Fatalf("facturaElectronicaPendienteDebeSincronizarCliente() = %v, want %v", got, tt.want)
			}
		})
	}
}
