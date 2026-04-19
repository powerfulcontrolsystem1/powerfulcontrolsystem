package db

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func openChatIATestDB(t *testing.T) *sql.DB {
	t.Helper()
	t.Setenv("DB_DIALECT", "sqlite")
	t.Setenv("PCS_DB_DIALECT", "sqlite")
	dbPath := filepath.Join(t.TempDir(), "chat_ia_test.db")
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

func TestEmpresaAIModeloPreferidoUpsertAndGet(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureEmpresaAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}

	modelID, err := GetEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com")
	if err != nil {
		t.Fatalf("get modelo preferido vacio: %v", err)
	}
	if modelID != "" {
		t.Fatalf("expected modelo preferido vacio, got %q", modelID)
	}

	if err := UpsertEmpresaAIModeloPreferido(dbConn, 10, "Admin@Example.com", "google:gemini-2.0-flash", "tester"); err != nil {
		t.Fatalf("upsert modelo preferido inicial: %v", err)
	}
	modelID, err = GetEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com")
	if err != nil {
		t.Fatalf("get modelo preferido inicial: %v", err)
	}
	if modelID != "google:gemini-2.0-flash" {
		t.Fatalf("expected model google:gemini-2.0-flash, got %q", modelID)
	}

	if err := UpsertEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com", "google:gemini-1.5-flash", "tester"); err != nil {
		t.Fatalf("upsert modelo preferido update: %v", err)
	}
	modelID, err = GetEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com")
	if err != nil {
		t.Fatalf("get modelo preferido update: %v", err)
	}
	if modelID != "google:gemini-1.5-flash" {
		t.Fatalf("expected model google:gemini-1.5-flash, got %q", modelID)
	}

	var provider string
	err = dbConn.QueryRow(`SELECT COALESCE(provider, '') FROM empresa_ai_modelo_preferido WHERE empresa_id = ? AND admin_email = ? LIMIT 1`, 10, "admin@example.com").Scan(&provider)
	if err != nil {
		t.Fatalf("query provider preferido: %v", err)
	}
	if provider != "google" {
		t.Fatalf("expected provider google, got %q", provider)
	}

	if err := UpsertEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com", "google:gemini-2.0-flash", "tester"); err != nil {
		t.Fatalf("upsert modelo preferido gemini: %v", err)
	}
	modelID, err = GetEmpresaAIModeloPreferido(dbConn, 10, "admin@example.com")
	if err != nil {
		t.Fatalf("get modelo preferido gemini: %v", err)
	}
	if modelID != "google:gemini-2.0-flash" {
		t.Fatalf("expected model google:gemini-2.0-flash, got %q", modelID)
	}
}

