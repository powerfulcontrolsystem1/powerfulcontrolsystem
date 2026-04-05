package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	if err := dbpkg.EnsureEmpresaAsistenciaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaAsistenciaSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaNominaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaNominaSchema: %v", err)
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
		{format: "pdf", expectedContentType: "application/pdf", expectedFragment: "%PDF-"},
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

	reqBadFormat := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=export&empresa_id=8&dataset=empresarial_tablero&format=docx", nil)
	rrBadFormat := httptest.NewRecorder()
	handler.ServeHTTP(rrBadFormat, reqBadFormat)
	if rrBadFormat.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid format, got %d", rrBadFormat.Code)
	}
}

func TestEmpresaReportesHandlerDatasetNominaLiquidaciones(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_nomina_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(19)
	empleadoNominaID, err := dbpkg.CreateEmpresaNominaEmpleado(dbEmp, dbpkg.EmpresaNominaEmpleado{
		EmpresaID:                empresaID,
		EmpleadoID:               1901,
		EmpleadoCodigo:           "EMP-1901",
		EmpleadoNombre:           "Nora Luna",
		EmpleadoDocumento:        "900190190",
		Cargo:                    "Analista",
		TipoContrato:             "indefinido",
		SalarioBasicoMensual:     2800000,
		AuxilioTransporteMensual: 200000,
		JornadaHorasDia:          8,
		UsuarioCreador:           "qa_nomina",
		Estado:                   "activo",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaNominaEmpleado: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaAsistenciaEmpleado(dbEmp, dbpkg.EmpresaAsistenciaEmpleado{
		EmpresaID:         empresaID,
		EmpleadoID:        1901,
		EmpleadoCodigo:    "EMP-1901",
		EmpleadoNombre:    "Nora Luna",
		EmpleadoDocumento: "900190190",
		Cargo:             "Analista",
		FechaAsistencia:   "2026-04-08",
		HoraEntrada:       "08:00:00",
		HoraSalida:        "17:00:00",
		HorasTrabajadas:   9,
		EstadoAsistencia:  "presente",
		UsuarioCreador:    "qa_nomina",
	}); err != nil {
		t.Fatalf("CreateEmpresaAsistenciaEmpleado: %v", err)
	}

	result, err := dbpkg.GenerateEmpresaNominaLiquidaciones(dbEmp, dbpkg.EmpresaNominaCalculoRequest{
		EmpresaID:        empresaID,
		PeriodoDesde:     "2026-04-08",
		PeriodoHasta:     "2026-04-08",
		EmpleadoNominaID: empleadoNominaID,
		Overwrite:        true,
		UsuarioCreador:   "qa_nomina",
	})
	if err != nil {
		t.Fatalf("GenerateEmpresaNominaLiquidaciones: %v", err)
	}
	if result == nil || result.Calculados != 1 {
		t.Fatalf("calculo nomina invalido: %+v", result)
	}

	handler := EmpresaReportesHandler(dbEmp)
	url := "/api/empresa/reportes?action=dataset&empresa_id=19&dataset=contable_nomina_liquidaciones&desde=2026-04-08&hasta=2026-04-08&empleado_nomina_id=" + strconv.FormatInt(empleadoNominaID, 10)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset contable_nomina_liquidaciones status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal contable_nomina_liquidaciones: %v", err)
	}
	if ds.Key != reporteDatasetContableNomina {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetContableNomina, ds.Key)
	}
	if ds.RowCount != 1 {
		t.Fatalf("row_count esperado=1 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	row := ds.Rows[0]
	if strings.TrimSpace(row["empleado_nombre"].(string)) != "Nora Luna" {
		t.Fatalf("empleado_nombre inesperado: %v", row["empleado_nombre"])
	}
	if int64(row["empleado_nomina_id"].(float64)) != empleadoNominaID {
		t.Fatalf("empleado_nomina_id esperado=%d obtenido=%v", empleadoNominaID, row["empleado_nomina_id"])
	}

	totalNeto, _ := ds.Summary["total_neto"].(float64)
	if totalNeto <= 0 {
		t.Fatalf("summary total_neto debe ser mayor que cero")
	}
	if ds.Summary["moneda"].(string) == "" {
		t.Fatalf("summary moneda debe estar informado")
	}
	filtroEmpleado, _ := ds.Summary["filtro_empleado_nomina_id"].(float64)
	if int64(filtroEmpleado) != empleadoNominaID {
		t.Fatalf("filtro_empleado_nomina_id esperado=%d obtenido=%v", empleadoNominaID, ds.Summary["filtro_empleado_nomina_id"])
	}
}

func TestEmpresaReportesHandlerDatasetContableFlujoCaja(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_flujo_caja_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(27)
	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "ingreso",
		Codigo:          "ING-2701",
		FechaMovimiento: "2026-04-01 09:00:00",
		Categoria:       "ventas",
		Concepto:        "Venta mostrador",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           1000,
		Total:           1000,
		TotalNeto:       1000,
		UsuarioCreador:  "qa_flujo",
		Estado:          "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaFinanzasMovimiento ingreso dia1: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "egreso",
		Codigo:          "EGR-2701",
		FechaMovimiento: "2026-04-01 13:00:00",
		Categoria:       "compras",
		Concepto:        "Compra insumos",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           200,
		Total:           200,
		TotalNeto:       200,
		UsuarioCreador:  "qa_flujo",
		Estado:          "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaFinanzasMovimiento egreso dia1: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "ingreso",
		Codigo:          "ING-2702",
		FechaMovimiento: "2026-04-02 10:00:00",
		Categoria:       "ventas",
		Concepto:        "Venta mostrador",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           500,
		Total:           500,
		TotalNeto:       500,
		UsuarioCreador:  "qa_flujo",
		Estado:          "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaFinanzasMovimiento ingreso dia2: %v", err)
	}

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=27&dataset=contable_flujo_caja&desde=2026-04-01&hasta=2026-04-02", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset contable_flujo_caja status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal contable_flujo_caja: %v", err)
	}
	if ds.Key != reporteDatasetContableFlujoCaja {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetContableFlujoCaja, ds.Key)
	}
	if ds.RowCount != 2 {
		t.Fatalf("row_count esperado=2 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	rowDia1 := ds.Rows[0]
	if rowDia1["fecha"].(string) != "2026-04-01" {
		t.Fatalf("fecha fila 1 esperada=2026-04-01 obtenida=%v", rowDia1["fecha"])
	}
	if rowDia1["ingresos"].(float64) != 1000 {
		t.Fatalf("ingresos dia1 esperado=1000 obtenido=%v", rowDia1["ingresos"])
	}
	if rowDia1["egresos"].(float64) != 200 {
		t.Fatalf("egresos dia1 esperado=200 obtenido=%v", rowDia1["egresos"])
	}
	if rowDia1["neto_dia"].(float64) != 800 {
		t.Fatalf("neto_dia dia1 esperado=800 obtenido=%v", rowDia1["neto_dia"])
	}
	if rowDia1["saldo_acumulado"].(float64) != 800 {
		t.Fatalf("saldo_acumulado dia1 esperado=800 obtenido=%v", rowDia1["saldo_acumulado"])
	}

	rowDia2 := ds.Rows[1]
	if rowDia2["fecha"].(string) != "2026-04-02" {
		t.Fatalf("fecha fila 2 esperada=2026-04-02 obtenida=%v", rowDia2["fecha"])
	}
	if rowDia2["saldo_acumulado"].(float64) != 1300 {
		t.Fatalf("saldo_acumulado dia2 esperado=1300 obtenido=%v", rowDia2["saldo_acumulado"])
	}

	if ds.Summary["dias"].(float64) != 2 {
		t.Fatalf("summary dias esperado=2 obtenido=%v", ds.Summary["dias"])
	}
	if ds.Summary["movimientos_total"].(float64) != 3 {
		t.Fatalf("summary movimientos_total esperado=3 obtenido=%v", ds.Summary["movimientos_total"])
	}
	if ds.Summary["total_ingresos"].(float64) != 1500 {
		t.Fatalf("summary total_ingresos esperado=1500 obtenido=%v", ds.Summary["total_ingresos"])
	}
	if ds.Summary["total_egresos"].(float64) != 200 {
		t.Fatalf("summary total_egresos esperado=200 obtenido=%v", ds.Summary["total_egresos"])
	}
	if ds.Summary["neto_periodo"].(float64) != 1300 {
		t.Fatalf("summary neto_periodo esperado=1300 obtenido=%v", ds.Summary["neto_periodo"])
	}
}

func TestEmpresaReportesHandlerDatasetContableFlujoCajaFiltros(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_flujo_caja_filtros_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(28)
	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "ingreso",
		Codigo:          "ING-2801",
		FechaMovimiento: "2026-04-03 09:00:00",
		Categoria:       "ventas",
		Concepto:        "Venta 1",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           900,
		Total:           900,
		TotalNeto:       900,
		UsuarioCreador:  "qa_flujo_filtro",
		Estado:          "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaFinanzasMovimiento ingreso ventas efectivo: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "ingreso",
		Codigo:          "ING-2802",
		FechaMovimiento: "2026-04-03 10:00:00",
		Categoria:       "servicios",
		Concepto:        "Venta 2",
		MetodoPago:      "tarjeta_credito",
		Moneda:          "COP",
		Monto:           400,
		Total:           400,
		TotalNeto:       400,
		UsuarioCreador:  "qa_flujo_filtro",
		Estado:          "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaFinanzasMovimiento ingreso servicios tarjeta: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "egreso",
		Codigo:          "EGR-2801",
		FechaMovimiento: "2026-04-03 11:00:00",
		Categoria:       "compras",
		Concepto:        "Compra 1",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           100,
		Total:           100,
		TotalNeto:       100,
		UsuarioCreador:  "qa_flujo_filtro",
		Estado:          "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaFinanzasMovimiento egreso compras efectivo: %v", err)
	}

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=28&dataset=contable_flujo_caja&desde=2026-04-03&hasta=2026-04-03&categoria=ventas&metodo_pago=efectivo", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset contable_flujo_caja filtros status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal contable_flujo_caja filtros: %v", err)
	}
	if ds.RowCount != 1 {
		t.Fatalf("row_count esperado=1 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}
	if ds.Rows[0]["ingresos"].(float64) != 900 {
		t.Fatalf("ingresos filtrados esperado=900 obtenido=%v", ds.Rows[0]["ingresos"])
	}
	if ds.Rows[0]["egresos"].(float64) != 0 {
		t.Fatalf("egresos filtrados esperado=0 obtenido=%v", ds.Rows[0]["egresos"])
	}
	if ds.Summary["total_ingresos"].(float64) != 900 {
		t.Fatalf("summary total_ingresos esperado=900 obtenido=%v", ds.Summary["total_ingresos"])
	}
	if ds.Summary["total_egresos"].(float64) != 0 {
		t.Fatalf("summary total_egresos esperado=0 obtenido=%v", ds.Summary["total_egresos"])
	}
	if strings.TrimSpace(ds.Summary["filtro_categoria"].(string)) != "ventas" {
		t.Fatalf("summary filtro_categoria esperado=ventas obtenido=%v", ds.Summary["filtro_categoria"])
	}
	if strings.TrimSpace(ds.Summary["filtro_metodo_pago"].(string)) != "efectivo" {
		t.Fatalf("summary filtro_metodo_pago esperado=efectivo obtenido=%v", ds.Summary["filtro_metodo_pago"])
	}
}

func TestEmpresaReportesHandlerDatasetReporteTurno(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_turno_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(11)

	cierreID, err := dbpkg.CreateEmpresaCierreCaja(dbEmp, dbpkg.EmpresaCierreCaja{
		EmpresaID:        empresaID,
		CajaCodigo:       "CAJA-01",
		Turno:            "manana",
		FechaOperacion:   "2026-04-05",
		FechaApertura:    "2026-04-05 08:00:00",
		FechaCierre:      "2026-04-05 12:00:00",
		EstadoCierre:     "cerrado",
		AperturaMonto:    50000,
		IngresosEfectivo: 35000,
		EgresosEfectivo:  3000,
		RetirosEfectivo:  2000,
		CajaFisica:       79000,
		UsuarioCreador:   "cajero_turno",
		CerradoPor:       "cajero_turno",
		Moneda:           "COP",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaCierreCaja: %v", err)
	}

	carritoID, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:      empresaID,
		Codigo:         "TURNO-001",
		Nombre:         "Mesa 1",
		EstadoCarrito:  "abierto",
		UsuarioCreador: "cajero_turno",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateCarritoCompra: %v", err)
	}

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:      empresaID,
		CarritoID:      carritoID,
		TipoItem:       "producto",
		Descripcion:    "Agua 600ml",
		Cantidad:       2,
		PrecioUnitario: 10000,
		UsuarioCreador: "cajero_turno",
	}); err != nil {
		t.Fatalf("CreateCarritoCompraItem producto: %v", err)
	}

	if _, err := dbpkg.CreateCarritoCompraItem(dbEmp, dbpkg.CarritoCompraItem{
		EmpresaID:      empresaID,
		CarritoID:      carritoID,
		TipoItem:       "servicio",
		Descripcion:    "Lavado premium",
		Cantidad:       1,
		PrecioUnitario: 15000,
		UsuarioCreador: "cajero_turno",
	}); err != nil {
		t.Fatalf("CreateCarritoCompraItem servicio: %v", err)
	}

	if err := dbpkg.PayCarritoStationSession(dbEmp, empresaID, carritoID, "efectivo", "RCB-001", "", "", 0, 0, 35000, 0); err != nil {
		t.Fatalf("PayCarritoStationSession: %v", err)
	}

	if _, err := dbEmp.Exec(`UPDATE carritos_compras
	SET activado_en = ?, pagado_en = ?, fecha_actualizacion = ?, usuario_creador = ?
	WHERE empresa_id = ? AND id = ?`,
		"2026-04-05 09:00:00",
		"2026-04-05 11:00:00",
		"2026-04-05 11:00:00",
		"cajero_turno",
		empresaID,
		carritoID,
	); err != nil {
		t.Fatalf("update carritos_compras fechas turno: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "egreso",
		Monto:           4000,
		Concepto:        "Compra de insumo de turno",
		Categoria:       "compras",
		MetodoPago:      "efectivo",
		FechaMovimiento: "2026-04-05 10:30:00",
		UsuarioCreador:  "cajero_turno",
		Estado:          "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaFinanzasMovimiento: %v", err)
	}

	handler := EmpresaReportesHandler(dbEmp)
	url := "/api/empresa/reportes?action=dataset&empresa_id=11&dataset=reporte_de_turno&desde=2026-04-05&hasta=2026-04-05&caja_codigo=CAJA-01&turno=manana&usuario=cajero_turno&cierre_id=" + strconv.FormatInt(cierreID, 10)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset reporte_de_turno status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal reporte_de_turno: %v", err)
	}
	if ds.Key != reporteDatasetOperativoTurno {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoTurno, ds.Key)
	}
	if ds.RowCount != 1 {
		t.Fatalf("row_count esperado=1 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	row := ds.Rows[0]
	if strings.TrimSpace(row["activado_en"].(string)) == "" {
		t.Fatalf("activado_en debe estar informado")
	}
	if row["caja_codigo"].(string) != "CAJA-01" {
		t.Fatalf("caja_codigo esperado=CAJA-01 obtenido=%v", row["caja_codigo"])
	}
	if row["turno"].(string) != "manana" {
		t.Fatalf("turno esperado=manana obtenido=%v", row["turno"])
	}

	totalProductos, _ := row["total_productos"].(float64)
	totalServicios, _ := row["total_servicios"].(float64)
	totalCarrito, _ := row["total_carrito"].(float64)
	if totalProductos <= 0 {
		t.Fatalf("total_productos debe ser mayor que cero")
	}
	if totalServicios <= 0 {
		t.Fatalf("total_servicios debe ser mayor que cero")
	}
	if totalCarrito <= 0 {
		t.Fatalf("total_carrito debe ser mayor que cero")
	}

	ventasProductos, _ := ds.Summary["ventas_productos"].(float64)
	ventasServicios, _ := ds.Summary["ventas_servicios"].(float64)
	gastosTurno, _ := ds.Summary["gastos_turno"].(float64)
	efectivoDebeHaber, _ := ds.Summary["efectivo_deberia_haber"].(float64)
	if ventasProductos <= 0 {
		t.Fatalf("summary ventas_productos debe ser mayor que cero")
	}
	if ventasServicios <= 0 {
		t.Fatalf("summary ventas_servicios debe ser mayor que cero")
	}
	if gastosTurno <= 0 {
		t.Fatalf("summary gastos_turno debe ser mayor que cero")
	}
	if efectivoDebeHaber <= 0 {
		t.Fatalf("summary efectivo_deberia_haber debe ser mayor que cero")
	}

	if ds.Summary["filtro_caja_codigo"].(string) != "CAJA-01" {
		t.Fatalf("summary filtro_caja_codigo esperado=CAJA-01 obtenido=%v", ds.Summary["filtro_caja_codigo"])
	}
}
