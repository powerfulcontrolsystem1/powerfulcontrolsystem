package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaImpresorasAgenteHandlerLimitado(t *testing.T) {
	raw, err := os.ReadFile("empresa_impresoras.go")
	if err != nil {
		t.Fatalf("read handler: %v", err)
	}
	src := string(raw)
	start := strings.Index(src, "func EmpresaImpresorasAgenteHandler(")
	if start < 0 {
		t.Fatal("debe existir EmpresaImpresorasAgenteHandler para el agente local")
	}
	body := src[start:]
	for _, required := range []string{
		"TomarEmpresaImpresoraTrabajos",
		"ActualizarEmpresaImpresoraTrabajoEstado",
		"agente_id requerido",
		"Metodo no permitido",
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("EmpresaImpresorasAgenteHandler debe conservar %s en: %s", required, body)
		}
	}
	for _, forbidden := range []string{
		"UpsertEmpresaImpresora(",
		"UpsertEmpresaImpresoraFuncionalidad",
		"UpsertEmpresaImpresoraProducto",
		"SetEmpresaImpresoraPredeterminada",
	} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("el handler de agente no debe administrar configuracion; encontro %s", forbidden)
		}
	}
}
