package handlers

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestSoporteComprasIAFileIntegrityAndDownloadMIME(t *testing.T) {
	content := []byte("%PDF-1.7\ncontenido controlado\n%%EOF")
	sum := sha256.Sum256(content)
	expected := hex.EncodeToString(sum[:])
	before := SupportFileOperationalMetrics()
	if err := verifySoporteComprasIABytesIntegrity(content, expected); err != nil {
		t.Fatalf("archivo integro rechazado: %v", err)
	}
	if err := verifySoporteComprasIABytesIntegrity(content, strings.ToUpper(expected)); err != nil {
		t.Fatalf("hash mayuscula integro rechazado: %v", err)
	}
	if err := verifySoporteComprasIABytesIntegrity(content, ""); err != nil {
		t.Fatalf("archivo legado sin hash rechazado: %v", err)
	}
	if err := verifySoporteComprasIABytesIntegrity(append(content, 'x'), expected); !errors.Is(err, errSoporteComprasIAIntegrity) {
		t.Fatalf("archivo alterado = %v", err)
	}
	if err := verifySoporteComprasIABytesIntegrity(content, "hash-invalido"); !errors.Is(err, errSoporteComprasIAIntegrity) {
		t.Fatalf("hash invalido = %v", err)
	}
	if got := SupportFileOperationalMetrics() - before; got != 2 {
		t.Fatalf("fallos de integridad = %d, want 2", got)
	}

	path := filepath.Join(t.TempDir(), "factura.pdf")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if _, err := file.Write(content); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	if err := verifySoporteComprasIAReaderIntegrity(file, expected); err != nil {
		t.Fatalf("lector integro rechazado: %v", err)
	}
	first := make([]byte, 4)
	if _, err := io.ReadFull(file, first); err != nil || string(first) != "%PDF" {
		t.Fatalf("lector no regreso al inicio: %q err=%v", first, err)
	}

	for _, tc := range []struct{ mime, path, want string }{
		{"application/pdf", "factura.pdf", "application/pdf"},
		{"text/html", "factura.pdf", "application/octet-stream"},
		{"text/xml", "factura.xml", "application/xml"},
		{"image/png", "factura.exe", "application/octet-stream"},
	} {
		if got := safeSoporteComprasIADownloadMIME(tc.mime, tc.path); got != tc.want {
			t.Fatalf("MIME seguro (%s,%s) = %q, want %q", tc.mime, tc.path, got, tc.want)
		}
	}
}

