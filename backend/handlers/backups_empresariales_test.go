package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func ensureEmpresaBackupsHandlerTestTable(t *testing.T, dbConn *sql.DB) {
	t.Helper()
	_, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS empresa_backup_items_test (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		codigo TEXT,
		valor INTEGER DEFAULT 0,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`)
	if err != nil {
		t.Fatalf("create empresa_backup_items_test: %v", err)
	}
}

func TestEmpresaBackupsHandlerCreateListExportRestoreYToggle(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresa_backups_handler.db")
	if err := dbpkg.EnsureEmpresaBackupsSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaBackupsSchema: %v", err)
	}
	ensureEmpresaBackupsHandlerTestTable(t, dbEmp)

	if _, err := dbEmp.Exec(`INSERT INTO empresa_backup_items_test (empresa_id, codigo, valor, usuario_creador, estado) VALUES
		(71, 'X-1', 15, 'seed', 'activo'),
		(71, 'X-2', 25, 'seed', 'activo'),
		(98, 'Y-1', 88, 'seed', 'activo')`); err != nil {
		t.Fatalf("seed table: %v", err)
	}

	h := EmpresaBackupsHandler(dbEmp)

	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/backups?empresa_id=71&action=crear", strings.NewReader(`{
		"empresa_id":71,
		"nombre":"Backup handler 71",
		"include_tables":["empresa_backup_items_test"]
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRR.Code, createRR.Body.String())
	}

	var createResp struct {
		Backup dbpkg.EmpresaBackup `json:"backup"`
	}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Backup.ID <= 0 {
		t.Fatalf("expected backup id > 0, got %+v", createResp.Backup)
	}
	backupID := createResp.Backup.ID

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/backups?empresa_id=71", nil)
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRR.Code, listRR.Body.String())
	}
	var listResp struct {
		Total int64                 `json:"total"`
		Rows  []dbpkg.EmpresaBackup `json:"rows"`
	}
	if err := json.Unmarshal(listRR.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Rows) != 1 {
		t.Fatalf("expected one backup in list, total=%d len=%d", listResp.Total, len(listResp.Rows))
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/empresa/backups?empresa_id=71&action=detalle&id="+itoa64(backupID)+"&include_snapshot=1", nil)
	detailRR := httptest.NewRecorder()
	h.ServeHTTP(detailRR, detailReq)
	if detailRR.Code != http.StatusOK {
		t.Fatalf("detail status=%d body=%s", detailRR.Code, detailRR.Body.String())
	}
	var detailResp struct {
		Payload dbpkg.EmpresaBackupPayload `json:"payload"`
	}
	if err := json.Unmarshal(detailRR.Body.Bytes(), &detailResp); err != nil {
		t.Fatalf("decode detail response: %v", err)
	}
	if detailResp.Payload.TotalTables != 1 || detailResp.Payload.TotalRows != 2 {
		t.Fatalf("unexpected payload totals: %+v", detailResp.Payload)
	}

	exportReq := httptest.NewRequest(http.MethodGet, "/api/empresa/backups?empresa_id=71&action=export&id="+itoa64(backupID)+"&format=csv", nil)
	exportRR := httptest.NewRecorder()
	h.ServeHTTP(exportRR, exportReq)
	if exportRR.Code != http.StatusOK {
		t.Fatalf("export status=%d body=%s", exportRR.Code, exportRR.Body.String())
	}
	if !strings.Contains(strings.ToLower(exportRR.Header().Get("Content-Type")), "text/csv") {
		t.Fatalf("expected csv content-type, got=%s", exportRR.Header().Get("Content-Type"))
	}

	if _, err := dbEmp.Exec(`UPDATE empresa_backup_items_test SET valor = 900 WHERE empresa_id = 71 AND codigo = 'X-1'`); err != nil {
		t.Fatalf("mutate before restore: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO empresa_backup_items_test (empresa_id, codigo, valor, usuario_creador, estado) VALUES (71, 'X-3', 33, 'seed', 'activo')`); err != nil {
		t.Fatalf("insert extra before restore: %v", err)
	}

	restoreReq := httptest.NewRequest(http.MethodPost, "/api/empresa/backups?empresa_id=71&action=restaurar", strings.NewReader(`{"empresa_id":71,"backup_id":`+itoa64(backupID)+`}`))
	restoreReq.Header.Set("Content-Type", "application/json")
	restoreRR := httptest.NewRecorder()
	h.ServeHTTP(restoreRR, restoreReq)
	if restoreRR.Code != http.StatusOK {
		t.Fatalf("restore status=%d body=%s", restoreRR.Code, restoreRR.Body.String())
	}

	rows, err := dbEmp.Query(`SELECT codigo, valor FROM empresa_backup_items_test WHERE empresa_id = 71 ORDER BY codigo`)
	if err != nil {
		t.Fatalf("query restored rows: %v", err)
	}
	defer rows.Close()

	type pair struct {
		Codigo string
		Valor  int64
	}
	items := make([]pair, 0)
	for rows.Next() {
		var p pair
		if err := rows.Scan(&p.Codigo, &p.Valor); err != nil {
			t.Fatalf("scan restored row: %v", err)
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate restored rows: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 restored rows, got %d", len(items))
	}
	if items[0].Codigo != "X-1" || items[0].Valor != 15 {
		t.Fatalf("unexpected first restored row: %+v", items[0])
	}
	if items[1].Codigo != "X-2" || items[1].Valor != 25 {
		t.Fatalf("unexpected second restored row: %+v", items[1])
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/backups?empresa_id=71&id="+itoa64(backupID), nil)
	deleteRR := httptest.NewRecorder()
	h.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleteRR.Code, deleteRR.Body.String())
	}

	listAfterDeleteReq := httptest.NewRequest(http.MethodGet, "/api/empresa/backups?empresa_id=71", nil)
	listAfterDeleteRR := httptest.NewRecorder()
	h.ServeHTTP(listAfterDeleteRR, listAfterDeleteReq)
	if listAfterDeleteRR.Code != http.StatusOK {
		t.Fatalf("list after delete status=%d body=%s", listAfterDeleteRR.Code, listAfterDeleteRR.Body.String())
	}
	var listAfterDelete struct {
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(listAfterDeleteRR.Body.Bytes(), &listAfterDelete); err != nil {
		t.Fatalf("decode list after delete response: %v", err)
	}
	if listAfterDelete.Total != 0 {
		t.Fatalf("expected active backups total=0 after delete, got %d", listAfterDelete.Total)
	}
}

func TestEmpresaBackupsHandlerRestoreNotFound(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresa_backups_handler_not_found.db")
	if err := dbpkg.EnsureEmpresaBackupsSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaBackupsSchema: %v", err)
	}

	h := EmpresaBackupsHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/backups?empresa_id=17&action=restaurar", strings.NewReader(`{"empresa_id":17,"backup_id":9999}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status=%d, got=%d body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
}

func TestEmpresaBackupsHandlerPurgeByDate(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresa_backups_handler_purge_date.db")
	if err := dbpkg.EnsureEmpresaBackupsSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaBackupsSchema: %v", err)
	}
	ensureEmpresaBackupsHandlerTestTable(t, dbEmp)

	if _, err := dbEmp.Exec(`INSERT INTO empresa_backup_items_test (
		empresa_id, codigo, valor, fecha_creacion, fecha_actualizacion, usuario_creador, estado
	) VALUES
		(77, 'OLD', 1, '2026-01-01 10:00:00', '2026-01-01 10:00:00', 'seed', 'activo'),
		(77, 'NEW', 2, '2026-04-08 10:00:00', '2026-04-08 10:00:00', 'seed', 'activo'),
		(78, 'EXT', 3, '2026-01-01 10:00:00', '2026-01-01 10:00:00', 'seed', 'activo')`); err != nil {
		t.Fatalf("seed rows for purge handler: %v", err)
	}

	h := EmpresaBackupsHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/backups?empresa_id=77&action=depurar_fecha", strings.NewReader(`{
		"empresa_id":77,
		"fecha_corte":"2026-03-01",
		"include_tables":["empresa_backup_items_test"],
		"crear_backup_previo":false
	}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("purge by date status=%d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Resultado struct {
			RegistrosEliminados int64 `json:"registros_eliminados"`
		} `json:"resultado"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode purge response: %v", err)
	}
	if resp.Resultado.RegistrosEliminados != 1 {
		t.Fatalf("expected registros_eliminados=1, got %d", resp.Resultado.RegistrosEliminados)
	}

	var empresa77Count int64
	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM empresa_backup_items_test WHERE empresa_id = 77`).Scan(&empresa77Count); err != nil {
		t.Fatalf("count empresa 77 after purge: %v", err)
	}
	if empresa77Count != 1 {
		t.Fatalf("expected one remaining row for empresa 77, got %d", empresa77Count)
	}

	var empresa78Count int64
	if err := dbEmp.QueryRow(`SELECT COUNT(1) FROM empresa_backup_items_test WHERE empresa_id = 78`).Scan(&empresa78Count); err != nil {
		t.Fatalf("count empresa 78 after purge: %v", err)
	}
	if empresa78Count != 1 {
		t.Fatalf("expected empresa 78 untouched, got %d", empresa78Count)
	}
}
