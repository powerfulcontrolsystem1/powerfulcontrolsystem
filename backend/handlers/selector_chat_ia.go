package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

// SelectorAIChatController is deliberately separate from the super-admin chat.
// It builds a read-only aggregate exclusively from companies the signed-in user
// can already open. Client-provided company IDs are never accepted here.
type SelectorAIChatController struct {
	base *EmpresaAIChatController
}

func NewSelectorAIChatController(dbEmp, dbSuper *sql.DB) *SelectorAIChatController {
	return &SelectorAIChatController{base: NewEmpresaAIChatController(dbEmp, dbSuper)}
}

func (c *SelectorAIChatController) currentAccount(r *http.Request) (string, error) {
	admin, err := currentAdminFromSession(r, c.base.dbSuper)
	if err != nil || admin == nil || strings.TrimSpace(admin.Email) == "" {
		return "", fmt.Errorf("unauthenticated")
	}
	return strings.ToLower(strings.TrimSpace(admin.Email)), nil
}

func (c *SelectorAIChatController) accessibleEmpresaIDs(account string) ([]int64, error) {
	rows, err := c.base.dbEmp.Query(`SELECT id FROM empresas WHERE COALESCE(estado, 'activo') <> 'inactivo' ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := make([]int64, 0, 8)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		allowed, err := dbpkg.CanAdminAccessEmpresaIA(c.base.dbEmp, c.base.dbSuper, account, id)
		if err != nil {
			return nil, err
		}
		if allowed {
			ids = append(ids, id)
		}
		if len(ids) >= 25 { // bounded prompt and predictable cost.
			break
		}
	}
	return ids, rows.Err()
}

func selectorAIContext(contexts []string) string {
	var out strings.Builder
	out.WriteString("CONTEXTO_SELECTOR_EMPRESAS_AUTORIZADAS\n")
	out.WriteString("Solo contiene resúmenes de empresas a las que el usuario autenticado ya tiene acceso. No contiene identificadores técnicos, NIT, secretos ni permisos de escritura.\n")
	for _, context := range contexts {
		lines := strings.Split(context, "\n")
		var kept []string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			lower := strings.ToLower(trimmed)
			if trimmed == "" || strings.HasPrefix(lower, "- empresa_id:") || strings.HasPrefix(lower, "- nit:") || strings.HasPrefix(lower, "- tablas_contexto_disponibles:") {
				continue
			}
			kept = append(kept, trimmed)
		}
		if len(kept) == 0 {
			continue
		}
		out.WriteString("\nEMPRESA_AUTORIZADA\n")
		out.WriteString(strings.Join(kept, "\n"))
		out.WriteString("\n")
	}
	return out.String()
}

func (c *SelectorAIChatController) contextForAccount(account string) (int64, string, int, error) {
	ids, err := c.accessibleEmpresaIDs(account)
	if err != nil {
		return 0, "", 0, err
	}
	if len(ids) == 0 {
		return 0, "", 0, fmt.Errorf("no hay empresas activas autorizadas")
	}
	contexts := make([]string, 0, len(ids))
	for _, empresaID := range ids {
		context, err := dbpkg.BuildEmpresaAIContexto(c.base.dbEmp, empresaID)
		if err != nil {
			return 0, "", 0, err
		}
		contexts = append(contexts, context)
	}
	return ids[0], selectorAIContext(contexts), len(ids), nil
}

func (c *SelectorAIChatController) ModelosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	account, err := c.currentAccount(r)
	if err != nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	primaryID, _, count, err := c.contextForAccount(account)
	if err != nil {
		http.Error(w, "No hay empresas autorizadas para el chat del selector", http.StatusForbidden)
		return
	}
	catalog := availableEmpresaAIModelCatalog(c.base.dbSuper)
	available := availableEmpresaAIModelMap(c.base.dbSuper)
	prefs, _ := dbpkg.GetEmpresaAIUsuarioPreferencias(c.base.dbEmp, account)
	preferred := prefs.ModelID
	if _, ok := available[preferred]; !ok {
		preferred = firstAvailableEmpresaAIModelID(c.base.dbSuper)
	}
	max, _, _, _ := getChatIAEmpresaMaxConsultasDia(c.base.dbSuper)
	today := time.Now().Format("2006-01-02")
	items := make([]map[string]interface{}, 0, len(catalog))
	for _, model := range catalog {
		usage, _ := dbpkg.GetEmpresaAIUsoDiario(c.base.dbEmp, primaryID, model.Provider, model.ID, today)
		limit := effectiveDailyLimitBySuperConfig(max, model.FreeDailyLimit)
		remaining := limit - usage.Consultas
		if remaining < 0 {
			remaining = 0
		}
		items = append(items, map[string]interface{}{
			"id": model.ID, "provider": model.Provider, "display_name": model.DisplayName,
			"upstream_model": model.UpstreamModel, "description": model.Description,
			"usage": map[string]interface{}{"daily_used": usage.Consultas, "daily_limit": limit, "daily_remaining": remaining},
		})
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "scope": "selector_empresas_autorizadas", "empresa_count": count,
		"modelo_preferido": preferred, "modo_preferido": normalizeAIAssistantMode(prefs.ModoAsistente),
		"streaming_enabled": false, "attachments_enabled": false, "modelos": items,
	})
}

func (c *SelectorAIChatController) ModeloPreferidoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	account, err := c.currentAccount(r)
	if err != nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	var payload struct {
		ModelID       string `json:"model_id"`
		ModoAsistente string `json:"modo_asistente"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	payload.ModelID = strings.TrimSpace(payload.ModelID)
	if _, ok := availableEmpresaAIModelMap(c.base.dbSuper)[payload.ModelID]; !ok {
		http.Error(w, "model_id no soportado o desactivado", http.StatusBadRequest)
		return
	}
	if err := dbpkg.UpsertEmpresaAIUsuarioPreferencias(c.base.dbEmp, account, payload.ModelID, normalizeAIAssistantMode(payload.ModoAsistente), "general", account); err != nil {
		http.Error(w, "No se pudo guardar la preferencia", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "scope": "selector_empresas_autorizadas", "model_id": payload.ModelID})
}

func (c *SelectorAIChatController) ConsultarHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !isSuperAIEnabled(c.base.dbSuper) {
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"ok": false, "code": "ai_disabled", "error": "La IA está desactivada desde configuración avanzada."})
		return
	}
	account, err := c.currentAccount(r)
	if err != nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}
	var payload superAIChatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 128<<10)).Decode(&payload); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}
	payload.Pregunta = strings.TrimSpace(payload.Pregunta)
	if payload.Pregunta == "" || len([]rune(payload.Pregunta)) > 2500 {
		http.Error(w, "pregunta invalida", http.StatusBadRequest)
		return
	}
	modelID := strings.TrimSpace(payload.ModelID)
	if modelID == "" {
		prefs, _ := dbpkg.GetEmpresaAIUsuarioPreferencias(c.base.dbEmp, account)
		modelID = prefs.ModelID
	}
	if modelID == "" {
		modelID = firstAvailableEmpresaAIModelID(c.base.dbSuper)
	}
	model, ok := availableEmpresaAIModelMap(c.base.dbSuper)[modelID]
	if !ok {
		http.Error(w, "model_id no soportado o desactivado", http.StatusBadRequest)
		return
	}
	primaryID, context, count, err := c.contextForAccount(account)
	if err != nil {
		http.Error(w, "No se pudo construir el contexto autorizado", http.StatusForbidden)
		return
	}
	max, _, _, _ := getChatIAEmpresaMaxConsultasDia(c.base.dbSuper)
	today := time.Now().Format("2006-01-02")
	usage, err := dbpkg.GetEmpresaAIUsoDiario(c.base.dbEmp, primaryID, model.Provider, model.ID, today)
	if err != nil {
		http.Error(w, "No se pudo consultar el límite de IA", http.StatusInternalServerError)
		return
	}
	limit := effectiveDailyLimitBySuperConfig(max, model.FreeDailyLimit)
	if limit == 0 || usage.Consultas >= limit {
		c.base.writeLimitReached(w, primaryID, model, usage.Consultas)
		return
	}
	system := "Eres el asistente del selector de empresas de PCS. Responde solo lo preguntado, en español claro, usando exclusivamente los resúmenes autorizados. No ejecutes acciones, no inventes empresas ni reveles IDs, NIT, secretos, credenciales o información de empresas no incluidas. Cuando el usuario pida operar una empresa, indícale que debe abrirla primero.\n\n" + buildAIAssistantModeInstruction(payload.ModoAsistente) + "\n\n" + context
	answer, promptTokens, completionTokens, err := c.base.generateResponseWithSystemPrompt(model, payload.Pregunta, payload.Historial, system)
	if err != nil {
		http.Error(w, publicAIProviderError(err), http.StatusBadGateway)
		return
	}
	answer = strings.TrimSpace(answer)
	if answer == "" {
		http.Error(w, "El proveedor no devolvio contenido", http.StatusBadGateway)
		return
	}
	_ = dbpkg.UpsertEmpresaAIUsuarioPreferencias(c.base.dbEmp, account, model.ID, normalizeAIAssistantMode(payload.ModoAsistente), "general", account)
	_, err = dbpkg.RegisterEmpresaAIConsulta(c.base.dbEmp, dbpkg.EmpresaAIConsulta{
		EmpresaID: primaryID, Provider: model.Provider, ModelID: model.ID, Pregunta: payload.Pregunta, Respuesta: answer,
		PromptTokens: promptTokens, CompletionTokens: completionTokens, TotalTokens: promptTokens + completionTokens,
		FechaConsulta: time.Now().Format("2006-01-02 15:04:05"), PlanActual: "selector_compartido", UsuarioCreador: account,
		Estado: "activo", Observaciones: "consulta selector multiempresa autorizada",
	})
	if err != nil {
		http.Error(w, "No se pudo registrar el consumo de IA", http.StatusInternalServerError)
		return
	}
	updated, _ := dbpkg.GetEmpresaAIUsoDiario(c.base.dbEmp, primaryID, model.Provider, model.ID, today)
	remaining := limit - updated.Consultas
	if remaining < 0 {
		remaining = 0
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "scope": "selector_empresas_autorizadas", "empresa_count": count,
		"provider": model.Provider, "model_id": model.ID, "display_name": model.DisplayName, "respuesta": answer,
		"usage": map[string]interface{}{"daily_used": updated.Consultas, "daily_limit": limit, "daily_remaining": remaining},
	})
}
