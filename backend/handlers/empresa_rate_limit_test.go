package handlers

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestCheckEmpresaRateLimitAtBlocksWithinMinute(t *testing.T) {
	empresaRateLimitMu.Lock()
	empresaRateLimitBuckets = map[string]empresaRateLimitBucket{}
	empresaRateLimitMu.Unlock()

	now := time.Date(2026, 4, 30, 10, 15, 20, 0, time.UTC)
	for i := 0; i < 3; i++ {
		allowed, remaining, retryAfter, current := checkEmpresaRateLimitAt(now, 77, "api", 3)
		if !allowed {
			t.Fatalf("solicitud %d bloqueada antes del limite", i+1)
		}
		if current != int64(i+1) {
			t.Fatalf("contador inesperado: got %d want %d", current, i+1)
		}
		if retryAfter != 0 {
			t.Fatalf("retry-after inesperado antes del bloqueo: %d", retryAfter)
		}
		if remaining != int64(2-i) {
			t.Fatalf("remaining inesperado: got %d want %d", remaining, 2-i)
		}
	}

	allowed, remaining, retryAfter, current := checkEmpresaRateLimitAt(now, 77, "api", 3)
	if allowed {
		t.Fatal("debe bloquear cuando se excede el limite por minuto")
	}
	if remaining != 0 {
		t.Fatalf("remaining al bloquear debe ser 0, got %d", remaining)
	}
	if retryAfter <= 0 {
		t.Fatalf("retry-after debe ser positivo, got %d", retryAfter)
	}
	if current != 3 {
		t.Fatalf("contador bloqueado inesperado: got %d want 3", current)
	}
}

func TestCheckEmpresaRateLimitAtResetsNextMinute(t *testing.T) {
	empresaRateLimitMu.Lock()
	empresaRateLimitBuckets = map[string]empresaRateLimitBucket{}
	empresaRateLimitMu.Unlock()

	now := time.Date(2026, 4, 30, 10, 15, 58, 0, time.UTC)
	if allowed, _, _, _ := checkEmpresaRateLimitAt(now, 88, "db_admin", 1); !allowed {
		t.Fatal("primera consulta debe ser permitida")
	}
	if allowed, _, _, _ := checkEmpresaRateLimitAt(now, 88, "db_admin", 1); allowed {
		t.Fatal("segunda consulta en la misma ventana debe bloquearse")
	}
	if allowed, remaining, _, current := checkEmpresaRateLimitAt(now.Add(3*time.Second), 88, "db_admin", 1); !allowed || remaining != 0 || current != 1 {
		t.Fatalf("la ventana siguiente debe reiniciar el contador: allowed=%v remaining=%d current=%d", allowed, remaining, current)
	}
}

func TestEmpresaRateLimitScopeForRequest(t *testing.T) {
	req := &http.Request{URL: &url.URL{Path: "/api/empresa/db_admin"}}
	if got := empresaRateLimitScopeForRequest(req); got != "db_admin" {
		t.Fatalf("scope db_admin inesperado: %s", got)
	}
	req.URL.Path = "/api/empresa/productos"
	if got := empresaRateLimitScopeForRequest(req); got != "api" {
		t.Fatalf("scope api inesperado: %s", got)
	}
}

func TestEmpresaRateLimitMaxForRequestUsesDefaultsWithoutDB(t *testing.T) {
	req := &http.Request{URL: &url.URL{Path: "/api/empresa/db_admin"}}
	if got := empresaRateLimitMaxForRequest(nil, req); got != defaultEmpresaDBQueriesPerMinute {
		t.Fatalf("default db_admin inesperado: got %d want %d", got, defaultEmpresaDBQueriesPerMinute)
	}
	req.URL.Path = "/api/empresa/productos"
	if got := empresaRateLimitMaxForRequest(nil, req); got != defaultEmpresaAPIRequestsPerMinute {
		t.Fatalf("default api inesperado: got %d want %d", got, defaultEmpresaAPIRequestsPerMinute)
	}
}
