package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	paginaPrincipalConfigKey          = "super.pagina_principal.cards.v1"
	paginaPrincipalConfigUpdatedByKey = "super.pagina_principal.cards.v1.updated_by"
	paginaPrincipalDefaultCardLimit   = 12
)

const (
	paginaPrincipalVisualSizeSmall  = "pequeno"
	paginaPrincipalVisualSizeMedium = "mediano"
	paginaPrincipalVisualSizeLarge  = "grande"
)

type paginaPrincipalCard struct {
	Titulo            string   `json:"titulo"`
	Descripcion       string   `json:"descripcion"`
	ImagenURL         string   `json:"imagen_url"`
	ImagenSecundaria  string   `json:"imagen_secundaria_url,omitempty"`
	Enlace            string   `json:"enlace"`
	YouTubeURL        string   `json:"youtube_url,omitempty"`
	DetalleEtiqueta   string   `json:"detalle_etiqueta"`
	DetalleTitular    string   `json:"detalle_titular"`
	DetalleParrafoUno string   `json:"detalle_parrafo_uno"`
	DetalleParrafoDos string   `json:"detalle_parrafo_dos"`
	DetallePuntos     []string `json:"detalle_puntos"`
}

type paginaPrincipalBannerCard struct {
	ImagenURL string `json:"imagen_url"`
	Enlace    string `json:"enlace,omitempty"`
}

type paginaPrincipalVisualSettings struct {
	IndexCardSize   string `json:"index_card_size"`
	IndexTextSize   string `json:"index_text_size"`
	LandingCardSize string `json:"landing_card_size"`
	LandingTextSize string `json:"landing_text_size"`
}

type paginaPrincipalConfig struct {
	Cantidad int                           `json:"cantidad"`
	Tarjetas []paginaPrincipalCard         `json:"tarjetas"`
	Estilos  paginaPrincipalVisualSettings `json:"estilos"`
}

func paginaPrincipalDefaultWhatsAppContactNumber() string {
	return "573043306506"
}

func normalizePortalWhatsAppContactNumber(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	builder := strings.Builder{}
	for idx, r := range trimmed {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
			continue
		}
		if r == '+' && idx == 0 {
			continue
		}
	}
	normalized := builder.String()
	if len(normalized) < 10 || len(normalized) > 15 {
		return ""
	}
	return normalized
}

func paginaPrincipalLoadWhatsAppContactNumber(dbSuper *sql.DB) string {
	configured, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, "portal.whatsapp_contact_number")
	if err != nil {
		return paginaPrincipalDefaultWhatsAppContactNumber()
	}
	normalized := normalizePortalWhatsAppContactNumber(configured)
	if normalized == "" {
		return paginaPrincipalDefaultWhatsAppContactNumber()
	}
	return normalized
}

func paginaPrincipalDefaultVisualSettings() paginaPrincipalVisualSettings {
	return paginaPrincipalVisualSettings{
		IndexCardSize:   paginaPrincipalVisualSizeMedium,
		IndexTextSize:   paginaPrincipalVisualSizeMedium,
		LandingCardSize: paginaPrincipalVisualSizeMedium,
		LandingTextSize: paginaPrincipalVisualSizeMedium,
	}
}

