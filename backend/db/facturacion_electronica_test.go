package db

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func seedFacturacionLegalConfig(t *testing.T, dbConn *sql.DB, empresaID int64, fechaDesde, fechaHasta string, consecutivoDesde, consecutivoHasta, proximoConsecutivo int64, estadoFE string) {
	t.Helper()

	if err := EnsureEmpresaConfiguracionAvanzadaSchema(dbConn); err != nil {
		t.Fatalf("ensure configuracion avanzada schema: %v", err)
	}
	if err := EnsureEmpresaFacturacionElectronicaSchema(dbConn); err != nil {
		t.Fatalf("ensure facturacion electronica schema: %v", err)
	}

	if _, err := UpsertEmpresaConfiguracionAvanzada(dbConn, EmpresaConfiguracionAvanzada{
		EmpresaID:            empresaID,
		TipoDocumentoEmisor:  "NIT",
		NIT:                  "900123456",
		RazonSocial:          "Empresa QA Facturacion",
		PaisCodigo:           "CO",
		AmbienteFE:           "produccion",
		PrefijoFactura:       "FE",
		ResolucionNumero:     "18760000000001",
		ResolucionFechaDesde: fechaDesde,
		ResolucionFechaHasta: fechaHasta,
		ConsecutivoDesde:     consecutivoDesde,
		ConsecutivoHasta:     consecutivoHasta,
		ProximoConsecutivo:   proximoConsecutivo,
		UsuarioCreador:       "facturacion@test.com",
		Estado:               "activo",
	}); err != nil {
		t.Fatalf("upsert configuracion avanzada: %v", err)
	}

	if strings.TrimSpace(estadoFE) == "" {
		estadoFE = "activo"
	}
	if _, err := UpsertFacturacionElectronicaPaisConfig(dbConn, FacturacionElectronicaPaisConfig{
		EmpresaID:           empresaID,
		PaisCodigo:          "CO",
		Proveedor:           "manual",
		Ambiente:            "produccion",
		TipoDocumentoEmisor: "NIT",
		IdentificadorFiscal: "900123456",
		RazonSocial:         "Empresa QA Facturacion",
		PrefijoFactura:      "FE",
		ResolucionNumero:    "18760000000001",
		Estado:              estadoFE,
		UsuarioCreador:      "facturacion@test.com",
	}); err != nil {
		t.Fatalf("upsert config facturacion pais: %v", err)
	}
}

func TestPrepareFacturacionDocumentoLegalSuccessAndConsecutivo(t *testing.T) {
	dbConn := openFinanzasTestDB(t)

	fechaDesde := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	fechaHasta := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	seedFacturacionLegalConfig(t, dbConn, 401, fechaDesde, fechaHasta, 1, 999999, 7, "activo")

	doc1, err := PrepareFacturacionDocumentoLegal(dbConn, 401, "CO", "FAC-401-01", 150000, "cop")
	if err != nil {
		t.Fatalf("prepare legal doc #1: %v", err)
	}
	if doc1.NumeroLegal != "FE-7" {
		t.Fatalf("expected numero_legal FE-7, got %q", doc1.NumeroLegal)
	}
	if doc1.ConsecutivoAsignado != 7 {
		t.Fatalf("expected consecutivo_asignado=7, got %d", doc1.ConsecutivoAsignado)
	}
	if len(doc1.CodigoValidacion) != 64 {
		t.Fatalf("expected codigo_validacion length 64, got %d", len(doc1.CodigoValidacion))
	}
	if doc1.CodigoValidacion != strings.ToUpper(doc1.CodigoValidacion) {
		t.Fatalf("expected codigo_validacion uppercase, got %q", doc1.CodigoValidacion)
	}

	var proximo1 int64
	if err := dbConn.QueryRow(`SELECT COALESCE(proximo_consecutivo, 0) FROM empresa_configuracion_avanzada WHERE empresa_id = ?`, 401).Scan(&proximo1); err != nil {
		t.Fatalf("query proximo_consecutivo #1: %v", err)
	}
	if proximo1 != 8 {
		t.Fatalf("expected proximo_consecutivo=8 after first emisión, got %d", proximo1)
	}

	doc2, err := PrepareFacturacionDocumentoLegal(dbConn, 401, "CO", "FAC-401-02", 80000, "COP")
	if err != nil {
		t.Fatalf("prepare legal doc #2: %v", err)
	}
	if doc2.NumeroLegal != "FE-8" {
		t.Fatalf("expected numero_legal FE-8, got %q", doc2.NumeroLegal)
	}

	var proximo2 int64
	if err := dbConn.QueryRow(`SELECT COALESCE(proximo_consecutivo, 0) FROM empresa_configuracion_avanzada WHERE empresa_id = ?`, 401).Scan(&proximo2); err != nil {
		t.Fatalf("query proximo_consecutivo #2: %v", err)
	}
	if proximo2 != 9 {
		t.Fatalf("expected proximo_consecutivo=9 after second emisión, got %d", proximo2)
	}
}

func TestPrepareFacturacionDocumentoLegalRejectsExpiredResolution(t *testing.T) {
	dbConn := openFinanzasTestDB(t)

	fechaDesde := time.Now().AddDate(-2, 0, 0).Format("2006-01-02")
	fechaHasta := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	seedFacturacionLegalConfig(t, dbConn, 402, fechaDesde, fechaHasta, 1, 999999, 10, "activo")

	_, err := PrepareFacturacionDocumentoLegal(dbConn, 402, "CO", "FAC-402-01", 45000, "COP")
	if err == nil {
		t.Fatalf("expected error for resolucion vencida")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "vencida") {
		t.Fatalf("expected vencida error, got %v", err)
	}

	var proximo int64
	if err := dbConn.QueryRow(`SELECT COALESCE(proximo_consecutivo, 0) FROM empresa_configuracion_avanzada WHERE empresa_id = ?`, 402).Scan(&proximo); err != nil {
		t.Fatalf("query proximo_consecutivo: %v", err)
	}
	if proximo != 10 {
		t.Fatalf("expected proximo_consecutivo unchanged=10, got %d", proximo)
	}
}

func TestPrepareFacturacionDocumentoLegalRejectsConfigInactivaAndRangoAgotado(t *testing.T) {
	dbConn := openFinanzasTestDB(t)

	fechaDesde := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	fechaHasta := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	seedFacturacionLegalConfig(t, dbConn, 403, fechaDesde, fechaHasta, 1, 10, 1, "inactivo")

	_, err := PrepareFacturacionDocumentoLegal(dbConn, 403, "CO", "FAC-403-01", 30000, "COP")
	if err == nil {
		t.Fatalf("expected error for configuracion inactiva")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "inactiva") {
		t.Fatalf("expected inactiva error, got %v", err)
	}

	seedFacturacionLegalConfig(t, dbConn, 404, fechaDesde, fechaHasta, 1, 1, 1, "activo")

	if _, err := PrepareFacturacionDocumentoLegal(dbConn, 404, "CO", "FAC-404-01", 30000, "COP"); err != nil {
		t.Fatalf("expected first legal doc in bounded range, got %v", err)
	}

	_, err = PrepareFacturacionDocumentoLegal(dbConn, 404, "CO", "FAC-404-02", 25000, "COP")
	if err == nil {
		t.Fatalf("expected error for consecutivo agotado")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "agotado") {
		t.Fatalf("expected agotado error, got %v", err)
	}
}
