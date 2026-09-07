package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestDomoticaSSHTargetRequiresExplicitPrivateCIDR(t *testing.T) {
	t.Setenv("PCS_DOMOTICA_SSH_ALLOWED_CIDRS", "")
	if _, err := resolveDomoticaSSHTarget("192.168.1.16", 22); err == nil {
		t.Fatal("una IP privada no autorizada no debe ser alcanzable por SSH desde el VPS")
	}
	t.Setenv("PCS_DOMOTICA_SSH_ALLOWED_CIDRS", "192.168.1.0/24")
	if target, err := resolveDomoticaSSHTarget("192.168.1.16", 22); err != nil || target != "192.168.1.16:22" {
		t.Fatalf("CIDR VPN autorizado no aceptado: target=%q err=%v", target, err)
	}
	if _, err := resolveDomoticaSSHTarget("127.0.0.1", 22); err == nil {
		t.Fatal("loopback siempre debe rechazarse")
	}
}

func TestDomoticaSSHHandlerDoesNotSerializePasswords(t *testing.T) {
	raw, err := os.ReadFile("control_electrico_ssh_install.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, forbidden := range []string{`"password": payload.Password`, `"sudo_password": payload.SudoPassword`, "log.Printf(payload.Password"} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("handler SSH expone secreto mediante %q", forbidden)
		}
	}
	for _, required := range []string{"HostKeyFingerprint", "subtle.ConstantTimeCompare", "SaveEmpresaControlElectricoSSHCredentials", "ResolveEmpresaControlElectricoSSHCredentials"} {
		if !strings.Contains(source, required) {
			t.Fatalf("handler SSH sin control requerido %q", required)
		}
	}
}
