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

func ensureEmpresaReportesSchema(t *testing.T, dbEmp *sql.DB) {
	t.Helper()

	if err := dbpkg.EnsureEmpresaClientesSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaClientesSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaProductosSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaCarritosSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaFinanzasSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaEventosContablesSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaDocumentosTransaccionalesSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaFacturacionElectronicaSchema: %v", err)
	}
}

func TestEmpresaReportesHandlerCatalogoSuiteDataset(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	handler := EmpresaReportesHandler(dbEmp)

	reqCatalog := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=catalogo&empresa_id=5", nil)
	rrCatalog := httptest.NewRecorder()
	handler.ServeHTTP(rrCatalog, reqCatalog)
	if rrCatalog.Code != http.StatusOK {
		t.Fatalf("catalogo status=%d body=%s", rrCatalog.Code, rrCatalog.Body.String())
	}
	var catalogResp struct {
		EmpresaID int64                        `json:"empresa_id"`
		Datasets  []empresaReporteCatalogoItem `json:"datasets"`
	}
	if err := json.Unmarshal(rrCatalog.Body.Bytes(), &catalogResp); err != nil {
		t.Fatalf("unmarshal catalogo: %v", err)
	}
	if catalogResp.EmpresaID != 5 {
		t.Fatalf("empresa_id esperado=5 obtenido=%d", catalogResp.EmpresaID)
	}
	if len(catalogResp.Datasets) < 8 {
		t.Fatalf("catalogo incompleto: %d datasets", len(catalogResp.Datasets))
	}

	reqSuite := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=suite&empresa_id=5&max_rows=120", nil)
	rrSuite := httptest.NewRecorder()
	handler.ServeHTTP(rrSuite, reqSuite)
	if rrSuite.Code != http.StatusOK {
		t.Fatalf("suite status=%d body=%s", rrSuite.Code, rrSuite.Body.String())
	}
	var suite empresaReportesSuiteResponse
	if err := json.Unmarshal(rrSuite.Body.Bytes(), &suite); err != nil {
		t.Fatalf("unmarshal suite: %v", err)
	}
	if suite.EmpresaID != 5 {
		t.Fatalf("suite empresa_id esperado=5 obtenido=%d", suite.EmpresaID)
	}
	if len(suite.Datasets) < 8 {
		t.Fatalf("suite incompleta: %d datasets", len(suite.Datasets))
	}

	reqDataset := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=5&dataset=empresarial_tablero", nil)
	rrDataset := httptest.NewRecorder()
	handler.ServeHTTP(rrDataset, reqDataset)
	if rrDataset.Code != http.StatusOK {
		t.Fatalf("dataset status=%d body=%s", rrDataset.Code, rrDataset.Body.String())
	}
	var ds empresaReporteDataset
	if err := json.Unmarshal(rrDataset.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal dataset: %v", err)
	}
	if ds.Key != "empresarial_tablero" {
		t.Fatalf("dataset key esperado=empresarial_tablero obtenido=%s", ds.Key)
	}
	if ds.RowCount != len(ds.Rows) {
		t.Fatalf("row_count inconsistente: row_count=%d rows=%d", ds.RowCount, len(ds.Rows))
	}
	if ds.RowCount == 0 {
		t.Fatalf("dataset empresarial_tablero no genero filas")
	}
}

func TestEmpresaReportesHandlerExportes(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_exportes_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	handler := EmpresaReportesHandler(dbEmp)

	cases := []struct {
		format              string
		expectedContentType string
		expectedFragment    string
	}{
		{format: "json", expectedContentType: "application/json", expectedFragment: "\"key\":\"empresarial_tablero\""},
		{format: "csv", expectedContentType: "text/csv", expectedFragment: "empresa_id"},
		{format: "txt", expectedContentType: "text/plain", expectedFragment: "Reporte:"},
		{format: "xls", expectedContentType: "application/vnd.ms-excel", expectedFragment: "empresa_id"},
	}

	for _, tc := range cases {
		url := "/api/empresa/reportes?action=export&empresa_id=8&dataset=empresarial_tablero&format=" + tc.format
		req := httptest.NewRequest(http.MethodGet, url, nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("format=%s status=%d body=%s", tc.format, rr.Code, rr.Body.String())
		}
		if ct := rr.Header().Get("Content-Type"); !strings.Contains(ct, tc.expectedContentType) {
			t.Fatalf("format=%s content-type inesperado: %s", tc.format, ct)
		}
		if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(strings.ToLower(cd), "."+tc.format) {
			t.Fatalf("format=%s content-disposition invalido: %s", tc.format, cd)
		}
		if !strings.Contains(rr.Body.String(), tc.expectedFragment) {
			t.Fatalf("format=%s contenido inesperado: %s", tc.format, rr.Body.String())
		}
	}

	reqSuiteJSON := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=export&empresa_id=8&format=json", nil)
	rrSuiteJSON := httptest.NewRecorder()
	handler.ServeHTTP(rrSuiteJSON, reqSuiteJSON)
	if rrSuiteJSON.Code != http.StatusOK {
		t.Fatalf("suite json export status=%d body=%s", rrSuiteJSON.Code, rrSuiteJSON.Body.String())
	}
	if !strings.Contains(rrSuiteJSON.Body.String(), "\"datasets\"") {
		t.Fatalf("suite json export sin datasets")
	}

	reqBadFormat := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=export&empresa_id=8&dataset=empresarial_tablero&format=pdf", nil)
	rrBadFormat := httptest.NewRecorder()
	handler.ServeHTTP(rrBadFormat, reqBadFormat)
	if rrBadFormat.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid format, got %d", rrBadFormat.Code)
	}
}
