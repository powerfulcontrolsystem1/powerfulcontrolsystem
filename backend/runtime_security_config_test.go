package main

import "testing"

func setValidProductionSecurityEnv(t *testing.T) {
	t.Helper()
	t.Setenv("PCS_ENV", "production")
	t.Setenv("PCS_TRUSTED_PROXY_CIDRS", "127.0.0.1/32")
	t.Setenv("CONFIG_ENC_KEY_ID", "key-current")
	t.Setenv("PCS_CSRF_ALLOWED_ORIGINS", "https://example.com")
	t.Setenv("SESSION_TIMEOUT", "12h")
	t.Setenv("MAX_REQUEST_BODY_BYTES", "67108864")
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", t.TempDir())
	t.Setenv("PCS_API_REPLICAS", "1")
	t.Setenv("PCS_PRIVATE_STORAGE_MODE", "local")
	t.Setenv("HTTP_READ_TIMEOUT", "30s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "60s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "2m")
}

func TestValidateProductionSecurityConfigRejectsLocalStorageWithMultipleReplicas(t *testing.T) {
	setValidProductionSecurityEnv(t)
	t.Setenv("PCS_API_REPLICAS", "2")
	t.Setenv("PCS_PRIVATE_STORAGE_MODE", "local")
	if err := validateProductionSecurityConfig(); err == nil {
		t.Fatal("expected local private storage to be rejected with multiple API replicas")
	}
	t.Setenv("PCS_PRIVATE_STORAGE_MODE", "shared")
	if err := validateProductionSecurityConfig(); err != nil {
		t.Fatalf("shared private storage rejected: %v", err)
	}
}

func TestValidateProductionSecurityConfigAcceptsCompleteConfig(t *testing.T) {
	setValidProductionSecurityEnv(t)
	if err := validateProductionSecurityConfig(); err != nil {
		t.Fatalf("valid production config rejected: %v", err)
	}
}

func TestValidateProductionSecurityConfigRejectsMissingCriticalValue(t *testing.T) {
	setValidProductionSecurityEnv(t)
	t.Setenv("HTTP_WRITE_TIMEOUT", "")
	if err := validateProductionSecurityConfig(); err == nil {
		t.Fatal("missing critical value was accepted")
	}
}

func TestValidateProductionSecurityConfigRejectsPartialOrigin(t *testing.T) {
	setValidProductionSecurityEnv(t)
	t.Setenv("PCS_CSRF_ALLOWED_ORIGINS", "https://example.com.attacker.invalid")
	if err := validateProductionSecurityConfig(); err != nil {
		t.Fatalf("a syntactically valid exact origin should be accepted: %v", err)
	}
	t.Setenv("PCS_CSRF_ALLOWED_ORIGINS", "example.com")
	if err := validateProductionSecurityConfig(); err == nil {
		t.Fatal("origin without scheme was accepted")
	}
}
