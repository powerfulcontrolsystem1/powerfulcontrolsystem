package db

import (
	"testing"
	"time"
)

func TestEmpresaDeleteBuildEmpresaIDPredicateUsesTextArgForTextColumns(t *testing.T) {
	predicate, arg := empresaDeleteBuildEmpresaIDPredicate(empresaDeleteCandidateTable{
		Name:              "empresa_configuracion_avanzada",
		EmpresaIDDataType: "text",
		EmpresaIDUDTName:  "text",
	}, 29)

	if predicate != "TRIM(empresa_id) = ?" {
		t.Fatalf("expected text predicate, got %q", predicate)
	}
	if value, ok := arg.(string); !ok || value != "29" {
		t.Fatalf("expected string empresa id arg 29, got %#v", arg)
	}
}

func TestEmpresaDeleteBuildEmpresaIDPredicateUsesNumericArgForNumericColumns(t *testing.T) {
	predicate, arg := empresaDeleteBuildEmpresaIDPredicate(empresaDeleteCandidateTable{
		Name:              "productos",
		EmpresaIDDataType: "bigint",
		EmpresaIDUDTName:  "int8",
	}, 29)

	if predicate != "empresa_id = ?" {
		t.Fatalf("expected numeric predicate, got %q", predicate)
	}
	if value, ok := arg.(int64); !ok || value != 29 {
		t.Fatalf("expected int64 empresa id arg 29, got %#v", arg)
	}
}

func TestEmpresaDeleteEmpresaIDIsTextColumnCoversPostgresAliases(t *testing.T) {
	textColumns := []empresaDeleteCandidateTable{
		{EmpresaIDDataType: "character varying"},
		{EmpresaIDDataType: "character"},
		{EmpresaIDUDTName: "varchar"},
		{EmpresaIDUDTName: "bpchar"},
	}

	for _, column := range textColumns {
		if !empresaDeleteEmpresaIDIsTextColumn(column) {
			t.Fatalf("expected column metadata %#v to be treated as text", column)
		}
	}
}

func TestInvalidateEmpresaByScopeCacheForEmpresaRemovesAllIDs(t *testing.T) {
	empresaByScopeCacheMu.Lock()
	previous := empresaByScopeCache
	empresaByScopeCache = map[int64]cachedEmpresaByScope{
		7:  {LoadedAt: time.Now()},
		44: {LoadedAt: time.Now()},
		99: {LoadedAt: time.Now()},
	}
	empresaByScopeCacheMu.Unlock()
	t.Cleanup(func() {
		empresaByScopeCacheMu.Lock()
		empresaByScopeCache = previous
		empresaByScopeCacheMu.Unlock()
	})

	InvalidateEmpresaByScopeCacheForEmpresa(7, 44)

	empresaByScopeCacheMu.Lock()
	defer empresaByScopeCacheMu.Unlock()
	if _, ok := empresaByScopeCache[7]; ok {
		t.Fatalf("expected physical empresa id cache to be removed")
	}
	if _, ok := empresaByScopeCache[44]; ok {
		t.Fatalf("expected logical empresa id cache to be removed")
	}
	if _, ok := empresaByScopeCache[99]; !ok {
		t.Fatalf("expected unrelated empresa cache to remain")
	}
}

func TestInvalidateAdminEmpresaCompartidaAccessCacheForEmpresa(t *testing.T) {
	adminEmpresaCompartidaAccessCacheMu.Lock()
	previous := adminEmpresaCompartidaAccessCache
	adminEmpresaCompartidaAccessCache = map[string]cachedAdminEmpresaCompartidaAccess{
		adminEmpresaCompartidaAccessCacheKey(7, "uno@example.com"):  {LoadedAt: time.Now()},
		adminEmpresaCompartidaAccessCacheKey(7, "dos@example.com"):  {LoadedAt: time.Now()},
		adminEmpresaCompartidaAccessCacheKey(8, "tres@example.com"): {LoadedAt: time.Now()},
	}
	adminEmpresaCompartidaAccessCacheMu.Unlock()
	t.Cleanup(func() {
		adminEmpresaCompartidaAccessCacheMu.Lock()
		adminEmpresaCompartidaAccessCache = previous
		adminEmpresaCompartidaAccessCacheMu.Unlock()
	})

	InvalidateAdminEmpresaCompartidaAccessCacheForEmpresa(7)

	adminEmpresaCompartidaAccessCacheMu.Lock()
	defer adminEmpresaCompartidaAccessCacheMu.Unlock()
	if _, ok := adminEmpresaCompartidaAccessCache[adminEmpresaCompartidaAccessCacheKey(7, "uno@example.com")]; ok {
		t.Fatalf("expected first shared access cache to be removed")
	}
	if _, ok := adminEmpresaCompartidaAccessCache[adminEmpresaCompartidaAccessCacheKey(7, "dos@example.com")]; ok {
		t.Fatalf("expected second shared access cache to be removed")
	}
	if _, ok := adminEmpresaCompartidaAccessCache[adminEmpresaCompartidaAccessCacheKey(8, "tres@example.com")]; !ok {
		t.Fatalf("expected unrelated shared access cache to remain")
	}
}
