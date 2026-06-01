package handlers

import "testing"

func TestAdminEmpresaCompartidaCanManageShares(t *testing.T) {
	t.Parallel()

	if !adminEmpresaCompartidaCanManageShares(true, false) {
		t.Fatal("owner should manage shares")
	}
	if !adminEmpresaCompartidaCanManageShares(false, true) {
		t.Fatal("super administrator should manage shares even when not owner")
	}
	if adminEmpresaCompartidaCanManageShares(false, false) {
		t.Fatal("non-owner non-super should not manage shares")
	}
}

func TestAdminEmpresaCompartidaActorEmail(t *testing.T) {
	t.Parallel()

	if got := adminEmpresaCompartidaActorEmail("super@example.com", "", true); got != "super@example.com" {
		t.Fatalf("super actor email = %q", got)
	}
	if got := adminEmpresaCompartidaActorEmail("delegado@example.com", "principal@example.com", false); got != "principal@example.com" {
		t.Fatalf("delegated actor email = %q", got)
	}
	if got := adminEmpresaCompartidaActorEmail("admin@example.com", "", false); got != "admin@example.com" {
		t.Fatalf("owner actor fallback = %q", got)
	}
}
