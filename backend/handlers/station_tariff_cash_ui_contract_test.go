package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEveryStationCanExposeCashAndTariffActions(t *testing.T) {
	files := map[string][]string{
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html"): {
			"funciona_como_caja", "data-caja-idx", "data-caja-codigo-idx", "Funciona como caja",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "estaciones.html"): {
			"data-station-tarifa", "Cambiar tarifa", "data-station-caja", "Abrir caja", "abrir_cambio_tarifa=1", "caja_codigo=",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html"): {
			"stationCashCode", "abrir_cambio_tarifa", "openCarritoTarifaPanel", "defaultPaymentCashRegisterCode",
		},
	}
	for path, markers := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source := string(raw)
		for _, marker := range markers {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s no contiene %q", path, marker)
			}
		}
	}
}

func TestCashboxStationScopeAndActivationOnlyUIContracts(t *testing.T) {
	files := map[string][]string{
		filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html"): {
			"Estaciones visibles", "data-caja-limitar", "data-caja-estaciones", "solo_activar", "parseStationRangeList",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "administrar_usuarios.html"): {
			"stationAccessCashbox", "caja_codigo", "Caja asignada", "Solo activar",
		},
		filepath.Join("..", "..", "web", "administrar_empresa", "estaciones.html"): {
			"isActivationOnlyCashbox", "allowedGroups.every", "caja_codigo", "solo_activar",
		},
	}
	for path, markers := range files {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		source := string(raw)
		for _, marker := range markers {
			if !strings.Contains(source, marker) {
				t.Fatalf("%s no contiene %q", path, marker)
			}
		}
	}
}
