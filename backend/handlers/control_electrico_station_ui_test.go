package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDomoticaStationUIHasSingleEntryAndVisibilityCheck(t *testing.T) {
	stationPage, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "estaciones.html"))
	if err != nil {
		t.Fatal(err)
	}
	stationConfig, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html"))
	if err != nil {
		t.Fatal(err)
	}
	stationSource := string(stationPage)
	configSource := string(stationConfig)
	for _, marker := range []string{"stationCardUI", "mostrar_boton_domotica", "estacion_id", "handleStationCardActivation"} {
		if !strings.Contains(stationSource, marker) {
			t.Fatalf("la pagina de estaciones no contiene %q", marker)
		}
	}
	for _, marker := range []string{"stShowDomoticaButton", "mostrar_boton_domotica", "equipos y sensores asociados"} {
		if !strings.Contains(configSource, marker) {
			t.Fatalf("la configuracion de estaciones no contiene %q", marker)
		}
	}
	if strings.Contains(stationSource, "data-open-station-domotica") || strings.Contains(stationSource, "station-domotica-button") {
		t.Fatal("la pagina de estaciones no debe mostrar un boton Domotica dentro de la tarjeta")
	}
	cartPage, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html"))
	if err != nil {
		t.Fatal(err)
	}
	cartSource := string(cartPage)
	for _, marker := range []string{"carrito-action-select-row", "id=\"carritoBtnControlElectrico\"", "aria-label=\"Abrir Domótica de esta estación\""} {
		if !strings.Contains(cartSource, marker) {
			t.Fatalf("el carrito no contiene %q", marker)
		}
	}
}

func TestDomoticaStationPanelShowsDevicesSensorsAndMultipleRaspberry(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "carrito_control_electrico.html"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(page)
	for _, marker := range []string{
		"payload.reles",
		"payload.sensores",
		"payload.eventos",
		"payload.raspberry_pis",
		"Raspberry conectadas",
		"sensor_input",
		"data-control-toggle",
		"raspberryStatusHTML",
		"device-toggle-action",
		"programado",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("el panel operativo de estacion no contiene %q", marker)
		}
	}
}
