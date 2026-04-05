package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaReservasHotelHandlerCRUDAndDisponibilidad(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reservas_hotel_handler.db")
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaReservasHotelSchema(dbEmp); err != nil {
		t.Fatalf("ensure reservas schema: %v", err)
	}

	if _, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:         1,
		Codigo:            "EST-1-1",
		Nombre:            "Habitacion 1",
		CanalVenta:        "estacion",
		Moneda:            "COP",
		ReferenciaExterna: "ESTACION_1",
		UsuarioCreador:    "test",
		Estado:            "activo",
	}); err != nil {
		t.Fatalf("create station 1: %v", err)
	}
	if _, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:         1,
		Codigo:            "EST-1-2",
		Nombre:            "Habitacion 2",
		CanalVenta:        "estacion",
		Moneda:            "COP",
		ReferenciaExterna: "ESTACION_2",
		UsuarioCreador:    "test",
		Estado:            "activo",
	}); err != nil {
		t.Fatalf("create station 2: %v", err)
	}

	h := EmpresaReservasHotelHandler(dbEmp)
	entrada := time.Now().Add(2 * time.Hour).Format("2006-01-02 15:04:05")
	salida := time.Now().Add(6 * time.Hour).Format("2006-01-02 15:04:05")

	createBody := fmt.Sprintf(`{"empresa_id":1,"estacion_id":1,"cliente_nombre":"Cliente Handler","cliente_documento":"1111","cliente_email":"cliente-handler@test.com","cantidad_huespedes":2,"fecha_entrada":"%s","fecha_salida":"%s","monto_total":120000,"moneda":"COP"}`,
		entrada,
		salida,
	)
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/reservas_hotel", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create expected=%d got=%d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var created dbpkg.ReservaHotel
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected id > 0, got %d", created.ID)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/reservas_hotel?empresa_id=1&limit=20", nil)
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list expected=%d got=%d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var listRows []dbpkg.ReservaHotel
	if err := json.Unmarshal(listRR.Body.Bytes(), &listRows); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listRows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(listRows))
	}

	dispURL := "/api/empresa/reservas_hotel?empresa_id=1&action=disponibilidad&fecha_entrada=" +
		url.QueryEscape(entrada) + "&fecha_salida=" + url.QueryEscape(salida)
	dispReq := httptest.NewRequest(http.MethodGet, dispURL, nil)
	dispRR := httptest.NewRecorder()
	h.ServeHTTP(dispRR, dispReq)
	if dispRR.Code != http.StatusOK {
		t.Fatalf("disponibilidad expected=%d got=%d body=%s", http.StatusOK, dispRR.Code, dispRR.Body.String())
	}
	var estaciones []dbpkg.ReservaHotelEstacion
	if err := json.Unmarshal(dispRR.Body.Bytes(), &estaciones); err != nil {
		t.Fatalf("decode disponibilidad response: %v", err)
	}
	if len(estaciones) < 2 {
		t.Fatalf("expected at least 2 stations, got %d", len(estaciones))
	}
	station1Found := false
	for _, st := range estaciones {
		if st.EstacionID == 1 {
			station1Found = true
			if st.Disponible {
				t.Fatalf("expected estacion 1 unavailable while pending reservation")
			}
		}
	}
	if !station1Found {
		t.Fatal("expected estacion 1 in disponibilidad response")
	}

	updateBody := fmt.Sprintf(`{"empresa_id":1,"id":%d,"estacion_id":2,"cliente_nombre":"Cliente Handler Editado","fecha_entrada":"%s","fecha_salida":"%s","cantidad_huespedes":3,"monto_total":130000,"moneda":"COP","observaciones":"update test"}`,
		created.ID,
		entrada,
		salida,
	)
	updateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/reservas_hotel", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	h.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("update expected=%d got=%d body=%s", http.StatusOK, updateRR.Code, updateRR.Body.String())
	}

	detailURL := fmt.Sprintf("/api/empresa/reservas_hotel?empresa_id=1&action=detalle&id=%d", created.ID)
	detailReq := httptest.NewRequest(http.MethodGet, detailURL, nil)
	detailRR := httptest.NewRecorder()
	h.ServeHTTP(detailRR, detailReq)
	if detailRR.Code != http.StatusOK {
		t.Fatalf("detail expected=%d got=%d body=%s", http.StatusOK, detailRR.Code, detailRR.Body.String())
	}
	var detail dbpkg.ReservaHotel
	if err := json.Unmarshal(detailRR.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode detail response: %v", err)
	}
	if detail.EstacionID != 2 {
		t.Fatalf("expected estacion_id=2 after update, got %d", detail.EstacionID)
	}

	confirmBody := fmt.Sprintf(`{"empresa_id":1,"id":%d,"referencia_pago":"TRX-HANDLER-1","observaciones":"confirmada en test"}`,
		created.ID,
	)
	confirmReq := httptest.NewRequest(http.MethodPut, "/api/empresa/reservas_hotel?action=confirmar_pago", strings.NewReader(confirmBody))
	confirmReq.Header.Set("Content-Type", "application/json")
	confirmRR := httptest.NewRecorder()
	h.ServeHTTP(confirmRR, confirmReq)
	if confirmRR.Code != http.StatusOK {
		t.Fatalf("confirm expected=%d got=%d body=%s", http.StatusOK, confirmRR.Code, confirmRR.Body.String())
	}

	detailReq = httptest.NewRequest(http.MethodGet, detailURL, nil)
	detailRR = httptest.NewRecorder()
	h.ServeHTTP(detailRR, detailReq)
	if detailRR.Code != http.StatusOK {
		t.Fatalf("detail after confirm expected=%d got=%d body=%s", http.StatusOK, detailRR.Code, detailRR.Body.String())
	}
	if err := json.Unmarshal(detailRR.Body.Bytes(), &detail); err != nil {
		t.Fatalf("decode detail after confirm: %v", err)
	}
	if detail.EstadoReserva != "confirmada" || detail.EstadoPago != "confirmado" {
		t.Fatalf("expected confirmada/confirmado, got %s/%s", detail.EstadoReserva, detail.EstadoPago)
	}

	disableURL := fmt.Sprintf("/api/empresa/reservas_hotel?empresa_id=1&id=%d&action=desactivar", created.ID)
	disableReq := httptest.NewRequest(http.MethodPut, disableURL, nil)
	disableRR := httptest.NewRecorder()
	h.ServeHTTP(disableRR, disableReq)
	if disableRR.Code != http.StatusOK {
		t.Fatalf("desactivar expected=%d got=%d body=%s", http.StatusOK, disableRR.Code, disableRR.Body.String())
	}

	deleteURL := fmt.Sprintf("/api/empresa/reservas_hotel?empresa_id=1&id=%d", created.ID)
	deleteReq := httptest.NewRequest(http.MethodDelete, deleteURL, nil)
	deleteRR := httptest.NewRecorder()
	h.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete expected=%d got=%d body=%s", http.StatusOK, deleteRR.Code, deleteRR.Body.String())
	}
}
