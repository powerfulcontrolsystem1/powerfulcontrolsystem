package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	reportesIAModelTextUsageID   = "openai:gpt-5.4-mini:reportes_texto"
	reportesIAModelReportUsageID = "openai:gpt-5.4-mini:reportes_export"
	reportesIATextDailyLimit     = 10
	reportesIAReportDailyLimit   = 2
)

// EmpresaReportesIAChatHandler agrega un chat compacto dentro de reportes.
// - modo texto: GPT-5.4 mini, 10 preguntas/dia por empresa.
// - modo reporte: GPT-5.4 mini, 2 reportes/dia por empresa; interpreta dataset/formato y devuelve URL de exportacion.
func EmpresaReportesIAChatHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if !isSuperAIEnabled(dbSuper) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":    false,
				"code":  "ai_disabled",
				"error": "La IA esta desactivada desde super administrador.",
			})
			return
		}
		empresaChatEnabled, _, _, err := getChatIAEmpresaEnabled(dbSuper)
		if err != nil || !empresaChatEnabled {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":    false,
				"code":  "ai_empresa_disabled",
				"error": "El chat IA empresarial esta desactivado.",
			})
			return
		}

		var payload struct {
			EmpresaID int64                  `json:"empresa_id"`
			Modo      string                 `json:"modo"`
			Pregunta  string                 `json:"pregunta"`
			Dataset   string                 `json:"dataset"`
			Format    string                 `json:"format"`
			Desde     string                 `json:"desde"`
			Hasta     string                 `json:"hasta"`
			Historial []empresaAIChatMensaje `json:"historial"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		pregunta := strings.TrimSpace(payload.Pregunta)
		if pregunta == "" {
			http.Error(w, "pregunta es obligatoria", http.StatusBadRequest)
			return
		}
		if len([]rune(pregunta)) > 2500 {
			r := []rune(pregunta)
			pregunta = string(r[:2500])
		}

		modo := strings.ToLower(strings.TrimSpace(payload.Modo))
		if modo == "" {
			modo = "texto"
		}
		switch modo {
		case "texto", "pregunta", "chat":
			handleEmpresaReportesIATexto(w, r, dbEmp, dbSuper, payload.EmpresaID, pregunta, payload.Historial, payload.Desde, payload.Hasta)
		case "reporte", "export", "exportar":
			handleEmpresaReportesIAReporte(w, r, dbEmp, dbSuper, payload.EmpresaID, pregunta, payload.Historial, payload.Dataset, payload.Format, payload.Desde, payload.Hasta)
		case "nuevo", "personalizado", "custom", "reporte_nuevo":
			handleEmpresaReportesIANuevo(w, r, dbEmp, dbSuper, payload.EmpresaID, pregunta, payload.Historial, payload.Desde, payload.Hasta)
		default:
			http.Error(w, "modo invalido (use texto, reporte o nuevo)", http.StatusBadRequest)
		}
	}
}

func handleEmpresaReportesIATexto(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, pregunta string, historial []empresaAIChatMensaje, desde, hasta string) {
	model, ok := availableEmpresaAIModelMap(dbSuper)["openai:gpt-5.4-mini"]
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "code": "ai_model_missing", "error": "GPT-5.4 mini no esta disponible."})
		return
	}
	fechaUso := time.Now().Format("2006-01-02")
	uso, err := dbpkg.GetEmpresaAIUsoDiario(dbEmp, empresaID, model.Provider, reportesIAModelTextUsageID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario de reportes IA", http.StatusInternalServerError)
		return
	}
	if uso.Consultas >= reportesIATextDailyLimit {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok": false, "code": "reportes_ia_text_limit_reached",
			"error": "Se alcanzo el limite diario de 10 preguntas de reportes IA para esta empresa.",
			"usage": reportesIAUsagePayload(uso.Consultas, reportesIATextDailyLimit, 0, 0),
		})
		return
	}

	builder := newReportesAIBuilder(dbEmp, empresaID, desde, hasta)
	contexto, err := buildReportesAITextContext(builder)
	if err != nil {
		http.Error(w, "No se pudo construir contexto de reportes", http.StatusInternalServerError)
		return
	}
	ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 45 * time.Second}}
	system := "Eres un analista de reportes del sistema POS multiempresa. Responde en espanol, breve y con datos del contexto. No inventes cifras fuera del contexto. Si el usuario pide generar/exportar un archivo, indica que use el modo Reporte IA.\n\n" + contexto
	resp, pt, ct, err := ctrl.generateResponseWithSystemPrompt(model, pregunta, sanitizeHistorial(historial, 6), system)
	if err != nil {
		http.Error(w, "No se pudo generar respuesta IA: "+err.Error(), http.StatusBadGateway)
		return
	}
	if _, err := dbpkg.RegisterEmpresaAIConsulta(dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID: empresaID, Provider: model.Provider, ModelID: reportesIAModelTextUsageID,
		Pregunta: pregunta, Respuesta: strings.TrimSpace(resp), PromptTokens: pt, CompletionTokens: ct, TotalTokens: pt + ct,
		FechaConsulta: time.Now().Format("2006-01-02 15:04:05"), PlanActual: strings.TrimSpace(uso.PlanActual),
		UsuarioCreador: adminEmailFromRequest(r), Estado: "activo", Observaciones: "reportes_ia_texto",
	}); err != nil {
		http.Error(w, "No se pudo registrar uso IA", http.StatusInternalServerError)
		return
	}
	usoNuevo, _ := dbpkg.GetEmpresaAIUsoDiario(dbEmp, empresaID, model.Provider, reportesIAModelTextUsageID, fechaUso)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "modo": "texto", "respuesta": strings.TrimSpace(resp),
		"usage": reportesIAUsagePayload(usoNuevo.Consultas, reportesIATextDailyLimit, pt, ct),
	})
}

func handleEmpresaReportesIAReporte(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, pregunta string, historial []empresaAIChatMensaje, dataset, format, desde, hasta string) {
	model, ok := availableEmpresaAIModelMap(dbSuper)["openai:gpt-5.4-mini"]
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "code": "ai_model_missing", "error": "GPT-5.4 mini no esta disponible."})
		return
	}
	fechaUso := time.Now().Format("2006-01-02")
	uso, err := dbpkg.GetEmpresaAIUsoDiario(dbEmp, empresaID, model.Provider, reportesIAModelReportUsageID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario de reportes IA", http.StatusInternalServerError)
		return
	}
	if uso.Consultas >= reportesIAReportDailyLimit {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok": false, "code": "reportes_ia_report_limit_reached",
			"error": "Se alcanzo el limite diario de 2 reportes IA para esta empresa.",
			"usage": reportesIAUsagePayload(uso.Consultas, reportesIAReportDailyLimit, 0, 0),
		})
		return
	}

	builder := newReportesAIBuilder(dbEmp, empresaID, desde, hasta)
	contexto := buildReportesIAReportContext(dataset, format)
	ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 60 * time.Second}}
	system := "Eres un asistente de reportes. Debes elegir el dataset y formato mas apropiado para exportar. Responde SOLO JSON valido con keys: dataset, format, title, message. No uses markdown. Si el usuario ya envio dataset/format validos, respetalos.\n\n" + contexto
	raw, pt, ct, err := ctrl.generateResponseWithSystemPrompt(model, pregunta, sanitizeHistorial(historial, 4), system)
	if err != nil {
		http.Error(w, "No se pudo interpretar reporte IA: "+err.Error(), http.StatusBadGateway)
		return
	}
	choice := parseReportesIAChoice(raw)
	if strings.TrimSpace(dataset) != "" {
		choice.Dataset = strings.ToLower(strings.TrimSpace(dataset))
	}
	if strings.TrimSpace(format) != "" {
		choice.Format = strings.ToLower(strings.TrimSpace(format))
	}
	choice.Dataset = strings.ToLower(strings.TrimSpace(choice.Dataset))
	choice.Format = normalizeReportesIAFormat(choice.Format)
	if choice.Dataset == "" {
		choice.Dataset = reporteDatasetEmpresarialTablero
	}
	if _, err := builder.buildDataset(choice.Dataset); err != nil {
		http.Error(w, "dataset sugerido no soportado: "+err.Error(), http.StatusBadRequest)
		return
	}
	exportURL := buildReportesIAExportURL(empresaID, choice.Dataset, choice.Format, desde, hasta)
	msg := strings.TrimSpace(choice.Message)
	if msg == "" {
		msg = "Reporte listo para exportar."
	}
	if _, err := dbpkg.RegisterEmpresaAIConsulta(dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID: empresaID, Provider: model.Provider, ModelID: reportesIAModelReportUsageID,
		Pregunta: pregunta, Respuesta: strings.TrimSpace(raw), PromptTokens: pt, CompletionTokens: ct, TotalTokens: pt + ct,
		FechaConsulta: time.Now().Format("2006-01-02 15:04:05"), PlanActual: strings.TrimSpace(uso.PlanActual),
		UsuarioCreador: adminEmailFromRequest(r), Estado: "activo", Observaciones: "reportes_ia_export:" + choice.Dataset + ":" + choice.Format,
	}); err != nil {
		http.Error(w, "No se pudo registrar uso IA", http.StatusInternalServerError)
		return
	}
	usoNuevo, _ := dbpkg.GetEmpresaAIUsoDiario(dbEmp, empresaID, model.Provider, reportesIAModelReportUsageID, fechaUso)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "modo": "reporte", "respuesta": msg,
		"dataset": choice.Dataset, "format": choice.Format, "title": strings.TrimSpace(choice.Title),
		"export_url": exportURL,
		"usage":      reportesIAUsagePayload(usoNuevo.Consultas, reportesIAReportDailyLimit, pt, ct),
	})
}

// handleEmpresaReportesIANuevo crea una vista nueva a partir de un contrato
// declarativo. La IA solo propone un ReportSpec; el servidor lo valida y lo
// ejecuta sobre datasets canonicos aislados por empresa.
func handleEmpresaReportesIANuevo(w http.ResponseWriter, r *http.Request, dbEmp, dbSuper *sql.DB, empresaID int64, pregunta string, historial []empresaAIChatMensaje, desde, hasta string) {
	model, ok := availableEmpresaAIModelMap(dbSuper)["openai:gpt-5.4-mini"]
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "code": "ai_model_missing", "error": "GPT-5.4 mini no esta disponible."})
		return
	}
	fechaUso := time.Now().Format("2006-01-02")
	uso, err := dbpkg.GetEmpresaAIUsoDiario(dbEmp, empresaID, model.Provider, reportesIAModelReportUsageID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario de reportes IA", http.StatusInternalServerError)
		return
	}
	if uso.Consultas >= reportesIAReportDailyLimit {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok": false, "code": "reportes_ia_report_limit_reached",
			"error": "Se alcanzo el limite diario de 2 reportes IA para esta empresa.",
			"usage": reportesIAUsagePayload(uso.Consultas, reportesIAReportDailyLimit, 0, 0),
		})
		return
	}

	builder := newReportesAIBuilder(dbEmp, empresaID, desde, hasta)
	ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 60 * time.Second}}
	system := "Eres un asistente contable de PCS. Genera un reporte nuevo SOLO como JSON valido con keys title, message, report_spec. report_spec debe incluir source_dataset, columns opcional, filters opcional, group_by opcional, metrics opcional, formulas opcional, order_by opcional y limit. No escribas SQL, no propongas otra empresa, no inventes campos y usa solo el catalogo semantico entregado. Si falta informacion, devuelve report_spec vacio y explica que aclaracion necesitas.\n\n" + buildReportesIANuevoContext()
	raw, pt, ct, err := ctrl.generateResponseWithSystemPrompt(model, pregunta, sanitizeHistorial(historial, 4), system)
	if err != nil {
		http.Error(w, "No se pudo interpretar reporte IA: "+err.Error(), http.StatusBadGateway)
		return
	}
	choice := parseReportesIANuevoChoice(raw)
	if strings.TrimSpace(choice.ReportSpec.SourceDataset) == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"ok": false, "modo": "nuevo", "respuesta": reportesIAFirstNonBlank(choice.Message, "Necesito que indiques el periodo, la fuente o el criterio del reporte."),
			"needs_clarification": true,
		})
		return
	}
	if strings.TrimSpace(choice.ReportSpec.Title) == "" {
		choice.ReportSpec.Title = strings.TrimSpace(choice.Title)
	}
	ds, err := builder.buildPersonalizedDataset(choice.ReportSpec)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]interface{}{
			"ok": false, "modo": "nuevo", "respuesta": "No puedo generar ese reporte con los datos autorizados: " + err.Error(),
			"needs_clarification": true,
		})
		return
	}
	if _, err := dbpkg.RegisterEmpresaAIConsulta(dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID: empresaID, Provider: model.Provider, ModelID: reportesIAModelReportUsageID,
		Pregunta: pregunta, Respuesta: strings.TrimSpace(raw), PromptTokens: pt, CompletionTokens: ct, TotalTokens: pt + ct,
		FechaConsulta: time.Now().Format("2006-01-02 15:04:05"), PlanActual: strings.TrimSpace(uso.PlanActual),
		UsuarioCreador: adminEmailFromRequest(r), Estado: "activo", Observaciones: reportesIACustomAuditObservations(choice.ReportSpec),
	}); err != nil {
		http.Error(w, "No se pudo registrar uso IA", http.StatusInternalServerError)
		return
	}
	usoNuevo, _ := dbpkg.GetEmpresaAIUsoDiario(dbEmp, empresaID, model.Provider, reportesIAModelReportUsageID, fechaUso)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "modo": "nuevo", "respuesta": reportesIAFirstNonBlank(choice.Message, "Vista previa generada. Revisa filtros, columnas y totales antes de exportar."),
		"title": ds.Title, "report_spec": choice.ReportSpec, "preview": ds,
		"usage": reportesIAUsagePayload(usoNuevo.Consultas, reportesIAReportDailyLimit, pt, ct),
	})
}

type reportesIAChoice struct {
	Dataset string `json:"dataset"`
	Format  string `json:"format"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type reportesIANuevoChoice struct {
	Title      string                          `json:"title"`
	Message    string                          `json:"message"`
	ReportSpec empresaReportePersonalizadoSpec `json:"report_spec"`
}

func parseReportesIAChoice(raw string) reportesIAChoice {
	var c reportesIAChoice
	text := strings.TrimSpace(raw)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		text = text[start : end+1]
	}
	_ = json.Unmarshal([]byte(text), &c)
	return c
}