func paginaPrincipalDefaultConfig() paginaPrincipalConfig {
	cards := []paginaPrincipalCard{
		{
			Titulo:            "Punto de venta",
			Descripcion:       "Solucion completa para ventas rapidas y facturacion electronica.",
			ImagenURL:         "/img/punto_venta.png",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=punto_venta",
			DetalleEtiqueta:   "Retail y mostrador",
			DetalleTitular:    "Vende rapido, factura mejor y controla la caja desde una sola pantalla.",
			DetalleParrafoUno: "La solucion de Punto de Venta esta pensada para negocios que necesitan registrar ventas agiles sin perder trazabilidad. Cada movimiento puede quedar asociado a empresa, usuario, estacion, cliente, metodo de pago y documento emitido, lo que facilita operar mostrador, caja rapida o atencion general desde una misma interfaz.",
			DetalleParrafoDos: "Ademas de cobrar, el sistema ayuda a controlar inventario, descuentos, precios, clientes frecuentes, cierres de caja y reportes operativos. Esto permite que el comercio no solo venda mas rapido, sino que tambien mantenga orden financiero y visibilidad real sobre lo que esta ocurriendo en el negocio.",
			DetallePuntos: []string{
				"Facturacion electronica y documentos de venta con trazabilidad por empresa.",
				"Carritos de compra rapidos con productos, servicios, combos y descuentos controlados.",
				"Cierres de caja, metodos de pago y conciliacion operativa para supervision diaria.",
				"Inventario sincronizado, historial de ventas y reportes para decisiones comerciales.",
			},
		},
		{
			Titulo:            "Motel",
			Descripcion:       "Gestion por tiempo de servicio y facturacion tarifada por estancia.",
			ImagenURL:         "/img/motel.png",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=motel",
			DetalleEtiqueta:   "Operacion por estancias",
			DetalleTitular:    "Controla habitaciones, tiempos de ocupacion y consumos sin perder detalle.",
			DetalleParrafoUno: "El sistema para Motel esta orientado a negocios donde el cobro depende del tiempo de uso, la disponibilidad de habitaciones y los consumos agregados durante la estancia. Permite conocer que habitaciones estan libres, ocupadas, reservadas o listas para limpieza, facilitando la operacion en tiempo real.",
			DetalleParrafoDos: "La plataforma combina tarifas por minutos, tarifas por bloques, cargos adicionales, consumos de minibar o servicios y seguimiento de pagos por habitacion. Esto mejora la rotacion, reduce errores de cobro y entrega una trazabilidad clara para auditoria interna y control administrativo.",
			DetallePuntos: []string{
				"Tarifas por tiempo, reglas por bloques y calculo automatico del valor a cobrar.",
				"Control de habitaciones con estados operativos y consumos asociados a la estancia.",
				"Carritos simultaneos por estacion o habitacion con aislamiento por empresa.",
				"Reportes de ocupacion, ingresos por turno y seguimiento detallado de servicios.",
			},
		},
		{
			Titulo:            "Restaurante",
			Descripcion:       "Gestion de mesas, pedidos y facturacion para restaurantes.",
			ImagenURL:         "/img/restaurante.png",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=restaurante",
			DetalleEtiqueta:   "Mesas y cocina",
			DetalleTitular:    "Administra mesas, pedidos, cocina y cobro final con flujo continuo.",
			DetalleParrafoUno: "La solucion para Restaurante ayuda a organizar la atencion desde que el cliente se sienta hasta que se factura la cuenta. El personal puede trabajar por mesas, estaciones o usuarios, tomar pedidos rapidamente y mantener control sobre productos, tiempos de despacho y consumos acumulados.",
			DetalleParrafoDos: "Tambien facilita dividir cuentas, manejar propinas, emitir facturas, enviar ordenes a cocina o barra y consultar reportes por turno. Con esto, el restaurante gana velocidad en servicio, reduce reprocesos y mejora la coordinacion entre salon, caja y produccion.",
			DetallePuntos: []string{
				"Gestion de mesas, pedidos abiertos y consumos acumulados por cliente o grupo.",
				"Impresion o resolucion de comandos para cocina, barra o puntos de preparacion.",
				"Propinas, descuentos y metodos de pago integrados al cierre de la cuenta.",
				"Indicadores de ventas, rotacion de mesas y control operativo por turno.",
			},
		},
		{
			Titulo:            "Control por sensor",
			Descripcion:       "Integracion y alertas con sensores para control de accesos.",
			ImagenURL:         "/img/sensor.png",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=sensor",
			DetalleEtiqueta:   "Monitoreo y automatizacion",
			DetalleTitular:    "Conecta eventos fisicos con alertas, estados y control operativo centralizado.",
			DetalleParrafoUno: "La solucion de Control por Sensor esta diseñada para operaciones donde un evento fisico debe producir una accion o una evidencia digital. Puede servir para accesos, aperturas, cierres, confirmaciones de paso, sensores de puerta o estados que deban registrarse automaticamente para soporte operativo o seguridad.",
			DetalleParrafoDos: "En lugar de depender solo de verificaciones manuales, el sistema centraliza senales, alertas y trazabilidad para que supervisores y administradores sepan que ocurrio, cuando ocurrio y en que punto operativo sucedio. Esto mejora la reaccion, la auditoria y la consistencia de la operacion diaria.",
			DetallePuntos: []string{
				"Registro de eventos de sensores con relacion a estaciones o puntos de control.",
				"Alertas operativas y visibilidad de estados recientes para soporte inmediato.",
				"Trazabilidad de accesos, aperturas o incidencias con evidencia temporal.",
				"Integracion con flujos operativos que requieren validacion fisica o automatizada.",
			},
		},
		{
			Titulo:            "Hotel",
			Descripcion:       "Administracion de empresas, roles y permisos para operacion hotelera.",
			ImagenURL:         "/img/settings-color.svg",
			ImagenSecundaria:  "/img/sistema punto de venta.png",
			Enlace:            "/administrar_empresa.html?module=configuracion",
			DetalleEtiqueta:   "Reservas y hospedaje",
			DetalleTitular:    "Gestiona reservas, check-in, check-out y facturacion hotelera en un mismo flujo.",
			DetalleParrafoUno: "La solucion para Hotel permite administrar habitaciones, reservas futuras, ocupacion actual y cargos adicionales dentro de una operacion unificada. De esta forma, recepcion y administracion pueden trabajar con informacion consistente sobre disponibilidad, tiempos de entrada y salida y consumos por huesped.",
			DetalleParrafoDos: "El sistema ayuda a estructurar el ciclo completo del hospedaje: reserva, confirmacion, asignacion de habitacion, facturacion, cargos por servicios y control posterior. Esto lo convierte en una herramienta util tanto para pequeños hoteles como para operaciones que necesitan mayor orden en caja, reportes y servicio al cliente.",
			DetallePuntos: []string{
				"Control de reservas por rango de fechas y disponibilidad real de habitaciones.",
				"Check-in y check-out con cargos diarios, servicios adicionales y seguimiento por huesped.",
				"Facturacion, reportes de ocupacion y trazabilidad por empresa y periodo.",
				"Soporte para operaciones multiusuario con historial claro de movimientos y cobros.",
			},
		},
	}
	return paginaPrincipalConfig{
		Cantidad: len(cards),
		Tarjetas: cards,
		Estilos:  paginaPrincipalDefaultVisualSettings(),
	}
}

