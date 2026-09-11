package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStationAIOrdersSurfaceIsRetired(t *testing.T) {
	retiredFiles := []string{
		filepath.Join("ia_pedidos_estacion.go"),
		filepath.Join("..", "..", "web", "administrar_empresa", "estacion_ia_pedidos.html"),
	}
	for _, path := range retiredFiles {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("retired station AI orders file still exists: %s", path)
		}
	}

	forbidden := []string{
		"ia_pedidos_enabled",
		"ia_pedidos_placement",
		"ia_pedidos_estacion",
		"estacion_ia_pedidos",
	}
	files := []string{
		filepath.Join("chat_con_inteligencia_artificial_router.go"),
		filepath.Join("empresa_permisos.go"),
		filepath.Join("empresa_preconfiguracion.go"),
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html"),
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_carrito_de_compra_empresa.html"),
		filepath.Join("..", "..", "web", "administrar_empresa", "estaciones.html"),
		filepath.Join("..", "..", "web", "js", "ai_chat_drawer.js"),
	}
	for _, path := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source := strings.ToLower(string(raw))
		for _, marker := range forbidden {
			if strings.Contains(source, marker) {
				t.Fatalf("%s still contains retired marker %q", path, marker)
			}
		}
	}
}

func TestStationCardClientFlagAndCartTopActionsContract(t *testing.T) {
	configurationPath := filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html")
	configurationRaw, err := os.ReadFile(configurationPath)
	if err != nil {
		t.Fatalf("read %s: %v", configurationPath, err)
	}
	configuration := string(configurationRaw)
	for _, marker := range []string{"masterMostrarCliente", "data-mostrar-cliente-idx", "mostrar_cliente_nombre"} {
		if !strings.Contains(configuration, marker) {
			t.Fatalf("station configuration does not contain %q", marker)
		}
	}

	stationsPath := filepath.Join("..", "..", "web", "administrar_empresa", "estaciones.html")
	stationsRaw, err := os.ReadFile(stationsPath)
	if err != nil {
		t.Fatalf("read %s: %v", stationsPath, err)
	}
	if !strings.Contains(string(stationsRaw), "cfg.mostrar_cliente_nombre !== false") {
		t.Fatal("station card does not apply its per-station client visibility flag")
	}

	cartPath := filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html")
	cartRaw, err := os.ReadFile(cartPath)
	if err != nil {
		t.Fatalf("read %s: %v", cartPath, err)
	}
	cart := string(cartRaw)
	searchIndex := strings.Index(cart, "Buscar producto")
	selectIndex := strings.Index(cart, `id="carritoActionSelect"`)
	domoticsIndex := strings.Index(cart, `id="carritoBtnControlElectrico"`)
	fullscreenIndex := strings.Index(cart, `id="directSaleFullscreenCard"`)
	if searchIndex < 0 || selectIndex < searchIndex || domoticsIndex < selectIndex || fullscreenIndex < domoticsIndex {
		t.Fatal("cart action selector and Domotics button are not ordered beside the product search heading")
	}
	if strings.Count(cart, `id="carritoActionSelect"`) != 1 || strings.Count(cart, `id="carritoBtnControlElectrico"`) != 1 {
		t.Fatal("cart action selector or Domotics button has duplicate DOM ids")
	}

	stylesPath := filepath.Join("..", "..", "web", "estilos.css")
	stylesRaw, err := os.ReadFile(stylesPath)
	if err != nil {
		t.Fatalf("read %s: %v", stylesPath, err)
	}
	styles := strings.ReplaceAll(string(stylesRaw), "\r\n", "\n")
	if strings.Contains(styles, "body.carrito-flat-page #stationCartActionsCard{\n  display:none") {
		t.Fatal("cart action panels cannot open because their parent is permanently hidden")
	}
}
