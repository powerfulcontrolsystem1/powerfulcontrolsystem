package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestSelectedHTTPFlowsDoNotBootstrapSchema(t *testing.T) {
	tests := []struct {
		file      string
		forbidden string
		required  string
	}{
		{"nextcloud.go", "dbpkg.EnsureEmpresaNextcloudSchema(", "dbpkg.EmpresaNextcloudSchemaReady("},
		{"super_correos_masivos.go", "dbpkg.EnsureEmpresaUsuariosAuthSchema(", "dbpkg.EmpresaUsuariosAuthSchemaReady("},
		{"creditos.go", "dbpkg.EnsureEmpresaEventosContablesSchema(", "dbpkg.EmpresaEventosContablesSchemaReady("},
		{"modulos_faltantes.go", "dbpkg.EnsureEmpresaEventosContablesSchema(", "dbpkg.EmpresaEventosContablesSchemaReady("},
		{"empresa_permisos.go", "dbpkg.EnsureEmpresaPermisosFinosSchema(", "dbpkg.EmpresaPermisosFinosSchemaReady("},
		{"rappi.go", "dbpkg.EnsureEmpresaRappiSchema(", "dbpkg.EmpresaRappiSchemaReady("},
		{"roles_tipos_usuario.go", "dbpkg.EnsureRolesPermisosSchema(", "dbpkg.RolesPermisosSchemaReady("},
	}
	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			body, err := os.ReadFile(test.file)
			if err != nil {
				t.Fatalf("read %s: %v", test.file, err)
			}
			source := string(body)
			if strings.Contains(source, test.forbidden) {
				t.Fatalf("HTTP flow must not bootstrap schema through %s", test.forbidden)
			}
			if !strings.Contains(source, test.required) {
				t.Fatalf("HTTP flow must fail closed through %s", test.required)
			}
		})
	}
}
