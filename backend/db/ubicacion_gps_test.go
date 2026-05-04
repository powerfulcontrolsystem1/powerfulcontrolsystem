package db

import "testing"

func TestNormalizeEmpresaGPSDispositivoProfessionalFields(t *testing.T) {
	got := normalizeEmpresaGPSDispositivo(EmpresaGPSDispositivo{
		Nombre:                "  Tracker Norte  ",
		Marca:                 " Teltonika ",
		Modelo:                " FMB920 ",
		TipoDispositivo:       "GPS Tracker",
		Proveedor:             " Traccar ",
		IdentificadorHardware: " 123456789012345 ",
		TelefonoSIM:           " +573001112233 ",
		PlacaActivo:           " abc123 ",
		ActivoReferencia:      " Camion 1 ",
		IntervaloReporteSeg:   2,
		Protocolo:             "HTTP Push",
		Estado:                "otro",
	})

	if got.Nombre != "Tracker Norte" || got.Marca != "Teltonika" || got.Modelo != "FMB920" {
		t.Fatalf("campos base no normalizados: %#v", got)
	}
	if got.TipoDispositivo != "gps_tracker" || got.Protocolo != "http_push" {
		t.Fatalf("catalogos inesperados: tipo=%q protocolo=%q", got.TipoDispositivo, got.Protocolo)
	}
	if got.IntervaloReporteSeg != 5 {
		t.Fatalf("intervalo minimo debe ser 5 segundos, got %d", got.IntervaloReporteSeg)
	}
	if got.PlacaActivo != "ABC123" || got.Estado != "activo" {
		t.Fatalf("placa o estado inesperado: placa=%q estado=%q", got.PlacaActivo, got.Estado)
	}
}

func TestNormalizeEmpresaGPSRecorridoTelemetry(t *testing.T) {
	got := normalizeEmpresaGPSRecorrido(EmpresaGPSRecorrido{
		PrecisionMetros:   -1,
		VelocidadKMH:      -20,
		RumboGrados:       500,
		BateriaPorcentaje: 150,
		SenalPorcentaje:   -10,
		Evento:            " Motor Encendido ",
		Fuente:            "",
		Estado:            "raro",
		Observaciones:     "  ok  ",
	})

	if got.PrecisionMetros != 0 || got.VelocidadKMH != 0 || got.RumboGrados != 359.99 {
		t.Fatalf("telemetria numerica inesperada: %#v", got)
	}
	if got.BateriaPorcentaje != 100 || got.SenalPorcentaje != 0 {
		t.Fatalf("bateria/senal fuera de rango: bateria=%v senal=%v", got.BateriaPorcentaje, got.SenalPorcentaje)
	}
	if got.Evento != "motor_encendido" || got.Fuente != "manual" || got.Estado != "activo" || got.Observaciones != "ok" {
		t.Fatalf("catalogos no normalizados: %#v", got)
	}
}
