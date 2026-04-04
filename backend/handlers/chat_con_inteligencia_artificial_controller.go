package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
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
			FreeDailyLimit: 20,
			Description:    "Modelo popular y rapido. Disponible si la cuenta OpenAI tiene cuota activa.",
		},
		{
			ID:             "openai:gpt-4.1-mini",
			Provider:       "openai",
			DisplayName:    "OpenAI GPT-4.1 mini",
			UpstreamModel:  "gpt-4.1-mini",
			Endpoint:       "https://api.openai.com/v1/chat/completions",
			ApiKeyEnv:      "OPENAI_API_KEY",
			Famous:         true,
			FreeDailyLimit: 20,
			Description:    "Modelo conocido para razonamiento general con bajo costo.",
		},
		{
			ID:             "deepseek:deepseek-chat",
			Provider:       "deepseek",
			DisplayName:    "DeepSeek Chat",
			UpstreamModel:  "deepseek-chat",
			Endpoint:       "https://api.deepseek.com/chat/completions",
			ApiKeyEnv:      "DEEPSEEK_API_KEY",
			Famous:         true,
			FreeDailyLimit: 50,
			Description:    "Modelo popular para chat general.",
		},
		{
			ID:             "deepseek:deepseek-reasoner",
			Provider:       "deepseek",
			DisplayName:    "DeepSeek Reasoner",
			UpstreamModel:  "deepseek-reasoner",
			Endpoint:       "https://api.deepseek.com/chat/completions",
			ApiKeyEnv:      "DEEPSEEK_API_KEY",
			Famous:         true,
			FreeDailyLimit: 40,
			Description:    "Modelo reconocido para razonamiento mas profundo.",
		},
		{
			ID:             "huggingface:meta-llama/Llama-3.1-8B-Instruct",
			Provider:       "huggingface",
			DisplayName:    "Meta Llama 3.1 8B Instruct",
			UpstreamModel:  "meta-llama/Llama-3.1-8B-Instruct",
			Endpoint:       "https://api-inference.huggingface.co/models/meta-llama/Llama-3.1-8B-Instruct",
			ApiKeyEnv:      "HUGGINGFACE_API_KEY",
			Famous:         true,
			FreeDailyLimit: 30,
			Description:    "Modelo famoso de Meta disponible via Hugging Face Inference.",
		},
		{
			ID:             "huggingface:mistralai/Mistral-7B-Instruct-v0.3",
			Provider:       "huggingface",
			DisplayName:    "Mistral 7B Instruct",
			UpstreamModel:  "mistralai/Mistral-7B-Instruct-v0.3",
			Endpoint:       "https://api-inference.huggingface.co/models/mistralai/Mistral-7B-Instruct-v0.3",
			ApiKeyEnv:      "HUGGINGFACE_API_KEY",
			Famous:         true,
			FreeDailyLimit: 30,
			Description:    "Modelo muy usado para chat y tareas generales.",
		},
		{
			ID:             "huggingface:google/gemma-2-9b-it",
			Provider:       "huggingface",
			DisplayName:    "Google Gemma 2 9B IT",
			UpstreamModel:  "google/gemma-2-9b-it",
			Endpoint:       "https://api-inference.huggingface.co/models/google/gemma-2-9b-it",
			ApiKeyEnv:      "HUGGINGFACE_API_KEY",
			Famous:         true,
			FreeDailyLimit: 30,
			Description:    "Modelo famoso de Google para dialogo e instrucciones.",
		},
		{
			ID:             "huggingface:Qwen/Qwen2.5-7B-Instruct",
			Provider:       "huggingface",
			DisplayName:    "Qwen 2.5 7B Instruct",
			UpstreamModel:  "Qwen/Qwen2.5-7B-Instruct",
			Endpoint:       "https://api-inference.huggingface.co/models/Qwen/Qwen2.5-7B-Instruct",
			ApiKeyEnv:      "HUGGINGFACE_API_KEY",
			Famous:         true,
			FreeDailyLimit: 30,
			Description:    "Modelo muy conocido para chat multilingue.",
		},
	}
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

