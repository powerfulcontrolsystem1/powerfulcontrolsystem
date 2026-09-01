package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDomoticaAndSolarHaveIndependentMainMenu(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	for _, marker := range []string{"Domótica y Energía Solar", "linkControlElectrico", "linkEquiposDomotica", "linkEnergiaSolar", "linkTutorialDomotica"} {
		if !strings.Contains(body, marker) {
			t.Fatalf("menu principal sin %q", marker)
		}
	}
}

func TestDomoticaConfigExposesDisconnectAlertsAndSensorModes(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "control_electrico.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	for _, marker := range []string{"disconnectAlertEnabled", "disconnectAlertEmail", "disconnectGraceMinutes", "encender_temporizado", "encender_programado", "ruleTimerSeconds"} {
		if !strings.Contains(body, marker) {
			t.Fatalf("configuracion de domotica sin %q", marker)
		}
	}
}

func TestDomoticaTutorialHasAccuratePinoutAndInstallFlow(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "tutorial_domotica.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	for _, marker := range []string{"raspberry-pi-40-pin-bcm.svg", "raspberry-pi-40-pin-tutorial.png", "GPIO2 está en el pin físico 3", "IP local es opcional", "encender por un temporizador", "programación guardada"} {
		if !strings.Contains(body, marker) {
			t.Fatalf("tutorial sin %q", marker)
		}
	}
	pinout, err := os.ReadFile(filepath.Join("..", "..", "web", "img", "raspberry-pi-40-pin-bcm.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(pinout), `class="pin `) != 40 {
		t.Fatalf("el diagrama no contiene exactamente 40 pines")
	}
}

func TestDoorSensorTutorialExplainsMatrixPinoutAndElectricalSafety(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "tutorial_sensores_puertas_raspberry.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	for _, marker := range []string{
		"GPIO0, GPIO1, GPIO2 y GPIO3",
		"GPIO4 a GPIO19",
		"OUT16 → puertas 61–64",
		"Los GPIO aceptan lógica de <strong>3,3 V</strong>",
		"módulo de 4 relés",
		"activo en alto",
		"0 como puerta cerrada",
		"sudo sh ~/Downloads/instalar-pcs-sensores-puertas-*.sh",
		"Raspberry Pi · GPIO BCM",
	} {
		if !strings.Contains(body, marker) {
			t.Fatalf("tutorial de sensores sin %q", marker)
		}
	}
	for _, page := range []string{"control_electrico.html", "configuracion_sensores_raspberry.html"} {
		linked, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", page))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(linked), "tutorial_sensores_puertas_raspberry.html") {
			t.Fatalf("%s no enlaza el tutorial de sensores", page)
		}
	}
	menu, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "modulo_menu.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(menu), "Tutorial sensores de puertas") || !strings.Contains(string(menu), "tutorial_sensores_puertas_raspberry.html") {
		t.Fatal("el submenu Domotica no expone el tutorial de sensores")
	}
}

func TestSuperDomoticaTrafficHasCompanyLimitsAndAbuseAlarm(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "web", "super", "domotica_raspberry_trafico.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(source)
	for _, marker := range []string{"Control y alarmas por empresa", "kpiAlerts", "abuseAlert", "policy-limit", "bloquear_al_superar", "method:'PUT'", "month_human"} {
		if !strings.Contains(body, marker) {
			t.Fatalf("panel super sin %q", marker)
		}
	}
}
