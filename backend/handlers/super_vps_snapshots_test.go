package handlers

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeSuperVPSSnapshotRemotePath(t *testing.T) {
	got := sanitizeSuperVPSSnapshotRemotePath("gdrive:PCS/backups\r\n--delete")
	if strings.ContainsAny(got, "\r\n") {
		t.Fatalf("remote path should not contain line breaks: %q", got)
	}
	if !strings.Contains(got, "gdrive:PCS/backups") {
		t.Fatalf("remote path lost expected prefix: %q", got)
	}
}

func TestSafeSuperVPSSnapshotPath(t *testing.T) {
	inside := filepath.Join(superVPSSnapshotDir(), "pcs-vps-snapshot-test.tar.gz")
	if _, ok := safeSuperVPSSnapshotPath(inside); !ok {
		t.Fatalf("expected path inside snapshot dir to be accepted")
	}
	outside := filepath.Join(resolveProjectRootDir(), "deploy", ".env.platform")
	if _, ok := safeSuperVPSSnapshotPath(outside); ok {
		t.Fatalf("expected path outside snapshot dir to be rejected")
	}
}
