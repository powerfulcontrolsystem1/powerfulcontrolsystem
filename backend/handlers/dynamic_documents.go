package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	htmltmpl "html/template"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	texttmpl "text/template"
	"time"

	"github.com/xuri/excelize/v2"
	dbpkg "github.com/you/pos-backend/db"
)

const dynamicDocumentModelID = "openai:gpt-5.4-mini"

type DynamicDocumentRequest struct {
	EmpresaID    int64                  `json:"empresa_id"`
	Title        string                 `json:"title"`
	Prompt       string                 `json:"prompt"`
	Content      string                 `json:"content"`
	InputFormat  string                 `json:"input_format"`
	TemplateName string                 `json:"template_name"`
	ModelID      string                 `json:"model_id"`
	Variables    map[string]interface{} `json:"variables"`
	Formats      []string               `json:"formats"`
	Metadata     map[string]interface{} `json:"metadata"`
}

type DynamicDocumentChatExportRequest struct {
	EmpresaID      int64                  `json:"empresa_id"`
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	InputFormat    string                 `json:"input_format"`
	Format         string                 `json:"format"`
	DocumentType   string                 `json:"document_type"`
	SourceModule   string                 `json:"source_module"`
	ConversationID string                 `json:"conversation_id"`
	MessageID      string                 `json:"message_id"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type DynamicDocumentEmailShareRequest struct {
	EmpresaID  int64  `json:"empresa_id"`
	DocumentID string `json:"document_id"`
	Format     string `json:"format"`
	ToEmail    string `json:"to_email"`
	Subject    string `json:"subject"`
	Message    string `json:"message"`
}

type dynamicDocumentRecord struct {
	ID           string                 `json:"id"`
	EmpresaID    int64                  `json:"empresa_id"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	InputFormat  string                 `json:"input_format"`
	TemplateName string                 `json:"template_name"`
	HTML         string                 `json:"html"`
	PlainText    string                 `json:"plain_text"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ModelID      string                 `json:"model_id,omitempty"`
	CreatedAt    string                 `json:"created_at"`
	CreatedBy    string                 `json:"created_by"`
}

type docTemplateData struct {
	Title          string
	EmpresaID      int64
	EmpresaNombre  string
	CreatedAt      string
	CreatedBy      string
	TemplateName   string
	DocumentNumber string
	ContentHTML    htmltmpl.HTML
	Variables      []docVariablePair
	Metadata       []docVariablePair
}

type docVariablePair struct {
	Key   string
	Value string
}

// DynamicDocumentGenerateHandler recibe contenido o prompt IA y crea el documento base temporal.
func DynamicDocumentGenerateHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if !voiceStreamRequireSession(w, r, dbSuper) {
			return
		}

		var payload DynamicDocumentRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2*1024*1024))
		if err := dec.Decode(&payload); err != nil {
			http.Error(w, "JSON invalido: "+err.Error(), http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if parsed, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && parsed > 0 {
				payload.EmpresaID = parsed
			}
		}
		payload.Title = firstNonEmptyString(payload.Title, "Documento generado con IA")
		payload.InputFormat = normalizeDynamicDocumentInputFormat(payload.InputFormat)
		payload.TemplateName = normalizeDynamicDocumentTemplateName(payload.TemplateName)
		payload.ModelID = firstNonEmptyString(payload.ModelID, dynamicDocumentModelID)
		if payload.Variables == nil {
			payload.Variables = map[string]interface{}{}
		}
		if payload.Metadata == nil {
			payload.Metadata = map[string]interface{}{}
		}
		empresaNombre := resolveDynamicDocumentEmpresaName(dbEmp, payload.EmpresaID)
		if empresaNombre != "" {
			payload.Metadata["empresa_nombre"] = empresaNombre
		}

		content := strings.TrimSpace(payload.Content)
		modelID := ""
		var promptTokens, completionTokens int64
		if content == "" {
			var err error
			content, modelID, promptTokens, completionTokens, err = generateDynamicDocumentAIContent(r.Context(), dbEmp, dbSuper, payload)
			if err != nil {
				writeJSON(w, http.StatusBadGateway, map[string]interface{}{
					"ok":    false,
					"code":  "ai_generation_failed",
					"error": "No se pudo generar el contenido con IA: " + err.Error(),
				})
				return
			}
		}
		if strings.TrimSpace(content) == "" {
			http.Error(w, "content o prompt es obligatorio", http.StatusBadRequest)
			return
		}
		content = redactDynamicDocumentSecrets(content)
		renderedContent, err := renderDynamicTextVariables(content, payload.Variables)
		if err != nil {
			http.Error(w, "No se pudieron aplicar variables al contenido: "+err.Error(), http.StatusBadRequest)
			return
		}

		contentHTML := buildDynamicDocumentContentHTML(renderedContent, payload.InputFormat)
		htmlDoc, err := renderDynamicDocumentHTML(docTemplateData{
			Title:          payload.Title,
			EmpresaID:      payload.EmpresaID,
			EmpresaNombre:  empresaNombre,
			CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
			CreatedBy:      adminEmailFromRequest(r),
			TemplateName:   payload.TemplateName,
			DocumentNumber: "DOC-" + time.Now().Format("20060102-150405"),
			ContentHTML:    htmltmpl.HTML(contentHTML),
			Variables:      sortedDocPairs(payload.Variables),
			Metadata:       sortedDocPairs(payload.Metadata),
		})
		if err != nil {
			http.Error(w, "No se pudo renderizar HTML: "+err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := newDynamicDocumentID()
		if err != nil {
			http.Error(w, "No se pudo generar id del documento", http.StatusInternalServerError)
			return
		}
		record := dynamicDocumentRecord{
			ID:           id,
			EmpresaID:    payload.EmpresaID,
			Title:        payload.Title,
			Content:      renderedContent,
			InputFormat:  payload.InputFormat,
			TemplateName: payload.TemplateName,
			HTML:         htmlDoc,
			PlainText:    htmlToPlainText(htmlDoc),
			Variables:    payload.Variables,
			Metadata:     payload.Metadata,
			ModelID:      firstNonEmptyString(modelID, payload.ModelID),
			CreatedAt:    time.Now().Format(time.RFC3339),
			CreatedBy:    adminEmailFromRequest(r),
		}
		if err := saveDynamicDocumentRecord(record); err != nil {
			http.Error(w, "No se pudo guardar documento temporal: "+err.Error(), http.StatusInternalServerError)
			return
		}

		formatLinks := map[string]string{}
		for _, format := range normalizeDynamicDocumentFormats(payload.Formats) {
			formatLinks[format] = "/download?id=" + id + "&type=" + format
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                true,
			"document_id":       id,
			"title":             record.Title,
			"model_id":          record.ModelID,
			"input_format":      record.InputFormat,
			"template_name":     record.TemplateName,
			"download_urls":     formatLinks,
			"html_preview_path": "/download?id=" + id + "&type=html",
			"preview_text":      truncateText(record.PlainText, 1400),
			"available_formats": normalizeDynamicDocumentFormats(payload.Formats),
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		})
	}
}

// DynamicDocumentDownloadHandler descarga el documento en el formato solicitado.
func DynamicDocumentDownloadHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if !voiceStreamRequireSession(w, r, dbSuper) {
			return
		}
		id := sanitizeDynamicDocumentID(r.URL.Query().Get("id"))
		format := normalizeDynamicDocumentFormat(r.URL.Query().Get("type"))
		if id == "" {
			http.Error(w, "id es obligatorio", http.StatusBadRequest)
			return
		}
		if format == "" {
			http.Error(w, "type invalido. Use pdf, docx, xlsx, html, txt o json", http.StatusBadRequest)
			return
		}
		record, err := loadDynamicDocumentRecord(id)
		if err != nil {
			http.Error(w, "Documento no encontrado o expirado", http.StatusNotFound)
			return
		}
		path, contentType, err := ensureDynamicDocumentFile(r.Context(), record, format)
		if err != nil {
			http.Error(w, "No se pudo generar archivo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		filename := dynamicDocumentDownloadFilename(record, format)
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
		http.ServeFile(w, r, path)
	}
}

// DynamicDocumentChatExportHandler convierte una respuesta/conversacion del chat IA en documento descargable.
func DynamicDocumentChatExportHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if !voiceStreamRequireSession(w, r, dbSuper) {
			return
		}

		var payload DynamicDocumentChatExportRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2*1024*1024))
		if err := dec.Decode(&payload); err != nil {
			http.Error(w, "JSON invalido: "+err.Error(), http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if parsed, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && parsed > 0 {
				payload.EmpresaID = parsed
			}
		}
		format := normalizeDynamicDocumentFormat(payload.Format)
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		if format == "" {
			http.Error(w, "format invalido. Use pdf, docx, xlsx, txt o json", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(payload.Content) == "" {
			http.Error(w, "content es obligatorio", http.StatusBadRequest)
			return
		}

		empresaNombre := resolveDynamicDocumentEmpresaName(dbEmp, payload.EmpresaID)
		docType := normalizeDynamicDocumentType(payload.DocumentType, payload.Title, payload.Content)
		title := firstNonEmptyString(payload.Title, dynamicDocumentTitleFromType(docType))
		sourceModule := normalizeDynamicDocumentSourceModule(payload.SourceModule)
		content := redactDynamicDocumentSecrets(payload.Content)
		inputFormat := normalizeDynamicDocumentInputFormat(payload.InputFormat)
		if inputFormat == "text" && strings.Contains(content, "|") {
			inputFormat = "markdown"
		}
		if payload.Metadata == nil {
			payload.Metadata = map[string]interface{}{}
		}
		payload.Metadata["origin"] = "chat_ia"
		payload.Metadata["source_module"] = sourceModule
		payload.Metadata["document_type"] = docType
		payload.Metadata["requested_format"] = format
		payload.Metadata["empresa_nombre"] = empresaNombre
		payload.Metadata["conversation_id"] = strings.TrimSpace(payload.ConversationID)
		payload.Metadata["message_id"] = strings.TrimSpace(payload.MessageID)
		payload.Metadata["download_filename"] = buildDynamicDocumentProfessionalBaseFilename(empresaNombre, docType, time.Now())

		record, err := buildDynamicDocumentRecordFromContent(DynamicDocumentRequest{
			EmpresaID:    payload.EmpresaID,
			Title:        title,
			Content:      content,
			InputFormat:  inputFormat,
			TemplateName: normalizeDynamicDocumentTemplateName(docType),
			ModelID:      dynamicDocumentModelID,
			Variables:    map[string]interface{}{},
			Metadata:     payload.Metadata,
		}, adminEmailFromRequest(r), empresaNombre)
		if err != nil {
			auditDynamicDocumentChatExport(dbEmp, r, payload.EmpresaID, format, "", "error", err.Error(), payload.Metadata)
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":       false,
				"code":     "document_generation_failed",
				"error":    "No se pudo preparar el documento. El chat sigue disponible; intenta otro formato o descarga TXT/JSON.",
				"fallback": "txt",
			})
			return
		}
		if err := saveDynamicDocumentRecord(record); err != nil {
			auditDynamicDocumentChatExport(dbEmp, r, payload.EmpresaID, format, "", "error", err.Error(), payload.Metadata)
			http.Error(w, "No se pudo guardar documento temporal: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if _, _, err := ensureDynamicDocumentFile(r.Context(), record, format); err != nil {
			auditDynamicDocumentChatExport(dbEmp, r, payload.EmpresaID, format, dynamicDocumentDownloadFilename(record, format), "error", err.Error(), payload.Metadata)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                true,
				"document_id":       record.ID,
				"title":             record.Title,
				"requested_format":  format,
				"download_url":      "/download?id=" + record.ID + "&type=txt",
				"fallback_format":   "txt",
				"conversion_failed": true,
				"warning":           "El conversor de " + strings.ToUpper(format) + " fallo. Se preparo una descarga TXT como respaldo.",
			})
			return
		}

		filename := dynamicDocumentDownloadFilename(record, format)
		auditDynamicDocumentChatExport(dbEmp, r, payload.EmpresaID, format, filename, "ok", "", payload.Metadata)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":           true,
			"document_id":  record.ID,
			"title":        record.Title,
			"format":       format,
			"filename":     filename,
			"download_url": "/download?id=" + record.ID + "&type=" + format,
			"fallback":     false,
		})
	}
}

// DynamicDocumentEmailShareHandler envia por correo un documento generado desde el chat IA.
func DynamicDocumentEmailShareHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if !voiceStreamRequireSession(w, r, dbSuper) {
			return
		}

		var payload DynamicDocumentEmailShareRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512*1024))
		if err := dec.Decode(&payload); err != nil {
			http.Error(w, "JSON invalido: "+err.Error(), http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if parsed, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && parsed > 0 {
				payload.EmpresaID = parsed
			}
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		payload.DocumentID = sanitizeDynamicDocumentID(payload.DocumentID)
		if payload.DocumentID == "" {
			http.Error(w, "document_id es obligatorio", http.StatusBadRequest)
			return
		}
		format := normalizeDynamicDocumentFormat(payload.Format)
		if format == "" {
			format = "pdf"
		}

		record, err := loadDynamicDocumentRecord(payload.DocumentID)
		if err != nil {
			http.Error(w, "Documento no encontrado o expirado", http.StatusNotFound)
			return
		}
		if record.EmpresaID != payload.EmpresaID {
			http.Error(w, "El documento no pertenece a la empresa activa", http.StatusForbidden)
			return
		}
		path, contentType, err := ensureDynamicDocumentFile(r.Context(), record, format)
		if err != nil {
			http.Error(w, "No se pudo generar archivo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		content, err := os.ReadFile(path)
		if err != nil {
			http.Error(w, "No se pudo leer el archivo generado", http.StatusInternalServerError)
			return
		}
		filename := dynamicDocumentDownloadFilename(record, format)
		subject := strings.TrimSpace(payload.Subject)
		if subject == "" {
			subject = "Documento generado desde chat IA: " + strings.TrimSpace(record.Title)
		}
		message := strings.TrimSpace(payload.Message)
		if message == "" {
			message = "Adjunto encontraras el documento generado desde el chat IA."
		}
		metaJSON := fmt.Sprintf(`{"scope":"chat_documentos_email","empresa_id":%d,"document_id":%q,"format":%q}`, payload.EmpresaID, record.ID, format)
		if err := sendReportesEmailWithAttachment(r, dbSuper, payload.EmpresaID, payload.ToEmail, subject, message, filename, contentType, content, metaJSON); err != nil {
			http.Error(w, "no se pudo enviar el correo: "+err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":       true,
			"to_email": strings.TrimSpace(payload.ToEmail),
			"filename": filename,
			"format":   format,
		})
	}
}

func generateDynamicDocumentAIContent(ctx context.Context, dbEmp, dbSuper *sql.DB, payload DynamicDocumentRequest) (string, string, int64, int64, error) {
	if !isSuperAIEnabled(dbSuper) {
		return "", "", 0, 0, fmt.Errorf("la IA esta desactivada desde super administrador")
	}
	model, ok := availableEmpresaAIModelMap(dbSuper)[dynamicDocumentModelID]
	if !ok {
		return "", "", 0, 0, fmt.Errorf("GPT-5.4 mini no esta disponible o el proveedor esta desactivado")
	}
	var varsJSON string
	if len(payload.Variables) > 0 {
		raw, _ := json.MarshalIndent(payload.Variables, "", "  ")
		varsJSON = string(raw)
	}
	system := "Eres un generador profesional de documentos empresariales. Devuelve solo el contenido del documento en Markdown limpio, sin ``` ni explicaciones externas. Usa titulos, listas y tablas cuando aporten claridad. Respeta variables y datos suministrados; no inventes datos criticos como identificaciones, valores legales o firmas."
	prompt := strings.TrimSpace(payload.Prompt)
	if prompt == "" {
		prompt = "Genera un documento empresarial profesional."
	}
	if varsJSON != "" {
		prompt += "\n\nVariables disponibles en JSON:\n" + varsJSON
	}
	ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 70 * time.Second}}
	done := make(chan struct{})
	var content string
	var pt, ct int64
	var err error
	go func() {
		defer close(done)
		content, pt, ct, err = ctrl.generateResponseWithSystemPrompt(model, prompt, nil, system)
	}()
	select {
	case <-ctx.Done():
		return "", model.ID, 0, 0, ctx.Err()
	case <-done:
		return strings.TrimSpace(content), model.ID, pt, ct, err
	}
}

