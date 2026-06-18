package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

var superMantenimientoAgentesRunMu sync.Mutex

type dianNewsCandidate struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Source string `json:"source"`
}

type dianNewsAssessment struct {
	Title        string `json:"title"`
	URL          string `json:"url"`
	Relevant     bool   `json:"relevant"`
	Summary      string `json:"summary"`
	Impact       string `json:"impact"`
	Relevance    string `json:"relevance"`
	PublishedAt  string `json:"published_at"`
	Source       string `json:"source"`
	ContentHash  string `json:"content_hash"`
	AssessedByAI bool   `json:"assessed_by_ai"`
}

func SuperMantenimientoAgentesHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if err := dbpkg.EnsureSuperMantenimientoAgentesSchema(dbSuper); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		action := strings.TrimSpace(r.URL.Query().Get("action"))
		switch r.Method {
		case http.MethodGet:
			agents, err := dbpkg.ListSuperMantenimientoAgentes(dbSuper)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
				return
			}
			findings, err := dbpkg.ListSuperMantenimientoHallazgos(dbSuper, dbpkg.SuperAgenteDIANNoticiasCodigo, 80)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "agents": agents, "findings": findings})
			return
		case http.MethodPut, http.MethodPost:
			if action == "run_now" {
				result := RunSuperDIANNoticiasAgent(dbSuper, adminEmail, true)
				status := http.StatusOK
				if !result["ok"].(bool) {
					status = http.StatusBadGateway
				}
				writeJSON(w, status, result)
				return
			}
			var payload struct {
				Codigo            string `json:"codigo"`
				Habilitado        bool   `json:"habilitado"`
				HoraEjecucion     string `json:"hora_ejecucion"`
				EmailNotificacion string `json:"email_notificacion"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(payload.Codigo) == "" {
				payload.Codigo = dbpkg.SuperAgenteDIANNoticiasCodigo
			}
			payload.HoraEjecucion = normalizeAgentHour(payload.HoraEjecucion)
			payload.EmailNotificacion = strings.TrimSpace(payload.EmailNotificacion)
			if payload.EmailNotificacion != "" {
				if _, err := mail.ParseAddress(payload.EmailNotificacion); err != nil {
					http.Error(w, "email_notificacion invalido", http.StatusBadRequest)
					return
				}
			}
			err := dbpkg.UpsertSuperMantenimientoAgente(dbSuper, dbpkg.SuperMantenimientoAgente{
				Codigo:            payload.Codigo,
				Nombre:            "Noticias DIAN",
				Descripcion:       "Revisa noticias oficiales de la DIAN y alerta cambios relevantes para el sistema.",
				Habilitado:        payload.Habilitado,
				HoraEjecucion:     payload.HoraEjecucion,
				EmailNotificacion: payload.EmailNotificacion,
				UsuarioCreador:    adminEmail,
				Estado:            "activo",
			})
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func StartSuperMantenimientoAgentesWorker(dbSuper *sql.DB, interval time.Duration, stop <-chan struct{}) {
	if dbSuper == nil {
		return
	}
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				runScheduledSuperMaintenanceAgents(dbSuper)
			case <-stop:
				return
			}
		}
	}()
}

func runScheduledSuperMaintenanceAgents(dbSuper *sql.DB) {
	agent, err := dbpkg.GetSuperMantenimientoAgente(dbSuper, dbpkg.SuperAgenteDIANNoticiasCodigo)
	if err != nil || agent == nil || !agent.Habilitado {
		return
	}
	now := time.Now()
	if strings.TrimSpace(agent.UltimaEjecucionDia) == now.Format("2006-01-02") {
		return
	}
	hour := normalizeAgentHour(agent.HoraEjecucion)
	if now.Format("15:04") < hour {
		return
	}
	RunSuperDIANNoticiasAgent(dbSuper, "sistema", false)
}

func RunSuperDIANNoticiasAgent(dbSuper *sql.DB, actor string, manual bool) map[string]any {
	superMantenimientoAgentesRunMu.Lock()
	defer superMantenimientoAgentesRunMu.Unlock()

	agent, err := dbpkg.GetSuperMantenimientoAgente(dbSuper, dbpkg.SuperAgenteDIANNoticiasCodigo)
	if err != nil {
		return map[string]any{"ok": false, "error": err.Error()}
	}
	candidates, fetchErr := fetchDIANNewsCandidates()
	if fetchErr != nil {
		_ = dbpkg.UpdateSuperMantenimientoAgenteRun(dbSuper, agent.Codigo, "error", fetchErr.Error(), time.Now())
		return map[string]any{"ok": false, "error": fetchErr.Error()}
	}
	assessments := assessDIANNewsWithAI(dbSuper, actor, candidates)
	newCount := 0
	for _, a := range assessments {
		if !a.Relevant {
			continue
		}
		created, _, err := dbpkg.CreateSuperMantenimientoHallazgoIfNew(dbSuper, dbpkg.SuperMantenimientoHallazgo{
			AgenteCodigo:     agent.Codigo,
			Titulo:           a.Title,
			URL:              a.URL,
			Fuente:           a.Source,
			Resumen:          a.Summary,
			ImpactoSistema:   a.Impact,
			Relevancia:       a.Relevance,
			FechaPublicacion: a.PublishedAt,
			HashContenido:    a.ContentHash,
			UsuarioCreador:   actor,
			Estado:           "nuevo",
		})
		if err != nil {
			log.Printf("[super_mantenimiento_agentes] no se pudo guardar hallazgo DIAN: %v", err)
			continue
		}
		if created {
			newCount++
		}
	}
	obs := fmt.Sprintf("candidatos=%d relevantes=%d nuevos=%d manual=%t", len(candidates), countRelevantAssessments(assessments), newCount, manual)
	_ = dbpkg.UpdateSuperMantenimientoAgenteRun(dbSuper, agent.Codigo, "ok", obs, time.Now())
	if newCount > 0 && strings.TrimSpace(agent.EmailNotificacion) != "" {
		if err := sendDIANAgentNotificationEmail(dbSuper, agent.EmailNotificacion, assessments); err != nil {
			log.Printf("[super_mantenimiento_agentes] no se pudo enviar correo DIAN: %v", err)
		}
	}
	return map[string]any{"ok": true, "candidates": len(candidates), "relevant": countRelevantAssessments(assessments), "new_findings": newCount}
}

func fetchDIANNewsCandidates() ([]dianNewsCandidate, error) {
	sources := []string{
		"https://www.dian.gov.co/Prensa/Paginas/Noticias.aspx",
		"https://www.dian.gov.co/Prensa/Paginas/Comunicados-de-Prensa.aspx",
		"https://micrositios.dian.gov.co/sistema-de-facturacion-electronica/",
	}
	client := &http.Client{Timeout: 25 * time.Second}
	seen := map[string]bool{}
	out := []dianNewsCandidate{}
	var lastErr error
	for _, source := range sources {
		req, _ := http.NewRequest(http.MethodGet, source, nil)
		req.Header.Set("User-Agent", "PowerfulControlSystem-DIAN-Agent/1.0")
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 700000))
		_ = resp.Body.Close()
		if err != nil || resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("DIAN %s HTTP %d", source, resp.StatusCode)
			continue
		}
		for _, item := range extractDIANLinks(source, string(body)) {
			key := strings.ToLower(strings.TrimSpace(item.URL + "|" + item.Title))
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, item)
			if len(out) >= 40 {
				return out, nil
			}
		}
	}
	if len(out) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return out, nil
}

var dianHrefRE = regexp.MustCompile(`(?is)<a[^>]+href=["']([^"']+)["'][^>]*>(.*?)</a>`)
var htmlTagRE = regexp.MustCompile(`(?is)<[^>]+>`)

func extractDIANLinks(sourceURL, raw string) []dianNewsCandidate {
	matches := dianHrefRE.FindAllStringSubmatch(raw, 160)
	out := []dianNewsCandidate{}
	for _, m := range matches {
		title := strings.TrimSpace(html.UnescapeString(htmlTagRE.ReplaceAllString(m[2], " ")))
		title = strings.Join(strings.Fields(title), " ")
		if len([]rune(title)) < 12 || len([]rune(title)) > 180 {
			continue
		}
		url := normalizeDIANURL(sourceURL, m[1])
		folded := strings.ToLower(title + " " + url)
		if !strings.Contains(folded, "dian") && !dianLooksRelevant(folded) {
			continue
		}
		out = append(out, dianNewsCandidate{Title: title, URL: url, Source: sourceURL})
	}
	return out
}

func normalizeDIANURL(baseURL, href string) string {
	href = strings.TrimSpace(html.UnescapeString(href))
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	if strings.HasPrefix(href, "/") {
		if strings.Contains(baseURL, "micrositios.dian.gov.co") {
			return "https://micrositios.dian.gov.co" + href
		}
		return "https://www.dian.gov.co" + href
	}
	idx := strings.LastIndex(baseURL, "/")
	if idx > 8 {
		return baseURL[:idx+1] + href
	}
	return href
}

func assessDIANNewsWithAI(dbSuper *sql.DB, actor string, candidates []dianNewsCandidate) []dianNewsAssessment {
	out := fallbackDIANAssessments(candidates)
	if dbSuper == nil || !isSuperAIEnabled(dbSuper) || len(candidates) == 0 {
		return out
	}
	catalog := availableEmpresaAIModelMap(dbSuper)
	var model empresaAIModelDef
	found := false
	for _, it := range catalog {
		if strings.EqualFold(it.Provider, "openai") {
			model = it
			found = true
			break
		}
	}
	if !found {
		return out
	}
	payload, _ := json.Marshal(candidates)
	prompt := "Analiza estas noticias/enlaces oficiales de la DIAN para Powerful Control System. Devuelve SOLO JSON con un arreglo llamado items. Marca relevant=true solo si afecta facturacion electronica Colombia, impuestos, documentos electronicos, resoluciones, calendario tributario, seguridad o integraciones que requieran revisar el sistema. Campos por item: title,url,relevant,summary,impact,relevance,published_at.\n\n" + string(payload)
	ctrl := &EmpresaAIChatController{dbSuper: dbSuper, client: &http.Client{Timeout: 45 * time.Second}}
	resp, promptTokens, completionTokens, err := ctrl.generateResponseWithSystemPrompt(model, prompt, nil, "Eres auditor tecnico tributario Colombia. Responde JSON estricto, sin markdown.")
	if err != nil {
		return out
	}
	_, _ = dbpkg.RegisterSuperAIConsulta(dbSuper, dbpkg.SuperAIConsulta{
		AdminEmail:       actor,
		Provider:         model.Provider,
		ModelID:          model.ID,
		Pregunta:         "Agente DIAN: clasificacion de noticias oficiales",
		Respuesta:        truncateAgentText(resp, 4000),
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
		FechaConsulta:    time.Now().Format("2006-01-02 15:04:05"),
		PlanActual:       "super",
		UsuarioCreador:   actor,
		Estado:           "activo",
		Observaciones:    "agente_mantenimiento_dian",
	})
	parsed := parseDIAssessmentJSON(resp, candidates)
	if len(parsed) == 0 {
		return out
	}
	return parsed
}

func fallbackDIANAssessments(candidates []dianNewsCandidate) []dianNewsAssessment {
	out := make([]dianNewsAssessment, 0, len(candidates))
	for _, c := range candidates {
		folded := strings.ToLower(c.Title + " " + c.URL)
		relevant := dianLooksRelevant(folded)
		out = append(out, dianNewsAssessment{
			Title:       c.Title,
			URL:         c.URL,
			Source:      c.Source,
			Relevant:    relevant,
			Summary:     c.Title,
			Impact:      "Revisar si exige ajuste operativo o normativo en el sistema.",
			Relevance:   "media",
			ContentHash: hashDIANContent(c.Title, c.URL),
		})
	}
	return out
}

func dianLooksRelevant(v string) bool {
	keys := []string{"factura", "facturacion", "electronica", "documento equivalente", "nomina electronica", "resolucion", "tribut", "impuesto", "radian", "ubl", "dian", "calendario"}
	for _, k := range keys {
		if strings.Contains(v, k) {
			return true
		}
	}
	return false
}

func parseDIAssessmentJSON(raw string, candidates []dianNewsCandidate) []dianNewsAssessment {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return nil
	}
	var payload struct {
		Items []dianNewsAssessment `json:"items"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &payload); err != nil {
		return nil
	}
	byURL := map[string]dianNewsCandidate{}
	for _, c := range candidates {
		byURL[strings.TrimSpace(c.URL)] = c
	}
	out := []dianNewsAssessment{}
	for _, item := range payload.Items {
		item.Title = strings.TrimSpace(item.Title)
		item.URL = strings.TrimSpace(item.URL)
		if item.Title == "" && item.URL != "" {
			item.Title = byURL[item.URL].Title
		}
		if item.Source == "" && item.URL != "" {
			item.Source = byURL[item.URL].Source
		}
		if item.ContentHash == "" {
			item.ContentHash = hashDIANContent(item.Title, item.URL)
		}
		item.AssessedByAI = true
		if item.Relevance == "" {
			item.Relevance = "media"
		}
		if item.Title != "" {
			out = append(out, item)
		}
	}
	return out
}

