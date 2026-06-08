package handlers

import "testing"

func TestExtractOCRFieldsDIAN(t *testing.T) {
	text := `TestSetId db98ef26-0c2a-468f-a3d0-31667aba47e1
Prefijo SETP
Numero Resolucion 18760000001
Rango desde 990000000
Rango hasta 995000000
Clave tecnica fc8eac422eba16e22ffd8c6f94b3f40a6e38162c
Pin 12345`
	fields := extractOCRFields("dian", text)
	seen := map[string]string{}
	for _, field := range fields {
		seen[field.Campo] = field.Valor
	}
	for _, key := range []string{"test_set_id", "prefijo", "numero_resolucion", "rango_desde", "rango_hasta", "clave_tecnica", "pin_software"} {
		if seen[key] == "" {
			t.Fatalf("campo OCR DIAN no detectado: %s; fields=%#v", key, fields)
		}
	}
	if seen["prefijo"] != "SETP" {
		t.Fatalf("prefijo esperado SETP, got %q", seen["prefijo"])
	}
}
