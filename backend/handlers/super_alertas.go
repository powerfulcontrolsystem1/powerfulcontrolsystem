package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

type superAlertCandidate struct {
	Tipo            string  `json:"tipo"`
	Severidad       string  `json:"severidad"`
	Titulo          string  `json:"titulo"`
	Detalle         string  `json:"detalle"`
	Valor           float64 `json:"valor"`
	Umbral          float64 `json:"umbral"`
	Unidad          string  `json:"unidad"`
	Triggered       bool    `json:"triggered"`
	CorreoEnviado   bool    `json:"correo_enviado"`
	CorreoError     string  `json:"correo_error,omitempty"`
	SkippedCooldown bool    `json:"skipped_cooldown,omitempty"`
}

type superAlertEvaluation struct {
	OK             bool                    `json:"ok"`
	Config         dbpkg.SuperAlertaConfig `json:"config"`
	Metric         *dbpkg.Metric           `json:"metric,omitempty"`
	DiskPercent    float64                 `json:"disk_percent"`
	TrafficTotalGB float64                 `json:"traffic_total_gb"`
	TrafficLimitGB float64                 `json:"traffic_limit_gb"`
	TrafficPercent float64                 `json:"traffic_percent"`
	ActiveSessions int64                   `json:"active_sessions"`
	DBConnections  int64                   `json:"db_connections"`
	Candidates     []superAlertCandidate   `json:"candidates"`
	EvaluatedAt    string                  `json:"evaluated_at"`
	Error          string                  `json:"error,omitempty"`
}

func SuperAlertasSistemaHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if dbSuper == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "db super no disponible"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			cfg, err := dbpkg.GetSuperAlertasConfig(dbSuper)
			if err != nil {
				writeSuperAlertasPublicError(w, r, http.StatusInternalServerError, "cargar configuracion", err, nil)
				return
			}
			eval := EvaluateSuperAlertasSistema(dbSuper, false)
			redactSuperAlertasEvaluationError(&eval, r)
			events, err := dbpkg.ListSuperAlertaEventos(dbSuper, 100)
			if err != nil {
				writeSuperAlertasPublicError(w, r, http.StatusInternalServerError, "listar eventos", err, nil)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":         true,
				"config":     cfg,
				"evaluation": eval,
				"events":     events,
			})
			return

		case http.MethodPut:
			var payload dbpkg.SuperAlertaConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}
			payload.UsuarioCreador = adminEmail
			if _, err := mail.ParseAddress(strings.TrimSpace(payload.RecipientEmail)); err != nil {
				http.Error(w, "recipient_email invalido", http.StatusBadRequest)
				return
			}
			if payload.DiskThresholdPct <= 0 || payload.DiskThresholdPct > 100 {
				http.Error(w, "disk_threshold_pct debe estar entre 1 y 100", http.StatusBadRequest)
				return
			}
			if payload.TrafficThresholdPct < 0 || payload.TrafficThresholdPct > 100 {
				http.Error(w, "traffic_threshold_pct debe estar entre 0 y 100", http.StatusBadRequest)
				return
			}
			if payload.TrafficThresholdGB < 0 {
				http.Error(w, "traffic_threshold_gb no puede ser negativo", http.StatusBadRequest)
				return
			}
			if payload.SessionsThreshold <= 0 || payload.DBConnectionsThreshold <= 0 || payload.CooldownMinutes <= 0 {
				http.Error(w, "umbrales numericos deben ser mayores que cero", http.StatusBadRequest)
				return
			}
			if err := dbpkg.SaveSuperAlertasConfig(dbSuper, payload); err != nil {
				writeSuperAlertasPublicError(w, r, http.StatusInternalServerError, "guardar configuracion", err, nil)
				return
			}
			cfg, _ := dbpkg.GetSuperAlertasConfig(dbSuper)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": cfg})
			return

		case http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			switch action {
			case "test", "probar":
				cfg, err := dbpkg.GetSuperAlertasConfig(dbSuper)
				if err != nil {
					writeSuperAlertasPublicError(w, r, http.StatusInternalServerError, "cargar configuracion de prueba", err, nil)
					return
				}
				subject := "[PCS] Prueba de alertas del sistema"
				body := "Prueba de alertas del sistema Powerful Control System.\r\n\r\n" +
					"Destino configurado: " + cfg.RecipientEmail + "\r\n" +
					"Usuario: " + adminEmail + "\r\n" +
					"Fecha: " + time.Now().Format("2006-01-02 15:04:05") + "\r\n\r\n" +
					"Si recibes este correo, el canal SMTP para alertas esta operativo."
				sent, sendErr := sendSuperSystemAlertEmail(dbSuper, cfg.RecipientEmail, subject, body, "prueba_alerta_sistema", adminEmail)
				event := dbpkg.SuperAlertaEvento{
					Tipo:           "prueba_alerta_sistema",
					Severidad:      "info",
					Titulo:         "Prueba manual de alertas",
					Detalle:        "Correo de prueba solicitado desde super administrador.",
					Destinatario:   cfg.RecipientEmail,
					Asunto:         subject,
					Cuerpo:         body,
					CorreoEnviado:  sent,
					UsuarioCreador: adminEmail,
					Observaciones:  "prueba_manual",
				}
				if sendErr != nil {
					event.CorreoError = sendErr.Error()
				}
				_, _ = dbpkg.CreateSuperAlertaEvento(dbSuper, event)
				if sendErr != nil {
					writeSuperAlertasPublicError(w, r, http.StatusInternalServerError, "enviar alerta de prueba", sendErr, map[string]interface{}{"sent": false})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "sent": true, "recipient": cfg.RecipientEmail})
				return
			case "evaluate", "evaluar":
				eval := EvaluateSuperAlertasSistema(dbSuper, true)
				status := http.StatusOK
				if eval.Error != "" {
					status = http.StatusInternalServerError
				}
				redactSuperAlertasEvaluationError(&eval, r)
				writeJSON(w, status, eval)
				return
			default:
				http.Error(w, "action no soportada", http.StatusBadRequest)
				return
			}

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func writeSuperAlertasPublicError(w http.ResponseWriter, r *http.Request, status int, operation string, err error, extra map[string]interface{}) {
	requestID := resolveAuditoriaRequestID(r)
	log.Printf("[super_alertas] operation=%s request_id=%s error_type=%T", operation, requestID, err)
	payload := map[string]interface{}{
		"ok":    false,
		"code":  "super_alertas_error",
		"error": "No se pudo completar la operacion de alertas.",
	}
	for key, value := range extra {
		payload[key] = value
	}
	if requestID != "" {
		payload["request_id"] = requestID
	}
	writeJSON(w, status, payload)
}

func redactSuperAlertasEvaluationError(eval *superAlertEvaluation, r *http.Request) {
	if eval == nil || strings.TrimSpace(eval.Error) == "" {
		return
	}
	log.Printf("[super_alertas] operation=evaluar request_id=%s error_present=true", resolveAuditoriaRequestID(r))
	eval.Error = "No se pudo evaluar el sistema de alertas."
}