func hashDIANContent(title, url string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(title)) + "|" + strings.TrimSpace(url)))
	return hex.EncodeToString(sum[:])
}

func countRelevantAssessments(items []dianNewsAssessment) int {
	total := 0
	for _, item := range items {
		if item.Relevant {
			total++
		}
	}
	return total
}

func sendDIANAgentNotificationEmail(dbSuper *sql.DB, to string, items []dianNewsAssessment) error {
	lines := []string{"El agente de mantenimiento DIAN detecto noticias relevantes:", ""}
	count := 0
	for _, item := range items {
		if !item.Relevant {
			continue
		}
		count++
		lines = append(lines, "- "+item.Title)
		if strings.TrimSpace(item.URL) != "" {
			lines = append(lines, "  "+item.URL)
		}
		if strings.TrimSpace(item.Impact) != "" {
			lines = append(lines, "  Impacto: "+item.Impact)
		}
	}
	if count == 0 {
		return nil
	}
	return sendEmpresaUsuarioGmailPlain(dbSuper, to, "Alerta DIAN - Agente de mantenimiento PCS", strings.Join(lines, "\r\n"))
}

func normalizeAgentHour(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 5 {
		v = v[:5]
	}
	if _, err := time.Parse("15:04", v); err == nil {
		return v
	}
	return "07:00"
}

func truncateAgentText(v string, limit int) string {
	r := []rune(strings.TrimSpace(v))
	if len(r) <= limit {
		return string(r)
	}
	return string(r[:limit])
}
