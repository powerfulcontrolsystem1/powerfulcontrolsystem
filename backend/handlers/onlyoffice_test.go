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
