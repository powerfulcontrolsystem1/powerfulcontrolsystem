package handlers

import (
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaAIWantsUserSummaryQuestion(t *testing.T) {
	cases := []struct {
		name string
		text string
		want bool
	}{
		{name: "usuarios registrados", text: "Cuantos usuarios hay registrados en el sistema?", want: true},
		{name: "total usuarios", text: "Dame el total de usuarios de esta empresa", want: true},
		{name: "sin conteo", text: "Crear un usuario nuevo", want: false},
		{name: "otro dominio", text: "Cuantos productos hay registrados?", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := empresaAIWantsUserSummaryQuestion(foldEmpresaAICommandText(tc.text))
			if got != tc.want {
				t.Fatalf("empresaAIWantsUserSummaryQuestion()=%v want %v", got, tc.want)
			}
		})
	}
}

func TestEmpresaAIAdministrativeDBReadRole(t *testing.T) {
	allowed := []string{"super_administrador", "administrador_total", "administrador", "admin_empresa"}
	for _, role := range allowed {
		if !empresaAIAdministrativeDBReadRole(role) {
			t.Fatalf("role %q debe tener lectura administrativa por empresa", role)
		}
	}
	denied := []string{"cajero", "vendedor", "contador", "inventario", "responsable_bodega", ""}
	for _, role := range denied {
		if empresaAIAdministrativeDBReadRole(role) {
			t.Fatalf("role %q no debe tener lectura total administrativa por IA", role)
		}
	}
}

func TestFormatEmpresaAIUsuariosResumen(t *testing.T) {
	resp := formatEmpresaAIUsuariosResumen(12, "admin_empresa", []dbpkg.EmpresaUsuario{
		{EmpresaID: 12, RolNombre: "administrador", Estado: "activo", EmailConfirmado: 1, PasswordSet: 1},
		{EmpresaID: 12, RolNombre: "cajero", Estado: "activo", EmailConfirmado: 0, PasswordSet: 0},
		{EmpresaID: 12, RolNombre: "cajero", Estado: "inactivo", EmailConfirmado: 1, PasswordSet: 1},
	})
	for _, needle := range []string{
		"**3 usuarios registrados**",
		"Activos: 2",
		"Inactivos u otros estados: 1",
		"Pendientes de confirmar correo: 1",
		"admin_empresa: 1",
		"cajero: 2",
		"empresa_id",
	} {
		if !strings.Contains(resp, needle) {
			t.Fatalf("respuesta no contiene %q:\n%s", needle, resp)
		}
	}
}
