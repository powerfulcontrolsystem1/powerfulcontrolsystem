package handlers

import (
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestParseSuperOutboxEventIDsRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"0", "-1", "abc", "1,,2"} {
		if _, err := parseSuperOutboxEventIDs(raw); err == nil {
			t.Fatalf("raw=%q must fail", raw)
		}
	}
}

func TestBuildSuperOutboxRecoveryItemsDoesNotExposeRawPayload(t *testing.T) {
	items := buildSuperOutboxRecoveryItems([]dbpkg.OutboxRecoveryEvent{{
		ID:          5,
		EmpresaID:   12,
		Topic:       dbpkg.EmpresaCxPPaymentOutboxTopic,
		PayloadJSON: `{"cuenta_por_pagar_id":14,"pago_id":4,"movimiento_finanzas_id":35,"monto":0.01,"token":"never-return"}`,
		Status:      dbpkg.OutboxDead,
		Attempts:    5,
		MaxAttempts: 5,
		LastError:   "outbox topic has no enabled worker handler",
	}})
	if len(items) != 1 {
		t.Fatalf("items=%d", len(items))
	}
	item := items[0]
	if item.EmpresaID != 12 || item.PagoID != 4 || item.CuentaPorPagarID != 14 || item.MovimientoFinanzasID != 35 || item.Monto != 0.01 {
		t.Fatalf("unexpected summary: %+v", item)
	}
}

func TestValidateSuperOutboxRecoveryRequestRequiresUniqueBoundedIDsAndReason(t *testing.T) {
	valid := superOutboxRecoveryRequest{
		EmpresaID: 12,
		Topic:     dbpkg.EmpresaCxPPaymentOutboxTopic,
		EventIDs:  []int64{5, 6},
		Reason:    "Manejador CxP desplegado y verificado",
	}
	if err := validateSuperOutboxRecoveryRequest(valid); err != nil {
		t.Fatal(err)
	}
	for name, request := range map[string]superOutboxRecoveryRequest{
		"empty":     {EventIDs: nil, Reason: valid.Reason},
		"duplicate": {EventIDs: []int64{5, 5}, Reason: valid.Reason},
		"negative":  {EventIDs: []int64{-5}, Reason: valid.Reason},
		"reason":    {EventIDs: []int64{5}, Reason: "corta"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateSuperOutboxRecoveryRequest(request); err == nil {
				t.Fatal("request must fail")
			}
		})
	}
}
