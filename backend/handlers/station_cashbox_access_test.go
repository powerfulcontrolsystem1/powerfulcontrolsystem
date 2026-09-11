package handlers

import (
	"net/http"
	"testing"
)

func TestApplyCarritoCashboxPolicyIntersectsUserAndCashboxStations(t *testing.T) {
	policy := carritoStationAccessPolicy{
		Enabled:       true,
		LimitStations: true,
		AllowCaja:     true,
		Stations:      map[int64]bool{3: true, 4: true, 6: true},
	}
	root := map[string]interface{}{
		"cajas_config": []interface{}{
			map[string]interface{}{
				"codigo":             "CAJA-1",
				"limitar_estaciones": true,
				"estaciones":         []interface{}{float64(1), float64(2), float64(3), float64(4), float64(5)},
				"modo_estaciones":    "operar",
			},
		},
	}
	entry := map[string]interface{}{"caja_codigo": "caja-1"}

	got := applyCarritoCashboxPolicy(policy, root, entry)
	if !got.LimitStations || len(got.Stations) != 2 || !got.Stations[3] || !got.Stations[4] {
		t.Fatalf("la interseccion esperada era {3,4}, resultado: %+v", got.Stations)
	}
	if got.ActivationOnly || !got.AllowCaja || got.CajaCode != "CAJA-1" {
		t.Fatalf("politica operativa inesperada: %+v", got)
	}
}

func TestApplyCarritoCashboxPolicySupportsActivationOnly(t *testing.T) {
	policy := carritoStationAccessPolicy{Enabled: true, AllowCaja: true, Stations: map[int64]bool{}}
	root := map[string]interface{}{
		"cajas_config": []interface{}{
			map[string]interface{}{
				"codigo":             "PORTERIA",
				"limitar_estaciones": true,
				"estaciones":         []interface{}{float64(1), float64(2), float64(6)},
				"modo_estaciones":    "solo_activar",
			},
		},
	}

	got := applyCarritoCashboxPolicy(policy, root, map[string]interface{}{"caja_codigo": "PORTERIA"})
	if !got.ActivationOnly || got.AllowCaja || !got.LimitStations || len(got.Stations) != 3 {
		t.Fatalf("politica de porteria inesperada: %+v", got)
	}
	if isActivationOnlyPolicyRestricted(got, http.MethodGet, "") {
		t.Fatal("porteria debe poder consultar el tablero")
	}
	if isActivationOnlyPolicyRestricted(got, http.MethodPut, "activar_estacion") {
		t.Fatal("porteria debe poder activar estaciones")
	}
	for _, request := range []struct{ method, action string }{
		{http.MethodGet, "totales_pago"},
		{http.MethodPut, "pagar_estacion"},
		{http.MethodPost, ""},
		{http.MethodDelete, ""},
	} {
		if !isActivationOnlyPolicyRestricted(got, request.method, request.action) {
			t.Fatalf("porteria no debe permitir %s action=%q", request.method, request.action)
		}
	}
}

func TestApplyCarritoCashboxPolicyKeepsLegacyUserOnlyConfiguration(t *testing.T) {
	policy := carritoStationAccessPolicy{
		Enabled:       true,
		LimitStations: true,
		AllowCaja:     true,
		Stations:      map[int64]bool{7: true, 8: true},
	}
	got := applyCarritoCashboxPolicy(policy, map[string]interface{}{}, map[string]interface{}{})
	if got.CajaCode != "" || got.ActivationOnly || len(got.Stations) != 2 || !got.Stations[7] || !got.Stations[8] {
		t.Fatalf("la configuracion historica por usuario debe conservarse: %+v", got)
	}
}
