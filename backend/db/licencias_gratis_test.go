package db

import "testing"

func TestIsApprovedLicenciaPaymentStatusCoversGatewayAliases(t *testing.T) {
	approved := []string{
		"APPROVED",
		"accepted",
		"ACREDITADA",
		"Aceptada",
		"Aprobado",
		"success",
		"1",
		"MANUAL",
	}
	for _, status := range approved {
		if !isApprovedLicenciaPaymentStatus(status) {
			t.Fatalf("expected status %q to be treated as approved", status)
		}
	}

	rejected := []string{"PENDING", "DECLINED", "REJECTED", "ERROR", "", "0"}
	for _, status := range rejected {
		if isApprovedLicenciaPaymentStatus(status) {
			t.Fatalf("did not expect status %q to be treated as approved", status)
		}
	}
}

func TestLicenciaPaymentRecordMatchesExclusion(t *testing.T) {
	if !licenciaPaymentRecordMatchesExclusion("TX-1", "REF-1", "TX-1", "") {
		t.Fatal("expected matching transaction to be excluded")
	}
	if !licenciaPaymentRecordMatchesExclusion("TX-1", "REF-1", "", "REF-1") {
		t.Fatal("expected matching reference to be excluded")
	}
	if licenciaPaymentRecordMatchesExclusion("TX-1", "REF-1", "TX-2", "REF-2") {
		t.Fatal("did not expect a different payment record to be excluded")
	}
}
