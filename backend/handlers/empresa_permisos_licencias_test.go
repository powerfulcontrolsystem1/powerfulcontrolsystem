package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestLicenciaModulosCSVControlsModuleAccess(t *testing.T) {
	allowed, ordered := parseLicenciaModulosCSV(" ventas, SEGURIDAD, carnets, modulo_desconocido, ventas ")
	if len(ordered) != 3 {
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
	if !isModuloPermitidoByLicencia(permModuleCarnets, allowed) {
		t.Fatal("expected carnets to be enabled by licencia")
	}
}

func TestLicenciaModulosLegacyFallbacksKeepSplitModulesEnabled(t *testing.T) {
	allowed, _ := parseLicenciaModulosCSV(strings.Join([]string{
		permModuleVentas,
		permModuleSeguridad,
		permModuleFinanzas,
		permModuleInventario,
		permModuleClientes,
	}, ","))

	cases := []string{
		permModuleReservasHotel,
		permModuleChatTareas,
		permModuleHorariosTrab,
		permModuleAsistenciaEmpleados,
		permModuleVehiculosRegistro,
		permModuleHojaVidaOperativa,
		permModuleUbicacionGPS,
		permModuleNominaSueldos,
		permModuleReportes,
		permModuleAuditoria,
		permModuleBackups,
		permModuleDocumentosOnlyOffice,
		permModuleCRMUnificado,
	}
	for _, module := range cases {
		if !isModuloPermitidoByLicencia(module, allowed) {
			t.Fatalf("expected legacy licencia modules to enable split module %s", module)
		}
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

func TestApplyAdminEmpresaCompartidaScopeSoloVerDisablesWriteActions(t *testing.T) {
	rows := []permissionModuleMatrixRow{
		{
			Modulo: permModuleVentas,
			Acciones: map[string]bool{
				permActionRead:    true,
				permActionCreate:  true,
				permActionUpdate:  true,
				permActionDelete:  true,
				permActionApprove: true,
			},
		},
	}

	filtered := applyAdminEmpresaCompartidaScopeToModuleRows(rows, &dbpkg.AdminEmpresaCompartidaAcceso{NivelAcceso: "solo_ver"})

	if !filtered[0].Acciones[permActionRead] {
		t.Fatal("expected read action to remain enabled for solo_ver shared access")
	}
	for _, action := range []string{permActionCreate, permActionUpdate, permActionDelete, permActionApprove} {
		if filtered[0].Acciones[action] {
			t.Fatalf("expected action %s to be disabled for solo_ver shared access", action)
		}
	}
}

func TestApplyAdminEmpresaCompartidaScopeSelectedModules(t *testing.T) {
	rows := []permissionModuleMatrixRow{
		{Modulo: permModuleVentas, Acciones: map[string]bool{permActionRead: true, permActionCreate: true}},
		{Modulo: permModuleFinanzas, Acciones: map[string]bool{permActionRead: true, permActionCreate: true}},
	}

	filtered := applyAdminEmpresaCompartidaScopeToModuleRows(rows, &dbpkg.AdminEmpresaCompartidaAcceso{
		NivelAcceso:       "modulos",
		ModulosPermitidos: permModuleVentas,
	})

	if !filtered[0].Acciones[permActionRead] || !filtered[0].Acciones[permActionCreate] {
		t.Fatal("expected selected shared module to keep existing actions")
	}
	for _, action := range permissionActionsCatalogOrdered {
		if filtered[1].Acciones[action] {
			t.Fatalf("expected non-selected shared module action %s to be disabled", action)
		}
	}
}

func TestEmpresaVerticalScopeKeepsCoreAndOnlyChosenVertical(t *testing.T) {
	rows := []permissionModuleMatrixRow{
		{Modulo: permModuleVentas, Acciones: map[string]bool{permActionRead: true, permActionCreate: true}},
		{Modulo: permModuleInventario, Acciones: map[string]bool{permActionRead: true, permActionCreate: true}},
		{Modulo: permModuleGimnasio, Acciones: map[string]bool{permActionRead: true, permActionCreate: true}},
		{Modulo: permModuleOdontologia, Acciones: map[string]bool{permActionRead: true, permActionCreate: true}},
	}
	scope := empresaVerticalScope{
		Enabled:     true,
		Allowed:     map[string]bool{permModuleGimnasio: true},
		AllowedList: []string{permModuleGimnasio},
	}

	filtered := applyEmpresaVerticalScopeToModuleRows(rows, scope)

	if !filtered[0].Acciones[permActionRead] || !filtered[1].Acciones[permActionCreate] {
		t.Fatal("expected core modules to remain enabled under vertical scope")
	}
	if !filtered[2].Acciones[permActionCreate] {
		t.Fatal("expected selected gimnasio vertical to remain enabled")
	}
	for _, action := range permissionActionsCatalogOrdered {
		if filtered[3].Acciones[action] {
			t.Fatalf("expected odontologia action %s to be hidden for gimnasio company", action)
		}
	}
}

func TestNormalizeVerticalScopeAliases(t *testing.T) {
	cases := map[string]string{
		"consultorio_odontologico": permModuleOdontologia,
		"taxi":                     permModuleTaxiSystem,
		"constructora":             permModuleAIUConstruccion,
		"apartamento-turistico":    permModuleApartTuristicos,
	}
	for raw, want := range cases {
		if got := normalizeVerticalScopeModule(raw); got != want {
			t.Fatalf("normalizeVerticalScopeModule(%q)=%q, want %q", raw, got, want)
		}
	}
}

func TestValidateEmpresaIDConsistencyRejectsBodyQueryMismatch(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/empresa/productos?empresa_id=7", bytes.NewBufferString(`{"empresa_id":8,"nombre":"Producto"}`))
	r.Header.Set("Content-Type", "application/json")

	if err := validateEmpresaIDConsistency(r, 7); err == nil {
		t.Fatal("expected mismatch between query empresa_id and JSON empresa_id to be rejected")
	}

	body := new(bytes.Buffer)
	if _, err := body.ReadFrom(r.Body); err != nil {
		t.Fatalf("reading restored body: %v", err)
	}
	if got := body.String(); !strings.Contains(got, `"empresa_id":8`) {
		t.Fatalf("expected JSON body to be restored for downstream handlers, got %q", got)
	}
}

func TestValidateEmpresaIDConsistencyAllowsMatchingSources(t *testing.T) {
	r := httptest.NewRequest(http.MethodPut, "/api/empresa/productos?empresa_id=7", bytes.NewBufferString(`{"empresaId":7,"nombre":"Producto"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Empresa-ID", "7")

	if err := validateEmpresaIDConsistency(r, 7); err != nil {
		t.Fatalf("expected matching empresa_id sources to pass, got %v", err)
	}
}

func TestResolvePermissionPageKeyForDirectSaleCart(t *testing.T) {
	cases := []string{
		"/api/empresa/carritos_compra?empresa_id=7&modo=venta_directa",
		"/api/empresa/carritos_compra?empresa_id=7&carrito_codigo=VENTA-DIRECTA-7",
		"/api/empresa/carritos_compra?empresa_id=7&perm_page=linkVentaDirecta",
		"/api/empresa/carritos_compra/items?empresa_id=7&modo=venta_directa&carrito_id=3",
		"/api/empresa/carritos_compra/items?empresa_id=7&carrito_codigo=VENTA-DIRECTA-7-0&carrito_id=3",
	}
	for _, rawURL := range cases {
		r := httptest.NewRequest(http.MethodGet, rawURL, nil)
		if got := resolvePermissionPageKeyForRequest(r); got != "linkVentaDirecta" {
			t.Fatalf("resolvePermissionPageKeyForRequest(%q)=%q, want linkVentaDirecta", rawURL, got)
		}
	}
}

func TestResolvePermissionPageKeyForRegularCart(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=7", nil)
	if got := resolvePermissionPageKeyForRequest(r); got != "linkCarritoCompras" {
		t.Fatalf("resolvePermissionPageKeyForRequest regular cart=%q, want linkCarritoCompras", got)
	}
}

func TestPermissionPagesCatalogExposesOnlyUniversalBusinessGroups(t *testing.T) {
	moduleRows := make([]permissionModuleMatrixRow, 0, len(permissionModulesCatalogOrdered))
	for _, module := range permissionModulesCatalogOrdered {
		actions := make(map[string]bool, len(permissionActionsCatalogOrdered))
		for _, action := range permissionActionsCatalogOrdered {
			actions[action] = true
		}
		moduleRows = append(moduleRows, permissionModuleMatrixRow{
			Modulo:   module,
			Acciones: actions,
		})
	}

	rows := buildPermissionPagesCatalogFromModuleRows(moduleRows, nil)
	if len(rows) == 0 {
		t.Fatal("expected permission page catalog rows")
	}

	legacyGroups := map[string]bool{
		"Operación diaria y ventas":                true,
		"Operación y venta":                        true,
		"Verticales de negocio":                    true,
		"Inventario y compras":                     true,
		"Inventario y catálogo":                    true,
		"Compras":                                  true,
		"Centro financiero y contable":             true,
		"Finanzas y reportes":                      true,
		"Finanzas y nómina":                        true,
		"Administración y configuración":           true,
		"Seguridad e integración":                  true,
		"Configuración":                            true,
		"Facturación electrónica":                  true,
		"Facturación DIAN":                         true,
		"Gestión de Relaciones con Clientes (CRM)": true,
		"Clientes":                                 true,
		"Personas y activos":                       true,
		"Análisis y control":                       true,
		"Analisis y control":                       true,
		"Documentos, nube y soporte":               true,
	}

	for _, row := range rows {
		if legacyGroups[row.Grupo] {
			t.Fatalf("permission page %s leaked legacy group %q", row.PaginaClave, row.Grupo)
		}
		if row.Grupo != "Acceso general" && row.Grupo != "Otras" && !strings.Contains(strings.ToLower(row.Grupo), "universal") {
			t.Fatalf("permission page %s group %q should be universal", row.PaginaClave, row.Grupo)
		}
	}
}

func TestPermissionModuleDisplayNamesExposeUniversalCoreLabels(t *testing.T) {
	labels := PermissionModuleDisplayNameMap()

	cases := map[string]string{
		permModuleVentas:      "Ventas universales",
		permModuleInventario:  "Inventario universal",
		permModuleFinanzas:    "Finanzas universales",
		permModuleClientes:    "CRM universal",
		permModuleCompras:     "Compras universales",
		permModuleFacturacion: "Facturación electrónica universal",
		permModuleSeguridad:   "Administración universal",
		permModuleAlquileres:  "Alquiler universal",
	}
	for module, want := range cases {
		if got := labels[module]; !strings.Contains(got, want) {
			t.Fatalf("module %s label=%q, want it to contain %q", module, got, want)
		}
	}
}
