package main

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type roleUser struct {
	Role     string
	Email    string
	Name     string
	Document string
}

type endpointResult struct {
	Role       string `json:"role"`
	Module     string `json:"module,omitempty"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	OK         bool   `json:"ok"`
	Expected   bool   `json:"expected"`
	BodySample string `json:"body_sample,omitempty"`
}

type runReport struct {
	EmpresaID      int64            `json:"empresa_id"`
	StartedAt      string           `json:"started_at"`
	UsersPrepared  []roleUser       `json:"users_prepared"`
	LoginByRole    map[string]int   `json:"login_by_role"`
	PermissionInfo map[string]any   `json:"permission_info,omitempty"`
	Endpoints      []endpointResult `json:"endpoints"`
	Failures       []endpointResult `json:"failures"`
}

type qaEndpoint struct {
	Module string
	Path   string
}

func loadEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n") {
		raw := strings.TrimSpace(line)
		if raw == "" || strings.HasPrefix(raw, "#") || !strings.Contains(raw, "=") {
			continue
		}
		parts := strings.SplitN(raw, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
		if key != "" && os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}

func rewriteDSN(raw string) string {
	if strings.TrimSpace(os.Getenv("DB_VPS_TUNNEL_ENABLED")) != "1" {
		return raw
	}
	port := strings.TrimSpace(os.Getenv("DB_VPS_LOCAL_PORT"))
	if port == "" {
		port = "15432"
	}
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	u.Host = "127.0.0.1:" + port
	return u.String()
}

func hashPassword(password, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":" + password))
	return hex.EncodeToString(sum[:])
}

func scalarInt64(db *sql.DB, q string, args ...any) int64 {
	var out sql.NullInt64
	if err := db.QueryRow(q, args...).Scan(&out); err != nil || !out.Valid {
		return 0
	}
	return out.Int64
}

func empresaTipoID(db *sql.DB, empresaID int64) int64 {
	candidates := []string{
		"SELECT COALESCE(tipo_empresa_id,0) FROM empresas WHERE id=$1",
		"SELECT COALESCE(tipo_id,0) FROM empresas WHERE id=$1",
	}
	for _, q := range candidates {
		id := scalarInt64(db, q, empresaID)
		if id > 0 {
			return id
		}
	}
	// Fallback: buscar tipo Motel.
	id := scalarInt64(db, "SELECT id FROM tipos_de_empresas WHERE lower(nombre) LIKE '%motel%' ORDER BY id LIMIT 1")
	if id > 0 {
		return id
	}
	return 1
}

func prepareRole(dbEmp, dbSuper *sql.DB, tipoEmpresaID int64, role string) (int64, error) {
	if err := dbpkg.EnsureRolesDeUsuarioSchema(dbSuper); err != nil {
		return 0, err
	}
	if err := dbpkg.EnsureRolesPermisosSchema(dbSuper); err != nil {
		return 0, err
	}
	id, _, err := dbpkg.UpsertRolDeUsuarioByTipoNombre(dbSuper, tipoEmpresaID, role, "QA operativo Calipso", "qa_operativo")
	return id, err
}

func prepareUser(dbEmp, dbSuper *sql.DB, empresaID, roleID int64, u roleUser, password string) error {
	if err := dbpkg.EnsureEmpresaUsuariosAuthSchema(dbEmp); err != nil {
		return err
	}
	if err := dbpkg.EnsureAdministradoresAuthSchema(dbSuper); err != nil {
		return err
	}
	if err := dbpkg.EnsureAdminEmpresaCompartidaSchema(dbSuper); err != nil {
		return err
	}
	salt := "qa_calipso_" + strings.ReplaceAll(u.Role, "_", "")
	hash := hashPassword(password, salt)
	_, err := dbEmp.Exec(`INSERT INTO users (
		email, name, role, empresa_id, documento_identidad, rol_usuario_id,
		email_confirmado, email_confirmado_en, password_hash, password_salt, password_set,
		password_actualizada_en, acepta_contrato, contrato_version_aceptada, fecha_acepta_contrato,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES ($1,$2,$3,$4,$5,$6,1,CAST(CURRENT_TIMESTAMP AS TEXT),$7,$8,1,CAST(CURRENT_TIMESTAMP AS TEXT),1,1,CAST(CURRENT_TIMESTAMP AS TEXT),$9,'activo',$10,CAST(CURRENT_TIMESTAMP AS TEXT),CAST(CURRENT_TIMESTAMP AS TEXT))
	ON CONFLICT(email) DO UPDATE SET
		name=EXCLUDED.name, role=EXCLUDED.role, empresa_id=EXCLUDED.empresa_id,
		documento_identidad=EXCLUDED.documento_identidad, rol_usuario_id=EXCLUDED.rol_usuario_id,
		email_confirmado=1, email_confirmado_en=CAST(CURRENT_TIMESTAMP AS TEXT),
		password_hash=EXCLUDED.password_hash, password_salt=EXCLUDED.password_salt, password_set=1,
		password_actualizada_en=CAST(CURRENT_TIMESTAMP AS TEXT), acepta_contrato=1,
		contrato_version_aceptada=1, fecha_acepta_contrato=CAST(CURRENT_TIMESTAMP AS TEXT),
		estado='activo', observaciones=EXCLUDED.observaciones, fecha_actualizacion=CAST(CURRENT_TIMESTAMP AS TEXT)`,
		u.Email, u.Name, u.Role, empresaID, u.Document, roleID, hash, salt, "qa_operativo", "QA operativo Motel Calipso")
	if err != nil {
		return err
	}
	if err := dbpkg.UpsertAdministrador(dbSuper, u.Email, u.Name, u.Role, ""); err != nil {
		return err
	}
	_, err = dbpkg.UpsertAdminEmpresaCompartidaAcceso(dbSuper, dbpkg.AdminEmpresaCompartidaAcceso{
		EmpresaID: empresaID, AdminEmail: u.Email, CompartidoPorEmail: "qa_operativo", FechaAceptada: time.Now().Format("2006-01-02 15:04:05"), UsuarioCreador: "qa_operativo", Estado: "activo", Observaciones: "QA operativo Calipso",
	})
	return err
}

func login(client *http.Client, base string, empresaID int64, u roleUser, password string) (int, string) {
	body, _ := json.Marshal(map[string]any{"empresa_id": empresaID, "email": u.Email, "password": password, "accept_contract": true, "recaptcha_token": "dev-bypass"})
	resp, err := client.Post(base+"/api/empresa/usuarios/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return resp.StatusCode, string(b)
}

func doReq(client *http.Client, base, method string, ep qaEndpoint) endpointResult {
	path := ep.Path
	req, _ := http.NewRequest(method, base+path, nil)
	resp, err := client.Do(req)
	if err != nil {
		return endpointResult{Module: ep.Module, Method: method, Path: path, Status: 0, OK: false, BodySample: err.Error()}
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 240))
	sample := strings.TrimSpace(strings.ReplaceAll(string(b), "\n", " "))
	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	return endpointResult{Module: ep.Module, Method: method, Path: path, Status: resp.StatusCode, OK: ok, Expected: ok, BodySample: sample}
}

func main() {
	var empresaID int64
	var base string
	var password string
	flag.Int64Var(&empresaID, "empresa_id", 7, "empresa id")
	flag.StringVar(&base, "base", "http://127.0.0.1:8080", "base URL")
	flag.StringVar(&password, "password", "QaCalipso2026", "QA password")
	flag.Parse()

	loadEnv(".env.local")
	dbEmp, err := sql.Open(dbpkg.PostgresCompatDriverName(), rewriteDSN(os.Getenv("DB_EMPRESAS_DSN")))
	if err != nil {
		log.Fatal(err)
	}
	defer dbEmp.Close()
	dbSuper, err := sql.Open(dbpkg.PostgresCompatDriverName(), rewriteDSN(os.Getenv("DB_SUPERADMIN_DSN")))
	if err != nil {
		log.Fatal(err)
	}
	defer dbSuper.Close()
	if err := dbEmp.Ping(); err != nil {
		log.Fatal("empresas ping: ", err)
	}
	if err := dbSuper.Ping(); err != nil {
		log.Fatal("super ping: ", err)
	}
	_ = dbpkg.EnsurePostgresRuntimeCompat(dbEmp)
	_ = dbpkg.EnsurePostgresRuntimeCompat(dbSuper)

	tipoID := empresaTipoID(dbEmp, empresaID)
	users := []roleUser{
		{Role: "admin_empresa", Email: "qa.admin.calipso@powerfulcontrolsystem.local", Name: "QA Admin Calipso", Document: "QA-ADMIN-7"},
		{Role: "cajero", Email: "qa.cajero.calipso@powerfulcontrolsystem.local", Name: "QA Cajero Calipso", Document: "QA-CAJERO-7"},
		{Role: "contabilidad", Email: "qa.contabilidad.calipso@powerfulcontrolsystem.local", Name: "QA Contabilidad Calipso", Document: "QA-CONT-7"},
		{Role: "auditor", Email: "qa.auditor.calipso@powerfulcontrolsystem.local", Name: "QA Auditor Calipso", Document: "QA-AUD-7"},
	}

	for _, u := range users {
		roleID, err := prepareRole(dbEmp, dbSuper, tipoID, u.Role)
		if err != nil {
			log.Fatalf("prepare role %s: %v", u.Role, err)
		}
		if err := prepareUser(dbEmp, dbSuper, empresaID, roleID, u, password); err != nil {
			log.Fatalf("prepare user %s: %v", u.Email, err)
		}
	}

	report := runReport{EmpresaID: empresaID, StartedAt: time.Now().Format(time.RFC3339), UsersPrepared: users, LoginByRole: map[string]int{}, PermissionInfo: map[string]any{}}
	endpoints := []qaEndpoint{
		{"usuarios", "/api/empresa/usuarios?empresa_id=7&include_inactive=1"},
		{"carritos", "/api/empresa/carritos_compra?empresa_id=7&include_inactive=1&estacion_id=1"},
		{"productos", "/api/empresa/productos?empresa_id=7&include_inactive=1"},
		{"clientes", "/api/empresa/clientes?empresa_id=7&include_inactive=1"},
		{"compras", "/api/empresa/compras/documentos?empresa_id=7"},
		{"proveedores", "/api/empresa/proveedores?empresa_id=7&include_inactive=1"},
		{"facturacion_electronica", "/api/empresa/facturacion_electronica?empresa_id=7"},
		{"impuestos", "/api/empresa/impuestos?action=context&empresa_id=7"},
		{"reservas_hotel", "/api/empresa/reservas_hotel?empresa_id=7"},
		{"tarifas_por_minutos", "/api/empresa/tarifas_por_minutos?empresa_id=7&include_inactive=1"},
		{"tarifas_por_dia", "/api/empresa/tarifas_por_dia?empresa_id=7&include_inactive=1"},
		{"tarifas_motel", "/api/empresa/tarifas_motel?empresa_id=7&include_inactive=1"},
		{"gimnasio", "/api/empresa/gimnasio?empresa_id=7&action=dashboard"},
		{"gimnasio_preconfiguracion", "/api/empresa/gimnasio?empresa_id=7&action=preconfiguracion"},
		{"alquileres", "/api/empresa/alquileres?empresa_id=7&action=dashboard"},
		{"odontologia", "/api/empresa/odontologia?empresa_id=7&action=dashboard"},
		{"taxi_system", "/api/empresa/taxi_system?empresa_id=7&action=dashboard"},
		{"turnos_atencion", "/api/empresa/turnos_atencion?empresa_id=7&action=dashboard"},
		{"horarios_trabajadores", "/api/empresa/horarios_trabajadores?empresa_id=7"},
		{"asistencia_empleados", "/api/empresa/asistencia_empleados?empresa_id=7"},
		{"nomina", "/api/empresa/nomina?empresa_id=7"},
		{"vehiculos_registro", "/api/empresa/vehiculos_registro?empresa_id=7"},
		{"auditoria", "/api/empresa/auditoria/eventos?empresa_id=7&limit=5"},
		{"finanzas", "/api/empresa/finanzas/configuracion?empresa_id=7"},
		{"propinas", "/api/empresa/propinas?empresa_id=7&action=resumen"},
		{"comisiones", "/api/empresa/comisiones?empresa_id=7&action=resumen"},
		{"graficos_estadisticas", "/api/empresa/graficos_estadisticas?empresa_id=7"},
	}

	adminCoveredModules := false
	for _, u := range users {
		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar, Timeout: 45 * time.Second}
		status, body := login(client, base, empresaID, u, password)
		report.LoginByRole[u.Role] = status
		if status < 200 || status >= 300 {
			report.Failures = append(report.Failures, endpointResult{Role: u.Role, Method: "POST", Path: "/api/empresa/usuarios/login", Status: status, BodySample: body})
			continue
		}
		perms := qaEndpoint{Module: "permisos_contexto", Path: fmt.Sprintf("/api/empresa/permisos_contexto?empresa_id=%d&include_matrix=1", empresaID)}
		res := doReq(client, base, "GET", perms)
		res.Role = u.Role
		report.Endpoints = append(report.Endpoints, res)
		if !res.Expected || res.Status >= 500 || res.Status == 404 {
			report.Failures = append(report.Failures, res)
		}

		if u.Role != "admin_empresa" && adminCoveredModules {
			continue
		}
		if u.Role == "admin_empresa" {
			adminCoveredModules = true
		}
		for _, ep := range endpoints {
			res := doReq(client, base, "GET", ep)
			res.Role = u.Role
			report.Endpoints = append(report.Endpoints, res)
			if !res.Expected || res.Status >= 500 || res.Status == 404 {
				report.Failures = append(report.Failures, res)
			}
		}
	}
	sort.Slice(report.Failures, func(i, j int) bool { return report.Failures[i].Path < report.Failures[j].Path })
	out, _ := json.MarshalIndent(report, "", "  ")
	outPath := filepath.Join("tmp_tools", "qa_calipso_operativo", "reporte_operativo_calipso.json")
	_ = os.WriteFile(outPath, out, 0600)
	fmt.Printf("RESULTADO_QA_CALIPSO usuarios=%d endpoints=%d failures=%d reporte=%s\n", len(users), len(report.Endpoints), len(report.Failures), outPath)
	for _, f := range report.Failures {
		fmt.Printf("FAIL role=%s status=%d path=%s sample=%s\n", f.Role, f.Status, f.Path, f.BodySample)
	}
}
