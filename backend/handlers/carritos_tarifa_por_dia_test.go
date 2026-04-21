package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaCarritosCompraListIncluyeTarifaPorDiaAutomatica(t *testing.T) {
	dbEmp := openTestSQLite(t, "carritos_tarifa_por_dia_handler.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaTarifasPorDiaSchema(dbEmp); err != nil {
		t.Fatalf("ensure tarifas por dia schema: %v", err)
	}

	carritoID, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:         1,
		Codigo:            "EST-1-55",
		Nombre:            "Habitacion 55",
		CanalVenta:        "estacion",
		Moneda:            "COP",
		ReferenciaExterna: "ESTACION_55",
		UsuarioCreador:    "qa@empresa.com",
		Estado:            "activo",
	})
	if err != nil {
		t.Fatalf("create carrito: %v", err)
	}

	if _, err := dbEmp.Exec(`UPDATE carritos_compras SET
		estado = 'activo',
		estado_carrito = 'abierto',
		activado_en = ?,
		pagado_en = NULL
	WHERE empresa_id = ? AND id = ?`, "2026-04-01 16:00:00", 1, carritoID); err != nil {
		t.Fatalf("seed activado_en: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaTarifaPorDia(dbEmp, dbpkg.EmpresaTarifaPorDia{
		EmpresaID:              1,
		EstacionID:             55,
		EstacionCodigo:         "EST-1-55",
		EstacionNombre:         "Habitacion 55",
		ServicioNombre:         "hotel",
		ValorDia:               100000,
		HoraCheckIn:            "15:00",
		HoraCheckOut:           "12:00",
		Moneda:                 "COP",
		Prioridad:              1,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa@empresa.com",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("create tarifa por dia: %v", err)
	}

	h := EmpresaCarritosCompraHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=1&include_inactive=1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list expected=%d got=%d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var rows []dbpkg.CarritoCompra
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 carrito, got %d", len(rows))
	}
	if rows[0].Total+0.001 < 100000 {
		t.Fatalf("expected total >= 100000 with tarifa diaria aplicada, got %.2f", rows[0].Total)
	}
}

func TestEmpresaCarritosCompraListIncluyeTarifaPorMinutosAutomatica(t *testing.T) {
	dbEmp := openTestSQLite(t, "carritos_tarifa_por_minutos_handler.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaTarifasPorMinutosSchema(dbEmp); err != nil {
		t.Fatalf("ensure tarifas por minutos schema: %v", err)
	}

	carritoID, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:         1,
		Codigo:            "EST-1-1",
		Nombre:            "Habitacion 1",
		CanalVenta:        "estacion",
		Moneda:            "COP",
		ReferenciaExterna: "ESTACION_1",
		UsuarioCreador:    "qa@empresa.com",
		Estado:            "activo",
	})
	if err != nil {
		t.Fatalf("create carrito: %v", err)
	}

	if _, err := dbEmp.Exec(`UPDATE carritos_compras SET
		estado = 'activo',
		estado_carrito = 'abierto',
		activado_en = ?,
		pagado_en = NULL
	WHERE empresa_id = ? AND id = ?`, time.Now().Add(-150*time.Minute).Format("2006-01-02 15:04:05"), 1, carritoID); err != nil {
		t.Fatalf("seed activado_en: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaTarifaPorMinutos(dbEmp, dbpkg.EmpresaTarifaPorMinutos{
		EmpresaID:      1,
		EstacionID:     1,
		EstacionCodigo: "EST-1-1",
		EstacionNombre: "Habitacion 1",
		DiaSemanaDesde: 1,
		DiaSemanaHasta: 7,
		MinutosBase:    120,
		ValorBase:      55000,
		MinutosExtra:   60,
		ValorExtra:     30000,
		Moneda:         "COP",
		Prioridad:      1,
		UsuarioCreador: "qa@empresa.com",
		Estado:         "activo",
	}); err != nil {
		t.Fatalf("create tarifa por minutos: %v", err)
	}

	h := EmpresaCarritosCompraHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/carritos_compra?empresa_id=1&include_inactive=1", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("list expected=%d got=%d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var rows []dbpkg.CarritoCompra
	if err := json.Unmarshal(rr.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 carrito, got %d", len(rows))
	}
	if rows[0].Total+0.001 < 85000 {
		t.Fatalf("expected total >= 85000 with tarifa por minutos aplicada, got %.2f", rows[0].Total)
	}
	if rows[0].TarifaPorMinutos == nil {
		t.Fatalf("expected tarifa_por_minutos summary in list response")
	}
	if rows[0].TarifaPorMinutos.BloquesExtra != 1 {
		t.Fatalf("expected 1 bloque extra in summary, got %d", rows[0].TarifaPorMinutos.BloquesExtra)
	}
	if rows[0].TarifaPorMinutos.FechaFinTarifaActual == "" {
		t.Fatalf("expected fecha_fin_tarifa_actual in summary")
	}
}
