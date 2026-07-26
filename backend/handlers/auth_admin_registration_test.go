package handlers

import (
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestPendingAdminScopedInvitationRequiresCreator(t *testing.T) {
	selfRegistration := &dbpkg.Admin{
		EmailConfirmado:   0,
		EmailConfirmToken: "hashed-email-confirmation-token",
	}
	if isPendingAdminScopedInvitation(selfRegistration) {
		t.Fatal("self registration must remain eligible for a confirmation email retry")
	}

	scopedInvitation := &dbpkg.Admin{
		EmailConfirmado:   0,
		EmailConfirmToken: "hashed-invitation-token",
		UsuarioCreador:    "principal@example.test",
	}
	if !isPendingAdminScopedInvitation(scopedInvitation) {
		t.Fatal("scoped administrator invitation must require its invitation token")
	}

	scopedInvitation.EmailConfirmado = 1
	if isPendingAdminScopedInvitation(scopedInvitation) {
		t.Fatal("confirmed administrator is no longer a pending invitation")
	}
}
