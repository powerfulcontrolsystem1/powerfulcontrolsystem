package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func signalingRequest(target string, empresaID int64) *http.Request {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req.Header.Set("Origin", "https://example.com")
	return req.WithContext(context.WithValue(req.Context(), "empresaID", empresaID))
}

func TestSoporteRemotoHTTPPropagaContextoEnConfiguracionYAlta(t *testing.T) {
	raw, err := os.ReadFile("soporte_remoto.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, forbidden := range []string{
		"dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp",
		"dbpkg.UpsertEmpresaSoporteRemotoConfig(dbEmp",
		"dbpkg.CreateEmpresaSoporteRemotoDispositivo(dbEmp",
		"dbpkg.GetEmpresaSoporteRemotoUso(dbEmp",
		"dbpkg.ListEmpresaSoporteRemotoDispositivos(dbEmp",
		"dbpkg.GetEmpresaSoporteRemotoDispositivoByID(dbEmp",
		"dbpkg.UpdateEmpresaSoporteRemotoDispositivo(dbEmp",
		"dbpkg.SetEmpresaSoporteRemotoDispositivoEstadoByID(dbEmp",
		"dbpkg.RegisterEmpresaSoporteRemotoDispositivoHeartbeat(dbEmp",
		"dbpkg.ValidateEmpresaSoporteRemotoDispositivoAccess(dbEmp",
		"dbpkg.CreateEmpresaSoporteRemotoSession(\n",
		"dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp",
		"dbpkg.ListEmpresaSoporteRemotoSesiones(dbEmp",
		"dbpkg.ResolveEmpresaSoporteRemotoViewerSession(dbEmp",
		"dbpkg.SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbEmp",
		"dbpkg.CreateEmpresaSoporteRemotoSignalingCredential(\n",
	} {
		if strings.Contains(source, forbidden) {
			t.Fatalf("support handler bypasses request context with %q", forbidden)
		}
	}
	for _, expected := range []string{
		"GetEmpresaSoporteRemotoConfigContext(r.Context()",
		"UpsertEmpresaSoporteRemotoConfigContext(r.Context()",
		"CreateEmpresaSoporteRemotoDispositivoContext(r.Context()",
		"GetEmpresaSoporteRemotoUsoContext(r.Context()",
		"ListEmpresaSoporteRemotoDispositivosContext(r.Context()",
		"GetEmpresaSoporteRemotoDispositivoByIDContext(r.Context()",
		"UpdateEmpresaSoporteRemotoDispositivoContext(r.Context()",
		"SetEmpresaSoporteRemotoDispositivoEstadoByIDContext(r.Context()",
		"RegisterEmpresaSoporteRemotoDispositivoHeartbeatContext(\n\t\tr.Context()",
		"ValidateEmpresaSoporteRemotoDispositivoAccessContext(r.Context()",
		"CreateEmpresaSoporteRemotoSessionContext(\n\t\tr.Context()",
		"GetEmpresaSoporteRemotoSessionByCodigoContext(r.Context()",
		"ListEmpresaSoporteRemotoSesionesContext(r.Context()",
		"ResolveEmpresaSoporteRemotoViewerSessionContext(r.Context()",
		"SetEmpresaSoporteRemotoSessionEstadoByCodigoContext(r.Context()",
		"CreateEmpresaSoporteRemotoSignalingCredentialContext(\n\t\tr.Context()",
	} {
		if !strings.Contains(source, expected) {
			t.Fatalf("missing cancellable support operation %q", expected)
		}
	}
}

func TestSoporteRemotoHTTPPropagaContextoEnSuperYWebRTC(t *testing.T) {
	for _, file := range []string{"super_soporte_remoto.go", "soporte_remoto_webrtc.go"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(raw)
		for _, forbidden := range []string{
			"dbpkg.GetEmpresaSoporteRemotoConfig(dbEmp",
			"dbpkg.GetEmpresaSoporteRemotoUso(dbEmp",
			"dbpkg.ListEmpresaSoporteRemotoDispositivos(dbEmp",
			"dbpkg.UpsertEmpresaSoporteRemotoConfig(dbEmp",
			"dbpkg.CreateEmpresaSoporteRemotoSession(dbEmp",
			"dbpkg.GetEmpresaSoporteRemotoSessionByCodigo(dbEmp",
			"dbpkg.ListEmpresaSoporteRemotoSesiones(dbEmp",
			"dbpkg.SetEmpresaSoporteRemotoSessionEstadoByCodigo(dbEmp",
			"dbpkg.ConsumeEmpresaSoporteRemotoSignalingCredential(dbEmp",
			"dbpkg.IsEmpresaSoporteRemotoSessionActive(dbEmp",
		} {
			if strings.Contains(source, forbidden) {
				t.Fatalf("%s bypasses request context with %q", file, forbidden)
			}
		}
	}
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
