package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestControlElectricoStationOperationUsesVentasAuthorizationOnly(t *testing.T) {
	tests := []struct {
		method string
		action string
		module string
		perm   string
		ok     bool
	}{
		{http.MethodGet, "estacion_controls", permModuleVentas, permActionRead, true},
		{http.MethodPost, "probar_rele", permModuleVentas, permActionUpdate, true},
		{http.MethodPost, "temporizador_rele", permModuleVentas, permActionUpdate, true},
		{http.MethodGet, "config", "", "", false},
		{http.MethodPut, "rele", "", "", false},
		{http.MethodPost, "probar_gpio", "", "", false},
		{http.MethodGet, "reles", "", "", false},
	}
	for _, test := range tests {
		r := httptest.NewRequest(test.method, "/api/empresa/control_electrico?action="+test.action, nil)
		module, action, ok := resolveControlElectricoStationOperationAuthorization(r)
		if ok != test.ok || module != test.module || action != test.perm {
			t.Fatalf("%s %s: got (%q, %q, %v), want (%q, %q, %v)", test.method, test.action, module, action, ok, test.module, test.perm, test.ok)
		}
	}
}

func TestControlElectricoStationOperationRejectsOtherEndpoint(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/empresa/otro?action=estacion_controls", nil)
	if _, _, ok := resolveControlElectricoStationOperationAuthorization(r); ok {
		t.Fatal("estacion_controls no debe relajar permisos fuera del endpoint de control electrico")
	}
}

func TestControlElectricoStationOperationalPolicyRespectsCashboxScope(t *testing.T) {
	allowed := carritoStationAccessPolicy{Enabled: true, LimitStations: true, Stations: map[int64]bool{1: true}}
	if err := validateControlElectricoStationOperationalPolicy(allowed, 1); err != nil {
		t.Fatalf("estacion permitida rechazada: %v", err)
	}
	if err := validateControlElectricoStationOperationalPolicy(allowed, 2); err != errCarritoStationAccessDenied {
		t.Fatalf("estacion fuera del rango: got %v, want %v", err, errCarritoStationAccessDenied)
	}
	activationOnly := carritoStationAccessPolicy{Enabled: true, ActivationOnly: true, Stations: map[int64]bool{1: true}}
	if err := validateControlElectricoStationOperationalPolicy(activationOnly, 1); err != errControlElectricoActivationOnly {
		t.Fatalf("caja solo activar debe rechazar domotica: got %v", err)
	}
}
