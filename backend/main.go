package main

import (
	"database/sql"
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
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbEmpresas.Exec(createEmpresas); err != nil {
		log.Fatalf("failed to create empresas table in empresas db: %v", err)
	}

	createTipos := `CREATE TABLE IF NOT EXISTS tipos_de_empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT NOT NULL,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if _, err := dbEmpresas.Exec(createTipos); err != nil {
		log.Fatalf("failed to create tipos_de_empresas table in empresas db: %v", err)
	}

	// Crear tablas en dbSuper (superadministrador)
	createAdmins := `CREATE TABLE IF NOT EXISTS administradores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT UNIQUE,
		name TEXT,
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
		nombre TEXT NOT NULL,
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
		fecha_inicio TEXT,
		fecha_fin TEXT,
		activo INTEGER DEFAULT 1,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime'))
	);`
	if _, err := dbSuper.Exec(createLic); err != nil {
		log.Fatalf("failed to create licencias table in super db: %v", err)
	}

	http.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin(clientID, redirectURL))
	// Pasar la conexión de la base `empresas` al callback para persistir usuarios y empresas
	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback(dbEmpresas, clientID, clientSecret, redirectURL))

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

	// Wrap DefaultServeMux with a logging middleware
	handler := utils.LoggingMiddleware(http.DefaultServeMux)

	// Respetar la variable de entorno PORT si está definida; por defecto usar 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Println("Servidor arrancado en", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
