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
