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

func TestEmpresaFacturacionElectronicaReintentosYReconciliacion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_facturacion_reintentos_handler.db")
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaConfiguracionAvanzadaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion avanzada schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}
	seedFacturacionCumplimientoConfig(t, dbEmp, 71)

	h := EmpresaFacturacionElectronicaHandler(dbEmp, nil)

	// Primera emision en proveedor manual para validar flujo estable.
	reqEmitOK := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=emitir", strings.NewReader(`{"empresa_id":71,"documento_codigo":"FAC-RT-OK-71","estado_actual":"borrador","monto_total":125000,"moneda":"COP","periodo_contable":"2026-04"}`))
	reqEmitOK = reqEmitOK.WithContext(context.WithValue(reqEmitOK.Context(), "adminEmail", "facturacion@test.com"))
	reqEmitOK.Header.Set("Content-Type", "application/json")
	rrEmitOK := httptest.NewRecorder()
	h.ServeHTTP(rrEmitOK, reqEmitOK)
	if rrEmitOK.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrEmitOK.Code, rrEmitOK.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrEmitOK.Body.String()), `"estado_envio":"enviado"`) {
		t.Fatalf("expected integracion estado_envio enviado, got body=%s", rrEmitOK.Body.String())
	}

	// Cambia a proveedor externo sin URL para forzar falla controlada y poblar cola de reintentos.
	if _, err := dbpkg.UpsertFacturacionElectronicaPaisConfig(dbEmp, dbpkg.FacturacionElectronicaPaisConfig{
		EmpresaID:           71,
		PaisCodigo:          "CO",
		Proveedor:           "externo",
		Ambiente:            "produccion",
		TipoDocumentoEmisor: "NIT",
		IdentificadorFiscal: "900123456",
		RazonSocial:         "Empresa Facturacion QA SAS",
		PrefijoFactura:      "FE",
		ResolucionNumero:    "18760000000001",
		Estado:              "activo",
		UsuarioCreador:      "facturacion@test.com",
	}); err != nil {
		t.Fatalf("upsert config proveedor externo: %v", err)
	}

	reqEmitFail := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=emitir", strings.NewReader(`{"empresa_id":71,"documento_codigo":"FAC-RT-FAIL-71","estado_actual":"borrador","monto_total":99000,"moneda":"COP","periodo_contable":"2026-04"}`))
	reqEmitFail = reqEmitFail.WithContext(context.WithValue(reqEmitFail.Context(), "adminEmail", "facturacion@test.com"))
	reqEmitFail.Header.Set("Content-Type", "application/json")
	rrEmitFail := httptest.NewRecorder()
	h.ServeHTTP(rrEmitFail, reqEmitFail)
	if rrEmitFail.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrEmitFail.Code, rrEmitFail.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrEmitFail.Body.String()), `"estado_envio":"fallido"`) {
		t.Fatalf("expected integracion estado_envio fallido, got body=%s", rrEmitFail.Body.String())
	}

	retryFail, err := dbpkg.GetFacturacionElectronicaRetryByDocumento(dbEmp, 71, "factura_electronica", "FAC-RT-FAIL-71")
	if err != nil {
		t.Fatalf("get retry fail item: %v", err)
	}
	retryFail.ProximoIntento = time.Now().Add(-2 * time.Minute).In(time.Local).Format("2006-01-02 15:04:05")
	retryFail.EstadoEnvio = "fallido"
	retryFail.UsuarioCreador = "facturacion@test.com"
	if _, err := dbpkg.UpsertFacturacionElectronicaRetry(dbEmp, *retryFail); err != nil {
		t.Fatalf("force retry due item: %v", err)
	}

	reqQueue := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica?action=procesar_reintentos&empresa_id=71&limit=20", nil)
	reqQueue = reqQueue.WithContext(context.WithValue(reqQueue.Context(), "adminEmail", "facturacion@test.com"))
	rrQueue := httptest.NewRecorder()
	h.ServeHTTP(rrQueue, reqQueue)
	if rrQueue.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrQueue.Code, rrQueue.Body.String())
	}
	var queueResp map[string]interface{}
	if err := json.Unmarshal(rrQueue.Body.Bytes(), &queueResp); err != nil {
		t.Fatalf("decode queue response: %v", err)
	}
	if v, ok := queueResp["procesados"].(float64); !ok || v < 1 {
		t.Fatalf("expected procesados >= 1, got %+v", queueResp)
	}

	reqList := httptest.NewRequest(http.MethodGet, "/api/empresa/facturacion_electronica?action=reintentos&empresa_id=71&limit=20", nil)
	rrList := httptest.NewRecorder()
	h.ServeHTTP(rrList, reqList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrList.Code, rrList.Body.String())
	}
	var listResp map[string]interface{}
	if err := json.Unmarshal(rrList.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if _, ok := listResp["items"].([]interface{}); !ok {
		t.Fatalf("expected items array in list response, got %+v", listResp)
	}

	reqReconGet := httptest.NewRequest(http.MethodGet, "/api/empresa/facturacion_electronica?action=reconciliacion&empresa_id=71", nil)
	rrReconGet := httptest.NewRecorder()
	h.ServeHTTP(rrReconGet, reqReconGet)
	if rrReconGet.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrReconGet.Code, rrReconGet.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrReconGet.Body.String()), `"documentos_evaluados"`) {
		t.Fatalf("expected reconciliacion resumen, got body=%s", rrReconGet.Body.String())
	}

	reqReconApply := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica?action=reconciliar_estados&empresa_id=71&aplicar=true", nil)
	reqReconApply = reqReconApply.WithContext(context.WithValue(reqReconApply.Context(), "adminEmail", "facturacion@test.com"))
	rrReconApply := httptest.NewRecorder()
	h.ServeHTTP(rrReconApply, reqReconApply)
	if rrReconApply.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrReconApply.Code, rrReconApply.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrReconApply.Body.String()), `"aplicar":true`) {
		t.Fatalf("expected aplicar=true in response, got body=%s", rrReconApply.Body.String())
	}
}
