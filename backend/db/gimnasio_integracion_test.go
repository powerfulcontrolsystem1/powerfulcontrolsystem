package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGimnasioCoreCodeEsEstableParaServicios(t *testing.T) {
	got := gymCoreCode("GYM-PLAN", "12", "Mensual estándar")
	if got != "GYM-PLAN-12-MENSUAL-EST-NDAR" {
		t.Fatalf("gymCoreCode() = %q", got)
	}
	if len(got) > 51 {
		t.Fatalf("codigo demasiado largo: %q", got)
	}
}

func TestNormalizeGymPagoUsaMetodoPagoCentral(t *testing.T) {
	row, err := normalizeGymPago(EmpresaGimnasioPago{
		EmpresaID:  7,
		SocioID:    3,
		Concepto:   "Mensualidad",
		Monto:      120000,
		MetodoPago: "transferencia",
	})
	if err != nil {
		t.Fatalf("normalizeGymPago() error = %v", err)
	}
	if row.MetodoPago != "transferencia_bancaria" {
		t.Fatalf("MetodoPago = %q", row.MetodoPago)
	}
	if row.Moneda != "COP" || row.Estado != "pagado" {
		t.Fatalf("defaults invalidos: %+v", row)
	}
}

func TestGimnasioEnsureCreaIndicesDeIntegracionDespuesDeColumnas(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("gimnasio.go"))
	if err != nil {
		t.Fatalf("read gimnasio.go: %v", err)
	}
	src := string(raw)
	ensurePos := strings.Index(src, `ensureColumnIfMissing(dbConn, group.table, column.name, column.def)`)
	if ensurePos < 0 {
		t.Fatal("no se encontro ensureColumnIfMissing en EnsureEmpresaGimnasioSchema")
	}
	for _, indexStmt := range []string{
		`ix_empresa_gimnasio_planes_servicio`,
		`ix_empresa_gimnasio_socios_cliente`,
		`ix_empresa_gimnasio_pagos_carrito`,
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
