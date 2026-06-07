package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const bolsaYahooChartURL = "https://query1.finance.yahoo.com/v8/finance/chart/"

type bolsaIndicatorDef struct {
	Symbol  string
	Nombre  string
	Grupo   string
	Mercado string
	Tipo    string
}

type bolsaIndicator struct {
	Symbol        string  `json:"symbol"`
	Nombre        string  `json:"nombre"`
	Grupo         string  `json:"grupo"`
	Mercado       string  `json:"mercado"`
	Tipo          string  `json:"tipo"`
	Precio        float64 `json:"precio,omitempty"`
	Anterior      float64 `json:"anterior,omitempty"`
	Cambio        float64 `json:"cambio,omitempty"`
	CambioPct     float64 `json:"cambio_pct,omitempty"`
	Moneda        string  `json:"moneda,omitempty"`
	Exchange      string  `json:"exchange,omitempty"`
	ActualizadoEn string  `json:"actualizado_en,omitempty"`
	Estado        string  `json:"estado"`
	Error         string  `json:"error,omitempty"`
}

type bolsaResponse struct {
	OK            bool                  `json:"ok"`
	EmpresaID     int64                 `json:"empresa_id"`
	Pais          dbpkg.PaisFacturacion `json:"pais"`
	PaisFuente    string                `json:"pais_fuente"`
	ActualizadoEn string                `json:"actualizado_en"`
	FuenteDatos   string                `json:"fuente_datos"`
	Internacional []bolsaIndicator      `json:"internacional"`
	Local         []bolsaIndicator      `json:"local"`
	Resumen       map[string]any        `json:"resumen"`
	Advertencias  []string              `json:"advertencias,omitempty"`
}

type bolsaCacheEntry struct {
	ExpiresAt time.Time
	Data      bolsaResponse
}

var (
	bolsaHTTPClient = &http.Client{Timeout: 12 * time.Second}
	bolsaCacheMu    sync.Mutex
	bolsaCache      = map[string]bolsaCacheEntry{}
)

func EmpresaBolsaHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		empresaID, err := parseEmpresaIDQuery(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pais, paisFuente, err := dbpkg.DetectFacturacionPais(dbEmp, empresaID, r.URL.Query().Get("tz"), r.URL.Query().Get("lang"))
		if err != nil || strings.TrimSpace(pais.Codigo) == "" {
			pais = dbpkg.PaisFacturacion{Codigo: "CO", Nombre: "Colombia", Bandera: "CO", Moneda: "COP"}
			paisFuente = "fallback"
		}
		pais.Codigo = strings.ToUpper(strings.TrimSpace(pais.Codigo))
		if pais.Codigo == "" {
			pais.Codigo = "CO"
		}

		cacheKey := pais.Codigo
		if cached, ok := bolsaReadCache(cacheKey); ok {
			cached.EmpresaID = empresaID
			writeJSON(w, http.StatusOK, cached)
			return
		}

		intDefs := bolsaInternationalIndicators()
		localDefs := bolsaLocalIndicators(pais.Codigo)
		resp := bolsaResponse{
			OK:            true,
			EmpresaID:     empresaID,
			Pais:          pais,
			PaisFuente:    paisFuente,
			ActualizadoEn: time.Now().UTC().Format(time.RFC3339),
			FuenteDatos:   "Yahoo Finance chart API",
			Internacional: bolsaFetchIndicatorSet(r.Context(), intDefs),
			Local:         bolsaFetchIndicatorSet(r.Context(), localDefs),
		}
		resp.Advertencias = bolsaBuildWarnings(resp)
		resp.Resumen = bolsaBuildSummary(resp)
		bolsaWriteCache(cacheKey, resp, 60*time.Second)
		writeJSON(w, http.StatusOK, resp)
	}
}

