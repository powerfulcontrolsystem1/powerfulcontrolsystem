package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type superAIChatRequest struct {
	ModelID        string                 `json:"model_id"`
	Pregunta       string                 `json:"pregunta"`
	Historial      []empresaAIChatMensaje `json:"historial"`
	Temperatura    float64                `json:"temperatura"`
	PaginaContexto string                 `json:"pagina_contexto,omitempty"`
}

type superAIModeloPreferidoPayload struct {
	ModelID string `json:"model_id"`
}

type SuperAIChatController struct {
	base *EmpresaAIChatController
}

func NewSuperAIChatController(dbEmp, dbSuper *sql.DB) *SuperAIChatController {
	return &SuperAIChatController{base: NewEmpresaAIChatController(dbEmp, dbSuper)}
}

func (c *SuperAIChatController) ModelosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.base.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.base.dbSuper),
		})
		return
	}

	adminEmail, ok := c.requireSuperAdmin(w, r)
	if !ok {
		return
	}

	modeloPreferido, err := dbpkg.GetSuperAIModeloPreferido(c.base.dbSuper, adminEmail)
	if err != nil {
		http.Error(w, "No se pudo consultar el modelo preferido", http.StatusInternalServerError)
		return
	}

	catalog := availableEmpresaAIModelCatalog(c.base.dbSuper)
	if len(catalog) == 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_models_unavailable",
			"error":          "No hay proveedores IA habilitados para super administrador.",
			"service_status": superAIServiceStatus(c.base.dbSuper),
		})
		return
	}
	availableMap := availableEmpresaAIModelMap(c.base.dbSuper)
	if _, ok := availableMap[modeloPreferido]; !ok {
		modeloPreferido = firstAvailableEmpresaAIModelID(c.base.dbSuper)
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
			"plan_url":         "/super/licencias.html",
		})
	}

	streamingEnabled, _, _, _ := getChatIASuperStreamingEnabled(c.base.dbSuper)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":                true,
		"admin_email":       adminEmail,
		"scope":             "global_superadministrador",
		"modelo_preferido":  modeloPreferido,
		"streaming_enabled": streamingEnabled,
		"modelos":           items,
	})
}

func (c *SuperAIChatController) ModeloPreferidoHandler(w http.ResponseWriter, r *http.Request) {
	if !isSuperAIEnabled(c.base.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.base.dbSuper),
		})
		return
	}
	switch r.Method {
	case http.MethodGet:
		adminEmail, ok := c.requireSuperAdmin(w, r)
		if !ok {
			return
		}

		modelID, err := dbpkg.GetSuperAIModeloPreferido(c.base.dbSuper, adminEmail)
		if err != nil {
			http.Error(w, "No se pudo consultar el modelo preferido", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":          true,
			"admin_email": adminEmail,
			"model_id":    modelID,
		})
		return

	case http.MethodPut:
		adminEmail, ok := c.requireSuperAdmin(w, r)
		if !ok {
			return
		}

		var payload superAIModeloPreferidoPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		payload.ModelID = strings.TrimSpace(payload.ModelID)
		if payload.ModelID == "" {
			payload.ModelID = firstAvailableEmpresaAIModelID(c.base.dbSuper)
		}

		catalog := availableEmpresaAIModelMap(c.base.dbSuper)
		if len(catalog) == 0 {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":             false,
				"code":           "ai_models_unavailable",
				"error":          "No hay proveedores IA habilitados para super administrador.",
				"service_status": superAIServiceStatus(c.base.dbSuper),
			})
			return
		}
		if _, found := catalog[payload.ModelID]; !found {
			http.Error(w, "model_id no soportado o desactivado", http.StatusBadRequest)
			return
		}

		if err := dbpkg.UpsertSuperAIModeloPreferido(c.base.dbSuper, adminEmail, payload.ModelID, adminEmail); err != nil {
			http.Error(w, "No se pudo registrar el modelo preferido", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":          true,
			"admin_email": adminEmail,
			"model_id":    payload.ModelID,
			"saved":       true,
		})
		return
	}

	http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
}

