package handlers

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type openAIStreamEvent struct {
	Delta string `json:"delta"`
	Done  bool   `json:"done,omitempty"`
	Error string `json:"error,omitempty"`
}

func sseWriteJSON(w http.ResponseWriter, payload interface{}) error {
	b, _ := json.Marshal(payload)
	_, err := w.Write([]byte("data: " + string(b) + "\n\n"))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	return err
}

func (c *EmpresaAIChatController) callOpenAIStreamChatCompletions(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, systemPrompt string, onDelta func(string)) (string, error) {
	apiKey, err := c.resolveModelAPIKey(model)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(apiKey) == "" {
		return "", fmt.Errorf("OPENAI_API_KEY no configurada")
	}

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	messages := make([]msg, 0, 12)
	messages = append(messages, msg{Role: "system", Content: systemPrompt})
	clean := sanitizeHistorial(historial, 8)
	for _, h := range clean {
		role := strings.ToLower(strings.TrimSpace(h.Rol))
		if role != "user" && role != "assistant" {
			continue
		}
		messages = append(messages, msg{Role: role, Content: strings.TrimSpace(h.Contenido)})
	}
	messages = append(messages, msg{Role: "user", Content: strings.TrimSpace(pregunta)})

	body := map[string]interface{}{
		"model":                 strings.TrimSpace(model.UpstreamModel),
		"messages":              messages,
		"temperature":           0.2,
		"max_completion_tokens": 700,
		"stream":                true,
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("no se pudo crear solicitud al proveedor")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("no se pudo contactar proveedor: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		return "", &aiProviderHTTPError{Provider: "openai", Status: resp.StatusCode, Body: truncateText(string(raw), 600)}
	}

	var out strings.Builder
	sc := bufio.NewScanner(resp.Body)
	// Permitir chunks grandes.
	sc.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			break
		}
		var parsed struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &parsed); err != nil {
			continue
		}
		if len(parsed.Choices) == 0 {
			continue
		}
		d := parsed.Choices[0].Delta.Content
		if strings.TrimSpace(d) == "" {
			continue
		}
		out.WriteString(d)
		if onDelta != nil {
			onDelta(d)
		}
	}
	if err := sc.Err(); err != nil {
		return "", fmt.Errorf("stream interrumpido: %v", err)
	}
	return out.String(), nil
}

type EmpresaAIChatController struct {
	dbEmp   *sql.DB
	dbSuper *sql.DB
	client  *http.Client
}

type aiProviderHTTPError struct {
	Provider string
	Status   int
	Body     string
}

func (e *aiProviderHTTPError) Error() string {
	body := strings.TrimSpace(e.Body)
	if body == "" {
		return fmt.Sprintf("error proveedor %s (%d)", strings.TrimSpace(e.Provider), e.Status)
	}
	return fmt.Sprintf("error proveedor %s (%d): %s", strings.TrimSpace(e.Provider), e.Status, body)
}

type empresaAIModelDef struct {
	ID             string `json:"id"`
	Provider       string `json:"provider"`
	DisplayName    string `json:"display_name"`
	UpstreamModel  string `json:"upstream_model"`
	Endpoint       string `json:"endpoint"`
	ApiKeyEnv      string `json:"-"`
	Famous         bool   `json:"famous"`
	FreeDailyLimit int    `json:"free_daily_limit"`
	Description    string `json:"description"`
	PlanURL        string `json:"plan_url"`
}

type empresaAIChatMensaje struct {
	Rol       string `json:"rol"`
	Contenido string `json:"contenido"`
}

type empresaAIChatRequest struct {
	EmpresaID   int64                  `json:"empresa_id"`
	ModelID     string                 `json:"model_id"`
	Pregunta    string                 `json:"pregunta"`
	Historial   []empresaAIChatMensaje `json:"historial"`
	Temperatura float64                `json:"temperatura"`
}

type aiAttachment struct {
	Filename string
	MimeType string
	Bytes    []byte
}

type empresaAIModeloPreferidoPayload struct {
	EmpresaID int64  `json:"empresa_id"`
	ModelID   string `json:"model_id"`
}

func NewEmpresaAIChatController(dbEmp, dbSuper *sql.DB) *EmpresaAIChatController {
	return &EmpresaAIChatController{
		dbEmp:   dbEmp,
		dbSuper: dbSuper,
		client:  &http.Client{Timeout: 45 * time.Second},
	}
}

func empresaAIModelCatalog() []empresaAIModelDef {
	return []empresaAIModelDef{
		{
			ID:             "openai:gpt-5.4-mini",
			Provider:       "openai",
			DisplayName:    "OpenAI GPT-5.4 mini",
			UpstreamModel:  "gpt-5.4-mini",
			Endpoint:       "https://api.openai.com/v1/chat/completions",
			ApiKeyEnv:      "OPENAI_API_KEY",
			Famous:         true,
			FreeDailyLimit: 120,
			Description:    "Chat empresarial con OpenAI, restringido por empresa_id y con API key cifrada en el panel super.",
		},
		{
			ID:             "openai:gpt-5.5",
			Provider:       "openai",
			DisplayName:    "OpenAI GPT-5.5 (documentos)",
			UpstreamModel:  "gpt-5.5",
			Endpoint:       "https://api.openai.com/v1/responses",
			ApiKeyEnv:      "OPENAI_API_KEY",
			Famous:         true,
			FreeDailyLimit: 20,
			Description:    "Procesamiento de documentos e imágenes con OpenAI. Se recomienda limitar su uso diario por empresa desde el panel super.",
		},
	}
}

func availableEmpresaAIModelCatalog(dbSuper *sql.DB) []empresaAIModelDef {
	catalog := empresaAIModelCatalog()
	available := make([]empresaAIModelDef, 0, len(catalog))
	for _, item := range catalog {
		if !isAIProviderEnabled(dbSuper, item.Provider) {
			continue
		}
		available = append(available, item)
	}
	return available
}

