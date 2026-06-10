package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCorporateEmailAppendThemePreservesSnappyMailSSOQuery(t *testing.T) {
	got := corporateEmailAppendThemeToURI("/webmail/index.php?sso&hash=abc123", "dark")
	if !strings.HasPrefix(got, "/webmail/index.php?sso&hash=abc123&") {
		t.Fatalf("SnappyMail SSO query must keep sso first, got %q", got)
	}
	if !strings.Contains(got, "theme=dark") {
		t.Fatalf("expected theme parameter in %q", got)
	}
	if !strings.Contains(got, "mail_theme=PCSDark%40custom") {
		t.Fatalf("expected SnappyMail theme parameter in %q", got)
	}
}

func TestCorporateEmailAppendThemeRegularURL(t *testing.T) {
	got := corporateEmailAppendThemeToURI("/webmail/?_task=mail", "light")
	if !strings.Contains(got, "_task=mail") {
		t.Fatalf("expected existing query to be preserved, got %q", got)
	}
	if !strings.Contains(got, "theme=light") {
		t.Fatalf("expected light theme parameter in %q", got)
	}
}

func TestCorporateEmailMaxAccountsDefaultAndBounds(t *testing.T) {
	if got := getCorporateEmailConfig(nil).MaxAccounts; got != corporateEmailDefaultMax {
		t.Fatalf("default max accounts per empresa = %d, want %d", got, corporateEmailDefaultMax)
	}
	cases := []struct {
		in   int
		want int
	}{
		{0, corporateEmailDefaultMax},
		{-3, corporateEmailDefaultMax},
		{5, 5},
		{12, 12},
		{501, 500},
	}
	for _, tc := range cases {
		if got := normalizeCorporateEmailMaxAccounts(tc.in); got != tc.want {
			t.Fatalf("normalizeCorporateEmailMaxAccounts(%d)=%d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestCorporateEmailParseStatusLineUnread(t *testing.T) {
	status := parseCorporateEmailStatusLine(`* STATUS INBOX (MESSAGES 14 RECENT 1 UNSEEN 3)`)
	if !status.Checked || !status.OK {
		t.Fatalf("expected checked OK status, got %+v", status)
	}
	if status.Messages != 14 || status.Recent != 1 || status.Unseen != 3 {
		t.Fatalf("unexpected IMAP counters: %+v", status)
	}
}

func TestCorporateEmailSuperPageIncludesMaxAccountsField(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "web", "super", "email_corporativo.html"))
	if err != nil {
		t.Fatalf("read email_corporativo.html: %v", err)
	}
	html := string(raw)
	required := []string{
		`id="maxAccountsPerEmpresa"`,
		`config.max_accounts_per_empresa || 5`,
		`max_accounts_per_empresa: Number(fields.maxAccountsPerEmpresa.value || 5)`,
	}
	for _, expected := range required {
		if !strings.Contains(html, expected) {
			t.Fatalf("email_corporativo.html debe exponer y guardar el cupo de cuentas por empresa; falta %q", expected)
		}
	}
}
