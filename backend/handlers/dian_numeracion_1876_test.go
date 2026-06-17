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