func (c *SuperAIChatController) ConsultarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.base.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.base.dbSuper),
		})
		return
	}

	adminEmail, ok := c.requireSuperAdmin(w, r)
	if !ok {
		return
	}

	var payload superAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	payload.ModelID = strings.TrimSpace(payload.ModelID)
	if payload.ModelID == "" {
		payload.ModelID = firstAvailableEmpresaAIModelID(c.base.dbSuper)
	}
	catalog := availableEmpresaAIModelMap(c.base.dbSuper)
	if len(catalog) == 0 {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_models_unavailable",
			"error":          "No hay proveedores IA habilitados para super administrador.",
			"service_status": superAIServiceStatus(c.base.dbSuper),
		})
		return
	}
	model, found := catalog[payload.ModelID]
	if !found {
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
	usoActual, err := dbpkg.GetSuperAIUsoDiario(c.base.dbSuper, adminEmail, model.Provider, model.ID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario", http.StatusInternalServerError)
		return
	}
	planActual := strings.ToLower(strings.TrimSpace(usoActual.PlanActual))
	if planActual == "" {
		planActual = "free"
	}

	superChatEnabled, _, _, err := getChatIASuperEnabled(c.base.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de chat IA", http.StatusInternalServerError)
		return
	}
	if !superChatEnabled {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_super_chat_disabled",
			"error": "El chat global de super administrador está desactivado desde configuración lógica del chat con IA.",
		})
		return
	}

	superMaxConsultas, _, _, err := getChatIASuperMaxConsultasDia(c.base.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de límites IA", http.StatusInternalServerError)
		return
	}
	effectiveLimit := effectiveDailyLimitBySuperConfig(superMaxConsultas, model.FreeDailyLimit)
	if effectiveLimit == 0 {
		writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
			"ok":    false,
			"code":  "ai_super_chat_blocked",
			"error": "El chat global está bloqueado por configuración (límite en 0).",
		})
		return
	}

	if usoActual.Consultas >= effectiveLimit {
		c.writeLimitReached(w, model, usoActual.Consultas)
		return
	}

	empresaRO, _, _, err := getChatIAEmpresaSoloLectura(c.base.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de contexto IA", http.StatusInternalServerError)
		return
	}

	contexto, err := dbpkg.BuildSuperAIContextoForQuestion(c.base.dbEmp, c.base.dbSuper, adminEmail, payload.Pregunta, dbpkg.SuperAIContextoOpts{
		EmpresaSoloLectura: empresaRO,
	})
	if err != nil {
		http.Error(w, "No se pudo construir contexto global", http.StatusBadRequest)
		return
	}

	superMeta := c.base.dbSuper != nil
	respuesta, promptTokens, completionTokens, err := c.base.generateResponseWithSystemPrompt(model, payload.Pregunta, payload.Historial, buildSuperAISystemPrompt(contexto, superMeta, empresaRO))
	if err != nil {
		if isProviderLimitError(err) {
			c.writeLimitReached(w, model, usoActual.Consultas)
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

	if err := dbpkg.UpsertSuperAIModeloPreferido(c.base.dbSuper, adminEmail, model.ID, adminEmail); err != nil {
		http.Error(w, "No se pudo registrar modelo preferido", http.StatusInternalServerError)
		return
	}
	_, err = dbpkg.RegisterSuperAIConsulta(c.base.dbSuper, dbpkg.SuperAIConsulta{
		AdminEmail:       adminEmail,
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
		Observaciones:    "consulta desde chat_con_ia_global super",
	})
	if err != nil {
		http.Error(w, "No se pudo registrar auditoria de consulta", http.StatusInternalServerError)
		return
	}

	usoActualizado, err := dbpkg.GetSuperAIUsoDiario(c.base.dbSuper, adminEmail, model.Provider, model.ID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo obtener uso actualizado", http.StatusInternalServerError)
		return
	}
	restante := effectiveLimit - usoActualizado.Consultas
	if restante < 0 {
		restante = 0
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"admin_email": adminEmail,
		"provider":    model.Provider,
		"model_id":    model.ID,
		"respuesta":   respuesta,
		"usage": map[string]interface{}{
			"plan":              planActual,
			"daily_used":        usoActualizado.Consultas,
			"daily_limit":       effectiveLimit,
			"daily_remaining":   restante,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		},
		"scope": map[string]interface{}{
			"global_superadministrador": true,
			"restricted_by_super_role":  true,
		},
	})
}