func paginaPrincipalNormalizeVisualSize(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "small", "pequeno":
		return paginaPrincipalVisualSizeSmall
	case "medium", "mediano":
		return paginaPrincipalVisualSizeMedium
	case "large", "grande":
		return paginaPrincipalVisualSizeLarge
	default:
		return paginaPrincipalVisualSizeMedium
	}
}

func paginaPrincipalNormalizeVisualSettings(raw paginaPrincipalVisualSettings) paginaPrincipalVisualSettings {
	defaults := paginaPrincipalDefaultVisualSettings()
	indexCardSize := paginaPrincipalNormalizeVisualSize(raw.IndexCardSize)
	if indexCardSize == "" {
		indexCardSize = defaults.IndexCardSize
	}
	indexTextSize := paginaPrincipalNormalizeVisualSize(raw.IndexTextSize)
	if indexTextSize == "" {
		indexTextSize = defaults.IndexTextSize
	}
	landingCardSize := paginaPrincipalNormalizeVisualSize(raw.LandingCardSize)
	if landingCardSize == "" {
		landingCardSize = defaults.LandingCardSize
	}
	landingTextSize := paginaPrincipalNormalizeVisualSize(raw.LandingTextSize)
	if landingTextSize == "" {
		landingTextSize = defaults.LandingTextSize
	}
	return paginaPrincipalVisualSettings{
		IndexCardSize:   indexCardSize,
		IndexTextSize:   indexTextSize,
		LandingCardSize: landingCardSize,
		LandingTextSize: landingTextSize,
	}
}

func paginaPrincipalNormalizeImageURL(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		value = "/img/punto_venta.png"
	}
	if strings.Contains(value, "..") {
		return "/img/punto_venta.png"
	}
	if !strings.HasPrefix(value, "/img/") {
		return "/img/punto_venta.png"
	}
	return value
}

func paginaPrincipalNormalizeBannerCards(raw []paginaPrincipalBannerCard) []paginaPrincipalBannerCard {
	source := raw
	if len(source) == 0 {
		return []paginaPrincipalBannerCard{}
	}
	out := make([]paginaPrincipalBannerCard, 0, len(source))
	for _, it := range source {
		img := paginaPrincipalNormalizeImageURL(it.ImagenURL, "/img/baner_ia.png")
		out = append(out, paginaPrincipalBannerCard{
			ImagenURL: img,
			Enlace:    paginaPrincipalNormalizeLink(it.Enlace, ""),
		})
	}
	return out
}

