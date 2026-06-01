package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func findPermissionModuleRowForTest(t *testing.T, rows []permissionModuleMatrixRow, modulo string) permissionModuleMatrixRow {
	t.Helper()
	for _, row := range rows {
		if row.Modulo == modulo {
			return row
		}
	}
	t.Fatalf("modulo %s no encontrado", modulo)
	return permissionModuleMatrixRow{}
}

func TestPorteroRoleOnlyAllowsStationViewAndActivation(t *testing.T) {
	if got := normalizePermissionRole("vigilante"); got != "portero" {
		t.Fatalf("expected vigilante to normalize as portero, got %q", got)
	}

	rows := buildPermissionModuleMatrixForRole("portero")
	ventas := findPermissionModuleRowForTest(t, rows, permModuleVentas)
	if !ventas.Read || !ventas.Approve {
		t.Fatalf("portero debe poder leer ventas y activar estaciones: %+v", ventas)
	}
	if ventas.Create || ventas.Update || ventas.Delete {
		t.Fatalf("portero no debe poder crear, editar ni eliminar ventas: %+v", ventas)
	}

	inventario := findPermissionModuleRowForTest(t, rows, permModuleInventario)
	if inventario.Read || inventario.Create || inventario.Update || inventario.Delete || inventario.Approve {
		t.Fatalf("portero no debe tener permisos de inventario: %+v", inventario)
	}

	pages := buildPermissionPagesMapForRoleDynamic(nil, "portero", rows)
	if !pages["linkEstaciones"] {
		t.Fatal("portero debe ver la pagina de estaciones")
	}
	for _, page := range []string{"linkPanelEmpresa", "linkVentaDirecta", "linkCarritoCompras", "linkCorteCaja", "linkConfigEstaciones", "linkChatIA"} {
		if pages[page] {
			t.Fatalf("portero no debe ver %s", page)
		}
	}
}

