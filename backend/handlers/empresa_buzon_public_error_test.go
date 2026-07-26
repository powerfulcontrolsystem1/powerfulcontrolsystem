package handlers

import (
	"github.com/you/pos-backend/db"
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

func TestEmpresaBuzonSanitizaAlertasDIANHistoricasConRutaPrivada(t *testing.T) {
	t.Parallel()
	items := sanitizeEmpresaBuzonMessagesForResponse([]db.EmpresaBuzonMensaje{{
		Modulo:  "facturacion_electronica",
		Mensaje: "firmar XML: open ../web/uploads/empresas/firma_privada.pem: permission denied",
	}})
	if len(items) != 1 || strings.Contains(items[0].Mensaje, "../web/uploads") || !strings.Contains(items[0].Mensaje, "detalle técnico fue ocultado") {
		t.Fatalf("historical DIAN alert must be sanitized: %#v", items)
	}
}
