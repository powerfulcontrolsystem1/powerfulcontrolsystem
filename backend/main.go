package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/handlers"
	"github.com/you/pos-backend/metrics"
	"github.com/you/pos-backend/utils"
)

var (
	clientID       = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret   = os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL    = os.Getenv("GOOGLE_REDIRECT_URL") // e.g. http://localhost:8080/auth/google/callback
	dbEmpresasPath = os.Getenv("DB_EMPRESAS_PATH")
	dbSuperPath    = os.Getenv("DB_SUPERADMIN_PATH")
	dbEmpresas     *sql.DB
	dbSuper        *sql.DB
)

func getenvIntRange(key string, defaultVal, minVal, maxVal int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("warning: %s invalido (%q), se usa valor por defecto %d", key, raw, defaultVal)
		return defaultVal
	}
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func resolveAsientosWorkerPolicy() (time.Duration, int, int) {
	intervalMinutes := getenvIntRange("ASIENTOS_WORKER_INTERVAL_MINUTES", 15, 1, 1440)
	batchSize := getenvIntRange("ASIENTOS_WORKER_BATCH_SIZE", 100, 1, 500)
	maxRetries := getenvIntRange("ASIENTOS_WORKER_MAX_RETRIES", 5, 1, 50)
	return time.Duration(intervalMinutes) * time.Minute, batchSize, maxRetries
}

func readConfigValueFromDB(dbConn *sql.DB, keys []string) (string, string, error) {
	for _, key := range keys {
		val, enc, err := dbpkg.GetConfigValue(dbConn, key)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return "", "", err
		}

		clean := strings.TrimSpace(val)
		if clean == "" {
			continue
		}

		if enc {
			dec, derr := utils.DecryptString(clean)
			if derr != nil {
				log.Printf("warning: no se pudo descifrar la configuración %s: %v", key, derr)
				continue
			}
			clean = strings.TrimSpace(dec)
			if clean == "" {
				continue
			}
		}

		return clean, key, nil
	}

	return "", "", nil
}

func loadGoogleOAuthFromDB(dbConn *sql.DB) {
	clientIDKeys := []string{
		"google.client_id",
		"oauth.google.client_id",
		"auth.google.client_id",
		"google_oauth.client_id",
		"GOOGLE_CLIENT_ID",
	}
	clientSecretKeys := []string{
		"google.client_secret",
		"oauth.google.client_secret",
		"auth.google.client_secret",
		"google_oauth.client_secret",
		"GOOGLE_CLIENT_SECRET",
	}
	redirectURLKeys := []string{
		"google.redirect_url",
		"oauth.google.redirect_url",
		"auth.google.redirect_url",
		"google_oauth.redirect_url",
		"GOOGLE_REDIRECT_URL",
	}

	dbClientID, clientIDKey, err := readConfigValueFromDB(dbConn, clientIDKeys)
	if err != nil {
		log.Printf("warning: no se pudo leer GOOGLE_CLIENT_ID desde DB: %v", err)
	}
	dbClientSecret, clientSecretKey, err := readConfigValueFromDB(dbConn, clientSecretKeys)
	if err != nil {
		log.Printf("warning: no se pudo leer GOOGLE_CLIENT_SECRET desde DB: %v", err)
	}
	dbRedirectURL, redirectURLKey, err := readConfigValueFromDB(dbConn, redirectURLKeys)
	if err != nil {
		log.Printf("warning: no se pudo leer GOOGLE_REDIRECT_URL desde DB: %v", err)
	}

	// Si la DB tiene client_id + client_secret, tomarlos como fuente de verdad.
	if dbClientID != "" && dbClientSecret != "" {
		clientID = dbClientID
		clientSecret = dbClientSecret
		if dbRedirectURL != "" {
			redirectURL = dbRedirectURL
		}
		log.Printf("INFO: OAuth Google cargado desde DB (%s, %s)", clientIDKey, clientSecretKey)
		if dbRedirectURL != "" {
			log.Printf("INFO: redirect OAuth cargado desde DB (%s)", redirectURLKey)
		}
		return
	}

	if clientID == "" && dbClientID != "" {
		clientID = dbClientID
		log.Printf("INFO: GOOGLE_CLIENT_ID completado desde DB (%s)", clientIDKey)
	}
	if clientSecret == "" && dbClientSecret != "" {
		clientSecret = dbClientSecret
		log.Printf("INFO: GOOGLE_CLIENT_SECRET completado desde DB (%s)", clientSecretKey)
	}
	if redirectURL == "" && dbRedirectURL != "" {
		redirectURL = dbRedirectURL
		log.Printf("INFO: GOOGLE_REDIRECT_URL completado desde DB (%s)", redirectURLKey)
	}
}

