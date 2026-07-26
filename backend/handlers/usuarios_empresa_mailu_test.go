package handlers

import "testing"

func TestEmpresaUsuarioMailuSMTPAddress(t *testing.T) {
	t.Setenv("PCS_MAILU_SMTP_ADDR", "")
	if got := empresaUsuarioMailuSMTPAddress(); got != "mailu-smtp:25" {
		t.Fatalf("default SMTP address = %q", got)
	}

	t.Setenv("PCS_MAILU_SMTP_ADDR", " 192.168.203.3:25 ")
	if got := empresaUsuarioMailuSMTPAddress(); got != "192.168.203.3:25" {
		t.Fatalf("configured SMTP address = %q", got)
	}
}
