package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPermissionChangeRequiresApprovalForSecurityEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		action string
		want   bool
	}{
		{
			name:   "crear usuario operativo no requiere codigo de aprobacion",
			method: http.MethodPost,
			path:   "/api/empresa/usuarios",
			action: permActionCreate,
			want:   false,
		},
		{
			name:   "actualizar usuario operativo no requiere codigo de aprobacion",
			method: http.MethodPut,
			path:   "/api/empresa/usuarios?id=10",
			action: permActionUpdate,
			want:   false,
		},
		{
			name:   "reenviar confirmacion de usuario no requiere codigo de aprobacion",
			method: http.MethodPut,
			path:   "/api/empresa/usuarios?action=reenviar_confirmacion&id=10",
			action: permActionApprove,
			want:   false,
		},
		{
			name:   "cambiar roles de usuario requiere codigo de aprobacion",
			method: http.MethodPost,
			path:   "/api/empresa/roles_de_usuario",
			action: permActionCreate,
			want:   true,
		},
		{
			name:   "cambiar permisos finos requiere codigo de aprobacion",
			method: http.MethodPut,
			path:   "/api/empresa/permisos_empresa",
			action: permActionUpdate,
			want:   true,
		},
		{
			name:   "leer permisos finos no requiere codigo de aprobacion",
			method: http.MethodGet,
			path:   "/api/empresa/permisos_empresa?empresa_id=44",
			action: permActionRead,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			got := permissionChangeRequiresApproval(permModuleSeguridad, req, tt.action)
			if got != tt.want {
				t.Fatalf("permissionChangeRequiresApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}
