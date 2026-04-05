package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openCarritoInventarioTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "carritos_inventario_test.db")
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

func stockTotalByProducto(t *testing.T, dbConn *sql.DB, empresaID, productoID int64) float64 {
	t.Helper()
	var total float64
	err := dbConn.QueryRow(`SELECT COALESCE(SUM(cantidad), 0) FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ?`, empresaID, productoID).Scan(&total)
	if err != nil {
		t.Fatalf("stock total query: %v", err)
	}
	return total
}

func TestCarritoProductoDescuentaInventarioYVentaMantieneStock(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{
		EmpresaID: 1,
		Codigo:    "BOD-T-001",
		Nombre:    "Bodega Test",
		Estado:    "activo",
	})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	productoID, err := CreateProducto(dbConn, Producto{
		EmpresaID:          1,
		BodegaPrincipalID:  bodegaID,
		SKU:                "SKU-T-001",
		CodigoBarras:       "7701000000001",
		Nombre:             "Producto Test",
		UnidadMedida:       "unidad",
		Costo:              1000,
		Precio:             2000,
		ImpuestoPorcentaje: 19,
		Estado:             "activo",
	}, 10, "TEST_INVENTARIO_CARRITO")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:      1,
		Codigo:         "CAR-T-001",
		Nombre:         "Carrito Test",
		CanalVenta:     "mostrador",
		Moneda:         "COP",
		UsuarioCreador: "test",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("create carrito: %v", err)
	}

	stockInicial := stockTotalByProducto(t, dbConn, 1, productoID)
	if stockInicial != 10 {
		t.Fatalf("expected initial stock 10, got %.2f", stockInicial)
	}

	_, err = CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:          1,
		CarritoID:          carritoID,
		TipoItem:           "producto",
		ReferenciaID:       productoID,
		CodigoItem:         "SKU-T-001",
		Descripcion:        "Producto Test",
		UnidadMedida:       "unidad",
		Cantidad:           2,
		PrecioUnitario:     2000,
		ImpuestoPorcentaje: 19,
		ImpuestoCodigo:     "IVA",
		UsuarioCreador:     "test",
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create carrito item: %v", err)
	}

	stockTrasAgregar := stockTotalByProducto(t, dbConn, 1, productoID)
	if stockTrasAgregar != 8 {
		t.Fatalf("expected stock 8 after add-to-cart, got %.2f", stockTrasAgregar)
	}

	carrito, err := GetCarritoCompraByID(dbConn, 1, carritoID)
	if err != nil {
		t.Fatalf("get carrito: %v", err)
	}
	if err := PayCarritoStationSession(dbConn, 1, carritoID, "efectivo", "", "", "", 0, 0, carrito.Total, 0); err != nil {
		t.Fatalf("pay carrito: %v", err)
	}

	stockTrasPago := stockTotalByProducto(t, dbConn, 1, productoID)
	if stockTrasPago != 8 {
		t.Fatalf("expected stock 8 after sale payment, got %.2f", stockTrasPago)
	}
}

func TestCarritoProductoStockInsuficiente(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{
		EmpresaID: 1,
		Codigo:    "BOD-T-002",
		Nombre:    "Bodega Test 2",
		Estado:    "activo",
	})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	productoID, err := CreateProducto(dbConn, Producto{
		EmpresaID:         1,
		BodegaPrincipalID: bodegaID,
		SKU:               "SKU-T-002",
		CodigoBarras:      "7701000000002",
		Nombre:            "Producto Stock Corto",
		UnidadMedida:      "unidad",
		Costo:             500,
		Precio:            900,
		Estado:            "activo",
	}, 1, "TEST_INVENTARIO_STOCK")
	if err != nil {
		t.Fatalf("create producto: %v", err)
	}

	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:      1,
		Codigo:         "CAR-T-002",
		Nombre:         "Carrito Test 2",
		CanalVenta:     "mostrador",
		Moneda:         "COP",
		UsuarioCreador: "test",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("create carrito: %v", err)
	}

	err = nil
	_, err = CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:      1,
		CarritoID:      carritoID,
		TipoItem:       "producto",
		ReferenciaID:   productoID,
		Descripcion:    "Producto Stock Corto",
		UnidadMedida:   "unidad",
		Cantidad:       2,
		PrecioUnitario: 900,
		Estado:         "activo",
	})
	if err == nil {
		t.Fatal("expected stock insufficient error, got nil")
	}
	if err != ErrStockInsuficiente {
		t.Fatalf("expected ErrStockInsuficiente, got %v", err)
	}

	stockFinal := stockTotalByProducto(t, dbConn, 1, productoID)
	if stockFinal != 1 {
		t.Fatalf("expected unchanged stock 1 after failed add-to-cart, got %.2f", stockFinal)
	}
}

