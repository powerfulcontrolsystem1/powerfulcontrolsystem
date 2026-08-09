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
		ProveedorNombre:   "  Papeleria Central  ",
		ProveedorNIT:      "  900123456-7 ",
		DocumentoNumero:   "  FAC-100  ",
		Subtotal:          100000,
		ImpuestoIVA:       19000,
		RetencionFuente:   2500,
		RetencionICA:      700,
		RetencionIVA:      -900,
		Total:             -1,
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
	if row.RetencionIVA != 0 {
		t.Fatalf("retencion negativa saneada = %.2f", row.RetencionIVA)
	}
	if row.ProveedorNombre != "Papeleria Central" || row.ProveedorNIT != "900123456-7" || row.DocumentoNumero != "FAC-100" {
		t.Fatalf("campos de tercero/documento no saneados: %#v", row)
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

func TestSoporteComprasIAEstadoAbierto(t *testing.T) {
	for _, estado := range []string{"radicado", "extraido", "en_revision", "aprobado"} {
		if !soporteIAEstadoAbierto(estado) {
			t.Fatalf("estado %q debe considerarse abierto", estado)
		}
	}
	for _, estado := range []string{"contabilizado", "rechazado", "duplicado"} {
		if soporteIAEstadoAbierto(estado) {
			t.Fatalf("estado %q no debe considerarse abierto", estado)
		}
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

func TestSoporteComprasIAEstadoFiltroVacioListaTodos(t *testing.T) {
	if got := normalizeSoporteIAEstadoFiltro(""); got != "" {
		t.Fatalf("filtro vacio = %q, want sin filtro", got)
	}
	if got := normalizeSoporteIAEstadoFiltro(" extraido "); got != "extraido" {
		t.Fatalf("filtro extraido = %q", got)
	}
	if got := normalizeSoporteIAEstadoFiltro("desconocido"); got != "radicado" {
		t.Fatalf("filtro invalido = %q, want radicado cerrado", got)
	}
}

func TestSoporteComprasIAEstadoExtraibleExcludesConfirmedStates(t *testing.T) {
	for _, estado := range []string{"radicado", "extraido", "en_revision"} {
		if !soporteIAEstadoExtraible(estado) {
			t.Fatalf("estado %q debe permitir extraccion", estado)
		}
	}
	for _, estado := range []string{"aprobado", "rechazado", "contabilizado", "duplicado"} {
		if soporteIAEstadoExtraible(estado) {
			t.Fatalf("estado confirmado %q no debe permitir extraccion", estado)
		}
	}
}

func TestSoporteComprasIAStateTransitionsAreIdempotentAndClosed(t *testing.T) {
	tests := []struct {
		current, next string
		idempotent    bool
		allowed       bool
	}{
		{"radicado", "aprobado", false, true},
		{"extraido", "aprobado", false, true},
		{"en_revision", "aprobado", false, true},
		{"aprobado", "aprobado", true, true},
		{"aprobado", "rechazado", false, true},
		{"rechazado", "rechazado", true, true},
		{"rechazado", "aprobado", false, false},
		{"duplicado", "aprobado", false, false},
		{"contabilizado", "rechazado", false, false},
		{"radicado", "contabilizado", false, false},
	}
	for _, tt := range tests {
		gotIdempotent, err := validateSoporteIAStateTransition(tt.current, tt.next)
		if (err == nil) != tt.allowed || gotIdempotent != tt.idempotent {
			t.Fatalf("transition %s -> %s: idempotent=%v err=%v", tt.current, tt.next, gotIdempotent, err)
		}
	}
}

func TestSoporteComprasIARegistroTransitionsPreserveAccountingTrace(t *testing.T) {
	tests := []struct {
		name, current, workflow, next string
		converted                     int64
		idempotent                    bool
		allowed                       bool
	}{
		{"eliminar radicado", "activo", "radicado", "eliminado", 0, false, true},
		{"eliminar rechazado", "activo", "rechazado", "eliminado", 0, false, true},
		{"recuperar", "eliminado", "en_revision", "activo", 0, false, true},
		{"eliminar idempotente", "eliminado", "radicado", "eliminado", 0, true, true},
		{"recuperar idempotente", "activo", "radicado", "activo", 0, true, true},
		{"bloquear contabilizado", "activo", "contabilizado", "eliminado", 12, false, false},
		{"bloquear convertido", "activo", "aprobado", "eliminado", 12, false, false},
		{"bloquear archivado", "archivado", "radicado", "activo", 0, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIdempotent, err := validateSoporteIARegistroTransition(tt.current, tt.workflow, tt.converted, tt.next)
			if (err == nil) != tt.allowed || gotIdempotent != tt.idempotent {
				t.Fatalf("transition %s -> %s: idempotent=%v err=%v", tt.current, tt.next, gotIdempotent, err)
			}
		})
	}
}

func TestNormalizeSoporteIAEstadoRegistroKeepsRecoverableAndLegacyStates(t *testing.T) {
	for input, want := range map[string]string{
		"":          "activo",
		"ACTIVO":    "activo",
		"eliminado": "eliminado",
		"archivado": "archivado",
		"inactivo":  "inactivo",
		"inventado": "activo",
	} {
		if got := normalizeSoporteIAEstadoRegistro(input); got != want {
			t.Fatalf("normalize %q = %q, want %q", input, got, want)
		}
	}
}

func TestSoporteComprasIARegistroFiltroFailsClosedToActive(t *testing.T) {
	if got := normalizeSoporteIARegistroFiltro("eliminado"); got != "eliminado" {
		t.Fatalf("papelera = %q", got)
	}
	for _, input := range []string{"", "activo", "archivado", "inactivo", "todos", "inventado"} {
		if got := normalizeSoporteIARegistroFiltro(input); got != "activo" {
			t.Fatalf("filtro %q = %q, want activo", input, got)
		}
	}
}

func TestSoporteComprasIAOperacionFailsClosedForUnknownRecordState(t *testing.T) {
	for _, input := range []string{"", "eliminado", "archivado", "inactivo", "inventado"} {
		if soporteIARegistroActivo(input) {
			t.Fatalf("estado %q no debe habilitar acciones operativas", input)
		}
	}
	if !soporteIARegistroActivo(" ACTIVO ") {
		t.Fatal("estado activo normalizado debe habilitar acciones")
	}
	if _, err := validateSoporteIARegistroTransition("inventado", "radicado", 0, "eliminado"); err == nil {
		t.Fatal("un estado persistido desconocido no debe convertirse implicitamente en activo")
	}
}
