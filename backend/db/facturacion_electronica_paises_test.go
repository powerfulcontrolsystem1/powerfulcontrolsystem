package db

import (
	"encoding/json"
	"testing"
)

func TestListPaisesFacturacionDisponiblesIncluyePerfilesMultiPais(t *testing.T) {
	got := ListPaisesFacturacionDisponibles()
	want := map[string]string{
		"CO": "COP",
		"EC": "USD",
		"PA": "PAB",
		"CR": "CRC",
		"AR": "ARS",
		"VE": "VES",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d countries, got %d: %#v", len(want), len(got), got)
	}
	for _, pais := range got {
		if want[pais.Codigo] != pais.Moneda {
			t.Fatalf("unexpected country profile: %#v", pais)
		}
		delete(want, pais.Codigo)
	}
	if len(want) != 0 {
		t.Fatalf("missing countries: %#v", want)
	}
}

func TestDetectFacturacionPaisPorTimezoneEIdioma(t *testing.T) {
	cases := []struct {
		name     string
		timezone string
		language string
		wantCode string
		wantFrom string
	}{
		{name: "costa rica timezone", timezone: "America/Costa_Rica", wantCode: "CR", wantFrom: "timezone"},
		{name: "argentina timezone", timezone: "America/Argentina/Buenos_Aires", wantCode: "AR", wantFrom: "timezone"},
		{name: "venezuela timezone", timezone: "America/Caracas", wantCode: "VE", wantFrom: "timezone"},
		{name: "panama language", language: "es-PA,es;q=0.9", wantCode: "PA", wantFrom: "language"},
		{name: "default colombia", wantCode: "CO", wantFrom: "default"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, source, err := DetectFacturacionPais(nil, 0, tc.timezone, tc.language)
			if err != nil {
				t.Fatalf("DetectFacturacionPais returned error: %v", err)
			}
			if got.Codigo != tc.wantCode || source != tc.wantFrom {
				t.Fatalf("expected %s/%s, got %s/%s", tc.wantCode, tc.wantFrom, got.Codigo, source)
			}
		})
	}
}

func TestDefaultFacturacionConfigPaisAplicaProveedorYCampos(t *testing.T) {
	cases := []struct {
		code       string
		proveedor  string
		documento  string
		prefijo    string
		jsonKey    string
		jsonValue  string
		monedaCode string
	}{
		{code: "CR", proveedor: "hacienda_cr", documento: "CEDULA", prefijo: "001-00001", jsonKey: "integracion", jsonValue: "hacienda_api_xml", monedaCode: "CRC"},
		{code: "AR", proveedor: "arca_wsfev1", documento: "CUIT", prefijo: "0001", jsonKey: "ws_servicio", jsonValue: "wsfev1", monedaCode: "ARS"},
		{code: "VE", proveedor: "seniat_imprenta_digital", documento: "RIF", prefijo: "A", jsonKey: "integracion", jsonValue: "seniat_facturacion_digital", monedaCode: "VES"},
		{code: "EC", proveedor: "sri_ecuador", documento: "RUC", prefijo: "001-001", jsonKey: "integracion", jsonValue: "sri_xml_firmado", monedaCode: "USD"},
		{code: "PA", proveedor: "dgi_panama_pac", documento: "RUC", prefijo: "FE", jsonKey: "modalidad", jsonValue: "pac_o_facturador_gratuito", monedaCode: "PAB"},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			cfg := defaultFacturacionConfig(1, tc.code)
			if cfg.Proveedor != tc.proveedor || cfg.TipoDocumentoEmisor != tc.documento || cfg.PrefijoFactura != tc.prefijo || cfg.MonedaCodigo != tc.monedaCode {
				t.Fatalf("unexpected defaults for %s: %#v", tc.code, cfg)
			}
			var extra map[string]interface{}
			if err := json.Unmarshal([]byte(cfg.CamposPaisJSON), &extra); err != nil {
				t.Fatalf("invalid country json: %v", err)
			}
			if extra[tc.jsonKey] != tc.jsonValue {
				t.Fatalf("expected JSON %s=%q, got %#v", tc.jsonKey, tc.jsonValue, extra)
			}
		})
	}
}
