package db

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// Only synthetic records in a disposable PostgreSQL database. No DIAN transport.
func fiscalNumberingTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	admin, err := sql.Open(PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("fiscal_numbering_qa_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	conn, err := sql.Open(PostgresCompatDriverName(), u.String())
	if err != nil {
		t.Fatal(err)
	}
	conn.SetMaxOpenConns(16)
	t.Cleanup(func() {
		conn.Close()
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	for _, ensure := range []func(*sql.DB) error{EnsureEmpresaConfiguracionAvanzadaSchema, EnsureEmpresaFacturacionElectronicaSchema, EnsureEmpresaDocumentosTransaccionalesSchema} {
		if err = ensure(conn); err != nil {
			t.Fatal(err)
		}
	}
	for _, tenant := range []int64{71001, 71002} {
		if _, err = UpsertEmpresaConfiguracionAvanzada(conn, fiscalNumberingTestConfig(tenant)); err != nil {
			t.Fatal(err)
		}
		cfg := defaultFacturacionConfig(tenant, "CO")
		cfg.TipoDocumentoEmisor, cfg.IdentificadorFiscal, cfg.RazonSocial = "NIT", "900000001", "QA FISCAL SINTETICO"
		cfg.Proveedor, cfg.Ambiente, cfg.PrefijoFactura, cfg.ResolucionNumero = "dian", "produccion", "QA", "TEST-NOT-AUTHORIZED"
		if _, err = UpsertFacturacionElectronicaPaisConfig(conn, cfg); err != nil {
			t.Fatal(err)
		}
	}
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = applyEmpresaFacturacionNumeracionReservasTx(context.Background(), tx); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return conn
}

func TestFiscalNumberingInvalidAmountBeforeDatabase(t *testing.T) {
	for _, amount := range []float64{-1, math.NaN(), math.Inf(1), 1e16} {
		if _, err := PrepareFacturacionDocumentoLegal(nil, 71001, "CO", "FV-QA", amount, "COP"); err == nil {
			t.Fatal("invalid amount accepted")
		}
	}
}

func TestFiscalNumberingPostgresReplayConflictAndTenantIsolation(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	first, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-SHARED", 100, "COP")
	if err != nil {
		t.Fatal(err)
	}
	other, err := PrepareFacturacionDocumentoLegal(conn, 71002, "CO", "FV-QA-SHARED", 100, "COP")
	if err != nil {
		t.Fatal(err)
	}
	if first.NumeroLegal != other.NumeroLegal || first.EmpresaID == other.EmpresaID {
		t.Fatal("tenant numbering is not independent")
	}
	for _, input := range []struct {
		amount   float64
		currency string
	}{{101, "COP"}, {100, "USD"}} {
		if _, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-SHARED", input.amount, input.currency); err == nil {
			t.Fatal("changed replay was accepted")
		}
	}
	var count, next int
	if err = conn.QueryRow(`SELECT COUNT(*) FROM empresa_facturacion_reservas_numeracion WHERE empresa_id=71001`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if err = conn.QueryRow(`SELECT proximo_consecutivo FROM empresa_configuracion_avanzada WHERE empresa_id=71001`).Scan(&next); err != nil {
		t.Fatal(err)
	}
	if count != 1 || next != 2 {
		t.Fatalf("replay advanced state: reservations=%d next=%d", count, next)
	}
}

func TestFiscalNumberingPostgresConcurrentReplayAndExhaustion(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	if _, err := conn.Exec(`UPDATE empresa_configuracion_avanzada SET consecutivo_hasta=1 WHERE empresa_id=71001`); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-LAST", 100, "COP")
			if err != nil {
				t.Error(err)
				return
			}
			if got.NumeroLegal != "QA1" {
				t.Error("replay changed number")
			}
		}()
	}
	wg.Wait()
	if _, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-EXHAUSTED", 100, "COP"); err == nil {
		t.Fatal("exhausted range accepted")
	}
}

