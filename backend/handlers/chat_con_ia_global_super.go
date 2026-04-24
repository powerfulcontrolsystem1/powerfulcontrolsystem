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
	ModelID     string                 `json:"model_id"`
	Pregunta    string                 `json:"pregunta"`
	Historial   []empresaAIChatMensaje `json:"historial"`
	Temperatura float64                `json:"temperatura"`
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

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok":               true,
		"admin_email":      adminEmail,
		"scope":            "global_superadministrador",
		"modelo_preferido": modeloPreferido,
		"modelos":          items,
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

	contexto, err := dbpkg.BuildSuperAIContextoForQuestion(c.base.dbEmp, c.base.dbSuper, adminEmail, payload.Pregunta)
	if err != nil {
		http.Error(w, "No se pudo construir contexto global", http.StatusBadRequest)
		return
	}

	respuesta, promptTokens, completionTokens, err := c.base.generateResponseWithSystemPrompt(model, payload.Pregunta, payload.Historial, buildSuperAISystemPrompt(contexto))
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

func buildSuperAISystemPrompt(contexto string) string {
	return "Eres un asistente global del sistema POS multiempresa para uso exclusivo de super administracion. " +
		"Responde en espanol claro y accionable. Usa solo el contexto agregado validado del sistema completo. " +
		"No reveles secretos, credenciales, hashes, tokens, llaves privadas ni datos sensibles. " +
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
