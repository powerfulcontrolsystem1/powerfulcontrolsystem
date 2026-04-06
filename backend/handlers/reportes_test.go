package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

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
	if err := dbpkg.EnsureEmpresaReservasHotelSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaReservasHotelSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaTarifasPorMinutosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaTarifasPorMinutosSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaTarifasPorDiaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaTarifasPorDiaSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaPropinasSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaPropinasSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaComisionesServicioSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaComisionesServicioSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaAuditoriaSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaModulosFaltantesSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaModulosFaltantesSchema: %v", err)
	}
}

func reporteDatasetFindRowByModuloKey(rows []map[string]interface{}, key string) (map[string]interface{}, bool) {
	key = strings.ToLower(strings.TrimSpace(key))
	for _, row := range rows {
		if strings.ToLower(strings.TrimSpace(row["modulo_key"].(string))) == key {
			return row, true
		}
	}
	return nil, false
}

func reporteDatasetToInt64(v interface{}) int64 {
	switch value := v.(type) {
	case int:
		return int64(value)
	case int32:
		return int64(value)
	case int64:
		return value
	case float64:
		return int64(value)
	case json.Number:
		i, _ := value.Int64()
		return i
	case string:
		parsed, _ := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
		return parsed
	default:
		return 0
	}
}

func reporteDatasetToFloat64(v interface{}) float64 {
	switch value := v.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case int32:
		return float64(value)
	case json.Number:
		f, _ := value.Float64()
		return f
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return parsed
	default:
		return 0
	}
}

func reporteDatasetFindRowByIntField(rows []map[string]interface{}, field string, expected int64) (map[string]interface{}, bool) {
	for _, row := range rows {
		if got := reporteDatasetToInt64(row[field]); got == expected {
			return row, true
		}
	}
	return nil, false
}

