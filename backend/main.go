package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"github.com/you/pos-backend/handlers"
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

func main() {
	if clientID == "" || clientSecret == "" || redirectURL == "" {
		log.Println("Warning: GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET or GOOGLE_REDIRECT_URL not set")
		// Intentar cargar credenciales desde documentos/descripcion_del_proyecto si existen
		if cid, csec := utils.TryLoadCredsFromDocs(); cid != "" || csec != "" {
			if clientID == "" {
				clientID = cid
			}
			if clientSecret == "" {
				clientSecret = csec
			}
		}
	}
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

	createEmpresas := `CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
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
		addIfMissing("fecha_actualizacion TEXT", "fecha_actualizacion")
		addIfMissing("usuario_creador TEXT", "usuario_creador")
		addIfMissing("estado TEXT DEFAULT 'activo'", "estado")
		addIfMissing("observaciones TEXT", "observaciones")
	}
	ensureEmpresasSchema(dbEmpresas)
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

	// Crear tablas en dbSuper (superadministrador)
	createAdmins := `CREATE TABLE IF NOT EXISTS administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
		role TEXT DEFAULT 'administrador',
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbSuper.Exec(createAdmins); err != nil {
		log.Fatalf("failed to create administradores table in super db: %v", err)
	}

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

	http.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin(clientID, redirectURL))
	// Pasar la conexión de la base `empresas` al callback para persistir usuarios y empresas
	// Pasar tanto la conexión de empresas como la de superadministrador al callback
	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback(dbEmpresas, dbSuper, clientID, clientSecret, redirectURL))

	// Endpoints para administración y auditoría (listar administradores y sesiones)
	http.HandleFunc("/super/administradores", handlers.ListAdministradoresHandler(dbSuper))
	http.HandleFunc("/super/sesiones", handlers.ListSesionesHandler(dbSuper))

	// Endpoints CRUD para tipos de empresas
	http.HandleFunc("/super/api/tipos_empresas", handlers.TiposEmpresasHandler(dbSuper))
	// Endpoint CRUD para empresas (guardadas en empresas.db)
	http.HandleFunc("/super/api/empresas", handlers.EmpresasHandler(dbEmpresas))
	// Endpoint CRUD para administradores (API)
	http.HandleFunc("/super/api/administradores", handlers.AdministradoresHandler(dbSuper))
	// Endpoint CRUD para licencias (nuevo)
	http.HandleFunc("/super/api/licencias", handlers.LicenciasHandler(dbSuper))

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

	// Wrap DefaultServeMux with authentication and logging middleware
	handler := utils.LoggingMiddleware(utils.AuthMiddleware(dbSuper, http.DefaultServeMux))

	// Respetar la variable de entorno PORT si está definida; por defecto usar 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Println("Servidor arrancado en", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