// ConsultarConAdjuntoHandler permite enviar una imagen/documento al chat global super usando GPT-5.5.
func (c *SuperAIChatController) ConsultarConAdjuntoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.base.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.base.dbSuper),
		})
		return
	}

	adminEmail, ok := c.requireSuperAdmin(w, r)
	if !ok {
		return
	}

	superChatEnabled, _, _, err := getChatIASuperEnabled(c.base.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de chat IA", http.StatusInternalServerError)
		return
	}
	if !superChatEnabled {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_super_chat_disabled",
			"error": "El chat global de super administrador está desactivado desde configuración lógica del chat con IA.",
		})
		return
	}

	att, err := parseSingleAttachmentFromMultipart(r, "file", 8<<20)
	if err != nil {
		http.Error(w, "adjunto inválido: "+err.Error(), http.StatusBadRequest)
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

	catalog := availableEmpresaAIModelMap(c.base.dbSuper)
	model, okModel := catalog["openai:gpt-5.5"]
	if !okModel {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":    false,
			"code":  "ai_gpt55_unavailable",
			"error": "GPT-5.5 no está disponible o el proveedor OpenAI está deshabilitado.",
		})
		return
	}

	empresaRO, _, _, err := getChatIAEmpresaSoloLectura(c.base.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de contexto IA", http.StatusInternalServerError)
		return
	}

	contexto, err := dbpkg.BuildSuperAIContextoForQuestion(c.base.dbEmp, c.base.dbSuper, adminEmail, pregunta, dbpkg.SuperAIContextoOpts{
		EmpresaSoloLectura: empresaRO,
	})
	if err != nil {
		http.Error(w, "No se pudo construir contexto global", http.StatusBadRequest)
		return
	}

	preguntaFinal := pregunta
	if att != nil && strings.TrimSpace(att.Filename) != "" {
		preguntaFinal = "Adjunto: " + strings.TrimSpace(att.Filename) + " (" + strings.TrimSpace(att.MimeType) + ")\n\n" + pregunta
	}

	superMeta := c.base.dbSuper != nil
	respuesta, promptTokens, completionTokens, err := c.base.generateResponseWithSystemPromptAndAttachment(model, preguntaFinal, historial, buildSuperAISystemPrompt(contexto, superMeta, empresaRO), att)
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

	fechaUso := time.Now().Format("2006-01-02")
	usoActualizado, _ := dbpkg.GetSuperAIUsoDiario(c.base.dbSuper, adminEmail, model.Provider, model.ID, fechaUso)
	planActual := strings.ToLower(strings.TrimSpace(usoActualizado.PlanActual))
	if planActual == "" {
		planActual = "free"
	}

	_, _ = dbpkg.RegisterSuperAIConsulta(c.base.dbSuper, dbpkg.SuperAIConsulta{
		AdminEmail:       adminEmail,
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
		Observaciones:    "consulta con adjunto/gpt55 desde chat_con_ia_global super",
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"admin_email": adminEmail,
		"provider":    model.Provider,
		"model_id":    model.ID,
		"respuesta":   respuesta,
		"usage": map[string]interface{}{
			"plan":              planActual,
			"prompt_tokens":     promptTokens,
			"completion_tokens": completionTokens,
		},
	})
}

