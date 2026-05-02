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
	if !secure && strings.EqualFold(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")), "https") {
		secure = true
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
		"Tu objetivo es responder SOLO preguntas sobre Powerful Control System, su plataforma SaaS POS multiempresa, modulos, precios o planes, novedades, condiciones de pago, soporte, y como comprar o activar licencias. " +
		"No inventes precios exactos si no se te proporcionan; si no sabes un precio, indica que depende del plan y ofrece el canal de contacto. " +
		"No pidas datos sensibles ni reveles credenciales. " +
		"No hables de la base de datos, endpoints internos, ni ejecutes acciones. " +
		"Responde en espanol, breve, y puedes sugerir WhatsApp o email para atencion personalizada. " +
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
	lics, err := dbpkg.GetLicenciasFilteredByPais(dbSuper, true, "", false, "")
	if err != nil || len(lics) == 0 {
		return ""
	}
	lines := make([]string, 0, len(lics)+1)
	lines = append(lines, "Precios de licencias segun base de datos del sistema:")
	for _, l := range lics {
		pais := strings.ToUpper(strings.TrimSpace(l.PaisCodigo))
		if pais == "" {
			pais = "CO"
		}
		lines = append(lines, "- "+strings.TrimSpace(l.Nombre)+" ("+pais+") - "+strconv.Itoa(l.DuracionDias)+" dias - valor: "+fmt.Sprintf("%.2f", l.Valor))
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

func buildPortalPublicStoreSystemPrompt(cfg dbpkg.EmpresaVentaPublicaConfig, pages []dbpkg.EmpresaVentaPublicaPagina, items []dbpkg.EmpresaVentaPublicaItem) string {
	var b strings.Builder
	b.WriteString("Eres un asistente publico de una pagina de venta publica empresarial dentro de Powerful Control System. ")
	b.WriteString("Responde solo sobre los productos, servicios, precios, disponibilidad publicada, paginas visibles y forma de compra de esta empresa. ")
	b.WriteString("No hables de otras empresas, no inventes catalogo, no des informacion interna, no menciones bases de datos, endpoints ni configuracion administrativa. ")
	b.WriteString("Si en el contexto existe CATALOGO_PUBLICO, asumelo como la fuente principal y menciona explicitamente los productos o servicios visibles antes de hablar de paginas generales. ")
	b.WriteString("Nunca digas que no ves catalogo, productos o servicios si CATALOGO_PUBLICO trae al menos un item. ")
	b.WriteString("Si te preguntan que vende la empresa, responde primero con los items exactos del catalogo publico, incluyendo nombre y precio cuando este disponible. ")
	b.WriteString("Si no ves un dato en el contexto, dilo con claridad. ")
	b.WriteString("Puedes sugerir carrito, pago, WhatsApp o mensaje privado cuando ayude al cliente. ")
	b.WriteString("No pidas contrasenas, tarjetas ni datos sensibles. Responde en espanol breve y comercial.\n\n")
	b.WriteString("TIENDA_PUBLICA:\n")
	b.WriteString("- empresa_slug: " + strings.TrimSpace(cfg.EmpresaSlug) + "\n")
	b.WriteString("- nombre_tienda: " + strings.TrimSpace(cfg.NombreTienda) + "\n")
	b.WriteString("- descripcion: " + strings.TrimSpace(cfg.DescripcionTienda) + "\n")
	b.WriteString("- moneda: " + strings.TrimSpace(cfg.Moneda) + "\n")
	if len(pages) > 0 {
		b.WriteString("PAGINAS_PUBLICAS:\n")
		for _, page := range pages {
			b.WriteString("- " + strings.TrimSpace(page.Nombre) + " | slug=" + strings.TrimSpace(page.Slug) + " | " + strings.TrimSpace(page.Descripcion) + "\n")
		}
	}
	if len(items) > 0 {
		b.WriteString("CATALOGO_PUBLICO:\n")
		for _, item := range items {
			b.WriteString(fmt.Sprintf("- %s | precio=%0.0f %s | pagina=%s | destacado=%t | descripcion=%s\n",
				strings.TrimSpace(item.Nombre),
				item.Precio,
				strings.TrimSpace(item.Moneda),
				strings.TrimSpace(item.PaginaNombre),
				item.Destacado,
				strings.TrimSpace(item.Descripcion),
			))
		}
	}
	return b.String()
}

func portalPublicQuestionWantsCatalog(question string) bool {
	q := strings.ToLower(strings.TrimSpace(question))
	if q == "" {
		return false
	}
	keywords := []string{
		"catalogo", "catálogo", "producto", "productos", "servicio", "servicios",
		"promocion", "promoción", "promociones", "vende", "ofrece", "precio", "precios",
		"que tiene", "qué tiene", "que venden", "qué venden", "que ofrece", "qué ofrece",
	}
	for _, kw := range keywords {
		if strings.Contains(q, kw) {
			return true
		}
	}
	return false
}

func portalPublicQuestionWantsPages(question string) bool {
	q := strings.ToLower(strings.TrimSpace(question))
	if q == "" {
		return false
	}
	keywords := []string{
		"paginas", "páginas", "pagina", "página", "secciones", "categorias", "categorías",
		"que paginas", "qué páginas", "que secciones", "qué secciones",
	}
	for _, kw := range keywords {
		if strings.Contains(q, kw) {
			return true
		}
	}
	return false
}

func portalPublicQuestionWantsPromotions(question string) bool {
	q := strings.ToLower(strings.TrimSpace(question))
	if q == "" {
		return false
	}
	keywords := []string{
		"promo", "promos", "promocion", "promoción", "promociones", "destacado", "destacados", "oferta", "ofertas",
	}
	for _, kw := range keywords {
		if strings.Contains(q, kw) {
			return true
		}
	}
	return false
}

func portalPublicQuestionWantsPrices(question string) bool {
	q := strings.ToLower(strings.TrimSpace(question))
	if q == "" {
		return false
	}
	keywords := []string{
		"precio", "precios", "valor", "valores", "cuanto cuesta", "cuánto cuesta", "cuanto vale", "cuánto vale",
	}
	for _, kw := range keywords {
		if strings.Contains(q, kw) {
			return true
		}
	}
	return false
}

func buildPortalPublicCatalogAnswer(cfg dbpkg.EmpresaVentaPublicaConfig, pages []dbpkg.EmpresaVentaPublicaPagina, items []dbpkg.EmpresaVentaPublicaItem) string {
	var b strings.Builder
	storeName := strings.TrimSpace(cfg.NombreTienda)
	if storeName == "" {
		storeName = "esta tienda"
	}
	if len(items) > 0 {
		b.WriteString(storeName + " tiene publicado este catalogo publico:\n\n")
		limit := len(items)
		if limit > 8 {
			limit = 8
		}
		for i := 0; i < limit; i++ {
			item := items[i]
			b.WriteString("- **" + strings.TrimSpace(item.Nombre) + "**")
			if item.Precio > 0 {
				moneda := strings.TrimSpace(item.Moneda)
				if moneda == "" {
					moneda = strings.TrimSpace(cfg.Moneda)
				}
				if moneda == "" {
					moneda = "COP"
				}
				b.WriteString(fmt.Sprintf(" - %0.0f %s", item.Precio, moneda))
			}
			if page := strings.TrimSpace(item.PaginaNombre); page != "" {
				b.WriteString(" | pagina: " + page)
			}
			if desc := strings.TrimSpace(item.Descripcion); desc != "" {
				b.WriteString(" | " + desc)
			}
			b.WriteString("\n")
		}
		if len(items) > limit {
			b.WriteString(fmt.Sprintf("\nHay %d publicaciones activas en total. ", len(items)))
		} else {
			b.WriteString("\n")
		}
		b.WriteString("Si quieres, tambien puedo ayudarte a revisar una pagina especifica o indicarte como comprar o reservar.")
		return b.String()
	}

	if len(pages) > 0 {
		b.WriteString("Ahora mismo no veo productos individuales publicados, pero si estas paginas publicas activas:\n\n")
		for _, page := range pages {
			b.WriteString("- **" + strings.TrimSpace(page.Nombre) + "**")
			if desc := strings.TrimSpace(page.Descripcion); desc != "" {
				b.WriteString(" - " + desc)
			}
			b.WriteString("\n")
		}
		b.WriteString("\nSi quieres, puedo ayudarte a revisar una de esas paginas o indicarte como comprar o reservar.")
		return b.String()
	}

	return "Ahora mismo no veo productos ni paginas publicas activas para esta tienda. Si quieres, puedo ayudarte a intentar otra consulta o indicarte como contactar a la empresa."
}

func buildPortalPublicPagesAnswer(cfg dbpkg.EmpresaVentaPublicaConfig, pages []dbpkg.EmpresaVentaPublicaPagina) string {
	storeName := strings.TrimSpace(cfg.NombreTienda)
	if storeName == "" {
		storeName = "Esta tienda"
	}
	if len(pages) == 0 {
		return storeName + " no tiene paginas publicas activas en este momento."
	}
	var b strings.Builder
	b.WriteString(storeName + " tiene estas paginas publicas activas:\n\n")
	for _, page := range pages {
		b.WriteString("- **" + strings.TrimSpace(page.Nombre) + "**")
		if slug := strings.TrimSpace(page.Slug); slug != "" {
			b.WriteString(" | slug: " + slug)
		}
		if desc := strings.TrimSpace(page.Descripcion); desc != "" {
			b.WriteString(" | " + desc)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nSi quieres, tambien puedo mostrarte las promociones o el catalogo publicado.")
	return b.String()
}

func buildPortalPublicPromotionsAnswer(cfg dbpkg.EmpresaVentaPublicaConfig, pages []dbpkg.EmpresaVentaPublicaPagina, items []dbpkg.EmpresaVentaPublicaItem) string {
	storeName := strings.TrimSpace(cfg.NombreTienda)
	if storeName == "" {
		storeName = "Esta tienda"
	}
	featured := make([]dbpkg.EmpresaVentaPublicaItem, 0, len(items))
	for _, item := range items {
		if item.Destacado {
			featured = append(featured, item)
		}
	}
	if len(featured) == 0 {
		if len(items) == 0 {
			return buildPortalPublicPagesAnswer(cfg, pages)
		}
		return buildPortalPublicCatalogAnswer(cfg, pages, items)
	}
	var b strings.Builder
	b.WriteString(storeName + " tiene estas promociones o destacados visibles:\n\n")
	limit := len(featured)
	if limit > 8 {
		limit = 8
	}
	for i := 0; i < limit; i++ {
		item := featured[i]
		b.WriteString("- **" + strings.TrimSpace(item.Nombre) + "**")
		if item.Precio > 0 {
			moneda := strings.TrimSpace(item.Moneda)
			if moneda == "" {
				moneda = strings.TrimSpace(cfg.Moneda)
			}
			if moneda == "" {
				moneda = "COP"
			}
			b.WriteString(fmt.Sprintf(" - %0.0f %s", item.Precio, moneda))
		}
		if desc := strings.TrimSpace(item.Descripcion); desc != "" {
			b.WriteString(" | " + desc)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nSi quieres, tambien puedo mostrarte el catalogo completo o las paginas publicas.")
	return b.String()
}

func buildPortalPublicPricesAnswer(cfg dbpkg.EmpresaVentaPublicaConfig, items []dbpkg.EmpresaVentaPublicaItem) string {
	if len(items) == 0 {
		return "Ahora mismo no veo precios publicos activos para esta tienda."
	}
	var b strings.Builder
	b.WriteString("Estos son los precios publicos visibles en este momento:\n\n")
	limit := len(items)
	if limit > 8 {
		limit = 8
	}
	for i := 0; i < limit; i++ {
		item := items[i]
		moneda := strings.TrimSpace(item.Moneda)
		if moneda == "" {
			moneda = strings.TrimSpace(cfg.Moneda)
		}
		if moneda == "" {
			moneda = "COP"
		}
		b.WriteString("- **" + strings.TrimSpace(item.Nombre) + "**")
		if item.Precio > 0 {
			b.WriteString(fmt.Sprintf(" - %0.0f %s", item.Precio, moneda))
		}
		if page := strings.TrimSpace(item.PaginaNombre); page != "" {
			b.WriteString(" | pagina: " + page)
		}
		b.WriteString("\n")
	}
	b.WriteString("\nSi quieres, tambien puedo mostrarte las promociones o el catalogo completo.")
	return b.String()
}
func loadPortalPublicStoreItems(dbEmp *sql.DB, empresaID int64, pages []dbpkg.EmpresaVentaPublicaPagina) ([]dbpkg.EmpresaVentaPublicaItem, error) {
	items, _, err := dbpkg.ListEmpresaVentaPublicaItems(dbEmp, empresaID, dbpkg.EmpresaVentaPublicaItemsFilter{
		IncludeInactive: false,
		Limit:           120,
		Offset:          0,
	})
	if err != nil {
		return nil, err
	}
	if len(items) > 0 || len(pages) == 0 {
		return items, nil
	}

	merged := make([]dbpkg.EmpresaVentaPublicaItem, 0, len(pages)*4)
	seen := map[int64]struct{}{}
	for _, page := range pages {
		pageItems, _, err := dbpkg.ListEmpresaVentaPublicaItems(dbEmp, empresaID, dbpkg.EmpresaVentaPublicaItemsFilter{
			IncludeInactive: false,
			PaginaSlug:      page.Slug,
			Limit:           120,
			Offset:          0,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range pageItems {
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			merged = append(merged, item)
		}
	}
	return merged, nil
}

func pickPortalChatModel(dbSuper *sql.DB, question string, wantsVision bool) (empresaAIModelDef, bool) {
	modelMap := empresaAIModelMap()
	defs := aiCredentialCatalogModels()
	if len(defs) == 0 {
		return empresaAIModelDef{}, false
	}

	normalizedQuestion := strings.ToLower(strings.TrimSpace(question))
	preferResponses := wantsVision ||
		len([]rune(normalizedQuestion)) > 700 ||
		strings.Contains(normalizedQuestion, "analiza") ||
		strings.Contains(normalizedQuestion, "documento") ||
		strings.Contains(normalizedQuestion, "imagen") ||
		strings.Contains(normalizedQuestion, "foto") ||
		strings.Contains(normalizedQuestion, "archivo")
	if preferResponses {
		candidates := []string{"openai:gpt-5.5", "openai:gpt-5.4-mini"}
		for _, candidate := range candidates {
			if model, ok := modelMap[candidate]; ok && isAIProviderEnabled(dbSuper, model.Provider) {
				return model, true
			}
		}
	}

	for _, wantProvider := range []string{"openai", "google"} {
		for _, d := range defs {
			if !strings.EqualFold(strings.TrimSpace(d.Provider), wantProvider) {
				continue
			}
			model, ok := modelMap[d.ModelID]
			if !ok {
				continue
			}
			if isAIProviderEnabled(dbSuper, model.Provider) {
				return model, true
			}
		}
	}
	for _, d := range defs {
		model, ok := modelMap[d.ModelID]
		if ok && isAIProviderEnabled(dbSuper, model.Provider) {
			return model, true
		}
	}
	return empresaAIModelDef{}, false
}

// PublicPortalCompanyChatHandler expone un chat publico contextual:
// - portal general: preguntas comerciales del index
// - venta_publica: preguntas solo sobre el catalogo publicado de una empresa
// Limite: 10 consultas por 5 minutos por usuario y contexto.
func PublicPortalCompanyChatHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if !isPortalPublicChatEnabled(dbSuper) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":    false,
				"code":  "chat_portal_disabled",
				"error": "El chat publico del portal esta deshabilitado.",
			})
			return
		}
		if !isSuperAIEnabled(dbSuper) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":    false,
				"code":  "ai_disabled",
				"error": "El chat esta temporalmente deshabilitado.",
			})
			return
		}

		var body struct {
			Pregunta    string                 `json:"pregunta"`
			Historial   []empresaAIChatMensaje `json:"historial"`
			Scope       string                 `json:"scope"`
			EmpresaSlug string                 `json:"empresa_slug"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}

		scope := strings.ToLower(strings.TrimSpace(body.Scope))
		if scope == "" {
			scope = "portal"
		}
		rateScope := "portal"
		if scope == "venta_publica" {
			rateScope = "venta_publica:" + dbpkg.NormalizeEmpresaPublicSlug(body.EmpresaSlug)
		}

		clientID := portalChatGetOrSetClientID(w, r)
		ip := portalChatClientIP(r)
		key := rateScope + ":" + clientID + ":" + ip
		ok, remaining, retry := portalChatLimiter.allow(key, 10, 5*time.Minute)
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"ok":                      false,
				"code":                    "rate_limited",
				"error":                   "Has alcanzado el limite de 10 preguntas. Espera unos 5 minutos e intentalo de nuevo, o escribenos por WhatsApp para atencion mas personalizada.",
				"retry_after_seconds":     retry,
				"remaining_in_window":     0,
				"window_seconds":          300,
				"public_contact_whatsapp": "https://wa.me/573043306506",
				"public_contact_email":    "powerfulcontrolsystem@gmail.com",
			})
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

		model, okModel := pickPortalChatModel(dbSuper, p, false)
		if !okModel {
			writeJSON(w, http.StatusBadGateway, map[string]interface{}{
				"ok":    false,
				"code":  "ai_model_missing",
				"error": "Modelo de IA no disponible.",
			})
			return
		}

		ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 35 * time.Second}}
		systemPrompt := buildPortalCompanySystemPrompt()
		if scope == "venta_publica" {
			empresaID, err := dbpkg.ResolveVentaPublicaEmpresaIDFromAny(dbEmp, 0, body.EmpresaSlug)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]interface{}{
					"ok":    false,
					"code":  "empresa_publica_no_encontrada",
					"error": "La tienda publica solicitada no existe.",
				})
				return
			}
			cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "No se pudo cargar la tienda publica", http.StatusInternalServerError)
				return
			}
			pages, err := dbpkg.ListEmpresaVentaPublicaPaginas(dbEmp, empresaID, false)
			if err != nil {
				http.Error(w, "No se pudieron cargar las paginas publicas", http.StatusInternalServerError)
				return
			}
			items, err := loadPortalPublicStoreItems(dbEmp, empresaID, pages)
			if err != nil {
				http.Error(w, "No se pudo cargar el catalogo publico", http.StatusInternalServerError)
				return
			}
			if portalPublicQuestionWantsCatalog(p) {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                  true,
					"respuesta":           buildPortalPublicCatalogAnswer(cfg, pages, items),
					"remaining_in_window": remaining,
					"window_seconds":      300,
					"scope":               scope,
				})
				return
			}
			if portalPublicQuestionWantsPrices(p) {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                  true,
					"respuesta":           buildPortalPublicPricesAnswer(cfg, items),
					"remaining_in_window": remaining,
					"window_seconds":      300,
					"scope":               scope,
				})
				return
			}
			if portalPublicQuestionWantsPromotions(p) {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                  true,
					"respuesta":           buildPortalPublicPromotionsAnswer(cfg, pages, items),
					"remaining_in_window": remaining,
					"window_seconds":      300,
					"scope":               scope,
				})
				return
			}
			if portalPublicQuestionWantsPages(p) {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":                  true,
					"respuesta":           buildPortalPublicPagesAnswer(cfg, pages),
					"remaining_in_window": remaining,
					"window_seconds":      300,
					"scope":               scope,
				})
				return
			}
			systemPrompt = buildPortalPublicStoreSystemPrompt(cfg, pages, items)
		} else {
			if extra := portalChatLoadExtraInfo(dbSuper); extra != "" {
				systemPrompt += "\n\n=== Informacion oficial editable (super administrador) ===\n" + extra
			}
			if pricing := portalChatBuildLicenciasPriceSummary(dbSuper); pricing != "" {
				systemPrompt += "\n\n=== Precios y planes desde base de datos ===\n" + pricing
			}
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
		if scope != "venta_publica" {
			closing := "No olvides que puedes probar ya mismo totalmente gratis el sistema con solo registrarte."
			if !strings.Contains(strings.ToLower(answer), strings.ToLower(closing)) {
				answer = strings.TrimSpace(answer) + "\n\n" + closing
			}
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":                  true,
			"respuesta":           answer,
			"remaining_in_window": remaining,
			"window_seconds":      300,
			"scope":               scope,
		})
	}
}

// PublicPortalCompanyChatStreamHandler expone un chat pÃºblico contextual con soporte para streaming SSE.
func PublicPortalCompanyChatStreamHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if !isPortalPublicChatEnabled(dbSuper) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":    false,
				"code":  "chat_portal_disabled",
				"error": "El chat publico del portal esta deshabilitado.",
			})
			return
		}
		if !isSuperAIEnabled(dbSuper) {
			writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"ok":    false,
				"code":  "ai_disabled",
				"error": "El chat esta temporalmente deshabilitado.",
			})
			return
		}

		var body struct {
			Pregunta    string                 `json:"pregunta"`
			Historial   []empresaAIChatMensaje `json:"historial"`
			Scope       string                 `json:"scope"`
			EmpresaSlug string                 `json:"empresa_slug"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}

		scope := strings.ToLower(strings.TrimSpace(body.Scope))
		if scope == "" {
			scope = "portal"
		}
		rateScope := "portal"
		if scope == "venta_publica" {
			rateScope = "venta_publica:" + dbpkg.NormalizeEmpresaPublicSlug(body.EmpresaSlug)
		}

		clientID := portalChatGetOrSetClientID(w, r)
		ip := portalChatClientIP(r)
		key := rateScope + ":" + clientID + ":" + ip
		ok, _, retry := portalChatLimiter.allow(key, 10, 5*time.Minute)
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			writeJSON(w, http.StatusTooManyRequests, map[string]interface{}{
				"ok":                      false,
				"code":                    "rate_limited",
				"error":                   "Has alcanzado el limite de 10 preguntas. Espera unos 5 minutos e intentalo de nuevo, o escribenos por WhatsApp para atencion mas personalizada.",
				"retry_after_seconds":     retry,
				"remaining_in_window":     0,
				"window_seconds":          300,
				"public_contact_whatsapp": "https://wa.me/573043306506",
				"public_contact_email":    "powerfulcontrolsystem@gmail.com",
			})
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

		model, okModel := pickPortalChatModel(dbSuper, p, false)
		if !okModel {
			writeJSON(w, http.StatusBadGateway, map[string]interface{}{
				"ok":    false,
				"code":  "ai_model_missing",
				"error": "Modelo de IA no disponible.",
			})
			return
		}
		ctrl := &EmpresaAIChatController{dbEmp: dbEmp, dbSuper: dbSuper, client: &http.Client{Timeout: 35 * time.Second}}
		systemPrompt := buildPortalCompanySystemPrompt()
		if scope == "venta_publica" {
			empresaID, err := dbpkg.ResolveVentaPublicaEmpresaIDFromAny(dbEmp, 0, body.EmpresaSlug)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]interface{}{
					"ok":    false,
					"code":  "empresa_publica_no_encontrada",
					"error": "La tienda publica solicitada no existe.",
				})
				return
			}
			cfg, err := dbpkg.GetEmpresaVentaPublicaConfig(dbEmp, empresaID)
			if err != nil {
				http.Error(w, "No se pudo cargar la tienda publica", http.StatusInternalServerError)
				return
			}
			pages, err := dbpkg.ListEmpresaVentaPublicaPaginas(dbEmp, empresaID, false)
			if err != nil {
				http.Error(w, "No se pudieron cargar las paginas publicas", http.StatusInternalServerError)
				return
			}
			items, err := loadPortalPublicStoreItems(dbEmp, empresaID, pages)
			if err != nil {
				http.Error(w, "No se pudo cargar el catalogo publico", http.StatusInternalServerError)
				return
			}
			if portalPublicQuestionWantsCatalog(p) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(http.StatusOK)
				text := buildPortalPublicCatalogAnswer(cfg, pages, items)
				payload, _ := json.Marshal(map[string]interface{}{"text": text})
				fmt.Fprintf(w, "data: %s\n\n", payload)
				fmt.Fprintf(w, "data: [DONE]\n\n")
				if flusher, okFlusher := w.(http.Flusher); okFlusher {
					flusher.Flush()
				}
				return
			}
			if portalPublicQuestionWantsPrices(p) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(http.StatusOK)
				text := buildPortalPublicPricesAnswer(cfg, items)
				payload, _ := json.Marshal(map[string]interface{}{"text": text})
				fmt.Fprintf(w, "data: %s\n\n", payload)
				fmt.Fprintf(w, "data: [DONE]\n\n")
				if flusher, okFlusher := w.(http.Flusher); okFlusher {
					flusher.Flush()
				}
				return
			}
			if portalPublicQuestionWantsPromotions(p) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(http.StatusOK)
				text := buildPortalPublicPromotionsAnswer(cfg, pages, items)
				payload, _ := json.Marshal(map[string]interface{}{"text": text})
				fmt.Fprintf(w, "data: %s\n\n", payload)
				fmt.Fprintf(w, "data: [DONE]\n\n")
				if flusher, okFlusher := w.(http.Flusher); okFlusher {
					flusher.Flush()
				}
				return
			}
			if portalPublicQuestionWantsPages(p) {
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.WriteHeader(http.StatusOK)
				text := buildPortalPublicPagesAnswer(cfg, pages)
				payload, _ := json.Marshal(map[string]interface{}{"text": text})
				fmt.Fprintf(w, "data: %s\n\n", payload)
				fmt.Fprintf(w, "data: [DONE]\n\n")
				if flusher, okFlusher := w.(http.Flusher); okFlusher {
					flusher.Flush()
				}
				return
			}
			systemPrompt = buildPortalPublicStoreSystemPrompt(cfg, pages, items)
		} else {
			if extra := portalChatLoadExtraInfo(dbSuper); extra != "" {
				systemPrompt += "\n\n=== Informacion oficial editable (super administrador) ===\n" + extra
			}
			if pricing := portalChatBuildLicenciasPriceSummary(dbSuper); pricing != "" {
				systemPrompt += "\n\n=== Precios y planes desde base de datos ===\n" + pricing
			}
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		h := sanitizeHistorial(body.Historial, 6)
		answer, err := ctrl.callOpenAIStreamChatCompletions(model, p, h, systemPrompt, func(delta string) {
			b, _ := json.Marshal(map[string]interface{}{"text": delta})
			fmt.Fprintf(w, "data: %s\n\n", b)
			if flusher, okFlusher := w.(http.Flusher); okFlusher {
				flusher.Flush()
			}
		})
		if err != nil {
			errJSON, _ := json.Marshal(map[string]interface{}{"error": "No se pudo generar respuesta: " + err.Error()})
			fmt.Fprintf(w, "data: %s\n\n", errJSON)
		} else {
			answer = strings.TrimSpace(answer)
			if scope != "venta_publica" {
				closing := "No olvides que puedes probar ya mismo totalmente gratis el sistema con solo registrarte."
				if !strings.Contains(strings.ToLower(answer), strings.ToLower(closing)) {
					b, _ := json.Marshal(map[string]interface{}{"text": "\n\n" + closing})
					fmt.Fprintf(w, "data: %s\n\n", b)
				}
			}
		}
		fmt.Fprintf(w, "data: [DONE]\n\n")
		if flusher, okFlusher := w.(http.Flusher); okFlusher {
			flusher.Flush()
		}
	}
}
