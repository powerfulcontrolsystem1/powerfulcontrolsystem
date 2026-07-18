package handlers

import (
	"archive/zip"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOnlyOfficeRuntimeResolversDoNotPersistDefaults(t *testing.T) {
	body, err := os.ReadFile("onlyoffice_super_config.go")
	if err != nil {
		t.Fatalf("read onlyoffice config source: %v", err)
	}
	source := string(body)
	if strings.Contains(source, "onlyOfficeEnsureJWTSecret") {
		t.Fatal("OnlyOffice must not generate or persist JWT secrets during HTTP requests")
	}
	for _, function := range []string{"onlyOfficeResolveJWTSecret", "onlyOfficeResolveDocumentServerURL"} {
		start := strings.Index(source, "func "+function)
		if start < 0 {
			t.Fatalf("%s not found", function)
		}
		section := source[start:]
		if next := strings.Index(section[1:], "\nfunc "); next >= 0 {
			section = section[:next+1]
		}
		if strings.Contains(section, "SetConfigValue(") {
			t.Fatalf("%s must be read-only at runtime", function)
		}
	}
}

func TestOnlyOfficeEmpresaDocsDirUsesConfiguredEmpresaRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PCS_DATA_ROOT", filepath.Join(root, "empresas"))

	got, err := onlyOfficeEmpresaDocsDir(32)
	if err != nil {
		t.Fatalf("onlyOfficeEmpresaDocsDir returned error: %v", err)
	}
	want := filepath.Join(root, "empresas", "32", "documentos")
	if got != want {
		t.Fatalf("unexpected docs dir\nwant: %s\n got: %s", want, got)
	}
}

func TestOnlyOfficeEmpresaDocsDirAddsEmpresasWhenRootIsGeneric(t *testing.T) {
	root := t.TempDir()
	t.Setenv("PCS_DATA_ROOT", root)

	got, err := onlyOfficeEmpresaDocsDir(7)
	if err != nil {
		t.Fatalf("onlyOfficeEmpresaDocsDir returned error: %v", err)
	}
	want := filepath.Join(root, "empresas", "7", "documentos")
	if got != want {
		t.Fatalf("unexpected docs dir\nwant: %s\n got: %s", want, got)
	}
}

