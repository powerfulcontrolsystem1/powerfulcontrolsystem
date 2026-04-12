package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func openSensorTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "sensor_puertas_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = dbConn.Close() })
	return dbConn
}

func TestSensorPuertasSchemaUpsertAndHeartbeat(t *testing.T) {
	dbConn := openSensorTestDB(t)
	if err := EnsureEmpresaSensorPuertasSchema(dbConn); err != nil {
		t.Fatalf("ensure sensor schema: %v", err)
	}

	p := EmpresaSensorDevice{
		EmpresaID:      7,
		DeviceID:       "RPI-TEST-001",
		EstacionID:     3,
		UsuarioCreador: "tester",
		Estado:         "activo",
	}
	id, err := UpsertEmpresaSensorDevice(dbConn, &p)
	if err != nil {
		t.Fatalf("upsert device: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected inserted id > 0, got %d", id)
	}

	got, err := GetEmpresaSensorByDeviceID(dbConn, "RPI-TEST-001")
	if err != nil {
		t.Fatalf("get device by id: %v", err)
	}
	if got.EmpresaID != p.EmpresaID {
		t.Fatalf("expected empresa %d got %d", p.EmpresaID, got.EmpresaID)
	}
	if got.EstacionID != p.EstacionID {
		t.Fatalf("expected estacion %d got %d", p.EstacionID, got.EstacionID)
	}

	empresaID, estacionID, err := UpdateDeviceHeartbeat(dbConn, "rpi-test-001", "on")
	if err != nil {
		t.Fatalf("heartbeat update: %v", err)
	}
	if empresaID != p.EmpresaID {
		t.Fatalf("heartbeat empresa mismatch expected %d got %d", p.EmpresaID, empresaID)
	}
	if estacionID != p.EstacionID {
		t.Fatalf("heartbeat estacion mismatch expected %d got %d", p.EstacionID, estacionID)
	}

	got2, err := GetEmpresaSensorByDeviceID(dbConn, "rpi-test-001")
	if err != nil {
		t.Fatalf("get device after heartbeat: %v", err)
	}
	if got2.LastState == "" {
		t.Fatalf("expected last_state updated, got empty")
	}
	if got2.LastSeen == "" {
		t.Fatalf("expected last_seen updated, got empty")
	}
}
