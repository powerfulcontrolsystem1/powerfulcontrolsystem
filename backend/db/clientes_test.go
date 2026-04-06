package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openClientesTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "clientes_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = dbConn.Close()
	})
	return dbConn
}

func TestGetClientePerfilComercialByEmpresaAndHistorial(t *testing.T) {
	dbConn := openClientesTestDB(t)
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	clienteID, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         301,
		TipoDocumento:     "CC",
		NumeroDocumento:   "123456789",
		NombreRazonSocial: "Cliente Perfil",
		UsuarioCreador:    "tester",
	})
	if err != nil {
		t.Fatalf("create cliente: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO carritos_compras (
		empresa_id, codigo, nombre, cliente_id, estado_carrito, total, total_pagado, pagado_en, fecha_creacion, fecha_actualizacion, estado
	) VALUES (?, 'CAR-C1', 'Venta 1', ?, 'cerrado', 120000, 120000, datetime('now','-10 day'), datetime('now','-10 day'), datetime('now','-10 day'), 'activo')`, 301, clienteID); err != nil {
		t.Fatalf("insert carrito 1: %v", err)
	}
	var carritoID1 int64
	if err := dbConn.QueryRow(`SELECT last_insert_rowid()`).Scan(&carritoID1); err != nil {
		t.Fatalf("last_insert_rowid carrito 1: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO carrito_compra_items (empresa_id, carrito_id, descripcion, cantidad, precio_unitario, total_linea, estado) VALUES (?, ?, 'Item A', 1, 120000, 120000, 'activo')`, 301, carritoID1); err != nil {
		t.Fatalf("insert item carrito 1: %v", err)
	}

	if _, err := dbConn.Exec(`INSERT INTO carritos_compras (
		empresa_id, codigo, nombre, cliente_id, estado_carrito, total, total_pagado, pagado_en, fecha_creacion, fecha_actualizacion, estado
	) VALUES (?, 'CAR-C2', 'Venta 2', ?, 'cerrado', 90000, 90000, datetime('now','-5 day'), datetime('now','-5 day'), datetime('now','-5 day'), 'activo')`, 301, clienteID); err != nil {
		t.Fatalf("insert carrito 2: %v", err)
	}
	var carritoID2 int64
	if err := dbConn.QueryRow(`SELECT last_insert_rowid()`).Scan(&carritoID2); err != nil {
		t.Fatalf("last_insert_rowid carrito 2: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO carrito_compra_items (empresa_id, carrito_id, descripcion, cantidad, precio_unitario, total_linea, estado) VALUES (?, ?, 'Item B', 2, 45000, 90000, 'activo')`, 301, carritoID2); err != nil {
		t.Fatalf("insert item carrito 2: %v", err)
	}

	perfil, err := GetClientePerfilComercialByEmpresa(dbConn, 301, clienteID)
	if err != nil {
		t.Fatalf("get perfil cliente: %v", err)
	}
	if perfil.Cliente.ID != clienteID {
		t.Fatalf("expected perfil cliente_id=%d, got %d", clienteID, perfil.Cliente.ID)
	}
	if perfil.NumeroCompras != 2 {
		t.Fatalf("expected numero_compras=2, got %d", perfil.NumeroCompras)
	}
	if perfil.MontoCompras <= 0 {
		t.Fatalf("expected monto_compras > 0, got %.2f", perfil.MontoCompras)
	}
	if perfil.Segmento == "nuevo" {
		t.Fatalf("expected segmento distinto de nuevo, got %q", perfil.Segmento)
	}

	historial, err := GetClienteHistorialComprasByEmpresa(dbConn, 301, clienteID, 10)
	if err != nil {
		t.Fatalf("get historial cliente: %v", err)
	}
	if len(historial) != 2 {
		t.Fatalf("expected historial len=2, got %d", len(historial))
	}
	if historial[0].Codigo != "CAR-C2" {
		t.Fatalf("expected primera compra reciente CAR-C2, got %+v", historial[0])
	}

	segmentos, err := GetClientesSegmentacionByEmpresa(dbConn, 301, true, "")
	if err != nil {
		t.Fatalf("get segmentacion clientes: %v", err)
	}
	if len(segmentos) == 0 {
		t.Fatalf("expected al menos un segmento")
	}
	if segmentos[0].Clientes <= 0 {
		t.Fatalf("expected clientes > 0 en segmento, got %+v", segmentos[0])
	}
}

func TestGetClienteByID(t *testing.T) {
	dbConn := openClientesTestDB(t)
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}

	clienteID, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         450,
		TipoDocumento:     "CC",
		NumeroDocumento:   "999111333",
		NombreRazonSocial: "Cliente Email FE",
		Email:             "cliente.fe@test.com",
		UsuarioCreador:    "tester",
	})
	if err != nil {
		t.Fatalf("create cliente: %v", err)
	}

	item, err := GetClienteByID(dbConn, 450, clienteID)
	if err != nil {
		t.Fatalf("get cliente by id: %v", err)
	}
	if item == nil {
		t.Fatalf("expected cliente item, got nil")
	}
	if item.ID != clienteID {
		t.Fatalf("expected cliente id=%d, got %d", clienteID, item.ID)
	}
	if item.Email != "cliente.fe@test.com" {
		t.Fatalf("expected email cliente.fe@test.com, got %q", item.Email)
	}
}