func empresaAIModelMap() map[string]empresaAIModelDef {
	catalog := empresaAIModelCatalog()
	m := make(map[string]empresaAIModelDef, len(catalog))
	for _, it := range catalog {
		it.PlanURL = "/pagar_licencia.html"
		m[it.ID] = it
	}
	return m
}

func availableEmpresaAIModelMap(dbSuper *sql.DB) map[string]empresaAIModelDef {
	catalog := availableEmpresaAIModelCatalog(dbSuper)
	m := make(map[string]empresaAIModelDef, len(catalog))
	for _, it := range catalog {
		it.PlanURL = "/pagar_licencia.html"
		m[it.ID] = it
	}
	return m
}

func firstAvailableEmpresaAIModelID(dbSuper *sql.DB) string {
	catalog := availableEmpresaAIModelCatalog(dbSuper)
	if len(catalog) == 0 {
		return ""
	}
	return catalog[0].ID
}

func (c *EmpresaAIChatController) ModelosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}

	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}

	googleAccount := googleAccountFromRequest(r)
	if googleAccount == "" {
		http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
		return
	}
	if err := c.ensureEmpresaAccessByAccount(googleAccount, empresaID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	modeloPreferido, err := dbpkg.GetEmpresaAIModeloPreferido(c.dbEmp, empresaID, googleAccount)
	if err != nil {
		http.Error(w, "No se pudo consultar el modelo preferido", http.StatusInternalServerError)
		return
	}

	catalog := availableEmpresaAIModelCatalog(c.dbSuper)
	if len(catalog) == 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_models_unavailable",
			"error":          "No hay proveedores IA habilitados para esta empresa.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}
	availableMap := availableEmpresaAIModelMap(c.dbSuper)
	if _, ok := availableMap[modeloPreferido]; !ok {
		modeloPreferido = firstAvailableEmpresaAIModelID(c.dbSuper)
	}

	items := make([]map[string]interface{}, 0, len(catalog))
	for _, it := range catalog {
		items = append(items, map[string]interface{}{
			"id":               it.ID,
			"provider":         it.Provider,
			"display_name":     it.DisplayName,
			"upstream_model":   it.UpstreamModel,
			"famous":           it.Famous,
			"free_daily_limit": it.FreeDailyLimit,
			"description":      it.Description,
			"plan_url":         it.PlanURL + "?empresa_id=" + fmt.Sprintf("%d", empresaID),
		})
	}

	streamingEnabled, _, _, _ := getChatIAEmpresaStreamingEnabled(c.dbSuper)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                true,
		"empresa_id":        empresaID,
		"google_account":    googleAccount,
		"modelo_preferido":  modeloPreferido,
		"streaming_enabled": streamingEnabled,
		"modelos":           items,
	})
}

func (c *EmpresaAIChatController) ModeloPreferidoHandler(w http.ResponseWriter, r *http.Request) {
	if !isSuperAIEnabled(c.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}
	switch r.Method {
	case http.MethodGet:
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		googleAccount := googleAccountFromRequest(r)
		if googleAccount == "" {
			http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
			return
		}
		if err := c.ensureEmpresaAccessByAccount(googleAccount, empresaID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		modelID, err := dbpkg.GetEmpresaAIModeloPreferido(c.dbEmp, empresaID, googleAccount)
		if err != nil {
			http.Error(w, "No se pudo consultar el modelo preferido", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":             true,
			"empresa_id":     empresaID,
			"google_account": googleAccount,
			"model_id":       modelID,
		})
		return

	case http.MethodPut:
		var payload empresaAIModeloPreferidoPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if payload.EmpresaID <= 0 {
			if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
				payload.EmpresaID = empresaID
			}
		}
		if payload.EmpresaID <= 0 {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}

		payload.ModelID = strings.TrimSpace(payload.ModelID)
		if payload.ModelID == "" {
			payload.ModelID = firstAvailableEmpresaAIModelID(c.dbSuper)
		}

		googleAccount := googleAccountFromRequest(r)
		if googleAccount == "" {
			http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
			return
		}
		if err := c.ensureEmpresaAccessByAccount(googleAccount, payload.EmpresaID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		catalog := availableEmpresaAIModelMap(c.dbSuper)
		if len(catalog) == 0 {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":             false,
				"code":           "ai_models_unavailable",
				"error":          "No hay proveedores IA habilitados para esta empresa.",
				"service_status": superAIServiceStatus(c.dbSuper),
			})
			return
		}
		if _, ok := catalog[payload.ModelID]; !ok {
			http.Error(w, "model_id no soportado o desactivado", http.StatusBadRequest)
			return
		}

		if err := dbpkg.UpsertEmpresaAIModeloPreferido(c.dbEmp, payload.EmpresaID, googleAccount, payload.ModelID, googleAccount); err != nil {
			http.Error(w, "No se pudo registrar el modelo preferido", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":             true,
			"empresa_id":     payload.EmpresaID,
			"google_account": googleAccount,
			"model_id":       payload.ModelID,
			"saved":          true,
		})
		return
	}

	http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
}