func EvaluateSuperAlertasSistema(dbSuper *sql.DB, forceSend bool) superAlertEvaluation {
	eval := superAlertEvaluation{
		OK:          true,
		EvaluatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Candidates:  []superAlertCandidate{},
	}
	if dbSuper == nil {
		eval.OK = false
		eval.Error = "db super no disponible"
		return eval
	}
	cfg, err := dbpkg.GetSuperAlertasConfig(dbSuper)
	if err != nil {
		eval.OK = false
		eval.Error = err.Error()
		return eval
	}
	eval.Config = cfg
	if !cfg.Enabled {
		return eval
	}

	if metric, err := dbpkg.GetLatestMetric(dbSuper); err == nil && metric != nil {
		eval.Metric = metric
		eval.DiskPercent = safeFloat(metric.DiskPercent)
		eval.TrafficTotalGB = safeFloat(float64(metric.NetRecv+metric.NetSent) / 1024 / 1024 / 1024)
	} else if err != nil && err != sql.ErrNoRows {
		eval.Candidates = append(eval.Candidates, superAlertCandidate{
			Tipo:      "metricas_no_disponibles",
			Severidad: "warning",
			Titulo:    "Metricas no disponibles",
			Detalle:   err.Error(),
		})
	}
	eval.TrafficLimitGB = dbpkg.GetFloatConfig(dbSuper, "hostinger.bandwidth.limit_gb")
	if eval.TrafficLimitGB > 0 {
		eval.TrafficPercent = safeFloat((eval.TrafficTotalGB / eval.TrafficLimitGB) * 100)
	}
	if total, err := dbpkg.CountActiveAdminSessions(dbSuper); err == nil {
		eval.ActiveSessions = total
	}
	if total, err := dbpkg.CountDatabaseConnections(dbSuper); err == nil {
		eval.DBConnections = total
	}

	if cfg.DiskEnabled && eval.Metric != nil {
		c := superAlertCandidate{
			Tipo:      "disco_vps",
			Severidad: severityForPercent(eval.DiskPercent, cfg.DiskThresholdPct),
			Titulo:    "Capacidad del disco VPS alta",
			Detalle:   fmt.Sprintf("El disco del VPS esta en %.2f%%. Umbral configurado: %.2f%%.", eval.DiskPercent, cfg.DiskThresholdPct),
			Valor:     superAlertRound2(eval.DiskPercent),
			Umbral:    superAlertRound2(cfg.DiskThresholdPct),
			Unidad:    "%",
			Triggered: eval.DiskPercent >= cfg.DiskThresholdPct,
		}
		eval.dispatchCandidate(dbSuper, cfg, c, forceSend)
	}

	if cfg.TrafficEnabled {
		if cfg.TrafficThresholdGB > 0 {
			c := superAlertCandidate{
				Tipo:      "trafico_vps_gb",
				Severidad: "warning",
				Titulo:    "Trafico acumulado alto",
				Detalle:   fmt.Sprintf("El trafico acumulado del VPS esta en %.2f GB. Umbral configurado: %.2f GB.", eval.TrafficTotalGB, cfg.TrafficThresholdGB),
				Valor:     superAlertRound2(eval.TrafficTotalGB),
				Umbral:    superAlertRound2(cfg.TrafficThresholdGB),
				Unidad:    "GB",
				Triggered: eval.TrafficTotalGB >= cfg.TrafficThresholdGB,
			}
			eval.dispatchCandidate(dbSuper, cfg, c, forceSend)
		} else if eval.TrafficLimitGB > 0 && cfg.TrafficThresholdPct > 0 {
			c := superAlertCandidate{
				Tipo:      "trafico_vps_pct",
				Severidad: severityForPercent(eval.TrafficPercent, cfg.TrafficThresholdPct),
				Titulo:    "Trafico del VPS cerca del limite",
				Detalle:   fmt.Sprintf("El trafico del VPS esta en %.2f%% del limite Hostinger configurado (%.2f de %.2f GB). Umbral: %.2f%%.", eval.TrafficPercent, eval.TrafficTotalGB, eval.TrafficLimitGB, cfg.TrafficThresholdPct),
				Valor:     superAlertRound2(eval.TrafficPercent),
				Umbral:    superAlertRound2(cfg.TrafficThresholdPct),
				Unidad:    "%",
				Triggered: eval.TrafficPercent >= cfg.TrafficThresholdPct,
			}
			eval.dispatchCandidate(dbSuper, cfg, c, forceSend)
		} else {
			eval.Candidates = append(eval.Candidates, superAlertCandidate{
				Tipo:      "trafico_vps_sin_limite",
				Severidad: "info",
				Titulo:    "Trafico sin limite configurado",
				Detalle:   "Para alertas de trafico por porcentaje configura Hostinger: ancho de banda limite (GB), o define un umbral en GB en este modulo.",
			})
		}
	}

	if cfg.SessionsEnabled {
		c := superAlertCandidate{
			Tipo:      "sesiones_admin",
			Severidad: "warning",
			Titulo:    "Maximo de sesiones administrativas cerca del limite",
			Detalle:   fmt.Sprintf("Hay %d sesiones administrativas activas. Umbral configurado: %d.", eval.ActiveSessions, cfg.SessionsThreshold),
			Valor:     float64(eval.ActiveSessions),
			Umbral:    float64(cfg.SessionsThreshold),
			Unidad:    "sesiones",
			Triggered: eval.ActiveSessions >= cfg.SessionsThreshold,
		}
		eval.dispatchCandidate(dbSuper, cfg, c, forceSend)
	}

	if cfg.DBConnectionsEnabled {
		c := superAlertCandidate{
			Tipo:      "conexiones_postgres",
			Severidad: "warning",
			Titulo:    "Conexiones PostgreSQL altas",
			Detalle:   fmt.Sprintf("PostgreSQL reporta %d conexiones. Umbral configurado: %d.", eval.DBConnections, cfg.DBConnectionsThreshold),
			Valor:     float64(eval.DBConnections),
			Umbral:    float64(cfg.DBConnectionsThreshold),
			Unidad:    "conexiones",
			Triggered: eval.DBConnections >= cfg.DBConnectionsThreshold,
		}
		eval.dispatchCandidate(dbSuper, cfg, c, forceSend)
	}

	return eval
}

