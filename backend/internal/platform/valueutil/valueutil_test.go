package valueutil

import (
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestValueNormalizers(t *testing.T) {
	if got := FirstNonBlank(" ", " valor "); got != "valor" {
		t.Fatalf("FirstNonBlank() = %q", got)
	}
	if got := HostWithoutPort("[::1]:8080"); got != "::1" {
		t.Fatalf("HostWithoutPort() = %q", got)
	}
	if !IsSafeSQLIdentifier("empresa_id") || IsSafeSQLIdentifier("empresa-id") {
		t.Fatal("IsSafeSQLIdentifier() no aplico el contrato")
	}
	if got := ParsePositiveInt64(" 12 "); got != 12 {
		t.Fatalf("ParsePositiveInt64() = %d", got)
	}
	if got := Truncate("área", 3); got != "áre" {
		t.Fatalf("Truncate() = %q", got)
	}
	if got := TrimmedPrefix(" 2026-08-26T10:00:00 ", 10); got != "2026-08-26" {
		t.Fatalf("TrimmedPrefix() = %q", got)
	}
	if got := UniqueSortedNonBlank([]string{" b ", "a", "b", ""}); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("UniqueSortedNonBlank() = %#v", got)
	}
	if !IsHexLength(strings.Repeat("a", 96), 96) || IsHexLength(strings.Repeat("z", 96), 96) {
		t.Fatal("IsHexLength() no aplico el contrato")
	}
	if got := NormalizeAllowed(" HOST ", "host", "viewer"); got != "host" {
		t.Fatalf("NormalizeAllowed() = %q", got)
	}
	if got := Clamp(120, 0, 100); got != 100 {
		t.Fatalf("Clamp() = %d", got)
	}
	if got := FirstPositive(float64(0), -1, 2.5); math.Abs(got-2.5) > 0.001 {
		t.Fatalf("FirstPositive() = %f", got)
	}
}

func TestMarshalJSONOrFallback(t *testing.T) {
	if got := MarshalJSONOr(map[string]int{"ok": 1}, "{}"); got != `{"ok":1}` {
		t.Fatalf("MarshalJSONOr() = %q", got)
	}
	if got := MarshalJSONOr(make(chan int), "{}"); got != "{}" {
		t.Fatalf("MarshalJSONOr() fallback = %q", got)
	}
}
