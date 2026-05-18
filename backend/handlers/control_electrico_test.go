package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestBuildControlElectricoEndpoint(t *testing.T) {
	cfg := &dbpkg.EmpresaControlElectricoConfig{
		RaspberryIP:   "192.168.10.20",
		RaspberryPort: 8090,
		APIPath:       "/api/gpio/relay",
	}
	got, err := buildControlElectricoEndpoint(cfg)
	if err != nil {
		t.Fatalf("build endpoint: %v", err)
	}
	want := "http://192.168.10.20:8090/api/gpio/relay"
	if got != want {
		t.Fatalf("endpoint = %q, want %q", got, want)
	}
}

func TestSendControlElectricoRelayCommand(t *testing.T) {
	var received controlElectricoCommandPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Fatalf("Authorization = %q", got)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cfg := &dbpkg.EmpresaControlElectricoConfig{
		EmpresaID:     77,
		RaspberryIP:   server.URL,
		APIPath:       "/api/gpio/relay",
		APIToken:      "secret-token",
		TimeoutMS:     1000,
		RaspberryPort: dbpkg.DefaultControlElectricoPort,
	}
	rele := &dbpkg.EmpresaControlElectricoRele{
		ID:             9,
		RaspberryID:    3,
		EstacionID:     4,
		EstacionCodigo: "EST-77-4",
		EstacionNombre: "Estacion 4",
		GPIOPin:        17,
		RelayName:      "Luz estacion 4",
		ActiveHigh:     true,
		PulsoMS:        0,
	}

	result := sendControlElectricoRelayCommand(cfg, rele, "on", "tester", "prueba_manual")
	if !result.OK {
		t.Fatalf("result not ok: %+v", result)
	}
	if received.EmpresaID != 77 || received.EstacionID != 4 || received.GPIOPin != 17 || received.Estado != "on" {
		t.Fatalf("payload inesperado: %+v", received)
	}
	if received.RaspberryID != 3 {
		t.Fatalf("raspberry_id = %d, want 3", received.RaspberryID)
	}
	if !received.ActiveHigh || received.Actor != "tester" || received.Origen != "prueba_manual" {
		t.Fatalf("payload runtime inesperado: %+v", received)
	}
}

func TestControlElectricoConfigFromRaspberry(t *testing.T) {
	base := &dbpkg.EmpresaControlElectricoConfig{
		EmpresaID:     12,
		Habilitado:    true,
		RaspberryIP:   "192.168.1.10",
		RaspberryPort: 8081,
		APIPath:       "/api/default",
		APIToken:      "base-token",
		TimeoutMS:     1000,
	}
	pi := &dbpkg.EmpresaControlElectricoRaspberry{
		ID:            5,
		RaspberryIP:   "192.168.1.55",
		RaspberryPort: 8099,
		APIPath:       "/api/custom",
		APIToken:      "pi-token",
		TimeoutMS:     3400,
	}

	got := controlElectricoConfigFromRaspberry(base, pi)
	if got == base {
		t.Fatalf("expected copied config, got original pointer")
	}
	if got.EmpresaID != 12 || got.RaspberryIP != "192.168.1.55" || got.RaspberryPort != 8099 || got.APIPath != "/api/custom" {
		t.Fatalf("config raspberry inesperada: %+v", got)
	}
	if got.APIToken != "pi-token" || got.TimeoutMS != 3400 {
		t.Fatalf("credenciales raspberry inesperadas: %+v", got)
	}
	if base.RaspberryIP != "192.168.1.10" || base.APIToken != "base-token" {
		t.Fatalf("base config fue mutada: %+v", base)
	}
}

func TestControlElectricoProgramacionDue(t *testing.T) {
	now := time.Date(2026, 5, 4, 8, 30, 0, 0, time.UTC) // lunes 03:30 Bogota.
	rele := dbpkg.EmpresaControlElectricoRele{
		ID:                     10,
		EmpresaID:              7,
		ProgramacionHabilitada: true,
		HoraEncendido:          "03:30",
		HoraApagado:            "04:00",
		ProgramacionDias:       "lunes_viernes",
		ProgramacionTimezone:   "America/Bogota",
	}
	due := controlElectricoProgramacionDue(rele, now)
	if len(due) != 1 || due[0].EstadoObjetivo != "on" {
		t.Fatalf("due = %+v, want one on", due)
	}
	rele.UltimaProgramacionOn = "2026-05-04 03:30:00"
	if got := controlElectricoProgramacionDue(rele, now); len(got) != 0 {
		t.Fatalf("due repetido = %+v, want none", got)
	}
}

func TestControlElectricoProgramacionDiaActivo(t *testing.T) {
	if !controlElectricoProgramacionDiaActivo("sabado_domingo", time.Sunday) {
		t.Fatalf("domingo debe estar activo para fines de semana")
	}
	if controlElectricoProgramacionDiaActivo("lunes_viernes", time.Sunday) {
		t.Fatalf("domingo no debe estar activo para lunes_viernes")
	}
	if !controlElectricoProgramacionDiaActivo("1,3,5", time.Friday) {
		t.Fatalf("viernes debe estar activo en lista 1,3,5")
	}
}
