package db

import "testing"

func TestControlElectricoDoorSensorNormalizationAndGPIOLayout(t *testing.T) {
	if got := NormalizeControlElectricoUsoTipo("Puertas"); got != ControlElectricoUsoSensorPuertas {
		t.Fatalf("uso normalizado = %q", got)
	}
	if got := NormalizeControlElectricoUsoTipo("otro"); got != ControlElectricoUsoDomotica {
		t.Fatalf("uso por defecto = %q", got)
	}
	cfg := BuildEmpresaControlElectricoDoorScanConfig(16, 250)
	if len(cfg.InputPins) != 4 || cfg.InputPins[0] != 0 || cfg.InputPins[3] != 3 {
		t.Fatalf("entradas inesperadas: %#v", cfg.InputPins)
	}
	if len(cfg.OutputPins) != 16 || cfg.OutputPins[0] != 4 || cfg.OutputPins[15] != 19 {
		t.Fatalf("salidas inesperadas: %#v", cfg.OutputPins)
	}
	if cfg.DelayMS != 250 || cfg.InputPull != "up" {
		t.Fatalf("configuracion inesperada: %#v", cfg)
	}
}

func TestControlElectricoDoorSensorLimitsAndDeviceIDs(t *testing.T) {
	if NormalizeControlElectricoDoorOutputCount(99) != 16 || NormalizeControlElectricoDoorOutputCount(0) != 16 {
		t.Fatal("la cantidad de salidas debe limitarse a 1..16 con default 16")
	}
	if NormalizeControlElectricoDoorDelayMS(1) != 100 || NormalizeControlElectricoDoorDelayMS(9000) != 5000 {
		t.Fatal("el delay debe usar default seguro y maximo 5000 ms")
	}
	first := empresaControlElectricoDoorDeviceID(7, 12, 1, 1)
	second := empresaControlElectricoDoorDeviceID(7, 12, 1, 2)
	otherTenant := empresaControlElectricoDoorDeviceID(8, 12, 1, 1)
	if first == second || first == otherTenant || first == "" {
		t.Fatalf("IDs de canal no aislados: %q %q %q", first, second, otherTenant)
	}
}

func TestEmpresaCatalogIncludesDoorSensorMigration(t *testing.T) {
	migrations, err := PlatformMigrations(MigrationTargetEmpresas)
	if err != nil {
		t.Fatal(err)
	}
	for _, migration := range migrations {
		if migration.Version != "20260831-002-raspberry-door-sensor-v1" {
			continue
		}
		if migration.Body != empresaControlElectricoDoorSensorSchemaFingerprint || migration.Apply == nil {
			t.Fatal("la migracion de sensores de puertas debe ser inmutable y ejecutable")
		}
		return
	}
	t.Fatal("falta la migracion versionada de sensores de puertas Raspberry")
}
