package handlers

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSuperVPS2HandlerRequiresSuperAdmin(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/super/api/vps2?action=status", nil)
	res := httptest.NewRecorder()

	SuperVPS2Handler(nil).ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestBuildSuperVPS2FilesResponseDegradesWithoutReachableVPS2(t *testing.T) {
	t.Setenv("PCS_VPS2_STATUS_FILE", filepath.Join(t.TempDir(), "missing.json"))
	cfg := superVPS2Config{
		Host:              "127.0.0.1",
		Port:              1,
		NextcloudDataPath: "/srv/data/nextcloud/data",
	}
	status, payload := buildSuperVPS2FilesResponse(cfg, "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want %d for degraded monitoring state", status, http.StatusOK)
	}
	if payload.OK {
		t.Fatal("payload must preserve ok=false when VPS2 and its snapshot are unavailable")
	}
	if len(payload.Errors) == 0 {
		t.Fatal("degraded payload must explain why the file index is unavailable")
	}
}
