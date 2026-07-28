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

func TestPublicDownloadsAreExplicitlyAllowlisted(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "nginx", "pcs.conf"))
	if err != nil {
		t.Fatalf("read frontend nginx config: %v", err)
	}
	config := string(raw)
	for _, name := range []string{
		"rustdesk-cliente-windows-x64",
		"rustdesk-cliente-linux-amd64",
		"rustdesk-cliente-macos-x64",
		"rustdesk-servidor-windows-x64",
		"rustdesk-servidor-linux-amd64",
	} {
		if !strings.Contains(config, name) {
			t.Fatalf("public download %q is not explicitly allowlisted", name)
		}
	}
	if !strings.Contains(config, "location /descargas/") || !strings.Contains(config, "return 404;") {
		t.Fatal("non-allowlisted public downloads must be rejected by nginx")
	}
}

func TestNextcloudFramePolicyUsesExactOrigins(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "nginx", "pcs.conf"))
	if err != nil {
		t.Fatalf("read frontend nginx config: %v", err)
	}
	config := string(raw)
	staticHeadersRaw, err := os.ReadFile(filepath.Join("..", "deploy", "nginx", "pcs-static-security-headers.inc"))
	if err != nil {
		t.Fatalf("read static nginx security headers: %v", err)
	}
	staticHeaders := string(staticHeadersRaw)
	if !strings.Contains(config, "include /etc/nginx/conf.d/pcs-static-security-headers.inc;") {
		t.Fatal("static frontend locations must include the static security header policy")
	}
	origin := "https://nextcloud.powerfulcontrolsystem.com"
	if count := strings.Count(staticHeaders, origin); count != 2 {
		t.Fatalf("Nextcloud origin must appear once in enforced CSP and once in report-only CSP; got %d", count)
	}
	if strings.Contains(staticHeaders, "*.powerfulcontrolsystem.com") {
		t.Fatal("Nextcloud framing must not rely on a wildcard company origin")
	}
	if strings.Contains(config, "add_header Content-Security-Policy") {
		t.Fatal("server-level frontend CSP would duplicate backend API CSP")
	}
	for _, required := range []string{
		`proxy_hide_header X-Content-Type-Options;`,
		`add_header X-Content-Type-Options "nosniff" always;`,
	} {
		if !strings.Contains(config, required) {
			t.Fatalf("frontend proxy must normalize dynamic X-Content-Type-Options with %q", required)
		}
	}

	scriptRaw, err := os.ReadFile(filepath.Join("..", "deploy", "scripts", "vps-configure-nextcloud-host-nginx.sh"))
	if err != nil {
		t.Fatalf("read Nextcloud host Nginx script: %v", err)
	}
	script := string(scriptRaw)
	for _, required := range []string{
		"frame-ancestors 'self' $EMBED_ORIGIN",
		"Nextcloud ya emite frame-ancestors",
		"curl -kfsSL --max-time 15 -D - -o /dev/null",
		"nginx -t",
		"cp -a \"$backup\" \"$SITE_AVAILABLE\"",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("Nextcloud host policy script is missing %q", required)
		}
	}
	if strings.Contains(script, "proxy_hide_header Content-Security-Policy") || strings.Contains(script, "docker compose") {
		t.Fatal("frame policy script must preserve vendor CSP and must not recreate Nextcloud")
	}
}

func TestStagingEdgeKeepsOnlyTransportHeaders(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "scripts", "vps-configure-staging-nginx.sh"))
	if err != nil {
		t.Fatalf("read staging nginx script: %v", err)
	}
	script := string(raw)
	for _, required := range []string{
		`add_header X-Frame-Options "SAMEORIGIN" always;`,
		`add_header Strict-Transport-Security "max-age=15552000; includeSubDomains" always;`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("staging edge header missing %q", required)
		}
	}
	for _, forbidden := range []string{
		`add_header Content-Security-Policy`,
		`add_header X-Content-Type-Options`,
		`add_header Referrer-Policy`,
		`add_header Permissions-Policy`,
	} {
		if strings.Contains(script, forbidden) {
			t.Fatalf("staging edge must not duplicate application header %q", forbidden)
		}
	}
}

func TestPanelGuidedSetupUsesCSRFAndNextcloudKeepsEmpresaContext(t *testing.T) {
	panel, err := os.ReadFile(filepath.Join("..", "web", "administrar_empresa", "panel.html"))
	if err != nil {
		t.Fatalf("read empresa panel: %v", err)
	}
	if !strings.Contains(string(panel), "X-CSRF-Token") || !strings.Contains(string(panel), "readCookieValue(\"pcs_csrf\")") {
		t.Fatal("guided setup mutations must include the browser CSRF token")
	}
	nextcloud, err := os.ReadFile(filepath.Join("..", "web", "js", "nextcloud_empresa.js"))
	if err != nil {
		t.Fatalf("read Nextcloud company script: %v", err)
	}
	if !strings.Contains(string(nextcloud), "__resolveEmpresaIdContext") || !strings.Contains(string(nextcloud), "__empresaModuleGuard.resolveEmpresaId") {
		t.Fatal("Nextcloud must resolve the validated company context from its parent shell")
	}
}

func TestEmpresaSubmenuContextInstallsCSRFForDirectOperationalPages(t *testing.T) {
	contextScript, err := os.ReadFile(filepath.Join("..", "web", "js", "empresa_submenu_context.js"))
	if err != nil {
		t.Fatalf("read empresa submenu context: %v", err)
	}
	content := string(contextScript)
	for _, required := range []string{"installCSRFFetch", "pcs_csrf", "X-CSRF-Token", "__pcsCSRFFetchInstalled"} {
		if !strings.Contains(content, required) {
			t.Fatalf("empresa submenu context must install CSRF fetch support; missing %q", required)
		}
	}

	carrito, err := os.ReadFile(filepath.Join("..", "web", "administrar_empresa", "carrito_de_compras.html"))
	if err != nil {
		t.Fatalf("read carrito page: %v", err)
	}
	if !strings.Contains(string(carrito), "/js/empresa_submenu_context.js") {
		t.Fatal("carrito must load the shared empresa context before operational mutations")
	}
}

func TestSuperPageToolsDoNotCoverMobileControls(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "web", "js", "super_page_tools.js"))
	if err != nil {
		t.Fatalf("read super page tools: %v", err)
	}
	content := string(raw)
	for _, required := range []string{
		"@media(max-width:560px)",
		".super-page-tools{position:static",
		"width:max-content",
		"justify-content:flex-end",
	} {
		if !strings.Contains(content, required) {
			t.Fatalf("mobile super page tools must stay in document flow; missing %q", required)
		}
	}
	if strings.Contains(content, "@media(max-width:560px){.super-page-tools{right:8px;bottom:8px}") {
		t.Fatal("mobile super page tools must not remain fixed over page controls")
	}
}

func TestEmpresaShellFullscreenPermissionHasNoDuplicateAttribute(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "web", "administrar_empresa.html"))
	if err != nil {
		t.Fatalf("read empresa shell: %v", err)
	}
	content := string(raw)
	if !strings.Contains(content, `allow="geolocation; fullscreen"`) {
		t.Fatal("empresa content iframe must keep the fullscreen permission policy")
	}
	if strings.Contains(content, `allow="geolocation; fullscreen" allowfullscreen`) {
		t.Fatal("empresa content iframe must not emit duplicate fullscreen attributes")
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
