package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSanitizeChatFlotanteRadioStations(t *testing.T) {
	raw := json.RawMessage(`[
		{"name":"Mi Emisora","genre":"Noticias","countryCode":"PA","streamUrl":"https://example.com/radio.mp3","sourceUrl":"https://example.com"},
		{"name":"Sin URL","streamUrl":"ftp://example.com/audio"},
		{"name":"Ecuador Radio","countryCode":"EC","streamUrl":"http://example.org/live"}
	]`)

	items, encoded, err := sanitizeChatFlotanteRadioStations(raw)
	if err != nil {
		t.Fatalf("sanitizeChatFlotanteRadioStations returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items)=%d, want 2", len(items))
	}
	if items[0].CountryCode != "PA" || !items[0].Custom {
		t.Fatalf("first station=%+v, want PA custom station", items[0])
	}
	if items[1].CountryCode != "EC" || items[1].SourceURL != "" {
		t.Fatalf("second station=%+v, want EC without source", items[1])
	}
	var decoded []chatFlotanteRadioStationPref
	if err := json.Unmarshal([]byte(encoded), &decoded); err != nil {
		t.Fatalf("encoded stations are invalid JSON: %v", err)
	}
}

func TestNormalizeChatFlotanteRadioCountry(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want string
	}{
		{raw: "pa", want: "PA"},
		{raw: "Panama", want: "PA"},
		{raw: "EC", want: "EC"},
		{raw: "Colombia", want: ""},
	} {
		if got := normalizeChatFlotanteRadioCountry(tc.raw); got != tc.want {
			t.Fatalf("normalizeChatFlotanteRadioCountry(%q)=%q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestNormalizeChatFlotanteAppearance(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{raw: "corporate", want: "corporativo"},
		{raw: "OCEANO", want: "oceano"},
		{raw: "unknown", want: "normal"},
	} {
		if got := normalizeChatFlotanteTheme(tc.raw); got != tc.want {
			t.Fatalf("theme %q = %q, want %q", tc.raw, got, tc.want)
		}
	}
	for _, tc := range []struct{ raw, want string }{
		{raw: "small", want: "pequeno"},
		{raw: "grande", want: "grande"},
		{raw: "invalid", want: "mediano"},
	} {
		if got := normalizeChatFlotanteTextSize(tc.raw); got != tc.want {
			t.Fatalf("text size %q = %q, want %q", tc.raw, got, tc.want)
		}
	}
}

func TestChatFlotanteChatEnabledAlwaysActiveForEmpresa(t *testing.T) {
	if got := getChatFlotanteBoolForEmpresa(nil, nil, 12, chatFlotanteChatEnabledKey, false); !got {
		t.Fatalf("chat flotante empresarial desactivado; want activo para empresas")
	}
	if got := chatFlotantePrefsResponse(nil, nil, 12)["chat_enabled"]; got != true {
		t.Fatalf("chat_enabled=%v, want true para empresa", got)
	}
	if got := chatFlotantePrefsResponse(nil, nil, 12)["theme"]; got != "normal" {
		t.Fatalf("theme=%v, want normal", got)
	}
	if got := chatFlotantePrefsResponse(nil, nil, 12)["text_size"]; got != "mediano" {
		t.Fatalf("text_size=%v, want mediano", got)
	}
}

func TestChatFlotanteRequestUsesContextEmpresaBeforeClientValue(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "/api/chat_flotante/preferencias?empresa_id=99", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.WithValue(req.Context(), "empresaID", int64(12)))
	got, err := parseInt64QueryOptional(req, "empresa_id")
	if err != nil || got != 12 {
		t.Fatalf("empresa id = %d, err=%v; want authenticated context 12", got, err)
	}
}
