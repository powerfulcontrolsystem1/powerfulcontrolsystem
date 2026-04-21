package handlers

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

func ensureEmpresasCoreSchemaForSuper(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	stmt := `CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre TEXT,
		nit TEXT,
		tipo_id INTEGER,
		tipo_nombre TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	if dbpkg.IsPostgresDialect() {
		stmt = `CREATE TABLE IF NOT EXISTS empresas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT,
			nombre TEXT,
			nit TEXT,
			tipo_id BIGINT,
			tipo_nombre TEXT,
			fecha_creacion TEXT,
			fecha_actualizacion TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`
	}
	_, err := dbEmp.Exec(stmt)
	if err != nil {
		t.Fatalf("create empresas schema: %v", err)
	}
}

func ensureEmpresasImpactSchemaForSuper(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	_, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		email TEXT,
		estado TEXT DEFAULT 'activo'
	);`)
	if err != nil {
		t.Fatalf("create users impact schema: %v", err)
	}

	_, err = dbEmp.Exec(`CREATE TABLE IF NOT EXISTS carritos_compras (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		estado_carrito TEXT,
		estado TEXT DEFAULT 'activo'
	);`)
	if err != nil {
		t.Fatalf("create carritos impact schema: %v", err)
	}

	_, err = dbEmp.Exec(`CREATE TABLE IF NOT EXISTS reservas_hotel (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		estado_reserva TEXT,
		estado TEXT DEFAULT 'activo'
	);`)
	if err != nil {
		t.Fatalf("create reservas impact schema: %v", err)
	}
}

func ensureSuperConfigSchemaForSuper(t *testing.T, dbSuper *sql.DB) {
	t.Helper()
	configStmt := `CREATE TABLE IF NOT EXISTS configuraciones (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_key TEXT UNIQUE,
		value TEXT,
		encrypted INTEGER DEFAULT 0,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT
	);`
	if dbpkg.IsPostgresDialect() {
		configStmt = `CREATE TABLE IF NOT EXISTS configuraciones (
			id BIGSERIAL PRIMARY KEY,
			config_key TEXT UNIQUE,
			value TEXT,
			encrypted INTEGER DEFAULT 0,
			fecha_creacion TEXT,
			fecha_actualizacion TEXT
		)`
	}
	_, err := dbSuper.Exec(configStmt)
	if err != nil {
		t.Fatalf("create configuraciones schema: %v", err)
	}

	licStmt := `CREATE TABLE IF NOT EXISTS licencias (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		tipo_id INTEGER,
		nombre TEXT,
		descripcion TEXT,
		valor REAL,
		duracion_dias INTEGER,
		modulos_habilitados TEXT,
		super_rol_habilitado INTEGER DEFAULT 0,
		fecha_inicio TEXT,
		activo INTEGER DEFAULT 1,
		fecha_fin TEXT,
		fecha_creacion TEXT
	);`
	if dbpkg.IsPostgresDialect() {
		licStmt = `CREATE TABLE IF NOT EXISTS licencias (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT,
			tipo_id BIGINT,
			nombre TEXT,
			descripcion TEXT,
			valor DOUBLE PRECISION,
			duracion_dias INTEGER,
			modulos_habilitados TEXT,
			super_rol_habilitado INTEGER DEFAULT 0,
			fecha_inicio TEXT,
			activo INTEGER DEFAULT 1,
			fecha_fin TEXT,
			fecha_creacion TEXT
		)`
	}
	_, err = dbSuper.Exec(licStmt)
	if err != nil {
		t.Fatalf("create licencias schema: %v", err)
	}

	tiposStmt := `CREATE TABLE IF NOT EXISTS tipos_de_empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT,
		observaciones TEXT,
		estado TEXT DEFAULT 'activo',
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		usuario_creador TEXT
	);`
	if dbpkg.IsPostgresDialect() {
		tiposStmt = `CREATE TABLE IF NOT EXISTS tipos_de_empresas (
			id BIGSERIAL PRIMARY KEY,
			nombre TEXT,
			observaciones TEXT,
			estado TEXT DEFAULT 'activo',
			fecha_creacion TEXT,
			fecha_actualizacion TEXT,
			usuario_creador TEXT
		)`
	}
	_, err = dbSuper.Exec(tiposStmt)
	if err != nil {
		t.Fatalf("create tipos_de_empresas schema: %v", err)
	}
}

func seedEmpresaEstadoForSuper(t *testing.T, dbEmp *sql.DB, id int64, nombre, estado string) {
	t.Helper()
	nowValue := time.Now().Format("2006-01-02 15:04:05")
	_, err := dbpkg.ExecCompat(dbEmp, `
		INSERT INTO empresas (id, empresa_id, nombre, estado, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, id, nombre, estado, nowValue, nowValue)
	if err != nil {
		t.Fatalf("insert empresa seed: %v", err)
	}
}

func TestEmpresasHandlerDesactivarConImpactoYForce(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_super_estado.db")
	dbSuper := openTestSQLite(t, "super_super_estado.db")

	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureEmpresasImpactSchemaForSuper(t, dbEmp)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	seedEmpresaEstadoForSuper(t, dbEmp, 1, "Empresa Uno", "activo")
	if _, err := dbEmp.Exec(`INSERT INTO users (empresa_id, email, estado) VALUES (1, 'user@empresa.com', 'activo')`); err != nil {
		t.Fatalf("insert users impact: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO carritos_compras (empresa_id, estado_carrito, estado) VALUES (1, 'abierto', 'activo')`); err != nil {
		t.Fatalf("insert carritos impact: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO reservas_hotel (empresa_id, estado_reserva, estado) VALUES (1, 'confirmada', 'activo')`); err != nil {
		t.Fatalf("insert reservas impact: %v", err)
	}
	if _, err := dbSuper.Exec(`INSERT INTO licencias (empresa_id, activo) VALUES (1, 1)`); err != nil {
		t.Fatalf("insert licencias impact: %v", err)
	}

	h := EmpresasHandler(dbEmp, dbSuper)

	req := httptest.NewRequest(http.MethodPut, "/super/api/empresas?id=1&action=desactivar", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d without force, got %d body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}

	var conflictBody map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &conflictBody); err != nil {
		t.Fatalf("decode conflict body: %v body=%s", err, rr.Body.String())
	}
	if ok, _ := conflictBody["requiere_confirmacion"].(bool); !ok {
		t.Fatalf("expected requiere_confirmacion=true, got %v", conflictBody["requiere_confirmacion"])
	}

	var estado string
	if err := dbEmp.QueryRow(`SELECT estado FROM empresas WHERE id = 1`).Scan(&estado); err != nil {
		t.Fatalf("query empresa estado after conflict: %v", err)
	}
	if strings.TrimSpace(strings.ToLower(estado)) != "activo" {
		t.Fatalf("estado should remain activo after conflict, got %q", estado)
	}

	forceReq := httptest.NewRequest(http.MethodPut, "/super/api/empresas?id=1&action=desactivar&force=1", nil)
	forceRR := httptest.NewRecorder()
	h.ServeHTTP(forceRR, forceReq)

	if forceRR.Code != http.StatusOK {
		t.Fatalf("expected status %d with force, got %d body=%s", http.StatusOK, forceRR.Code, forceRR.Body.String())
	}

	if err := dbEmp.QueryRow(`SELECT estado FROM empresas WHERE id = 1`).Scan(&estado); err != nil {
		t.Fatalf("query empresa estado after force: %v", err)
	}
	if strings.TrimSpace(strings.ToLower(estado)) != "inactivo" {
		t.Fatalf("estado should be inactivo after force deactivation, got %q", estado)
	}
}

