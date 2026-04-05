package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	_ "modernc.org/sqlite"
)

const seedUser = "seed_motel_malibu"

type productSeed struct {
	SKU          string
	CodigoBarras string
	Nombre       string
	Descripcion  string
	Categoria    string
	Marca        string
	Unidad       string
	Costo        float64
	Precio       float64
	Impuesto     float64
	StockInicial float64
}

type clientSeed struct {
	TipoDocumento     string
	NumeroDocumento   string
	TipoPersona       string
	NombreRazonSocial string
	NombreComercial   string
	RegimenFiscal     string
	Responsabilidad   string
	Email             string
	Telefono          string
	Direccion         string
	Departamento      string
	Municipio         string
	CodigoPostal      string
	Observaciones     string
}

type userSeed struct {
	Email       string
	Nombre      string
	Documento   string
	Rol         string
	Observacion string
}

type productIndex struct {
	ID       int64
	SKU      string
	Nombre   string
	Unidad   string
	Precio   float64
	Impuesto float64
}

type saleSummary struct {
	CarritoID     int64
	Codigo        string
	ClienteNombre string
	Total         float64
	PagadoEn      string
	Items         []dbpkg.CarritoCompraItem
}

type saleStockValidation struct {
	SKU       string
	BeforeAdd float64
	AfterAdd  float64
	AfterPay  float64
	SoldQty   float64
}

var demoProducts = []productSeed{
	{SKU: "MM-COCA-350", CodigoBarras: "7700000000011", Nombre: "Coca-Cola 350ml", Descripcion: "Gaseosa Coca-Cola personal", Categoria: "Bebidas", Marca: "Coca-Cola", Unidad: "unidad", Costo: 2200, Precio: 3500, Impuesto: 19, StockInicial: 40},
	{SKU: "MM-COCA-1500", CodigoBarras: "7700000000028", Nombre: "Coca-Cola 1.5L", Descripcion: "Gaseosa Coca-Cola familiar", Categoria: "Bebidas", Marca: "Coca-Cola", Unidad: "unidad", Costo: 5200, Precio: 7800, Impuesto: 19, StockInicial: 25},
	{SKU: "MM-PEPSI-400", CodigoBarras: "7700000000035", Nombre: "Pepsi 400ml", Descripcion: "Gaseosa Pepsi botella", Categoria: "Bebidas", Marca: "Pepsi", Unidad: "unidad", Costo: 2000, Precio: 3200, Impuesto: 19, StockInicial: 35},
	{SKU: "MM-AGUA-600", CodigoBarras: "7700000000042", Nombre: "Agua Manantial 600ml", Descripcion: "Agua sin gas", Categoria: "Bebidas", Marca: "Manantial", Unidad: "unidad", Costo: 1500, Precio: 2500, Impuesto: 19, StockInicial: 35},
	{SKU: "MM-HIT-MANGO", CodigoBarras: "7700000000059", Nombre: "Jugo Hit Mango 500ml", Descripcion: "Jugo Hit sabor mango", Categoria: "Bebidas", Marca: "Hit", Unidad: "unidad", Costo: 1800, Precio: 3000, Impuesto: 19, StockInicial: 30},
	{SKU: "MM-PAPAS-40", CodigoBarras: "7700000000066", Nombre: "Papas Margarita 40g", Descripcion: "Papas fritas paquete personal", Categoria: "Snacks", Marca: "Margarita", Unidad: "unidad", Costo: 1600, Precio: 2800, Impuesto: 19, StockInicial: 45},
	{SKU: "MM-FESTIVAL-6", CodigoBarras: "7700000000073", Nombre: "Galletas Festival 6u", Descripcion: "Galletas rellenas surtidas", Categoria: "Snacks", Marca: "Festival", Unidad: "unidad", Costo: 1700, Precio: 3000, Impuesto: 19, StockInicial: 30},
	{SKU: "MM-JET-12", CodigoBarras: "7700000000080", Nombre: "Chocolatina Jet 12g", Descripcion: "Chocolate en barra personal", Categoria: "Snacks", Marca: "Jet", Unidad: "unidad", Costo: 700, Precio: 1200, Impuesto: 19, StockInicial: 80},
	{SKU: "MM-AGUILA-330", CodigoBarras: "7700000000097", Nombre: "Cerveza Aguila lata 330ml", Descripcion: "Cerveza rubia en lata", Categoria: "Bebidas", Marca: "Aguila", Unidad: "unidad", Costo: 2800, Precio: 4500, Impuesto: 19, StockInicial: 35},
	{SKU: "MM-DUREX-3", CodigoBarras: "7700000000103", Nombre: "Preservativo Durex 3u", Descripcion: "Caja preservativos x3", Categoria: "Habitacion", Marca: "Durex", Unidad: "caja", Costo: 11000, Precio: 18000, Impuesto: 19, StockInicial: 20},
}

