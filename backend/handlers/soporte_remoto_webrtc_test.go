package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func signalingRequest(target string, empresaID int64) *http.Request {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Header.Set("Origin", "https://example.com")
	return req.WithContext(context.WithValue(req.Context(), "empresaID", empresaID))
}

func TestSoporteRemotoSignalingRejectsMissingOrigin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/api/public/webrtc/signaling?empresa_id=12", nil)
	res := httptest.NewRecorder()
	SoporteRemotoSignalingHandler(nil).ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for missing origin, got %d", res.Code)
	}
}

func TestSoporteRemotoSignalingRejectsCrossTenantQuery(t *testing.T) {
	req := signalingRequest("https://example.com/api/public/webrtc/signaling?empresa_id=13&codigo_sesion=x&role=host&token=x&nonce=x", 12)
	res := httptest.NewRecorder()
	SoporteRemotoSignalingHandler(nil).ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for cross-tenant query, got %d", res.Code)
	}
}

func TestSoporteRemotoSignalingRequiresCompleteCredential(t *testing.T) {
	req := signalingRequest("https://example.com/api/public/webrtc/signaling?empresa_id=12&codigo_sesion=x&role=host", 12)
	res := httptest.NewRecorder()
	SoporteRemotoSignalingHandler(nil).ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for incomplete credential, got %d", res.Code)
	}
}

func TestSoporteRemotoSignalingKeysAreTenantScoped(t *testing.T) {
	keyA := soporteRemotoSignalingPeerKey(12, "session", "host")
	keyB := soporteRemotoSignalingPeerKey(13, "session", "host")
	if keyA == keyB {
		t.Fatal("peer key does not isolate company")
	}
}

func TestSoporteRemotoSignalingAllowsOnePeerPerRole(t *testing.T) {
	key := soporteRemotoSignalingPeerKey(987654, "unit-test-session", "viewer")
	soporteRemotoSignalingRelease(key, nil)
	if !soporteRemotoSignalingReserve(key) {
		t.Fatal("first reservation was rejected")
	}
	defer soporteRemotoSignalingRelease(key, nil)
	if soporteRemotoSignalingReserve(key) {
		t.Fatal("duplicate reservation was accepted")
	}
}
