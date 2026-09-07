package db

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestShouldRefreshEmpresaCarritosSchemaForMissingObjects(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "missing base column",
			err:  errors.New(`pq: column c.cierre_caja_id does not exist`),
			want: true,
		},
		{
			name: "missing base relation",
			err:  errors.New(`pq: relation "carritos_compras" does not exist`),
			want: true,
		},
		{
			name: "business validation",
			err:  errors.New("stock insuficiente para cerrar la venta"),
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldRefreshEmpresaCarritosSchema(tc.err); got != tc.want {
				t.Fatalf("shouldRefreshEmpresaCarritosSchema()=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestBuildCarritosCompraByEmpresaQueryWithoutItemCountsDoesNotReferenceAlias(t *testing.T) {
	query, args := buildCarritosCompraByEmpresaQuery(32, true, "", true, false)
	if len(args) != 1 || args[0] != int64(32) {
		t.Fatalf("args=%v, want empresa_id only", args)
	}
	if containsSQLToken(query, "ic.") {
		t.Fatalf("query without item counts must not reference ic alias: %s", query)
	}
}

func TestValidateCarritoCompraItemCantidadAllowsWeightDecimalsOnlyForWeightUnits(t *testing.T) {
	if err := validateCarritoCompraItemCantidad(0.375, "kg"); err != nil {
		t.Fatalf("kg decimal must be valid: %v", err)
	}
	if err := validateCarritoCompraItemCantidad(250, "g"); err != nil {
		t.Fatalf("gram quantity must be valid: %v", err)
	}
	if err := validateCarritoCompraItemCantidad(1.5, "unidad"); err == nil {
		t.Fatalf("unit decimal must be rejected")
	}
	if err := validateCarritoCompraItemCantidad(2, "unidad"); err != nil {
		t.Fatalf("integer unit quantity must be valid: %v", err)
	}
}

func TestCarritosPostgresFechaExpressionsCastToTextBeforeCoalesce(t *testing.T) {
	raw, err := os.ReadFile("carritos_compras.go")
	if err != nil {
		t.Fatalf("read carritos_compras.go: %v", err)
	}
	src := string(raw)
	for _, want := range []string{
		"COALESCE(CAST(activado_en AS TEXT), '')",
		"COALESCE(CAST(pagado_en AS TEXT), '')",
		"COALESCE(NULLIF(?, ''), CAST(CURRENT_TIMESTAMP AS TEXT))",
		"COALESCE(CAST(fecha_evento AS TEXT), CAST(fecha_creacion AS TEXT), CAST(CURRENT_TIMESTAMP AS TEXT))",
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("falta conversion compatible con PostgreSQL: %s", want)
		}
	}
}

func TestBuildCarritoSaleOperationCodeSeparatesReusableCartSessions(t *testing.T) {
	first := BuildCarritoSaleOperationCode("VENTA-DIRECTA-12", 12, 117, "2026-08-26 00:21:51.175963-05")
	second := BuildCarritoSaleOperationCode("VENTA-DIRECTA-12", 12, 117, "2026-08-26 00:25:09.100001-05")
	if first == second {
		t.Fatalf("reused cart sales must have distinct operation codes: %q", first)
	}
	if again := BuildCarritoSaleOperationCode("VENTA-DIRECTA-12", 12, 117, "2026-08-26 00:21:51.175963-05"); again != first {
		t.Fatalf("same paid timestamp must generate a stable code: got %q want %q", again, first)
	}
	for _, want := range []string{"VENTA-DIRECTA-12", "CRT-117", "PG-2026082600215117596305"} {
		if !strings.Contains(first, want) {
			t.Fatalf("operation code %q does not contain %q", first, want)
		}
	}
}

func TestResolveCarritoAttentionDurationSecondsAcceptsPostgresFractionAndTimezone(t *testing.T) {
	got := ResolveCarritoAttentionDurationSeconds(
		"2026-08-26 00:17:07.000000-05",
		"2026-08-26 00:21:51.175963-05",
	)
	if got != 284 {
		t.Fatalf("duration=%d, want 284 seconds", got)
	}
}

func TestCarritoStationMetricInsertArgumentsMatchQuery(t *testing.T) {
	input, err := normalizeCarritoStationMetricInput(CarritoStationMetricInput{
		EmpresaID:       12,
		CarritoID:       117,
		EventoOperacion: "venta_pagada",
		PagadoEn:        "2026-08-26 00:21:51.175963-05:00",
	})
	if err != nil {
		t.Fatalf("normalize metric input: %v", err)
	}
	args := carritoStationMetricInsertArgs(input)
	if got, want := len(args), strings.Count(insertCarritoStationMetricQuery, "?"); got != want {
		t.Fatalf("metric insert args=%d, placeholders=%d", got, want)
	}
	if input.DetalleItemsJSON != "[]" {
		t.Fatalf("empty detail must be normalized to JSON array, got %q", input.DetalleItemsJSON)
	}
}

func TestEmpresaCatalogIncludesImmutableCartSaleHistoryMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260826-002-cart-sale-history-v1" {
			continue
		}
		if migration.Apply == nil || migration.Body != empresaCarritoSaleHistorySchemaFingerprint {
			t.Fatal("cart sale history migration must be executable and immutable")
		}
		return
	}
	t.Fatal("cart sale history migration is missing from empresa catalog")
}

func containsSQLToken(query, token string) bool {
	return strings.Contains(strings.ToLower(query), strings.ToLower(token))
}
