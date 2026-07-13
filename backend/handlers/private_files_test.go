package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testPNGHeader = []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d}

func TestSaveEmpresaPrivateUploadIsTenantScopedAndRandom(t *testing.T) {
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", t.TempDir())
	name1, path1, size1, err := saveEmpresaPrivateUpload(10, "buzon", ".png", strings.NewReader(string(testPNGHeader)), 1024)
	if err != nil {
		t.Fatalf("save first private upload: %v", err)
	}
	name2, path2, _, err := saveEmpresaPrivateUpload(10, "buzon", ".png", strings.NewReader(string(testPNGHeader)), 1024)
	if err != nil {
		t.Fatalf("save second private upload: %v", err)
	}
	if name1 == name2 || path1 == path2 {
		t.Fatal("private uploads must use collision-resistant names")
	}
	if size1 != int64(len(testPNGHeader)) {
		t.Fatalf("unexpected stored size: %d", size1)
	}
	if _, err := resolveEmpresaPrivateFile(11, "buzon", name1); err == nil {
		t.Fatal("another tenant must not resolve the private file")
	}
	if _, err := resolveEmpresaPrivateFile(10, "buzon", name1); err != nil {
		t.Fatalf("owner tenant should resolve its private file: %v", err)
	}
}

func TestSaveEmpresaPrivateUploadRejectsUnsafeContent(t *testing.T) {
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", t.TempDir())
	tests := []struct {
		name string
		ext  string
		body string
		max  int64
	}{
		{name: "dangerous extension", ext: ".html", body: "plain", max: 1024},
		{name: "active html", ext: ".txt", body: "<html><script>alert(1)</script></html>", max: 1024},
		{name: "active svg", ext: ".txt", body: "<svg onload='alert(1)'></svg>", max: 1024},
		{name: "mime mismatch", ext: ".jpg", body: "%PDF-1.7", max: 1024},
		{name: "too large", ext: ".txt", body: "123456", max: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, _, _, err := saveEmpresaPrivateUpload(10, "buzon", tt.ext, strings.NewReader(tt.body), tt.max); err == nil {
				t.Fatal("unsafe private upload was accepted")
			}
		})
	}
}

func TestResolveEmpresaPrivateFileRejectsTraversalAndSymlink(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", root)
	name, _, _, err := saveEmpresaPrivateUpload(10, "buzon", ".txt", strings.NewReader("safe text"), 1024)
	if err != nil {
		t.Fatalf("save private file: %v", err)
	}
	if _, err := resolveEmpresaPrivateFile(10, "buzon", "../"+name); err == nil {
		t.Fatal("path traversal was accepted")
	}
	outside := filepath.Join(root, "outside.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	tenantRoot, err := empresaPrivateCategoryRoot(10, "buzon")
	if err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(tenantRoot, "link.txt")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink creation is unavailable: %v", err)
	}
	if _, err := resolveEmpresaPrivateFile(10, "buzon", "link.txt"); err == nil {
		t.Fatal("symlink was accepted")
	}
}

func TestServeEmpresaPrivateFileSetsDownloadHeaders(t *testing.T) {
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", t.TempDir())
	name, _, _, err := saveEmpresaPrivateUpload(10, "buzon", ".txt", strings.NewReader("safe text"), 1024)
	if err != nil {
		t.Fatalf("save private file: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/buzon/archivo?ref="+name, nil)
	rec := httptest.NewRecorder()
	serveEmpresaPrivateFile(rec, req, 10, "buzon")
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("private download security headers are missing")
	}
	if !strings.HasPrefix(rec.Header().Get("Content-Disposition"), "attachment;") {
		t.Fatal("private file must be served as an attachment")
	}
}