func TestEmpresasHandlerImpactoDesactivacion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_super_impacto.db")
	dbSuper := openTestSQLite(t, "super_super_impacto.db")

	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureEmpresasImpactSchemaForSuper(t, dbEmp)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	seedEmpresaEstadoForSuper(t, dbEmp, 5, "Empresa Impacto", "activo")
	if _, err := dbEmp.Exec(`INSERT INTO users (empresa_id, email, estado) VALUES (5, 'u1@empresa.com', 'activo')`); err != nil {
		t.Fatalf("insert users impacto: %v", err)
	}

	h := EmpresasHandler(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/super/api/empresas?id=5&action=impacto_desactivacion", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode impacto response: %v body=%s", err, rr.Body.String())
	}
	impacto, ok := body["impacto"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected impacto object, got %T", body["impacto"])
	}
	if got, _ := impacto["empresa_id"].(float64); int64(got) != 5 {
		t.Fatalf("expected impacto.empresa_id=5, got %v", impacto["empresa_id"])
	}
	if got, _ := impacto["usuarios_activos"].(float64); int64(got) != 1 {
		t.Fatalf("expected impacto.usuarios_activos=1, got %v", impacto["usuarios_activos"])
	}
}

func TestEmpresasHandlerResumenDescargaYExport(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_super_descarga.db")
	dbSuper := openTestSQLite(t, "super_super_descarga.db")

	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if _, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS clientes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre TEXT,
		email TEXT
	)`); err != nil {
		t.Fatalf("create clientes schema: %v", err)
	}

	seedEmpresaEstadoForSuper(t, dbEmp, 7, "Empresa Exportable", "activo")
	if _, err := dbEmp.Exec(`UPDATE empresas SET nit = '900777111', tipo_nombre = 'Motel', observaciones = 'empresa para exportes' WHERE id = 7`); err != nil {
		t.Fatalf("update empresa seed: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO clientes (empresa_id, nombre, email) VALUES (7, 'Cliente Uno', 'cliente1@test.com')`); err != nil {
		t.Fatalf("insert cliente: %v", err)
	}
	if _, err := dbSuper.Exec(`INSERT INTO licencias (empresa_id, activo, fecha_fin) VALUES (7, 1, '2026-12-31')`); err != nil {
		t.Fatalf("insert licencia export: %v", err)
	}

	h := EmpresasHandler(dbEmp, dbSuper)

	resumenReq := httptest.NewRequest(http.MethodGet, "/super/api/empresas?id=7&action=resumen_descarga", nil)
	resumenRR := httptest.NewRecorder()
	h.ServeHTTP(resumenRR, resumenReq)

	if resumenRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, resumenRR.Code, resumenRR.Body.String())
	}

	var resumenBody map[string]interface{}
	if err := json.Unmarshal(resumenRR.Body.Bytes(), &resumenBody); err != nil {
		t.Fatalf("decode resumen body: %v body=%s", err, resumenRR.Body.String())
	}
	snapshot, ok := resumenBody["snapshot"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected snapshot object, got %T", resumenBody["snapshot"])
	}
	if got, _ := snapshot["total_tables"].(float64); int64(got) < 2 {
		t.Fatalf("expected at least 2 tables, got %v", snapshot["total_tables"])
	}
	if got, _ := snapshot["total_rows"].(float64); int64(got) < 3 {
		t.Fatalf("expected at least 3 rows consolidated, got %v", snapshot["total_rows"])
	}

	exportReq := httptest.NewRequest(http.MethodGet, "/super/api/empresas?id=7&action=exportar_informacion&format=pdf", nil)
	exportRR := httptest.NewRecorder()
	h.ServeHTTP(exportRR, exportReq)

	if exportRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, exportRR.Code, exportRR.Body.String())
	}
	if ct := strings.ToLower(exportRR.Header().Get("Content-Type")); !strings.Contains(ct, "application/pdf") {
		t.Fatalf("expected content-type application/pdf, got %q", exportRR.Header().Get("Content-Type"))
	}
	if cd := strings.ToLower(exportRR.Header().Get("Content-Disposition")); !strings.Contains(cd, ".pdf") {
		t.Fatalf("expected content-disposition with .pdf, got %q", exportRR.Header().Get("Content-Disposition"))
	}
}