var demoClients = []clientSeed{
	{TipoDocumento: "CC", NumeroDocumento: "1020304050", TipoPersona: "natural", NombreRazonSocial: "Juan Carlos Perez", NombreComercial: "Juan Perez", RegimenFiscal: "No responsable de IVA", Responsabilidad: "R-99-PN", Email: "juan.perez@example.com", Telefono: "3001112233", Direccion: "Calle 50 # 20-10", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050001", Observaciones: "Cliente frecuente"},
	{TipoDocumento: "CC", NumeroDocumento: "1030304050", TipoPersona: "natural", NombreRazonSocial: "Maria Fernanda Gomez", NombreComercial: "Maria Gomez", RegimenFiscal: "No responsable de IVA", Responsabilidad: "R-99-PN", Email: "maria.gomez@example.com", Telefono: "3012223344", Direccion: "Carrera 34 # 45-12", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050021", Observaciones: "Cliente de pruebas"},
	{TipoDocumento: "CE", NumeroDocumento: "50012345", TipoPersona: "natural", NombreRazonSocial: "Carlos Andres Ruiz", NombreComercial: "Carlos Ruiz", RegimenFiscal: "No responsable de IVA", Responsabilidad: "R-99-PN", Email: "carlos.ruiz@example.com", Telefono: "3023334455", Direccion: "Avenida 80 # 10-55", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050041", Observaciones: "Documento extranjero"},
	{TipoDocumento: "NIT", NumeroDocumento: "900222333", TipoPersona: "juridica", NombreRazonSocial: "Empresas Viajeras SAS", NombreComercial: "Empresas Viajeras", RegimenFiscal: "Responsable de IVA", Responsabilidad: "O-13", Email: "facturas@empresasviajeras.com", Telefono: "6044445566", Direccion: "Calle 10 # 40-90", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050024", Observaciones: "Cliente corporativo"},
	{TipoDocumento: "CC", NumeroDocumento: "999999999", TipoPersona: "natural", NombreRazonSocial: "Consumidor Final Motel Malibu", NombreComercial: "Consumidor Final", RegimenFiscal: "No responsable de IVA", Responsabilidad: "R-99-PN", Email: "consumidor.final@example.com", Telefono: "3009990001", Direccion: "Sin direccion", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050000", Observaciones: "Cliente mostrador"},
	{TipoDocumento: "CC", NumeroDocumento: "1040304050", TipoPersona: "natural", NombreRazonSocial: "Laura Paola Ramirez", NombreComercial: "Laura Ramirez", RegimenFiscal: "No responsable de IVA", Responsabilidad: "R-99-PN", Email: "laura.ramirez@example.com", Telefono: "3045556677", Direccion: "Calle 33 # 70-44", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050031", Observaciones: "Cliente recurrente"},
	{TipoDocumento: "CC", NumeroDocumento: "1050304050", TipoPersona: "natural", NombreRazonSocial: "Andres Felipe Quintero", NombreComercial: "Andres Quintero", RegimenFiscal: "No responsable de IVA", Responsabilidad: "R-99-PN", Email: "andres.quintero@example.com", Telefono: "3056667788", Direccion: "Calle 74 # 52-11", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050012", Observaciones: "Cliente nocturno"},
	{TipoDocumento: "NIT", NumeroDocumento: "901444555", TipoPersona: "juridica", NombreRazonSocial: "Logistica Express LATAM SAS", NombreComercial: "Logistica Express", RegimenFiscal: "Responsable de IVA", Responsabilidad: "O-13", Email: "compras@logisticaexpress.com", Telefono: "6047778899", Direccion: "Carrera 65 # 98-21", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050052", Observaciones: "Cliente corporativo premium"},
	{TipoDocumento: "CC", NumeroDocumento: "1060304050", TipoPersona: "natural", NombreRazonSocial: "Paula Andrea Cardenas", NombreComercial: "Paula Cardenas", RegimenFiscal: "No responsable de IVA", Responsabilidad: "R-99-PN", Email: "paula.cardenas@example.com", Telefono: "3101122334", Direccion: "Transversal 39 # 45-60", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050025", Observaciones: "Cliente recomendada"},
	{TipoDocumento: "CC", NumeroDocumento: "1070304050", TipoPersona: "natural", NombreRazonSocial: "Miguel Angel Restrepo", NombreComercial: "Miguel Restrepo", RegimenFiscal: "No responsable de IVA", Responsabilidad: "R-99-PN", Email: "miguel.restrepo@example.com", Telefono: "3112233445", Direccion: "Diagonal 78 # 65-20", Departamento: "Antioquia", Municipio: "Medellin", CodigoPostal: "050045", Observaciones: "Cliente corporativo ocasional"},
}

var demoUsers = []userSeed{
	{Email: "recepcion1@motelmalibu.local", Nombre: "Recepcion Uno", Documento: "MM-USER-001", Rol: "recepcion", Observacion: "Usuario operativo de recepción"},
	{Email: "recepcion2@motelmalibu.local", Nombre: "Recepcion Dos", Documento: "MM-USER-002", Rol: "recepcion", Observacion: "Turno alterno de recepción"},
	{Email: "cajero1@motelmalibu.local", Nombre: "Cajero Principal", Documento: "MM-USER-003", Rol: "cajero", Observacion: "Caja principal"},
	{Email: "cajero2@motelmalibu.local", Nombre: "Cajero Nocturno", Documento: "MM-USER-004", Rol: "cajero", Observacion: "Caja nocturna"},
	{Email: "supervisor@motelmalibu.local", Nombre: "Supervisor Operativo", Documento: "MM-USER-005", Rol: "supervisor", Observacion: "Control de operación"},
	{Email: "ventas1@motelmalibu.local", Nombre: "Asesor Ventas Uno", Documento: "MM-USER-006", Rol: "vendedor", Observacion: "Ventas mostrador"},
	{Email: "ventas2@motelmalibu.local", Nombre: "Asesor Ventas Dos", Documento: "MM-USER-007", Rol: "vendedor", Observacion: "Ventas turno tarde"},
	{Email: "inventario@motelmalibu.local", Nombre: "Auxiliar Inventario", Documento: "MM-USER-008", Rol: "inventario", Observacion: "Control de bodega"},
	{Email: "facturacion@motelmalibu.local", Nombre: "Analista Facturación", Documento: "MM-USER-009", Rol: "facturacion", Observacion: "Emisión de documentos"},
	{Email: "gerencia@motelmalibu.local", Nombre: "Gerencia Empresa", Documento: "MM-USER-010", Rol: "administrador", Observacion: "Gestión general"},
}

func normalizeText(v string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(strings.ToLower(v))), " ")
}

func shortCode(v string) string {
	clean := strings.ToUpper(strings.TrimSpace(v))
	clean = strings.ReplaceAll(clean, " ", "")
	if len(clean) <= 4 {
		return clean
	}
	return clean[:4]
}

func getTableColumns(dbConn *sql.DB, tableName string) (map[string]bool, error) {
	rows, err := dbConn.Query(`PRAGMA table_info(` + tableName + `)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		out[strings.TrimSpace(name)] = true
	}
	return out, rows.Err()
}

func ensureUsersSchema(dbConn *sql.DB) error {
	createUsers := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT DEFAULT 'administrador',
		empresa_id INTEGER,
		documento_identidad TEXT,
		password_hash TEXT,
		password_salt TEXT,
		password_set INTEGER DEFAULT 0,
		password_actualizada_en TEXT,
		rol_usuario_id INTEGER,
		email_confirmado INTEGER DEFAULT 0,
		email_confirm_token TEXT,
		email_confirm_expira TEXT,
		email_confirmado_en TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbConn.Exec(createUsers); err != nil {
		return err
	}

	columns, err := getTableColumns(dbConn, "users")
	if err != nil {
		return err
	}

	required := []struct {
		name string
		def  string
	}{
		{name: "documento_identidad", def: "documento_identidad TEXT"},
		{name: "rol_usuario_id", def: "rol_usuario_id INTEGER"},
		{name: "email_confirmado", def: "email_confirmado INTEGER DEFAULT 0"},
		{name: "email_confirm_token", def: "email_confirm_token TEXT"},
		{name: "email_confirm_expira", def: "email_confirm_expira TEXT"},
		{name: "email_confirmado_en", def: "email_confirmado_en TEXT"},
		{name: "password_hash", def: "password_hash TEXT"},
		{name: "password_salt", def: "password_salt TEXT"},
		{name: "password_set", def: "password_set INTEGER DEFAULT 0"},
		{name: "password_actualizada_en", def: "password_actualizada_en TEXT"},
		{name: "usuario_creador", def: "usuario_creador TEXT"},
		{name: "estado", def: "estado TEXT DEFAULT 'activo'"},
		{name: "observaciones", def: "observaciones TEXT"},
		{name: "fecha_actualizacion", def: "fecha_actualizacion TEXT"},
	}

	for _, col := range required {
		if columns[col.name] {
			continue
		}
		if _, err := dbConn.Exec(`ALTER TABLE users ADD COLUMN ` + col.def); err != nil {
			return err
		}
	}

	return nil
}

func ensureSchemas(dbConn *sql.DB) error {
	if err := dbpkg.EnsureEmpresasScopeReferences(dbConn); err != nil {
		return err
	}
	if err := ensureUsersSchema(dbConn); err != nil {
		return err
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbConn); err != nil {
		return err
	}
	if err := dbpkg.EnsureEmpresaClientesSchema(dbConn); err != nil {
		return err
	}
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbConn); err != nil {
		return err
	}
	if err := dbpkg.EnsureEmpresaConfiguracionAvanzadaSchema(dbConn); err != nil {
		return err
	}
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbConn); err != nil {
		return err
	}
	return nil
}

func findOrCreateEmpresa(dbConn *sql.DB, objetivo string) (int64, string, bool, error) {
	empresas, err := dbpkg.GetEmpresas(dbConn)
	if err != nil {
		return 0, "", false, err
	}
	normObjetivo := normalizeText(objetivo)
	for _, e := range empresas {
		normNombre := normalizeText(e.Nombre)
		if normNombre == normObjetivo || strings.Contains(normNombre, normObjetivo) {
			if e.EmpresaID > 0 {
				return e.EmpresaID, e.Nombre, false, nil
			}
			return e.ID, e.Nombre, false, nil
		}
	}

	nit := fmt.Sprintf("901%06d-1", time.Now().Unix()%1000000)
	id, err := dbpkg.CreateEmpresa(dbConn, 0, "Hospedaje", objetivo, nit, "Empresa creada para seed de datos demo", seedUser)
	if err != nil {
		return 0, "", false, err
	}
	return id, objetivo, true, nil
}

func ensureBodegaPrincipal(dbConn *sql.DB, empresaID int64) (int64, bool, error) {
	bodegas, err := dbpkg.GetBodegasByEmpresa(dbConn, empresaID, true)
	if err != nil {
		return 0, false, err
	}
	for _, b := range bodegas {
		if normalizeText(b.Nombre) == "bodega principal" || strings.EqualFold(strings.TrimSpace(b.Codigo), "BOD-MALIBU-PRINCIPAL") {
			return b.ID, false, nil
		}
	}
	id, err := dbpkg.CreateBodega(dbConn, dbpkg.Bodega{
		EmpresaID:      empresaID,
		Codigo:         "BOD-MALIBU-PRINCIPAL",
		Nombre:         "Bodega principal",
		Ubicacion:      "Bodega central Motel Malibu",
		Responsable:    "Administrador",
		UsuarioCreador: seedUser,
		Estado:         "activo",
		Observaciones:  "Bodega semilla para inventario demo",
	})
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func ensureProveedorPrincipal(dbConn *sql.DB, empresaID int64) (int64, bool, error) {
	proveedores, err := dbpkg.GetProveedoresByEmpresa(dbConn, empresaID, true)
	if err != nil {
		return 0, false, err
	}
	for _, p := range proveedores {
		if normalizeText(p.Nombre) == "proveedor local medellin" || strings.EqualFold(strings.TrimSpace(p.Codigo), "PROV-MALIBU") {
			return p.ID, false, nil
		}
	}
	id, err := dbpkg.CreateProveedor(dbConn, dbpkg.Proveedor{
		EmpresaID:      empresaID,
		Codigo:         "PROV-MALIBU",
		Nombre:         "Proveedor Local Medellin",
		Documento:      "NIT 900111222",
		Contacto:       "Bodega comercial",
		Telefono:       "6045556677",
		Email:          "compras@proveedorlocal.com",
		Direccion:      "Medellin, Antioquia",
		UsuarioCreador: seedUser,
		Estado:         "activo",
		Observaciones:  "Proveedor semilla para pruebas",
	})
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func ensureCategorias(dbConn *sql.DB, empresaID int64, categorias []string) (map[string]int64, int, error) {
	existing, err := dbpkg.GetCategoriasProductoByEmpresa(dbConn, empresaID, true, "")
	if err != nil {
		return nil, 0, err
	}
	byName := make(map[string]int64)
	for _, c := range existing {
		byName[normalizeText(c.Nombre)] = c.ID
	}

	created := 0
	for i, nombre := range categorias {
		key := normalizeText(nombre)
		if key == "" || byName[key] > 0 {
			continue
		}
		id, err := dbpkg.CreateCategoriaProducto(dbConn, dbpkg.CategoriaProducto{
			EmpresaID:      empresaID,
			Codigo:         "CAT-" + shortCode(fmt.Sprintf("%s-%d", nombre, i+1)),
			Nombre:         nombre,
			Descripcion:    "Categoria semilla para catalogo Motel Malibu",
			ColorHex:       "#1f8ef1",
			Orden:          int64(i + 1),
			UsuarioCreador: seedUser,
			Estado:         "activo",
			Observaciones:  "Creada por seeder",
		})
		if err != nil {
			return nil, created, err
		}
		byName[key] = id
		created++
	}
	return byName, created, nil
}

func loadProductIndex(dbConn *sql.DB, empresaID int64) (map[string]productIndex, error) {
	rows, err := dbConn.Query(`SELECT id, COALESCE(sku, ''), COALESCE(nombre, ''), COALESCE(unidad_medida, 'unidad'), COALESCE(precio, 0), COALESCE(impuesto_porcentaje, 0)
		FROM productos WHERE empresa_id = ?`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]productIndex)
	for rows.Next() {
		var p productIndex
		if err := rows.Scan(&p.ID, &p.SKU, &p.Nombre, &p.Unidad, &p.Precio, &p.Impuesto); err != nil {
			return nil, err
		}
		key := strings.ToUpper(strings.TrimSpace(p.SKU))
		if key != "" {
			out[key] = p
		}
	}
	return out, rows.Err()
}

func seedProductos(dbConn *sql.DB, empresaID, bodegaID, proveedorID int64, categoriaIDs map[string]int64) (int, int, map[string]productIndex, error) {
	existing, err := loadProductIndex(dbConn, empresaID)
	if err != nil {
		return 0, 0, nil, err
	}

	created := 0
	skipped := 0
	for _, item := range demoProducts {
		sku := strings.ToUpper(strings.TrimSpace(item.SKU))
		if sku == "" {
			continue
		}
		if _, ok := existing[sku]; ok {
			skipped++
			continue
		}
		catID := categoriaIDs[normalizeText(item.Categoria)]
		prod := dbpkg.Producto{
			EmpresaID:            empresaID,
			BodegaPrincipalID:    bodegaID,
			ProveedorPrincipalID: proveedorID,
			CategoriaID:          catID,
			SKU:                  item.SKU,
			CodigoBarras:         item.CodigoBarras,
			Nombre:               item.Nombre,
			Descripcion:          item.Descripcion,
			Categoria:            item.Categoria,
			Marca:                item.Marca,
			UnidadMedida:         item.Unidad,
			Costo:                item.Costo,
			Precio:               item.Precio,
			ImpuestoPorcentaje:   item.Impuesto,
			StockMinimo:          2,
			StockMaximo:          120,
			UsuarioCreador:       seedUser,
			Estado:               "activo",
			Observaciones:        "Producto semilla de referencia de precios COL",
		}
		if _, err := dbpkg.CreateProducto(dbConn, prod, item.StockInicial, "SEED_MOTEL_MALIBU"); err != nil {
			return created, skipped, nil, err
		}
		created++
	}

	finalIndex, err := loadProductIndex(dbConn, empresaID)
	if err != nil {
		return created, skipped, nil, err
	}
	return created, skipped, finalIndex, nil
}

func seedClientes(dbConn *sql.DB, empresaID int64) (int, int, int64, error) {
	existingKeys := map[string]bool{}
	rows, err := dbConn.Query(`SELECT COALESCE(tipo_documento, 'NIT'), COALESCE(numero_documento, '') FROM clientes WHERE empresa_id = ?`, empresaID)
	if err != nil {
		return 0, 0, 0, err
	}
	for rows.Next() {
		var tipo, numero string
		if err := rows.Scan(&tipo, &numero); err != nil {
			rows.Close()
			return 0, 0, 0, err
		}
		existingKeys[strings.ToUpper(strings.TrimSpace(tipo))+"|"+strings.TrimSpace(numero)] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, 0, 0, err
	}
	if err := rows.Close(); err != nil {
		return 0, 0, 0, err
	}

	created := 0
	skipped := 0
	for _, c := range demoClients {
		key := strings.ToUpper(strings.TrimSpace(c.TipoDocumento)) + "|" + strings.TrimSpace(c.NumeroDocumento)
		if existingKeys[key] {
			skipped++
			continue
		}
		payload := dbpkg.Cliente{
			EmpresaID:                 empresaID,
			TipoDocumento:             c.TipoDocumento,
			NumeroDocumento:           c.NumeroDocumento,
			TipoPersona:               c.TipoPersona,
			NombreRazonSocial:         c.NombreRazonSocial,
			NombreComercial:           c.NombreComercial,
			RegimenFiscal:             c.RegimenFiscal,
			ResponsabilidadTributaria: c.Responsabilidad,
			Email:                     c.Email,
			Telefono:                  c.Telefono,
			Direccion:                 c.Direccion,
			Pais:                      "CO",
			Departamento:              c.Departamento,
			Municipio:                 c.Municipio,
			CodigoPostal:              c.CodigoPostal,
			UsuarioCreador:            seedUser,
			Estado:                    "activo",
			Observaciones:             c.Observaciones,
		}
		if _, err := dbpkg.CreateCliente(dbConn, payload); err != nil {
			return created, skipped, 0, err
		}
		created++
		existingKeys[key] = true
	}

	var clienteID int64
	if err := dbConn.QueryRow(`SELECT id FROM clientes WHERE empresa_id = ? AND COALESCE(estado, 'activo') = 'activo' ORDER BY id ASC LIMIT 1`, empresaID).Scan(&clienteID); err != nil {
		return created, skipped, 0, err
	}
	return created, skipped, clienteID, nil
}

func seedUsuariosEmpresa(dbConn *sql.DB, empresaID int64) (int, int, error) {
	existing := map[string]bool{}
	rows, err := dbConn.Query(`SELECT lower(trim(email)) FROM users WHERE empresa_id = ?`, empresaID)
	if err != nil {
		return 0, 0, err
	}
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			rows.Close()
			return 0, 0, err
		}
		existing[strings.TrimSpace(strings.ToLower(email))] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, 0, err
	}
	if err := rows.Close(); err != nil {
		return 0, 0, err
	}

	created := 0
	skipped := 0
	for _, u := range demoUsers {
		email := strings.TrimSpace(strings.ToLower(u.Email))
		if email == "" {
			skipped++
			continue
		}
		if existing[email] {
			skipped++
			continue
		}

		_, err := dbConn.Exec(`INSERT INTO users (
			email,
			name,
			role,
			empresa_id,
			documento_identidad,
			rol_usuario_id,
			email_confirmado,
			email_confirmado_en,
			password_set,
			usuario_creador,
			estado,
			observaciones,
			fecha_creacion,
			fecha_actualizacion
		) VALUES (?, ?, ?, ?, ?, 0, 1, datetime('now','localtime'), 0, ?, 'activo', ?, datetime('now','localtime'), datetime('now','localtime'))`,
			u.Email,
			u.Nombre,
			u.Rol,
			empresaID,
			u.Documento,
			seedUser,
			u.Observacion,
		)
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				skipped++
				continue
			}
			return created, skipped, err
		}

		existing[email] = true
		created++
	}

	return created, skipped, nil
}

func getProductoStockTotal(dbConn *sql.DB, empresaID, productoID int64) (float64, error) {
	var total float64
	err := dbConn.QueryRow(`SELECT COALESCE(SUM(cantidad), 0) FROM inventario_existencias WHERE empresa_id = ? AND producto_id = ?`, empresaID, productoID).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func createVentaDemo(dbConn *sql.DB, empresaID, clienteID int64, products map[string]productIndex) (*saleSummary, []saleStockValidation, error) {
	if clienteID <= 0 {
		return nil, nil, fmt.Errorf("no hay cliente activo para asociar la venta")
	}

	saleItemsConfig := []struct {
		SKU      string
		Cantidad float64
	}{
		{SKU: "MM-COCA-350", Cantidad: 2},
		{SKU: "MM-PAPAS-40", Cantidad: 1},
		{SKU: "MM-DUREX-3", Cantidad: 1},
	}

	for _, item := range saleItemsConfig {
		if _, ok := products[item.SKU]; !ok {
			return nil, nil, fmt.Errorf("producto requerido para venta demo no encontrado: %s", item.SKU)
		}
	}

	stockBeforeAdd := map[string]float64{}
	for _, cfg := range saleItemsConfig {
		prod := products[cfg.SKU]
		stock, err := getProductoStockTotal(dbConn, empresaID, prod.ID)
		if err != nil {
			return nil, nil, err
		}
		stockBeforeAdd[cfg.SKU] = stock
	}

	carritoCode := "MM-VTA-" + time.Now().Format("20060102150405")
	carritoNombre := "Venta demo Motel Malibu " + time.Now().Format("20060102150405")
	carritoID, err := dbpkg.CreateCarritoCompra(dbConn, dbpkg.CarritoCompra{
		EmpresaID:         empresaID,
		Codigo:            carritoCode,
		Nombre:            carritoNombre,
		CanalVenta:        "mostrador",
		ClienteID:         clienteID,
		Moneda:            "COP",
		ReferenciaExterna: "SEED_MOTEL_MALIBU",
		UsuarioCreador:    seedUser,
		Observaciones:     "Venta de prueba automatica para reportes",
	})
	if err != nil {
		return nil, nil, err
	}

	for _, cfg := range saleItemsConfig {
		prod := products[cfg.SKU]
		_, err := dbpkg.CreateCarritoCompraItem(dbConn, dbpkg.CarritoCompraItem{
			EmpresaID:          empresaID,
			CarritoID:          carritoID,
			TipoItem:           "producto",
			ReferenciaID:       prod.ID,
			CodigoItem:         prod.SKU,
			Descripcion:        prod.Nombre,
			UnidadMedida:       prod.Unidad,
			Cantidad:           cfg.Cantidad,
			PrecioUnitario:     prod.Precio,
			ImpuestoPorcentaje: prod.Impuesto,
			ImpuestoCodigo:     "IVA",
			UsuarioCreador:     seedUser,
			Estado:             "activo",
			Observaciones:      "Item semilla para venta demo",
		})
		if err != nil {
			return nil, nil, err
		}
	}

	stockAfterAdd := map[string]float64{}
	for _, cfg := range saleItemsConfig {
		prod := products[cfg.SKU]
		stock, err := getProductoStockTotal(dbConn, empresaID, prod.ID)
		if err != nil {
			return nil, nil, err
		}
		stockAfterAdd[cfg.SKU] = stock
		expected := stockBeforeAdd[cfg.SKU] - cfg.Cantidad
		if absFloat(stock-stockBeforeAdd[cfg.SKU]) < 0.000001 {
			return nil, nil, fmt.Errorf("el stock no cambio al agregar al carrito para %s", cfg.SKU)
		}
		if absFloat(stock-expected) > 0.000001 {
			return nil, nil, fmt.Errorf("stock inesperado para %s despues de agregar al carrito: esperado %.2f, actual %.2f", cfg.SKU, expected, stock)
		}
	}

	beforePay, err := dbpkg.GetCarritoCompraByID(dbConn, empresaID, carritoID)
	if err != nil {
		return nil, nil, err
	}
	if err := dbpkg.PayCarritoStationSession(dbConn, empresaID, carritoID, "efectivo", "", "", "", 0, 0, beforePay.Total, 0); err != nil {
		return nil, nil, err
	}
	afterPay, err := dbpkg.GetCarritoCompraByID(dbConn, empresaID, carritoID)
	if err != nil {
		return nil, nil, err
	}
	if strings.ToLower(strings.TrimSpace(afterPay.EstadoCarrito)) != "cerrado" {
		return nil, nil, fmt.Errorf("la venta demo no quedo cerrada")
	}

	stockAfterPay := map[string]float64{}
	for _, cfg := range saleItemsConfig {
		prod := products[cfg.SKU]
		stock, err := getProductoStockTotal(dbConn, empresaID, prod.ID)
		if err != nil {
			return nil, nil, err
		}
		stockAfterPay[cfg.SKU] = stock
		if absFloat(stock-stockAfterAdd[cfg.SKU]) > 0.000001 {
			return nil, nil, fmt.Errorf("el stock cambio inesperadamente al pagar para %s", cfg.SKU)
		}
	}

	items, err := dbpkg.GetCarritoCompraItems(dbConn, empresaID, carritoID, true)
	if err != nil {
		return nil, nil, err
	}

	summary := &saleSummary{
		CarritoID:     afterPay.ID,
		Codigo:        afterPay.Codigo,
		ClienteNombre: afterPay.ClienteNombre,
		Total:         afterPay.TotalPagado,
		PagadoEn:      afterPay.PagadoEn,
		Items:         items,
	}

	validations := make([]saleStockValidation, 0, len(saleItemsConfig))
	for _, cfg := range saleItemsConfig {
		validations = append(validations, saleStockValidation{
			SKU:       cfg.SKU,
			BeforeAdd: stockBeforeAdd[cfg.SKU],
			AfterAdd:  stockAfterAdd[cfg.SKU],
			AfterPay:  stockAfterPay[cfg.SKU],
			SoldQty:   cfg.Cantidad,
		})
	}

	return summary, validations, nil
}

func ensurePrintConfig(dbConn *sql.DB, empresaID int64, empresaNombre string) (*dbpkg.EmpresaConfiguracionAvanzada, error) {
	cfg, err := dbpkg.GetEmpresaConfiguracionAvanzada(dbConn, empresaID)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &dbpkg.EmpresaConfiguracionAvanzada{}
	}

	cfg.EmpresaID = empresaID
	if strings.TrimSpace(cfg.RazonSocial) == "" {
		cfg.RazonSocial = empresaNombre
	}
	if strings.TrimSpace(cfg.NombreComercial) == "" {
		cfg.NombreComercial = empresaNombre
	}
	if strings.TrimSpace(cfg.NIT) == "" {
		cfg.NIT = "901999888-1"
	}
	if strings.TrimSpace(cfg.PaisCodigo) == "" {
		cfg.PaisCodigo = "CO"
	}
	if strings.TrimSpace(cfg.EmailFacturacion) == "" {
		cfg.EmailFacturacion = "facturacion@motelmalibu.com"
	}
	if strings.TrimSpace(cfg.TelefonoFacturacion) == "" {
		cfg.TelefonoFacturacion = "6045558899"
	}
	if strings.TrimSpace(cfg.DireccionFiscal) == "" {
		cfg.DireccionFiscal = "Calle 48 # 73-21"
	}

	cfg.FormatoImpresion = "pos"
	cfg.ImprimirCopiaFactura = true
	cfg.MostrarLogo = true
	if strings.TrimSpace(cfg.LogoURL) == "" {
		cfg.LogoURL = "/img/company-briefcase.svg"
	}
	if strings.TrimSpace(cfg.PieFactura) == "" {
		cfg.PieFactura = "Gracias por su visita. Regrese pronto."
	}
	cfg.UsuarioCreador = seedUser
	if strings.TrimSpace(cfg.Estado) == "" {
		cfg.Estado = "activo"
	}
	cfg.Observaciones = strings.TrimSpace(cfg.Observaciones + " | validado formato impresion por seeder")

	if _, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbConn, *cfg); err != nil {
		return nil, err
	}
	return dbpkg.GetEmpresaConfiguracionAvanzada(dbConn, empresaID)
}

func renderPosPreview(empresaNombre string, sale *saleSummary, cfg *dbpkg.EmpresaConfiguracionAvanzada) string {
	var b strings.Builder
	b.WriteString("FORMATO POS (SIMULACION)\n")
	b.WriteString("Empresa: " + empresaNombre + "\n")
	b.WriteString("Carrito: " + sale.Codigo + "\n")
	b.WriteString("Cliente: " + sale.ClienteNombre + "\n")
	b.WriteString("-------------------------------\n")
	for _, it := range sale.Items {
		line := fmt.Sprintf("%0.0fx %s = $%0.0f\n", it.Cantidad, it.Descripcion, it.TotalLinea)
		b.WriteString(line)
	}
	b.WriteString("-------------------------------\n")
	b.WriteString(fmt.Sprintf("TOTAL: $%0.0f COP\n", sale.Total))
	if cfg != nil && cfg.ImprimirCopiaFactura {
		b.WriteString("Imprimir copia: SI\n")
	}
	if cfg != nil && strings.TrimSpace(cfg.PieFactura) != "" {
		b.WriteString("Pie: " + cfg.PieFactura + "\n")
	}
	return b.String()
}

func renderCartaPreview(empresaNombre string, sale *saleSummary, cfg *dbpkg.EmpresaConfiguracionAvanzada) string {
	var b strings.Builder
	b.WriteString("FORMATO CARTA (SIMULACION)\n")
	b.WriteString("Empresa: " + empresaNombre + "\n")
	b.WriteString("Fecha pago: " + sale.PagadoEn + "\n")
	b.WriteString("Carrito: " + sale.Codigo + "\n")
	b.WriteString("Cliente: " + sale.ClienteNombre + "\n\n")
	b.WriteString("DETALLE\n")
	for _, it := range sale.Items {
		b.WriteString(fmt.Sprintf("- %s | Cant: %0.2f | Vlr U: $%0.0f | Total: $%0.0f\n", it.Descripcion, it.Cantidad, it.PrecioUnitario, it.TotalLinea))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Total pagado: $%0.0f COP\n", sale.Total))
	if cfg != nil && strings.TrimSpace(cfg.NotasLegales) != "" {
		b.WriteString("Notas legales: " + cfg.NotasLegales + "\n")
	}
	return b.String()
}

func seedMovimientosFinancieros(dbConn *sql.DB, empresaID int64) (int, int, float64, float64, error) {
	cfg, err := dbpkg.GetEmpresaFinanzasConfiguracion(dbConn, empresaID)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	cfg.EmpresaID = empresaID
	cfg.HabilitarIngresos = true
	cfg.HabilitarEgresos = true
	cfg.Moneda = "COP"
	cfg.CategoriasIngreso = "ventas\nservicios\notros ingresos"
	cfg.CategoriasEgreso = "compras\nservicios\nnomina\notros gastos"
	cfg.PrefijoIngreso = "ING"
	cfg.PrefijoEgreso = "EGR"
	cfg.FormatoImpresion = "carta"
	cfg.RequiereAprobacion = false
	cfg.UsuarioCreador = seedUser
	cfg.Estado = "activo"
	cfg.Observaciones = strings.TrimSpace(cfg.Observaciones + " | configuracion financiera validada por seeder")
	if _, err := dbpkg.UpsertEmpresaFinanzasConfiguracion(dbConn, *cfg); err != nil {
		return 0, 0, 0, 0, err
	}

	var existing int
	err = dbConn.QueryRow(`SELECT COUNT(1) FROM empresa_finanzas_movimientos WHERE empresa_id = ? AND COALESCE(referencia_externa, '') = 'SEED_FINANZAS_MALIBU'`, empresaID).Scan(&existing)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	if existing > 0 {
		return 0, existing, 0, 0, nil
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	demo := []dbpkg.EmpresaFinanzasMovimiento{
		{
			EmpresaID:         empresaID,
			TipoMovimiento:    "ingreso",
			FechaMovimiento:   now,
			Categoria:         "ventas",
			Concepto:          "Ingreso por ventas mostrador",
			Descripcion:       "Ingreso de prueba para validacion del modulo financiero",
			MetodoPago:        "efectivo",
			Moneda:            "COP",
			Monto:             255000,
			Impuesto:          0,
			Total:             255000,
			TerceroNombre:     "Cliente mostrador",
			TipoComprobante:   "recibo_interno",
			NumeroComprobante: "",
			ReferenciaExterna: "SEED_FINANZAS_MALIBU",
			UsuarioCreador:    seedUser,
			Estado:            "activo",
		},
		{
			EmpresaID:         empresaID,
			TipoMovimiento:    "ingreso",
			FechaMovimiento:   now,
			Categoria:         "servicios",
			Concepto:          "Ingreso por servicio adicional",
			Descripcion:       "Ingreso de servicio para pruebas",
			MetodoPago:        "transferencia",
			Moneda:            "COP",
			Monto:             98000,
			Impuesto:          0,
			Total:             98000,
			TerceroNombre:     "Cliente corporativo",
			TipoComprobante:   "factura",
			NumeroComprobante: "",
			ReferenciaExterna: "SEED_FINANZAS_MALIBU",
			UsuarioCreador:    seedUser,
			Estado:            "activo",
		},
		{
			EmpresaID:         empresaID,
			TipoMovimiento:    "egreso",
			FechaMovimiento:   now,
			Categoria:         "compras",
			Concepto:          "Compra de abastecimiento",
			Descripcion:       "Egreso de prueba para inventario",
			MetodoPago:        "transferencia",
			Moneda:            "COP",
			Monto:             124500,
			Impuesto:          0,
			Total:             124500,
			TerceroNombre:     "Proveedor Local Medellin",
			TipoComprobante:   "factura",
			NumeroComprobante: "",
			ReferenciaExterna: "SEED_FINANZAS_MALIBU",
			UsuarioCreador:    seedUser,
			Estado:            "activo",
		},
		{
			EmpresaID:         empresaID,
			TipoMovimiento:    "egreso",
			FechaMovimiento:   now,
			Categoria:         "servicios",
			Concepto:          "Pago servicios publicos",
			Descripcion:       "Egreso de servicios para pruebas",
			MetodoPago:        "efectivo",
			Moneda:            "COP",
			Monto:             76000,
			Impuesto:          0,
			Total:             76000,
			TerceroNombre:     "Empresa de servicios",
			TipoComprobante:   "soporte_externo",
			NumeroComprobante: "",
			ReferenciaExterna: "SEED_FINANZAS_MALIBU",
			UsuarioCreador:    seedUser,
			Estado:            "activo",
		},
	}

	created := 0
	totalIngresos := 0.0
	totalEgresos := 0.0
	for _, mov := range demo {
		if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbConn, mov); err != nil {
			return created, 0, totalIngresos, totalEgresos, err
		}
		created++
		if strings.EqualFold(strings.TrimSpace(mov.TipoMovimiento), "ingreso") {
			totalIngresos += mov.Total
		} else {
			totalEgresos += mov.Total
		}
	}

	return created, 0, totalIngresos, totalEgresos, nil
}

func sortedCategoryNames() []string {
	namesMap := map[string]bool{}
	for _, p := range demoProducts {
		namesMap[p.Categoria] = true
	}
	names := make([]string, 0, len(namesMap))
	for k := range namesMap {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func main() {
	dbPath := os.Getenv("DB_EMPRESAS_PATH")
	if strings.TrimSpace(dbPath) == "" {
		dbPath = "db/empresas.db"
	}

	fmt.Println("Seeder Motel Malibu")
	fmt.Println("DB empresas:", dbPath)

	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("no se pudo abrir la db: %v", err)
	}
	defer dbConn.Close()
	dbConn.SetMaxOpenConns(1)

	if err := ensureSchemas(dbConn); err != nil {
		log.Fatalf("error asegurando esquemas: %v", err)
	}

	empresaID, empresaNombre, createdEmpresa, err := findOrCreateEmpresa(dbConn, "Motel Malibu")
	if err != nil {
		log.Fatalf("error buscando/creando empresa: %v", err)
	}
	if createdEmpresa {
		fmt.Printf("Empresa creada: %s (empresa_id=%d)\n", empresaNombre, empresaID)
	} else {
		fmt.Printf("Empresa encontrada: %s (empresa_id=%d)\n", empresaNombre, empresaID)
	}

	bodegaID, createdBodega, err := ensureBodegaPrincipal(dbConn, empresaID)
	if err != nil {
		log.Fatalf("error asegurando bodega: %v", err)
	}
	proveedorID, createdProveedor, err := ensureProveedorPrincipal(dbConn, empresaID)
	if err != nil {
		log.Fatalf("error asegurando proveedor: %v", err)
	}
	fmt.Printf("Bodega principal id=%d (creada=%v) | Proveedor principal id=%d (creado=%v)\n", bodegaID, createdBodega, proveedorID, createdProveedor)

	categoriaIDs, createdCategorias, err := ensureCategorias(dbConn, empresaID, sortedCategoryNames())
	if err != nil {
		log.Fatalf("error asegurando categorias: %v", err)
	}
	fmt.Printf("Categorias creadas: %d\n", createdCategorias)

	createdProd, skippedProd, productCatalog, err := seedProductos(dbConn, empresaID, bodegaID, proveedorID, categoriaIDs)
	if err != nil {
		log.Fatalf("error cargando productos: %v", err)
	}
	fmt.Printf("Productos procesados: total=%d | creados=%d | existentes=%d\n", len(demoProducts), createdProd, skippedProd)

	createdCli, skippedCli, clienteID, err := seedClientes(dbConn, empresaID)
	if err != nil {
		log.Fatalf("error cargando clientes: %v", err)
	}
	fmt.Printf("Clientes procesados: total=%d | creados=%d | existentes=%d | cliente_venta_id=%d\n", len(demoClients), createdCli, skippedCli, clienteID)

	createdUsers, skippedUsers, err := seedUsuariosEmpresa(dbConn, empresaID)
	if err != nil {
		log.Fatalf("error cargando usuarios de empresa: %v", err)
	}
	fmt.Printf("Usuarios empresa procesados: total=%d | creados=%d | existentes=%d\n", len(demoUsers), createdUsers, skippedUsers)

	createdFinanzas, skippedFinanzas, totalIngresosFin, totalEgresosFin, err := seedMovimientosFinancieros(dbConn, empresaID)
	if err != nil {
		log.Fatalf("error cargando movimientos financieros: %v", err)
	}
	if skippedFinanzas > 0 {
		fmt.Printf("Movimientos financieros demo: existentes=%d (sin crear nuevos)\n", skippedFinanzas)
	} else {
		fmt.Printf("Movimientos financieros demo creados=%d | ingresos=%.0f | egresos=%.0f | balance=%.0f\n", createdFinanzas, totalIngresosFin, totalEgresosFin, totalIngresosFin-totalEgresosFin)
	}

	sale, stockChecks, err := createVentaDemo(dbConn, empresaID, clienteID, productCatalog)
	if err != nil {
		log.Fatalf("error creando venta demo: %v", err)
	}
	fmt.Printf("Venta de prueba creada y cerrada: carrito_id=%d codigo=%s total_pagado=%.0f fecha=%s\n", sale.CarritoID, sale.Codigo, sale.Total, sale.PagadoEn)
	fmt.Println("Validacion de inventario (agregar carrito y venta):")
	for _, check := range stockChecks {
		fmt.Printf("  - %s | antes=%.2f | tras agregar=%.2f | tras pagar=%.2f | vendido=%.2f\n", check.SKU, check.BeforeAdd, check.AfterAdd, check.AfterPay, check.SoldQty)
	}

	cfg, err := ensurePrintConfig(dbConn, empresaID, empresaNombre)
	if err != nil {
		log.Fatalf("error guardando configuracion de impresion: %v", err)
	}
	fmt.Printf("Formato de impresion validado: formato=%s | imprimir_copia=%v | mostrar_logo=%v\n", cfg.FormatoImpresion, cfg.ImprimirCopiaFactura, cfg.MostrarLogo)

	fmt.Println("\n--- Vista previa POS ---")
	fmt.Println(renderPosPreview(empresaNombre, sale, cfg))
	fmt.Println("--- Vista previa CARTA ---")
	fmt.Println(renderCartaPreview(empresaNombre, sale, cfg))
	fmt.Println("Validacion de impresion: vista previa POS y CARTA generadas correctamente.")

	fmt.Println("Proceso finalizado correctamente.")
}