func (c *EmpresaAIChatController) ModelosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}

	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}
	if err := c.ensureEmpresaAccess(r, empresaID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	googleAccount := googleAccountFromRequest(r)
	if googleAccount == "" {
		http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
		return
	}

	modeloPreferido, err := dbpkg.GetEmpresaAIModeloPreferido(c.dbEmp, empresaID, googleAccount)
	if err != nil {
		http.Error(w, "No se pudo consultar el modelo preferido", http.StatusInternalServerError)
		return
	}

	catalog := empresaAIModelCatalog()
	sort.SliceStable(catalog, func(i, j int) bool {
		if catalog[i].Famous == catalog[j].Famous {
			return catalog[i].DisplayName < catalog[j].DisplayName
		}
		return catalog[i].Famous && !catalog[j].Famous
	})

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
	switch r.Method {
	case http.MethodGet:
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
			return
		}
		if err := c.ensureEmpresaAccess(r, empresaID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		googleAccount := googleAccountFromRequest(r)
		if googleAccount == "" {
			http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
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
		if err := c.ensureEmpresaAccess(r, payload.EmpresaID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		payload.ModelID = strings.TrimSpace(payload.ModelID)
		catalog := empresaAIModelMap()
		if _, ok := catalog[payload.ModelID]; !ok {
			http.Error(w, "model_id no soportado", http.StatusBadRequest)
			return
		}

		googleAccount := googleAccountFromRequest(r)
		if googleAccount == "" {
			http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
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
	if err := c.ensureEmpresaAccess(r, payload.EmpresaID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	payload.ModelID = strings.TrimSpace(payload.ModelID)
	catalog := empresaAIModelMap()
	model, ok := catalog[payload.ModelID]
	if !ok {
		http.Error(w, "model_id no soportado", http.StatusBadRequest)
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

	contexto, err := dbpkg.BuildEmpresaAIContexto(c.dbEmp, payload.EmpresaID)
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

	adminEmail := googleAccountFromRequest(r)
	if adminEmail == "" {
		http.Error(w, "No se pudo identificar la cuenta de Google del usuario autenticado", http.StatusUnauthorized)
		return
	}
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

	empresaID, err := parseEmpresaIDQuery(r)
	if err != nil {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}
	if err := c.ensureEmpresaAccess(r, empresaID); err != nil {
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
	switch model.Provider {
	case "openai", "deepseek":
		return c.callOpenAICompatible(model, pregunta, historial, contexto)
	case "huggingface":
		return c.callHuggingFace(model, pregunta, historial, contexto)
	default:
		return "", 0, 0, fmt.Errorf("proveedor no soportado")
	}
}

func (c *EmpresaAIChatController) callOpenAICompatible(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, contexto string) (string, int64, int64, error) {
	apiKey := strings.TrimSpace(os.Getenv(model.ApiKeyEnv))
	if apiKey == "" {
		return "", 0, 0, fmt.Errorf("la credencial %s no esta configurada en servidor", model.ApiKeyEnv)
	}

	messages := make([]map[string]string, 0, 12)
	messages = append(messages, map[string]string{
		"role":    "system",
		"content": "Eres un asistente empresarial. Responde en espanol claro. Usa solo el contexto de la empresa entregado. Si no hay datos suficientes, dilo explicitamente.",
	})
	messages = append(messages, map[string]string{
		"role":    "system",
		"content": "Contexto validado por empresa_id:\n" + contexto,
	})

	for _, h := range sanitizeHistorial(historial, 8) {
		messages = append(messages, map[string]string{
			"role":    h.Rol,
			"content": h.Contenido,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": pregunta,
	})

	body := map[string]interface{}{
		"model":       model.UpstreamModel,
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
	req.Header.Set("Authorization", "Bearer "+apiKey)

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
		return "", 0, 0, fmt.Errorf("el proveedor no devolvio choices")
	}
	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if text == "" {
		return "", 0, 0, fmt.Errorf("el proveedor devolvio respuesta vacia")
	}
	return text, parsed.Usage.PromptTokens, parsed.Usage.CompletionTokens, nil
}

func (c *EmpresaAIChatController) callHuggingFace(model empresaAIModelDef, pregunta string, historial []empresaAIChatMensaje, contexto string) (string, int64, int64, error) {
	apiKey := strings.TrimSpace(os.Getenv(model.ApiKeyEnv))
	if apiKey == "" {
		return "", 0, 0, fmt.Errorf("la credencial %s no esta configurada en servidor", model.ApiKeyEnv)
	}

	prompt := buildHuggingFacePrompt(pregunta, historial, contexto)
	body := map[string]interface{}{
		"inputs": prompt,
		"parameters": map[string]interface{}{
			"max_new_tokens":   700,
			"temperature":      0.2,
			"return_full_text": false,
		},
	}
	payload, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, model.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", 0, 0, fmt.Errorf("no se pudo crear solicitud al proveedor")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", 0, 0, fmt.Errorf("no se pudo contactar proveedor: %v", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", 0, 0, fmt.Errorf("error proveedor (%d): %s", resp.StatusCode, truncateText(string(raw), 600))
	}

	text := extractHuggingFaceText(raw)
	text = strings.TrimSpace(text)
	if text == "" {
		return "", 0, 0, fmt.Errorf("el proveedor no devolvio texto generado")
	}
	return text, 0, 0, nil
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

func buildHuggingFacePrompt(pregunta string, historial []empresaAIChatMensaje, contexto string) string {
	var b strings.Builder
	b.WriteString("SISTEMA:\n")
	b.WriteString("Responde solo con informacion util para negocio. Si faltan datos, indicalo.\n\n")
	b.WriteString("CONTEXTO_EMPRESA:\n")
	b.WriteString(contexto)
	b.WriteString("\n\n")

	h := sanitizeHistorial(historial, 6)
	if len(h) > 0 {
		b.WriteString("HISTORIAL_RECIENTE:\n")
		for _, item := range h {
			prefix := "Usuario"
			if item.Rol == "assistant" {
				prefix = "Asistente"
			}
			b.WriteString(prefix + ": " + item.Contenido + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("PREGUNTA_ACTUAL:\n")
	b.WriteString(strings.TrimSpace(pregunta))
	b.WriteString("\n\nRESPUESTA:")
	return b.String()
}

func extractHuggingFaceText(raw []byte) string {
	var asArray []map[string]interface{}
	if err := json.Unmarshal(raw, &asArray); err == nil && len(asArray) > 0 {
		if v, ok := asArray[0]["generated_text"].(string); ok {
			return v
		}
	}

	var asObj map[string]interface{}
	if err := json.Unmarshal(raw, &asObj); err == nil {
		if v, ok := asObj["generated_text"].(string); ok {
			return v
		}
		if v, ok := asObj["error"].(string); ok {
			return "ERROR_HF: " + v
		}
	}

	return string(raw)
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
