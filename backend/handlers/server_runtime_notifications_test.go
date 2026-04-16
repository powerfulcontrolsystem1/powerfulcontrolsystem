package handlers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestRegisterServerStartupEventCapturesNotificationAndState(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_server_startup_event.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_test_mode", "1", false); err != nil {
		t.Fatalf("seed gmail.smtp_test_mode: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.restart_alert_to", "ops@empresa.com", false); err != nil {
		t.Fatalf("seed gmail.restart_alert_to: %v", err)
	}

	backendDir := t.TempDir()
	markStopped, err := RegisterServerStartupEvent(dbSuper, ServerStartupRegistration{
		BackendDir:  backendDir,
		ListenAddr:  ":8080",
		StartReason: "inicio_script_iniciar_servidor",
	})
	if err != nil {
		t.Fatalf("RegisterServerStartupEvent returned error: %v", err)
	}
	if markStopped == nil {
		t.Fatal("expected markStopped callback")
	}

	var motivo string
	var reinicioInesperado int
	var correoDestino string
	var correoEnviado int
	if err := dbSuper.QueryRow(`
		SELECT
			COALESCE(motivo, ''),
			COALESCE(reinicio_inesperado, 0),
			COALESCE(correo_destino, ''),
			COALESCE(correo_enviado, 0)
		FROM super_servidor_eventos
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&motivo, &reinicioInesperado, &correoDestino, &correoEnviado); err != nil {
		t.Fatalf("query super_servidor_eventos: %v", err)
	}
	if strings.TrimSpace(motivo) != "inicio_script_iniciar_servidor" {
		t.Fatalf("expected motivo %q, got %q", "inicio_script_iniciar_servidor", motivo)
	}
	if reinicioInesperado != 0 {
		t.Fatalf("expected reinicio_inesperado=0, got %d", reinicioInesperado)
	}
	if strings.TrimSpace(correoDestino) != "ops@empresa.com" {
		t.Fatalf("expected correo_destino %q, got %q", "ops@empresa.com", correoDestino)
	}
	if correoEnviado != 1 {
		t.Fatalf("expected correo_enviado=1 in test mode capture, got %d", correoEnviado)
	}

	notifications, err := dbpkg.ListSuperCorreoNotificacionesPrueba(dbSuper, dbpkg.SuperCorreoNotificacionPruebaFilter{
		Tipo:  dbpkg.SuperCorreoNotificacionTipoInicioServidor,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("list super correo notificaciones: %v", err)
	}
	if len(notifications) == 0 {
		t.Fatal("expected captured startup email notification in test mode")
	}
	if strings.TrimSpace(notifications[0].Destinatario) != "ops@empresa.com" {
		t.Fatalf("expected destinatario %q, got %q", "ops@empresa.com", notifications[0].Destinatario)
	}

	statePath := filepath.Join(backendDir, "logs", "server_runtime_state.json")
	rawState, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read runtime state before stop: %v", err)
	}
	var runningState serverRuntimeState
	if err := json.Unmarshal(rawState, &runningState); err != nil {
		t.Fatalf("decode running state: %v", err)
	}
	if strings.TrimSpace(runningState.Status) != "running" {
		t.Fatalf("expected state running, got %q", runningState.Status)
	}

	markStopped("signal_test")
	rawStoppedState, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read runtime state after stop: %v", err)
	}
	var stoppedState serverRuntimeState
	if err := json.Unmarshal(rawStoppedState, &stoppedState); err != nil {
		t.Fatalf("decode stopped state: %v", err)
	}
	if strings.TrimSpace(stoppedState.Status) != "stopped" {
		t.Fatalf("expected state stopped, got %q", stoppedState.Status)
	}
	if strings.TrimSpace(stoppedState.LastStopReason) != "signal_test" {
		t.Fatalf("expected last stop reason %q, got %q", "signal_test", stoppedState.LastStopReason)
	}
}

func TestRegisterServerStartupEventDetectsUnexpectedRestart(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_server_unexpected_restart.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_test_mode", "1", false); err != nil {
		t.Fatalf("seed gmail.smtp_test_mode: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.restart_alert_to", "ops@empresa.com", false); err != nil {
		t.Fatalf("seed gmail.restart_alert_to: %v", err)
	}

	backendDir := t.TempDir()
	statePath := filepath.Join(backendDir, "logs", "server_runtime_state.json")
	prevState := serverRuntimeState{
		Status:             "running",
		Hostname:           "host-previo",
		ProcessID:          4455,
		ListenAddr:         ":8080",
		LastStartAt:        "2026-04-08 07:00:00",
		LastStopAt:         "",
		LastStopReason:     "",
		LastEventID:        33,
		LastKnownServerErr: "panic: sqlite is locked",
	}
	if err := saveServerRuntimeState(statePath, prevState); err != nil {
		t.Fatalf("seed previous runtime state: %v", err)
	}

	markStopped, err := RegisterServerStartupEvent(dbSuper, ServerStartupRegistration{
		BackendDir:  backendDir,
		ListenAddr:  ":8080",
		StartReason: "inicio_script_iniciar_servidor",
	})
	if err != nil {
		t.Fatalf("RegisterServerStartupEvent returned error: %v", err)
	}
	if markStopped == nil {
		t.Fatal("expected markStopped callback")
	}

	var motivo string
	var reinicioInesperado int
	if err := dbSuper.QueryRow(`
		SELECT COALESCE(motivo, ''), COALESCE(reinicio_inesperado, 0)
		FROM super_servidor_eventos
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&motivo, &reinicioInesperado); err != nil {
		t.Fatalf("query latest startup event: %v", err)
	}
	if strings.TrimSpace(motivo) != "reinicio_inesperado_detectado" {
		t.Fatalf("expected motivo %q, got %q", "reinicio_inesperado_detectado", motivo)
	}
	if reinicioInesperado != 1 {
		t.Fatalf("expected reinicio_inesperado=1, got %d", reinicioInesperado)
	}
}

