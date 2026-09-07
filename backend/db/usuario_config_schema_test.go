package db

import (
	"reflect"
	"testing"
)

func TestNormalizeUsuarioSelectorEmpresaIDs(t *testing.T) {
	got := normalizeUsuarioSelectorEmpresaIDs([]int64{7, 0, -1, 3, 7, 44, 3})
	want := []int64{7, 3, 44}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("orden normalizado inesperado: got %v want %v", got, want)
	}
}

func TestRemoveEmpresaIDFromSelectorOrderRaw(t *testing.T) {
	got, changed := removeEmpresaIDFromSelectorOrderRaw(`[7,44,7,12,44]`, 44)
	if !changed {
		t.Fatalf("expected selector order to change")
	}
	if got != `[7,12]` {
		t.Fatalf("orden depurado inesperado: got %s", got)
	}
}

func TestRemoveEmpresaIDFromSelectorOrderRawNoop(t *testing.T) {
	got, changed := removeEmpresaIDFromSelectorOrderRaw(`[7,12]`, 44)
	if changed {
		t.Fatalf("expected selector order to remain unchanged")
	}
	if got != `[7,12]` {
		t.Fatalf("orden inesperado: got %s", got)
	}
}