func (eval *superAlertEvaluation) dispatchCandidate(dbSuper *sql.DB, cfg dbpkg.SuperAlertaConfig, c superAlertCandidate, forceSend bool) {
	if !c.Triggered {
		eval.Candidates = append(eval.Candidates, c)
		return
	}
	if !forceSend {
		if recent, err := dbpkg.SuperAlertaRecentlySent(dbSuper, c.Tipo, cfg.CooldownMinutes); err == nil && recent {
			c.SkippedCooldown = true
			eval.Candidates = append(eval.Candidates, c)
			return
		}
	}

	subject := "[PCS] " + c.Titulo
	body := buildSuperSystemAlertBody(c, eval)
	sent, sendErr := sendSuperSystemAlertEmail(dbSuper, cfg.RecipientEmail, subject, body, c.Tipo, "sistema")
	c.CorreoEnviado = sent
	if sendErr != nil {
		c.CorreoError = sendErr.Error()
	}
	metadata, _ := json.Marshal(map[string]interface{}{
		"evaluated_at":     eval.EvaluatedAt,
		"traffic_total_gb": eval.TrafficTotalGB,
		"traffic_limit_gb": eval.TrafficLimitGB,
		"traffic_percent":  eval.TrafficPercent,
		"active_sessions":  eval.ActiveSessions,
		"db_connections":   eval.DBConnections,
		"cooldown_minutes": cfg.CooldownMinutes,
	})
	event := dbpkg.SuperAlertaEvento{
		Tipo:           c.Tipo,
		Severidad:      c.Severidad,
		Titulo:         c.Titulo,
		Detalle:        c.Detalle,
		Valor:          c.Valor,
		Umbral:         c.Umbral,
		Unidad:         c.Unidad,
		Destinatario:   cfg.RecipientEmail,
		Asunto:         subject,
		Cuerpo:         body,
		CorreoEnviado:  sent,
		MetadataJSON:   string(metadata),
		UsuarioCreador: "sistema",
		Observaciones:  "evaluacion_automatica_alertas",
	}
	if sendErr != nil {
		event.CorreoError = sendErr.Error()
	}
	if _, err := dbpkg.CreateSuperAlertaEvento(dbSuper, event); err != nil && sendErr == nil {
		c.CorreoError = err.Error()
	}
	eval.Candidates = append(eval.Candidates, c)
}

