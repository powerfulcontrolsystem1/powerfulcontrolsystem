package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func ensureSuperEmpresasSchemaForReportesGlobales(t *testing.T, dbSuper *sql.DB) {
	t.Helper()
	if _, err := dbSuper.Exec(`CREATE TABLE IF NOT EXISTS empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER,
		tipo_id INTEGER,
		tipo_nombre TEXT,
		nombre TEXT NOT NULL,
		nit TEXT,
		fecha_creacion TEXT,
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`); err != nil {
		t.Fatalf("create super empresas schema: %v", err)
	}
}

func insertVentaReporteGlobal(t *testing.T, dbEmp *sql.DB, empresaID int64, codigo string, total float64, clienteID int64) {
	t.Helper()
	if _, err := dbEmp.Exec(`INSERT INTO carritos_compras (
		empresa_id, codigo, nombre, cliente_id, estado_carrito, total, total_pagado, canal_venta, pagado_en, fecha_creacion, fecha_actualizacion, estado
	) VALUES (?, ?, ?, ?, 'cerrado', ?, ?, 'mostrador', datetime('now','localtime'), datetime('now','localtime'), datetime('now','localtime'), 'activo')`, empresaID, codigo, codigo, clienteID, total, total); err != nil {
		t.Fatalf("insert venta reporte global: %v", err)
	}
}