func TestEmpresasHandlerEliminarTotalPurgaDatosRelacionados(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_super_eliminacion_total.db")
	dbSuper := openTestSQLite(t, "super_super_eliminacion_total.db")

	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if _, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS clientes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre TEXT
	)`); err != nil {
		t.Fatalf("create clientes schema: %v", err)
	}
	if _, err := dbEmp.Exec(`CREATE TABLE IF NOT EXISTS empresa_backups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		codigo TEXT
	)`); err != nil {
		t.Fatalf("create empresa_backups schema: %v", err)
	}
	if _, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS pagos_wompi (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		reference TEXT
	)`); err != nil {
		t.Fatalf("create pagos_wompi schema: %v", err)
	}

	seedEmpresaEstadoForSuper(t, dbEmp, 10, "Empresa Purga", "activo")
	seedEmpresaEstadoForSuper(t, dbEmp, 11, "Empresa Vecina", "activo")
	if _, err := dbEmp.Exec(`INSERT INTO clientes (empresa_id, nombre) VALUES (10, 'Cliente Purga'), (11, 'Cliente Vecino')`); err != nil {
		t.Fatalf("insert clientes: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO empresa_backups (empresa_id, codigo) VALUES (10, 'BKP-10'), (11, 'BKP-11')`); err != nil {
		t.Fatalf("insert empresa_backups: %v", err)
	}
	if _, err := dbSuper.Exec(`INSERT INTO licencias (empresa_id, activo) VALUES (10, 1), (11, 1)`); err != nil {
		t.Fatalf("insert licencias: %v", err)
	}
	if _, err := dbSuper.Exec(`INSERT INTO pagos_wompi (empresa_id, reference) VALUES (10, 'ref-10'), (11, 'ref-11')`); err != nil {
		t.Fatalf("insert pagos_wompi: %v", err)
	}

	h := EmpresasHandler(dbEmp, dbSuper)
	req := httptest.NewRequest(http.MethodDelete, "/super/api/empresas?id=10&action=eliminar_total", strings.NewReader(`{"confirmacion_nombre":"Empresa Purga"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var empresasCount int
	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM empresas WHERE id = 10`).Scan(&empresasCount); err != nil {
		t.Fatalf("count empresa purge: %v", err)
	}
	if empresasCount != 0 {
		t.Fatalf("expected empresa 10 deleted, got %d rows", empresasCount)
	}
	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM clientes WHERE empresa_id = 10`).Scan(&empresasCount); err != nil {
		t.Fatalf("count clientes purge: %v", err)
	}
	if empresasCount != 0 {
		t.Fatalf("expected clientes purge for empresa 10, got %d rows", empresasCount)
	}
	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM empresa_backups WHERE empresa_id = 10`).Scan(&empresasCount); err != nil {
		t.Fatalf("count backups purge: %v", err)
	}
	if empresasCount != 0 {
		t.Fatalf("expected backups purge for empresa 10, got %d rows", empresasCount)
	}
	if err := dbSuper.QueryRow(`SELECT COUNT(1) FROM licencias WHERE empresa_id = 10`).Scan(&empresasCount); err != nil {
		t.Fatalf("count licencias purge: %v", err)
	}
	if empresasCount != 0 {
		t.Fatalf("expected licencias purge for empresa 10, got %d rows", empresasCount)
	}
	if err := dbSuper.QueryRow(`SELECT COUNT(1) FROM pagos_wompi WHERE empresa_id = 10`).Scan(&empresasCount); err != nil {
		t.Fatalf("count pagos_wompi purge: %v", err)
	}
	if empresasCount != 0 {
		t.Fatalf("expected pagos_wompi purge for empresa 10, got %d rows", empresasCount)
	}

	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM empresas WHERE id = 11`).Scan(&empresasCount); err != nil {
		t.Fatalf("count empresa vecina: %v", err)
	}
	if empresasCount != 1 {
		t.Fatalf("expected empresa 11 intacta, got %d rows", empresasCount)
	}
}

func TestEmpresasHandlerFiltraEmpresasPorAdministradorPrincipal(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_scope_handler.db")
	dbSuper := openTestSQLite(t, "super_scope_handler.db")

	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureAdminAuthTestSchema(t, dbSuper)

	if err := dbpkg.UpsertAdministradorConCreador(dbSuper, "principal@empresa.com", "Principal", "super_administrador", "", ""); err != nil {
		t.Fatalf("upsert principal: %v", err)
	}
	if err := dbpkg.UpsertAdministradorConCreador(dbSuper, "delegado@empresa.com", "Delegado", "super_administrador", "", "principal@empresa.com"); err != nil {
		t.Fatalf("upsert delegado: %v", err)
	}
	if err := dbpkg.UpsertAdministradorConCreador(dbSuper, "externo@empresa.com", "Externo", "super_administrador", "", ""); err != nil {
		t.Fatalf("upsert externo: %v", err)
	}

	if _, err := dbEmp.Exec(`
		INSERT INTO empresas (id, empresa_id, nombre, usuario_creador, estado, fecha_creacion, fecha_actualizacion)
		VALUES
		(1, 1, 'Empresa Principal', 'principal@empresa.com', 'activo', datetime('now','localtime'), datetime('now','localtime')),
		(2, 2, 'Empresa Delegada Legacy', 'delegado@empresa.com', 'activo', datetime('now','localtime'), datetime('now','localtime')),
		(3, 3, 'Empresa Externa', 'externo@empresa.com', 'activo', datetime('now','localtime'), datetime('now','localtime'))
	`); err != nil {
		t.Fatalf("seed empresas: %v", err)
	}

	h := EmpresasHandler(dbEmp, dbSuper)
	listReq := httptest.NewRequest(http.MethodGet, "/super/api/empresas", nil)
	listReq = listReq.WithContext(context.WithValue(listReq.Context(), "adminEmail", "delegado@empresa.com"))
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)

	if listRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}

	var empresas []dbpkg.Empresa
	if err := json.Unmarshal(listRR.Body.Bytes(), &empresas); err != nil {
		t.Fatalf("decode empresas response: %v body=%s", err, listRR.Body.String())
	}
	if len(empresas) != 2 {
		t.Fatalf("expected 2 empresas in delegated scope, got %d: %+v", len(empresas), empresas)
	}
	for _, empresa := range empresas {
		if strings.EqualFold(empresa.Nombre, "Empresa Externa") {
			t.Fatalf("empresa externa must not be visible inside delegated scope: %+v", empresas)
		}
	}

	body := `{"nombre":"Empresa Nueva Compartida","usuario_creador":"delegado@empresa.com"}`
	postReq := httptest.NewRequest(http.MethodPost, "/super/api/empresas", strings.NewReader(body))
	postReq = postReq.WithContext(context.WithValue(postReq.Context(), "adminEmail", "delegado@empresa.com"))
	postReq.Header.Set("Content-Type", "application/json")
	postRR := httptest.NewRecorder()
	h.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, postRR.Code, postRR.Body.String())
	}

	var createdID int64
	if err := dbEmp.QueryRow(`SELECT id FROM empresas WHERE nombre = ?`, "Empresa Nueva Compartida").Scan(&createdID); err != nil {
		t.Fatalf("query created empresa id: %v", err)
	}
	var creator string
	if err := dbEmp.QueryRow(`SELECT COALESCE(usuario_creador, '') FROM empresas WHERE id = ?`, createdID).Scan(&creator); err != nil {
		t.Fatalf("query created empresa creator: %v", err)
	}
	if !strings.EqualFold(creator, "principal@empresa.com") {
		t.Fatalf("expected empresa creator principal@empresa.com, got %q", creator)
	}

	forbiddenReq := httptest.NewRequest(http.MethodGet, "/super/api/empresas?id=3", nil)
	forbiddenReq = forbiddenReq.WithContext(context.WithValue(forbiddenReq.Context(), "adminEmail", "delegado@empresa.com"))
	forbiddenRR := httptest.NewRecorder()
	h.ServeHTTP(forbiddenRR, forbiddenReq)

	if forbiddenRR.Code != http.StatusForbidden {
		t.Fatalf("expected status %d for out-of-scope empresa, got %d body=%s", http.StatusForbidden, forbiddenRR.Code, forbiddenRR.Body.String())
	}
}

func TestSuperConfigBackupHandlerExportYRestore(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_config_backup.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "wompi.public_key", "pub_test_demo", false); err != nil {
		t.Fatalf("seed wompi.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.mode", "sandbox", false); err != nil {
		t.Fatalf("seed wompi.mode: %v", err)
	}

	h := SuperConfigBackupHandler(dbSuper)
	exportReq := httptest.NewRequest(http.MethodGet, "/super/api/config/backup", nil)
	exportRR := httptest.NewRecorder()
	h.ServeHTTP(exportRR, exportReq)

	if exportRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on export, got %d body=%s", http.StatusOK, exportRR.Code, exportRR.Body.String())
	}

	var backup superConfigBackupPayload
	if err := json.Unmarshal(exportRR.Body.Bytes(), &backup); err != nil {
		t.Fatalf("decode backup export: %v body=%s", err, exportRR.Body.String())
	}
	if backup.Version != superConfigBackupVersion {
		t.Fatalf("expected backup version %q, got %q", superConfigBackupVersion, backup.Version)
	}

	updatedMode := false
	for i := range backup.Items {
		if strings.TrimSpace(backup.Items[i].Key) == "wompi.mode" {
			backup.Items[i].Value = "production"
			backup.Items[i].Configured = true
			updatedMode = true
			break
		}
	}
	if !updatedMode {
		t.Fatalf("expected wompi.mode key inside backup payload")
	}

	rawRestore, err := json.Marshal(backup)
	if err != nil {
		t.Fatalf("encode restore payload: %v", err)
	}

	restoreReq := httptest.NewRequest(http.MethodPut, "/super/api/config/backup", strings.NewReader(string(rawRestore)))
	restoreReq.Header.Set("Content-Type", "application/json")
	restoreRR := httptest.NewRecorder()
	h.ServeHTTP(restoreRR, restoreReq)

	if restoreRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on restore, got %d body=%s", http.StatusOK, restoreRR.Code, restoreRR.Body.String())
	}

	modeValue, _, err := dbpkg.GetConfigValue(dbSuper, "wompi.mode")
	if err != nil {
		t.Fatalf("read restored wompi.mode: %v", err)
	}
	if strings.TrimSpace(modeValue) != "production" {
		t.Fatalf("expected wompi.mode restored to production, got %q", modeValue)
	}
}

func TestSuperConfigBackupHandlerRestoreEncryptsSensitivePlaintext(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(i + 11)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_config_backup_encrypt_sensitive.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	payload := superConfigBackupPayload{
		Version:   superConfigBackupVersion,
		Scope:     "super_config_critica",
		CreatedAt: "2026-04-08T00:00:00Z",
		Items: []superConfigBackupItem{
			{
				Key:        "ai.model.deepseek.deepseek_chat.api_key",
				Value:      "deepseek_test_secret",
				Encrypted:  false,
				Configured: true,
			},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal restore payload: %v", err)
	}

	h := SuperConfigBackupHandler(dbSuper)
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/backup", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on restore, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "ai.model.deepseek.deepseek_chat.api_key")
	if err != nil {
		t.Fatalf("read restored sensitive key: %v", err)
	}
	if !encrypted {
		t.Fatal("expected restored sensitive key to be marked as encrypted")
	}
	if strings.TrimSpace(stored) == "deepseek_test_secret" {
		t.Fatal("expected stored sensitive key to be encrypted payload, got plaintext")
	}

	decrypted, decErr := utils.DecryptString(stored)
	if decErr != nil {
		t.Fatalf("decrypt restored sensitive key: %v", decErr)
	}
	if decrypted != "deepseek_test_secret" {
		t.Fatalf("expected decrypted sensitive value %q, got %q", "deepseek_test_secret", decrypted)
	}
}

func TestSuperConfigBackupHandlerRestoreSensitivePlaintextRequiresEncryptionKey(t *testing.T) {
	t.Setenv("CONFIG_ENC_KEY", "")

	dbSuper := openTestSQLite(t, "super_config_backup_encrypt_required.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	payload := superConfigBackupPayload{
		Version:   superConfigBackupVersion,
		Scope:     "super_config_critica",
		CreatedAt: "2026-04-08T00:00:00Z",
		Items: []superConfigBackupItem{
			{
				Key:        "gmail.smtp_app_password",
				Value:      "app-password-demo",
				Encrypted:  false,
				Configured: true,
			},
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal restore payload: %v", err)
	}

	h := SuperConfigBackupHandler(dbSuper)
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/backup", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d when CONFIG_ENC_KEY is missing, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestAIModelsConfigHandlerSaveGeminiEncrypted(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(77 + i)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_ai_config_handler.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := AIModelsConfigHandler(dbSuper)
	body := `{"credentials":[{"model_id":"google:gemini-2.0-flash","api_key":"gemini_test_key"}]}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/ai", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Email", "super@empresa.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on ai save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "ai.model.google.gemini_2_0_flash.api_key")
	if err != nil {
		t.Fatalf("read ai.model.google.gemini_2_0_flash.api_key: %v", err)
	}
	if !encrypted {
		t.Fatal("expected Gemini model key to be encrypted")
	}

	decrypted, decErr := utils.DecryptString(stored)
	if decErr != nil {
		t.Fatalf("decrypt Gemini model key: %v", decErr)
	}
	if decrypted != "gemini_test_key" {
		t.Fatalf("expected decrypted Gemini key %q, got %q", "gemini_test_key", decrypted)
	}

	providerStored, providerEncrypted, err := dbpkg.GetConfigValue(dbSuper, "ai.provider.google.api_key")
	if err != nil {
		t.Fatalf("read ai.provider.google.api_key: %v", err)
	}
	if !providerEncrypted {
		t.Fatal("expected provider Gemini key to be encrypted")
	}
	providerDecrypted, providerDecErr := utils.DecryptString(providerStored)
	if providerDecErr != nil {
		t.Fatalf("decrypt provider Gemini key: %v", providerDecErr)
	}
	if providerDecrypted != "gemini_test_key" {
		t.Fatalf("expected decrypted provider Gemini key %q, got %q", "gemini_test_key", providerDecrypted)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/ai", nil)
	getReq.Header.Set("X-Admin-Email", "super@empresa.com")
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on ai get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode ai get response: %v body=%s", err, getRR.Body.String())
	}
	modelos, ok := getBody["modelos"].([]interface{})
	if !ok || len(modelos) != 1 {
		t.Fatalf("expected modelos in ai get response, got %T", getBody["modelos"])
	}
	seenGemini := false
	for _, raw := range modelos {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		switch item["model_id"] {
		case "google:gemini-2.0-flash":
			seenGemini = true
			if masked, _ := item["masked"].(string); masked != "********" {
				t.Fatalf("expected masked value ******** for gemini, got %q", masked)
			}
		}
	}
	if !seenGemini {
		t.Fatal("expected gemini model in ai get response")
	}
}

