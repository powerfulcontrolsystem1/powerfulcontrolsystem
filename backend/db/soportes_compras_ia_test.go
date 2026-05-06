package db

import "testing"

func TestEmpresaSoporteComprasIAHashBytes(t *testing.T) {
	got := EmpresaSoporteComprasIAHashBytes([]byte("factura-demo"))
	want := "53b5a2def129dc2cc2d3929409704e353c9580470408239c7240d60da4939e90"
	if got != want {
		t.Fatalf("hash = %q, want %q", got, want)
	}
}

func TestNormalizeEmpresaSoporteComprasIA(t *testing.T) {
	row := NormalizeEmpresaSoporteComprasIA(EmpresaSoporteComprasIA{
		TipoSoporte:       "desconocido",
		DocumentoTipo:     "otro",
		Subtotal:          100000,
		ImpuestoIVA:       19000,
		RetencionFuente:   2500,
		RetencionICA:      700,
		ConfianzaIA:       1.5,
		ImpactaInventario: true,
	})
	if row.TipoSoporte != "gasto" {
		t.Fatalf("tipo default = %q", row.TipoSoporte)
	}
	if row.DocumentoTipo != "otro" {
		t.Fatalf("documento tipo = %q", row.DocumentoTipo)
	}
	if row.Total != 115800 {
		t.Fatalf("total calculado = %.2f", row.Total)
	}
	if row.Moneda != "COP" {
		t.Fatalf("moneda default = %q", row.Moneda)
	}
	if row.ModeloIA != EmpresaSoporteComprasIAModeloDefault {
		t.Fatalf("modelo default = %q", row.ModeloIA)
	}
	if row.ConfianzaIA != 1 {
		t.Fatalf("confianza limitada = %.2f", row.ConfianzaIA)
	}
	if row.EstadoSoporte != "radicado" {
		t.Fatalf("estado soporte default = %q", row.EstadoSoporte)
	}
}

func TestSoporteComprasIANormalizaciones(t *testing.T) {
	if got := normalizeSoporteIAOrigen("pdf"); got != "pdf" {
		t.Fatalf("origen pdf = %q", got)
	}
	if got := normalizeSoporteIAEstado("extraido"); got != "extraido" {
		t.Fatalf("estado extraido = %q", got)
	}
	if got := normalizeSoporteIADocumentoTipo("cuenta_cobro"); got != "cuenta_cobro" {
		t.Fatalf("documento cuenta cobro = %q", got)
	}
	if got := normalizeSoporteIAOrigen("correo-fisico"); got != "manual" {
		t.Fatalf("origen default = %q", got)
	}
}
