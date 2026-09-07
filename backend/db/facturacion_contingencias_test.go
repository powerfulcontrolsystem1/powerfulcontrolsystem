package db

import (
	"context"
	"database/sql"
	"testing"
)

func TestFacturacionContingenciaConfigurationRejectsUnsafeActivation(t *testing.T) {
	base := EmpresaFacturacionContingenciaConfiguracion{
		EmpresaID: 12, PaisCodigo: "CO", Prefijo: "CTG", ResolucionNumero: "QA-NO-AUTORIZADA",
		FechaDesde: "2026-01-01", FechaHasta: "2026-12-31", RangoDesde: 1, RangoHasta: 100,
		ProximoNumero: 1, Estado: "configurando",
	}
	if err := validateFacturacionContingenciaConfiguracion(&base); err != nil {
		t.Fatal(err)
	}
	invalid := base
	invalid.PaisCodigo = "PA"
	if err := validateFacturacionContingenciaConfiguracion(&invalid); err == nil {
		t.Fatal("non-Colombia contingency configuration accepted")
	}
	invalid = base
	invalid.ProximoNumero = 101
	invalid.Estado = "activo"
	if err := validateFacturacionContingenciaConfiguracion(&invalid); err == nil {
		t.Fatal("exhausted paper authorization activated")
	}
}

func TestFacturacionContingenciaMigrationRegistered(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version == "20260906-002-facturacion-contingencias-v1" && migration.Body == empresaFacturacionContingenciasFingerprint && migration.Apply != nil {
			return
		}
	}
	t.Fatal("facturacion contingency migration not registered")
}

func TestApplyFacturacionContingenciasRejectsNilTransaction(t *testing.T) {
	if err := applyEmpresaFacturacionContingenciasTx(context.Background(), (*sql.Tx)(nil)); err == nil {
		t.Fatal("nil transaction accepted")
	}
}