func TestAIModelsConfigHandlerSavesProviderEnabledState(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_ai_provider_enabled.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := AIModelsConfigHandler(dbSuper)
	body := `{"provider_enabled":{"google":false}}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/ai", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Email", "super@empresa.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on ai provider save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	deepseekEnabled, _, err := dbpkg.GetConfigValue(dbSuper, "ai.provider.google.enabled")
	if err != nil {
		t.Fatalf("read ai.provider.google.enabled: %v", err)
	}
	if strings.TrimSpace(deepseekEnabled) != "0" {
		t.Fatalf("expected google enabled flag 0, got %q", deepseekEnabled)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/ai", nil)
	getReq.Header.Set("X-Admin-Email", "super@empresa.com")
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on ai get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode ai get response: %v body=%s", err, getRR.Body.String())
	}
	modelos, ok := getBody["modelos"].([]interface{})
	if !ok {
		t.Fatalf("expected modelos in ai get response, got %T", getBody["modelos"])
	}
	for _, raw := range modelos {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if item["model_id"] == "google:gemini-2.0-flash" {
			if enabled, _ := item["enabled"].(bool); enabled {
				t.Fatal("expected gemini to appear disabled in ai catalog")
			}
		}
	}
}

func TestAIModelsConfigHandlerTogglesGlobalServiceState(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_ai_toggle_handler.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := AIModelsConfigHandler(dbSuper)
	body := `{"enabled":false}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/ai", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Email", "super@empresa.com")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on ai toggle save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, superAIEnabledConfigKey)
	if err != nil {
		t.Fatalf("read %s: %v", superAIEnabledConfigKey, err)
	}
	if encrypted {
		t.Fatal("expected ai.global.enabled to be stored as plaintext flag")
	}
	if strings.TrimSpace(stored) != "0" {
		t.Fatalf("expected ai.global.enabled=0, got %q", stored)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/ai", nil)
	getReq.Header.Set("X-Admin-Email", "super@empresa.com")
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on ai get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode ai get response: %v body=%s", err, getRR.Body.String())
	}
	serviceStatus, ok := getBody["service_status"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected service_status object, got %#v", getBody["service_status"])
	}
	enabled, _ := serviceStatus["enabled"].(bool)
	if enabled {
		t.Fatal("expected service_status.enabled=false after toggle off")
	}
}