func TestRegisterEmpresaAIConsultaAcumulaUsoDiario(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureEmpresaAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}

	_, err := RegisterEmpresaAIConsulta(dbConn, EmpresaAIConsulta{
		EmpresaID:        6,
		Provider:         "google",
		ModelID:          "google:gemini-2.0-flash",
		Pregunta:         "primera pregunta",
		Respuesta:        "primera respuesta",
		PromptTokens:     8,
		CompletionTokens: 12,
		TotalTokens:      20,
		FechaConsulta:    "2026-04-03 10:00:00",
		PlanActual:       "free",
		UsuarioCreador:   "admin@example.com",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("register consulta 1: %v", err)
	}

	_, err = RegisterEmpresaAIConsulta(dbConn, EmpresaAIConsulta{
		EmpresaID:        6,
		Provider:         "google",
		ModelID:          "google:gemini-2.0-flash",
		Pregunta:         "segunda pregunta",
		Respuesta:        "segunda respuesta",
		PromptTokens:     5,
		CompletionTokens: 7,
		TotalTokens:      12,
		FechaConsulta:    "2026-04-03 11:00:00",
		PlanActual:       "free",
		UsuarioCreador:   "admin@example.com",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("register consulta 2: %v", err)
	}

	uso, err := GetEmpresaAIUsoDiario(dbConn, 6, "google", "google:gemini-2.0-flash", "2026-04-03")
	if err != nil {
		t.Fatalf("get uso diario: %v", err)
	}
	if uso.Consultas != 2 {
		t.Fatalf("expected 2 consultas, got %d", uso.Consultas)
	}
	if uso.TokensTotal != 32 {
		t.Fatalf("expected 32 tokens acumulados, got %d", uso.TokensTotal)
	}

	rows, err := ListEmpresaAIConsultasRecientes(dbConn, 6, 10)
	if err != nil {
		t.Fatalf("list consultas recientes: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 consultas recientes, got %d", len(rows))
	}
}

func TestSuperAIModeloPreferidoUpsertAndGet(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureSuperAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure super chat ia schema: %v", err)
	}

	modelID, err := GetSuperAIModeloPreferido(dbConn, "super@example.com")
	if err != nil {
		t.Fatalf("get super modelo preferido vacio: %v", err)
	}
	if modelID != "" {
		t.Fatalf("expected modelo preferido vacio, got %q", modelID)
	}

	if err := UpsertSuperAIModeloPreferido(dbConn, "Super@Example.com", "google:gemini-2.0-flash", "tester"); err != nil {
		t.Fatalf("upsert super modelo preferido inicial: %v", err)
	}
	modelID, err = GetSuperAIModeloPreferido(dbConn, "super@example.com")
	if err != nil {
		t.Fatalf("get super modelo preferido inicial: %v", err)
	}
	if modelID != "google:gemini-2.0-flash" {
		t.Fatalf("expected google:gemini-2.0-flash, got %q", modelID)
	}

	if err := UpsertSuperAIModeloPreferido(dbConn, "super@example.com", "google:gemini-1.5-flash", "tester"); err != nil {
		t.Fatalf("upsert super modelo preferido update: %v", err)
	}
	modelID, err = GetSuperAIModeloPreferido(dbConn, "super@example.com")
	if err != nil {
		t.Fatalf("get super modelo preferido update: %v", err)
	}
	if modelID != "google:gemini-1.5-flash" {
		t.Fatalf("expected google:gemini-1.5-flash, got %q", modelID)
	}

	var provider string
	err = dbConn.QueryRow(`SELECT COALESCE(provider, '') FROM super_ai_modelo_preferido WHERE admin_email = ? LIMIT 1`, "super@example.com").Scan(&provider)
	if err != nil {
		t.Fatalf("query provider super preferido: %v", err)
	}
	if provider != "google" {
		t.Fatalf("expected provider google, got %q", provider)
	}
}

