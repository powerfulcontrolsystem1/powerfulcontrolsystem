package db

import (
	"database/sql"
	"errors"
	"testing"
)

func TestNormalizeRolCatalogKeyCajaDeduplicaComoCajero(t *testing.T) {
	for _, raw := range []string{"Caja", "Caja principal", "caja_turno", "Cajero"} {
		if got := normalizeRolCatalogKey(raw); got != "cajero" {
			t.Fatalf("normalizeRolCatalogKey(%q)=%q, want cajero", raw, got)
		}
	}
}

func TestNormalizeRolCatalogPreservesDifferentOperationalCapabilities(t *testing.T) {
	if normalizeRolCatalogKey("tecnico") == normalizeRolCatalogKey("tecnico_solar") {
		t.Fatal("generic and solar technician must remain distinct roles")
	}
	if normalizeRolCatalogKey("bodega") != "responsable_bodega" {
		t.Fatal("warehouse alias must match the authorization engine")
	}
	if normalizeRolCatalogKey("responsable_bodega") == normalizeRolCatalogKey("jefe_bodega") {
		t.Fatal("warehouse responsibility roles must preserve distinct IDs")
	}
}

func TestRolesCatalogPrefersCompanyTypeWithoutChangingPermissionIDs(t *testing.T) {
	roles := []RolDeUsuario{
		{ID: 1, TipoEmpresaID: 1, TipoEmpresaNombre: "Motel", Nombre: "Cajero", Estado: "activo"},
		{ID: 90, TipoEmpresaID: 2, TipoEmpresaNombre: "Taller", Nombre: "Caja", Estado: "activo"},
		{ID: 91, TipoEmpresaID: 0, Nombre: "Cajero", Estado: "activo"},
	}
	selected := selectRolesGlobalesParaTipo(roles, 2, false)
	if len(selected) != 1 || selected[0].ID != 90 || selected[0].TipoEmpresaID != 2 || selected[0].Nombre != "Caja" {
		t.Fatalf("type-specific role ID or metadata replaced: %+v", selected)
	}
	selected = selectRolesGlobalesParaTipo(roles, 3, false)
	if len(selected) != 1 || selected[0].ID != 91 {
		t.Fatalf("real universal role not preferred: %+v", selected)
	}
	selected = selectRolesGlobalesParaTipo(roles[:2], 3, false)
	if len(selected) != 2 || selected[0].NombreVisible == selected[1].NombreVisible {
		t.Fatalf("foreign matrices were collapsed without explicit choice: %+v", selected)
	}
}

func TestRolesCatalogPreservesDuplicateIDsAndRejectsInactivePreference(t *testing.T) {
	roles := []RolDeUsuario{
		{ID: 10, TipoEmpresaID: 2, TipoEmpresaNombre: "Taller", Nombre: "Cajero", Estado: "inactivo"},
		{ID: 11, TipoEmpresaID: 0, Nombre: "Cajero", Estado: "activo"},
		{ID: 12, TipoEmpresaID: 2, TipoEmpresaNombre: "Taller", Nombre: "Supervisor", Estado: "activo"},
		{ID: 13, TipoEmpresaID: 2, TipoEmpresaNombre: "Taller", Nombre: "supervisor_sucursal", Estado: "activo"},
	}
	selected := selectRolesGlobalesParaTipo(roles, 2, true)
	ids := map[int64]RolDeUsuario{}
	for _, rol := range selected {
		ids[rol.ID] = rol
	}
	if len(ids) != 3 || ids[11].ID != 11 || ids[12].ID != 12 || ids[13].ID != 13 {
		t.Fatalf("active universal or distinct local matrix lost: %+v", selected)
	}
	if ids[12].NombreVisible == ids[13].NombreVisible {
		t.Fatal("local duplicate variants need explicit names")
	}
}

func TestEmpresaRolAsignableRejectsPlatformAndInactive(t *testing.T) {
	for _, nombre := range []string{"super_administrador", "Super Administrador", "superadmin", "super", "administrador_total"} {
		if IsRolDeUsuarioAsignable(&RolDeUsuario{Nombre: nombre, Estado: "activo"}) {
			t.Fatalf("platform role %q is assignable", nombre)
		}
	}
	if IsRolDeUsuarioAsignable(nil) || IsRolDeUsuarioAsignable(&RolDeUsuario{Nombre: "cajero", Estado: "inactivo"}) {
		t.Fatal("inactive role is assignable")
	}
	if !IsRolDeUsuarioAsignable(&RolDeUsuario{Nombre: "Coordinador de sucursal nuevo", Estado: "activo"}) {
		t.Fatal("custom role names must not require a whitelist")
	}
}

func TestEmpresaRolScopesRejectMissingTenantBeforeDatabase(t *testing.T) {
	for _, ids := range [][2]int64{{0, 1}, {-1, 1}, {1, 0}} {
		if _, err := GetRolDeUsuarioByIDEmpresaScope(nil, ids[0], ids[1]); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("global/custom scope accepted missing identity: %v", err)
		}
		if _, err := GetRolDeUsuarioEmpresaByID(nil, ids[0], ids[1]); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("own scope accepted missing identity: %v", err)
		}
		if err := SetEmpresaRolDeUsuarioEstado(nil, ids[0], ids[1], "activo"); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("state mutation accepted missing identity: %v", err)
		}
	}
}

func TestRolPermisosInputRejectsDuplicateAndForeignRole(t *testing.T) {
	cases := []struct {
		modulos []RolPermisoModulo
		paginas []RolPermisoPagina
	}{
		{modulos: []RolPermisoModulo{{RolID: 2, Modulo: "ventas", Accion: "R"}}},
		{modulos: []RolPermisoModulo{{Modulo: "ventas", Accion: "R"}, {Modulo: " VENTAS ", Accion: "r"}}},
		{modulos: []RolPermisoModulo{{Modulo: "ventas", Accion: "DROP"}}},
		{paginas: []RolPermisoPagina{{RolID: 2, PaginaClave: "linkEstaciones"}}},
		{paginas: []RolPermisoPagina{{PaginaClave: "linkEstaciones"}, {PaginaClave: " linkEstaciones "}}},
	}
	for _, tc := range cases {
		if err := validateRolPermisosInput(1, tc.modulos, tc.paginas); err == nil {
			t.Fatal("invalid matrix accepted")
		}
	}
	if err := validateRolPermisosInput(1, []RolPermisoModulo{{Modulo: "new_module", Accion: "R"}}, nil); err != nil {
		t.Fatalf("DB must permit catalog expansion: %v", err)
	}
}

func TestPreferredRolCatalogDisplayRankPrefiereCajero(t *testing.T) {
	if preferredRolCatalogDisplayRank("cajero", "Cajero") >= preferredRolCatalogDisplayRank("cajero", "Caja principal") {
		t.Fatal("el catalogo global debe preferir mostrar Cajero antes que variantes como Caja principal")
	}
}
