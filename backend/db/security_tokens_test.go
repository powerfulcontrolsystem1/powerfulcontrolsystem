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

func TestAdministradorEmailConfirmationTokenUsesOnlyVerifier(t *testing.T) {
	token := "confirmacion-temporal"
	hash := hashOneTimeSecret(token)
	if !AdministradorEmailConfirmTokenMatches(hash, token) {
		t.Fatal("valid email confirmation token did not match verifier")
	}
	if AdministradorEmailConfirmTokenMatches(hash, "otro-token") {
		t.Fatal("incorrect email confirmation token matched verifier")
	}
}
