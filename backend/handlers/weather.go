package handlers

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const weatherResponseLimit = 2 << 20

var (
	weatherHTTPClient        = &http.Client{Timeout: 8 * time.Second}
	weatherForecastEndpoint  = "https://api.open-meteo.com/v1/forecast"
	weatherGeocodingEndpoint = "https://geocoding-api.open-meteo.com/v1/search"
)

// EmpresaWeatherHandler consulta Open-Meteo desde el backend para que el panel
// no dependa de conexiones de terceros iniciadas directamente por el navegador.
// El wrapper de la ruta valida sesion, licencia y empresa antes de llegar aqui.
func EmpresaWeatherHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mode := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("mode")))
		switch mode {
		case "forecast":
			latitude, err := weatherCoordinate(r.URL.Query().Get("latitude"), -90, 90)
			if err != nil {
				http.Error(w, "invalid latitude", http.StatusBadRequest)
				return
			}
			longitude, err := weatherCoordinate(r.URL.Query().Get("longitude"), -180, 180)
			if err != nil {
				http.Error(w, "invalid longitude", http.StatusBadRequest)
				return
			}
			query := url.Values{
				"latitude":  {strconv.FormatFloat(latitude, 'f', -1, 64)},
				"longitude": {strconv.FormatFloat(longitude, 'f', -1, 64)},
				"current":   {"temperature_2m,relative_humidity_2m,apparent_temperature,precipitation,weather_code,cloud_cover,wind_speed_10m,is_day"},
				"timezone":  {"auto"},
			}
			proxyWeatherJSON(w, r, weatherForecastEndpoint, query, "public, max-age=180")
		case "geocoding":
			name := strings.TrimSpace(r.URL.Query().Get("name"))
			if name == "" || utf8.RuneCountInString(name) > 120 {
				http.Error(w, "invalid location", http.StatusBadRequest)
				return
			}
			query := url.Values{
				"name":     {name},
				"count":    {"1"},
				"language": {"es"},
				"format":   {"json"},
			}
			proxyWeatherJSON(w, r, weatherGeocodingEndpoint, query, "public, max-age=86400")
		default:
			http.Error(w, "invalid weather mode", http.StatusBadRequest)
		}
	}
}

func weatherCoordinate(raw string, min, max float64) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || value < min || value > max {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}

func proxyWeatherJSON(w http.ResponseWriter, r *http.Request, endpoint string, query url.Values, cacheControl string) {
	target := endpoint + "?" + query.Encode()
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		http.Error(w, "weather service unavailable", http.StatusBadGateway)
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "PowerfulControlSystem/1.0 weather-proxy")

	resp, err := weatherHTTPClient.Do(req)
	if err != nil {
		http.Error(w, "weather service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		http.Error(w, "weather service unavailable", http.StatusBadGateway)
		return
	}

	var payload interface{}
	decoder := json.NewDecoder(io.LimitReader(resp.Body, weatherResponseLimit))
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, "weather service unavailable", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", cacheControl)
	encodeJSONResponse(w, payload)
}