func buildSuperSystemAlertBody(c superAlertCandidate, eval *superAlertEvaluation) string {
	var b strings.Builder
	b.WriteString("Alerta del sistema Powerful Control System.\r\n\r\n")
	b.WriteString("Alerta: " + c.Titulo + "\r\n")
	b.WriteString("Severidad: " + c.Severidad + "\r\n")
	b.WriteString("Detalle: " + c.Detalle + "\r\n")
	b.WriteString(fmt.Sprintf("Valor: %.2f %s\r\n", c.Valor, c.Unidad))
	b.WriteString(fmt.Sprintf("Umbral: %.2f %s\r\n", c.Umbral, c.Unidad))
	b.WriteString("Fecha evaluacion: " + eval.EvaluatedAt + "\r\n\r\n")
	b.WriteString("Estado operativo:\r\n")
	b.WriteString(fmt.Sprintf("- Disco VPS: %.2f%%\r\n", eval.DiskPercent))
	if eval.TrafficLimitGB > 0 {
		b.WriteString(fmt.Sprintf("- Trafico: %.2f GB / %.2f GB (%.2f%%)\r\n", eval.TrafficTotalGB, eval.TrafficLimitGB, eval.TrafficPercent))
	} else {
		b.WriteString(fmt.Sprintf("- Trafico acumulado: %.2f GB\r\n", eval.TrafficTotalGB))
	}
	b.WriteString(fmt.Sprintf("- Sesiones administrativas activas: %d\r\n", eval.ActiveSessions))
	b.WriteString(fmt.Sprintf("- Conexiones PostgreSQL: %d\r\n", eval.DBConnections))
	b.WriteString("\r\nRevisa el modulo Super administrador > Alertas del sistema.")
	return b.String()
}

func NotifySuperAdminAdminRegistered(dbSuper *sql.DB, adminID int64, email, name, telefono, pais, ciudad string) {
	notifySuperAdminBusinessEvent(dbSuper, "admin_registrado_login", func(cfg dbpkg.SuperAlertaConfig) bool {
		return cfg.AdminRegisterEnabled
	}, "Nuevo administrador registrado", map[string]string{
		"ID administrador": strconv.FormatInt(adminID, 10),
		"Nombre":           strings.TrimSpace(name),
		"Correo":           strings.TrimSpace(email),
		"Telefono":         strings.TrimSpace(telefono),
		"Pais":             strings.TrimSpace(pais),
		"Ciudad":           strings.TrimSpace(ciudad),
		"Origen":           "login.html / registro administrador",
	})
}

func NotifySuperAdminEmpresaNueva(dbSuper *sql.DB, empresaID, tipoID int64, nombre, nit, tipoNombre, usuarioCreador string, preconfigAplicada bool, preconfigError string) {
	notifySuperAdminBusinessEvent(dbSuper, "empresa_nueva_admin", func(cfg dbpkg.SuperAlertaConfig) bool {
		return cfg.EmpresaNuevaEnabled
	}, "Nueva empresa creada", map[string]string{
		"ID empresa":             strconv.FormatInt(empresaID, 10),
		"Nombre":                 strings.TrimSpace(nombre),
		"NIT":                    strings.TrimSpace(nit),
		"Tipo ID":                strconv.FormatInt(tipoID, 10),
		"Tipo":                   strings.TrimSpace(tipoNombre),
		"Administrador":          strings.TrimSpace(usuarioCreador),
		"Preconfiguracion":       boolLabel(preconfigAplicada),
		"Preconfiguracion aviso": strings.TrimSpace(preconfigError),
		"Origen":                 "seleccionar_empresa.html / agregar empresa",
	})
}

func notifySuperAdminBusinessEvent(dbSuper *sql.DB, tipo string, enabled func(dbpkg.SuperAlertaConfig) bool, titulo string, fields map[string]string) {
	if dbSuper == nil {
		return
	}
	go func() {
		cfg, err := dbpkg.GetSuperAlertasConfig(dbSuper)
		if err != nil {
			log.Printf("super_alertas: no se pudo leer configuracion para %s: %v", tipo, err)
			return
		}
		if !cfg.Enabled || !enabled(cfg) {
			return
		}
		subject := "[PCS] " + titulo
		body := buildSuperBusinessAlertBody(titulo, fields)
		sent, sendErr := sendSuperSystemAlertEmail(dbSuper, cfg.RecipientEmail, subject, body, tipo, "sistema")
		metadata, _ := json.Marshal(fields)
		event := dbpkg.SuperAlertaEvento{
			Tipo:           tipo,
			Severidad:      "info",
			Titulo:         titulo,
			Detalle:        firstNonEmptySuperAlertField(fields["Nombre"], fields["Correo"], tipo),
			Destinatario:   cfg.RecipientEmail,
			Asunto:         subject,
			Cuerpo:         body,
			CorreoEnviado:  sent,
			MetadataJSON:   string(metadata),
			UsuarioCreador: "sistema",
			Observaciones:  "evento_negocio_super_alertas",
		}
		if sendErr != nil {
			event.CorreoError = sendErr.Error()
			log.Printf("super_alertas: envio %s error: %v", tipo, sendErr)
		}
		if _, err := dbpkg.CreateSuperAlertaEvento(dbSuper, event); err != nil {
			log.Printf("super_alertas: registrar evento %s error: %v", tipo, err)
		}
	}()
}

