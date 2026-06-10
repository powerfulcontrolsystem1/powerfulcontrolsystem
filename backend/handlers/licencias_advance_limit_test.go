package handlers

import (
	"testing"
	"time"
)

func TestLicenciaAdvancePurchaseBlocked(t *testing.T) {
	now := time.Date(2026, 6, 9, 10, 0, 0, 0, time.Local)

	blocked, windows := licenciaAdvancePurchaseBlocked(now, now.AddDate(0, 0, 365*3).Format("2006-01-02 15:04:05"), 365, 2)
	if !blocked {
		t.Fatalf("expected third active/future annual window to block new advance purchase; windows=%d", windows)
	}
	if windows != 3 {
		t.Fatalf("windows = %d, want 3", windows)
	}

	blocked, windows = licenciaAdvancePurchaseBlocked(now, now.AddDate(0, 0, 365*2).Format("2006-01-02 15:04:05"), 365, 2)
	if blocked {
		t.Fatalf("expected two annual windows to allow one more advance purchase; windows=%d", windows)
	}
	if windows != 2 {
		t.Fatalf("windows = %d, want 2", windows)
	}
}