func bolsaReadCache(key string) (bolsaResponse, bool) {
	bolsaCacheMu.Lock()
	defer bolsaCacheMu.Unlock()
	item, ok := bolsaCache[key]
	if !ok || time.Now().After(item.ExpiresAt) {
		return bolsaResponse{}, false
	}
	return item.Data, true
}

func bolsaWriteCache(key string, data bolsaResponse, ttl time.Duration) {
	bolsaCacheMu.Lock()
	defer bolsaCacheMu.Unlock()
	bolsaCache[key] = bolsaCacheEntry{ExpiresAt: time.Now().Add(ttl), Data: data}
}

func bolsaInternationalIndicators() []bolsaIndicatorDef {
	return []bolsaIndicatorDef{
		{Symbol: "^GSPC", Nombre: "S&P 500", Grupo: "internacional", Mercado: "Estados Unidos", Tipo: "Indice"},
		{Symbol: "^IXIC", Nombre: "Nasdaq Composite", Grupo: "internacional", Mercado: "Estados Unidos", Tipo: "Indice"},
		{Symbol: "^DJI", Nombre: "Dow Jones", Grupo: "internacional", Mercado: "Estados Unidos", Tipo: "Indice"},
		{Symbol: "^FTSE", Nombre: "FTSE 100", Grupo: "internacional", Mercado: "Reino Unido", Tipo: "Indice"},
		{Symbol: "^N225", Nombre: "Nikkei 225", Grupo: "internacional", Mercado: "Japon", Tipo: "Indice"},
		{Symbol: "GC=F", Nombre: "Oro", Grupo: "internacional", Mercado: "Commodities", Tipo: "Materia prima"},
		{Symbol: "CL=F", Nombre: "Petroleo WTI", Grupo: "internacional", Mercado: "Commodities", Tipo: "Materia prima"},
		{Symbol: "BTC-USD", Nombre: "Bitcoin", Grupo: "internacional", Mercado: "Cripto", Tipo: "Criptoactivo"},
		{Symbol: "EURUSD=X", Nombre: "EUR/USD", Grupo: "internacional", Mercado: "Divisas", Tipo: "Divisa"},
	}
}