func TestPorteroCarritoRestrictions(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?action=activar_estacion", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminRoleEfectivo", "portero"))
	if isPorteroRestrictedCarritoRequest(req, "activar_estacion") {
		t.Fatal("portero debe poder activar estacion")
	}

	req = httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?action=pagar_estacion", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminRoleEfectivo", "portero"))
	if !isPorteroRestrictedCarritoRequest(req, "pagar_estacion") {
		t.Fatal("portero no debe poder pagar estaciones")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminRoleEfectivo", "portero"))
	if !isPorteroRestrictedCarritoRequest(req, "") {
		t.Fatal("portero no debe poder crear carritos")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=1&estacion_id=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminRoleEfectivo", "portero"))
	if isPorteroRestrictedCarritoRequest(req, "") {
		t.Fatal("portero debe poder consultar estado de estaciones")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?action=totales_pago&empresa_id=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminRoleEfectivo", "portero"))
	if !isPorteroRestrictedCarritoRequest(req, "totales_pago") {
		t.Fatal("portero no debe poder consultar totales de caja")
	}
}

func TestContadorRoleOnlyAllowsFinanceAndTaxesRead(t *testing.T) {
	if got := normalizePermissionRole("contador"); got != "contador" {
		t.Fatalf("expected contador to stay as contador, got %q", got)
	}

	rows := buildPermissionModuleMatrixForRole("contador")
	finanzas := findPermissionModuleRowForTest(t, rows, permModuleFinanzas)
	if !finanzas.Read || finanzas.Create || finanzas.Update || finanzas.Delete || finanzas.Approve {
		t.Fatalf("contador debe tener solo lectura de finanzas: %+v", finanzas)
	}
	facturacion := findPermissionModuleRowForTest(t, rows, permModuleFacturacion)
	if !facturacion.Read || facturacion.Create || facturacion.Update || facturacion.Delete || facturacion.Approve {
		t.Fatalf("contador debe tener solo lectura de facturacion para impuestos: %+v", facturacion)
	}
	ventas := findPermissionModuleRowForTest(t, rows, permModuleVentas)
	if ventas.Read || ventas.Create || ventas.Update || ventas.Delete || ventas.Approve {
		t.Fatalf("contador no debe tener permisos de ventas: %+v", ventas)
	}

	pages := buildPermissionPagesMapForRoleDynamic(nil, "contador", rows)
	for _, page := range []string{"linkFinanzas", "linkFinanzasMain", "linkImpuestos"} {
		if !pages[page] {
			t.Fatalf("contador debe ver %s", page)
		}
	}
	for _, page := range []string{"linkPanelEmpresa", "linkVentaDirecta", "linkEstaciones", "linkCarritoCompras", "linkCorteCaja", "linkContabilidadColombia", "linkDeclaracionesTributarias"} {
		if pages[page] {
			t.Fatalf("contador no debe ver %s", page)
		}
	}
}

func TestEmpresarioRoleOnlyAllowsExecutiveReportsRead(t *testing.T) {
	if got := normalizePermissionRole("propietario"); got != "empresario" {
		t.Fatalf("expected propietario to normalize as empresario, got %q", got)
	}

	rows := buildPermissionModuleMatrixForRole("empresario")
	reportes := findPermissionModuleRowForTest(t, rows, permModuleReportes)
	if !reportes.Read || reportes.Create || reportes.Update || reportes.Delete || reportes.Approve {
		t.Fatalf("empresario debe tener solo lectura de reportes: %+v", reportes)
	}

	for _, modulo := range []string{permModuleVentas, permModuleFinanzas, permModuleInventario, permModuleFacturacion, permModuleSeguridad} {
		row := findPermissionModuleRowForTest(t, rows, modulo)
		if row.Read || row.Create || row.Update || row.Delete || row.Approve {
			t.Fatalf("empresario no debe tener permisos de %s: %+v", modulo, row)
		}
	}

	pages := buildPermissionPagesMapForRoleDynamic(nil, "empresario", rows)
	for _, page := range []string{"linkReportes", "linkReportesEjecutivos"} {
		if !pages[page] {
			t.Fatalf("empresario debe ver %s", page)
		}
	}
	for _, page := range []string{"linkPanelEmpresa", "linkVentaDirecta", "linkEstaciones", "linkCarritoCompras", "linkCorteCaja", "linkReportesTurnos", "linkFinanzas", "linkImpuestos"} {
		if pages[page] {
			t.Fatalf("empresario no debe ver %s", page)
		}
	}
}

func TestServicioLimpiezaRoleOnlyAllowsStationsAndCleaning(t *testing.T) {
	if got := normalizePermissionRole("Servicio de limpieza"); got != "servicio_limpieza" {
		t.Fatalf("expected Servicio de limpieza to normalize as servicio_limpieza, got %q", got)
	}

	rows := buildPermissionModuleMatrixForRole("servicio_limpieza")
	ventas := findPermissionModuleRowForTest(t, rows, permModuleVentas)
	if !ventas.Read || ventas.Create || ventas.Update || ventas.Delete || ventas.Approve {
		t.Fatalf("servicio_limpieza debe tener solo lectura de estaciones/ventas: %+v", ventas)
	}
	for _, modulo := range []string{permModuleFinanzas, permModuleInventario, permModuleFacturacion, permModuleSeguridad, permModuleReportes} {
		row := findPermissionModuleRowForTest(t, rows, modulo)
		if row.Read || row.Create || row.Update || row.Delete || row.Approve {
			t.Fatalf("servicio_limpieza no debe tener permisos de %s: %+v", modulo, row)
		}
	}

	pages := buildPermissionPagesMapForRoleDynamic(nil, "servicio_limpieza", rows)
	if !pages["linkEstaciones"] {
		t.Fatal("servicio_limpieza debe ver la pagina de estaciones")
	}
	for _, page := range []string{"linkPanelEmpresa", "linkVentaDirecta", "linkCarritoCompras", "linkCorteCaja", "linkConfigEstaciones", "linkReportes"} {
		if pages[page] {
			t.Fatalf("servicio_limpieza no debe ver %s", page)
		}
	}
}

func TestServicioLimpiezaCarritoRestrictions(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminRoleEfectivo", "servicio_limpieza"))
	if isServicioLimpiezaRestrictedCarritoRequest(req, "") {
		t.Fatal("servicio_limpieza debe poder consultar el tablero de estaciones")
	}

	req = httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?action=activar_estacion", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminRoleEfectivo", "servicio_limpieza"))
	if !isServicioLimpiezaRestrictedCarritoRequest(req, "activar_estacion") {
		t.Fatal("servicio_limpieza no debe poder activar estaciones")
	}

	req = httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?action=totales_pago&empresa_id=1", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminRoleEfectivo", "servicio_limpieza"))
	if !isServicioLimpiezaRestrictedCarritoRequest(req, "totales_pago") {
		t.Fatal("servicio_limpieza no debe poder consultar totales de caja")
	}
}