func TestSoporteComprasIAClamAVCleanMalwareAndFailClosed(t *testing.T) {
	att := &aiAttachment{Filename: "factura.pdf", MimeType: "application/pdf", Bytes: []byte("%PDF-1.7\n%%EOF")}
	before := SupportAntivirusOperationalMetrics()
	t.Run("clean", func(t *testing.T) {
		t.Setenv("PCS_SUPPORTS_CLAMAV_ADDR", startFakeSoporteClamd(t, "stream: OK"))
		t.Setenv("PCS_SUPPORTS_CLAMAV_REQUIRED", "true")
		if err := scanSoporteComprasIAAttachment(att); err != nil {
			t.Fatalf("clean attachment rejected: %v", err)
		}
	})
	t.Run("malware", func(t *testing.T) {
		t.Setenv("PCS_SUPPORTS_CLAMAV_ADDR", startFakeSoporteClamd(t, "stream: Eicar-Signature FOUND"))
		t.Setenv("PCS_SUPPORTS_CLAMAV_REQUIRED", "true")
		if err := scanSoporteComprasIAAttachment(att); !errors.Is(err, errSoporteComprasIAMalware) {
			t.Fatalf("malware result = %v", err)
		}
	})
	t.Run("required without address", func(t *testing.T) {
		t.Setenv("PCS_SUPPORTS_CLAMAV_ADDR", "")
		t.Setenv("PCS_SUPPORTS_CLAMAV_REQUIRED", "true")
		if err := scanSoporteComprasIAAttachment(att); !errors.Is(err, errSoporteComprasIAAntivirusUnavailable) {
			t.Fatalf("required scanner result = %v", err)
		}
	})
	t.Run("optional without address", func(t *testing.T) {
		t.Setenv("PCS_SUPPORTS_CLAMAV_ADDR", "")
		t.Setenv("PCS_SUPPORTS_CLAMAV_REQUIRED", "false")
		if err := scanSoporteComprasIAAttachment(att); err != nil {
			t.Fatalf("optional scanner disabled = %v", err)
		}
	})
	after := SupportAntivirusOperationalMetrics()
	if after.Clean-before.Clean != 1 || after.Malware-before.Malware != 1 ||
		after.Unavailable-before.Unavailable != 1 || after.Bypassed-before.Bypassed != 1 {
		t.Fatalf("antivirus metrics delta = clean:%d malware:%d unavailable:%d bypassed:%d, want 1 each",
			after.Clean-before.Clean, after.Malware-before.Malware,
			after.Unavailable-before.Unavailable, after.Bypassed-before.Bypassed)
	}
	t.Setenv("PCS_SUPPORTS_CLAMAV_ADDR", "127.0.0.1:3310")
	t.Setenv("PCS_SUPPORTS_CLAMAV_REQUIRED", "true")
	configured := SupportAntivirusOperationalMetrics()
	if !configured.Required || !configured.Configured {
		t.Fatalf("antivirus configuration metrics = required:%v configured:%v, want true/true", configured.Required, configured.Configured)
	}
}

func TestSupportAntivirusMetricsAreConcurrentAndTenantFree(t *testing.T) {
	t.Setenv("PCS_SUPPORTS_CLAMAV_ADDR", "")
	t.Setenv("PCS_SUPPORTS_CLAMAV_REQUIRED", "false")
	before := SupportAntivirusOperationalMetrics()
	const workers = 64
	done := make(chan struct{}, workers)
	for i := 0; i < workers; i++ {
		go func() {
			_ = scanSoporteComprasIAAttachment(nil)
			done <- struct{}{}
		}()
	}
	for i := 0; i < workers; i++ {
		<-done
	}
	after := SupportAntivirusOperationalMetrics()
	if got := after.Bypassed - before.Bypassed; got != workers {
		t.Fatalf("concurrent bypassed scans = %d, want %d", got, workers)
	}
}

func startFakeSoporteClamd(t *testing.T, response string) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		if _, err := reader.ReadString(0); err != nil {
			return
		}
		for {
			var size uint32
			if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
				return
			}
			if size == 0 {
				break
			}
			if _, err := io.CopyN(io.Discard, reader, int64(size)); err != nil {
				return
			}
		}
		_, _ = conn.Write(append([]byte(response), 0))
	}()
	return listener.Addr().String()
}

func TestSoporteComprasIAAttachmentValidatesSignatureAndCanonicalMIME(t *testing.T) {
	tests := []struct {
		name, filename, wantMIME string
		content                  []byte
	}{
		{"png", "factura.png", "image/png", append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0}, 24)...)},
		{"jpeg", "factura.jpg", "image/jpeg", []byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\xff\xd9")},
		{"webp", "factura.webp", "image/webp", []byte("RIFF\x04\x00\x00\x00WEBPVP8 ")},
		{"pdf", "factura.pdf", "application/pdf", []byte("%PDF-1.7\n1 0 obj\n<<>>\nendobj\n%%EOF")},
		{"xml", "factura.xml", "application/xml", []byte(`<?xml version="1.0"?><Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"><ID>FV-1</ID></Invoice>`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			att := &aiAttachment{Filename: tt.filename, MimeType: "application/octet-stream", Bytes: tt.content}
			if err := validateSoporteComprasIAAttachment(att); err != nil {
				t.Fatalf("valid attachment rejected: %v", err)
			}
			if att.MimeType != tt.wantMIME {
				t.Fatalf("canonical MIME = %q, want %q", att.MimeType, tt.wantMIME)
			}
		})
	}
}

