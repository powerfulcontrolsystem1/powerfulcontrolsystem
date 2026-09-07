package db

import (
	"os"
	"strings"
	"testing"
)

func TestPaymentCheckoutReferenceIsStableAndOpaque(t *testing.T) {
	key := "checkout-wompi-20260826-0001"
	first, err := PaymentCheckoutReference("wompi", 10, 20, key)
	if err != nil {
		t.Fatal(err)
	}
	second, err := PaymentCheckoutReference("wompi", 10, 20, key)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("la referencia debe ser estable: %q != %q", first, second)
	}
	if strings.Contains(first, key) || !strings.Contains(first, "WOMPI-LIC-10-EMP-20-IDEM-") {
		t.Fatalf("la referencia no debe exponer la clave: %q", first)
	}
	other, err := PaymentCheckoutReference("wompi", 10, 20, key+"-other")
	if err != nil {
		t.Fatal(err)
	}
	if other == first {
		t.Fatal("claves distintas no pueden producir la misma referencia")
	}
}

func TestPaymentCheckoutProviderIsRestricted(t *testing.T) {
	if _, err := PaymentCheckoutReference("otro", 10, 20, "checkout-valid-key-0001"); err == nil {
		t.Fatal("se acepto un proveedor no soportado")
	}
}

func TestLicenciaActivationProcessingIsNotAutomaticallyReclaimed(t *testing.T) {
	raw, err := os.ReadFile("db.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func TryBeginLicenciaPaymentActivation")
	if start < 0 {
		t.Fatal("no se encontro TryBeginLicenciaPaymentActivation")
	}
	end := strings.Index(source[start:], "func FinishLicenciaPaymentActivation")
	if end < 0 {
		t.Fatal("no se encontro el final del claim de activacion")
	}
	body := source[start : start+end]
	if !strings.Contains(body, "COALESCE(licencia_activation_status, '') <> 'processing'") {
		t.Fatal("una activacion incierta podria reclamarse y volver a sumar vigencia")
	}
	if strings.Contains(body, "licencia_activation_lease_until < CURRENT_TIMESTAMP") {
		t.Fatal("una activacion processing no debe reintentarse automaticamente por vencimiento")
	}
}

func TestPaymentPostEffectDoesNotReclaimUnknownRemoteOutcome(t *testing.T) {
	raw, err := os.ReadFile("payment_checkout_idempotency.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func ClaimPaymentPostEffect")
	if start < 0 {
		t.Fatal("no se encontro ClaimPaymentPostEffect")
	}
	end := strings.Index(source[start:], "func FinishPaymentPostEffect")
	if end < 0 {
		t.Fatal("no se encontro el final del claim de efecto")
	}
	body := source[start : start+end]
	if !strings.Contains(body, "WHERE payment_post_effect_idempotencia.estado = 'fallido'") {
		t.Fatal("el claim solo debe reintentar fallos confirmados")
	}
	if strings.Contains(body, "fecha_actualizacion <") || strings.Contains(body, "interval '") {
		t.Fatal("un efecto remoto incierto o en proceso no debe reclamarse automaticamente por tiempo")
	}
}

func TestSuperCatalogIncludesPaymentIdempotencyMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetSuper)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260826-001-payment-idempotency-v1" {
			continue
		}
		if migration.Apply == nil || migration.Body != paymentIdempotencySchemaFingerprint {
			t.Fatal("la migracion de idempotencia de pagos debe ser ejecutable e inmutable")
		}
		return
	}
	t.Fatal("falta la migracion de idempotencia de pagos en el catalogo super")
}
