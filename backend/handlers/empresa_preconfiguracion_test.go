package handlers

import "testing"

func TestEmpresaPreconfigHasStations(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{name: "empty", raw: ``, want: false},
		{name: "invalid json", raw: `{`, want: false},
		{name: "without stations", raw: `{"cantidad":0}`, want: false},
		{name: "empty stations", raw: `{"estaciones":[]}`, want: false},
		{name: "configured station", raw: `{"estaciones":[{"id":1,"nombre":"Habitacion 1"}]}`, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := empresaPreconfigHasStations(tt.raw); got != tt.want {
				t.Fatalf("empresaPreconfigHasStations(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
