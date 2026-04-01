package utils

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
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
		// Rutas públicas exactas (no usar prefijo "/" porque abriría todo el sistema).
		publicExact := map[string]struct{}{
			"/":                     {},
			"/index.html":           {},
			"/login.html":           {},
			"/auth/google/login":    {},
			"/auth/google/callback": {},
			"/auth/logout":          {},
			"/estilos.css":          {},
			"/menu.js":              {},
			"/favicon.ico":          {},
		}
		if _, ok := publicExact[path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		// Recursos estáticos públicos
		publicPrefixes := []string{"/assets/", "/img/", "/ayuda/"}
		for _, p := range publicPrefixes {
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

// getEncKeyFromEnv intenta obtener la clave de cifrado desde la variable `CONFIG_ENC_KEY`.
// Se admite una cadena Base64 (preferida) o una cadena de al menos 32 bytes.
func getEncKeyFromEnv() ([]byte, error) {
	s := os.Getenv("CONFIG_ENC_KEY")
	if s == "" {
		return nil, fmt.Errorf("CONFIG_ENC_KEY not set")
	}
	// intentar decodificar base64
	if decoded, err := base64.StdEncoding.DecodeString(s); err == nil {
		if len(decoded) >= 16 {
			// normalizar a 32 bytes (AES-256) si es necesario
			if len(decoded) >= 32 {
				return decoded[:32], nil
			}
			key := make([]byte, 32)
			copy(key, decoded)
			return key, nil
		}
	}
	// usar bytes directos si la longitud es suficiente
	if len(s) >= 32 {
		return []byte(s)[:32], nil
	}
	return nil, fmt.Errorf("CONFIG_ENC_KEY invalid length; provide base64 or >=32 bytes")
}

// EncryptString cifra texto plano usando AES-GCM y devuelve Base64(nonce|ciphertext)
func EncryptString(plain string) (string, error) {
	key, err := getEncKeyFromEnv()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nil, nonce, []byte(plain), nil)
	out := append(nonce, ct...)
	return base64.StdEncoding.EncodeToString(out), nil
}

// DecryptString descifra un payload Base64(nonce|ciphertext) usando AES-GCM
func DecryptString(payload string) (string, error) {
	key, err := getEncKeyFromEnv()
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", fmt.Errorf("invalid payload")
	}
	nonce := raw[:ns]
	ct := raw[ns:]
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

// EncryptionAvailable devuelve true si la variable de entorno CONFIG_ENC_KEY está disponible y válida.
func EncryptionAvailable() bool {
	_, err := getEncKeyFromEnv()
	return err == nil
}
