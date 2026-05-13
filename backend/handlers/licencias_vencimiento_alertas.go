package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	licenciaVencimientoConfigEnabled      = "licencias.vencimiento_alertas.enabled"
	licenciaVencimientoConfigDiasAviso    = "licencias.vencimiento_alertas.dias_aviso"
	licenciaVencimientoConfigMaxPorRun    = "licencias.vencimiento_alertas.max_por_ejecucion"
	licenciaVencimientoConfigUltimoRun    = "licencias.vencimiento_alertas.ultimo_run"
	licenciaVencimientoConfigUltimoResult = "licencias.vencimiento_alertas.ultimo_resultado"
	licenciaVencimientoMailType           = "licencia_vencimiento_alerta"
)

var licenciaVencimientoWorkerRunning int32

type licenciaVencimientoAlertasConfig struct {
	Enabled         bool   `json:"enabled"`
	DiasAvisoRaw    string `json:"dias_aviso_raw"`
	DiasAviso       []int  `json:"dias_aviso"`
	MaxPorEjecucion int    `json:"max_por_ejecucion"`
	UltimoRun       string `json:"ultimo_run,omitempty"`
	UltimoResultado string `json:"ultimo_resultado,omitempty"`
}

type licenciaVencimientoRunResult struct {
	OK              bool                                       `json:"ok"`
	DryRun          bool                                       `json:"dry_run"`
	Enabled         bool                                       `json:"enabled"`
	Evaluadas       int                                        `json:"evaluadas"`
	Pendientes      int                                        `json:"pendientes"`
	Enviadas        int                                        `json:"enviadas"`
	Capturadas      int                                        `json:"capturadas"`
	Errores         int                                        `json:"errores"`
	SinCorreo       int                                        `json:"sin_correo"`
	OmitidasPrevias int                                        `json:"omitidas_previas"`
	Candidatas      []dbpkg.LicenciaVencimientoCandidate       `json:"candidatas,omitempty"`
	Logs            []dbpkg.LicenciaVencimientoNotificationLog `json:"logs,omitempty"`
	StartedAt       string                                     `json:"started_at"`
	FinishedAt      string                                     `json:"finished_at"`
	Message         string                                     `json:"message,omitempty"`
	Error           string                                     `json:"error,omitempty"`
}

func getLicenciaVencimientoAlertasConfig(dbSuper *sql.DB) licenciaVencimientoAlertasConfig {
	enabledRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaVencimientoConfigEnabled)
	diasRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaVencimientoConfigDiasAviso)
	maxRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaVencimientoConfigMaxPorRun)
	lastRun, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaVencimientoConfigUltimoRun)
	lastResult, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaVencimientoConfigUltimoResult)

	dias := parseLicenciaVencimientoDiasAviso(diasRaw)
	if strings.TrimSpace(diasRaw) == "" {
		diasRaw = "15,7,3,1"
	}
	maxRun := 100
	if v, err := strconv.Atoi(strings.TrimSpace(maxRaw)); err == nil && v > 0 {
		maxRun = v
	}
	if maxRun > 500 {
		maxRun = 500
	}
	return licenciaVencimientoAlertasConfig{
		Enabled:         parseEmpresaUsuarioBool(enabledRaw, true),
		DiasAvisoRaw:    diasRaw,
		DiasAviso:       dias,
		MaxPorEjecucion: maxRun,
		UltimoRun:       strings.TrimSpace(lastRun),
		UltimoResultado: strings.TrimSpace(lastResult),
	}
}

func parseLicenciaVencimientoDiasAviso(raw string) []int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []int{15, 7, 3, 1}
	}
	seen := map[int]bool{}
	out := make([]int, 0)
	for _, part := range strings.FieldsFunc(raw, func(r rune) bool { return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t' }) {
		v, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil || v <= 0 || v > 365 || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	if len(out) == 0 {
		out = []int{15, 7, 3, 1}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(out)))
	return out
}

