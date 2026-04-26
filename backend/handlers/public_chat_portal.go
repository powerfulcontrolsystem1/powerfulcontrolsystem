package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type publicPortalChatLimiter struct {
	mu      sync.Mutex
	records map[string]*publicPortalChatUsage
}

type publicPortalChatUsage struct {
	ResetAt time.Time
	Used    int
}

var portalChatLimiter = &publicPortalChatLimiter{records: map[string]*publicPortalChatUsage{}}

func portalChatClientIP(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); v != "" {
		parts := strings.Split(v, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if v := strings.TrimSpace(r.Header.Get("X-Real-IP")); v != "" {
		return v
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func portalChatGetOrSetClientID(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("pcs_public_chat_id"); err == nil {
		v := strings.TrimSpace(c.Value)
		if len(v) >= 16 && len(v) <= 80 {
			return v
		}
	}
	var b [16]byte
	_, _ = rand.Read(b[:])
	id := hex.EncodeToString(b[:])

	secure := r.TLS != nil
	if !secure {
		if strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") {
			secure = true
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "pcs_public_chat_id",
		Value:    id,
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		HttpOnly: true,
	})
	return id
}

func (l *publicPortalChatLimiter) allow(key string, max int, window time.Duration) (bool, int, int) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.records == nil {
		l.records = map[string]*publicPortalChatUsage{}
	}
	u := l.records[key]
	if u == nil || now.After(u.ResetAt) {
		u = &publicPortalChatUsage{ResetAt: now.Add(window), Used: 0}
		l.records[key] = u
	}
	if u.Used >= max {
		retry := int(time.Until(u.ResetAt).Seconds())
		if retry < 1 {
			retry = 1
		}
		return false, 0, retry
	}
	u.Used++
	remaining := max - u.Used
	retry := int(time.Until(u.ResetAt).Seconds())
	if retry < 0 {
		retry = 0
	}
	return true, remaining, retry
}

func buildPortalCompanySystemPrompt() string {
	return "Eres un asistente comercial del sitio web Powerful Control System. " +
		"Tu objetivo es responder SOLO preguntas sobre Powerful Control System, su plataforma SaaS POS multiempresa, módulos, precios/planes (mensual/anual), novedades, condiciones de pago, soporte, y cómo comprar/activar licencias. " +
		"No inventes precios exactos si no se te proporcionan; si no sabes un precio, indica que el precio depende del plan y ofrece el canal de contacto. " +
		"No pidas datos sensibles (contraseñas, tarjetas, tokens) y no reveles credenciales. " +
		"No hables de la base de datos, endpoints internos, ni ejecutes acciones (no PCS_ACTION). " +
		"Responde en español, breve, con bullets cuando convenga y siempre puedes sugerir WhatsApp o email para atención personalizada. " +
		"Regla obligatoria de cierre: termina tu respuesta con esta frase exacta: " +
		"\"No olvides que puedes probar ya mismo totalmente gratis el sistema con solo registrarte.\""
}

func portalChatLoadExtraInfo(dbSuper *sql.DB) string {
	if dbSuper == nil {
		return ""
	}
	raw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, superPortalChatIAInfoKey)
	return strings.TrimSpace(raw)
}

func portalChatBuildLicenciasPriceSummary(dbSuper *sql.DB) string {
	if dbSuper == nil {
		return ""
	}
	// Usamos licencias activas (catálogo) sin empresa asignada.
	lics, err := dbpkg.GetLicenciasFilteredByPais(dbSuper, true, "", false, "")
	if err != nil || len(lics) == 0 {
		return ""
	}
	// Resumen simple (evita inventar): nombre + duración + valor + país
	lines := make([]string, 0, len(lics))
	lines = append(lines, "Precios de licencias (según base de datos del sistema):")
	for _, l := range lics {
		pais := strings.ToUpper(strings.TrimSpace(l.PaisCodigo))
		if pais == "" {
			pais = "CO"
		}
		lines = append(lines, "- "+strings.TrimSpace(l.Nombre)+" ("+pais+") — "+strconv.Itoa(l.DuracionDias)+" días — valor: "+fmt.Sprintf("%.2f", l.Valor))
	}
	return strings.Join(lines, "\n")
}

func isPortalPublicChatEnabled(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return defaultChatIAPortalEnabled
	}
	v, _, _, err := getChatIAPortalEnabled(dbSuper)
	if err != nil {
		return defaultChatIAPortalEnabled
	}
	return v
}