func (c *EmpresaAIChatController) ConsultarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}

	var payload empresaAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 {
		if empresaID, err := parseInt64QueryOptional(r, "empresa_id"); err == nil && empresaID > 0 {
			payload.EmpresaID = empresaID
		}
	}
	if payload.EmpresaID <= 0 {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}
	googleAccount := googleAccountFromRequest(r)
	if googleAccount == "" {
		http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
		return
	}
	if err := c.ensureEmpresaAccessByAccount(googleAccount, payload.EmpresaID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	payload.ModelID = strings.TrimSpace(payload.ModelID)
	if payload.ModelID == "" {
		payload.ModelID = firstAvailableEmpresaAIModelID(c.dbSuper)
	}
	catalog := availableEmpresaAIModelMap(c.dbSuper)
	if len(catalog) == 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_models_unavailable",
			"error":          "No hay proveedores IA habilitados para esta empresa.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}
	model, ok := catalog[payload.ModelID]
	if !ok {
		http.Error(w, "model_id no soportado o desactivado", http.StatusBadRequest)
		return
	}

	payload.Pregunta = strings.TrimSpace(payload.Pregunta)
	if payload.Pregunta == "" {
		http.Error(w, "pregunta es obligatoria", http.StatusBadRequest)
		return
	}
	if len([]rune(payload.Pregunta)) > 2500 {
		http.Error(w, "pregunta supera el maximo permitido (2500 caracteres)", http.StatusBadRequest)
		return
	}

	fechaUso := time.Now().Format("2006-01-02")
	usoActual, err := dbpkg.GetEmpresaAIUsoDiario(c.dbEmp, payload.EmpresaID, model.Provider, model.ID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario", http.StatusInternalServerError)
		return
	}
	planActual := strings.ToLower(strings.TrimSpace(usoActual.PlanActual))
	if planActual == "" {
		planActual = "free"
	}

	empresaChatEnabled, _, _, err := getChatIAEmpresaEnabled(c.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de chat IA", http.StatusInternalServerError)
		return
	}
	if !empresaChatEnabled {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_chat_disabled",
			"error": "El chat con IA para empresas está desactivado desde configuración lógica del chat con IA.",
		})
		return
	}

	empresaMaxConsultas, _, _, err := getChatIAEmpresaMaxConsultasDia(c.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de límites IA", http.StatusInternalServerError)
		return
	}
	effectiveLimit := effectiveDailyLimitBySuperConfig(empresaMaxConsultas, model.FreeDailyLimit)
	if effectiveLimit == 0 {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_chat_blocked",
			"error": "El chat con IA para empresas está bloqueado por configuración (límite en 0).",
		})
		return
	}

	if usoActual.Consultas >= effectiveLimit {
		c.writeLimitReached(w, payload.EmpresaID, model, usoActual.Consultas)
		return
	}

	contexto, err := dbpkg.BuildEmpresaAIContextoForQuestion(c.dbEmp, payload.EmpresaID, payload.Pregunta, googleAccount)
	if err != nil {
		http.Error(w, "No se pudo construir contexto de empresa", http.StatusBadRequest)
		return
	}

	respuesta, promptTokens, completionTokens, err := c.generateResponse(model, payload.Pregunta, payload.Historial, contexto)
	if err != nil {
		if isProviderLimitError(err) {
			c.writeLimitReached(w, payload.EmpresaID, model, usoActual.Consultas)
			return
		}
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	respuesta = strings.TrimSpace(respuesta)
	if respuesta == "" {
		http.Error(w, "El proveedor no devolvio contenido", http.StatusBadGateway)
		return
	}
	if len([]rune(respuesta)) > 12000 {
		r := []rune(respuesta)
		respuesta = string(r[:12000])
	}

	adminEmail := googleAccount
	if err := dbpkg.UpsertEmpresaAIModeloPreferido(c.dbEmp, payload.EmpresaID, adminEmail, model.ID, adminEmail); err != nil {
		http.Error(w, "No se pudo registrar modelo para la cuenta de Google", http.StatusInternalServerError)
		return
	}
	_, err = dbpkg.RegisterEmpresaAIConsulta(c.dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID:        payload.EmpresaID,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         payload.Pregunta,
		Respuesta:        respuesta,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		FechaConsulta:    time.Now().Format("2006-01-02 15:04:05"),
		PlanActual:       planActual,
		UsuarioCreador:   adminEmail,
		Estado:           "activo",
		Observaciones:    "consulta desde chat_con_inteligencia_artificial",
	})
	if err != nil {
		http.Error(w, "No se pudo registrar auditoria de consulta", http.StatusInternalServerError)
		return
	}

	usoActualizado, err := dbpkg.GetEmpresaAIUsoDiario(c.dbEmp, payload.EmpresaID, model.Provider, model.ID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo obtener uso actualizado", http.StatusInternalServerError)
		return
	}
	restante := effectiveLimit - usoActualizado.Consultas
	if restante < 0 {
		restante = 0
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                       true,
		"empresa_id":               payload.EmpresaID,
		"google_account":           adminEmail,
		"modelo_registrado_google": true,
		"provider":                 model.Provider,
		"model_id":                 model.ID,
		"respuesta":                respuesta,
		"usage": map[string]interface{}{
			"plan":              planActual,
			"daily_used":        usoActualizado.Consultas,
			"daily_limit":       effectiveLimit,
			"daily_remaining":   restante,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		},
		"scope": map[string]interface{}{
			"restricted_by_empresa_id": true,
			"empresa_id":               payload.EmpresaID,
		},
	})
}

