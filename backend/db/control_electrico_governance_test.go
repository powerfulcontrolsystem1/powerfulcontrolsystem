package db

import (
	"regexp"
	"testing"
	"time"
)

func TestControlElectricoTunnelDeviceUIDIsOpaqueAndUnique(t *testing.T) {
	pattern := regexp.MustCompile(`^RPI-[A-F0-9]{32}$`)
	seen := map[string]bool{}
	for i := 0; i < 512; i++ {
		uid, err := generateControlElectricoTunnelDeviceUID()
		if err != nil {
			t.Fatal(err)
		}
		if !pattern.MatchString(uid) {
			t.Fatalf("formato de ID no profesional: %q", uid)
		}
		if seen[uid] {
			t.Fatalf("ID repetido: %s", uid)
		}
		seen[uid] = true
	}
}

func TestControlElectricoSensorActionsPreserveTimerAndSchedule(t *testing.T) {
	for _, action := range []string{"encender_temporizado", "encender_programado"} {
		item := EmpresaControlElectricoRegla{Accion: action, TemporizadorSegundos: 900, EntradaGPIOPin: 23}
		normalizeEmpresaControlElectricoRegla(&item)
		if item.Accion != action {
			t.Fatalf("accion %q normalizada como %q", action, item.Accion)
		}
	}
}

func TestControlElectricoBandwidthLevelsAndBlocking(t *testing.T) {
	policy := EmpresaControlElectricoTunnelPolicy{EmpresaID: 7, LimiteMensualMB: 100, AlertaPorcentaje: 80, BloquearAlSuperar: true}
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	status := BuildEmpresaControlElectricoTunnelBandwidthStatus(policy, 79*1024*1024, now)
	if status.Nivel != "normal" || status.Bloqueado {
		t.Fatalf("79%% debe ser normal: %+v", status)
	}
	status = BuildEmpresaControlElectricoTunnelBandwidthStatus(policy, 80*1024*1024, now)
	if status.Nivel != "advertencia" || status.Bloqueado {
		t.Fatalf("80%% debe advertir sin bloquear: %+v", status)
	}
	status = BuildEmpresaControlElectricoTunnelBandwidthStatus(policy, 100*1024*1024, now)
	if status.Nivel != "excedido" || !status.Bloqueado || status.EmpresaID != 7 || status.Mes != "2026-08" {
		t.Fatalf("100%% debe exceder y bloquear: %+v", status)
	}
}
