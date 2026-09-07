package db

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSuperPaymentAuditDoesNotSelectSensitiveLedgerColumns(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file")
	}
	body, err := os.ReadFile(filepath.Join(filepath.Dir(file), "payment_audit.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	for _, forbidden := range []string{"raw_payload", "clave_hash", "solicitud_hash", "respuesta_json", "ultimo_error"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("payment audit must not reference sensitive ledger column %q", forbidden)
		}
	}
}

func TestPaymentAuditWhereUsesBoundValues(t *testing.T) {
	where, args := paymentAuditWhere(SuperPaymentAuditFilters{Provider: "epayco", Status: "APPROVED", Search: "INV-1", EmpresaID: 9}, "status")
	if strings.Contains(where, "INV-1") || strings.Contains(where, "epayco") || len(args) != 7 {
		t.Fatalf("expected bound audit filters, where=%q args=%#v", where, args)
	}
}

func TestPaymentUpdatesAreMonotonicAndRequestPathsDoNotRepairSchema(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file")
	}
	body, err := os.ReadFile(filepath.Join(filepath.Dir(file), "db.go"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	guard := "UPPER(COALESCE(status, '')) = 'APPROVED' AND UPPER(?) <> 'APPROVED'"
	if strings.Count(text, guard) != 4 {
		t.Fatalf("expected monotonic approval guard in four payment updates, got %d", strings.Count(text, guard))
	}
	for _, functionName := range []string{
		"CreateWompiPaymentRecord", "UpdateWompiPaymentRecordByTransaction", "UpdateWompiPaymentRecordByReference",
		"CreateEpaycoPaymentRecord", "UpdateEpaycoPaymentRecordByTransaction", "UpdateEpaycoPaymentRecordByReference",
		"TryBeginLicenciaPaymentActivation", "FinishLicenciaPaymentActivation",
	} {
		start := strings.Index(text, "func "+functionName+"(")
		if start < 0 {
			t.Fatalf("missing function %s", functionName)
		}
		next := strings.Index(text[start+5:], "\nfunc ")
		block := text[start:]
		if next >= 0 {
			block = text[start : start+5+next]
		}
		if strings.Contains(block, "EnsurePaymentGatewaySchema") {
			t.Fatalf("request-path function %s must not repair schema", functionName)
		}
	}
}