func TestSuperReportesGlobalesHandlerFiltraYConsolidaPorAdministrador(t *testing.T) {
	dbEmp := openTestSQLite(t, "reportes_globales_empresas.db")
	dbSuper := openTestSQLite(t, "reportes_globales_super.db")
	ensureEmpresaReportesSchema(t, dbEmp)
	ensureSuperEmpresasSchemaForReportesGlobales(t, dbSuper)

	empresaUno, err := dbpkg.CreateEmpresa(dbSuper, 0, "", "Empresa Uno", "NIT-1", "", "admin1@test.com")
	if err != nil {
		t.Fatalf("CreateEmpresa empresaUno: %v", err)
	}
	empresaDos, err := dbpkg.CreateEmpresa(dbSuper, 0, "", "Empresa Dos", "NIT-2", "", "admin1@test.com")
	if err != nil {
		t.Fatalf("CreateEmpresa empresaDos: %v", err)
	}
	if _, err := dbpkg.CreateEmpresa(dbSuper, 0, "", "Empresa Tres", "NIT-3", "", "admin2@test.com"); err != nil {
		t.Fatalf("CreateEmpresa empresaTres: %v", err)
	}

	clienteUno, err := dbpkg.CreateCliente(dbEmp, dbpkg.Cliente{EmpresaID: empresaUno, NombreRazonSocial: "Cliente Uno", Estado: "activo"})
	if err != nil {
		t.Fatalf("CreateCliente uno: %v", err)
	}
	clienteDos, err := dbpkg.CreateCliente(dbEmp, dbpkg.Cliente{EmpresaID: empresaDos, NombreRazonSocial: "Cliente Dos", Estado: "activo"})
	if err != nil {
		t.Fatalf("CreateCliente dos: %v", err)
	}
	insertVentaReporteGlobal(t, dbEmp, empresaUno, "CAR-UNO", 150000, clienteUno)
	insertVentaReporteGlobal(t, dbEmp, empresaDos, "CAR-DOS", 90000, clienteDos)

	handler := SuperReportesGlobalesHandler(dbEmp, dbSuper)

	reqCatalog := httptest.NewRequest(http.MethodGet, "/super/api/reportes_globales?action=catalogo", nil)
	reqCatalog = reqCatalog.WithContext(context.WithValue(reqCatalog.Context(), "adminEmail", "admin1@test.com"))
	rrCatalog := httptest.NewRecorder()
	handler.ServeHTTP(rrCatalog, reqCatalog)
	if rrCatalog.Code != http.StatusOK {
		t.Fatalf("catalogo status=%d body=%s", rrCatalog.Code, rrCatalog.Body.String())
	}
	var catalogResp struct {
		AdminEmail string          `json:"admin_email"`
		Empresas   []dbpkg.Empresa `json:"empresas"`
	}
	if err := json.Unmarshal(rrCatalog.Body.Bytes(), &catalogResp); err != nil {
		t.Fatalf("unmarshal catalogo: %v", err)
	}
	if catalogResp.AdminEmail != "admin1@test.com" {
		t.Fatalf("admin_email inesperado: %s", catalogResp.AdminEmail)
	}
	if len(catalogResp.Empresas) != 2 {
		t.Fatalf("se esperaban 2 empresas del admin1, se obtuvieron %d", len(catalogResp.Empresas))
	}

	reqDataset := httptest.NewRequest(http.MethodGet, "/super/api/reportes_globales?action=dataset&dataset=operativo_ventas_detalle&modo=consolidado", nil)
	reqDataset = reqDataset.WithContext(context.WithValue(reqDataset.Context(), "adminEmail", "admin1@test.com"))
	rrDataset := httptest.NewRecorder()
	handler.ServeHTTP(rrDataset, reqDataset)
	if rrDataset.Code != http.StatusOK {
		t.Fatalf("dataset status=%d body=%s", rrDataset.Code, rrDataset.Body.String())
	}
	var datasetResp superReportesDatasetResponse
	if err := json.Unmarshal(rrDataset.Body.Bytes(), &datasetResp); err != nil {
		t.Fatalf("unmarshal dataset: %v", err)
	}
	if datasetResp.Combinado.RowCount != 2 {
		t.Fatalf("filas consolidadas esperadas=2 obtenidas=%d", datasetResp.Combinado.RowCount)
	}
	if got := strings.TrimSpace(reportesStringValue(datasetResp.Combinado.Summary["ingresos"])); got != "240000.00" {
		t.Fatalf("ingresos consolidados inesperados: %s", got)
	}
	rowUno, ok := reporteDatasetFindRowByStringField(datasetResp.Combinado.Rows, "empresa_nombre", "Empresa Uno")
	if !ok {
		t.Fatalf("no se encontro fila consolidada para Empresa Uno")
	}
	if got := reporteDatasetToFloat64(rowUno["total"]); got != 150000 {
		t.Fatalf("total empresa uno inesperado: %.2f", got)
	}

	reqIndividual := httptest.NewRequest(http.MethodGet, "/super/api/reportes_globales?action=dataset&dataset=operativo_ventas_detalle&modo=individual&empresa_ids="+strings.Join([]string{strconv.FormatInt(empresaUno, 10), strconv.FormatInt(empresaDos, 10)}, ","), nil)
	reqIndividual = reqIndividual.WithContext(context.WithValue(reqIndividual.Context(), "adminEmail", "admin1@test.com"))
	rrIndividual := httptest.NewRecorder()
	handler.ServeHTTP(rrIndividual, reqIndividual)
	if rrIndividual.Code != http.StatusOK {
		t.Fatalf("individual status=%d body=%s", rrIndividual.Code, rrIndividual.Body.String())
	}
	var individualResp superReportesDatasetResponse
	if err := json.Unmarshal(rrIndividual.Body.Bytes(), &individualResp); err != nil {
		t.Fatalf("unmarshal individual: %v", err)
	}
	if len(individualResp.Individuales) != 2 {
		t.Fatalf("datasets individuales esperados=2 obtenidos=%d", len(individualResp.Individuales))
	}
	for _, item := range individualResp.Individuales {
		if item.Empresa.UsuarioCreador != "admin1@test.com" {
			t.Fatalf("empresa fuera de alcance devuelta en individual: %+v", item.Empresa)
		}
		if item.Dataset.RowCount != 1 {
			t.Fatalf("cada dataset individual debe tener 1 fila, obtenido=%d para empresa=%d", item.Dataset.RowCount, item.Empresa.ID)
		}
	}

	reqSingle := httptest.NewRequest(http.MethodGet, "/super/api/reportes_globales?action=dataset&dataset=operativo_ventas_detalle&modo=consolidado&empresa_id="+strconv.FormatInt(empresaDos, 10), nil)
	reqSingle = reqSingle.WithContext(context.WithValue(reqSingle.Context(), "adminEmail", "admin1@test.com"))
	rrSingle := httptest.NewRecorder()
	handler.ServeHTTP(rrSingle, reqSingle)
	if rrSingle.Code != http.StatusOK {
		t.Fatalf("single status=%d body=%s", rrSingle.Code, rrSingle.Body.String())
	}
	var singleResp superReportesDatasetResponse
	if err := json.Unmarshal(rrSingle.Body.Bytes(), &singleResp); err != nil {
		t.Fatalf("unmarshal single: %v", err)
	}
	if len(singleResp.Empresas) != 1 || singleResp.Empresas[0].ID != empresaDos {
		t.Fatalf("seleccion singular inesperada: %+v", singleResp.Empresas)
	}
	if singleResp.Combinado.RowCount != 1 {
		t.Fatalf("filas consolidadas singulares esperadas=1 obtenidas=%d", singleResp.Combinado.RowCount)
	}
	rowDos, ok := reporteDatasetFindRowByStringField(singleResp.Combinado.Rows, "empresa_nombre", "Empresa Dos")
	if !ok {
		t.Fatalf("no se encontro fila singular para Empresa Dos")
	}
	if got := reporteDatasetToFloat64(rowDos["total"]); got != 90000 {
		t.Fatalf("total empresa dos inesperado: %.2f", got)
	}
}
