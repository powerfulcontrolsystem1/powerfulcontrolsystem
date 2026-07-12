package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var publicMarketSymbols = map[string]struct{}{
	"spy.us": {},
	"qqq.us": {},
	"gc.f":   {},
	"cl.f":   {},
}

// PublicMarketSymbolHandler consulta cotizaciones externas desde el servidor para evitar CORS en el navegador.
func PublicMarketSymbolHandler() http.HandlerFunc {
	client := &http.Client{Timeout: 7 * time.Second}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		symbol := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("s")))
		if _, ok := publicMarketSymbols[symbol]; !ok {
			http.Error(w, "symbol not allowed", http.StatusBadRequest)
			return
		}

		url := fmt.Sprintf("https://stooq.com/q/l/?s=%s&f=sd2t2ohlcvn&h&e=json", symbol)
		req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
		if err != nil {
			http.Error(w, "failed to build request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("User-Agent", "PowerfulControlSystem/1.0 market-proxy")

		resp, err := client.Do(req)
		if err != nil {
			writeEmptyMarketSymbol(w)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			writeEmptyMarketSymbol(w)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			writeEmptyMarketSymbol(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=120")
		_ = encodeJSONResponse(w, payload)
	}
}

func writeEmptyMarketSymbol(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=30")
	w.WriteHeader(http.StatusOK)
	_ = encodeJSONResponse(w, map[string]interface{}{"symbols": []interface{}{}})
}
