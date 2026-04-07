package db

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEmpresaCreditosFlowCrearCuotasAbonoYResumen(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCreditosSchema(dbConn); err != nil {
		t.Fatalf("ensure creditos schema: %v", err)
	}

	creditoID, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:             77,
		ClienteNombre:         "Cliente Credito QA",
		TipoCredito:           "cuotas",
		MontoAprobado:         1200,
		CupoCredito:           1500,
		TasaInteres:           12,
		PlazoCuotas:           3,
		FechaInicio:           "2026-04-07",
		FechaVencimiento:      "2026-07-07",
		BloqueoAutomaticoMora: true,
		EstadoCredito:         "activo",
		UsuarioCreador:        "qa@empresa.com",
		Estado:                "activo",
	})
	if err != nil {
		t.Fatalf("create credito: %v", err)
	}
	if creditoID <= 0 {
		t.Fatalf("expected credito id > 0, got %d", creditoID)
	}

	credito, err := GetEmpresaCreditoByID(dbConn, 77, creditoID)
	if err != nil {
		t.Fatalf("get credito: %v", err)
	}
	if credito.SaldoActual <= 0 {
		t.Fatalf("expected saldo_actual > 0, got %+v", credito)
	}
	if credito.Codigo == "" {
		t.Fatalf("expected generated codigo, got %+v", credito)
	}

	cuotas, err := ListEmpresaCreditoCuotas(dbConn, 77, creditoID, false)
	if err != nil {
		t.Fatalf("list cuotas: %v", err)
	}
	if len(cuotas) != 3 {
		t.Fatalf("expected 3 cuotas, got %d", len(cuotas))
	}
	if cuotas[0].NumeroCuota != 1 {
		t.Fatalf("expected first cuota numero=1, got %+v", cuotas[0])
	}

	movID, updated, err := RegisterEmpresaCreditoAbono(dbConn, EmpresaCreditoAbonoInput{
		EmpresaID:      77,
		CreditoID:      creditoID,
		Monto:          450,
		MetodoPago:     "transferencia_bancaria",
		ReferenciaPago: "ABN-001",
		UsuarioCreador: "qa@empresa.com",
	})
	if err != nil {
		t.Fatalf("register abono: %v", err)
	}
	if movID <= 0 {
		t.Fatalf("expected movimiento id > 0, got %d", movID)
	}
	if updated == nil {
		t.Fatalf("expected updated credito after abono")
	}
	if updated.SaldoActual >= credito.SaldoActual {
		t.Fatalf("expected saldo reduced after abono, before=%.2f after=%.2f", credito.SaldoActual, updated.SaldoActual)
	}

	movs, err := ListEmpresaCreditoMovimientos(dbConn, 77, creditoID, false, 20)
	if err != nil {
		t.Fatalf("list movimientos: %v", err)
	}
	if len(movs) != 1 {
		t.Fatalf("expected 1 movimiento, got %d", len(movs))
	}
	if movs[0].Monto <= 0 {
		t.Fatalf("expected movimiento monto > 0, got %+v", movs[0])
	}

	resumen, err := GetEmpresaCreditosCarteraResumen(dbConn, 77, false)
	if err != nil {
		t.Fatalf("resumen cartera: %v", err)
	}
	if resumen.TotalCreditos != 1 {
		t.Fatalf("expected total_creditos=1, got %+v", resumen)
	}
	if resumen.SaldoTotal <= 0 {
		t.Fatalf("expected saldo_total > 0 after partial abono, got %+v", resumen)
	}
}