func buildDynamicDocumentRecordFromContent(payload DynamicDocumentRequest, createdBy, empresaNombre string) (dynamicDocumentRecord, error) {
	payload.Title = firstNonEmptyString(payload.Title, "Documento generado con IA")
	payload.InputFormat = normalizeDynamicDocumentInputFormat(payload.InputFormat)
	payload.TemplateName = normalizeDynamicDocumentTemplateName(payload.TemplateName)
	payload.ModelID = firstNonEmptyString(payload.ModelID, dynamicDocumentModelID)
	if payload.Variables == nil {
		payload.Variables = map[string]interface{}{}
	}
	if payload.Metadata == nil {
		payload.Metadata = map[string]interface{}{}
	}
	if empresaNombre != "" {
		payload.Metadata["empresa_nombre"] = empresaNombre
	}
	content := redactDynamicDocumentSecrets(payload.Content)
	if strings.TrimSpace(content) == "" {
		return dynamicDocumentRecord{}, fmt.Errorf("content es obligatorio")
	}
	renderedContent, err := renderDynamicTextVariables(content, payload.Variables)
	if err != nil {
		return dynamicDocumentRecord{}, fmt.Errorf("variables invalidas: %w", err)
	}
	contentHTML := buildDynamicDocumentContentHTML(renderedContent, payload.InputFormat)
	now := time.Now()
	htmlDoc, err := renderDynamicDocumentHTML(docTemplateData{
		Title:          payload.Title,
		EmpresaID:      payload.EmpresaID,
		EmpresaNombre:  empresaNombre,
		CreatedAt:      now.Format("2006-01-02 15:04:05"),
		CreatedBy:      createdBy,
		TemplateName:   payload.TemplateName,
		DocumentNumber: "DOC-" + now.Format("20060102-150405"),
		ContentHTML:    htmltmpl.HTML(contentHTML),
		Variables:      sortedDocPairs(payload.Variables),
		Metadata:       sortedDocPairs(payload.Metadata),
	})
	if err != nil {
		return dynamicDocumentRecord{}, err
	}
	id, err := newDynamicDocumentID()
	if err != nil {
		return dynamicDocumentRecord{}, err
	}
	return dynamicDocumentRecord{
		ID:           id,
		EmpresaID:    payload.EmpresaID,
		Title:        payload.Title,
		Content:      renderedContent,
		InputFormat:  payload.InputFormat,
		TemplateName: payload.TemplateName,
		HTML:         htmlDoc,
		PlainText:    htmlToPlainText(htmlDoc),
		Variables:    payload.Variables,
		Metadata:     payload.Metadata,
		ModelID:      payload.ModelID,
		CreatedAt:    now.Format(time.RFC3339),
		CreatedBy:    createdBy,
	}, nil
}

