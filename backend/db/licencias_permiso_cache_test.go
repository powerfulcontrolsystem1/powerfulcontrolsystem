package db

import (
	"testing"
	"time"
)

func TestInvalidateLicenciaPermisoPolicyCacheForEmpresa(t *testing.T) {
	licenciaPermisoPolicyCacheMu.Lock()
	licenciaPermisoPolicyCache = map[int64]cachedLicenciaPermisoPolicy{
		77: {
			Policy: &LicenciaPermisoPolicy{
				LicenciaID:         10,
				Nombre:             "Plan QA",
				ModulosHabilitados: "ventas",
			},
			LoadedAt: time.Now(),
		},
		88: {
			Policy:   nil,
			LoadedAt: time.Now(),
		},
	}
	licenciaPermisoPolicyCacheMu.Unlock()

	InvalidateLicenciaPermisoPolicyCacheForEmpresa(77)

	licenciaPermisoPolicyCacheMu.Lock()
	defer licenciaPermisoPolicyCacheMu.Unlock()
	if _, ok := licenciaPermisoPolicyCache[77]; ok {
		t.Fatal("expected empresa 77 licencia permission cache to be invalidated")
	}
	if _, ok := licenciaPermisoPolicyCache[88]; !ok {
		t.Fatal("expected unrelated empresa cache to remain untouched")
	}
}
