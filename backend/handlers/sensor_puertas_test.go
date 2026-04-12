package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
	_ "modernc.org/sqlite"
)

func openHandlerSensorTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "handler_sensor_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	dbConn.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = dbConn.Close() })
	return dbConn
}

func TestPublicSensorHeartbeat(t *testing.T) {
	dbConn := openHandlerSensorTestDB(t)
	if err := dbpkg.EnsureEmpresaSensorPuertasSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	// seed a device mapping
	id, err := dbpkg.UpsertEmpresaSensorDevice(dbConn, &dbpkg.EmpresaSensorDevice{
		EmpresaID:      999,
		DeviceID:       "rpi-hb-1",
		EstacionID:     11,
		UsuarioCreador: "tester",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("seed device: %v", err)
	}
	if id <= 0 {
		t.Fatalf("invalid seeded id: %d", id)
	}

	payload := map[string]string{"device_id": "rpi-hb-1", "state": "on"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/?action=heartbeat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h := PublicSensorPuertasHandler(dbConn)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	var res map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if ok, _ := res["ok"].(bool); !ok {
		t.Fatalf("expected ok true, got %v", res)
	}
	// json numbers decode to float64
	if int64(res["empresa_id"].(float64)) != 999 {
		t.Fatalf("empresa_id mismatch: %v", res["empresa_id"])
	}
	if int64(res["estacion_id"].(float64)) != 11 {
		t.Fatalf("estacion_id mismatch: %v", res["estacion_id"])
	}
}

func TestEmpresaSensorConfigHandler_CreateAndList(t *testing.T) {
	dbConn := openHandlerSensorTestDB(t)
	if err := dbpkg.EnsureEmpresaSensorPuertasSchema(dbConn); err != nil {
		t.Fatalf("ensure schema: %v", err)
	}

	// Create via handler
	payload := map[string]interface{}{"device_id": "rpi-new-1", "estacion_id": 13}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/?empresa_id=42", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Admin-Email", "admin@test.local")
	rr := httptest.NewRecorder()

	h := EmpresaSensorConfigHandler(dbConn)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 creating device, got %d body=%s", rr.Code, rr.Body.String())
	}
	var createRes map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&createRes); err != nil {
		t.Fatalf("decode create res: %v", err)
	}
	if ok, _ := createRes["ok"].(bool); !ok {
		t.Fatalf("expected ok true create, got %v", createRes)
	}

	// List via GET
	req2 := httptest.NewRequest(http.MethodGet, "/?empresa_id=42", nil)
	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 listing devices, got %d body=%s", rr2.Code, rr2.Body.String())
	}
	var listRes []map[string]interface{}
	if err := json.NewDecoder(rr2.Body).Decode(&listRes); err != nil {
		t.Fatalf("decode list res: %v", err)
	}
	found := false
	for _, d := range listRes {
		if d["device_id"] == "rpi-new-1" || d["estacion_id"] == float64(13) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("created device not found in list: %v", listRes)
	}
}
