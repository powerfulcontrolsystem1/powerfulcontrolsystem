package handlers

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestNewRoleDefaultsDenyEveryModuleIncludingVida(t *testing.T) {
	for _, role := range []string{"rol_futuro", "mesero", "", "sin_rol"} {
		for _, row := range buildPermissionModuleMatrixForRole(role) {
			for action, allowed := range row.Acciones {
				if allowed {
					t.Fatalf("unconfigured role %q implicitly received %s:%s", role, row.Modulo, action)
				}
			}
		}
	}
	if !roleAllowsModuleAction("vendedor", permModuleVida, permActionRead) {
		t.Fatal("known role defaults should retain their existing personal module policy")
	}
}

func TestOperationalPermissionsNeverRestoreDeniedActions(t *testing.T) {
	for _, role := range []string{"cajero", "portero", "contador", "empresario", "servicio_limpieza", "tecnico_solar", "jefe_bodega", "responsable_bodega", "recursos_humanos"} {
		t.Run(role, func(t *testing.T) {
			rows := buildPermissionModuleMatrixForRole(role)
			for idx := range rows {
				for _, action := range permissionActionsCatalogOrdered {
					setPermissionActionOnModuleRow(&rows[idx], action, false)
				}
			}
			for _, row := range restrictPermissionModuleRowsForOperationalRole(role, rows) {
				for action, allowed := range row.Acciones {
					if allowed {
						t.Fatalf("%s restored denied %s:%s", role, row.Modulo, action)
					}
				}
			}
		})
	}
}

func TestRolePagePresetsRespectLicenseAndCompanyDenials(t *testing.T) {
	rolePages := map[string]bool{"linkEstaciones": true, "linkVentaDirecta": true, "linkProductos": true}
	rows := buildPermissionModuleMatrixForRole("admin_empresa")
	rows = applyLicenciaRestriccionesToModuleRows(rows, map[string]bool{permModuleInventario: true})
	pages := intersectRolePagesWithModuleRows(rolePages, rows)
	if pages["linkEstaciones"] || pages["linkVentaDirecta"] || !pages["linkProductos"] {
		t.Fatalf("navigation disregarded license: %+v", pages)
	}
	pages = applyEmpresaPageRestrictionsToMap(pages, map[string]bool{"linkProductos": false})
	if pages["linkProductos"] {
		t.Fatal("company denial must override the role preset")
	}
	if !rolePages["linkProductos"] {
		t.Fatal("company B must not mutate company A role template")
	}
}

func TestLicenseCannotPromoteSupervisorToAdministrator(t *testing.T) {
	for _, policy := range []*dbpkg.LicenciaPermisoPolicy{nil, {SuperRolHabilitado: false}, {SuperRolHabilitado: true}} {
		role := resolveEffectiveRoleByLicencia("supervisor_sucursal", policy)
		if role != "supervisor_sucursal" || roleAllowsModuleAction(role, permModuleSeguridad, permActionUpdate) {
			t.Fatalf("license promoted a supervisor to %s", role)
		}
	}
}

func TestTypedEmpresaIdentityRejectsCrossTenantAndIncompletePrincipal(t *testing.T) {
	for _, tc := range []struct {
		name                                          string
		principalID, sessionEmpresa, requestedEmpresa int64
	}{
		{"foreign company", 5, 12, 99}, {"missing principal", 0, 12, 12}, {"missing company", 5, 0, 12},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/empresa/productos", nil)
			ctx := context.WithValue(r.Context(), "sessionPrincipalType", "empresa_usuario")
			ctx = context.WithValue(ctx, "sessionPrincipalID", tc.principalID)
			ctx = context.WithValue(ctx, "sessionEmpresaID", tc.sessionEmpresa)
			if _, err := getEmpresaPermissionSnapshotForRequest(r.WithContext(ctx), nil, nil, "qa@example.invalid", tc.requestedEmpresa); !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("invalid typed identity should fail before database access: %v", err)
			}
		})
	}
}

func TestAuthorizationDoesNotReusePriorRequestSnapshot(t *testing.T) {
	const key = "qa@example.invalid|12"
	empresaPermissionCacheMu.Lock()
	empresaPermissionCache[key] = empresaPermissionSnapshot{EmpresaID: 12, AdminEmail: "qa@example.invalid", CanAccess: true, LoadedAt: time.Now()}
	empresaPermissionCacheMu.Unlock()
	t.Cleanup(func() {
		empresaPermissionCacheMu.Lock()
		delete(empresaPermissionCache, key)
		empresaPermissionCacheMu.Unlock()
	})
	if _, err := getEmpresaPermissionSnapshot(nil, nil, "qa@example.invalid", 12); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("a stale process snapshot must not authenticate a new request: %v", err)
	}
}

func TestPermissionsContextUsesAuthorizedSnapshotAndProtectsMatrix(t *testing.T) {
	snapshot := empresaPermissionSnapshot{EmpresaID: 12, AdminEmail: "qa@example.invalid", AdminRole: "cajero", EffectiveRole: "cajero", CanAccess: true,
		ModuleRows: buildPermissionModuleMatrixForRole("cajero"), AllowedPages: map[string]bool{"linkEstaciones": false}, RoleModuleActions: map[string]bool{}}
	for _, tc := range []struct {
		query  string
		status int
	}{{"empresa_id=12", 200}, {"empresa_id=12&include_matrix=1", 403}, {"empresa_id=99", 403}} {
		r := httptest.NewRequest(http.MethodGet, "/api/empresa/permisos_contexto?"+tc.query, nil)
		r = r.WithContext(context.WithValue(r.Context(), empresaPermissionSnapshotContextKey{}, snapshot))
		w := httptest.NewRecorder()
		EmpresaPermisosContextoHandler(nil)(w, r)
		if w.Code != tc.status {
			t.Fatalf("%s status=%d want=%d", tc.query, w.Code, tc.status)
		}
		if w.Code == http.StatusOK {
			var actual empresaPermisosContextResponse
			if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
				t.Fatal(err)
			}
			if actual.RolEfectivo != "cajero" || actual.Paginas["linkEstaciones"] {
				t.Fatal("context recomputed or regranted denied access")
			}
		}
	}
}

type permissionsUnavailableDriver struct{}
type permissionsUnavailableConn struct{}

var permissionsUnavailableRegister sync.Once

func (permissionsUnavailableDriver) Open(string) (driver.Conn, error) {
	return permissionsUnavailableConn{}, nil
}
func (permissionsUnavailableConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("authorization store unavailable")
}
func (permissionsUnavailableConn) Close() error { return nil }
func (permissionsUnavailableConn) Begin() (driver.Tx, error) {
	return nil, errors.New("authorization store unavailable")
}
func (permissionsUnavailableConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("authorization store unavailable")
}

func TestUnavailableAuthorizationStoreFailsClosed(t *testing.T) {
	permissionsUnavailableRegister.Do(func() { sql.Register("pcs_permissions_unavailable", permissionsUnavailableDriver{}) })
	conn, err := sql.Open("pcs_permissions_unavailable", "")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, _, _, err := loadEmpresaPermissionOverridesStrict(conn, 12); err == nil {
		t.Fatal("tenant restrictions must not fail open")
	}
	if _, _, _, err := loadEmpresaRolePermissionMatrix(conn, 12, 0, "admin_empresa"); err == nil {
		t.Fatal("role grants must not fail open")
	}
	if _, _, _, err := loadEmpresaRolePermissionMatrix(conn, 12, 123, "admin_empresa"); err == nil {
		t.Fatal("assigned role errors must not fall back to global admin")
	}
}
