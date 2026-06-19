package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
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

	licenciaRetencionEmpresaConfigEnabled      = "licencias.retencion_empresas_vencidas.enabled"
	licenciaRetencionEmpresaConfigDiasEspera   = "licencias.retencion_empresas_vencidas.dias_espera"
	licenciaRetencionEmpresaConfigDiasPreaviso = "licencias.retencion_empresas_vencidas.dias_preaviso"
	licenciaRetencionEmpresaConfigMaxPorRun    = "licencias.retencion_empresas_vencidas.max_por_ejecucion"
	licenciaRetencionEmpresaConfigUltimoRun    = "licencias.retencion_empresas_vencidas.ultimo_run"
	licenciaRetencionEmpresaConfigUltimoResult = "licencias.retencion_empresas_vencidas.ultimo_resultado"
	licenciaRetencionEmpresaMailType           = "licencia_empresa_eliminacion_preaviso"
)

var licenciaVencimientoWorkerRunning int32

type licenciaVencimientoAlertasConfig struct {
	Enabled         bool                            `json:"enabled"`
	DiasAvisoRaw    string                          `json:"dias_aviso_raw"`
	DiasAviso       []int                           `json:"dias_aviso"`
	MaxPorEjecucion int                             `json:"max_por_ejecucion"`
	UltimoRun       string                          `json:"ultimo_run,omitempty"`
	UltimoResultado string                          `json:"ultimo_resultado,omitempty"`
	Retencion       licenciaRetencionEmpresasConfig `json:"retencion_empresas,omitempty"`
}

type licenciaRetencionEmpresasConfig struct {
	Enabled         bool   `json:"enabled"`
	DiasEspera      int    `json:"dias_espera"`
	DiasPreaviso    int    `json:"dias_preaviso"`
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

type licenciaRetencionEmpresasRunResult struct {
	OK          bool                                      `json:"ok"`
	DryRun      bool                                      `json:"dry_run"`
	Enabled     bool                                      `json:"enabled"`
	Evaluadas   int                                       `json:"evaluadas"`
	Preavisadas int                                       `json:"preavisadas"`
	Capturadas  int                                       `json:"capturadas"`
	Eliminadas  int                                       `json:"eliminadas"`
	Errores     int                                       `json:"errores"`
	SinCorreo   int                                       `json:"sin_correo"`
	Candidatas  []dbpkg.LicenciaEmpresaRetencionCandidate `json:"candidatas,omitempty"`
	Logs        []dbpkg.LicenciaEmpresaRetencionLog       `json:"logs,omitempty"`
	StartedAt   string                                    `json:"started_at"`
	FinishedAt  string                                    `json:"finished_at"`
	Message     string                                    `json:"message,omitempty"`
	Error       string                                    `json:"error,omitempty"`
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
		Retencion:       getLicenciaRetencionEmpresasConfig(dbSuper),
	}
}

func getLicenciaRetencionEmpresasConfig(dbSuper *sql.DB) licenciaRetencionEmpresasConfig {
	enabledRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaRetencionEmpresaConfigEnabled)
	diasEsperaRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaRetencionEmpresaConfigDiasEspera)
	diasPreavisoRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaRetencionEmpresaConfigDiasPreaviso)
	maxRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaRetencionEmpresaConfigMaxPorRun)
	lastRun, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaRetencionEmpresaConfigUltimoRun)
	lastResult, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, licenciaRetencionEmpresaConfigUltimoResult)
	diasEspera := 365
	if v, err := strconv.Atoi(strings.TrimSpace(diasEsperaRaw)); err == nil && v > 0 {
		diasEspera = v
	}
	if diasEspera > 3650 {
		diasEspera = 3650
	}
	diasPreaviso := 1
	if v, err := strconv.Atoi(strings.TrimSpace(diasPreavisoRaw)); err == nil && v > 0 {
		diasPreaviso = v
	}
	if diasPreaviso > diasEspera {
		diasPreaviso = diasEspera
	}
	maxRun := 25
	if v, err := strconv.Atoi(strings.TrimSpace(maxRaw)); err == nil && v > 0 {
		maxRun = v
	}
	if maxRun > 200 {
		maxRun = 200
	}
	return licenciaRetencionEmpresasConfig{
		Enabled:         parseEmpresaUsuarioBool(enabledRaw, false),
		DiasEspera:      diasEspera,
		DiasPreaviso:    diasPreaviso,
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
	if err := dbpkg.SetConfigValue(dbSuper, licenciaVencimientoConfigMaxPorRun, strconv.Itoa(maxRun), false); err != nil {
		return err
	}
	return saveLicenciaRetencionEmpresasConfig(dbSuper, cfg.Retencion)
}

