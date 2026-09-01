package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestDomoticaTunnelSeenRecentlyAcceptsPostgresOffsetWithoutColon(t *testing.T) {
	bogota := time.FixedZone("COT", -5*60*60)
	recent := time.Now().In(bogota).Add(-time.Second).Format("2006-01-02 15:04:05.999999-07")
	if !domoticaTunnelSeenRecently(recent, 90*time.Second) {
		t.Fatalf("heartbeat PostgreSQL reciente con offset -05 marcado desconectado: %s", recent)
	}
}
