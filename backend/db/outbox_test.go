package db

import "testing"

func TestValidateOutboxEventRejectsUnsafeInput(t *testing.T) {
	t.Parallel()
	if err := ValidateOutboxEvent(OutboxEvent{EmpresaID: -1, Topic: "sale.confirmed"}); err == nil {
		t.Fatal("expected empresa validation error")
	}
	if err := ValidateOutboxEvent(OutboxEvent{EmpresaID: 1, Topic: "", MaxAttempts: 1}); err == nil {
		t.Fatal("expected topic validation error")
	}
	if err := ValidateOutboxEvent(OutboxEvent{EmpresaID: 1, Topic: "sale.confirmed", MaxAttempts: 26}); err == nil {
		t.Fatal("expected retry validation error")
	}
}

func TestValidateOutboxEventAcceptsTenantEvent(t *testing.T) {
	t.Parallel()
	err := ValidateOutboxEvent(OutboxEvent{EmpresaID: 12, Topic: "sale.confirmed", Version: 1, PayloadJSON: `{"sale_id":4}`, MaxAttempts: 5})
	if err != nil {
		t.Fatalf("ValidateOutboxEvent: %v", err)
	}
}