func TestGmailConfigHandlerSaveRestartAlertTo(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_gmail_restart_alert.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := GmailConfigHandler(dbSuper)
	body := `{"restart_alert_to":"ops@empresa.com"}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/gmail", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "gmail.restart_alert_to")
	if err != nil {
		t.Fatalf("read gmail.restart_alert_to: %v", err)
	}
	if strings.TrimSpace(stored) != "ops@empresa.com" {
		t.Fatalf("expected gmail.restart_alert_to %q, got %q", "ops@empresa.com", stored)
	}
	if encrypted {
		t.Fatal("expected gmail.restart_alert_to to be non-encrypted")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/gmail", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode gmail get response: %v body=%s", err, getRR.Body.String())
	}
	if got := strings.TrimSpace(fmt.Sprint(getBody["restart_alert_to"])); got != "ops@empresa.com" {
		t.Fatalf("expected restart_alert_to in response %q, got %q", "ops@empresa.com", got)
	}
	if setFlag, _ := getBody["restart_alert_to_set"].(bool); !setFlag {
		t.Fatalf("expected restart_alert_to_set=true, got %v", getBody["restart_alert_to_set"])
	}
}

func TestGmailConfigHandlerSaveWhatsAppContactNumber(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_whatsapp_contact_number.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := GmailConfigHandler(dbSuper)
	body := `{"whatsapp_contact_number":"+57 300 111 2233"}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/gmail", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on whatsapp save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "portal.whatsapp_contact_number")
	if err != nil {
		t.Fatalf("read portal.whatsapp_contact_number: %v", err)
	}
	if strings.TrimSpace(stored) != "573001112233" {
		t.Fatalf("expected portal.whatsapp_contact_number %q, got %q", "573001112233", stored)
	}
	if encrypted {
		t.Fatal("expected portal.whatsapp_contact_number to be non-encrypted")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/gmail", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode gmail get response: %v body=%s", err, getRR.Body.String())
	}
	if got := strings.TrimSpace(fmt.Sprint(getBody["whatsapp_contact_number"])); got != "573001112233" {
		t.Fatalf("expected whatsapp_contact_number in response %q, got %q", "573001112233", got)
	}
	if setFlag, _ := getBody["whatsapp_contact_number_set"].(bool); !setFlag {
		t.Fatalf("expected whatsapp_contact_number_set=true, got %v", getBody["whatsapp_contact_number_set"])
	}
}

func TestGmailConfigHandlerSaveRestartAlertToggle(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_gmail_restart_alert_toggle.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := GmailConfigHandler(dbSuper)
	body := `{"restart_alert_enabled":false}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/gmail", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail toggle save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "gmail.restart_alert_enabled")
	if err != nil {
		t.Fatalf("read gmail.restart_alert_enabled: %v", err)
	}
	if strings.TrimSpace(stored) != "0" {
		t.Fatalf("expected gmail.restart_alert_enabled %q, got %q", "0", stored)
	}
	if encrypted {
		t.Fatal("expected gmail.restart_alert_enabled to be non-encrypted")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/gmail", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var getBody map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &getBody); err != nil {
		t.Fatalf("decode gmail get response: %v body=%s", err, getRR.Body.String())
	}
	if enabled, _ := getBody["restart_alert_enabled"].(bool); enabled {
		t.Fatalf("expected restart_alert_enabled=false, got %v", getBody["restart_alert_enabled"])
	}
	if setFlag, _ := getBody["restart_alert_enabled_set"].(bool); !setFlag {
		t.Fatalf("expected restart_alert_enabled_set=true, got %v", getBody["restart_alert_enabled_set"])
	}
}

func TestGmailConfigHandlerTestActionCapturesNotification(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(41 + i)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_gmail_test_action.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_test_mode", "1", false); err != nil {
		t.Fatalf("seed gmail.smtp_test_mode: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_email", "mailer@powerfulcontrolsystem.com", false); err != nil {
		t.Fatalf("seed gmail.smtp_email: %v", err)
	}
	encPass, err := utils.EncryptString("app-pass-demo")
	if err != nil {
		t.Fatalf("encrypt gmail.smtp_app_password: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_app_password", encPass, true); err != nil {
		t.Fatalf("seed gmail.smtp_app_password: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_host", "smtp.gmail.com", false); err != nil {
		t.Fatalf("seed gmail.smtp_host: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_port", "587", false); err != nil {
		t.Fatalf("seed gmail.smtp_port: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_from_name", "Powerful Control System", false); err != nil {
		t.Fatalf("seed gmail.smtp_from_name: %v", err)
	}

	h := GmailConfigHandler(dbSuper)
	req := httptest.NewRequest(http.MethodPost, "/super/api/config/gmail?action=test", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on gmail test action, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode gmail test response: %v body=%s", err, rr.Body.String())
	}
	if sent, _ := body["sent"].(bool); !sent {
		t.Fatalf("expected sent=true, got %v", body["sent"])
	}
	if got := strings.TrimSpace(fmt.Sprint(body["recipient"])); got != superGmailTestRecipient {
		t.Fatalf("expected recipient %q, got %q", superGmailTestRecipient, got)
	}

	notifications, err := dbpkg.ListSuperCorreoNotificacionesPrueba(dbSuper, dbpkg.SuperCorreoNotificacionPruebaFilter{
		Tipo:         superCorreoNotificacionTipoPruebaGmail,
		Destinatario: superGmailTestRecipient,
		Limit:        10,
	})
	if err != nil {
		t.Fatalf("list super correo notificaciones: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("expected 1 captured notification, got %d", len(notifications))
	}
	if !strings.Contains(notifications[0].Cuerpo, "Prueba del boton Probar Gmail") && !strings.Contains(notifications[0].Cuerpo, "prueba del boton Probar Gmail") {
		t.Fatalf("expected captured body to mention gmail test button, got %q", notifications[0].Cuerpo)
	}
}

func TestSuperEmailTemplatesHandlerSaveAndGet(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_email_templates_handler.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := SuperEmailTemplatesHandler(dbSuper)
	body := `{"templates":[{"key":"empresa_user_confirmation","subject":"Bienvenido {{name}}","body_text":"Confirma aquí: {{confirm_url}}","body_html":"<p>Confirma aquí <a href=\"{{confirm_url}}\">ingresando</a></p>"},{"key":"licencia_activation_payment","subject":"Licencia activa para {{company_name}}","body_text":"Empresa: {{company_name}}","body_html":""}]}`
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/email_templates", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/email_templates", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var payload struct {
		Templates []superEmailTemplateItem `json:"templates"`
	}
	if err := json.Unmarshal(getRR.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode email templates response: %v body=%s", err, getRR.Body.String())
	}
	if len(payload.Templates) == 0 {
		t.Fatalf("expected templates in response, got %s", getRR.Body.String())
	}
	var foundEmpresa bool
	var foundLicencia bool
	for _, item := range payload.Templates {
		switch item.Key {
		case superEmailTemplateKeyEmpresaConfirmation:
			foundEmpresa = true
			if item.Subject != "Bienvenido {{name}}" {
				t.Fatalf("expected custom empresa subject, got %q", item.Subject)
			}
		case superEmailTemplateKeyLicenciaActivation:
			foundLicencia = true
			if item.Subject != "Licencia activa para {{company_name}}" {
				t.Fatalf("expected custom licencia subject, got %q", item.Subject)
			}
		}
	}
	if !foundEmpresa || !foundLicencia {
		t.Fatalf("expected saved templates in response, got %+v", payload.Templates)
	}
}

func TestApplySuperEmailTemplateUsesConfiguredValues(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_email_template_render.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, superEmailTemplateConfigKey(superEmailTemplateKeyLicenciaActivation, "subject"), "Licencia activa para {{company_name}}", false); err != nil {
		t.Fatalf("save custom subject: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, superEmailTemplateConfigKey(superEmailTemplateKeyLicenciaActivation, "body_text"), "Empresa {{company_name}} con ref {{reference}}", false); err != nil {
		t.Fatalf("save custom body_text: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, superEmailTemplateConfigKey(superEmailTemplateKeyLicenciaActivation, "body_html"), "", false); err != nil {
		t.Fatalf("save empty body_html: %v", err)
	}

	subject, bodyText, bodyHTML, err := applySuperEmailTemplate(dbSuper, superEmailTemplateKeyLicenciaActivation, map[string]string{
		"company_name": "Hotel Demo",
		"reference":    "REF-123",
	})
	if err != nil {
		t.Fatalf("apply template: %v", err)
	}
	if subject != "Licencia activa para Hotel Demo" {
		t.Fatalf("expected rendered subject, got %q", subject)
	}
	if bodyText != "Empresa Hotel Demo con ref REF-123" {
		t.Fatalf("expected rendered body_text, got %q", bodyText)
	}
	if !strings.Contains(bodyHTML, "Empresa Hotel Demo con ref REF-123") {
		t.Fatalf("expected generated html body to include rendered text, got %q", bodyHTML)
	}
}

func TestPublicLicenciasPaymentMethodsHandlerOrdersAndAvailability(t *testing.T) {
	rawKey := make([]byte, 32)
	for i := range rawKey {
		rawKey[i] = byte(31 + i)
	}
	t.Setenv("CONFIG_ENC_KEY", base64.StdEncoding.EncodeToString(rawKey))

	dbSuper := openTestSQLite(t, "super_payment_methods_status.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "epayco.public_key", "pub_test_demo", false); err != nil {
		t.Fatalf("seed epayco.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.customer_id", "1579238", false); err != nil {
		t.Fatalf("seed epayco.customer_id: %v", err)
	}
	encEpaycoKey, err := utils.EncryptString("epayco_secret_demo")
	if err != nil {
		t.Fatalf("encrypt epayco.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.private_key", encEpaycoKey, true); err != nil {
		t.Fatalf("seed epayco.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.enabled", "1", false); err != nil {
		t.Fatalf("seed epayco.enabled: %v", err)
	}

	if err := dbpkg.SetConfigValue(dbSuper, "wompi.public_key", "pub_test_demo", false); err != nil {
		t.Fatalf("seed wompi.public_key: %v", err)
	}
	encWompiPrivate, err := utils.EncryptString("prv_test_demo")
	if err != nil {
		t.Fatalf("encrypt wompi.private_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.private_key", encWompiPrivate, true); err != nil {
		t.Fatalf("seed wompi.private_key: %v", err)
	}
	encWompiIntegrity, err := utils.EncryptString("test_integrity_demo")
	if err != nil {
		t.Fatalf("encrypt wompi.integrity_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.integrity_key", encWompiIntegrity, true); err != nil {
		t.Fatalf("seed wompi.integrity_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "wompi.enabled", "0", false); err != nil {
		t.Fatalf("seed wompi.enabled: %v", err)
	}

	h := PublicLicenciasPaymentMethodsHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/api/public/licencias/payment_methods", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body struct {
		Providers []struct {
			ID         string `json:"id"`
			Enabled    bool   `json:"enabled"`
			Configured bool   `json:"configured"`
			Available  bool   `json:"available"`
		} `json:"providers"`
		DefaultMethod string `json:"default_method"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode payment methods response: %v body=%s", err, rr.Body.String())
	}
	if len(body.Providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(body.Providers))
	}
	if body.Providers[0].ID != "epayco" || body.Providers[1].ID != "wompi" {
		t.Fatalf("expected order epayco,wompi got %+v", body.Providers)
	}
	if !body.Providers[0].Enabled || !body.Providers[0].Configured || !body.Providers[0].Available {
		t.Fatalf("expected epayco available, got %+v", body.Providers[0])
	}
	if !body.Providers[1].Configured || body.Providers[1].Available || body.Providers[1].Enabled {
		t.Fatalf("expected wompi configured but disabled, got %+v", body.Providers[1])
	}
	if body.DefaultMethod != "epayco" {
		t.Fatalf("expected default_method epayco, got %q", body.DefaultMethod)
	}
}

