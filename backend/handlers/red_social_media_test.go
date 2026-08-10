package handlers

import (
	"os"
	"path/filepath"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestAvailableRedSocialPhotoURL(t *testing.T) {
	root := t.TempDir()
	validRelative := filepath.Join("uploads", "red_social", "empresa_12", "post.png")
	validAbsolute := filepath.Join(root, validRelative)
	if err := os.MkdirAll(filepath.Dir(validAbsolute), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(validAbsolute, []byte("png"), 0o600); err != nil {
		t.Fatal(err)
	}

	if got := availableRedSocialPhotoURL(root, "/uploads/red_social/empresa_12/post.png"); got != "/uploads/red_social/empresa_12/post.png" {
		t.Fatalf("se esperaba conservar archivo existente, got %q", got)
	}
	if got := availableRedSocialPhotoURL(root, "/uploads/red_social/empresa_12/missing.png"); got != "" {
		t.Fatalf("se esperaba ocultar archivo inexistente, got %q", got)
	}
	if got := availableRedSocialPhotoURL(root, "https://cdn.example.test/post.png"); got != "https://cdn.example.test/post.png" {
		t.Fatalf("se esperaba conservar URL externa, got %q", got)
	}
	if got := availableRedSocialPhotoURL(root, "/uploads/red_social/../../secreto.txt"); got != "" {
		t.Fatalf("se esperaba rechazar traversal, got %q", got)
	}
}

func TestFilterUnavailableRedSocialPhotos(t *testing.T) {
	root := t.TempDir()
	pubs := []dbpkg.PublicacionRedSocial{
		{FotoURL: "/uploads/red_social/empresa_7/ausente.png"},
		{FotoURL: "https://cdn.example.test/vigente.png"},
	}
	filterUnavailableRedSocialPhotos(root, pubs)
	if pubs[0].FotoURL != "" {
		t.Fatalf("foto local ausente no fue filtrada: %q", pubs[0].FotoURL)
	}
	if pubs[1].FotoURL == "" {
		t.Fatal("URL externa fue filtrada incorrectamente")
	}
}