func (c *SuperAIChatController) ConsultarStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.base.dbSuper) {
		http.Error(w, "IA desactivada", http.StatusServiceUnavailable)
		return
	}
	adminEmail, ok := c.requireSuperAdmin(w, r)
	if !ok {
		return
	}
	enabled, _, _, err := getChatIASuperStreamingEnabled(c.base.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración streaming", http.StatusInternalServerError)
		return
	}
	if !enabled {
		http.Error(w, "Streaming desactivado", http.StatusServiceUnavailable)
		return
	}

	var payload superAIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	payload.ModelID = strings.TrimSpace(payload.ModelID)
	if payload.ModelID == "" {
		payload.ModelID = firstAvailableEmpresaAIModelID(c.base.dbSuper)
	}
	catalog := availableEmpresaAIModelMap(c.base.dbSuper)
	model, found := catalog[payload.ModelID]
	if !found {
		http.Error(w, "model_id no soportado o desactivado", http.StatusBadRequest)
		return
	}
	if !strings.Contains(strings.ToLower(model.Endpoint), "/v1/chat/completions") {
		http.Error(w, "modelo no soporta streaming", http.StatusBadRequest)
		return
	}

	payload.Pregunta = strings.TrimSpace(payload.Pregunta)
	if payload.Pregunta == "" {
		http.Error(w, "pregunta es obligatoria", http.StatusBadRequest)
		return
	}

	superChatEnabled, _, _, err := getChatIASuperEnabled(c.base.dbSuper)
	if err != nil || !superChatEnabled {
		http.Error(w, "chat super desactivado", http.StatusServiceUnavailable)
		return
	}

	superMaxConsultas, _, _, err := getChatIASuperMaxConsultasDia(c.base.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar límites IA", http.StatusInternalServerError)
		return
	}
	fechaUso := time.Now().Format("2006-01-02")
	usoActual, err := dbpkg.GetSuperAIUsoDiario(c.base.dbSuper, adminEmail, model.Provider, model.ID, fechaUso)
	if err != nil {
		http.Error(w, "No se pudo consultar uso diario", http.StatusInternalServerError)
		return
	}
	effectiveLimit := effectiveDailyLimitBySuperConfig(superMaxConsultas, model.FreeDailyLimit)
	if effectiveLimit == 0 || usoActual.Consultas >= effectiveLimit {
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_ = sseWriteJSON(w, openAIStreamEvent{Error: "Límite de uso alcanzado."})
		_ = sseWriteJSON(w, openAIStreamEvent{Done: true})
		return
	}

	empresaRO, _, _, err := getChatIAEmpresaSoloLectura(c.base.dbSuper)
	if err != nil {
		http.Error(w, "No se pudo consultar configuración de contexto IA", http.StatusInternalServerError)
		return
	}
	contexto, err := dbpkg.BuildSuperAIContextoForQuestion(c.base.dbEmp, c.base.dbSuper, adminEmail, payload.Pregunta, dbpkg.SuperAIContextoOpts{
		EmpresaSoloLectura: empresaRO,
	})
	if err != nil {
		http.Error(w, "No se pudo construir contexto global", http.StatusBadRequest)
		return
	}
	superMeta := c.base.dbSuper != nil
	systemPrompt := buildSuperAISystemPrompt(contexto, superMeta, empresaRO)

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var full strings.Builder
	_, err = c.base.callOpenAIStreamChatCompletions(model, payload.Pregunta, payload.Historial, systemPrompt, func(delta string) {
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
	_, _ = dbpkg.RegisterSuperAIConsulta(c.base.dbSuper, dbpkg.SuperAIConsulta{
		AdminEmail:       adminEmail,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         payload.Pregunta,
		Respuesta:        text,
		PromptTokens:     0,
		CompletionTokens: 0,
		TotalTokens:      0,
		FechaConsulta:    time.Now().Format("2006-01-02 15:04:05"),
		PlanActual:       planActual,
		UsuarioCreador:   adminEmail,
		Estado:           "activo",
		Observaciones:    "consulta streaming desde chat_con_ia_global super",
	})
	_ = sseWriteJSON(w, openAIStreamEvent{Done: true})
}

func (c *SuperAIChatController) HistorialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.base.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(c.base.dbSuper),
		})
		return
	}

	adminEmail, ok := c.requireSuperAdmin(w, r)
	if !ok {
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

	rows, err := dbpkg.ListSuperAIConsultasRecientes(c.base.dbSuper, adminEmail, limit)
	if err != nil {
		http.Error(w, "No se pudo consultar historial", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":          true,
		"admin_email": adminEmail,
		"items":       rows,
	})
}

