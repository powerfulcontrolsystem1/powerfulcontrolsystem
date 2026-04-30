package db

import (
	"strings"
	"testing"
)

func TestNormalizeEmpresaSensorDeviceID(t *testing.T) {
	got := NormalizeEmpresaSensorDeviceID(" RPI Mesa #1 / Puerta Principal ")
	if got != "rpi-mesa-1-puerta-principal" {
		t.Fatalf("device_id normalizado = %q", got)
	}
	if NormalizeEmpresaSensorDeviceID(" !!! ") != "" {
		t.Fatalf("device_id invalido debe quedar vacio")
	}
}

func TestGenerateEmpresaSensorToken(t *testing.T) {
	token, err := GenerateEmpresaSensorToken()
	if err != nil {
		t.Fatalf("GenerateEmpresaSensorToken: %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("token debe tener 64 caracteres hex, got %d", len(token))
	}
	if strings.Trim(token, "0123456789abcdef") != "" {
		t.Fatalf("token no parece hexadecimal: %q", token)
	}
}

func TestGenerateEmpresaSensorDeviceID(t *testing.T) {
	id, err := GenerateEmpresaSensorDeviceID(25, 3)
	if err != nil {
		t.Fatalf("GenerateEmpresaSensorDeviceID: %v", err)
	}
	if !strings.HasPrefix(id, "rpi-e3-25-") {
		t.Fatalf("device_id generado inesperado: %q", id)
	}
	if _, err := GenerateEmpresaSensorDeviceID(0, 1); err == nil {
		t.Fatalf("empresa_id invalido debe fallar")
	}
}
