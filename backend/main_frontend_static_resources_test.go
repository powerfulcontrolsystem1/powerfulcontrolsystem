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

func TestPrometheusMetricsAreNotExposedByPublicFrontend(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "nginx", "pcs.conf"))
	if err != nil {
		t.Fatalf("read frontend nginx config: %v", err)
	}
	config := string(raw)
	block := regexp.MustCompile(`(?s)location\s*=\s*/metrics\s*\{[^}]*return\s+404;[^}]*\}`)
	if !block.MatchString(config) {
		t.Fatal("public frontend must reject /metrics; Prometheus scrapes the backend over the private Docker network")
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
	reportOnlyHeader := ""
	for _, line := range strings.Split(staticHeaders, "\n") {
		if strings.HasPrefix(line, "add_header Content-Security-Policy-Report-Only ") {
			reportOnlyHeader = line
			break
		}
	}
	if reportOnlyHeader == "" {
		t.Fatal("static frontend must define a report-only CSP header")
	}
	for _, forbidden := range []string{
		`img-src 'self' data: blob: https:;`,
		`style-src 'self' 'unsafe-inline'`,
		`script-src 'self' 'unsafe-inline'`,
	} {
		if strings.Contains(reportOnlyHeader, forbidden) {
			t.Fatalf("static report-only CSP must keep explicit origins and omit inline compatibility: %q", forbidden)
		}
	}
	for _, required := range []string{
		`Content-Security-Policy-Report-Only`,
		`img-src 'self' data: blob: https://lh3.googleusercontent.com`,
		`style-src 'self' https://unpkg.com https://fonts.googleapis.com`,
		`script-src 'self' https://accounts.google.com`,
		`connect-src 'self' https://api.openai.com`,
	} {
		if !strings.Contains(reportOnlyHeader, required) {
			t.Fatalf("static strict report-only CSP is missing %q", required)
		}
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
	for _, required := range []string{
		`CERT_NAME="${CERT_NAME:-pcs-staging}"`,
		`openssl x509 -in "$CERT_FILE" -noout -checkend 0`,
		`ssl_certificate $CERT_FILE;`,
		`ssl_certificate_key $CERT_KEY;`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("staging certificate contract missing %q", required)
		}
	}
	if strings.Contains(script, "powerfulcontrolsystem.com-0001") {
		t.Fatal("staging edge must not restore the expired wildcard certificate")
	}
}

func TestStagingDigestPromotionRequiresAllExactImagesBeforeRecreate(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "scripts", "vps-staging-digest-up.sh"))
	if err != nil {
		t.Fatalf("read staging digest promotion script: %v", err)
	}
	script := string(raw)
	for _, required := range []string{
		"PLATFORM_COMPOSE_FILE",
		"STAGING_COMPOSE_FILE",
		"RELEASE_COMPOSE_FILE",
		`config --images`,
		`grep -Fqx "$image"`,
		`up -d --no-build postgres migrate backend worker frontend`,
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("staging digest promotion must enforce %q", required)
		}
	}
	if strings.Contains(script, `"${compose[@]}" up -d --no-build`+"\n") {
		t.Fatal("staging digest promotion must not recreate the entire platform stack")
	}
}

func TestOperationalVPSBackupRequiresPrivateStorage(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "deploy", "scripts", "vps-backup-operacion.sh"))
	if err != nil {
		t.Fatalf("read operational VPS backup script: %v", err)
	}
	script := string(raw)
	for _, required := range []string{
		"powerful-control-system_pcs_private_storage",
		"powerful-control-system_pcs_private_storage.tar.gz",
		"powerful-control-system_mailu_certs.tar.gz",
		"powerful-control-system_pcs_onlyoffice_data.tar.gz",
		"powerful-control-system_pcs_onlyoffice_lib.tar.gz",
		"powerful-control-system_pcs_onlyoffice_logs.tar.gz",
		"[ERROR] Backup VPS incompleto",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("operational VPS backup must require private storage artifact %q", required)
		}
	}
}