func buildSuperAISystemPrompt(contexto string, superEsquemaCompleto, empresaSoloLectura bool) string {
	extra := ""
	if superEsquemaCompleto {
		extra += "El contexto incluye inventario de la base superadministrador (conteos por tabla, columnas nombre:tipo, reparto de administradores por rol). No se inyectan valores de fila ni secretos. No asumas que puedes ejecutar SQL ni acceder fuera de lo resumido. "
	}
	if empresaSoloLectura {
		extra += "Datos de empresas en el contexto provienen solo de consultas de lectura en el servidor. No sugieras ni describas operaciones de escritura, UPDATE, DELETE, ni PCS_ACTION que modifiquen datos de negocio; limítate a analizar o explicar lo mostrado. "
	} else {
		extra += "Si el usuario pide ejecutar acciones operativas (por ejemplo crear productos en una empresa, ajustar precios o registrar egresos), NO ejecutes nada directamente. " +
			"Las acciones solo se materializan como bloque PCS_ACTION tras confirmacion humana en el hilo. Regla obligatoria: " +
			"NUNCA incluyas el bloque PCS_ACTION en la primera respuesta donde propones o describes un cambio. " +
			"En ese turno solo explica el impacto, lista que haria cada llamada (endpoint, metodo, datos clave) y termina preguntando de forma explicita si el usuario confirma aplicar esos cambios. " +
			"Unicamente en un turno POSTERIOR, cuando el ultimo mensaje del usuario sea una confirmacion explicita (por ejemplo: si, confirmo, de acuerdo, adelante, procede, ejecuta) referida a esa propuesta ya aclarada, puedes incluir al FINAL el bloque literal con prefijo EXACTO `PCS_ACTION` y JSON valido (mismo formato que el chat empresarial: version, actions, note). " +
			"Regla de seguridad: NO propongas acciones de eliminacion (DELETE) ni operaciones destructivas; limita PCS_ACTION a GET/OPEN/POST/PUT. " +
			"Si el historial no muestra que el usuario confirmo tras tu pregunta, o si la intencion sigue ambigua, no emitas PCS_ACTION: aclara o vuelve a pedir confirmacion. " +
			"Si falta cualquier dato obligatorio, pregunta primero y NO emitas PCS_ACTION. " +
			"Si la operacion es riesgosa o destructiva, pide confirmacion adicional antes de cualquier PCS_ACTION. "
	}
	return "Eres un asistente global del sistema POS multiempresa para uso exclusivo de super administracion. " +
		"Responde en espanol claro y accionable. Usa solo el contexto agregado validado del sistema completo. " +
		"No reveles secretos, credenciales, hashes, tokens, llaves privadas ni datos sensibles. " +
		extra + "\n\n" +
		"Si existe la seccion CONSULTAS_SEGURAS_GLOBALES_RESUELTAS, priorizala como fuente principal para responder la pregunta actual. " +
		"Si faltan datos, dilo explicitamente y sugiere el siguiente reporte o consulta a revisar.\n\nCONTEXTO_GLOBAL_VALIDADO:\n" + contexto
}

func (c *SuperAIChatController) writeLimitReached(w http.ResponseWriter, model empresaAIModelDef, used int64) {
	writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
		"ok":              false,
		"code":            "free_tier_limit_reached",
		"error":           "Se alcanzo el limite del plan gratuito para este modelo.",
		"model_id":        model.ID,
		"provider":        model.Provider,
		"daily_used":      used,
		"daily_limit":     model.FreeDailyLimit,
		"upgrade_url":     "/super/licencias.html",
		"upgrade_message": "Puedes ajustar licencias o configuracion global para ampliar capacidad.",
	})
}

func (c *SuperAIChatController) requireSuperAdmin(w http.ResponseWriter, r *http.Request) (string, bool) {
	if c.base.dbSuper == nil {
		http.Error(w, "super db no disponible", http.StatusInternalServerError)
		return "", false
	}
	if adminEmail := strings.TrimSpace(adminEmailFromRequest(r)); adminEmail != "" && !strings.EqualFold(adminEmail, "sistema") {
		role := strings.TrimSpace(adminRoleFromRequest(r))
		if paginaPrincipalRoleIsSuper(role) {
			return strings.ToLower(adminEmail), true
		}
	}
	return paginaPrincipalRequireSuperAdmin(w, r, c.base.dbSuper)
}
