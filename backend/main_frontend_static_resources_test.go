package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestFrontendStaticResourcesExist(t *testing.T) {
	references, err := collectFrontendStaticResourceReferences(filepath.Join("..", "web"))
	if err != nil {
		t.Fatalf("collect frontend static resources: %v", err)
	}

	missing := make([]string, 0)
	for _, reference := range references {
		if _, err := os.Stat(reference.targetPath); err == nil {
			continue
		}
		if isAllowedPendingStaticResource(reference.targetPath) {
			continue
		}
		missing = append(missing, fmt.Sprintf("%s references %s -> %s", reference.sourcePath, reference.rawReference, reference.targetPath))
	}

	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("frontend static resources not found:\n%s", strings.Join(missing, "\n"))
	}
}

type staticResourceReference struct {
	sourcePath   string
	rawReference string
	targetPath   string
}

func collectFrontendStaticResourceReferences(webDir string) ([]staticResourceReference, error) {
	htmlAttrPattern := regexp.MustCompile(`(?i)(?:href|src|action)\s*=\s*["']([^"']+)["']`)
	cssURLPattern := regexp.MustCompile(`(?i)url\(\s*["']?([^"')]+)["']?\s*\)`)
	scriptBlockPattern := regexp.MustCompile(`(?is)<script[\s\S]*?</script>`)
	templateBlockPattern := regexp.MustCompile(`(?is)<template[\s\S]*?</template>`)

	references := make([]staticResourceReference, 0)
	err := filepath.WalkDir(webDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules", "vendor":
				return filepath.SkipDir
			default:
				return nil
			}
		}

		extension := strings.ToLower(filepath.Ext(path))
		if extension != ".html" && extension != ".css" {
			return nil
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		content := string(raw)
		pattern := cssURLPattern
		if extension == ".html" {
			content = scriptBlockPattern.ReplaceAllString(content, "")
			content = templateBlockPattern.ReplaceAllString(content, "")
			pattern = htmlAttrPattern
		}

		for _, match := range pattern.FindAllStringSubmatch(content, -1) {
			reference := normalizeStaticResourceReference(match[1])
			if reference == "" {
				continue
			}
			references = append(references, staticResourceReference{
				sourcePath:   filepath.ToSlash(path),
				rawReference: reference,
				targetPath:   filepath.ToSlash(resolveStaticResourceReference(path, webDir, reference)),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return references, nil
}

func normalizeStaticResourceReference(reference string) string {
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return ""
	}

	lowerReference := strings.ToLower(reference)
	skippedPrefixes := []string{
		"http:", "https:", "mailto:", "tel:", "javascript:", "data:", "blob:", "about:", "#",
		"/api/", "/super/api/", "/epayco/", "/wompi/", "/nequi/", "/auth/",
	}
	for _, prefix := range skippedPrefixes {
		if strings.HasPrefix(lowerReference, prefix) {
			return ""
		}
	}
	if reference == "/generate" || reference == "/download" {
		return ""
	}
	if strings.Contains(reference, "${") || strings.Contains(reference, "{{") || strings.Contains(reference, "+") || strings.Contains(reference, "`") || strings.Contains(reference, " ") {
		return ""
	}

	reference = strings.Split(reference, "#")[0]
	reference = strings.Split(reference, "?")[0]
	reference = strings.TrimSpace(reference)
	if reference == "" || reference == "/" {
		return ""
	}
	return reference
}

func resolveStaticResourceReference(sourcePath string, webDir string, reference string) string {
	if strings.HasPrefix(reference, "/") {
		return filepath.Join(webDir, strings.TrimPrefix(reference, "/"))
	}
	return filepath.Clean(filepath.Join(filepath.Dir(sourcePath), reference))
}

func isAllowedPendingStaticResource(targetPath string) bool {
	targetPath = filepath.ToSlash(targetPath)
	allowed := []string{
		"../web/descargas/rustdesk-cliente-windows-x64.exe",
		"../web/descargas/rustdesk-cliente-linux-amd64.deb",
		"../web/descargas/rustdesk-cliente-macos-x64.dmg",
		"../web/descargas/rustdesk-servidor-windows-x64.zip",
		"../web/descargas/rustdesk-servidor-linux-amd64.zip",
	}
	for _, value := range allowed {
		if targetPath == value {
			return true
		}
	}
	return false
}
