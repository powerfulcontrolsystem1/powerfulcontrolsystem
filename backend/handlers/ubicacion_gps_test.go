package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaUbicacionGPSHandlersCRUDFlow(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_gps_handlers.db")
	if err := dbpkg.EnsureEmpresaUbicacionGPSSchema(dbEmp); err != nil {
		t.Fatalf("ensure gps schema: %v", err)
	}

	dispositivosHandler := EmpresaUbicacionGPSDispositivosHandler(dbEmp)
	recorridosHandler := EmpresaUbicacionGPSRecorridosHandler(dbEmp)

	createDeviceReq := httptest.NewRequest(http.MethodPost, "/api/empresa/ubicacion_gps/dispositivos", strings.NewReader(`{"empresa_id":9,"codigo":"GPS-01","nombre":"Unidad 01"}`))
	createDeviceReq.Header.Set("Content-Type", "application/json")
	createDeviceRR := httptest.NewRecorder()
	dispositivosHandler.ServeHTTP(createDeviceRR, createDeviceReq)
	if createDeviceRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, createDeviceRR.Code, createDeviceRR.Body.String())
	}

	var createDeviceResp map[string]interface{}
	if err := json.Unmarshal(createDeviceRR.Body.Bytes(), &createDeviceResp); err != nil {
		t.Fatalf("decode create device response: %v", err)
	}
	dispositivoID := int64(createDeviceResp["id"].(float64))
	if dispositivoID <= 0 {
		t.Fatalf("expected dispositivo id > 0, got %d", dispositivoID)
	}

	listDevicesReq := httptest.NewRequest(http.MethodGet, "/api/empresa/ubicacion_gps/dispositivos?empresa_id=9&include_inactive=1", nil)
	listDevicesRR := httptest.NewRecorder()
	dispositivosHandler.ServeHTTP(listDevicesRR, listDevicesReq)
	if listDevicesRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, listDevicesRR.Code, listDevicesRR.Body.String())
	}

	createPointBody := `{"empresa_id":9,"dispositivo_id":` + strconv.FormatInt(dispositivoID, 10) + `,"latitud":4.618,"longitud":-74.11,"precision_metros":6,"velocidad_kmh":12,"fuente":"simulado_10s"}`
	createPointReq := httptest.NewRequest(http.MethodPost, "/api/empresa/ubicacion_gps/recorridos", strings.NewReader(createPointBody))
	createPointReq.Header.Set("Content-Type", "application/json")
	createPointRR := httptest.NewRecorder()
	recorridosHandler.ServeHTTP(createPointRR, createPointReq)
	if createPointRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, createPointRR.Code, createPointRR.Body.String())
	}

	var createPointResp map[string]interface{}
	if err := json.Unmarshal(createPointRR.Body.Bytes(), &createPointResp); err != nil {
		t.Fatalf("decode create point response: %v", err)
	}
	puntoID := int64(createPointResp["id"].(float64))
	if puntoID <= 0 {
		t.Fatalf("expected punto id > 0, got %d", puntoID)
	}

	listPointsReq := httptest.NewRequest(http.MethodGet, "/api/empresa/ubicacion_gps/recorridos?empresa_id=9&dispositivo_id="+strconv.FormatInt(dispositivoID, 10)+"&limit=50", nil)
	listPointsRR := httptest.NewRecorder()
	recorridosHandler.ServeHTTP(listPointsRR, listPointsReq)
	if listPointsRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, listPointsRR.Code, listPointsRR.Body.String())
	}

	var puntos []dbpkg.EmpresaGPSRecorrido
	if err := json.Unmarshal(listPointsRR.Body.Bytes(), &puntos); err != nil {
		t.Fatalf("decode list points response: %v", err)
	}
	if len(puntos) != 1 {
		t.Fatalf("expected 1 punto, got %d", len(puntos))
	}

	togglePointReq := httptest.NewRequest(http.MethodPut, "/api/empresa/ubicacion_gps/recorridos?empresa_id=9&id="+strconv.FormatInt(puntoID, 10)+"&action=desactivar", nil)
	togglePointRR := httptest.NewRecorder()
	recorridosHandler.ServeHTTP(togglePointRR, togglePointReq)
	if togglePointRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, togglePointRR.Code, togglePointRR.Body.String())
	}

	toggleDeviceReq := httptest.NewRequest(http.MethodPut, "/api/empresa/ubicacion_gps/dispositivos?empresa_id=9&id="+strconv.FormatInt(dispositivoID, 10)+"&action=desactivar", nil)
	toggleDeviceRR := httptest.NewRecorder()
	dispositivosHandler.ServeHTTP(toggleDeviceRR, toggleDeviceReq)
	if toggleDeviceRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, toggleDeviceRR.Code, toggleDeviceRR.Body.String())
	}

	deletePointReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/ubicacion_gps/recorridos?empresa_id=9&id="+strconv.FormatInt(puntoID, 10), nil)
	deletePointRR := httptest.NewRecorder()
	recorridosHandler.ServeHTTP(deletePointRR, deletePointReq)
	if deletePointRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, deletePointRR.Code, deletePointRR.Body.String())
	}

	deleteDeviceReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/ubicacion_gps/dispositivos?empresa_id=9&id="+strconv.FormatInt(dispositivoID, 10), nil)
	deleteDeviceRR := httptest.NewRecorder()
	dispositivosHandler.ServeHTTP(deleteDeviceRR, deleteDeviceReq)
	if deleteDeviceRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, deleteDeviceRR.Code, deleteDeviceRR.Body.String())
	}
}
