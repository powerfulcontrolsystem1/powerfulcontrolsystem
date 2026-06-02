package handlers

import (
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaUsuarioEstadoBloqueaPrimerIngresoPermitePendienteInactivo(t *testing.T) {
	item := &dbpkg.EmpresaUsuario{
		Estado:          "inactivo",
		EmailConfirmado: 0,
	}

	if empresaUsuarioEstadoBloqueaPrimerIngreso(item) {
		t.Fatal("usuario pendiente con invitacion valida no debe bloquear el primer ingreso")
	}
}

func TestEmpresaUsuarioEstadoBloqueaPrimerIngresoBloqueaConfirmadoInactivo(t *testing.T) {
	item := &dbpkg.EmpresaUsuario{
		Estado:          "inactivo",
		EmailConfirmado: 1,
	}

	if !empresaUsuarioEstadoBloqueaPrimerIngreso(item) {
		t.Fatal("usuario confirmado e inactivo debe quedar bloqueado")
	}
}

func TestNormalizePermissionRoleCajaEsCajero(t *testing.T) {
	for _, raw := range []string{"Caja", "caja", "Caja principal", "caja_turno"} {
		if got := normalizePermissionRole(raw); got != "cajero" {
			t.Fatalf("normalizePermissionRole(%q)=%q, want cajero", raw, got)
		}
	}
}

func TestCajeroSoloVePaginasOperativas(t *testing.T) {
	allowed := []string{"linkVentaDirecta", "linkEstaciones", "linkCorteCaja"}
	for _, page := range allowed {
		if !isAllowedPageForOperationalRole("cajero", page) {
			t.Fatalf("cajero debe poder ver %s", page)
		}
	}
	blocked := []string{"linkPanelEmpresa", "linkUsuarios", "linkProductos", "linkFinanzas", "linkConfiguracion", "linkReportes"}
	for _, page := range blocked {
		if isAllowedPageForOperationalRole("cajero", page) {
			t.Fatalf("cajero no debe poder ver %s", page)
		}
	}
}