func TestTecnicoSolarRoleOnlyAllowsSolarRead(t *testing.T) {
	if got := normalizePermissionRole("Tecnico solar"); got != "tecnico_solar" {
		t.Fatalf("expected Tecnico solar to normalize as tecnico_solar, got %q", got)
	}

	rows := buildPermissionModuleMatrixForRole("tecnico_solar")
	solar := findPermissionModuleRowForTest(t, rows, permModuleEnergiaSolar)
	if !solar.Read || solar.Create || solar.Update || solar.Delete || solar.Approve {
		t.Fatalf("tecnico_solar debe tener solo lectura de energia solar: %+v", solar)
	}
	for _, modulo := range []string{permModuleVentas, permModuleInventario, permModuleFinanzas, permModuleSeguridad, permModuleReportes} {
		row := findPermissionModuleRowForTest(t, rows, modulo)
		if row.Read || row.Create || row.Update || row.Delete || row.Approve {
			t.Fatalf("tecnico_solar no debe tener permisos de %s: %+v", modulo, row)
		}
	}

	pages := buildPermissionPagesMapForRoleDynamic(nil, "tecnico_solar", rows)
	if !pages["linkEnergiaSolar"] {
		t.Fatal("tecnico_solar debe ver Energia solar")
	}
	for _, page := range []string{"linkPanelEmpresa", "linkEstaciones", "linkControlElectrico", "linkConfiguracion", "linkReportes"} {
		if pages[page] {
			t.Fatalf("tecnico_solar no debe ver %s", page)
		}
	}
}

func TestJefeBodegaRoleAllowsWarehouseInventory(t *testing.T) {
	if got := normalizePermissionRole("Jefe de bodega"); got != "jefe_bodega" {
		t.Fatalf("expected Jefe de bodega to normalize as jefe_bodega, got %q", got)
	}

	rows := buildPermissionModuleMatrixForRole("jefe_bodega")
	inventario := findPermissionModuleRowForTest(t, rows, permModuleInventario)
	if !inventario.Read || !inventario.Create || !inventario.Update || !inventario.Approve || inventario.Delete {
		t.Fatalf("jefe_bodega debe administrar inventario sin eliminar: %+v", inventario)
	}
	compras := findPermissionModuleRowForTest(t, rows, permModuleCompras)
	if !compras.Read || compras.Create || compras.Update || compras.Delete || compras.Approve {
		t.Fatalf("jefe_bodega debe tener solo lectura de compras: %+v", compras)
	}
	ventas := findPermissionModuleRowForTest(t, rows, permModuleVentas)
	if ventas.Read || ventas.Create || ventas.Update || ventas.Delete || ventas.Approve {
		t.Fatalf("jefe_bodega no debe tener permisos de ventas: %+v", ventas)
	}

	pages := buildPermissionPagesMapForRoleDynamic(nil, "jefe_bodega", rows)
	for _, page := range []string{"linkProductos", "linkInventarioAvanzado", "linkBodegas", "linkCategorias", "linkPreciosHistorial"} {
		if !pages[page] {
			t.Fatalf("jefe_bodega debe ver %s", page)
		}
	}
	for _, page := range []string{"linkPanelEmpresa", "linkVentaDirecta", "linkEstaciones", "linkCorteCaja", "linkConfiguracion"} {
		if pages[page] {
			t.Fatalf("jefe_bodega no debe ver %s", page)
		}
	}
}

func TestRecursosHumanosRoleAllowsPeopleOperations(t *testing.T) {
	rows := buildPermissionModuleMatrixForRole("recursos_humanos")
	for _, modulo := range []string{permModuleHorariosTrab, permModuleAsistenciaEmpleados, permModuleNominaSueldos} {
		row := findPermissionModuleRowForTest(t, rows, modulo)
		if !row.Read || !row.Create || !row.Update || row.Delete || row.Approve {
			t.Fatalf("recursos_humanos debe gestionar %s sin eliminar/aprobar: %+v", modulo, row)
		}
	}
	ventas := findPermissionModuleRowForTest(t, rows, permModuleVentas)
	if ventas.Read || ventas.Create || ventas.Update || ventas.Delete || ventas.Approve {
		t.Fatalf("recursos_humanos no debe tener ventas: %+v", ventas)
	}
}