func TestSoporteComprasIAAttachmentRejectsSpoofedAndActiveContent(t *testing.T) {
	tests := []aiAttachment{
		{Filename: "falso.pdf", MimeType: "application/pdf", Bytes: []byte("<html>not pdf</html>")},
		{Filename: "falso.png", MimeType: "image/png", Bytes: []byte("not an image")},
		{Filename: "factura.exe", MimeType: "application/pdf", Bytes: []byte("%PDF-1.7")},
		{Filename: "xxe.xml", MimeType: "application/xml", Bytes: []byte(`<!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><Invoice>&xxe;</Invoice>`)},
		{Filename: "activo.xml", MimeType: "application/xml", Bytes: []byte(`<Invoice><script>alert(1)</script></Invoice>`)},
		{Filename: "pi.xml", MimeType: "application/xml", Bytes: []byte(`<?run command?><Invoice/>`)},
		{Filename: "roto.xml", MimeType: "application/xml", Bytes: []byte(`<Invoice><ID></Invoice>`)},
	}
	for i := range tests {
		if err := validateSoporteComprasIAAttachment(&tests[i]); err == nil {
			t.Fatalf("hostile attachment %d was accepted", i)
		}
	}
}

func TestCleanupUnpersistedSoporteComprasIAAttachmentRemovesOnlyPrivateFile(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	att := &aiAttachment{Filename: "factura.pdf", MimeType: "application/pdf", Bytes: []byte("%PDF-1.7\n%%EOF")}
	if err := validateSoporteComprasIAAttachment(att); err != nil {
		t.Fatal(err)
	}
	privateURL, _, _, _, _, err := saveSoporteComprasIAAttachment(att, 12)
	if err != nil {
		t.Fatal(err)
	}
	path, err := safeSoporteComprasIAPathFromURL(privateURL)
	if err != nil {
		t.Fatal(err)
	}
	cleanupUnpersistedSoporteComprasIAAttachment(privateURL)
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unpersisted private file still exists: %v", err)
	}
	outside := filepath.Join(storage, "outside.pdf")
	if err := os.WriteFile(outside, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	cleanupUnpersistedSoporteComprasIAAttachment(outside)
	if _, err := os.Stat(outside); err != nil {
		t.Fatalf("cleanup touched an out-of-scope path: %v", err)
	}
}

func TestSoporteComprasIARetentionPreviewCountsOnlySafePrivateFiles(t *testing.T) {
	storage := t.TempDir()
	t.Setenv("PCS_PRIVATE_STORAGE_DIR", storage)
	att := &aiAttachment{Filename: "factura.pdf", MimeType: "application/pdf", Bytes: []byte("%PDF-1.7\n%%EOF")}
	if err := validateSoporteComprasIAAttachment(att); err != nil {
		t.Fatal(err)
	}
	privateURL, _, _, _, _, err := saveSoporteComprasIAAttachment(att, 12)
	if err != nil {
		t.Fatal(err)
	}
	rows := []dbpkg.EmpresaSoporteComprasIA{{EmpresaID: 12, ArchivoURL: privateURL}, {EmpresaID: 12, ArchivoURL: "private://soportes_compras_ia/empresa_53/ajeno.pdf"}, {EmpresaID: 12, ArchivoURL: "https://example.com/file.pdf"}}
	if got, want := soporteComprasIARetentionBytes(rows), int64(len(att.Bytes)); got != want {
		t.Fatalf("retention bytes = %d, want %d", got, want)
	}
	for input, want := range map[string]int{"": 90, "0": 90, "3651": 90, "abc": 90, "1": 1, "3650": 3650, "120": 120} {
		if got := parseSoporteComprasIARetentionDays(input); got != want {
			t.Fatalf("retention days %q = %d, want %d", input, got, want)
		}
	}
}

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
