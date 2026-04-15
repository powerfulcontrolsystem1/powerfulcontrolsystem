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

type paginaPrincipalCard struct {
	Titulo      string `json:"titulo"`
	Descripcion string `json:"descripcion"`
	ImagenURL   string `json:"imagen_url"`
	Enlace      string `json:"enlace"`
}

type paginaPrincipalConfig struct {
	Cantidad int                   `json:"cantidad"`
	Tarjetas []paginaPrincipalCard `json:"tarjetas"`
}

func paginaPrincipalDefaultConfig() paginaPrincipalConfig {
	cards := []paginaPrincipalCard{
		{
			Titulo:      "Punto de venta",
			Descripcion: "Solucion completa para ventas rapidas y facturacion electronica.",
			ImagenURL:   "/img/punto_venta.png",
			Enlace:      "/administrar_empresa.html?module=punto_venta",
		},
		{
			Titulo:      "Motel",
			Descripcion: "Gestion por tiempo de servicio y facturacion tarifada por estancia.",
			ImagenURL:   "/img/motel.png",
			Enlace:      "/administrar_empresa.html?module=motel",
		},
		{
			Titulo:      "Restaurante",
			Descripcion: "Gestion de mesas, pedidos y facturacion para restaurantes.",
			ImagenURL:   "/img/restaurante.png",
			Enlace:      "/administrar_empresa.html?module=restaurante",
		},
		{
			Titulo:      "Control por sensor",
			Descripcion: "Integracion y alertas con sensores para control de accesos.",
			ImagenURL:   "/img/sensor.png",
			Enlace:      "/administrar_empresa.html?module=sensor",
		},
		{
			Titulo:      "Hotel",
			Descripcion: "Administracion de empresas, roles y permisos para operacion hotelera.",
			ImagenURL:   "/img/settings-color.svg",
			Enlace:      "/administrar_empresa.html?module=configuracion",
		},
	}
	return paginaPrincipalConfig{
		Cantidad: len(cards),
		Tarjetas: cards,
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
			Titulo:      title,
			Descripcion: description,
			ImagenURL:   paginaPrincipalNormalizeImageURL(current.ImagenURL, base.ImagenURL),
			Enlace:      paginaPrincipalNormalizeLink(current.Enlace, base.Enlace),
		})
	}

	return paginaPrincipalConfig{
		Cantidad: cfg.Cantidad,
		Tarjetas: normalized,
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

// SuperPaginaPrincipalHandler administra las tarjetas configurables del index para el panel super.
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

// PublicPaginaPrincipalHandler expone tarjetas del index para visualizacion publica.
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
			"updated_at": updatedAt,
		})
	}
}