func TestCarritoComboDescuentaIngredientesYRevierteAlEliminar(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{
		EmpresaID: 1,
		Codigo:    "BOD-CMB-001",
		Nombre:    "Bodega Combos",
		Estado:    "activo",
	})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	arrozID, err := CreateProducto(dbConn, Producto{
		EmpresaID:         1,
		BodegaPrincipalID: bodegaID,
		SKU:               "ING-ARROZ",
		Nombre:            "Arroz",
		UnidadMedida:      "gramo",
		Costo:             6,
		Precio:            12,
		Estado:            "activo",
	}, 20, "TEST_COMBO_ARROZ")
	if err != nil {
		t.Fatalf("create arroz: %v", err)
	}

	proteinaID, err := CreateProducto(dbConn, Producto{
		EmpresaID:         1,
		BodegaPrincipalID: bodegaID,
		SKU:               "ING-PROT",
		Nombre:            "Proteina",
		UnidadMedida:      "unidad",
		Costo:             900,
		Precio:            1800,
		Estado:            "activo",
	}, 10, "TEST_COMBO_PROT")
	if err != nil {
		t.Fatalf("create proteina: %v", err)
	}

	comboID, err := CreateComboProducto(dbConn, ComboProducto{
		EmpresaID:          1,
		Codigo:             "CMB-PLT-001",
		Nombre:             "Plato combo",
		UnidadMedida:       "plato",
		Precio:             22000,
		ImpuestoPorcentaje: 19,
		Estado:             "activo",
	}, []ComboProductoDetalle{
		{ProductoID: arrozID, Cantidad: 2, UnidadMedida: "gramo"},
		{ProductoID: proteinaID, Cantidad: 1, UnidadMedida: "unidad"},
	})
	if err != nil {
		t.Fatalf("create combo: %v", err)
	}

	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:      1,
		Codigo:         "CAR-CMB-001",
		Nombre:         "Carrito Combo",
		CanalVenta:     "mostrador",
		Moneda:         "COP",
		UsuarioCreador: "test",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("create carrito: %v", err)
	}

	itemID, err := CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:          1,
		CarritoID:          carritoID,
		TipoItem:           "combo",
		ReferenciaID:       comboID,
		CodigoItem:         "CMB-PLT-001",
		Descripcion:        "Plato combo",
		UnidadMedida:       "plato",
		Cantidad:           3,
		PrecioUnitario:     22000,
		ImpuestoPorcentaje: 19,
		ImpuestoCodigo:     "IVA",
		UsuarioCreador:     "test",
		Estado:             "activo",
	})
	if err != nil {
		t.Fatalf("create combo item: %v", err)
	}

	stockArroz := stockTotalByProducto(t, dbConn, 1, arrozID)
	if stockArroz != 14 {
		t.Fatalf("expected arroz stock 14 after combo reserve, got %.2f", stockArroz)
	}
	stockProteina := stockTotalByProducto(t, dbConn, 1, proteinaID)
	if stockProteina != 7 {
		t.Fatalf("expected proteina stock 7 after combo reserve, got %.2f", stockProteina)
	}

	if err := SetCarritoCompraItemEstado(dbConn, 1, carritoID, itemID, "inactivo"); err != nil {
		t.Fatalf("set item inactivo: %v", err)
	}
	if got := stockTotalByProducto(t, dbConn, 1, arrozID); got != 20 {
		t.Fatalf("expected arroz stock 20 after deactivate item, got %.2f", got)
	}
	if got := stockTotalByProducto(t, dbConn, 1, proteinaID); got != 10 {
		t.Fatalf("expected proteina stock 10 after deactivate item, got %.2f", got)
	}

	if err := SetCarritoCompraItemEstado(dbConn, 1, carritoID, itemID, "activo"); err != nil {
		t.Fatalf("set item activo: %v", err)
	}
	if got := stockTotalByProducto(t, dbConn, 1, arrozID); got != 14 {
		t.Fatalf("expected arroz stock 14 after reactivation, got %.2f", got)
	}
	if got := stockTotalByProducto(t, dbConn, 1, proteinaID); got != 7 {
		t.Fatalf("expected proteina stock 7 after reactivation, got %.2f", got)
	}

	if err := DeleteCarritoCompraItem(dbConn, 1, carritoID, itemID); err != nil {
		t.Fatalf("delete combo item: %v", err)
	}
	if got := stockTotalByProducto(t, dbConn, 1, arrozID); got != 20 {
		t.Fatalf("expected arroz stock 20 after delete item, got %.2f", got)
	}
	if got := stockTotalByProducto(t, dbConn, 1, proteinaID); got != 10 {
		t.Fatalf("expected proteina stock 10 after delete item, got %.2f", got)
	}
}

