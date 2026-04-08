package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestRolesDeUsuarioPermisosHandlerGetAplicaOverrides(t *testing.T) {
	dbSuper := openPermsTestDB(t, "super_roles_permisos_get.db")
	ensurePermsRoleConfigSchema(t, dbSuper)
	rolID := seedPermsRolDeUsuario(t, dbSuper, "contabilidad")

	err := dbpkg.ReplaceRolPermisosDeUsuario(
		dbSuper,
		rolID,
		[]dbpkg.RolPermisoModulo{
			{RolID: rolID, Modulo: permModuleFinanzas, Accion: permActionApprove, Permitido: false},
			{RolID: rolID, Modulo: permModuleSeguridad, Accion: permActionRead, Permitido: true},
		},
		[]dbpkg.RolPermisoPagina{
			{RolID: rolID, PaginaClave: "linkFinanzas", Permitido: false},
			{RolID: rolID, PaginaClave: "linkReportes", Permitido: true},
		},
		"qa",
	)
	if err != nil {
		t.Fatalf("seed ReplaceRolPermisosDeUsuario: %v", err)
	}

	h := RolesDeUsuarioPermisosHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/super/api/roles_de_usuario/permisos?rol_id="+strconv.FormatInt(rolID, 10), nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status=200, got=%d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		RolID   int64                       `json:"rol_id"`
		Modulos []permissionModuleMatrixRow `json:"modulos"`
		Paginas []permissionPageAccessRow   `json:"paginas"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if resp.RolID != rolID {
		t.Fatalf("expected rol_id=%d got=%d", rolID, resp.RolID)
	}

	finanzas, ok := findPermissionModuleRow(resp.Modulos, permModuleFinanzas)
	if !ok {
		t.Fatalf("finanzas module must be returned")
	}
	if finanzas.Approve {
		t.Fatalf("expected finanzas approve=false by override")
	}

	finanzasPage, ok := findPagePermissionRow(resp.Paginas, "linkFinanzas")
	if !ok {
		t.Fatalf("linkFinanzas page must be returned")
	}
	if finanzasPage.Permitido {
		t.Fatalf("expected linkFinanzas permitido=false by override")
	}
}

func TestRolesDeUsuarioPermisosHandlerPutGuardaReglas(t *testing.T) {
	dbSuper := openPermsTestDB(t, "super_roles_permisos_put.db")
	ensurePermsRoleConfigSchema(t, dbSuper)
	rolID := seedPermsRolDeUsuario(t, dbSuper, "cajero")

	h := RolesDeUsuarioPermisosHandler(dbSuper)
	body := `{
		"rol_id": ` + strconv.FormatInt(rolID, 10) + `,
		"permisos_modulo": [
			{"modulo": "finanzas", "accion": "A", "permitido": false},
			{"modulo": "ventas", "accion": "C", "permitido": true}
		],
		"permisos_pagina": [
			{"pagina_clave": "linkFinanzas", "permitido": false},
			{"pagina_clave": "linkVentas", "permitido": true}
		]
	}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/roles_de_usuario/permisos", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status=204, got=%d body=%s", rr.Code, rr.Body.String())
	}

	modulos, err := dbpkg.ListRolPermisosModuloByRolID(dbSuper, rolID)
	if err != nil {
		t.Fatalf("ListRolPermisosModuloByRolID: %v", err)
	}
	if len(modulos) != 2 {
		t.Fatalf("expected 2 modulo rules, got %d", len(modulos))
	}

	found, permitido, err := dbpkg.LookupRolPermisoModuloByRolID(dbSuper, rolID, "finanzas", "A")
	if err != nil {
		t.Fatalf("LookupRolPermisoModuloByRolID finanzas/A: %v", err)
	}
	if !found || permitido {
		t.Fatalf("expected finanzas/A override found=true permitido=false, got found=%v permitido=%v", found, permitido)
	}

	paginas, err := dbpkg.ListRolPermisosPaginaByRolID(dbSuper, rolID)
	if err != nil {
		t.Fatalf("ListRolPermisosPaginaByRolID: %v", err)
	}
	if len(paginas) != 2 {
		t.Fatalf("expected 2 page rules, got %d", len(paginas))
	}

	found, permitido, err = dbpkg.LookupRolPermisoPaginaByRolID(dbSuper, rolID, "linkFinanzas")
	if err != nil {
		t.Fatalf("LookupRolPermisoPaginaByRolID linkFinanzas: %v", err)
	}
	if !found || permitido {
		t.Fatalf("expected linkFinanzas override found=true permitido=false, got found=%v permitido=%v", found, permitido)
	}
}

func TestRolesDeUsuarioPermisosHandlerGetRolNoEncontrado(t *testing.T) {
	dbSuper := openPermsTestDB(t, "super_roles_permisos_404.db")
	ensurePermsRoleConfigSchema(t, dbSuper)

	h := RolesDeUsuarioPermisosHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/super/api/roles_de_usuario/permisos?rol_id=999", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status=404, got=%d body=%s", rr.Code, rr.Body.String())
	}
}

func findPagePermissionRow(rows []permissionPageAccessRow, page string) (permissionPageAccessRow, bool) {
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.PaginaClave), strings.TrimSpace(page)) {
			return row, true
		}
	}
	return permissionPageAccessRow{}, false
}
