package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSuperQueueCapacitySurfaceAndMetricsAreRegistered(t *testing.T) {
	mainRaw, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	mainSource := string(mainRaw)
	for _, required := range []string{
		`/super/api/capacidad_colas`,
		`WithSuperAuditoria(dbSuper, "super_capacidad_colas"`,
		`pcs_queue_lane_saturation_percent`,
		`pcs_queue_lane_oldest_seconds`,
	} {
		if !strings.Contains(mainSource, required) {
			t.Fatalf("registro de capacidad ausente: %s", required)
		}
	}

	menuRaw, err := os.ReadFile(filepath.Join("..", "web", "super_administrador.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(menuRaw), `href="/super/capacidad_colas.html"`) {
		t.Fatal("el Super administrador no enlaza la capacidad de colas")
	}
	pageRaw, err := os.ReadFile(filepath.Join("..", "web", "super", "capacidad_colas.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(pageRaw)
	for _, lane := range []string{"printing", "product_add", "fiscal"} {
		if !strings.Contains(page, lane) {
			t.Fatalf("la UI no contempla el carril %s", lane)
		}
	}
}