func TestOnlyOfficeBrowserDocumentServerURLRewritesInternalDockerHost(t *testing.T) {
	t.Setenv("ONLYOFFICE_PUBLIC_DOCUMENT_SERVER_URL", "")
	t.Setenv("ONLYOFFICE_BROWSER_DOCUMENT_SERVER_URL", "")
	req := httptest.NewRequest("GET", "http://powerfulcontrolsystem.com/api/empresa/documentos", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	got, rewritten := onlyOfficeBrowserDocumentServerURL(req, nil, "http://onlyoffice-documentserver:80")
	if !rewritten {
		t.Fatalf("expected internal document server URL to be rewritten")
	}
	want := "https://onlyoffice.powerfulcontrolsystem.com"
	if got != want {
		t.Fatalf("unexpected browser document server URL\nwant: %s\n got: %s", want, got)
	}
}

func TestOnlyOfficeBrowserDocumentServerURLKeepsPublicHost(t *testing.T) {
	t.Setenv("ONLYOFFICE_PUBLIC_DOCUMENT_SERVER_URL", "")
	t.Setenv("ONLYOFFICE_BROWSER_DOCUMENT_SERVER_URL", "")
	req := httptest.NewRequest("GET", "https://powerfulcontrolsystem.com/api/empresa/documentos", nil)

	got, rewritten := onlyOfficeBrowserDocumentServerURL(req, nil, "https://onlyoffice.powerfulcontrolsystem.com")
	if rewritten {
		t.Fatalf("did not expect public document server URL to be rewritten")
	}
	want := "https://onlyoffice.powerfulcontrolsystem.com"
	if got != want {
		t.Fatalf("unexpected browser document server URL\nwant: %s\n got: %s", want, got)
	}
}

func TestOnlyOfficeAttachConfigTokenUsesTopLevelTokenOnly(t *testing.T) {
	cfg := map[string]any{
		"documentType": "word",
		"document": map[string]any{
			"fileType": "docx",
			"key":      "key-1",
			"title":    "Prueba.docx",
			"url":      "https://example.com/file.docx",
			"token":    "stale-document-token",
		},
		"editorConfig": map[string]any{
			"mode":        "edit",
			"callbackUrl": "https://example.com/callback",
			"token":       "stale-editor-token",
		},
	}

	jwt, err := onlyOfficeAttachConfigToken("secret-for-test", cfg)
	if err != nil {
		t.Fatalf("onlyOfficeAttachConfigToken returned error: %v", err)
	}
	if jwt == "" {
		t.Fatalf("expected jwt")
	}
	if cfg["token"] != jwt {
		t.Fatalf("expected top-level config token")
	}
	if doc, ok := cfg["document"].(map[string]any); !ok || doc["token"] != nil {
		t.Fatalf("document.token must not be sent to Document Server")
	}
	if ed, ok := cfg["editorConfig"].(map[string]any); !ok || ed["token"] != nil {
		t.Fatalf("editorConfig.token must not be sent to Document Server")
	}
}

func TestCopyOnlyOfficeCallbackFileRejectsOversizedDocument(t *testing.T) {
	var dst bytes.Buffer
	tooLarge := bytes.Repeat([]byte("x"), int(onlyOfficeCallbackMaxBytes)+1)
	if err := copyOnlyOfficeCallbackFile(&dst, bytes.NewReader(tooLarge)); err == nil {
		t.Fatal("expected oversized callback document to be rejected")
	}
	if got := int64(dst.Len()); got != onlyOfficeCallbackMaxBytes+1 {
		t.Fatalf("expected size probe to stop at max + 1, got %d", got)
	}
}

func TestCopyOnlyOfficeCallbackFileKeepsCompleteAllowedDocument(t *testing.T) {
	var dst bytes.Buffer
	source := []byte("valid onlyoffice document")
	if err := copyOnlyOfficeCallbackFile(&dst, bytes.NewReader(source)); err != nil {
		t.Fatalf("expected allowed callback document: %v", err)
	}
	if !bytes.Equal(dst.Bytes(), source) {
		t.Fatal("callback document changed during copy")
	}
}

func TestOnlyOfficeTemporaryTokenHandlersDisableCaching(t *testing.T) {
	for _, handler := range []http.HandlerFunc{
		OnlyOfficeFilePublicHandler(nil),
		OnlyOfficeCallbackPublicHandler(nil),
	} {
		rec := httptest.NewRecorder()
		handler(rec, httptest.NewRequest(http.MethodGet, "/api/onlyoffice/file?token=temporary", nil))
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Fatalf("expected no-store for temporary OnlyOffice token, got %q", got)
		}
	}
}

func TestOnlyOfficeBuildBlankPPTXIncludesPresentationMinimumParts(t *testing.T) {
	raw, err := onlyOfficeBuildBlankPPTX()
	if err != nil {
		t.Fatalf("onlyOfficeBuildBlankPPTX returned error: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("pptx must be a valid zip: %v", err)
	}
	seen := make(map[string]bool)
	for _, f := range zr.File {
		seen[f.Name] = true
	}
	required := []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"ppt/presentation.xml",
		"ppt/_rels/presentation.xml.rels",
		"ppt/slides/slide1.xml",
		"ppt/slides/_rels/slide1.xml.rels",
		"ppt/slideMasters/slideMaster1.xml",
		"ppt/slideMasters/_rels/slideMaster1.xml.rels",
		"ppt/slideLayouts/slideLayout1.xml",
		"ppt/slideLayouts/_rels/slideLayout1.xml.rels",
		"ppt/theme/theme1.xml",
	}
	for _, name := range required {
		if !seen[name] {
			t.Fatalf("pptx missing required part %s", name)
		}
	}
}
