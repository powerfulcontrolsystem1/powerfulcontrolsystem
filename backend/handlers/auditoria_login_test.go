package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoginAuditBuildsProfessionalSuccessEventWithoutSecrets(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/empresa/usuarios/login", strings.NewReader(`{"password":"no-debe-aparecer"}`))
	r.Header.Set("User-Agent", "PCS-Test")
	attempt := &loginAuditAttempt{
		EmpresaID:     12,
		Email:         " Operador@Example.com ",
		PrincipalType: "usuario_empresa",
		AuthMethod:    "password",
	}
	attempt.markAuthenticated("cajero")
	event := attempt.buildEvent(r, http.StatusOK)
	if event.Modulo != loginAuditModule || event.Accion != "inicio_sesion_exitoso" || event.Resultado != "ok" {
		t.Fatalf("evento exitoso invalido: %+v", event)
	}
	if event.EmpresaID != 12 || event.UsuarioCreador != "operador@example.com" || event.CodigoHTTP != http.StatusOK {
		t.Fatalf("alcance o identidad incorrectos: %+v", event)
	}
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(event.MetadataJSON), &metadata); err != nil {
		t.Fatalf("metadata invalida: %v", err)
	}
	serialized := strings.ToLower(event.MetadataJSON + event.Observaciones)
	if strings.Contains(serialized, "no-debe-aparecer") || strings.Contains(serialized, `"password":"`) {
		t.Fatalf("la auditoria no debe copiar credenciales: %s", serialized)
	}
}

func TestLoginAuditSuperUIContract(t *testing.T) {
	read := func(parts ...string) string {
		raw, err := os.ReadFile(filepath.Join(parts...))
		if err != nil {
			t.Fatal(err)
		}
		return string(raw)
	}
	menu := read("..", "..", "web", "super_administrador.html")
	page := read("..", "..", "web", "super", "auditoria_login.html")
	allowlist := read("..", "..", "web", "js", "super_administrador.js")
	if !strings.Contains(menu, `href="/super/auditoria_login.html"`) || !strings.Contains(allowlist, `"/super/auditoria_login.html": true`) {
		t.Fatal("el panel super debe exponer y permitir la pagina de auditoria de login")
	}
	for _, marker := range []string{`data-audit-scope="super_panel"`, `data-audit-default-modulo="autenticacion"`, `id="resultado"`, `id="empresaId"`, `id="btnCSV"`, `id="btnJSON"`, "nunca almacena contraseñas ni tokens"} {
		if !strings.Contains(page, marker) {
			t.Fatalf("auditoria_login.html no contiene %q", marker)
		}
	}
}

func TestLoginAuditClassifiesRejectedAndInternalAttempts(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/super/api/administradores/login", nil)
	attempt := &loginAuditAttempt{Email: "admin@example.com", PrincipalType: "administrador"}
	rejected := attempt.buildEvent(r, http.StatusUnauthorized)
	if rejected.Resultado != "rechazado" || rejected.Accion != "inicio_sesion_fallido" || !strings.Contains(rejected.MetadataJSON, "credenciales_invalidas") {
		t.Fatalf("rechazo mal clasificado: %+v", rejected)
	}
	internal := attempt.buildEvent(r, http.StatusServiceUnavailable)
	if internal.Resultado != "error" || !strings.Contains(internal.MetadataJSON, "error_interno") {
		t.Fatalf("error interno mal clasificado: %+v", internal)
	}
}