func TestPublicLicenciasPaymentMethodsHandlerRequiresPrivateKeyForEpaycoAvailability(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_payment_methods_epayco_public_only.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "epayco.public_key", "pub_test_only_public", false); err != nil {
		t.Fatalf("seed epayco.public_key: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "epayco.enabled", "1", false); err != nil {
		t.Fatalf("seed epayco.enabled: %v", err)
	}

	h := PublicLicenciasPaymentMethodsHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/api/public/licencias/payment_methods", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body struct {
		Providers []struct {
			ID         string `json:"id"`
			Configured bool   `json:"configured"`
			Available  bool   `json:"available"`
			Enabled    bool   `json:"enabled"`
		} `json:"providers"`
		DefaultMethod string `json:"default_method"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode payment methods response: %v body=%s", err, rr.Body.String())
	}
	if len(body.Providers) == 0 {
		t.Fatalf("expected providers in response, got %s", rr.Body.String())
	}
	if body.Providers[0].ID != "epayco" || body.Providers[0].Configured || body.Providers[0].Available || !body.Providers[0].Enabled {
		t.Fatalf("expected epayco enabled but unavailable without private key, got %+v", body.Providers[0])
	}
	if body.DefaultMethod != "" {
		t.Fatalf("expected no default_method when epayco is incomplete, got %q", body.DefaultMethod)
	}
}

func TestWompiConfigHandlerPersistsEnabledFlag(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_wompi_enabled_toggle.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	h := WompiConfigHandler(dbSuper)
	req := httptest.NewRequest(http.MethodPut, "/super/api/config/wompi", strings.NewReader(`{"enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d on wompi enabled save, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	stored, encrypted, err := dbpkg.GetConfigValue(dbSuper, "wompi.enabled")
	if err != nil {
		t.Fatalf("read wompi.enabled: %v", err)
	}
	if encrypted {
		t.Fatal("expected wompi.enabled to be non-encrypted")
	}
	if strings.TrimSpace(stored) != "1" {
		t.Fatalf("expected wompi.enabled stored as 1, got %q", stored)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/super/api/config/wompi", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on wompi get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(getRR.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode wompi get response: %v body=%s", err, getRR.Body.String())
	}
	if enabled, _ := body["enabled"].(bool); !enabled {
		t.Fatalf("expected enabled=true in wompi get response, got %v", body["enabled"])
	}
}

