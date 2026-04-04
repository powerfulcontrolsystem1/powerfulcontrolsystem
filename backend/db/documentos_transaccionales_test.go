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
		EmpresaID:        31,
		TipoDocumento:    "factura_electronica",
		DocumentoCodigo:  "fac-1001",
		NumeroLegal:      "FE-1",
		CodigoValidacion: "ABC123",
		PaisCodigo:       "co",
		AmbienteFE:       "Sandbox",
		EstadoDocumento:  "emitida",
		EstadoAnterior:   "borrador",
		EventoUltimo:     "factura_emitida",
		PeriodoContable:  "2026-04",
		MontoTotal:       120000,
		Moneda:           "cop",
		UsuarioCreador:   "tester",
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
	if created.NumeroLegal != "FE-1" {
		t.Fatalf("expected numero_legal FE-1, got %q", created.NumeroLegal)
	}
	if created.CodigoValidacion != "ABC123" {
		t.Fatalf("expected codigo_validacion ABC123, got %q", created.CodigoValidacion)
	}
	if created.PaisCodigo != "CO" {
		t.Fatalf("expected pais_codigo CO, got %q", created.PaisCodigo)
	}
	if created.AmbienteFE != "sandbox" {
		t.Fatalf("expected ambiente_fe sandbox, got %q", created.AmbienteFE)
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

func TestEmpresaDocumentoCompraListAndSetEstadoByCodigo(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	if _, err := UpsertEmpresaDocumentoCompra(dbConn, EmpresaDocumentoCompra{
		EmpresaID:       77,
		ProveedorID:     501,
		TipoDocumento:   "orden_compra",
		DocumentoCodigo: "oc-a1",
		EstadoDocumento: "borrador",
		EventoUltimo:    "orden_compra_creada",
		PeriodoContable: "2026-04",
		MontoTotal:      200000,
		Moneda:          "cop",
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("upsert compra a1: %v", err)
	}
	if _, err := UpsertEmpresaDocumentoCompra(dbConn, EmpresaDocumentoCompra{
		EmpresaID:       77,
		ProveedorID:     502,
		TipoDocumento:   "orden_compra",
		DocumentoCodigo: "oc-a2",
		EstadoDocumento: "emitida",
		EstadoAnterior:  "borrador",
		EventoUltimo:    "orden_compra_emitida",
		PeriodoContable: "2026-04",
		MontoTotal:      450000,
		Moneda:          "cop",
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("upsert compra a2: %v", err)
	}

	rows, err := ListEmpresaDocumentosCompraByEmpresa(dbConn, 77, "orden_compra", 0, "", false, "", 50, 0)
	if err != nil {
		t.Fatalf("list compras activos inicial: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows before deactivate, got %d", len(rows))
	}

	if err := SetEmpresaDocumentoCompraEstadoByCodigo(dbConn, 77, "orden_compra", "OC-A1", "inactivo"); err != nil {
		t.Fatalf("set estado inactivo: %v", err)
	}

	rows, err = ListEmpresaDocumentosCompraByEmpresa(dbConn, 77, "orden_compra", 0, "", false, "", 50, 0)
	if err != nil {
		t.Fatalf("list compras activos after deactivate: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after deactivate, got %d", len(rows))
	}
	if rows[0].DocumentoCodigo != "OC-A2" {
		t.Fatalf("expected remaining codigo OC-A2, got %q", rows[0].DocumentoCodigo)
	}

	rows, err = ListEmpresaDocumentosCompraByEmpresa(dbConn, 77, "orden_compra", 0, "", true, "A1", 50, 0)
	if err != nil {
		t.Fatalf("list compras include inactive by search: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row for search A1, got %d", len(rows))
	}
	if rows[0].Estado != "inactivo" {
		t.Fatalf("expected row estado inactivo, got %q", rows[0].Estado)
	}

	if err := SetEmpresaDocumentoCompraEstadoByCodigo(dbConn, 77, "orden_compra", "OC-404", "inactivo"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows when codigo not found, got %v", err)
	}
}
