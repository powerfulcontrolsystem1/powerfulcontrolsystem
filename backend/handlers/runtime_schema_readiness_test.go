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
		{"super_alertas.go", "dbpkg.EnsureSuperAlertasSchema(", "dbpkg.SuperAlertasSchemaReady("},
		{"super_mantenimiento_agentes.go", "dbpkg.EnsureSuperMantenimientoAgentesSchema(", "dbpkg.SuperMantenimientoAgentesSchemaReady("},
		{"empresa_plantillas_nuevas.go", "dbpkg.EnsureNuevasPlantillasProduccionMasivaLicencias(", "dbpkg.ProvisionNuevasPlantillasProduccionMasivaLicencias("},
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

func TestEmpresaCorporateEmailGETDoesNotProvisionAccount(t *testing.T) {
	body, err := os.ReadFile("email_corporativo_handlers.go")
	if err != nil {
		t.Fatalf("read corporate email handler: %v", err)
	}
	source := string(body)
	start := strings.Index(source, "account, err := dbpkg.GetEmpresaEmailCorporativoByEmpresa(dbSuper, empresaID)")
	end := strings.Index(source, "func EmpresaEmailCorporativoAutologinHandler")
	if start < 0 || end <= start {
		t.Fatal("corporate email GET section not found")
	}
	section := source[start:end]
	for _, forbidden := range []string{
		"ProvisionEmpresaCorporateEmailAfterCreate(",
		"SyncEmpresaEmailRowsForExistingEmpresas(",
		"UpsertEmpresaEmailCorporativo(",
	} {
		if strings.Contains(section, forbidden) {
			t.Fatalf("corporate email GET must remain read-only; found %s", forbidden)
		}
	}
}