func TestWompiTermsHandlerRejectsWhenDisabled(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_wompi_terms_disabled.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "wompi.enabled", "0", false); err != nil {
		t.Fatalf("seed wompi.enabled: %v", err)
	}

	h := WompiTermsHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/wompi/terms", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusPreconditionFailed, rr.Code, rr.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode wompi disabled response: %v body=%s", err, rr.Body.String())
	}
	if got := strings.TrimSpace(fmt.Sprint(body["provider"])); got != "wompi" {
		t.Fatalf("expected provider wompi, got %q", got)
	}
}

func TestSuperEndpointsPermisosPorRol(t *testing.T) {
	dbEmp := openTestPostgres(t, "DB_EMPRESAS_DSN", "empresas_super_roles")
	dbSuper := openTestPostgres(t, "DB_SUPERADMIN_DSN", "super_super_roles")
	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureSuperSchema(t, dbSuper)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	mux := http.NewServeMux()
	mux.HandleFunc("/super/api/empresas", EmpresasHandler(dbEmp, dbSuper))
	mux.HandleFunc("/super/api/tipos_empresas", TiposEmpresasHandler(dbSuper))
	mux.HandleFunc("/super/api/licencias", LicenciasHandler(dbSuper))
	mux.HandleFunc("/super/api/config/wompi", WompiConfigHandler(dbSuper))
	mux.HandleFunc("/super/api/config/gmail", GmailConfigHandler(dbSuper))
	mux.HandleFunc("/super/api/config/ai", AIModelsConfigHandler(dbSuper))
	mux.HandleFunc("/super/api/config/backup", SuperConfigBackupHandler(dbSuper))
	mux.HandleFunc("/super/api/soporte_remoto", SuperSoporteRemotoHandler(dbEmp))
	mux.HandleFunc("/super/licencias.html", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	protected := utils.AuthMiddleware(dbSuper, mux)

	if err := dbpkg.UpsertAdministrador(dbSuper, "powerfulcontrolsystem@gmail.com", "Super", "super_administrador", ""); err != nil {
		t.Fatalf("upsert super admin: %v", err)
	}
	if err := dbpkg.CreateSession(dbSuper, "powerfulcontrolsystem@gmail.com", "127.0.0.1", "test-agent", "token-super"); err != nil {
		t.Fatalf("create super session: %v", err)
	}

	rolesBloqueados := []string{"administrador", "contabilidad", "cajero", "supervisor_sucursal", "auditor"}
	for i, role := range rolesBloqueados {
		email := fmt.Sprintf("role-%d@empresa.com", i+1)
		token := fmt.Sprintf("token-%d", i+1)
		if err := dbpkg.UpsertAdministrador(dbSuper, email, "RoleUser", role, ""); err != nil {
			t.Fatalf("upsert role admin %s: %v", role, err)
		}
		if err := dbpkg.CreateSession(dbSuper, email, "127.0.0.1", "test-agent", token); err != nil {
			t.Fatalf("create role session %s: %v", role, err)
		}
	}

	superOnlyEndpoints := []string{
		"/super/api/config/wompi",
		"/super/api/config/gmail",
		"/super/api/config/ai",
		"/super/api/config/backup",
		"/super/api/soporte_remoto",
	}

	for _, endpoint := range superOnlyEndpoints {
		reqNoAuth := httptest.NewRequest(http.MethodGet, endpoint, nil)
		rrNoAuth := httptest.NewRecorder()
		protected.ServeHTTP(rrNoAuth, reqNoAuth)
		if rrNoAuth.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 without token for %s, got %d body=%s", endpoint, rrNoAuth.Code, rrNoAuth.Body.String())
		}

		reqSuper := httptest.NewRequest(http.MethodGet, endpoint, nil)
		reqSuper.AddCookie(&http.Cookie{Name: "session_token", Value: "token-super"})
		rrSuper := httptest.NewRecorder()
		protected.ServeHTTP(rrSuper, reqSuper)
		if rrSuper.Code != http.StatusOK {
			t.Fatalf("expected 200 for super_admin in %s, got %d body=%s", endpoint, rrSuper.Code, rrSuper.Body.String())
		}

		for i := range rolesBloqueados {
			token := fmt.Sprintf("token-%d", i+1)
			reqRole := httptest.NewRequest(http.MethodGet, endpoint, nil)
			reqRole.AddCookie(&http.Cookie{Name: "session_token", Value: token})
			rrRole := httptest.NewRecorder()
			protected.ServeHTTP(rrRole, reqRole)
			if rrRole.Code != http.StatusForbidden {
				t.Fatalf("expected 403 for role=%s in %s, got %d body=%s", rolesBloqueados[i], endpoint, rrRole.Code, rrRole.Body.String())
			}
		}
	}

	adminReadableEndpoints := []string{
		"/super/licencias.html",
		"/super/api/empresas",
		"/super/api/tipos_empresas",
		"/super/api/licencias",
	}

	for _, endpoint := range adminReadableEndpoints {
		reqAdmin := httptest.NewRequest(http.MethodGet, endpoint, nil)
		reqAdmin.AddCookie(&http.Cookie{Name: "session_token", Value: "token-1"})
		rrAdmin := httptest.NewRecorder()
		protected.ServeHTTP(rrAdmin, reqAdmin)
		if rrAdmin.Code != http.StatusOK {
			t.Fatalf("expected 200 for administrador in %s, got %d body=%s", endpoint, rrAdmin.Code, rrAdmin.Body.String())
		}

		for _, blockedToken := range []string{"token-2", "token-3", "token-4", "token-5"} {
			reqRole := httptest.NewRequest(http.MethodGet, endpoint, nil)
			reqRole.AddCookie(&http.Cookie{Name: "session_token", Value: blockedToken})
			rrRole := httptest.NewRecorder()
			protected.ServeHTTP(rrRole, reqRole)
			if rrRole.Code != http.StatusForbidden {
				t.Fatalf("expected 403 for non-admin reader token=%s in %s, got %d body=%s", blockedToken, endpoint, rrRole.Code, rrRole.Body.String())
			}
		}
	}
}

