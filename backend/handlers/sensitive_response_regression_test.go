package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestSensitiveAuthAndAdminResponsesDoNotExposeProviderOrDatabaseErrors(t *testing.T) {
	tests := []struct {
		path    string
		blocked []string
	}{
		{
			path:    "auth_admin_handlers.go",
			blocked: []string{`"error":      err.Error()`, `"error": err.Error()`},
		},
		{
			path: "administradores_frecuencia_fe_handlers.go",
			blocked: []string{
				`"failed to load: "+err.Error()`,
				`"invalid payload: "+err.Error()`,
				`"failed to save: "+err.Error()`,
			},
		},
	}
	for _, test := range tests {
		contents, err := os.ReadFile(test.path)
		if err != nil {
			t.Fatalf("read %s: %v", test.path, err)
		}
		for _, blocked := range test.blocked {
			if strings.Contains(string(contents), blocked) {
				t.Fatalf("%s still exposes an internal error pattern %q", test.path, blocked)
			}
		}
	}
}
