package handlers

import (
	"encoding/json"
	"testing"
)

func TestDefaultEmpresaPreconfigCarritoUISimplificado(t *testing.T) {
	cfg := defaultEmpresaPreconfigCarritoUI()

	wantTrue := []string{
		"mostrar_boton_buscar_productos",
		"mostrar_busqueda_catalogo",
		"mostrar_codigo_manual_item",
		"mostrar_observaciones_item",
		"mostrar_selector_cliente",
		"mostrar_impuestos_item",
		"mostrar_lector_codigo_barras",
		"permitir_pago_mixto",
		"mostrar_resumen_totales_carrito",
		"mostrar_resumen_productos",
		"mostrar_boton_pagar",
		"mostrar_tarjetas_pago",
		"mostrar_tarjeta_lector_codigo",
		"mostrar_tarjeta_items_carrito",
		"mostrar_tarjeta_totales_detalles",
		"mostrar_tarjeta_acciones_carrito",
		"mostrar_control_electrico_carrito",
		"mostrar_tarjeta_valores_pago",
		"mostrar_tarjeta_vip_cliente",
	}
	for _, key := range wantTrue {
		if got, _ := cfg[key].(bool); !got {
			t.Fatalf("%s debe quedar activo por defecto", key)
		}
	}

	wantFalse := []string{
		"modo_pantalla_tactil",
		"mostrar_descuentos",
		"mostrar_propina",
		"mostrar_comision",
		"mostrar_desglose_cobro",
		"mostrar_tarjeta_cobro_estados",
		"mostrar_tarjeta_comision",
	}
	for _, key := range wantFalse {
		if got, _ := cfg[key].(bool); got {
			t.Fatalf("%s debe quedar apagado por defecto", key)
		}
	}
}

func TestApplyDefaultCarritoUIPresetToConfigActualizaEmpresasViejas(t *testing.T) {
	raw := `{
		"cantidad": 2,
		"card_size": "medium",
		"station_card_ui": {"mostrar_total": true},
		"carrito_ui_global": {
			"mostrar_descuentos": true,
			"mostrar_tarjeta_cobro_estados": true,
			"mostrar_tarjeta_comision": true
		},
		"estaciones": [
			{
				"id": 1,
				"nombre": "Estacion 1",
				"carrito": {
					"usar_configuracion_global": false,
					"configuracion": {
						"mostrar_descuentos": true,
						"mostrar_tarjeta_cobro_estados": true
					}
				}
			}
		]
	}`

	nextRaw, changed, err := applyDefaultCarritoUIPresetToConfig(raw, defaultEmpresaPreconfigCarritoUI())
	if err != nil {
		t.Fatalf("applyDefaultCarritoUIPresetToConfig error: %v", err)
	}
	if !changed {
		t.Fatal("se esperaba cambio en configuracion antigua")
	}

	var cfg map[string]any
	if err := json.Unmarshal([]byte(nextRaw), &cfg); err != nil {
		t.Fatalf("json resultante invalido: %v", err)
	}
	global := cfg["carrito_ui_global"].(map[string]any)
	if got, _ := global["mostrar_tarjeta_cobro_estados"].(bool); got {
		t.Fatal("carrito_ui_global debe apagar cobro y estados")
	}
	if got, _ := global["mostrar_tarjeta_valores_pago"].(bool); !got {
		t.Fatal("carrito_ui_global debe conservar valores por medio de pago")
	}
	if cfg["cantidad"].(float64) != 2 {
		t.Fatal("no debe cambiar la cantidad de estaciones")
	}
	if _, ok := cfg["station_card_ui"].(map[string]any); !ok {
		t.Fatal("debe preservar station_card_ui")
	}
	stations := cfg["estaciones"].([]any)
	station := stations[0].(map[string]any)
	carrito := station["carrito"].(map[string]any)
	if got, _ := carrito["usar_configuracion_global"].(bool); !got {
		t.Fatal("debe activar usar_configuracion_global para unificar empresas viejas")
	}
	stationCfg := carrito["configuracion"].(map[string]any)
	if got, _ := stationCfg["mostrar_tarjeta_comision"].(bool); got {
		t.Fatal("configuracion por estacion debe apagar lavador/comision")
	}
}
