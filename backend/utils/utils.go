package utils

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware registra método, ruta, remoto y código de respuesta
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}
		log.Printf("-> %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(lrw, r)
		log.Printf("<- %d %s %s", lrw.status, r.Method, r.URL.Path)
	})
}

// TryLoadCredsFromDocs busca credenciales en documentos/descripcion_del_proyecto
// Devuelve clientID y clientSecret (vacíos si no se encontraron)
func TryLoadCredsFromDocs() (string, string) {
	candidates := []string{
		"../documentos/descripcion_del_proyecto",
		"./../documentos/descripcion_del_proyecto",
		"../../documentos/descripcion_del_proyecto",
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(exeDir, "..", "documentos", "descripcion_del_proyecto"))
	}

	reClient := regexp.MustCompile(`([0-9]+[-][A-Za-z0-9._-]+apps\.googleusercontent\.com)`) // Google client id pattern
	reSecret := regexp.MustCompile(`(?i)Secreto[^:\n\r]*[:\-\s]*([A-Za-z0-9_\-]+)`)          // busca 'Secreto' y captura token

	var clientID, clientSecret string
	for _, p := range candidates {
		if b, err := os.ReadFile(p); err == nil {
			s := string(b)
			if clientID == "" {
				if m := reClient.FindStringSubmatch(s); len(m) > 1 {
					clientID = m[1]
					log.Println("Cargado GOOGLE_CLIENT_ID desde", p)
				}
			}
			if clientSecret == "" {
				if m2 := reSecret.FindStringSubmatch(s); len(m2) > 1 {
					clientSecret = m2[1]
					log.Println("Cargado GOOGLE_CLIENT_SECRET desde", p)
				}
			}
			if clientID != "" && clientSecret != "" {
				return clientID, clientSecret
			}
		}
	}
	return clientID, clientSecret
}

// GenerateSecureToken devuelve un token seguro en hex de `n` bytes.
func GenerateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// AuthMiddleware protege rutas usando la tabla sesiones y administradores en la BD superadministrador.
// Permite un conjunto público de rutas (login/callback/activos). Para rutas que comienzan con /super/
// exige rol 'super_administrador'. Añade `adminEmail` en el contexto de la petición.
func AuthMiddleware(dbSuper *sql.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Rutas públicas
		public := []string{"/", "/index.html", "/login.html", "/auth/google/login", "/auth/google/callback", "/assets/"}
		for _, p := range public {
			if strings.HasPrefix(path, p) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Obtener token desde cookie o header Authorization
		var token string
		if c, err := r.Cookie("session_token"); err == nil {
			token = c.Value
		} else if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token = strings.TrimPrefix(auth, "Bearer ")
		}
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		sess, err := dbpkg.GetSessionByToken(dbSuper, token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		admin, err := dbpkg.GetAdminByEmail(dbSuper, sess.AdminEmail)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Rutas /super/ requieren rol super_administrador
		if strings.HasPrefix(path, "/super/") && admin.Role != "super_administrador" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Propagar información del admin en el contexto
		ctx := context.WithValue(r.Context(), "adminEmail", admin.Email)
		r = r.WithContext(ctx)
		// Añadir cabecera informativa
		r.Header.Set("X-Admin-Email", admin.Email)

		next.ServeHTTP(w, r)
	})
}
