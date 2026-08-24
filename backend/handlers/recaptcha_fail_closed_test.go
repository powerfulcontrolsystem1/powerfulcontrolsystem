package handlers

import "testing"

func TestRecaptchaDevBypassCannotEnableInProduction(t *testing.T) {
	t.Setenv("PCS_ENV", "production")
	t.Setenv("RECAPTCHA_DEV_BYPASS", "1")
	if recaptchaDevBypassEnabled() {
		t.Fatal("production must ignore the reCAPTCHA development bypass")
	}
}

func TestRecaptchaDevBypassRemainsAvailableForIsolatedDevelopment(t *testing.T) {
	t.Setenv("PCS_ENV", "development")
	t.Setenv("APP_ENV", "development")
	t.Setenv("RECAPTCHA_DEV_BYPASS", "1")
	if !recaptchaDevBypassEnabled() {
		t.Fatal("isolated development may explicitly enable the reCAPTCHA bypass")
	}
}
