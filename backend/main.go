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
	clientID     = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL  = os.Getenv("GOOGLE_REDIRECT_URL") // e.g. http://localhost:8080/auth/google/callback
	dbPath       = os.Getenv("DB_PATH")
	db           *sql.DB
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
	if dbPath == "" {
		dbPath = "pos.db"
	}
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/google/callback"
		log.Println("INFO: usando redirectURL por defecto:", redirectURL)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open sqlite db: %v", err)
	}
	// Asegurar tabla users
	createSQL := `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT UNIQUE,
        name TEXT,
        role TEXT DEFAULT 'administrador',
        created_at TEXT DEFAULT (datetime('now','localtime'))
    );`
	if _, err := db.Exec(createSQL); err != nil {
		log.Fatalf("failed to create users table: %v", err)
	}

	http.HandleFunc("/auth/google/login", handlers.HandleGoogleLogin(clientID, redirectURL))
	http.HandleFunc("/auth/google/callback", handlers.HandleGoogleCallback(db, clientID, clientSecret, redirectURL))

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
