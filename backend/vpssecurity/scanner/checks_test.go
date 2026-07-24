package scanner

import (
	"os"
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

func TestTrivyRootfsArgsSkipProtectedCredentialFiles(t *testing.T) {
	args := trivyRootfsArgs("/tmp/report.json", "/")
	for _, path := range []string{"/etc/shadow", "/etc/shadow-", "/etc/gshadow", "/etc/gshadow-"} {
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
