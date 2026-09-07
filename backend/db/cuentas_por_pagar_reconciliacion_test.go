package db

import "testing"

func TestReconcileEmpresaCxPRowsIsReadOnlyAndTenantScoped(t *testing.T) {
	canonical := []empresaCxPReconciliacionRow{
		{Proveedor: "Proveedor A", Documento: "FAC-1", Moneda: "COP", Original: 100, Saldo: 60},
		{Proveedor: "Proveedor B", Documento: "FAC-2", Moneda: "COP", Original: 50, Saldo: 50},
	}
	historical := []empresaCxPReconciliacionRow{
		{Proveedor: "Proveedor A", Documento: "FAC-1", Moneda: "COP", Original: 100, Saldo: 70},
		{Proveedor: "Proveedor C", Documento: "FAC-3", Moneda: "COP", Original: 40, Saldo: 40},
	}
	report := reconcileEmpresaCxPRows(12, canonical, historical)
	if report.EmpresaID != 12 || !report.SoloLectura {
		t.Fatalf("unexpected tenant/read-only state: %#v", report)
	}
	if report.SoloCanonica != 1 || report.SoloHistorica != 1 || report.ConDiferencias != 1 {
		t.Fatalf("unexpected reconciliation counts: %#v", report)
	}
	if !report.RequiereRevisionHumana || len(report.Items) != 3 {
		t.Fatalf("differences must require human review: %#v", report)
	}
}

func TestEmpresaCxPReconciliacionKeyNormalizesDocumentAndSupplier(t *testing.T) {
	a := empresaCxPReconciliacionKey(empresaCxPReconciliacionRow{Documento: " fac-1 ", Proveedor: "Proveedor A"})
	b := empresaCxPReconciliacionKey(empresaCxPReconciliacionRow{Documento: "FAC-1", Proveedor: " proveedor a "})
	if a != b {
		t.Fatalf("same document/supplier must share a key: %q != %q", a, b)
	}
}
