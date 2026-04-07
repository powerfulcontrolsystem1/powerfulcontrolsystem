package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaCreditosHandlerFlujoBasico(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_creditos_handler.db")
	if err := dbpkg.EnsureEmpresaCreditosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaCreditosSchema: %v", err)
	}

	h := EmpresaCreditosHandler(dbEmp)

	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?empresa_id=88", strings.NewReader(`{
		"empresa_id":88,
		"cliente_nombre":"Cliente Credito Handler",
		"tipo_credito":"cuotas",
		"monto_aprobado":1800,
		"cupo_credito":2000,
		"tasa_interes":10,
		"plazo_cuotas":6,
		"fecha_inicio":"2026-04-07",
		"fecha_vencimiento":"2026-10-07"
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createRR.Code, createRR.Body.String())
	}

	var createResp struct {
		Credito dbpkg.EmpresaCredito `json:"credito"`
	}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Credito.ID <= 0 {
		t.Fatalf("expected credito id > 0, got %+v", createResp)
	}

	estadoCuentaReq := httptest.NewRequest(http.MethodGet, "/api/empresa/creditos?action=estado_cuenta&empresa_id=88&id="+itoa64(createResp.Credito.ID), nil)
	estadoCuentaRR := httptest.NewRecorder()
	h.ServeHTTP(estadoCuentaRR, estadoCuentaReq)
	if estadoCuentaRR.Code != http.StatusOK {
		t.Fatalf("estado_cuenta status=%d body=%s", estadoCuentaRR.Code, estadoCuentaRR.Body.String())
	}

	abonoReq := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?action=abono&empresa_id=88", strings.NewReader(`{
		"empresa_id":88,
		"credito_id":`+itoa64(createResp.Credito.ID)+`,
		"monto":500,
		"metodo_pago":"transferencia_bancaria",
		"referencia_pago":"AB-500"
	}`))
	abonoReq.Header.Set("Content-Type", "application/json")
	abonoRR := httptest.NewRecorder()
	h.ServeHTTP(abonoRR, abonoReq)
	if abonoRR.Code != http.StatusOK {
		t.Fatalf("abono status=%d body=%s", abonoRR.Code, abonoRR.Body.String())
	}

	var abonoResp struct {
		Credito dbpkg.EmpresaCredito `json:"credito"`
	}
	if err := json.Unmarshal(abonoRR.Body.Bytes(), &abonoResp); err != nil {
		t.Fatalf("decode abono response: %v", err)
	}
	if abonoResp.Credito.SaldoActual >= createResp.Credito.SaldoActual {
		t.Fatalf("expected saldo reduced after abono, before=%.2f after=%.2f", createResp.Credito.SaldoActual, abonoResp.Credito.SaldoActual)
	}

	resumenReq := httptest.NewRequest(http.MethodGet, "/api/empresa/creditos?action=resumen_cartera&empresa_id=88", nil)
	resumenRR := httptest.NewRecorder()
	h.ServeHTTP(resumenRR, resumenReq)
	if resumenRR.Code != http.StatusOK {
		t.Fatalf("resumen status=%d body=%s", resumenRR.Code, resumenRR.Body.String())
	}

	reporteReq := httptest.NewRequest(http.MethodGet, "/api/empresa/creditos?action=reporte&empresa_id=88&format=csv", nil)
	reporteRR := httptest.NewRecorder()
	h.ServeHTTP(reporteRR, reporteReq)
	if reporteRR.Code != http.StatusOK {
		t.Fatalf("reporte status=%d body=%s", reporteRR.Code, reporteRR.Body.String())
	}
	if !strings.Contains(strings.ToLower(reporteRR.Header().Get("Content-Type")), "text/csv") {
		t.Fatalf("expected csv content type, got=%s", reporteRR.Header().Get("Content-Type"))
	}

	estadoReq := httptest.NewRequest(http.MethodPut, "/api/empresa/creditos?action=estado&empresa_id=88&id="+itoa64(createResp.Credito.ID), strings.NewReader(`{"estado_credito":"suspendido"}`))
	estadoReq.Header.Set("Content-Type", "application/json")
	estadoRR := httptest.NewRecorder()
	h.ServeHTTP(estadoRR, estadoReq)
	if estadoRR.Code != http.StatusOK {
		t.Fatalf("estado status=%d body=%s", estadoRR.Code, estadoRR.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/creditos?empresa_id=88&id="+itoa64(createResp.Credito.ID), nil)
	deleteRR := httptest.NewRecorder()
	h.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete status=%d body=%s", deleteRR.Code, deleteRR.Body.String())
	}
}

func TestEmpresaCreditosHandlerAlertasMoraYReporte(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_creditos_handler_alertas.db")
	if err := dbpkg.EnsureEmpresaCreditosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaCreditosSchema: %v", err)
	}

	today := time.Now()
	if _, err := dbpkg.CreateEmpresaCredito(dbEmp, dbpkg.EmpresaCredito{
		EmpresaID:        89,
		ClienteNombre:    "Cliente Proximo Handler",
		TipoCredito:      "cuotas",
		MontoAprobado:    450,
		CupoCredito:      500,
		PlazoCuotas:      2,
		FechaInicio:      today.AddDate(0, 0, -3).Format("2006-01-02"),
		FechaVencimiento: today.AddDate(0, 0, 2).Format("2006-01-02"),
		EstadoCredito:    "activo",
		Estado:           "activo",
	}); err != nil {
		t.Fatalf("create credito proximo: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaCredito(dbEmp, dbpkg.EmpresaCredito{
		EmpresaID:        89,
		ClienteNombre:    "Cliente Vencido Handler",
		TipoCredito:      "cuotas",
		MontoAprobado:    900,
		CupoCredito:      1000,
		PlazoCuotas:      4,
		FechaInicio:      today.AddDate(0, -1, 0).Format("2006-01-02"),
		FechaVencimiento: today.AddDate(0, 0, -12).Format("2006-01-02"),
		EstadoCredito:    "activo",
		Estado:           "activo",
	}); err != nil {
		t.Fatalf("create credito vencido: %v", err)
	}

	h := EmpresaCreditosHandler(dbEmp)

	alertReq := httptest.NewRequest(http.MethodGet, "/api/empresa/creditos?action=alertas_mora&empresa_id=89&dias_proximos=7&top=5", nil)
	alertRR := httptest.NewRecorder()
	h.ServeHTTP(alertRR, alertReq)
	if alertRR.Code != http.StatusOK {
		t.Fatalf("alertas_mora status=%d body=%s", alertRR.Code, alertRR.Body.String())
	}
	var alertResp struct {
		Alertas dbpkg.EmpresaCreditosMoraDashboard `json:"alertas"`
	}
	if err := json.Unmarshal(alertRR.Body.Bytes(), &alertResp); err != nil {
		t.Fatalf("decode alertas response: %v", err)
	}
	if alertResp.Alertas.TotalProximosVencer < 1 {
		t.Fatalf("expected proximos >= 1, got %+v", alertResp.Alertas)
	}
	if alertResp.Alertas.TotalVencidos < 1 {
		t.Fatalf("expected vencidos >= 1, got %+v", alertResp.Alertas)
	}

	rankingReq := httptest.NewRequest(http.MethodGet, "/api/empresa/creditos?action=ranking_morosidad&empresa_id=89&dias_proximos=7&top=5", nil)
	rankingRR := httptest.NewRecorder()
	h.ServeHTTP(rankingRR, rankingReq)
	if rankingRR.Code != http.StatusOK {
		t.Fatalf("ranking_morosidad status=%d body=%s", rankingRR.Code, rankingRR.Body.String())
	}

	reporteReq := httptest.NewRequest(http.MethodGet, "/api/empresa/creditos?action=reporte&empresa_id=89&tipo=morosidad&dias_proximos=7&top=5&format=csv", nil)
	reporteRR := httptest.NewRecorder()
	h.ServeHTTP(reporteRR, reporteReq)
	if reporteRR.Code != http.StatusOK {
		t.Fatalf("reporte morosidad status=%d body=%s", reporteRR.Code, reporteRR.Body.String())
	}
	if !strings.Contains(strings.ToLower(reporteRR.Header().Get("Content-Type")), "text/csv") {
		t.Fatalf("expected morosidad csv content type, got=%s", reporteRR.Header().Get("Content-Type"))
	}
	body := strings.ToLower(reporteRR.Body.String())
	if !strings.Contains(body, "grupo,id,codigo") {
		t.Fatalf("expected morosidad report columns in csv, body=%s", reporteRR.Body.String())
	}
}

func TestEmpresaCreditosHandlerAbonoIntegraContabilidadYAsientos(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_creditos_handler_contable.db")
	if err := dbpkg.EnsureEmpresaCreditosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaCreditosSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaFinanzasSchema: %v", err)
	}

	creditoID, err := dbpkg.CreateEmpresaCredito(dbEmp, dbpkg.EmpresaCredito{
		EmpresaID:        90,
		ClienteNombre:    "Cliente Contable Handler",
		TipoCredito:      "cuotas",
		MontoAprobado:    1200,
		CupoCredito:      1200,
		PlazoCuotas:      3,
		FechaInicio:      time.Now().AddDate(0, 0, -10).Format("2006-01-02"),
		FechaVencimiento: time.Now().AddDate(0, 0, 20).Format("2006-01-02"),
		EstadoCredito:    "activo",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaCredito: %v", err)
	}

	h := EmpresaCreditosHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?action=abono&empresa_id=90&procesar_asientos=true&asientos_limit=20", strings.NewReader(`{
		"empresa_id":90,
		"credito_id":`+itoa64(creditoID)+`,
		"monto":400,
		"metodo_pago":"transferencia_bancaria",
		"referencia_pago":"TRX-400"
	}`))
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "creditos@test.com"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("abono status=%d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		MovimientoID        int64                  `json:"movimiento_id"`
		IntegracionContable map[string]interface{} `json:"integracion_contable"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode abono contable response: %v", err)
	}
	if resp.MovimientoID <= 0 {
		t.Fatalf("expected movimiento_id > 0, got %+v", resp)
	}
	if registrado, _ := resp.IntegracionContable["evento_registrado"].(bool); !registrado {
		t.Fatalf("expected evento_registrado=true, resp=%+v", resp.IntegracionContable)
	}
	if procesados, _ := resp.IntegracionContable["asientos_procesados"].(bool); !procesados {
		t.Fatalf("expected asientos_procesados=true, resp=%+v", resp.IntegracionContable)
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 90, dbpkg.EmpresaEventoContableFilter{Modulo: "creditos", Limit: 20})
	if err != nil {
		t.Fatalf("ListEmpresaEventosContables creditos: %v", err)
	}
	if len(eventos) == 0 {
		t.Fatalf("expected at least one evento contable de creditos")
	}
	encontroEvento := false
	for _, ev := range eventos {
		if ev.Evento == "credito_abono_registrado" && ev.EntidadID == resp.MovimientoID {
			encontroEvento = true
			break
		}
	}
	if !encontroEvento {
		t.Fatalf("expected credito_abono_registrado linked to movimiento_id=%d", resp.MovimientoID)
	}

	asientos, err := dbpkg.ListEmpresaAsientosContables(dbEmp, 90, dbpkg.EmpresaAsientoContableFilter{Modulo: "creditos", Evento: "credito_abono_registrado", Limit: 20})
	if err != nil {
		t.Fatalf("ListEmpresaAsientosContables creditos: %v", err)
	}
	if len(asientos) == 0 {
		t.Fatalf("expected at least one asiento contable de creditos")
	}
	if asientos[0].TotalDebito <= 0 || asientos[0].TotalCredito <= 0 {
		t.Fatalf("expected asiento con debito y credito positivos, got %+v", asientos[0])
	}
}

