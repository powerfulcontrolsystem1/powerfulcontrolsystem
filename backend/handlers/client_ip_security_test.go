package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientIPConsumersRejectSpoofing(t *testing.T) {
	t.Setenv("PCS_TRUSTED_PROXY_CIDRS", "127.0.0.1/32")
	consumers := map[string]func(*http.Request) string{
		"public chat": portalChatClientIP,
		"public messages": publicClientIP,
		"certificates": clientIPForCertificados,
		"audit": resolveAuditoriaIP,
		"device tunnel": domoticaTunnelRemoteIP,
	}
	for name, consumer := range consumers {
		t.Run(name, func(t *testing.T) {
			for _, peer := range []string{"198.51.100.7:4567", "127.0.0.1:8080"} {
				r := httptest.NewRequest(http.MethodGet, "https://service.test/", nil)
				r.RemoteAddr = peer
				r.Header.Set("X-Forwarded-For", "203.0.113.8, 198.51.100.7")
				r.Header.Set("X-Real-IP", "203.0.113.9")
				if got := consumer(r); got != "198.51.100.7" {
					t.Fatalf("peer %q resolved to forged IP %q", peer, got)
				}
			}
		})
	}
}