func parseReportesIANuevoChoice(raw string) reportesIANuevoChoice {
	var choice reportesIANuevoChoice
	text := strings.TrimSpace(raw)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start >= 0 && end > start {
		text = text[start : end+1]
	}
	_ = json.Unmarshal([]byte(text), &choice)
	return choice
}

func newReportesAIBuilder(dbEmp *sql.DB, empresaID int64, desde, hasta string) *reportesBuilder {
	return &reportesBuilder{db: dbEmp, empresaID: empresaID, desde: strings.TrimSpace(desde), hasta: strings.TrimSpace(hasta), maxRows: 120, itemsCache: make(map[int64][]dbpkg.CarritoCompraItem)}
}

func buildReportesAITextContext(builder *reportesBuilder) (string, error) {
	tablero, err := builder.getTableroResumen()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Empresa ID: %d\nRango: %s a %s\n", builder.empresaID, builder.desde, builder.hasta))
	raw, _ := json.Marshal(tablero)
	b.WriteString("TABLERO_JSON:\n")
	b.WriteString(truncateText(string(raw), 9000))
	b.WriteString("\n\nCATALOGO_DATASETS:\n")
	for _, it := range reportesCatalogo {
		b.WriteString("- " + it.Key + ": " + it.Title + " (" + strings.Join(it.Formats, ",") + ")\n")
	}
	return b.String(), nil
}

