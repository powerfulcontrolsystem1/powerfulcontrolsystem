package scanner

import (
	"os"
	"testing"
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
