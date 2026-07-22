package handlers

import (
	"os"
	"strings"
	"testing"
)

func TestEmpresaBuzonNoExponeErroresInternosEnCodigoPublico(t *testing.T) {
	content, err := os.ReadFile("empresa_buzon.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(content)
	for _, fragment := range []string{
		`"No se pudo guardar configuracion: "+err.Error()`,
		`"No se pudo limpiar archivos: "+err.Error()`,
	} {
		if strings.Contains(source, fragment) {
			t.Fatalf("el handler volvio a exponer causa interna: %s", fragment)
		}
	}
}