func buildSuperBusinessAlertBody(titulo string, fields map[string]string) string {
	var b strings.Builder
	b.WriteString("Notificacion del sistema Powerful Control System.\r\n\r\n")
	b.WriteString("Evento: " + strings.TrimSpace(titulo) + "\r\n")
	b.WriteString("Fecha: " + time.Now().Format("2006-01-02 15:04:05") + "\r\n\r\n")
	order := []string{"ID administrador", "ID empresa", "Nombre", "Correo", "Telefono", "Pais", "Ciudad", "NIT", "Tipo ID", "Tipo", "Administrador", "Preconfiguracion", "Preconfiguracion aviso", "Origen"}
	written := map[string]bool{}
	for _, key := range order {
		value := strings.TrimSpace(fields[key])
		if value == "" {
			continue
		}
		b.WriteString("- " + key + ": " + value + "\r\n")
		written[key] = true
	}
	for key, value := range fields {
		if written[key] {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		b.WriteString("- " + key + ": " + value + "\r\n")
	}
	b.WriteString("\r\nPuedes desactivar este aviso en Super administrador > Alertas del sistema.")
	return b.String()
}

func boolLabel(v bool) string {
	if v {
		return "si"
	}
	return "no"
}

func firstNonEmptySuperAlertField(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func sendSuperSystemAlertEmail(dbSuper *sql.DB, toEmail, subject, body, tipo, usuario string) (bool, error) {
	toEmail = strings.TrimSpace(toEmail)
	if _, err := mail.ParseAddress(toEmail); err != nil {
		return false, fmt.Errorf("correo destino invalido: %w", err)
	}
	if dbSuper != nil && isEmpresaUsuarioMailTestMode(dbSuper) {
		metadataJSON := fmt.Sprintf(`{"tipo":%q,"modo":"test"}`, strings.TrimSpace(tipo))
		if err := captureEmpresaUsuarioMailNotification(dbSuper, dbpkg.SuperCorreoNotificacionTipoSistemaAlerta, 0, toEmail, subject, body, strings.TrimSpace(tipo), metadataJSON, usuario); err != nil {
			return false, err
		}
		return true, nil
	}
	if err := sendServerStartupEmail(dbSuper, toEmail, subject, body); err != nil {
		return false, err
	}
	return true, nil
}

func StartSuperAlertasWorker(dbSuper *sql.DB, interval time.Duration, stopCh <-chan struct{}) {
	if interval <= 0 {
		interval = time.Minute
	}
	if err := dbpkg.EnsureSuperAlertasSchema(dbSuper); err != nil {
		log.Printf("super_alertas: no se pudo inicializar esquema: %v", err)
		utils.ReportProcessError("super.alertas", "schema_init", "No se pudo inicializar el modulo de alertas", err, utils.ErrorLevelError, nil)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	EvaluateSuperAlertasSistema(dbSuper, false)
	for {
		select {
		case <-ticker.C:
			EvaluateSuperAlertasSistema(dbSuper, false)
		case <-stopCh:
			log.Println("super_alertas: worker stopped")
			return
		}
	}
}

func severityForPercent(value, threshold float64) string {
	if threshold <= 0 {
		return "warning"
	}
	if value >= math.Min(100, threshold+15) {
		return "critical"
	}
	return "warning"
}

func superAlertRound2(v float64) float64 {
	return math.Round(safeFloat(v)*100) / 100
}

func safeFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
