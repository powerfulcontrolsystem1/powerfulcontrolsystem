package handlers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStationConfigurationSyncPreservesExistingSessionLifecycle(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_de_estaciones.html"))
	if err != nil {
		t.Fatal(err)
	}

	text := string(source)
	start := strings.Index(text, "async function syncCarritosFromStations")
	end := strings.Index(text[start:], "async function inactivarTodosLosCarritosEstacion")
	if start < 0 || end < 0 {
		t.Fatal("no se encontro el bloque de sincronizacion de carritos por estacion")
	}
	block := text[start : start+end]
	if strings.Contains(block, "ensureCarritoClosedAndInactive(currentID, current)") {
		t.Fatal("la sincronizacion de configuracion no debe cerrar una sesion existente")
	}
	if !strings.Contains(block, "ensureCarritoClosedAndInactive(createdID, createdItem)") {
		t.Fatal("un carrito nuevo debe conservar la inicializacion cerrada e inactiva")
	}
}