func TestFiscalNumberingPostgresLegacyCollisionRollsBack(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	if _, err := conn.Exec(`INSERT INTO empresa_facturacion_documentos (empresa_id,tipo_documento,documento_codigo,numero_legal,estado_documento) VALUES (71001,'factura_electronica','FV-LEGACY','QA-1','pendiente')`); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-NEW", 100, "COP"); err == nil {
		t.Fatal("legacy number reused")
	}
	var next int
	if err := conn.QueryRow(`SELECT proximo_consecutivo FROM empresa_configuracion_avanzada WHERE empresa_id=71001`).Scan(&next); err != nil {
		t.Fatal(err)
	}
	if next != 1 {
		t.Fatal("failed reservation advanced counter")
	}
}

func TestFiscalNumberingPostgresMissingOrChangedConfiguration(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	if _, err := conn.Exec(`UPDATE facturacion_electronica_pais SET prefijo_factura='OTHER' WHERE empresa_id=71001`); err != nil {
		t.Fatal(err)
	}
	if _, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-MISMATCH", 100, "COP"); err == nil {
		t.Fatal("mixed country/advanced numbering accepted")
	}
	if _, err := PrepareFacturacionDocumentoLegal(conn, 71003, "CO", "FV-QA-NO-TENANT", 100, "COP"); err == nil {
		t.Fatal("missing tenant configuration accepted")
	}
}

func TestFiscalNumberingPostgresPendingDocumentCanReserve(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	if _, err := conn.Exec(`INSERT INTO empresa_facturacion_documentos (empresa_id,tipo_documento,documento_codigo,estado_documento) VALUES (71001,'factura_electronica','FV-PENDING','pendiente_emision')`); err != nil {
		t.Fatal(err)
	}
	got, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-PENDING", 100, "COP")
	if err != nil || got.NumeroLegal != "QA1" {
		t.Fatalf("unnumbered pending document cannot reserve: %v", err)
	}
}

func TestOfflinePostgresConcurrentClaimsAndPayloadConflict(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	if err := EnsureEmpresaVentasOfflineSchema(conn); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	claims := make(chan bool, 32)
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, claimed, err := ClaimEmpresaVentaOfflineSync(conn, 71001, "OFF-QA-SAME", `{"total":100}`, "2026-09-06", "qa@example.invalid", "isolated test")
			if err != nil {
				t.Error(err)
				return
			}
			claims <- claimed
		}()
	}
	wg.Wait()
	close(claims)
	count := 0
	for claimed := range claims {
		if claimed {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("same offline sale claimed %d times", count)
	}
	if _, _, err := ClaimEmpresaVentaOfflineSync(conn, 71001, "OFF-QA-SAME", `{"total":101}`, "2026-09-06", "qa@example.invalid", ""); err == nil {
		t.Fatal("offline payload conflict accepted")
	}
	if _, claimed, err := ClaimEmpresaVentaOfflineSync(conn, 71002, "OFF-QA-SAME", `{"total":100}`, "2026-09-06", "qa@example.invalid", ""); err != nil || !claimed {
		t.Fatalf("other tenant was blocked: %v", err)
	}
}

func fiscalNumberingTestConfig(tenant int64) EmpresaConfiguracionAvanzada {
	return EmpresaConfiguracionAvanzada{EmpresaID: tenant, PaisCodigo: "CO", AmbienteFE: "produccion", TipoDocumentoEmisor: "NIT", NIT: "900000001", RazonSocial: "QA FISCAL SINTETICO", PrefijoFactura: "QA", ResolucionNumero: "TEST-NOT-AUTHORIZED", ResolucionFechaDesde: "2020-01-01", ResolucionFechaHasta: "2099-12-31", ConsecutivoDesde: 1, ConsecutivoHasta: 10000, ProximoConsecutivo: 1, MonedaCodigo: "COP"}
}

func TestFiscalNumberingPostgresConcurrentCashRegisters(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	const registers = 32
	var wg sync.WaitGroup
	numbers := make(chan string, registers)
	for i := 0; i < registers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			got, err := PrepareFacturacionDocumentoLegalContext(ctx, conn, 71001, "CO", fmt.Sprintf("FV-QA-%d", i), 100, "COP")
			if err != nil {
				t.Error(err)
				return
			}
			numbers <- got.NumeroLegal
		}(i)
	}
	wg.Wait()
	close(numbers)
	seen := map[string]bool{}
	for n := range numbers {
		if seen[n] {
			t.Errorf("duplicate number %s", n)
		}
		seen[n] = true
	}
	if len(seen) != registers {
		t.Fatalf("numbers=%d want=%d", len(seen), registers)
	}
}

