package handlers

import (
	"encoding/json"
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

func TestChatFlotanteChatEnabledAlwaysActiveForEmpresa(t *testing.T) {
	if got := getChatFlotanteBoolForEmpresa(nil, nil, 12, chatFlotanteChatEnabledKey, false); !got {
		t.Fatalf("chat flotante empresarial desactivado; want activo para empresas")
	}
	if got := chatFlotantePrefsResponse(nil, nil, 12)["chat_enabled"]; got != true {
		t.Fatalf("chat_enabled=%v, want true para empresa", got)
	}
}