func TestCarritoComboStockInsuficienteEnIngrediente(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaProductosSchema(dbConn); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	bodegaID, err := CreateBodega(dbConn, Bodega{
		EmpresaID: 1,
		Codigo:    "BOD-CMB-002",
		Nombre:    "Bodega Combos 2",
		Estado:    "activo",
	})
	if err != nil {
		t.Fatalf("create bodega: %v", err)
	}

	insumoID, err := CreateProducto(dbConn, Producto{
		EmpresaID:         1,
		BodegaPrincipalID: bodegaID,
		SKU:               "ING-LIM",
		Nombre:            "Insumo limitado",
		UnidadMedida:      "unidad",
		Costo:             200,
		Precio:            450,
		Estado:            "activo",
	}, 5, "TEST_COMBO_LIMIT")
	if err != nil {
		t.Fatalf("create insumo limitado: %v", err)
	}

	comboID, err := CreateComboProducto(dbConn, ComboProducto{
		EmpresaID:          1,
		Codigo:             "CMB-LIM-001",
		Nombre:             "Combo limitado",
		Precio:             3500,
		ImpuestoPorcentaje: 0,
		Estado:             "activo",
	}, []ComboProductoDetalle{
		{ProductoID: insumoID, Cantidad: 2, UnidadMedida: "unidad"},
	})
	if err != nil {
		t.Fatalf("create combo limitado: %v", err)
	}

	carritoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:      1,
		Codigo:         "CAR-CMB-002",
		Nombre:         "Carrito Combo Limitado",
		CanalVenta:     "mostrador",
		Moneda:         "COP",
		UsuarioCreador: "test",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("create carrito: %v", err)
	}

	_, err = CreateCarritoCompraItem(dbConn, CarritoCompraItem{
		EmpresaID:      1,
		CarritoID:      carritoID,
		TipoItem:       "combo",
		ReferenciaID:   comboID,
		Descripcion:    "Combo limitado",
		UnidadMedida:   "combo",
		Cantidad:       3,
		PrecioUnitario: 3500,
		Estado:         "activo",
	})
	if err == nil {
		t.Fatal("expected stock insufficient for combo ingredients, got nil")
	}
	if err != ErrStockInsuficiente {
		t.Fatalf("expected ErrStockInsuficiente, got %v", err)
	}

	stockFinal := stockTotalByProducto(t, dbConn, 1, insumoID)
	if stockFinal != 5 {
		t.Fatalf("expected unchanged ingredient stock 5 after failed combo add, got %.2f", stockFinal)
	}
}

func TestCarritoEstadoVentaLifecycle(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCarritosSchema(dbConn); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	carritoAbiertoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:      1,
		Codigo:         "CAR-LIFE-001",
		Nombre:         "Carrito Lifecycle A",
		CanalVenta:     "mostrador",
		Moneda:         "COP",
		UsuarioCreador: "test",
	})
	if err != nil {
		t.Fatalf("create carrito abierto: %v", err)
	}

	carritoAbierto, err := GetCarritoCompraByID(dbConn, 1, carritoAbiertoID)
	if err != nil {
		t.Fatalf("get carrito abierto: %v", err)
	}
	if carritoAbierto.EstadoVenta != "venta_abierta" {
		t.Fatalf("expected venta_abierta, got %q", carritoAbierto.EstadoVenta)
	}

	if err := SetCarritoOperacionEstado(dbConn, 1, carritoAbiertoID, "cerrado"); err != nil {
		t.Fatalf("cerrar carrito: %v", err)
	}
	carritoCerrado, err := GetCarritoCompraByID(dbConn, 1, carritoAbiertoID)
	if err != nil {
		t.Fatalf("get carrito cerrado: %v", err)
	}
	if carritoCerrado.EstadoVenta != "venta_cerrada" {
		t.Fatalf("expected venta_cerrada, got %q", carritoCerrado.EstadoVenta)
	}

	if err := PayCarritoStationSession(dbConn, 1, carritoAbiertoID, "efectivo", "", "", "", 0, 0, carritoCerrado.Total, 0); err != nil {
		t.Fatalf("pagar carrito: %v", err)
	}
	carritoPagado, err := GetCarritoCompraByID(dbConn, 1, carritoAbiertoID)
	if err != nil {
		t.Fatalf("get carrito pagado: %v", err)
	}
	if carritoPagado.EstadoVenta != "venta_pagada" {
		t.Fatalf("expected venta_pagada, got %q", carritoPagado.EstadoVenta)
	}

	carritoSuspendidoID, err := CreateCarritoCompra(dbConn, CarritoCompra{
		EmpresaID:      1,
		Codigo:         "CAR-LIFE-002",
		Nombre:         "Carrito Lifecycle B",
		CanalVenta:     "mostrador",
		Moneda:         "COP",
		UsuarioCreador: "test",
	})
	if err != nil {
		t.Fatalf("create carrito suspendido: %v", err)
	}
	if err := SetCarritoCompraEstado(dbConn, 1, carritoSuspendidoID, "inactivo"); err != nil {
		t.Fatalf("desactivar carrito: %v", err)
	}
	carritoSuspendido, err := GetCarritoCompraByID(dbConn, 1, carritoSuspendidoID)
	if err != nil {
		t.Fatalf("get carrito suspendido: %v", err)
	}
	if carritoSuspendido.EstadoVenta != "venta_suspendida" {
		t.Fatalf("expected venta_suspendida, got %q", carritoSuspendido.EstadoVenta)
	}
}
