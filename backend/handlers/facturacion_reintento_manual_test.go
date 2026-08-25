package handlers

import (
	"context"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestFacturacionManualTerminalRetryContextIsExplicit(t *testing.T) {
	if facturacionManualTerminalRetryAllowed(context.Background()) {
		t.Fatal("automatic processing must not reactivate terminal retries")
	}
	ctx := context.WithValue(context.Background(), facturacionManualTerminalRetryContextKey{}, true)
	if !facturacionManualTerminalRetryAllowed(ctx) {
		t.Fatal("explicit manual retry context must be recognized")
	}
}

func TestFacturacionPrepareManualTerminalRetryReactivatesWithoutFiscalReference(t *testing.T) {
	retry := &dbpkg.FacturacionElectronicaRetryItem{
		EmpresaID:       12,
		TipoDocumento:   "factura_electronica",
		DocumentoCodigo: "FV-123",
		EstadoEnvio:     "fallido_terminal",
		Intentos:        5,
		MaxIntentos:     5,
		UltimoError:     "fallo local de persistencia",
		Estado:          "activo",
	}
	doc := dbpkg.EmpresaDocumentoFacturacion{EmpresaID: 12, DocumentoCodigo: "FV-123"}

	got, changed, err := facturacionPrepareManualTerminalRetry(retry, doc, "qa@example.test")
	if err != nil || !changed {
		t.Fatalf("expected safe manual reactivation, changed=%v err=%v", changed, err)
	}
	if got.EstadoEnvio != "fallido" || got.Intentos != 0 || got.UltimoError != "" {
		t.Fatalf("unexpected reactivated retry: %#v", got)
	}
	if !strings.Contains(got.Observaciones, "reactivacion manual DIAN") || !strings.Contains(got.Observaciones, "intentos previos=5") {
		t.Fatalf("manual retry audit is missing: %q", got.Observaciones)
	}
	if retry.EstadoEnvio != "fallido_terminal" || retry.Intentos != 5 {
		t.Fatalf("input retry must remain immutable: %#v", retry)
	}
}

func TestFacturacionPrepareManualTerminalRetryBlocksExistingFiscalEvidence(t *testing.T) {
	cases := []struct {
		name  string
		retry dbpkg.FacturacionElectronicaRetryItem
		doc   dbpkg.EmpresaDocumentoFacturacion
	}{
		{name: "track", retry: dbpkg.FacturacionElectronicaRetryItem{EstadoEnvio: "fallido_terminal", ReferenciaExterna: "TRACK-REAL"}},
		{name: "retry cufe", retry: dbpkg.FacturacionElectronicaRetryItem{EstadoEnvio: "fallido_terminal", CodigoValidacion: "CUFE-REAL"}},
		{name: "document cufe", retry: dbpkg.FacturacionElectronicaRetryItem{EstadoEnvio: "fallido_terminal"}, doc: dbpkg.EmpresaDocumentoFacturacion{CodigoValidacion: "CUFE-REAL"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, changed, err := facturacionPrepareManualTerminalRetry(&tc.retry, tc.doc, "qa@example.test")
			if err == nil || changed || got != &tc.retry {
				t.Fatalf("expected fiscal evidence to block resend, changed=%v err=%v", changed, err)
			}
		})
	}
}

func TestFacturacionPrepareManualTerminalRetryIgnoresNonTerminal(t *testing.T) {
	retry := &dbpkg.FacturacionElectronicaRetryItem{EstadoEnvio: "fallido", Intentos: 2}
	got, changed, err := facturacionPrepareManualTerminalRetry(retry, dbpkg.EmpresaDocumentoFacturacion{}, "qa@example.test")
	if err != nil || changed || got != retry {
		t.Fatalf("non-terminal retry must remain untouched, changed=%v err=%v", changed, err)
	}
}