func TestRegisterSuperAIConsultaAcumulaUsoDiario(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureSuperAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure super chat ia schema: %v", err)
	}

	_, err := RegisterSuperAIConsulta(dbConn, SuperAIConsulta{
		AdminEmail:       "super@example.com",
		Provider:         "google",
		ModelID:          "google:gemini-2.0-flash",
		Pregunta:         "primera pregunta global",
		Respuesta:        "primera respuesta global",
		PromptTokens:     9,
		CompletionTokens: 11,
		TotalTokens:      20,
		FechaConsulta:    "2026-04-18 10:00:00",
		PlanActual:       "free",
		UsuarioCreador:   "super@example.com",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("register super consulta 1: %v", err)
	}

	_, err = RegisterSuperAIConsulta(dbConn, SuperAIConsulta{
		AdminEmail:       "super@example.com",
		Provider:         "google",
		ModelID:          "google:gemini-2.0-flash",
		Pregunta:         "segunda pregunta global",
		Respuesta:        "segunda respuesta global",
		PromptTokens:     4,
		CompletionTokens: 6,
		TotalTokens:      10,
		FechaConsulta:    "2026-04-18 11:00:00",
		PlanActual:       "free",
		UsuarioCreador:   "super@example.com",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("register super consulta 2: %v", err)
	}

	uso, err := GetSuperAIUsoDiario(dbConn, "super@example.com", "google", "google:gemini-2.0-flash", "2026-04-18")
	if err != nil {
		t.Fatalf("get super uso diario: %v", err)
	}
	if uso.Consultas != 2 {
		t.Fatalf("expected 2 consultas, got %d", uso.Consultas)
	}
	if uso.TokensTotal != 30 {
		t.Fatalf("expected 30 tokens acumulados, got %d", uso.TokensTotal)
	}

	rows, err := ListSuperAIConsultasRecientes(dbConn, "super@example.com", 10)
	if err != nil {
		t.Fatalf("list super consultas recientes: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 consultas recientes, got %d", len(rows))
	}
	if rows[0].AdminEmail != "super@example.com" {
		t.Fatalf("expected admin_email super@example.com, got %q", rows[0].AdminEmail)
	}
}

func TestGetSuperAIModeloPreferidoRepairsMissingSchema(t *testing.T) {
	dbConn := openChatIATestDB(t)

	modelID, err := GetSuperAIModeloPreferido(dbConn, "super@example.com")
	if err != nil {
		t.Fatalf("get super modelo preferido con schema faltante: %v", err)
	}
	if modelID != "" {
		t.Fatalf("expected modelo preferido vacio con schema inicial, got %q", modelID)
	}

	if err := UpsertSuperAIModeloPreferido(dbConn, "super@example.com", "google:gemini-2.0-flash", "tester"); err != nil {
		t.Fatalf("upsert super modelo preferido tras reparacion: %v", err)
	}

	modelID, err = GetSuperAIModeloPreferido(dbConn, "super@example.com")
	if err != nil {
		t.Fatalf("get super modelo preferido tras reparacion: %v", err)
	}
	if modelID != "google:gemini-2.0-flash" {
		t.Fatalf("expected google:gemini-2.0-flash after repair, got %q", modelID)
	}
}

func TestGetSuperAIUsoDiarioRepairsMissingSchema(t *testing.T) {
	dbConn := openChatIATestDB(t)

	uso, err := GetSuperAIUsoDiario(dbConn, "super@example.com", "google", "google:gemini-2.0-flash", "2026-04-18")
	if err != nil {
		t.Fatalf("get super uso diario con schema faltante: %v", err)
	}
	if uso.Consultas != 0 {
		t.Fatalf("expected 0 consultas on repaired empty schema, got %d", uso.Consultas)
	}
	if uso.PlanActual != "free" {
		t.Fatalf("expected free plan on repaired empty schema, got %q", uso.PlanActual)
	}
}

func TestRegisterSuperAIConsultaRepairsMissingSchema(t *testing.T) {
	dbConn := openChatIATestDB(t)

	_, err := RegisterSuperAIConsulta(dbConn, SuperAIConsulta{
		AdminEmail:       "super@example.com",
		Provider:         "google",
		ModelID:          "google:gemini-2.0-flash",
		Pregunta:         "puedes resumir el contexto global disponible",
		Respuesta:        "si, solo contexto consolidado seguro",
		PromptTokens:     3,
		CompletionTokens: 5,
		TotalTokens:      8,
		FechaConsulta:    "2026-04-18 12:00:00",
		PlanActual:       "free",
		UsuarioCreador:   "super@example.com",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("register super consulta con schema faltante: %v", err)
	}

	uso, err := GetSuperAIUsoDiario(dbConn, "super@example.com", "google", "google:gemini-2.0-flash", "2026-04-18")
	if err != nil {
		t.Fatalf("get super uso diario tras reparacion: %v", err)
	}
	if uso.Consultas != 1 {
		t.Fatalf("expected 1 consulta after repair insert, got %d", uso.Consultas)
	}
	if uso.TokensTotal != 8 {
		t.Fatalf("expected 8 tokens after repair insert, got %d", uso.TokensTotal)
	}
}

func TestBuildEmpresaAIContextoIncluyeResumenOperativo(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureEmpresaAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE empresas (id INTEGER PRIMARY KEY AUTOINCREMENT, nombre TEXT, nit TEXT, estado TEXT, usuario_creador TEXT)`); err != nil {
		t.Fatalf("create empresas: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE clientes (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, numero_documento TEXT, nombre_razon_social TEXT, estado TEXT)`); err != nil {
		t.Fatalf("create clientes: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE productos (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, nombre TEXT, stock_minimo REAL DEFAULT 0, estado TEXT)`); err != nil {
		t.Fatalf("create productos: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE inventario_existencias (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, producto_id INTEGER, cantidad REAL DEFAULT 0)`); err != nil {
		t.Fatalf("create inventario_existencias: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE carritos_compras (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, cliente_id INTEGER, codigo TEXT, estado_carrito TEXT, total REAL DEFAULT 0, metodo_pago TEXT, pagado_en TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT)`); err != nil {
		t.Fatalf("create carritos_compras: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE carrito_compra_items (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, carrito_id INTEGER, descripcion TEXT, codigo_item TEXT, cantidad REAL DEFAULT 0, total_linea REAL DEFAULT 0)`); err != nil {
		t.Fatalf("create carrito_compra_items: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `CREATE TABLE empresa_finanzas_movimientos (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, tipo_movimiento TEXT, concepto TEXT, descripcion TEXT, numero_comprobante TEXT, monto REAL DEFAULT 0, total REAL DEFAULT 0, total_neto REAL DEFAULT 0, fecha_movimiento TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT, estado TEXT)`); err != nil {
		t.Fatalf("create empresa_finanzas_movimientos: %v", err)
	}

	if _, err := execSQLCompat(dbConn, `INSERT INTO empresas (id, nombre, nit, estado, usuario_creador) VALUES (?, ?, ?, ?, ?)`, 1, "Empresa Demo", "900100", "activo", "admin@example.com"); err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO clientes (id, empresa_id, numero_documento, nombre_razon_social, estado) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)`,
		1, 1, "1001", "Cliente Uno", "activo",
		2, 1, "1002", "Cliente Dos", "activo"); err != nil {
		t.Fatalf("insert clientes: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO productos (id, empresa_id, nombre, stock_minimo, estado) VALUES (?, ?, ?, ?, ?), (?, ?, ?, ?, ?)`,
		10, 1, "Cafe", 5, "activo",
		11, 1, "Pan", 3, "activo"); err != nil {
		t.Fatalf("insert productos: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO inventario_existencias (empresa_id, producto_id, cantidad) VALUES (?, ?, ?), (?, ?, ?)`,
		1, 10, 2,
		1, 11, 9); err != nil {
		t.Fatalf("insert inventario_existencias: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO carritos_compras (id, empresa_id, cliente_id, codigo, estado_carrito, total, metodo_pago, pagado_en, fecha_creacion, fecha_actualizacion) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		100, 1, 1, "VTA-001", "cerrado", 50000, "efectivo", "2026-04-18 10:00:00", "2026-04-18 09:50:00", "2026-04-18 10:00:00",
		101, 1, 2, "VTA-002", "pagado", 30000, "tarjeta", "2026-04-18 11:00:00", "2026-04-18 10:40:00", "2026-04-18 11:00:00"); err != nil {
		t.Fatalf("insert carritos: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO carrito_compra_items (empresa_id, carrito_id, descripcion, codigo_item, cantidad, total_linea) VALUES (?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?)`,
		1, 100, "Cafe", "CAF", 4, 40000,
		1, 101, "Pan", "PAN", 6, 30000); err != nil {
		t.Fatalf("insert items: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO empresa_finanzas_movimientos (empresa_id, tipo_movimiento, concepto, total, total_neto, fecha_movimiento, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		1, "ingreso", "Venta mostrador", 80000, 78000, "2026-04-18 11:30:00", "2026-04-18 11:30:00", "2026-04-18 11:30:00", "activo",
		1, "egreso", "Compra insumos", 20000, 20000, "2026-04-18 12:00:00", "2026-04-18 12:00:00", "2026-04-18 12:00:00", "activo"); err != nil {
		t.Fatalf("insert finanzas: %v", err)
	}

	contexto, err := BuildEmpresaAIContexto(dbConn, 1)
	if err != nil {
		t.Fatalf("build contexto empresa: %v", err)
	}

	for _, expected := range []string{
		"clientes_activos: 2",
		"TOP_PRODUCTOS_VENTA",
		"Cafe | cantidad=4.00 | total=40000.00",
		"TOP_CLIENTES",
		"Cliente Uno | compras=1 | total=50000.00",
		"ALERTAS_INVENTARIO",
		"Cafe | stock=2.00 | stock_minimo=5.00",
		"MOVIMIENTOS_FINANCIEROS_RECIENTES",
		"ingreso | concepto=Venta mostrador | total=78000.00",
	} {
		if !strings.Contains(contexto, expected) {
			t.Fatalf("expected contexto to contain %q, got:\n%s", expected, contexto)
		}
	}
}

func TestBuildEmpresaAIContextoForQuestionIncluyeConsultaSeguraClienteYProducto(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureEmpresaAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	for _, stmt := range []string{
		`CREATE TABLE empresas (id INTEGER PRIMARY KEY AUTOINCREMENT, nombre TEXT, nit TEXT, estado TEXT, usuario_creador TEXT)`,
		`CREATE TABLE clientes (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, numero_documento TEXT, nombre_razon_social TEXT, estado TEXT)`,
		`CREATE TABLE productos (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, nombre TEXT, sku TEXT, codigo_barras TEXT, stock_minimo REAL DEFAULT 0, estado TEXT)`,
		`CREATE TABLE inventario_existencias (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, producto_id INTEGER, cantidad REAL DEFAULT 0)`,
		`CREATE TABLE carritos_compras (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, cliente_id INTEGER, codigo TEXT, estado_carrito TEXT, total REAL DEFAULT 0, metodo_pago TEXT, pagado_en TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT)`,
		`CREATE TABLE carrito_compra_items (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, carrito_id INTEGER, descripcion TEXT, codigo_item TEXT, cantidad REAL DEFAULT 0, total_linea REAL DEFAULT 0)`,
	} {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			t.Fatalf("exec schema %q: %v", stmt, err)
		}
	}

	if _, err := execSQLCompat(dbConn, `INSERT INTO empresas (id, nombre, nit, estado, usuario_creador) VALUES (?, ?, ?, ?, ?)`, 1, "Empresa Demo", "900100", "activo", "admin@example.com"); err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO clientes (id, empresa_id, numero_documento, nombre_razon_social, estado) VALUES (?, ?, ?, ?, ?)`, 1, 1, "1001", "Cliente Uno", "activo"); err != nil {
		t.Fatalf("insert cliente: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO productos (id, empresa_id, nombre, sku, codigo_barras, stock_minimo, estado) VALUES (?, ?, ?, ?, ?, ?, ?)`, 10, 1, "Cafe Premium", "CAF-01", "7701", 4, "activo"); err != nil {
		t.Fatalf("insert producto: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO inventario_existencias (empresa_id, producto_id, cantidad) VALUES (?, ?, ?)`, 1, 10, 3); err != nil {
		t.Fatalf("insert existencias: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO carritos_compras (id, empresa_id, cliente_id, codigo, estado_carrito, total, metodo_pago, pagado_en, fecha_creacion, fecha_actualizacion) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, 100, 1, 1, "VTA-001", "cerrado", 42000, "efectivo", "2026-04-18 10:00:00", "2026-04-18 09:50:00", "2026-04-18 10:00:00"); err != nil {
		t.Fatalf("insert carrito: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO carrito_compra_items (empresa_id, carrito_id, descripcion, codigo_item, cantidad, total_linea) VALUES (?, ?, ?, ?, ?, ?)`, 1, 100, "Cafe Premium", "CAF-01", 2, 42000); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	contextoCliente, err := BuildEmpresaAIContextoForQuestion(dbConn, 1, `¿Qué ventas tiene el cliente "Cliente Uno"?`)
	if err != nil {
		t.Fatalf("build contexto cliente: %v", err)
	}
	for _, expected := range []string{
		"CONSULTAS_SEGURAS_RESUELTAS",
		"CONSULTA_SEGURA_CLIENTES",
		"Cliente Uno | ventas=1 | total=42000.00",
	} {
		if !strings.Contains(contextoCliente, expected) {
			t.Fatalf("expected contexto cliente to contain %q, got:\n%s", expected, contextoCliente)
		}
	}

	contextoProducto, err := BuildEmpresaAIContextoForQuestion(dbConn, 1, `¿Cómo va el producto "Cafe Premium"?`)
	if err != nil {
		t.Fatalf("build contexto producto: %v", err)
	}
	for _, expected := range []string{
		"CONSULTA_SEGURA_PRODUCTOS",
		"Cafe Premium | sku=CAF-01 | stock=3.00 | stock_minimo=4.00 | vendido=2.00 | total_vendido=42000.00",
	} {
		if !strings.Contains(contextoProducto, expected) {
			t.Fatalf("expected contexto producto to contain %q, got:\n%s", expected, contextoProducto)
		}
	}
}

func TestBuildEmpresaAIContextoForQuestionIncluyeProductosSinRotacion(t *testing.T) {
	dbConn := openChatIATestDB(t)
	if err := EnsureEmpresaAIChatSchema(dbConn); err != nil {
		t.Fatalf("ensure chat ia schema: %v", err)
	}
	for _, stmt := range []string{
		`CREATE TABLE empresas (id INTEGER PRIMARY KEY AUTOINCREMENT, nombre TEXT, nit TEXT, estado TEXT, usuario_creador TEXT)`,
		`CREATE TABLE productos (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, nombre TEXT, sku TEXT, stock_minimo REAL DEFAULT 0, estado TEXT)`,
		`CREATE TABLE inventario_existencias (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, producto_id INTEGER, cantidad REAL DEFAULT 0)`,
		`CREATE TABLE carritos_compras (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, cliente_id INTEGER, codigo TEXT, estado_carrito TEXT, total REAL DEFAULT 0, metodo_pago TEXT, pagado_en TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT)`,
		`CREATE TABLE carrito_compra_items (id INTEGER PRIMARY KEY AUTOINCREMENT, empresa_id INTEGER, carrito_id INTEGER, descripcion TEXT, codigo_item TEXT, cantidad REAL DEFAULT 0, total_linea REAL DEFAULT 0)`,
	} {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			t.Fatalf("exec schema %q: %v", stmt, err)
		}
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO empresas (id, nombre, nit, estado, usuario_creador) VALUES (?, ?, ?, ?, ?)`, 1, "Empresa Demo", "900100", "activo", "admin@example.com"); err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO productos (id, empresa_id, nombre, sku, stock_minimo, estado) VALUES (?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?)`,
		10, 1, "Cafe", "CAF-01", 4, "activo",
		11, 1, "Te", "TE-01", 2, "activo"); err != nil {
		t.Fatalf("insert productos: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO inventario_existencias (empresa_id, producto_id, cantidad) VALUES (?, ?, ?), (?, ?, ?)`,
		1, 10, 3,
		1, 11, 7); err != nil {
		t.Fatalf("insert existencias: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO carritos_compras (id, empresa_id, cliente_id, codigo, estado_carrito, total, metodo_pago, pagado_en, fecha_creacion, fecha_actualizacion) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		100, 1, 1, "VTA-001", "cerrado", 20000, "efectivo", "2026-04-18 10:00:00", "2026-04-18 09:50:00", "2026-04-18 10:00:00"); err != nil {
		t.Fatalf("insert carrito: %v", err)
	}
	if _, err := execSQLCompat(dbConn, `INSERT INTO carrito_compra_items (empresa_id, carrito_id, descripcion, codigo_item, cantidad, total_linea) VALUES (?, ?, ?, ?, ?, ?)`,
		1, 100, "Cafe", "CAF-01", 1, 20000); err != nil {
		t.Fatalf("insert item: %v", err)
	}

	contexto, err := BuildEmpresaAIContextoForQuestion(dbConn, 1, `¿Qué productos están sin rotación?`)
	if err != nil {
		t.Fatalf("build contexto sin rotacion: %v", err)
	}
	for _, expected := range []string{
		"CONSULTA_SEGURA_PRODUCTOS_SIN_ROTACION_30D",
		"Te | stock=7.00 | stock_minimo=2.00 | sin_ventas_30d=true",
	} {
		if !strings.Contains(contexto, expected) {
			t.Fatalf("expected contexto sin rotacion to contain %q, got:\n%s", expected, contexto)
		}
	}
}