func bolsaLocalIndicators(pais string) []bolsaIndicatorDef {
	switch strings.ToUpper(strings.TrimSpace(pais)) {
	case "AR":
		return []bolsaIndicatorDef{
			{Symbol: "^MERV", Nombre: "S&P Merval", Grupo: "local", Mercado: "Argentina", Tipo: "Indice"},
			{Symbol: "ARS=X", Nombre: "USD/ARS", Grupo: "local", Mercado: "Argentina", Tipo: "Divisa"},
			{Symbol: "ARGT", Nombre: "ETF Argentina", Grupo: "local", Mercado: "Argentina", Tipo: "ETF"},
			{Symbol: "GGAL", Nombre: "Grupo Financiero Galicia ADR", Grupo: "local", Mercado: "Argentina", Tipo: "Accion"},
			{Symbol: "YPF", Nombre: "YPF ADR", Grupo: "local", Mercado: "Argentina", Tipo: "Accion"},
		}
	case "CR":
		return []bolsaIndicatorDef{
			{Symbol: "USDCRC=X", Nombre: "USD/CRC", Grupo: "local", Mercado: "Costa Rica", Tipo: "Divisa"},
			{Symbol: "^GSPC", Nombre: "S&P 500 referencia USD", Grupo: "local", Mercado: "Costa Rica", Tipo: "Referencia"},
			{Symbol: "GC=F", Nombre: "Oro referencia", Grupo: "local", Mercado: "Costa Rica", Tipo: "Referencia"},
			{Symbol: "CL=F", Nombre: "Petroleo WTI referencia", Grupo: "local", Mercado: "Costa Rica", Tipo: "Referencia"},
		}
	case "EC":
		return []bolsaIndicatorDef{
			{Symbol: "EURUSD=X", Nombre: "EUR/USD", Grupo: "local", Mercado: "Ecuador", Tipo: "Divisa"},
			{Symbol: "BTC-USD", Nombre: "Bitcoin referencia USD", Grupo: "local", Mercado: "Ecuador", Tipo: "Referencia"},
			{Symbol: "GC=F", Nombre: "Oro referencia", Grupo: "local", Mercado: "Ecuador", Tipo: "Referencia"},
			{Symbol: "CL=F", Nombre: "Petroleo WTI referencia", Grupo: "local", Mercado: "Ecuador", Tipo: "Referencia"},
		}
	case "PA":
		return []bolsaIndicatorDef{
			{Symbol: "PABUSD=X", Nombre: "PAB/USD", Grupo: "local", Mercado: "Panama", Tipo: "Divisa"},
			{Symbol: "EURUSD=X", Nombre: "EUR/USD", Grupo: "local", Mercado: "Panama", Tipo: "Divisa"},
			{Symbol: "BTC-USD", Nombre: "Bitcoin referencia USD", Grupo: "local", Mercado: "Panama", Tipo: "Referencia"},
			{Symbol: "GC=F", Nombre: "Oro referencia", Grupo: "local", Mercado: "Panama", Tipo: "Referencia"},
		}
	case "VE":
		return []bolsaIndicatorDef{
			{Symbol: "USDVES=X", Nombre: "USD/VES", Grupo: "local", Mercado: "Venezuela", Tipo: "Divisa"},
			{Symbol: "BTC-USD", Nombre: "Bitcoin referencia USD", Grupo: "local", Mercado: "Venezuela", Tipo: "Referencia"},
			{Symbol: "GC=F", Nombre: "Oro referencia", Grupo: "local", Mercado: "Venezuela", Tipo: "Referencia"},
			{Symbol: "CL=F", Nombre: "Petroleo WTI referencia", Grupo: "local", Mercado: "Venezuela", Tipo: "Referencia"},
		}
	default:
		return []bolsaIndicatorDef{
			{Symbol: "USDCOP=X", Nombre: "USD/COP", Grupo: "local", Mercado: "Colombia", Tipo: "Divisa"},
			{Symbol: "ICOLCAP.CL", Nombre: "ETF iCOLCAP BVC", Grupo: "local", Mercado: "Colombia", Tipo: "ETF"},
			{Symbol: "GXG", Nombre: "ETF Colombia Global X", Grupo: "local", Mercado: "Colombia", Tipo: "ETF"},
			{Symbol: "ECOPETROL.CL", Nombre: "Ecopetrol BVC", Grupo: "local", Mercado: "Colombia", Tipo: "Accion"},
			{Symbol: "CIB", Nombre: "Bancolombia ADR", Grupo: "local", Mercado: "Colombia", Tipo: "Accion"},
			{Symbol: "EC", Nombre: "Ecopetrol ADR", Grupo: "local", Mercado: "Colombia", Tipo: "Accion"},
		}
	}
}

func bolsaFetchIndicatorSet(ctx context.Context, defs []bolsaIndicatorDef) []bolsaIndicator {
	out := make([]bolsaIndicator, 0, len(defs))
	for _, def := range defs {
		out = append(out, bolsaFetchIndicator(ctx, def))
	}
	return out
}

