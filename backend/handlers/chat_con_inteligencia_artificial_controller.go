package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type EmpresaAIChatController struct {
	dbEmp   *sql.DB
	dbSuper *sql.DB
	client  *http.Client
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
			ID:             "openai:gpt-4o-mini",
			Provider:       "openai",
			DisplayName:    "OpenAI GPT-4o mini",
			UpstreamModel:  "gpt-4o-mini",
			Endpoint:       "https://api.openai.com/v1/chat/completions",
			ApiKeyEnv:      "OPENAI_API_KEY",
			Famous:         true,
			FreeDailyLimit: 120,
			Description:    "Chat empresarial con OpenAI, restringido por empresa_id y con API key cifrada en el panel super.",
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

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"empresa_id":       empresaID,
		"google_account":   googleAccount,
		"modelo_preferido": modeloPreferido,
		"modelos":          items,
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

	if planActual == "free" && model.FreeDailyLimit > 0 && usoActual.Consultas >= int64(model.FreeDailyLimit) {
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
	restante := int64(model.FreeDailyLimit) - usoActualizado.Consultas
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
			"daily_limit":       model.FreeDailyLimit,
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
		return c.callOpenAIWithSystemPrompt(model, pregunta, historial, systemPrompt)
	}
	return "", 0, 0, fmt.Errorf("proveedor no soportado")
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
		"model":       strings.TrimSpace(model.UpstreamModel),
		"messages":    messages,
		"temperature": 0.2,
		"max_tokens":  700,
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
		return "", 0, 0, fmt.Errorf("error proveedor (%d): %s", resp.StatusCode, truncateText(string(raw), 600))
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
