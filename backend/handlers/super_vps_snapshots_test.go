package handlers

import (
	"path/filepath"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
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

func TestSanitizeSuperVPSSnapshotLogRemovesInternalDiagnostics(t *testing.T) {
	item := sanitizeSuperVPSSnapshotLog(dbpkg.SuperVPSSnapshotLog{
		FilePath:     "D:/private/snapshots/backup.tar.gz",
		ManifestJSON: `{"project_root":"D:/powerfulcontrolsystem"}`,
		Error:        "rclone failed with credential details",
		CloudMensaje: "provider diagnostic",
	})
	if item.FilePath != "" || item.ManifestJSON != "" || item.Error != "" || item.CloudMensaje != "" {
		t.Fatalf("snapshot response retained internal diagnostics: %#v", item)
	}
}

func TestSuperVPSSnapshotFailureMessageIsGeneric(t *testing.T) {
	message := superVPSSnapshotFailureMessage()
	if strings.Contains(strings.ToLower(message), "rclone") || strings.Contains(strings.ToLower(message), "error:") {
		t.Fatalf("snapshot public error must not expose an internal diagnostic: %q", message)
	}
}
