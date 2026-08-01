package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/you/pos-backend/vpssecurity/parser"
)

func TestShadowPermissionPolicyAllowsUbuntuDefault(t *testing.T) {
	for _, mode := range []os.FileMode{0o600, 0o640} {
		if mode&0o037 != 0 {
			t.Fatalf("mode %04o must be accepted for /etc/shadow", mode)
		}
	}
	for _, mode := range []os.FileMode{0o660, 0o644, 0o640 | 0o001} {
		if mode&0o037 == 0 {
			t.Fatalf("mode %04o must be rejected for /etc/shadow", mode)
		}
	}
}

func TestNginxConfigPathsIncludesExtensionlessEnabledSites(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"nginx.conf", "conf.d/security.conf", "sites-enabled/powerfulcontrolsystem", "sites-available/not-enabled"} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(path, []byte("# fixture\n"), 0o600); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	got := nginxConfigPaths(root)
	joined := strings.Join(got, "\n")
	if !strings.Contains(joined, filepath.Join("sites-enabled", "powerfulcontrolsystem")) {
		t.Fatalf("extensionless enabled site missing: %v", got)
	}
	if strings.Contains(joined, filepath.Join("sites-available", "not-enabled")) {
		t.Fatalf("disabled extensionless site must be ignored: %v", got)
	}
}

func TestTrivyRootfsArgsSkipProtectedCredentialFiles(t *testing.T) {
	args := trivyRootfsArgs("/tmp/report.json", "/")
	for _, path := range []string{"/etc/shadow", "/etc/shadow-", "/etc/gshadow", "/etc/gshadow-", "/lib/apk/db/lock"} {
		found := false
		for index := 0; index+1 < len(args); index++ {
			if args[index] == "--skip-files" && args[index+1] == path {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Trivy args do not skip %s: %#v", path, args)
		}
	}
	if got := args[len(args)-1]; got != "/" {
		t.Fatalf("Trivy target = %q, want /", got)
	}
}

func TestNmapLoopbackPortIsInternalInformation(t *testing.T) {
	raw := []byte(`<nmaprun><host><address addr="127.0.0.1"/><ports><port protocol="tcp" portid="8080"><state state="open"/><service name="http" product="PCS API"/></port></ports></host></nmaprun>`)
	findings, openPorts, _, err := parser.ParseNmapXML(raw, "127.0.0.1")
	if err != nil {
		t.Fatalf("parse Nmap XML: %v", err)
	}
	if len(findings) != 1 || len(openPorts) != 1 || openPorts[0] != 8080 {
		t.Fatalf("unexpected Nmap result: findings=%#v ports=%#v", findings, openPorts)
	}
	if findings[0].Severity != "INFO" {
		t.Fatalf("loopback severity = %q, want INFO", findings[0].Severity)
	}
	if !strings.Contains(strings.ToLower(findings[0].Title), "interno") {
		t.Fatalf("loopback title does not identify internal scope: %q", findings[0].Title)
	}
	if strings.Contains(strings.ToLower(findings[0].Title), "expuesto") {
		t.Fatalf("loopback title must not claim public exposure: %q", findings[0].Title)
	}
}
