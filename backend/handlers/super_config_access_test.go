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