func TestEmpresaCreditosMoraDashboard(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCreditosSchema(dbConn); err != nil {
		t.Fatalf("ensure creditos schema: %v", err)
	}

	today := time.Now()
	proximo := today.AddDate(0, 0, 2).Format("2006-01-02")
	vencidoSuave := today.AddDate(0, 0, -4).Format("2006-01-02")
	vencidoSevero := today.AddDate(0, 0, -20).Format("2006-01-02")

	if _, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        91,
		ClienteNombre:    "Cliente Proximo",
		TipoCredito:      "cuotas",
		MontoAprobado:    500,
		CupoCredito:      600,
		PlazoCuotas:      2,
		FechaInicio:      today.Format("2006-01-02"),
		FechaVencimiento: proximo,
		EstadoCredito:    "activo",
		Estado:           "activo",
	}); err != nil {
		t.Fatalf("create credito proximo: %v", err)
	}

	if _, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        91,
		ClienteNombre:    "Cliente Mora 1",
		TipoCredito:      "cuotas",
		MontoAprobado:    700,
		CupoCredito:      800,
		PlazoCuotas:      3,
		FechaInicio:      today.AddDate(0, -1, 0).Format("2006-01-02"),
		FechaVencimiento: vencidoSuave,
		EstadoCredito:    "activo",
		Estado:           "activo",
	}); err != nil {
		t.Fatalf("create credito vencido suave: %v", err)
	}

	if _, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        91,
		ClienteNombre:    "Cliente Mora 2",
		TipoCredito:      "cuotas",
		MontoAprobado:    1400,
		CupoCredito:      1600,
		PlazoCuotas:      4,
		FechaInicio:      today.AddDate(0, -2, 0).Format("2006-01-02"),
		FechaVencimiento: vencidoSevero,
		EstadoCredito:    "activo",
		Estado:           "activo",
	}); err != nil {
		t.Fatalf("create credito vencido severo: %v", err)
	}

	dashboard, err := GetEmpresaCreditosMoraDashboard(dbConn, 91, 7, 5, false)
	if err != nil {
		t.Fatalf("GetEmpresaCreditosMoraDashboard: %v", err)
	}

	if dashboard.TotalProximosVencer != 1 {
		t.Fatalf("expected proximos=1, got %+v", dashboard)
	}
	if dashboard.TotalVencidos != 2 {
		t.Fatalf("expected vencidos=2, got %+v", dashboard)
	}
	if len(dashboard.RankingMorosidad) != 2 {
		t.Fatalf("expected ranking length 2, got %d", len(dashboard.RankingMorosidad))
	}
	if dashboard.RankingMorosidad[0].DiasMora < dashboard.RankingMorosidad[1].DiasMora {
		t.Fatalf("expected ranking sorted by dias_mora desc, got first=%d second=%d",
			dashboard.RankingMorosidad[0].DiasMora,
			dashboard.RankingMorosidad[1].DiasMora,
		)
	}
}