func renderDynamicTextVariables(content string, vars map[string]interface{}) (string, error) {
	tpl, err := texttmpl.New("document-content").Option("missingkey=zero").Parse(content)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, vars); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

func renderDynamicDocumentHTML(data docTemplateData) (string, error) {
	tpl, err := htmltmpl.New("dynamic-document").Parse(dynamicDocumentHTMLTemplate)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

const dynamicDocumentHTMLTemplate = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<title>{{.Title}}</title>
<style>
  @page { size: A4; margin: 18mm; }
  body { font-family: Arial, Helvetica, sans-serif; color: #182033; line-height: 1.48; margin: 0; background: #fff; }
  .doc-shell { max-width: 900px; margin: 0 auto; }
  .doc-header { border-bottom: 3px solid #1f8ef1; padding-bottom: 14px; margin-bottom: 22px; }
  .doc-kicker { color: #1f8ef1; font-size: 12px; font-weight: 800; letter-spacing: .08em; text-transform: uppercase; margin: 0 0 6px; }
  h1 { font-size: 28px; margin: 0; line-height: 1.16; }
  h2 { font-size: 20px; margin: 22px 0 8px; color: #12345f; }
  h3 { font-size: 16px; margin: 18px 0 8px; color: #20304a; }
  p { margin: 0 0 10px; }
  ul, ol { margin: 8px 0 12px 22px; padding: 0; }
  table { width: 100%; border-collapse: collapse; margin: 14px 0; font-size: 13px; }
  th { background: #e8f2ff; color: #12345f; text-align: left; }
  th, td { border: 1px solid #c9d8ea; padding: 8px 9px; vertical-align: top; }
  .doc-meta { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px 16px; margin-top: 12px; color: #5a6b84; font-size: 12px; }
  .doc-vars { margin-top: 26px; border-top: 1px solid #d9e2ee; padding-top: 12px; font-size: 12px; color: #5a6b84; }
  .doc-vars strong { color: #20304a; }
  .doc-footer { margin-top: 28px; padding-top: 12px; border-top: 1px solid #d9e2ee; color: #6b7b90; font-size: 11px; }
</style>
</head>
<body>
<main class="doc-shell">
  <header class="doc-header">
    <p class="doc-kicker">Documento dinamico generado con IA</p>
    <h1>{{.Title}}</h1>
    <div class="doc-meta">
      <span><strong>Empresa:</strong> {{if .EmpresaNombre}}{{.EmpresaNombre}}{{else}}{{.EmpresaID}}{{end}}</span>
      <span><strong>Fecha:</strong> {{.CreatedAt}}</span>
      <span><strong>Numero:</strong> {{.DocumentNumber}}</span>
      <span><strong>Plantilla:</strong> {{.TemplateName}}</span>
      <span><strong>Usuario:</strong> {{.CreatedBy}}</span>
    </div>
  </header>
  <section class="doc-content">
    {{.ContentHTML}}
  </section>
  {{if .Variables}}
  <section class="doc-vars">
    <strong>Variables usadas</strong>
    <table><tbody>{{range .Variables}}<tr><th>{{.Key}}</th><td>{{.Value}}</td></tr>{{end}}</tbody></table>
  </section>
  {{end}}
  <footer class="doc-footer">Powerful Control System - generador temporal de documentos.</footer>
</main>
</body>
</html>`

func buildDynamicDocumentContentHTML(content, inputFormat string) string {
	switch normalizeDynamicDocumentInputFormat(inputFormat) {
	case "html":
		return sanitizeDynamicDocumentHTML(content)
	default:
		return markdownLikeToHTML(content)
	}
}

func markdownLikeToHTML(content string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var b strings.Builder
	inList := false
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
			continue
		}
		if isMarkdownTableLine(line) && i+1 < len(lines) && isMarkdownSeparatorLine(strings.TrimSpace(lines[i+1])) {
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
			table, next := consumeMarkdownTable(lines, i)
			b.WriteString(table)
			i = next - 1
			continue
		}
		if strings.HasPrefix(line, "### ") {
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
			b.WriteString("<h3>" + htmltmpl.HTMLEscapeString(strings.TrimSpace(line[4:])) + "</h3>")
			continue
		}
		if strings.HasPrefix(line, "## ") {
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
			b.WriteString("<h2>" + htmltmpl.HTMLEscapeString(strings.TrimSpace(line[3:])) + "</h2>")
			continue
		}
		if strings.HasPrefix(line, "# ") {
			if inList {
				b.WriteString("</ul>")
				inList = false
			}
			b.WriteString("<h2>" + htmltmpl.HTMLEscapeString(strings.TrimSpace(line[2:])) + "</h2>")
			continue
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			if !inList {
				b.WriteString("<ul>")
				inList = true
			}
			b.WriteString("<li>" + inlineMarkdownToHTML(strings.TrimSpace(line[2:])) + "</li>")
			continue
		}
		if inList {
			b.WriteString("</ul>")
			inList = false
		}
		b.WriteString("<p>" + inlineMarkdownToHTML(line) + "</p>")
	}
	if inList {
		b.WriteString("</ul>")
	}
	return b.String()
}

func inlineMarkdownToHTML(raw string) string {
	escaped := htmltmpl.HTMLEscapeString(raw)
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	return re.ReplaceAllString(escaped, "<strong>$1</strong>")
}

func isMarkdownTableLine(line string) bool {
	return strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") && strings.Count(line, "|") >= 2
}

func isMarkdownSeparatorLine(line string) bool {
	if !isMarkdownTableLine(line) {
		return false
	}
	clean := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(line, "|", ""), "-", ""), ":", "")
	return strings.TrimSpace(clean) == ""
}

func consumeMarkdownTable(lines []string, start int) (string, int) {
	header := splitMarkdownTableRow(lines[start])
	i := start + 2
	rows := [][]string{}
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if !isMarkdownTableLine(line) {
			break
		}
		rows = append(rows, splitMarkdownTableRow(line))
		i++
	}
	var b strings.Builder
	b.WriteString("<table><thead><tr>")
	for _, cell := range header {
		b.WriteString("<th>" + inlineMarkdownToHTML(cell) + "</th>")
	}
	b.WriteString("</tr></thead><tbody>")
	for _, row := range rows {
		b.WriteString("<tr>")
		for _, cell := range row {
			b.WriteString("<td>" + inlineMarkdownToHTML(cell) + "</td>")
		}
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table>")
	return b.String(), i
}

func splitMarkdownTableRow(line string) []string {
	line = strings.Trim(strings.TrimSpace(line), "|")
	parts := strings.Split(line, "|")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, strings.TrimSpace(part))
	}
	return out
}

func sanitizeDynamicDocumentHTML(raw string) string {
	reScript := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	reStyle := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	reEvents := regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	reJavascriptURLs := regexp.MustCompile(`(?i)\s+(href|src)\s*=\s*("javascript:[^"]*"|'javascript:[^']*'|javascript:[^\s>]+)`)
	out := reScript.ReplaceAllString(raw, "")
	out = reStyle.ReplaceAllString(out, "")
	out = reEvents.ReplaceAllString(out, "")
	out = reJavascriptURLs.ReplaceAllString(out, "")
	return strings.TrimSpace(out)
}

func htmlToPlainText(raw string) string {
	reBreaks := regexp.MustCompile(`(?i)</(p|h1|h2|h3|li|tr)>`)
	reTags := regexp.MustCompile(`<[^>]+>`)
	text := reBreaks.ReplaceAllString(raw, "\n")
	text = reTags.ReplaceAllString(text, " ")
	reSpaces := regexp.MustCompile(`[ \t]+`)
	text = reSpaces.ReplaceAllString(text, " ")
	reLines := regexp.MustCompile(`\n{3,}`)
	return strings.TrimSpace(reLines.ReplaceAllString(text, "\n\n"))
}

func ensureDynamicDocumentFile(ctx context.Context, record dynamicDocumentRecord, format string) (string, string, error) {
	dir, err := ensureDynamicDocumentDir()
	if err != nil {
		return "", "", err
	}
	path := filepath.Join(dir, record.ID+"."+format)
	if _, err := os.Stat(path); err == nil {
		return path, dynamicDocumentContentType(format), nil
	}
	switch format {
	case "html":
		err = os.WriteFile(path, []byte(record.HTML), 0600)
	case "txt":
		err = os.WriteFile(path, []byte(record.PlainText), 0600)
	case "json":
		var raw []byte
		raw, err = json.MarshalIndent(record, "", "  ")
		if err == nil {
			err = os.WriteFile(path, raw, 0600)
		}
	case "xlsx":
		err = writeDynamicDocumentXLSX(record, path)
	case "docx":
		err = writeDynamicDocumentDOCX(record, path)
	case "pdf":
		err = writeDynamicDocumentPDF(ctx, record, path)
	default:
		err = fmt.Errorf("formato no soportado")
	}
	if err != nil {
		return "", "", err
	}
	return path, dynamicDocumentContentType(format), nil
}

func writeDynamicDocumentXLSX(record dynamicDocumentRecord, path string) error {
	f := excelize.NewFile()
	sheet := "Documento"
	f.SetSheetName("Sheet1", sheet)
	rows := extractFirstMarkdownTable(record.Content)
	if len(rows) > 0 {
		for r, row := range rows {
			for c, cell := range row {
				axis, _ := excelize.CoordinatesToCellName(c+1, r+1)
				_ = f.SetCellValue(sheet, axis, cell)
			}
		}
	} else {
		_ = f.SetCellValue(sheet, "A1", "Linea")
		_ = f.SetCellValue(sheet, "B1", "Contenido")
		lines := strings.Split(record.PlainText, "\n")
		row := 2
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			_ = f.SetCellValue(sheet, fmt.Sprintf("A%d", row), row-1)
			_ = f.SetCellValue(sheet, fmt.Sprintf("B%d", row), line)
			row++
		}
	}
	if len(record.Variables) > 0 {
		varSheet := "Variables"
		_, _ = f.NewSheet(varSheet)
		_ = f.SetCellValue(varSheet, "A1", "Variable")
		_ = f.SetCellValue(varSheet, "B1", "Valor")
		for i, pair := range sortedDocPairs(record.Variables) {
			_ = f.SetCellValue(varSheet, fmt.Sprintf("A%d", i+2), pair.Key)
			_ = f.SetCellValue(varSheet, fmt.Sprintf("B%d", i+2), pair.Value)
		}
	}
	_ = f.SetColWidth(sheet, "A", "A", 14)
	_ = f.SetColWidth(sheet, "B", "Z", 26)
	return f.SaveAs(path)
}

func extractFirstMarkdownTable(content string) [][]string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	for i := 0; i+1 < len(lines); i++ {
		if isMarkdownTableLine(strings.TrimSpace(lines[i])) && isMarkdownSeparatorLine(strings.TrimSpace(lines[i+1])) {
			rows := [][]string{splitMarkdownTableRow(lines[i])}
			i += 2
			for i < len(lines) && isMarkdownTableLine(strings.TrimSpace(lines[i])) {
				rows = append(rows, splitMarkdownTableRow(lines[i]))
				i++
			}
			return rows
		}
	}
	return nil
}

func writeDynamicDocumentDOCX(record dynamicDocumentRecord, path string) error {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := map[string]string{
		"[Content_Types].xml":          docxContentTypesXML,
		"_rels/.rels":                  docxRelsXML,
		"docProps/core.xml":            buildDOCXCoreXML(record),
		"docProps/app.xml":             docxAppXML,
		"word/_rels/document.xml.rels": docxDocumentRelsXML,
		"word/document.xml":            buildDOCXDocumentXML(record),
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, files[name]); err != nil {
			return err
		}
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0600)
}

func buildDOCXDocumentXML(record dynamicDocumentRecord) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	b.WriteString(docxParagraph(record.Title, true))
	for _, line := range strings.Split(record.PlainText, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		b.WriteString(docxParagraph(line, false))
	}
	b.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func docxParagraph(text string, title bool) string {
	var escaped bytes.Buffer
	_ = xml.EscapeText(&escaped, []byte(text))
	if title {
		return `<w:p><w:pPr><w:pStyle w:val="Title"/></w:pPr><w:r><w:rPr><w:b/><w:sz w:val="32"/></w:rPr><w:t>` + escaped.String() + `</w:t></w:r></w:p>`
	}
	return `<w:p><w:r><w:t xml:space="preserve">` + escaped.String() + `</w:t></w:r></w:p>`
}

func buildDOCXCoreXML(record dynamicDocumentRecord) string {
	title := xmlEscapeString(record.Title)
	created := time.Now().UTC().Format(time.RFC3339)
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/" xmlns:dcmitype="http://purl.org/dc/dcmitype/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"><dc:title>` + title + `</dc:title><dc:creator>Powerful Control System</dc:creator><cp:lastModifiedBy>Powerful Control System</cp:lastModifiedBy><dcterms:created xsi:type="dcterms:W3CDTF">` + created + `</dcterms:created><dcterms:modified xsi:type="dcterms:W3CDTF">` + created + `</dcterms:modified></cp:coreProperties>`
}

const docxContentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/><Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/><Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/></Types>`
const docxRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/><Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/></Relationships>`
const docxDocumentRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`
const docxAppXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?><Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties" xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes"><Application>Powerful Control System</Application></Properties>`

func writeDynamicDocumentPDF(ctx context.Context, record dynamicDocumentRecord, path string) error {
	dir, err := ensureDynamicDocumentDir()
	if err != nil {
		return err
	}
	htmlPath := filepath.Join(dir, record.ID+".html")
	if err := os.WriteFile(htmlPath, []byte(record.HTML), 0600); err != nil {
		return err
	}
	wkhtml := strings.TrimSpace(os.Getenv("WKHTMLTOPDF_PATH"))
	if wkhtml == "" {
		wkhtml, _ = exec.LookPath("wkhtmltopdf")
	}
	if wkhtml != "" {
		cctx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		cmd := exec.CommandContext(cctx, wkhtml, "--quiet", "--encoding", "utf-8", htmlPath, path)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}
	return writeBasicPDF(record.Title, record.PlainText, path)
}

func writeBasicPDF(title, text, path string) error {
	lines := wrapPDFLines(title+"\n\n"+text, 92)
	var stream strings.Builder
	stream.WriteString("BT\n/F1 11 Tf\n50 790 Td\n14 TL\n")
	for i, line := range lines {
		if i > 0 && i%52 == 0 {
			stream.WriteString("ET\nBT\n/F1 11 Tf\n50 790 Td\n14 TL\n")
		}
		stream.WriteString("(" + escapePDFString(line) + ") Tj\nT*\n")
	}
	stream.WriteString("ET")
	content := stream.String()
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	writeObj := func(id int, body string) {
		offsets = append(offsets, out.Len())
		out.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", id, body))
	}
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>")
	writeObj(4, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")
	writeObj(5, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content))
	xref := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(offsets)))
	for i := 1; i < len(offsets); i++ {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(offsets), xref))
	return os.WriteFile(path, out.Bytes(), 0600)
}

func wrapPDFLines(text string, max int) []string {
	words := strings.Fields(strings.ReplaceAll(text, "\n", " \n "))
	lines := []string{}
	current := ""
	for _, word := range words {
		if word == "\n" {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, "")
			continue
		}
		if len([]rune(current+" "+word)) > max {
			lines = append(lines, current)
			current = word
			continue
		}
		if current == "" {
			current = word
		} else {
			current += " " + word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func escapePDFString(s string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "(", "\\(", ")", "\\)")
	return replacer.Replace(s)
}

func saveDynamicDocumentRecord(record dynamicDocumentRecord) error {
	dir, err := ensureDynamicDocumentDir()
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, record.ID+".record.json"), raw, 0600)
}

func loadDynamicDocumentRecord(id string) (dynamicDocumentRecord, error) {
	var record dynamicDocumentRecord
	dir, err := ensureDynamicDocumentDir()
	if err != nil {
		return record, err
	}
	raw, err := os.ReadFile(filepath.Join(dir, id+".record.json"))
	if err != nil {
		return record, err
	}
	if err := json.Unmarshal(raw, &record); err != nil {
		return record, err
	}
	return record, nil
}

func ensureDynamicDocumentDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "powerfulcontrolsystem-documents")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func newDynamicDocumentID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func sanitizeDynamicDocumentID(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if matched, _ := regexp.MatchString(`^[a-f0-9]{32}$`, value); matched {
		return value
	}
	return ""
}

func normalizeDynamicDocumentInputFormat(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "html":
		return "html"
	case "md", "markdown":
		return "markdown"
	default:
		return "text"
	}
}

func normalizeDynamicDocumentTemplateName(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "factura", "contrato", "reporte", "acta", "cotizacion", "tabla", "documento":
		return value
	case "profesional":
		return value
	default:
		return "profesional"
	}
}

func normalizeDynamicDocumentType(raw, title, content string) string {
	text := strings.ToLower(strings.TrimSpace(raw + " " + title + " " + content))
	switch {
	case strings.Contains(text, "factura"):
		return "factura"
	case strings.Contains(text, "contrato"):
		return "contrato"
	case strings.Contains(text, "cotizacion") || strings.Contains(text, "cotización"):
		return "cotizacion"
	case strings.Contains(text, "acta"):
		return "acta"
	case strings.Contains(text, "reporte") || strings.Contains(text, "informe"):
		return "reporte"
	case strings.Contains(text, "excel") || strings.Contains(text, "tabla") || strings.Contains(text, "|"):
		return "tabla"
	default:
		return "documento"
	}
}

func dynamicDocumentTitleFromType(docType string) string {
	switch normalizeDynamicDocumentType(docType, "", "") {
	case "factura":
		return "Factura generada desde chat IA"
	case "contrato":
		return "Contrato generado desde chat IA"
	case "cotizacion":
		return "Cotizacion generada desde chat IA"
	case "acta":
		return "Acta generada desde chat IA"
	case "reporte":
		return "Reporte generado desde chat IA"
	case "tabla":
		return "Tabla generada desde chat IA"
	default:
		return "Documento generado desde chat IA"
	}
}

func normalizeDynamicDocumentSourceModule(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	value = strings.ReplaceAll(value, " ", "_")
	value = regexp.MustCompile(`[^a-z0-9_-]+`).ReplaceAllString(value, "")
	if value == "" {
		return "chat_ia"
	}
	if len(value) > 80 {
		value = value[:80]
	}
	return value
}

func normalizeDynamicDocumentFormat(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "pdf", "docx", "xlsx", "html", "txt", "json":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func normalizeDynamicDocumentFormats(raw []string) []string {
	if len(raw) == 0 {
		return []string{"pdf", "docx", "xlsx", "txt", "json"}
	}
	seen := map[string]bool{}
	out := []string{}
	for _, item := range raw {
		format := normalizeDynamicDocumentFormat(item)
		if format == "" || seen[format] {
			continue
		}
		seen[format] = true
		out = append(out, format)
	}
	if len(out) == 0 {
		return []string{"pdf", "docx", "xlsx", "txt", "json"}
	}
	return out
}

func dynamicDocumentContentType(format string) string {
	switch format {
	case "pdf":
		return "application/pdf"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "html":
		return "text/html; charset=utf-8"
	case "json":
		return "application/json; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}

func sanitizeDownloadFilename(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		value = "documento"
	}
	value = regexp.MustCompile(`[^a-z0-9._-]+`).ReplaceAllString(value, "_")
	value = strings.Trim(value, "._-")
	if value == "" {
		value = "documento"
	}
	if len(value) > 80 {
		value = value[:80]
	}
	return value
}

func dynamicDocumentDownloadFilename(record dynamicDocumentRecord, format string) string {
	base := ""
	if record.Metadata != nil {
		if value, ok := record.Metadata["download_filename"]; ok && value != nil {
			base = strings.TrimSpace(fmt.Sprint(value))
		}
	}
	if base == "" {
		base = record.Title
	}
	return sanitizeDownloadFilename(base) + "." + normalizeDynamicDocumentFormat(format)
}

func buildDynamicDocumentProfessionalBaseFilename(empresaNombre, docType string, now time.Time) string {
	empresa := sanitizeDownloadFilename(firstNonEmptyString(empresaNombre, "empresa"))
	tipo := sanitizeDownloadFilename(firstNonEmptyString(normalizeDynamicDocumentType(docType, "", ""), "documento"))
	fecha := now.Format("2006-01-02")
	return strings.Trim(empresa+"_"+tipo+"_"+fecha, "_")
}

func resolveDynamicDocumentEmpresaName(dbEmp *sql.DB, empresaID int64) string {
	if dbEmp == nil || empresaID <= 0 {
		return ""
	}
	empresa, err := dbpkg.GetEmpresaByScopeID(dbEmp, empresaID)
	if err != nil || empresa == nil {
		return ""
	}
	return strings.TrimSpace(empresa.Nombre)
}

func redactDynamicDocumentSecrets(content string) string {
	out := strings.TrimSpace(content)
	if out == "" {
		return ""
	}
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|contraseña|contrasena|authorization|bearer|private[_-]?key|p[_-]?key)\s*[:=]\s*["']?[^"'\s,;]+["']?`),
		regexp.MustCompile(`(?i)("?(api[_-]?key|secret|token|password|contraseña|contrasena|authorization|private[_-]?key|p[_-]?key)"?\s*:\s*)"[^"]+"`),
	}
	for _, re := range patterns {
		out = re.ReplaceAllStringFunc(out, func(match string) string {
			if strings.Contains(match, ":") {
				parts := strings.SplitN(match, ":", 2)
				return strings.TrimSpace(parts[0]) + ": [REDACTADO]"
			}
			if strings.Contains(match, "=") {
				parts := strings.SplitN(match, "=", 2)
				return strings.TrimSpace(parts[0]) + "=[REDACTADO]"
			}
			return "[REDACTADO]"
		})
	}
	return out
}