func buildReportesIAReportContext(dataset, format string) string {
	var b strings.Builder
	b.WriteString("Dataset solicitado por UI: " + strings.TrimSpace(dataset) + "\n")
	b.WriteString("Formato solicitado por UI: " + strings.TrimSpace(format) + "\n")
	b.WriteString("Formatos validos: json, csv, txt, xls, pdf.\n")
	b.WriteString("Catalogo:\n")
	for _, it := range reportesCatalogo {
		b.WriteString("- " + it.Key + " | " + it.Title + " | " + it.Description + " | formatos: " + strings.Join(it.Formats, ",") + "\n")
	}
	return b.String()
}

func buildReportesIANuevoContext() string {
	var b strings.Builder
	b.WriteString("Fuentes y campos autorizados para reportes nuevos:\n")
	keys := make([]string, 0, len(reportesPersonalizadosSources))
	for key := range reportesPersonalizadosSources {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		b.WriteString("- " + key + ": " + strings.Join(reportesPersonalizadosSources[key], ", ") + "\n")
	}
	b.WriteString("Operaciones metricas: count, sum, average, min, max. Operaciones formula: add, subtract, divide, percentage. Filtros: eq, ne, contains, gt, gte, lt, lte. Maximo 1000 filas.\n")
	return b.String()
}

