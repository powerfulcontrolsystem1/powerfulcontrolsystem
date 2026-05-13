package db

import (
	"errors"
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

func containsSQLToken(query, token string) bool {
	return strings.Contains(strings.ToLower(query), strings.ToLower(token))
}
