package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSuperAdminPanelExposesPrivateHelpButton(t *testing.T) {
	htmlPath := filepath.Join("..", "web", "super_administrador.html")
	raw, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read super admin html: %v", err)
	}
	html := string(raw)
	if !strings.Contains(html, `id="superHelpNavItem"`) {
		t.Fatal("expected super admin panel to include private help nav item")
	}
	if !strings.Contains(html, `href="/ayuda/ayuda.html"`) || !strings.Contains(html, `target="contentFrame"`) {
		t.Fatal("expected private help button to open /ayuda/ayuda.html inside the super admin frame")
	}

	jsPath := filepath.Join("..", "web", "js", "super_administrador.js")
	raw, err = os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("read super admin js: %v", err)
	}
	js := string(raw)
	if !strings.Contains(js, `normalized === "/ayuda/ayuda.html"`) {
		t.Fatal("expected super admin navigation to allow persisting the private help page")
	}
	if strings.Contains(js, `"/ayuda/ayuda.html": true`) {
		t.Fatal("control_super_administrador allow list must not include private super admin help")
	}
}