func paginaPrincipalNormalizeLink(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
	}
	if value == "" {
		return "/login.html"
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return value
	}
	if strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + strings.TrimLeft(value, "/")
}

func paginaPrincipalNormalizeYouTubeURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "https://youtu.be/") ||
		strings.HasPrefix(lower, "https://www.youtube.com/") ||
		strings.HasPrefix(lower, "https://youtube.com/") ||
		strings.HasPrefix(lower, "http://youtu.be/") ||
		strings.HasPrefix(lower, "http://www.youtube.com/") ||
		strings.HasPrefix(lower, "http://youtube.com/") {
		// Normalizar a https para evitar mixed-content en el portal.
		value = strings.TrimSpace(value)
		if strings.HasPrefix(strings.ToLower(value), "http://") {
			value = "https://" + strings.TrimPrefix(value, "http://")
		}
		return value
	}
	return ""
}

func paginaPrincipalNormalizeText(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}

func paginaPrincipalNormalizePoints(raw, fallback []string) []string {
	normalized := make([]string, 0, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	if len(normalized) > 0 {
		return normalized
	}
	for _, item := range fallback {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func paginaPrincipalNormalizeConfig(cfg paginaPrincipalConfig) paginaPrincipalConfig {
	defaults := paginaPrincipalDefaultConfig()
	if cfg.Cantidad <= 0 {
		cfg.Cantidad = len(cfg.Tarjetas)
	}
	if cfg.Cantidad <= 0 {
		cfg.Cantidad = defaults.Cantidad
	}
	if cfg.Cantidad > paginaPrincipalDefaultCardLimit {
		cfg.Cantidad = paginaPrincipalDefaultCardLimit
	}

	normalized := make([]paginaPrincipalCard, 0, cfg.Cantidad)
	for i := 0; i < cfg.Cantidad; i++ {
		base := defaults.Tarjetas[i%len(defaults.Tarjetas)]
		var current paginaPrincipalCard
		if i < len(cfg.Tarjetas) {
			current = cfg.Tarjetas[i]
		}
		title := strings.TrimSpace(current.Titulo)
		if title == "" {
			title = base.Titulo
		}
		description := strings.TrimSpace(current.Descripcion)
		if description == "" {
			description = base.Descripcion
		}
		normalized = append(normalized, paginaPrincipalCard{
			Titulo:            title,
			Descripcion:       description,
			ImagenURL:         paginaPrincipalNormalizeImageURL(current.ImagenURL, base.ImagenURL),
			ImagenSecundaria:  paginaPrincipalNormalizeImageURL(current.ImagenSecundaria, base.ImagenSecundaria),
			Enlace:            paginaPrincipalNormalizeLink(current.Enlace, base.Enlace),
			YouTubeURL:        paginaPrincipalNormalizeYouTubeURL(current.YouTubeURL),
			DetalleEtiqueta:   paginaPrincipalNormalizeText(current.DetalleEtiqueta, base.DetalleEtiqueta),
			DetalleTitular:    paginaPrincipalNormalizeText(current.DetalleTitular, base.DetalleTitular),
			DetalleParrafoUno: paginaPrincipalNormalizeText(current.DetalleParrafoUno, base.DetalleParrafoUno),
			DetalleParrafoDos: paginaPrincipalNormalizeText(current.DetalleParrafoDos, base.DetalleParrafoDos),
			DetallePuntos:     paginaPrincipalNormalizePoints(current.DetallePuntos, base.DetallePuntos),
		})
	}

	return paginaPrincipalConfig{
		Cantidad: cfg.Cantidad,
		Tarjetas: normalized,
		Estilos:  paginaPrincipalNormalizeVisualSettings(cfg.Estilos),
	}
}

func paginaPrincipalLoadConfig(dbSuper *sql.DB) (paginaPrincipalConfig, string, string, error) {
	cfg := paginaPrincipalDefaultConfig()
	stored, _, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, paginaPrincipalConfigKey)
	if err != nil {
		return cfg, "", "", err
	}
	updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, paginaPrincipalConfigUpdatedByKey)
	if strings.TrimSpace(stored) == "" {
		return cfg, "", strings.TrimSpace(updatedBy), nil
	}

	var decoded paginaPrincipalConfig
	if err := json.Unmarshal([]byte(stored), &decoded); err != nil {
		log.Printf("[pagina_principal] invalid config JSON, fallback defaults: %v", err)
		return cfg, strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
	}

	return paginaPrincipalNormalizeConfig(decoded), strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
}

func paginaPrincipalSaveConfig(dbSuper *sql.DB, cfg paginaPrincipalConfig, updatedBy string) error {
	normalized := paginaPrincipalNormalizeConfig(cfg)
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	if err := dbpkg.SetConfigValue(dbSuper, paginaPrincipalConfigKey, string(encoded), false); err != nil {
		return err
	}
	actor := strings.TrimSpace(updatedBy)
	if actor == "" {
		actor = "sistema"
	}
	if err := dbpkg.SetConfigValue(dbSuper, paginaPrincipalConfigUpdatedByKey, actor, false); err != nil {
		return err
	}
	return nil
}

func paginaPrincipalListImageURLs(webDir string) ([]string, error) {
	imgDir := filepath.Join(strings.TrimSpace(webDir), "img")
	entries, err := os.ReadDir(imgDir)
	if err != nil {
		return nil, err
	}
	images := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		switch ext {
		case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg":
			images = append(images, "/img/"+entry.Name())
		}
	}
	sort.Slice(images, func(i, j int) bool {
		return strings.ToLower(images[i]) < strings.ToLower(images[j])
	})
	return images, nil
}

