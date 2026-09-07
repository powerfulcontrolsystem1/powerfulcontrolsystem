package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClientIPConsumersRejectSpoofing(t *testing.T) {
	t.Setenv("PCS_TRUSTED_PROXY_CIDRS", "127.0.0.1/32")
	consumers := map[string]func(*http.Request) string{
		"public chat":     portalChatClientIP,
		"public messages": publicClientIP,
		"certificates":    clientIPForCertificados,
		"audit":           resolveAuditoriaIP,
		"device tunnel":   domoticaTunnelRemoteIP,
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

func TestPublicChatCookieRotationCannotResetIPBudget(t *testing.T) {
	l := &publicPortalChatLimiter{}
	for i := 0; i < 30; i++ {
		if ok, _, _ := l.allowClient(fmt.Sprint(i), fmt.Sprint(i), "198.51.100.7"); !ok {
			t.Fatalf("request %d rejected before budget", i)
		}
	}
	if ok, _, retry := l.allowClient("new-store", "new-cookie", "198.51.100.7"); ok || retry < 1 {
		t.Fatal("rotated cookie bypassed IP limit")
	}
}

func TestPublicChatLimiterBoundsMemoryAndReclaimsExpiredBuckets(t *testing.T) {
	l := &publicPortalChatLimiter{records: map[string]*publicPortalChatUsage{}}
	for i := 0; i < publicPortalChatMaxBuckets; i++ {
		l.records[fmt.Sprint(i)] = &publicPortalChatUsage{ResetAt: time.Now().Add(time.Minute)}
	}
	if ok, _, _ := l.allow("new", 10, time.Minute); ok {
		t.Fatal("saturated limiter admitted a new bucket")
	}
	l.records["0"].ResetAt = time.Now().Add(-time.Second)
	l.nextCleanup = time.Time{}
	if ok, _, _ := l.allow("new", 10, time.Minute); !ok || len(l.records) != publicPortalChatMaxBuckets {
		t.Fatal("expired bucket was not reclaimed")
	}
}
