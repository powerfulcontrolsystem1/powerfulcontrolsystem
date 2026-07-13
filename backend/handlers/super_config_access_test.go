package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuperConfigHandlersRequireSuperAdmin(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"chat_ia":         SuperChatIALogicaConfigHandler(nil, nil),
		"config_backup":   SuperConfigBackupHandler(nil),
		"domotica":        SuperDomoticaStorageHandler(nil, nil),
		"email_templates": SuperEmailTemplatesHandler(nil),
		"limitaciones":    SuperEmpresaLimitacionesConfigHandler(nil),
		"mantenimiento":   SuperMantenimientoConfigHandler(nil),
		"recordatorios":   SuperRecordatoriosInfraestructuraHandler(nil),
	}
	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/super/api/config", nil)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestWithSuperAuditoriaRequiresSuperAdmin(t *testing.T) {
	called := false
	handler := WithSuperAuditoria(nil, "prueba", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/super/api/protegido", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
	if called {
		t.Fatal("wrapped handler ran without an authenticated super administrator")
	}
}

func TestUnwrappedSuperHandlersRequireSuperAdmin(t *testing.T) {
	handlers := map[string]http.HandlerFunc{
		"plantillas_nuevas":      SuperPlantillasNuevosCatalogoHandler(nil),
		"plantillas_integracion": SuperPlantillasIntegracionCatalogoHandler(nil),
		"roles":                  RolesDeUsuarioHandler(nil),
		"roles_permisos":         RolesDeUsuarioPermisosHandler(nil),
		"licencias_adicionales":  EmpresaLicenciasAdicionalesHandler(nil),
		"venta_digital":          SuperVentaDigitalHandler(nil),
		"metrics_current":        MetricsCurrentHandler(nil),
		"metrics_history":        MetricsHistoryHandler(nil),
	}
	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/super/api/protegido", nil)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
			}
		})
	}
}
