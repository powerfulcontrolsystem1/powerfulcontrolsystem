package handlers

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestValidateRecaptchaActivationRejectsIncompleteOrUnreadableSecrets(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		siteKey   string
		secretKey string
		readErr   error
		wantError bool
	}{
		{name: "disabled permits incomplete configuration", enabled: false},
		{name: "enabled accepts complete configuration", enabled: true, siteKey: "site", secretKey: "private"},
		{name: "enabled rejects missing site key", enabled: true, secretKey: "private", wantError: true},
		{name: "enabled rejects missing private key", enabled: true, siteKey: "site", wantError: true},
		{name: "enabled rejects unreadable stored private key", enabled: true, siteKey: "site", readErr: errors.New("cipher mismatch"), wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateRecaptchaActivation(test.enabled, test.siteKey, test.secretKey, test.readErr)
			if (err != nil) != test.wantError {
				t.Fatalf("validateRecaptchaActivation() error=%v wantError=%v", err, test.wantError)
			}
		})
	}
}

func TestRecaptchaConfigurationUsesAtomicWritesAndVisibleActivationGuard(t *testing.T) {
	backendSource, err := os.ReadFile("recaptcha.go")
	if err != nil {
		t.Fatal(err)
	}
	backend := string(backendSource)
	for _, required := range []string{
		"validateRecaptchaActivation(enabled, siteKey, secretKey, secretReadErr)",
		"dbpkg.SetConfigValuesAtomic(dbSuper, entries)",
		"http.StatusUnprocessableEntity",
	} {
		if !strings.Contains(backend, required) {
			t.Fatalf("recaptcha backend is missing activation guard contract %q", required)
		}
	}

	frontendSource, err := os.ReadFile("../../web/super/configuracion_avanzada.html")
	if err != nil {
		t.Fatal(err)
	}
	frontend := string(frontendSource)
	for _, required := range []string{
		"recaptchaUnreadableStoredSecret",
		"no puede descifrarse",
		"pega una nueva clave privada válida antes de activarlo",
	} {
		if !strings.Contains(frontend, required) {
			t.Fatalf("recaptcha frontend is missing visible recovery contract %q", required)
		}
	}
}