func main() {
	if dbEmpresasPath == "" {
		dbEmpresasPath = "empresas.db"
	}
	if dbSuperPath == "" {
		dbSuperPath = "superadministrador.db"
	}
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/google/callback"
		log.Println("INFO: usando redirectURL por defecto:", redirectURL)
	}

	var err error
	// Abrir base de datos para empresas
	dbEmpresas, err = sql.Open("sqlite", dbEmpresasPath)
	if err != nil {
		log.Fatalf("failed to open empresas sqlite db: %v", err)
	}
	// Abrir base de datos para superadministrador
	dbSuper, err = sql.Open("sqlite", dbSuperPath)
	if err != nil {
		log.Fatalf("failed to open superadministrador sqlite db: %v", err)
	}

	if err := dbpkg.EnsureSchemaMigrationsTable(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure schema_migrations in empresas db: %v", err)
	}
	if err := dbpkg.EnsureSchemaMigrationsTable(dbSuper); err != nil {
		log.Fatalf("failed to ensure schema_migrations in super db: %v", err)
	}

	// Crear tablas en dbEmpresas
	createUsers := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT DEFAULT 'administrador',
		empresa_id INTEGER,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbEmpresas.Exec(createUsers); err != nil {
		log.Fatalf("failed to create users table in empresas db: %v", err)
	}

	// Asegurar esquema mínimo de users para gestión de usuarios por empresa y confirmación por correo.
	ensureUsersSchema := func(db *sql.DB) {
		rows, err := db.Query("PRAGMA table_info(users);")
		if err != nil {
			log.Printf("warning: unable to inspect users schema: %v", err)
			return
		}
		defer rows.Close()
		existing := map[string]bool{}
		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				log.Printf("warning: scan pragma table_info users error: %v", err)
				return
			}
			existing[name] = true
		}

		addIfMissing := func(colDef string, name string) {
			if !existing[name] {
				q := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s;", colDef)
				if _, err := db.Exec(q); err != nil {
					log.Printf("failed to add column %s to users: %v", name, err)
				} else {
					log.Printf("added missing column %s to users", name)
				}
			}
		}

		addIfMissing("documento_identidad TEXT", "documento_identidad")
		addIfMissing("rol_usuario_id INTEGER", "rol_usuario_id")
		addIfMissing("email_confirmado INTEGER DEFAULT 0", "email_confirmado")
		addIfMissing("email_confirm_token TEXT", "email_confirm_token")
		addIfMissing("email_confirm_expira TEXT", "email_confirm_expira")
		addIfMissing("email_confirmado_en TEXT", "email_confirmado_en")
		addIfMissing("password_hash TEXT", "password_hash")
		addIfMissing("password_salt TEXT", "password_salt")
		addIfMissing("password_set INTEGER DEFAULT 0", "password_set")
		addIfMissing("password_actualizada_en TEXT", "password_actualizada_en")
		addIfMissing("fecha_actualizacion TEXT", "fecha_actualizacion")
		addIfMissing("usuario_creador TEXT", "usuario_creador")
		addIfMissing("estado TEXT DEFAULT 'activo'", "estado")
		addIfMissing("observaciones TEXT", "observaciones")
	}
	ensureUsersSchema(dbEmpresas)

	createEmpresas := `CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre TEXT NOT NULL,
		nit TEXT,
		tipo_id INTEGER,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbEmpresas.Exec(createEmpresas); err != nil {
		log.Fatalf("failed to create empresas table in empresas db: %v", err)
	}

	// Asegurar esquema mínimo de la tabla empresas (migraciones simples)
	ensureEmpresasSchema := func(db *sql.DB) {
		rows, err := db.Query("PRAGMA table_info(empresas);")
		if err != nil {
			log.Printf("warning: unable to inspect empresas schema: %v", err)
			return
		}
		defer rows.Close()
		existing := map[string]bool{}
		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				log.Printf("warning: scan pragma table_info error: %v", err)
				return
			}
			existing[name] = true
		}

		addIfMissing := func(colDef string, name string) {
			if !existing[name] {
				q := fmt.Sprintf("ALTER TABLE empresas ADD COLUMN %s;", colDef)
				if _, err := db.Exec(q); err != nil {
					log.Printf("failed to add column %s to empresas: %v", name, err)
				} else {
					log.Printf("added missing column %s to empresas", name)
				}
			}
		}

		addIfMissing("tipo_id INTEGER", "tipo_id")
		addIfMissing("tipo_nombre TEXT", "tipo_nombre")
		addIfMissing("empresa_id INTEGER", "empresa_id")
		addIfMissing("fecha_actualizacion TEXT", "fecha_actualizacion")
		addIfMissing("usuario_creador TEXT", "usuario_creador")
		addIfMissing("estado TEXT DEFAULT 'activo'", "estado")
		addIfMissing("observaciones TEXT", "observaciones")

		if _, err := db.Exec("UPDATE empresas SET empresa_id = id WHERE empresa_id IS NULL OR empresa_id <= 0"); err != nil {
			log.Printf("warning: unable to backfill empresa_id in empresas table: %v", err)
		}
		if _, err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS ux_empresas_empresa_id ON empresas(empresa_id)"); err != nil {
			log.Printf("warning: unable to create ux_empresas_empresa_id: %v", err)
		}
	}
	ensureEmpresasSchema(dbEmpresas)
	if err := dbpkg.EnsureEmpresasScopeReferences(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure scope references in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure productos schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaClientesSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure clientes schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure carritos schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaConfiguracionAvanzadaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure empresa_configuracion_avanzada schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure facturacion_electronica schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaChatTareasSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure chat_tareas schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaUbicacionGPSSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure ubicacion_gps schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure finanzas schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure eventos contables schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure documentos transaccionales schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaAIChatSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure chat IA schema in empresas db: %v", err)
	}
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmpresas); err != nil {
		log.Fatalf("failed to ensure auditoria schema in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-01-001-baseline", "baseline schema snapshot: users, empresas, productos, clientes, carritos, configuracion_avanzada"); err != nil {
		log.Fatalf("failed to register schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-02-001-chat-tareas", "chat y tareas por empresa: conversaciones, participantes, mensajes, adjuntos y tareas"); err != nil {
		log.Fatalf("failed to register chat_tareas schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-01-002-empresa-scope-and-fe", "asegura referencia empresa_id en tablas base y agrega modulo de facturacion electronica por pais"); err != nil {
		log.Fatalf("failed to register empresas scope/fe schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-02-002-ubicacion-gps", "modulo de ubicacion gps por empresa: dispositivos y recorridos con tracking periodico"); err != nil {
		log.Fatalf("failed to register ubicacion_gps schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-03-003-finanzas", "modulo financiero por empresa: ingresos, egresos, comprobantes y configuracion"); err != nil {
		log.Fatalf("failed to register finanzas schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-03-004-finanzas-periodos-retenciones", "periodos contables, bloqueo por cierre, retenciones y reportes contables avanzados"); err != nil {
		log.Fatalf("failed to register finanzas periodos schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-03-005-chat-ia-empresa", "chat con inteligencia artificial por empresa, modelos externos y control de uso diario"); err != nil {
		log.Fatalf("failed to register chat ia schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-04-006-chat-ia-modelo-preferido", "persistencia de modelo preferido por empresa y cuenta Google autenticada"); err != nil {
		log.Fatalf("failed to register chat ia preferred model schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-04-007-eventos-contables", "contrato de eventos contables por modulo y trazabilidad de ventas"); err != nil {
		log.Fatalf("failed to register eventos contables schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-04-008-documentos-transaccionales", "persistencia canonica de documentos transaccionales de facturacion y compras"); err != nil {
		log.Fatalf("failed to register documentos transaccionales schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-04-009-cierres-caja", "flujo operativo de cierre de caja por sucursal y arqueo de efectivo"); err != nil {
		log.Fatalf("failed to register cierres caja schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-04-010-asientos-canonicos", "persistencia canonica de asientos por evento procesado con control de idempotencia y reintentos"); err != nil {
		log.Fatalf("failed to register asientos canonicos schema migration in empresas db: %v", err)
	}
	if err := dbpkg.RegisterSchemaMigration(dbEmpresas, "empresas", "2026-04-04-011-auditoria-empresa", "registro de auditoria por empresa para acciones criticas con consulta filtrable y politica de retencion"); err != nil {
		log.Fatalf("failed to register auditoria empresa schema migration in empresas db: %v", err)
	}
	// Crear tipos_de_empresas en la base de datos de superadministrador (ubicación centralizada)
	createTiposSuper := `CREATE TABLE IF NOT EXISTS tipos_de_empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT NOT NULL UNIQUE,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createTiposSuper); err != nil {
		log.Fatalf("failed to create tipos_de_empresas table in superadministrador db: %v", err)
	}

	createRolesDeUsuario := `CREATE TABLE IF NOT EXISTS roles_de_usuario (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tipo_empresa_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		descripcion TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createRolesDeUsuario); err != nil {
		log.Fatalf("failed to create roles_de_usuario table in super db: %v", err)
	}

	createTiposDeUsuario := `CREATE TABLE IF NOT EXISTS tipos_de_usuario (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tipo_empresa_id INTEGER NOT NULL,
		rol_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		descripcion TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createTiposDeUsuario); err != nil {
		log.Fatalf("failed to create tipos_de_usuario table in super db: %v", err)
	}

	// Crear tablas en dbSuper (superadministrador)
	createAdmins := `CREATE TABLE IF NOT EXISTS administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT DEFAULT 'administrador',
		photo TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createAdmins); err != nil {
		log.Fatalf("failed to create administradores table in super db: %v", err)
	}

	// Asegurar columna 'photo' en administradores para almacenar URL de avatar
	ensureAdminsSchema := func(db *sql.DB) {
		rows, err := db.Query("PRAGMA table_info(administradores);")
		if err != nil {
			log.Printf("warning: unable to inspect administradores schema: %v", err)
			return
		}
		defer rows.Close()
		existing := map[string]bool{}
		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				log.Printf("warning: scan pragma table_info error: %v", err)
				return
			}
			existing[name] = true
		}

		addIfMissing := func(colDef string, name string) {
			if !existing[name] {
				q := fmt.Sprintf("ALTER TABLE administradores ADD COLUMN %s;", colDef)
				if _, err := db.Exec(q); err != nil {
					log.Printf("failed to add column %s to administradores: %v", name, err)
				} else {
					log.Printf("added missing column %s to administradores", name)
				}
			}
		}

		addIfMissing("photo TEXT", "photo")
	}
	ensureAdminsSchema(dbSuper)

	createTiposLic := `CREATE TABLE IF NOT EXISTS tipos_de_licencia (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT NOT NULL UNIQUE,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createTiposLic); err != nil {
		log.Fatalf("failed to create tipos_de_licencia table in super db: %v", err)
	}

	// licencias, configuracion, sesiones (super)
	createLic := `CREATE TABLE IF NOT EXISTS licencias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		tipo_id INTEGER,
		nombre TEXT,
		descripcion TEXT,
		valor REAL DEFAULT 0,
		duracion_dias INTEGER DEFAULT 0,
		fecha_inicio TEXT,
		fecha_fin TEXT,
		activo INTEGER DEFAULT 1,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime'))
	);`
	if _, err := dbSuper.Exec(createLic); err != nil {
		log.Fatalf("failed to create licencias table in super db: %v", err)
	}

	// Asegurar esquema mínimo de la tabla licencias (migraciones simples)
	ensureLicenciasSchema := func(db *sql.DB) {
		// Obtener columnas actuales
		rows, err := db.Query("PRAGMA table_info(licencias);")
		if err != nil {
			log.Printf("warning: unable to inspect licencias schema: %v", err)
			return
		}
		defer rows.Close()
		existing := map[string]bool{}
		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				log.Printf("warning: scan pragma table_info error: %v", err)
				return
			}
			existing[name] = true
		}

		addIfMissing := func(colDef string, name string) {
			if !existing[name] {
				q := fmt.Sprintf("ALTER TABLE licencias ADD COLUMN %s;", colDef)
				if _, err := db.Exec(q); err != nil {
					log.Printf("failed to add column %s to licencias: %v", name, err)
				} else {
					log.Printf("added missing column %s to licencias", name)
				}
			}
		}

		addIfMissing("empresa_id INTEGER", "empresa_id")
		addIfMissing("tipo_id INTEGER", "tipo_id")
		addIfMissing("nombre TEXT", "nombre")
		addIfMissing("descripcion TEXT", "descripcion")
		addIfMissing("valor REAL DEFAULT 0", "valor")
		addIfMissing("duracion_dias INTEGER DEFAULT 0", "duracion_dias")
		addIfMissing("fecha_inicio TEXT", "fecha_inicio")
		addIfMissing("fecha_fin TEXT", "fecha_fin")
		addIfMissing("activo INTEGER DEFAULT 1", "activo")
		addIfMissing("fecha_creacion TEXT DEFAULT (datetime('now','localtime'))", "fecha_creacion")
	}
	ensureLicenciasSchema(dbSuper)

	// Tabla para registrar preferencias/pagos de Mercado Pago
	createPagos := `CREATE TABLE IF NOT EXISTS pagos_mercadopago (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		licencia_id INTEGER,
		empresa_id INTEGER,
		preference_id TEXT,
		payment_id TEXT,
		status TEXT,
		raw_payload TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createPagos); err != nil {
		log.Fatalf("failed to create pagos_mercadopago table in super db: %v", err)
	}

	// Asegurar columnas nuevas en pagos_mercadopago para compatibilidad con instalaciones previas.
	ensurePagosSchema := func(db *sql.DB) {
		rows, err := db.Query("PRAGMA table_info(pagos_mercadopago);")
		if err != nil {
			log.Printf("warning: unable to inspect pagos_mercadopago schema: %v", err)
			return
		}
		defer rows.Close()
		existing := map[string]bool{}
		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				log.Printf("warning: scan pragma table_info pagos_mercadopago error: %v", err)
				return
			}
			existing[name] = true
		}

		addIfMissing := func(colDef string, name string) {
			if !existing[name] {
				q := fmt.Sprintf("ALTER TABLE pagos_mercadopago ADD COLUMN %s;", colDef)
				if _, err := db.Exec(q); err != nil {
					log.Printf("failed to add column %s to pagos_mercadopago: %v", name, err)
				} else {
					log.Printf("added missing column %s to pagos_mercadopago", name)
				}
			}
		}

		addIfMissing("fecha_actualizacion TEXT", "fecha_actualizacion")
		addIfMissing("usuario_creador TEXT", "usuario_creador")
		addIfMissing("estado TEXT DEFAULT 'activo'", "estado")
		addIfMissing("observaciones TEXT", "observaciones")
	}
	ensurePagosSchema(dbSuper)

	// Tabla para registrar transacciones/pagos de Wompi (Nequi)
	createPagosWompi := `CREATE TABLE IF NOT EXISTS pagos_wompi (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		licencia_id INTEGER,
		empresa_id INTEGER,
		transaction_id TEXT,
		reference TEXT,
		status TEXT,
		raw_payload TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createPagosWompi); err != nil {
		log.Fatalf("failed to create pagos_wompi table in super db: %v", err)
	}

	ensurePagosWompiSchema := func(db *sql.DB) {
		rows, err := db.Query("PRAGMA table_info(pagos_wompi);")
		if err != nil {
			log.Printf("warning: unable to inspect pagos_wompi schema: %v", err)
			return
		}
		defer rows.Close()
		existing := map[string]bool{}
		for rows.Next() {
			var cid int
			var name string
			var ctype string
			var notnull int
			var dflt sql.NullString
			var pk int
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				log.Printf("warning: scan pragma table_info pagos_wompi error: %v", err)
				return
			}
			existing[name] = true
		}

		addIfMissing := func(colDef string, name string) {
			if !existing[name] {
				q := fmt.Sprintf("ALTER TABLE pagos_wompi ADD COLUMN %s;", colDef)
				if _, err := db.Exec(q); err != nil {
					log.Printf("failed to add column %s to pagos_wompi: %v", name, err)
				} else {
					log.Printf("added missing column %s to pagos_wompi", name)
				}
			}
		}

		addIfMissing("fecha_actualizacion TEXT", "fecha_actualizacion")
		addIfMissing("usuario_creador TEXT", "usuario_creador")
		addIfMissing("estado TEXT DEFAULT 'activo'", "estado")
		addIfMissing("observaciones TEXT", "observaciones")
	}
	ensurePagosWompiSchema(dbSuper)

	// Tabla para almacenar configuraciones/k-v (ej. credenciales cifradas)
	createConfiguraciones := `CREATE TABLE IF NOT EXISTS configuraciones (
		config_key TEXT PRIMARY KEY,
		value TEXT,
		encrypted INTEGER DEFAULT 0,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createConfiguraciones); err != nil {
		log.Fatalf("failed to create configuraciones table in super db: %v", err)
	}
	loadGoogleOAuthFromDB(dbSuper)
	if clientID == "" || clientSecret == "" {
		log.Println("Warning: GOOGLE_CLIENT_ID o GOOGLE_CLIENT_SECRET no configurados (entorno/DB)")
	}

	// Crear tabla de sesiones en la base superadministrador
	createSesiones := `CREATE TABLE IF NOT EXISTS sesiones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		admin_email TEXT,
		token TEXT,
		ip TEXT,
		user_agent TEXT,
		fecha_inicio TEXT DEFAULT (datetime('now','localtime')),
		fecha_fin TEXT,
		activo INTEGER DEFAULT 1,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime'))
	);`
	if _, err := dbSuper.Exec(createSesiones); err != nil {
		log.Fatalf("failed to create sesiones table in super db: %v", err)
	}

	if err := dbpkg.RegisterSchemaMigration(dbSuper, "superadministrador", "2026-04-01-001-baseline", "baseline schema snapshot: administradores, licencias, configuraciones, sesiones, pagos"); err != nil {
		log.Fatalf("failed to register schema migration in super db: %v", err)
	}

	// Inicializar tabla de métricas y arrancar collector periódico
	if err := dbpkg.InitMetricsTable(dbSuper); err != nil {
		log.Printf("warning: failed to init metrics table: %v", err)
	}
	metricsInterval := metrics.DefaultIntervalSeconds()
	stopMetrics := make(chan struct{})
	go metrics.StartCollector(dbSuper, metricsInterval, stopMetrics)

	stopAuditRetention := make(chan struct{})
	go dbpkg.StartEmpresaAuditoriaRetentionWorker(dbEmpresas, 12*time.Hour, stopAuditRetention)

	asientosInterval, asientosBatchSize, asientosMaxRetries := resolveAsientosWorkerPolicy()
	log.Printf("[asientos_worker] policy interval=%s batch=%d max_reintentos=%d", asientosInterval, asientosBatchSize, asientosMaxRetries)
	stopAsientosWorker := make(chan struct{})
	go dbpkg.StartEmpresaAsientosContablesWorker(dbEmpresas, asientosInterval, asientosBatchSize, asientosMaxRetries, stopAsientosWorker)

	http.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin(clientID, redirectURL))
	// Pasar la conexión de la base `empresas` al callback para persistir usuarios y empresas
	// Pasar tanto la conexión de empresas como la de superadministrador al callback
	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback(dbEmpresas, dbSuper, clientID, clientSecret, redirectURL))

	// Endpoints para administración y auditoría (listar administradores y sesiones)
	http.HandleFunc("/super/administradores", handlers.ListAdministradoresHandler(dbSuper))
	http.HandleFunc("/super/sesiones", handlers.ListSesionesHandler(dbSuper))

	// Endpoints CRUD para tipos de empresas
	http.HandleFunc("/super/api/tipos_empresas", handlers.TiposEmpresasHandler(dbSuper))
	http.HandleFunc("/super/api/roles_de_usuario", handlers.RolesDeUsuarioHandler(dbSuper))
	http.HandleFunc("/super/api/tipos_de_usuario", handlers.TiposDeUsuarioHandler(dbSuper))
	// Endpoint CRUD para empresas (guardadas en empresas.db)
	http.HandleFunc("/super/api/empresas", handlers.EmpresasHandler(dbEmpresas))
	// Módulo de productos por empresa (empresas.db)
	http.HandleFunc("/api/empresa/bodegas", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaBodegasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/categorias_productos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaCategoriasProductosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/productos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaProductosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/productos/imagen", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaProductoImagenUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/existencias", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioExistenciasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/alertas", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioAlertasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/resumen", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioResumenHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/tendencia", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioTendenciaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/balance_bodegas", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioBalanceBodegasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/movimientos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioMovimientosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/transferir", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioTransferHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/ajustar", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioAjusteHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/inventario/cambiar_producto", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaInventarioCambioProductoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/productos/precios_historial", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaProductoPrecioHistorialHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/proveedores", handlers.WithEmpresaComprasPermissions(dbEmpresas, dbSuper, handlers.EmpresaProveedoresHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/servicios", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaServiciosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/usuarios/login", handlers.EmpresaUsuarioLoginHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/empresa/usuarios/establecer_password", handlers.EmpresaUsuarioSetPasswordHandler(dbEmpresas, dbSuper))
	http.HandleFunc("/api/empresa/usuarios", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaUsuariosHandler(dbEmpresas, dbSuper)))
	http.HandleFunc("/api/empresa/clientes", handlers.WithEmpresaClientesPermissions(dbEmpresas, dbSuper, handlers.EmpresaClientesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/carritos_compra", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCarritosCompraHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/carritos_compra/items", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaCarritoItemsHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/configuracion_avanzada", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaConfiguracionAvanzadaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica/pais_detectado", handlers.WithEmpresaFacturacionPermissions(dbEmpresas, dbSuper, handlers.EmpresaFacturacionElectronicaPaisDetectadoHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/facturacion_electronica/paises_disponibles", handlers.EmpresaFacturacionElectronicaPaisesDisponiblesHandler())
	http.HandleFunc("/api/empresa/chat_tareas/conversaciones", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasConversacionesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/participantes", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasParticipantesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/mensajes", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasMensajesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/mensajes/adjunto", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasAdjuntoUploadHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/chat_tareas/tareas", handlers.WithEmpresaVentasPermissions(dbEmpresas, dbSuper, handlers.EmpresaChatTareasTareasHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/ubicacion_gps/dispositivos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaUbicacionGPSDispositivosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/ubicacion_gps/recorridos", handlers.WithEmpresaInventarioPermissions(dbEmpresas, dbSuper, handlers.EmpresaUbicacionGPSRecorridosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/movimientos", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasMovimientosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/configuracion", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasConfiguracionHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/periodos", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasPeriodosHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/asientos_contables", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasAsientosContablesHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/finanzas/cierres_caja", handlers.WithEmpresaFinanzasPermissions(dbEmpresas, dbSuper, handlers.EmpresaFinanzasCierresCajaHandler(dbEmpresas)))
	http.HandleFunc("/api/empresa/auditoria/eventos", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaAuditoriaEventosHandler(dbEmpresas)))
	handlers.RegisterEmpresaChatIARoutes(dbEmpresas, dbSuper)
	http.HandleFunc("/api/empresa/roles_de_usuario", handlers.WithEmpresaSeguridadPermissions(dbEmpresas, dbSuper, handlers.EmpresaRolesDeUsuarioHandler(dbEmpresas, dbSuper)))
	// Endpoint para obtener admin actual desde la cookie de sesión
	http.HandleFunc("/me", handlers.MeHandler(dbSuper))
	// Endpoint CRUD para administradores (API)
	http.HandleFunc("/super/api/administradores", handlers.AdministradoresHandler(dbSuper))
	// Endpoint CRUD para licencias (nuevo)
	http.HandleFunc("/super/api/licencias", handlers.LicenciasHandler(dbSuper))
	// Endpoint para gestionar credenciales de Mercado Pago (GET/PUT)
	http.HandleFunc("/super/api/config/mercadopago", handlers.MercadoPagoConfigHandler(dbSuper))
	// Endpoint para gestionar credenciales de Wompi (GET/PUT)
	http.HandleFunc("/super/api/config/wompi", handlers.WompiConfigHandler(dbSuper))
	// Endpoint para gestionar SMTP Gmail (GET/PUT)
	http.HandleFunc("/super/api/config/gmail", handlers.GmailConfigHandler(dbSuper))
	// Endpoint para gestionar credenciales IA de modelos populares (GET/PUT)
	http.HandleFunc("/super/api/config/ai", handlers.AIModelsConfigHandler(dbSuper))
	// Endpoints for Mercado Pago integration (crear preferencia y webhook)
	http.HandleFunc("/mercadopago/create_preference", handlers.MercadoPagoCreatePreferenceHandler(dbSuper))
	http.HandleFunc("/mercadopago/webhook", handlers.MercadoPagoWebhookHandler(dbSuper))
	http.HandleFunc("/mercadopago/test_preference", handlers.MercadoPagoTestPreferenceHandler(dbSuper))
	// Endpoints Wompi (Nequi): crear transacción y consultar estado
	http.HandleFunc("/wompi/terms", handlers.WompiTermsHandler(dbSuper))
	http.HandleFunc("/wompi/create_transaction_nequi", handlers.WompiCreateNequiTransactionHandler(dbSuper))
	http.HandleFunc("/wompi/transaction_status", handlers.WompiTransactionStatusHandler(dbSuper))
	// Activación manual de licencia sin pago (uso interno de avance/prototipo)
	http.HandleFunc("/licencias/activar_sin_pago", handlers.ActivateLicenciaSinPagoHandler(dbSuper))
	// Confirmación de correo para usuarios de empresa.
	http.HandleFunc("/auth/confirmar_correo", handlers.ConfirmarCorreoUsuarioHandler(dbEmpresas))

	// Endpoints de métricas (actual y histórico)
	http.HandleFunc("/super/api/metrics/current", handlers.MetricsCurrentHandler(dbSuper))
	http.HandleFunc("/super/api/metrics/history", handlers.MetricsHistoryHandler(dbSuper))
	// Endpoint de seguridad: escaneo de puertos
	http.HandleFunc("/super/api/security/ports", handlers.SecurityPortsHandler(dbSuper))
	// Endpoint de seguridad: listado de procesos en memoria RAM
	http.HandleFunc("/super/api/security/processes", handlers.SecurityProcessesHandler(dbSuper))

	// Logout handler: limpiar cookie de sesión (si existe) y redirigir a la página de login
	http.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		// Invalidate common session cookie names
		cookies := []string{"session", "sid", "auth"}
		for _, name := range cookies {
			// set cookie expired
			http.SetCookie(w, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1})
		}
		// also clear our session_token cookie with same attributes
		http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode})
		// Redirigir al login
		http.Redirect(w, r, "/login.html", http.StatusFound)
	})

	// Determinar carpeta `web` (probar ./web, ../web, y relativo al ejecutable)
	webDir := "web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		alt := "../web"
		if _, err2 := os.Stat(alt); err2 == nil {
			webDir = alt
		} else if exe, err3 := os.Executable(); err3 == nil {
			cand := filepath.Join(filepath.Dir(exe), "..", "web")
			if _, err4 := os.Stat(cand); err4 == nil {
				webDir = cand
			}
		}
	}

	// Servir assets centralizados (CSS, JS)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(webDir))))

	// Servir páginas estáticas desde la carpeta `web` detectada
	// Verificar existencia de index.html y loguear la ruta usada
	indexPath := filepath.Join(webDir, "index.html")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		log.Printf("Warning: index.html no encontrado en %s\n", indexPath)
	} else if err != nil {
		log.Printf("Warning: error comprobando index.html en %s: %v\n", indexPath, err)
	} else {
		log.Printf("index.html encontrado en %s\n", indexPath)
	}
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Wrap DefaultServeMux with authentication, JSON error normalization and logging middleware
	handler := utils.LoggingMiddleware(utils.JSONErrorMiddleware(utils.AuthMiddleware(dbSuper, http.DefaultServeMux)))

	// Respetar la variable de entorno PORT si está definida; por defecto usar 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Println("Servidor arrancado en", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