func TestRegisterServerStartupEventSkipsEmailWhenAlertsDisabled(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_server_alert_disabled.db")
	ensureSuperConfigSchemaForSuper(t, dbSuper)

	if err := dbpkg.SetConfigValue(dbSuper, "gmail.smtp_test_mode", "1", false); err != nil {
		t.Fatalf("seed gmail.smtp_test_mode: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.restart_alert_to", "ops@empresa.com", false); err != nil {
		t.Fatalf("seed gmail.restart_alert_to: %v", err)
	}
	if err := dbpkg.SetConfigValue(dbSuper, "gmail.restart_alert_enabled", "0", false); err != nil {
		t.Fatalf("seed gmail.restart_alert_enabled: %v", err)
	}

	backendDir := t.TempDir()
	markStopped, err := RegisterServerStartupEvent(dbSuper, ServerStartupRegistration{
		BackendDir:  backendDir,
		ListenAddr:  ":8080",
		StartReason: "inicio_script_iniciar_servidor",
	})
	if err != nil {
		t.Fatalf("RegisterServerStartupEvent returned error: %v", err)
	}
	if markStopped == nil {
		t.Fatal("expected markStopped callback")
	}

	var correoDestino string
	var correoEnviado int
	var correoError string
	if err := dbSuper.QueryRow(`
		SELECT
			COALESCE(correo_destino, ''),
			COALESCE(correo_enviado, 0),
			COALESCE(correo_error, '')
		FROM super_servidor_eventos
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&correoDestino, &correoEnviado, &correoError); err != nil {
		t.Fatalf("query super_servidor_eventos: %v", err)
	}
	if strings.TrimSpace(correoDestino) != "ops@empresa.com" {
		t.Fatalf("expected correo_destino %q, got %q", "ops@empresa.com", correoDestino)
	}
	if correoEnviado != 0 {
		t.Fatalf("expected correo_enviado=0 when alerts disabled, got %d", correoEnviado)
	}
	if !strings.Contains(strings.ToLower(correoError), "desactivada") {
		t.Fatalf("expected correo_error to mention disabled alert, got %q", correoError)
	}

	notifications, err := dbpkg.ListSuperCorreoNotificacionesPrueba(dbSuper, dbpkg.SuperCorreoNotificacionPruebaFilter{
		Tipo:  dbpkg.SuperCorreoNotificacionTipoInicioServidor,
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("list super correo notificaciones: %v", err)
	}
	if len(notifications) != 0 {
		t.Fatalf("expected no captured startup email notification when alerts are disabled, got %d", len(notifications))
	}
}
