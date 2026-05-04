package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddlewarePublicAndProtectedSuperRoutes(t *testing.T) {
	t.Parallel()

	handler := AuthMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	publicPaths := []string{
		"/super/api/administradores/register",
		"/super/api/administradores/login",
		"/super/api/administradores/solicitar_recuperacion",
		"/super/api/administradores/restablecer_password",
		"/api/public/venta_publica",
		"/api/public/market_symbol",
		"/api/onlyoffice/file",
		"/api/onlyoffice/callback",
		"/red_social_comercial.html",
		"/emulador",
		"/emulador/",
		"/emulador/emulator/data/loader.js",
		"/emulador/api/roms",
	}
	for _, path := range publicPaths {
		path := path
		t.Run("public "+path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusNoContent {
				t.Fatalf("expected public route %s to pass without session, got status %d", path, rec.Code)
			}
		})
	}

	protectedPaths := []string{
		"/super/api/empresas",
		"/super/api/licencias",
		"/super/api/config/epayco",
		"/super/api/reportes_globales",
	}
	for _, path := range protectedPaths {
		path := path
		t.Run("protected "+path, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected protected route %s to require session, got status %d", path, rec.Code)
			}
		})
	}
}
