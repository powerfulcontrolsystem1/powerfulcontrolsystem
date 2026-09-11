package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// TestRolesVisualPostgresQA is an explicitly enabled, local visual harness.
// It uses the actual role editor and authorization/CRUD handlers against a new
// synthetic PostgreSQL schema. The actor selector replaces login only; it never
// injects an authorization snapshot or TenantContext. This is not a login or
// production end-to-end test. See the permissions matrix's QA commands.
//
// Run with PCS_ROLES_VISUAL_QA=1, PCS_TEST_POSTGRES_DSN pointing to an isolated
// local test database and, optionally, PCS_ROLES_VISUAL_WEB_ROOT pointing to the
// integrated worktree's web directory. Use go test -timeout 12m -count=1 -v
// ./handlers -run ^TestRolesVisualPostgresQA$. POST /qa/stop ends the run early.
func TestRolesVisualPostgresQA(t *testing.T) {
	if os.Getenv("PCS_ROLES_VISUAL_QA") != "1" {
		t.Skip("interactive local visual QA is disabled")
	}
	dsn := strings.TrimSpace(os.Getenv("PCS_TEST_POSTGRES_DSN"))
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || (u.Hostname() != "127.0.0.1" && u.Hostname() != "localhost" && u.Hostname() != "::1") {
		t.Fatal("visual QA requires a PostgreSQL URL for a local isolated test database")
	}
	webRoot := strings.TrimSpace(os.Getenv("PCS_ROLES_VISUAL_WEB_ROOT"))
	if webRoot == "" {
		webRoot = filepath.Join("..", "..", "web")
	}
	webRoot, err = filepath.Abs(webRoot)
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(filepath.Join(webRoot, "administrar_empresa", "administrar_usuarios.html"))
	if err != nil {
		t.Fatal("read real role editor page: ", err)
	}
	panel := regexp.MustCompile(`(?s)<section\b[^>]*\bid="rolePermissionsPanel"[^>]*>.*?</section>`).Find(source)
	if len(panel) == 0 {
		t.Fatal("real role editor panel was not found")
	}
	pageStyle := regexp.MustCompile(`(?s)<style\b[^>]*>.*?</style>`).Find(source)
	if _, err := os.Stat(filepath.Join(webRoot, "js", "empresa_role_permissions.js")); err != nil {
		t.Fatal("real role editor script is required: ", err)
	}

	admin, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("roles_visual_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer admin.Close()
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("clean visual QA schema: %v", err)
		}
	})
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	conn, err := sql.Open(dbpkg.PostgresCompatDriverName(), u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	conn.SetMaxOpenConns(4)
	setupRolesVisualPostgresFixture(t, conn)

	stop := make(chan struct{})
	var stopOnce sync.Once
	mux := http.NewServeMux()
	mux.HandleFunc("/api/empresa/permisos_contexto", WithEmpresaSeguridadPermissions(conn, conn, EmpresaPermisosContextoHandler(conn)))
	mux.HandleFunc("/api/empresa/roles_de_usuario", WithEmpresaSeguridadPermissions(conn, conn, EmpresaRolDeUsuarioPermisosHandler(conn)))
	// This endpoint probes the real sales authorization middleware without
	// creating a sale or fabricating a successful business operation.
	mux.HandleFunc("/api/empresa/roles_visual_read", WithEmpresaVentasPermissions(conn, conn, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"authorized": true, "probe": "ventas:R", "empresa_id": extractEmpresaIDForPermissions(r)})
	}))
	mux.HandleFunc("/qa/actor", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		actor := r.URL.Query().Get("actor")
		if actor != "admin" && actor != "cajero" && actor != "empresa_b" {
			http.Error(w, "unknown synthetic actor", http.StatusBadRequest)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "pcs_roles_qa_actor", Value: actor, Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode})
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/qa/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		stopOnce.Do(func() { close(stop) })
	})
	for _, asset := range []string{"/estilos.css", "/js/empresa_role_permissions.js"} {
		asset := asset
		mux.HandleFunc(asset, func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, filepath.Join(webRoot, filepath.FromSlash(strings.TrimPrefix(asset, "/"))))
		})
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		actor := rolesVisualActor(r)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		body := strings.NewReplacer("__REAL_STYLE__", string(pageStyle), "__REAL_PANEL__", string(panel), "__ACTOR__", html.EscapeString(actor)).Replace(rolesVisualHTML)
		fmt.Fprint(w, body)
	})
	qaHandler := rolesVisualSyntheticIdentity(mux)
	for _, check := range []struct {
		actor string
		path  string
		allow bool
	}{
		{"admin", "/api/empresa/permisos_contexto?empresa_id=101", true},
		{"admin", "/api/empresa/permisos_contexto?empresa_id=202", false},
		{"cajero", "/api/empresa/permisos_contexto?empresa_id=101", true},
		{"empresa_b", "/api/empresa/permisos_contexto?empresa_id=202", true},
		{"cajero", "/api/empresa/roles_visual_read?empresa_id=101", true},
		{"empresa_b", "/api/empresa/roles_visual_read?empresa_id=202", true},
		{"cajero", "/api/empresa/roles_visual_read?empresa_id=202", false},
	} {
		r := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8875"+check.path, nil)
		r.RemoteAddr = "127.0.0.1:12345"
		r.AddCookie(&http.Cookie{Name: "pcs_roles_qa_actor", Value: check.actor})
		w := httptest.NewRecorder()
		qaHandler.ServeHTTP(w, r)
		allowed := w.Code >= 200 && w.Code < 300
		denied := w.Code == http.StatusForbidden || w.Code == http.StatusBadRequest
		if (check.allow && !allowed) || (!check.allow && !denied) {
			t.Fatalf("synthetic preflight %s %s: HTTP %d: %s", check.actor, check.path, w.Code, w.Body.String())
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:8875")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: qaHandler, ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 30 * time.Second}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			server.Close()
			t.Errorf("stop visual QA server: %v", err)
		}
	})
	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(listener) }()
	t.Log("Synthetic visual QA ready: http://127.0.0.1:8875 (real editor, Go handlers and PostgreSQL; synthetic login; expires in 10 minutes)")
	timer := time.NewTimer(10 * time.Minute)
	defer timer.Stop()
	select {
	case <-stop:
	case <-timer.C:
	case err := <-serveErr:
		if err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}
}

