package handlers

import (
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestSuperReportesGlobalesUsaMismoAlcanceDelSelector(t *testing.T) {
	empresas := []dbpkg.Empresa{
		{ID: 10, EmpresaID: 10, Nombre: "Empresa propia", UsuarioCreador: "super@example.com"},
		{ID: 20, EmpresaID: 20, Nombre: "Empresa ajena", UsuarioCreador: "otro@example.com"},
	}

	got, err := superReportesFiltrarEmpresasPermitidas(nil, "super@example.com", "", empresas)
	if err != nil {
		t.Fatalf("superReportesFiltrarEmpresasPermitidas returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected only visible selector empresas, got %d", len(got))
	}
	if got[0].ID != 10 || got[0].AccessSource != "owner" {
		t.Fatalf("expected own empresa marked as owner, got id=%d source=%q", got[0].ID, got[0].AccessSource)
	}
}

func TestSuperReportesGlobalesIncluyeEmpresasDelegadasDelSelector(t *testing.T) {
	empresas := []dbpkg.Empresa{
		{ID: 10, EmpresaID: 10, Nombre: "Principal A", UsuarioCreador: "principal@example.com"},
		{ID: 20, EmpresaID: 20, Nombre: "Ajena", UsuarioCreador: "otro@example.com"},
	}

	got, err := superReportesFiltrarEmpresasPermitidas(nil, "delegado@example.com", "principal@example.com", empresas)
	if err != nil {
		t.Fatalf("superReportesFiltrarEmpresasPermitidas returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected only delegated principal empresa, got %d", len(got))
	}
	if got[0].ID != 10 || got[0].AccessSource != "delegated" {
		t.Fatalf("expected delegated empresa, got id=%d source=%q", got[0].ID, got[0].AccessSource)
	}
}