func TestEmpresaCreditosWorkflowReversoAprobadoEjecutaReversion(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCreditosSchema(dbConn); err != nil {
		t.Fatalf("ensure creditos schema: %v", err)
	}

	creditoID, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        120,
		ClienteNombre:    "Cliente Reverso Workflow",
		TipoCredito:      "cuotas",
		MontoAprobado:    1000,
		CupoCredito:      1000,
		PlazoCuotas:      4,
		FechaInicio:      "2026-04-01",
		FechaVencimiento: "2026-08-01",
		EstadoCredito:    "activo",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("create credito: %v", err)
	}

	movAbonoID, creditoPostAbono, err := RegisterEmpresaCreditoAbono(dbConn, EmpresaCreditoAbonoInput{
		EmpresaID:      120,
		CreditoID:      creditoID,
		Monto:          250,
		MetodoPago:     "efectivo",
		ReferenciaPago: "ABN-REV-001",
		UsuarioCreador: "caja@empresa.com",
	})
	if err != nil {
		t.Fatalf("register abono: %v", err)
	}
	if movAbonoID <= 0 {
		t.Fatalf("expected movAbonoID > 0")
	}
	if creditoPostAbono == nil {
		t.Fatalf("expected creditoPostAbono")
	}

	workflowID, err := CreateEmpresaCreditoWorkflowSolicitud(dbConn, EmpresaCreditoWorkflowSolicitudInput{
		EmpresaID:                120,
		CreditoID:                creditoID,
		TipoSolicitud:            "reverso_abono",
		MovimientoOrigenID:       movAbonoID,
		NivelAprobacionRequerido: 1,
		MotivoSolicitud:          "Error operativo de caja",
		UsuarioCreador:           "supervisor@empresa.com",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaCreditoWorkflowSolicitud reverso: %v", err)
	}

	workflow, err := AprobarEmpresaCreditoWorkflow(dbConn, EmpresaCreditoWorkflowAprobacionInput{
		EmpresaID:        120,
		WorkflowID:       workflowID,
		AprobadoPor:      "gerencia@empresa.com",
		CodigoAprobacion: "APR-REV-001",
		MotivoAprobacion: "Autorizado por control interno",
		EjecutadoPor:     "gerencia@empresa.com",
		UsuarioCreador:   "gerencia@empresa.com",
	})
	if err != nil {
		t.Fatalf("AprobarEmpresaCreditoWorkflow reverso: %v", err)
	}
	if workflow == nil {
		t.Fatalf("expected workflow after approve")
	}
	if workflow.EstadoSolicitud != "ejecutada" {
		t.Fatalf("expected workflow ejecutada, got %+v", workflow)
	}
	if workflow.MovimientoResultadoID <= 0 {
		t.Fatalf("expected movimiento_resultado_id > 0, got %+v", workflow)
	}

	creditoPostReverso, err := GetEmpresaCreditoByID(dbConn, 120, creditoID)
	if err != nil {
		t.Fatalf("GetEmpresaCreditoByID post reverso: %v", err)
	}
	if creditoPostReverso.SaldoActual <= creditoPostAbono.SaldoActual {
		t.Fatalf("expected saldo to increase after reverso, before=%.2f after=%.2f", creditoPostAbono.SaldoActual, creditoPostReverso.SaldoActual)
	}

	movs, err := ListEmpresaCreditoMovimientos(dbConn, 120, creditoID, false, 20)
	if err != nil {
		t.Fatalf("ListEmpresaCreditoMovimientos: %v", err)
	}
	encontroReverso := false
	for _, mov := range movs {
		if mov.ID == workflow.MovimientoResultadoID && mov.TipoMovimiento == "reverso" {
			encontroReverso = true
			break
		}
	}
	if !encontroReverso {
		t.Fatalf("expected reverso movimiento id=%d in movimientos", workflow.MovimientoResultadoID)
	}
}