func bolsaFetchIndicator(ctx context.Context, def bolsaIndicatorDef) bolsaIndicator {
	item := bolsaIndicator{
		Symbol:  def.Symbol,
		Nombre:  def.Nombre,
		Grupo:   def.Grupo,
		Mercado: def.Mercado,
		Tipo:    def.Tipo,
		Estado:  "error",
	}
	reqURL := bolsaYahooChartURL + url.PathEscape(def.Symbol) + "?range=1d&interval=5m&includePrePost=false"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		item.Error = "no se pudo preparar consulta"
		return item
	}
	req.Header.Set("User-Agent", "PowerfulControlSystem/1.0")
	res, err := bolsaHTTPClient.Do(req)
	if err != nil {
		item.Error = "proveedor no disponible"
		return item
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode > 299 {
		item.Error = fmt.Sprintf("proveedor HTTP %d", res.StatusCode)
		return item
	}
	var raw struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency                   string  `json:"currency"`
					Symbol                     string  `json:"symbol"`
					ExchangeName               string  `json:"exchangeName"`
					FullExchangeName           string  `json:"fullExchangeName"`
					InstrumentType             string  `json:"instrumentType"`
					RegularMarketPrice         float64 `json:"regularMarketPrice"`
					RegularMarketTime          int64   `json:"regularMarketTime"`
					ChartPreviousClose         float64 `json:"chartPreviousClose"`
					RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
				} `json:"meta"`
			} `json:"result"`
			Error *struct {
				Code        string `json:"code"`
				Description string `json:"description"`
			} `json:"error"`
		} `json:"chart"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		item.Error = "respuesta no interpretable"
		return item
	}
	if raw.Chart.Error != nil {
		item.Error = strings.TrimSpace(raw.Chart.Error.Description)
		if item.Error == "" {
			item.Error = raw.Chart.Error.Code
		}
		return item
	}
	if len(raw.Chart.Result) == 0 {
		item.Error = "sin datos"
		return item
	}
	meta := raw.Chart.Result[0].Meta
	price := meta.RegularMarketPrice
	prev := meta.RegularMarketPreviousClose
	if prev == 0 {
		prev = meta.ChartPreviousClose
	}
	if price == 0 {
		item.Error = "sin precio vigente"
		return item
	}
	item.Estado = "ok"
	item.Error = ""
	item.Precio = bolsaRound(price, 4)
	item.Anterior = bolsaRound(prev, 4)
	if prev != 0 {
		change := price - prev
		item.Cambio = bolsaRound(change, 4)
		item.CambioPct = bolsaRound((change/prev)*100, 4)
	}
	item.Moneda = strings.TrimSpace(meta.Currency)
	item.Exchange = strings.TrimSpace(meta.FullExchangeName)
	if item.Exchange == "" {
		item.Exchange = strings.TrimSpace(meta.ExchangeName)
	}
	if meta.InstrumentType != "" {
		item.Tipo = meta.InstrumentType
	}
	if meta.RegularMarketTime > 0 {
		item.ActualizadoEn = time.Unix(meta.RegularMarketTime, 0).UTC().Format(time.RFC3339)
	}
	return item
}

func bolsaRound(v float64, places int) float64 {
	if places < 0 {
		places = 0
	}
	pow := math.Pow10(places)
	return math.Round(v*pow) / pow
}

func bolsaBuildWarnings(resp bolsaResponse) []string {
	warnings := []string{
		"Informacion de mercado para referencia operativa. No constituye recomendacion de inversion.",
	}
	failed := 0
	total := 0
	for _, group := range [][]bolsaIndicator{resp.Internacional, resp.Local} {
		for _, item := range group {
			total++
			if item.Estado != "ok" {
				failed++
			}
		}
	}
	if failed > 0 {
		warnings = append(warnings, fmt.Sprintf("%d de %d indicadores no respondieron desde el proveedor.", failed, total))
	}
	return warnings
}

func bolsaBuildSummary(resp bolsaResponse) map[string]any {
	total := 0
	ok := 0
	up := 0
	down := 0
	for _, group := range [][]bolsaIndicator{resp.Internacional, resp.Local} {
		for _, item := range group {
			total++
			if item.Estado == "ok" {
				ok++
				if item.CambioPct > 0 {
					up++
				} else if item.CambioPct < 0 {
					down++
				}
			}
		}
	}
	return map[string]any{
		"total":        total,
		"disponibles":  ok,
		"al_alza":      up,
		"a_la_baja":    down,
		"sin_cambio":   ok - up - down,
		"cache_seg":    60,
		"pais_codigo":  resp.Pais.Codigo,
		"pais_nombre":  resp.Pais.Nombre,
		"moneda_local": resp.Pais.Moneda,
	}
}
