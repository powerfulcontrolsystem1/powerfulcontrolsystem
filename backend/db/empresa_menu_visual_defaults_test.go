package db

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEmpresaMenuVisualDefaultConfig(t *testing.T) {
	var config struct {
		Version     int      `json:"version"`
		Enabled     bool     `json:"enabled"`
		HiddenLinks []string `json:"hidden_links"`
	}
	if err := json.Unmarshal([]byte(EmpresaMenuVisualDefaultConfigJSON), &config); err != nil {
		t.Fatalf("default menu config must be valid JSON: %v", err)
	}
	if config.Version != 4 || !config.Enabled {
		t.Fatalf("unexpected default menu config: %+v", config)
	}
	wantHidden := []string{
		"linkImportacionesCosteo", "linkProduccionMRP", "linkLogisticaWMS",
		"linkCRMComercial",
		"linkMiHorario", "linkHorariosTrabajadores", "linkAsistenciaEmpleados", "linkCarnets",
		"linkVehiculosRegistro", "linkHojaVidaOperativa",
	}
	if len(config.HiddenLinks) != len(wantHidden) {
		t.Fatalf("hidden links=%d, want %d", len(config.HiddenLinks), len(wantHidden))
	}
	seen := make(map[string]bool, len(config.HiddenLinks))
	for _, id := range config.HiddenLinks {
		if seen[id] {
			t.Fatalf("duplicate hidden link %q", id)
		}
		seen[id] = true
	}
	for _, id := range wantHidden {
		if !seen[id] {
			t.Fatalf("default hidden link missing %q", id)
		}
	}
	for _, visible := range []string{
		"linkPanelEmpresa", "linkVida", "linkVentaDirecta", "linkProductos", "linkFinanzas",
		"linkUsuarios", "linkClientes", "linkPortalUsuarios",
		"linkVentaPublica", "linkControlElectrico", "linkDocumentosOnlyOffice",
		"linkUbicacionGPS", "linkConfiguracionGPS", "linkAuditoria", "linkCalidadProcesos",
		"linkCamaras", "linkGrafologia", "linkBolsa", "linkConfiguracion", "linkNoticias", "linkVolverEmpresas",
		"linkLicenciaSistema",
	} {
		if seen[visible] {
			t.Fatalf("required default visible link %q cannot be hidden", visible)
		}
	}
}

func TestEmpresaCatalogIncludesMenuVisualDefaultsMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260914-002-menu-visual-defaults-v3" {
			continue
		}
		if migration.Apply == nil || migration.Body != empresaMenuVisualDefaultsFingerprint {
			t.Fatal("menu visual defaults migration must be executable and immutable")
		}
		for _, marker := range []string{"FROM empresas e", "menu_visual_config", "ON CONFLICT", empresaMenuVisualDefaultConfigV3JSON} {
			if !strings.Contains(migration.Body, marker) {
				t.Fatalf("menu visual migration missing %q", marker)
			}
		}
		return
	}
	t.Fatal("menu visual defaults migration is missing from empresas catalog")
}

func TestEmpresaCatalogIncludesMenuVisualDefaultsV4Migration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260914-003-menu-visual-defaults-v4" {
			continue
		}
		if migration.Apply == nil || migration.Body != empresaMenuVisualDefaultsV4Fingerprint {
			t.Fatal("menu visual v4 migration must be executable and immutable")
		}
		if !strings.Contains(migration.Body, EmpresaMenuVisualDefaultConfigJSON) {
			t.Fatal("menu visual v4 migration must contain the current default")
		}
		return
	}
	t.Fatal("menu visual defaults v4 migration is missing from empresas catalog")
}
