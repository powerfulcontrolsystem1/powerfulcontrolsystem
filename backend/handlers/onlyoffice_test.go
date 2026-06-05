package handlers

import (
	"net/http/httptest"
	"path/filepath"
	"testing"
)

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