func TestEmpresaCreditosHandlerWorkflowReversoSolicitudYAprobacion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_creditos_handler_workflow_reverso.db")
	if err := dbpkg.EnsureEmpresaCreditosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaCreditosSchema: %v", err)
	}

	creditoID, err := dbpkg.CreateEmpresaCredito(dbEmp, dbpkg.EmpresaCredito{
		EmpresaID:        91,
		ClienteNombre:    "Cliente Workflow Handler",
		TipoCredito:      "cuotas",
		MontoAprobado:    900,
		CupoCredito:      900,
		PlazoCuotas:      3,
		FechaInicio:      "2026-04-01",
		FechaVencimiento: "2026-07-01",
		EstadoCredito:    "activo",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaCredito: %v", err)
	}

	movAbonoID, _, err := dbpkg.RegisterEmpresaCreditoAbono(dbEmp, dbpkg.EmpresaCreditoAbonoInput{
		EmpresaID:      91,
		CreditoID:      creditoID,
		Monto:          300,
		MetodoPago:     "efectivo",
		ReferenciaPago: "ABN-WF-001",
		UsuarioCreador: "caja@empresa.com",
	})
	if err != nil {
		t.Fatalf("RegisterEmpresaCreditoAbono: %v", err)
	}

	h := EmpresaCreditosHandler(dbEmp)

	solicitarReq := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?action=solicitar_reverso&empresa_id=91", strings.NewReader(`{
		"empresa_id":91,
		"credito_id":`+itoa64(creditoID)+`,
		"movimiento_origen_id":`+itoa64(movAbonoID)+`,
		"motivo_solicitud":"Abono registrado en caja equivocada",
		"nivel_aprobacion_requerido":1
	}`))
	solicitarReq.Header.Set("Content-Type", "application/json")
	solicitarReq = solicitarReq.WithContext(context.WithValue(solicitarReq.Context(), "adminEmail", "analista@empresa.com"))
	solicitarRR := httptest.NewRecorder()
	h.ServeHTTP(solicitarRR, solicitarReq)
	if solicitarRR.Code != http.StatusCreated {
		t.Fatalf("solicitar_reverso status=%d body=%s", solicitarRR.Code, solicitarRR.Body.String())
	}

	var solicitarResp struct {
		Workflow dbpkg.EmpresaCreditoWorkflow `json:"workflow"`
	}
	if err := json.Unmarshal(solicitarRR.Body.Bytes(), &solicitarResp); err != nil {
		t.Fatalf("decode solicitar workflow response: %v", err)
	}
	if solicitarResp.Workflow.ID <= 0 {
		t.Fatalf("expected workflow id > 0, got %+v", solicitarResp)
	}

	aprobarReq := httptest.NewRequest(http.MethodPut, "/api/empresa/creditos?action=aprobar_workflow&empresa_id=91", strings.NewReader(`{
		"empresa_id":91,
		"workflow_id":`+itoa64(solicitarResp.Workflow.ID)+`,
		"aprobado_por":"gerencia@empresa.com",
		"codigo_aprobacion":"APR-HND-001",
		"motivo_aprobacion":"Validado por gerencia"
	}`))
	aprobarReq.Header.Set("Content-Type", "application/json")
	aprobarReq = aprobarReq.WithContext(context.WithValue(aprobarReq.Context(), "adminEmail", "gerencia@empresa.com"))
	aprobarRR := httptest.NewRecorder()
	h.ServeHTTP(aprobarRR, aprobarReq)
	if aprobarRR.Code != http.StatusOK {
		t.Fatalf("aprobar_workflow status=%d body=%s", aprobarRR.Code, aprobarRR.Body.String())
	}

	var aprobarResp struct {
		Workflow dbpkg.EmpresaCreditoWorkflow `json:"workflow"`
	}
	if err := json.Unmarshal(aprobarRR.Body.Bytes(), &aprobarResp); err != nil {
		t.Fatalf("decode aprobar workflow response: %v", err)
	}
	if aprobarResp.Workflow.EstadoSolicitud != "ejecutada" {
		t.Fatalf("expected workflow ejecutada after approve, got %+v", aprobarResp.Workflow)
	}
	if aprobarResp.Workflow.MovimientoResultadoID <= 0 {
		t.Fatalf("expected movimiento_resultado_id > 0, got %+v", aprobarResp.Workflow)
	}
}

func TestEmpresaCreditosHandlerLimitesClienteYBloqueo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_creditos_handler_limites.db")
	if err := dbpkg.EnsureEmpresaCreditosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaCreditosSchema: %v", err)
	}

	h := EmpresaCreditosHandler(dbEmp)
	clienteID, err := dbpkg.CreateCliente(dbEmp, dbpkg.Cliente{
		EmpresaID:         92,
		TipoDocumento:     "CC",
		NumeroDocumento:   "10992001",
		NombreRazonSocial: "Cliente Limite Handler",
		UsuarioCreador:    "qa_creditos",
	})
	if err != nil {
		t.Fatalf("CreateCliente: %v", err)
	}

	reqDenied := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?action=limite_cliente&empresa_id=92", strings.NewReader(`{
		"empresa_id":92,
		"cliente_id":`+itoa64(clienteID)+`,
		"limite_saldo_total":700,
		"max_creditos_activos":2
	}`))
	reqDenied.Header.Set("Content-Type", "application/json")
	reqDenied = reqDenied.WithContext(context.WithValue(reqDenied.Context(), "adminRole", "cajero"))
	reqDenied = reqDenied.WithContext(context.WithValue(reqDenied.Context(), "adminEmail", "cajero@empresa.com"))
	rrDenied := httptest.NewRecorder()
	h.ServeHTTP(rrDenied, reqDenied)
	if rrDenied.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for cajero upsert limite, got=%d body=%s", rrDenied.Code, rrDenied.Body.String())
	}

	reqUpsert := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?action=limite_cliente&empresa_id=92", strings.NewReader(`{
		"empresa_id":92,
		"cliente_id":`+itoa64(clienteID)+`,
		"limite_saldo_total":700,
		"max_creditos_activos":2,
		"requiere_aprobacion_exceso":false
	}`))
	reqUpsert.Header.Set("Content-Type", "application/json")
	reqUpsert = reqUpsert.WithContext(context.WithValue(reqUpsert.Context(), "adminRole", "contabilidad"))
	reqUpsert = reqUpsert.WithContext(context.WithValue(reqUpsert.Context(), "adminEmail", "contabilidad@empresa.com"))
	rrUpsert := httptest.NewRecorder()
	h.ServeHTTP(rrUpsert, reqUpsert)
	if rrUpsert.Code != http.StatusCreated && rrUpsert.Code != http.StatusOK {
		t.Fatalf("upsert limite status=%d body=%s", rrUpsert.Code, rrUpsert.Body.String())
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/api/empresa/creditos?action=limites_cliente&empresa_id=92&cliente_id="+itoa64(clienteID), nil)
	rrGet := httptest.NewRecorder()
	h.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("get limite status=%d body=%s", rrGet.Code, rrGet.Body.String())
	}

	createA := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?empresa_id=92", strings.NewReader(`{
		"empresa_id":92,
		"cliente_id":`+itoa64(clienteID)+`,
		"cliente_nombre":"Cliente Limite Handler",
		"tipo_credito":"cuotas",
		"monto_aprobado":500,
		"plazo_cuotas":4,
		"fecha_inicio":"2026-04-01",
		"fecha_vencimiento":"2026-08-01"
	}`))
	createA.Header.Set("Content-Type", "application/json")
	rrCreateA := httptest.NewRecorder()
	h.ServeHTTP(rrCreateA, createA)
	if rrCreateA.Code != http.StatusCreated {
		t.Fatalf("create credito A status=%d body=%s", rrCreateA.Code, rrCreateA.Body.String())
	}

	createB := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?empresa_id=92", strings.NewReader(`{
		"empresa_id":92,
		"cliente_id":`+itoa64(clienteID)+`,
		"cliente_nombre":"Cliente Limite Handler",
		"tipo_credito":"cuotas",
		"monto_aprobado":250,
		"plazo_cuotas":3,
		"fecha_inicio":"2026-05-01",
		"fecha_vencimiento":"2026-08-01"
	}`))
	createB.Header.Set("Content-Type", "application/json")
	rrCreateB := httptest.NewRecorder()
	h.ServeHTTP(rrCreateB, createB)
	if rrCreateB.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 by limite_saldo_total, got=%d body=%s", rrCreateB.Code, rrCreateB.Body.String())
	}

	reqDelete := httptest.NewRequest(http.MethodDelete, "/api/empresa/creditos?action=limite_cliente&empresa_id=92&cliente_id="+itoa64(clienteID), nil)
	reqDelete = reqDelete.WithContext(context.WithValue(reqDelete.Context(), "adminRole", "contabilidad"))
	reqDelete = reqDelete.WithContext(context.WithValue(reqDelete.Context(), "adminEmail", "contabilidad@empresa.com"))
	rrDelete := httptest.NewRecorder()
	h.ServeHTTP(rrDelete, reqDelete)
	if rrDelete.Code != http.StatusOK {
		t.Fatalf("delete limite status=%d body=%s", rrDelete.Code, rrDelete.Body.String())
	}

	createBAfterDelete := httptest.NewRequest(http.MethodPost, "/api/empresa/creditos?empresa_id=92", strings.NewReader(`{
		"empresa_id":92,
		"cliente_id":`+itoa64(clienteID)+`,
		"cliente_nombre":"Cliente Limite Handler",
		"tipo_credito":"cuotas",
		"monto_aprobado":250,
		"plazo_cuotas":3,
		"fecha_inicio":"2026-05-01",
		"fecha_vencimiento":"2026-08-01"
	}`))
	createBAfterDelete.Header.Set("Content-Type", "application/json")
	rrCreateBAfterDelete := httptest.NewRecorder()
	h.ServeHTTP(rrCreateBAfterDelete, createBAfterDelete)
	if rrCreateBAfterDelete.Code != http.StatusCreated {
		t.Fatalf("expected second credit created after deleting limit, got=%d body=%s", rrCreateBAfterDelete.Code, rrCreateBAfterDelete.Body.String())
	}
}

func TestEmpresaCreditosHandlerWorkflowPermisoFinoPorTipo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_creditos_handler_fine_permissions.db")
	if err := dbpkg.EnsureEmpresaCreditosSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaCreditosSchema: %v", err)
	}

	creditoID, err := dbpkg.CreateEmpresaCredito(dbEmp, dbpkg.EmpresaCredito{
		EmpresaID:        93,
		ClienteNombre:    "Cliente Fine Permission",
		TipoCredito:      "cuotas",
		MontoAprobado:    1200,
		CupoCredito:      1200,
		PlazoCuotas:      4,
		FechaInicio:      "2026-04-01",
		FechaVencimiento: "2026-08-01",
		EstadoCredito:    "activo",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaCredito: %v", err)
	}

	movAbonoID, _, err := dbpkg.RegisterEmpresaCreditoAbono(dbEmp, dbpkg.EmpresaCreditoAbonoInput{
		EmpresaID:      93,
		CreditoID:      creditoID,
		Monto:          300,
		MetodoPago:     "efectivo",
		ReferenciaPago: "ABN-FINE-001",
		UsuarioCreador: "caja@empresa.com",
	})
	if err != nil {
		t.Fatalf("RegisterEmpresaCreditoAbono: %v", err)
	}

	workflowRefID, err := dbpkg.CreateEmpresaCreditoWorkflowSolicitud(dbEmp, dbpkg.EmpresaCreditoWorkflowSolicitudInput{
		EmpresaID:                93,
		CreditoID:                creditoID,
		TipoSolicitud:            "refinanciacion",
		NivelAprobacionRequerido: 1,
		MotivoSolicitud:          "Refinanciar flujo",
		PayloadJSON:              `{"nuevo_plazo_cuotas":6,"nueva_tasa_interes":5.5}`,
		UsuarioCreador:           "analista@empresa.com",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaCreditoWorkflowSolicitud refinanciacion: %v", err)
	}

	h := EmpresaCreditosHandler(dbEmp)

	reqRefDenied := httptest.NewRequest(http.MethodPut, "/api/empresa/creditos?action=aprobar_workflow&empresa_id=93", strings.NewReader(`{
		"empresa_id":93,
		"workflow_id":`+itoa64(workflowRefID)+`,
		"aprobado_por":"contabilidad@empresa.com",
		"codigo_aprobacion":"APR-FINE-001",
		"motivo_aprobacion":"Intento contabilidad"
	}`))
	reqRefDenied.Header.Set("Content-Type", "application/json")
	reqRefDenied = reqRefDenied.WithContext(context.WithValue(reqRefDenied.Context(), "adminRole", "contabilidad"))
	reqRefDenied = reqRefDenied.WithContext(context.WithValue(reqRefDenied.Context(), "adminEmail", "contabilidad@empresa.com"))
	rrRefDenied := httptest.NewRecorder()
	h.ServeHTTP(rrRefDenied, reqRefDenied)
	if rrRefDenied.Code != http.StatusForbidden {
		t.Fatalf("expected forbidden for refinanciacion by contabilidad, got=%d body=%s", rrRefDenied.Code, rrRefDenied.Body.String())
	}

	workflowRef, err := dbpkg.GetEmpresaCreditoWorkflowByID(dbEmp, 93, workflowRefID)
	if err != nil {
		t.Fatalf("GetEmpresaCreditoWorkflowByID refinanciacion: %v", err)
	}
	if workflowRef.EstadoSolicitud != "pendiente_aprobacion" {
		t.Fatalf("expected workflow refinanciacion pendiente_aprobacion, got=%s", workflowRef.EstadoSolicitud)
	}

	workflowRevID, err := dbpkg.CreateEmpresaCreditoWorkflowSolicitud(dbEmp, dbpkg.EmpresaCreditoWorkflowSolicitudInput{
		EmpresaID:                93,
		CreditoID:                creditoID,
		TipoSolicitud:            "reverso_abono",
		MovimientoOrigenID:       movAbonoID,
		NivelAprobacionRequerido: 1,
		MotivoSolicitud:          "Reversar abono",
		UsuarioCreador:           "analista@empresa.com",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaCreditoWorkflowSolicitud reverso: %v", err)
	}

	reqRevAllowed := httptest.NewRequest(http.MethodPut, "/api/empresa/creditos?action=aprobar_workflow&empresa_id=93", strings.NewReader(`{
		"empresa_id":93,
		"workflow_id":`+itoa64(workflowRevID)+`,
		"aprobado_por":"contabilidad@empresa.com",
		"codigo_aprobacion":"APR-FINE-002",
		"motivo_aprobacion":"Aprobado por contabilidad"
	}`))
	reqRevAllowed.Header.Set("Content-Type", "application/json")
	reqRevAllowed = reqRevAllowed.WithContext(context.WithValue(reqRevAllowed.Context(), "adminRole", "contabilidad"))
	reqRevAllowed = reqRevAllowed.WithContext(context.WithValue(reqRevAllowed.Context(), "adminEmail", "contabilidad@empresa.com"))
	rrRevAllowed := httptest.NewRecorder()
	h.ServeHTTP(rrRevAllowed, reqRevAllowed)
	if rrRevAllowed.Code != http.StatusOK {
		t.Fatalf("expected reverso approval by contabilidad to succeed, got=%d body=%s", rrRevAllowed.Code, rrRevAllowed.Body.String())
	}

	workflowRev, err := dbpkg.GetEmpresaCreditoWorkflowByID(dbEmp, 93, workflowRevID)
	if err != nil {
		t.Fatalf("GetEmpresaCreditoWorkflowByID reverso: %v", err)
	}
	if workflowRev.EstadoSolicitud != "ejecutada" {
		t.Fatalf("expected workflow reverso ejecutada, got=%s", workflowRev.EstadoSolicitud)
	}
}