// ConsultarConAdjuntoHandler permite consultas con documentos/imagenes y/o forzar uso GPT-5.5.
// Este endpoint está pensado para "2 consultas diarias" (configurable) por empresa usando GPT-5.5.
func (c *EmpresaAIChatController) ConsultarConAdjuntoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}

	// multipart/form-data:
	// - empresa_id
	// - pregunta
	// - historial (json)
	// - use_gpt55 (1/true)
	// - file (opcional)
	att, err := parseSingleAttachmentFromMultipart(r, "file", 8<<20)
	if err != nil {
		http.Error(w, "adjunto inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	empresaID, err := parseInt64FormOptional(r, "empresa_id")
	if err != nil || empresaID <= 0 {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}

	pregunta := strings.TrimSpace(r.FormValue("pregunta"))
	if pregunta == "" {
		http.Error(w, "pregunta es obligatoria", http.StatusBadRequest)
		return
	}
	if len([]rune(pregunta)) > 2500 {
		http.Error(w, "pregunta supera el maximo permitido (2500 caracteres)", http.StatusBadRequest)
		return
	}

	var historial []empresaAIChatMensaje
	if raw := strings.TrimSpace(r.FormValue("historial")); raw != "" {
		_ = json.Unmarshal([]byte(raw), &historial)
	}

	useGPT55 := queryBool(r, "use_gpt55") || queryBool(r, "gpt55") || queryBool(r, "premium")
	if v := strings.TrimSpace(strings.ToLower(r.FormValue("use_gpt55"))); v == "1" || v == "true" || v == "si" || v == "sí" {
		useGPT55 = true
	}
	if att != nil {
		useGPT55 = true
	}
	if !useGPT55 {
		http.Error(w, "Debe adjuntar un archivo o activar use_gpt55=1", http.StatusBadRequest)
		return
	}

	googleAccount := googleAccountFromRequest(r)
	if googleAccount == "" {
		http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
		return
	}
	if err := c.ensureEmpresaAccessByAccount(googleAccount, empresaID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	empresaChatEnabled, _, _, err := getChatIAEmpresaEnabled(c.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de chat IA", http.StatusInternalServerError)
		return
	}
	if !empresaChatEnabled {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_chat_disabled",
			"error": "El chat con IA para empresas está desactivado desde configuración lógica del chat con IA.",
		})
		return
	}

	// Enforce GPT-5.5 daily limit (configurable)
	maxGPT55, _, _, err := getChatIAEmpresaMaxGPT55ConsultasDia(c.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar límite GPT-5.5", http.StatusInternalServerError)
		return
	}
	if maxGPT55 == 0 {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_gpt55_blocked",
			"error": "Las consultas con GPT-5.5 están bloqueadas para empresas (límite en 0).",
		})
		return
	}

	catalog := availableEmpresaAIModelMap(c.dbSuper)
	model, ok := catalog["openai:gpt-5.5"]
	if !ok {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_gpt55_unavailable",
			"error": "GPT-5.5 no está disponible o el proveedor OpenAI está deshabilitado.",
		})
		return
	}

	fechaUso := time.Now().Format("2006-01-02")
	usoActual, err := dbpkg.GetEmpresaAIUsoDiario(c.dbEmp, empresaID, model.Provider, model.ID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario", http.StatusInternalServerError)
		return
	}
	if usoActual.Consultas >= maxGPT55 {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok":    false,
			"code":  "ai_empresa_gpt55_limit_reached",
			"error": "Se alcanzó el límite diario de consultas con GPT-5.5 para esta empresa. Intenta mañana o solicita ampliar el límite en Super Administrador.",
			"usage": map[string]interface{}{
				"daily_used":      usoActual.Consultas,
				"daily_limit":     maxGPT55,
				"daily_remaining": 0,
			},
		})
		return
	}

	contexto, err := dbpkg.BuildEmpresaAIContextoForQuestion(c.dbEmp, empresaID, pregunta, googleAccount)
	if err != nil {
		http.Error(w, "No se pudo construir contexto de empresa", http.StatusBadRequest)
		return
	}

	preguntaFinal := pregunta
	if att != nil && strings.TrimSpace(att.Filename) != "" {
		preguntaFinal = "Adjunto: " + strings.TrimSpace(att.Filename) + " (" + strings.TrimSpace(att.MimeType) + ")\n\n" + pregunta
	}

	respuesta, promptTokens, completionTokens, err := c.generateResponseWithSystemPromptAndAttachment(model, preguntaFinal, historial, buildEmpresaAISystemPrompt(contexto), att)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	respuesta = strings.TrimSpace(respuesta)
	if respuesta == "" {
		http.Error(w, "El proveedor no devolvio contenido", http.StatusBadGateway)
		return
	}
	if len([]rune(respuesta)) > 12000 {
		rn := []rune(respuesta)
		respuesta = string(rn[:12000])
	}

	planActual := strings.ToLower(strings.TrimSpace(usoActual.PlanActual))
	if planActual == "" {
		planActual = "free"
	}
	adminEmail := googleAccount
	_, err = dbpkg.RegisterEmpresaAIConsulta(c.dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID:        empresaID,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         preguntaFinal,
		Respuesta:        respuesta,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		FechaConsulta:    time.Now().Format("2006-01-02 15:04:05"),
		PlanActual:       planActual,
		UsuarioCreador:   adminEmail,
		Estado:           "activo",
		Observaciones:    "consulta con adjunto/gpt55 desde chat_con_inteligencia_artificial",
	})
	if err != nil {
		http.Error(w, "No se pudo registrar auditoria de consulta", http.StatusInternalServerError)
		return
	}

	usoActualizado, _ := dbpkg.GetEmpresaAIUsoDiario(c.dbEmp, empresaID, model.Provider, model.ID, fechaUso)
	restante := maxGPT55 - usoActualizado.Consultas
	if restante < 0 {
		restante = 0
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"provider":   model.Provider,
		"model_id":   model.ID,
		"respuesta":  respuesta,
		"usage": map[string]interface{}{
			"plan":              planActual,
			"daily_used":        usoActualizado.Consultas,
			"daily_limit":       maxGPT55,
			"daily_remaining":   restante,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		},
	})
}

