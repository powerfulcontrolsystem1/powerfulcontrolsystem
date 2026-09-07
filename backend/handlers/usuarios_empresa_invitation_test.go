package handlers

import (
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaUsuarioEstadoBloqueaPrimerIngresoPermitePendienteInactivo(t *testing.T) {
	item := &dbpkg.EmpresaUsuario{
		Estado:          "inactivo",
		EmailConfirmado: 0,
	}

	if empresaUsuarioEstadoBloqueaPrimerIngreso(item) {
		t.Fatal("usuario pendiente con invitacion valida no debe bloquear el primer ingreso")
	}
}

func TestEmpresaUsuarioEstadoBloqueaPrimerIngresoBloqueaConfirmadoInactivo(t *testing.T) {
	item := &dbpkg.EmpresaUsuario{
		Estado:          "inactivo",
		EmailConfirmado: 1,
		PasswordSet:     1,
		PasswordHash:    "hash",
		PasswordSalt:    "salt",
	}

	if !empresaUsuarioEstadoBloqueaPrimerIngreso(item) {
		t.Fatal("usuario confirmado e inactivo debe quedar bloqueado")
	}
}

func TestEmpresaUsuarioEstadoBloqueaPrimerIngresoPermiteConfirmadoSinPassword(t *testing.T) {
	item := &dbpkg.EmpresaUsuario{
		Estado:          "inactivo",
		EmailConfirmado: 1,
		PasswordSet:     0,
	}

	if empresaUsuarioEstadoBloqueaPrimerIngreso(item) {
		t.Fatal("usuario confirmado por invitacion pero sin password debe completar primer ingreso")
	}
}

func TestEmpresaUsuarioPasswordConfirmPayloadAceptaAliases(t *testing.T) {
	cases := []empresaUsuarioPasswordConfirmPayload{
		{PasswordConfirm: "fixture-confirmation-value"},
		{PasswordConfirmation: "fixture-confirmation-value"},
		{ConfirmPassword: "fixture-confirmation-value"},
		{ConfirmarPassword: "fixture-confirmation-value"},
		{ConfirmarContrasena: "fixture-confirmation-value"},
		{ConfirmarContrasenia: "fixture-confirmation-value"},
		{ConfirmacionContrasena: "fixture-confirmation-value"},
	}
	for _, tc := range cases {
		if got := tc.value(); got != "fixture-confirmation-value" {
			t.Fatalf("confirmacion normalizada=%q, want ClaveSegura123", got)
		}
	}
}

func TestNormalizePermissionRoleCajaEsCajero(t *testing.T) {
	for _, raw := range []string{"Caja", "caja", "Caja principal", "caja_turno"} {
		if got := normalizePermissionRole(raw); got != "cajero" {
			t.Fatalf("normalizePermissionRole(%q)=%q, want cajero", raw, got)
		}
	}
}

func TestCajeroSoloVePaginasOperativas(t *testing.T) {
	allowed := []string{"linkVentaDirecta", "linkEstaciones", "linkCorteCaja", "linkVentas", "linkFacturasElectronicas", "linkFacturacionElectronica"}
	for _, page := range allowed {
		if !isAllowedPageForOperationalRole("cajero", page) {
			t.Fatalf("cajero debe poder ver %s", page)
		}
	}
	blocked := []string{"linkPanelEmpresa", "linkUsuarios", "linkProductos", "linkFinanzas", "linkConfiguracion", "linkReportes"}
	for _, page := range blocked {
		if isAllowedPageForOperationalRole("cajero", page) {
			t.Fatalf("cajero no debe poder ver %s", page)
		}
	}
}

func TestCajeroTienePermisosOperativosParaCarritoCompleto(t *testing.T) {
	rows := restrictPermissionModuleRowsForOperationalRole("cajero", buildPermissionModuleMatrixForRole("admin_empresa"))
	byModule := map[string]permissionModuleMatrixRow{}
	for _, row := range rows {
		byModule[row.Modulo] = row
	}

	if !byModule[permModuleVentas].Acciones[permActionCreate] {
		t.Fatal("cajero debe poder crear/cobrar ventas desde carrito")
	}
	if !byModule[permModuleFinanzas].Acciones[permActionCreate] {
		t.Fatal("cajero debe poder registrar cobros de caja desde carrito")
	}
	if !byModule[permModuleFacturacion].Acciones[permActionCreate] {
		t.Fatal("cajero debe poder emitir facturacion operativa desde carrito")
	}
	if !byModule[permModuleInventario].Acciones[permActionRead] || byModule[permModuleInventario].Acciones[permActionCreate] {
		t.Fatal("cajero debe consultar catalogo de inventario sin administrar productos")
	}
	if !byModule[permModuleClientes].Acciones[permActionRead] || !byModule[permModuleClientes].Acciones[permActionCreate] || byModule[permModuleClientes].Acciones[permActionDelete] {
		t.Fatal("cajero debe leer/crear clientes desde carrito sin eliminarlos")
	}
	if isAllowedPageForOperationalRole("cajero", "linkClientes") || isAllowedPageForOperationalRole("cajero", "linkProductos") {
		t.Fatal("cajero no debe ganar paginas de Clientes o Productos en el menu")
	}
}

func TestCajeroPuedeUsarAPIAuxiliaresDelCarritoSinPaginaDeMenu(t *testing.T) {
	allowed := []string{
		"/api/empresa/clientes",
		"/api/empresa/productos",
		"/api/empresa/servicios",
		"/api/empresa/recetas_productos",
		"/api/empresa/codigos_de_descuento",
		"/api/empresa/propinas",
		"/api/empresa/comisiones",
		"/api/empresa/chat_con_inteligencia_artificial/modelos",
		"/api/empresa/chat_con_inteligencia_artificial/consultar",
		"/api/empresa/chat_con_inteligencia_artificial/consultar_stream",
		"/api/empresa/ia_pedidos_estacion/ejecutar",
		"/api/empresa/ia_radio/activar",
	}
	for _, path := range allowed {
		if !isCajeroCartAuxiliaryAPIRequest("cajero", path) {
			t.Fatalf("cajero debe poder usar API auxiliar de carrito %s", path)
		}
	}
	if isCajeroCartAuxiliaryAPIRequest("cajero", "/api/empresa/usuarios") {
		t.Fatal("cajero no debe saltarse pagina para APIs administrativas")
	}
	if isCajeroCartAuxiliaryAPIRequest("contador", "/api/empresa/clientes") {
		t.Fatal("la excepcion de APIs auxiliares aplica solo a cajero")
	}
}