func reportesIAFirstNonBlank(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

// reportesIACustomAuditObservations conserva una huella verificable de la
// especificacion ejecutada sin ampliar el esquema ni guardar datos de negocio
// duplicados en la auditoria.
func reportesIACustomAuditObservations(spec empresaReportePersonalizadoSpec) string {
	payload, err := json.Marshal(spec)
	if err != nil {
		return "reportes_ia_custom:" + strings.TrimSpace(spec.SourceDataset) + ":spec_sha256_error"
	}
	hash := sha256.Sum256(payload)
	return fmt.Sprintf("reportes_ia_custom:%s:spec_sha256:%x", strings.TrimSpace(spec.SourceDataset), hash)
}

func normalizeReportesIAFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "csv", "txt", "xls", "pdf", "json":
		return strings.ToLower(strings.TrimSpace(format))
	default:
		return "pdf"
	}
}

func buildReportesIAExportURL(empresaID int64, dataset, format, desde, hasta string) string {
	q := fmt.Sprintf("/api/empresa/reportes?action=export&empresa_id=%d&dataset=%s&format=%s", empresaID, urlQueryEscapeLite(dataset), urlQueryEscapeLite(format))
	if strings.TrimSpace(desde) != "" {
		q += "&desde=" + urlQueryEscapeLite(desde)
	}
	if strings.TrimSpace(hasta) != "" {
		q += "&hasta=" + urlQueryEscapeLite(hasta)
	}
	return q
}

func urlQueryEscapeLite(v string) string {
	r := strings.NewReplacer(" ", "%20", "#", "%23", "&", "%26", "?", "%3F", "=", "%3D", "/", "%2F", ":", "%3A")
	return r.Replace(strings.TrimSpace(v))
}

func reportesIAUsagePayload(used int64, limit int64, promptTokens, completionTokens int64) map[string]interface{} {
	remaining := limit - used
	if remaining < 0 {
		remaining = 0
	}
	return map[string]interface{}{
		"daily_used":        used,
		"daily_limit":       limit,
		"daily_remaining":   remaining,
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
	}
}
