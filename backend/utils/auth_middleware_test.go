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
		"/api/public/informacion_de_modulos",
		"/api/public/plantillas_nuevas/catalogo",
		"/api/public/plantillas_integracion/catalogo",
		"/api/public/juegos/records",
		"/api/public/licencias/payment_methods",
		"/api/public/licencias/checkout_summary",
		"/licencias/activar_sin_pago",
		"/api/onlyoffice/file",
		"/api/onlyoffice/callback",
		"/descripcion_de_los_sistemas.html",
		"/descripcion_de_los_sistemas.ht",
		"/red_social_comercial.html",
		"/visualizar_productos_y_precios_publico.html",
		"/motel-calipso/visualizar_productos_y_precios_publico.html",
		"/emulador",
		"/emulador/",
		"/emulador/emulator/data/loader.js",
		"/emulador/api/roms",
		"/ayuda/login_administradores.html",
		"/manifest.webmanifest",
		"/sw.js",
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
		"/ayuda/ayuda.html",
		"/super_administrador.html",
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

func TestSuperControlRouteAllowList(t *testing.T) {
	t.Parallel()

	allowed := []string{
		"/super/licencias_resumen.html",
		"/super/administradores.html",
		"/super/api/administradores",
		"/super/api/security/vps/status",
	}
	for _, path := range allowed {
		if !allowSuperControlRoute(path, http.MethodGet, SuperControlRole) {
			t.Fatalf("expected control role to access %s", path)
		}
	}

	blocked := []string{
		"/super/licencias.html",
		"/super/configuracion_avanzada.html",
		"/super/api/config/epayco",
		"/super/api/licencias",
	}
	for _, path := range blocked {
		if allowSuperControlRoute(path, http.MethodGet, SuperControlRole) {
			t.Fatalf("expected control role to be blocked from %s", path)
		}
	}
}

func TestAdminLimitedRouteAllowListIncludesScopedAdministradores(t *testing.T) {
	t.Parallel()

	allowed := []struct {
		path   string
		method string
	}{
		{"/super/administradores.html", http.MethodGet},
		{"/super/api/administradores", http.MethodGet},
		{"/super/api/administradores", http.MethodPost},
		{"/super/api/administradores", http.MethodPut},
		{"/super/api/administradores", http.MethodDelete},
	}
	for _, tc := range allowed {
		if !allowAdminLimitedSuperRoute(tc.path, tc.method, "administrador") {
			t.Fatalf("expected administrador to access %s %s", tc.method, tc.path)
		}
	}

	if allowAdminLimitedSuperRoute("/super/api/administradores", http.MethodGet, "cajero") {
		t.Fatalf("expected non administrador role to be blocked")
	}
}
