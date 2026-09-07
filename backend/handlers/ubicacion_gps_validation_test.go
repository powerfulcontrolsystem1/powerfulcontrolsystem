package handlers

import (
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestValidateGPSCoordinates(t *testing.T) {
	for _, tc := range []struct {
		name string
		lat  float64
		lng  float64
		want bool
	}{
		{name: "bogota", lat: 4.711, lng: -74.0721},
		{name: "limits", lat: 90, lng: -180},
		{name: "invalid_latitude", lat: 90.01, lng: 0, want: true},
		{name: "invalid_longitude", lat: 0, lng: 180.01, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := validateGPSCoordinates(tc.lat, tc.lng); (got != nil) != tc.want {
				t.Fatalf("validateGPSCoordinates(%v, %v) error=%v wantError=%v", tc.lat, tc.lng, got, tc.want)
			}
		})
	}
}

func TestValidateGPSTelemetry(t *testing.T) {
	valid := dbpkg.EmpresaGPSRecorrido{
		PrecisionMetros:   5,
		VelocidadKMH:      80,
		RumboGrados:       120,
		AltitudMetros:     2600,
		BateriaPorcentaje: 75,
		SenalPorcentaje:   88,
	}
	if err := validateGPSTelemetry(valid); err != nil {
		t.Fatalf("valid telemetry rejected: %v", err)
	}
	invalid := valid
	invalid.BateriaPorcentaje = 101
	if err := validateGPSTelemetry(invalid); err == nil {
		t.Fatal("invalid battery percentage accepted")
	}
}
