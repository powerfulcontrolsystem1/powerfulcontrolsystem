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

func TestEmpresaDocumentoFacturacionListByEmpresaFiltrosClienteFecha(t *testing.T) {
	dbConn := openFinanzasTestDB(t)
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}
	if err := EnsureEmpresaDocumentosTransaccionalesSchema(dbConn); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	resAna, err := dbConn.Exec(`INSERT INTO clientes (empresa_id, tipo_documento, numero_documento, nombre_razon_social, email, estado)
	VALUES (?, 'CC', '1001', 'Ana Gomez', 'ana@example.com', 'activo')`, 91)
	if err != nil {
		t.Fatalf("insert cliente ana: %v", err)
	}
	anaID, _ := resAna.LastInsertId()

	resCarlos, err := dbConn.Exec(`INSERT INTO clientes (empresa_id, tipo_documento, numero_documento, nombre_razon_social, email, estado)
	VALUES (?, 'CC', '1002', 'Carlos Ruiz', 'carlos@example.com', 'activo')`, 91)
	if err != nil {
		t.Fatalf("insert cliente carlos: %v", err)
	}
	carlosID, _ := resCarlos.LastInsertId()

	if _, err := UpsertEmpresaDocumentoFacturacion(dbConn, EmpresaDocumentoFacturacion{
		EmpresaID:            91,
		TipoDocumento:        "factura_electronica",
		DocumentoCodigo:      "FAC-9101",
		NumeroLegal:          "FE-9101",
		CodigoValidacion:     "VAL9101",
		EstadoDocumento:      "emitida",
		EntidadRelacionadaID: anaID,
		FechaDocumento:       "2026-04-05",
		Moneda:               "COP",
		MontoTotal:           98000,
		Estado:               "activo",
	}); err != nil {
		t.Fatalf("upsert facturacion FAC-9101: %v", err)
	}

	if _, err := UpsertEmpresaDocumentoFacturacion(dbConn, EmpresaDocumentoFacturacion{
		EmpresaID:            91,
		TipoDocumento:        "factura_electronica",
		DocumentoCodigo:      "FAC-9102",
		NumeroLegal:          "FE-9102",
		CodigoValidacion:     "VAL9102",
		EstadoDocumento:      "emitida",
		EntidadRelacionadaID: carlosID,
		FechaDocumento:       "2026-04-03",
		Moneda:               "COP",
		MontoTotal:           123000,
		Estado:               "inactivo",
	}); err != nil {
		t.Fatalf("upsert facturacion FAC-9102: %v", err)
	}

	rows, err := ListEmpresaDocumentosFacturacionByEmpresa(dbConn, EmpresaDocumentoFacturacionListFilter{
		EmpresaID:    91,
		ClienteQuery: "ana",
	})
	if err != nil {
		t.Fatalf("list facturacion by cliente query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row by cliente query with activos, got %d", len(rows))
	}
	if rows[0].DocumentoCodigo != "FAC-9101" {
		t.Fatalf("expected FAC-9101, got %q", rows[0].DocumentoCodigo)
	}
	if rows[0].ClienteNombre != "Ana Gomez" {
		t.Fatalf("expected cliente Ana Gomez, got %q", rows[0].ClienteNombre)
	}
	if rows[0].ClienteEmail != "ana@example.com" {
		t.Fatalf("expected cliente email ana@example.com, got %q", rows[0].ClienteEmail)
	}

	rows, err = ListEmpresaDocumentosFacturacionByEmpresa(dbConn, EmpresaDocumentoFacturacionListFilter{
		EmpresaID:       91,
		IncludeInactive: true,
		FechaHasta:      "2026-04-04",
	})
	if err != nil {
		t.Fatalf("list facturacion by fecha_hasta include inactive: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row by fecha_hasta include inactive, got %d", len(rows))
	}
	if rows[0].DocumentoCodigo != "FAC-9102" {
		t.Fatalf("expected FAC-9102 by fecha_hasta filter, got %q", rows[0].DocumentoCodigo)
	}

	rows, err = ListEmpresaDocumentosFacturacionByEmpresa(dbConn, EmpresaDocumentoFacturacionListFilter{
		EmpresaID:       91,
		IncludeInactive: true,
		DocumentoQuery:  "FE-9101",
	})
	if err != nil {
		t.Fatalf("list facturacion by documento query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row by documento query, got %d", len(rows))
	}
	if rows[0].DocumentoCodigo != "FAC-9101" {
		t.Fatalf("expected FAC-9101 by documento query, got %q", rows[0].DocumentoCodigo)
	}
}
