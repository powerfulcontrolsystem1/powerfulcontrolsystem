package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSoporteComprasIAStorageAdmissionRespectsCompanyQuota(t *testing.T) {
	cfg := empresaStorageConfig{QuotaEnabled: true, BlockUploads: true, DefaultLimitMB: 1, MaxUploadMB: 10}
	usage := empresaStorageUsage{EmpresaID: 12, UsedBytes: 900, LimitBytes: 1_000}
	if err := validateSoporteComprasIAStorageUploadWithUsage(cfg, usage, 101); !errors.Is(err, errSoporteComprasIAStorageQuota) {
		t.Fatalf("quota must reject the incoming tenant attachment: %v", err)
	}
	if err := validateSoporteComprasIAStorageUploadWithUsage(cfg, usage, 100); err != nil {
		t.Fatalf("attachment that exactly reaches quota must be accepted: %v", err)
	}
}

func TestSoporteComprasIAStorageAdmissionHonorsQuotaSwitches(t *testing.T) {
	usage := empresaStorageUsage{EmpresaID: 12, UsedBytes: 1_000, LimitBytes: 1_000}
	for name, cfg := range map[string]empresaStorageConfig{
		"quota disabled":    {QuotaEnabled: false, BlockUploads: true, DefaultLimitMB: 1, MaxUploadMB: 10},
		"blocking disabled": {QuotaEnabled: true, BlockUploads: false, DefaultLimitMB: 1, MaxUploadMB: 10},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateSoporteComprasIAStorageUploadWithUsage(cfg, usage, 1); err != nil {
				t.Fatalf("storage policy must allow upload: %v", err)
			}
		})
	}
}

func TestSoporteComprasIAStorageAdmissionHonorsConfiguredMaxUpload(t *testing.T) {
	cfg := empresaStorageConfig{QuotaEnabled: true, BlockUploads: true, DefaultLimitMB: 100, MaxUploadMB: 1}
	usage := empresaStorageUsage{EmpresaID: 12, LimitBytes: 100 << 20}
	if err := validateSoporteComprasIAStorageUploadWithUsage(cfg, usage, (1<<20)+1); err == nil {
		t.Fatal("configured maximum upload must reject oversized attachment")
	}
	if err := validateSoporteComprasIAStorageUploadWithUsage(cfg, usage, 1<<20); err != nil {
		t.Fatalf("attachment at configured maximum must be accepted: %v", err)
	}
}

func TestEmpresaSoportesComprasIAStorageBytesAreTenantScoped(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	company12 := filepath.Join(storage, "soportes_compras_ia", "empresa_12")
	company7 := filepath.Join(storage, "soportes_compras_ia", "empresa_7")
	if err := os.MkdirAll(company12, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(company7, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(company12, "a.pdf"), []byte("12345"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(company7, "b.pdf"), []byte("1234567"), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := empresaSoportesComprasIAStorageBytes(12); got != 5 {
		t.Fatalf("company 12 private usage = %d, want 5", got)
	}
}

func TestSoporteComprasIAStorageLockIsPostgresBackedAndTenantScoped(t *testing.T) {
	raw, err := os.ReadFile("soportes_compras_ia.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, want := range []string{
		"func acquireSoporteComprasIAStorageLock", "pg_try_advisory_lock",
		"pg_advisory_unlock", "soporteComprasIAStorageNamespace, empresaID",
		"errSoporteComprasIAStorageBusy", "defer release()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("storage concurrency contract missing %q", want)
		}
	}
}

func TestSoporteComprasIAAllowedExtensionsRejectActiveContent(t *testing.T) {
	for _, ext := range []string{".html", ".htm", ".svg", ".js"} {
		if soporteComprasIAAllowedExt[ext] {
			t.Fatalf("active content extension must stay blocked: %s", ext)
		}
	}
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp", ".pdf", ".xml"} {
		if !soporteComprasIAAllowedExt[ext] {
			t.Fatalf("business attachment extension missing: %s", ext)
		}
	}
}
