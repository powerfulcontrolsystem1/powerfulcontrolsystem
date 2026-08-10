package db

import (
	"strings"
	"testing"
)

func TestNormalizeOutboxRecoveryInputRequiresTenantAndExplicitIDs(t *testing.T) {
	if _, _, _, err := normalizeOutboxRecoveryInput(0, "topic", []int64{1}, 1, true); err == nil {
		t.Fatal("empresa_id must be required")
	}
	if _, _, _, err := normalizeOutboxRecoveryInput(12, "", []int64{1}, 1, true); err == nil {
		t.Fatal("topic must be required")
	}
	if _, _, _, err := normalizeOutboxRecoveryInput(12, "topic", nil, 1, true); err == nil {
		t.Fatal("explicit event IDs must be required for recovery")
	}
}

func TestNormalizeOutboxRecoveryInputSortsAndBounds(t *testing.T) {
	ids, topic, limit, err := normalizeOutboxRecoveryInput(12, "  cuentas_por_pagar.pago_registrado  ", []int64{9, 3, 5}, 1, true)
	if err != nil {
		t.Fatal(err)
	}
	if topic != EmpresaCxPPaymentOutboxTopic {
		t.Fatalf("topic=%q", topic)
	}
	want := []int64{3, 5, 9}
	if len(ids) != len(want) {
		t.Fatalf("ids=%v", ids)
	}
	for index := range want {
		if ids[index] != want[index] {
			t.Fatalf("ids=%v want=%v", ids, want)
		}
	}
	if limit != len(want) {
		t.Fatalf("limit=%d want=%d", limit, len(want))
	}
	if _, _, _, err := normalizeOutboxRecoveryInput(12, "topic", []int64{9, 9}, 2, true); err == nil {
		t.Fatal("duplicate recovery IDs must be rejected")
	}

	tooMany := make([]int64, MaxOutboxRecoveryEvents+1)
	for index := range tooMany {
		tooMany[index] = int64(index + 1)
	}
	if _, _, _, err := normalizeOutboxRecoveryInput(12, "topic", tooMany, len(tooMany), true); err == nil {
		t.Fatal("recovery batch must be bounded")
	}
}

func TestOutboxRecoveryIDClauseUsesBoundParameters(t *testing.T) {
	clause, args := outboxRecoveryIDClause([]int64{4, 8})
	if clause != " AND id IN (?,?)" {
		t.Fatalf("clause=%q", clause)
	}
	if len(args) != 2 || args[0] != int64(4) || args[1] != int64(8) {
		t.Fatalf("args=%v", args)
	}
	if strings.Contains(clause, "4") || strings.Contains(clause, "8") {
		t.Fatalf("ids must not be interpolated: %q", clause)
	}
}
