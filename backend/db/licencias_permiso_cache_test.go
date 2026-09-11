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

func TestLicenciaPermissionRequiresDatabaseAndTenant(t *testing.T) {
	if _, err := GetLicenciaPermisoPolicyByEmpresa(nil, 71001); err == nil {
		t.Fatal("missing database must not be treated as an unrestricted license")
	}
}

func TestLicenciaPermissionPostgresRevocationAndSchemaFailure(t *testing.T) {
	conn := empresaRolesTestDB(t)
	if _, err := GetLicenciaPermisoPolicyByEmpresa(conn, 71001); err == nil {
		t.Fatal("missing license table silently accepted")
	}
	_, err := conn.Exec(`CREATE TABLE licencias(id BIGINT PRIMARY KEY, empresa_id BIGINT, nombre TEXT, modulos_habilitados TEXT, super_rol_habilitado INTEGER DEFAULT 0, activo INTEGER DEFAULT 1, fecha_inicio TEXT, fecha_fin TEXT);
		INSERT INTO licencias(id,empresa_id,nombre,modulos_habilitados) VALUES(1,71001,'QA','ventas'),(2,0,'Addon QA','inventario'),(3,71002,'Other tenant','seguridad')`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := GetLicenciaPermisoPolicyByEmpresa(conn, 71001); err == nil {
		t.Fatal("missing addon schema silently accepted")
	}
	_, err = conn.Exec(`CREATE TABLE empresa_licencias_adicionales(id BIGINT PRIMARY KEY, empresa_id BIGINT, licencia_id BIGINT, activo INTEGER DEFAULT 1, fecha_inicio TEXT, fecha_fin TEXT);
		INSERT INTO empresa_licencias_adicionales(id,empresa_id,licencia_id) VALUES(1,71001,2),(2,71001,3)`)
	if err != nil {
		t.Fatal(err)
	}
	policy, err := GetLicenciaPermisoPolicyByEmpresa(conn, 71001)
	if err != nil || policy == nil || policy.ModulosHabilitados != "ventas,inventario" {
		t.Fatalf("policy=%+v err=%v", policy, err)
	}
	if _, err := conn.Exec(`UPDATE empresa_licencias_adicionales SET activo = 0 WHERE empresa_id = 71001 AND id = 1`); err != nil {
		t.Fatal(err)
	}
	policy, err = GetLicenciaPermisoPolicyByEmpresa(conn, 71001)
	if err != nil || policy == nil || policy.ModulosHabilitados != "ventas" {
		t.Fatalf("revoked addon kept access: policy=%+v err=%v", policy, err)
	}
	if _, err := conn.Exec(`UPDATE licencias SET activo = 0 WHERE empresa_id = 71001`); err != nil {
		t.Fatal(err)
	}
	policy, err = GetLicenciaPermisoPolicyByEmpresa(conn, 71001)
	if err != nil || policy != nil {
		t.Fatalf("revoked base kept cached access: policy=%+v err=%v", policy, err)
	}
	if _, err := conn.Exec(`ALTER TABLE licencias DROP COLUMN modulos_habilitados`); err != nil {
		t.Fatal(err)
	}
	if _, err := GetLicenciaPermisoPolicyByEmpresa(conn, 71001); err == nil {
		t.Fatal("missing required column silently accepted")
	}
}