func auditDynamicDocumentChatExport(dbEmp *sql.DB, r *http.Request, empresaID int64, format, filename, resultado, detail string, metadata map[string]interface{}) {
	if dbEmp == nil || empresaID <= 0 {
		return
	}
	meta := map[string]interface{}{
		"origin":        "chat_ia",
		"format":        strings.ToLower(strings.TrimSpace(format)),
		"filename":      strings.TrimSpace(filename),
		"source_module": "chat_ia",
	}
	for key, value := range metadata {
		switch strings.ToLower(strings.TrimSpace(key)) {
		case "origin", "source_module", "document_type", "conversation_id", "message_id", "requested_format":
			meta[key] = value
		}
	}
	if strings.TrimSpace(detail) != "" {
		meta["detail"] = detail
	}
	raw, _ := json.Marshal(meta)
	status := int64(http.StatusOK)
	if strings.ToLower(strings.TrimSpace(resultado)) == "error" {
		status = int64(http.StatusInternalServerError)
	}
	_, _ = dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      empresaID,
		Modulo:         "documentos_ia",
		Accion:         "exportar_chat",
		Recurso:        "dynamic_documents",
		MetodoHTTP:     http.MethodPost,
		Endpoint:       "/api/empresa/chat_documentos/exportar",
		Resultado:      resultado,
		CodigoHTTP:     status,
		RequestID:      resolveAuditoriaRequestID(r),
		IPOrigen:       resolveAuditoriaIP(r),
		UserAgent:      strings.TrimSpace(r.UserAgent()),
		MetadataJSON:   string(raw),
		RetencionDias:  normalizeRetencionDiasForHandler(0),
		UsuarioCreador: strings.TrimSpace(adminEmailFromRequest(r)),
		Estado:         "activo",
		Observaciones:  "exportacion de documento generado desde chat IA",
	})
}

func sortedDocPairs(values map[string]interface{}) []docVariablePair {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]docVariablePair, 0, len(keys))
	for _, key := range keys {
		out = append(out, docVariablePair{Key: key, Value: fmt.Sprint(values[key])})
	}
	return out
}

func xmlEscapeString(value string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(value))
	return b.String()
}