func maxLicenciaVencimientoDiasAviso(values []int) int {
	max := 0
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func selectLicenciaVencimientoDiaAviso(diasRestantes int, thresholds []int) int {
	if diasRestantes < 0 {
		return 0
	}
	if len(thresholds) == 0 {
		return 0
	}
	copyVals := append([]int(nil), thresholds...)
	sort.Ints(copyVals)
	for _, threshold := range copyVals {
		if diasRestantes <= threshold {
			return threshold
		}
	}
	return 0
}

func saveLicenciaVencimientoAlertasConfig(dbSuper *sql.DB, cfg licenciaVencimientoAlertasConfig) error {
	if dbSuper == nil {
		return fmt.Errorf("db super no disponible")
	}
	if err := dbpkg.SetConfigValue(dbSuper, licenciaVencimientoConfigEnabled, strconv.FormatBool(cfg.Enabled), false); err != nil {
		return err
	}
	diasRaw := strings.TrimSpace(cfg.DiasAvisoRaw)
	dias := parseLicenciaVencimientoDiasAviso(diasRaw)
	if diasRaw == "" {
		parts := make([]string, 0, len(dias))
		for _, d := range dias {
			parts = append(parts, strconv.Itoa(d))
		}
		diasRaw = strings.Join(parts, ",")
	}
	if err := dbpkg.SetConfigValue(dbSuper, licenciaVencimientoConfigDiasAviso, diasRaw, false); err != nil {
		return err
	}
	maxRun := cfg.MaxPorEjecucion
	if maxRun <= 0 {
		maxRun = 100
	}
	if maxRun > 500 {
		maxRun = 500
	}
	return dbpkg.SetConfigValue(dbSuper, licenciaVencimientoConfigMaxPorRun, strconv.Itoa(maxRun), false)
}

func buildLicenciaVencimientoPending(dbSuper, dbEmp *sql.DB, cfg licenciaVencimientoAlertasConfig, now time.Time, limit int) ([]dbpkg.LicenciaVencimientoCandidate, int, int, error) {
	maxDays := maxLicenciaVencimientoDiasAviso(cfg.DiasAviso)
	candidates, err := dbpkg.ListLicenciaVencimientoCandidates(dbSuper, dbEmp, maxDays, now)
	if err != nil {
		return nil, 0, 0, err
	}
	pending := make([]dbpkg.LicenciaVencimientoCandidate, 0)
	alreadySent := 0
	for _, c := range candidates {
		c.DiasAviso = selectLicenciaVencimientoDiaAviso(c.DiasRestantes, cfg.DiasAviso)
		if c.DiasAviso <= 0 {
			continue
		}
		sent, sentErr := dbpkg.LicenciaVencimientoNotificationSent(dbSuper, c.LicenciaTipo, c.LicenciaID, c.EmpresaID, c.AdminEmail, c.DiasAviso, c.FechaFin)
		if sentErr != nil {
			return nil, len(candidates), alreadySent, sentErr
		}
		if sent {
			alreadySent++
			continue
		}
		pending = append(pending, c)
		if limit > 0 && len(pending) >= limit {
			break
		}
	}
	return pending, len(candidates), alreadySent, nil
}

func ProcessLicenciaVencimientoAlertas(dbSuper, dbEmp *sql.DB, dryRun bool, actor string) licenciaVencimientoRunResult {
	start := time.Now()
	result := licenciaVencimientoRunResult{
		OK:        true,
		DryRun:    dryRun,
		StartedAt: start.Format("2006-01-02 15:04:05"),
	}
	if dbSuper == nil {
		result.OK = false
		result.Error = "db super no disponible"
		result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
		return result
	}
	cfg := getLicenciaVencimientoAlertasConfig(dbSuper)
	result.Enabled = cfg.Enabled
	if !cfg.Enabled && !dryRun {
		result.Message = "alertas de vencimiento desactivadas"
		result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
		return result
	}
	limit := cfg.MaxPorEjecucion
	if dryRun {
		limit = 50
	}
	pending, evaluated, alreadySent, err := buildLicenciaVencimientoPending(dbSuper, dbEmp, cfg, start, limit)
	if err != nil {
		result.OK = false
		result.Error = err.Error()
		result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
		return result
	}
	result.Evaluadas = evaluated
	result.OmitidasPrevias = alreadySent
	result.Pendientes = len(pending)
	if dryRun {
		result.Candidatas = pending
		result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
		return result
	}

	for _, c := range pending {
		if strings.TrimSpace(c.AdminEmail) == "" {
			result.SinCorreo++
			_ = dbpkg.UpsertLicenciaVencimientoNotificationResult(dbSuper, c, "error", "empresa sin correo administrador", actor)
			continue
		}
		estado, err := sendLicenciaVencimientoAlertEmail(dbSuper, c, actor)
		if err != nil {
			result.Errores++
			_ = dbpkg.UpsertLicenciaVencimientoNotificationResult(dbSuper, c, "error", err.Error(), actor)
			continue
		}
		if strings.EqualFold(estado, "capturado") {
			result.Capturadas++
		} else {
			result.Enviadas++
		}
		_ = dbpkg.UpsertLicenciaVencimientoNotificationResult(dbSuper, c, estado, "", actor)
	}
	result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
	summary := fmt.Sprintf("evaluadas=%d pendientes=%d enviadas=%d capturadas=%d errores=%d sin_correo=%d", result.Evaluadas, result.Pendientes, result.Enviadas, result.Capturadas, result.Errores, result.SinCorreo)
	_ = dbpkg.SetConfigValue(dbSuper, licenciaVencimientoConfigUltimoRun, result.FinishedAt, false)
	_ = dbpkg.SetConfigValue(dbSuper, licenciaVencimientoConfigUltimoResult, summary, false)
	result.Message = summary
	return result
}

func sendLicenciaVencimientoAlertEmail(dbSuper *sql.DB, c dbpkg.LicenciaVencimientoCandidate, actor string) (string, error) {
	if _, err := mail.ParseAddress(strings.TrimSpace(c.AdminEmail)); err != nil {
		return "", fmt.Errorf("correo administrador invalido")
	}
	baseURL := resolveLicenciaVencimientoBaseURL(dbSuper)
	renewURL := buildLicenciaVencimientoRenewURL(baseURL, c.EmpresaID)
	adminName := strings.TrimSpace(c.AdminNombre)
	if adminName == "" {
		adminName = "administrador"
	}
	daysLabel := strconv.Itoa(c.DiasRestantes)
	if c.DiasRestantes == 1 {
		daysLabel = "1"
	}
	subject, bodyPlain, bodyHTML, err := applySuperEmailTemplate(dbSuper, superEmailTemplateKeyLicenciaExpiryWarning, map[string]string{
		"name":             adminName,
		"company_name":     c.EmpresaNombre,
		"license_name":     c.LicenciaNombre,
		"license_type":     c.LicenciaTipo,
		"days_remaining":   daysLabel,
		"notice_threshold": strconv.Itoa(c.DiasAviso),
		"end_date":         c.FechaFin,
		"renew_url":        renewURL,
	})
	if err != nil {
		return "", err
	}
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadata, _ := json.Marshal(map[string]interface{}{
			"mail_mode":      "test",
			"empresa_id":     c.EmpresaID,
			"licencia_tipo":  c.LicenciaTipo,
			"licencia_id":    c.LicenciaID,
			"dias_restantes": c.DiasRestantes,
			"dias_aviso":     c.DiasAviso,
			"fecha_fin":      c.FechaFin,
			"renew_url":      renewURL,
		})
		if err := captureEmpresaUsuarioMailNotification(dbSuper, licenciaVencimientoMailType, c.EmpresaID, c.AdminEmail, subject, bodyPlain, "", string(metadata), actor); err != nil {
			return "", err
		}
		return "capturado", nil
	}
	if err := sendLicenciaVencimientoSMTP(dbSuper, c.AdminEmail, subject, bodyPlain, bodyHTML, baseURL); err != nil {
		return "", err
	}
	return "enviado", nil
}