func TestEmpresaCreditosWorkflowRefinanciacionAprobadaRegeneraCuotas(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCreditosSchema(dbConn); err != nil {
		t.Fatalf("ensure creditos schema: %v", err)
	}

	creditoID, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        121,
		ClienteNombre:    "Cliente Refinanciacion Workflow",
		TipoCredito:      "cuotas",
		MontoAprobado:    1500,
		CupoCredito:      1600,
		PlazoCuotas:      5,
		FechaInicio:      "2026-04-01",
		FechaVencimiento: "2026-09-01",
		EstadoCredito:    "activo",
		Estado:           "activo",
	})
	if err != nil {
		t.Fatalf("create credito: %v", err)
	}

	if _, _, err := RegisterEmpresaCreditoAbono(dbConn, EmpresaCreditoAbonoInput{
		EmpresaID:      121,
		CreditoID:      creditoID,
		Monto:          300,
		MetodoPago:     "transferencia_bancaria",
		ReferenciaPago: "ABN-REF-001",
		UsuarioCreador: "cobranza@empresa.com",
	}); err != nil {
		t.Fatalf("register abono: %v", err)
	}

	payloadRef := map[string]interface{}{
		"nuevo_plazo_cuotas":      8,
		"nueva_tasa_interes":      7.5,
		"nueva_tasa_mora":         2.5,
		"nueva_fecha_inicio":      "2026-05-01",
		"nueva_fecha_vencimiento": "2027-01-01",
	}

	workflowID, err := CreateEmpresaCreditoWorkflowSolicitud(dbConn, EmpresaCreditoWorkflowSolicitudInput{
		EmpresaID:                121,
		CreditoID:                creditoID,
		TipoSolicitud:            "refinanciacion",
		NivelAprobacionRequerido: 1,
		MotivoSolicitud:          "Refinanciacion por flujo de caja",
		PayloadJSON:              creditoMarshalJSON(payloadRef, "{}"),
		UsuarioCreador:           "analista@empresa.com",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaCreditoWorkflowSolicitud refinanciacion: %v", err)
	}

	workflow, err := AprobarEmpresaCreditoWorkflow(dbConn, EmpresaCreditoWorkflowAprobacionInput{
		EmpresaID:        121,
		WorkflowID:       workflowID,
		AprobadoPor:      "director.finanzas@empresa.com",
		CodigoAprobacion: "APR-REF-001",
		MotivoAprobacion: "Aprobado por comité",
		EjecutadoPor:     "director.finanzas@empresa.com",
		UsuarioCreador:   "director.finanzas@empresa.com",
	})
	if err != nil {
		t.Fatalf("AprobarEmpresaCreditoWorkflow refinanciacion: %v", err)
	}
	if workflow == nil {
		t.Fatalf("expected workflow after approve")
	}
	if workflow.EstadoSolicitud != "ejecutada" {
		t.Fatalf("expected workflow ejecutada, got %+v", workflow)
	}

	creditoPostRef, err := GetEmpresaCreditoByID(dbConn, 121, creditoID)
	if err != nil {
		t.Fatalf("GetEmpresaCreditoByID post refin: %v", err)
	}
	if creditoPostRef.PlazoCuotas != 8 {
		t.Fatalf("expected plazo_cuotas=8, got %+v", creditoPostRef)
	}
	if creditoPostRef.TasaInteres != 7.5 {
		t.Fatalf("expected tasa_interes=7.5, got %+v", creditoPostRef)
	}

	cuotasActivas, err := ListEmpresaCreditoCuotas(dbConn, 121, creditoID, false)
	if err != nil {
		t.Fatalf("ListEmpresaCreditoCuotas activas: %v", err)
	}
	pendientes := 0
	for _, cuota := range cuotasActivas {
		if cuota.EstadoCuota == "pendiente" {
			pendientes++
		}
	}
	if pendientes != 8 {
		t.Fatalf("expected 8 cuotas pendientes nuevas after refinanciacion, got %d", pendientes)
	}
}