// ConsultarStreamHandler entrega respuesta en vivo (SSE) para texto (GPT-5.4 mini).
func (c *EmpresaAIChatController) ConsultarStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.dbSuper) {
		http.Error(w, "IA desactivada", http.StatusServiceUnavailable)
		return
	}
	enabled, _, _, err := getChatIAEmpresaStreamingEnabled(c.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración streaming", http.StatusInternalServerError)
		return
	}
	if !enabled {
		http.Error(w, "Streaming desactivado", http.StatusServiceUnavailable)
		return
	}

	var payload empresaAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	if payload.EmpresaID <= 0 {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}
	googleAccount := googleAccountFromRequest(r)
	if googleAccount == "" {
		http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
		return
	}
	if err := c.ensureEmpresaAccessByAccount(googleAccount, payload.EmpresaID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	payload.Pregunta = strings.TrimSpace(payload.Pregunta)
	if payload.Pregunta == "" {
		http.Error(w, "pregunta es obligatoria", http.StatusBadRequest)
		return
	}

	empresaChatEnabled, _, _, err := getChatIAEmpresaEnabled(c.dbSuper)
	if err != nil || !empresaChatEnabled {
		http.Error(w, "chat empresarial desactivado", http.StatusServiceUnavailable)
		return
	}

	// Modelo: mantener el preferido actual (normalmente GPT-5.4 mini).
	payload.ModelID = strings.TrimSpace(payload.ModelID)
	if payload.ModelID == "" {
		payload.ModelID = firstAvailableEmpresaAIModelID(c.dbSuper)
	}
	catalog := availableEmpresaAIModelMap(c.dbSuper)
	model, ok := catalog[payload.ModelID]
	if !ok {
		http.Error(w, "model_id no soportado o desactivado", http.StatusBadRequest)
		return
	}
	// Solo streaming para chat/completions.
	if !strings.Contains(strings.ToLower(model.Endpoint), "/v1/chat/completions") {
		http.Error(w, "modelo no soporta streaming", http.StatusBadRequest)
		return
	}

	fechaUso := time.Now().Format("2006-01-02")
	usoActual, err := dbpkg.GetEmpresaAIUsoDiario(c.dbEmp, payload.EmpresaID, model.Provider, model.ID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario", http.StatusInternalServerError)
		return
	}
	empresaMaxConsultas, _, _, err := getChatIAEmpresaMaxConsultasDia(c.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar límites IA", http.StatusInternalServerError)
		return
	}
	effectiveLimit := effectiveDailyLimitBySuperConfig(empresaMaxConsultas, model.FreeDailyLimit)
	if effectiveLimit == 0 || usoActual.Consultas >= effectiveLimit {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_ = sseWriteJSON(w, openAIStreamEvent{Error: "Límite de uso alcanzado."})
		_ = sseWriteJSON(w, openAIStreamEvent{Done: true})
		return
	}

	contexto, err := dbpkg.BuildEmpresaAIContextoForQuestion(c.dbEmp, payload.EmpresaID, payload.Pregunta, googleAccount)
	if err != nil {
		http.Error(w, "No se pudo construir contexto de empresa", http.StatusBadRequest)
		return
	}
	systemPrompt := buildEmpresaAISystemPrompt(contexto)

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var full strings.Builder
	_, err = c.callOpenAIStreamChatCompletions(model, payload.Pregunta, payload.Historial, systemPrompt, func(delta string) {
		full.WriteString(delta)
		_ = sseWriteJSON(w, openAIStreamEvent{Delta: delta})
	})
	if err != nil {
		_ = sseWriteJSON(w, openAIStreamEvent{Error: err.Error()})
		_ = sseWriteJSON(w, openAIStreamEvent{Done: true})
		return
	}
	text := strings.TrimSpace(full.String())
	if text == "" {
		_ = sseWriteJSON(w, openAIStreamEvent{Error: "Respuesta vacía"})
		_ = sseWriteJSON(w, openAIStreamEvent{Done: true})
		return
	}

	planActual := strings.ToLower(strings.TrimSpace(usoActual.PlanActual))
	if planActual == "" {
		planActual = "free"
	}
	// Registrar consulta (tokens desconocidos en streaming → 0).
	_, _ = dbpkg.RegisterEmpresaAIConsulta(c.dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID:        payload.EmpresaID,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         payload.Pregunta,
		Respuesta:        text,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		FechaConsulta:    time.Now().Format("2006-01-02 15:04:05"),
		PlanActual:       planActual,
		UsuarioCreador:   googleAccount,
		Estado:           "activo",
		Observaciones:    "consulta streaming desde chat_con_inteligencia_artificial",
	})

	_ = sseWriteJSON(w, openAIStreamEvent{Done: true})
}

func (c *EmpresaAIChatController) HistorialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.dbSuper),
		})
		return
	}

	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}
	googleAccount := googleAccountFromRequest(r)
	if googleAccount == "" {
		http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
		return
	}
	if err := c.ensureEmpresaAccessByAccount(googleAccount, empresaID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	limit, err := parseIntQueryOptional(r, "limit")
	if err != nil {
		http.Error(w, "limit invalido", http.StatusBadRequest)
		return
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := dbpkg.ListEmpresaAIConsultasRecientes(c.dbEmp, empresaID, limit)
	if err != nil {
		http.Error(w, "No se pudo consultar historial", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":         true,
		"empresa_id": empresaID,
		"items":      rows,
	})
}

func (c *EmpresaAIChatController) ensureEmpresaAccess(r *http.Request, empresaID int64) error {
	adminEmail := googleAccountFromRequest(r)
	return c.ensureEmpresaAccessByAccount(adminEmail, empresaID)
}

func (c *EmpresaAIChatController) ensureEmpresaAccessByAccount(adminEmail string, empresaID int64) error {
	adminEmail = strings.TrimSpace(strings.ToLower(adminEmail))
	if adminEmail == "" {
		return fmt.Errorf("no se pudo identificar la cuenta de Google del usuario autenticado")
	}
	ok, err := dbpkg.CanAdminAccessEmpresaIA(c.dbEmp, c.dbSuper, adminEmail, empresaID)
	if err != nil {
		return fmt.Errorf("no se pudo validar alcance de empresa")
	}
	if !ok {
		return fmt.Errorf("empresa_id fuera del alcance del usuario autenticado")
	}
	return nil
}

func (c *EmpresaAIChatController) writeLimitReached(w http.ResponseWriter, empresaID int64, model empresaAIModelDef, used int64) {
	writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
		"ok":              false,
		"code":            "free_tier_limit_reached",
		"error":           "Se alcanzo el limite del plan gratuito para este modelo.",
		"model_id":        model.ID,
		"provider":        model.Provider,
		"daily_used":      used,
		"daily_limit":     model.FreeDailyLimit,
		"upgrade_url":     model.PlanURL + "?empresa_id=" + fmt.Sprintf("%d", empresaID),
		"upgrade_message": "Puedes adquirir un plan para ampliar limites y capacidad.",
	})
}

