package handlers

import (
	"reflect"
	"testing"
)

func TestParseLicenciaVencimientoDiasAviso(t *testing.T) {
	got := parseLicenciaVencimientoDiasAviso("7, 15;3 1,7,0,400,x")
	want := []int{15, 7, 3, 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dias aviso = %v, want %v", got, want)
	}
}

func TestSelectLicenciaVencimientoDiaAviso(t *testing.T) {
	thresholds := []int{15, 7, 3, 1}
	cases := []struct {
		name string
		days int
		want int
	}{
		{name: "inside upper window", days: 10, want: 15},
		{name: "exact threshold", days: 7, want: 7},
		{name: "near expiry", days: 2, want: 3},
		{name: "today", days: 0, want: 1},
		{name: "outside window", days: 16, want: 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := selectLicenciaVencimientoDiaAviso(tc.days, thresholds); got != tc.want {
				t.Fatalf("threshold = %d, want %d", got, tc.want)
			}
		})
	}
}
