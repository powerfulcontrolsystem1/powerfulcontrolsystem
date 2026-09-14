package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func readFrontendContractFile(t *testing.T, parts ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{"..", "web"}, parts...)...)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func TestPublicPortalUsesSevenPreconfigurationsAndKeepsTallerMotos(t *testing.T) {
	catalog := readFrontendContractFile(t, "js", "preconfiguraciones_basicas_catalogo.js")
	featuredStart := strings.Index(catalog, "var featured = [")
	if featuredStart < 0 {
		t.Fatal("public catalog must separate featured systems from basic preconfigurations")
	}
	basicCatalog := catalog[:featuredStart]
	modulePattern := regexp.MustCompile(`module:\s*"([^"]+)"`)
	matches := modulePattern.FindAllStringSubmatch(basicCatalog, -1)
	want := []string{"hotel", "motel", "restaurante", "bar", "pymes", "salon_belleza", "lavadero_autos"}
	if len(matches) != len(want) {
		t.Fatalf("basic catalog must contain exactly seven modules, got %d", len(matches))
	}
	for i, key := range want {
		if matches[i][1] != key {
			t.Fatalf("basic catalog item %d = %q, want %q", i, matches[i][1], key)
		}
	}
	if !strings.Contains(catalog[featuredStart:], `module: "taller_mecanico"`) || !strings.Contains(catalog[featuredStart:], `title: "Taller de motos"`) {
		t.Fatal("Taller de motos must remain as the single featured business system")
	}

	for _, page := range []string{"index.html", "descripcion_de_los_sistemas.html"} {
		html := readFrontendContractFile(t, page)
		if !strings.Contains(html, "/js/preconfiguraciones_basicas_catalogo.js") {
			t.Fatalf("%s must load the current public catalog", page)
		}
		if strings.Contains(html, "/js/plantillas_nuevas_catalogo.js") || strings.Contains(html, "/api/public/plantillas_nuevas/catalogo") {
			t.Fatalf("%s still loads the retired mass-template catalog", page)
		}
		for _, retired := range []string{"Domicilios y entregas", "Veterinaria y pet shop", "Alquileres de activos", "Parqueaderos con ticket QR", "Lavanderia y tintoreria", "Parque recreativo"} {
			if strings.Contains(html, retired) {
				t.Fatalf("%s still advertises retired card %q", page, retired)
			}
		}
	}
}

func TestAdminCompanyNavigationStaysInsideContentFrame(t *testing.T) {
	admin := readFrontendContractFile(t, "administrar_empresa.html")
	anchorPattern := regexp.MustCompile(`<a id="(link[^"]+)"[^>]*>`)
	matches := anchorPattern.FindAllStringSubmatch(admin, -1)
	if len(matches) == 0 {
		t.Fatal("admin sidebar links were not found")
	}
	for _, match := range matches {
		anchor := match[0]
		id := match[1]
		if id == "linkVolverEmpresas" {
			continue
		}
		if !strings.Contains(anchor, `target="contentFrame"`) {
			t.Fatalf("admin navigation %s must stay inside contentFrame: %s", id, anchor)
		}
	}
	if strings.Contains(admin, `target="_blank"`) {
		t.Fatal("Administrar empresa sidebar must not open business pages in a new tab")
	}
	for _, expected := range []string{
		`id="linkEmailCorporativo" href="/administrar_empresa/email_corporativo.html" target="contentFrame"`,
		`id="linkPortalUsuarios" href="/login_usuario.html" target="contentFrame"`,
		`name="contentFrame" id="contentFrame"`,
	} {
		if !strings.Contains(admin, expected) {
			t.Fatalf("admin frame contract is missing %q", expected)
		}
	}

	adminJS := readFrontendContractFile(t, "js", "administrar_empresa.js")
	if strings.Contains(adminJS, `portalUsuariosLink.target = "_blank"`) || strings.Contains(adminJS, `window.open(buildPortalUsuariosURL`) {
		t.Fatal("user portal must not escape the Administrar empresa shell")
	}
	if !strings.Contains(adminJS, "frame.src = url;") {
		t.Fatal("dynamic user portal URL must be assigned to contentFrame")
	}

	menu := readFrontendContractFile(t, "menu.js")
	for _, expected := range []string{
		"window.open('/calculadora.html?compact=1', 'pcs_calculadora'",
		"window.open('/juegos.html', 'pcs_juegos'",
	} {
		if !strings.Contains(menu, expected) {
			t.Fatalf("allowed accessory popup is missing: %s", expected)
		}
	}
	if got := strings.Count(menu, "window.open('/"); got != 2 {
		t.Fatalf("only Calculadora and Juegos may open same-origin popup windows; got %d literal launchers", got)
	}
}

func TestKeyPageContractsExposeLightAndDarkAppearanceNodes(t *testing.T) {
	pages := [][]string{
		{"index.html"},
		{"descripcion_de_los_sistemas.html"},
		{"administrar_empresa.html"},
		{"administrar_empresa", "email_corporativo.html"},
		{"calculadora.html"},
		{"juegos.html"},
		{"juegos", "creditos.html"},
	}
	for _, parts := range pages {
		html := readFrontendContractFile(t, parts...)
		name := strings.Join(parts, "/")
		if !strings.Contains(html, `data-appearance-contract="light-dark"`) {
			t.Fatalf("%s must declare its light/dark appearance contract", name)
		}
		if !strings.Contains(html, `data-appearance-node="`) {
			t.Fatalf("%s must expose semantic appearance nodes", name)
		}
	}

	bootstrap := readFrontendContractFile(t, "js", "theme_bootstrap.js")
	for _, expected := range []string{`"data-appearance-mode"`, `light ? "light" : "dark"`, `root.style.colorScheme`} {
		if !strings.Contains(bootstrap, expected) {
			t.Fatalf("theme bootstrap is missing %q", expected)
		}
	}
	gamesCSS := readFrontendContractFile(t, "juegos", "sala.css")
	if !strings.Contains(gamesCSS, `html[data-appearance-mode="light"]`) || !strings.Contains(gamesCSS, `html[data-appearance-mode="dark"]`) {
		t.Fatal("PCS PLAY must define professional surfaces for both appearance modes")
	}
	gamesJS := readFrontendContractFile(t, "juegos", "sala.js")
	for _, node := range []string{"game-card", "toolbar", "controls", "status", "card"} {
		if !strings.Contains(gamesJS, `'data-appearance-node','`+node+`'`) && !strings.Contains(gamesJS, `data-appearance-node="`+node+`"`) {
			t.Fatalf("PCS PLAY is missing appearance node %q", node)
		}
	}
}
