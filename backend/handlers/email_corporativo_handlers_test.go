package handlers

import (
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
