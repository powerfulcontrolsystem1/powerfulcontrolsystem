package handlers

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestIsPublicOutboundIPRejectsNonPublicRanges(t *testing.T) {
	for _, raw := range []string{
		"127.0.0.1",
		"10.0.0.1",
		"172.16.0.1",
		"192.168.1.1",
		"169.254.169.254",
		"100.64.0.1",
		"192.0.2.1",
		"198.18.0.1",
		"198.51.100.1",
		"203.0.113.1",
		"224.0.0.1",
		"240.0.0.1",
		"::1",
		"fc00::1",
		"fe80::1",
		"2001:db8::1",
	} {
		if isPublicOutboundIP(net.ParseIP(raw)) {
			t.Fatalf("non-public address accepted: %s", raw)
		}
	}
	for _, raw := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"} {
		if !isPublicOutboundIP(net.ParseIP(raw)) {
			t.Fatalf("public address rejected: %s", raw)
		}
	}
}

func TestNormalizePublicIntegracionEndpointRejectsUnsafeTargets(t *testing.T) {
	for _, raw := range []string{
		"file:///etc/passwd",
		"ftp://example.com/file",
		"http://localhost/admin",
		"http://service.local/health",
		"http://127.0.0.1/health",
		"http://169.254.169.254/latest/meta-data",
		"https://user:password@example.com/api",
	} {
		if got := normalizePublicIntegracionEndpoint(raw); got != "" {
			t.Fatalf("unsafe endpoint accepted: input=%q normalized=%q", raw, got)
		}
	}
	if got := normalizePublicIntegracionEndpoint("api.example.com/v1#private"); got != "https://api.example.com/v1" {
		t.Fatalf("unexpected safe normalization: %q", got)
	}
}

func TestDIANAcuseEndpointMustMatchConfiguredOrigin(t *testing.T) {
	configured := "https://api.provider.example/v1/dian"
	if got, ok := dianAcuseEndpointAllowed(configured, "https://api.provider.example/v1/acuse?id=1"); !ok || got == "" {
		t.Fatalf("same-origin acuse rejected: endpoint=%q ok=%v", got, ok)
	}
	for _, candidate := range []string{
		"http://api.provider.example/v1/acuse",
		"https://other.example/v1/acuse",
		"https://api.provider.example:8443/v1/acuse",
		"http://127.0.0.1/internal",
	} {
		if got, ok := dianAcuseEndpointAllowed(configured, candidate); ok || got != "" {
			t.Fatalf("out-of-origin acuse accepted: %q", candidate)
		}
	}
}

func TestDIANConfiguredEndpointCannotBeOverriddenAcrossOrigins(t *testing.T) {
	cfg := map[string]interface{}{"url_dian": "https://vpfe.dian.gov.co/WcfDianCustomerServices.svc"}
	if got := dianConfiguredEndpoint(cfg, nil); got == "" {
		t.Fatal("configured DIAN endpoint was rejected")
	}
	if got := dianConfiguredEndpoint(cfg, map[string]interface{}{
		"url_dian": "https://vpfe.dian.gov.co/WcfDianCustomerServices.svc?singleWsdl=1",
	}); got == "" {
		t.Fatal("same-origin DIAN endpoint override was rejected")
	}
	for _, raw := range []string{
		"https://vpfe.dian.gov.co.evil.example/collect",
		"http://vpfe.dian.gov.co/WcfDianCustomerServices.svc",
		"http://127.0.0.1/internal",
	} {
		if got := dianConfiguredEndpoint(cfg, map[string]interface{}{"url_dian": raw}); got != "" {
			t.Fatalf("unsafe DIAN endpoint override accepted: %q", raw)
		}
	}
	if got := dianConfiguredEndpoint(map[string]interface{}{}, map[string]interface{}{"url_dian": "https://vpfe.dian.gov.co/service"}); got != "" {
		t.Fatal("ad-hoc DIAN endpoint was accepted without company configuration")
	}
}

func TestIsDIANOfficialEndpointRequiresExactHTTPSDomainSuffix(t *testing.T) {
	for _, raw := range []string{
		"https://dian.gov.co/service",
		"https://vpfe.dian.gov.co/WcfDianCustomerServices.svc",
	} {
		if !isDIANOfficialEndpoint(raw) {
			t.Fatalf("official DIAN endpoint rejected: %s", raw)
		}
	}
	for _, raw := range []string{
		"http://vpfe.dian.gov.co/service",
		"https://vpfe.dian.gov.co.evil.example/service",
		"https://evil-dian.gov.co.example/service",
		"http://127.0.0.1/dian.gov.co",
	} {
		if isDIANOfficialEndpoint(raw) {
			t.Fatalf("non-official DIAN endpoint accepted: %s", raw)
		}
	}
}

func TestPublicOutboundRedirectPolicyStaysOnOrigin(t *testing.T) {
	origin, _ := url.Parse("https://api.provider.example/v1/start")
	client := publicOutboundHTTPClient(time.Second, origin)
	for _, raw := range []string{
		"https://api.provider.example/v1/final",
		"https://api.provider.example/other",
	} {
		req := &http.Request{URL: mustParseURLForTest(t, raw)}
		if err := client.CheckRedirect(req, []*http.Request{{URL: origin}}); err != nil {
			t.Fatalf("same-origin redirect rejected: %s: %v", raw, err)
		}
	}
	for _, raw := range []string{
		"http://api.provider.example/v1/final",
		"https://other.example/final",
		"http://127.0.0.1/internal",
	} {
		req := &http.Request{URL: mustParseURLForTest(t, raw)}
		if err := client.CheckRedirect(req, []*http.Request{{URL: origin}}); err == nil {
			t.Fatalf("unsafe redirect accepted: %s", raw)
		}
	}
}

func TestRunIntegracionProbeDoesNotReachLoopback(t *testing.T) {
	hit := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case hit <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	status, reachable, _, _ := runIntegracionProbe(server.URL)
	if status != 0 || reachable {
		t.Fatalf("loopback probe was treated as reachable: status=%d reachable=%v", status, reachable)
	}
	select {
	case <-hit:
		t.Fatal("loopback probe reached the blocked server")
	default:
	}
	if conn, err := dialPublicOutbound(context.Background(), "tcp", "127.0.0.1:80"); err == nil || conn != nil {
		if conn != nil {
			_ = conn.Close()
		}
		t.Fatal("direct loopback dial was accepted")
	}
}

func TestFacturacionProveedorHTTPDoesNotReachLoopback(t *testing.T) {
	hit := make(chan struct{}, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case hit <- struct{}{}:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dispatch := dispatchFacturacionProveedorHTTP(server.URL, map[string]interface{}{"documento": "prueba"})
	if dispatch.Success || !dispatch.ConnectivityFailure {
		t.Fatalf("loopback fiscal dispatch was not blocked: %#v", dispatch)
	}
	status := facturacionProveedorConnectionStatus(&dbpkg.FacturacionElectronicaPaisConfig{
		EmpresaID:  12,
		PaisCodigo: "CO",
		Ambiente:   "produccion",
		Proveedor:  "proveedor_externo",
		APIBaseURL: server.URL,
		Estado:     "activo",
	})
	if status["estado_conexion"] != "sin_endpoint" {
		t.Fatalf("loopback fiscal health endpoint was not blocked: %#v", status)
	}
	select {
	case <-hit:
		t.Fatal("blocked fiscal provider received an outbound request")
	default:
	}
}

func mustParseURLForTest(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse test URL: %v", err)
	}
	return parsed
}
