package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaClientesHandlerPerfilHistorialSegmentacion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_clientes_perfil_historial_handler.db")
	if err := dbpkg.EnsureEmpresaClientesSchema(dbEmp); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	clienteID, err := dbpkg.CreateCliente(dbEmp, dbpkg.Cliente{
		EmpresaID:         410,
		TipoDocumento:     "CC",
		NumeroDocumento:   "777222111",
		NombreRazonSocial: "Cliente Handler",
		UsuarioCreador:    "tester",
	})
	if err != nil {
		t.Fatalf("create cliente: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO carritos_compras (
		empresa_id, codigo, nombre, cliente_id, estado_carrito, total, total_pagado, pagado_en, fecha_creacion, fecha_actualizacion, estado
	) VALUES (410, 'CAR-H1', 'Venta H1', ?, 'cerrado', 150000, 150000, datetime('now','-3 day'), datetime('now','-3 day'), datetime('now','-3 day'), 'activo')`, clienteID); err != nil {
		t.Fatalf("insert carrito historial: %v", err)
	}

	h := EmpresaClientesHandler(dbEmp)

	reqSegment := httptest.NewRequest(http.MethodGet, "/api/empresa/clientes?empresa_id=410&action=segmentacion", nil)
	rrSegment := httptest.NewRecorder()
	h.ServeHTTP(rrSegment, reqSegment)
	if rrSegment.Code != http.StatusOK {
		t.Fatalf("expected segmentacion status %d, got %d body=%s", http.StatusOK, rrSegment.Code, rrSegment.Body.String())
	}
	var segmentos []dbpkg.ClienteSegmentacionResumen
	if err := json.Unmarshal(rrSegment.Body.Bytes(), &segmentos); err != nil {
		t.Fatalf("decode segmentacion: %v", err)
	}
	if len(segmentos) == 0 {
		t.Fatalf("expected segmentacion con datos")
	}

	reqPerfil := httptest.NewRequest(http.MethodGet, "/api/empresa/clientes?empresa_id=410&action=perfil&id="+strconv.FormatInt(clienteID, 10), nil)
	rrPerfil := httptest.NewRecorder()
	h.ServeHTTP(rrPerfil, reqPerfil)
	if rrPerfil.Code != http.StatusOK {
		t.Fatalf("expected perfil status %d, got %d body=%s", http.StatusOK, rrPerfil.Code, rrPerfil.Body.String())
	}
	var perfil dbpkg.ClientePerfilComercial
	if err := json.Unmarshal(rrPerfil.Body.Bytes(), &perfil); err != nil {
		t.Fatalf("decode perfil: %v", err)
	}
	if perfil.Cliente.ID != clienteID {
		t.Fatalf("expected perfil cliente_id=%d, got %d", clienteID, perfil.Cliente.ID)
	}
	if perfil.NumeroCompras != 1 {
		t.Fatalf("expected numero_compras=1, got %d", perfil.NumeroCompras)
	}

	reqHistorial := httptest.NewRequest(http.MethodGet, "/api/empresa/clientes?empresa_id=410&action=historial&id="+strconv.FormatInt(clienteID, 10)+"&limit=5", nil)
	rrHistorial := httptest.NewRecorder()
	h.ServeHTTP(rrHistorial, reqHistorial)
	if rrHistorial.Code != http.StatusOK {
		t.Fatalf("expected historial status %d, got %d body=%s", http.StatusOK, rrHistorial.Code, rrHistorial.Body.String())
	}
	var historial []dbpkg.ClienteCompraHistorial
	if err := json.Unmarshal(rrHistorial.Body.Bytes(), &historial); err != nil {
		t.Fatalf("decode historial: %v", err)
	}
	if len(historial) != 1 {
		t.Fatalf("expected historial len=1, got %d", len(historial))
	}

	reqBad := httptest.NewRequest(http.MethodGet, "/api/empresa/clientes?empresa_id=410&action=perfil", nil)
	rrBad := httptest.NewRecorder()
	h.ServeHTTP(rrBad, reqBad)
	if rrBad.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request status %d, got %d body=%s", http.StatusBadRequest, rrBad.Code, rrBad.Body.String())
	}

	reqNotFound := httptest.NewRequest(http.MethodGet, "/api/empresa/clientes?empresa_id=410&action=perfil&id=999999", nil)
	rrNotFound := httptest.NewRecorder()
	h.ServeHTTP(rrNotFound, reqNotFound)
	if rrNotFound.Code != http.StatusNotFound {
		t.Fatalf("expected not found status %d, got %d body=%s", http.StatusNotFound, rrNotFound.Code, rrNotFound.Body.String())
	}
}
