package db

import "testing"

func TestNormalizeEmpresaConfiguracionGeneralClima(t *testing.T) {
	cfg := normalizeEmpresaConfiguracionGeneral(EmpresaConfiguracionGeneral{
		EmpresaID:           29,
		ClimaCiudad:         "  Medellin  ",
		ClimaRegion:         " Antioquia ",
		ClimaPais:           " Colombia ",
		ClimaPaisCodigo:     "colombia",
		ClimaMoneda:         "cop-extra",
		ClimaLatitud:        6.2442,
		ClimaLongitud:       -75.5812,
		ClimaNombre:         " Medellin, Antioquia, Colombia ",
		ClimaFuente:         "GPS",
		CopiasOrdenServicio: 1,
	})

	if cfg.ClimaCiudad != "Medellin" {
		t.Fatalf("ciudad normalizada incorrecta: %q", cfg.ClimaCiudad)
	}
	if cfg.ClimaRegion != "Antioquia" {
		t.Fatalf("region normalizada incorrecta: %q", cfg.ClimaRegion)
	}
	if cfg.ClimaPais != "Colombia" {
		t.Fatalf("pais normalizado incorrecto: %q", cfg.ClimaPais)
	}
	if cfg.ClimaPaisCodigo != "CO" {
		t.Fatalf("codigo pais debe quedar de dos letras: %q", cfg.ClimaPaisCodigo)
	}
	if cfg.ClimaMoneda != "COP" {
		t.Fatalf("moneda debe quedar de tres letras: %q", cfg.ClimaMoneda)
	}
	if cfg.ClimaLatitud != 6.2442 || cfg.ClimaLongitud != -75.5812 {
		t.Fatalf("coordenadas no deben cambiar: %f,%f", cfg.ClimaLatitud, cfg.ClimaLongitud)
	}
	if cfg.ClimaNombre != "Medellin, Antioquia, Colombia" {
		t.Fatalf("nombre normalizado incorrecto: %q", cfg.ClimaNombre)
	}
	if cfg.ClimaFuente != "gps" {
		t.Fatalf("fuente normalizada incorrecta: %q", cfg.ClimaFuente)
	}
}
