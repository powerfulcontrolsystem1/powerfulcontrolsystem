package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSafeSoporteComprasIAPathUsesPrivateTenantRoot(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	root := filepath.Join(storage, "soportes_compras_ia")
	tenantDir := filepath.Join(root, "empresa_12")
	if err := os.MkdirAll(tenantDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(tenantDir, "abcdef.pdf")
	if err := os.WriteFile(path, []byte("%PDF-test"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := safeSoporteComprasIAPathFromURL("private://soportes_compras_ia/empresa_12/abcdef.pdf")
	if err != nil || got == "" {
		t.Fatalf("private tenant file rejected: path=%q err=%v", got, err)
	}
	if _, err := safeSoporteComprasIAPathFromURL("/uploads/soportes_compras_ia/empresa_12/abcdef.pdf"); err == nil {
		t.Fatal("legacy public upload URL was accepted")
	}
	if _, err := safeSoporteComprasIAPathFromURL("private://soportes_compras_ia/../outside.pdf"); err == nil {
		t.Fatal("private traversal was accepted")
	}
}

func TestResolveExistingPrivateFileUnderRootRejectsEscapes(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "documento.pdf")
	if err := os.WriteFile(inside, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := resolveExistingPrivateFileUnderRoot(root, inside); err != nil || got == "" {
		t.Fatalf("inside file rejected: path=%q err=%v", got, err)
	}
	if _, err := resolveExistingPrivateFileUnderRoot(root, filepath.Join(root, "..", "outside.pdf")); err == nil {
		t.Fatal("path outside root was accepted")
	}

	outside := filepath.Join(t.TempDir(), "outside.pdf")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.pdf")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := resolveExistingPrivateFileUnderRoot(root, link); err == nil {
		t.Fatal("symlink outside root was accepted")
	}
}