func saveLicenciaRetencionEmpresasConfig(dbSuper *sql.DB, cfg licenciaRetencionEmpresasConfig) error {
	if dbSuper == nil {
		return fmt.Errorf("db super no disponible")
	}
	if err := dbpkg.SetConfigValue(dbSuper, licenciaRetencionEmpresaConfigEnabled, strconv.FormatBool(cfg.Enabled), false); err != nil {
		return err
	}
	diasEspera := cfg.DiasEspera
	if diasEspera <= 0 {
		diasEspera = 365
	}
	if diasEspera > 3650 {
		diasEspera = 3650
	}
	diasPreaviso := cfg.DiasPreaviso
	if diasPreaviso <= 0 {
		diasPreaviso = 1
	}
	if diasPreaviso > diasEspera {
		diasPreaviso = diasEspera
	}
	maxRun := cfg.MaxPorEjecucion
	if maxRun <= 0 {
		maxRun = 25
	}
	if maxRun > 200 {
		maxRun = 200
	}
	if err := dbpkg.SetConfigValue(dbSuper, licenciaRetencionEmpresaConfigDiasEspera, strconv.Itoa(diasEspera), false); err != nil {
		return err
	}
	if err := dbpkg.SetConfigValue(dbSuper, licenciaRetencionEmpresaConfigDiasPreaviso, strconv.Itoa(diasPreaviso), false); err != nil {
		return err
	}
	return dbpkg.SetConfigValue(dbSuper, licenciaRetencionEmpresaConfigMaxPorRun, strconv.Itoa(maxRun), false)
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

func ProcessLicenciaRetencionEmpresas(dbSuper, dbEmp *sql.DB, dryRun bool, actor string) licenciaRetencionEmpresasRunResult {
	start := time.Now()
	result := licenciaRetencionEmpresasRunResult{
		OK:        true,
		DryRun:    dryRun,
		StartedAt: start.Format("2006-01-02 15:04:05"),
	}
	if dbSuper == nil || dbEmp == nil {
		result.OK = false
		result.Error = "db super u operativa no disponible"
		result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
		return result
	}
	cfg := getLicenciaRetencionEmpresasConfig(dbSuper)
	result.Enabled = cfg.Enabled
	if !cfg.Enabled && !dryRun {
		result.Message = "retencion de empresas vencidas desactivada"
		result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
		return result
	}
	limit := cfg.MaxPorEjecucion
	if dryRun {
		limit = 50
	}
	candidates, err := dbpkg.ListLicenciaEmpresaRetencionCandidates(dbSuper, dbEmp, cfg.DiasEspera, cfg.DiasPreaviso, start, limit)
	if err != nil {
		result.OK = false
		result.Error = err.Error()
		result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
		return result
	}
	result.Evaluadas = len(candidates)
	if dryRun {
		result.Candidatas = candidates
		result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
		return result
	}
	for _, c := range candidates {
		if c.DebePreavisar {
			if strings.TrimSpace(c.AdminEmail) == "" {
				result.SinCorreo++
				_ = dbpkg.UpsertLicenciaEmpresaRetencionLog(dbSuper, c, "error", "empresa sin correo administrador para preaviso", "", actor)
				continue
			}
			estado, err := sendLicenciaRetencionEmpresaNoticeEmail(dbSuper, c, actor)
			if err != nil {
				result.Errores++
				_ = dbpkg.UpsertLicenciaEmpresaRetencionLog(dbSuper, c, "error", err.Error(), "", actor)
				continue
			}
			if strings.EqualFold(estado, "preaviso_capturado") {
				result.Capturadas++
			} else {
				result.Preavisadas++
			}
			_ = dbpkg.UpsertLicenciaEmpresaRetencionLog(dbSuper, c, estado, "", "", actor)
			continue
		}
		if c.DebeEliminar {
			_ = dbpkg.UpsertLicenciaEmpresaRetencionLog(dbSuper, c, "eliminacion_en_proceso", "", "", actor)
			deleteResult, err := dbpkg.DeleteEmpresaCascade(dbEmp, dbSuper, c.EmpresaID)
			if err != nil {
				result.Errores++
				_ = dbpkg.UpsertLicenciaEmpresaRetencionLog(dbSuper, c, "error", err.Error(), "", actor)
				continue
			}
			fileCleanup := cleanupEmpresaOwnedFiles(c.EmpresaID)
			markResult := map[string]interface{}{
				"db":       deleteResult,
				"archivos": fileCleanup,
			}
			_ = dbpkg.UpsertLicenciaEmpresaRetencionDeleted(dbSuper, c, markResult, actor)
			result.Eliminadas++
		}
	}
	result.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
	summary := fmt.Sprintf("evaluadas=%d preavisadas=%d capturadas=%d eliminadas=%d errores=%d sin_correo=%d", result.Evaluadas, result.Preavisadas, result.Capturadas, result.Eliminadas, result.Errores, result.SinCorreo)
	_ = dbpkg.SetConfigValue(dbSuper, licenciaRetencionEmpresaConfigUltimoRun, result.FinishedAt, false)
	_ = dbpkg.SetConfigValue(dbSuper, licenciaRetencionEmpresaConfigUltimoResult, summary, false)
	result.Message = summary
	return result
}

func sendLicenciaRetencionEmpresaNoticeEmail(dbSuper *sql.DB, c dbpkg.LicenciaEmpresaRetencionCandidate, actor string) (string, error) {
	if _, err := mail.ParseAddress(strings.TrimSpace(c.AdminEmail)); err != nil {
		return "", fmt.Errorf("correo administrador invalido")
	}
	baseURL := resolveLicenciaVencimientoBaseURL(dbSuper)
	renewURL := buildLicenciaVencimientoRenewURL(baseURL, c.EmpresaID)
	adminName := strings.TrimSpace(c.AdminNombre)
	if adminName == "" {
		adminName = "administrador"
	}
	subject, bodyPlain, bodyHTML, err := applySuperEmailTemplate(dbSuper, superEmailTemplateKeyLicenciaEmpresaDeletionWarning, map[string]string{
		"name":             adminName,
		"company_name":     c.EmpresaNombre,
		"last_license_end": c.UltimaLicenciaFin,
		"deletion_date":    c.FechaProgramadaEliminacion,
		"retention_days":   strconv.Itoa(c.RetencionDias),
		"notice_days":      strconv.Itoa(c.PreavisoDias),
		"renew_url":        renewURL,
		"support_url":      strings.TrimRight(baseURL, "/") + "/ayuda/ayuda.html",
	})
	if err != nil {
		return "", err
	}
	if isEmpresaUsuarioMailTestMode(dbSuper) {
		metadata, _ := json.Marshal(map[string]interface{}{
			"mail_mode":                    "test",
			"empresa_id":                   c.EmpresaID,
			"ultima_licencia_fin":          c.UltimaLicenciaFin,
			"fecha_programada_eliminacion": c.FechaProgramadaEliminacion,
			"retencion_dias":               c.RetencionDias,
			"preaviso_dias":                c.PreavisoDias,
			"renew_url":                    renewURL,
		})
		if err := captureEmpresaUsuarioMailNotification(dbSuper, licenciaRetencionEmpresaMailType, c.EmpresaID, c.AdminEmail, subject, bodyPlain, "", string(metadata), actor); err != nil {
			return "", err
		}
		return "preaviso_capturado", nil
	}
	if err := sendLicenciaVencimientoSMTP(dbSuper, c.AdminEmail, subject, bodyPlain, bodyHTML, baseURL); err != nil {
		return "", err
	}
	return "preaviso_enviado", nil
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
	fromName, fromEmail := corporateSystemSenderAddress(dbSuper, "ventas")
	msg := buildEmpresaUsuarioMultipartMessage(dbSuper, baseURL, fromName, fromEmail, strings.TrimSpace(toEmail), sanitizeEmailHeader(subject), bodyPlain, bodyHTML)
	return sendEmpresaUsuarioMailuMessage(dbSuper, fromEmail, strings.TrimSpace(toEmail), []byte(msg))
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
			retencionPreview, _ := dbpkg.ListLicenciaEmpresaRetencionCandidates(dbSuper, dbEmp, cfg.Retencion.DiasEspera, cfg.Retencion.DiasPreaviso, time.Now(), 25)
			retencionLogs, _ := dbpkg.ListLicenciaEmpresaRetencionLogs(dbSuper, 25, "")
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                true,
				"config":            cfg,
				"preview":           preview,
				"evaluadas":         evaluated,
				"omitidas_previas":  alreadySent,
				"logs":              logs,
				"retencion_preview": retencionPreview,
				"retencion_logs":    retencionLogs,
			})
			return
		case http.MethodPut, http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "retencion_run_now" || action == "retencion_preview" {
				dryRun := action == "retencion_preview"
				result := ProcessLicenciaRetencionEmpresas(dbSuper, dbEmp, dryRun, adminEmail)
				if logs, err := dbpkg.ListLicenciaEmpresaRetencionLogs(dbSuper, 50, ""); err == nil {
					result.Logs = logs
				}
				status := http.StatusOK
				if !result.OK {
					status = http.StatusInternalServerError
				}
				writeJSON(w, status, result)
				return
			}
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
		retencionResult := ProcessLicenciaRetencionEmpresas(dbSuper, dbEmp, false, "sistema."+origin)
		if !result.OK {
			log.Printf("[licencias_vencimiento] worker %s error: %s", origin, result.Error)
		}
		if !retencionResult.OK {
			log.Printf("[licencias_retencion_empresas] worker %s error: %s", origin, retencionResult.Error)
		}
		if result.Enviadas > 0 || result.Capturadas > 0 || result.Errores > 0 {
			log.Printf("[licencias_vencimiento] worker %s: evaluadas=%d pendientes=%d enviadas=%d capturadas=%d errores=%d sin_correo=%d", origin, result.Evaluadas, result.Pendientes, result.Enviadas, result.Capturadas, result.Errores, result.SinCorreo)
		}
		if retencionResult.Preavisadas > 0 || retencionResult.Capturadas > 0 || retencionResult.Eliminadas > 0 || retencionResult.Errores > 0 {
			log.Printf("[licencias_retencion_empresas] worker %s: evaluadas=%d preavisadas=%d capturadas=%d eliminadas=%d errores=%d sin_correo=%d", origin, retencionResult.Evaluadas, retencionResult.Preavisadas, retencionResult.Capturadas, retencionResult.Eliminadas, retencionResult.Errores, retencionResult.SinCorreo)
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