func (c *EmpresaAIChatController) generateResponse(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, contexto string) (string, int64, int64, error) {
	systemPrompt := buildEmpresaAISystemPrompt(contexto)
	return c.generateResponseWithSystemPrompt(model, pregunta, historial, systemPrompt)
}

func (c *EmpresaAIChatController) generateResponseWithSystemPrompt(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, systemPrompt string) (string, int64, int64, error) {
	if strings.EqualFold(model.Provider, "google") {
		return c.callGoogleGeminiWithSystemPrompt(model, pregunta, historial, systemPrompt)
	}
	if strings.EqualFold(model.Provider, "openai") {
		// Si el modelo apunta al endpoint de Responses, usar el flujo nuevo.
		if strings.Contains(strings.ToLower(model.Endpoint), "/v1/responses") {
			return c.callOpenAIResponsesWithSystemPrompt(model, pregunta, historial, systemPrompt, nil)
		}
		return c.callOpenAIWithSystemPrompt(model, pregunta, historial, systemPrompt)
	}
	return "", 0, 0, fmt.Errorf("proveedor no soportado")
}

func (c *EmpresaAIChatController) generateResponseWithSystemPromptAndAttachment(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, systemPrompt string, att *aiAttachment) (string, int64, int64, error) {
	if strings.EqualFold(model.Provider, "openai") {
		return c.callOpenAIResponsesWithSystemPrompt(model, pregunta, historial, systemPrompt, att)
	}
	return c.generateResponseWithSystemPrompt(model, pregunta, historial, systemPrompt)
}

