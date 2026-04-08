package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// ServerStartupRegistration define el contexto operativo del arranque actual.
type ServerStartupRegistration struct {
	BackendDir  string
	ListenAddr  string
	StartReason string
}

type serverRuntimeState struct {
	Status                string `json:"status"`
	Hostname              string `json:"hostname,omitempty"`
	ProcessID             int    `json:"process_id,omitempty"`
	ListenAddr            string `json:"listen_addr,omitempty"`
	LastStartAt           string `json:"last_start_at,omitempty"`
	LastStartReason       string `json:"last_start_reason,omitempty"`
	LastStartReasonDetail string `json:"last_start_reason_detail,omitempty"`
	LastStopAt            string `json:"last_stop_at,omitempty"`
	LastStopReason        string `json:"last_stop_reason,omitempty"`
	LastUnexpectedRestart bool   `json:"last_unexpected_restart"`
	LastKnownServerErr    string `json:"last_known_server_err,omitempty"`
	LastEventID           int64  `json:"last_event_id,omitempty"`
}

// RegisterServerStartupEvent registra el arranque en DB/log local y notifica por correo si hay destino configurado.
// Retorna un callback idempotente para marcar cierre controlado del servidor.
func RegisterServerStartupEvent(dbSuper *sql.DB, opts ServerStartupRegistration) (func(string), error) {
	backendDir := strings.TrimSpace(opts.BackendDir)
	if backendDir == "" {
		backendDir = "."
	}
	if absBackendDir, err := filepath.Abs(backendDir); err == nil {
		backendDir = absBackendDir
	}

	statePath := filepath.Join(backendDir, "logs", "server_runtime_state.json")
	logPath := filepath.Join(backendDir, "logs", "server_reinicio.log")

	prevState, err := loadServerRuntimeState(statePath)
	if err != nil {
		prevState = serverRuntimeState{}
	}

	hostname := "desconocido"
	if h, hostErr := os.Hostname(); hostErr == nil {
		hostname = strings.TrimSpace(h)
	}

	now := time.Now()
	nowText := now.Format("2006-01-02 15:04:05")
	listenAddr := strings.TrimSpace(opts.ListenAddr)
	pid := os.Getpid()

	motivo, motivoDetalle, reinicioInesperado := inferServerStartupReason(strings.TrimSpace(opts.StartReason), prevState)
	serverErrHint := readServerErrHint(backendDir)
	if serverErrHint != "" {
		if motivoDetalle != "" {
			motivoDetalle += " "
		}
		motivoDetalle += "Ultimo indicio en server.err: " + serverErrHint
	}

	correoDestino := ""
	correoDestinoReadErr := ""
	if dbSuper != nil {
		if configured, readErr := getDecryptedConfigValue(dbSuper, "gmail.restart_alert_to"); readErr != nil {
			correoDestinoReadErr = readErr.Error()
		} else {
			correoDestino = strings.TrimSpace(configured)
		}
	} else {
		correoDestinoReadErr = "db super no disponible"
	}

	asunto := fmt.Sprintf("[PCS] Inicio de servidor detectado (%s)", hostname)
	cuerpo := buildServerStartupEmailBody(nowText, hostname, listenAddr, motivo, motivoDetalle, reinicioInesperado, prevState)

	correoEnviado := false
	correoError := ""
	if correoDestino == "" {
		if correoDestinoReadErr != "" {
			correoError = "no se pudo leer gmail.restart_alert_to: " + correoDestinoReadErr
		} else {
			correoError = "gmail.restart_alert_to no configurado"
		}
	} else if _, addrErr := mail.ParseAddress(correoDestino); addrErr != nil {
		correoError = "gmail.restart_alert_to invalido: " + addrErr.Error()
	} else if dbSuper != nil && isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"motivo":%q,"reinicio_inesperado":%t,"listen_addr":%q}`, motivo, reinicioInesperado, listenAddr)
		if captureErr := captureEmpresaUsuarioMailNotification(
			dbSuper,
			dbpkg.SuperCorreoNotificacionTipoInicioServidor,
			0,
			correoDestino,
			asunto,
			cuerpo,
			"",
			metadataJSON,
			"sistema",
		); captureErr != nil {
			correoError = captureErr.Error()
		} else {
			correoEnviado = true
		}
	} else {
		if sendErr := sendServerStartupEmail(dbSuper, correoDestino, asunto, cuerpo); sendErr != nil {
			correoError = sendErr.Error()
		} else {
			correoEnviado = true
		}
	}

	metadataRaw, _ := json.Marshal(map[string]interface{}{
		"listen_addr":               listenAddr,
		"start_reason_input":        strings.TrimSpace(opts.StartReason),
		"previous_status":           strings.TrimSpace(prevState.Status),
		"previous_process_id":       prevState.ProcessID,
		"previous_start_at":         strings.TrimSpace(prevState.LastStartAt),
		"previous_stop_at":          strings.TrimSpace(prevState.LastStopAt),
		"previous_stop_reason":      strings.TrimSpace(prevState.LastStopReason),
		"previous_server_err_hint":  serverErrHint,
		"correo_destino_read_error": correoDestinoReadErr,
	})

	var eventID int64
	eventErr := error(nil)
	if dbSuper != nil {
		eventID, eventErr = dbpkg.CreateSuperServidorEvento(dbSuper, dbpkg.SuperServidorEvento{
			TipoEvento:         dbpkg.SuperServidorEventoTipoInicio,
			Motivo:             motivo,
			MotivoDetalle:      motivoDetalle,
			OrigenArranque:     strings.TrimSpace(opts.StartReason),
			Hostname:           hostname,
			ProcessID:          int64(pid),
			ListenAddr:         listenAddr,
			ReinicioInesperado: reinicioInesperado,
			PrevioEstado:       strings.TrimSpace(prevState.Status),
			PrevioProcessID:    int64(prevState.ProcessID),
			PrevioInicioEn:     strings.TrimSpace(prevState.LastStartAt),
			PrevioFinEn:        strings.TrimSpace(prevState.LastStopAt),
			CorreoDestino:      correoDestino,
			CorreoEnviado:      correoEnviado,
			CorreoError:        correoError,
			MetadataJSON:       string(metadataRaw),
			FechaEvento:        nowText,
			UsuarioCreador:     "sistema",
			Estado:             "activo",
			Observaciones:      "registro_automatico_arranque_servidor",
		})
	}

	appendPayload := map[string]interface{}{
		"timestamp":           nowText,
		"tipo_evento":         dbpkg.SuperServidorEventoTipoInicio,
		"motivo":              motivo,
		"motivo_detalle":      motivoDetalle,
		"hostname":            hostname,
		"process_id":          pid,
		"listen_addr":         listenAddr,
		"reinicio_inesperado": reinicioInesperado,
		"correo_destino":      correoDestino,
		"correo_enviado":      correoEnviado,
		"correo_error":        correoError,
		"event_id":            eventID,
	}
	_ = appendServerRuntimeLog(logPath, appendPayload)

	state := serverRuntimeState{
		Status:                "running",
		Hostname:              hostname,
		ProcessID:             pid,
		ListenAddr:            listenAddr,
		LastStartAt:           nowText,
		LastStartReason:       motivo,
		LastStartReasonDetail: motivoDetalle,
		LastUnexpectedRestart: reinicioInesperado,
		LastKnownServerErr:    serverErrHint,
		LastEventID:           eventID,
	}
	stateErr := saveServerRuntimeState(statePath, state)

	var stopOnce sync.Once
	markStopped := func(stopReason string) {
		stopOnce.Do(func() {
			state.Status = "stopped"
			state.LastStopAt = time.Now().Format("2006-01-02 15:04:05")
			state.LastStopReason = strings.TrimSpace(stopReason)
			if state.LastStopReason == "" {
				state.LastStopReason = "apagado_controlado"
			}
			_ = saveServerRuntimeState(statePath, state)
			_ = appendServerRuntimeLog(logPath, map[string]interface{}{
				"timestamp":   state.LastStopAt,
				"tipo_evento": "cierre_servidor",
				"hostname":    state.Hostname,
				"process_id":  state.ProcessID,
				"motivo":      state.LastStopReason,
			})
		})
	}

	if eventErr != nil {
		if stateErr != nil {
			return markStopped, fmt.Errorf("registro arranque db y estado fallidos: %v | %v", eventErr, stateErr)
		}
		return markStopped, eventErr
	}
	if stateErr != nil {
		return markStopped, stateErr
	}
	if err != nil {
		return markStopped, err
	}
	return markStopped, nil
}

func inferServerStartupReason(requested string, prevState serverRuntimeState) (string, string, bool) {
	requested = strings.TrimSpace(requested)
	prevStatus := strings.ToLower(strings.TrimSpace(prevState.Status))
	if prevStatus == "running" {
		detail := fmt.Sprintf(
			"Se detecto estado previo en ejecucion sin cierre limpio. previo_pid=%d previo_inicio=%s previo_host=%s previo_motivo_cierre=%s.",
			prevState.ProcessID,
			strings.TrimSpace(prevState.LastStartAt),
			strings.TrimSpace(prevState.Hostname),
			strings.TrimSpace(prevState.LastStopReason),
		)
		if requested != "" {
			detail += " Motivo reportado por entorno: " + requested + "."
		}
		return "reinicio_inesperado_detectado", detail, true
	}
	if requested != "" {
		return requested, "Motivo reportado por entorno/sistema de arranque.", false
	}
	return "inicio_normal", "Inicio sin evidencia de reinicio inesperado previo.", false
}

func buildServerStartupEmailBody(nowText, hostname, listenAddr, motivo, motivoDetalle string, reinicioInesperado bool, prevState serverRuntimeState) string {
	builder := strings.Builder{}
	builder.WriteString("Inicio de servidor detectado.\r\n\r\n")
	builder.WriteString("Fecha evento: " + nowText + "\r\n")
	builder.WriteString("Host: " + hostname + "\r\n")
	if listenAddr != "" {
		builder.WriteString("Direccion escucha: " + listenAddr + "\r\n")
	}
	builder.WriteString("Motivo: " + motivo + "\r\n")
	builder.WriteString(fmt.Sprintf("Reinicio inesperado: %t\r\n", reinicioInesperado))
	if strings.TrimSpace(motivoDetalle) != "" {
		builder.WriteString("Detalle: " + strings.TrimSpace(motivoDetalle) + "\r\n")
	}
	if strings.TrimSpace(prevState.Status) != "" {
		builder.WriteString("\r\nEstado previo: " + strings.TrimSpace(prevState.Status) + "\r\n")
	}
	if strings.TrimSpace(prevState.LastStartAt) != "" {
		builder.WriteString("Inicio previo: " + strings.TrimSpace(prevState.LastStartAt) + "\r\n")
	}
	if strings.TrimSpace(prevState.LastStopAt) != "" {
		builder.WriteString("Fin previo: " + strings.TrimSpace(prevState.LastStopAt) + "\r\n")
	}
	if strings.TrimSpace(prevState.LastStopReason) != "" {
		builder.WriteString("Motivo cierre previo: " + strings.TrimSpace(prevState.LastStopReason) + "\r\n")
	}
	builder.WriteString("\r\nMensaje generado automaticamente por el backend PCS.")
	return builder.String()
}

func loadServerRuntimeState(path string) (serverRuntimeState, error) {
	state := serverRuntimeState{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return state, err
	}
	if strings.TrimSpace(string(data)) == "" {
		return state, nil
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, err
	}
	return state, nil
}

func saveServerRuntimeState(path string, state serverRuntimeState) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	return os.WriteFile(path, raw, 0600)
}

func appendServerRuntimeLog(path string, payload map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(raw, '\n')); err != nil {
		return err
	}
	return nil
}

func readServerErrHint(backendDir string) string {
	path := filepath.Join(backendDir, "server.err")
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "panic") || strings.Contains(lower, "fatal") || strings.Contains(lower, "error") {
			return truncateServerErrLine(line)
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return truncateServerErrLine(line)
		}
	}
	return ""
}

func truncateServerErrLine(line string) string {
	line = strings.TrimSpace(line)
	if len(line) <= 360 {
		return line
	}
	return line[:360] + "..."
}

func sendServerStartupEmail(dbSuper *sql.DB, toEmail, subject, body string) error {
	if dbSuper == nil {
		return fmt.Errorf("db super no disponible")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(toEmail)); err != nil {
		return fmt.Errorf("correo destino invalido: %w", err)
	}

	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return fmt.Errorf("gmail.smtp_email no configurado")
	}

	smtpPass, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_app_password")
	if err != nil {
		return err
	}
	smtpPass = strings.TrimSpace(smtpPass)
	if smtpPass == "" {
		return fmt.Errorf("gmail.smtp_app_password no configurado")
	}

	smtpHost, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_host")
	smtpPort, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_port")
	fromName, _ := getDecryptedConfigValue(dbSuper, "gmail.smtp_from_name")

	smtpHost = strings.TrimSpace(smtpHost)
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}
	smtpPort = strings.TrimSpace(smtpPort)
	if smtpPort == "" {
		smtpPort = "587"
	}
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Powerful Control System"
	}

	mailHostForAuth := smtpHost
	if strings.Contains(smtpHost, ":") {
		if h, _, splitErr := net.SplitHostPort(smtpHost); splitErr == nil {
			mailHostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}

	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	msg := "From: " + fromName + " <" + smtpEmail + ">\r\n" +
		"To: " + strings.TrimSpace(toEmail) + "\r\n" +
		"Subject: " + strings.TrimSpace(subject) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
		body

	return smtp.SendMail(addr, auth, smtpEmail, []string{strings.TrimSpace(toEmail)}, []byte(msg))
}
