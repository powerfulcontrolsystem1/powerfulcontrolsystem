package handlers

import (
	dbpkg "github.com/you/pos-backend/db"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
	recovered, err := prepareSoporteComprasIAPurgeFile(dbpkg.EmpresaSoporteComprasIA{EmpresaID: 12, ArchivoURL: url, Estado: "eliminado"})
	if err != nil || recovered.quarantine != quarantine.quarantine {
		t.Fatalf("retry did not recover quarantine: %#v err=%v", recovered, err)
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

func TestSoporteComprasIAPurgePendingCanFinalizeAfterFileRemoval(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	tenantDir := filepath.Join(storage, "soportes_compras_ia", "empresa_12")
	if err := os.MkdirAll(tenantDir, 0o700); err != nil {
		t.Fatal(err)
	}
	row := dbpkg.EmpresaSoporteComprasIA{
		EmpresaID: 12, Estado: "purga_pendiente",
		ArchivoURL: "private://soportes_compras_ia/empresa_12/already-removed.pdf",
	}
	quarantine, err := prepareSoporteComprasIAPurgeFile(row)
	if err != nil || quarantine.original != "" || quarantine.quarantine != "" {
		t.Fatalf("pending metadata retry should need no file: %#v err=%v", quarantine, err)
	}
	if err := (soporteComprasIAPurgeFile{quarantine: filepath.Join(tenantDir, "already-gone.purge-aaaaaaaaaaaaaaaaaaaaaaaa")}).commit(); err != nil {
		t.Fatalf("concurrent idempotent commit rejected missing quarantine: %v", err)
	}
}

func TestSoporteComprasIAPurgeRejectsAmbiguousQuarantines(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	tenantDir := filepath.Join(storage, "soportes_compras_ia", "empresa_12")
	if err := os.MkdirAll(tenantDir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(tenantDir, "ambiguous.pdf")
	for _, suffix := range []string{"aaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbb"} {
		if err := os.WriteFile(path+".purge-"+suffix, []byte("quarantine"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	_, err := prepareSoporteComprasIAPurgeFile(dbpkg.EmpresaSoporteComprasIA{
		EmpresaID: 12, Estado: "purga_pendiente", ArchivoURL: "private://soportes_compras_ia/empresa_12/ambiguous.pdf",
	})
	if err == nil || !strings.Contains(err.Error(), "varias cuarentenas") {
		t.Fatalf("ambiguous quarantine was accepted: %v", err)
	}
}

func TestSoporteComprasIAQuarantineStatsAreTenantScoped(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	company12 := filepath.Join(storage, "soportes_compras_ia", "empresa_12")
	company53 := filepath.Join(storage, "soportes_compras_ia", "empresa_53")
	for _, dir := range []string{company12, company53} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	validName := "invoice.pdf.purge-aaaaaaaaaaaaaaaaaaaaaaaa"
	if err := os.WriteFile(filepath.Join(company12, validName), []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(company12, "not-quarantine.tmp"), []byte("ignored"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(company53, "foreign.pdf.purge-bbbbbbbbbbbbbbbbbbbbbbbb"), []byte("foreign"), 0o600); err != nil {
		t.Fatal(err)
	}
	count, bytes, err := soporteComprasIAQuarantineStats(12)
	if err != nil || count != 1 || bytes != 5 {
		t.Fatalf("tenant quarantine stats: count=%d bytes=%d err=%v", count, bytes, err)
	}
	for _, name := range []string{"x.purge-short", "x.purge-gggggggggggggggggggggggg", ".purge-aaaaaaaaaaaaaaaaaaaaaaaa"} {
		if isSoporteComprasIAQuarantineName(name) {
			t.Fatalf("invalid quarantine name accepted: %q", name)
		}
	}
}

func TestSoporteComprasIAStalePendingUsesBoundedThreshold(t *testing.T) {
	if got := parseSoporteComprasIAQuarantineThreshold("4"); got != 15 {
		t.Fatalf("threshold below minimum = %d", got)
	}
	if got := parseSoporteComprasIAQuarantineThreshold("60"); got != 60 {
		t.Fatalf("valid threshold = %d", got)
	}
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	rows := []dbpkg.EmpresaSoporteComprasIA{
		{FechaActualizacion: "2026-08-09T11:30:00Z"},
		{FechaActualizacion: "2026-08-09 11:55:00"},
		{FechaActualizacion: "invalid"},
	}
	if got := countSoporteComprasIAStalePending(rows, now, 15*time.Minute); got != 1 {
		t.Fatalf("stale pending = %d, want 1", got)
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
