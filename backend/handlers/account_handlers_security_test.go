package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAccountCredentialChangesRotateSessions(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("account_handlers.go"))
	if err != nil {
		t.Fatalf("read account handlers: %v", err)
	}
	source := string(raw)
	for _, handler := range []string{
		"AccountChangePasswordHandler",
		"AccountSetGooglePasswordHandler",
	} {
		start := strings.Index(source, "func "+handler)
		if start < 0 {
			t.Fatalf("missing %s", handler)
		}
		section := source[start:]
		if !strings.Contains(section, "issueReplacementAdminSession") {
			t.Fatalf("%s must rotate the browser session after changing credentials", handler)
		}
	}
	if !strings.Contains(source, "A change of login identifier is a security event") || !strings.Contains(source, "issueReplacementAdminSession(w, r, dbSuper, newEmail)") {
		t.Fatal("email change must revoke old sessions and issue a replacement")
	}
	adminSource, err := os.ReadFile(filepath.Join("auth_admin_handlers.go"))
	if err != nil {
		t.Fatalf("read administrative handlers: %v", err)
	}
	if !strings.Contains(string(adminSource), "RevokeSessionsByAdminEmail(dbSuper, existing.Email)") {
		t.Fatal("role changes must revoke existing administrator sessions")
	}
	deleteStart := strings.Index(string(adminSource), "case http.MethodDelete:")
	if deleteStart < 0 {
		t.Fatal("administrative delete handler is missing")
	}
	deleteSection := string(adminSource)[deleteStart:]
	revokeIndex := strings.Index(deleteSection, "RevokeSessionsByAdminEmail(dbSuper, targetAdmin.Email)")
	deleteIndex := strings.Index(deleteSection, "DeleteAdministrador(dbSuper, id)")
	invalidateIndex := strings.Index(deleteSection, "InvalidateAuthCacheForAdmin(targetAdmin.Email)")
	if revokeIndex < 0 || deleteIndex < 0 || invalidateIndex < 0 || revokeIndex > deleteIndex || deleteIndex > invalidateIndex {
		t.Fatal("administrator deletion must revoke sessions before deletion and invalidate authorization cache afterwards")
	}
}