func TestCarritoDocumentIntentPostgresSerializesAutomaticFrequencyAcrossCashRegisters(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	const tenant = int64(71001)
	if _, err := conn.Exec(`UPDATE empresa_configuracion_avanzada SET
		modo_documento_venta='factura_electronica',facturacion_frecuencia_automatica_activa=1,
		facturacion_frecuencia_cada_n_no=3,facturacion_frecuencia_contador=0 WHERE empresa_id=$1`, tenant); err != nil {
		t.Fatal(err)
	}

	const registers = 32
	var wg sync.WaitGroup
	var mu sync.Mutex
	invoices := 0
	receipts := 0
	errs := make([]error, 0)
	for i := 0; i < registers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx, err := conn.BeginTx(context.Background(), nil)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}
			intent, err := resolveCarritoPaidDocumentIntentTx(tx, tenant, "")
			if err == nil {
				err = tx.Commit()
			} else {
				_ = tx.Rollback()
			}
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			if intent.ResolvedMode == "factura_electronica" {
				invoices++
			} else if intent.ResolvedMode == "comprobante_pago" {
				receipts++
			} else {
				errs = append(errs, fmt.Errorf("unexpected mode %q", intent.ResolvedMode))
			}
		}()
	}
	wg.Wait()
	if len(errs) > 0 {
		t.Fatalf("concurrent document intents failed: %v", errs[0])
	}
	if invoices != 8 || receipts != 24 {
		t.Fatalf("unexpected automatic distribution invoices=%d receipts=%d", invoices, receipts)
	}
	var counter int64
	if err := conn.QueryRow(`SELECT facturacion_frecuencia_contador FROM empresa_configuracion_avanzada WHERE empresa_id=$1`, tenant).Scan(&counter); err != nil {
		t.Fatal(err)
	}
	if counter != 0 {
		t.Fatalf("frequency counter lost updates: %d", counter)
	}
}

func TestFacturacionContingenciaPostgresLifecycleAndTenantIsolation(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = applyEmpresaFacturacionContingenciasTx(context.Background(), tx); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err = EmpresaFacturacionContingenciasSchemaReady(conn); err != nil {
		t.Fatal(err)
	}
	incident, err := OpenEmpresaFacturacionContingencia(conn, 71001, FacturacionContingenciaFallaDIAN,
		"Interrupcion sintetica suficientemente detallada", "QA-INTEGRATION-EVIDENCE", "qa", "sin transporte externo")
	if err != nil {
		t.Fatal(err)
	}
	if incident.EmpresaID != 71001 || incident.Tipo != FacturacionContingenciaFallaDIAN {
		t.Fatal("incident crossed tenant or type")
	}
	if _, err = RegisterEmpresaFacturacionContingenciaDocumento(conn, 71001, FacturacionContingenciaFallaDIAN, "factura_electronica", "FV-QA-CONT", 0); err != nil {
		t.Fatal(err)
	}
	if _, err = GetActiveEmpresaFacturacionContingencia(conn, 71002, FacturacionContingenciaFallaDIAN); err == nil {
		t.Fatal("incident visible in another tenant")
	}
	if err = RecoverEmpresaFacturacionContingencia(conn, 71001, incident.ID, "qa"); err != nil {
		t.Fatal(err)
	}
	if err = CloseEmpresaFacturacionContingencia(conn, 71001, incident.ID, "qa"); err == nil {
		t.Fatal("incident closed with pending documents")
	}
	if err = SetEmpresaFacturacionContingenciaDocumentoEstado(conn, 71001, "factura_electronica", "FV-QA-CONT", "aceptado"); err != nil {
		t.Fatal(err)
	}
	if err = CloseEmpresaFacturacionContingencia(conn, 71001, incident.ID, "qa"); err != nil {
		t.Fatal(err)
	}
	items, err := ListEmpresaFacturacionContingencias(conn, 71001, 10)
	if err != nil || len(items) != 1 || items[0].Estado != "cerrada" || items[0].DocumentosPendientes != 0 {
		t.Fatalf("unexpected lifecycle state: items=%v err=%v", items, err)
	}
}

