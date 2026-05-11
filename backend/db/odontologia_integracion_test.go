package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOdontoCoreCodeEsEstableParaServicios(t *testing.T) {
	got := odontoCoreCode("OD-TRAT", "15", "Ortodoncia estetica")
	if got != "OD-TRAT-15-ORTODONCIA-ESTETICA" {
		t.Fatalf("odontoCoreCode() = %q", got)
	}
	if len(got) > 50 {
		t.Fatalf("codigo demasiado largo: %q", got)
	}
}

func TestOdontoPagoUsaMetodoPagoCentral(t *testing.T) {
	got := NormalizeMetodoPagoCarrito("transferencia")
	if got != "transferencia_bancaria" {
		t.Fatalf("NormalizeMetodoPagoCarrito() = %q", got)
	}
}

func TestOdontoPagoCarritoReferenciaUsaIDParaSincronizacion(t *testing.T) {
	pago := EmpresaOdontologiaPago{ID: 42, PacienteID: 7, Concepto: "Abono", Referencia: "REF-123", FechaPago: "2026-05-12 10:00:00", Monto: 30000}
	if got := odontoPagoCarritoReferencia(pago); got != "odontologia:pago:42" {
		t.Fatalf("odontoPagoCarritoReferencia() = %q", got)
	}
	if got := odontoPagoCarritoNombre(pago); got != "Odontologia - Abono #42" {
		t.Fatalf("odontoPagoCarritoNombre() = %q", got)
	}
}

func TestOdontologiaEnsureCreaIndicesDeIntegracionDespuesDeColumnas(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("odontologia.go"))
	if err != nil {
		t.Fatalf("read odontologia.go: %v", err)
	}
	src := string(raw)
	ensurePos := strings.Index(src, `ensureColumnIfMissing(dbConn, group.table, column.name, column.def)`)
	if ensurePos < 0 {
		t.Fatal("no se encontro ensureColumnIfMissing en EnsureEmpresaOdontologiaSchema")
	}
	for _, indexStmt := range []string{
		`ix_empresa_odontologia_pacientes_cliente`,
		`ix_empresa_odontologia_tratamientos_servicio`,
		`ix_empresa_odontologia_pagos_carrito`,
	} {
		indexPos := strings.Index(src, indexStmt)
		if indexPos < 0 {
			t.Fatalf("no se encontro indice de integracion %s", indexStmt)
		}
		if indexPos < ensurePos {
			t.Fatalf("el indice %s se crea antes de asegurar columnas de integracion", indexStmt)
		}
	}
}
