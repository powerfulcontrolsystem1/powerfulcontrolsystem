package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

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
