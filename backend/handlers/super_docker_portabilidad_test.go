package handlers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSuperDockerPortableRootAndExclusions(t *testing.T) {
	root := t.TempDir()
	mustWrite := func(rel, content string) {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	mustWrite("deploy/docker-compose.platform.yml", "services:\n  postgres:\n")
	mustWrite("deploy/.env.platform.example", "POSTGRES_PASSWORD=change-me\n")
	mustWrite("deploy/.env.platform", "POSTGRES_PASSWORD=secret\n")
	mustWrite("deploy/docker/backend.Dockerfile", "FROM scratch\n")
	mustWrite("deploy/docker/frontend.Dockerfile", "FROM nginx\n")
	mustWrite("backend/go.mod", "module example\n")
	mustWrite("backend/.env.local", "CONFIG_ENC_KEY=secret\n")
	mustWrite("web/index.html", "<!doctype html>")
	mustWrite("web/uploads/private.txt", "private")
	mustWrite("documentos/docker_vps_operacion.md", "doc")
	mustWrite("scripts/rs.ps1", "Write-Host ok")
	mustWrite(".dockerignore", ".env\n")
	mustWrite("CHANGELOG.md", "changelog")

	if err := superDockerPortableAssertReadyForTests(root); err != nil {
		t.Fatalf("expected valid portable root: %v", err)
	}

	files, err := superDockerPortableSortedIncludedForTests(root)
	if err != nil {
		t.Fatalf("walk portable root: %v", err)
	}
	seen := map[string]bool{}
	for _, file := range files {
		seen[file] = true
	}
	for _, want := range []string{
		"deploy/docker-compose.platform.yml",
		"deploy/.env.platform.example",
		"backend/go.mod",
		"web/index.html",
		"documentos/docker_vps_operacion.md",
		"scripts/rs.ps1",
		".dockerignore",
		"CHANGELOG.md",
	} {
		if !seen[want] {
			t.Fatalf("expected %s in portable package, got %#v", want, files)
		}
	}
	for _, blocked := range []string{
		"deploy/.env.platform",
		"backend/.env.local",
		"web/uploads/private.txt",
	} {
		if seen[blocked] {
			t.Fatalf("did not expect %s in portable package, got %#v", blocked, files)
		}
	}
}