func resolveLicenciaVencimientoBaseURL(dbSuper *sql.DB) string {
	baseURL := ""
	if dbSuper != nil {
		baseURL, _ = getDecryptedConfigValue(dbSuper, "gmail.confirm_base_url")
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "https://powerfulcontrolsystem.com"
	}
	return strings.TrimRight(baseURL, "/")
}

func buildLicenciaVencimientoRenewURL(baseURL string, empresaID int64) string {
	u, err := url.Parse(strings.TrimRight(baseURL, "/") + "/elegir_licencia.html")
	if err != nil {
		return strings.TrimRight(baseURL, "/") + "/elegir_licencia.html"
	}
	q := u.Query()
	if empresaID > 0 {
		q.Set("empresa_id", strconv.FormatInt(empresaID, 10))
	}
	q.Set("origen", "alerta_vencimiento")
	u.RawQuery = q.Encode()
	return u.String()
}

func sendLicenciaVencimientoSMTP(dbSuper *sql.DB, toEmail, subject, bodyPlain, bodyHTML, baseURL string) error {
	smtpEmail, err := getDecryptedConfigValue(dbSuper, "gmail.smtp_email")
	if err != nil {
		return err
	}
	smtpEmail = strings.TrimSpace(smtpEmail)
	if smtpEmail == "" {
		return fmt.Errorf("gmail.smtp_email no configurado")
	}
	if _, err := mail.ParseAddress(smtpEmail); err != nil {
		return fmt.Errorf("gmail.smtp_email invalido")
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
		if h, _, err := net.SplitHostPort(smtpHost); err == nil {
			mailHostForAuth = h
		}
	}
	addr := smtpHost
	if !strings.Contains(addr, ":") {
		addr = smtpHost + ":" + smtpPort
	}
	auth := smtp.PlainAuth("", smtpEmail, smtpPass, mailHostForAuth)
	boundary := "==PCS_LICENSE_EXPIRY_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	listUnsub := ""
	if u, err := url.Parse(baseURL); err == nil {
		host := u.Host
		if strings.Contains(host, ":") {
			host, _, _ = net.SplitHostPort(host)
		}
		if host != "" {
			listUnsub = "<mailto:postmaster@" + host + ">"
		}
	}
	from := (&mail.Address{Name: fromName, Address: smtpEmail}).String()
	headers := "From: " + from + "\r\n" +
		"To: " + strings.TrimSpace(toEmail) + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("utf-8", sanitizeEmailHeader(subject)) + "\r\n"
	if listUnsub != "" {
		headers += "List-Unsubscribe: " + listUnsub + "\r\n"
	}
	headers += "MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n\r\n"
	msg := headers +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n\r\n" +
		bodyPlain + "\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n\r\n" +
		bodyHTML + "\r\n" +
		"--" + boundary + "--\r\n"
	return smtp.SendMail(addr, auth, smtpEmail, []string{strings.TrimSpace(toEmail)}, []byte(msg))
}

func sanitizeEmailHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

func SuperLicenciaVencimientoAlertasHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			cfg := getLicenciaVencimientoAlertasConfig(dbSuper)
			preview, evaluated, alreadySent, err := buildLicenciaVencimientoPending(dbSuper, dbEmp, cfg, time.Now(), 25)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": err.Error()})
				return
			}
			logs, _ := dbpkg.ListLicenciaVencimientoNotificationLogs(dbSuper, 25)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":               true,
				"config":           cfg,
				"preview":          preview,
				"evaluadas":        evaluated,
				"omitidas_previas": alreadySent,
				"logs":             logs,
			})
			return
		case http.MethodPut, http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "run_now" || action == "preview" {
				dryRun := action == "preview"
				result := ProcessLicenciaVencimientoAlertas(dbSuper, dbEmp, dryRun, adminEmail)
				status := http.StatusOK
				if !result.OK {
					status = http.StatusInternalServerError
				}
				writeJSON(w, status, result)
				return
			}
			var payload licenciaVencimientoAlertasConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := saveLicenciaVencimientoAlertasConfig(dbSuper, payload); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": err.Error()})
				return
			}
			cfg := getLicenciaVencimientoAlertasConfig(dbSuper)
			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "config": cfg})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func StartLicenciaVencimientoAlertasWorker(dbSuper, dbEmp *sql.DB, interval time.Duration, stop <-chan struct{}) {
	if dbSuper == nil {
		return
	}
	if interval <= 0 {
		interval = 12 * time.Hour
	}
	run := func(origin string) {
		if !atomic.CompareAndSwapInt32(&licenciaVencimientoWorkerRunning, 0, 1) {
			return
		}
		defer atomic.StoreInt32(&licenciaVencimientoWorkerRunning, 0)
		result := ProcessLicenciaVencimientoAlertas(dbSuper, dbEmp, false, "sistema."+origin)
		if !result.OK {
			log.Printf("[licencias_vencimiento] worker %s error: %s", origin, result.Error)
			return
		}
		if result.Enviadas > 0 || result.Capturadas > 0 || result.Errores > 0 {
			log.Printf("[licencias_vencimiento] worker %s: evaluadas=%d pendientes=%d enviadas=%d capturadas=%d errores=%d sin_correo=%d", origin, result.Evaluadas, result.Pendientes, result.Enviadas, result.Capturadas, result.Errores, result.SinCorreo)
		}
	}
	run("startup")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			run("periodic")
		case <-stop:
			return
		}
	}
}