func TestAdministradorPuedeEditarYEliminarEmpresaDesdeRutaSuperProtegida(t *testing.T) {
	dbEmp := openTestPostgres(t, "DB_EMPRESAS_DSN", "empresas_super_admin_manage")
	dbSuper := openTestPostgres(t, "DB_SUPERADMIN_DSN", "super_super_admin_manage")
	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureSuperSchema(t, dbSuper)
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	clientesStmt := `CREATE TABLE IF NOT EXISTS clientes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		nombre TEXT
	)`
	if dbpkg.IsPostgresDialect() {
		clientesStmt = `CREATE TABLE IF NOT EXISTS clientes (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT,
			nombre TEXT
		)`
	}
	if _, err := dbEmp.Exec(clientesStmt); err != nil {
		t.Fatalf("create clientes schema: %v", err)
	}

	if err := dbpkg.UpsertAdministrador(dbSuper, "admin_scope@empresa.com", "Admin Scope", "administrador", ""); err != nil {
		t.Fatalf("upsert admin: %v", err)
	}
	if err := dbpkg.CreateSession(dbSuper, "admin_scope@empresa.com", "127.0.0.1", "test-agent", "token-admin-scope"); err != nil {
		t.Fatalf("create session admin: %v", err)
	}

	nowValue := time.Now().Format("2006-01-02 15:04:05")
	if _, err := dbpkg.ExecCompat(dbEmp, `
		INSERT INTO empresas (id, empresa_id, nombre, nit, tipo_id, tipo_nombre, usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 41, 41, "Empresa Editable", "90041", 2, "Hotel", "admin_scope@empresa.com", "activo", "descripcion original", nowValue, nowValue); err != nil {
		t.Fatalf("insert empresa editable: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO clientes (empresa_id, nombre) VALUES (41, 'Cliente Interno')`); err != nil {
		t.Fatalf("insert cliente editable: %v", err)
	}
	if _, err := dbSuper.Exec(`INSERT INTO licencias (empresa_id, activo) VALUES (41, 1)`); err != nil {
		t.Fatalf("insert licencia editable: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/super/api/empresas", EmpresasHandler(dbEmp, dbSuper))
	protected := utils.AuthMiddleware(dbSuper, mux)

	updateReq := httptest.NewRequest(http.MethodPut, "/super/api/empresas?id=41", strings.NewReader(`{"tipo_id":2,"tipo_nombre":"Hotel","nombre":"Empresa Editada","nit":"90041","observaciones":"descripcion actualizada"}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(&http.Cookie{Name: "session_token", Value: "token-admin-scope"})
	updateRR := httptest.NewRecorder()
	protected.ServeHTTP(updateRR, updateReq)

	if updateRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d on update, got %d body=%s", http.StatusNoContent, updateRR.Code, updateRR.Body.String())
	}

	var nombre, observaciones string
	if err := dbEmp.QueryRow(`SELECT nombre, COALESCE(observaciones, '') FROM empresas WHERE id = 41`).Scan(&nombre, &observaciones); err != nil {
		t.Fatalf("query updated empresa: %v", err)
	}
	if nombre != "Empresa Editada" || observaciones != "descripcion actualizada" {
		t.Fatalf("unexpected update result nombre=%q observaciones=%q", nombre, observaciones)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/super/api/empresas?id=41&action=eliminar_total", strings.NewReader(`{"confirmacion_nombre":"Empresa Editada"}`))
	deleteReq.Header.Set("Content-Type", "application/json")
	deleteReq.AddCookie(&http.Cookie{Name: "session_token", Value: "token-admin-scope"})
	deleteRR := httptest.NewRecorder()
	protected.ServeHTTP(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on delete, got %d body=%s", http.StatusOK, deleteRR.Code, deleteRR.Body.String())
	}

	var total int
	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM empresas WHERE id = 41`).Scan(&total); err != nil {
		t.Fatalf("count empresa after delete: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected empresa deleted, got %d rows", total)
	}
	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM clientes WHERE empresa_id = 41`).Scan(&total); err != nil {
		t.Fatalf("count clientes after delete: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected clientes deleted, got %d rows", total)
	}
	if err := dbSuper.QueryRow(`SELECT COUNT(1) FROM licencias WHERE empresa_id = 41`).Scan(&total); err != nil {
		t.Fatalf("count licencias after delete: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected licencias deleted, got %d rows", total)
	}
}

func TestNuevoAdminRegistradoPuedeCrearSuPrimeraEmpresaViaRutaSuperProtegida(t *testing.T) {
	dbEmp := openTestPostgres(t, "DB_EMPRESAS_DSN", "empresas_nuevo_admin_create")
	dbSuper := openTestPostgres(t, "DB_SUPERADMIN_DSN", "super_nuevo_admin_create")
	ensureEmpresasCoreSchemaForSuper(t, dbEmp)
	ensureAdminAuthTestSchema(t, dbSuper)

	registerBody := `{"email":"nuevo_flujo@empresa.com","name":"Nuevo Flujo","telefono":"3001234567","pais":"Colombia","ciudad":"Bogota","password":"ClaveSegura99"}`
	registerReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/super/api/administradores/register", strings.NewReader(registerBody))
	registerRR := httptest.NewRecorder()
	AdminRegisterHandler(dbSuper).ServeHTTP(registerRR, registerReq)

	if registerRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on register, got %d body=%s", http.StatusOK, registerRR.Code, registerRR.Body.String())
	}

	if _, err := dbpkg.ExecCompat(dbSuper, `UPDATE administradores SET email_confirmado = 1 WHERE lower(email) = lower(?)`, "nuevo_flujo@empresa.com"); err != nil {
		t.Fatalf("confirm new admin: %v", err)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "http://localhost:8080/super/api/administradores/login", strings.NewReader(`{"email":"nuevo_flujo@empresa.com","password":"ClaveSegura99"}`))
	loginRR := httptest.NewRecorder()
	AdminLoginHandler(dbSuper).ServeHTTP(loginRR, loginReq)

	if loginRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on login, got %d body=%s", http.StatusOK, loginRR.Code, loginRR.Body.String())
	}

	var sessionCookie *http.Cookie
	for _, cookie := range loginRR.Result().Cookies() {
		if cookie.Name == "session_token" && strings.TrimSpace(cookie.Value) != "" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session_token cookie for new admin login")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/super/api/empresas", EmpresasHandler(dbEmp, dbSuper))
	protected := utils.AuthMiddleware(dbSuper, mux)

	createBody := `{"tipo_id":7,"tipo_nombre":"Hotel","nombre":"Hotel Nuevo Flujo","nit":"901234567","observaciones":"Primera empresa del usuario nuevo"}`
	createReq := httptest.NewRequest(http.MethodPost, "/super/api/empresas", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createRR := httptest.NewRecorder()
	protected.ServeHTTP(createRR, createReq)

	if createRR.Code != http.StatusOK {
		t.Fatalf("expected status %d on protected super create empresa, got %d body=%s", http.StatusOK, createRR.Code, createRR.Body.String())
	}

	var totalEmpresas int
	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM empresas`).Scan(&totalEmpresas); err != nil {
		t.Fatalf("count empresas: %v", err)
	}
	if totalEmpresas != 1 {
		t.Fatalf("expected 1 empresa created for generic admin, got %d", totalEmpresas)
	}

	var empresaNombre, empresaCreador string
	if err := dbEmp.QueryRow(`SELECT COALESCE(nombre, ''), COALESCE(usuario_creador, '') FROM empresas LIMIT 1`).Scan(&empresaNombre, &empresaCreador); err != nil {
		t.Fatalf("query created empresa: %v", err)
	}
	if !strings.EqualFold(strings.TrimSpace(empresaNombre), "Hotel Nuevo Flujo") {
		t.Fatalf("expected created empresa name Hotel Nuevo Flujo, got %q", empresaNombre)
	}
	if !strings.EqualFold(strings.TrimSpace(empresaCreador), "nuevo_flujo@empresa.com") {
		t.Fatalf("expected usuario_creador nuevo_flujo@empresa.com, got %q", empresaCreador)
	}

	admin, err := dbpkg.GetAdminByEmailFull(dbSuper, "nuevo_flujo@empresa.com")
	if err != nil {
		t.Fatalf("reload admin after create empresa: %v", err)
	}
	if !strings.EqualFold(admin.Role, "administrador") {
		t.Fatalf("expected new admin role administrador, got %q", admin.Role)
	}
}

func TestNuevoAdminRegistradoNoObtieneAccesoSuperParaCrearEmpresa(t *testing.T) {
	TestNuevoAdminRegistradoPuedeCrearSuPrimeraEmpresaViaRutaSuperProtegida(t)
}
