package main

import (
	"encoding/json"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaMenuVisualDefaultsFrontendContract(t *testing.T) {
	read := func(parts ...string) string {
		raw, err := os.ReadFile(filepath.Join(append([]string{".."}, parts...)...))
		if err != nil {
			t.Fatalf("read %s: %v", filepath.Join(parts...), err)
		}
		return string(raw)
	}
	defaultsJS := read("web", "js", "menu_visual_defaults.js")
	adminHTML := read("web", "administrar_empresa.html")
	adminJS := read("web", "js", "administrar_empresa.js")
	configHTML := read("web", "administrar_empresa", "configuracion", "menu_visual.html")
	configMenuHTML := read("web", "administrar_empresa", "configuracion_menu.html")
	selectorJS := read("web", "js", "seleccionar_empresa.js")
	panelHTML := read("web", "administrar_empresa", "panel.html")

	var config struct {
		HiddenLinks []string `json:"hidden_links"`
	}
	if err := json.Unmarshal([]byte(dbpkg.EmpresaMenuVisualDefaultConfigJSON), &config); err != nil {
		t.Fatal(err)
	}
	hidden := make(map[string]bool, len(config.HiddenLinks))
	for _, id := range config.HiddenLinks {
		hidden[id] = true
		if !strings.Contains(defaultsJS, `"`+id+`"`) {
			t.Fatalf("frontend default catalog missing %s", id)
		}
	}

	groupPattern := regexp.MustCompile(`(?s)<li class="admin-nav-group[^\"]*">\s*<button[^>]*>(.*?)</button>\s*<ul[^>]*>(.*?)</ul>\s*</li>`)
	linkPattern := regexp.MustCompile(`<a id="(link[^"]+)"`)
	hiddenGroups := map[string]bool{
		"Producción": true, "CRM y clientes": true, "Usuarios, clientes y personas": true,
		"Control de asistencia y horarios": true,
	}
	visibleGroups := map[string]bool{
		"Operación y ventas": true, "Inventario y compras": true, "Finanzas y cumplimiento": true,
		"Canales digitales y colaboración": true, "Ubicación GPS": true, "Análisis y control": true,
		"Domótica y Energía Solar": true, "Documentos, nube y soporte": true, "Administración": true,
		"Licencia": true,
	}
	seenGroups := map[string]bool{}
	for _, match := range groupPattern.FindAllStringSubmatch(adminHTML, -1) {
		groupName := strings.TrimSpace(html.UnescapeString(regexp.MustCompile(`<[^>]+>`).ReplaceAllString(match[1], "")))
		seenGroups[groupName] = true
		for _, linkMatch := range linkPattern.FindAllStringSubmatch(match[2], -1) {
			id := linkMatch[1]
			if hiddenGroups[groupName] && !hidden[id] {
				t.Fatalf("default-hidden group %q leaves %s visible", groupName, id)
			}
			if visibleGroups[groupName] && hidden[id] {
				t.Fatalf("default-visible group %q hides %s", groupName, id)
			}
		}
	}
	for groupName := range hiddenGroups {
		if !seenGroups[groupName] {
			t.Fatalf("default-hidden group missing from admin menu: %q", groupName)
		}
	}
	for groupName := range visibleGroups {
		if !seenGroups[groupName] {
			t.Fatalf("default-visible group missing from admin menu: %q", groupName)
		}
	}
	for _, id := range []string{"linkPanelEmpresa", "linkVida", "linkNoticias", "linkVolverEmpresas"} {
		if hidden[id] {
			t.Fatalf("required standalone menu link is hidden by default: %s", id)
		}
	}
	for _, removed := range []string{"Soluciones por negocio", "adminBusinessVerticalsMount", "linkParqueadero", "linkEventosBoleteria"} {
		if strings.Contains(adminHTML, removed) || strings.Contains(configHTML, removed) {
			t.Fatalf("business solutions menu must not exist: %s", removed)
		}
	}
	for _, removed := range []string{"linkPlantillasIntegracion", "linkConfiguracionGuiada", "Adaptacion por tipo", "Configuracion guiada"} {
		if strings.Contains(configMenuHTML, removed) {
			t.Fatalf("business-type setup must not remain in the regular configuration menu: %s", removed)
		}
	}
	for _, required := range []string{"setConfigurationAssistantPending(createData.id, preconfig)", "tipo_empresa_nombre"} {
		if !strings.Contains(selectorJS, required) {
			t.Fatalf("new-company business setup trigger missing: %s", required)
		}
	}
	for _, required := range []string{"hasGuidedPendingFlag()", "!estado.configurada_anterior", "/api/empresa/configuracion_guiada"} {
		if !strings.Contains(panelHTML, required) {
			t.Fatalf("first-entry guided setup contract missing: %s", required)
		}
	}
	for _, required := range []string{
		`/js/menu_visual_defaults.js?v=20260914-menu-defaults-v3`,
		`/js/administrar_empresa.js?v=20260914-menu-defaults-v3`,
	} {
		if !strings.Contains(adminHTML, required) {
			t.Fatalf("admin shell missing %q", required)
		}
	}
	for _, required := range []string{
		"data-menu-group",
		"Mostrar grupo",
		"btnAplicarPredeterminado",
		"Aplicar selección predeterminada",
		"const DEFAULT_CONFIG = window.PCS_MENU_VISUAL_DEFAULTS",
		"Noticias Beta",
	} {
		if !strings.Contains(configHTML, required) {
			t.Fatalf("menu configuration missing %q", required)
		}
	}
	if strings.Contains(adminJS, "isSuperAdminMenuContext()) return false") {
		t.Fatal("super administrators must see the same company menu selection")
	}
	for _, recoverable := range []string{"linkConfiguracion: true", "linkVolverEmpresas: true"} {
		if !strings.Contains(adminJS, recoverable) || !strings.Contains(configHTML, recoverable) {
			t.Fatalf("recovery link must remain non-hideable: %s", recoverable)
		}
	}
}
