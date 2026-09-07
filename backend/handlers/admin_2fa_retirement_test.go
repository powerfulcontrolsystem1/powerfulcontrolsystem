package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAdministrator2FAHasNoRuntimeOrUISurface(t *testing.T) {
	t.Parallel()
	checks := map[string][]string{
		"auth_admin_handlers.go":                       {"OTPCode", "two_factor_required", "adminTOTP"},
		"recaptcha.go":                                 {"ADMIN_2FA_LOGIN_ENABLED", "isAdminTOTPLoginEnabled"},
		filepath.Join("..", "main.go"):                 {"/super/api/administradores/2fa", "/super/api/config/admin_2fa"},
		filepath.Join("..", "..", "web", "login.html"): {"adminOtpCode", "Código 2FA"},
		filepath.Join("..", "..", "web", "js", "login.js"): {
			"otp_code", "two_factor_required", "ADMIN_2FA_LOGIN_ENABLED",
		},
		filepath.Join("..", "..", "web", "super_administrador.html"): {"configuracion/login_2fa.html"},
		filepath.Join("..", "..", "web", "super", "configuracion_avanzada.html"): {
			"admin2FAConfigCard", "/super/api/config/admin_2fa",
		},
	}
	for path, forbidden := range checks {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source := string(raw)
		for _, marker := range forbidden {
			if strings.Contains(source, marker) {
				t.Fatalf("%s still exposes retired 2FA marker %q", path, marker)
			}
		}
	}
	removedPage := filepath.Join("..", "..", "web", "super", "configuracion", "login_2fa.html")
	if _, err := os.Stat(removedPage); !os.IsNotExist(err) {
		t.Fatalf("retired 2FA page must not exist: %v", err)
	}
}
