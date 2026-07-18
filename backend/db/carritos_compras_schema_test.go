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

func containsSQLToken(query, token string) bool {
	return strings.Contains(strings.ToLower(query), strings.ToLower(token))
}
