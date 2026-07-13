package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedactEmailForLog(t *testing.T) {
	tests := map[string]string{
		"ana@example.com":   "a***@example.com",
		"a@example.com":     "a***@example.com",
		" invalid-address ": "[redacted]",
		"":                  "[redacted]",
	}
	for input, want := range tests {
		if got := redactEmailForLog(input); got != want {
			t.Fatalf("redactEmailForLog(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSensitiveLogValuesAreRedacted(t *testing.T) {
	if got := redactPersonalDocumentForLog("1.234.567.890"); got != "1***90" {
		t.Fatalf("document redaction = %q", got)
	}
	if got := redactPhoneForLog("+57 300 123 4567"); got != "5***67" {
		t.Fatalf("phone redaction = %q", got)
	}
	if got := redactTokenForLog("very-secret-token"); got != "[redacted-token]" {
		t.Fatalf("token redaction = %q", got)
	}
	if got := redactAuthorizationForLog("Bearer very-secret-token"); got != "[authorization-present]" {
		t.Fatalf("authorization redaction = %q", got)
	}
	if got := normalizeLogValue("line1\r\nline2"); got != "line1\\r\\nline2" {
		t.Fatalf("log normalization = %q", got)
	}
}

func TestSensitiveRequestHeadersAreRedacted(t *testing.T) {
	headers := http.Header{
		"Authorization": {"Bearer secret"},
		"Cookie":        {"session=secret"},
		"X-Request-ID":  {"req-123"},
	}
	redacted := redactRequestHeadersForLog(headers)
	for key, value := range redacted {
		if strings.Contains(value, "secret") {
			t.Fatalf("header %s leaked secret: %q", key, value)
		}
	}
	if got := redacted["x-request-id"]; got != "req-123" {
		t.Fatalf("request id = %q", got)
	}
}

func TestLegacyEpaycoRuntimeToolDoesNotReadOrPrintSensitiveData(t *testing.T) {
	_, err := os.Stat(filepath.Join("..", "tools", "query_epayco_runtime.go"))
	if !os.IsNotExist(err) {
		t.Fatalf("legacy sensitive payment runtime tool must remain removed, stat err=%v", err)
	}
}

func TestLegacyAdminLoginInspectionToolRemainsRemoved(t *testing.T) {
	_, err := os.Stat(filepath.Join("..", "tools", "inspect_admin_login", "main.go"))
	if !os.IsNotExist(err) {
		t.Fatalf("legacy admin login inspection tool must remain removed, stat err=%v", err)
	}
}

func TestLegacyPlaintextRustDeskDeviceImplementationRemainsRemoved(t *testing.T) {
	for _, path := range []string{
		filepath.Join("rustdesk.go"),
		filepath.Join("..", "db", "rustdesk.go"),
	} {
		_, err := os.Stat(path)
		if !os.IsNotExist(err) {
			t.Fatalf("legacy plaintext RustDesk implementation must remain removed: %s, stat err=%v", path, err)
		}
	}
}

func TestStationPreferencesDoNotLogRawUserEmail(t *testing.T) {
	path := filepath.Join("estaciones_columnas_pref.go")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if strings.Contains(string(contents), "user=%s error: %v\", empresaID, p.UsuarioEmail") {
		t.Fatal("station preferences must redact user email before logging")
	}
}

func TestProviderMailFailuresDoNotLogProviderErrors(t *testing.T) {
	for path, forbidden := range map[string]string{
		"payments_handlers.go":           "email:\", mailErr",
		"email_corporativo_handlers.go":  "provision warning: %s",
		"super_mantenimiento_agentes.go": "enviar correo DIAN: %v",
		"system_empresas_handlers.go":    "email corporativo warning: %v",
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if strings.Contains(string(contents), forbidden) {
			t.Fatalf("provider mail failure must not log raw provider error: %s", path)
		}
	}
}

func TestRemoteSupportConfigurationDoesNotExposePersistenceErrors(t *testing.T) {
	contents, err := os.ReadFile("super_soporte_remoto.go")
	if err != nil {
		t.Fatalf("read remote support handler: %v", err)
	}
	if strings.Contains(string(contents), "No se pudo guardar configuracion de soporte remoto: \"+err.Error()") {
		t.Fatal("remote support configuration must not expose persistence errors")
	}
}

func TestCompanyConfigurationHandlersDoNotExposePersistenceErrors(t *testing.T) {
	for path, forbidden := range map[string][]string{
		"configuracion_guiada.go": {
			"no se pudo cargar la configuracion guiada: \"+err.Error()",
			"no se pudo cargar el contexto guiado: \"+err.Error()",
			"no se pudo posponer la configuracion guiada: \"+err.Error()",
			"no se pudo aplicar la configuracion guiada: \"+err.Error()",
		},
		"panel_empresa_config.go": {
			"no se pudo guardar favoritos: \"+err.Error()",
			"no se pudo guardar email corporativo: \"+err.Error()",
			"no se pudo guardar noticias: \"+err.Error()",
			"no se pudo guardar buzon: \"+err.Error()",
			"no se pudo guardar chat: \"+err.Error()",
		},
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, value := range forbidden {
			if strings.Contains(string(contents), value) {
				t.Fatalf("company configuration handler must not expose persistence error: %s", path)
			}
		}
	}
}

func TestAIConfigurationDoesNotExposeCredentialOrProviderDiagnostics(t *testing.T) {
	contents, err := os.ReadFile("ai_config_handlers.go")
	if err != nil {
		t.Fatalf("read AI configuration handler: %v", err)
	}
	for _, forbidden := range []string{
		`"error":          err.Error()`,
		`publicErr = err.Error()`,
		`publicErr = perr.Error()`,
		`"failed to save "+superAIEnabledConfigKey+": "+err.Error()`,
		`"encryption failed: "+err.Error()`,
		`"failed to save provider key "+providerKey+": "+err.Error()`,
	} {
		if strings.Contains(string(contents), forbidden) {
			t.Fatalf("AI configuration must not expose internal or provider diagnostic: %s", forbidden)
		}
	}
}

func TestCorporateEmailDoesNotExposeProvisionOrAutologinDiagnostics(t *testing.T) {
	contents, err := os.ReadFile("email_corporativo_handlers.go")
	if err != nil {
		t.Fatalf("read corporate email handler: %v", err)
	}
	for _, forbidden := range []string{
		`"Autologin no disponible: " + err.Error()`,
		`"No se pudo guardar configuracion: "+err.Error()`,
		`"El buzon corporativo todavia no pudo provisionarse en Mailu: "+result.Error`,
		`"No se pudo iniciar sesion automaticamente en la bandeja de correo. "+redirectErr.Error()`,
		`"No se pudo iniciar sesion automaticamente en la bandeja de correo. "+err.Error()`,
	} {
		if strings.Contains(string(contents), forbidden) {
			t.Fatalf("corporate email must not expose provision or autologin diagnostic: %s", forbidden)
		}
	}
}

func TestPrivilegedIntegrationConfigurationsDoNotExposePersistenceErrors(t *testing.T) {
	for path, forbidden := range map[string][]string{
		"recaptcha.go": {
			`"failed to save "+superRecaptchaEnabledConfigKey+": "+err.Error()`,
			`"failed to encrypt "+superRecaptchaSecretKeyConfigKey+": "+err.Error()`,
		},
		"onlyoffice_super_config.go": {
			`"No se pudo guardar onlyoffice.enabled: "+err.Error()`,
			`"No se pudo cifrar jwt_secret: "+err.Error()`,
			`"No se pudo guardar onlyoffice.jwt_secret: "+err.Error()`,
		},
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, value := range forbidden {
			if strings.Contains(string(contents), value) {
				t.Fatalf("privileged integration configuration must not expose persistence error: %s", path)
			}
		}
	}
}

func TestPaymentConfigurationsDoNotExposeSecretsOrPersistenceErrors(t *testing.T) {
	contents, err := os.ReadFile("payments_handlers.go")
	if err != nil {
		t.Fatalf("read payments handler: %v", err)
	}
	for _, forbidden := range []string{
		`"failed to save wompi.public_key: "+err.Error()`,
		`"failed to save wompi.private_key: "+err.Error()`,
		`"failed to save epayco.private_key: "+err.Error()`,
		`"failed to encrypt epayco.checkout_key: "+err.Error()`,
	} {
		if strings.Contains(string(contents), forbidden) {
			t.Fatalf("payment configuration must not expose secret or persistence diagnostic: %s", forbidden)
		}
	}
}
