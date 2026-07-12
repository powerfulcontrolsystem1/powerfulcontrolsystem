package db

import "testing"

func TestAdministradorPasswordResetTokenUsesOnlyVerifier(t *testing.T) {
	token := "token-de-prueba-no-persistible"
	hash := hashOneTimeSecret(token)
	if hash == token || !isSHA256Hex(hash) {
		t.Fatal("reset token was not converted to a SHA-256 verifier")
	}
	if !AdministradorPasswordResetTokenMatches(hash, token) {
		t.Fatal("valid reset token did not match verifier")
	}
	if AdministradorPasswordResetTokenMatches(hash, token+"x") {
		t.Fatal("incorrect reset token matched verifier")
	}
}