func paginaPrincipalRoleIsSuper(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "super_administrador", "superadministrador", "superadmin", "super":
		return true
	default:
		return false
	}
}

func paginaPrincipalRequireSuperAdmin(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) (string, bool) {
	cookie, err := r.Cookie("session_token")
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return "", false
	}
	session, err := dbpkg.GetSessionByToken(dbSuper, cookie.Value)
	if err != nil || session == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return "", false
	}
	admin, err := dbpkg.GetAdminByEmail(dbSuper, strings.TrimSpace(session.AdminEmail))
	if err != nil {
		http.Error(w, "failed to resolve admin session", http.StatusInternalServerError)
		return "", false
	}
	if admin == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return "", false
	}
	if !paginaPrincipalRoleIsSuper(admin.Role) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return "", false
	}
	return strings.TrimSpace(admin.Email), true
}

// SuperPaginaPrincipalHandler administra las tarjetas configurables del portal principal y su landing descriptiva.
func SuperPaginaPrincipalHandler(dbSuper *sql.DB, webDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		if action == "" {
			action = "config"
		}

		switch r.Method {
		case http.MethodGet:
			switch action {
			case "config", "get", "listar":
				cfg, updatedAt, updatedBy, err := paginaPrincipalLoadConfig(dbSuper)
				if err != nil {
					http.Error(w, "failed to read pagina principal config: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":          true,
					"config":      cfg,
					"updated_at":  updatedAt,
					"updated_by":  updatedBy,
					"admin_email": adminEmail,
				})
				return
			case "imagenes", "images", "listar_imagenes":
				images, err := paginaPrincipalListImageURLs(webDir)
				if err != nil {
					http.Error(w, "failed to list images: "+err.Error(), http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"ok":       true,
					"imagenes": images,
					"total":    len(images),
				})
				return
			default:
				http.Error(w, "action not supported", http.StatusBadRequest)
				return
			}

		case http.MethodPut, http.MethodPost:
			if action != "config" && action != "save" && action != "guardar" {
				http.Error(w, "action not supported", http.StatusBadRequest)
				return
			}
			var payload paginaPrincipalConfig
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			if payload.Cantidad <= 0 {
				http.Error(w, "cantidad must be greater than 0", http.StatusBadRequest)
				return
			}
			normalized := paginaPrincipalNormalizeConfig(payload)
			if err := paginaPrincipalSaveConfig(dbSuper, normalized, adminEmail); err != nil {
				http.Error(w, "failed to save pagina principal config: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"saved":      true,
				"config":     normalized,
				"updated_by": adminEmail,
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// PublicPaginaPrincipalHandler expone tarjetas del portal para visualizacion publica del index y la landing descriptiva.
func PublicPaginaPrincipalHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cfg, updatedAt, _, err := paginaPrincipalLoadConfig(dbSuper)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "failed to read pagina principal config: "+err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":         true,
			"cantidad":   cfg.Cantidad,
			"tarjetas":   cfg.Tarjetas,
			"estilos":    cfg.Estilos,
			"whatsapp_contact_number": paginaPrincipalLoadWhatsAppContactNumber(dbSuper),
			"updated_at": updatedAt,
		})
	}
}