func TestFacturacionContingenciaPostgresPaperSaleAdvancesOnlyTenantAuthorization(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	if err := EnsureEmpresaCarritosSchema(conn); err != nil {
		t.Fatal(err)
	}
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, apply := range []func(context.Context, *sql.Tx) error{
		applyEmpresaFacturacionArtefactosTx,
		applyEmpresaFacturacionFuenteFiscalTx,
		applyEmpresaFacturacionContingenciasTx,
	} {
		if err = apply(context.Background(), tx); err != nil {
			_ = tx.Rollback()
			t.Fatal(err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}

	bogota := time.FixedZone("America/Bogota", -5*60*60)
	today := time.Now().In(bogota)
	date := today.Format("2006-01-02")
	for _, tenant := range []int64{71001, 71002} {
		_, err = UpsertEmpresaFacturacionContingenciaConfiguracion(conn, EmpresaFacturacionContingenciaConfiguracion{
			EmpresaID: tenant, PaisCodigo: "CO", Prefijo: "CTG", ResolucionNumero: "TEST-NOT-AUTHORIZED",
			FechaDesde: today.AddDate(0, 0, -1).Format("2006-01-02"), FechaHasta: today.AddDate(0, 0, 1).Format("2006-01-02"),
			RangoDesde: 1, RangoHasta: 10, ProximoNumero: 1, Estado: "activo", UsuarioCreador: "qa",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	incident, err := OpenEmpresaFacturacionContingencia(conn, 71001, FacturacionContingenciaFallaFacturador,
		"Interrupcion sintetica suficientemente detallada", "QA-INTEGRATION-EVIDENCE", "qa", "sin transporte externo")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = UpsertEmpresaFacturacionContingenciaConfiguracion(conn, EmpresaFacturacionContingenciaConfiguracion{
		EmpresaID: 71001, PaisCodigo: "CO", Prefijo: "NEW", ResolucionNumero: "TEST-SECOND-AUTHORIZATION",
		FechaDesde: today.AddDate(0, 0, -1).Format("2006-01-02"), FechaHasta: today.AddDate(0, 0, 1).Format("2006-01-02"),
		RangoDesde: 1, RangoHasta: 10, ProximoNumero: 1, Estado: "activo", UsuarioCreador: "qa",
	}); err == nil {
		t.Fatal("active paper authorization was replaced during a facturer incident")
	}
	var cartID int64
	err = conn.QueryRow(`INSERT INTO carritos_compras
		(empresa_id,codigo,nombre,estado_carrito,moneda,total,pagado_en,total_pagado,estado)
		VALUES (71001,'QA-CONT-PAPER','QA sintetico','cerrado','COP',119,CURRENT_TIMESTAMP::text,119,'inactivo') RETURNING id`).Scan(&cartID)
	if err != nil {
		t.Fatal(err)
	}
	documentCode := fmt.Sprintf("CP-QA-CONT-PAPER-CRT-%d", cartID)
	if _, err = conn.Exec(`INSERT INTO empresa_facturacion_documentos
		(empresa_id,tipo_documento,documento_codigo,numero_legal,estado_documento,monto_total,moneda)
		VALUES (71001,'comprobante_pago',$1,$1,'emitida',119,'COP')`, documentCode); err != nil {
		t.Fatal(err)
	}
	if _, err = conn.Exec(`INSERT INTO empresa_facturacion_artefactos
		(empresa_id,tipo_documento,documento_codigo,tipo_artefacto,storage_ref,sha256,mime_type,tamano_bytes,estado)
		VALUES (71001,'comprobante_pago',$1,'fuente_fiscal_json','qa/private-source.json',$2,'application/json',2,'activo')`, documentCode, strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
	registered, err := RegisterEmpresaFacturacionTalonarioSale(conn, 71001, incident.ID, cartID, documentCode, "CTG1", date, "qa")
	if err != nil {
		t.Fatal(err)
	}
	if registered.EmpresaID != 71001 || registered.ConfiguracionID != incident.ConfiguracionTalonarioID || registered.NumeroPapel != "CTG1" || registered.EstadoTransmision != "pendiente" {
		t.Fatalf("unexpected registered paper sale: %+v", registered)
	}
	if _, err = RegisterEmpresaFacturacionTalonarioSale(conn, 71001, incident.ID, cartID, documentCode, "CTG1", date, "qa"); err == nil {
		t.Fatal("paper sale replay consumed or reused a number")
	}
	for tenant, want := range map[int64]int64{71001: 2, 71002: 1} {
		var next int64
		if err = conn.QueryRow(`SELECT proximo_numero FROM empresa_facturacion_contingencia_configuracion WHERE empresa_id=$1`, tenant).Scan(&next); err != nil {
			t.Fatal(err)
		}
		if next != want {
			t.Fatalf("tenant %d next paper number=%d want=%d", tenant, next, want)
		}
	}
	documents, err := ListEmpresaFacturacionContingenciaDocumentos(conn, 71001, 10)
	if err != nil || len(documents) != 1 || documents[0].CarritoID != cartID {
		t.Fatalf("unexpected contingency documents: %+v err=%v", documents, err)
	}
	otherDocuments, err := ListEmpresaFacturacionContingenciaDocumentos(conn, 71002, 10)
	if err != nil || len(otherDocuments) != 0 {
		t.Fatalf("paper sale leaked to other tenant: %+v err=%v", otherDocuments, err)
	}
	if err = RecoverEmpresaFacturacionContingencia(conn, 71001, incident.ID, "qa"); err != nil {
		t.Fatal(err)
	}
	if _, err = UpsertEmpresaFacturacionContingenciaConfiguracion(conn, EmpresaFacturacionContingenciaConfiguracion{
		EmpresaID: 71001, PaisCodigo: "CO", Prefijo: "NEW", ResolucionNumero: "TEST-SECOND-AUTHORIZATION",
		FechaDesde: today.AddDate(0, 0, -1).Format("2006-01-02"), FechaHasta: today.AddDate(0, 0, 1).Format("2006-01-02"),
		RangoDesde: 1, RangoHasta: 10, ProximoNumero: 1, Estado: "activo", UsuarioCreador: "qa",
	}); err != nil {
		t.Fatal(err)
	}
	var configCount, activeCount int
	if err = conn.QueryRow(`SELECT COUNT(*),COUNT(*) FILTER (WHERE estado='activo')
		FROM empresa_facturacion_contingencia_configuracion WHERE empresa_id=71001`).Scan(&configCount, &activeCount); err != nil {
		t.Fatal(err)
	}
	if configCount != 2 || activeCount != 1 {
		t.Fatalf("authorization history was not preserved: total=%d active=%d", configCount, activeCount)
	}
}

func TestFiscalNumberingPostgresReplaySurvivesDocumentPersistenceFailure(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	first, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-REPLAY", 100, "COP")
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a crash before the later document INSERT: only reservation committed.
	second, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-REPLAY", 100, "COP")
	if err != nil {
		t.Fatal(err)
	}
	if *first != *second {
		t.Fatalf("retry changed durable legal identity: %s -> %s", first.NumeroLegal, second.NumeroLegal)
	}
}

func TestFiscalNumberingPostgresStaleConfigurationCannotRewind(t *testing.T) {
	conn := fiscalNumberingTestDB(t)
	first, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-FIRST", 100, "COP")
	if err != nil {
		t.Fatal(err)
	}
	// An administrator saves a form opened before another cash register invoiced.
	if _, err = UpsertEmpresaConfiguracionAvanzada(conn, fiscalNumberingTestConfig(71001)); err != nil {
		t.Fatal(err)
	}
	second, err := PrepareFacturacionDocumentoLegal(conn, 71001, "CO", "FV-QA-NEXT", 100, "COP")
	if err != nil {
		t.Fatal(err)
	}
	if first.NumeroLegal == second.NumeroLegal {
		t.Fatal("stale configuration reused a fiscal number")
	}
}
