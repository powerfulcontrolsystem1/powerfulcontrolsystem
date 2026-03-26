package utils

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
