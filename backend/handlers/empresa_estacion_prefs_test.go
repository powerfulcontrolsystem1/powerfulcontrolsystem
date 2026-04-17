package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

type estacionPrefsResponse struct {
	OK    bool                        `json:"ok"`
	Prefs []dbpkg.EmpresaEstacionPref `json:"prefs"`
}

func TestEmpresaEstacionPrefsHandler_UpsertAndIsolationByEmpresa(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_estacion_prefs_handler.db")
	ensureClientesSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaEstacionPrefsSchema(dbEmp); err != nil {
		t.Fatalf("ensure estacion prefs schema: %v", err)
	}

	h := EmpresaEstacionPrefsHandler(dbEmp)

	buildConfig := func(cantidad int, prefix string) string {
		estaciones := make([]map[string]interface{}, 0, cantidad)
		for i := 1; i <= cantidad; i++ {
			estaciones = append(estaciones, map[string]interface{}{
				"id":                      i,
				"nombre":                  prefix + " " + string(rune('A'+i-1)),
				"venta_simple_habilitada": i%2 == 0,
				"mostrar_total":           true,
			})
		}
		cfg := map[string]interface{}{
			"cantidad":   cantidad,
			"estaciones": estaciones,
			"card_size":  "medium",
		}
		by, err := json.Marshal(cfg)
		if err != nil {
			t.Fatalf("marshal cfg: %v", err)
		}
		return string(by)
	}

	putPref := func(empresaID int64, valor string) {
		payload := map[string]interface{}{
			"estacion_id": 0,
			"clave":       "estaciones_config",
			"valor":       valor,
		}
		by, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		req := httptest.NewRequest(http.MethodPut, "/api/empresa/estacion_prefs?empresa_id="+strconv.FormatInt(empresaID, 10), bytes.NewReader(by))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d for empresa %d, got %d body=%s", http.StatusOK, empresaID, rr.Code, rr.Body.String())
		}
	}

	putPref(701, buildConfig(10, "Estacion"))
	putPref(702, buildConfig(3, "Mesa"))

	carritos701, err := dbpkg.GetCarritosCompraByEmpresa(dbEmp, 701, true, "")
	if err != nil {
		t.Fatalf("list synced carritos empresa 701: %v", err)
	}
	if len(carritos701) != 10 {
		t.Fatalf("expected 10 synced carritos for empresa 701, got %d", len(carritos701))
	}
	for _, item := range carritos701 {
		if item.ReferenciaExterna == "ESTACION_1" {
			if item.Codigo != "EST-701-1" {
				t.Fatalf("expected linked carrito code EST-701-1, got %+v", item)
			}
			if item.Nombre != "Estacion A" {
				t.Fatalf("expected linked carrito name Estacion A, got %+v", item)
			}
		}
	}

	getPrefs := func(empresaID int64) estacionPrefsResponse {
		req := httptest.NewRequest(http.MethodGet, "/api/empresa/estacion_prefs?empresa_id="+strconv.FormatInt(empresaID, 10), nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d on list empresa %d, got %d body=%s", http.StatusOK, empresaID, rr.Code, rr.Body.String())
		}
		var resp estacionPrefsResponse
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode prefs response empresa %d: %v", empresaID, err)
		}
		return resp
	}

	resp701 := getPrefs(701)
	if !resp701.OK {
		t.Fatalf("expected ok=true for empresa 701, got %+v", resp701)
	}
	if len(resp701.Prefs) == 0 {
		t.Fatalf("expected prefs for empresa 701")
	}
	for _, pref := range resp701.Prefs {
		if pref.EmpresaID != 701 {
			t.Fatalf("expected empresa_id=701 in prefs, got %d", pref.EmpresaID)
		}
	}

	var pref701 *dbpkg.EmpresaEstacionPref
	for i := range resp701.Prefs {
		if resp701.Prefs[i].Clave == "estaciones_config" && resp701.Prefs[i].EstacionID == 0 {
			pref701 = &resp701.Prefs[i]
			break
		}
	}
	if pref701 == nil {
		t.Fatalf("expected estaciones_config for empresa 701")
	}

	var cfg701 struct {
		Cantidad   int `json:"cantidad"`
		Estaciones []struct {
			ID int `json:"id"`
		} `json:"estaciones"`
	}
	if err := json.Unmarshal([]byte(pref701.Valor), &cfg701); err != nil {
		t.Fatalf("decode estaciones_config empresa 701: %v", err)
	}
	if cfg701.Cantidad != 10 {
		t.Fatalf("expected cantidad=10 for empresa 701, got %d", cfg701.Cantidad)
	}
	if len(cfg701.Estaciones) != 10 {
		t.Fatalf("expected 10 estaciones in empresa 701 config, got %d", len(cfg701.Estaciones))
	}

	resp702 := getPrefs(702)
	if !resp702.OK {
		t.Fatalf("expected ok=true for empresa 702, got %+v", resp702)
	}
	if len(resp702.Prefs) == 0 {
		t.Fatalf("expected prefs for empresa 702")
	}
	for _, pref := range resp702.Prefs {
		if pref.EmpresaID != 702 {
			t.Fatalf("expected empresa_id=702 in prefs, got %d", pref.EmpresaID)
		}
	}
}