func pickPortalChatModel(dbSuper *sql.DB) (empresaAIModelDef, bool) {
	modelMap := empresaAIModelMap()
	defs := aiCredentialCatalogModels()
	if len(defs) == 0 {
		return empresaAIModelDef{}, false
	}

	// Preferir OpenAI (si está habilitado/configurado). Si no, usar el primero disponible del catálogo.
	for _, wantProvider := range []string{"openai", "google"} {
		for _, d := range defs {
			if !strings.EqualFold(strings.TrimSpace(d.Provider), wantProvider) {
				continue
			}
			m, ok := modelMap[d.ModelID]
			if !ok {
				continue
			}
			if isAIProviderEnabled(dbSuper, m.Provider) {
				return m, true
			}
		}
	}
	for _, d := range defs {
		m, ok := modelMap[d.ModelID]
		if ok && isAIProviderEnabled(dbSuper, m.Provider) {
			return m, true
		}
	}
	return empresaAIModelDef{}, false
}

// PublicPortalCompanyChatHandler: chat público del index con límite por usuario.
// - 10 consultas por ventana de 5 minutos.
func PublicPortalCompanyChatHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if !isPortalPublicChatEnabled(dbSuper) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":    false,
				"code":  "chat_portal_disabled",
				"error": "El chat público del portal está deshabilitado.",
			})
			return
		}
		if !isSuperAIEnabled(dbSuper) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":    false,
				"code":  "ai_disabled",
				"error": "El chat está temporalmente deshabilitado.",
			})
			return
		}

		clientID := portalChatGetOrSetClientID(w, r)
		ip := portalChatClientIP(r)
		key := "portal:" + clientID + ":" + ip
		ok, remaining, retry := portalChatLimiter.allow(key, 10, 5*time.Minute)
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"ok":                   false,
				"code":                 "rate_limited",
				"error":                "Has alcanzado el límite de 10 preguntas. Espera unos 5 minutos e inténtalo de nuevo, o escríbenos por WhatsApp para atención más personalizada.",
				"retry_after_seconds":   retry,
				"remaining_in_window":   0,
				"window_seconds":        300,
				"public_contact_whatsapp": "https://wa.me/573043306506",
				"public_contact_email":  "powerfulcontrolsystem@gmail.com",
			})
			return
		}

		var body struct {
			Pregunta  string                `json:"pregunta"`
			Historial []empresaAIChatMensaje `json:"historial"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		p := strings.TrimSpace(body.Pregunta)
		if p == "" {
			http.Error(w, "pregunta es obligatoria", http.StatusBadRequest)
			return
		}
		if len(p) > 2000 {
			p = p[:2000]
		}

		model, okModel := pickPortalChatModel(dbSuper)
		if !okModel {
			writeJSON(w, http.StatusBadGateway, map[string]interface{}{
				"ok":    false,
				"code":  "ai_model_missing",
				"error": "Modelo de IA no disponible.",
			})
			return
		}

		ctrl := &EmpresaAIChatController{dbSuper: dbSuper, client: &http.Client{Timeout: 35 * time.Second}}
		systemPrompt := buildPortalCompanySystemPrompt()
		// Inyectar info editable + precios de licencias del catálogo.
		if extra := portalChatLoadExtraInfo(dbSuper); extra != "" {
			systemPrompt += "\n\n=== Información oficial editable (super administrador) ===\n" + extra
		}
		if pricing := portalChatBuildLicenciasPriceSummary(dbSuper); pricing != "" {
			systemPrompt += "\n\n=== Precios/planes desde base de datos ===\n" + pricing
		}
		h := sanitizeHistorial(body.Historial, 6)

		answer, _, _, err := ctrl.generateResponseWithSystemPrompt(model, p, h, systemPrompt)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]interface{}{
				"ok":                  false,
				"code":                "ai_provider_error",
				"error":               "No se pudo generar respuesta en este momento. Intenta nuevamente en unos minutos.",
				"remaining_in_window": remaining,
				"retry_after_seconds": 0,
			})
			return
		}
		answer = strings.TrimSpace(answer)
		if answer == "" {
			answer = "No pude generar una respuesta en este momento. Intenta de nuevo."
		}
		closing := "No olvides que puedes probar ya mismo totalmente gratis el sistema con solo registrarte."
		if !strings.Contains(strings.ToLower(answer), strings.ToLower(closing)) {
			answer = strings.TrimSpace(answer) + "\n\n" + closing
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                  true,
			"respuesta":           answer,
			"remaining_in_window": remaining,
			"window_seconds":      300,
		})
	}
}

