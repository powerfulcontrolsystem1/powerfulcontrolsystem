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
	for _, marker := range []string{"carrito-action-select-row", "gap:18px", "id=\"carritoBtnControlElectrico\"", "aria-label=\"Abrir Domótica de esta estación\"", "const estacionID = resolveCarritoEstacionID(selected)", "return_to: 'carrito'"} {
		if !strings.Contains(cartSource, marker) {
			t.Fatalf("el carrito no contiene %q", marker)
		}
	}
}

func TestDomoticaStationEntryPreferenceKeepsCartDefaultAndRedirectsCompanyMenu(t *testing.T) {
	files := map[string][]string{
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_carrito_de_compra_empresa.html"): {
			"carritoCfgAbrirDomoticaEstaciones",
			"abrir_domotica_al_entrar_estacion: false",
			"Abrir equipos electronicos al entrar a una estacion o a Venta directa",
			"pcs-station-entry-navigation-updated",
			"window.top && window.top !== window",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "estaciones.html"): {
			"function openStationOperationalDestination",
			"openStationDomotica(stationID, stationName)",
			"openStationOperationalDestination(stationID, stationName)",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html"): {
			"nextConfig.abrir_domotica_al_entrar_estacion = !!prev.abrir_domotica_al_entrar_estacion",
		},
		filepath.Join("..", "..", "web", "js", "administrar_empresa.js"): {
			"stationEntryDomoticaEnabled",
			"link.id === \"linkVentaDirecta\"",
			"target.pathname = \"/administrar_empresa/carrito_control_electrico.html\"",
			"target.searchParams.set(\"vista\", \"todas\")",
			"function refreshDirectSaleDestination",
		},
	}
	for path, markers := range files {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source := string(content)
		for _, marker := range markers {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s no contiene %q", path, marker)
			}
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
		"programacion_inicio",
		"programacion_fin",
		"datetime-local",
		"conectadas');",
		"device-footer.has-photo",
		"function timerInfo(releID)",
		"rele.ultimo_estado='off'",
		"const isOn = info.cls === 'is-on' || timer.active",
		"function scheduleState(rele)",
		"Programado · Funcionando ahora",
		"Programado · En espera",
		"function waitForRelayConfirmation(releID, estado)",
		"Comando enviado; esperando confirmación de la Raspberry",
		"El comando sigue en cola. Revisa la conexión de la Raspberry.",
		"function refreshLiveTimerLabels()",
		"touch-action: manipulation",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("el panel operativo de estacion no contiene %q", marker)
		}
	}
	if strings.Contains(source, ">Editar</a>") || strings.Contains(source, "control_electrico.html?empresa_id=") {
		t.Fatal("la pagina operativa de equipos no debe ofrecer edicion; debe quedar en configuracion")
	}
	if strings.Contains(source, "Modo: ' + escapeHtml(rele.modo") || strings.Contains(source, "raspberryStatusHTML(rele.raspberry_id)") {
		t.Fatal("las tarjetas no deben repetir modo ni estado textual de Raspberry")
	}
}

func TestSuperAdminMobileStartsWithCompanySelector(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "web", "super_administrador.html"))
	if err != nil {
		t.Fatalf("read super admin: %v", err)
	}
	source := string(content)
	selector := `<li class="admin-nav-standalone super-select-company-first">`
	panel := `<a href="/super/licencias_resumen.html"`
	if !strings.Contains(source, selector) || !strings.Contains(source, `href="/seleccionar_empresa.html"`) {
		t.Fatal("el menu super debe exponer Seleccionar empresa como acceso principal")
	}
	if strings.Index(source, selector) > strings.Index(source, panel) {
		t.Fatal("Seleccionar empresa debe aparecer antes del Panel en el menu movil del super administrador")
	}
}

func TestDomoticaConfigurationExposesSafeRaspberryOperations(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "control_electrico.html"))
	if err != nil {
		t.Fatalf("read domotica configuration: %v", err)
	}
	for _, marker := range []string{"raspberry-connection-test", "raspberry-restart", "raspberry-shutdown", "api('raspberry_operacion')", "window.confirm"} {
		if !strings.Contains(string(content), marker) {
			t.Fatalf("domotica configuration missing raspberry operation marker %q", marker)
		}
	}
}

func TestDomoticaRaspberryConfigHasSafeGPIODiagnostic(t *testing.T) {
	page, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "control_electrico.html"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(page)
	for _, marker := range []string{"raspberry-gpio-toggle", "raspberry-gpio-test", "api('probar_gpio')", "pulso de prueba de un segundo"} {
		if !strings.Contains(source, marker) {
			t.Fatalf("la configuracion de Raspberry no contiene %q", marker)
		}
	}
	for _, marker := range []string{"activationDelaySeconds", "activation_delay_seconds", "Cola única por empresa"} {
		if !strings.Contains(source, marker) {
			t.Fatalf("la configuracion no contiene %q", marker)
		}
	}
	for _, marker := range []string{"Array.from({ length: 28 }", "GPIO 0 y GPIO 1 corresponden a ID_SDA/ID_SCL", "cableado verificado", "assignedPins", "hasConfiguredGPIO", "Selecciona una Raspberry", "Sin Raspberry asignada"} {
		if !strings.Contains(source, marker) {
			t.Fatalf("la configuracion no conserva GPIO 0 ni la lista neutral de Raspberry: falta %q", marker)
		}
	}
	for _, legacy := range []string{"Raspberry principal/global", "disabled>Principal</button>"} {
		if strings.Contains(source, legacy) {
			t.Fatalf("la configuracion no debe clasificar controladores como principal/secundario: encontro %q", legacy)
		}
	}
}

func TestDomoticaSummaryHidesInactiveRaspberryControllers(t *testing.T) {
	content, err := os.ReadFile("control_electrico.go")
	if err != nil {
		t.Fatalf("read domotica handler: %v", err)
	}
	source := string(content)
	for _, marker := range []string{
		"raspberries, err := dbpkg.ListEmpresaControlElectricoRaspberryContext(r.Context(), dbEmp, empresaID, false)",
		"rows, _ := dbpkg.ListEmpresaControlElectricoRaspberry(dbEmp, empresaID, false)",
	} {
		if !strings.Contains(source, marker) {
			t.Fatalf("la respuesta operativa de Domotica debe ocultar Raspberry inactivas: falta %q", marker)
		}
	}
}
