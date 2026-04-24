package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
)

func normalizeCountryCode(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if len(value) != 2 {
		return ""
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return ""
		}
	}
	return value
}

// PublicGeoHandler expone el país detectado (best-effort) para páginas públicas.
// Prioriza headers típicos de reverse-proxy/CDN (ej. Cloudflare: CF-IPCountry).
// Fallback: "CO".
func PublicGeoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Override opcional para pruebas locales: /api/public/geo?pais=CO
		if test := normalizeCountryCode(r.URL.Query().Get("pais")); test != "" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"pais_codigo": test,
				"source":      "query",
			})
			return
		}

		candidates := []struct {
			header string
			label  string
		}{
			{header: "CF-IPCountry", label: "cf-ipcountry"},
			{header: "X-Country-Code", label: "x-country-code"},
			{header: "X-Geo-Country", label: "x-geo-country"},
		}

		country := ""
		source := "default"
		for _, c := range candidates {
			if v := normalizeCountryCode(r.Header.Get(c.header)); v != "" {
				country = v
				source = c.label
				break
			}
		}
		if country == "" {
			country = "CO"
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"pais_codigo": country,
			"source":      source,
		})
	}
}

