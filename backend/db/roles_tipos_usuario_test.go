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