func rolesVisualActor(r *http.Request) string {
	if cookie, err := r.Cookie("pcs_roles_qa_actor"); err == nil && (cookie.Value == "cajero" || cookie.Value == "empresa_b") {
		return cookie.Value
	}
	return "admin"
}

func rolesVisualSyntheticIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(host)
		if err != nil || ip == nil || !ip.IsLoopback() || r.Host != "127.0.0.1:8875" {
			http.Error(w, "local QA only", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != "http://127.0.0.1:8875" {
			http.Error(w, "same-origin QA only", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		principalType, principalID, empresaID := "administrador", int64(1), int64(0)
		switch rolesVisualActor(r) {
		case "cajero":
			principalType, principalID, empresaID = "empresa_usuario", 501, 101
		case "empresa_b":
			principalType, principalID, empresaID = "empresa_usuario", 502, 202
		}
		ctx := context.WithValue(r.Context(), "adminEmail", "shared@example.invalid")
		ctx = context.WithValue(ctx, "sessionPrincipalType", principalType)
		ctx = context.WithValue(ctx, "sessionPrincipalID", principalID)
		ctx = context.WithValue(ctx, "sessionEmpresaID", empresaID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func setupRolesVisualPostgresFixture(t *testing.T, conn *sql.DB) {
	t.Helper()
	for _, statement := range []string{
		`CREATE TABLE tipos_de_empresas (id BIGINT PRIMARY KEY, nombre TEXT)`,
		`CREATE TABLE roles_de_usuario (id BIGINT PRIMARY KEY, empresa_id BIGINT DEFAULT 0, tipo_empresa_id BIGINT DEFAULT 0, nombre TEXT, descripcion TEXT, origen TEXT DEFAULT 'global', rol_base_id BIGINT DEFAULT 0, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT)`,
		`CREATE TABLE empresa_permisos_modulos (empresa_id BIGINT, modulo TEXT, accion TEXT, permitido INT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE empresa_permisos_paginas (empresa_id BIGINT, pagina_clave TEXT, permitido INT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE empresas (id BIGINT PRIMARY KEY, empresa_id BIGINT, nombre TEXT, nit TEXT, tipo_id BIGINT, tipo_nombre TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT)`,
		`CREATE TABLE licencias (id BIGINT PRIMARY KEY, empresa_id BIGINT, nombre TEXT, modulos_habilitados TEXT, super_rol_habilitado INT, activo INT, fecha_inicio TEXT, fecha_fin TEXT)`,
		`CREATE TABLE empresa_licencias_adicionales (id BIGINT PRIMARY KEY, empresa_id BIGINT, licencia_id BIGINT, activo INT, fecha_inicio TEXT, fecha_fin TEXT)`,
		`CREATE TABLE users (id BIGINT PRIMARY KEY, empresa_id BIGINT, email TEXT, name TEXT, documento_identidad TEXT, rol_usuario_id BIGINT, role TEXT, foto_url TEXT, control_aseo_estaciones INT, email_confirmado INT, email_confirm_token TEXT, email_confirm_expira TEXT, email_confirmado_en TEXT, acepta_contrato INT, contrato_version_aceptada INT, fecha_acepta_contrato TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT)`,
		`CREATE TABLE administradores (id BIGINT PRIMARY KEY, email TEXT, name TEXT, role TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE admin_principal_delegaciones (admin_email TEXT, principal_email TEXT, estado TEXT, fecha_revocada TEXT)`,
		`CREATE TABLE admin_empresa_compartida (id BIGINT PRIMARY KEY, empresa_id BIGINT, admin_email TEXT, compartido_por_email TEXT, invitacion_id BIGINT, nivel_acceso TEXT, modulos_permitidos TEXT, puede_compartir BOOL, fecha_aceptada TEXT, fecha_revocada TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT)`,
		`CREATE TABLE configuraciones (config_key TEXT PRIMARY KEY, value TEXT, encrypted INT, fecha_creacion TEXT, fecha_actualizacion TEXT)`,
		`CREATE TABLE empresa_api_rate_limits (empresa_id BIGINT, scope TEXT, window_start TIMESTAMPTZ, request_count BIGINT, updated_at TIMESTAMPTZ, PRIMARY KEY (empresa_id, scope))`,
		`INSERT INTO roles_de_usuario (id,nombre) VALUES (1,'admin_empresa'),(2,'cajero')`,
		`INSERT INTO roles_de_usuario (id,nombre,empresa_id,rol_base_id,origen) VALUES (11,'Caja PCS A',101,2,'empresa'),(12,'Caja PCS B',202,2,'empresa')`,
		`INSERT INTO administradores (id,email,name,role) VALUES (1,'shared@example.invalid','Administrador QA A','administrador'),(2,'other@example.invalid','Administrador QA B','administrador')`,
		`INSERT INTO empresas (id,empresa_id,nombre,usuario_creador) VALUES (101,101,'PCS QA A','shared@example.invalid'),(202,202,'PCS QA B','other@example.invalid')`,
		`INSERT INTO licencias (id,empresa_id,nombre,modulos_habilitados,super_rol_habilitado,activo) VALUES (101,101,'Licencia QA','',1,1),(202,202,'Licencia QA','',1,1)`,
		`INSERT INTO users (id,empresa_id,email,name,rol_usuario_id,email_confirmado) VALUES (501,101,'shared@example.invalid','Cajero QA A',11,1),(502,202,'shared@example.invalid','Cajero QA B',12,1)`,
	} {
		if _, err := conn.Exec(statement); err != nil {
			t.Fatal("create synthetic visual fixture: ", err)
		}
	}
	// DDL is limited to creating this disposable test fixture, before serving.
	for _, ensure := range []func(*sql.DB) error{dbpkg.EnsurePostgresRuntimeCompat, dbpkg.EnsureRolesPermisosSchema, dbpkg.EnsureEmpresaConfiguracionOperativaSchema, dbpkg.EnsureEmpresaAuditoriaSchema} {
		if err := ensure(conn); err != nil {
			t.Fatal("prepare synthetic visual fixture: ", err)
		}
	}
}

const rolesVisualHTML = `<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>QA local · Roles y permisos</title><link rel="stylesheet" href="/estilos.css">__REAL_STYLE__
<style>body{margin:0;background:#f3f6fa;color:#182a43}main{max-width:1180px;margin:auto;padding:24px}.qa-banner{background:#172c4d;color:white;padding:20px;border-radius:12px}.qa-banner h1{margin-top:0;font-size:1.4rem}.qa-controls{display:flex;gap:12px;flex-wrap:wrap;align-items:center;margin:18px 0}.qa-controls button,.qa-controls select{padding:10px;min-height:42px}#qaResult{background:#fff;padding:16px;white-space:pre-wrap;overflow-wrap:anywhere;max-height:360px;overflow:auto;border:1px solid #ccd5e1;border-radius:8px}#rolePermissionsPanel{margin-top:22px}select{max-width:100%}@media(max-width:600px){main{padding:12px}.qa-controls>*{width:100%}}</style></head><body><main>
<div class="qa-banner"><h1>Roles y permisos · QA sintética local</h1><p>Editor del producto, handlers Go y PostgreSQL reales. La identidad se selecciona para pruebas; no se está validando el inicio de sesión ni producción. No hay cuentas, correos enviados ni datos reales.</p><p>Las tres identidades usan el mismo correo sintético. El administrador pertenece solo a A; cada cajero tiene su propia empresa y rol.</p></div>
<div class="qa-controls"><label for="qaActor">Identidad de prueba</label><select id="qaActor"><option value="admin">Administrador global de A</option><option value="cajero">Cajero operativo de A</option><option value="empresa_b">Cajero operativo de B</option></select><button id="qaStop" type="button">Terminar QA y limpiar datos</button></div>
<div class="qa-controls"><button type="button" class="custom-role-permissions" data-empresa-id="101" data-role-id="11" data-role-name="Caja PCS A">Abrir editor real · Caja A</button><button type="button" class="custom-role-permissions" data-empresa-id="202" data-role-id="12" data-role-name="Caja PCS B">Abrir editor real · Caja B</button></div>
<p>Para comprobar revocación: como administrador, desactiva Consultar en Ventas del rol A y guarda. Cambia a Cajero A y consulta ventas; después prueba Cajero B. La sonda consulta el permiso real y no crea ventas.</p>
<div class="qa-controls"><button type="button" data-probe="/api/empresa/permisos_contexto?empresa_id=101">Contexto real A</button><button type="button" data-probe="/api/empresa/permisos_contexto?empresa_id=202">Contexto real B</button><button type="button" data-probe="/api/empresa/roles_visual_read?empresa_id=101">Probar permiso Ventas A</button><button type="button" data-probe="/api/empresa/roles_visual_read?empresa_id=202">Probar permiso Ventas B</button><button type="button" id="qaDeniedWrite">Intentar escritura sin alcance en B</button></div>
<pre id="qaResult" role="status" aria-live="polite">Listo. Selecciona identidad, editor o sonda.</pre>
__REAL_PANEL__
</main><script>
'use strict';const actor=document.getElementById('qaActor');actor.value='__ACTOR__';const result=document.getElementById('qaResult');
async function probe(path,options){try{const response=await fetch(path,{cache:'no-store',credentials:'same-origin',...options});const raw=await response.text();let body=raw;try{body=JSON.stringify(JSON.parse(raw),null,2)}catch{}result.textContent=(options?.method||'GET')+' '+path+'\nHTTP '+response.status+'\n'+body;return response}catch(error){result.textContent=error.message;return null}}
actor.addEventListener('change',async()=>{const response=await probe('/qa/actor?actor='+encodeURIComponent(actor.value),{method:'POST'});if(response?.ok)location.reload()});
document.querySelectorAll('[data-probe]').forEach(button=>button.addEventListener('click',()=>probe(button.dataset.probe)));
document.getElementById('qaDeniedWrite').addEventListener('click',()=>{if(actor.value==='empresa_b'){result.textContent='Selecciona Administrador de A o Cajero A para esta prueba de escritura entre empresas.';return}probe('/api/empresa/roles_de_usuario?action=permisos&empresa_id=202&rol_id=12',{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify({empresa_id:202,rol_id:12,permisos_modulo:[],permisos_pagina:[]})})});
document.getElementById('qaStop').addEventListener('click',async()=>{await probe('/qa/stop',{method:'POST'});result.textContent='Servidor detenido. La prueba limpia el esquema sintético.'});
</script><script src="/js/empresa_role_permissions.js"></script></body></html>`