func reporteDatasetFindRowByStringField(rows []map[string]interface{}, field string, expected string) (map[string]interface{}, bool) {
	expected = strings.ToLower(strings.TrimSpace(expected))
	for _, row := range rows {
		value, _ := row[field].(string)
		if strings.ToLower(strings.TrimSpace(value)) == expected {
			return row, true
		}
	}
	return nil, false
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

func TestEmpresaReportesHandlerDatasetOperativoModulosResumen(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_modulos_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(71)

	if _, err := dbEmp.Exec(`
		INSERT INTO crm_leads (empresa_id, codigo, nombre, estado_lead, fecha_creacion, estado)
		VALUES (?, ?, ?, ?, ?, ?)
	`, empresaID, "LEAD-MOD-001", "Lead Abril", "nuevo", "2026-04-05 10:00:00", "activo"); err != nil {
		t.Fatalf("insert crm_leads LEAD-MOD-001: %v", err)
	}

	if _, err := dbEmp.Exec(`
		INSERT INTO crm_leads (empresa_id, codigo, nombre, estado_lead, fecha_creacion, estado)
		VALUES (?, ?, ?, ?, ?, ?)
	`, empresaID, "LEAD-MOD-002", "Lead Marzo", "nuevo", "2026-03-10 08:00:00", "activo"); err != nil {
		t.Fatalf("insert crm_leads LEAD-MOD-002: %v", err)
	}

	carritoID, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:      empresaID,
		Codigo:         "MOD-CAR-001",
		Nombre:         "Carrito módulo",
		EstadoCarrito:  "abierto",
		UsuarioCreador: "qa_modulos",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateCarritoCompra: %v", err)
	}

	if _, err := dbEmp.Exec(`
		UPDATE carritos_compras
		SET fecha_creacion = ?, fecha_actualizacion = ?, estado = ?
		WHERE empresa_id = ? AND id = ?
	`, "2026-04-06 09:30:00", "2026-04-06 09:30:00", "activo", empresaID, carritoID); err != nil {
		t.Fatalf("update carritos_compras para reporte modulos: %v", err)
	}

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=71&dataset=operativo_modulos_resumen&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_modulos_resumen status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_modulos_resumen: %v", err)
	}
	if ds.Key != reporteDatasetOperativoModulos {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoModulos, ds.Key)
	}
	if ds.RowCount == 0 {
		t.Fatalf("row_count esperado > 0, obtenido=%d", ds.RowCount)
	}

	crmRow, ok := reporteDatasetFindRowByModuloKey(ds.Rows, "crm_leads")
	if !ok {
		t.Fatalf("no se encontro fila crm_leads en dataset de modulos")
	}
	if got := reporteDatasetToInt64(crmRow["registros_totales"]); got != 2 {
		t.Fatalf("crm_leads registros_totales esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(crmRow["registros_activos"]); got != 2 {
		t.Fatalf("crm_leads registros_activos esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(crmRow["registros_rango"]); got != 1 {
		t.Fatalf("crm_leads registros_rango esperado=1 obtenido=%d", got)
	}
	if strings.TrimSpace(crmRow["ultimo_registro"].(string)) == "" {
		t.Fatalf("crm_leads ultimo_registro no debe venir vacio")
	}

	if got := reporteDatasetToInt64(ds.Summary["modulos_total"]); got != int64(ds.RowCount) {
		t.Fatalf("summary modulos_total esperado=%d obtenido=%d", ds.RowCount, got)
	}
	if got := reporteDatasetToInt64(ds.Summary["modulos_con_datos"]); got < 1 {
		t.Fatalf("summary modulos_con_datos esperado >= 1, obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["registros_totales"]); got < 3 {
		t.Fatalf("summary registros_totales esperado >= 3, obtenido=%d", got)
	}
}

func TestEmpresaReportesHandlerDatasetOperativoReservasOcupacion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_reservas_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(72)

	carritoA, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:      empresaID,
		Codigo:         "EST-72-1",
		Nombre:         "Habitacion 101",
		EstadoCarrito:  "abierto",
		UsuarioCreador: "qa_reservas",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateCarritoCompra A: %v", err)
	}
	carritoB, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:      empresaID,
		Codigo:         "EST-72-2",
		Nombre:         "Habitacion 102",
		EstadoCarrito:  "abierto",
		UsuarioCreador: "qa_reservas",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateCarritoCompra B: %v", err)
	}

	insertSQL := `INSERT INTO reservas_hotel (
		empresa_id, carrito_id, estacion_id, codigo_reserva, cliente_nombre,
		fecha_entrada, fecha_salida, monto_total, moneda, estado_reserva, estado_pago,
		cantidad_huespedes, fecha_expiracion, estado, usuario_creador, observaciones,
		fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	rows := []struct {
		CarritoID       int64
		EstacionID      int64
		Codigo          string
		Cliente         string
		Entrada         string
		Salida          string
		Monto           float64
		EstadoReserva   string
		EstadoPago      string
		Huespedes       int64
		FechaExpiracion string
	}{
		{CarritoID: carritoA, EstacionID: 1, Codigo: "RSV-72-001", Cliente: "Cliente Confirmado", Entrada: "2026-04-10 14:00:00", Salida: "2026-04-12 10:00:00", Monto: 200, EstadoReserva: "confirmada", EstadoPago: "confirmado", Huespedes: 2, FechaExpiracion: "2026-04-10 13:00:00"},
		{CarritoID: carritoA, EstacionID: 1, Codigo: "RSV-72-002", Cliente: "Cliente Pendiente", Entrada: "2026-04-15 15:00:00", Salida: "2026-04-16 09:00:00", Monto: 120, EstadoReserva: "pendiente_pago", EstadoPago: "pendiente", Huespedes: 1, FechaExpiracion: "2099-04-15 14:00:00"},
		{CarritoID: carritoB, EstacionID: 2, Codigo: "RSV-72-003", Cliente: "Cliente Cancelado", Entrada: "2026-04-18 12:00:00", Salida: "2026-04-19 11:00:00", Monto: 90, EstadoReserva: "cancelada", EstadoPago: "cancelado", Huespedes: 1, FechaExpiracion: "2026-04-18 11:00:00"},
		{CarritoID: carritoB, EstacionID: 2, Codigo: "RSV-72-004", Cliente: "Cliente Expirado", Entrada: "2026-04-22 12:00:00", Salida: "2026-04-23 11:00:00", Monto: 80, EstadoReserva: "expirada", EstadoPago: "expirado", Huespedes: 1, FechaExpiracion: "2026-04-22 11:00:00"},
	}

	for _, item := range rows {
		if _, err := dbEmp.Exec(insertSQL,
			empresaID,
			item.CarritoID,
			item.EstacionID,
			item.Codigo,
			item.Cliente,
			item.Entrada,
			item.Salida,
			item.Monto,
			"COP",
			item.EstadoReserva,
			item.EstadoPago,
			item.Huespedes,
			item.FechaExpiracion,
			"activo",
			"qa_reservas",
			"seed",
			item.Entrada,
			item.Entrada,
		); err != nil {
			t.Fatalf("insert reservas_hotel %s: %v", item.Codigo, err)
		}
	}

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=72&dataset=operativo_reservas_ocupacion&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_reservas_ocupacion status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_reservas_ocupacion: %v", err)
	}
	if ds.Key != reporteDatasetOperativoReservas {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoReservas, ds.Key)
	}
	if ds.RowCount != 2 {
		t.Fatalf("row_count esperado=2 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	rowEst1, ok := reporteDatasetFindRowByIntField(ds.Rows, "estacion_id", 1)
	if !ok {
		t.Fatalf("no se encontro fila de estacion_id=1")
	}
	if got := reporteDatasetToInt64(rowEst1["reservas_totales"]); got != 2 {
		t.Fatalf("estacion 1 reservas_totales esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowEst1["reservas_confirmadas"]); got != 1 {
		t.Fatalf("estacion 1 reservas_confirmadas esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowEst1["reservas_pendientes"]); got != 1 {
		t.Fatalf("estacion 1 reservas_pendientes esperado=1 obtenido=%d", got)
	}
	cumplimiento := reporteDatasetToFloat64(rowEst1["cumplimiento_pct"])
	if cumplimiento < 49.9 || cumplimiento > 50.1 {
		t.Fatalf("estacion 1 cumplimiento_pct esperado~50 obtenido=%v", rowEst1["cumplimiento_pct"])
	}

	if got := reporteDatasetToInt64(ds.Summary["estaciones_con_reservas"]); got != 2 {
		t.Fatalf("summary estaciones_con_reservas esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["reservas_totales"]); got != 4 {
		t.Fatalf("summary reservas_totales esperado=4 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["reservas_confirmadas"]); got != 1 {
		t.Fatalf("summary reservas_confirmadas esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["reservas_pendientes"]); got != 1 {
		t.Fatalf("summary reservas_pendientes esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["reservas_canceladas"]); got != 1 {
		t.Fatalf("summary reservas_canceladas esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["reservas_expiradas"]); got != 1 {
		t.Fatalf("summary reservas_expiradas esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["ingresos_potenciales"]); got != 490 {
		t.Fatalf("summary ingresos_potenciales esperado=490 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["ingresos_confirmados"]); got != 200 {
		t.Fatalf("summary ingresos_confirmados esperado=200 obtenido=%d", got)
	}
}

func TestEmpresaReportesHandlerDatasetOperativoTarifasIngresos(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_tarifas_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(73)

	if _, err := dbpkg.CreateEmpresaTarifaPorDia(dbEmp, dbpkg.EmpresaTarifaPorDia{
		EmpresaID:              empresaID,
		EstacionID:             1,
		EstacionCodigo:         "EST-73-1",
		EstacionNombre:         "Habitacion 101",
		ServicioNombre:         "hospedaje",
		ValorDia:               150,
		HoraCheckIn:            "15:00",
		HoraCheckOut:           "12:00",
		Moneda:                 "COP",
		Prioridad:              1,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa_tarifas",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaTarifaPorDia: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaTarifaPorMinutos(dbEmp, dbpkg.EmpresaTarifaPorMinutos{
		EmpresaID:      empresaID,
		EstacionID:     2,
		EstacionCodigo: "EST-73-2",
		EstacionNombre: "Habitacion 102",
		DiaSemanaDesde: 1,
		DiaSemanaHasta: 7,
		MinutosBase:    120,
		ValorBase:      60,
		MinutosExtra:   60,
		ValorExtra:     20,
		Moneda:         "COP",
		Prioridad:      1,
		UsuarioCreador: "qa_tarifas",
		Estado:         "activo",
	}); err != nil {
		t.Fatalf("CreateEmpresaTarifaPorMinutos: %v", err)
	}

	createVentaCerrada := func(codigo string, referenciaExterna string, total float64, pagadoEn string) {
		t.Helper()
		carritoID, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
			EmpresaID:         empresaID,
			Codigo:            codigo,
			Nombre:            codigo,
			EstadoCarrito:     "abierto",
			ReferenciaExterna: referenciaExterna,
			UsuarioCreador:    "qa_tarifas",
			Estado:            "activo",
		})
		if err != nil {
			t.Fatalf("CreateCarritoCompra %s: %v", codigo, err)
		}

		if _, err := dbEmp.Exec(`
			UPDATE carritos_compras
			SET estado_carrito = 'cerrado', total = ?, total_pagado = ?, pagado_en = ?, referencia_externa = ?, fecha_actualizacion = ?, estado = 'activo'
			WHERE empresa_id = ? AND id = ?
		`, total, total, pagadoEn, referenciaExterna, pagadoEn, empresaID, carritoID); err != nil {
			t.Fatalf("update carrito cerrado %s: %v", codigo, err)
		}
	}

	createVentaCerrada("EST-73-1", "", 300, "2026-04-11 10:00:00")
	createVentaCerrada("EST-73-2", "", 120, "2026-04-12 12:00:00")
	createVentaCerrada("VENTA-73-3", "", 80, "2026-04-13 14:00:00")

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=73&dataset=operativo_tarifas_ingresos&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_tarifas_ingresos status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_tarifas_ingresos: %v", err)
	}
	if ds.Key != reporteDatasetOperativoTarifas {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoTarifas, ds.Key)
	}
	if ds.RowCount != 3 {
		t.Fatalf("row_count esperado=3 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	rowDia, ok := reporteDatasetFindRowByStringField(ds.Rows, "modelo_tarifa", "tarifa_por_dia")
	if !ok {
		t.Fatalf("no se encontro fila tarifa_por_dia")
	}
	if got := reporteDatasetToInt64(rowDia["carritos_cerrados"]); got != 1 {
		t.Fatalf("tarifa_por_dia carritos_cerrados esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowDia["ingresos_totales"]); got != 300 {
		t.Fatalf("tarifa_por_dia ingresos_totales esperado=300 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowDia["tarifas_configuradas"]); got != 1 {
		t.Fatalf("tarifa_por_dia tarifas_configuradas esperado=1 obtenido=%d", got)
	}

	rowMin, ok := reporteDatasetFindRowByStringField(ds.Rows, "modelo_tarifa", "tarifa_por_minutos")
	if !ok {
		t.Fatalf("no se encontro fila tarifa_por_minutos")
	}
	if got := reporteDatasetToInt64(rowMin["carritos_cerrados"]); got != 1 {
		t.Fatalf("tarifa_por_minutos carritos_cerrados esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowMin["ingresos_totales"]); got != 120 {
		t.Fatalf("tarifa_por_minutos ingresos_totales esperado=120 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowMin["tarifas_configuradas"]); got != 1 {
		t.Fatalf("tarifa_por_minutos tarifas_configuradas esperado=1 obtenido=%d", got)
	}

	rowSinModelo, ok := reporteDatasetFindRowByStringField(ds.Rows, "modelo_tarifa", "sin_modelo")
	if !ok {
		t.Fatalf("no se encontro fila sin_modelo")
	}
	if got := reporteDatasetToInt64(rowSinModelo["carritos_cerrados"]); got != 1 {
		t.Fatalf("sin_modelo carritos_cerrados esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowSinModelo["ingresos_totales"]); got != 80 {
		t.Fatalf("sin_modelo ingresos_totales esperado=80 obtenido=%d", got)
	}

	if got := reporteDatasetToInt64(ds.Summary["carritos_cerrados_total"]); got != 3 {
		t.Fatalf("summary carritos_cerrados_total esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["carritos_con_modelo"]); got != 2 {
		t.Fatalf("summary carritos_con_modelo esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["carritos_sin_modelo"]); got != 1 {
		t.Fatalf("summary carritos_sin_modelo esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["ingresos_total"]); got != 500 {
		t.Fatalf("summary ingresos_total esperado=500 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["ingresos_tarifa_por_dia"]); got != 300 {
		t.Fatalf("summary ingresos_tarifa_por_dia esperado=300 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["ingresos_tarifa_por_minutos"]); got != 120 {
		t.Fatalf("summary ingresos_tarifa_por_minutos esperado=120 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["ingresos_sin_modelo"]); got != 80 {
		t.Fatalf("summary ingresos_sin_modelo esperado=80 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["tarifas_por_dia_configuradas"]); got != 1 {
		t.Fatalf("summary tarifas_por_dia_configuradas esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["tarifas_por_minutos_configuradas"]); got != 1 {
		t.Fatalf("summary tarifas_por_minutos_configuradas esperado=1 obtenido=%d", got)
	}

	cobertura := reporteDatasetToFloat64(ds.Summary["cobertura_modelo_tarifa_pct"])
	if cobertura < 66.6 || cobertura > 66.7 {
		t.Fatalf("summary cobertura_modelo_tarifa_pct esperado~66.67 obtenido=%v", ds.Summary["cobertura_modelo_tarifa_pct"])
	}
}

func TestEmpresaReportesHandlerDatasetOperativoCadenaCumplimiento(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_cadena_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(74)

	if _, err := dbEmp.Exec(`
		INSERT INTO crm_leads (empresa_id, codigo, nombre, estado_lead, valor_potencial, fecha_creacion, estado)
		VALUES (?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?)
	`,
		empresaID, "LEAD-74-001", "Lead Ganado", "ganado", 1000.0, "2026-04-10 10:00:00", "activo",
		empresaID, "LEAD-74-002", "Lead Calificado", "calificado", 500.0, "2026-04-11 11:00:00", "activo",
	); err != nil {
		t.Fatalf("insert crm_leads: %v", err)
	}

	if _, err := dbEmp.Exec(`
		INSERT INTO produccion_ordenes (empresa_id, codigo, producto_nombre, cantidad_programada, estado_orden, costo_real, fecha_programada, fecha_creacion, estado)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		empresaID, "OP-74-001", "Producto A", 10, "cerrado", 300.0, "2026-04-12", "2026-04-12 09:00:00", "activo",
		empresaID, "OP-74-002", "Producto B", 8, "en_proceso", 200.0, "2026-04-13", "2026-04-13 09:00:00", "activo",
	); err != nil {
		t.Fatalf("insert produccion_ordenes: %v", err)
	}

	if _, err := dbEmp.Exec(`
		INSERT INTO logistica_envios (empresa_id, codigo, cliente_nombre, direccion_entrega, estado_envio, costo_envio, fecha_programada, fecha_creacion, estado)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?), (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		empresaID, "ENV-74-001", "Cliente 1", "Direccion 1", "entregado", 80.0, "2026-04-14", "2026-04-14 08:00:00", "activo",
		empresaID, "ENV-74-002", "Cliente 2", "Direccion 2", "programado", 40.0, "2026-04-15", "2026-04-15 08:00:00", "activo",
	); err != nil {
		t.Fatalf("insert logistica_envios: %v", err)
	}

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=74&dataset=operativo_cadena_cumplimiento&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_cadena_cumplimiento status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_cadena_cumplimiento: %v", err)
	}
	if ds.Key != reporteDatasetOperativoCadena {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoCadena, ds.Key)
	}
	if ds.RowCount != 3 {
		t.Fatalf("row_count esperado=3 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	crmRow, ok := reporteDatasetFindRowByModuloKey(ds.Rows, "crm_leads")
	if !ok {
		t.Fatalf("no se encontro fila crm_leads")
	}
	if got := reporteDatasetToInt64(crmRow["registros_rango"]); got != 2 {
		t.Fatalf("crm_leads registros_rango esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(crmRow["finalizados"]); got != 1 {
		t.Fatalf("crm_leads finalizados esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(crmRow["en_proceso"]); got != 1 {
		t.Fatalf("crm_leads en_proceso esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(crmRow["monto_referencia"]); got != 1500 {
		t.Fatalf("crm_leads monto_referencia esperado=1500 obtenido=%d", got)
	}

	prodRow, ok := reporteDatasetFindRowByModuloKey(ds.Rows, "produccion_ordenes")
	if !ok {
		t.Fatalf("no se encontro fila produccion_ordenes")
	}
	if got := reporteDatasetToInt64(prodRow["finalizados"]); got != 1 {
		t.Fatalf("produccion_ordenes finalizados esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(prodRow["en_proceso"]); got != 1 {
		t.Fatalf("produccion_ordenes en_proceso esperado=1 obtenido=%d", got)
	}

	logRow, ok := reporteDatasetFindRowByModuloKey(ds.Rows, "logistica_envios")
	if !ok {
		t.Fatalf("no se encontro fila logistica_envios")
	}
	if got := reporteDatasetToInt64(logRow["finalizados"]); got != 1 {
		t.Fatalf("logistica_envios finalizados esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(logRow["en_proceso"]); got != 1 {
		t.Fatalf("logistica_envios en_proceso esperado=1 obtenido=%d", got)
	}

	if got := reporteDatasetToInt64(ds.Summary["registros_totales"]); got != 6 {
		t.Fatalf("summary registros_totales esperado=6 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["registros_rango"]); got != 6 {
		t.Fatalf("summary registros_rango esperado=6 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["finalizados_totales"]); got != 3 {
		t.Fatalf("summary finalizados_totales esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["en_proceso_totales"]); got != 3 {
		t.Fatalf("summary en_proceso_totales esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["monto_referencia_total"]); got != 2120 {
		t.Fatalf("summary monto_referencia_total esperado=2120 obtenido=%d", got)
	}

	if got := reporteDatasetToFloat64(ds.Summary["crm_conversion_pct"]); got < 49.9 || got > 50.1 {
		t.Fatalf("summary crm_conversion_pct esperado~50 obtenido=%v", ds.Summary["crm_conversion_pct"])
	}
	if got := reporteDatasetToFloat64(ds.Summary["produccion_cumplimiento_pct"]); got < 49.9 || got > 50.1 {
		t.Fatalf("summary produccion_cumplimiento_pct esperado~50 obtenido=%v", ds.Summary["produccion_cumplimiento_pct"])
	}
	if got := reporteDatasetToFloat64(ds.Summary["logistica_cumplimiento_pct"]); got < 49.9 || got > 50.1 {
		t.Fatalf("summary logistica_cumplimiento_pct esperado~50 obtenido=%v", ds.Summary["logistica_cumplimiento_pct"])
	}
	if got := reporteDatasetToFloat64(ds.Summary["cumplimiento_global_pct"]); got < 49.9 || got > 50.1 {
		t.Fatalf("summary cumplimiento_global_pct esperado~50 obtenido=%v", ds.Summary["cumplimiento_global_pct"])
	}
}

func TestEmpresaReportesHandlerDatasetOperativoInventarioBodega(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_inventario_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(75)

	bodegaCentralID, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{
		EmpresaID:      empresaID,
		Codigo:         "BDG-75-CEN",
		Nombre:         "Bodega Central",
		UsuarioCreador: "qa_inventario",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateBodega central: %v", err)
	}
	bodegaNorteID, err := dbpkg.CreateBodega(dbEmp, dbpkg.Bodega{
		EmpresaID:      empresaID,
		Codigo:         "BDG-75-NOR",
		Nombre:         "Bodega Norte",
		UsuarioCreador: "qa_inventario",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateBodega norte: %v", err)
	}

	cafeID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{
		EmpresaID:         empresaID,
		BodegaPrincipalID: bodegaCentralID,
		SKU:               "INV-75-CAF",
		CodigoBarras:      "INV-75-CAF-EAN",
		Nombre:            "Cafe Molido",
		Costo:             5,
		Precio:            9,
		StockMinimo:       10,
		StockMaximo:       40,
		UsuarioCreador:    "qa_inventario",
		Estado:            "activo",
	}, 18, "seed-cafe")
	if err != nil {
		t.Fatalf("CreateProducto cafe: %v", err)
	}

	azucarID, err := dbpkg.CreateProducto(dbEmp, dbpkg.Producto{
		EmpresaID:         empresaID,
		BodegaPrincipalID: bodegaNorteID,
		SKU:               "INV-75-AZU",
		CodigoBarras:      "INV-75-AZU-EAN",
		Nombre:            "Azucar",
		Costo:             2,
		Precio:            4,
		StockMinimo:       8,
		StockMaximo:       30,
		UsuarioCreador:    "qa_inventario",
		Estado:            "activo",
	}, 25, "seed-azucar")
	if err != nil {
		t.Fatalf("CreateProducto azucar: %v", err)
	}

	if err := dbpkg.RegistrarMovimientoInventario(dbEmp, empresaID, cafeID, bodegaCentralID, "salida", 12, "SAL-75-CAF", "qa_inventario", "consumo operativo"); err != nil {
		t.Fatalf("RegistrarMovimientoInventario cafe salida: %v", err)
	}
	if err := dbpkg.RegistrarMovimientoInventario(dbEmp, empresaID, azucarID, bodegaNorteID, "salida", 5, "SAL-75-AZU", "qa_inventario", "consumo operativo"); err != nil {
		t.Fatalf("RegistrarMovimientoInventario azucar salida: %v", err)
	}

	desde := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	hasta := time.Now().Format("2006-01-02")

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=75&dataset=operativo_inventario_bodega&desde="+desde+"&hasta="+hasta, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_inventario_bodega status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_inventario_bodega: %v", err)
	}
	if ds.Key != reporteDatasetOperativoInventario {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoInventario, ds.Key)
	}
	if ds.RowCount != 2 {
		t.Fatalf("row_count esperado=2 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	rowCafe, ok := reporteDatasetFindRowByStringField(ds.Rows, "producto", "Cafe Molido")
	if !ok {
		t.Fatalf("no se encontro fila de producto Cafe Molido")
	}
	if estado, _ := rowCafe["estado_stock"].(string); strings.ToLower(strings.TrimSpace(estado)) != "bajo_minimo" {
		t.Fatalf("estado_stock cafe esperado=bajo_minimo obtenido=%v", rowCafe["estado_stock"])
	}
	if estado, _ := rowCafe["estado_proyeccion"].(string); strings.ToLower(strings.TrimSpace(estado)) != "bajo_minimo" {
		t.Fatalf("estado_proyeccion cafe esperado=bajo_minimo obtenido=%v", rowCafe["estado_proyeccion"])
	}
	if got := reporteDatasetToInt64(rowCafe["valorizacion_costo"]); got != 30 {
		t.Fatalf("cafe valorizacion_costo esperado=30 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowCafe["valorizacion_venta"]); got != 54 {
		t.Fatalf("cafe valorizacion_venta esperado=54 obtenido=%d", got)
	}
	if got := reporteDatasetToFloat64(rowCafe["salida_promedio_diaria"]); got <= 0 {
		t.Fatalf("cafe salida_promedio_diaria esperado>0 obtenido=%v", rowCafe["salida_promedio_diaria"])
	}
	if got := reporteDatasetToFloat64(rowCafe["indice_rotacion_30d"]); got < 1.9 || got > 2.1 {
		t.Fatalf("cafe indice_rotacion_30d esperado~2 obtenido=%v", rowCafe["indice_rotacion_30d"])
	}

	if got := reporteDatasetToInt64(ds.Summary["sin_stock"]); got != 0 {
		t.Fatalf("summary sin_stock esperado=0 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["bajo_minimo"]); got != 1 {
		t.Fatalf("summary bajo_minimo esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["valorizacion_costo_total"]); got != 70 {
		t.Fatalf("summary valorizacion_costo_total esperado=70 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["valorizacion_venta_total"]); got != 134 {
		t.Fatalf("summary valorizacion_venta_total esperado=134 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["alertas_total"]); got != 1 {
		t.Fatalf("summary alertas_total esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["bajo_minimo_proyeccion"]); got < 1 {
		t.Fatalf("summary bajo_minimo_proyeccion esperado>=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["movimientos_salida"]); got != 2 {
		t.Fatalf("summary movimientos_salida esperado=2 obtenido=%d", got)
	}
}

func TestEmpresaReportesHandlerDatasetOperativoComprasMovimientos(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_compras_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(76)
	proveedorAID, err := dbpkg.CreateProveedor(dbEmp, dbpkg.Proveedor{
		EmpresaID:      empresaID,
		Codigo:         "PRV-76-A",
		Nombre:         "Proveedor A",
		UsuarioCreador: "qa_compras",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateProveedor A: %v", err)
	}
	proveedorBID, err := dbpkg.CreateProveedor(dbEmp, dbpkg.Proveedor{
		EmpresaID:      empresaID,
		Codigo:         "PRV-76-B",
		Nombre:         "Proveedor B",
		UsuarioCreador: "qa_compras",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateProveedor B: %v", err)
	}

	upsertDoc := func(codigo, estadoDocumento string, proveedorID int64, monto float64, fecha string) {
		t.Helper()
		if _, upsertErr := dbpkg.UpsertEmpresaDocumentoCompra(dbEmp, dbpkg.EmpresaDocumentoCompra{
			EmpresaID:       empresaID,
			ProveedorID:     proveedorID,
			TipoDocumento:   "orden_compra",
			DocumentoCodigo: codigo,
			EstadoDocumento: estadoDocumento,
			MontoTotal:      monto,
			Moneda:          "COP",
			FechaDocumento:  fecha,
			UsuarioCreador:  "qa_compras",
			Estado:          "activo",
		}); upsertErr != nil {
			t.Fatalf("UpsertEmpresaDocumentoCompra %s: %v", codigo, upsertErr)
		}
	}

	upsertDoc("OC-76-A-01", "emitida", proveedorAID, 1000, "2026-04-10")
	upsertDoc("OC-76-A-02", "recepcionada", proveedorAID, 500, "2026-04-11")
	upsertDoc("OC-76-A-03", "contabilizada", proveedorAID, 300, "2026-04-12")

	upsertDoc("OC-76-B-01", "emitida", proveedorBID, 2000, "2026-04-09")
	upsertDoc("OC-76-B-02", "recepcionada", proveedorBID, 500, "2026-04-15")
	upsertDoc("OC-76-B-03", "borrador", proveedorBID, 700, "2026-04-16")
	upsertDoc("OC-76-B-04", "contabilizada", proveedorBID, 900, "2026-03-28")

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=76&dataset=operativo_compras_movimientos&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_compras_movimientos status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_compras_movimientos: %v", err)
	}
	if ds.Key != reporteDatasetOperativoCompras {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoCompras, ds.Key)
	}
	if ds.RowCount != 2 {
		t.Fatalf("row_count esperado=2 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	rowA, ok := reporteDatasetFindRowByStringField(ds.Rows, "proveedor", "Proveedor A")
	if !ok {
		t.Fatalf("no se encontro fila de Proveedor A")
	}
	if got := reporteDatasetToInt64(rowA["ordenes_emitidas"]); got != 3 {
		t.Fatalf("Proveedor A ordenes_emitidas esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowA["recepciones"]); got != 2 {
		t.Fatalf("Proveedor A recepciones esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowA["contabilizaciones"]); got != 1 {
		t.Fatalf("Proveedor A contabilizaciones esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowA["monto_ordenado"]); got != 1800 {
		t.Fatalf("Proveedor A monto_ordenado esperado=1800 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowA["monto_recepcionado"]); got != 800 {
		t.Fatalf("Proveedor A monto_recepcionado esperado=800 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowA["brecha_monto"]); got != 1000 {
		t.Fatalf("Proveedor A brecha_monto esperado=1000 obtenido=%d", got)
	}
	if got := reporteDatasetToFloat64(rowA["cumplimiento_recepcion_pct"]); got < 66.6 || got > 66.7 {
		t.Fatalf("Proveedor A cumplimiento_recepcion_pct esperado~66.67 obtenido=%v", rowA["cumplimiento_recepcion_pct"])
	}

	rowB, ok := reporteDatasetFindRowByStringField(ds.Rows, "proveedor", "Proveedor B")
	if !ok {
		t.Fatalf("no se encontro fila de Proveedor B")
	}
	if got := reporteDatasetToInt64(rowB["documentos"]); got != 3 {
		t.Fatalf("Proveedor B documentos esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowB["ordenes_emitidas"]); got != 2 {
		t.Fatalf("Proveedor B ordenes_emitidas esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowB["recepciones"]); got != 1 {
		t.Fatalf("Proveedor B recepciones esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowB["monto_ordenado"]); got != 2500 {
		t.Fatalf("Proveedor B monto_ordenado esperado=2500 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowB["monto_recepcionado"]); got != 500 {
		t.Fatalf("Proveedor B monto_recepcionado esperado=500 obtenido=%d", got)
	}

	if got := reporteDatasetToInt64(ds.Summary["documentos"]); got != 6 {
		t.Fatalf("summary documentos esperado=6 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["ordenes_emitidas"]); got != 5 {
		t.Fatalf("summary ordenes_emitidas esperado=5 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["recepciones"]); got != 3 {
		t.Fatalf("summary recepciones esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["contabilizaciones"]); got != 1 {
		t.Fatalf("summary contabilizaciones esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["monto_ordenado"]); got != 4300 {
		t.Fatalf("summary monto_ordenado esperado=4300 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["monto_recepcionado"]); got != 1300 {
		t.Fatalf("summary monto_recepcionado esperado=1300 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["brecha_monto"]); got != 3000 {
		t.Fatalf("summary brecha_monto esperado=3000 obtenido=%d", got)
	}
	if got := reporteDatasetToFloat64(ds.Summary["cumplimiento_recepcion_pct"]); got < 59.9 || got > 60.1 {
		t.Fatalf("summary cumplimiento_recepcion_pct esperado~60 obtenido=%v", ds.Summary["cumplimiento_recepcion_pct"])
	}
	if got := reporteDatasetToFloat64(ds.Summary["cumplimiento_monto_pct"]); got < 30.2 || got > 30.3 {
		t.Fatalf("summary cumplimiento_monto_pct esperado~30.23 obtenido=%v", ds.Summary["cumplimiento_monto_pct"])
	}
}

func TestEmpresaReportesHandlerDatasetOperativoPropinasAcumulado(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_propinas_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(77)
	if _, err := dbpkg.UpsertEmpresaPropinasConfiguracion(dbEmp, dbpkg.EmpresaPropinasConfiguracion{
		EmpresaID:              empresaID,
		HabilitarPropina:       true,
		PorcentajePropina:      10,
		ModoDistribucion:       dbpkg.EmpresaPropinaModoPorUsuario,
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa_propinas",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("UpsertEmpresaPropinasConfiguracion: %v", err)
	}

	setFecha := func(id int64, fecha string) {
		t.Helper()
		if _, err := dbEmp.Exec(`
			UPDATE empresa_propinas_movimientos
			SET fecha_movimiento = ?, fecha_creacion = ?, fecha_actualizacion = ?
			WHERE empresa_id = ? AND id = ?
		`, fecha, fecha, fecha, empresaID, id); err != nil {
			t.Fatalf("update fecha propina id=%d: %v", id, err)
		}
	}

	createMov := func(mov dbpkg.EmpresaPropinaMovimiento) int64 {
		t.Helper()
		id, err := dbpkg.CreateEmpresaPropinaMovimiento(dbEmp, mov)
		if err != nil {
			t.Fatalf("CreateEmpresaPropinaMovimiento %+v: %v", mov, err)
		}
		return id
	}

	id1 := createMov(dbpkg.EmpresaPropinaMovimiento{
		EmpresaID:         empresaID,
		VentaReferencia:   "VTA-77-001",
		UsuarioOrigen:     "cajero_a",
		UsuarioAsignado:   "alice",
		ModoDistribucion:  dbpkg.EmpresaPropinaModoPorUsuario,
		Moneda:            "COP",
		BaseCobro:         1000,
		PorcentajePropina: 10,
		MontoPropina:      100,
		UsuarioCreador:    "qa_propinas",
		Estado:            "activo",
	})
	id2 := createMov(dbpkg.EmpresaPropinaMovimiento{
		EmpresaID:         empresaID,
		VentaReferencia:   "VTA-77-002",
		UsuarioOrigen:     "cajero_a",
		UsuarioAsignado:   "bob",
		ModoDistribucion:  dbpkg.EmpresaPropinaModoPorUsuario,
		Moneda:            "COP",
		BaseCobro:         500,
		PorcentajePropina: 10,
		MontoPropina:      50,
		UsuarioCreador:    "qa_propinas",
		Estado:            "activo",
	})
	id3 := createMov(dbpkg.EmpresaPropinaMovimiento{
		EmpresaID:         empresaID,
		VentaReferencia:   "VTA-77-003",
		UsuarioOrigen:     "cajero_a",
		ModoDistribucion:  dbpkg.EmpresaPropinaModoUniversal,
		Moneda:            "COP",
		BaseCobro:         250,
		PorcentajePropina: 10,
		MontoPropina:      25,
		UsuarioCreador:    "qa_propinas",
		Estado:            "activo",
	})
	id4 := createMov(dbpkg.EmpresaPropinaMovimiento{
		EmpresaID:         empresaID,
		VentaReferencia:   "VTA-77-004",
		UsuarioOrigen:     "cajero_a",
		UsuarioAsignado:   "alice",
		ModoDistribucion:  dbpkg.EmpresaPropinaModoPorUsuario,
		Moneda:            "COP",
		BaseCobro:         400,
		PorcentajePropina: 10,
		MontoPropina:      40,
		UsuarioCreador:    "qa_propinas",
		Estado:            "activo",
	})

	setFecha(id1, "2026-04-05 10:00:00")
	setFecha(id2, "2026-04-06 11:00:00")
	setFecha(id3, "2026-04-07 12:00:00")
	setFecha(id4, "2026-03-25 09:00:00")

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=77&dataset=operativo_propinas_acumulado&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_propinas_acumulado status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_propinas_acumulado: %v", err)
	}
	if ds.Key != reporteDatasetOperativoPropinas {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoPropinas, ds.Key)
	}
	if ds.RowCount != 2 {
		t.Fatalf("row_count esperado=2 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	rowAlice, ok := reporteDatasetFindRowByStringField(ds.Rows, "usuario_clave", "alice")
	if !ok {
		t.Fatalf("no se encontro fila para usuario_clave=alice")
	}
	if got := reporteDatasetToInt64(rowAlice["movimientos"]); got != 1 {
		t.Fatalf("alice movimientos esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowAlice["propina_total"]); got != 100 {
		t.Fatalf("alice propina_total esperado=100 obtenido=%d", got)
	}
	if got := reporteDatasetToFloat64(rowAlice["participacion_pct"]); got < 57.1 || got > 57.2 {
		t.Fatalf("alice participacion_pct esperado~57.14 obtenido=%v", rowAlice["participacion_pct"])
	}

	if got := reporteDatasetToInt64(ds.Summary["movimientos_total"]); got != 3 {
		t.Fatalf("summary movimientos_total esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["total_propinas"]); got != 175 {
		t.Fatalf("summary total_propinas esperado=175 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["total_propinas_por_usuario"]); got != 150 {
		t.Fatalf("summary total_propinas_por_usuario esperado=150 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["total_propinas_universal"]); got != 25 {
		t.Fatalf("summary total_propinas_universal esperado=25 obtenido=%d", got)
	}
	if got := strings.ToLower(strings.TrimSpace(ds.Summary["usuario_top"].(string))); got != "alice" {
		t.Fatalf("summary usuario_top esperado=alice obtenido=%s", got)
	}
}

func TestEmpresaReportesHandlerDatasetOperativoComisionesLavador(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_comisiones_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(78)
	if _, err := dbpkg.UpsertEmpresaComisionesServicioConfiguracion(dbEmp, dbpkg.EmpresaComisionesServicioConfiguracion{
		EmpresaID:              empresaID,
		HabilitarComisiones:    true,
		PorcentajeComision:     10,
		FiltroServicio:         "lavado",
		AplicarAutomaticamente: true,
		UsuarioCreador:         "qa_comisiones",
		Estado:                 "activo",
	}); err != nil {
		t.Fatalf("UpsertEmpresaComisionesServicioConfiguracion: %v", err)
	}

	setFecha := func(id int64, fecha string) {
		t.Helper()
		if _, err := dbEmp.Exec(`
			UPDATE empresa_comisiones_servicio_movimientos
			SET fecha_movimiento = ?, fecha_creacion = ?, fecha_actualizacion = ?
			WHERE empresa_id = ? AND id = ?
		`, fecha, fecha, fecha, empresaID, id); err != nil {
			t.Fatalf("update fecha comision id=%d: %v", id, err)
		}
	}

	createMov := func(mov dbpkg.EmpresaComisionServicioMovimiento) int64 {
		t.Helper()
		id, err := dbpkg.CreateEmpresaComisionServicioMovimiento(dbEmp, mov)
		if err != nil {
			t.Fatalf("CreateEmpresaComisionServicioMovimiento %+v: %v", mov, err)
		}
		return id
	}

	id1 := createMov(dbpkg.EmpresaComisionServicioMovimiento{
		EmpresaID:          empresaID,
		CarritoItemID:      7801,
		ServicioCodigo:     "LAV-A",
		ServicioNombre:     "Lavado A",
		ServicioCategoria:  "lavado",
		UsuarioOrigen:      "cajero",
		UsuarioLavador:     "lavador_a",
		VentaReferencia:    "VTA-78-001",
		Moneda:             "COP",
		BaseServicio:       300,
		PorcentajeComision: 10,
		MontoComision:      30,
		UsuarioCreador:     "qa_comisiones",
		Estado:             "activo",
	})
	id2 := createMov(dbpkg.EmpresaComisionServicioMovimiento{
		EmpresaID:          empresaID,
		CarritoItemID:      7802,
		ServicioCodigo:     "LAV-B",
		ServicioNombre:     "Lavado B",
		ServicioCategoria:  "lavado",
		UsuarioOrigen:      "cajero",
		UsuarioLavador:     "lavador_a",
		VentaReferencia:    "VTA-78-002",
		Moneda:             "COP",
		BaseServicio:       200,
		PorcentajeComision: 10,
		MontoComision:      20,
		UsuarioCreador:     "qa_comisiones",
		Estado:             "activo",
	})
	id3 := createMov(dbpkg.EmpresaComisionServicioMovimiento{
		EmpresaID:          empresaID,
		CarritoItemID:      7803,
		ServicioCodigo:     "LAV-C",
		ServicioNombre:     "Lavado C",
		ServicioCategoria:  "lavado",
		UsuarioOrigen:      "cajero",
		UsuarioLavador:     "lavador_b",
		VentaReferencia:    "VTA-78-003",
		Moneda:             "COP",
		BaseServicio:       100,
		PorcentajeComision: 10,
		MontoComision:      10,
		UsuarioCreador:     "qa_comisiones",
		Estado:             "activo",
	})
	id4 := createMov(dbpkg.EmpresaComisionServicioMovimiento{
		EmpresaID:          empresaID,
		CarritoItemID:      7804,
		ServicioCodigo:     "LAV-D",
		ServicioNombre:     "Lavado D",
		ServicioCategoria:  "lavado",
		UsuarioOrigen:      "cajero",
		UsuarioLavador:     "lavador_b",
		VentaReferencia:    "VTA-78-004",
		Moneda:             "COP",
		BaseServicio:       90,
		PorcentajeComision: 10,
		MontoComision:      9,
		UsuarioCreador:     "qa_comisiones",
		Estado:             "activo",
	})

	setFecha(id1, "2026-04-02 08:00:00")
	setFecha(id2, "2026-04-03 09:00:00")
	setFecha(id3, "2026-04-04 10:00:00")
	setFecha(id4, "2026-03-20 10:00:00")

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=78&dataset=operativo_comisiones_lavador&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_comisiones_lavador status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_comisiones_lavador: %v", err)
	}
	if ds.Key != reporteDatasetOperativoComisiones {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoComisiones, ds.Key)
	}
	if ds.RowCount != 2 {
		t.Fatalf("row_count esperado=2 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	rowLavadorA, ok := reporteDatasetFindRowByStringField(ds.Rows, "usuario_lavador", "lavador_a")
	if !ok {
		t.Fatalf("no se encontro fila para usuario_lavador=lavador_a")
	}
	if got := reporteDatasetToInt64(rowLavadorA["movimientos"]); got != 2 {
		t.Fatalf("lavador_a movimientos esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowLavadorA["base_servicios"]); got != 500 {
		t.Fatalf("lavador_a base_servicios esperado=500 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowLavadorA["monto_comision"]); got != 50 {
		t.Fatalf("lavador_a monto_comision esperado=50 obtenido=%d", got)
	}
	if got := reporteDatasetToFloat64(rowLavadorA["participacion_pct"]); got < 83.3 || got > 83.4 {
		t.Fatalf("lavador_a participacion_pct esperado~83.33 obtenido=%v", rowLavadorA["participacion_pct"])
	}

	if got := reporteDatasetToInt64(ds.Summary["movimientos_total"]); got != 3 {
		t.Fatalf("summary movimientos_total esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["total_base_servicios"]); got != 600 {
		t.Fatalf("summary total_base_servicios esperado=600 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["total_comisiones"]); got != 60 {
		t.Fatalf("summary total_comisiones esperado=60 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["lavadores_con_comision"]); got != 2 {
		t.Fatalf("summary lavadores_con_comision esperado=2 obtenido=%d", got)
	}
	if got := strings.ToLower(strings.TrimSpace(ds.Summary["lavador_top"].(string))); got != "lavador_a" {
		t.Fatalf("summary lavador_top esperado=lavador_a obtenido=%s", got)
	}
}

func TestEmpresaReportesHandlerDatasetOperativoFacturacionTrazabilidad(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_facturacion_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(79)

	upsertDoc := func(payload dbpkg.EmpresaDocumentoFacturacion) {
		t.Helper()
		if _, err := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, payload); err != nil {
			t.Fatalf("UpsertEmpresaDocumentoFacturacion %+v: %v", payload, err)
		}
	}

	upsertDoc(dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:        empresaID,
		TipoDocumento:    "factura_electronica",
		DocumentoCodigo:  "FAC-79-001",
		EstadoDocumento:  "emitida",
		EventoUltimo:     "emitir",
		MontoTotal:       1000,
		Moneda:           "COP",
		FechaDocumento:   "2026-04-05",
		NumeroLegal:      "FE-79-001",
		CodigoValidacion: "CV-79-001",
		UsuarioCreador:   "qa_facturacion",
		Estado:           "activo",
	})
	upsertDoc(dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:        empresaID,
		TipoDocumento:    "factura_electronica",
		DocumentoCodigo:  "FAC-79-002",
		EstadoDocumento:  "anulada",
		EventoUltimo:     "anular",
		MontoTotal:       800,
		Moneda:           "COP",
		FechaDocumento:   "2026-04-07",
		NumeroLegal:      "FE-79-002",
		CodigoValidacion: "CV-79-002",
		UsuarioCreador:   "qa_facturacion",
		Estado:           "activo",
	})
	upsertDoc(dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:       empresaID,
		TipoDocumento:   "factura_electronica",
		DocumentoCodigo: "FAC-79-003",
		EstadoDocumento: "borrador",
		EventoUltimo:    "crear",
		MontoTotal:      600,
		Moneda:          "COP",
		FechaDocumento:  "2026-04-09",
		UsuarioCreador:  "qa_facturacion",
		Estado:          "activo",
	})
	upsertDoc(dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:        empresaID,
		TipoDocumento:    "nota_credito",
		DocumentoCodigo:  "NC-79-001",
		EstadoDocumento:  "emitida",
		EventoUltimo:     "nota_credito_emitida",
		MontoTotal:       200,
		Moneda:           "COP",
		FechaDocumento:   "2026-04-10",
		NumeroLegal:      "NC-79-001",
		CodigoValidacion: "NCV-79-001",
		UsuarioCreador:   "qa_facturacion",
		Estado:           "activo",
	})
	upsertDoc(dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:        empresaID,
		TipoDocumento:    "factura_electronica",
		DocumentoCodigo:  "FAC-79-OLD",
		EstadoDocumento:  "emitida",
		EventoUltimo:     "emitir",
		MontoTotal:       500,
		Moneda:           "COP",
		FechaDocumento:   "2026-03-25",
		NumeroLegal:      "FE-79-OLD",
		CodigoValidacion: "CV-79-OLD",
		UsuarioCreador:   "qa_facturacion",
		Estado:           "activo",
	})

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=79&dataset=operativo_facturacion_trazabilidad&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_facturacion_trazabilidad status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_facturacion_trazabilidad: %v", err)
	}
	if ds.Key != reporteDatasetOperativoFacturacion {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoFacturacion, ds.Key)
	}
	if ds.RowCount != 2 {
		t.Fatalf("row_count esperado=2 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	rowFactura, ok := reporteDatasetFindRowByStringField(ds.Rows, "tipo_documento", "factura_electronica")
	if !ok {
		t.Fatalf("no se encontro fila para tipo_documento=factura_electronica")
	}
	if got := reporteDatasetToInt64(rowFactura["documentos"]); got != 3 {
		t.Fatalf("factura_electronica documentos esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowFactura["emitidas"]); got != 1 {
		t.Fatalf("factura_electronica emitidas esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowFactura["anuladas"]); got != 1 {
		t.Fatalf("factura_electronica anuladas esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowFactura["pendientes"]); got != 1 {
		t.Fatalf("factura_electronica pendientes esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowFactura["con_trazabilidad"]); got != 2 {
		t.Fatalf("factura_electronica con_trazabilidad esperado=2 obtenido=%d", got)
	}

	rowNC, ok := reporteDatasetFindRowByStringField(ds.Rows, "tipo_documento", "nota_credito")
	if !ok {
		t.Fatalf("no se encontro fila para tipo_documento=nota_credito")
	}
	if got := reporteDatasetToInt64(rowNC["notas_credito"]); got != 1 {
		t.Fatalf("nota_credito notas_credito esperado=1 obtenido=%d", got)
	}

	if got := reporteDatasetToInt64(ds.Summary["documentos_total"]); got != 4 {
		t.Fatalf("summary documentos_total esperado=4 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["documentos_emitidos"]); got != 2 {
		t.Fatalf("summary documentos_emitidos esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["documentos_anulados"]); got != 1 {
		t.Fatalf("summary documentos_anulados esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["documentos_pendientes"]); got != 1 {
		t.Fatalf("summary documentos_pendientes esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["notas_credito"]); got != 1 {
		t.Fatalf("summary notas_credito esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["documentos_trazables"]); got != 3 {
		t.Fatalf("summary documentos_trazables esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToFloat64(ds.Summary["trazabilidad_pct"]); got < 74.9 || got > 75.1 {
		t.Fatalf("summary trazabilidad_pct esperado~75 obtenido=%v", ds.Summary["trazabilidad_pct"])
	}
	if got := reporteDatasetToInt64(ds.Summary["monto_total"]); got != 2600 {
		t.Fatalf("summary monto_total esperado=2600 obtenido=%d", got)
	}
}

func TestEmpresaReportesHandlerDatasetOperativoAuditoriaAcciones(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_auditoria_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	empresaID := int64(80)
	createEvento := func(in dbpkg.EmpresaAuditoriaEvento) {
		t.Helper()
		if _, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, in); err != nil {
			t.Fatalf("CreateEmpresaAuditoriaEvento %+v: %v", in, err)
		}
	}

	createEvento(dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "clientes",
		Accion:         "crear",
		Resultado:      "ok",
		CodigoHTTP:     201,
		FechaEvento:    "2026-04-05 09:00:00",
		UsuarioCreador: "admin_auditoria",
		Estado:         "activo",
	})
	createEvento(dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "clientes",
		Accion:         "eliminar",
		Resultado:      "error",
		CodigoHTTP:     500,
		FechaEvento:    "2026-04-06 10:00:00",
		UsuarioCreador: "admin_auditoria",
		RequestID:      "REQ-80-001",
		Recurso:        "cliente",
		RecursoID:      1001,
		Estado:         "activo",
	})
	createEvento(dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "clientes",
		Accion:         "anular",
		Resultado:      "error",
		CodigoHTTP:     409,
		FechaEvento:    "2026-04-07 10:00:00",
		UsuarioCreador: "admin_auditoria",
		RequestID:      "REQ-80-002",
		Recurso:        "factura",
		RecursoID:      1002,
		Estado:         "activo",
	})
	createEvento(dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "inventario",
		Accion:         "ajustar",
		Resultado:      "ok",
		CodigoHTTP:     200,
		FechaEvento:    "2026-04-08 10:00:00",
		UsuarioCreador: "operador_auditoria",
		RequestID:      "REQ-80-003",
		Estado:         "activo",
	})
	createEvento(dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "inventario",
		Accion:         "ajustar",
		Resultado:      "error",
		CodigoHTTP:     500,
		FechaEvento:    "2026-03-20 10:00:00",
		UsuarioCreador: "operador_auditoria",
		RequestID:      "REQ-80-OLD",
		Estado:         "activo",
	})

	handler := EmpresaReportesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=dataset&empresa_id=80&dataset=operativo_auditoria_acciones&desde=2026-04-01&hasta=2026-04-30", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("dataset operativo_auditoria_acciones status=%d body=%s", rr.Code, rr.Body.String())
	}

	var ds empresaReporteDataset
	if err := json.Unmarshal(rr.Body.Bytes(), &ds); err != nil {
		t.Fatalf("unmarshal operativo_auditoria_acciones: %v", err)
	}
	if ds.Key != reporteDatasetOperativoAuditoria {
		t.Fatalf("dataset key esperado=%s obtenido=%s", reporteDatasetOperativoAuditoria, ds.Key)
	}
	if ds.RowCount != 2 {
		t.Fatalf("row_count esperado=2 obtenido=%d body=%s", ds.RowCount, rr.Body.String())
	}

	var rowClientes map[string]interface{}
	for _, row := range ds.Rows {
		modulo, _ := row["modulo"].(string)
		usuario, _ := row["usuario"].(string)
		if strings.EqualFold(strings.TrimSpace(modulo), "clientes") && strings.EqualFold(strings.TrimSpace(usuario), "admin_auditoria") {
			rowClientes = row
			break
		}
	}
	if rowClientes == nil {
		t.Fatalf("no se encontro fila para modulo=clientes y usuario=admin_auditoria")
	}
	if got := reporteDatasetToInt64(rowClientes["eventos"]); got != 3 {
		t.Fatalf("clientes/admin eventos esperado=3 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowClientes["errores"]); got != 2 {
		t.Fatalf("clientes/admin errores esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowClientes["http_4xx"]); got != 1 {
		t.Fatalf("clientes/admin http_4xx esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(rowClientes["http_5xx"]); got != 1 {
		t.Fatalf("clientes/admin http_5xx esperado=1 obtenido=%d", got)
	}

	if got := reporteDatasetToInt64(ds.Summary["eventos_total"]); got != 4 {
		t.Fatalf("summary eventos_total esperado=4 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["modulos_total"]); got != 2 {
		t.Fatalf("summary modulos_total esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["usuarios_total"]); got != 2 {
		t.Fatalf("summary usuarios_total esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["parejas_modulo_usuario"]); got != 2 {
		t.Fatalf("summary parejas_modulo_usuario esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["errores_total"]); got != 2 {
		t.Fatalf("summary errores_total esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["http_4xx_total"]); got != 1 {
		t.Fatalf("summary http_4xx_total esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["http_5xx_total"]); got != 1 {
		t.Fatalf("summary http_5xx_total esperado=1 obtenido=%d", got)
	}
	if got := reporteDatasetToInt64(ds.Summary["acciones_criticas_total"]); got != 2 {
		t.Fatalf("summary acciones_criticas_total esperado=2 obtenido=%d", got)
	}
	if got := reporteDatasetToFloat64(ds.Summary["error_global_pct"]); got < 49.9 || got > 50.1 {
		t.Fatalf("summary error_global_pct esperado~50 obtenido=%v", ds.Summary["error_global_pct"])
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
