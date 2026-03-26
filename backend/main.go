package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
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
	}
	if dbPath == "" {
		dbPath = "pos.db"
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

	http.HandleFunc("/auth/google/login", handleGoogleLogin)
	http.HandleFunc("/auth/google/callback", handleGoogleCallback)

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
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Wrap DefaultServeMux with a logging middleware
	handler := loggingMiddleware(http.DefaultServeMux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Println("Servidor arrancado en", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}

// loggingResponseWriter captura el código de estado para logging
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware registra método, ruta, remoto y código de respuesta
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		log.Printf("-> %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(lrw, r)
		log.Printf("<- %d %s %s", lrw.status, r.Method, r.URL.Path)
	})
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := "state-token" // en producción generar state aleatorio y validarlo
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?%s",
		url.Values{
			"client_id":     {clientID},
			"redirect_uri":  {redirectURL},
			"response_type": {"code"},
			"scope":         {"openid email profile"},
			"access_type":   {"offline"},
			"state":         {state},
		}.Encode())

	http.Redirect(w, r, authURL, http.StatusFound)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if errStr := q.Get("error"); errStr != "" {
		http.Error(w, "error from provider: "+errStr, http.StatusBadRequest)
		return
	}
	code := q.Get("code")
	if code == "" {
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}

	// Intercambiar código por token
	tokenResp, err := exchangeCodeForToken(code)
	if err != nil {
		log.Println("token exchange error:", err)
		http.Error(w, "token exchange failed", http.StatusInternalServerError)
		return
	}

	// Solicitar userinfo
	userinfo, err := fetchUserInfo(tokenResp.AccessToken)
	if err != nil {
		log.Println("fetch userinfo error:", err)
		http.Error(w, "failed to fetch userinfo", http.StatusInternalServerError)
		return
	}

	// Guardar/actualizar usuario en SQLite usando driver Go
	if err := upsertUser(db, userinfo.Email, userinfo.Name); err != nil {
		log.Println("db upsert error:", err)
		// No bloquear al usuario; solo loguear
	}

	// Redirigir al flujo de selección de empresa
	http.Redirect(w, r, "/seleccionar_empresa.html", http.StatusFound)
}

// exchangeCodeForToken realiza POST a Google token endpoint
func exchangeCodeForToken(code string) (*tokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token endpoint returned %d: %s", resp.StatusCode, string(b))
	}
	var tr tokenResponse
	if err := json.Unmarshal(b, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// fetchUserInfo solicita el endpoint userinfo
func fetchUserInfo(accessToken string) (*userInfo, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo endpoint %d: %s", resp.StatusCode, string(b))
	}
	var u userInfo
	if err := json.Unmarshal(b, &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func upsertUser(db *sql.DB, email, name string) error {
	// Insertar si no existe (rol por defecto 'administrador' en la definición de tabla)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT OR IGNORE INTO users (email, name) VALUES (?, ?)", email, name); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE users SET name = ? WHERE email = ?", name, email); err != nil {
		return err
	}
	return tx.Commit()
}

// Tipos para respuestas
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	// Otros campos omitidos
}

type userInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}
