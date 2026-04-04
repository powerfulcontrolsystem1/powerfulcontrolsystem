package db

import (
	"database/sql"
	"errors"
	"testing"
)

func TestEmpresaDocumentoFacturacionUpsertAndGet(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	created, err := UpsertEmpresaDocumentoFacturacion(dbConn, EmpresaDocumentoFacturacion{
		EmpresaID:       31,
		TipoDocumento:   "factura_electronica",
		DocumentoCodigo: "fac-1001",
		EstadoDocumento: "emitida",
		EstadoAnterior:  "borrador",
		EventoUltimo:    "factura_emitida",
		PeriodoContable: "2026-04",
		MontoTotal:      120000,
		Moneda:          "cop",
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("upsert facturacion document: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected document id > 0")
	}
	if created.DocumentoCodigo != "FAC-1001" {
		t.Fatalf("expected documento_codigo FAC-1001, got %q", created.DocumentoCodigo)
	}

	updated, err := UpsertEmpresaDocumentoFacturacion(dbConn, EmpresaDocumentoFacturacion{
		EmpresaID:       31,
		TipoDocumento:   "factura_electronica",
		DocumentoCodigo: "FAC-1001",
		EstadoDocumento: "anulada",
		EstadoAnterior:  "emitida",
		EventoUltimo:    "factura_anulada",
	})
	if err != nil {
		t.Fatalf("upsert facturacion update: %v", err)
	}
	if updated.ID != created.ID {
		t.Fatalf("expected same id %d after update, got %d", created.ID, updated.ID)
	}
	if updated.EstadoDocumento != "anulada" {
		t.Fatalf("expected estado_documento anulada, got %q", updated.EstadoDocumento)
	}
}

func TestEmpresaDocumentoCompraUpsertAndGet(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	created, err := UpsertEmpresaDocumentoCompra(dbConn, EmpresaDocumentoCompra{
		EmpresaID:       12,
		ProveedorID:     44,
		TipoDocumento:   "orden_compra",
		DocumentoCodigo: "oc-1001",
		EstadoDocumento: "emitida",
		EstadoAnterior:  "borrador",
		EventoUltimo:    "orden_compra_emitida",
		PeriodoContable: "2026-04",
		MontoTotal:      500000,
		Moneda:          "cop",
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("upsert compra document: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected document id > 0")
	}
	if created.DocumentoCodigo != "OC-1001" {
		t.Fatalf("expected documento_codigo OC-1001, got %q", created.DocumentoCodigo)
	}

	loaded, err := GetEmpresaDocumentoCompraByCodigo(dbConn, 12, "orden_compra", "OC-1001")
	if err != nil {
		t.Fatalf("get compra document: %v", err)
	}
	if loaded.ID != created.ID {
		t.Fatalf("expected loaded id %d, got %d", created.ID, loaded.ID)
	}

	_, err = GetEmpresaDocumentoCompraByCodigo(dbConn, 12, "orden_compra", "OC-9999")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for missing document, got %v", err)
	}
}
