package db

import (
	"database/sql"
	"testing"
)

func ensureEmpresaBackupItemsTestTable(t *testing.T, dbConn *sql.DB) {
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

func TestEmpresaBackupsSnapshotYRestoreFlow(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaBackupsSchema(dbConn); err != nil {
		t.Fatalf("EnsureEmpresaBackupsSchema: %v", err)
	}
	ensureEmpresaBackupItemsTestTable(t, dbConn)

	if _, err := dbConn.Exec(`INSERT INTO empresa_backup_items_test (empresa_id, codigo, valor, usuario_creador, estado) VALUES
		(501, 'A-1', 10, 'qa_seed', 'activo'),
		(501, 'A-2', 20, 'qa_seed', 'activo'),
		(777, 'B-1', 99, 'qa_seed', 'activo')`); err != nil {
		t.Fatalf("seed empresa_backup_items_test: %v", err)
	}

	backupID, err := CreateEmpresaBackupSnapshot(dbConn, 501, "Backup QA 501", "snapshot inicial", "qa_user", EmpresaBackupBuildOptions{
		IncludeTables: []string{"empresa_backup_items_test"},
	})
	if err != nil {
		t.Fatalf("CreateEmpresaBackupSnapshot: %v", err)
	}
	if backupID <= 0 {
		t.Fatalf("expected backup id > 0, got %d", backupID)
	}

	if _, err := dbConn.Exec(`UPDATE empresa_backup_items_test SET valor = 999 WHERE empresa_id = 501 AND codigo = 'A-1'`); err != nil {
		t.Fatalf("update rows before restore: %v", err)
	}
	if _, err := dbConn.Exec(`INSERT INTO empresa_backup_items_test (empresa_id, codigo, valor, usuario_creador, estado) VALUES (501, 'A-3', 30, 'qa_user', 'activo')`); err != nil {
		t.Fatalf("insert extra row before restore: %v", err)
	}

	result, err := RestoreEmpresaBackupByID(dbConn, 501, backupID, "qa_restore", "restore de prueba")
	if err != nil {
		t.Fatalf("RestoreEmpresaBackupByID: %v", err)
	}
	if result.TablasRestauradas != 1 {
		t.Fatalf("expected tablas_restauradas=1, got %d", result.TablasRestauradas)
	}
	if result.RegistrosRestaurados != 2 {
		t.Fatalf("expected registros_restaurados=2, got %d", result.RegistrosRestaurados)
	}

	rows, err := dbConn.Query(`SELECT codigo, valor FROM empresa_backup_items_test WHERE empresa_id = 501 ORDER BY codigo`)
	if err != nil {
		t.Fatalf("query restored rows empresa 501: %v", err)
	}
	defer rows.Close()

	type pair struct {
		codigo string
		valor  int64
	}
	got := make([]pair, 0)
	for rows.Next() {
		var item pair
		if err := rows.Scan(&item.codigo, &item.valor); err != nil {
			t.Fatalf("scan restored row: %v", err)
		}
		got = append(got, item)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate restored rows: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 restored rows, got %d", len(got))
	}
	if got[0].codigo != "A-1" || got[0].valor != 10 {
		t.Fatalf("unexpected first row after restore: %+v", got[0])
	}
	if got[1].codigo != "A-2" || got[1].valor != 20 {
		t.Fatalf("unexpected second row after restore: %+v", got[1])
	}

	var empresa777Count int64
	if err := dbConn.QueryRow(`SELECT COUNT(1) FROM empresa_backup_items_test WHERE empresa_id = 777 AND codigo = 'B-1'`).Scan(&empresa777Count); err != nil {
		t.Fatalf("count empresa 777 rows: %v", err)
	}
	if empresa777Count != 1 {
		t.Fatalf("expected empresa 777 rows untouched, got %d", empresa777Count)
	}

	backupRow, err := GetEmpresaBackupByID(dbConn, 501, backupID, false)
	if err != nil {
		t.Fatalf("GetEmpresaBackupByID: %v", err)
	}
	if backupRow.RestauradoPor != "qa_restore" {
		t.Fatalf("expected restaurado_por=qa_restore, got %q", backupRow.RestauradoPor)
	}

	var restoreHistoryCount int64
	if err := dbConn.QueryRow(`SELECT COUNT(1) FROM empresa_backups_restauraciones WHERE empresa_id = 501 AND backup_id = ?`, backupID).Scan(&restoreHistoryCount); err != nil {
		t.Fatalf("count restore history: %v", err)
	}
	if restoreHistoryCount != 1 {
		t.Fatalf("expected one restore history row, got %d", restoreHistoryCount)
	}
}

func TestEmpresaBackupsListYPayload(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaBackupsSchema(dbConn); err != nil {
		t.Fatalf("EnsureEmpresaBackupsSchema: %v", err)
	}
	ensureEmpresaBackupItemsTestTable(t, dbConn)

	if _, err := dbConn.Exec(`INSERT INTO empresa_backup_items_test (empresa_id, codigo, valor, usuario_creador, estado) VALUES
		(620, 'L-1', 5, 'qa_seed', 'activo'),
		(620, 'L-2', 7, 'qa_seed', 'activo')`); err != nil {
		t.Fatalf("seed empresa 620 rows: %v", err)
	}

	backupOne, err := CreateEmpresaBackupSnapshot(dbConn, 620, "Backup Uno", "base", "qa_user", EmpresaBackupBuildOptions{IncludeTables: []string{"empresa_backup_items_test"}})
	if err != nil {
		t.Fatalf("CreateEmpresaBackupSnapshot one: %v", err)
	}
	backupTwo, err := CreateEmpresaBackupSnapshot(dbConn, 620, "Backup Dos", "operativo", "qa_user", EmpresaBackupBuildOptions{IncludeTables: []string{"empresa_backup_items_test"}})
	if err != nil {
		t.Fatalf("CreateEmpresaBackupSnapshot two: %v", err)
	}

	if err := SetEmpresaBackupEstadoByID(dbConn, 620, backupOne, "inactivo"); err != nil {
		t.Fatalf("SetEmpresaBackupEstadoByID: %v", err)
	}

	activeRows, totalActive, err := ListEmpresaBackups(dbConn, 620, EmpresaBackupFilter{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListEmpresaBackups active: %v", err)
	}
	if totalActive != 1 || len(activeRows) != 1 {
		t.Fatalf("expected one active backup, total=%d len=%d", totalActive, len(activeRows))
	}
	if activeRows[0].ID != backupTwo {
		t.Fatalf("expected active backup id=%d, got %d", backupTwo, activeRows[0].ID)
	}

	allRows, totalAll, err := ListEmpresaBackups(dbConn, 620, EmpresaBackupFilter{IncludeInactive: true, Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListEmpresaBackups include inactive: %v", err)
	}
	if totalAll != 2 || len(allRows) != 2 {
		t.Fatalf("expected two backups including inactive, total=%d len=%d", totalAll, len(allRows))
	}

	backupMeta, payload, err := GetEmpresaBackupPayloadByID(dbConn, 620, backupTwo)
	if err != nil {
		t.Fatalf("GetEmpresaBackupPayloadByID: %v", err)
	}
	if backupMeta.ID != backupTwo {
		t.Fatalf("expected backup id=%d, got %d", backupTwo, backupMeta.ID)
	}
	if payload.EmpresaID != 620 {
		t.Fatalf("expected payload empresa_id=620, got %d", payload.EmpresaID)
	}
	if payload.TotalTables != 1 {
		t.Fatalf("expected payload total_tables=1, got %d", payload.TotalTables)
	}
	if payload.TotalRows != 2 {
		t.Fatalf("expected payload total_rows=2, got %d", payload.TotalRows)
	}
	if len(payload.Tables) != 1 || payload.Tables[0].Table != "empresa_backup_items_test" {
		t.Fatalf("unexpected payload tables: %+v", payload.Tables)
	}
}