func (c *EmpresaAIChatController) callOpenAIResponsesWithSystemPrompt(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, systemPrompt string, att *aiAttachment) (string, int64, int64, error) {
	apiKey, err := c.resolveModelAPIKey(model)
	if err != nil {
		return "", 0, 0, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return "", 0, 0, fmt.Errorf("OPENAI_API_KEY no configurada")
	}

	type inMsg struct {
		Role    string      `json:"role"`
		Content interface{} `json:"content"`
	}
	type inPart struct {
		Type     string `json:"type"`
		Text     string `json:"text,omitempty"`
		ImageURL string `json:"image_url,omitempty"`
		Filename string `json:"filename,omitempty"`
		FileData string `json:"file_data,omitempty"`
	}

	parts := make([]inPart, 0, 4)
	parts = append(parts, inPart{Type: "input_text", Text: strings.TrimSpace(pregunta)})
	if att != nil && len(att.Bytes) > 0 {
		mt := strings.ToLower(strings.TrimSpace(att.MimeType))
		if mt == "" {
			mt = "application/octet-stream"
		}
		if strings.HasPrefix(mt, "image/") {
			parts = append(parts, inPart{
				Type:     "input_image",
				ImageURL: "data:" + mt + ";base64," + base64.StdEncoding.EncodeToString(att.Bytes),
			})
		} else {
			parts = append(parts, inPart{
				Type:     "input_file",
				Filename: strings.TrimSpace(att.Filename),
				FileData: base64.StdEncoding.EncodeToString(att.Bytes),
			})
		}
	}

	messages := make([]inMsg, 0, 12)
	messages = append(messages, inMsg{Role: "system", Content: strings.TrimSpace(systemPrompt)})
	clean := sanitizeHistorial(historial, 8)
	for _, h := range clean {
		role := strings.ToLower(strings.TrimSpace(h.Rol))
		if role != "user" && role != "assistant" {
			continue
		}
		messages = append(messages, inMsg{Role: role, Content: strings.TrimSpace(h.Contenido)})
	}
	messages = append(messages, inMsg{Role: "user", Content: parts})

	body := map[string]interface{}{
		"model": strings.TrimSpace(model.UpstreamModel),
		"input": messages,
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, strings.TrimSpace(model.Endpoint), bytes.NewReader(payload))
	if err != nil {
		return "", 0, 0, fmt.Errorf("no se pudo crear solicitud al proveedor")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("no se pudo contactar proveedor: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", 0, 0, &aiProviderHTTPError{
			Provider: "openai",
			Status:   resp.StatusCode,
			Body:     truncateText(string(raw), 600),
		}
	}

	var parsed struct {
		Output []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
		Usage struct {
			InputTokens  int64 `json:"input_tokens"`
			OutputTokens int64 `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", 0, 0, fmt.Errorf("respuesta del proveedor no es JSON valido")
	}

	var textParts []string
	for _, out := range parsed.Output {
		for _, c := range out.Content {
			if strings.EqualFold(strings.TrimSpace(c.Type), "output_text") && strings.TrimSpace(c.Text) != "" {
				textParts = append(textParts, strings.TrimSpace(c.Text))
			}
		}
	}
	text := strings.TrimSpace(strings.Join(textParts, "\n\n"))
	if text == "" {
		return "", 0, 0, fmt.Errorf("el proveedor devolvio respuesta vacia")
	}
	return text, parsed.Usage.InputTokens, parsed.Usage.OutputTokens, nil
}

func parseSingleAttachmentFromMultipart(r *http.Request, field string, maxBytes int64) (*aiAttachment, error) {
	if maxBytes <= 0 {
		maxBytes = 8 << 20
	}
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		return nil, err
	}
	f, fh, err := r.FormFile(field)
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var b []byte
	limited := io.LimitReader(f, maxBytes+1)
	b, _ = io.ReadAll(limited)
	if int64(len(b)) > maxBytes {
		return nil, fmt.Errorf("adjunto supera el máximo permitido (%d bytes)", maxBytes)
	}
	mt := strings.TrimSpace(fh.Header.Get("Content-Type"))
	if mt == "" {
		mt = mime.TypeByExtension(strings.ToLower(strings.TrimSpace(filepath.Ext(fh.Filename))))
	}
	if mt == "" {
		mt = "application/octet-stream"
	}
	return &aiAttachment{
		Filename: strings.TrimSpace(fh.Filename),
		MimeType: mt,
		Bytes:    b,
	}, nil
}

func (c *EmpresaAIChatController) callOpenAIWithSystemPrompt(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, systemPrompt string) (string, int64, int64, error) {
	apiKey, err := c.resolveModelAPIKey(model)
	if err != nil {
		return "", 0, 0, err
	}
	if strings.TrimSpace(apiKey) == "" {
		return "", 0, 0, fmt.Errorf("OPENAI_API_KEY no configurada")
	}

	type msg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	messages := make([]msg, 0, 12)
	messages = append(messages, msg{Role: "system", Content: systemPrompt})

	clean := sanitizeHistorial(historial, 8)
	for _, h := range clean {
		role := strings.ToLower(strings.TrimSpace(h.Rol))
		if role != "user" && role != "assistant" {
			continue
		}
		messages = append(messages, msg{Role: role, Content: strings.TrimSpace(h.Contenido)})
	}
	messages = append(messages, msg{Role: "user", Content: strings.TrimSpace(pregunta)})

	body := map[string]interface{}{
		"model":                 strings.TrimSpace(model.UpstreamModel),
		"messages":              messages,
		"temperature":           0.2,
		"max_completion_tokens": 700,
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, model.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", 0, 0, fmt.Errorf("no se pudo crear solicitud al proveedor")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("no se pudo contactar proveedor: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", 0, 0, &aiProviderHTTPError{
			Provider: "openai",
			Status:   resp.StatusCode,
			Body:     truncateText(string(raw), 600),
		}
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int64 `json:"prompt_tokens"`
			CompletionTokens int64 `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", 0, 0, fmt.Errorf("respuesta del proveedor no es JSON valido")
	}
	if len(parsed.Choices) == 0 {
		return "", 0, 0, fmt.Errorf("el proveedor devolvio respuesta vacia")
	}
	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if text == "" {
		return "", 0, 0, fmt.Errorf("el proveedor devolvio respuesta vacia")
	}
	return text, parsed.Usage.PromptTokens, parsed.Usage.CompletionTokens, nil
}

func buildEmpresaAISystemPrompt(contexto string) string {
	return "Eres un asistente empresarial para el sistema POS multiempresa. " +
		"Responde en espanol claro y accionable. Usa solo el contexto validado por empresa_id. " +
		"No inventes consultas SQL ni afirmes acceso a otras empresas. " +
		"Cuando el usuario pida ejecutar acciones operativas o de base de datos (consultar, crear, editar o eliminar), NO ejecutes nada directamente. " +
		"En su lugar, debes proponer una accion como una sugerencia estructurada para que el usuario la confirme. " +
		"Regla de seguridad: puedes proponer acciones GET/OPEN/POST/PUT/DELETE, pero TODA accion debe pedir confirmacion previa; y si es DELETE o afecta muchos datos, pide confirmacion adicional. " +
		"Para operaciones de base de datos genericas, usa el endpoint protegido /api/empresa/db_admin con action=schema|columns|select|insert|update|delete (siempre aislado por empresa_id). " +
		"Importante (foto de carta/lista de precios y egresos): cuando el usuario adjunte una foto y pida registrar productos o egresos, primero extrae y presenta una lista estructurada para revision humana. " +
		"Solo tras una confirmacion explicita del usuario (por ejemplo: 'si, confirma y guarda'), genera UNA sola accion PCS_ACTION que llame a POST /api/empresa/ia/importar_desde_foto con empresa_id y un arreglo de productos y/o egresos. " +
		"Usa ese endpoint (no llames directamente /api/empresa/productos ni /api/empresa/finanzas/movimientos) para que el servidor aplique la importacion solo si el usuario es administrador. " +
		"Solo cuando tengas TODOS los datos obligatorios, incluye al FINAL de tu respuesta un bloque literal con el prefijo EXACTO `PCS_ACTION` en una linea aparte, seguido por un JSON valido. " +
		"Formato requerido:\n" +
		"PCS_ACTION\n" +
		"{\"version\":1,\"actions\":[{\"id\":\"...\",\"title\":\"...\",\"endpoint\":\"/api/empresa/...\",\"method\":\"GET|OPEN|POST|PUT|DELETE\",\"body\":{...},\"requires_confirmation\":true}],\"note\":\"...\"}\n" +
		"- No incluyas Markdown dentro del JSON. - No incluyas comentarios. - El JSON debe ser parseable.\n" +
		"Si te falta un dato (por ejemplo categoria_id, impuesto, monto, fecha, estacion_id), NO generes PCS_ACTION: pregunta lo que falta primero. " +
		"Si la operacion es riesgosa o destructiva, pregunta confirmacion adicional.\n\n" +
		"Si existe la seccion CONSULTAS_SEGURAS_RESUELTAS, priorizala como fuente principal para responder la pregunta actual. " +
		"Si faltan datos, dilo explicitamente y sugiere que dato consultar.\n\nCONTEXTO_EMPRESA_VALIDADO:\n" + contexto
}

func (c *EmpresaAIChatController) callGoogleGemini(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, contexto string) (string, int64, int64, error) {
	systemPrompt := buildEmpresaAISystemPrompt(contexto)
	return c.callGoogleGeminiWithSystemPrompt(model, pregunta, historial, systemPrompt)
}

func (c *EmpresaAIChatController) callGoogleGeminiWithSystemPrompt(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, systemPrompt string) (string, int64, int64, error) {
	apiKey, err := c.resolveModelAPIKey(model)
	if err != nil {
		return "", 0, 0, err
	}

	endpoint := model.Endpoint
	sep := "?"
	if strings.Contains(endpoint, "?") {
		sep = "&"
	}
	endpoint = endpoint + sep + "key=" + url.QueryEscape(apiKey)

	contents := buildGeminiContents(pregunta, historial)
	body := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{{"text": systemPrompt}},
		},
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"temperature":     0.2,
			"maxOutputTokens": 700,
		},
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", 0, 0, fmt.Errorf("no se pudo crear solicitud al proveedor")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("no se pudo contactar proveedor: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", 0, 0, fmt.Errorf("error proveedor (%d): %s", resp.StatusCode, truncateText(string(raw), 600))
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int64 `json:"promptTokenCount"`
			CandidatesTokenCount int64 `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", 0, 0, fmt.Errorf("respuesta del proveedor no es JSON valido")
	}
	text := extractGeminiText(parsed.Candidates)
	if text == "" {
		return "", 0, 0, fmt.Errorf("el proveedor devolvio respuesta vacia")
	}
	return text, parsed.UsageMetadata.PromptTokenCount, parsed.UsageMetadata.CandidatesTokenCount, nil
}

func buildGeminiContents(pregunta string, historial []empresaAIChatMensaje) []map[string]interface{} {
	clean := sanitizeHistorial(historial, 8)
	out := make([]map[string]interface{}, 0, len(clean)+1)

	for _, h := range clean {
		role := "user"
		if h.Rol == "assistant" {
			role = "model"
		}
		out = append(out, map[string]interface{}{
			"role":  role,
			"parts": []map[string]string{{"text": h.Contenido}},
		})
	}

	out = append(out, map[string]interface{}{
		"role":  "user",
		"parts": []map[string]string{{"text": strings.TrimSpace(pregunta)}},
	})
	return out
}

func extractGeminiText(candidates []struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}) string {
	if len(candidates) == 0 {
		return ""
	}
	parts := candidates[0].Content.Parts
	if len(parts) == 0 {
		return ""
	}
	chunks := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p.Text)
		if t == "" {
			continue
		}
		chunks = append(chunks, t)
	}
	return strings.TrimSpace(strings.Join(chunks, "\n"))
}

func sanitizeHistorial(in []empresaAIChatMensaje, max int) []empresaAIChatMensaje {
	if max <= 0 {
		max = 6
	}
	out := make([]empresaAIChatMensaje, 0, max)
	for _, item := range in {
		role := strings.ToLower(strings.TrimSpace(item.Rol))
		if role != "user" && role != "assistant" {
			continue
		}
		msg := strings.TrimSpace(item.Contenido)
		if msg == "" {
			continue
		}
		if len([]rune(msg)) > 1500 {
			r := []rune(msg)
			msg = string(r[:1500])
		}
		out = append(out, empresaAIChatMensaje{Rol: role, Contenido: msg})
	}
	if len(out) <= max {
		return out
	}
	return out[len(out)-max:]
}

func (c *EmpresaAIChatController) resolveModelAPIKey(model empresaAIModelDef) (string, error) {
	if strings.TrimSpace(model.ApiKeyEnv) == "" {
		return "", nil
	}
	if c.dbSuper != nil {
		if def, ok := aiCredentialByModelID()[model.ID]; ok {
			if key, err := getDecryptedConfigValue(c.dbSuper, def.ConfigKey); err == nil {
				if strings.TrimSpace(key) != "" {
					return strings.TrimSpace(key), nil
				}
			} else {
				log.Printf("[chat_ia] warning: no se pudo leer config_key=%s: %v", def.ConfigKey, err)
			}
		}

		providerKey := aiProviderConfigKey(model.Provider)
		if providerKey != "" {
			if key, err := getDecryptedConfigValue(c.dbSuper, providerKey); err == nil {
				if strings.TrimSpace(key) != "" {
					return strings.TrimSpace(key), nil
				}
			} else {
				log.Printf("[chat_ia] warning: no se pudo leer provider_key=%s: %v", providerKey, err)
			}
		}
	}

	apiKey := strings.TrimSpace(os.Getenv(model.ApiKeyEnv))
	if apiKey != "" {
		return apiKey, nil
	}
	return "", fmt.Errorf("la credencial %s no esta configurada en servidor", model.ApiKeyEnv)
}

func truncateText(v string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(strings.TrimSpace(v))
	if len(r) <= max {
		return string(r)
	}
	return string(r[:max])
}

func isProviderLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	needles := []string{
		"rate limit",
		"too many requests",
		"quota",
		"insufficient_quota",
		"resource_exhausted",
		"quota exceeded",
		"insufficient balance",
		"insufficient_balance",
		"invalid_request_error",
		"payment required",
		"insufficient_funds",
		"free tier",
	}
	for _, n := range needles {
		if strings.Contains(msg, n) {
			return true
		}
	}
	return false
}

func googleAccountFromRequest(r *http.Request) string {
	if v := r.Context().Value("adminEmail"); v != nil {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(strings.ToLower(s))
			if s != "" && s != "sistema" {
				return s
			}
		}
	}
	h := strings.TrimSpace(strings.ToLower(r.Header.Get("X-Admin-Email")))
	if h != "" && h != "sistema" {
		return h
	}
	return ""
}