func TestEmpresaCreditosClienteLimitesBloqueaExceso(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCreditosSchema(dbConn); err != nil {
		t.Fatalf("ensure creditos schema: %v", err)
	}

	if _, err := UpsertEmpresaCreditoClienteLimite(dbConn, EmpresaCreditoClienteLimite{
		EmpresaID:          140,
		ClienteID:          501,
		MaxCreditosActivos: 1,
		UsuarioCreador:     "qa_creditos",
		Estado:             "activo",
	}); err != nil {
		t.Fatalf("upsert limite max_creditos_activos: %v", err)
	}

	if _, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        140,
		ClienteID:        501,
		ClienteNombre:    "Cliente Limite Uno",
		TipoCredito:      "cuotas",
		MontoAprobado:    300,
		PlazoCuotas:      3,
		FechaInicio:      "2026-04-01",
		FechaVencimiento: "2026-07-01",
		EstadoCredito:    "activo",
		Estado:           "activo",
	}); err != nil {
		t.Fatalf("create credito cliente 501 (primero): %v", err)
	}

	_, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        140,
		ClienteID:        501,
		ClienteNombre:    "Cliente Limite Uno",
		TipoCredito:      "cuotas",
		MontoAprobado:    200,
		PlazoCuotas:      2,
		FechaInicio:      "2026-05-01",
		FechaVencimiento: "2026-07-01",
		EstadoCredito:    "activo",
		Estado:           "activo",
	})
	if err == nil || !strings.Contains(err.Error(), "max_creditos_activos") {
		t.Fatalf("expected max_creditos_activos error, got=%v", err)
	}

	if _, err := UpsertEmpresaCreditoClienteLimite(dbConn, EmpresaCreditoClienteLimite{
		EmpresaID:        140,
		ClienteID:        502,
		LimiteSaldoTotal: 700,
		UsuarioCreador:   "qa_creditos",
		Estado:           "activo",
		Observaciones:    "limite saldo para pruebas",
	}); err != nil {
		t.Fatalf("upsert limite_saldo_total: %v", err)
	}

	if _, err := CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        140,
		ClienteID:        502,
		ClienteNombre:    "Cliente Limite Dos",
		TipoCredito:      "cuotas",
		MontoAprobado:    500,
		PlazoCuotas:      4,
		FechaInicio:      "2026-04-01",
		FechaVencimiento: "2026-08-01",
		EstadoCredito:    "activo",
		Estado:           "activo",
	}); err != nil {
		t.Fatalf("create credito cliente 502 (primero): %v", err)
	}

	_, err = CreateEmpresaCredito(dbConn, EmpresaCredito{
		EmpresaID:        140,
		ClienteID:        502,
		ClienteNombre:    "Cliente Limite Dos",
		TipoCredito:      "cuotas",
		MontoAprobado:    250,
		PlazoCuotas:      2,
		FechaInicio:      "2026-05-01",
		FechaVencimiento: "2026-07-01",
		EstadoCredito:    "activo",
		Estado:           "activo",
	})
	if err == nil || !strings.Contains(err.Error(), "limite_saldo_total") {
		t.Fatalf("expected limite_saldo_total error, got=%v", err)
	}
}

func TestEmpresaCreditosClienteLimitesCRUD(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	if err := EnsureEmpresaCreditosSchema(dbConn); err != nil {
		t.Fatalf("ensure creditos schema: %v", err)
	}

	limiteID, err := UpsertEmpresaCreditoClienteLimite(dbConn, EmpresaCreditoClienteLimite{
		EmpresaID:                141,
		ClienteID:                601,
		LimiteSaldoTotal:         1500,
		MaxCreditosActivos:       3,
		RequiereAprobacionExceso: true,
		UsuarioCreador:           "qa_creditos",
		Estado:                   "activo",
		Observaciones:            "crud limites",
	})
	if err != nil {
		t.Fatalf("upsert limite: %v", err)
	}
	if limiteID <= 0 {
		t.Fatalf("expected limiteID > 0, got=%d", limiteID)
	}

	row, err := GetEmpresaCreditoClienteLimite(dbConn, 141, 601, false)
	if err != nil {
		t.Fatalf("GetEmpresaCreditoClienteLimite: %v", err)
	}
	if row == nil || row.ClienteID != 601 || row.MaxCreditosActivos != 3 || !row.RequiereAprobacionExceso {
		t.Fatalf("unexpected limite row: %+v", row)
	}

	rows, total, err := ListEmpresaCreditoClienteLimites(dbConn, 141, EmpresaCreditoClienteLimiteFilter{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("ListEmpresaCreditoClienteLimites: %v", err)
	}
	if total != 1 || len(rows) != 1 {
		t.Fatalf("expected total=1 len=1, got total=%d len=%d", total, len(rows))
	}

	if err := SetEmpresaCreditoClienteLimiteRowEstado(dbConn, 141, 601, "inactivo"); err != nil {
		t.Fatalf("SetEmpresaCreditoClienteLimiteRowEstado: %v", err)
	}

	_, err = GetEmpresaCreditoClienteLimite(dbConn, 141, 601, false)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows for inactive limit, got=%v", err)
	}

	rowInactive, err := GetEmpresaCreditoClienteLimite(dbConn, 141, 601, true)
	if err != nil {
		t.Fatalf("GetEmpresaCreditoClienteLimite includeInactive: %v", err)
	}
	if rowInactive == nil || rowInactive.Estado != "inactivo" {
		t.Fatalf("expected inactive row, got %+v", rowInactive)
	}
}
