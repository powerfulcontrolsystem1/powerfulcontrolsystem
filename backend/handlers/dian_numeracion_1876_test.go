package handlers

import "testing"

func TestParseDIANNumeracion1876Text(t *testing.T) {
	text := `
4. Numero de formulario
18764111203411
      8 4 4 5 6 7 7 9 1 CAYON GUARNIZO IVAN FRANCISCO
Impuestos y Aduanas de Santa Marta 1 9
SUBDIRECCION DE FACTURA ELECTRONICA Y SOLUCI
2 0 2 6 -0 6 -1 6 /0 1 :5 6 :5 4
Rangos de numeracion para autorizar, habilitar o inhabilitar
FACTURA ELECTRONICA DE VENTA 4
PCS
1
1,000,000
AUTORIZACION  1
24
`
	fields, warnings := parseDIANNumeracion1876Text(text)
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v", warnings)
	}
	assertField := func(key string, want interface{}) {
		t.Helper()
		if got := fields[key]; got != want {
			t.Fatalf("%s = %#v, want %#v; fields=%#v", key, got, want, fields)
		}
	}
	assertField("numero_formulario", "18764111203411")
	assertField("resolucion_numero", "18764111203411")
	assertField("nit", "84456779")
	assertField("dv", "1")
	assertField("razon_social", "CAYON GUARNIZO IVAN FRANCISCO")
	assertField("prefijo", "PCS")
	assertField("rango_desde", int64(1))
	assertField("rango_hasta", int64(1000000))
	assertField("resolucion_fecha_desde", "2026-06-16")
	assertField("resolucion_fecha_hasta", "2028-06-16")
	assertField("tipo_ambiente", "produccion")
}

func TestParseDIANNumeracion1876TextKeepsNumericPrefix(t *testing.T) {
	text := `
4. Numero de formulario
18764111318575
      8 4 4 5 6 7 7 9 1 CAYON GUARNIZO IVAN FRANCISCO
2 0 2 6 -0 6 -1 7 /0 1 :5 1 :1 1
Rangos de numeracion para autorizar, habilitar o inhabilitar
FACTURA ELECTRONICA DE VENTA 4



1PCS
1



100,000



AUTORIZACION  1



24
`
	fields, warnings := parseDIANNumeracion1876Text(text)
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v", warnings)
	}
	if got := fields["prefijo"]; got != "1PCS" {
		t.Fatalf("prefijo = %#v, want 1PCS; fields=%#v", got, fields)
	}
	if got := fields["numero_formulario"]; got != "18764111318575" {
		t.Fatalf("numero_formulario = %#v, want 18764111318575", got)
	}
	if got := fields["rango_hasta"]; got != int64(100000) {
		t.Fatalf("rango_hasta = %#v, want 100000", got)
	}
}
