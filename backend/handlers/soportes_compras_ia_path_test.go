package handlers

import (
	dbpkg "github.com/you/pos-backend/db"
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

func TestSafeSoporteComprasIAPathForEmpresaRejectsCrossTenantReference(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	foreignDir := filepath.Join(storage, "soportes_compras_ia", "empresa_53")
	if err := os.MkdirAll(foreignDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreignDir, "foreign.pdf"), []byte("%PDF-test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := safeSoporteComprasIAPathForEmpresa("private://soportes_compras_ia/empresa_53/foreign.pdf", 12); err == nil {
		t.Fatal("cross-tenant private reference was accepted")
	}
}

func TestSoporteComprasIAPurgeFileRollbackAndCommit(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	tenantDir := filepath.Join(storage, "soportes_compras_ia", "empresa_12")
	if err := os.MkdirAll(tenantDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(tenantDir, "purge.pdf")
	url := "private://soportes_compras_ia/empresa_12/purge.pdf"
	if err := os.WriteFile(path, []byte("%PDF-test"), 0o600); err != nil {
		t.Fatal(err)
	}
	quarantine, err := prepareSoporteComprasIAPurgeFile(dbpkg.EmpresaSoporteComprasIA{EmpresaID: 12, ArchivoURL: url})
	if err != nil || quarantine.quarantine == "" {
		t.Fatalf("prepare quarantine: %#v err=%v", quarantine, err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("original should be quarantined: %v", err)
	}
	if err := quarantine.rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("rollback did not restore original: %v", err)
	}
	quarantine, err = prepareSoporteComprasIAPurgeFile(dbpkg.EmpresaSoporteComprasIA{EmpresaID: 12, ArchivoURL: url})
	if err != nil {
		t.Fatal(err)
	}
	if err := quarantine.commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("committed purge retained original: %v", err)
	}
}

func TestSanitizeManualSoporteComprasIAClearsClientPrivateMetadata(t *testing.T) {
	row := sanitizeManualSoporteComprasIA(dbpkg.EmpresaSoporteComprasIA{
		ArchivoURL: "private://soportes_compras_ia/empresa_53/foreign.pdf", ArchivoNombre: "foreign.pdf",
		ArchivoMime: "application/pdf", ArchivoHash: "forged", Origen: "api",
	})
	if row.ArchivoURL != "" || row.ArchivoNombre != "" || row.ArchivoMime != "" || row.ArchivoHash != "" || row.Origen != "manual" {
		t.Fatalf("manual support retained client attachment metadata: %#v", row)
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