func TestOperationalVPSRestoreCanVerifyCriticalTenantDataAndPrivateChecksums(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "scripts", "vps_restore_validation.ps1"))
	if err != nil {
		t.Fatalf("read operational VPS restore script: %v", err)
	}
	script := string(raw)
	for _, required := range []string{
		"VerifyCriticalData",
		"VerifyCriticalData requiere ExecuteDrill",
		"pcs_empresas','pcs_superadministrador",
		"empresa_cuentas_por_pagar empresa_asientos_contables empresa_ai_memoria empresa_dian_configuracion empresa_documentos_gestion",
		"WHERE empresa_id=12",
		"empresa_soportes_compras_ia",
		"private://soportes_compras_ia/empresa_",
		"sha256sum",
		"trap cleanup EXIT",
	} {
		if !strings.Contains(script, required) {
			t.Fatalf("operational VPS restore critical audit must enforce %q", required)
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

func TestEmpresaPagesWithMutatingFetchInstallCSRFSynchronizer(t *testing.T) {
	root := filepath.Join("..", "web", "administrar_empresa")
	mutatingFetch := regexp.MustCompile(`(?s)fetch\s*\(.{0,800}?method\s*:\s*["'](?:POST|PUT|PATCH|DELETE)["']`)
	missing := make([]string, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || !strings.EqualFold(filepath.Ext(path), ".html") {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		content := string(raw)
		if !mutatingFetch.MatchString(content) {
			return nil
		}
		usesSharedSynchronizer := strings.Contains(content, "/js/empresa_submenu_context.js")
		usesExplicitToken := strings.Contains(content, "X-CSRF-Token") && strings.Contains(content, "pcs_csrf")
		if !usesSharedSynchronizer && !usesExplicitToken {
			relative, relErr := filepath.Rel(root, path)
			if relErr != nil {
				relative = path
			}
			missing = append(missing, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan empresa HTML pages for CSRF coverage: %v", err)
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("empresa pages with mutating fetch must install CSRF support:\n%s", strings.Join(missing, "\n"))
	}
}

func TestPlan108FullSweepFrontendRegressions(t *testing.T) {
	root := filepath.Clean("..")
	products, err := os.ReadFile(filepath.Join(root, "web", "administrar_empresa", "administrar_productos.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, corrupted := range []string{
		"/api/empresa/inventario/resumenú",
		"/api/empresa/inventario/plan_reposicion_resumenú",
	} {
		if strings.Contains(string(products), corrupted) {
			t.Fatalf("products keeps corrupted query separator %q", corrupted)
		}
	}

	moduleScript, err := os.ReadFile(filepath.Join(root, "web", "js", "modulo_colombia_admin.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(moduleScript), "if (!state.modulo)") || !strings.Contains(string(moduleScript), "Selecciona un módulo") {
		t.Fatal("generic Colombia module entry must stop before calling an empty /api/empresa/ route")
	}

	for _, rel := range []string{
		"venta_publica.html",
		filepath.Join("administrar_empresa", "alquileres.html"),
		filepath.Join("administrar_empresa", "domicilios.html"),
		filepath.Join("administrar_empresa", "taxi_system.html"),
		filepath.Join("administrar_empresa", "ubicacion_gps.html"),
		"taxi_system.html",
		"taxi_system_conductor.html",
	} {
		content, readErr := os.ReadFile(filepath.Join(root, "web", rel))
		if readErr != nil {
			t.Fatal(readErr)
		}
		if strings.Contains(string(content), "unpkg.com/leaflet@1.9.4") && (!strings.Contains(string(content), "sha256-p4NxAoJBhIIN+") || !strings.Contains(string(content), "sha256-20nQCchB9co0qIjJZRGuk2/")) {
			t.Fatalf("%s must pin Leaflet CSS and JavaScript with SRI", rel)
		}
	}

	chartPage, err := os.ReadFile(filepath.Join(root, "web", "super", "administrar_base_de_datos.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(chartPage), "chart.js@4.5.1/dist/chart.umd.min.js") || !strings.Contains(string(chartPage), "integrity=\"sha384-") {
		t.Fatal("PostgreSQL dashboard must pin Chart.js with SRI")
	}

	staticHeaders, err := os.ReadFile(filepath.Join(root, "deploy", "nginx", "pcs-static-security-headers.inc"))
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range []string{
		"https://unpkg.com",
		"https://cdn.jsdelivr.net",
		"https://fonts.googleapis.com",
		"https://fonts.gstatic.com",
	} {
		if !strings.Contains(string(staticHeaders), origin) {
			t.Fatalf("frontend CSP must allow the pinned visual resource origin %s", origin)
		}
	}
	if !strings.Contains(string(staticHeaders), "font-src 'self' data:") {
		t.Fatal("frontend CSP must declare an explicit font-src")
	}

	domicilios, err := os.ReadFile(filepath.Join(root, "web", "js", "domicilios.js"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(domicilios), "function asArray(v)") || !strings.Contains(string(domicilios), "state.menu=asArray(menuData)") {
		t.Fatal("Domicilios must render an empty menu response as an empty list")
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