func TestGetClientePerfilComercialByEmpresaSinComprasSegmentoNuevo(t *testing.T) {
	dbConn := openClientesTestDB(t)
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	clienteID, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         302,
		TipoDocumento:     "CC",
		NumeroDocumento:   "444999111",
		NombreRazonSocial: "Cliente Nuevo",
		UsuarioCreador:    "tester",
	})
	if err != nil {
		t.Fatalf("create cliente: %v", err)
	}

	perfil, err := GetClientePerfilComercialByEmpresa(dbConn, 302, clienteID)
	if err != nil {
		t.Fatalf("get perfil cliente: %v", err)
	}
	if perfil.NumeroCompras != 0 {
		t.Fatalf("expected numero_compras=0, got %d", perfil.NumeroCompras)
	}
	if perfil.Segmento != "nuevo" {
		t.Fatalf("expected segmento nuevo, got %q", perfil.Segmento)
	}
}

func TestCreateClienteDeduplicacionDocumentoCorreoTelefono(t *testing.T) {
	dbConn := openClientesTestDB(t)
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}

	if _, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         510,
		TipoDocumento:     "CC",
		NumeroDocumento:   "900123",
		NombreRazonSocial: "Cliente Base",
		Email:             "cliente.base@test.com",
		Telefono:          "300-111-2233",
		UsuarioCreador:    "tester",
	}); err != nil {
		t.Fatalf("create base cliente: %v", err)
	}

	_, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         510,
		TipoDocumento:     "CC",
		NumeroDocumento:   "900123",
		NombreRazonSocial: "Cliente Documento Duplicado",
		Email:             "otro.correo@test.com",
		Telefono:          "3009998877",
		UsuarioCreador:    "tester",
	})
	if err == nil {
		t.Fatalf("expected duplicate documento error")
	}
	var dupErr *ClienteDuplicadoError
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected ClienteDuplicadoError for documento, got %v", err)
	}
	if dupErr.Campo != "documento" {
		t.Fatalf("expected campo=documento, got %q", dupErr.Campo)
	}

	_, err = CreateCliente(dbConn, Cliente{
		EmpresaID:         510,
		TipoDocumento:     "CC",
		NumeroDocumento:   "900124",
		NombreRazonSocial: "Cliente Correo Duplicado",
		Email:             "CLIENTE.BASE@TEST.COM",
		Telefono:          "3005551122",
		UsuarioCreador:    "tester",
	})
	if err == nil {
		t.Fatalf("expected duplicate correo error")
	}
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected ClienteDuplicadoError for correo, got %v", err)
	}
	if dupErr.Campo != "correo" {
		t.Fatalf("expected campo=correo, got %q", dupErr.Campo)
	}

	_, err = CreateCliente(dbConn, Cliente{
		EmpresaID:         510,
		TipoDocumento:     "CC",
		NumeroDocumento:   "900125",
		NombreRazonSocial: "Cliente Telefono Duplicado",
		Email:             "cliente.telefono@test.com",
		Telefono:          "(300) 111 2233",
		UsuarioCreador:    "tester",
	})
	if err == nil {
		t.Fatalf("expected duplicate telefono error")
	}
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected ClienteDuplicadoError for telefono, got %v", err)
	}
	if dupErr.Campo != "telefono" {
		t.Fatalf("expected campo=telefono, got %q", dupErr.Campo)
	}
}

func TestUpdateClienteDeduplicacionCorreoTelefono(t *testing.T) {
	dbConn := openClientesTestDB(t)
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}

	clienteAID, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         511,
		TipoDocumento:     "CC",
		NumeroDocumento:   "910001",
		NombreRazonSocial: "Cliente A",
		Email:             "cliente.a@test.com",
		Telefono:          "3017008899",
		UsuarioCreador:    "tester",
	})
	if err != nil {
		t.Fatalf("create cliente A: %v", err)
	}

	clienteBID, err := CreateCliente(dbConn, Cliente{
		EmpresaID:         511,
		TipoDocumento:     "CC",
		NumeroDocumento:   "910002",
		NombreRazonSocial: "Cliente B",
		Email:             "cliente.b@test.com",
		Telefono:          "3018009900",
		UsuarioCreador:    "tester",
	})
	if err != nil {
		t.Fatalf("create cliente B: %v", err)
	}

	clienteA, err := GetClienteByID(dbConn, 511, clienteAID)
	if err != nil {
		t.Fatalf("get cliente A: %v", err)
	}
	clienteB, err := GetClienteByID(dbConn, 511, clienteBID)
	if err != nil {
		t.Fatalf("get cliente B: %v", err)
	}

	clienteB.Email = "CLIENTE.A@TEST.COM"
	err = UpdateCliente(dbConn, *clienteB)
	if err == nil {
		t.Fatalf("expected duplicate correo error on update")
	}
	var dupErr *ClienteDuplicadoError
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected ClienteDuplicadoError on correo update, got %v", err)
	}
	if dupErr.Campo != "correo" {
		t.Fatalf("expected campo=correo, got %q", dupErr.Campo)
	}

	clienteB.Email = "cliente.b@test.com"
	clienteB.Telefono = "301 700 8899"
	err = UpdateCliente(dbConn, *clienteB)
	if err == nil {
		t.Fatalf("expected duplicate telefono error on update")
	}
	if !errors.As(err, &dupErr) {
		t.Fatalf("expected ClienteDuplicadoError on telefono update, got %v", err)
	}
	if dupErr.Campo != "telefono" {
		t.Fatalf("expected campo=telefono, got %q", dupErr.Campo)
	}

	clienteA.Email = "cliente.a@test.com"
	clienteA.Telefono = "301-700-8899"
	if err := UpdateCliente(dbConn, *clienteA); err != nil {
		t.Fatalf("expected update self values without duplicate conflict, got %v", err)
	}
}
