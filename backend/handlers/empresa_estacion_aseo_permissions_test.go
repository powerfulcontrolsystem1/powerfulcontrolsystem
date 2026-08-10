package handlers

import "testing"

func TestEmpresaAseoRoleCanManage(t *testing.T) {
	allowed := []string{"super_administrador", "administrador_total", "admin_empresa", "supervisor_sucursal", "auditor"}
	for _, role := range allowed {
		if !empresaAseoRoleCanManage(role) {
			t.Fatalf("el rol administrativo %q debe poder consultar el reporte", role)
		}
	}
	for _, role := range []string{"cajero", "servicio_limpieza", "vendedor", ""} {
		if empresaAseoRoleCanManage(role) {
			t.Fatalf("el rol operativo %q no debe administrar el reporte", role)
		}
	}
}
