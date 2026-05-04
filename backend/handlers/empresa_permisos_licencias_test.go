package handlers

import "testing"

func TestLicenciaModulosCSVControlsModuleAccess(t *testing.T) {
	allowed, ordered := parseLicenciaModulosCSV(" ventas, SEGURIDAD, modulo_desconocido, ventas ")
	if len(ordered) != 2 {
		t.Fatalf("expected only known unique modules, got %v", ordered)
	}
	if !isModuloPermitidoByLicencia(permModuleVentas, allowed) {
		t.Fatal("expected ventas to be enabled by licencia")
	}
	if !isModuloPermitidoByLicencia(permModuleSeguridad, allowed) {
		t.Fatal("expected seguridad to be enabled by licencia")
	}
	if isModuloPermitidoByLicencia(permModuleFinanzas, allowed) {
		t.Fatal("expected finanzas to be disabled when omitted from licencia")
	}
}

func TestApplyLicenciaRestriccionesDisablesActionsForInactiveModules(t *testing.T) {
	rows := []permissionModuleMatrixRow{
		{
			Modulo: permModuleVentas,
			Acciones: map[string]bool{
				permActionRead:   true,
				permActionCreate: true,
			},
		},
		{
			Modulo: permModuleFinanzas,
			Acciones: map[string]bool{
				permActionRead:    true,
				permActionCreate:  true,
				permActionUpdate:  true,
				permActionDelete:  true,
				permActionApprove: true,
			},
		},
	}
	allowed, _ := parseLicenciaModulosCSV(permModuleVentas)

	filtered := applyLicenciaRestriccionesToModuleRows(rows, allowed)

	if !filtered[0].Acciones[permActionRead] || !filtered[0].Acciones[permActionCreate] {
		t.Fatal("expected enabled licencia module actions to remain active")
	}
	for _, action := range permissionActionsCatalogOrdered {
		if filtered[1].Acciones[action] {
			t.Fatalf("expected finanzas action %s to be disabled by licencia", action)
		}
	}
}
