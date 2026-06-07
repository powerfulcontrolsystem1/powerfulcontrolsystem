package handlers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type bolsaRoundTripFunc func(*http.Request) (*http.Response, error)

func (f bolsaRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestBolsaFetchIndicatorCalculatesChange(t *testing.T) {
	oldClient := bolsaHTTPClient
	t.Cleanup(func() { bolsaHTTPClient = oldClient })
	bolsaHTTPClient = &http.Client{Transport: bolsaRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := `{"chart":{"result":[{"meta":{"currency":"USD","symbol":"TEST","exchangeName":"NYS","fullExchangeName":"NYSE","instrumentType":"INDEX","regularMarketPrice":110,"regularMarketPreviousClose":100,"regularMarketTime":1700000000}}],"error":null}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}

	got := bolsaFetchIndicator(context.Background(), bolsaIndicatorDef{
		Symbol:  "TEST",
		Nombre:  "Prueba",
		Grupo:   "internacional",
		Mercado: "Test",
		Tipo:    "Indice",
	})
	if got.Estado != "ok" {
		t.Fatalf("expected ok, got %#v", got)
	}
	if got.Precio != 110 || got.Anterior != 100 || got.Cambio != 10 || got.CambioPct != 10 {
		t.Fatalf("unexpected market math: %#v", got)
	}
	if got.Exchange != "NYSE" || got.Moneda != "USD" {
		t.Fatalf("unexpected metadata: %#v", got)
	}
}

func TestBolsaLocalIndicatorsColombia(t *testing.T) {
	got := bolsaLocalIndicators("CO")
	if len(got) < 5 {
		t.Fatalf("expected Colombia indicators, got %d", len(got))
	}
	want := map[string]bool{"USDCOP=X": false, "ICOLCAP.CL": false, "ECOPETROL.CL": false}
	for _, item := range got {
		if _, ok := want[item.Symbol]; ok {
			want[item.Symbol] = true
		}
	}
	for symbol, ok := range want {
		if !ok {
			t.Fatalf("missing Colombia indicator %s in %#v", symbol, got)
		}
	}
}
